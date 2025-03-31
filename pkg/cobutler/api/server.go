package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Server represents an HTTP server for the API
type Server struct {
	server *http.Server
	port   string
}

// NewServer creates a new server with the given handler and port
func NewServer(handler *Handler, port string) *Server {
	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	return &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mux,
		},
		port: port,
	}
}

// Start starts the server in a goroutine
func (s *Server) Start() error {
	go func() {
		slog.Info("Server starting", "port", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
		}
	}()
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	slog.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	slog.Info("Server stopped")
	return nil
}
