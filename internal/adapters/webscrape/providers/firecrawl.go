// internal/adapters/webscrape/providers/firecrawl.go
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// FirecrawlScrapingProvider implements scraping via Firecrawl API
type FirecrawlScrapingProvider struct {
	BaseProvider
	apiKey string
	apiURL string
}

// NewFirecrawlScrapingProvider creates a new Firecrawl scraping provider with storage support
func NewFirecrawlScrapingProvider(httpClient *http.Client, storageClient storage.Client, logger *zap.Logger) *FirecrawlScrapingProvider {
	apiURL := os.Getenv("FIRECRAWL_API_URL")
	if apiURL == "" {
		apiURL = "https://api.firecrawl.dev/v1"
	}

	logger.Info("In NewFirecrawlScrapingProvider",
		zap.String("url", apiURL),
	)

	return &FirecrawlScrapingProvider{
		BaseProvider: BaseProvider{
			httpClient:    httpClient,
			storageClient: storageClient,
			logger:        logger.With(zap.String("provider", "firecrawl")),
		},
		apiKey: os.Getenv("FIRECRAWL_API_KEY"),
		apiURL: apiURL,
	}
}

func (f *FirecrawlScrapingProvider) Name() string {
	return "firecrawl"
}

func (f *FirecrawlScrapingProvider) IsAvailable() bool {
	return f.apiKey != ""
}

// Scrape performs single page scraping
func (f *FirecrawlScrapingProvider) Scrape(ctx context.Context, url string, config map[string]interface{}) (map[string]interface{}, error) {
	f.logger.Info("Starting scrape", zap.String("url", url))

	// Build scrape configuration
	formats := []string{"markdown", "html"}
	if formatList, ok := config["formats"].([]string); ok {
		formats = formatList
	}

	captureScreenshot := true
	if capture, ok := config["capture_screenshot"].(bool); ok {
		captureScreenshot = capture
	}

	onlyMainContent := false
	if mainContent, ok := config["only_main_content"].(bool); ok {
		onlyMainContent = mainContent
	}

	waitFor := 0
	if wait, ok := config["wait_for"].(int); ok {
		waitFor = wait
	}

	// Build request payload
	payload := map[string]interface{}{
		"url":             url,
		"formats":         formats,
		"onlyMainContent": onlyMainContent,
		"waitFor":         waitFor,
		"includeRawHtml":  true,
	}

	// Add screenshot config if needed
	if captureScreenshot {
		screenshotConfig := map[string]interface{}{
			"fullPage": true,
		}

		if viewport, ok := config["viewport"].(map[string]interface{}); ok {
			if width, ok := viewport["width"].(float64); ok {
				screenshotConfig["width"] = int(width)
			}
			if height, ok := viewport["height"].(float64); ok {
				screenshotConfig["height"] = int(height)
			}
		}

		payload["screenshot"] = true
		payload["screenshotConfig"] = screenshotConfig
	}

	f.logger.Info("In Firecrawl go Scrape",
		zap.Any("DEBUGaa: payload", payload),
	)

	// Make API request
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", f.apiURL+"/scrape", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var apiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	f.logger.Info("In Firecrawl go Scrape - response",
		zap.Any("DEBUGaa: firecrawl apiResponse", apiResponse),
	)

	if resp.StatusCode != http.StatusOK {
		if errorMsg, ok := apiResponse["error"].(string); ok {
			return nil, fmt.Errorf("API error: %s", errorMsg)
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Extract and format results
	result := map[string]interface{}{
		"url":         url,
		"captured_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Add available data from response
	if success, ok := apiResponse["success"].(bool); ok && success {
		if data, ok := apiResponse["data"].(map[string]interface{}); ok {
			if markdown, ok := data["markdown"].(string); ok {
				result["markdown_content"] = markdown
			}
			if html, ok := data["html"].(string); ok {
				result["html_content"] = html
			}
			if rawHtml, ok := data["rawHtml"].(string); ok {
				result["raw_html"] = rawHtml
			}
			if screenshot, ok := data["screenshot"].(string); ok {
				result["screenshot_url"] = screenshot
			}
			if metadata, ok := data["metadata"].(map[string]interface{}); ok {
				result["metadata"] = metadata
				if title, ok := metadata["title"].(string); ok {
					result["title"] = title
				}
				if description, ok := metadata["description"].(string); ok {
					result["description"] = description
				}
			}
			if links, ok := data["links"].([]interface{}); ok {
				result["links"] = links
			}
			if content, ok := data["content"].(string); ok {
				result["clean_content"] = content
			}
		}
	}

	f.logger.Info("Scrape completed successfully", zap.String("url", url))
	return result, nil
}

// Crawl performs multi-page crawling
func (f *FirecrawlScrapingProvider) Crawl(ctx context.Context, url string, config map[string]interface{}) (map[string]interface{}, error) {
	f.logger.Info("Starting crawl", zap.String("url", url))

	// Build crawl configuration
	limit := 10
	if l, ok := config["limit"].(float64); ok {
		limit = int(l)
	}

	maxDepth := 2
	if depth, ok := config["max_depth"].(float64); ok {
		maxDepth = int(depth)
	}

	// Build request payload
	payload := map[string]interface{}{
		"url":      url,
		"limit":    limit,
		"maxDepth": maxDepth,
	}

	// Add optional parameters
	if excludePaths, ok := config["exclude_paths"].([]interface{}); ok {
		payload["excludePaths"] = excludePaths
	}
	if includePaths, ok := config["include_paths"].([]interface{}); ok {
		payload["includePaths"] = includePaths
	}
	if allowBackward, ok := config["allow_backward_links"].(bool); ok {
		payload["allowBackwardLinks"] = allowBackward
	}

	formats := []string{"markdown", "html"}
	if formatList, ok := config["formats"].([]interface{}); ok {
		stringFormats := make([]string, len(formatList))
		for i, f := range formatList {
			if str, ok := f.(string); ok {
				stringFormats[i] = str
			}
		}
		formats = stringFormats
	}
	payload["formats"] = formats

	f.logger.Info("In Firecrawl go Crawl",
		zap.Any("DEBUGaa: payload", payload),
	)

	// Start crawl job
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", f.apiURL+"/crawl", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var crawlResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&crawlResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	f.logger.Info("In Firecrawl go Crawl - response",
		zap.Any("DEBUGaa: firecrawl crawlResponse", crawlResponse),
	)

	if resp.StatusCode != http.StatusOK {
		if errorMsg, ok := crawlResponse["error"].(string); ok {
			return nil, fmt.Errorf("API error: %s", errorMsg)
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Get job ID
	jobID, ok := crawlResponse["id"].(string)
	if !ok {
		return nil, fmt.Errorf("no job ID in response")
	}

	f.logger.Info("Crawl job started", zap.String("job_id", jobID))

	// Poll for completion
	result, err := f.pollCrawlJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get crawl results: %w", err)
	}

	return result, nil
}

// pollCrawlJob polls for crawl job completion
func (f *FirecrawlScrapingProvider) pollCrawlJob(ctx context.Context, jobID string) (map[string]interface{}, error) {
	maxAttempts := 60
	pollInterval := 5 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while polling")
		case <-time.After(pollInterval):
			// Check job status
			req, err := http.NewRequestWithContext(ctx, "GET", f.apiURL+"/crawl/"+jobID, nil)
			if err != nil {
				return nil, err
			}

			req.Header.Set("Authorization", "Bearer "+f.apiKey)

			resp, err := f.httpClient.Do(req)
			if err != nil {
				f.logger.Warn("Failed to check crawl status", zap.Error(err))
				continue
			}

			var statusResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			status, _ := statusResponse["status"].(string)
			f.logger.Debug("Crawl job status",
				zap.String("job_id", jobID),
				zap.String("status", status))

			switch status {
			case "completed":
				return map[string]interface{}{
					"job_id":       jobID,
					"pages":        statusResponse["data"],
					"total_pages":  statusResponse["total"],
					"completed_at": time.Now().UTC().Format(time.RFC3339),
					"status":       "completed",
				}, nil
			case "failed":
				errorMsg, _ := statusResponse["error"].(string)
				return nil, fmt.Errorf("crawl job failed: %s", errorMsg)
			}
		}
	}

	return nil, fmt.Errorf("crawl job timeout after %d attempts", maxAttempts)
}

