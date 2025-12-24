package server

import (
	"database/sql"
	"errors"
	"net/http"
)

type APIFunc func(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error

// Handle wraps an [APIFunc] into an [http.HandlerFunc].
func (s *Server) Handle(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		conn, err := s.db.Conn(ctx)
		if err != nil {
			s.log.Error("Failed to open database conn.", "error", err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		if err := h(conn, w, r); err != nil {
			s.log.Error("Error in handler func", "error", err.Error(), "method", r.Method, "path", r.URL, "remoteAddr", r.RemoteAddr)

			var apiErr APIError
			if errors.As(err, &apiErr) {
				http.Error(w, err.Error(), apiErr.StatusCode)
			} else {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
	}
}
