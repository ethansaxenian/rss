-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS items_feed_id_ix ON items(feed_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE INDEX IF EXISTS items_feed_id_ix ON items(feed_id);
-- +goose StatementEnd
