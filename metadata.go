package rrpc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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
		p.Service(),
		p.Method(),
		tr,
		m,
	)
}

func (p Metadata) Version() uint16 {
	return binary.BigEndian.Uint16(p.raw())
}

func (p Metadata) Service() string {
	offset := 2
	raw := p.raw()

	serviceLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2

	return string(raw[offset : offset+serviceLen])
}

func (p Metadata) Method() string {
	offset := 2
	raw := p.raw()

	serviceLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2 + serviceLen

	methodLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2

	return string(raw[offset : offset+methodLen])
}

func (p Metadata) Tracing() []byte {
	offset := 2
	raw := p.raw()

	serviceLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2 + serviceLen

	methodLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2 + methodLen

	tracingLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2

	if tracingLen > 0 {
		return raw[offset : offset+tracingLen]
	} else {
		return nil
	}
}

func (p Metadata) Metadata() []byte {
	offset := 2
	raw := p.raw()

	serviceLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2 + serviceLen

	methodLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2 + methodLen

	tracingLen := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
	offset += 2 + tracingLen

	return raw[offset:]

}

func (p Metadata) raw() []byte {
	return ([]byte)(p)
}

func EncodeMetadata(service string, method string, tracing []byte, metadata []byte) (m Metadata, err error) {
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
	_, err = w.WriteString(service)
	if err != nil {
		return
	}
	// write method
	err = binary.Write(w, binary.BigEndian, uint16(len(method)))
	if err != nil {
		return
	}
	_, err = w.WriteString(method)
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
