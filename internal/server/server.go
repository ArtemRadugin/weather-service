package server

import (
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	context context.Context
	host    string
	port    string
	handler http.Handler
}

func NewServer(
	ctx context.Context,
	host string,
	port string,
	handler http.Handler,
) *Server {
	return &Server{
		context: ctx,
		host:    host,
		port:    port,
		handler: handler,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := http.ListenAndServe(addr, s.handler)
	if err != nil {
		panic(err)
	}
	return nil
}
