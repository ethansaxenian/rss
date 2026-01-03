package main

import (
	"fmt"
	"os"
	"strconv"
)

type config struct {
	dsn  string
	port int
}

func getConfig() (config, error) {
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		return config{}, fmt.Errorf("parsing SERVER_PORT: %w", err)
	}

	dsn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return config{}, fmt.Errorf("empty DATABASE_URL") //nolint: err113
	}

	c := config{
		dsn:  dsn,
		port: port,
	}

	return c, nil
}
