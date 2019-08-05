package rrpc

import "github.com/rsocket/rsocket-go"

type RrpcRSocket interface {
	rsocket.RSocket
	Name() string
}
