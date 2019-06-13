package rrpc

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go"
)

type RPCService interface {
	rsocket.RSocket
	Service() string
}

type serviceRegistry struct {
	m *sync.Map
}

func (p *serviceRegistry) Register(s RPCService) (err error) {
	name := s.Service()
	_, loaded := p.m.LoadOrStore(name, s)
	if loaded {
		err = errors.Errorf("register service failed: duplicated service name %s", name)
	}
	return
}

func (p *serviceRegistry) Load(name string) (s RPCService, ok bool) {
	v, ok := p.m.Load(name)
	if ok {
		s = v.(RPCService)
	}
	return
}

func newRegister() *serviceRegistry {
	return &serviceRegistry{
		m: &sync.Map{},
	}
}
