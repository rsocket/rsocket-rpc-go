package rrpc

import (
	"fmt"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
)

type RequestHandler struct {
	rsocket.RSocket
	handlers map[string]RrpcRSocket
}

type RequestHandlingRSocket interface {
	rsocket.RSocket
	Register(rsocket RrpcRSocket) error
}

func (r *RequestHandler) Register(rsocket RrpcRSocket) error {
	if r.handlers[rsocket.Name()] != nil {
		return fmt.Errorf("a service named %s already exists", rsocket.Name())
	}
	r.handlers[rsocket.Name()] = rsocket
	return nil
}

func NewRequestHandler() *RequestHandler {
	return &RequestHandler{
		handlers: make(map[string]RrpcRSocket),
	}
}

// FireAndForget is a single one-way message.
func (r *RequestHandler) FireAndForget(msg payload.Payload) {
	m, ok := msg.Metadata()
	if !ok {
		panic("missing metadata")
	}
	metadata := (Metadata)(m)
	srv := string(metadata.Service())

	rs := r.handlers[srv]
	if rs == nil {
		panic("missing srv")
	}

	rs.FireAndForget(msg)
}

// MetadataPush sends asynchronous Metadata frame.
func (r *RequestHandler) MetadataPush(msg payload.Payload) {
	m, ok := msg.Metadata()
	if !ok {
		panic("missing metadata")
	}
	metadata := (Metadata)(m)
	srv := string(metadata.Service())

	rs := r.handlers[srv]
	if rs == nil {
		panic("missing srv")
	}

	rs.MetadataPush(msg)
}

// RequestResponse request single response.
func (r *RequestHandler) RequestResponse(msg payload.Payload) mono.Mono {
	m, ok := msg.Metadata()
	if !ok {
		panic("missing metadata")
	}
	metadata := (Metadata)(m)
	srv := string(metadata.Service())

	rs := r.handlers[srv]
	if rs == nil {
		panic("missing srv")
	}

	return rs.RequestResponse(msg)
}

// RequestStream request a completable stream.
func (r *RequestHandler) RequestStream(msg payload.Payload) flux.Flux {
	m, ok := msg.Metadata()
	if !ok {
		panic("missing metadata")
	}
	metadata := (Metadata)(m)
	srv := string(metadata.Service())

	rs := r.handlers[srv]
	if rs == nil {
		panic("missing srv")
	}

	return rs.RequestStream(msg)
}

func (r *RequestHandler) RequestChannel(msgs rx.Publisher) flux.Flux {
	clone := flux.Clone(msgs)
	return clone.SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {
		msg, ok := s.Value()
		if !ok {
			panic("missing value")
		}

		m, ok := msg.Metadata()
		if !ok {
			panic("missing metadata")
		}

		metadata := (Metadata)(m)
		srv := string(metadata.Service())

		rs := r.handlers[srv]
		if rs == nil {
			panic("missing srv")
		}

		return rs.RequestChannel(f)
	})
}
