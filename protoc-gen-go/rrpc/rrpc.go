package rrpc

import (
	"fmt"
	"strconv"
	"strings"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

const (
	atomicPkgPath  = "sync/atomic"
	contextPkgPath = "context"
	rrpcPkgPath    = "github.com/rsocket/rsocket-rpc-go"
	rsocketPkgPath = "github.com/rsocket/rsocket-go"
	rxPkgPath      = "github.com/rsocket/rsocket-go/rx"
	payloadPkgPath = "github.com/rsocket/rsocket-go/payload"
)

var (
	atomicPkg  string
	contextPkg string
	rrpcPkg    string
	rsocketPkg string
	rxPkg      string
	payloadPkg string
)

var fluxNames []string

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
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	atomicPkg = string(g.gen.AddImport(atomicPkgPath))
	contextPkg = string(g.gen.AddImport(contextPkgPath))
	rrpcPkg = string(g.gen.AddImport(rrpcPkgPath))
	rsocketPkg = string(g.gen.AddImport(rsocketPkgPath))
	rxPkg = string(g.gen.AddImport(rxPkgPath))
	payloadPkg = string(g.gen.AddImport(payloadPkgPath))

	g.P("var _ ", contextPkg, ".Context")
	g.P("var _ ", rrpcPkg, ".ClientConn")
	g.P("var _ ", rsocketPkg, ".RSocket")
	g.P("var _ ", rxPkg, ".Flux")
	g.P("var _ ", payloadPkg, ".Payload")

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}

	finish := make(map[string]bool)
	for _, it := range fluxNames {
		_, ok := finish[it]
		if ok {
			continue
		}
		finish[it] = true
		g.generateFlux(it)
		// test
		g.generateMono(it)
	}
}

func (g *rrpc) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	path := fmt.Sprintf("6,%d", index) // 6 means service.
	origServName := service.GetName()
	fullServName := origServName
	if pkg := file.GetPackage(); pkg != "" {
		fullServName = pkg + "." + fullServName
	}
	servName := generator.CamelCase(origServName)
	g.P()

	// Client interface.
	g.P("type ", servName, "Client interface {")
	for i, method := range service.Method {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i))
		g.P(g.generateClientSignature(servName, method))
	}
	g.P("}")
	g.P()

	// Client structure.
	g.P("type ", unexport(servName), "Client struct {")
	g.P("cc *", rrpcPkg, ".ClientConn")
	g.P("}")
	g.P()

	// NewClient factory.
	g.P("func New", servName, "Client(s ", rsocketPkg, ".RSocket, m ", rrpcPkg, ".MeterRegistry, t ", rrpcPkg, ".Tracer) ", servName, "Client {")
	g.P("cc := ", rrpcPkg, ".NewClientConn(s, m, t)")
	g.P("return &", unexport(servName), "Client{cc}")
	g.P("}")
	g.P()

	serviceDescVar := "_" + servName + "_serviceDesc"
	for _, method := range service.Method {
		g.generateClientMethod(servName, fullServName, serviceDescVar, method)
	}

	// Server interface.
	serverType := servName + "Server"

	g.P("type ", serverType, " interface {")
	for i, method := range service.Method {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		g.P(g.generateServerSignature(servName, method))
	}
	g.P("}")
	g.P()

	// Server registration.
	g.P("func Register", servName, "Server(s *", rrpcPkg, ".Server, srv ", serverType, ") {")
	g.P("s.RegisterService(&", serviceDescVar, `, srv)`)
	g.P("}")
	g.P()

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Method {
		hname := g.generateServerMethod(servName, fullServName, method)
		handlerNames = append(handlerNames, hname)
	}

	// Service descriptor.
	g.P("var ", serviceDescVar, " = ", rrpcPkg, ".ServiceDesc {")
	g.P("Name: ", strconv.Quote(fullServName), ",")
	g.P("HandlerType: (*", serverType, ")(nil),")
	g.P("Methods: []", rrpcPkg, ".MethodDesc{")
	for i, method := range service.Method {
		if !method.GetClientStreaming() && !method.GetServerStreaming() {
			g.P("{")
			g.P("Name: ", strconv.Quote(method.GetName()), ",")
			g.P("Handler: ", handlerNames[i], ",")
			g.P("},")
		}
	}
	g.P("},")
	g.P("Streams: []", rrpcPkg, ".StreamDesc{")
	for i, method := range service.Method {
		if !method.GetClientStreaming() && method.GetServerStreaming() {
			g.P("{")
			g.P("Name: ", strconv.Quote(method.GetName()), ",")
			g.P("Handler: ", handlerNames[i], ",")
			g.P("},")
		}
	}
	g.P("},")
	g.P("}")
	g.P()
}

func (g *rrpc) fluxName(msgType string) string {
	typ := g.typeName(msgType)
	return "Flux" + typ
}

