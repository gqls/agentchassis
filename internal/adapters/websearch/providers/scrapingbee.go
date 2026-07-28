// FILE: internal/adapters/websearch/providers/scrapingbee.go
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.uber.org/zap"
)

type ScrapingBeeProvider struct {
	apiKey     string
	httpClient *http.Client
	apiURL     string
	logger     *zap.Logger
}

func NewScrapingBeeProvider(httpClient *http.Client, logger *zap.Logger) *ScrapingBeeProvider {
	return &ScrapingBeeProvider{
		apiKey:     os.Getenv("SCRAPING_BEE_API_KEY"),
		httpClient: httpClient,
		apiURL:     "https://app.scrapingbee.com/api/v1/store/google",
		logger:     logger.With(zap.String("provider", "scrapingbee")),
	}
}

func (s *ScrapingBeeProvider) Name() string {
	return "scrapingbee"
}

func (s *ScrapingBeeProvider) IsAvailable() bool {
	return s.apiKey != ""
}

func (s *ScrapingBeeProvider) Search(ctx context.Context, query string, numResults int, opts SearchOptions) ([]SearchResult, error) {
	if numResults == 0 {
		numResults = 10
	}

	// The Google store API supports web, news, maps and images via
	// search_type, but only news has response parsing here — decline the
	// rest rather than serving mislabelled web results (bugs_open/127).
	switch opts.SearchType {
	case "", "web", "news":
	default:
		return nil, fmt.Errorf("scrapingbee search_type %q not supported here: %w", opts.SearchType, ErrUnsupportedSearchType)
	}

	// Build search URL
	params := url.Values{}
	params.Add("api_key", s.apiKey)
	params.Add("search", query)
	params.Add("nb_results", fmt.Sprintf("%d", numResults))
	if opts.SearchType == "news" {
		params.Add("search_type", "news")
	}
	if opts.TimeRange != "" {
		// ScrapingBee forwards extra Google parameters verbatim, so tbs
		// applies Google's own recency filter.
		if tbs, ok := googleTbsByTimeRange[opts.TimeRange]; ok {
			params.Add("tbs", tbs)
		} else {
			s.logger.Warn("Unrecognised time_range, searching without a date filter",
				zap.String("time_range", opts.TimeRange))
		}
	}

	searchURL := fmt.Sprintf("%s?%s", s.apiURL, params.Encode())

	s.logger.Info("Executing search",
		zap.String("query", query),
		zap.Int("num_results", numResults),
		zap.String("search_type", opts.SearchType),
		zap.String("time_range", opts.TimeRange))

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response — news searches return news_results, web searches
	// organic_results; item shape is the same apart from description.
	type sbResult struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Snippet     string `json:"snippet"`
		Description string `json:"description"`
		Date        string `json:"date,omitempty"`
	}
	var apiResponse struct {
		OrganicResults []sbResult `json:"organic_results"`
		NewsResults    []sbResult `json:"news_results"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	raw := append(apiResponse.NewsResults, apiResponse.OrganicResults...)
	now := time.Now()
	results := make([]SearchResult, 0, len(raw))
	for _, r := range raw {
		snippet := r.Snippet
		if snippet == "" {
			snippet = r.Description
		}
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.Link,
			Snippet:     snippet,
			PublishedAt: normalisePublishedAt(r.Date, now),
			Source:      s.Name(),
		})
	}

	s.logger.Info("Search completed",
		zap.Int("results_count", len(results)))

	return results, nil
}
