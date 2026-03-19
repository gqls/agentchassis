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
	Postcode      string
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
	totalMatchedPass2 := 0

	// Minimum trigram similarity for name-only matching (pass 2)
	nameOnlyThreshold := 0.70
	if t, ok := config["name_only_threshold"].(float64); ok && t > 0 {
		nameOnlyThreshold = t
	}

	// Track unmatched businesses for pass 2
	var unmatchedForPass2 []businessForMatching

	// =====================================================================
	// PASS 1: Postcode prefix + name scoring
	// =====================================================================
	params.Logger.Info("CHLocalMatch: starting pass 1 (postcode + name)")

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
			unmatchedForPass2 = append(unmatchedForPass2, biz)
			totalProcessed++
			continue
		}

		// Score candidates
		best := scoreLocalCandidates(candidates, biz.Name, biz.Postcode, params.Logger)

		if best != nil && best.Score >= threshold {
			err := storeLocalMatch(ctx, params.DB, best.CompanyNumber, biz.ID, best.Score, best.Method)
			if err != nil {
				params.Logger.Warn("CHLocalMatch: failed to store match",
					zap.String("business", biz.Name),
					zap.String("company", best.CompanyName),
					zap.Error(err))
			} else {
				params.Logger.Info("CHLocalMatch: pass1 matched",
					zap.String("business", biz.Name),
					zap.String("company", best.CompanyName),
					zap.Float64("score", best.Score),
					zap.String("method", best.Method))
				totalMatched++
			}
		} else {
			unmatchedForPass2 = append(unmatchedForPass2, biz)
		}

		totalProcessed++
	}

	params.Logger.Info("CHLocalMatch: pass 1 complete",
		zap.Int("matched", totalMatched),
		zap.Int("unmatched_for_pass2", len(unmatchedForPass2)))

	// =====================================================================
	// PASS 2: Name-only matching via trigram similarity (no postcode requirement)
	// Uses GiST trigram index for fast per-business lookups (~4ms each).
	// Catches companies registered at accountant/HQ addresses.
	// =====================================================================
	params.Logger.Info("CHLocalMatch: starting pass 2 (name-only trigram)",
		zap.Int("businesses", len(unmatchedForPass2)),
		zap.Float64("name_only_threshold", nameOnlyThreshold))

	for _, biz := range unmatchedForPass2 {
		select {
		case <-ctx.Done():
			break
		default:
		}

		match, err := findByTrigramSimilarity(ctx, params.DB, biz.Name, nameOnlyThreshold)
		if err != nil {
			params.Logger.Warn("CHLocalMatch: trigram query failed",
				zap.String("business", biz.Name),
				zap.Error(err))
			totalNoMatch++
			continue
		}

		if match != nil {
			// Post-check: verify at least one distinctive word from the business
			// name appears in the CH name. This prevents false positives where
			// common vet words inflate the trigram score.
			if !hasDistinctiveWordOverlap(biz.Name, match.CompanyName) {
				params.Logger.Info("CHLocalMatch: pass2 rejected (no distinctive word overlap)",
					zap.String("business", biz.Name),
					zap.String("company", match.CompanyName),
					zap.Float64("score", match.Score))
				totalNoMatch++
				continue
			}

			err := storeLocalMatch(ctx, params.DB, match.CompanyNumber, biz.ID, match.Score, match.Method)
			if err != nil {
				params.Logger.Warn("CHLocalMatch: failed to store pass2 match",
					zap.String("business", biz.Name),
					zap.String("company", match.CompanyName),
					zap.Error(err))
			} else {
				params.Logger.Info("CHLocalMatch: pass2 matched",
					zap.String("business", biz.Name),
					zap.String("company", match.CompanyName),
					zap.Float64("score", match.Score),
					zap.String("method", match.Method))
				totalMatchedPass2++
				totalMatched++
			}
		} else {
			totalNoMatch++
		}
	}

	params.Logger.Info("CHLocalMatch: pass 2 complete",
		zap.Int("pass2_matched", totalMatchedPass2))

	// Query overall stats from the table
	stats := map[string]interface{}{}
	row := params.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) as total_ch_companies,
			   COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched,
			   COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched
		FROM business_intel.ch_vet_companies
		WHERE company_status = 'active'`)
	var totalCH, totalCHMatched, totalCHUnmatched int
	if err := row.Scan(&totalCH, &totalCHMatched, &totalCHUnmatched); err == nil {
		stats["total_ch_companies"] = totalCH
		stats["ch_matched"] = totalCHMatched
		stats["ch_unmatched"] = totalCHUnmatched
		if totalCH > 0 {
			stats["match_pct"] = fmt.Sprintf("%.1f", 100.0*float64(totalCHMatched)/float64(totalCH))
		}
	}

	// Notify scheduler that this task completed
	taskName := "ch-local-match"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, err = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`,
		taskName)
	if err != nil {
		params.Logger.Warn("CHLocalMatch: failed to notify scheduler", zap.Error(err))
	}

	params.Logger.Info("CHLocalMatch: complete",
		zap.Int("total_processed", totalProcessed),
		zap.Int("total_matched", totalMatched),
		zap.Int("pass1_matched", totalMatched-totalMatchedPass2),
		zap.Int("pass2_matched", totalMatchedPass2),
		zap.Int("total_no_match", totalNoMatch),
		zap.Any("stats", stats))

	return map[string]interface{}{
		"status":          "complete",
		"total_processed": totalProcessed,
		"total_matched":   totalMatched,
		"pass1_matched":   totalMatched - totalMatchedPass2,
		"pass2_matched":   totalMatchedPass2,
		"total_no_match":  totalNoMatch,
		"stats":           stats,
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
				Postcode:      c.Postcode,
				Score:         score,
				Method:        method,
			}
		}
	}

	return best
}