func (g *rrpc) generateMono(msgType string) {
	typ := g.typeName(msgType)

	g.P("type Mono", typ, " struct {")
	g.P("m ", rxPkg, ".Mono")
	g.P("}")
	g.P()
	g.P("func (p *Mono", typ, ") Raw() ", rxPkg, ".Mono {")
	g.P("return p.m")
	g.P("}")
	g.P()

	g.P("func (p *Mono", typ, ") DoOnSuccess(fn func(", contextPkg, ".Context, ", rxPkg, ".Subscription, *", typ, ")) *Mono", typ, " {")
	g.P("p.m.DoOnSuccess(func(ctx ", contextPkg, ".Context, s ", rxPkg, ".Subscription, elem ", payloadPkg, ".Payload) {")
	g.P("o := new(", typ, ")")
	g.P("if err := proto.Unmarshal(elem.Data(), o); err != nil {")
	g.P("panic(err)")
	g.P("}")
	g.P("fn(ctx, s, o)")
	g.P("})")
	g.P("return p")
	g.P("}")
	g.P()
	g.P("func (p *Mono", typ, ") SubscribeOn(s ", rxPkg, ".Scheduler) *Mono", typ, " {")
	g.P("p.m.SubscribeOn(s)")
	g.P("return p")
	g.P("}")
	g.P()
	g.P("func (p *Mono", typ, ") Subscribe(ctx ", contextPkg, ".Context) {")
	g.P("p.m.Subscribe(ctx)")
	g.P("}")
	g.P()
}

func (g *rrpc) generateFlux(msgType string) {
	typ := g.typeName(msgType)
	g.P("type Flux", typ, " struct {")
	g.P("f ", rxPkg, ".Flux")
	g.P("}")
	g.P()
	g.P("func (p *Flux", typ, ") Raw() ", rxPkg, ".Flux {")
	g.P("return p.f")
	g.P("}")
	g.P()
	g.P("func (p *Flux", typ, ") DoOnError(fn func(", contextPkg, ".Context, error)) *Flux", typ, " {")
	g.P("p.f.DoOnError(fn)")
	g.P("return p")
	g.P("}")
	g.P()
	g.P("func (p *Flux", typ, ") DoOnNext(fn func(", contextPkg, ".Context, ", rxPkg, ".Subscription, *", typ, ")) *Flux", typ, " {")
	g.P("p.f.DoOnNext(func(ctx ", contextPkg, ".Context, s ", rxPkg, ".Subscription, elem ", payloadPkg, ".Payload) {")
	g.P("o := new(", typ, ")")
	g.P("if err := proto.Unmarshal(elem.Data(), o); err != nil {")
	g.P("panic(err)")
	g.P("}")
	g.P("fn(ctx, s, o)")
	g.P("})")
	g.P("return p")
	g.P("}")
	g.P()
	g.P("func (p *Flux", typ, ") SubscribeOn(s ", rxPkg, ".Scheduler) *Flux", typ, " {")
	g.P("p.f.SubscribeOn(s)")
	g.P("return p")
	g.P("}")
	g.P()
	g.P("func (p *Flux", typ, ") Subscribe(ctx ", contextPkg, ".Context) {")
	g.P("p.f.Subscribe(ctx)")
	g.P("}")
	g.P()
	g.P("type Sink", typ, " interface {")
	g.P("Next(*", typ, ")")
	g.P("Error(error)")
	g.P("Complete()")
	g.P("}")
	g.P()
	g.P("type _Sink_", typ, " struct {")
	g.P("d ", rxPkg, ".Producer")
	g.P("n int32")
	g.P("}")
	g.P()
	g.P("func (p *_Sink_", typ, ") Error(err error) {")
	g.P("p.d.Error(err)")
	g.P("}")
	g.P()
	g.P("func (p *_Sink_", typ, ") Complete() {")
	g.P("p.d.Complete()")
	g.P("}")
	g.P()
	g.P("func (p *_Sink_", typ, ") Next(v *", typ, ") {")
	g.P("raw,err := proto.Marshal(v)")
	g.P("if err != nil {")
	g.P("panic(err)")
	g.P("}")
	g.P("if ", atomicPkg, ".AddInt32(&(p.n), 1) == 1 {")
	g.P("p.d.Next(", payloadPkg, ".New(raw, nil))")
	g.P("return")
	g.P("}")
	g.P("p.d.Next(", payloadPkg, ".New(raw, nil))")
	g.P("}")
	g.P()

	g.P("func NewFlux", typ, "(g func(", contextPkg, ".Context, Sink", typ, ")) *Flux", typ, " {")
	g.P("f := ", rxPkg, ".NewFlux(func(ctx ", contextPkg, ".Context, producer ", rxPkg, ".Producer) {")
	g.P("pd := &_Sink_", typ, "{")
	g.P("d: producer,")
	g.P("}")
	g.P("g(ctx, pd)")
	g.P("})")
	g.P("return &Flux", typ, "{f}")
	g.P("}")
	g.P()
}

