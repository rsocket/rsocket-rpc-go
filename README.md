# rsocket-rpc-go
RSocket RPC Golang

## NOTICE
It is still under development. **DO NOT USE IT IN A PRODUCTION ENVIRONMENT!!!**

## Install
1. Install Protocol Buffers v3. Please see [https://github.com/protocolbuffers/protobuf](https://github.com/protocolbuffers/protobuf).
2. We offer a `protoc-gen-go` with RSocket RPC support and it's 100% compatible with [official tools](https://github.com/golang/protobuf).
Please install by command below:

```bash
$ go get -u github.com/rsocket/rsocket-rpc-go/protoc-gen-go
```

## Generate Codes
Just use command below:
```bash
$ protoc --go_out=plugins=rrpc:./ping-pong ./ping-pong.proto
```

> NOTICE: you can find some sample codes in [examples](./examples)
