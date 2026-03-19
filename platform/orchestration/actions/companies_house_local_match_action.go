// FILE: platform/orchestration/actions/ch_local_match_action.go
// Matches businesses against the local ch_vet_companies table.
// No API calls — pure SQL + Go scoring. Safe to re-run (updates existing matches).
//
// Actions:
//   - ch_local_match: Match verified businesses to CH companies using postcode + name similarity

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// chCandidate represents a row from ch_vet_companies for scoring
type chCandidate struct {
	CompanyNumber  string
	CompanyName    string
	CompanyStatus  string
	Postcode       string
	PostcodePrefix string
	Locality       string
}

// chLocalMatchResult holds a scored match
type chLocalMatchResult struct {
	CompanyNumber string
	CompanyName   string
	Score         float64
	Method        string
}

// CHLocalMatchAction matches verified businesses against the local ch_vet_companies table.
// Processes all unmatched businesses in one action call — no workflow loop needed.
func CHLocalMatchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CHLocalMatchAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	// Configuration
	batchSize := 500
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	threshold := 0.40
	if t, ok := config["threshold"].(float64); ok && t > 0 {
		threshold = t
	}

	// Allow re-matching previously matched businesses
	rematch := false
	if r, ok := config["rematch"].(bool); ok {
		rematch = r
	}

	// Load unmatched businesses
	businesses, err := loadUnmatchedBusinesses(ctx, params.DB, batchSize, rematch, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load businesses: %w", err)
	}

	if len(businesses) == 0 {
		params.Logger.Info("CHLocalMatch: no unmatched businesses to process")
		return map[string]interface{}{
			"status":          "complete",
			"total_processed": 0,
			"total_matched":   0,
			"total_no_match":  0,
		}, nil
	}

	params.Logger.Info("CHLocalMatch: loaded businesses",
		zap.Int("count", len(businesses)))

	totalProcessed := 0
	totalMatched := 0
	totalNoMatch := 0

	for _, biz := range businesses {
		select {
		case <-ctx.Done():
			params.Logger.Warn("CHLocalMatch: context cancelled",
				zap.Int("processed", totalProcessed))
			return map[string]interface{}{
				"status":          "interrupted",
				"total_processed": totalProcessed,
				"total_matched":   totalMatched,
				"total_no_match":  totalNoMatch,
			}, nil
		default:
		}

		// Find candidates by postcode prefix
		candidates, err := findCandidatesByPostcode(ctx, params.DB, biz.PostcodePrefix)
		if err != nil {
			params.Logger.Warn("CHLocalMatch: failed to find candidates",
				zap.String("business", biz.Name),
				zap.Error(err))
			totalProcessed++
			totalNoMatch++
			continue
		}

		// Score candidates
		best := scoreLocalCandidates(candidates, biz.Name, biz.Postcode, params.Logger)

		if best != nil && best.Score >= threshold {
			// Store match
			err := storeLocalMatch(ctx, params.DB, best.CompanyNumber, biz.ID, best.Score, best.Method)
			if err != nil {
				params.Logger.Warn("CHLocalMatch: failed to store match",
					zap.String("business", biz.Name),
					zap.String("company", best.CompanyName),
					zap.Error(err))
			} else {
				params.Logger.Info("CHLocalMatch: matched",
					zap.String("business", biz.Name),
					zap.String("company", best.CompanyName),
					zap.Float64("score", best.Score),
					zap.String("method", best.Method))
				totalMatched++
			}
		} else {
			totalNoMatch++
			if best != nil {
				params.Logger.Info("CHLocalMatch: below threshold",
					zap.String("business", biz.Name),
					zap.String("best_candidate", best.CompanyName),
					zap.Float64("score", best.Score))
			}
		}

		totalProcessed++
	}

	params.Logger.Info("CHLocalMatch: complete",
		zap.Int("total_processed", totalProcessed),
		zap.Int("total_matched", totalMatched),
		zap.Int("total_no_match", totalNoMatch))

	return map[string]interface{}{
		"status":          "complete",
		"total_processed": totalProcessed,
		"total_matched":   totalMatched,
		"total_no_match":  totalNoMatch,
	}, nil
}

// businessForMatching is a minimal struct for matching
type businessForMatching struct {
	ID             string
	Name           string
	Postcode       string
	PostcodePrefix string
}

// loadUnmatchedBusinesses loads verified businesses that don't yet have a match
// in ch_vet_companies (or all businesses if rematch=true).
func loadUnmatchedBusinesses(ctx context.Context, db *sql.DB, limit int, rematch bool, logger *zap.Logger) ([]businessForMatching, error) {
	var query string
	if rematch {
		// Load all verified businesses
		query = `
			SELECT b.id, b.name, b.postcode,
				   SPLIT_PART(UPPER(b.postcode), ' ', 1) as postcode_prefix
			FROM business_intel.businesses b
			JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
			WHERE bv.slug = 'veterinary'
			  AND b.verification_status = 'verified'
			  AND b.postcode IS NOT NULL AND b.postcode != ''
			ORDER BY b.name
			LIMIT $1`
	} else {
		// Load only businesses not yet matched in ch_vet_companies
		query = `
			SELECT b.id, b.name, b.postcode,
				   SPLIT_PART(UPPER(b.postcode), ' ', 1) as postcode_prefix
			FROM business_intel.businesses b
			JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
			LEFT JOIN business_intel.ch_vet_companies ch 
				ON ch.matched_business_id = b.id::uuid
			WHERE bv.slug = 'veterinary'
			  AND b.verification_status = 'verified'
			  AND b.postcode IS NOT NULL AND b.postcode != ''
			  AND ch.company_number IS NULL
			ORDER BY b.name
			LIMIT $1`
	}

	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []businessForMatching
	for rows.Next() {
		var biz businessForMatching
		if err := rows.Scan(&biz.ID, &biz.Name, &biz.Postcode, &biz.PostcodePrefix); err != nil {
			logger.Warn("CHLocalMatch: failed to scan business", zap.Error(err))
			continue
		}
		businesses = append(businesses, biz)
	}
	return businesses, rows.Err()
}

