// FILE: platform/orchestration/actions/scan_discovery_candidates.go
//
// ScanDiscoveryCandidatesAction examines cached search results for the current
// business and checks each result against existing businesses.
// Unmatched results that look like vet practices get inserted into
// business_intel.discovery_candidates for later review.
//
// Runs as a local action at the end of the vet-practice-verifier workflow,
// after store_results and before complete.
//
// Workflow config:
//
//   "scan_discoveries": {
//       "action": "scan_discovery_candidates",
//       "config": {
//           "input_fields": ["business_id", "search_results", "search_practice"]
//       },
//       "next_step": "complete",
//       "description": "Scan search results for unknown vet practices",
//       "output_field": "discovery_scan"
//   }

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// Domains that are directories/aggregators, not individual practices
var skipDomains = map[string]bool{
	"google.com":             true,
	"google.co.uk":           true,
	"facebook.com":           true,
	"twitter.com":            true,
	"instagram.com":          true,
	"linkedin.com":           true,
	"youtube.com":            true,
	"yell.com":               true,
	"yelp.com":               true,
	"yelp.co.uk":             true,
	"tripadvisor.com":        true,
	"trustpilot.com":         true,
	"rcvs.org.uk":            true,
	"find-a-vet.rcvs.org.uk": true,
	"wikipedia.org":          true,
	"en.wikipedia.org":       true,
	"nhs.uk":                 true,
	"gov.uk":                 true,
	"bva.co.uk":              true,
	"vets-now.com":           true, // national chain, not individual
	"vets4pets.com":          true, // national chain directory
	"medivet.co.uk":          true, // national chain directory
	"cvsvets.com":            true, // group directory
	"ivcpractices.com":       true, // group directory
	"amazon.co.uk":           true,
	"ebay.co.uk":             true,
}

// Keywords that suggest a search result is about a vet practice
var vetKeywords = []string{
	"veterinary", "vets", "vet practice", "vet surgery",
	"vet clinic", "vet hospital", "animal hospital",
	"pet care", "vet centre", "vet center",
}

func ScanDiscoveryCandidatesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ScanDiscoveryCandidatesAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	// Get the business_id so we can record the source
	businessID := ""
	if id, ok := params.CollectedData["business_id"].(string); ok {
		businessID = id
	}
	if businessID == "" {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			businessID, _ = inputData["business_id"].(string)
		}
	}
	if businessID == "" {
		if raw, ok := params.CollectedData["__raw_message__"].(map[string]interface{}); ok {
			if inputData, ok := raw["input_data"].(map[string]interface{}); ok {
				businessID, _ = inputData["business_id"].(string)
			}
		}
	}

	// Get the source business's website URL to exclude self-matches
	sourceWebsite := ""
	if br, ok := params.CollectedData["business_record"].(map[string]interface{}); ok {
		if biz, ok := br["business"].(map[string]interface{}); ok {
			sourceWebsite, _ = biz["website_url"].(string)
		}
	}

	// Find search results in collected data
	var results []interface{}
	var query string

	for _, key := range []string{"search_practice", "search_results"} {
		stepData, ok := params.CollectedData[key].(map[string]interface{})
		if !ok {
			continue
		}
		resp, ok := stepData["response"].(map[string]interface{})
		if !ok {
			continue
		}
		if r, ok := resp["results"].([]interface{}); ok && len(r) > 0 {
			results = r
			query, _ = resp["query"].(string)
			break
		}
	}

	if len(results) == 0 {
		params.Logger.Info("ScanDiscoveryCandidatesAction: no search results to scan")
		return map[string]interface{}{
			"scanned":    0,
			"candidates": 0,
			"skipped":    0,
		}, nil
	}

	scanned := 0
	candidates := 0
	skipped := 0
	alreadyKnown := 0

	for _, r := range results {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		scanned++

		resultURL, _ := result["url"].(string)
		resultTitle, _ := result["title"].(string)
		resultSnippet, _ := result["snippet"].(string)
		if resultSnippet == "" {
			resultSnippet, _ = result["description"].(string)
		}

		if resultURL == "" {
			skipped++
			continue
		}

		// Skip directory/aggregator sites
		domain := extractDomain(resultURL)
		if skipDomains[domain] {
			skipped++
			continue
		}

		// Skip if this is the source business's own website
		if sourceWebsite != "" && sameWebsite(resultURL, sourceWebsite) {
			skipped++
			continue
		}

		// Check if this looks like a vet practice (by title or snippet)
		combined := strings.ToLower(resultTitle + " " + resultSnippet)
		looksLikeVet := false
		for _, kw := range vetKeywords {
			if strings.Contains(combined, kw) {
				looksLikeVet = true
				break
			}
		}
		if !looksLikeVet {
			skipped++
			continue
		}

		// Check if we already have this website in businesses
		rootURL := extractRootURL(resultURL)
		var existingID sql.NullString
		err := params.DB.QueryRowContext(ctx,
			`SELECT id FROM business_intel.businesses 
			 WHERE website_url ILIKE $1 OR website_url ILIKE $2
			 LIMIT 1`,
			rootURL+"%", "www."+rootURL+"%",
		).Scan(&existingID)

		if err == nil && existingID.Valid {
			alreadyKnown++
			continue
		}

		// Extract a name from the title (heuristic: take text before " - " or " | ")
		candidateName := extractPracticeName(resultTitle)

		// Try to extract postcode from snippet (UK postcode pattern)
		postcode := extractUKPostcode(resultSnippet)

		// Insert as discovery candidate
		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO business_intel.discovery_candidates 
				(name, website_url, address_snippet, postcode,
				 source_business_id, source_query, source_url, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', NOW())
			ON CONFLICT (source_url) DO UPDATE SET
				name = COALESCE(NULLIF(EXCLUDED.name, ''), business_intel.discovery_candidates.name),
				postcode = COALESCE(NULLIF(EXCLUDED.postcode, ''), business_intel.discovery_candidates.postcode),
				updated_at = NOW()`,
			candidateName, rootURL, resultSnippet, postcode,
			nullIfEmpty(businessID), query, resultURL,
		)
		if err != nil {
			params.Logger.Warn("ScanDiscoveryCandidatesAction: failed to insert candidate",
				zap.String("url", resultURL),
				zap.Error(err))
			continue
		}
		candidates++

		params.Logger.Info("ScanDiscoveryCandidatesAction: found candidate",
			zap.String("name", candidateName),
			zap.String("url", rootURL),
			zap.String("postcode", postcode))
	}

	params.Logger.Info("ScanDiscoveryCandidatesAction: complete",
		zap.Int("scanned", scanned),
		zap.Int("candidates", candidates),
		zap.Int("skipped", skipped),
		zap.Int("already_known", alreadyKnown))

	return map[string]interface{}{
		"scanned":       scanned,
		"candidates":    candidates,
		"skipped":       skipped,
		"already_known": alreadyKnown,
	}, nil
}

// extractDomain returns the root domain from a URL (e.g. "www.example.co.uk" -> "example.co.uk")
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host
}

// extractRootURL returns the scheme + host (without path) and without www
func extractRootURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	host := strings.TrimPrefix(parsed.Hostname(), "www.")
	return fmt.Sprintf("https://%s", host)
}

// sameWebsite checks if two URLs point to the same website
func sameWebsite(url1, url2 string) bool {
	d1 := extractDomain(url1)
	d2 := extractDomain(url2)
	return d1 != "" && d1 == d2
}

// extractPracticeName tries to pull a practice name from a search result title
// Titles tend to be "Practice Name - Location | Something" or "Practice Name | Vets in ..."
func extractPracticeName(title string) string {
	title = strings.TrimSpace(title)
	// Split on common separators
	for _, sep := range []string{" | ", " - ", " – ", " — "} {
		parts := strings.SplitN(title, sep, 2)
		if len(parts) > 0 && len(parts[0]) > 3 {
			return strings.TrimSpace(parts[0])
		}
	}
	return title
}

// extractUKPostcode finds a UK postcode pattern in text
// Pattern: 1-2 letters, 1-2 digits, optional space, digit, 2 letters
func extractUKPostcode(text string) string {
	text = strings.ToUpper(text)
	// Simple scan - look for postcode-shaped sequences
	words := strings.Fields(text)
	for i := 0; i < len(words)-1; i++ {
		candidate := words[i] + " " + words[i+1]
		if looksLikeUKPostcode(candidate) {
			return candidate
		}
		// Also try without space
		if looksLikeUKPostcode(words[i]) {
			// Insert space before last 3 chars
			if len(words[i]) >= 5 {
				pc := words[i][:len(words[i])-3] + " " + words[i][len(words[i])-3:]
				return pc
			}
		}
	}
	return ""
}

func looksLikeUKPostcode(s string) bool {
	s = strings.TrimSpace(s)
	// Must be 6-8 chars (with space)
	if len(s) < 6 || len(s) > 8 {
		return false
	}
	// Must contain a space
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return false
	}
	// Second part must be digit + 2 letters
	if len(parts[1]) != 3 {
		return false
	}
	if parts[1][0] < '0' || parts[1][0] > '9' {
		return false
	}
	if parts[1][1] < 'A' || parts[1][1] > 'Z' {
		return false
	}
	if parts[1][2] < 'A' || parts[1][2] > 'Z' {
		return false
	}
	// First part: 2-4 chars, starts with letter
	if len(parts[0]) < 2 || len(parts[0]) > 4 {
		return false
	}
	if parts[0][0] < 'A' || parts[0][0] > 'Z' {
		return false
	}
	return true
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
