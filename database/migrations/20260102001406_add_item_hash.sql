-- +goose Up
-- +goose StatementBegin
ALTER TABLE items ADD COLUMN hash TEXT;

CREATE TABLE items_temp (
  id INTEGER PRIMARY KEY,
  feed_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  link TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT CHECK( status IN ('read', 'unread') ) NOT NULL DEFAULT 'unread',
  published_at TIMESTAMP NOT NULL,
  hash TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP,

  FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

INSERT INTO
  items_temp
SELECT
  id,
  feed_id,
  title,
  link,
  description,
  status,
  published_at,
  link, -- temporarily set hash to link, hope this is ok
  created_at,
  updated_at
FROM items;

DROP TABLE items;

ALTER TABLE items_temp RENAME TO items;

CREATE INDEX IF NOT EXISTS items_feed_id_ix ON items(feed_id);
CREATE INDEX IF NOT EXISTS items_published_at_ix ON items(published_at);
CREATE UNIQUE INDEX IF NOT EXISTS items_feed_id_hash ON items(feed_id, hash);
CREATE TRIGGER IF NOT EXISTS items_set_updated_at
AFTER UPDATE ON items
FOR EACH ROW
BEGIN
  UPDATE items
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS items_feed_id_hash;

ALTER TABLE items DROP COLUMN hash;
-- +goose StatementEnd
