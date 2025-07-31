// FILE: internal/adapters/websearch/providers/scrapingbee.go
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"os"
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

func (s *ScrapingBeeProvider) Search(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	if numResults == 0 {
		numResults = 10
	}

	// Build search URL
	params := url.Values{}
	params.Add("api_key", s.apiKey)
	params.Add("search", query)
	params.Add("nb_results", fmt.Sprintf("%d", numResults))

	searchURL := fmt.Sprintf("%s?%s", s.apiURL, params.Encode())

	s.logger.Debug("Executing search",
		zap.String("query", query),
		zap.Int("num_results", numResults))

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

	// Parse response
	var apiResponse struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			Date    string `json:"date,omitempty"`
		} `json:"organic_results"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert to our format
	results := make([]SearchResult, 0, len(apiResponse.OrganicResults))
	for _, r := range apiResponse.OrganicResults {
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.Link,
			Snippet:     r.Snippet,
			PublishedAt: r.Date,
			Source:      s.Name(),
		})
	}

	s.logger.Info("Search completed",
		zap.Int("results_count", len(results)))

	return results, nil
}
