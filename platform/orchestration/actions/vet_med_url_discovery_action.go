// FILE: platform/orchestration/actions/med_url_discovery_action.go
//
// Discovers product URLs by scraping retailer category pages.
// For each retailer, scrapes category_urls from med_retailers, extracts
// product page links from the markdown, and inserts them into med_retailer_listings.
//
// Follows the same direct-Firecrawl-call pattern as med_price_scrape_action.go.
//
// Actions:
//   - med_discover_urls: Discover product URLs from retailer category pages

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// medRetailerForDiscovery holds a retailer row with its category URLs
type medRetailerForDiscovery struct {
	ID           string
	Name         string
	Domain       string
	BaseURL      string
	CategoryURLs []string
}

// MedDiscoverURLsAction scrapes retailer category pages and discovers product URLs.
func MedDiscoverURLsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("MedDiscoverURLsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	delayMs := 3000 // 3s between category page scrapes
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}

	// Optional: filter to a single retailer
	retailerFilter := ""
	if rf, ok := config["retailer_id"].(string); ok {
		retailerFilter = rf
	}

	// Override from input_data
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if d, ok := inputData["delay_ms"].(float64); ok && d > 0 {
			delayMs = int(d)
		}
		if rf, ok := inputData["retailer_id"].(string); ok && rf != "" {
			retailerFilter = rf
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

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	logCtx := chHTTPLogContext{
		DB:              params.DB,
		Logger:          params.Logger,
		AgentType:       params.AgentType,
		AgentID:         params.Headers["agent_id"],
		StepName:        params.ExecutionContext.StepName,
		OrchestrationID: params.ExecutionContext.OrchestrationID,
		CorrelationID:   params.ExecutionContext.CorrelationID,
		ActionName:      "med_discover_urls",
	}

	// Load retailers
	retailers, err := loadRetailersForDiscovery(ctx, params.DB, retailerFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to load retailers: %w", err)
	}

	if len(retailers) == 0 {
		params.Logger.Info("MedDiscoverURLs: no retailers to process")
		return map[string]interface{}{
			"status": "complete", "retailers_processed": 0, "urls_discovered": 0,
		}, nil
	}

	totalDiscovered := 0
	totalInserted := 0
	totalSkipped := 0

	for _, retailer := range retailers {
		if ctx.Err() != nil {
			break
		}

		for _, categoryURL := range retailer.CategoryURLs {
			if ctx.Err() != nil {
				break
			}

			params.Logger.Info("MedDiscoverURLs: scraping category page",
				zap.String("retailer", retailer.ID),
				zap.String("url", categoryURL))

			// Scrape the category page
			productLinks, err := scrapeCategoryPage(ctx, httpClient, firecrawlAPIKey, firecrawlAPIURL, categoryURL, retailer, logCtx)
			if err != nil {
				params.Logger.Warn("MedDiscoverURLs: category scrape failed",
					zap.String("retailer", retailer.ID),
					zap.String("url", categoryURL),
					zap.Error(err))
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
				continue
			}

			totalDiscovered += len(productLinks)

			// Insert discovered URLs
			inserted, skipped, err := insertDiscoveredListings(ctx, params.DB, retailer.ID, productLinks, params.Logger)
			if err != nil {
				params.Logger.Warn("MedDiscoverURLs: insert failed",
					zap.String("retailer", retailer.ID),
					zap.Error(err))
			}
			totalInserted += inserted
			totalSkipped += skipped

			params.Logger.Info("MedDiscoverURLs: category page processed",
				zap.String("retailer", retailer.ID),
				zap.String("url", categoryURL),
				zap.Int("links_found", len(productLinks)),
				zap.Int("inserted", inserted),
				zap.Int("skipped_existing", skipped))

			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	// Notify scheduler
	taskName := "med-discover-urls"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)

	params.Logger.Info("MedDiscoverURLs: complete",
		zap.Int("discovered", totalDiscovered),
		zap.Int("inserted", totalInserted),
		zap.Int("skipped", totalSkipped))

	return map[string]interface{}{
		"status":              "complete",
		"retailers_processed": len(retailers),
		"urls_discovered":     totalDiscovered,
		"urls_inserted":       totalInserted,
		"urls_skipped":        totalSkipped,
	}, nil
}

// loadRetailersForDiscovery loads active retailers with their category URLs.
func loadRetailersForDiscovery(ctx context.Context, db *sql.DB, retailerFilter string) ([]medRetailerForDiscovery, error) {
	query := `
		SELECT id, name, domain, base_url, category_urls
		FROM business_intel.med_retailers
		WHERE is_active = true
		  AND category_urls IS NOT NULL
		  AND array_length(category_urls, 1) > 0`

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
		var catURLs []byte // pg text[] comes as bytes
		if err := rows.Scan(&r.ID, &r.Name, &r.Domain, &r.BaseURL, &catURLs); err != nil {
			continue
		}
		// Parse Postgres text array using existing helper
		r.CategoryURLs = pgArrayToSlice(string(catURLs))
		if len(r.CategoryURLs) > 0 {
			retailers = append(retailers, r)
		}
	}
	return retailers, rows.Err()
}

