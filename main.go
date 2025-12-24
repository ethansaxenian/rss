package main

import (
	"context"
	"database/sql"
	"log"

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

	w := worker.New(db)
	go w.RunLoop(ctx)

	server, err := server.New(ctx, cfg.port, db, w)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close() //nolint:errcheck

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
