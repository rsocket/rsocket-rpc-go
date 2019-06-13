package example

import (
	"log"
	"reflect"

	rrpc "github.com/rsocket/rsocket-rpc-go"
	pb "github.com/rsocket/rsocket-rpc-go/example/ping"
)

type PingPong interface {
	Ping(*pb.Ping) *pb.Pong
}

func RegisterPingPongServer(server *rrpc.Server, service PingPong) {
	t := reflect.TypeOf(service)
	log.Println(t.Name())
}
