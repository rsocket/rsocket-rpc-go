package rrpc

import (
	"fmt"
	"strconv"
	"strings"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

const (
	contextPkgPath = "context"
	rrpcPkgPath    = "github.com/rsocket/rsocket-rpc-go"
	rsocketPkgPath = "github.com/rsocket/rsocket-go"
)

var (
	contextPkg string
	rrpcPkg    string
	rsocketPkg string
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
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	contextPkg = string(g.gen.AddImport(contextPkgPath))
	rrpcPkg = string(g.gen.AddImport(rrpcPkgPath))
	rsocketPkg = string(g.gen.AddImport(rsocketPkgPath))

	g.P("var _ ", contextPkg, ".Context")
	g.P("var _ ", rrpcPkg, ".ClientConn")
	g.P("var _ ", rsocketPkg, ".RSocket")

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
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
		g.P("{")
		g.P("Name: ", strconv.Quote(method.GetName()), ",")
		g.P("Handler: ", handlerNames[i], ",")
		g.P("},")
	}
	g.P("},")
	g.P("}")
	g.P()
}

func (g *rrpc) generateServerSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	var reqArgs []string
	reqArgs = append(reqArgs, contextPkg+".Context")
	ret := "(*" + g.typeName(method.GetOutputType()) + ", error)"
	reqArgs = append(reqArgs, "*"+g.typeName(method.GetInputType()))
	reqArgs = append(reqArgs, rrpcPkg+".Metadata")
	return methName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
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
	}
	// TODO: support stream
}

func (g *rrpc) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	// TODO: support stream
	reqArg := ", in *" + g.typeName(method.GetInputType())
	respName := "*" + g.typeName(method.GetOutputType())
	return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) (%s, error)", methName, contextPkg, reqArg, rrpcPkg, respName)
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

func (g *rrpc) GenerateImports(file *generator.FileDescriptor) {
}

func (g *rrpc) P(args ...interface{}) { g.gen.P(args...) }

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
