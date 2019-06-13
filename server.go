package rrpc

type Server struct {
	r *serviceRegistry
}

func (p *Server) Serve() error {
	panic("todo")
}

func NewServer() *Server {
	return &Server{
		r: newRegister(),
	}
}
