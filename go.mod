module github.com/rsocket/rsocket-rpc-go

go 1.12

require (
	github.com/golang/protobuf v1.3.2
	github.com/jjeffcaii/reactor-go v0.0.16
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/rsocket/rsocket-go v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.3.0
)

replace github.com/rsocket/rsocket-go => ../rsocket-go
