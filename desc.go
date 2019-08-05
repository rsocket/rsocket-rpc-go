package rrpc

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/rsocket/rsocket-go/payload"
)

type methodHandler func(ctx context.Context,
	handler interface{},
	deserialize func(payload payload.Payload) (proto.Message, error),
	metadata Metadata) (<-chan *proto.Message, <-chan error)

type ServiceDesc struct {
	Name        string
	HandlerType interface{}
	Methods     []MethodDesc

	Metadata interface{}
}

type MethodDesc struct {
	Name    string
	Handler methodHandler
}
