package rrpc

import (
	"log"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/require"
)

func TestUnmarshallTracing(t *testing.T) {
	//s := "0004666f757200013400036f6e65000131000374776f00013200057468726565000133000466697665000135"
	//raw, _ := hex.DecodeString(s)
	tracer := mocktracer.New()
	span := tracer.StartSpan("a", opentracing.Tags(map[string]interface{}{
		"one":   "1",
		"two":   "2",
		"three": "3",
		"four":  "4",
		"five":  "5",
	}))
	span.SetBaggageItem("foo", "bar")
	span.Finish()
	tracer.FinishedSpans()

	raw, err := MarshallTracing(span.Context())
	require.NoError(t, err, "marshall tracing failed")
	sc, err := UnmarshallTracing(tracer, raw)
	require.NoError(t, err, "unmarshall tracing failed")
	log.Println(sc)
}
