package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrCannotParseDate = errors.New("failed to parse date")

type RSSTime struct {
	time.Time
}

func (t *RSSTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}

	v = strings.TrimSpace(v)

	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, v); err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("rss: cannot parse date %q", v)
}

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string  `xml:"title"`
	Link        string  `xml:"link"`
	Description string  `xml:"description"`
	PubDate     RSSTime `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, url string) (Channel, error) {
	client := http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Channel{}, fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Channel{}, fmt.Errorf("fetching feed url: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	var feed RSS
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return Channel{}, fmt.Errorf("decoding feed body: %w", err)
	}

	return feed.Channel, nil
}
