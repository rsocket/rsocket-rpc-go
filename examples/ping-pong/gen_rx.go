package ping_pong

import (
	"context"

	"github.com/golang/protobuf/proto"
	xmono "github.com/jjeffcaii/reactor-go/mono"
	"github.com/rsocket/rsocket-go/payload"
)

type T = Ping

type MonoT interface {
	DoOnSuccess(func(*T)) MonoT
	DoOnError(func(error)) MonoT
	DoOnComplete(func()) MonoT
	Block(ctx context.Context) (*T, error)
}

type innerMonoT struct {
	origin xmono.Mono
}

func (p innerMonoT) DoOnComplete(callback func()) MonoT {
	return newMonoT(p.origin.DoOnComplete(callback), false)
}

func (p innerMonoT) DoOnError(callback func(error)) MonoT {
	return newMonoT(p.origin.DoOnError(callback), false)
}

func (p innerMonoT) DoOnSuccess(callback func(*T)) MonoT {
	return newMonoT(p.origin.DoOnNext(func(input interface{}) {
		callback(input.(*T))
	}), false)
}

func (p innerMonoT) Block(ctx context.Context) (*T, error) {
	v, err := p.origin.Block(ctx)
	if err != nil {
		return nil, err
	}
	return v.(*T), nil
}

func newMonoT(input xmono.Mono, autoMap bool) MonoT {
	if autoMap {
		input = input.Map(func(v interface{}) interface{} {
			ret := new(T)
			pa := v.(payload.Payload)
			if err := proto.Unmarshal(pa.Data(), ret); err != nil {
				panic(err)
			}
			return ret
		})
	}
	return innerMonoT{origin: input}
}

func NewMonoT(input *T) MonoT {
	return newMonoT(xmono.Create(func(ctx context.Context, s xmono.Sink) {
		bs, err := proto.Marshal(input)
		if err != nil {
			s.Error(err)
		} else {
			s.Success(payload.New(bs, nil))
		}
	}), true)
}
