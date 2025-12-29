package domain

import "time"

// News represents a weather forecast news article
type News struct {
	Title       string
	PublishedAt time.Time
	Content     string
}

// NewsRepository defines the interface for fetching news data
type NewsRepository interface {
	FetchLatestNews() (*News, error)
}
