package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/d6o/homeclip/internal/infrastructure/config"
)

type Server struct {
	httpServer *http.Server
	port       string
}

func NewServer(handler http.Handler, cfg *config.Config) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		port: cfg.Port,
	}
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Port() string {
	return s.port
}

func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost:%s", s.port)
}