package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ethansaxenian/rss/database"
	"github.com/ethansaxenian/rss/rss"
	"golang.org/x/sync/errgroup"
)

const (
	refreshLoopInterval     = 60 * time.Minute
	maxConcurrentRefreshes  = 5
	feedRefreshTimeout      = 15 * time.Second
	refreshThrottleInverval = 10 * time.Minute
)

type Worker struct {
	db          *sql.DB
	dbMu        sync.Mutex
	refreshChan chan struct{}
	log         *slog.Logger
}

func New(db *sql.DB, logger *slog.Logger) *Worker {
	return &Worker{
		db:          db,
		refreshChan: make(chan struct{}, 1),
		log:         logger,
	}
}

func (w *Worker) RefreshAll() {
	w.refreshChan <- struct{}{}
}

func (w *Worker) RunLoop(ctx context.Context) {
	w.log.Info("Starting worker")
	ticker := time.Tick(refreshLoopInterval)

	for {
		select {
		case <-ticker:
			if err := w.refreshFeeds(ctx); err != nil {
				w.log.Error("Error refreshing feeds", "error", err)
			}
		case <-w.refreshChan:
			if err := w.refreshFeeds(ctx); err != nil {
				w.log.Error("Error refreshing feeds", "error", err)
			}
		case <-ctx.Done():
			w.log.Info("Context cancelled, exiting.")
			return
		}
	}
}

func (w *Worker) refreshFeeds(ctx context.Context) error {
	var eg errgroup.Group
	eg.SetLimit(maxConcurrentRefreshes)

	q := database.New(w.db)
	feeds, err := q.ListFeeds(ctx)
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	w.log.Debug("Found feeds.", "num_feeds", len(feeds))

	for _, feed := range feeds {
		eg.Go(func() error {
			feedCtx, cancel := context.WithTimeout(ctx, feedRefreshTimeout)
			defer cancel()
			if err = w.refreshFeed(feedCtx, feed); err != nil {
				w.log.Error("Error refreshing feed", "feed_id", feed.ID, "url", feed.URL, "error", err)
			}
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("refreshing feeds: %w", err)
	}

	return nil
}

func (w *Worker) refreshFeed(ctx context.Context, feed database.Feed) error {
	logger := w.log.With("feed_id", feed.ID, "url", feed.URL)

	// now := time.Now().UTC()

	// if feed.LastRefreshedAt != nil && feed.LastRefreshedAt.Add(refreshThrottleInverval).After(now) {
	// 	logger.Warn("Refresh triggered too quickly. Try again later.", "can_refresh_at", feed.LastRefreshedAt.Add(refreshThrottleInverval).Local())
	// 	return nil
	// }

	logger.Info("Refreshing feed.")

	res, err := rss.FetchFeed(ctx, feed.URL)
	if err != nil {
		return fmt.Errorf("fetching feed URL: %w", err)
	}

	w.dbMu.Lock()
	defer w.dbMu.Unlock()

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	q := database.New(w.db).WithTx(tx)

	numNewItems, numUpdatedItems, err := rss.UpdateFeedItems(ctx, q, feed.ID, res, logger)
	if err != nil {
		return fmt.Errorf("updating feed items: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	logger.Info("Successfully refreshed feed.", "new_items", numNewItems, "updated_items", numUpdatedItems)

	return nil
}
