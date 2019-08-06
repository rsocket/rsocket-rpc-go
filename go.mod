module github.com/rsocket/rsocket-rpc-go

go 1.12

require (
	github.com/golang/protobuf v1.3.2
	github.com/hoisie/mustache v0.0.0-20120318181656-6dfe7cd5e765
	github.com/jjeffcaii/reactor-go v0.0.12
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/rsocket/rsocket-go v0.0.0-00010101000000-000000000000
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.3.0
)

replace github.com/rsocket/rsocket-go => ../rsocket-go
