// rss_feeds.go
package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// Create a new HTTP request with the provided context, where the HTTP method
	// is "GET", URL is feedURL, and body is nil since we don't need to send any data
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	// Set a custom User-Agent header to identify the application making the request
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}    // Create a new HTTP client to send the request
	resp, err := client.Do(req) // Send the HTTP request and receive the response
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // Ensure that the response body is closed after we're done with it to free up resources

	body, err := io.ReadAll(resp.Body) // Read the entire response body into memory as a byte slice
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed                                   // Create a variable to hold the parsed RSS feed data
	if err := xml.Unmarshal(body, &rssFeed); err != nil { // Parse the XML data from the response body into the rssFeed variable
		return nil, err
	}

	// Unescape HTML entities in Channel
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	// Unescape HTML entities in Items
	for i := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}
	return &rssFeed, nil
}
