package lab_test

import (
	"context"
	"log"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	rrpc "github.com/rsocket/rsocket-rpc-go"
	pb "github.com/rsocket/rsocket-rpc-go/example/ping"
	. "github.com/rsocket/rsocket-rpc-go/lab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock PingPong Server
type mockServer struct {
}

func (*mockServer) Ping(in *pb.Ping) (out *pb.Pong) {
	log.Println("rcv ping:", in)
	out = &pb.Pong{
		Ball: in.Ball,
	}
	log.Println("snd pong:", out)
	return
}

func TestMock(t *testing.T) {
	const addr = "tcp://127.0.0.1:8989"
	// Prepare server.
	s := rrpc.NewServer()
	RegisterPingPongServer(s, &mockServer{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Serve a server
	go func(ctx context.Context) {
		err := rsocket.Receive().
			Acceptor(s.Acceptor()).
			Transport(addr).
			Serve(context.Background())
		if err != nil {
			log.Printf("serve failed: %+v\n", err)
		}
	}(ctx)

	// Connect with raw client.
	cli, err := rsocket.Connect().
		Transport(addr).
		Start(context.Background())
	require.NoError(t, err, "create raw client failed")
	defer cli.Close()

	// Create request.
	req := &pb.Ping{
		Ball: "Football",
	}
	// Let's GO!
	rp, _ := rrpc.NewRequestPayload("PingPong", "ping", req)
	cli.RequestResponse(rp).
		DoOnSuccess(func(ctx context.Context, s rx.Subscription, elem payload.Payload) {
			pong := new(pb.Pong)
			err := proto.Unmarshal(elem.Data(), pong)
			assert.NoError(t, err, "bad response payload")
			log.Println("--> got response:", pong)
		}).
		Subscribe(ctx)
}
