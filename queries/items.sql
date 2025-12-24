-- name: CreateItem :exec
INSERT INTO items(feed_id, title, link, description, published_at) VALUES (?, ?, ?, ?, ?);

-- name: ListItems :many
SELECT items.*, sqlc.embed(feeds) FROM items
JOIN feeds ON items.feed_id = feeds.id
WHERE items.status = ?
ORDER BY items.published_at DESC
LIMIT ? OFFSET ?;

-- name: CountItems :one
SELECT COUNT(*) FROM items WHERE items.status = ?;

-- name: UpdateItemStatus :exec
UPDATE items SET status = ? WHERE id = ?;

-- name: MarkAllItemsAsRead :exec
UPDATE items SET status = "read" WHERE status = "unread";
