// FILE: internal/adapters/websearch/providers/firecrawl.go
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
		apiURL:     "https://api.firecrawl.dev/v0/search",
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

	payload := map[string]interface{}{
		"query": query,
		"pageOptions": map[string]interface{}{
			"onlyMainContent":  true,
			"includeHtml":      false,
			"fetchPageContent": true,
		},
		"searchOptions": map[string]interface{}{
			"limit": numResults,
		},
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

	var apiResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
		Data    []struct {
			Title    string                 `json:"title"`
			URL      string                 `json:"url"`
			Content  string                 `json:"content"`
			Markdown string                 `json:"markdown,omitempty"`
			Provider string                 `json:"provider"`
			Metadata map[string]interface{} `json:"metadata,omitempty"`
			Score    float64                `json:"score,omitempty"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("search failed: %s", apiResponse.Message)
	}

	results := make([]SearchResult, 0, len(apiResponse.Data))
	for _, r := range apiResponse.Data {
		// Create snippet from content
		snippet := r.Content
		if snippet == "" && r.Markdown != "" {
			snippet = r.Markdown
		}
		snippet = strings.TrimSpace(snippet)
		if len(snippet) > 200 {
			snippet = snippet[:197] + "..."
		}

		// Extract published date if available
		publishedAt := ""
		if r.Metadata != nil {
			if date, ok := r.Metadata["publishedDate"].(string); ok {
				publishedAt = date
			}
		}

		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Snippet:     snippet,
			PublishedAt: publishedAt,
			Source:      f.Name(),
		})
	}

	f.logger.Info("Search completed",
		zap.Int("results_count", len(results)))

	return results, nil
}
