package rrpc

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeMetadata(t *testing.T) {
	service := "foo"
	method := "bar"
	var md [20]byte
	_, _ = rand.Read(md[:])

	m, err := EncodeMetadata(service, method, nil, md[:])
	require.NoError(t, err, "encode failed")

	require.Equal(t, service, string(m.Service()), "bad service")
	require.Equal(t, method, string(m.Method()), "bad method")
	require.Equal(t, "", string(m.Tracing()), "bad tracing")
	require.Equal(t, md[:], m.Metadata(), "bad metadata")
}

func TestSliceEmptyMetadata(t *testing.T) {
	service := "foo"
	method := "bar"

	m, err := EncodeMetadata(service, method, nil, nil)
	require.NoError(t, err, "encode failed")

	md := m.Metadata()
	require.Equal(t, md, []byte{}, "should be empty byte array")
}

func TestStringWithEmptyMetadata(t *testing.T) {
	service := "foo"
	method := "bar"
	var md [20]byte
	_, _ = rand.Read(md[:])

	m, err := EncodeMetadata(service, method, nil, md[:])
	require.NoError(t, err, "encode failed")

	s := m.String()
	fmt.Println(s)
}

func TestStringWithNilMetadata(t *testing.T) {
	service := "foo"
	method := "bar"
	var md [20]byte
	_, _ = rand.Read(md[:])

	m, err := EncodeMetadata(service, method, nil, nil)
	require.NoError(t, err, "encode failed")

	s := m.String()
	fmt.Println(s)
}

