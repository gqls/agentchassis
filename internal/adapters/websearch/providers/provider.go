// FILE: internal/adapters/websearch/providers/provider.go
package providers

import (
	"context"
	"errors"
)

// SearchProvider defines the interface for web search providers.
// Every Search receives the caller's SearchOptions; a provider that cannot
// honour opts.SearchType must return ErrUnsupportedSearchType rather than
// serving a different kind of search under the requested label
// (bugs_open/127: search_type was typed, logged, and dropped at this hop,
// so every "news" feed in the fleet was a plain web search).
type SearchProvider interface {
	Search(ctx context.Context, query string, numResults int, opts SearchOptions) ([]SearchResult, error)
	Name() string
	IsAvailable() bool
}

// ErrUnsupportedSearchType is returned by a provider that cannot perform the
// requested search type. The adapter treats it as non-retryable and falls
// through to the next provider; if every provider declines, the request fails
// loudly instead of returning mislabelled results.
var ErrUnsupportedSearchType = errors.New("unsupported search type")

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

// googleTbsByTimeRange maps SearchOptions.TimeRange onto Google's tbs
// query-string values, used by the Google-backed providers (ScrapingBee
// forwards extra Google parameters verbatim; Firecrawl v2 accepts tbs
// directly).
var googleTbsByTimeRange = map[string]string{
	"day":   "qdr:d",
	"week":  "qdr:w",
	"month": "qdr:m",
	"year":  "qdr:y",
}
