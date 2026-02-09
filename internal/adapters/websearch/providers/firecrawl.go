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

func (f *FirecrawlProvider) Search(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	if numResults == 0 {
		numResults = 10
	}

	// v2 uses flat payload with "limit" at top level
	// Was: nested "pageOptions" and "searchOptions"
	payload := map[string]interface{}{
		"query": query,
		"limit": numResults,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	f.logger.Info("Executing search",
		zap.String("query", query),
		zap.Int("num_results", numResults))

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

	// Response struct matches v2 format
	// Was: Data []struct{...} (flat array)
	// Now: Data.Web []struct{...} (nested under "web" key)
	var apiResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
		Data    struct {
			Web []struct {
				URL         string `json:"url"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Position    int    `json:"position,omitempty"`
			} `json:"web"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("search failed: %s", apiResponse.Message)
	}

	// iterate Data.Web instead of Data
	results := make([]SearchResult, 0, len(apiResponse.Data.Web))
	for _, r := range apiResponse.Data.Web {
		snippet := strings.TrimSpace(r.Description)
		if len(snippet) > 200 {
			snippet = snippet[:197] + "..."
		}

		results = append(results, SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: snippet,
			Source:  f.Name(),
		})
	}

	f.logger.Info("Search completed",
		zap.Int("results_count", len(results)))

	return results, nil
}
