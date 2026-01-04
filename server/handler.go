package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ethansaxenian/rss/log"
)

type APIFunc func(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error

// Handle wraps an [APIFunc] into an [http.HandlerFunc].
func (s *Server) Handle(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		conn, err := s.db.Conn(ctx)
		if err != nil {
			log.Add(ctx, slog.GroupAttrs("error", slog.String("message", fmt.Sprintf("opening database conn: %v", err))))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		if err := h(conn, w, r); err != nil {
			log.Add(ctx, slog.GroupAttrs("error", slog.String("message", err.Error())))

			var apiErr APIError
			if errors.As(err, &apiErr) {
				http.Error(w, err.Error(), apiErr.StatusCode)
			} else {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
	}
}
