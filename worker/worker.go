package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/ethansaxenian/rss/database"
	"github.com/mmcdole/gofeed"
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
	triggerChan chan struct{}
	log         *slog.Logger
}

func New(db *sql.DB, logger *slog.Logger) *Worker {
	return &Worker{
		db:          db,
		triggerChan: make(chan struct{}, 1),
		log:         logger,
	}
}

func (w *Worker) ForceRefresh() {
	w.triggerChan <- struct{}{}
}

func (w *Worker) RunLoop(ctx context.Context) {
	w.log.Info("Starting worker")
	ticker := time.Tick(refreshLoopInterval)

	for {
		select {
		case <-ticker:
			if err := w.refreshFeeds(ctx, false); err != nil {
				w.log.Error("Error refreshing feeds", "error", err)
			}
		case <-w.triggerChan:
			if err := w.refreshFeeds(ctx, true); err != nil {
				w.log.Error("Error refreshing feeds", "error", err)
			}
		case <-ctx.Done():
			w.log.Info("Context cancelled, exiting.")
			return
		}
	}
}

func (w *Worker) refreshFeeds(ctx context.Context, force bool) error {
	if force {
		w.log.Info("Starting feed refresh.", "forced", true)
	} else {
		w.log.Info("Starting feed refresh.", "forced", false, "next_refresh", time.Now().UTC().Add(refreshLoopInterval).Local())
	}

	var eg errgroup.Group
	eg.SetLimit(maxConcurrentRefreshes)

	q := database.New(w.db)
	feeds, err := q.ListFeeds(ctx)
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	w.log.Info("Found feeds.", "num_feeds", len(feeds))

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

	now := time.Now().UTC()

	if feed.LastRefreshedAt != nil && feed.LastRefreshedAt.Add(refreshThrottleInverval).After(now) {
		logger.Warn("Refresh triggered too quickly. Try again later.", "can_refresh_at", feed.LastRefreshedAt.Add(refreshThrottleInverval).Local())
		return nil
	}

	logger.Info("Refreshing feed.")

	res, err := gofeed.NewParser().ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		return fmt.Errorf("parsing feed URL: %w", err)
	}

	w.dbMu.Lock()
	defer w.dbMu.Unlock()

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	q := database.New(w.db).WithTx(tx)

	var numNewItems int
	for _, item := range res.Items {
		if feed.LastRefreshedAt != nil && item.PublishedParsed.Before(*feed.LastRefreshedAt) {
			continue
		}

		if err := q.CreateItem(
			ctx,
			database.CreateItemParams{
				FeedID:      feed.ID,
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				PublishedAt: (*item.PublishedParsed).UTC(), // Store as UTC
			},
		); err != nil {
			return fmt.Errorf("creating item: %w", err)
		}

		numNewItems++
	}

	if res.Image != nil {
		if err := q.UpdateFeedImage(ctx, database.UpdateFeedImageParams{Image: &res.Image.URL, ID: feed.ID}); err != nil {
			logger.Error("Failed to update feeds.image.")
		}
	}

	if err := q.UpdateFeedLastRefreshedAt(ctx, feed.ID); err != nil {
		logger.Error("Failed to update feeds.last_refreshed_at.")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	logger.Info("Successfully refreshed feed.", "new_items", numNewItems)

	return nil
}