func (g *rrpc) generateServerSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)

	// RequestResponse
	if !method.GetClientStreaming() && !method.GetServerStreaming() {
		var reqArgs []string
		reqArgs = append(reqArgs, contextPkg+".Context")
		ret := "(*" + g.typeName(method.GetOutputType()) + ", error)"
		reqArgs = append(reqArgs, "*"+g.typeName(method.GetInputType()))
		reqArgs = append(reqArgs, rrpcPkg+".Metadata")
		return methName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
	}
	// RequestStream
	if !method.GetClientStreaming() {
		var reqArgs []string
		reqArgs = append(reqArgs, contextPkg+".Context")
		ret := "*" + g.fluxName(method.GetOutputType())
		reqArgs = append(reqArgs, "*"+g.typeName(method.GetInputType()))
		reqArgs = append(reqArgs, rrpcPkg+".Metadata")
		return methName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
	}
	// RequestChannel
	// TODO: support channel
	panic("todo: bi-stream")
}

func (g *rrpc) generateClientMethod(servName, fullServName, serviceDescVar string, method *pb.MethodDescriptorProto) {
	methodName := method.GetName()
	outType := g.typeName(method.GetOutputType())
	g.P("func (c *", unexport(servName), "Client) ", g.generateClientSignature(servName, method), " {")
	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		g.P("out := new(", outType, ")")
		g.P(`err := c.cc.Invoke(ctx, "`, fullServName, `", "`, methodName, `", in, out, opts...)`)
		g.P("if err != nil { return nil, err }")
		g.P("return out, nil")
		g.P("}")
		g.P()
		return
	}
	// RequestStream
	if !method.GetClientStreaming() {
		fluxOut := g.fluxName(method.GetOutputType())
		g.P("dec := func(f ", rxPkg, ".Flux) interface{} {")
		g.P("return &", fluxOut, "{f}")
		g.P("}")
		g.P(`out := c.cc.InvokeStream(ctx, "`, fullServName, `", "`, methodName, `", in, dec, opts...)`)
		g.P("return out.(*", fluxOut, ")")
		g.P("}")
		g.P()
		return
	}
	// TODO: support channel
	panic("todo: bi-stream")
}

func (g *rrpc) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	// RequestResponse
	if !method.GetClientStreaming() && !method.GetServerStreaming() {
		reqArg := ", in *" + g.typeName(method.GetInputType())
		respName := "(*" + g.typeName(method.GetOutputType()) + ", error)"
		return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) %s", methName, contextPkg, reqArg, rrpcPkg, respName)
	}
	// RequestStream
	if !method.GetClientStreaming() {
		fluxNames = append(fluxNames, method.GetOutputType())
		reqArg := ", in *" + g.typeName(method.GetInputType())
		return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) *%s", methName, contextPkg, reqArg, rrpcPkg, g.fluxName(method.GetOutputType()))
	}
	// RequestChannel
	// TODO: support channel
	panic("todo: bi-stream ")
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *rrpc) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

func (g *rrpc) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

func (g *rrpc) generateServerMethod(servName, fullServName string, method *pb.MethodDescriptorProto) string {
	methName := generator.CamelCase(method.GetName())
	if !method.GetClientStreaming() && !method.GetServerStreaming() {
		hname := fmt.Sprintf("_%s_%s_Handler", servName, methName)
		inType := g.typeName(method.GetInputType())
		g.P("func ", hname, "(ctx ", contextPkg, ".Context, srv interface{}, dec func(interface{}) error, md ", rrpcPkg, ".Metadata) (interface{}, error) {")
		g.P("in := new(", inType, ")")
		g.P("err := dec(in)")
		g.P("if err != nil {")
		g.P("return nil, err")
		g.P("}")
		g.P("return srv.(", servName, "Server).", methName, "(ctx, in, md)")
		g.P("}")
		return hname
	}
	// RequestStream
	if !method.GetClientStreaming() {
		hname := fmt.Sprintf("_%s_%s_Handler", servName, methName)
		inType := g.typeName(method.GetInputType())
		g.P("func ", hname, "(ctx ", contextPkg, ".Context, srv interface{}, dec func(interface{}) error, md ", rrpcPkg, ".Metadata) (interface{}, error) {")
		g.P("in := new(", inType, ")")
		g.P("err := dec(in)")
		g.P("if err != nil {")
		g.P("return nil, err")
		g.P("}")
		g.P("return srv.(", servName, "Server).", methName, "(ctx, in, md), nil")
		g.P("}")
		return hname
	}
	panic("todo: bi stream")
}

func (g *rrpc) GenerateImports(file *generator.FileDescriptor) {

}

func (g *rrpc) P(args ...interface{}) { g.gen.P(args...) }

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
