package ping_pong

import (
	"context"
	"log"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/stretchr/testify/assert"
)

func mockPing() *Ping {
	return &Ping{
		Ball:        "foo",
		AnotherBall: "bar",
	}
}

func TestX(t *testing.T) {
	req := mockPing()
	bs, _ := proto.Marshal(req)
	tgt := newMonoT(mono.Just(payload.New(bs, nil)).Raw(), true)
	res, err := tgt.
		DoOnSuccess(func(it *T) {
			assert.Equal(t, req.Ball, it.Ball)
			assert.Equal(t, req.AnotherBall, it.AnotherBall)
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, req.Ball, res.Ball)
	assert.Equal(t, req.AnotherBall, res.AnotherBall)
}

func Test_implMonoPing_Block(t *testing.T) {
	req := mockPing()
	v, err := NewMonoPing(req).
		DoOnComplete(func() {
			log.Println("complete")
		}).
		DoOnSuccess(func(ping *Ping) {
			log.Println("on next:", ping)
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, req.Ball, v.Ball)
	assert.Equal(t, req.AnotherBall, v.AnotherBall)
}
