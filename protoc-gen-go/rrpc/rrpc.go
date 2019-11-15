package rrpc

import (
	"log"
	"strconv"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

const (
	contextPkgPath = "context"
	rrpcPkgPath    = "github.com/rsocket/rsocket-rpc-go"
	rsocketPkgPath = "github.com/rsocket/rsocket-go"
	rawMonoPkgPath = "github.com/jjeffcaii/reactor-go/mono"
	rawFluxPkgPath = "github.com/jjeffcaii/reactor-go/flux"
	schPkgPath     = "github.com/jjeffcaii/reactor-go/scheduler"
	rxPkgPath      = "github.com/rsocket/rsocket-go/rx"
	fluxPkgPath    = "github.com/rsocket/rsocket-go/rx/flux"
	monoPkgPath    = "github.com/rsocket/rsocket-go/rx/mono"
	payloadPkgPath = "github.com/rsocket/rsocket-go/payload"
	errorsPkgPath  = "errors"
)

var (
	contextPkg string
	rrpcPkg    string
	rsocketPkg string
	rxPkg      string
	fluxPkg    string
	monoPkg    string
	payloadPkg string
	errorsPkg  string
	schPkg     string
	rawMonoPkg string
	rawFluxPkg string
)

func init() {
	generator.RegisterPlugin(new(rrpc))
}

type rrpc struct {
	gen *generator.Generator
}

func (g *rrpc) Name() string {
	return "rrpc"
}

func (g *rrpc) Init(gen *generator.Generator) {
	g.gen = gen
}

func (g *rrpc) Generate(file *generator.FileDescriptor) {
	defer func() {
		i := recover()
		if i != nil {
			log.Println(i)
		}
	}()
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	contextPkg = string(g.gen.AddImport(contextPkgPath))
	rrpcPkg = string(g.gen.AddImport(rrpcPkgPath))
	rsocketPkg = string(g.gen.AddImport(rsocketPkgPath))
	rxPkg = string(g.gen.AddImport(rxPkgPath))
	fluxPkg = string(g.gen.AddImport(fluxPkgPath))
	monoPkg = string(g.gen.AddImport(monoPkgPath))
	payloadPkg = string(g.gen.AddImport(payloadPkgPath))
	errorsPkg = string(g.gen.AddImport(errorsPkgPath))
	schPkg = string(g.gen.AddImport(schPkgPath))
	rawMonoPkg = string(g.gen.AddImport(rawMonoPkgPath))
	rawFluxPkg = string(g.gen.AddImport(rawFluxPkgPath))

	g.P("var _ ", contextPkg, ".Context")
	g.P("var _ ", rrpcPkg, ".ClientConn")
	g.P("var _ ", rsocketPkg, ".RSocket")
	g.P("var _ ", rxPkg, ".Subscription")
	g.P("var _ ", fluxPkg, ".Flux")
	g.P("var _ ", monoPkg, ".Mono")
	g.P("var _ ", payloadPkg, ".Payload")
	g.P("var _ ", schPkg, ".Scheduler")
	g.P("var _ ", rawMonoPkg, ".Mono")
	g.P("var _ ", rawFluxPkg, ".Flux")

	for _, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service)
	}
}

