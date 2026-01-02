-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS items_feed_id_ix ON items(feed_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS items_feed_id_ix;
-- +goose StatementEnd
