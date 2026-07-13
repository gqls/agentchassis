// FILE: platform/orchestration/actions/ch_detail_fetch_action.go
// Fetches detailed company data (profile, officers, PSC) from the CH API
// for confirmed matches in ch_vet_companies, and stores the enrichment
// in companies_house_data with derived succession risk signals.
//
// Reuses existing helpers from companies_house_actions.go:
//   chAPIGet, extractOfficersList, extractPSCList,
//   deriveOwnerSignals, deriveSuccessionRisk, nullStr, nullInt
//
// Actions:
//   - ch_detail_fetch: Fetch and store details for confirmed matches

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// chMatchedCompany holds a confirmed match ready for detail fetch
type chMatchedCompany struct {
	CompanyNumber   string
	CompanyName     string
	BusinessID      string
	MatchConfidence float64
	MatchMethod     string
}

// CHDetailFetchAction fetches officers/PSC for confirmed matches and stores
// the enrichment data. Processes a batch per invocation with rate limiting.
func CHDetailFetchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CHDetailFetchAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	batchSize := 50
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	delayMs := 15000 // 15s between API calls — ~2 companies/min, 3 calls each
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}

	fetchOfficers := true
	if fo, ok := config["fetch_officers"].(bool); ok {
		fetchOfficers = fo
	}
	fetchPSC := true
	if fp, ok := config["fetch_psc"].(bool); ok {
		fetchPSC = fp
	}

	apiKey := os.Getenv("COMPANIES_HOUSE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("COMPANIES_HOUSE_API_KEY not set")
	}

	// Load confirmed matches not yet fetched
	companies, err := loadUnfetchedMatches(ctx, params.DB, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load unfetched matches: %w", err)
	}

	if len(companies) == 0 {
		params.Logger.Info("CHDetailFetch: no unfetched matches")
		return map[string]interface{}{
			"status":          "complete",
			"fetched":         0,
			"failed":          0,
			"total_remaining": 0,
		}, nil
	}

	params.Logger.Info("CHDetailFetch: loaded companies",
		zap.Int("count", len(companies)))

	totalFetched := 0
	totalFailed := 0

	for _, co := range companies {
		select {
		case <-ctx.Done():
			params.Logger.Warn("CHDetailFetch: context cancelled",
				zap.Int("fetched", totalFetched))
			break
		default:
		}

		err := fetchAndStoreCompanyDetails(ctx, params.DB, apiKey, co, fetchOfficers, fetchPSC, delayMs, params.Logger)
		if err != nil {
			params.Logger.Warn("CHDetailFetch: failed",
				zap.String("company", co.CompanyName),
				zap.String("company_number", co.CompanyNumber),
				zap.Error(err))
			totalFailed++
		} else {
			totalFetched++
		}
	}

	// Count remaining
	var remaining int
	row := params.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM business_intel.ch_vet_companies
		WHERE matched_business_id IS NOT NULL
		  AND details_fetched = false
		  AND match_method NOT IN ('pending_llm_review', 'llm_uncertain')`)
	_ = row.Scan(&remaining)

	// Notify scheduler
	taskName := "ch-detail-fetch"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`,
		taskName)

	params.Logger.Info("CHDetailFetch: complete",
		zap.Int("fetched", totalFetched),
		zap.Int("failed", totalFailed),
		zap.Int("remaining", remaining))

	return map[string]interface{}{
		"status":          "complete",
		"fetched":         totalFetched,
		"failed":          totalFailed,
		"total_remaining": remaining,
	}, nil
}

// loadUnfetchedMatches loads confirmed matches that haven't had details fetched yet.
// Excludes pending_llm_review and llm_uncertain — only fetch for confident matches.
func loadUnfetchedMatches(ctx context.Context, db *sql.DB, limit int) ([]chMatchedCompany, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ch.company_number, ch.company_name, ch.matched_business_id::text,
			   COALESCE(ch.match_confidence, 0), COALESCE(ch.match_method, '')
		FROM business_intel.ch_vet_companies ch
		WHERE ch.matched_business_id IS NOT NULL
		  AND ch.details_fetched = false
		  AND ch.match_method NOT IN ('pending_llm_review', 'llm_uncertain')
		ORDER BY ch.match_confidence DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []chMatchedCompany
	for rows.Next() {
		var co chMatchedCompany
		if err := rows.Scan(&co.CompanyNumber, &co.CompanyName, &co.BusinessID,
			&co.MatchConfidence, &co.MatchMethod); err != nil {
			continue
		}
		companies = append(companies, co)
	}
	return companies, rows.Err()
}

