package rrpc_test

import (
	"log"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	. "github.com/rsocket/rsocket-rpc-go"
	"github.com/stretchr/testify/require"
)

func TestMarshallAndUnmarshallTracing(t *testing.T) {
	tracer := mocktracer.New()
	span := tracer.StartSpan("rsocket.request")
	span.SetBaggageItem("foo", "bar")
	span.Finish()

	carrier := opentracing.TextMapCarrier(map[string]string{
		"one":   "1",
		"two":   "2",
		"three": "3",
		"four":  "4",
		"five":  "5",
	})

	err := tracer.Inject(span.Context(), opentracing.TextMap, carrier)
	require.NoError(t, err, "inject failed")

	log.Println("carrier after inject:", carrier)

	raw, err := MarshallTracing(carrier)
	require.NoError(t, err, "marshall failed")

	ctx1 := span.Context()
	ctx2, err := UnmarshallTracing(tracer, raw)
	require.NoError(t, err, "unmarshall failed")
	require.Equal(t, ctx1, ctx2, "bad context")
	log.Printf("context1: %+v\n", ctx1)
	log.Printf("context2: %+v\n", ctx2)
}
