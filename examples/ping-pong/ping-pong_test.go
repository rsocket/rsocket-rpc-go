package ping_pong_test

import (
	"context"
	"testing"
	"time"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/logger"
	rrpc "github.com/rsocket/rsocket-rpc-go"
	pb "github.com/rsocket/rsocket-rpc-go/examples/ping-pong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pingPongServer struct {
}

func (p *pingPongServer) Ping(ctx context.Context, in *pb.Ping) (*pb.Pong, error) {
	return &pb.Pong{
		Ball: in.Ball,
	}, nil
}

func TestAll(t *testing.T) {
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
	c := pb.NewPingPongClient(rrpc.NewClientConn(rc, nil, nil))
	req := &pb.Ping{
		Ball: "Hello World!",
	}
	res, err := c.Ping(context.Background(), req, rrpc.WithMetadata([]byte("FROM_GOLANG")))
	assert.NoError(t, err, "cannot get response")
	assert.Equal(t, req.Ball, res.Ball, "bad response")
}

func doTest(addr string, t *testing.T) {
	//const addr =
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		ss := &pingPongServer{}
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
	c := pb.NewPingPongClient(rrpc.NewClientConn(rc, nil, nil))
	req := &pb.Ping{
		Ball: "Hello World!",
	}
	res, err := c.Ping(ctx, req)
	assert.NoError(t, err, "cannot get response")
	assert.Equal(t, req.Ball, res.Ball, "bad response")
}
