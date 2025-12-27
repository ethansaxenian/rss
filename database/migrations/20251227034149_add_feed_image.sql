-- +goose Up
-- +goose StatementBegin
ALTER TABLE feeds ADD COLUMN image TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE feeds DROP COLUMN image;
-- +goose StatementEnd
