package rrpc

import (
	"context"
	"reflect"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
)

type service struct {
	ss interface{}
	md map[string]*MethodDesc
	sm map[string]*StreamDesc
}

type Server struct {
	mu    sync.Mutex
	m     map[string]*service
	serve bool
}

func (p *Server) Acceptor() rsocket.ServerAcceptor {
	return func(setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) rsocket.RSocket {
		return rsocket.NewAbstractSocket(
			rsocket.RequestResponse(p.handleRequestResponse),
			rsocket.RequestStream(p.handleStream),
		)
	}
}

func (p *Server) handleRequestResponse(msg payload.Payload) rx.Mono {
	return rx.NewMono(func(ctx context.Context, sink rx.MonoProducer) {
		res, err := p.makeResponse(ctx, msg)
		if err != nil {
			sink.Error(err)
			return
		}
		msg, ok := res.(proto.Message)
		if !ok {
			sink.Error(errors.Errorf("rrpc: invalid response type: expect=proto.Message, actual=%v", msg))
			return
		}
		raw, err := proto.Marshal(msg)
		if err != nil {
			sink.Error(err)
			return
		}
		// TODO: fill response metadata.
		pd := payload.New(raw, nil)
		if err := sink.Success(pd); err != nil {
			pd.Release()
		}
	})
}

func (p *Server) handleStream(msg payload.Payload) rx.Flux {
	v, err := p.makeStream(context.Background(), msg)
	// TODO: process error
	if err != nil {
		panic(err)
	}
	return v.(rawFlux).Raw()
}

func (p *Server) makeStream(ctx context.Context, req payload.Payload) (res interface{}, err error) {
	m, ok := req.Metadata()
	if !ok {
		err = errors.New("rrpc: missing metadata in Payload")
		return
	}
	meta := (Metadata)(m)
	ss, ok := p.m[string(meta.Service())]
	if !ok {
		err = errors.Errorf("rrpc: no such service %s", string(meta.Service()))
		return
	}
	md, ok := ss.sm[string(meta.Method())]
	if !ok {
		err = errors.Errorf("rrpc: no such method %s", string(meta.Method()))
		return
	}
	res, err = md.Handler(ctx, ss.ss, p.getUnmarshaller(req.Data()), meta)
	return
}

func (p *Server) makeResponse(ctx context.Context, req payload.Payload) (res interface{}, err error) {
	m, ok := req.Metadata()
	if !ok {
		err = errors.New("rrpc: missing metadata in Payload")
		return
	}
	meta := (Metadata)(m)
	ss, ok := p.m[string(meta.Service())]
	if !ok {
		err = errors.Errorf("rrpc: no such service %s", string(meta.Service()))
		return
	}
	md, ok := ss.md[string(meta.Method())]
	if !ok {
		err = errors.Errorf("rrpc: no such method %s", string(meta.Method()))
		return
	}
	res, err = md.Handler(ctx, ss.ss, p.getUnmarshaller(req.Data()), meta)
	return
}

func (p *Server) RegisterService(sd *ServiceDesc, ss interface{}) {
	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		panic(errors.Errorf("rrpc: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht))
		return
	}
	p.register(sd, ss)
}

func (p *Server) register(sd *ServiceDesc, ss interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.serve {
		panic(errors.Errorf("rrpc: cannot register service %s after serving", sd.Name))
	}
	if _, ok := p.m[sd.Name]; ok {
		panic(errors.Errorf("rrpc: duplicated service %s", sd.Name))
	}
	srv := &service{
		ss: ss,
		md: make(map[string]*MethodDesc),
		sm: make(map[string]*StreamDesc),
	}
	for i := range sd.Methods {
		it := &sd.Methods[i]
		srv.md[it.Name] = it
	}
	for i := range sd.Streams {
		it := &sd.Streams[i]
		srv.sm[it.Name] = it
	}
	p.m[sd.Name] = srv
}

func (p *Server) getUnmarshaller(raw []byte) func(interface{}) error {
	return func(i interface{}) error {
		return proto.Unmarshal(raw, i.(proto.Message))
	}
}

func NewServer() *Server {
	return &Server{
		mu: sync.Mutex{},
		m:  make(map[string]*service),
	}
}