// fetchAndStoreCompanyDetails fetches profile/officers/PSC from CH API,
// derives succession signals, stores in companies_house_data, and marks
// the ch_vet_companies row as fetched.
// Reuses chAPIGet, extractOfficersList, extractPSCList, deriveOwnerSignals,
// deriveSuccessionRisk from companies_house_actions.go.
func fetchAndStoreCompanyDetails(
	ctx context.Context,
	db *sql.DB,
	apiKey string,
	co chMatchedCompany,
	fetchOfficers, fetchPSC bool,
	delayMs int,
	logger *zap.Logger,
) error {
	// 1. Company profile
	profile, err := chAPIGet(ctx, apiKey, fmt.Sprintf("/company/%s", co.CompanyNumber))
	if err != nil {
		return fmt.Errorf("fetch profile: %w", err)
	}

	// 2. Officers
	var officersData map[string]interface{}
	if fetchOfficers {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		officers, err := chAPIGet(ctx, apiKey,
			fmt.Sprintf("/company/%s/officers?items_per_page=50", co.CompanyNumber))
		if err != nil {
			logger.Warn("CHDetailFetch: officers fetch failed, continuing",
				zap.String("company_number", co.CompanyNumber),
				zap.Error(err))
			officersData = map[string]interface{}{"items": []interface{}{}}
		} else {
			officersData = officers
		}
	}

	// 3. PSC
	var pscData map[string]interface{}
	if fetchPSC {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		psc, err := chAPIGet(ctx, apiKey,
			fmt.Sprintf("/company/%s/persons-with-significant-control", co.CompanyNumber))
		if err != nil {
			logger.Warn("CHDetailFetch: PSC fetch failed, continuing",
				zap.String("company_number", co.CompanyNumber),
				zap.Error(err))
			pscData = map[string]interface{}{"items": []interface{}{}}
		} else {
			pscData = psc
		}
	}

	// Extract fields from profile
	companyStatus, _ := profile["company_status"].(string)
	companyType, _ := profile["company_type"].(string)
	dateOfCreation, _ := profile["date_of_creation"].(string)

	var sicCodes []string
	if sics, ok := profile["sic_codes"].([]interface{}); ok {
		for _, s := range sics {
			if str, ok := s.(string); ok {
				sicCodes = append(sicCodes, str)
			}
		}
	}

	registeredAddr, _ := json.Marshal(profile["registered_office_address"])

	// Derive signals using existing helpers
	officersList := extractOfficersList(officersData)
	officersJSON, _ := json.Marshal(officersList)

	pscList := extractPSCList(pscData)
	pscJSON, _ := json.Marshal(pscList)

	owner := deriveOwnerSignals(officersData, pscData)
	successionRisk := deriveSuccessionRisk(owner)

	var incorporationDate sql.NullTime
	if dateOfCreation != "" {
		if t, err := time.Parse("2006-01-02", dateOfCreation); err == nil {
			incorporationDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	rawResponse := map[string]interface{}{
		"profile": profile,
	}
	rawJSON, _ := json.Marshal(rawResponse)

	// Store in companies_house_data — same schema as StoreCHEnrichmentAction
	_, err = db.ExecContext(ctx, `
		INSERT INTO business_intel.companies_house_data (
			business_id, company_number, company_name, company_status,
			company_type, incorporation_date, sic_codes, registered_address,
			officers, psc,
			parent_company_name, parent_company_number,
			owner_name, owner_dob_year, owner_dob_month, owner_estimated_age,
			owner_appointment_date, owner_tenure_years,
			is_sole_director, is_corporate_owned, succession_risk,
			match_confidence, match_method,
			raw_response
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24
		)
		ON CONFLICT (business_id) DO UPDATE SET
			company_number = EXCLUDED.company_number,
			company_name = EXCLUDED.company_name,
			company_status = EXCLUDED.company_status,
			company_type = EXCLUDED.company_type,
			incorporation_date = EXCLUDED.incorporation_date,
			sic_codes = EXCLUDED.sic_codes,
			registered_address = EXCLUDED.registered_address,
			officers = EXCLUDED.officers,
			psc = EXCLUDED.psc,
			parent_company_name = EXCLUDED.parent_company_name,
			parent_company_number = EXCLUDED.parent_company_number,
			owner_name = EXCLUDED.owner_name,
			owner_dob_year = EXCLUDED.owner_dob_year,
			owner_dob_month = EXCLUDED.owner_dob_month,
			owner_estimated_age = EXCLUDED.owner_estimated_age,
			owner_appointment_date = EXCLUDED.owner_appointment_date,
			owner_tenure_years = EXCLUDED.owner_tenure_years,
			is_sole_director = EXCLUDED.is_sole_director,
			is_corporate_owned = EXCLUDED.is_corporate_owned,
			succession_risk = EXCLUDED.succession_risk,
			match_confidence = EXCLUDED.match_confidence,
			match_method = EXCLUDED.match_method,
			raw_response = EXCLUDED.raw_response,
			enriched_at = NOW()`,
		co.BusinessID,                  // $1
		co.CompanyNumber,               // $2
		co.CompanyName,                 // $3
		companyStatus,                  // $4
		companyType,                    // $5
		incorporationDate,              // $6
		pgArrayFromInterface(sicCodes), // $7
		registeredAddr,                 // $8
		officersJSON,                   // $9
		pscJSON,                        // $10
		owner.ParentCompanyName,        // $11
		owner.ParentCompanyNumber,      // $12
		nullStr(owner.Name),            // $13
		nullInt(owner.DOBYear),         // $14
		nullInt(owner.DOBMonth),        // $15
		nullInt(owner.EstimatedAge),    // $16
		owner.AppointmentDate,          // $17
		nullInt(owner.TenureYears),     // $18
		owner.IsSoleDirector,           // $19
		owner.IsCorporateOwned,         // $20
		successionRisk,                 // $21
		co.MatchConfidence,             // $22
		co.MatchMethod,                 // $23
		rawJSON,                        // $24
	)
	if err != nil {
		return fmt.Errorf("store enrichment: %w", err)
	}

	// Mark as fetched in ch_vet_companies
	_, err = db.ExecContext(ctx, `
		UPDATE business_intel.ch_vet_companies
		SET details_fetched = true,
			details_fetched_at = NOW(),
			updated_at = NOW()
		WHERE company_number = $1`,
		co.CompanyNumber)
	if err != nil {
		logger.Warn("CHDetailFetch: failed to mark as fetched",
			zap.String("company_number", co.CompanyNumber),
			zap.Error(err))
	}

	logger.Info("CHDetailFetch: stored",
		zap.String("business_id", co.BusinessID),
		zap.String("company_number", co.CompanyNumber),
		zap.String("company_name", co.CompanyName),
		zap.String("succession_risk", successionRisk),
		zap.Int("owner_age", owner.EstimatedAge))

	return nil
}
