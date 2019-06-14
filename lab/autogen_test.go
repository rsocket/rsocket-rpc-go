package lab

import (
	"log"
	"testing"

	rrpc "github.com/rsocket/rsocket-rpc-go"
	"github.com/rsocket/rsocket-rpc-go/example/ping"
	"github.com/stretchr/testify/require"
)

// Mock PingPong Server
type mockServer struct {
}

func (*mockServer) Ping(in *ping.Ping) (out *ping.Pong) {
	log.Println("rcv ping:", in)
	out = &ping.Pong{
		Ball: in.Ball,
	}
	log.Println("snd pong:", out)
	return
}

func TestMock(t *testing.T) {
	// Prepare server.
	s := rrpc.NewServer()
	RegisterPingPongServer(s, &mockServer{})

	// Create request.
	req := &ping.Ping{
		Ball: "Football",
	}
	// Let's GO!
	v, err := s.MockRequestResponse("PingPong", "ping", req)
	require.NoError(t, err, "bad handle request")
	res, ok := v.(*ping.Pong)
	if !ok {
		require.Fail(t, "bad response type")
	}
	require.Equal(t, req.Ball, res.Ball, "bad ping-pong ")
	log.Printf("eval result: %v\n", res)
}
