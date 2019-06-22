package rrpc

import (
	"context"

	"github.com/rsocket/rsocket-go/rx"
)

type rawFlux interface {
	Raw() rx.Flux
}

type methodHandler func(ctx context.Context, srv interface{}, dec func(interface{}) error, m Metadata) (interface{}, error)
type streamHandler func(ctx context.Context, srv interface{}, dec func(interface{}) error, m Metadata) (interface{}, error)

type ServiceDesc struct {
	Name        string
	HandlerType interface{}
	Methods     []MethodDesc
	Streams     []StreamDesc

	Metadata interface{}
}

type MethodDesc struct {
	Name    string
	Handler methodHandler
}

type StreamDesc struct {
	Name    string
	Handler streamHandler
}