func (g *rrpc) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	// Constants
	g.P("// -- Constants")
	monos := make(map[string]struct{}, 0)
	fluxs := make(map[string]struct{}, 0)
	g.P("const ", service.GetName(), `ServiceName = "`, file.GetPackage(), ".", service.GetName(), `"`)
	for _, method := range service.GetMethod() {
		g.P("const ", method.GetName(), `FunctionName = "`, method.GetName(), `"`)
		if method.GetClientStreaming() {
			fluxs[method.GetInputType()] = struct{}{}
		} else {
			monos[method.GetInputType()] = struct{}{}
		}
		if method.GetServerStreaming() {
			fluxs[method.GetOutputType()] = struct{}{}
		} else {
			monos[method.GetOutputType()] = struct{}{}
		}
	}
	g.P()

	g.P("// -- Mono start")
	for it := range monos {
		g.generateMono(it)
	}

	g.P()
	g.P("// -- client start")
	g.generateClientInterface(service)
	g.P()
	g.P("type ", service.GetName(), "ClientStruct struct {")
	g.P(service.GetName(), "Client")

	g.P("client ", rrpcPkg, ".ClientConn")
	g.P("}")
	g.P()
	g.generateClientFunctions(service)
	g.P()
	g.generateClientConstruct(service)
	g.P("// -- client end")
	g.P()
	g.P("// -- server start")
	g.generateServerInterface(service)
	g.P()
	g.generateServerRequestResponse(service)
	g.P()
	g.generateServerRequestStream(service)
	g.P()
	g.generateServerRequestChannel(service)
	//g.P()
	//g.generateServerFireAndForget(service)
	g.P()
	g.generateServerConstructor(service)
}

func cleanupType(t string) string {
	index := strings.LastIndex(t, ".")
	return t[index+1:]
}

func (g *rrpc) generateMono(inputType string) {
	typo := cleanupType(inputType)
	iface := "Mono" + typo
	g.P("type ", iface, " interface {")
	g.P("DoOnSuccess(func(*", typo, ")) ", iface)
	g.P("DoOnError(func(error)) ", iface)
	g.P("DoOnComplete(func()) ", iface)
	g.P("Block(", contextPkg, ".Context) (*", typo, ",error)")
	g.P("}")
	g.P()

	imp := "_impl" + iface
	impNew := "new" + iface

	g.P("type ", imp, " struct {")
	g.P("origin ", rawMonoPkg, ".Mono")
	g.P("}")
	g.P()
	g.P("func (p ", imp, ") DoOnComplete(callback func()) ", iface, " {")
	g.P("return ", impNew, "(p.origin.DoOnComplete(callback), false)")
	g.P("}")
	g.P()
	g.P("func (p ", imp, ") DoOnError(callback func(error)) ", iface, " {")
	g.P("return ", impNew, "(p.origin.DoOnError(callback), false)")
	g.P("}")
	g.P()
	g.P("func (p ", imp, ") DoOnSuccess(callback func(*", typo, ")) ", iface, " {")
	g.P("return ", impNew, "(p.origin.DoOnNext(func(input interface{}) {")
	g.P("callback(input.(*", typo, "))")
	g.P("}), false)")
	g.P("}")
	g.P()

	g.P("func (p ", imp, ") Block(ctx ", contextPkg, ".Context) (*", typo, ", error) {")
	g.P("v, err := p.origin.Block(ctx)")
	g.P("if err != nil {")
	g.P("return nil, err")
	g.P("}")
	g.P("return v.(*", typo, "), nil")
	g.P("}")
	g.P()
	g.P("func ", impNew, "(input ", rawMonoPkg, ".Mono, autoMap bool) ", iface, " {")
	g.P("if autoMap {")
	g.P("input = input.Map(func(v interface{}) interface{} {")
	g.P("ret := new(", typo, ")")
	g.P("pa := v.(", payloadPkg, ".Payload)")
	g.P("if err := proto.Unmarshal(pa.Data(), ret); err != nil {")
	g.P("panic(err)")
	g.P("}")
	g.P("return ret")
	g.P("})")
	g.P("}")
	g.P("return ", imp, "{origin: input}")
	g.P("}")
	g.P()
	g.P("func New", iface, "(input *", typo, ") ", iface, "{")
	g.P("return ", impNew, "(", rawMonoPkg, ".Create(func(ctx ", contextPkg, ".Context, s ", rawMonoPkg, ".Sink) {")
	g.P("bs, err := proto.Marshal(input)")
	g.P("if err != nil {")
	g.P("s.Error(err)")
	g.P("} else {")
	g.P("s.Success(", payloadPkg, ".New(bs, nil))")
	g.P("}")
	g.P("}), true)")
	g.P("}")
	g.P()
}

