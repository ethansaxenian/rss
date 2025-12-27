-- +goose Up
-- +goose StatementBegin
ALTER TABLE feeds ADD COLUMN last_refreshed_at TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE feeds DROP COLUMN last_refreshed_at;
-- +goose StatementEnd
