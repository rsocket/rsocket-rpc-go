package rrpc

import (
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/payload"
)

func NewRequestPayload(srv string, method string, d []byte, tracing []byte, metadata []byte) (req payload.Payload, err error) {

	m, err := EncodeMetadata(srv, method, tracing, metadata)
	if err != nil {
		err = errors.Wrap(err, "rrpc: encode request metadata failed")
		return
	}
	req = payload.New(d, m)
	return
}

