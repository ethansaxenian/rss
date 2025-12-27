package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"time"

	"embed"

	"github.com/pressly/goose/v3"
)

const (
	migrationTimeout = 10 * time.Second
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	migrations, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("reading migrations dir: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectSQLite3,
		db,
		migrations,
	)
	if err != nil {
		return fmt.Errorf("initializing provider: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, migrationTimeout)
	defer cancel()

	if _, err = provider.Up(ctx); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil

}
