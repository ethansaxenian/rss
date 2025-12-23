package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}

type Item struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

func fetchFeed(url string) (Channel, error) {
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return Channel{}, fmt.Errorf("fetching feed url: %w", err)
	}
	log.Println(resp.Header)
	defer resp.Body.Close()

	var feed RSS
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return Channel{}, fmt.Errorf("decoding feed body: %w", err)
	}

	return feed.Channel, nil
}
