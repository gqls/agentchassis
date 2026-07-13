// FILE: platform/orchestration/actions/ch_scrape_company_number_action.go
// Scrapes business websites for Companies House registration numbers.
// Fetches homepage HTML, searches the bottom portion (footer region) for
// company number patterns, and stores matches for CH lookup.
//
// Generic across verticals — every UK limited company has a registration number.
//
// Actions:
//   - ch_scrape_company_number: Batch scrape websites for company registration numbers

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Company number patterns found on UK websites.
// Order matters — more specific patterns first.
var companyNumberPatterns = []*regexp.Regexp{
	// "Company number: 12345678" / "Company no. 12345678" / "Company reg: 12345678"
	regexp.MustCompile(`(?i)company\s*(?:number|no\.?|reg(?:istration)?\.?)\s*[:.]?\s*(\d{7,8})`),
	// "Registered in England & Wales: 12345678" / "Registered in England no 12345678"
	regexp.MustCompile(`(?i)registered\s+(?:in\s+)?(?:england|wales|scotland|england\s*(?:&|and)\s*wales)[\s,]*(?:number|no\.?|reg\.?)?\s*[:.]?\s*(\d{7,8})`),
	// "Registered company 12345678"
	regexp.MustCompile(`(?i)registered\s+company\s*[:.]?\s*(\d{7,8})`),
	// "Registration number: 12345678"
	regexp.MustCompile(`(?i)registration\s+(?:number|no\.?)\s*[:.]?\s*(\d{7,8})`),
	// "Reg. No. 12345678" (standalone)
	regexp.MustCompile(`(?i)reg\.?\s*no\.?\s*[:.]?\s*(\d{7,8})`),
	// Scottish companies: "SC123456"
	regexp.MustCompile(`\b(SC\d{6})\b`),
	// Northern Irish companies: "NI123456"
	regexp.MustCompile(`\b(NI\d{6})\b`),
}

// CHScrapeCompanyNumberAction scrapes business websites for company registration numbers.
func CHScrapeCompanyNumberAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CHScrapeCompanyNumber: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	batchSize := 100
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	delayMs := 1000 // 1 second between requests — polite crawling
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}

	timeoutSec := 10 // per-request timeout
	if t, ok := config["request_timeout_sec"].(float64); ok && t > 0 {
		timeoutSec = int(t)
	}

	minConfidence := 0.40
	if mc, ok := config["min_confidence"].(float64); ok && mc > 0 {
		minConfidence = mc
	}

	verticalSlug := "veterinary"
	if vs, ok := config["vertical_slug"].(string); ok && vs != "" {
		verticalSlug = vs
	}

	// Load businesses to scrape: have website_url, not yet scraped, not low confidence
	businesses, err := loadBusinessesForScraping(ctx, params.DB, batchSize, minConfidence, verticalSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to load businesses: %w", err)
	}

	if len(businesses) == 0 {
		params.Logger.Info("CHScrapeCompanyNumber: no businesses to scrape")
		return map[string]interface{}{
			"status":  "complete",
			"scraped": 0, "found": 0, "matched": 0, "failed": 0,
		}, nil
	}

	params.Logger.Info("CHScrapeCompanyNumber: loaded businesses",
		zap.Int("count", len(businesses)))

	httpClient := &http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
		// Don't follow too many redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	totalScraped := 0
	totalFound := 0
	totalMatched := 0
	totalFailed := 0

	for _, biz := range businesses {
		select {
		case <-ctx.Done():
			params.Logger.Warn("CHScrapeCompanyNumber: context cancelled",
				zap.Int("scraped", totalScraped))
			break
		default:
		}

		// Rate limit
		if totalScraped > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		companyNumber, err := scrapeCompanyNumber(ctx, httpClient, biz.WebsiteURL)
		if err != nil {
			params.Logger.Warn("CHScrapeCompanyNumber: fetch failed",
				zap.String("business", biz.Name),
				zap.String("url", biz.WebsiteURL),
				zap.Error(err))
			totalFailed++
			// Mark as scraped even on failure so we don't retry every run
			_ = markCompanyNumberScraped(ctx, params.DB, biz.ID, "", false)
			totalScraped++
			continue
		}

		if companyNumber == "" {
			// No company number found — mark as scraped
			_ = markCompanyNumberScraped(ctx, params.DB, biz.ID, "", false)
			totalScraped++
			continue
		}

		// Found a company number — store it
		err = markCompanyNumberScraped(ctx, params.DB, biz.ID, companyNumber, true)
		if err != nil {
			params.Logger.Warn("CHScrapeCompanyNumber: failed to store",
				zap.String("business", biz.Name),
				zap.Error(err))
		}
		totalFound++

		// Try to match against ch_vet_companies
		matched, err := matchByCompanyNumber(ctx, params.DB, biz.ID, companyNumber)
		if err != nil {
			params.Logger.Warn("CHScrapeCompanyNumber: match failed",
				zap.String("business", biz.Name),
				zap.Error(err))
		} else if matched {
			totalMatched++
			params.Logger.Info("CHScrapeCompanyNumber: matched",
				zap.String("business", biz.Name),
				zap.String("company_number", companyNumber))
		} else {
			params.Logger.Info("CHScrapeCompanyNumber: found number but no CH match",
				zap.String("business", biz.Name),
				zap.String("company_number", companyNumber))
		}

		totalScraped++
	}

	// Notify scheduler
	taskName := "ch-scrape-company-number"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)

	params.Logger.Info("CHScrapeCompanyNumber: complete",
		zap.Int("scraped", totalScraped),
		zap.Int("found", totalFound),
		zap.Int("matched", totalMatched),
		zap.Int("failed", totalFailed))

	return map[string]interface{}{
		"status":  "complete",
		"scraped": totalScraped,
		"found":   totalFound,
		"matched": totalMatched,
		"failed":  totalFailed,
	}, nil
}

