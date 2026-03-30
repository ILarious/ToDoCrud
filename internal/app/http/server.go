package http

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Handler:        handler,
			MaxHeaderBytes: 1 << 20,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
		},
	}
}

func (s *Server) Run(port string) error {
	s.httpServer.Addr = ":" + port
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
