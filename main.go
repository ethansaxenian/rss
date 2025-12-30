package main

import (
	"context"
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

	db, err := database.Init(ctx, cfg.dsn)
	if err != nil {
		log.Fatalf("initializing db: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	w := worker.New(db, logger)
	go w.RunLoop(ctx)

	server := server.New(ctx, cfg.port, db, w, logger)
	defer server.Close() //nolint:errcheck

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
