-- name: CreateItem :exec
INSERT INTO items(feed_id, title, link, description, published_at) VALUES (?, ?, ?, ?, ?);

-- name: ListUnread :many
SELECT items.*, sqlc.embed(feeds) FROM items
JOIN feeds ON items.feed_id = feeds.id
WHERE items.status = 'unread'
ORDER BY items.published_at DESC
LIMIT ? OFFSET ?;

-- name: MarkRead :exec
UPDATE items SET status = 'read' WHERE id = ?;
