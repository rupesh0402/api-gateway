package server

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(addr string, handler http.Handler, readTimeout, writeTimeout, idleTimeout time.Duration) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}
}

func (s *Server) Start() {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", s.httpServer.Addr, err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server ...")
	return s.httpServer.Shutdown(ctx)
}
