package rrpc

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
)

type ClientConn struct {
	c rsocket.RSocket
	m MeterRegistry
	t Tracer
}

func (p *ClientConn) Invoke(
	ctx context.Context,
	srv string,
	method string,
	in proto.Message,
	out proto.Message,
	opts ...CallOption,
) (err error) {
	o := &callOption{}
	for i := range opts {
		opts[i](o)
	}
	sent, err := NewRequestPayload(srv, method, in, o.tracing, o.metadata)
	if err != nil {
		return
	}
	p.c.RequestResponse(sent).
		DoOnSuccess(func(ctx context.Context, s rx.Subscription, elem payload.Payload) {
			err = proto.Unmarshal(elem.Data(), out.(proto.Message))
		}).
		DoOnError(func(ctx context.Context, e error) {
			err = e
		}).
		Subscribe(ctx)
	return
}

func NewClientConn(c rsocket.RSocket, m MeterRegistry, t Tracer) *ClientConn {
	return &ClientConn{
		c: c,
		m: m,
		t: t,
	}
}

type callOption struct {
	tracing  []byte
	metadata []byte
}

type CallOption func(*callOption)

func WithTracing(tracing []byte) CallOption {
	return func(o *callOption) {
		o.tracing = tracing
	}
}

func WithMetadata(metadata []byte) CallOption {
	return func(o *callOption) {
		o.metadata = metadata
	}
}
