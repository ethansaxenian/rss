-- +goose Up
-- +goose StatementBegin

-- temp table
CREATE TABLE IF NOT EXISTS items_temp (
  id INTEGER PRIMARY KEY,
  feed_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  link TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT CHECK( status IN ('read', 'unread') ) NOT NULL DEFAULT 'unread',
  published_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP,

  FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- copy rows
INSERT INTO items_temp SELECT * FROM items;

-- rename table
DROP TABLE items;
ALTER TABLE items_temp RENAME TO items;

-- recreate index
CREATE INDEX IF NOT EXISTS items_published_at_ix ON items(published_at);

-- recreate trigger
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
CREATE TABLE IF NOT EXISTS items_temp (
  id INTEGER PRIMARY KEY,
  feed_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  link TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT CHECK( status IN ('read', 'unread') ) NOT NULL DEFAULT 'unread',
  published_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP,

  FOREIGN KEY(feed_id) REFERENCES feeds(id)
);

-- copy rows
INSERT INTO items_temp SELECT * FROM items;

-- rename table
DROP TABLE items;
ALTER TABLE items_temp RENAME TO items;

-- recreate index
CREATE INDEX IF NOT EXISTS items_published_at_ix ON items(published_at);

-- recreate trigger
CREATE TRIGGER IF NOT EXISTS items_set_updated_at
AFTER UPDATE ON items
FOR EACH ROW
BEGIN
  UPDATE items
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = NEW.id;
END;

-- +goose StatementEnd
