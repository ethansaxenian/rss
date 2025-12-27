package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/ethansaxenian/rss/components"
	"github.com/ethansaxenian/rss/contextkeys"
	"github.com/ethansaxenian/rss/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	defaultPageSize = 20
)

func (s *Server) NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}}))
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.NoCache)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/unread", http.StatusMovedPermanently)
	})

	r.Get("/unread", s.Handle(s.unreadPage))
	r.Get("/unread/list", s.Handle(s.unreadItemList))
	r.Get("/history", s.Handle(s.historyPage))
	r.Get("/history/list", s.Handle(s.historyItemList))
	r.Get("/feeds", s.Handle(s.feedsPage))
	r.Get("/feeds/{id:^[0-9]+}", s.Handle(s.feedPage))
	r.Get("/feeds/{id:^[0-9]+}/list", s.Handle(s.feedItemList))
	r.Post("/feeds/refresh", s.Handle(s.refreshFeeds))
	r.Put("/items/{id:^[0-9]+}/status", s.Handle(s.status))
	r.Post("/items/read-all", s.Handle(s.readAll))

	return r
}

func (s *Server) unreadPage(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	q := database.New(conn)
	count, err := q.CountItems(
		ctx,
		database.CountItemsParams{
			HasStatus: true,
			Status:    database.StatusUnread,
		},
	)
	if err != nil {
		return fmt.Errorf("counting unread items: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return components.UnreadPage(count).Render(ctx, w)
}

func (s *Server) historyPage(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	q := database.New(conn)
	count, err := q.CountItems(
		ctx,
		database.CountItemsParams{
			HasStatus: true,
			Status:    database.StatusRead,
		},
	)
	if err != nil {
		return fmt.Errorf("counting read items: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return components.HistoryPage(count).Render(ctx, w)
}

func (s *Server) feedPage(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, fmt.Errorf("parsing feed ID: %w", err))
	}

	q := database.New(conn)
	feed, err := q.GetFeed(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("getting feed: %w", err)
	}

	count, err := q.CountItems(
		ctx,
		database.CountItemsParams{
			HasFeedID: true,
			FeedID:    int64(id),
		},
	)
	if err != nil {
		return fmt.Errorf("counting read items: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return components.FeedPage(feed, count).Render(ctx, w)
}

func (s *Server) listItems(conn *sql.Conn, w http.ResponseWriter, r *http.Request, status database.Status, feedID int64) error {
	ctx := r.Context()

	query := r.URL.Query()
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		page = 0
	}

	q := database.New(conn)
	items, err := q.ListItems(
		ctx,
		database.ListItemsParams{
			HasStatus: status != database.StatusAny,
			Status:    status,
			HasFeedID: feedID != 0,
			FeedID:    feedID,
			Limit:     defaultPageSize,
			Offset:    int64(page) * defaultPageSize,
		},
	)
	if err != nil {
		return fmt.Errorf("listing %s items: %w", status, err)
	}

	ctx = contextkeys.WithRoutePathCtx(r.Context(), r.URL.Path)

	w.WriteHeader(http.StatusOK)
	return components.ItemsList(items, page).Render(ctx, w)
}

func (s *Server) unreadItemList(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	return s.listItems(conn, w, r, database.StatusUnread, 0)
}

func (s *Server) historyItemList(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	return s.listItems(conn, w, r, database.StatusRead, 0)
}

func (s *Server) feedItemList(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, fmt.Errorf("parsing feed ID: %w", err))
	}

	return s.listItems(conn, w, r, database.StatusAny, int64(id))
}

func (s *Server) status(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, fmt.Errorf("parsing item ID: %w", err))
	}

	query := r.URL.Query()
	status := database.Status(query.Get("status"))

	if !slices.Contains(database.AllStatusValues(), status) {
		return NewAPIError(http.StatusBadRequest, fmt.Errorf("unknown status: %s", status)) //nolint:err113
	}

	q := database.New(conn)
	item, err := q.UpdateItemStatus(ctx, database.UpdateItemStatusParams{Status: status, ID: int64(id)})
	if err != nil {
		return fmt.Errorf("updating item status: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return components.MarkAs(item).Render(ctx, w)
}

func (s *Server) feedsPage(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	q := database.New(conn)
	feeds, err := q.ListFeeds(ctx)
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return components.FeedsPage(feeds).Render(ctx, w)
}

func (s *Server) readAll(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	q := database.New(conn)
	err := q.MarkAllItemsAsRead(ctx)
	if err != nil {
		return fmt.Errorf("marking items as read: %w", err)
	}

	http.Redirect(w, r, "/unread", http.StatusFound)
	return nil
}

func (s *Server) refreshFeeds(conn *sql.Conn, w http.ResponseWriter, r *http.Request) error {
	s.worker.ForceRefresh()
	w.WriteHeader(http.StatusOK)
	return nil
}
