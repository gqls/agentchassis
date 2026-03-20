// FILE: platform/orchestration/actions/ch_local_match_action.go
// Matches businesses against the local ch_vet_companies table.
// No API calls — pure SQL + Go scoring. Safe to re-run.
//
// Matching cascade (run in order, each business falls through until matched):
//   Tier 1: Exact cleaned name (≥0.90 sim) + geographic confirmation (same town or postcode prefix)
//   Tier 2: Exact cleaned name (≥0.90 sim) + unique in CH (only one company with that name)
//   Tier 3: Same postcode prefix + name scoring (threshold 0.50)
//   Residual: Trigram 0.50-0.90 + distinctive word → pending_llm_review
//
// Actions:
//   - ch_local_match: Match verified businesses to CH companies

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

// businessForMatching is a minimal struct for matching
type businessForMatching struct {
	ID             string
	Name           string
	Postcode       string
	PostcodePrefix string
	Town           string
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

	batchSize := 3000
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	rematch := false
	if r, ok := config["rematch"].(bool); ok {
		rematch = r
	}

	// Pre-filter: minimum confidence to exclude directories and junk entries
	minConfidence := 0.40
	if mc, ok := config["min_confidence"].(float64); ok && mc > 0 {
		minConfidence = mc
	}

	// Tier thresholds
	nameAutoAccept := 0.90 // Tiers 1 & 2: cleaned name similarity for auto-accept
	if t, ok := config["name_auto_accept"].(float64); ok && t > 0 {
		nameAutoAccept = t
	}

	postcodeThreshold := 0.50 // Tier 3: postcode + name scoring threshold
	if t, ok := config["postcode_threshold"].(float64); ok && t > 0 {
		postcodeThreshold = t
	}

	reviewFloor := 0.50 // Residual: minimum trigram for LLM review
	if t, ok := config["review_floor"].(float64); ok && t > 0 {
		reviewFloor = t
	}

	// Load vertical profile
	verticalSlug := "veterinary"
	if vs, ok := config["vertical_slug"].(string); ok && vs != "" {
		verticalSlug = vs
	}
	profile := GetCHVerticalProfile(verticalSlug)

	params.Logger.Info("CHLocalMatch: config",
		zap.String("vertical", profile.Slug),
		zap.Float64("min_confidence", minConfidence),
		zap.Float64("name_auto_accept", nameAutoAccept),
		zap.Float64("postcode_threshold", postcodeThreshold),
		zap.Float64("review_floor", reviewFloor))

	// Load unmatched businesses (with pre-filter)
	businesses, err := loadUnmatchedBusinesses(ctx, params.DB, batchSize, rematch, minConfidence, profile, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load businesses: %w", err)
	}

	if len(businesses) == 0 {
		params.Logger.Info("CHLocalMatch: no unmatched businesses to process")
		return map[string]interface{}{
			"status":          "complete",
			"total_processed": 0,
		}, nil
	}

	params.Logger.Info("CHLocalMatch: loaded businesses",
		zap.Int("count", len(businesses)))

	// Counters
	tier1Matched := 0
	tier2Matched := 0
	tier3Matched := 0
	pendingReview := 0
	noMatch := 0

	// Track businesses remaining after each tier
	var afterTier1 []businessForMatching
	var afterTier2 []businessForMatching
	var afterTier3 []businessForMatching

	// =====================================================================
	// TIER 1: Exact cleaned name (≥0.90) + geographic confirmation
	// Strongest signal: the name matches and they're in the same area.
	// =====================================================================
	params.Logger.Info("CHLocalMatch: Tier 1 — exact name + geography",
		zap.Int("businesses", len(businesses)))

	for _, biz := range businesses {
		match, err := findByNameAndGeography(ctx, params.DB, biz, nameAutoAccept)
		if err != nil {
			params.Logger.Warn("CHLocalMatch: Tier 1 query failed",
				zap.String("business", biz.Name), zap.Error(err))
			afterTier1 = append(afterTier1, biz)
			continue
		}

		if match != nil {
			err := storeLocalMatch(ctx, params.DB, match.CompanyNumber, biz.ID, match.Score, "tier1_name_geo")
			if err == nil {
				tier1Matched++
			}
		} else {
			afterTier1 = append(afterTier1, biz)
		}
	}

	params.Logger.Info("CHLocalMatch: Tier 1 complete",
		zap.Int("matched", tier1Matched),
		zap.Int("remaining", len(afterTier1)))

	// =====================================================================
	// TIER 2: Exact cleaned name (≥0.90), unique in CH
	// Name matches strongly and there's only one such company — must be it.
	// =====================================================================
	params.Logger.Info("CHLocalMatch: Tier 2 — exact name, unique in CH",
		zap.Int("businesses", len(afterTier1)))

	for _, biz := range afterTier1 {
		match, err := findByNameUnique(ctx, params.DB, biz, nameAutoAccept, profile)
		if err != nil {
			afterTier2 = append(afterTier2, biz)
			continue
		}

		if match != nil {
			err := storeLocalMatch(ctx, params.DB, match.CompanyNumber, biz.ID, match.Score, "tier2_name_unique")
			if err == nil {
				tier2Matched++
			}
		} else {
			afterTier2 = append(afterTier2, biz)
		}
	}

	params.Logger.Info("CHLocalMatch: Tier 2 complete",
		zap.Int("matched", tier2Matched),
		zap.Int("remaining", len(afterTier2)))

	// =====================================================================
	// TIER 3: Same postcode prefix + name scoring
	// Weaker signal — requires real name overlap, not just vet words.
	// =====================================================================
	params.Logger.Info("CHLocalMatch: Tier 3 — postcode + name scoring",
		zap.Int("businesses", len(afterTier2)))

	for _, biz := range afterTier2 {
		candidates, err := findCandidatesByPostcode(ctx, params.DB, biz.PostcodePrefix)
		if err != nil || len(candidates) == 0 {
			afterTier3 = append(afterTier3, biz)
			continue
		}

		best := scoreLocalCandidates(candidates, biz.Name, biz.Postcode, profile, params.Logger)

		if best != nil && best.Score >= postcodeThreshold {
			err := storeLocalMatch(ctx, params.DB, best.CompanyNumber, biz.ID, best.Score, best.Method)
			if err == nil {
				tier3Matched++
			}
		} else {
			afterTier3 = append(afterTier3, biz)
		}
	}

	params.Logger.Info("CHLocalMatch: Tier 3 complete",
		zap.Int("matched", tier3Matched),
		zap.Int("remaining", len(afterTier3)))

	// =====================================================================
	// RESIDUAL: Trigram name search for LLM review candidates
	// Similarity 0.50–0.90 + distinctive word → pending_llm_review
	// =====================================================================
	params.Logger.Info("CHLocalMatch: Residual — LLM review candidates",
		zap.Int("businesses", len(afterTier3)))

	for _, biz := range afterTier3 {
		match, err := findByTrigramSimilarity(ctx, params.DB, biz.Name, reviewFloor)
		if err != nil || match == nil {
			noMatch++
			continue
		}

		if !hasDistinctiveWordOverlap(biz.Name, match.CompanyName, profile) {
			noMatch++
			continue
		}

		err = storeLocalMatch(ctx, params.DB, match.CompanyNumber, biz.ID, match.Score, "pending_llm_review")
		if err == nil {
			pendingReview++
		}
	}

	// Stats
	totalMatched := tier1Matched + tier2Matched + tier3Matched
	stats := map[string]interface{}{}
	row := params.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) as total_ch,
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

	// Notify scheduler
	taskName := "ch-local-match"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, err = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)
	if err != nil {
		params.Logger.Warn("CHLocalMatch: failed to notify scheduler", zap.Error(err))
	}

	params.Logger.Info("CHLocalMatch: complete",
		zap.Int("total_processed", len(businesses)),
		zap.Int("tier1_name_geo", tier1Matched),
		zap.Int("tier2_name_unique", tier2Matched),
		zap.Int("tier3_postcode_name", tier3Matched),
		zap.Int("total_matched", totalMatched),
		zap.Int("pending_llm_review", pendingReview),
		zap.Int("no_match", noMatch),
		zap.Any("stats", stats))

	return map[string]interface{}{
		"status":              "complete",
		"total_processed":     len(businesses),
		"tier1_name_geo":      tier1Matched,
		"tier2_name_unique":   tier2Matched,
		"tier3_postcode_name": tier3Matched,
		"total_matched":       totalMatched,
		"pending_llm_review":  pendingReview,
		"no_match":            noMatch,
		"stats":               stats,
	}, nil
}

