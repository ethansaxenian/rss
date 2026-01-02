package rss

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ethansaxenian/rss/database"
	"github.com/mmcdole/gofeed"
)

func FetchFeed(ctx context.Context, url string) (*gofeed.Feed, error) {
	return gofeed.NewParser().ParseURLWithContext(url, ctx) //nolint: wrapcheck
}

func GetItemHash(item *gofeed.Item) string {
	var toHash string
	switch {
	case item.GUID != "":
		toHash = item.GUID
	case item.Link != "":
		toHash = item.Link
	default:
		toHash = item.Title + item.Content
	}
	hashBytes := sha256.Sum256([]byte(toHash))
	hash := hex.EncodeToString(hashBytes[:])

	return hash
}

func shouldUpdateItem(item *gofeed.Item, existingItem database.Item) bool {
	return item.Title != existingItem.Title ||
		item.Link != existingItem.Link ||
		item.Description != existingItem.Description ||
		*item.PublishedParsed != existingItem.PublishedAt
}

func UpdateFeedItems(ctx context.Context, q *database.Queries, feedID int64, feed *gofeed.Feed, logger *slog.Logger) (int, int, error) {
	var numNewItems int
	var numUpdatedItems int
	for _, item := range feed.Items {
		if strings.HasPrefix(item.Link, "https://youtube.com/shorts/") {
			continue
		}

		hash := GetItemHash(item)

		existingItem, existsErr := q.CheckItemExists(ctx, database.CheckItemExistsParams{FeedID: feedID, Hash: hash})
		if existsErr == nil {
			if shouldUpdateItem(item, existingItem) {
				if err := q.UpdateItem(
					ctx,
					database.UpdateItemParams{
						Title:       item.Title,
						Link:        item.Link,
						Description: item.Description,
						PublishedAt: (*item.PublishedParsed).UTC(), // Store as UTC
						ID:          existingItem.ID,
					},
				); err != nil {
					return 0, 0, fmt.Errorf("updating item: %w", err)
				}

				numUpdatedItems++
			}

		} else if !errors.Is(existsErr, sql.ErrNoRows) {
			logger.Error("Error checking if item exists", "hash", hash, "error", existsErr)
			continue

		} else {
			if err := q.CreateItem(
				ctx,
				database.CreateItemParams{
					FeedID:      feedID,
					Title:       item.Title,
					Link:        item.Link,
					Hash:        hash,
					Description: item.Description,
					PublishedAt: (*item.PublishedParsed).UTC(), // Store as UTC
				},
			); err != nil {
				return 0, 0, fmt.Errorf("creating item: %w", err)
			}

			numNewItems++
		}
	}

	if feed.Image != nil {
		if err := q.UpdateFeedImage(ctx, database.UpdateFeedImageParams{Image: &feed.Image.URL, ID: feedID}); err != nil {
			logger.Error("Failed to update feeds.image.")
		}
	}

	if err := q.UpdateFeedLastRefreshedAt(ctx, feedID); err != nil {
		logger.Error("Failed to update feeds.last_refreshed_at.")
	}

	return numNewItems, numUpdatedItems, nil
}
