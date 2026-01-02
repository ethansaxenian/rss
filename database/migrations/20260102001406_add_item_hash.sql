-- +goose Up
-- +goose StatementBegin
DELETE FROM items; -- unavoidable :(
ALTER TABLE items ADD COLUMN hash TEXT NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS items_feed_id_hash ON items(feed_id, hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS items_feed_id_hash;

ALTER TABLE items DROP COLUMN hash;
-- +goose StatementEnd
