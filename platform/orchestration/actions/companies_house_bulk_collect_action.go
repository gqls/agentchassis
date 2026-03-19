// FILE: platform/orchestration/actions/ch_bulk_collect_action.go
// Bulk collects all Companies House companies with SIC 75000 (veterinary activities)
// into the local ch_vet_companies table. Handles pagination internally — no workflow
// loop needed. Safe to re-run (uses upsert).
//
// Actions:
//   - ch_bulk_collect: Paginate through CH advanced search, store all results locally

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CHBulkCollectAction paginates through all CH companies with the configured
// SIC code and stores them locally. Handles all pagination in a single action
// call to avoid orchestration state bloat.
func CHBulkCollectAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CHBulkCollectAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	// Configuration with defaults
	sicCode := "75000"
	if sc, ok := config["sic_code"].(string); ok && sc != "" {
		sicCode = sc
	}

	companyStatus := "active"
	if cs, ok := config["company_status"].(string); ok && cs != "" {
		companyStatus = cs
	}

	pageSize := 100
	if ps, ok := config["page_size"].(float64); ok && ps > 0 {
		pageSize = int(ps)
	}

	delayMs := 2000 // 2 seconds between pages — well within 600 req/5 min
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}

	// Allow resuming from a specific offset (e.g. if previous run was interrupted)
	startFrom := 0
	if sf, ok := config["start_from"].(float64); ok && sf > 0 {
		startFrom = int(sf)
	}
	// Input data override
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if sf, ok := inputData["start_from"].(float64); ok && sf > 0 {
			startFrom = int(sf)
		}
	}

	apiKey := os.Getenv("COMPANIES_HOUSE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("COMPANIES_HOUSE_API_KEY not set")
	}

	// Ensure table exists (idempotent)
	if err := ensureCHVetCompaniesTable(ctx, params.DB); err != nil {
		return nil, fmt.Errorf("failed to ensure table: %w", err)
	}

	// Paginate through all results
	startIndex := startFrom
	totalCollected := 0
	totalNew := 0
	totalUpdated := 0
	totalHits := 0
	pagesProcessed := 0

	for {
		// Check context for cancellation
		select {
		case <-ctx.Done():
			params.Logger.Warn("CHBulkCollect: context cancelled, stopping",
				zap.Int("collected_so_far", totalCollected),
				zap.Int("start_index", startIndex))
			return map[string]interface{}{
				"status":           "interrupted",
				"total_collected":  totalCollected,
				"total_new":        totalNew,
				"total_updated":    totalUpdated,
				"stopped_at_index": startIndex,
				"total_hits":       totalHits,
			}, nil
		default:
		}

		// Rate limiting between pages
		if pagesProcessed > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// Build the advanced search URL
		searchPath := fmt.Sprintf("/advanced-search/companies?sic_codes=%s&company_status=%s&size=%d&start_index=%d",
			sicCode, companyStatus, pageSize, startIndex)

		params.Logger.Info("CHBulkCollect: fetching page",
			zap.Int("start_index", startIndex),
			zap.Int("page_size", pageSize),
			zap.Int("pages_processed", pagesProcessed),
			zap.Int("total_collected", totalCollected))

		result, err := chAPIGet(ctx, apiKey, searchPath)
		if err != nil {
			// If rate limited, wait and retry once
			if strings.Contains(err.Error(), "429") {
				params.Logger.Warn("CHBulkCollect: rate limited, waiting 60s")
				time.Sleep(60 * time.Second)
				result, err = chAPIGet(ctx, apiKey, searchPath)
			}
			if err != nil {
				params.Logger.Error("CHBulkCollect: API error, stopping",
					zap.Error(err),
					zap.Int("start_index", startIndex))
				return map[string]interface{}{
					"status":           "error",
					"error":            err.Error(),
					"total_collected":  totalCollected,
					"total_new":        totalNew,
					"total_updated":    totalUpdated,
					"stopped_at_index": startIndex,
					"total_hits":       totalHits,
				}, nil
			}
		}

		// Extract total hits
		if hits, ok := result["hits"].(float64); ok {
			totalHits = int(hits)
		}

		// Extract items
		items, ok := result["items"].([]interface{})
		if !ok || len(items) == 0 {
			params.Logger.Info("CHBulkCollect: no more items, collection complete",
				zap.Int("total_collected", totalCollected),
				zap.Int("total_hits", totalHits))
			break
		}

		// Store this page
		newCount, updatedCount, err := storeCHVetCompaniesPage(ctx, params.DB, items, params.Logger)
		if err != nil {
			params.Logger.Error("CHBulkCollect: failed to store page",
				zap.Error(err),
				zap.Int("start_index", startIndex))
			// Continue to next page rather than failing entirely
		}

		totalCollected += len(items)
		totalNew += newCount
		totalUpdated += updatedCount
		pagesProcessed++

		// Move to next page
		startIndex += pageSize

		// Stop if we've fetched everything
		if startIndex >= totalHits {
			params.Logger.Info("CHBulkCollect: reached end of results",
				zap.Int("total_collected", totalCollected),
				zap.Int("total_hits", totalHits))
			break
		}
	}

	params.Logger.Info("CHBulkCollect: complete",
		zap.Int("total_collected", totalCollected),
		zap.Int("total_new", totalNew),
		zap.Int("total_updated", totalUpdated),
		zap.Int("total_hits", totalHits),
		zap.Int("pages_processed", pagesProcessed))

	// Query stats from the table
	stats := map[string]interface{}{}
	row := params.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) as total,
			   COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched,
			   COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched
		FROM business_intel.ch_vet_companies
		WHERE company_status = 'active'`)
	var totalCH, matched, unmatched int
	if err := row.Scan(&totalCH, &matched, &unmatched); err == nil {
		stats["total_ch_companies"] = totalCH
		stats["matched"] = matched
		stats["unmatched"] = unmatched
	}

	// Notify scheduler
	taskName := "ch-vet-collect"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`,
		taskName)

	return map[string]interface{}{
		"status":          "complete",
		"total_collected": totalCollected,
		"total_new":       totalNew,
		"total_updated":   totalUpdated,
		"total_hits":      totalHits,
		"pages_processed": pagesProcessed,
		"stats":           stats,
	}, nil
}