// discoveredProduct holds a product URL and name found on a category page
type discoveredProduct struct {
	URL  string
	Name string
}

// Regex to extract markdown links: [text](url)
var markdownLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// scrapeCategoryPage scrapes a category page and extracts product links.
func scrapeCategoryPage(
	ctx context.Context,
	httpClient *http.Client,
	apiKey, apiURL, categoryURL string,
	retailer medRetailerForDiscovery,
	logCtx chHTTPLogContext,
) ([]discoveredProduct, error) {

	// Scrape with markdown + links formats
	payload := map[string]interface{}{
		"url":             categoryURL,
		"formats":         []string{"markdown"},
		"onlyMainContent": false, // need full page for product listings
		"waitFor":         3000,  // wait for JS to render product grid
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	scrapeURL := apiURL + "/scrape"
	req, err := http.NewRequestWithContext(ctx, "POST", scrapeURL, bytes.NewReader(payloadBytes))
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
		Method: "POST", URL: categoryURL,
		StatusCode: statusCode, LatencyMs: latencyMs,
		Success:      err == nil && statusCode == 200,
		ErrorMessage: errString(err),
		Metadata:     map[string]interface{}{"retailer_id": retailer.ID},
	})

	if err != nil {
		return nil, fmt.Errorf("firecrawl request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("firecrawl status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var fcResp map[string]interface{}
	if err := json.Unmarshal(body, &fcResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Extract markdown
	markdown := ""
	if data, ok := fcResp["data"].(map[string]interface{}); ok {
		if md, ok := data["markdown"].(string); ok {
			markdown = md
		}
	}

	if markdown == "" {
		return nil, fmt.Errorf("no markdown in response")
	}

	// Extract product links from markdown
	products := extractProductLinks(markdown, retailer)

	return products, nil
}

// extractProductLinks finds product page URLs from category page markdown.
// Uses markdown link syntax [Product Name](URL) and filters to same-domain product pages.
func extractProductLinks(markdown string, retailer medRetailerForDiscovery) []discoveredProduct {
	matches := markdownLinkPattern.FindAllStringSubmatch(markdown, -1)
	seen := map[string]bool{}
	var products []discoveredProduct

	for _, m := range matches {
		linkText := strings.TrimSpace(m[1])
		linkURL := strings.TrimSpace(m[2])

		// Strip markdown link title: [text](url "title") → url
		if idx := strings.Index(linkURL, " \""); idx > 0 {
			linkURL = linkURL[:idx]
		}
		if idx := strings.Index(linkURL, " '"); idx > 0 {
			linkURL = linkURL[:idx]
		}
		// Also handle space without quotes (shouldn't happen but defensive)
		linkURL = strings.TrimRight(linkURL, " ")

		// Resolve relative URLs
		if strings.HasPrefix(linkURL, "/") {
			linkURL = retailer.BaseURL + linkURL
		}

		// Must be same domain
		if !isRetailerProductURL(linkURL, retailer) {
			continue
		}

		// Skip if already seen
		if seen[linkURL] {
			continue
		}
		seen[linkURL] = true

		// Skip short/generic link text
		if len(linkText) < 5 {
			continue
		}

		// Skip navigation/UI link text
		lowerText := strings.ToLower(linkText)
		skipTexts := []string{
			"account", "sign in", "create an account", "login", "register",
			"chat status", "cookie", "skip to", "toggle menu", "menu",
			"subscribe", "view all", "explore", "read article", "read more",
			"error", "upload prescription", "who we are", "our expert",
			"refer a friend", "modern slavery", "tax strategy",
			"pet care hub", "knowledgebase", "promotion",
			"prescription info", "buying prescription",
			"covid", "free delivery",
		}
		skip := false
		for _, st := range skipTexts {
			if strings.Contains(lowerText, st) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Skip URLs with fragments (chat widgets, anchors)
		if strings.Contains(linkURL, "#") {
			continue
		}

		products = append(products, discoveredProduct{
			URL:  linkURL,
			Name: linkText,
		})
	}

	return products
}

// isRetailerProductURL checks if a URL is a product page on this retailer's domain.
// Filters out category pages, navigation, images, and non-product URLs.
func isRetailerProductURL(rawURL string, retailer medRetailerForDiscovery) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Must be same domain (with or without www)
	host := strings.TrimPrefix(parsed.Hostname(), "www.")
	retailerDomain := strings.TrimPrefix(retailer.Domain, "www.")
	if host != retailerDomain {
		return false
	}

	path := strings.ToLower(parsed.Path)

	// Skip root
	if path == "/" || path == "" {
		return false
	}

	// Skip fragment-only links
	if parsed.Path == "" && parsed.Fragment != "" {
		return false
	}

	// Skip non-product paths
	skipPrefixes := []string{
		"/cart", "/basket", "/checkout", "/login", "/account", "/customer/",
		"/delivery", "/returns", "/help", "/contact", "/about",
		"/terms", "/privacy", "/cookie", "/faq", "/blog", "/pet-advice",
		"/brand", "/brands", "/special-offer", "/prescription-info",
		"/prescriptions", "/how-do-i-", "/how-to-",
		"/media/", "/static/", "/js/", "/css/", "/icons/",
		"/search", "/wishlist", "/compare", "/review",
		"/offers", "/subscription", "/refer-", "/tax-", "/who-we-",
		"/modern-slavery", "/promotion-", "/vet-prescription",
		"/email-signup", "/covid", "/our-expert", "/pet-accessories",
		"/prescription-upload", "/pet-prescriptions",
		"/flea-tick-worm", "/natural-flea",
	}
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}

	// Skip image/asset URLs
	skipSuffixes := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg",
		".pdf", ".css", ".js", ".ico",
	}
	for _, suffix := range skipSuffixes {
		if strings.HasSuffix(path, suffix) {
			return false
		}
	}

	// Split path into segments: "/dog/healthcare/ageing" → ["dog", "healthcare", "ageing"]
	segments := []string{}
	for _, s := range strings.Split(strings.Trim(path, "/"), "/") {
		if s != "" {
			segments = append(segments, s)
		}
	}

	if len(segments) == 0 {
		return false
	}

	// Category prefixes — first segment is a species/department, not a product
	categoryFirstSegments := []string{
		"dog", "dogs", "cat", "cats", "horse", "horses", "equine",
		"small-pet", "small-pets", "small-furries", "small-animals",
		"pigeon", "farm", "human", "home", "bird", "birds",
		"current-offers", "special-offers",
	}
	if len(segments) >= 2 {
		for _, cat := range categoryFirstSegments {
			if segments[0] == cat {
				return false // e.g. /dog/healthcare/ageing — category page
			}
		}
	}

	// Multi-segment paths are usually categories unless they match known product patterns
	// VioVet: /Product-Name/c{number}/ — product family page
	// Hyperdrug: /Product-Name/productinfo/SKU/ — product page
	if len(segments) >= 2 {
		secondSeg := segments[1]
		// Allow VioVet family pages: /Name/c{digits}/
		if len(secondSeg) > 1 && secondSeg[0] == 'c' && isNumeric(secondSeg[1:]) {
			return true
		}
		// Allow Hyperdrug product pages: /Name/productinfo/SKU/
		if secondSeg == "productinfo" {
			return true
		}
		// Everything else with 2+ segments is likely a category
		return false
	}

	// Single-segment paths: likely products if long enough
	// Short single-segment paths could be top-level categories
	if len(segments) == 1 && len(segments[0]) < 10 {
		return false // e.g. /dog, /cat, /brands — too short
	}

	return true
}

// isNumeric checks if a string is all digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// insertDiscoveredListings inserts new product URLs into med_retailer_listings.
// Returns (inserted, skipped, error). Skipped = already existed.
func insertDiscoveredListings(ctx context.Context, db *sql.DB, retailerID string, products []discoveredProduct, logger *zap.Logger) (int, int, error) {
	inserted := 0
	skipped := 0

	for _, p := range products {
		result, err := db.ExecContext(ctx, `
			INSERT INTO business_intel.med_retailer_listings
				(retailer_id, retailer_url, retailer_product_name, match_method, is_active)
			VALUES ($1, $2, $3, 'url_discovery', true)
			ON CONFLICT (retailer_id, retailer_url) DO NOTHING`,
			retailerID,
			p.URL,
			p.Name,
		)
		if err != nil {
			logger.Warn("MedDiscoverURLs: insert failed",
				zap.String("url", p.URL),
				zap.Error(err))
			continue
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			inserted++
		} else {
			skipped++
		}
	}

	return inserted, skipped, nil
}
