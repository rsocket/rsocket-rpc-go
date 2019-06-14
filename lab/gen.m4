divert(-1)

define(DEF_PKG,
`package $1')

# Ping(*Ping) *Pong
define(DEF_METHOD_IFACE,`$1(*$2) *$3')

define(DEF_METHOD_DESC,
`{
			Name: "$2",
			Handler: _$1_$2_handler,
		},')

define(DEF_SRV_DESC,
`var _$1_serviceDesc = rrpc.ServiceDesc{
	Name:		"$1",
	HandlerType:	(*$1Server)(nil),
	Methods:	[]rrpc.MethodDesc{
		DEF_METHOD_DESC($1,Ping)
	},
}')

define(DEF_SRV_METHOD,
`func _$1_$2_handler(srv interface{}, dec func(interface{}) error) (interface{}, error) {
	req := new($3)
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.($1Server).$2(req), nil
}')

define(DEF_SERVER,
`type $1Server interface {
}')

define(DEF_SERVER_REGISTER,
`func Register$1Server(s *rrpc.Server, srv $1Server) {
	s.RegisterService(&_$1_serverDesc, srv)
}')

divert(0)dnl

DEF_PKG(lab)

import (
	rrpc "github.com/rsocket/rsocket-rpc-go"
)

DEF_SRV_DESC(PingPong)

DEF_SRV_METHOD(PingPong,Ping,Ping)

DEF_SERVER(PingPong)

DEF_SERVER_REGISTER(PingPong)

DEF_METHOD_IFACE(Ping,Ping,Pong)


