// NOTICE: ALL CODES IN THIS FILE SHOULD BE AUTO GENERATED.
package lab

import (
	rrpc "github.com/rsocket/rsocket-rpc-go"
	pb "github.com/rsocket/rsocket-rpc-go/example/ping"
)

var _PingPong_serviceDesc = rrpc.ServiceDesc{
	Name:        "PingPong",
	HandlerType: (*PingPongServer)(nil),
	Methods: []rrpc.MethodDesc{
		{
			Name:    "ping",
			Handler: _PingPong_ping_handler,
		},
	},
}

func _PingPong_ping_handler(srv interface{}, dec func(interface{}) error) (interface{}, error) {
	req := new(pb.Ping)
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(PingPongServer).Ping(req), nil
}

type PingPongServer interface {
	Ping(*pb.Ping) *pb.Pong
}

func RegisterPingPongServer(s *rrpc.Server, srv PingPongServer) {
	s.RegisterService(&_PingPong_serviceDesc, srv)
}
