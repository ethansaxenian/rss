package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"

	"github.com/ethansaxenian/rss/database"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("GOOSE_DBSTRING")

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatal(err)
	}

	queries := database.New(db)

	// f, err := queries.CreateFeed(ctx, database.CreateFeedParams{Title: "Blog - Astral", URL: "https://astral.sh/blog/rss.xml"})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(f)

	feeds, err := queries.ListFeeds(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, feed := range feeds {
		res, err := fetchFeed(feed.URL)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(res)
	}
}
