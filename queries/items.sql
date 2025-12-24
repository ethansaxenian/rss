-- name: CreateItem :exec
INSERT INTO items(feed_id, title, link, description, published_at) VALUES (?, ?, ?, ?, ?);
