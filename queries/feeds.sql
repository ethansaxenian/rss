-- name: ListFeeds :many
SELECT * FROM feeds ORDER BY created_at DESC;

-- name: CreateFeed :exec
INSERT INTO feeds(title, url) VALUES (?, ?);

-- name: UpdateFeedLastRefreshedAt :exec
UPDATE feeds SET last_refreshed_at = CURRENT_TIMESTAMP WHERE id = ?;
