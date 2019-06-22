package ping_pong_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/logger"
	"github.com/rsocket/rsocket-go/rx"
	rrpc "github.com/rsocket/rsocket-rpc-go"
	pb "github.com/rsocket/rsocket-rpc-go/examples/ping-pong"
	"github.com/stretchr/testify/require"
)

func TestServe(t *testing.T) {
	logger.SetLevel(logger.LevelDebug)
	ss := &pingPongServer{}
	s := rrpc.NewServer()
	pb.RegisterPingPongServer(s, ss)
	err := rsocket.Receive().
		Acceptor(s.Acceptor()).
		Transport("tcp://127.0.0.1:7878").
		Serve(context.Background())
	if err != nil {
		panic(err)
	}
}

func TestNewPingPongClient(t *testing.T) {
	addr := "tcp://127.0.0.1:8081"
	logger.SetLevel(logger.LevelDebug)

	time.Sleep(500 * time.Millisecond)

	rc, err := rsocket.Connect().
		Transport(addr).
		Start(context.Background())
	require.NoError(t, err, "cannot create client")
	c := pb.NewPingPongClient(rc, nil, mocktracer.New())
	req := &pb.Ping{
		Ball: "Hello World!",
	}

	c.LotsOfPongs(context.Background(), req).
		DoOnNext(func(i context.Context, subscription rx.Subscription, pong *pb.Pong) {
			log.Println("pong:", pong.Ball)
		}).
		Subscribe(context.Background())

	//res, err := c.Ping(context.Background(), req, rrpc.WithMetadata([]byte("FROM_GOLANG")))
	//assert.NoError(t, err, "cannot get response")
	//assert.Equal(t, req.Ball, res.Ball, "bad response")
}
