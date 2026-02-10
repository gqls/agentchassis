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
// ALSO defines shared data structures and helpers used by process_area_sweep.go:
//   skipDomains, blockedDomainSuffixes, knownGroups, vetKeywords
//   extractDomain, extractRootURL, extractPracticeName,
//   extractUKPostcode, isBlockedDomain, detectGroup
//
// Workflow config:
//
//	"scan_discoveries": {
//	    "action": "scan_discovery_candidates",
//	    "config": {
//	        "input_fields": ["business_id", "search_results", "search_practice"]
//	    },
//	    "next_step": "complete",
//	    "description": "Scan search results for unknown vet practices",
//	    "output_field": "discovery_scan"
//	}

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Shared data structures (used by both scan_discovery_candidates and
// process_area_sweep)
// ---------------------------------------------------------------------------

// skipDomains - generic directories, social media, aggregators.
// NOT national vet chains — those go in knownGroups so we capture their
// individual practice pages.
var skipDomains = map[string]bool{
	// Search engines / maps
	"google.com":      true,
	"google.co.uk":    true,
	"maps.google.com": true,
	"mapquest.com":    true,
	"wheree.com":      true,
	// Social media
	"facebook.com":   true,
	"twitter.com":    true,
	"x.com":          true,
	"instagram.com":  true,
	"linkedin.com":   true,
	"youtube.com":    true,
	"tiktok.com":     true,
	"nextdoor.co.uk": true,
	"pinterest.com":  true,
	// Generic directories
	"yell.com":           true,
	"yelp.com":           true,
	"yelp.co.uk":         true,
	"thomsonlocal.com":   true,
	"192.com":            true,
	"scoot.co.uk":        true,
	"hotfrog.co.uk":      true,
	"cylex-uk.co.uk":     true,
	"brownbook.net":      true,
	"freeindex.co.uk":    true,
	"localmole.co.uk":    true,
	"lacartes.com":       true,
	"misterwhat.co.uk":   true,
	"opendi.co.uk":       true,
	"fyple.co.uk":        true,
	"tuugo.co.uk":        true,
	"bark.com":           true,
	"checkatrade.com":    true,
	"britaine.co.uk":     true,
	"places-near-me.com": true,
	// Review sites
	"tripadvisor.com": true,
	"trustpilot.com":  true,
	// Reference / government
	"wikipedia.org":    true,
	"en.wikipedia.org": true,
	"nhs.uk":           true,
	"gov.uk":           true,
	// Professional bodies (main domains — subdomains caught by suffix matching)
	"bva.co.uk": true,
	// Retail / irrelevant
	"amazon.co.uk":    true,
	"ebay.co.uk":      true,
	"gumtree.com":     true,
	"jobs.nhs.uk":     true,
	"indeed.co.uk":    true,
	"glassdoor.co.uk": true,
	"reed.co.uk":      true,
	"totaljobs.com":   true,
}

// blockedDomainSuffixes catches all subdomains of blocked domains.
// Checked via suffix matching so "find-a-vet.rcvs.org.uk" is caught.
var blockedDomainSuffixes = []string{
	".rcvs.org.uk",   // all RCVS subdomains
	".gov.uk",        // all government subdomains
	".wikipedia.org", // all wikipedia subdomains
}

// knownGroups maps domain patterns to group names.
// We want to CAPTURE these results (they're real practices) but tag them
// with their group affiliation.
var knownGroups = map[string]string{
	"cvsvets.com":               "CVS Group",
	"cvsukltd.co.uk":            "CVS Group",
	"ivcpractices.com":          "IVC Evidensia",
	"ivcevidensia.com":          "IVC Evidensia",
	"evidensia.co.uk":           "IVC Evidensia",
	"medivet.co.uk":             "Medivet",
	"vets4pets.com":             "Vets4Pets",
	"vets-now.com":              "Vets Now",
	"linnaeus.group":            "Linnaeus",
	"linnaeusgroup.co.uk":       "Linnaeus",
	"vetpartners.co.uk":         "VetPartners",
	"myfamilyvets.co.uk":        "My Family Vets",
	"goddardvetgroup.co.uk":     "Goddard",
	"eastcottvets.co.uk":        "Eastcott",
	"whitehorsevetcentre.co.uk": "White Horse",
}

// vetKeywords suggests a search result is about a vet practice
var vetKeywords = []string{
	"veterinary", "vets", "vet practice", "vet surgery",
	"vet clinic", "vet hospital", "animal hospital",
	"pet care", "vet centre", "vet center",
}

// ---------------------------------------------------------------------------
// ScanDiscoveryCandidatesAction
// ---------------------------------------------------------------------------

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

		// Skip blocked domains (directories, social, RCVS subdomains)
		domain := extractDomain(resultURL)
		if isBlockedDomain(domain) {
			skipped++
			continue
		}

		// Skip if this is the source business's own website
		if sourceWebsite != "" && sameWebsite(resultURL, sourceWebsite) {
			skipped++
			continue
		}

		// Check if this looks like a vet practice
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
		var existingID string
		err := params.DB.QueryRowContext(ctx,
			`SELECT id FROM business_intel.businesses
			 WHERE website_url ILIKE $1 OR website_url ILIKE $2
			 LIMIT 1`,
			rootURL+"%", "www."+strings.TrimPrefix(rootURL, "https://")+"%",
		).Scan(&existingID)

		if err != nil && err != sql.ErrNoRows {
			// Real DB error — log and skip this result
			params.Logger.Warn("ScanDiscoveryCandidatesAction: DB error checking businesses",
				zap.String("url", resultURL), zap.Error(err))
			skipped++
			continue
		}
		if err == nil {
			// Found in businesses table
			alreadyKnown++
			continue
		}
		// err == sql.ErrNoRows — not in businesses, continue to insert

		// Detect group affiliation
		detectedGroup, isGroup := detectGroup(domain)
		isIndependent := !isGroup

		candidateName := extractPracticeName(resultTitle)
		postcode := extractUKPostcode(resultSnippet)

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO business_intel.discovery_candidates
				(name, website_url, address_snippet, postcode,
				 source_business_id, source_query, source_url,
				 detected_group, is_independent,
				 status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,
				'pending', NOW(), NOW())
			ON CONFLICT (source_url) DO UPDATE SET
				name = COALESCE(NULLIF(EXCLUDED.name, ''), business_intel.discovery_candidates.name),
				postcode = COALESCE(NULLIF(EXCLUDED.postcode, ''), business_intel.discovery_candidates.postcode),
				detected_group = COALESCE(EXCLUDED.detected_group, business_intel.discovery_candidates.detected_group),
				is_independent = COALESCE(EXCLUDED.is_independent, business_intel.discovery_candidates.is_independent),
				updated_at = NOW()`,
			candidateName, rootURL, resultSnippet, postcode,
			nullIfEmpty(businessID), query, resultURL,
			nullIfEmpty(detectedGroup), nullIfFalse(isIndependent),
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
			zap.String("postcode", postcode),
			zap.String("group", detectedGroup))
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
