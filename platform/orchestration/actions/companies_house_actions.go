// FILE: platform/orchestration/actions/companies_house_actions.go
// Actions for Companies House enrichment pipeline.
// These enrich verified businesses with financial data, officers, ownership,
// and succession risk signals from the Companies House API.
//
// Actions:
//   - load_ch_enrichment_batch:  Load verified businesses not yet enriched
//   - companies_house_search:    Search CH by name, match by postcode + SIC
//   - companies_house_fetch:     Fetch company profile, officers, PSC
//   - store_ch_enrichment:       Store CH data and derive succession signals

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

const chAPIBase = "https://api.company-information.service.gov.uk"

// ---------------------------------------------------------------------------
// load_ch_enrichment_batch
// ---------------------------------------------------------------------------
// Loads verified businesses that don't yet have a row in companies_house_data.

func LoadCHEnrichmentBatchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadCHEnrichmentBatchAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	batchSize := 20
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}
	// Allow input_data override
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if bs, ok := inputData["batch_size"].(float64); ok && bs > 0 {
			batchSize = int(bs)
		}
	}

	verticalSlug, _ := config["vertical_slug"].(string)
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if vs, ok := inputData["vertical_slug"].(string); ok && vs != "" {
			verticalSlug = vs
		}
	}

	query := `
		SELECT b.id, b.name, b.postcode, b.town, b.website_url, b.group_name
		FROM business_intel.businesses b
		JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
		LEFT JOIN business_intel.companies_house_data ch ON ch.business_id = b.id
		WHERE b.verification_status = 'verified'
		  AND ch.business_id IS NULL
	`
	args := []interface{}{}
	argIdx := 1

	if verticalSlug != "" {
		query += fmt.Sprintf(" AND bv.slug = $%d", argIdx)
		args = append(args, verticalSlug)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY b.updated_at ASC LIMIT $%d", argIdx)
	args = append(args, batchSize)

	rows, err := params.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query unenriched businesses: %w", err)
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, name string
		var postcode, town, websiteURL, groupName sql.NullString

		if err := rows.Scan(&id, &name, &postcode, &town, &websiteURL, &groupName); err != nil {
			params.Logger.Warn("LoadCHEnrichmentBatch: scan error", zap.Error(err))
			continue
		}

		item := map[string]interface{}{
			"id":          id,
			"name":        name,
			"postcode":    postcode.String,
			"town":        town.String,
			"website_url": websiteURL.String,
			"group_name":  groupName.String,
		}
		items = append(items, item)
	}

	params.Logger.Info("LoadCHEnrichmentBatchAction: Loaded batch",
		zap.Int("count", len(items)),
		zap.Int("batch_size", batchSize),
		zap.String("vertical", verticalSlug))

	return map[string]interface{}{
		"items": items,
		"count": len(items),
	}, nil
}

// ---------------------------------------------------------------------------
// companies_house_search
// ---------------------------------------------------------------------------
// Searches Companies House by business name, scores matches by postcode
// proximity and SIC code. Returns the best match or "no match".

func CompaniesHouseSearchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompaniesHouseSearchAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Extract current_business using ExtractFields (handles loop variables correctly).
	// ExtractNestedField resolves to stale values during loops because it searches
	// deeply nested collected_data paths and can find a previous iteration's data.
	inputFields := []string{"current_business"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	business, _ := extracted["current_business"].(map[string]interface{})

	// Read name and postcode from the extracted business object
	nameKey := "name"
	if nf, ok := config["name_field"].(string); ok && nf != "" {
		// Support dotted paths like "current_business.name" — take the last part
		parts := strings.Split(nf, ".")
		nameKey = parts[len(parts)-1]
	}
	pcKey := "postcode"
	if pf, ok := config["postcode_field"].(string); ok && pf != "" {
		parts := strings.Split(pf, ".")
		pcKey = parts[len(parts)-1]
	}

	name, _ := business[nameKey].(string)
	postcode, _ := business[pcKey].(string)

	params.Logger.Info("CompaniesHouseSearch: extracted business data",
		zap.String("name", name),
		zap.String("postcode", postcode),
		zap.String("business_id", fmt.Sprintf("%v", business["id"])),
	)

	if name == "" {
		params.Logger.Warn("CompaniesHouseSearch: no business name")
		return map[string]interface{}{
			"matched": false,
			"reason":  "no business name provided",
		}, nil
	}

	// Rate limiting delay — be polite to Companies House
	delayMs := 15000
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}
	params.Logger.Info("CompaniesHouseSearch: waiting before API call",
		zap.Int("delay_ms", delayMs))
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// SIC code filter
	var sicFilter []string
	if sics, ok := config["sic_filter"].([]interface{}); ok {
		for _, s := range sics {
			if str, ok := s.(string); ok {
				sicFilter = append(sicFilter, str)
			}
		}
	}

	// Clean up the search name — remove common suffixes that hurt search
	searchName := cleanCompanySearchName(name)

	// Search Companies House API
	apiKey := os.Getenv("COMPANIES_HOUSE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("COMPANIES_HOUSE_API_KEY not set")
	}

	searchURL := fmt.Sprintf("%s/search/companies?q=%s&items_per_page=10",
		chAPIBase, url.QueryEscape(searchName))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(apiKey, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CH search API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CH search API returned %d: %s", resp.StatusCode, string(body))
	}

	var searchResp chSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse CH search response: %w", err)
	}

	params.Logger.Info("CompaniesHouseSearch: got results",
		zap.String("query", searchName),
		zap.Int("results", len(searchResp.Items)))

	// Score and rank matches
	bestMatch := scoreCHMatches(searchResp.Items, name, postcode, sicFilter, params.Logger)

	if bestMatch == nil {
		return map[string]interface{}{
			"matched":       false,
			"search_query":  searchName,
			"results_count": len(searchResp.Items),
		}, nil
	}

	params.Logger.Info("CompaniesHouseSearch: matched",
		zap.String("company_number", bestMatch.CompanyNumber),
		zap.String("company_name", bestMatch.Title),
		zap.Float64("score", bestMatch.Score),
		zap.String("method", bestMatch.Method))

	return map[string]interface{}{
		"matched":          true,
		"company_number":   bestMatch.CompanyNumber,
		"company_name":     bestMatch.Title,
		"company_status":   bestMatch.CompanyStatus,
		"company_type":     bestMatch.CompanyType,
		"sic_codes":        bestMatch.SICCodes,
		"date_of_creation": bestMatch.DateOfCreation,
		"match_confidence": bestMatch.Score,
		"match_method":     bestMatch.Method,
		"search_query":     searchName,
		"results_count":    len(searchResp.Items),
	}, nil
}

// ---------------------------------------------------------------------------
// companies_house_fetch
// ---------------------------------------------------------------------------
// Fetches company profile, officers, and PSC for a matched company.

func CompaniesHouseFetchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompaniesHouseFetchAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Extract ch_search using ExtractFields (handles loop variables correctly)
	inputFields := []string{"ch_search"}
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	chSearch, _ := extracted["ch_search"].(map[string]interface{})
	companyNumber, _ := chSearch["company_number"].(string)

	if companyNumber == "" {
		return nil, fmt.Errorf("company_number is required")
	}

	fetchOfficers := true
	if fo, ok := config["fetch_officers"].(bool); ok {
		fetchOfficers = fo
	}
	fetchPSC := true
	if fp, ok := config["fetch_psc"].(bool); ok {
		fetchPSC = fp
	}

	delayMs := 15000
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}

	apiKey := os.Getenv("COMPANIES_HOUSE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("COMPANIES_HOUSE_API_KEY not set")
	}

	result := map[string]interface{}{}

	// 1. Company profile
	params.Logger.Info("CompaniesHouseFetch: fetching profile",
		zap.String("company_number", companyNumber))

	profile, err := chAPIGet(ctx, apiKey, fmt.Sprintf("/company/%s", companyNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch company profile: %w", err)
	}
	result["profile"] = profile

	// 2. Officers (directors, secretaries)
	if fetchOfficers {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		params.Logger.Info("CompaniesHouseFetch: fetching officers")
		officers, err := chAPIGet(ctx, apiKey,
			fmt.Sprintf("/company/%s/officers?items_per_page=50", companyNumber))
		if err != nil {
			params.Logger.Warn("CompaniesHouseFetch: failed to fetch officers", zap.Error(err))
			result["officers"] = map[string]interface{}{"items": []interface{}{}}
		} else {
			result["officers"] = officers
		}
	}

	// 3. Persons of Significant Control
	if fetchPSC {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		params.Logger.Info("CompaniesHouseFetch: fetching PSC")
		psc, err := chAPIGet(ctx, apiKey,
			fmt.Sprintf("/company/%s/persons-with-significant-control", companyNumber))
		if err != nil {
			params.Logger.Warn("CompaniesHouseFetch: failed to fetch PSC", zap.Error(err))
			result["psc"] = map[string]interface{}{"items": []interface{}{}}
		} else {
			result["psc"] = psc
		}
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// store_ch_enrichment
// ---------------------------------------------------------------------------
// Stores Companies House data and derives ownership/succession signals.

func StoreCHEnrichmentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StoreCHEnrichmentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config
	inputFields := []string{"current_business", "ch_search", "ch_details"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	business, _ := extracted["current_business"].(map[string]interface{})
	businessID, _ := business["id"].(string)
	if businessID == "" {
		return nil, fmt.Errorf("business id is required")
	}

	searchResult, _ := extracted["ch_search"].(map[string]interface{})

	// Check if this is a no-match store
	noMatch, _ := config["no_match"].(bool)
	matched, _ := searchResult["matched"].(bool)

	if noMatch || !matched {
		searchQuery, _ := searchResult["search_query"].(string)
		resultsCount := 0
		if rc, ok := searchResult["results_count"].(float64); ok {
			resultsCount = int(rc)
		}

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO business_intel.companies_house_data
				(business_id, match_confidence, match_method, search_query,
				 search_results_count, succession_risk)
			VALUES ($1, 0, 'no_match', $2, $3, 'unknown')
			ON CONFLICT (business_id) DO UPDATE SET
				match_confidence = 0,
				match_method = 'no_match',
				enriched_at = NOW()`,
			businessID, searchQuery, resultsCount)
		if err != nil {
			return nil, fmt.Errorf("failed to store no-match: %w", err)
		}

		params.Logger.Info("StoreCHEnrichment: stored no-match",
			zap.String("business_id", businessID))

		return map[string]interface{}{
			"stored":  true,
			"matched": false,
		}, nil
	}

	// We have a match — extract full details
	details, _ := extracted["ch_details"].(map[string]interface{})
	profile, _ := details["profile"].(map[string]interface{})
	officersData, _ := details["officers"].(map[string]interface{})
	pscData, _ := details["psc"].(map[string]interface{})

	// Extract company fields from profile
	companyNumber, _ := searchResult["company_number"].(string)
	companyName, _ := searchResult["company_name"].(string)
	companyStatus, _ := searchResult["company_status"].(string)
	companyType, _ := searchResult["company_type"].(string)
	dateOfCreation, _ := searchResult["date_of_creation"].(string)

	matchConfidence := 0.0
	if mc, ok := searchResult["match_confidence"].(float64); ok {
		matchConfidence = mc
	}
	matchMethod, _ := searchResult["match_method"].(string)
	searchQuery, _ := searchResult["search_query"].(string)
	resultsCount := 0
	if rc, ok := searchResult["results_count"].(float64); ok {
		resultsCount = int(rc)
	}

	// SIC codes
	var sicCodes []string
	if sics, ok := profile["sic_codes"].([]interface{}); ok {
		for _, s := range sics {
			if str, ok := s.(string); ok {
				sicCodes = append(sicCodes, str)
			}
		}
	}

	// Registered address
	registeredAddr, _ := json.Marshal(profile["registered_office_address"])

	// Build officers JSON
	officersList := extractOfficersList(officersData)
	officersJSON, _ := json.Marshal(officersList)

	// Build PSC JSON
	pscList := extractPSCList(pscData)
	pscJSON, _ := json.Marshal(pscList)

	// Derive owner signals
	owner := deriveOwnerSignals(officersData, pscData)

	// Derive succession risk
	successionRisk := deriveSuccessionRisk(owner)

	// Parse incorporation date
	var incorporationDate sql.NullTime
	if dateOfCreation != "" {
		if t, err := time.Parse("2006-01-02", dateOfCreation); err == nil {
			incorporationDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Store raw response for debugging
	rawResponse := map[string]interface{}{
		"search":  searchResult,
		"profile": profile,
	}
	rawJSON, _ := json.Marshal(rawResponse)

	_, err := params.DB.ExecContext(ctx, `
		INSERT INTO business_intel.companies_house_data (
			business_id, company_number, company_name, company_status,
			company_type, incorporation_date, sic_codes, registered_address,
			officers, psc,
			parent_company_name, parent_company_number,
			owner_name, owner_dob_year, owner_dob_month, owner_estimated_age,
			owner_appointment_date, owner_tenure_years,
			is_sole_director, is_corporate_owned, succession_risk,
			match_confidence, match_method, search_query, search_results_count,
			raw_response
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26
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
		businessID,                     // $1
		companyNumber,                  // $2
		companyName,                    // $3
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
		matchConfidence,                // $22
		matchMethod,                    // $23
		searchQuery,                    // $24
		resultsCount,                   // $25
		rawJSON,                        // $26
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store CH enrichment: %w", err)
	}

	params.Logger.Info("StoreCHEnrichment: stored",
		zap.String("business_id", businessID),
		zap.String("company_number", companyNumber),
		zap.String("succession_risk", successionRisk),
		zap.Int("owner_age", owner.EstimatedAge))

	return map[string]interface{}{
		"stored":          true,
		"matched":         true,
		"company_number":  companyNumber,
		"company_name":    companyName,
		"succession_risk": successionRisk,
		"owner_age":       owner.EstimatedAge,
		"owner_name":      owner.Name,
	}, nil
}

// ===========================================================================
// Helper types and functions
// ===========================================================================

type chSearchResponse struct {
	Items []chSearchItem `json:"items"`
}

type chSearchItem struct {
	CompanyNumber           string   `json:"company_number"`
	Title                   string   `json:"title"`
	CompanyStatus           string   `json:"company_status"`
	CompanyType             string   `json:"company_type"`
	DateOfCreation          string   `json:"date_of_creation"`
	SICCodes                []string `json:"sic_codes"`
	RegisteredOfficeAddress struct {
		PostalCode   string `json:"postal_code"`
		Locality     string `json:"locality"`
		AddressLine1 string `json:"address_line_1"`
	} `json:"registered_office_address"`
}

type scoredMatch struct {
	chSearchItem
	Score  float64
	Method string
}

type ownerSignals struct {
	Name                string
	DOBYear             int
	DOBMonth            int
	EstimatedAge        int
	AppointmentDate     sql.NullTime
	TenureYears         int
	IsSoleDirector      bool
	IsCorporateOwned    bool
	ParentCompanyName   string
	ParentCompanyNumber string
}

// chAPIGet makes a GET request to the Companies House API
func chAPIGet(ctx context.Context, apiKey, path string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", chAPIBase+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiKey, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return map[string]interface{}{}, nil
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("CH API rate limited (429)")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CH API returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// cleanCompanySearchName strips common suffixes that hurt CH search accuracy
func cleanCompanySearchName(name string) string {
	name = strings.TrimSpace(name)
	original := name

	// Remove common trailing descriptions
	for _, suffix := range []string{
		" - Veterinary Practice",
		" Veterinary Surgery",
		" Veterinary Centre",
		" Veterinary Clinic",
		" Veterinary Hospital",
		" Veterinary Group",
		" Vets",
		" Ltd",
		" Limited",
	} {
		if strings.HasSuffix(strings.ToLower(name), strings.ToLower(suffix)) {
			name = strings.TrimSpace(name[:len(name)-len(suffix)])
		}
	}

	// If stripped name is too short to be a useful search query,
	// fall back to the original. Single-word queries like "Erne" or "Cave"
	// return too many irrelevant results from Companies House.
	if len(name) < 6 {
		return original
	}
	return name
}

// scoreCHMatches scores search results and returns the best match
func scoreCHMatches(items []chSearchItem, businessName, postcode string, sicFilter []string, logger *zap.Logger) *scoredMatch {
	if len(items) == 0 {
		return nil
	}

	postcodePrefix := ""
	if len(postcode) >= 2 {
		// Extract outward code (e.g. "BH10" from "BH10 4AQ", "SW1A" from "SW1A 1AA")
		parts := strings.Fields(strings.ToUpper(postcode))
		if len(parts) > 0 {
			postcodePrefix = parts[0]
		}
	}

	businessNameLower := strings.ToLower(businessName)
	var best *scoredMatch

	for _, item := range items {
		score := 0.0
		method := "name_search"

		// Score: postcode prefix match (+0.35)
		if postcodePrefix != "" && item.RegisteredOfficeAddress.PostalCode != "" {
			regPC := strings.ToUpper(item.RegisteredOfficeAddress.PostalCode)
			regParts := strings.Fields(regPC)
			if len(regParts) > 0 && regParts[0] == postcodePrefix {
				score += 0.35
				method = "name_postcode"
			}
		}

		// Score: SIC code match (+0.25)
		hasSICMatch := false
		if len(sicFilter) > 0 {
			for _, itemSIC := range item.SICCodes {
				for _, filterSIC := range sicFilter {
					if itemSIC == filterSIC {
						hasSICMatch = true
						score += 0.25
						break
					}
				}
				if hasSICMatch {
					break
				}
			}
		}

		// Score: company is active (+0.15)
		if item.CompanyStatus == "active" {
			score += 0.15
		}

		// Score: name similarity (+0.25)
		titleLower := strings.ToLower(item.Title)
		if titleLower == businessNameLower {
			score += 0.25
		} else if strings.Contains(titleLower, businessNameLower) || strings.Contains(businessNameLower, titleLower) {
			score += 0.15
		} else {
			// Check if key words overlap
			businessWords := strings.Fields(businessNameLower)
			matchedWords := 0
			for _, word := range businessWords {
				if len(word) > 3 && strings.Contains(titleLower, word) {
					matchedWords++
				}
			}
			if len(businessWords) > 0 {
				wordRatio := float64(matchedWords) / float64(len(businessWords))
				score += wordRatio * 0.2
			}
		}

		// If SIC filter provided but no match, penalise heavily
		if len(sicFilter) > 0 && !hasSICMatch {
			score -= 0.3
		}

		logger.Info("CompaniesHouseSearch: scored result",
			zap.String("company", item.Title),
			zap.String("number", item.CompanyNumber),
			zap.String("status", item.CompanyStatus),
			zap.Strings("sic", item.SICCodes),
			zap.String("reg_postcode", item.RegisteredOfficeAddress.PostalCode),
			zap.Float64("score", score))

		if score > 0.4 && (best == nil || score > best.Score) {
			best = &scoredMatch{
				chSearchItem: item,
				Score:        score,
				Method:       method,
			}
		}
	}

	return best
}

// extractOfficersList pulls the officers array from the API response
func extractOfficersList(officersData map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	items, ok := officersData["items"].([]interface{})
	if !ok {
		return result
	}

	for _, item := range items {
		officer, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entry := map[string]interface{}{}

		// Copy safe fields
		for _, field := range []string{"name", "officer_role", "appointed_on", "resigned_on", "nationality", "occupation"} {
			if v, ok := officer[field]; ok {
				entry[field] = v
			}
		}

		// DOB (month/year only — this is what CH discloses)
		if dob, ok := officer["date_of_birth"].(map[string]interface{}); ok {
			entry["dob_month"] = dob["month"]
			entry["dob_year"] = dob["year"]
		}

		result = append(result, entry)
	}

	return result
}

// extractPSCList pulls the PSC array from the API response
func extractPSCList(pscData map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	items, ok := pscData["items"].([]interface{})
	if !ok {
		return result
	}

	for _, item := range items {
		psc, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entry := map[string]interface{}{}

		for _, field := range []string{"name", "kind", "notified_on", "ceased_on", "natures_of_control"} {
			if v, ok := psc[field]; ok {
				entry[field] = v
			}
		}

		// Name elements for individuals
		if nameElements, ok := psc["name_elements"].(map[string]interface{}); ok {
			entry["name_elements"] = nameElements
		}

		// Identification for corporate entities
		if ident, ok := psc["identification"].(map[string]interface{}); ok {
			entry["identification"] = ident
		}

		result = append(result, entry)
	}

	return result
}

// deriveOwnerSignals extracts the primary owner/director and their age signals
func deriveOwnerSignals(officersData, pscData map[string]interface{}) ownerSignals {
	var owner ownerSignals

	// Check officers for active directors
	if items, ok := officersData["items"].([]interface{}); ok {
		activeDirectors := 0
		var earliestAppointment time.Time
		var primaryDirector map[string]interface{}

		for _, item := range items {
			officer, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			role, _ := officer["officer_role"].(string)
			resignedOn, _ := officer["resigned_on"].(string)

			if role != "director" || resignedOn != "" {
				continue
			}
			activeDirectors++

			// Track earliest appointment
			if appointed, _ := officer["appointed_on"].(string); appointed != "" {
				t, err := time.Parse("2006-01-02", appointed)
				if err == nil && (earliestAppointment.IsZero() || t.Before(earliestAppointment)) {
					earliestAppointment = t
					primaryDirector = officer
				}
			}
		}

		owner.IsSoleDirector = activeDirectors == 1

		// Extract DOB and name from the longest-serving director
		if primaryDirector != nil {
			owner.Name, _ = primaryDirector["name"].(string)

			if dob, ok := primaryDirector["date_of_birth"].(map[string]interface{}); ok {
				if year, ok := dob["year"].(float64); ok && int(year) > 0 {
					owner.DOBYear = int(year)
					owner.EstimatedAge = time.Now().Year() - int(year)
				}
				if month, ok := dob["month"].(float64); ok {
					owner.DOBMonth = int(month)
				}
			}

			if !earliestAppointment.IsZero() {
				owner.AppointmentDate = sql.NullTime{Time: earliestAppointment, Valid: true}
				owner.TenureYears = int(time.Since(earliestAppointment).Hours() / 8760)
			}
		}
	}

	// Check PSC for corporate ownership
	if items, ok := pscData["items"].([]interface{}); ok {
		for _, item := range items {
			psc, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			kind, _ := psc["kind"].(string)
			ceasedOn, _ := psc["ceased_on"].(string)

			if ceasedOn != "" {
				continue // skip ceased PSCs
			}

			if strings.Contains(kind, "corporate") {
				owner.IsCorporateOwned = true
				owner.ParentCompanyName, _ = psc["name"].(string)
				if ident, ok := psc["identification"].(map[string]interface{}); ok {
					owner.ParentCompanyNumber, _ = ident["registration_number"].(string)
				}
				break
			}
		}
	}

	return owner
}

// deriveSuccessionRisk scores the likelihood the owner is approaching retirement
func deriveSuccessionRisk(owner ownerSignals) string {
	if owner.IsCorporateOwned {
		return "acquired"
	}

	score := 0

	// Age signals
	if owner.EstimatedAge >= 60 {
		score += 3
	} else if owner.EstimatedAge >= 55 {
		score += 2
	} else if owner.EstimatedAge >= 50 {
		score += 1
	}

	// Tenure signals
	if owner.TenureYears >= 25 {
		score += 2
	} else if owner.TenureYears >= 15 {
		score += 1
	}

	// Sole director = no succession plan
	if owner.IsSoleDirector {
		score += 1
	}

	switch {
	case score >= 5:
		return "high"
	case score >= 3:
		return "medium"
	case score >= 1:
		return "low"
	default:
		return "unknown"
	}
}

// nullStr returns sql.NullString
func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullInt returns sql.NullInt32 for zero = null
func nullInt(i int) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(i), Valid: true}
}
