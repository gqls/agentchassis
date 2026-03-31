// FILE: platform/orchestration/actions/med_price_scrape_action.go
//
// Scrapes veterinary medicine prices from retailer product pages using Firecrawl.
// Follows the same direct-HTTP-call pattern as ch_fetch_accounts_action.go.
//
// Flow per listing:
//   1. Load batch of active listings from med_retailer_listings
//   2. For each listing, call Firecrawl /v1/scrape to get markdown
//   3. Parse markdown for price/variant data
//   4. Insert rows into med_price_snapshots
//   5. Update listing's last_scraped_at and last_price
//
// Requires FIRECRAWL_API_KEY env var on the business-intel pod.
//
// Actions:
//   - med_scrape_prices: Batch scrape prices for active listings

package actions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/storage"
)

// medPriceListing holds a listing row to scrape
type medPriceListing struct {
	ListingID   string
	RetailerID  string
	ProductID   sql.NullString
	RetailerURL string
	ProductName sql.NullString
}

// medExtractedVariant holds a price variant parsed from a product page
type medExtractedVariant struct {
	SizeVariant     string
	Price           float64
	TypicalVetPrice float64 // 0 if not found
	InStock         bool
}

// MedScrapePricesAction scrapes prices from retailer product pages.
func MedScrapePricesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("MedScrapePricesAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	batchSize := 20
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	delayMs := 2000 // 2s between scrapes — respectful rate
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}

	// Optional: filter to a single retailer for testing
	retailerFilter := ""
	if rf, ok := config["retailer_id"].(string); ok {
		retailerFilter = rf
	}

	// Override from input_data if provided (e.g. trigger message passes overrides)
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if bs, ok := inputData["batch_size"].(float64); ok && bs > 0 {
			batchSize = int(bs)
		}
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

	// Build HTTP logging context
	logCtx := chHTTPLogContext{
		DB:              params.DB,
		Logger:          params.Logger,
		AgentType:       params.AgentType,
		AgentID:         params.Headers["agent_id"],
		StepName:        params.ExecutionContext.StepName,
		OrchestrationID: params.ExecutionContext.OrchestrationID,
		CorrelationID:   params.ExecutionContext.CorrelationID,
		ActionName:      "med_scrape_prices",
	}

	// Load listings to scrape
	listings, err := loadMedListingsForScrape(ctx, params.DB, batchSize, retailerFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to load listings: %w", err)
	}

	if len(listings) == 0 {
		params.Logger.Info("MedScrapePrices: no listings to scrape")
		return map[string]interface{}{
			"status": "complete", "scraped": 0, "variants_stored": 0, "failed": 0,
		}, nil
	}

	params.Logger.Info("MedScrapePrices: loaded listings",
		zap.Int("count", len(listings)))

	totalScraped := 0
	totalVariants := 0
	totalFailed := 0

	for _, listing := range listings {
		if ctx.Err() != nil {
			params.Logger.Info("MedScrapePrices: context cancelled, stopping")
			break
		}

		variants, rawMarkdown, screenshotBytes, err := scrapeMedProductPage(ctx, httpClient, firecrawlAPIKey, firecrawlAPIURL, listing, logCtx)
		if err != nil {
			params.Logger.Warn("MedScrapePrices: scrape failed",
				zap.String("url", listing.RetailerURL),
				zap.String("retailer_id", listing.RetailerID),
				zap.Error(err))
			totalFailed++

			// Still update last_scraped_at so we don't hammer failed URLs
			_, _ = params.DB.ExecContext(ctx, `
				UPDATE business_intel.med_retailer_listings
				SET last_scraped_at = NOW(), updated_at = NOW()
				WHERE id = $1::uuid`, listing.ListingID)

			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			continue
		}

		// Upload screenshot from the same page load — uses chassis storage client
		screenshotURL := ""
		if len(screenshotBytes) > 0 && params.StorageClient != nil {
			ssURL, ssErr := uploadMedScreenshot(ctx, params.StorageClient, listing, screenshotBytes, params.Logger)
			if ssErr != nil {
				params.Logger.Warn("MedScrapePrices: screenshot upload failed (non-blocking)",
					zap.String("listing_id", listing.ListingID),
					zap.Error(ssErr))
			} else {
				screenshotURL = ssURL
			}
		}

		if len(variants) == 0 {
			params.Logger.Warn("MedScrapePrices: no variants found on page",
				zap.String("url", listing.RetailerURL),
				zap.String("retailer_id", listing.RetailerID))
			totalFailed++

			storeMedScrapeEvidence(ctx, params.DB, listing, rawMarkdown, screenshotURL, 0, 0, params.Logger)

			_, _ = params.DB.ExecContext(ctx, `
				UPDATE business_intel.med_retailer_listings
				SET last_scraped_at = NOW(), updated_at = NOW()
				WHERE id = $1::uuid`, listing.ListingID)

			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			continue
		}

		// Store price snapshots
		storedCount, err := storeMedPriceSnapshots(ctx, params.DB, listing, variants, rawMarkdown)
		if err != nil {
			params.Logger.Warn("MedScrapePrices: failed to store snapshots",
				zap.String("listing_id", listing.ListingID),
				zap.Error(err))
			totalFailed++
		} else {
			totalScraped++
			totalVariants += storedCount

			storeMedScrapeEvidence(ctx, params.DB, listing, rawMarkdown, screenshotURL, len(variants), storedCount, params.Logger)

			params.Logger.Info("MedScrapePrices: stored",
				zap.String("url", listing.RetailerURL),
				zap.Int("variants", storedCount),
				zap.Float64("first_price", variants[0].Price),
				zap.Bool("has_screenshot", screenshotURL != ""))
		}

		// Delay between requests
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	// Notify scheduler
	taskName := "med-scrape-prices"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)

	// Refresh materialized view if we stored anything
	if totalVariants > 0 {
		_, err := params.DB.ExecContext(ctx, `REFRESH MATERIALIZED VIEW CONCURRENTLY business_intel.med_price_current`)
		if err != nil {
			params.Logger.Warn("MedScrapePrices: failed to refresh materialized view", zap.Error(err))
		}
	}

	params.Logger.Info("MedScrapePrices: complete",
		zap.Int("scraped", totalScraped),
		zap.Int("variants_stored", totalVariants),
		zap.Int("failed", totalFailed))

	return map[string]interface{}{
		"status":          "complete",
		"scraped":         totalScraped,
		"variants_stored": totalVariants,
		"failed":          totalFailed,
	}, nil
}

