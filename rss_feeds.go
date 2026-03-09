// rss_feeds.go
package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Razimuth/gator/internal/database"
	"github.com/google/uuid"
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

func scrapeFeeds(s *state) {
	// Get next feed
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("Error fetching next feed: %v\n", err)
		return
	}

	// Mark as fetched
	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("Error marking feed as fetched: %v\n", err)
		return
	}

	// Fetch feed
	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Error parsing feed %s: %v\n", feed.Url, err)
		return
	}

	// Create posts
	for _, item := range rssFeed.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			FeedID:    feed.ID,
			Title:     item.Title,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			fmt.Printf("Couldn't create post: %v", err)
			continue
		}
		fmt.Printf(" - %s\n", item.Title)
	}
	fmt.Printf("Fetched: %s - %d posts\n", feed.Name, len(rssFeed.Channel.Item))
}
