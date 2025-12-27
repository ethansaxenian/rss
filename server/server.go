package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/ethansaxenian/rss/worker"
)

type Server struct {
	db     *sql.DB
	port   int
	server *http.Server
	log    *slog.Logger
	worker *worker.Worker
}

func (s *Server) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("closing database: %w", err)
	}

	if err := s.server.Close(); err != nil {
		return fmt.Errorf("closing server: %w", err)
	}

	return nil
}

func (s *Server) ListenAndServe() error {
	if err := s.server.ListenAndServe(); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	return nil
}

func New(ctx context.Context, port int, db *sql.DB, worker *worker.Worker, logger *slog.Logger) *Server {
	s := &Server{
		db:     db,
		port:   port,
		log:    logger,
		worker: worker,
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.NewRouter(),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	s.server = server

	return s
}
