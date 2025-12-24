package database

import (
	"os"
)

var DSN = os.Getenv("GOOSE_DBSTRING")
