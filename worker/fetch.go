package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	defaultFetchTimeout = 5 * time.Second
)

func fetchFeed(ctx context.Context, url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultFetchTimeout)
	defer cancel()

	fp := gofeed.NewParser()

	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("parsing feed URL: %w", url)
	}

	return feed, nil
}