// =========================================================================
// Business loader
// =========================================================================

// loadUnmatchedBusinesses loads verified businesses not yet matched,
// with pre-filter on confidence score and directory exclusion.
func loadUnmatchedBusinesses(ctx context.Context, db *sql.DB, limit int, rematch bool, minConfidence float64, profile *CHVerticalProfile, logger *zap.Logger) ([]businessForMatching, error) {
	var query string
	if rematch {
		query = `
			SELECT b.id, b.name, COALESCE(b.postcode, ''),
				   SPLIT_PART(UPPER(COALESCE(b.postcode, '')), ' ', 1),
				   COALESCE(b.town, '')
			FROM business_intel.businesses b
			JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
			WHERE bv.slug = $1
			  AND b.verification_status = 'verified'
			  AND b.confidence_score >= $2
			  AND b.business_type NOT ILIKE '%directory%'
			  AND b.business_type NOT ILIKE '%listing%'
			ORDER BY b.name
			LIMIT $3`
	} else {
		query = `
			SELECT b.id, b.name, COALESCE(b.postcode, ''),
				   SPLIT_PART(UPPER(COALESCE(b.postcode, '')), ' ', 1),
				   COALESCE(b.town, '')
			FROM business_intel.businesses b
			JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
			LEFT JOIN business_intel.ch_vet_companies ch 
				ON ch.matched_business_id = b.id::uuid
			WHERE bv.slug = $1
			  AND b.verification_status = 'verified'
			  AND b.confidence_score >= $2
			  AND b.business_type NOT ILIKE '%directory%'
			  AND b.business_type NOT ILIKE '%listing%'
			  AND ch.company_number IS NULL
			ORDER BY b.name
			LIMIT $3`
	}

	rows, err := db.QueryContext(ctx, query, profile.BusinessVerticalSlug, minConfidence, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []businessForMatching
	for rows.Next() {
		var biz businessForMatching
		if err := rows.Scan(&biz.ID, &biz.Name, &biz.Postcode, &biz.PostcodePrefix, &biz.Town); err != nil {
			logger.Warn("CHLocalMatch: failed to scan business", zap.Error(err))
			continue
		}
		businesses = append(businesses, biz)
	}
	return businesses, rows.Err()
}

// =========================================================================
// Tier 1: Exact name + geographic confirmation
// =========================================================================

// findByNameAndGeography finds a CH company with cleaned name similarity ≥ threshold
// that also shares the same postcode prefix OR the same town/locality.
func findByNameAndGeography(ctx context.Context, db *sql.DB, biz businessForMatching, minSim float64) (*chLocalMatchResult, error) {
	cleanedName := strings.ToLower(cleanCompanySearchName(biz.Name))
	if cleanedName == "" || len(cleanedName) < 4 {
		return nil, nil
	}

	bizTownLower := strings.ToLower(strings.TrimSpace(biz.Town))

	// Query: high name similarity + (same postcode prefix OR same locality)
	rows, err := db.QueryContext(ctx, `
		SELECT company_number, company_name, COALESCE(postcode, ''),
			   COALESCE(postcode_prefix, ''), COALESCE(locality, ''),
			   similarity(company_name_cleaned, $1) as sim
		FROM business_intel.ch_vet_companies
		WHERE company_status = 'active'
		  AND matched_business_id IS NULL
		  AND company_name_cleaned % $1
		  AND similarity(company_name_cleaned, $1) >= $2
		ORDER BY company_name_cleaned <-> $1
		LIMIT 5`,
		cleanedName, minSim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var companyNumber, companyName, postcode, postcodePrefix, locality string
		var sim float64
		if err := rows.Scan(&companyNumber, &companyName, &postcode, &postcodePrefix, &locality, &sim); err != nil {
			continue
		}

		// Geographic confirmation: same postcode prefix or same town
		geoMatch := false
		if biz.PostcodePrefix != "" && postcodePrefix == biz.PostcodePrefix {
			geoMatch = true
		}
		if !geoMatch && bizTownLower != "" && strings.ToLower(locality) == bizTownLower {
			geoMatch = true
		}

		if geoMatch {
			return &chLocalMatchResult{
				CompanyNumber: companyNumber,
				CompanyName:   companyName,
				Postcode:      postcode,
				Score:         sim,
				Method:        "tier1_name_geo",
			}, nil
		}
	}

	return nil, rows.Err()
}

// =========================================================================
// Tier 2: Exact name, unique in CH
// =========================================================================

// findByNameUnique finds a CH company with cleaned name similarity ≥ threshold
// where there's only one such company in the entire CH table. No geography needed.
// Also requires distinctive word overlap to prevent generic matches.
func findByNameUnique(ctx context.Context, db *sql.DB, biz businessForMatching, minSim float64, profile *CHVerticalProfile) (*chLocalMatchResult, error) {
	cleanedName := strings.ToLower(cleanCompanySearchName(biz.Name))
	if cleanedName == "" || len(cleanedName) < 4 {
		return nil, nil
	}

	// Find all CH companies matching at ≥ minSim
	rows, err := db.QueryContext(ctx, `
		SELECT company_number, company_name, COALESCE(postcode, ''),
			   similarity(company_name_cleaned, $1) as sim
		FROM business_intel.ch_vet_companies
		WHERE company_status = 'active'
		  AND matched_business_id IS NULL
		  AND company_name_cleaned % $1
		  AND similarity(company_name_cleaned, $1) >= $2
		ORDER BY sim DESC
		LIMIT 3`,
		cleanedName, minSim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type candidate struct {
		companyNumber, companyName, postcode string
		sim                                  float64
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.companyNumber, &c.companyName, &c.postcode, &c.sim); err != nil {
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Only accept if exactly one candidate with high similarity
	if len(candidates) != 1 {
		return nil, nil
	}

	c := candidates[0]

	// Require distinctive word overlap
	if !hasDistinctiveWordOverlap(biz.Name, c.companyName, profile) {
		return nil, nil
	}

	return &chLocalMatchResult{
		CompanyNumber: c.companyNumber,
		CompanyName:   c.companyName,
		Postcode:      c.postcode,
		Score:         c.sim,
		Method:        "tier2_name_unique",
	}, nil
}

// =========================================================================
// Tier 3: Postcode + name scoring (reused from original)
// =========================================================================

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
		  AND company_status = 'active'
		  AND matched_business_id IS NULL`,
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
// Uses profile for industry keyword bonus.
func scoreLocalCandidates(candidates []chCandidate, businessName, businessPostcode string, profile *CHVerticalProfile, logger *zap.Logger) *chLocalMatchResult {
	if len(candidates) == 0 {
		return nil
	}

	cleanedBizName := cleanCompanySearchName(businessName)
	bizNameLower := strings.ToLower(cleanedBizName)
	bizPostcodeUpper := strings.ToUpper(strings.TrimSpace(businessPostcode))

	var best *chLocalMatchResult

	for _, c := range candidates {
		score := 0.0
		method := "local_postcode"

		cleanedCHName := cleanCompanySearchName(c.CompanyName)
		chNameLower := strings.ToLower(cleanedCHName)

		// Postcode prefix match (always true from query): +0.20
		score += 0.20

		// Exact full postcode: +0.15
		if bizPostcodeUpper != "" && c.Postcode != "" {
			chPostcodeUpper := strings.ToUpper(strings.TrimSpace(c.Postcode))
			if strings.ReplaceAll(bizPostcodeUpper, " ", "") == strings.ReplaceAll(chPostcodeUpper, " ", "") {
				score += 0.15
				method = "local_postcode_exact"
			}
		}

		// Name similarity
		if chNameLower == bizNameLower {
			score += 0.30
		} else if strings.Contains(chNameLower, bizNameLower) || strings.Contains(bizNameLower, chNameLower) {
			score += 0.20
		} else {
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

		// Industry keyword bonus
		if len(profile.IndustryKeywords) > 0 && profile.IndustryKeywordBonus > 0 {
			if hasIndustryKeyword(c.CompanyName, profile.IndustryKeywords) {
				score += profile.IndustryKeywordBonus
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

// =========================================================================
// Residual: Trigram name search for LLM review
// =========================================================================

// findByTrigramSimilarity finds the best name match across ALL postcodes
// using the GiST trigram index on company_name_cleaned.
func findByTrigramSimilarity(ctx context.Context, db *sql.DB, businessName string, minSimilarity float64) (*chLocalMatchResult, error) {
	cleanedName := strings.ToLower(cleanCompanySearchName(businessName))
	if cleanedName == "" || len(cleanedName) < 4 {
		return nil, nil
	}

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

// =========================================================================
// Helpers
// =========================================================================

// hasIndustryKeyword checks if any industry keyword appears as a whole word.
func hasIndustryKeyword(companyName string, keywords []string) bool {
	words := strings.Fields(strings.ToLower(companyName))
	for _, w := range words {
		cleaned := strings.TrimRight(w, ".,;:()")
		for _, kw := range keywords {
			if cleaned == kw {
				return true
			}
		}
	}
	return false
}

// hasDistinctiveWordOverlap checks that at least one non-generic word from
// the business name appears in the CH company name.
func hasDistinctiveWordOverlap(businessName, companyName string, profile *CHVerticalProfile) bool {
	bizWords := strings.Fields(strings.ToLower(businessName))
	chNameLower := strings.ToLower(companyName)

	for _, word := range bizWords {
		word = strings.TrimRight(word, ".,;:()'-")
		word = strings.TrimLeft(word, "('")

		if len(word) < 3 {
			continue
		}
		if profile.GenericWords[word] {
			continue
		}

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