func (g *rrpc) generateClientInterface(service *descriptor.ServiceDescriptorProto) {
	g.P("type ", service.GetName(), "Client interface {")
	for _, method := range service.GetMethod() {
		g.P()
		var method = method
		if !method.GetClientStreaming() {
			g.P(strings.Title(method.GetName()), "(ctx context.Context, in *", cleanupType(method.GetInputType()), ", opts ...", rrpcPkg, ".CallOption) (<-chan *", cleanupType(cleanupType(method.GetOutputType())), ", <-chan error)")
		} else {
			g.P(strings.Title(method.GetName()), "(ctx context.Context, in chan *", cleanupType(method.GetInputType()), ", err chan error, opts ...", rrpcPkg, ".CallOption) (<-chan *", cleanupType(cleanupType(method.GetOutputType())), ", <-chan error)")
		}
	}
	g.P("}")
	g.P()
}

func (g *rrpc) generateClientFunctions(service *descriptor.ServiceDescriptorProto) {
	for _, method := range service.GetMethod() {
		var method = method
		if method.GetClientStreaming() {
			g.generateClientRequestChannelFunction(service, method)
		} else if method.GetServerStreaming() {
			g.generateClientRequestStreamFunction(service, method)
		} else {
			g.generateClientRequestReplyFunction(service, method)
		}
	}
}

func (g *rrpc) generateClientRequestReplyFunction(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	g.P("func (c *", service.GetName(), "ClientStruct) ", strings.Title(method.GetName()), "(ctx context.Context, in *", cleanupType(method.GetInputType()), ", opts ...", rrpcPkg, ".CallOption) (<-chan *", cleanupType(method.GetOutputType()), ", <-chan error) {")
	g.P("response := make(chan *", cleanupType(method.GetOutputType()), ", 1)")
	g.P("err := make(chan error, 1)")
	g.P("defer func() {")
	g.P("close(response)")
	g.P("close(err)")
	g.P("}()")
	g.P("d, e := proto.Marshal(in)")
	g.P("if e != nil {")
	g.P("return nil, err")
	g.P("}")
	g.P()
	g.P("payloads, errors := c.client.InvokeRequestResponse(ctx, ", service.GetName(), "ServiceName, ", method.GetName(), "FunctionName, d, opts...)")
	g.P("loop:")
	g.P("for {")
	g.P("select {")
	g.P("case p, ok := <-payloads:")
	g.P("if ok {")
	g.P("i := payload.Payload(p)")
	g.P("data := i.Data()")
	g.P("res := &", cleanupType(method.GetOutputType()), "{}")
	g.P("e := proto.Unmarshal(data, res)")
	g.P("if e != nil {")
	g.P("err <- e")
	g.P("break loop")
	g.P("} else {")
	g.P("response <- res")
	g.P("}")
	g.P("} else {")
	g.P("break loop")
	g.P("}")
	g.P("case e := <-errors:")
	g.P("if err != nil {")
	g.P("err <- e")
	g.P("break loop")
	g.P("}")
	g.P("}")
	g.P("}")
	g.P()
	g.P("return response, err")
	g.P("}")
	g.P()
}

