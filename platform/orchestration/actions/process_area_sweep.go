// FILE: platform/orchestration/actions/process_area_sweep.go
//
// ProcessAreaSweepAction processes web search results from a geographic area
// sweep, matches them against existing businesses, and inserts unmatched
// results as discovery candidates. Updates search_areas tracking table.
//
// This action shares helper functions with scan_discovery_candidates.go
// (same package): extractDomain, extractRootURL, extractPracticeName,
// extractUKPostcode, skipDomains, vetKeywords, nullIfEmpty.
// Do NOT redefine them here.
//
// Workflow step config:
//
//   "process_results": {
//       "action": "process_area_sweep",
//       "config": {
//           "input_fields": ["district_code", "area_name", "search_area_id", "search_results"]
//       },
//       "next_step": "complete",
//       "description": "Check results against known businesses",
//       "output_field": "sweep_result"
//   }

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func ProcessAreaSweepAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ProcessAreaSweepAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	// -- Extract inputs from collected_data --

	districtCode := getStringFromCollected(params.CollectedData, "district_code")
	areaName := getStringFromCollected(params.CollectedData, "area_name")
	searchAreaID := getStringFromCollected(params.CollectedData, "search_area_id")

	params.Logger.Info("ProcessAreaSweepAction: processing",
		zap.String("district_code", districtCode),
		zap.String("area_name", areaName))

	// -- Find search results --
	var results []interface{}
	var searchQuery string

	if sr, ok := params.CollectedData["search_results"].(map[string]interface{}); ok {
		// search_results may have response.results or just results
		if resp, ok := sr["response"].(map[string]interface{}); ok {
			if r, ok := resp["results"].([]interface{}); ok {
				results = r
			}
			searchQuery, _ = resp["query"].(string)
		}
		if results == nil {
			if r, ok := sr["results"].([]interface{}); ok {
				results = r
			}
			if searchQuery == "" {
				searchQuery, _ = sr["query"].(string)
			}
		}
	}

	if len(results) == 0 {
		params.Logger.Info("ProcessAreaSweepAction: no search results")
		updateSweepTracking(ctx, params.DB, searchAreaID, districtCode, 0, 0, params.Logger)
		return map[string]interface{}{
			"district_code": districtCode,
			"scanned":       0,
			"candidates":    0,
			"already_known": 0,
			"skipped":       0,
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

		// For area sweeps, we're less strict about vet keyword matching
		// because the search query itself is vet-specific. But still
		// filter out obviously non-vet results.
		combined := strings.ToLower(resultTitle + " " + resultSnippet)
		isDefinitelyNotVet := false
		nonVetIndicators := []string{
			"pet shop", "pet store", "dog grooming", "cat cafe",
			"pet food", "dog walking", "pet sitting",
			"wikipedia", "news article",
		}
		for _, indicator := range nonVetIndicators {
			if strings.Contains(combined, indicator) && !strings.Contains(combined, "veterinary") && !strings.Contains(combined, "vet ") {
				isDefinitelyNotVet = true
				break
			}
		}
		if isDefinitelyNotVet {
			skipped++
			continue
		}

		// Check if we already have this website in businesses table
		rootURL := extractRootURL(resultURL)
		var existingID sql.NullString
		err := params.DB.QueryRowContext(ctx,
			`SELECT id FROM business_intel.businesses 
			 WHERE website_url ILIKE $1 OR website_url ILIKE $2
			 LIMIT 1`,
			rootURL+"%", "www."+strings.TrimPrefix(rootURL, "https://")+"%",
		).Scan(&existingID)

		if err == nil && existingID.Valid {
			alreadyKnown++
			continue
		}

		// Also check if this URL is already a discovery candidate
		var existingCandidateID sql.NullString
		err = params.DB.QueryRowContext(ctx,
			`SELECT id FROM business_intel.discovery_candidates 
			 WHERE source_url = $1
			 LIMIT 1`,
			resultURL,
		).Scan(&existingCandidateID)

		if err == nil && existingCandidateID.Valid {
			// Already tracked as candidate, skip
			alreadyKnown++
			continue
		}

		// Extract practice name and postcode from result
		candidateName := extractPracticeName(resultTitle)
		postcode := extractUKPostcode(resultSnippet)

		// If no postcode found in snippet, use the district code as approximation
		if postcode == "" && districtCode != "" {
			postcode = districtCode
		}

		// Insert as discovery candidate
		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO business_intel.discovery_candidates 
				(name, website_url, address_snippet, postcode,
				 source_query, source_url, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'pending', NOW())
			ON CONFLICT (source_url) DO UPDATE SET
				name = COALESCE(NULLIF(EXCLUDED.name, ''), business_intel.discovery_candidates.name),
				postcode = COALESCE(NULLIF(EXCLUDED.postcode, ''), business_intel.discovery_candidates.postcode),
				updated_at = NOW()`,
			candidateName, rootURL, resultSnippet, postcode,
			searchQuery, resultURL,
		)
		if err != nil {
			params.Logger.Warn("ProcessAreaSweepAction: failed to insert candidate",
				zap.String("url", resultURL),
				zap.Error(err))
			continue
		}
		candidates++

		params.Logger.Info("ProcessAreaSweepAction: found candidate",
			zap.String("name", candidateName),
			zap.String("url", rootURL),
			zap.String("district", districtCode))
	}

	// Update sweep tracking
	updateSweepTracking(ctx, params.DB, searchAreaID, districtCode, len(results), candidates, params.Logger)

	params.Logger.Info("ProcessAreaSweepAction: complete",
		zap.String("district_code", districtCode),
		zap.Int("scanned", scanned),
		zap.Int("candidates", candidates),
		zap.Int("skipped", skipped),
		zap.Int("already_known", alreadyKnown))

	return map[string]interface{}{
		"district_code": districtCode,
		"area_name":     areaName,
		"scanned":       scanned,
		"candidates":    candidates,
		"already_known": alreadyKnown,
		"skipped":       skipped,
	}, nil
}

// updateSweepTracking updates the search_areas table after a sweep
func updateSweepTracking(ctx context.Context, db *sql.DB, searchAreaID, districtCode string, resultCount, candidatesFound int, logger *zap.Logger) {
	var err error

	if searchAreaID != "" {
		// Update by ID if provided
		_, err = db.ExecContext(ctx, `
			UPDATE business_intel.search_areas 
			SET last_swept_at = NOW(),
			    sweep_count = sweep_count + 1,
			    last_result_count = $1,
			    candidates_found = candidates_found + $2
			WHERE id = $3`,
			resultCount, candidatesFound, searchAreaID)
	} else if districtCode != "" {
		// Fall back to district_code lookup
		_, err = db.ExecContext(ctx, `
			UPDATE business_intel.search_areas 
			SET last_swept_at = NOW(),
			    sweep_count = sweep_count + 1,
			    last_result_count = $1,
			    candidates_found = candidates_found + $2
			WHERE district_code = $3 AND country = 'GB'`,
			resultCount, candidatesFound, districtCode)
	}

	if err != nil {
		logger.Warn("ProcessAreaSweepAction: failed to update sweep tracking",
			zap.String("district_code", districtCode),
			zap.Error(err))
	}
}

// getStringFromCollected extracts a string from collected_data, checking
// top-level, input_data, and __raw_message__.input_data
func getStringFromCollected(collected map[string]interface{}, key string) string {
	// Direct top-level
	if v, ok := collected[key].(string); ok && v != "" {
		return v
	}
	// input_data
	if inputData, ok := collected["input_data"].(map[string]interface{}); ok {
		if v, ok := inputData[key].(string); ok && v != "" {
			return v
		}
	}
	// __raw_message__
	if raw, ok := collected["__raw_message__"].(map[string]interface{}); ok {
		if inputData, ok := raw["input_data"].(map[string]interface{}); ok {
			if v, ok := inputData[key].(string); ok && v != "" {
				return v
			}
		}
	}
	return ""
}