// findCandidatesByPostcode returns all CH vet companies in the same postcode area.
func findCandidatesByPostcode(ctx context.Context, db *sql.DB, postcodePrefix string) ([]chCandidate, error) {
	if postcodePrefix == "" {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT company_number, company_name, company_status, 
			   COALESCE(postcode, ''), COALESCE(postcode_prefix, ''), COALESCE(locality, '')
		FROM business_intel.ch_vet_companies
		WHERE postcode_prefix = $1
		  AND company_status = 'active'`,
		postcodePrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []chCandidate
	for rows.Next() {
		var c chCandidate
		if err := rows.Scan(&c.CompanyNumber, &c.CompanyName, &c.CompanyStatus,
			&c.Postcode, &c.PostcodePrefix, &c.Locality); err != nil {
			continue
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

// scoreLocalCandidates scores CH company candidates against a business.
// Adapted from scoreCHMatches but simplified for local matching:
// - All candidates already have SIC 75000 — no SIC penalty
// - All are active — no active bonus needed (already filtered)
// - Postcode prefix already matched by query — focus on name similarity
func scoreLocalCandidates(candidates []chCandidate, businessName, businessPostcode string, logger *zap.Logger) *chLocalMatchResult {
	if len(candidates) == 0 {
		return nil
	}

	// Clean both names for comparison using the same cleaner
	cleanedBizName := cleanCompanySearchName(businessName)
	bizNameLower := strings.ToLower(cleanedBizName)

	// Also get full postcode for exact match bonus
	bizPostcodeUpper := strings.ToUpper(strings.TrimSpace(businessPostcode))

	var best *chLocalMatchResult

	for _, c := range candidates {
		score := 0.0
		method := "local_postcode"

		// Clean the CH company name too
		cleanedCHName := cleanCompanySearchName(c.CompanyName)
		chNameLower := strings.ToLower(cleanedCHName)

		// All candidates matched on postcode prefix already: +0.20
		// (lower than API scoring's +0.35 because prefix match is just the filter here)
		score += 0.20

		// Bonus: exact full postcode match (+0.15)
		if bizPostcodeUpper != "" && c.Postcode != "" {
			chPostcodeUpper := strings.ToUpper(strings.TrimSpace(c.Postcode))
			// Normalise: remove all spaces for comparison
			if strings.ReplaceAll(bizPostcodeUpper, " ", "") == strings.ReplaceAll(chPostcodeUpper, " ", "") {
				score += 0.15
				method = "local_postcode_exact"
			}
		}

		// Name similarity scoring (same logic as scoreCHMatches)
		if chNameLower == bizNameLower {
			score += 0.30 // exact match on cleaned names — strong signal
		} else if strings.Contains(chNameLower, bizNameLower) || strings.Contains(bizNameLower, chNameLower) {
			score += 0.20
		} else {
			// Word overlap
			bizWords := strings.Fields(bizNameLower)
			matchedWords := 0
			for _, word := range bizWords {
				if len(word) > 3 && strings.Contains(chNameLower, word) {
					matchedWords++
				}
			}
			if len(bizWords) > 0 {
				wordRatio := float64(matchedWords) / float64(len(bizWords))
				score += wordRatio * 0.25
			}
		}

		// Vet industry signal in CH company name (+0.10)
		// Lower bonus than API scoring because all candidates are SIC 75000
		// — "veterinary" in the name is confirmatory, not discriminating.
		chTitleWords := strings.Fields(strings.ToLower(c.CompanyName))
		for _, w := range chTitleWords {
			cleaned := strings.TrimRight(w, ".,;:()")
			if cleaned == "veterinary" || cleaned == "vet" || cleaned == "vets" {
				score += 0.10
				break
			}
		}

		if best == nil || score > best.Score {
			best = &chLocalMatchResult{
				CompanyNumber: c.CompanyNumber,
				CompanyName:   c.CompanyName,
				Score:         score,
				Method:        method,
			}
		}
	}

	return best
}

// storeLocalMatch updates ch_vet_companies with the matched business ID.
func storeLocalMatch(ctx context.Context, db *sql.DB, companyNumber, businessID string, confidence float64, method string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE business_intel.ch_vet_companies
		SET matched_business_id = $1,
			matched_at = NOW(),
			match_confidence = $2,
			match_method = $3,
			updated_at = NOW()
		WHERE company_number = $4`,
		businessID,
		confidence,
		method,
		companyNumber,
	)
	return err
}
