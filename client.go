package rrpc

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
)

type ClientConn struct {
	c rsocket.Client
}

func (p *ClientConn) Invoke(
	ctx context.Context,
	srv string,
	method string,
	in proto.Message,
	out proto.Message,
) (err error) {
	sent, err := NewRequestPayload(srv, method, in)
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

func NewClientConn(c rsocket.Client) *ClientConn {
	return &ClientConn{c}
}
