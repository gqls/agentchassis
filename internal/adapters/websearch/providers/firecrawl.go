// FILE: internal/adapters/websearch/providers/firecrawl.go
//
//   1. API URL: v0/search -> v2/search
//   2. Request payload: flat {query, limit} instead of nested pageOptions/searchOptions
//   3. Response struct: data.web[] array instead of flat data[]
//

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type FirecrawlProvider struct {
	apiKey     string
	httpClient *http.Client
	apiURL     string
	logger     *zap.Logger
}

func NewFirecrawlProvider(httpClient *http.Client, logger *zap.Logger) *FirecrawlProvider {
	return &FirecrawlProvider{
		apiKey:     os.Getenv("FIRECRAWL_API_KEY"),
		httpClient: httpClient,
		apiURL:     "https://api.firecrawl.dev/v2/search", // was v0/search
		logger:     logger.With(zap.String("provider", "firecrawl")),
	}
}

func (f *FirecrawlProvider) Name() string {
	return "firecrawl"
}

func (f *FirecrawlProvider) IsAvailable() bool {
	return f.apiKey != ""
}

func (f *FirecrawlProvider) Search(ctx context.Context, query string, numResults int, opts SearchOptions) ([]SearchResult, error) {
	if numResults == 0 {
		numResults = 10
	}

	// v2 uses flat payload with "limit" at top level
	// Was: nested "pageOptions" and "searchOptions"
	payload := map[string]interface{}{
		"query": query,
		"limit": numResults,
	}

	// v2 sources selects the result vertical; the default (absent) is web.
	// Images is a documented source but has no response parsing here, so
	// decline it rather than serve mislabelled results (bugs_open/127).
	switch opts.SearchType {
	case "", "web":
	case "news":
		payload["sources"] = []string{"news"}
	default:
		return nil, fmt.Errorf("firecrawl search_type %q not supported here: %w", opts.SearchType, ErrUnsupportedSearchType)
	}
	if opts.TimeRange != "" {
		if tbs, ok := googleTbsByTimeRange[opts.TimeRange]; ok {
			payload["tbs"] = tbs
		} else {
			f.logger.Warn("Unrecognised time_range, searching without a date filter",
				zap.String("time_range", opts.TimeRange))
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	f.logger.Info("Executing search",
		zap.String("query", query),
		zap.Int("num_results", numResults),
		zap.String("search_type", opts.SearchType),
		zap.String("time_range", opts.TimeRange))

	req, err := http.NewRequestWithContext(ctx, "POST", f.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	// Response struct matches v2 format: each requested source comes back
	// under its own key. Web items carry description; news items carry
	// snippet and a human-readable relative date ("3 months ago").
	type fcResult struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Snippet     string `json:"snippet"`
		Date        string `json:"date,omitempty"`
		Position    int    `json:"position,omitempty"`
	}
	var apiResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
		Data    struct {
			Web  []fcResult `json:"web"`
			News []fcResult `json:"news"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("search failed: %s", apiResponse.Message)
	}

	raw := append(apiResponse.Data.News, apiResponse.Data.Web...)
	now := time.Now()
	results := make([]SearchResult, 0, len(raw))
	for _, r := range raw {
		snippet := strings.TrimSpace(r.Description)
		if snippet == "" {
			snippet = strings.TrimSpace(r.Snippet)
		}
		if len(snippet) > 200 {
			snippet = snippet[:197] + "..."
		}

		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Snippet:     snippet,
			PublishedAt: normalisePublishedAt(r.Date, now),
			Source:      f.Name(),
		})
	}

	f.logger.Info("Search completed",
		zap.Int("results_count", len(results)))

	return results, nil
}