// scrapeBusiness holds what we need for scraping
type scrapeBusiness struct {
	ID         string
	Name       string
	WebsiteURL string
}

// loadBusinessesForScraping loads businesses with website URLs that haven't been
// scraped for company numbers yet. Uses the company_number_scraped column:
// NULL = not yet scraped, ” = scraped but not found, 'NNNNNNN' = found.
func loadBusinessesForScraping(ctx context.Context, db *sql.DB, limit int, minConfidence float64, verticalSlug string) ([]scrapeBusiness, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT b.id, b.name, b.website_url
		FROM business_intel.businesses b
		JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
		WHERE bv.slug = $1
		  AND b.verification_status = 'verified'
		  AND b.confidence_score >= $2
		  AND b.website_url IS NOT NULL AND b.website_url != ''
		  AND b.company_number_scraped IS NULL
		  AND b.business_type NOT ILIKE '%directory%'
		  AND b.business_type NOT ILIKE '%listing%'
		ORDER BY b.confidence_score DESC
		LIMIT $3`,
		verticalSlug, minConfidence, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []scrapeBusiness
	for rows.Next() {
		var biz scrapeBusiness
		if err := rows.Scan(&biz.ID, &biz.Name, &biz.WebsiteURL); err != nil {
			continue
		}
		businesses = append(businesses, biz)
	}
	return businesses, rows.Err()
}

// scrapeCompanyNumber fetches a website and extracts a company registration number.
// Reads the HTML body (limited to 500KB) and searches for patterns.
// Focuses on the bottom half of the content where footers live.
func scrapeCompanyNumber(ctx context.Context, client *http.Client, url string) (string, error) {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BusinessIntelBot/1.0)")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read limited body to avoid OOM on large pages
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024)) // 512KB max
	if err != nil {
		return "", err
	}

	html := string(body)

	// Focus on the bottom 40% of the HTML — company numbers are almost always in footers
	// But also check the full text in case it's in an "about" section higher up
	footerStart := len(html) * 60 / 100
	footerHTML := html[footerStart:]

	// Try footer first (more likely, fewer false positives)
	if num := extractCompanyNumber(footerHTML); num != "" {
		return num, nil
	}

	// Fall back to full page
	return extractCompanyNumber(html), nil
}

// extractCompanyNumber applies regex patterns to find a company registration number.
func extractCompanyNumber(html string) string {
	for _, pattern := range companyNumberPatterns {
		matches := pattern.FindStringSubmatch(html)
		if len(matches) >= 2 {
			num := strings.TrimSpace(matches[1])

			// For SC/NI prefixed numbers, return as-is
			if strings.HasPrefix(num, "SC") || strings.HasPrefix(num, "NI") {
				return num
			}

			// Validate: must be 7 or 8 digits
			if len(num) >= 7 && len(num) <= 8 {
				// Pad to 8 digits if 7
				if len(num) == 7 {
					num = "0" + num
				}
				return num
			}
		}
	}
	return ""
}

// markCompanyNumberScraped updates the business record with the scrape result.
// Empty string = scraped but not found. Non-empty = found.
func markCompanyNumberScraped(ctx context.Context, db *sql.DB, businessID, companyNumber string, found bool) error {
	if found {
		_, err := db.ExecContext(ctx, `
			UPDATE business_intel.businesses
			SET company_number_scraped = $1, updated_at = NOW()
			WHERE id = $2`,
			companyNumber, businessID)
		return err
	}
	// Mark as scraped but not found — use empty string to distinguish from NULL (not yet scraped)
	_, err := db.ExecContext(ctx, `
		UPDATE business_intel.businesses
		SET company_number_scraped = '', updated_at = NOW()
		WHERE id = $1`,
		businessID)
	return err
}

// matchByCompanyNumber directly matches a business to a CH company by company number.
// This is the highest confidence match possible — company numbers are unique.
func matchByCompanyNumber(ctx context.Context, db *sql.DB, businessID, companyNumber string) (bool, error) {
	result, err := db.ExecContext(ctx, `
		UPDATE business_intel.ch_vet_companies
		SET matched_business_id = $1,
			matched_at = NOW(),
			match_confidence = 1.0,
			match_method = 'company_number_scraped',
			updated_at = NOW()
		WHERE company_number = $2
		  AND company_status = 'active'`,
		businessID, companyNumber)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}
