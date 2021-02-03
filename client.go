package rrpc

import (
	"context"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
)

type releasable interface {
	IncRef() int32
	RefCnt() int32
	Release()
}

func pinPayload(input payload.Payload) error {
	if r, ok := input.(releasable); ok {
		r.IncRef()
	}
	return nil
}

// ClientConn struct
type ClientConn struct {
	rSocket       rsocket.RSocket
	meterRegistry MeterRegistry
	tracer        Tracer
}

// InvokeRequestResponse invoke request response
func (p *ClientConn) InvokeRequestResponse(
	ctx context.Context,
	srv string,
	method string,
	data *[]byte,
	opts ...CallOption,
) (<-chan payload.Payload, <-chan error) {
	o := &callOption{}
	for i := range opts {
		opt := opts[i]
		if opt != nil {
			opt(o)
		}
	}

	sent, e := NewRequestPayload(srv, method, *data, o.tracing, o.metadata)
	if e != nil {
		err := make(chan error, 1)
		err <- e
		close(err)
		return nil, err
	}

	return p.rSocket.RequestResponse(sent).DoOnSuccess(pinPayload).ToChan(ctx)
}

// InvokeRequestStream invoke request stream
func (p *ClientConn) InvokeRequestStream(
	ctx context.Context,
	srv string,
	method string,
	data *[]byte,
	opts ...CallOption,
) (<-chan payload.Payload, <-chan error) {
	o := &callOption{}
	for i := range opts {
		opt := opts[i]
		if opt != nil {
			opt(o)
		}
	}

	sent, e := NewRequestPayload(srv, method, *data, o.tracing, o.metadata)
	if e != nil {
		err := make(chan error, 1)
		err <- e
		close(err)
		return nil, err
	}

	return p.rSocket.RequestStream(sent).DoOnNext(pinPayload).ToChan(ctx, 1)
}

// NewClientConn creates new client
func NewClientConn(c rsocket.RSocket, m MeterRegistry, t Tracer) *ClientConn {
	return &ClientConn{
		rSocket:       c,
		meterRegistry: m,
		tracer:        t,
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

func (p *ClientConn) InvokeChannel(
	ctx context.Context,
	srv string,
	method string,
	datachan chan *[]byte,
	err chan error,
	opts ...CallOption) (<-chan payload.Payload, <-chan error) {

	o := &callOption{}
	for i := range opts {
		opt := opts[i]
		if opt != nil {
			opt(o)
		}
	}

	inchan := make(chan payload.Payload)
	inerr := make(chan error)
	scheduler.Parallel().Worker().Do(func() {
		defer close(inchan)
		defer close(inerr)
	loop:
		for {
			select {
			case data, ok := <-datachan:
				if ok {
					sent, e := NewRequestPayload(srv, method, *data, o.tracing, o.metadata)
					if e != nil {
						inerr <- e
					} else {
						inchan <- sent
					}
				} else {
					break loop
				}
			case e := <-err:
				if e != nil {
					inerr <- e
				}
			}
		}
	})

	influx := flux.CreateFromChannel(inchan, inerr)

	return p.rSocket.RequestChannel(influx).DoOnNext(pinPayload).ToChan(ctx, 1)
}
