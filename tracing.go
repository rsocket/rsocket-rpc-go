package rrpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

var (
	errEmptyTracing = errors.New("rrpc.tracing: tracing is empty")
	errBadTracing   = errors.New("rrpc.tracing: invalid tracing format")
)

type Tracer opentracing.Tracer
type SpanContext opentracing.SpanContext

func MarshallTracing(sc SpanContext) (b []byte, err error) {
	bf := &bytes.Buffer{}
	sc.ForeachBaggageItem(func(k, v string) bool {
		log.Printf("put: %s->%s\n", k, v)
		err = binary.Write(bf, binary.BigEndian, uint16(len(k)))
		if err != nil {
			return false
		}
		_, err = bf.WriteString(k)
		if err != nil {
			return false
		}
		err = binary.Write(bf, binary.BigEndian, uint16(len(v)))
		if err != nil {
			return false
		}
		_, err = bf.WriteString(v)
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return
	}
	b = bf.Bytes()
	return
}

func UnmarshallTracing(tracer Tracer, tracing []byte) (SpanContext, error) {
	if len(tracing) < 1 {
		return nil, errEmptyTracing
	}
	scanner := bufio.NewScanner(bytes.NewReader(tracing))
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return
		}
		if len(data) < 2 {
			return
		}
		kvLen := int(binary.BigEndian.Uint16(data))
		if kvLen < 1 {
			err = errBadTracing
			return
		}
		kvSize := kvLen + 2
		if kvSize <= len(data) {
			return kvSize, data[:kvSize], nil
		}
		return
	})
	var pairs []string
	for scanner.Scan() {
		pairs = append(pairs, scanner.Text())
	}
	m := make(map[string]string)
	for i, l := 0, len(pairs); i < l; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(m))
}
