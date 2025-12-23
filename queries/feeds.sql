-- name: ListFeeds :many
SELECT * FROM feeds;

-- name: CreateFeed :one
INSERT INTO feeds(title, url) VALUES (?, ?) RETURNING *;
