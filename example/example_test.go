package example

import (
	"testing"

	"github.com/rsocket/rsocket-rpc-go/example/ping"
)

type server struct {
}

func (*server) Ping(*ping.Ping) *ping.Pong {
	panic("implement me")
}

func TestName(t *testing.T) {
	RegisterPingPongServer(nil, &server{})
}