func (g *rrpc) generateClientRequestStreamFunction(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	g.P("func (c *", service.GetName(), "ClientStruct) ", strings.Title(method.GetName()), "(ctx context.Context, in *", cleanupType(method.GetInputType()), ", opts ...", rrpcPkg, ".CallOption) (<-chan *", cleanupType(method.GetOutputType()), ", <-chan error) {")
	g.P("err := make(chan error)")
	g.P("d, e := proto.Marshal(in)")
	g.P("if e != nil {")
	g.P("close(err)")
	g.P("return nil, err")
	g.P("}")
	g.P("payloads, errors := c.client.InvokeRequestStream(ctx, ", service.GetName(), "ServiceName, ", method.GetName(), "FunctionName, d, opts...)")
	g.P("response := make(chan *", cleanupType(method.GetOutputType()), ", len(payloads))")
	g.P("scheduler.Elastic().Worker().Do(func() {")
	g.P("defer func() {")
	g.P("close(response)")
	g.P("close(err)")
	g.P("}()")
	g.P("loop:")
	g.P("for {")
	g.P("select {")
	g.P("case p, ok := <-payloads:")
	g.P("if ok {")
	g.P("i := payload.Payload(p)")
	g.P("data := i.Data()")
	g.P("res := &", cleanupType(method.GetOutputType()), "{}")
	g.P("e := proto.Unmarshal(data, res)")
	g.P("if e != nil {")
	g.P("err <- e")
	g.P("break loop")
	g.P("} else {")
	g.P("response <- res")
	g.P("}")
	g.P("} else {")
	g.P("break loop")
	g.P("}")
	g.P("case e := <-errors:")
	g.P("err <- e")
	g.P("break loop")
	g.P("}")
	g.P("}")
	g.P("")
	g.P("})")
	g.P("return response, err")
	g.P("}")
	g.P()
}

func (g *rrpc) generateClientRequestChannelFunction(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	g.P("func (c *", service.GetName(), "ClientStruct) ", strings.Title(method.GetName()), "(ctx context.Context, in chan *", cleanupType(method.GetInputType()), ", err chan error", ", opts ...", rrpcPkg, ".CallOption) (<-chan *", cleanupType(method.GetOutputType()), ", <-chan error) {")
	g.P("bytesin := make(chan []byte)")
	g.P("errin := make(chan error)")
	g.P("scheduler.Elastic().Worker().Do(func() {")
	g.P("defer close(bytesin)")
	g.P("defer close(errin)")
	g.P("loop:")
	g.P("for {")
	g.P("select {")
	g.P("case p, o := <-in:")
	g.P("if o {")
	g.P("d, e := proto.Marshal(p)")
	g.P("if e != nil {")
	g.P("errin <- e")
	g.P("break loop")
	g.P("} else {")
	g.P("bytesin <- d")
	g.P("}")
	g.P("} else {")
	g.P("break loop")
	g.P("}")
	g.P("case e := <-err:")
	g.P("if e != nil {")
	g.P("errin <- e")
	g.P("}")
	g.P("}")
	g.P("}")
	g.P("})")
	g.P("payloads, chanerrors := c.client.InvokeChannel(ctx, ", service.GetName(), "ServiceName, ", method.GetName(), "FunctionName, bytesin, errin, opts...)")
	g.P("payloadsout := make(chan *", cleanupType(method.GetOutputType()), ", len(payloads))")
	g.P("errout := make(chan error)")
	g.P("scheduler.Elastic().Worker().Do(func() {")
	g.P("defer func() {")
	g.P("close(payloadsout)")
	g.P("close(errout)")
	g.P("}()")
	g.P("loop:")
	g.P("for {")
	g.P("select {")
	g.P("case p, ok := <-payloads:")
	g.P("if ok {")
	g.P("i := payload.Payload(p)")
	g.P("data := i.Data()")
	g.P("res := &", cleanupType(method.GetOutputType()), "{}")
	g.P("e := proto.Unmarshal(data, res)")
	g.P("if e != nil {")
	g.P("err <- e")
	g.P("break loop")
	g.P("} else {")
	g.P("payloadsout <- res")
	g.P("}")
	g.P("} else {")
	g.P("break loop")
	g.P("}")
	g.P("case e := <-chanerrors:")
	g.P("if err != nil {")
	g.P("err <- e")
	g.P("break loop")
	g.P("}")
	g.P("}")
	g.P("}")
	g.P("})")
	g.P("return payloadsout, errout")
	g.P("}")
	g.P()
}

func (g *rrpc) generateClientFireAndForgetFunction(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	// Todo - implement fire and forget
}

