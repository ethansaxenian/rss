package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethansaxenian/rss/database"
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
	triggerChan chan struct{}
}

func New(db *sql.DB) *Worker {
	return &Worker{
		db:          db,
		triggerChan: make(chan struct{}, 1),
	}
}

func (w *Worker) ForceRefresh() {
	w.triggerChan <- struct{}{}
}

func (w *Worker) RunLoop(ctx context.Context) {
	ticker := time.Tick(refreshLoopInterval)

	for {
		select {
		case <-ticker:
			if err := w.refreshFeeds(ctx, false); err != nil {
				slog.Error("Error refreshing feeds", "error", err)
			}
		case <-w.triggerChan:
			if err := w.refreshFeeds(ctx, true); err != nil {
				slog.Error("Error refreshing feeds", "error", err)
			}
		case <-ctx.Done():
			slog.Info("Context cancelled, exiting.")
			return
		}
	}
}

func (w *Worker) refreshFeeds(ctx context.Context, force bool) error {
	if force {
		slog.Info("Starting feed refresh.", "forced", true)
	} else {
		slog.Info("Starting feed refresh.", "forced", false, "next_refresh", time.Now().UTC().Add(refreshLoopInterval))
	}

	var eg errgroup.Group
	eg.SetLimit(maxConcurrentRefreshes)

	q := database.New(w.db)
	feeds, err := q.ListFeeds(ctx)
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	slog.Info("Found feeds.", "num_feeds", len(feeds))

	for _, feed := range feeds {
		eg.Go(func() error {
			feedCtx, cancel := context.WithTimeout(ctx, feedRefreshTimeout)
			defer cancel()
			return refreshFeed(feedCtx, w.db, feed)
		})
	}

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("refreshing feeds: %w", err)
	}

	return nil
}

func refreshFeed(ctx context.Context, db *sql.DB, feed database.Feed) error {
	logger := slog.With("feed_id", feed.ID, "url", feed.URL)

	now := time.Now().UTC()

	if feed.LastRefreshedAt != nil && feed.LastRefreshedAt.Add(refreshThrottleInverval).After(now) {
		logger.Warn("Refresh triggered too quickly. Try again later.", "can_refresh_at", feed.LastRefreshedAt.Add(refreshThrottleInverval))
		return nil
	}

	logger.Info("Refreshing feed.")

	res, err := fetchFeed(ctx, feed.URL)
	if err != nil {
		return fmt.Errorf("fetching feed: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	q := database.New(db).WithTx(tx)

	var newItems int
	for _, item := range res.Items {
		if feed.LastRefreshedAt != nil && item.PubDate.Before(*feed.LastRefreshedAt) {
			continue
		}

		if err := q.CreateItem(
			ctx,
			database.CreateItemParams{
				FeedID:      feed.ID,
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				PublishedAt: item.PubDate.Time,
			},
		); err != nil {
			return fmt.Errorf("creating item: %w", err)
		}

		newItems++
	}

	if err := q.UpdateFeedLastRefreshedAt(ctx, feed.ID); err != nil {
		logger.Error("Failed to update last_refreshed_at.")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	logger.Info("Successfully refreshed feed.", "new_items", newItems)

	return nil
}