// ExtractStructured extracts structured data using LLM
func (f *FirecrawlScrapingProvider) ExtractStructured(ctx context.Context, url string, schema map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	f.logger.Info("Starting structured extraction", zap.String("url", url))

	// Build extraction configuration
	systemPrompt := "Extract the requested information from the webpage content"
	if prompt, ok := config["system_prompt"].(string); ok {
		systemPrompt = prompt
	}

	userPrompt := ""
	if prompt, ok := config["prompt"].(string); ok {
		userPrompt = prompt
	}

	// Build request payload
	payload := map[string]interface{}{
		"url":     url,
		"formats": []string{"extract"},
		"extract": map[string]interface{}{
			"schema":       schema,
			"systemPrompt": systemPrompt,
			"prompt":       userPrompt,
		},
	}

	f.logger.Info("In Firecrawl go ExtractStructured",
		zap.Any("DEBUGaa: payload for extract structured", payload),
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", f.apiURL+"/scrape", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var apiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	f.logger.Info("In Firecrawl go ExtractStructured - response",
		zap.Any("DEBUGaa: firecrawl apiResponse", apiResponse),
	)

	if resp.StatusCode != http.StatusOK {
		if errorMsg, ok := apiResponse["error"].(string); ok {
			return nil, fmt.Errorf("API error: %s", errorMsg)
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	result := map[string]interface{}{
		"url":          url,
		"extracted_at": time.Now().UTC().Format(time.RFC3339),
	}

	if success, ok := apiResponse["success"].(bool); ok && success {
		if data, ok := apiResponse["data"].(map[string]interface{}); ok {
			if extracted, ok := data["extract"].(map[string]interface{}); ok {
				result["extracted_data"] = extracted
			}
			if metadata, ok := data["metadata"].(map[string]interface{}); ok {
				result["metadata"] = metadata
			}
		}
	}

	f.logger.Info("Extraction completed successfully", zap.String("url", url))
	return result, nil
}
