package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	defaultNewConnTimout = 5 * time.Second
)

func Init(ctx context.Context, dsn string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultNewConnTimout)
	defer cancel()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening DB: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("verifying DB connection: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = wal;",
		"PRAGMA synchronous = normal;",
	}

	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			return nil, fmt.Errorf("enabling %s: %w", p, err)
		}
	}

	if err := migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}
