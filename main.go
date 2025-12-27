package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/ethansaxenian/rss/database"
	"github.com/ethansaxenian/rss/server"
	"github.com/ethansaxenian/rss/worker"
	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()

	cfg, err := getConfig()
	if err != nil {
		log.Fatalf("getting config: %v", err)
	}

	db, err := sql.Open("sqlite", cfg.dsn)
	if err != nil {
		log.Fatalf("opening db: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Debug("Running migrations")
	if err := database.Migrate(ctx, db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	w := worker.New(db, logger)
	go w.RunLoop(ctx)

	server := server.New(ctx, cfg.port, db, w, logger)
	defer server.Close() //nolint:errcheck

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
