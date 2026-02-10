// FILE: platform/orchestration/actions/process_area_sweep.go
//
// ProcessAreaSweepAction processes web search results from a single district
// sweep. Checks each result against known businesses and existing candidates,
// inserts new ones into discovery_candidates, updates search_areas tracking.
//
// Uses shared helpers from scan_discovery_candidates.go (same package):
//   isBlockedDomain, detectGroup, extractDomain, extractRootURL,
//   extractPracticeName, extractUKPostcode, nullIfEmpty, nullIfFalse
// Do NOT redefine them here.
//
// Workflow config:
//
//	"process_results": {
//	    "action": "process_area_sweep",
//	    "config": {
//	        "input_fields": ["district_code", "area_name", "search_area_id", "search_results"]
//	    },
//	    "output_field": "sweep_result",
//	    "next_step": "complete"
//	}
//
// Registration:
//   "process_area_sweep": ProcessAreaSweepAction,

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ProcessAreaSweepInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"search_results"},
	Optional: []string{"district_code", "area_name", "search_area_id"},
}

func init() {
	datahelpers.RegisterActionInputSpec("process_area_sweep", ProcessAreaSweepInputSpec)
}

func ProcessAreaSweepAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ProcessAreaSweepAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ProcessAreaSweepInputSpec,
		params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	districtCode := inputs.Get("district_code")
	areaName := inputs.Get("area_name")
	searchAreaID := inputs.Get("search_area_id")

	params.Logger.Info("ProcessAreaSweepAction: processing",
		zap.String("district_code", districtCode),
		zap.String("area_name", areaName))

	// -- Extract search results from the nested structure --
	var results []interface{}
	var searchQuery string

	searchResults := inputs.GetMap("search_results")
	if searchResults != nil {
		// search_results may have response.results or just results
		if resp, ok := searchResults["response"].(map[string]interface{}); ok {
			if r, ok := resp["results"].([]interface{}); ok {
				results = r
			}
			searchQuery, _ = resp["query"].(string)
		}
		if results == nil {
			if r, ok := searchResults["results"].([]interface{}); ok {
				results = r
			}
			if searchQuery == "" {
				searchQuery, _ = searchResults["query"].(string)
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
	dbErrors := 0

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

		// Filter obviously non-vet results. The search query is already
		// vet-specific so we're lenient — only exclude clear mismatches
		// that don't also mention vet keywords.
		combined := strings.ToLower(resultTitle + " " + resultSnippet)
		nonVetIndicators := []string{
			"pet shop", "pet store", "dog grooming", "cat cafe",
			"pet food", "dog walking", "pet sitting",
			"wikipedia", "news article",
		}
		hasVetKeyword := strings.Contains(combined, "veterinary") ||
			strings.Contains(combined, "vet ") ||
			strings.Contains(combined, "vets ")
		isNotVet := false
		if !hasVetKeyword {
			for _, indicator := range nonVetIndicators {
				if strings.Contains(combined, indicator) {
					isNotVet = true
					break
				}
			}
		}
		if isNotVet {
			skipped++
			continue
		}

		// Check if already in businesses table
		rootURL := extractRootURL(resultURL)
		var existingID string
		err := params.DB.QueryRowContext(ctx,
			`SELECT id FROM business_intel.businesses
			 WHERE website_url ILIKE $1 OR website_url ILIKE $2
			 LIMIT 1`,
			rootURL+"%", "www."+strings.TrimPrefix(rootURL, "https://")+"%",
		).Scan(&existingID)

		if err != nil && err != sql.ErrNoRows {
			// Real DB error — log and skip
			params.Logger.Warn("ProcessAreaSweepAction: DB error checking businesses",
				zap.String("url", resultURL), zap.Error(err))
			dbErrors++
			continue
		}
		if err == nil {
			alreadyKnown++
			continue
		}
		// err == sql.ErrNoRows — not in businesses table

		// Check if already a discovery candidate (by source_url)
		var existingCandID string
		err = params.DB.QueryRowContext(ctx,
			`SELECT id FROM business_intel.discovery_candidates
			 WHERE source_url = $1 LIMIT 1`,
			resultURL,
		).Scan(&existingCandID)

		if err != nil && err != sql.ErrNoRows {
			params.Logger.Warn("ProcessAreaSweepAction: DB error checking candidates",
				zap.String("url", resultURL), zap.Error(err))
			dbErrors++
			continue
		}
		if err == nil {
			alreadyKnown++
			continue
		}
		// err == sql.ErrNoRows — not yet a candidate

		// Detect group affiliation
		detectedGroup, isGroup := detectGroup(domain)
		isIndependent := !isGroup

		// Extract name and postcode from result
		candidateName := extractPracticeName(resultTitle)
		postcode := extractUKPostcode(resultSnippet)
		if postcode == "" && districtCode != "" {
			postcode = districtCode
		}

		// Insert as discovery candidate
		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO business_intel.discovery_candidates
				(name, website_url, address_snippet, postcode,
				 source_query, source_url,
				 detected_group, is_independent,
				 status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
				'pending', NOW(), NOW())
			ON CONFLICT (source_url) DO UPDATE SET
				name = COALESCE(NULLIF(EXCLUDED.name, ''), business_intel.discovery_candidates.name),
				postcode = COALESCE(NULLIF(EXCLUDED.postcode, ''), business_intel.discovery_candidates.postcode),
				detected_group = COALESCE(EXCLUDED.detected_group, business_intel.discovery_candidates.detected_group),
				is_independent = COALESCE(EXCLUDED.is_independent, business_intel.discovery_candidates.is_independent),
				updated_at = NOW()`,
			candidateName, rootURL, resultSnippet, postcode,
			searchQuery, resultURL,
			nullIfEmpty(detectedGroup), nullIfFalse(isIndependent),
		)
		if err != nil {
			params.Logger.Warn("ProcessAreaSweepAction: failed to insert candidate",
				zap.String("url", resultURL), zap.Error(err))
			dbErrors++
			continue
		}
		candidates++

		params.Logger.Info("ProcessAreaSweepAction: found candidate",
			zap.String("name", candidateName),
			zap.String("url", rootURL),
			zap.String("district", districtCode),
			zap.String("group", detectedGroup))
	}

	// Update sweep tracking
	updateSweepTracking(ctx, params.DB, searchAreaID, districtCode, len(results), candidates, params.Logger)

	params.Logger.Info("ProcessAreaSweepAction: complete",
		zap.String("district_code", districtCode),
		zap.Int("scanned", scanned),
		zap.Int("candidates", candidates),
		zap.Int("skipped", skipped),
		zap.Int("already_known", alreadyKnown),
		zap.Int("db_errors", dbErrors))

	return map[string]interface{}{
		"district_code": districtCode,
		"area_name":     areaName,
		"scanned":       scanned,
		"candidates":    candidates,
		"already_known": alreadyKnown,
		"skipped":       skipped,
		"db_errors":     dbErrors,
	}, nil
}

// updateSweepTracking updates the search_areas table after a sweep
func updateSweepTracking(ctx context.Context, db *sql.DB, searchAreaID, districtCode string, resultCount, candidatesFound int, logger *zap.Logger) {
	var err error
	if searchAreaID != "" {
		_, err = db.ExecContext(ctx, `
			UPDATE business_intel.search_areas
			SET last_swept_at = NOW(),
			    sweep_count = sweep_count + 1,
			    last_result_count = $1,
			    candidates_found = candidates_found + $2
			WHERE id = $3`,
			resultCount, candidatesFound, searchAreaID)
	} else if districtCode != "" {
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
		logger.Warn("Failed to update sweep tracking",
			zap.String("district_code", districtCode), zap.Error(err))
	}
}
