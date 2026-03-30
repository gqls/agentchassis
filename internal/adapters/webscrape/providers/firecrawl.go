// internal/adapters/webscrape/providers/firecrawl.go
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

	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// FirecrawlScrapingProvider implements scraping via Firecrawl API v2
// https://docs.firecrawl.dev/migrate-to-v2
type FirecrawlScrapingProvider struct {
	BaseProvider
	apiKey string
	apiURL string
}

// NewFirecrawlScrapingProvider creates a new Firecrawl v2 scraping provider with storage support
func NewFirecrawlScrapingProvider(httpClient *http.Client, storageClient storage.Client, logger *zap.Logger) *FirecrawlScrapingProvider {
	apiURL := os.Getenv("FIRECRAWL_API_URL")
	if apiURL == "" {
		apiURL = "https://api.firecrawl.dev/v2"
	}

	logger.Info("In NewFirecrawlScrapingProvider",
		zap.String("url", apiURL),
		zap.String("version", "v2"),
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

// Scrape performs single page scraping using Firecrawl API v2
func (f *FirecrawlScrapingProvider) Scrape(ctx context.Context, url string, config map[string]interface{}) (map[string]interface{}, error) {
	f.logger.Info("Starting scrape", zap.String("url", url))

	// Build formats array (v2 format — at top level for /scrape endpoint)
	formats := []interface{}{"markdown", "html", "rawHtml", "links"}

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

	// Add screenshot as object format in formats array (v2 style)
	if captureScreenshot {
		screenshotObj := map[string]interface{}{
			"type":     "screenshot",
			"fullPage": true,
		}

		// Add viewport if specified
		if viewport, ok := config["viewport"].(map[string]interface{}); ok {
			screenshotObj["viewport"] = viewport
		}

		formats = append(formats, screenshotObj)
	}

	// Build request payload (v2 format — /scrape uses top-level formats)
	payload := map[string]interface{}{
		"url":     url,
		"formats": formats,
	}

	// Add optional parameters
	if onlyMainContent {
		payload["onlyMainContent"] = onlyMainContent
	}

	if waitFor > 0 {
		payload["waitFor"] = waitFor
	}

	f.logger.Info("Firecrawl Scrape request",
		zap.String("url", url),
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

	//f.logger.Info("In Firecrawl go Scrape - response")// zap.Any("DEBUGaa: firecrawl apiResponse", apiResponse),

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
			// v2 screenshot handling - could be base64 or URL
			if screenshot, ok := data["screenshot"].(string); ok {
				// Check if it's base64 or URL
				if len(screenshot) > 0 {
					if screenshot[:4] == "http" {
						result["screenshot_url"] = screenshot
					} else {
						result["screenshot_base64"] = screenshot
					}
				}
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
			// Extract images from response (Firecrawl v2)
			if images, ok := data["images"].([]interface{}); ok && len(images) > 0 {
				result["images"] = images
			}

			if links, ok := data["links"].([]interface{}); ok {
				result["links"] = links

				// Extract image URLs from links
				imageLinks := []map[string]interface{}{}
				for _, link := range links {
					if linkStr, ok := link.(string); ok {
						// Simple string URL - check if it's an image
						lowerLink := strings.ToLower(linkStr)
						if strings.HasSuffix(lowerLink, ".jpg") ||
							strings.HasSuffix(lowerLink, ".jpeg") ||
							strings.HasSuffix(lowerLink, ".png") ||
							strings.HasSuffix(lowerLink, ".gif") ||
							strings.HasSuffix(lowerLink, ".webp") ||
							strings.HasSuffix(lowerLink, ".tif") ||
							strings.HasSuffix(lowerLink, ".tiff") ||
							strings.HasSuffix(lowerLink, ".bmp") ||
							strings.HasSuffix(lowerLink, ".svg") {
							imageLinks = append(imageLinks, map[string]interface{}{
								"url": linkStr,
							})
						}
					} else if linkMap, ok := link.(map[string]interface{}); ok {
						// Object with href/url
						var href string
						if h, ok := linkMap["href"].(string); ok {
							href = h
						} else if h, ok := linkMap["url"].(string); ok {
							href = h
						}

						if href != "" {
							lowerHref := strings.ToLower(href)
							if strings.HasSuffix(lowerHref, ".jpg") ||
								strings.HasSuffix(lowerHref, ".jpeg") ||
								strings.HasSuffix(lowerHref, ".png") ||
								strings.HasSuffix(lowerHref, ".gif") ||
								strings.HasSuffix(lowerHref, ".webp") ||
								strings.HasSuffix(lowerHref, ".tif") ||
								strings.HasSuffix(lowerHref, ".tiff") ||
								strings.HasSuffix(lowerHref, ".bmp") ||
								strings.HasSuffix(lowerHref, ".svg") {
								imgInfo := map[string]interface{}{
									"url": href,
								}
								if alt, ok := linkMap["alt"].(string); ok {
									imgInfo["alt"] = alt
								}
								if text, ok := linkMap["text"].(string); ok {
									imgInfo["text"] = text
								}
								imageLinks = append(imageLinks, imgInfo)
							}
						}
					}
				}

				if len(imageLinks) > 0 {
					result["images"] = imageLinks
				}
			}
			if content, ok := data["content"].(string); ok {
				result["clean_content"] = content
			}
		}
	}

	f.logger.Info("Scrape completed successfully", zap.String("url", url))
	return result, nil
}

// Crawl performs multi-page crawling using Firecrawl API v2
func (f *FirecrawlScrapingProvider) Crawl(ctx context.Context, url string, config map[string]interface{}) (map[string]interface{}, error) {
	f.logger.Info("Starting crawl", zap.String("url", url))

	// Build crawl configuration
	limit := 10
	if l, ok := config["limit"].(float64); ok {
		limit = int(l)
	}

	// v2 uses maxDiscoveryDepth instead of maxDepth
	maxDiscoveryDepth := 2
	if depth, ok := config["max_depth"].(float64); ok {
		maxDiscoveryDepth = int(depth)
	} else if depth, ok := config["max_discovery_depth"].(float64); ok {
		maxDiscoveryDepth = int(depth)
	}

	// Build request payload (v2 format)
	payload := map[string]interface{}{
		"url":               url,
		"limit":             limit,
		"maxDiscoveryDepth": maxDiscoveryDepth,
	}

	// Add optional parameters
	if excludePaths, ok := config["exclude_paths"].([]interface{}); ok {
		payload["excludePaths"] = excludePaths
	}
	if includePaths, ok := config["include_paths"].([]interface{}); ok {
		payload["includePaths"] = includePaths
	}

	// v2: use crawlEntireDomain instead of allowBackwardCrawling
	if crawlEntire, ok := config["crawl_entire_domain"].(bool); ok {
		payload["crawlEntireDomain"] = crawlEntire
	}

	// v2: sitemap handling (can be "only", "skip", or "include")
	if sitemap, ok := config["sitemap"].(string); ok {
		payload["sitemap"] = sitemap
	}

	// v2: optional prompt for smart crawling
	if prompt, ok := config["prompt"].(string); ok {
		payload["prompt"] = prompt
	}

	// v2: for /crawl, formats go inside scrapeOptions, NOT at the top level.
	// The /scrape endpoint uses top-level formats, but /crawl wraps them.
	scrapeOptions := map[string]interface{}{
		"formats": []string{"markdown"},
	}
	if formatList, ok := config["formats"].([]interface{}); ok {
		stringFormats := make([]string, 0, len(formatList))
		for _, f := range formatList {
			if str, ok := f.(string); ok {
				stringFormats = append(stringFormats, str)
			}
		}
		if len(stringFormats) > 0 {
			scrapeOptions["formats"] = stringFormats
		}
	}
	if onlyMain, ok := config["only_main_content"].(bool); ok {
		scrapeOptions["onlyMainContent"] = onlyMain
	}
	payload["scrapeOptions"] = scrapeOptions

	f.logger.Info("Firecrawl Crawl request",
		zap.String("url", url),
		zap.Int("limit", limit),
		zap.Int("maxDiscoveryDepth", maxDiscoveryDepth),
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

	/*f.logger.Info("In Firecrawl go Crawl - response",
		zap.Any("DEBUGaa: firecrawl crawlResponse", crawlResponse),
	)*/

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
			// Check job status (v2 endpoint)
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
			f.logger.Info("Crawl job status",
				zap.String("job_id", jobID),
				zap.String("status", status),
				zap.Int("attempt", attempt+1))

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

// ExtractStructured extracts structured data using LLM (v2 format)
func (f *FirecrawlScrapingProvider) ExtractStructured(ctx context.Context, url string, schema map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	f.logger.Info("Starting structured extraction", zap.String("url", url))

	// Build extraction configuration
	extractPrompt := "Extract the requested information from the webpage content"
	if prompt, ok := config["prompt"].(string); ok {
		extractPrompt = prompt
	} else if prompt, ok := config["system_prompt"].(string); ok {
		extractPrompt = prompt
	}

	// Build JSON extraction format (v2 style)
	jsonFormat := map[string]interface{}{
		"type":   "json",
		"prompt": extractPrompt,
		"schema": schema,
	}

	// Build request payload (v2 format)
	payload := map[string]interface{}{
		"url":     url,
		"formats": []interface{}{jsonFormat},
	}

	f.logger.Info("Firecrawl ExtractStructured request",
		zap.String("url", url),
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
			// v2: extracted JSON data is in the "json" field
			if extracted, ok := data["json"].(map[string]interface{}); ok {
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
