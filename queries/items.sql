-- name: CreateItem :exec
INSERT INTO items(feed_id, title, link, description, hash, published_at) VALUES (?, ?, ?, ?, ?, ?);

-- name: ListItems :many
SELECT sqlc.embed(items), sqlc.embed(feeds) FROM items
JOIN feeds ON items.feed_id = feeds.id
WHERE (CAST (@has_status AS BOOL)  = 0 OR items.status  = @status)
AND   (CAST (@has_feed_id AS BOOL) = 0 OR items.feed_id = @feed_id)
ORDER BY items.published_at DESC
LIMIT ? OFFSET ?;

-- name: CountItems :one
SELECT COUNT(*) FROM items
WHERE (CAST (@has_status AS BOOL)  = 0 OR items.status  = @status)
AND   (CAST (@has_feed_id AS BOOL) = 0 OR items.feed_id = @feed_id);

-- name: UpdateItem :exec
UPDATE items SET title = ?, link = ?, description = ?, published_at = ? WHERE id = ?;

-- name: UpdateItemStatus :one
UPDATE items SET status = ? WHERE id = ? RETURNING *;

-- name: MarkAllItemsAsRead :exec
UPDATE items SET status = "read" WHERE status = "unread";

-- name: CheckItemExists :one
SELECT * FROM items WHERE feed_id = ? AND hash = ?;
