package main

import (
	"context"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()

	refreshLoop(ctx)
}