func (g *rrpc) generateClientConstruct(service *descriptor.ServiceDescriptorProto) {
	g.P("func New", service.GetName(), "Client(s rsocket_go.RSocket, m ", rrpcPkg, ".MeterRegistry, t ", rrpcPkg, ".Tracer) ", service.GetName(), "Client {")
	g.P("cc := *", rrpcPkg, ".NewClientConn(s, m, t)")
	g.P("return &", service.GetName(), "ClientStruct{client: cc}")
	g.P("}")
	g.P()
}

func (g *rrpc) generateServerInterface(service *descriptor.ServiceDescriptorProto) {
	g.P("type ", service.GetName(), " interface {")
	for _, method := range service.GetMethod() {
		var method = method
		if method.GetClientStreaming() {
			g.P(strings.Title(method.GetName()), "(context.Context, chan *", cleanupType(method.GetInputType()), ", chan error, []byte) (<-chan *", cleanupType(method.GetOutputType()), ", <-chan error)")
		} else {
			g.P(strings.Title(method.GetName()), "(context.Context, *", cleanupType(method.GetInputType()), ", []byte) (<-chan *", cleanupType(method.GetOutputType()), ", <-chan error)")
		}
	}
	g.P("}")
	g.P()
	g.P("type ", service.GetName(), "Server struct {")
	g.P("pp ", service.GetName())
	g.P(rrpcPkg, ".RrpcRSocket")
	g.P("}")
	g.P()
	g.P("func (p *", service.GetName(), "Server) Name() string {")
	g.P("return ", service.GetName(), "ServiceName")
	g.P("}")
}

func (g *rrpc) generateServerRequestResponse(service *descriptor.ServiceDescriptorProto) {
	g.P("func (p *", service.GetName(), "Server) RequestResponse(msg payload.Payload) mono.Mono {")

	var found = false
	for _, method := range service.GetMethod() {
		if method.GetClientStreaming() || method.GetServerStreaming() {
			continue
		} else {
			found = true
			break
		}
	}

	if !found {
		g.P("panic(\"request response not implemented\")")
		g.P("}")
		return
	}

	g.P("return mono.Create(func(ctx context.Context, sink mono.Sink) {")
	g.P("d := msg.Data()")
	g.P("m, ok := msg.Metadata()")
	g.P("if !ok {")
	g.P("sink.Error(errors.New(\"RSocket rpc: missing metadata in Payload for ", service.GetName(), " service\"))")
	g.P("return")
	g.P("}")
	g.P("metadata := (", rrpcPkg, ".Metadata)(m)")
	g.P("method := metadata.Method()")
	g.P("ud := metadata.Metadata()")
	g.P("switch method {")
	for i, method := range service.GetMethod() {
		if method.GetClientStreaming() || method.GetServerStreaming() {
			continue
		}

		var method = method
		var in = "_in" + strconv.Itoa(i)
		var out = "_out" + strconv.Itoa(i)
		var loop = "_loop" + strconv.Itoa(i)

		g.P("case ", method.GetName(), "FunctionName:")
		g.P(in, " := &", cleanupType(method.GetInputType()), "{}")
		g.P("e := proto.Unmarshal(d, ", in, ")")
		g.P("if e != nil {")
		g.P("sink.Error(e)")
		g.P("return")
		g.P("}")
		g.P("defer func() {")
		g.P("if err := recover(); err != nil {")
		g.P("sink.Error(fmt.Errorf(\"Error calling %s function: %s\", ", method.GetName(), "FunctionName, err))")
		g.P("}")
		g.P("}()")
		g.P(out, ", err := p.pp.", strings.Title(method.GetName()), "(ctx, ", in, ", ud)")
		g.P(loop, ":")
		g.P("for {")
		g.P("select {")
		g.P("case <-ctx.Done():")
		g.P("case r, ok := <-", out, ":")
		g.P("if ok {")
		g.P("bytes, e := proto.Marshal(r)")
		g.P("if e != nil {")
		g.P("sink.Error(e)")
		g.P("} else {")
		g.P("sink.Success(payload.New(bytes, nil))")
		g.P("break ", loop)
		g.P("}")
		g.P("} else {")
		g.P("break ", loop)
		g.P("}")
		g.P("case e := <-err:")
		g.P("if e != nil {")
		g.P("sink.Error(e)")
		g.P("}")
		g.P("}")
		g.P("}")
	}
	g.P("}")
	g.P("})")
	g.P("}")
}

