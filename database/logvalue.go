package database

import "log/slog"

func (i Feed) LogValue() slog.Attr {
	return slog.GroupAttrs("feed",
		slog.Int64("id", i.ID),
		slog.String("title", i.Title),
		slog.String("url", i.URL),
	)
}

func (i Item) LogValue() slog.Attr {
	return slog.GroupAttrs("feed",
		slog.Int64("id", i.ID),
		slog.Int64("feed_id", i.FeedID),
		slog.String("title", i.Title),
		slog.String("url", i.Link),
		slog.String("status", string(i.Status)),
		slog.String("hash", i.Hash),
	)
}
