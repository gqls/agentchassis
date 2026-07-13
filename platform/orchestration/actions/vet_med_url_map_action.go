// FILE: platform/orchestration/actions/vet_med_url_map_action.go
//
// Discovers product URLs using Firecrawl's /map endpoint for site-wide URL discovery.
// Unlike med_discover_urls (which scrapes specific category pages and follows pagination),
// this uses Firecrawl's built-in crawling and sitemap parsing to find all product URLs
// across an entire retailer site in a single API call.
//
// Use this for broad discovery; use med_discover_urls for targeted category-specific discovery.
//
// Actions:
//   - med_map_urls: Discover product URLs via Firecrawl /map endpoint

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// MedMapURLsAction discovers product URLs using Firecrawl's /map endpoint.
// This is the broad-discovery alternative to MedDiscoverURLsAction.
func MedMapURLsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("MedMapURLsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	// Merge input_data
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		for k, v := range inputData {
			config[k] = v
		}
	}

	firecrawlAPIKey := os.Getenv("FIRECRAWL_API_KEY")
	if firecrawlAPIKey == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_KEY not set")
	}
	firecrawlAPIURL := os.Getenv("FIRECRAWL_API_URL")
	if firecrawlAPIURL == "" {
		firecrawlAPIURL = "https://api.firecrawl.dev/v2"
	}

	// Optional retailer filter
	retailerFilter := ""
	if rf, ok := config["retailer_id"].(string); ok {
		retailerFilter = rf
	}

	// Optional search term to focus map results
	searchTerm := ""
	if st, ok := config["search"].(string); ok {
		searchTerm = st
	}

	// URL limit per retailer (default 5000, max allowed by Firecrawl)
	urlLimit := 5000
	if ul, ok := config["url_limit"].(float64); ok && ul > 0 {
		urlLimit = int(ul)
	}

	// Load retailers — uses the same loader but doesn't require category_urls
	retailers, err := loadRetailersForMap(ctx, params.DB, retailerFilter)
	if err != nil {
		return nil, fmt.Errorf("load retailers: %w", err)
	}

	if len(retailers) == 0 {
		params.Logger.Info("MedMapURLsAction: no active retailers found")
		return map[string]interface{}{"status": "no_retailers"}, nil
	}

	httpClient := &http.Client{Timeout: 120 * time.Second} // map can take a while

	logCtx := chHTTPLogContext{
		DB:              params.DB,
		Logger:          params.Logger,
		AgentType:       params.AgentType,
		AgentID:         params.Headers["agent_id"],
		StepName:        params.ExecutionContext.StepName,
		OrchestrationID: params.ExecutionContext.OrchestrationID,
		CorrelationID:   params.ExecutionContext.CorrelationID,
		ActionName:      "med_map_urls",
	}

	totalDiscovered := 0
	totalInserted := 0
	totalSkipped := 0
	totalFiltered := 0

	for _, retailer := range retailers {
		if ctx.Err() != nil {
			break
		}

		params.Logger.Info("MedMapURLsAction: mapping retailer site",
			zap.String("retailer", retailer.ID),
			zap.String("domain", retailer.Domain),
			zap.String("search", searchTerm),
			zap.Int("limit", urlLimit))

		links, err := firecrawlMapSite(ctx, httpClient, firecrawlAPIKey, firecrawlAPIURL, retailer, searchTerm, urlLimit, logCtx)
		if err != nil {
			params.Logger.Warn("MedMapURLsAction: map failed",
				zap.String("retailer", retailer.ID),
				zap.Error(err))
			continue
		}

		params.Logger.Info("MedMapURLsAction: map returned links",
			zap.String("retailer", retailer.ID),
			zap.Int("total_links", len(links)))

		// Filter to product URLs using the same deny-list as category-based discovery
		var products []discoveredProduct
		for _, link := range links {
			if isRetailerProductURL(link.URL, retailer) {
				products = append(products, discoveredProduct{
					URL:  link.URL,
					Name: link.Title,
				})
			} else {
				totalFiltered++
			}
		}

		totalDiscovered += len(products)

		if len(products) == 0 {
			params.Logger.Info("MedMapURLsAction: no product URLs after filtering",
				zap.String("retailer", retailer.ID),
				zap.Int("filtered_out", totalFiltered))
			continue
		}

		// Insert discovered URLs — reuses the same function as category-based discovery
		inserted, skipped, err := insertDiscoveredListings(ctx, params.DB, retailer.ID, products, params.Logger)
		if err != nil {
			params.Logger.Warn("MedMapURLsAction: insert failed",
				zap.String("retailer", retailer.ID),
				zap.Error(err))
		}
		totalInserted += inserted
		totalSkipped += skipped

		params.Logger.Info("MedMapURLsAction: retailer processed",
			zap.String("retailer", retailer.ID),
			zap.Int("map_links", len(links)),
			zap.Int("product_urls", len(products)),
			zap.Int("inserted", inserted),
			zap.Int("skipped_existing", skipped))
	}

	// Notify scheduler
	taskName := "med-map-urls"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)

	params.Logger.Info("MedMapURLsAction: complete",
		zap.Int("discovered", totalDiscovered),
		zap.Int("inserted", totalInserted),
		zap.Int("skipped", totalSkipped),
		zap.Int("filtered", totalFiltered))

	return map[string]interface{}{
		"status":     "complete",
		"discovered": totalDiscovered,
		"inserted":   totalInserted,
		"skipped":    totalSkipped,
		"filtered":   totalFiltered,
	}, nil
}

