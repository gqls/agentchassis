// FILE: platform/orchestration/actions/process_area_sweep.go
//
// ProcessAreaSweepAction processes web search results from a single district
// sweep. Checks each result against known businesses and existing candidates,
// inserts new ones into discovery_candidates, updates search_areas tracking.
//
// Shares helpers with scan_discovery_candidates.go (same package):
//   extractDomain, extractRootURL, extractPracticeName,
//   extractUKPostcode, skipDomains, vetKeywords
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

		// Skip directory/aggregator sites (reuses skipDomains from scan_discovery_candidates.go)
		domain := extractDomain(resultURL)
		if skipDomains[domain] {
			skipped++
			continue
		}

		// Filter obviously non-vet results (search query is already vet-specific
		// so we're lenient, just excluding clear mismatches)
		combined := strings.ToLower(resultTitle + " " + resultSnippet)
		nonVetIndicators := []string{
			"pet shop", "pet store", "dog grooming", "cat cafe",
			"pet food", "dog walking", "pet sitting",
			"wikipedia", "news article",
		}
		isNotVet := false
		for _, indicator := range nonVetIndicators {
			if strings.Contains(combined, indicator) &&
				!strings.Contains(combined, "veterinary") &&
				!strings.Contains(combined, "vet ") {
				isNotVet = true
				break
			}
		}
		if isNotVet {
			skipped++
			continue
		}

		// Check if already in businesses table
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

		// Check if already a discovery candidate
		var existingCandID sql.NullString
		err = params.DB.QueryRowContext(ctx,
			`SELECT id FROM business_intel.discovery_candidates 
			 WHERE source_url = $1 LIMIT 1`,
			resultURL,
		).Scan(&existingCandID)

		if err == nil && existingCandID.Valid {
			alreadyKnown++
			continue
		}

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
			params.Logger.Warn("Failed to insert candidate",
				zap.String("url", resultURL), zap.Error(err))
			continue
		}
		candidates++

		params.Logger.Info("Found candidate",
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
