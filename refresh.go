package main

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
	refreshLoopInterval    = 1 * time.Minute
	maxConcurrentRefreshes = 5
	feedRefreshTimeout     = 15 * time.Second
)

func refreshLoop(ctx context.Context) {
	ticker := time.Tick(refreshLoopInterval)

	for {
		if err := refreshFeeds(ctx); err != nil {
			slog.Error("Error refreshing feeds", "error", err)
		}
		select {
		case <-ticker:
			continue
		case <-ctx.Done():
			slog.Info("Context cancelled, exiting.")
			return
		}
	}
}

func refreshFeeds(ctx context.Context) error {
	slog.Info("Starting feed refresh.", "next_refresh", time.Now().Add(refreshLoopInterval))

	db, err := sql.Open("sqlite", database.DSN)
	if err != nil {
		return fmt.Errorf("opening db connection: %w", err)
	}

	var eg errgroup.Group
	eg.SetLimit(maxConcurrentRefreshes)

	q := database.New(db)
	feeds, err := q.ListFeeds(ctx)
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	slog.Info("Found feeds.", "num_feeds", len(feeds))

	for _, feed := range feeds {
		eg.Go(func() error {
			feedCtx, cancel := context.WithTimeout(ctx, feedRefreshTimeout)
			defer cancel()
			return refreshFeed(feedCtx, db, feed)
		})
	}

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("refreshing feeds: %w", err)
	}

	return nil
}

func refreshFeed(ctx context.Context, db *sql.DB, feed database.Feed) error {
	logger := slog.With("id", feed.ID, "url", feed.URL)
	logger.Info("Refreshing feed.")

	res, err := fetchFeed(ctx, feed.URL)
	if err != nil {
		return fmt.Errorf("fetching feed: %w", err)
	}

	tx, err := db.Begin()
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
