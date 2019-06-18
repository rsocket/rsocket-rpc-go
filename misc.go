package rrpc

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-rpc-go/internal/common"
	meta "github.com/rsocket/rsocket-rpc-go/internal/metadata"
)

func NewRequestPayload(srv string, method string, msg proto.Message, tracing []byte, metadata []byte) (req payload.Payload, err error) {
	m, err := meta.EncodeMetadata(common.Str2bytes(srv), common.Str2bytes(method), tracing, metadata)
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
