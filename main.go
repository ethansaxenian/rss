package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/ethansaxenian/rss/server"
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

	server, err := server.New(ctx, cfg.port, db)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close() //nolint:errcheck

	// go refreshLoop(ctx, db)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
