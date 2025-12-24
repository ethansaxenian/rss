-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);


CREATE TRIGGER IF NOT EXISTS feeds_set_updated_at
AFTER UPDATE ON feeds
FOR EACH ROW
BEGIN
  UPDATE feeds
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE TRIGGER IF EXISTS feeds_set_updated_at;
DELETE TABLE IF EXISTS feeds;
-- +goose StatementEnd

