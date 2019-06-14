package rrpc

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-rpc-go/internal/common"
	"github.com/rsocket/rsocket-rpc-go/internal/metadata"
)

func NewRequestPayload(srv string, method string, msg proto.Message) (req payload.Payload, err error) {
	m, err := metadata.EncodeMetadata(common.Str2bytes(srv), common.Str2bytes(method), nil, nil)
	if err != nil {
		err = errors.Wrap(err, "rrpc: encode request metadata failed")
		return
	}
	d, err := proto.Marshal(msg)
	if err != nil {
		err = errors.Wrap(err, "rrpc: encode protobuf message failed")
		return
	}
	req = payload.New(d, m)
	return
}