// mapLink represents a URL returned by Firecrawl's /map endpoint
type mapLink struct {
	URL         string
	Title       string
	Description string
}

// firecrawlMapSite calls the Firecrawl /map endpoint to discover all URLs on a site.
func firecrawlMapSite(
	ctx context.Context,
	httpClient *http.Client,
	apiKey, apiURL string,
	retailer medRetailerForDiscovery,
	searchTerm string,
	limit int,
	logCtx chHTTPLogContext,
) ([]mapLink, error) {

	payload := map[string]interface{}{
		"url":                   retailer.BaseURL,
		"limit":                 limit,
		"includeSubdomains":     false,
		"ignoreQueryParameters": true,
	}

	if searchTerm != "" {
		payload["search"] = searchTerm
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	mapURL := apiURL + "/map"
	req, err := http.NewRequestWithContext(ctx, "POST", mapURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	callStart := time.Now()
	resp, err := httpClient.Do(req)
	latencyMs := int(time.Since(callStart).Milliseconds())

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	LogHTTPRequest(logCtx.DB, logCtx.Logger, HTTPRequestLogParams{
		AgentType: logCtx.AgentType, AgentID: logCtx.AgentID,
		StepName: logCtx.StepName, OrchestrationID: logCtx.OrchestrationID,
		CorrelationID: logCtx.CorrelationID, ActionName: logCtx.ActionName,
		Method: "POST", URL: retailer.BaseURL,
		StatusCode: statusCode, LatencyMs: latencyMs,
		Success:      err == nil && statusCode == 200,
		ErrorMessage: errString(err),
		Metadata: map[string]interface{}{
			"retailer_id": retailer.ID,
			"endpoint":    "map",
			"search":      searchTerm,
			"limit":       limit,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("firecrawl map request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("firecrawl map status %d: %s", resp.StatusCode, truncate(string(body), 300))
	}

	// Parse response: {"success": true, "links": [{"url": "...", "title": "...", "description": "..."}]}
	var mapResp struct {
		Success bool `json:"success"`
		Links   []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"links"`
	}

	if err := json.Unmarshal(body, &mapResp); err != nil {
		return nil, fmt.Errorf("parse map response: %w", err)
	}

	if !mapResp.Success {
		return nil, fmt.Errorf("firecrawl map returned success=false")
	}

	var links []mapLink
	for _, l := range mapResp.Links {
		if l.URL != "" {
			links = append(links, mapLink{
				URL:         l.URL,
				Title:       l.Title,
				Description: l.Description,
			})
		}
	}

	return links, nil
}

// loadRetailersForMap loads active retailers — does not require category_urls.
func loadRetailersForMap(ctx context.Context, db *sql.DB, retailerFilter string) ([]medRetailerForDiscovery, error) {
	query := `
		SELECT id, name, domain, base_url, COALESCE(category_urls, '{}')
		FROM business_intel.med_retailers
		WHERE is_active = true`

	args := []interface{}{}
	if retailerFilter != "" {
		query += " AND id = $1"
		args = append(args, retailerFilter)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var retailers []medRetailerForDiscovery
	for rows.Next() {
		var r medRetailerForDiscovery
		var catURLs []byte
		if err := rows.Scan(&r.ID, &r.Name, &r.Domain, &r.BaseURL, &catURLs); err != nil {
			continue
		}
		r.CategoryURLs = pgArrayToSlice(string(catURLs))
		retailers = append(retailers, r)
	}
	return retailers, rows.Err()
}
