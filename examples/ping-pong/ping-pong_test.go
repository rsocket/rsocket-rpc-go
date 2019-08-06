package ping_pong_test

import (
	"context"
	"fmt"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	rrpc "github.com/rsocket/rsocket-rpc-go"
	ping_pong "github.com/rsocket/rsocket-rpc-go/examples/ping-pong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"strings"
	"testing"
	"time"
)

func TestAllInOne(t *testing.T) {
	defer func() {
		i := recover()
		fmt.Print("recover -> ")
		fmt.Println(i)
	}()

	m := map[string]string{
		"tcp": "tcp://127.0.0.1:7878",
		//"websocket": "ws://127.0.0.1:7878",
	}
	for k, v := range m {
		t.Run(k, func(t *testing.T) {
			doTest(v, t)
		})
	}
}

type PingPongTestServer struct {
	totals int
	ping_pong.PingPong
}

func (p *PingPongTestServer) Ping(ctx context.Context, in *ping_pong.Ping, m []byte) (<-chan *ping_pong.Pong, <-chan error) {
	log.Println("rcv metadata:", string(m))

	response := make(chan *ping_pong.Pong, 1)
	defer close(response)
	pong := &ping_pong.Pong{
		Ball: in.Ball,
	}

	response <- pong

	return response, nil
}

func (p *PingPongTestServer) LotsOfPongs(ctx context.Context, in *ping_pong.Ping, m []byte) (<-chan *ping_pong.Pong, <-chan error) {
	pongs := make(chan *ping_pong.Pong)
	go func() {
		defer func() {
			close(pongs)
		}()
		ball := in.GetBall()
		for i := 0; i < p.totals; i++ {
			sprintf := fmt.Sprintf("%s_%d", ball, i)
			fmt.Printf("sending a ball -> %s\n", sprintf)
			pong := &ping_pong.Pong{
				Ball: sprintf,
			}
			pongs <- pong
		}
	}()
	return pongs, nil
}

func doTest(addr string, t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context) {
		pingPongServer := ping_pong.NewPingPongServer(&PingPongTestServer{
			totals: 1000,
		})
		err := rsocket.
			Receive().
			Acceptor(
				func(setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) rsocket.RSocket {
					return pingPongServer
				}).
			Transport(addr).Serve(ctx)
		if err != nil {
			panic(err)
		}
	}(ctx)

	time.Sleep(500 * time.Millisecond)

	rSocket, err := rsocket.Connect().
		Transport(addr).
		Start(ctx)
	require.NoError(t, err, "cannot create client")

	pingPongClient := ping_pong.NewPingPongClient(rSocket, nil, nil)
	doPingTest(ctx, t, pingPongClient)
	doLotsOfPingsTest(ctx, t, pingPongClient)
}

func doPingTest(ctx context.Context, t *testing.T, pingPongClient ping_pong.PingPongClient) {
	req := &ping_pong.Ping{
		Ball: "Hello World!",
	}
	pongs, errors := pingPongClient.Ping(ctx, req, rrpc.WithMetadata([]byte("this_is_metadata")))
	select {
	case err := <-errors:
		assert.NoError(t, err, "cannot get response")
	case res, ok := <-pongs:
		if ok {
			assert.Equal(t, req.Ball, res.Ball, "bad response")
		}
	}
}

func doLotsOfPingsTest(ctx context.Context, t *testing.T, pingPongClient ping_pong.PingPongClient) {
	req := &ping_pong.Ping{
		Ball: "Hello World!",
	}
	pongs, errors := pingPongClient.LotsOfPongs(ctx, req, rrpc.WithMetadata([]byte("this_is_metadata")))

loop:
	for {
		//fmt.Println("got here looping")
		select {
		case err := <-errors:
			if err != nil {
				assert.NoError(t, err, "cannot get response")
				break loop
			}
		case res, ok := <-pongs:
			if ok {
				fmt.Printf("receiving a ball -> %s\n", res.Ball)
				assert.Equal(t, res.Ball[:strings.LastIndex(res.Ball, "_")], req.Ball, "bad response")
			} else {
				break loop
			}
		}
	}
}
