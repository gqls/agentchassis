// FILE: internal/adapters/websearch/providers/provider.go
package providers

import "context"

// SearchProvider defines the interface for web search providers
type SearchProvider interface {
	Search(ctx context.Context, query string, numResults int) ([]SearchResult, error)
	Name() string
	IsAvailable() bool
}

// SearchResult represents a single search result
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	PublishedAt string `json:"published_at,omitempty"`
	Source      string `json:"source"` // Which provider returned this
}

// SearchOptions for advanced search features
type SearchOptions struct {
	SearchType string // web, news, images
	Language   string // en, es, fr, etc.
	Region     string // us, uk, etc.
	TimeRange  string // day, week, month, year
	SafeSearch bool
	Domains    []string // Restrict to specific domains
}