func (g *rrpc) generateServerRequestStream(service *descriptor.ServiceDescriptorProto) {
	g.P("func (p *", service.GetName(), "Server) RequestStream(msg payload.Payload) flux.Flux {")

	var found = false
	for _, method := range service.GetMethod() {
		if method.GetClientStreaming() || !method.GetServerStreaming() {
			continue
		} else {
			found = true
			break
		}
	}

	if !found {
		g.P("panic(\"request stream not implemented\")")
		g.P("}")
		return
	}

	g.P("d := msg.Data()")
	g.P("m, ok := msg.Metadata()")
	g.P("if !ok {")
	g.P("return flux.Error(errors.New(\"RSocket rpc: missing metadata in Payload for ", service.GetName(), " service\"))")
	g.P("}")
	g.P()
	g.P("metadata := (", rrpcPkg, ".Metadata)(m)")
	g.P("method := metadata.Method()")
	g.P()
	g.P("ud := metadata.Metadata()")
	g.P("switch method {")
	for i, method := range service.GetMethod() {
		if method.GetClientStreaming() || !method.GetServerStreaming() {
			continue
		}
		var method = method
		var in = "_in" + strconv.Itoa(i)
		var out = "_out" + strconv.Itoa(i)

		g.P("case ", method.GetName(), "FunctionName:")
		g.P(in, " := &", cleanupType(method.GetInputType()), "{}")
		g.P("e := proto.Unmarshal(d, ", in, ")")
		g.P("if e != nil {")
		g.P("return flux.Error(e)")
		g.P("}")
		g.P()
		g.P("ctx := context.Background()")
		g.P(out, ", errors := p.pp.", strings.Title(method.GetName()), "(ctx, ", in, ", ud)")
		g.P("payloads := make(chan payload.Payload)")
		g.P("chanerrors := make(chan error)")
		g.P("scheduler.Elastic().Worker().Do(func() {")
		g.P("defer func() {")
		g.P("close(payloads)")
		g.P("close(chanerrors)")
		g.P("}()")
		g.P("loop:")
		g.P("for {")
		g.P("select {")
		g.P("case <-ctx.Done():")
		g.P("case r, ok := <-", out, ":")
		g.P("if ok {")
		g.P("bytes, e := proto.Marshal(r)")
		g.P("p := payload.New(bytes, nil)")
		g.P("payloads <- p")
		g.P("if e != nil {")
		g.P("chanerrors <- e")
		g.P("break loop")
		g.P("}")
		g.P("} else {")
		g.P("break loop")
		g.P("}")
		g.P("case e := <-errors:")
		g.P("chanerrors <- e")
		g.P("break loop")
		g.P("}")
		g.P("}")
		g.P("})")
		g.P("return flux.CreateFromChannel(payloads, chanerrors)")
	}
	g.P("default:")
	g.P("return flux.Error(fmt.Errorf(\"unknown method %s\", method))")
	g.P("}")
	g.P("}")
}

