package internal_test

import (
	"crypto/rand"
	"testing"

	. "github.com/rsocket/rsocket-rpc-go/internal"
	"github.com/stretchr/testify/require"
)

func TestEncodeMetadata(t *testing.T) {
	service := "foo"
	method := "bar"
	var md [20]byte
	_, _ = rand.Read(md[:])

	m, err := EncodeMetadata([]byte(service), []byte(method), nil, md[:])
	require.NoError(t, err, "encode failed")

	require.Equal(t, service, string(m.Service()), "bad service")
	require.Equal(t, method, string(m.Method()), "bad method")
	require.Equal(t, "", string(m.Tracing()), "bad tracing")
	require.Equal(t, md[:], m.Metadata(), "bad metadata")
}