// findByTrigramSimilarity finds the best name match across ALL postcodes
// using the GiST trigram index on company_name_cleaned. Returns nil if no
// match above threshold. Compares cleaned names (stripped of Ltd/Group/Surgery etc.)
// for much better similarity scores.
func findByTrigramSimilarity(ctx context.Context, db *sql.DB, businessName string, minSimilarity float64) (*chLocalMatchResult, error) {
	// Clean the business name the same way CH names are cleaned in the DB
	cleanedName := strings.ToLower(cleanCompanySearchName(businessName))
	if cleanedName == "" || len(cleanedName) < 4 {
		return nil, nil
	}

	// Query against company_name_cleaned using the GiST trigram index.
	// Both sides are now stripped of Ltd/Group/Surgery etc.
	rows, err := db.QueryContext(ctx, `
		SELECT company_number, company_name, COALESCE(postcode, ''),
			   similarity(company_name_cleaned, $1) as sim
		FROM business_intel.ch_vet_companies
		WHERE company_status = 'active'
		  AND matched_business_id IS NULL
		  AND company_name_cleaned % $1
		ORDER BY company_name_cleaned <-> $1
		LIMIT 3`,
		cleanedName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var best *chLocalMatchResult
	for rows.Next() {
		var companyNumber, companyName, postcode string
		var sim float64
		if err := rows.Scan(&companyNumber, &companyName, &postcode, &sim); err != nil {
			continue
		}

		if sim < minSimilarity {
			continue
		}

		if best == nil || sim > best.Score {
			best = &chLocalMatchResult{
				CompanyNumber: companyNumber,
				CompanyName:   companyName,
				Postcode:      postcode,
				Score:         sim,
				Method:        "name_trigram",
			}
		}
	}

	return best, rows.Err()
}

// hasDistinctiveWordOverlap checks that at least one non-generic word from
// the business name appears in the CH company name. This prevents false
// positives where common vet/business words inflate trigram similarity.
// E.g. "WW Mobile Veterinary Services" vs "HAYLOFT MOBILE VETERINARY SERVICES"
// — high trigram score but "WW" doesn't appear in the CH name.
func hasDistinctiveWordOverlap(businessName, companyName string) bool {
	// Words that appear in many vet company names — not distinctive
	genericWords := map[string]bool{
		"the": true, "and": true, "of": true, "for": true, "in": true, "at": true, "a": true,
		"veterinary": true, "vet": true, "vets": true, "vets4pets": true,
		"animal": true, "pet": true, "pets": true, "paws": true,
		"mobile": true, "services": true, "service": true,
		"clinic": true, "centre": true, "center": true, "surgery": true,
		"practice": true, "hospital": true, "group": true,
		"limited": true, "ltd": true, "llp": true, "plc": true,
		"equine": true, "farm": true, "emergency": true, "referrals": true,
	}

	bizWords := strings.Fields(strings.ToLower(businessName))
	chNameLower := strings.ToLower(companyName)

	for _, word := range bizWords {
		// Strip punctuation
		word = strings.TrimRight(word, ".,;:()'-")
		word = strings.TrimLeft(word, "('")

		if len(word) < 3 {
			continue
		}
		if genericWords[word] {
			continue
		}

		// This is a distinctive word — does it appear in the CH name?
		if strings.Contains(chNameLower, word) {
			return true
		}
	}

	return false
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