// ensureCHVetCompaniesTable creates the table if it doesn't exist.
// This makes the action self-contained — no separate migration step needed.
func ensureCHVetCompaniesTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS business_intel.ch_vet_companies (
			company_number      VARCHAR(10) PRIMARY KEY,
			company_name        TEXT NOT NULL,
			company_name_cleaned TEXT,
			company_status      TEXT,
			company_type        TEXT,
			date_of_creation    DATE,
			date_of_cessation   DATE,
			sic_codes           TEXT[],
			registered_address  JSONB,
			postcode            TEXT,
			postcode_prefix     TEXT,
			locality            TEXT,
			matched_business_id UUID,
			matched_at          TIMESTAMPTZ,
			match_confidence    NUMERIC(3,2),
			match_method        TEXT,
			details_fetched     BOOLEAN NOT NULL DEFAULT FALSE,
			details_fetched_at  TIMESTAMPTZ,
			is_discovered       BOOLEAN NOT NULL DEFAULT FALSE,
			discovery_status    TEXT DEFAULT 'pending',
			collected_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		-- Add column if table already exists without it
		ALTER TABLE business_intel.ch_vet_companies 
			ADD COLUMN IF NOT EXISTS company_name_cleaned TEXT;

		CREATE INDEX IF NOT EXISTS idx_ch_vet_postcode_prefix 
			ON business_intel.ch_vet_companies (postcode_prefix) 
			WHERE company_status = 'active';

		CREATE INDEX IF NOT EXISTS idx_ch_vet_unmatched 
			ON business_intel.ch_vet_companies (matched_business_id) 
			WHERE matched_business_id IS NULL AND company_status = 'active';

		CREATE INDEX IF NOT EXISTS idx_ch_vet_status 
			ON business_intel.ch_vet_companies (company_status);
	`)
	return err
}

// storeCHVetCompaniesPage stores a page of results into the table.
// Returns (new_count, updated_count, error).
func storeCHVetCompaniesPage(ctx context.Context, db *sql.DB, items []interface{}, logger *zap.Logger) (int, int, error) {
	newCount := 0
	updatedCount := 0

	for _, item := range items {
		company, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		companyNumber, _ := company["company_number"].(string)
		if companyNumber == "" {
			continue
		}

		companyName, _ := company["company_name"].(string)
		companyStatus, _ := company["company_status"].(string)
		companyType, _ := company["company_type"].(string)
		dateOfCreation, _ := company["date_of_creation"].(string)
		dateOfCessation, _ := company["date_of_cessation"].(string)

		// SIC codes
		var sicCodes []string
		if sics, ok := company["sic_codes"].([]interface{}); ok {
			for _, s := range sics {
				if str, ok := s.(string); ok {
					sicCodes = append(sicCodes, str)
				}
			}
		}

		// Address — the advanced search returns "registered_office_address"
		var addressJSON []byte
		var postcode, postcodePrefix, locality string

		if addr, ok := company["registered_office_address"].(map[string]interface{}); ok {
			addressJSON, _ = json.Marshal(addr)
			postcode, _ = addr["postal_code"].(string)
			locality, _ = addr["locality"].(string)

			// Extract outward code (e.g. "BT74" from "BT74 6HR")
			if postcode != "" {
				parts := strings.Fields(strings.ToUpper(postcode))
				if len(parts) > 0 {
					postcodePrefix = parts[0]
				}
			}
		}

		// Parse dates
		var creationDate sql.NullTime
		if dateOfCreation != "" {
			if t, err := time.Parse("2006-01-02", dateOfCreation); err == nil {
				creationDate = sql.NullTime{Time: t, Valid: true}
			}
		}
		var cessationDate sql.NullTime
		if dateOfCessation != "" {
			if t, err := time.Parse("2006-01-02", dateOfCessation); err == nil {
				cessationDate = sql.NullTime{Time: t, Valid: true}
			}
		}

		// Clean company name for trigram matching
		companyNameCleaned := strings.ToLower(cleanCompanySearchName(companyName))

		// Upsert — safe for re-runs
		result, err := db.ExecContext(ctx, `
			INSERT INTO business_intel.ch_vet_companies (
				company_number, company_name, company_status, company_type,
				date_of_creation, date_of_cessation, sic_codes,
				registered_address, postcode, postcode_prefix, locality,
				company_name_cleaned
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (company_number) DO UPDATE SET
				company_name = EXCLUDED.company_name,
				company_status = EXCLUDED.company_status,
				company_type = EXCLUDED.company_type,
				date_of_creation = EXCLUDED.date_of_creation,
				date_of_cessation = EXCLUDED.date_of_cessation,
				sic_codes = EXCLUDED.sic_codes,
				registered_address = EXCLUDED.registered_address,
				postcode = EXCLUDED.postcode,
				postcode_prefix = EXCLUDED.postcode_prefix,
				locality = EXCLUDED.locality,
				company_name_cleaned = EXCLUDED.company_name_cleaned,
				updated_at = NOW()`,
			companyNumber,
			companyName,
			companyStatus,
			companyType,
			creationDate,
			cessationDate,
			pgArrayFromInterface(sicCodes), // reuse existing helper from business_intel_actions.go
			addressJSON,
			postcode,
			postcodePrefix,
			locality,
			companyNameCleaned,
		)
		if err != nil {
			logger.Warn("CHBulkCollect: failed to upsert company",
				zap.String("company_number", companyNumber),
				zap.Error(err))
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			// Check if it was an insert or update by seeing if collected_at changed
			// Simplification: just count all as collected
			newCount++
		}
	}

	return newCount, updatedCount, nil
}
