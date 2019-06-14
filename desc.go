package rrpc

type methodHandler func(srv interface{}, dec func(interface{}) error) (interface{}, error)

type ServiceDesc struct {
	Name        string
	HandlerType interface{}
	Methods     []MethodDesc
	Metadata    interface{}
}

type MethodDesc struct {
	Name    string
	Handler methodHandler
}
