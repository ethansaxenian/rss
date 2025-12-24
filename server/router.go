package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ethansaxenian/rss/components"
	"github.com/ethansaxenian/rss/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	defaultPageSize = 5
)

func (s *Server) NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}}))
	r.Use(middleware.RedirectSlashes)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/unread", http.StatusMovedPermanently)
	})

	r.Get("/unread", s.Handle(unread))
	r.Post("/items/{id:^[0-9]+}/read", s.Handle(readItem))

	return r
}

func unread(ctx context.Context, log *slog.Logger, conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		log.Warn("Unable to parse page", "page", query.Get("page"))
	}

	q := database.New(conn)
	items, err := q.ListUnread(
		ctx,
		database.ListUnreadParams{
			Limit:  defaultPageSize,
			Offset: int64(page) * defaultPageSize,
		},
	)
	if err != nil {
		return fmt.Errorf("listing unread items: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return components.UnreadPage(items, page).Render(ctx, w)
}

func readItem(ctx context.Context, log *slog.Logger, conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, fmt.Errorf("parsing item ID: %w", err))
	}

	q := database.New(conn)
	err = q.MarkRead(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("updating item status: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