// loadMedListingsForScrape loads active listings ordered by least recently scraped.
func loadMedListingsForScrape(ctx context.Context, db *sql.DB, limit int, retailerFilter string) ([]medPriceListing, error) {
	query := `
		SELECT l.id::text, l.retailer_id, l.product_id, l.retailer_url, l.retailer_product_name
		FROM business_intel.med_retailer_listings l
		JOIN business_intel.med_retailers r ON r.id = l.retailer_id
		WHERE l.is_active = true
		  AND r.is_active = true`

	args := []interface{}{}
	argIdx := 1

	if retailerFilter != "" {
		query += fmt.Sprintf(" AND l.retailer_id = $%d", argIdx)
		args = append(args, retailerFilter)
		argIdx++
	}

	query += fmt.Sprintf(`
		ORDER BY l.last_scraped_at ASC NULLS FIRST
		LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []medPriceListing
	for rows.Next() {
		var l medPriceListing
		if err := rows.Scan(&l.ListingID, &l.RetailerID, &l.ProductID, &l.RetailerURL, &l.ProductName); err != nil {
			continue
		}
		listings = append(listings, l)
	}
	return listings, rows.Err()
}

// scrapeMedProductPage calls Firecrawl scrape API and parses the markdown for prices.
// Returns: variants, markdown, screenshot PNG bytes (nil if unavailable), error.
func scrapeMedProductPage(
	ctx context.Context,
	httpClient *http.Client,
	apiKey, apiURL string,
	listing medPriceListing,
	logCtx chHTTPLogContext,
) ([]medExtractedVariant, string, []byte, error) {

	// Build Firecrawl scrape request
	// v1: screenshot as string "screenshot@fullPage"
	// v2: screenshot as object {"type": "screenshot", "fullPage": true}
	var formats interface{}
	isV2 := strings.Contains(apiURL, "/v2")
	if isV2 {
		formats = []interface{}{
			"markdown",
			map[string]interface{}{
				"type":     "screenshot",
				"fullPage": true,
			},
		}
	} else {
		formats = []string{"markdown", "screenshot@fullPage"}
	}

	payload := map[string]interface{}{
		"url":             listing.RetailerURL,
		"formats":         formats,
		"onlyMainContent": true,
		"waitFor":         2000,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", nil, fmt.Errorf("marshal payload: %w", err)
	}

	scrapeURL := apiURL + "/scrape"
	req, err := http.NewRequestWithContext(ctx, "POST", scrapeURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, "", nil, fmt.Errorf("create request: %w", err)
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
		Method: "POST", URL: listing.RetailerURL,
		StatusCode: statusCode, LatencyMs: latencyMs,
		Success:      err == nil && statusCode == 200,
		ErrorMessage: errString(err),
		Metadata: map[string]interface{}{
			"listing_id":  listing.ListingID,
			"retailer_id": listing.RetailerID,
		},
	})

	if err != nil {
		return nil, "", nil, fmt.Errorf("firecrawl request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, "", nil, fmt.Errorf("firecrawl status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var fcResp map[string]interface{}
	if err := json.Unmarshal(body, &fcResp); err != nil {
		return nil, "", nil, fmt.Errorf("parse firecrawl response: %w", err)
	}

	// Extract markdown and screenshot from same page load
	markdown := ""
	var screenshotBytes []byte
	if data, ok := fcResp["data"].(map[string]interface{}); ok {
		if md, ok := data["markdown"].(string); ok {
			markdown = md
		}

		// Log what keys Firecrawl returned so we can see if screenshot is present
		dataKeys := make([]string, 0, len(data))
		for k := range data {
			dataKeys = append(dataKeys, k)
		}
		logCtx.Logger.Info("MedScrapePrices: firecrawl response keys",
			zap.String("url", listing.RetailerURL),
			zap.Bool("is_v2", strings.Contains(apiURL, "/v2")),
			zap.String("api_url", apiURL),
			zap.Strings("data_keys", dataKeys))

		if ss, ok := data["screenshot"].(string); ok && ss != "" {
			logCtx.Logger.Info("MedScrapePrices: screenshot data received",
				zap.Int("length", len(ss)),
				zap.String("prefix", truncate(ss, 50)))

			b64Data := ss
			// Strip data URI prefix if present: "data:image/png;base64,..."
			if idx := strings.Index(b64Data, ","); idx >= 0 && idx < 100 {
				b64Data = b64Data[idx+1:]
			}
			// Try standard base64
			if decoded, err := base64.StdEncoding.DecodeString(b64Data); err == nil {
				screenshotBytes = decoded
			} else {
				// Try URL-safe base64
				if decoded, err := base64.RawStdEncoding.DecodeString(b64Data); err == nil {
					screenshotBytes = decoded
				} else {
					logCtx.Logger.Warn("MedScrapePrices: screenshot base64 decode failed",
						zap.String("url", listing.RetailerURL),
						zap.Error(err))
				}
			}
		} else {
			logCtx.Logger.Info("MedScrapePrices: no screenshot in response",
				zap.String("url", listing.RetailerURL))
		}
	}

	if markdown == "" {
		return nil, "", nil, fmt.Errorf("no markdown in firecrawl response")
	}

	inStock := !strings.Contains(strings.ToLower(markdown), "out of stock")
	variants := parseMedPriceVariants(markdown, inStock)

	return variants, markdown, screenshotBytes, nil
}

// Price parsing regex patterns.
// These handle the observed Pet Drugs Online format:
//   + 10ml bottle Price: £3.89 Regular Price: £14.09 (TVP) Save £10.20
// And simpler formats like:
//   Price: £17.48
//   £17.48
var (
	// Pattern for variant lines: "SIZE Price: £X.XX ... (TVP) ..."
	medVariantPattern = regexp.MustCompile(
		`(?i)(?:\+\s*)?([^\n£]+?)(?:\s+)?Price:\s*£([\d,]+\.?\d*)\s*(?:Regular Price:\s*£([\d,]+\.?\d*))?`)

	// Pattern for a single price when no variants exist
	medSinglePricePattern = regexp.MustCompile(
		`(?i)Price:\s*£([\d,]+\.?\d*)`)

	// Pattern for TVP where amount comes BEFORE label: "£50.54 (TVP)"
	medTVPBeforeLabelPattern = regexp.MustCompile(
		`£([\d,]+\.?\d*)\s*\((?:TVP|Typical Vet Price|RRP|SRP)\)`)

	// Pattern for TVP where label comes BEFORE amount: "TVP: £50.54" or "SRP £3.10"
	// Note: no .*? wildcard — label must be directly followed by optional colon then £
	medTVPAfterLabelPattern = regexp.MustCompile(
		`(?i)(?:TVP|Typical Vet Price|RRP|SRP)\s*:?\s*£([\d,]+\.?\d*)`)

	// Fallback: price+TVP jammed together: "£28.99£50.54 TVP" or "£28.99£50.54TVP"
	medPriceTVPJoinedPattern = regexp.MustCompile(
		`£([\d,]+\.?\d*)£([\d,]+\.?\d*)\s*(?:TVP|SRP)`)

	// Fallback: standalone £XX.XX on its own line (not preceded by Save/delivery/over/under)
	medStandalonePricePattern = regexp.MustCompile(
		`(?m)^£([\d,]+\.?\d*)\s*$`)
)

// parseMedPriceVariants extracts size/price variants from markdown.
// Only parses the product info section — truncates before Description/Related Products
// to avoid matching prices from other products on the same page.
func parseMedPriceVariants(markdown string, inStock bool) []medExtractedVariant {
	// Truncate to just the product section — everything before the description
	// or related products sections. This prevents matching prices from
	// "You May Also Like", "Related Products", etc.
	productSection := extractProductSection(markdown)

	var variants []medExtractedVariant

	// Try variant pattern first (multi-size products like Metacam)
	matches := medVariantPattern.FindAllStringSubmatch(productSection, -1)
	for _, m := range matches {
		sizeLabel := strings.TrimSpace(m[1])
		price := parsePoundValue(m[2])
		if price <= 0 {
			continue
		}

		// Skip if the "size" is just generic text
		sizeLabel = cleanSizeLabel(sizeLabel)
		if sizeLabel == "" {
			continue
		}

		var tvp float64
		if len(m) > 3 && m[3] != "" {
			tvp = parsePoundValue(m[3])
		}

		variants = append(variants, medExtractedVariant{
			SizeVariant:     sizeLabel,
			Price:           price,
			TypicalVetPrice: tvp,
			InStock:         inStock,
		})
	}

	// If no variants found, try single-price pattern (has "Price:" prefix)
	if len(variants) == 0 {
		singleMatch := medSinglePricePattern.FindStringSubmatch(productSection)
		if singleMatch != nil {
			price := parsePoundValue(singleMatch[1])
			if price > 0 {
				var tvp float64
				tvpMatch := medTVPBeforeLabelPattern.FindStringSubmatch(productSection)
				if tvpMatch != nil {
					tvp = parsePoundValue(tvpMatch[1])
				} else {
					tvpMatch = medTVPAfterLabelPattern.FindStringSubmatch(productSection)
					if tvpMatch != nil {
						tvp = parsePoundValue(tvpMatch[1])
					}
				}

				variants = append(variants, medExtractedVariant{
					SizeVariant:     "standard",
					Price:           price,
					TypicalVetPrice: tvp,
					InStock:         inStock,
				})
			}
		}
	}

	// Fallback: no "Price:" on page — try joined format "£28.99£50.54 TVP"
	// This handles Pet Drugs Online pages that show price+TVP jammed together
	if len(variants) == 0 {
		joinedMatch := medPriceTVPJoinedPattern.FindStringSubmatch(productSection)
		if joinedMatch != nil {
			price := parsePoundValue(joinedMatch[1])
			tvp := parsePoundValue(joinedMatch[2])
			if price > 0 {
				variants = append(variants, medExtractedVariant{
					SizeVariant:     "standard",
					Price:           price,
					TypicalVetPrice: tvp,
					InStock:         inStock,
				})
			}
		}
	}

	// Last resort: standalone £XX.XX on its own line
	if len(variants) == 0 {
		standaloneMatch := medStandalonePricePattern.FindStringSubmatch(productSection)
		if standaloneMatch != nil {
			price := parsePoundValue(standaloneMatch[1])
			if price > 0 {
				// Try to find TVP nearby
				var tvp float64
				tvpMatch := medTVPBeforeLabelPattern.FindStringSubmatch(productSection)
				if tvpMatch != nil {
					tvp = parsePoundValue(tvpMatch[1])
				}

				variants = append(variants, medExtractedVariant{
					SizeVariant:     "standard",
					Price:           price,
					TypicalVetPrice: tvp,
					InStock:         inStock,
				})
			}
		}
	}

	return variants
}

// extractProductSection returns only the product info portion of the markdown,
// truncating before Description headings, Related Products, and similar sections.
// This prevents the price regex from matching prices in "You May Also Like" etc.
func extractProductSection(markdown string) string {
	// Section markers that indicate the end of product info.
	// These must be line-start patterns to avoid matching inside URLs
	// (e.g. "#product-long-description" should NOT trigger a cutoff).
	lineMarkers := []string{
		"##### description",
		"#### description",
		"### description",
		"## description",
		"# description",
		"**description**",
	}

	// These can match anywhere — they're distinctive enough not to appear in URLs
	anywhereMarkers := []string{
		"related products",
		"you may also like",
		"similar products",
		"customers also bought",
		"recommended products",
		"recently viewed",
	}

	lower := strings.ToLower(markdown)
	cutoff := len(markdown)

	// Line-start markers: find "\n" + marker
	for _, marker := range lineMarkers {
		idx := strings.Index(lower, "\n"+marker)
		if idx >= 0 && idx < cutoff {
			cutoff = idx
		}
	}

	// Anywhere markers
	for _, marker := range anywhereMarkers {
		idx := strings.Index(lower, marker)
		if idx > 0 && idx < cutoff {
			cutoff = idx
		}
	}

	return markdown[:cutoff]
}

// parsePoundValue parses "17.48" or "1,234.56" to float64.
func parsePoundValue(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// cleanSizeLabel normalises the size variant label.
func cleanSizeLabel(s string) string {
	s = strings.TrimSpace(s)

	// Remove markdown checkbox markup: [ ] or [x]
	s = strings.Replace(s, "[ ]", "", 1)
	s = strings.Replace(s, "[x]", "", 1)
	s = strings.Replace(s, "[X]", "", 1)
	s = strings.TrimSpace(s)

	// Remove leading bullet characters and list markers
	s = strings.TrimLeft(s, "•·-*")
	s = strings.TrimSpace(s)

	// Skip if it's just noise
	lower := strings.ToLower(s)
	skipPhrases := []string{
		"size options", "select size", "choose", "options",
		"add to cart", "add to basket", "buy now",
		"description", "in stock", "out of stock",
	}
	for _, skip := range skipPhrases {
		if strings.HasPrefix(lower, skip) {
			return ""
		}
	}

	// If it's too long, probably not a real size label
	if len(s) > 80 {
		return ""
	}

	return s
}

// storeMedPriceSnapshots inserts price snapshot rows and updates the listing.
func storeMedPriceSnapshots(ctx context.Context, db *sql.DB, listing medPriceListing, variants []medExtractedVariant, rawMarkdown string) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stored := 0
	var firstPrice float64

	for _, v := range variants {
		// Build raw_data for debugging
		rawData := map[string]interface{}{
			"markdown_length": len(rawMarkdown),
		}
		rawJSON, _ := json.Marshal(rawData)

		productID := sql.NullString{}
		if listing.ProductID.Valid {
			productID = listing.ProductID
		}

		tvp := sql.NullFloat64{}
		if v.TypicalVetPrice > 0 {
			tvp = sql.NullFloat64{Float64: v.TypicalVetPrice, Valid: true}
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO business_intel.med_price_snapshots
				(listing_id, product_id, retailer_id, size_variant, price, in_stock, typical_vet_price, collection_method, raw_data)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, 'scrape', $8::jsonb)
			ON CONFLICT (listing_id, size_variant, collected_at) DO NOTHING`,
			listing.ListingID,
			productID,
			listing.RetailerID,
			v.SizeVariant,
			v.Price,
			v.InStock,
			tvp,
			string(rawJSON),
		)
		if err != nil {
			return stored, fmt.Errorf("insert snapshot for %s: %w", v.SizeVariant, err)
		}

		stored++
		if firstPrice == 0 {
			firstPrice = v.Price
		}
	}

	// Update listing with last scraped time and lowest price
	_, err = tx.ExecContext(ctx, `
		UPDATE business_intel.med_retailer_listings
		SET last_scraped_at = NOW(),
		    last_price = $2,
		    updated_at = NOW()
		WHERE id = $1::uuid`,
		listing.ListingID,
		firstPrice,
	)
	if err != nil {
		return stored, fmt.Errorf("update listing: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return stored, nil
}

// uploadMedScreenshot uploads a screenshot PNG via the chassis storage client.
// Path: med-evidence/{retailer_id}/{date}/{listing_id}.png
func uploadMedScreenshot(ctx context.Context, client storage.Client, listing medPriceListing, pngBytes []byte, logger *zap.Logger) (string, error) {
	dateStr := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("med-evidence/%s/%s/%s.png", listing.RetailerID, dateStr, listing.ListingID)

	uri, err := client.Upload(ctx, key, "image/png", bytes.NewReader(pngBytes))
	if err != nil {
		return "", fmt.Errorf("upload screenshot: %w", err)
	}

	logger.Info("MedScrapePrices: screenshot uploaded",
		zap.String("key", key),
		zap.Int("bytes", len(pngBytes)))

	return uri, nil
}

// storeMedScrapeEvidence stores the raw markdown and optional screenshot URI as evidence.
// Fire-and-forget — errors are logged but don't affect the scrape result.
func storeMedScrapeEvidence(ctx context.Context, db *sql.DB, listing medPriceListing, markdown, screenshotURL string, variantsFound, pricesStored int, logger *zap.Logger) {
	hash := sha256.Sum256([]byte(markdown))
	contentHash := hex.EncodeToString(hash[:])

	metadata := map[string]interface{}{}
	if screenshotURL != "" {
		metadata["screenshot_url"] = screenshotURL
	}
	metadataJSON, _ := json.Marshal(metadata)

	_, err := db.ExecContext(ctx, `
		INSERT INTO business_intel.med_scrape_evidence
			(listing_id, retailer_id, url, markdown_content, content_hash, variants_found, prices_stored, metadata)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::jsonb)`,
		listing.ListingID,
		listing.RetailerID,
		listing.RetailerURL,
		markdown,
		contentHash,
		variantsFound,
		pricesStored,
		string(metadataJSON),
	)
	if err != nil {
		logger.Warn("MedScrapePrices: failed to store evidence",
			zap.String("listing_id", listing.ListingID),
			zap.String("url", listing.RetailerURL),
			zap.String("error", err.Error()))
	} else {
		logger.Info("MedScrapePrices: evidence stored",
			zap.String("listing_id", listing.ListingID),
			zap.Int("markdown_bytes", len(markdown)),
			zap.Bool("has_screenshot", screenshotURL != ""),
			zap.Int("variants_found", variantsFound))
	}
}

// truncate and errString are defined in existing action files
// (fetch_agent_questionnaire.go and ch_fetch_accounts_action.go respectively)
// and are reused here via same package.
