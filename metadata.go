package rrpc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/rsocket/rsocket-rpc-go/internal/common"
)

const Version = uint16(1)

type Metadata []byte

func (p Metadata) String() string {
	var tr string
	if b := p.Tracing(); len(b) < 1 {
		tr = "<nil>"
	} else {
		tr = "0x" + hex.EncodeToString(b)
	}

	var m string
	if b := p.Metadata(); len(b) < 1 {
		m = "<nil>"
	} else {
		m = "0x" + hex.EncodeToString(b)
	}
	return fmt.Sprintf(
		"Metadata{version=%d, service=%s, method=%s, tracing=%s, metadata=%s}",
		p.Version(),
		common.Bytes2str(p.Service()),
		common.Bytes2str(p.Method()),
		tr,
		m,
	)
}

func (p Metadata) Version() uint16 {
	raw := p.pp()
	_ = raw[1]
	return binary.BigEndian.Uint16(raw)
}

func (p Metadata) Service() []byte {
	a, b := p.seekNext(2)
	raw := p.pp()
	return raw[a:b]
}

func (p Metadata) Method() []byte {
	raw := p.pp()
	a, b := 0, 2
	for range [2]struct{}{} {
		a, b = p.seekNext(b)
	}
	return raw[a:b]
}

func (p Metadata) Tracing() []byte {
	a, b := 0, 2
	for range [3]struct{}{} {
		a, b = p.seekNext(b)
	}
	raw := p.pp()
	return raw[a:b]
}

func (p Metadata) Metadata() []byte {
	b := 2
	for range [3]struct{}{} {
		_, b = p.seekNext(b)
	}
	raw := p.pp()
	return raw[b:]
}

func (p Metadata) pp() []byte {
	return ([]byte)(p)
}

func (p Metadata) seekNext(offset int) (int, int) {
	raw := p.pp()
	l := binary.BigEndian.Uint16(raw[offset:])
	offset += 2
	return offset, offset + int(l)
}

func encodeMetadata(service, method, tracing, metadata []byte) (m Metadata, err error) {
	w := &bytes.Buffer{}
	// write version
	err = binary.Write(w, binary.BigEndian, Version)
	if err != nil {
		return
	}
	// write service
	err = binary.Write(w, binary.BigEndian, uint16(len(service)))
	if err != nil {
		return
	}
	_, err = w.Write(service)
	if err != nil {
		return
	}
	// write method
	err = binary.Write(w, binary.BigEndian, uint16(len(method)))
	if err != nil {
		return
	}
	_, err = w.Write(method)
	if err != nil {
		return
	}
	// write tracing
	lenTracing := uint16(len(tracing))
	err = binary.Write(w, binary.BigEndian, lenTracing)
	if err != nil {
		return
	}
	if lenTracing > 0 {
		_, err = w.Write(tracing)
		if err != nil {
			return
		}
	}
	// write metadata
	if l := len(metadata); l > 0 {
		_, err = w.Write(metadata)
		if err != nil {
			return
		}
	}
	m = w.Bytes()
	return
}
