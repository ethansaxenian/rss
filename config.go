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

	dsn, ok := os.LookupEnv("DATABASE_DSN")
	if !ok {
		return config{}, fmt.Errorf("empty DATABASE_DSN")
	}

	c := config{
		dsn:  dsn,
		port: port,
	}

	return c, nil
}
