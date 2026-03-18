package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/cybint/sao/internal/config"
	"github.com/cybint/sao/internal/entity"
	"github.com/cybint/sao/internal/health"
	"github.com/cybint/sao/internal/schema"
)

// Server wraps the HTTP server used by SAO.
type Server struct {
	httpServer *http.Server
}

// New creates a server instance with default routes.
func New(_ config.Config) *Server {
	mux := http.NewServeMux()
	entityStore := entity.NewStore()
	schemaRegistry := schema.NewRegistry()
	mux.Handle("/healthz", health.Handler("sao"))
	mux.Handle("/entities", entity.CollectionHandler(entityStore))
	mux.Handle("/entities/", entity.ItemHandler(entityStore))
	mux.Handle("/schemas", schema.Handler(schemaRegistry))

	return &Server{
		httpServer: &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// Serve runs the HTTP server on a provided listener.
func (s *Server) Serve(listener net.Listener) error {
	err := s.httpServer.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Handler returns the HTTP handler for testing and composition.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}