func (g *rrpc) generateServerRequestChannel(service *descriptor.ServiceDescriptorProto) {
	g.P("func (p *", service.GetName(), "Server) RequestChannel(msgs rx.Publisher) flux.Flux {")

	var found = false
	for _, method := range service.GetMethod() {
		if !method.GetClientStreaming() {
			continue
		} else {
			found = true
			break
		}
	}

	if !found {
		g.P("panic(\"request channel not implemented\")")
		g.P("}")
		return
	}

	g.P("return flux.Clone(msgs).SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {")
	g.P("msg, ok := s.Value()")
	g.P("if !ok {")
	g.P("return flux.Error(errors.New(\"RSocket rpc: missing payload to switch request on\"))")
	g.P("}")
	g.P("d := msg.Data()")
	g.P("m, ok := msg.Metadata()")
	g.P("if !ok {")
	g.P("return flux.Error(errors.New(\"RSocket rpc: missing metadata in Payload for PingPong service\"))")
	g.P("}")
	g.P("metadata := (", rrpcPkg, ".Metadata)(m)")
	g.P("method := metadata.Method()")
	g.P("ud := metadata.Metadata()")
	g.P("switch method {")
	for i, method := range service.GetMethod() {
		if !method.GetClientStreaming() {
			continue
		}

		var method = method
		var in = "_in" + strconv.Itoa(i)

		g.P("case ", method.GetName(), "FunctionName:")
		g.P(in, " := &", cleanupType(method.GetInputType()), "{}")
		g.P("e := proto.Unmarshal(d, ", in, ")")
		g.P("if e != nil {")
		g.P("return flux.Error(e)")
		g.P("}")
		g.P()
		g.P("ctx := context.Background()")
		g.P("inchan := make(chan *", cleanupType(method.GetInputType()), ")")
		g.P("inerr := make(chan error)")
		g.P("var sub rx.Subscription")
		g.P("f.DoOnSubscribe(func(s rx.Subscription) {")
		g.P("sub = s")
		g.P("}).SubscribeOn(scheduler.Elastic()).")
		g.P("DoOnNext(func(input payload.Payload) {")
		g.P("_in5 := &Ping{}")
		g.P("e := proto.Unmarshal(d, _in5)")
		g.P("if e != nil {")
		g.P("inerr <- e")
		g.P("if sub != nil {")
		g.P("sub.Cancel()")
		g.P("}")
		g.P("} else {")
		g.P("inchan <- _in5")
		g.P("}")
		g.P("}).")
		g.P("DoOnError(func(e error) {")
		g.P("inerr <- e")
		g.P("}).")
		g.P("DoFinally(func(s rx.SignalType) {")
		g.P("close(inchan)")
		g.P("close(inerr)")
		g.P("}).Subscribe(ctx)")
		g.P()
		g.P("outchan, outerr := p.pp.", strings.Title(method.GetName()), "(ctx, inchan, inerr, ud)")
		g.P("return flux.Create(func(ctx context.Context, sink flux.Sink) {")
		g.P("loop:")
		g.P("for {")
		g.P("select {")
		g.P("case i, o := <-outchan:")
		g.P("if o {")
		g.P("bytes, e := proto.Marshal(i)")
		g.P("if e != nil {")
		g.P("sink.Error(e)")
		g.P("break loop")
		g.P("}")
		g.P("p := payload.New(bytes, nil)")
		g.P("sink.Next(p)")
		g.P("} else {")
		g.P("break loop")
		g.P("}")
		g.P("case err := <-outerr:")
		g.P("if err != nil {")
		g.P("sink.Error(err)")
		g.P("}")
		g.P("}")
		g.P("}")
		g.P("})")
	}

	g.P("default:")
	g.P("return flux.Error(fmt.Errorf(\"unknown method %s\", method))")
	g.P("}")
	g.P("})")
	g.P("}")
}

func (g *rrpc) generateServerFireAndForget(service *descriptor.ServiceDescriptorProto) {

}

func (g *rrpc) generateServerConstructor(service *descriptor.ServiceDescriptorProto) {
	g.P("func New", service.GetName(), "Server(p ", service.GetName(), ") *", service.GetName(), "Server {")
	g.P("return &", service.GetName(), "Server{")
	g.P("pp: p,")
	g.P("}")
	g.P("}")

}

func (g *rrpc) GenerateImports(file *generator.FileDescriptor) {

}

func (g *rrpc) P(args ...interface{}) { g.gen.P(args...) }

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
