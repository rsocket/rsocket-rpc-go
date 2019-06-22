package ping_pong_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/rx"
	rrpc "github.com/rsocket/rsocket-rpc-go"
	pb "github.com/rsocket/rsocket-rpc-go/examples/ping-pong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pingPongServer struct {
	totals int
}

func (p *pingPongServer) LotsOfPongs(ctx context.Context, in *pb.Ping, m rrpc.Metadata) *pb.FluxPong {
	ball := in.GetBall()
	return pb.NewFluxPong(func(ctx context.Context, producer pb.SinkPong) {
		for i := 0; i < p.totals; i++ {
			producer.Next(&pb.Pong{
				Ball: fmt.Sprintf("%s_%d", ball, i),
			})
		}
		producer.Complete()
	})
}

func (p *pingPongServer) Ping(ctx context.Context, in *pb.Ping, m rrpc.Metadata) (*pb.Pong, error) {
	log.Println("rcv metadata:", m)
	if t := m.Tracing(); len(t) > 0 {
		if carrier, err := rrpc.UnmarshallTracingCarrier(t); err == nil {
			log.Println("carrier:", carrier)
		}
	}
	return &pb.Pong{
		Ball: in.Ball,
	}, nil
}

func TestAllInOne(t *testing.T) {
	m := map[string]string{
		"tcp":       "tcp://127.0.0.1:7878",
		"websocket": "ws://127.0.0.1:7878",
	}
	for k, v := range m {
		t.Run(k, func(t *testing.T) {
			doTest(v, t)
		})
	}
}

func doTest(addr string, t *testing.T) {
	//const addr =
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const streamTotals = 10

	go func(ctx context.Context) {
		ss := &pingPongServer{
			totals: streamTotals,
		}
		s := rrpc.NewServer()
		pb.RegisterPingPongServer(s, ss)
		err := rsocket.Receive().
			Acceptor(s.Acceptor()).
			Transport(addr).
			Serve(ctx)
		if err != nil {
			panic(err)
		}
	}(ctx)

	time.Sleep(500 * time.Millisecond)

	rc, err := rsocket.Connect().
		Transport(addr).
		Start(ctx)
	require.NoError(t, err, "cannot create client")
	c := pb.NewPingPongClient(rc, nil, nil)
	req := &pb.Ping{
		Ball: "Hello World!",
	}
	res, err := c.Ping(ctx, req, rrpc.WithMetadata([]byte("this_is_metadata")))
	assert.NoError(t, err, "cannot get response")
	assert.Equal(t, req.Ball, res.Ball, "bad response")

	var totals int32
	c.LotsOfPongs(ctx, req, rrpc.WithMetadata([]byte("this_is_metadata"))).
		DoOnNext(func(i context.Context, subscription rx.Subscription, pong *pb.Pong) {
			ball := pong.GetBall()
			log.Println("pong stream:", pong)
			assert.Equal(t, fmt.Sprintf("%s_%d", req.GetBall(), totals), ball)
			totals++
		}).
		Subscribe(ctx)
	assert.Equal(t, streamTotals, int(totals))
}

//func TestGeneratedFlux(t *testing.T) {
//	fx := pb.NewFluxPong(func(ctx context.Context, producer pb.SinkPong) {
//		for range [10]struct{}{} {
//			producer.Next(&pb.Pong{
//				Ball: time.Now().String(),
//			})
//		}
//		producer.Complete()
//	})
//
//	fx.
//		DoOnNext(func(ctx context.Context, subscription rx.Subscription, pong *pb.Pong) {
//			log.Println("pong:", pong.Ball)
//		}).
//		Subscribe(context.Background())
//
//}
