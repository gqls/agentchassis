// FILE: platform/orchestration/actions/ch_fetch_accounts_action.go
// Fetches and parses Companies House filed accounts (iXBRL) for enriched companies.
// Extracts financial data: net assets, total assets, employee count, turnover, profit/loss.
//
// Flow per company:
//   1. GET /company/{number}/filing-history?category=accounts&items_per_page=1
//   2. GET {document_metadata_url} → get content URL and available formats
//   3. GET {content_url} with Accept: application/xhtml+xml (follows redirects to S3)
//   4. Parse ix:nonFraction elements via regex
//   5. Update companies_house_data with financial fields
//
// Reuses chAPIGet for the filing history call. Document metadata and content
// downloads use a redirect-following HTTP client (CH document API redirects to S3).
//
// Actions:
//   - ch_fetch_accounts: Batch fetch and parse filed accounts

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const chDocumentAPIBase = "https://document-api.company-information.service.gov.uk"

// iXBRL tag suffixes mapped to DB fields.
// Multiple tag names can map to the same field — first match wins.
// Context: CY_END = current year-end (balance sheet instant), CY = current year period.
// Tag names vary between FRS-102, FRS-105, and older UK GAAP taxonomies.
var ixbrlFieldMappings = []struct {
	TagNames  []string // suffixes to match, e.g. "NetAssetsLiabilities"
	Context   string   // "CY_END" or "CY"
	DBField   string   // column in companies_house_data
	FieldType string   // "money" or "integer"
}{
	{[]string{
		"NetAssetsLiabilities",
		"NetCurrentAssetsLiabilities",
	}, "CY_END", "net_worth_gbp", "money"},

	{[]string{
		"TotalAssetsLessCurrentLiabilities",
	}, "CY_END", "total_assets_gbp", "money"},

	{[]string{
		"AverageNumberEmployeesDuringPeriod",
		"EmployeesTotal",
	}, "CY", "employee_count", "integer"},

	{[]string{
		"TurnoverRevenue",
		"TurnoverGrossIncome",
		"Turnover",
	}, "CY", "turnover_gbp", "money"},

	{[]string{
		"ProfitLossForPeriod",
		"ProfitLossOnOrdinaryActivitiesBeforeTax",
		"ProfitLossForFinancialYear",
		"ProfitLossForYear",
		"ProfitLoss",
	}, "CY", "profit_loss_gbp", "money"},
}

// Regex to extract ix:nonFraction elements from iXBRL.
// Captures the full opening tag (group 1) and text content (group 2).
// We parse name, contextRef, and sign from the tag attributes separately
// because attribute order varies between filing software.
var ixbrlTagPattern = regexp.MustCompile(
	`<ix:nonFraction([^>]*)>([^<]*)</ix:nonFraction>`)

// Attribute extractors
var attrName = regexp.MustCompile(`name="([^"]*)"`)
var attrContextRef = regexp.MustCompile(`contextRef="([^"]*)"`)
var attrSign = regexp.MustCompile(`sign="([^"]*)"`)

// accountsCompany holds a company ready for accounts fetch
type accountsCompany struct {
	CompanyNumber string
	BusinessID    string
}

// chHTTPLogContext carries agent context for HTTP request logging.
// Passed into helper functions so they can log without knowing about ActionParams.
type chHTTPLogContext struct {
	DB              *sql.DB
	Logger          *zap.Logger
	AgentType       string
	AgentID         interface{}
	StepName        string
	OrchestrationID string
	CorrelationID   string
	ActionName      string
}

// CHFetchAccountsAction fetches and parses filed accounts for enriched companies.
func CHFetchAccountsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CHFetchAccountsAction: Starting")

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

	delayMs := 15000
	if d, ok := config["delay_ms"].(float64); ok && d > 0 {
		delayMs = int(d)
	}

	apiKey := ""
	if ak, ok := config["api_key_env_var"].(string); ok && ak != "" {
		apiKey = os.Getenv(ak)
	}
	if apiKey == "" {
		apiKey = os.Getenv("COMPANIES_HOUSE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("COMPANIES_HOUSE_API_KEY not set")
	}

	// Load companies with details fetched but accounts not yet fetched
	companies, err := loadCompaniesForAccountsFetch(ctx, params.DB, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load companies: %w", err)
	}

	if len(companies) == 0 {
		params.Logger.Info("CHFetchAccounts: no companies to process")
		return map[string]interface{}{
			"status": "complete", "fetched": 0, "parsed": 0, "failed": 0,
		}, nil
	}

	params.Logger.Info("CHFetchAccounts: loaded companies",
		zap.Int("count", len(companies)))

	// HTTP client with redirect following for document downloads
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Build logging context for HTTP request tracking
	logCtx := chHTTPLogContext{
		DB:              params.DB,
		Logger:          params.Logger,
		AgentType:       params.AgentType,
		AgentID:         params.Headers["agent_id"],
		StepName:        params.ExecutionContext.StepName,
		OrchestrationID: params.ExecutionContext.OrchestrationID,
		CorrelationID:   params.ExecutionContext.CorrelationID,
		ActionName:      "ch_fetch_accounts",
	}

	totalFetched := 0
	totalParsed := 0
	totalFailed := 0
	totalNoAccounts := 0

	for _, co := range companies {
		// Check context cancellation — break must exit the for loop, not just the select
		if ctx.Err() != nil {
			params.Logger.Info("CHFetchAccounts: context cancelled, stopping")
			break
		}

		result, err := fetchAndParseAccounts(ctx, apiKey, httpClient, co.CompanyNumber, delayMs, logCtx)
		if err != nil {
			params.Logger.Warn("CHFetchAccounts: failed",
				zap.String("company_number", co.CompanyNumber),
				zap.Error(err))
			totalFailed++
			// Mark as attempted so we don't retry every run
			_ = markAccountsFetched(ctx, params.DB, co.CompanyNumber, co.BusinessID, nil, "error: "+err.Error())
			continue
		}

		totalFetched++

		if result == nil {
			// No accounts filing found
			_ = markAccountsFetched(ctx, params.DB, co.CompanyNumber, co.BusinessID, nil, "no_accounts_filing")
			totalNoAccounts++
			continue
		}

		// Store financial data
		err = storeAccountsData(ctx, params.DB, co.BusinessID, result)
		if err != nil {
			params.Logger.Warn("CHFetchAccounts: failed to store",
				zap.String("company_number", co.CompanyNumber),
				zap.Error(err))
			totalFailed++
			continue
		}

		_ = markAccountsFetched(ctx, params.DB, co.CompanyNumber, co.BusinessID, result, "")
		totalParsed++

		params.Logger.Info("CHFetchAccounts: parsed",
			zap.String("company_number", co.CompanyNumber),
			zap.Any("net_worth", result["net_worth_gbp"]),
			zap.Any("employees", result["employee_count"]),
			zap.String("accounts_type", fmt.Sprintf("%v", result["accounts_type"])))
	}

	// Notify scheduler
	taskName := "ch-fetch-accounts"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)

	params.Logger.Info("CHFetchAccounts: complete",
		zap.Int("fetched", totalFetched),
		zap.Int("parsed", totalParsed),
		zap.Int("no_accounts", totalNoAccounts),
		zap.Int("failed", totalFailed))

	return map[string]interface{}{
		"status":      "complete",
		"fetched":     totalFetched,
		"parsed":      totalParsed,
		"no_accounts": totalNoAccounts,
		"failed":      totalFailed,
	}, nil
}

// loadCompaniesForAccountsFetch loads companies that have been detail-fetched
// but don't yet have accounts data.
func loadCompaniesForAccountsFetch(ctx context.Context, db *sql.DB, limit int) ([]accountsCompany, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ch.company_number, ch.matched_business_id::text
		FROM business_intel.ch_vet_companies ch
		JOIN business_intel.companies_house_data chd ON chd.business_id = ch.matched_business_id
		WHERE ch.details_fetched = true
		  AND ch.accounts_fetched = false
		  AND ch.match_method NOT IN ('pending_llm_review', 'llm_uncertain')
		ORDER BY ch.match_confidence DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []accountsCompany
	for rows.Next() {
		var co accountsCompany
		if err := rows.Scan(&co.CompanyNumber, &co.BusinessID); err != nil {
			continue
		}
		companies = append(companies, co)
	}
	return companies, rows.Err()
}

// fetchAndParseAccounts does the three-step fetch: filing history → metadata → iXBRL.
func fetchAndParseAccounts(ctx context.Context, apiKey string, httpClient *http.Client, companyNumber string, delayMs int, lc chHTTPLogContext) (map[string]interface{}, error) {

	meta := map[string]interface{}{"company_number": companyNumber}

	// Step 1: Get latest accounts filing
	filingURL := fmt.Sprintf("/company/%s/filing-history?category=accounts&items_per_page=1", companyNumber)
	fullFilingURL := "https://api.company-information.service.gov.uk" + filingURL

	callStart := time.Now()
	filingResp, err := chAPIGet(ctx, apiKey, filingURL)
	latencyMs := int(time.Since(callStart).Milliseconds())

	LogHTTPRequest(lc.DB, lc.Logger, HTTPRequestLogParams{
		AgentType: lc.AgentType, AgentID: lc.AgentID, StepName: lc.StepName,
		OrchestrationID: lc.OrchestrationID, CorrelationID: lc.CorrelationID,
		ActionName: lc.ActionName, Method: "GET", URL: fullFilingURL,
		StatusCode: httpStatusFromErr(err, 200), LatencyMs: latencyMs,
		Success: err == nil, ErrorMessage: errString(err), Metadata: meta,
	})

	if err != nil {
		return nil, fmt.Errorf("filing history: %w", err)
	}

	items, _ := filingResp["items"].([]interface{})
	if len(items) == 0 {
		return nil, nil // No accounts filed
	}

	filing, _ := items[0].(map[string]interface{})
	if filing == nil {
		return nil, nil
	}

	// Extract accounts date and type
	accountsDate, _ := filing["date"].(string)
	accountsType, _ := filing["description"].(string)

	// Get document metadata links
	links, _ := filing["links"].(map[string]interface{})
	documentMetaURL, _ := links["document_metadata"].(string)
	if documentMetaURL == "" {
		return nil, fmt.Errorf("no document_metadata URL in filing")
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Step 2: Get document metadata to find content URL.
	resolvedMetaURL := resolveDocumentURL(documentMetaURL)
	callStart = time.Now()
	metaResp, err := chDocumentAPIGet(ctx, httpClient, apiKey, resolvedMetaURL)
	latencyMs = int(time.Since(callStart).Milliseconds())

	LogHTTPRequest(lc.DB, lc.Logger, HTTPRequestLogParams{
		AgentType: lc.AgentType, AgentID: lc.AgentID, StepName: lc.StepName,
		OrchestrationID: lc.OrchestrationID, CorrelationID: lc.CorrelationID,
		ActionName: lc.ActionName, Method: "GET", URL: resolvedMetaURL,
		StatusCode: httpStatusFromErr(err, 200), LatencyMs: latencyMs,
		Success: err == nil, ErrorMessage: errString(err), Metadata: meta,
	})

	if err != nil {
		return nil, fmt.Errorf("document metadata: %w", err)
	}

	// Check if iXBRL format is available.
	// CH document API: resources map lists available formats by MIME type.
	// The actual content is at /document/{id}/content with the appropriate Accept header.
	resources, _ := metaResp["resources"].(map[string]interface{})
	hasIXBRL := false
	for mimeType := range resources {
		if strings.Contains(mimeType, "xhtml") || strings.Contains(mimeType, "xml") {
			hasIXBRL = true
			break
		}
	}

	if !hasIXBRL {
		// Only PDF available — can't parse, but note it exists
		lc.Logger.Info("CHFetchAccounts: no iXBRL available, only PDF",
			zap.String("company_number", companyNumber))
		result := map[string]interface{}{
			"accounts_date": accountsDate,
			"accounts_type": accountsType,
			"format":        "pdf_only",
		}
		return result, nil
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Step 3: Download iXBRL content.
	// The content URL is at links.document in the metadata response,
	// or the metadata URL + "/content". Both point to the same place.
	// Request with Accept: application/xhtml+xml → CH 302s to S3.
	contentURL := resolvedMetaURL + "/content"
	if links, ok := metaResp["links"].(map[string]interface{}); ok {
		if docURL, ok := links["document"].(string); ok && docURL != "" {
			contentURL = docURL
		}
	}

	callStart = time.Now()
	ixbrlContent, err := downloadDocument(ctx, httpClient, apiKey, contentURL)
	latencyMs = int(time.Since(callStart).Milliseconds())

	LogHTTPRequest(lc.DB, lc.Logger, HTTPRequestLogParams{
		AgentType: lc.AgentType, AgentID: lc.AgentID, StepName: lc.StepName,
		OrchestrationID: lc.OrchestrationID, CorrelationID: lc.CorrelationID,
		ActionName: lc.ActionName, Method: "GET", URL: contentURL,
		StatusCode: httpStatusFromErr(err, 200), LatencyMs: latencyMs,
		ResponseBytes: len(ixbrlContent),
		Success:       err == nil, ErrorMessage: errString(err), Metadata: meta,
	})

	if err != nil {
		return nil, fmt.Errorf("download iXBRL: %w", err)
	}

	// Step 4: Parse financial values
	values := parseIXBRL(ixbrlContent)

	result := map[string]interface{}{
		"accounts_date": accountsDate,
		"accounts_type": accountsType,
		"format":        "ixbrl",
	}
	for k, v := range values {
		result[k] = v
	}

	return result, nil
}

// httpStatusFromErr returns defaultStatus if err is nil, or 0 if there was an error.
// For more precise status codes, the caller should extract from the HTTP response.
func httpStatusFromErr(err error, defaultStatus int) int {
	if err != nil {
		return 0
	}
	return defaultStatus
}

// errString returns the error message or empty string.
func errString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// resolveDocumentURL ensures a CH document URL is absolute.
// The filing history API returns document_metadata as a relative path
// (e.g. "/document/abc123") that needs the document API host prepended.
func resolveDocumentURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "http") {
		return rawURL
	}
	return chDocumentAPIBase + rawURL
}

// chDocumentAPIGet fetches from the CH document API (different host than company API).
func chDocumentAPIGet(ctx context.Context, client *http.Client, apiKey, fullURL string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiKey, "")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// downloadDocument fetches a CH document, following redirects.
// The document API returns a redirect to S3 where the actual content lives.
func downloadDocument(ctx context.Context, client *http.Client, apiKey, url string) (string, error) {
	url = resolveDocumentURL(url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(apiKey, "")
	req.Header.Set("Accept", "application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Limit to 2MB — iXBRL files are typically 50-200KB
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// parseIXBRL extracts financial values from iXBRL content using regex.
// Returns a map of DB field names to values.
func parseIXBRL(content string) map[string]interface{} {
	result := map[string]interface{}{}

	// Collect all ix:nonFraction matches
	type ixbrlMatch struct {
		Name       string
		ContextRef string
		Value      string
		Sign       string // "-" if negative, "" otherwise
	}
	var matches []ixbrlMatch

	for _, m := range ixbrlTagPattern.FindAllStringSubmatch(content, -1) {
		if len(m) < 3 {
			continue
		}
		attrs := m[1] // all attributes
		value := m[2] // text content

		// Extract individual attributes from the tag
		nameMatch := attrName.FindStringSubmatch(attrs)
		ctxMatch := attrContextRef.FindStringSubmatch(attrs)
		if nameMatch == nil || ctxMatch == nil {
			continue
		}

		sign := ""
		if signMatch := attrSign.FindStringSubmatch(attrs); signMatch != nil {
			sign = signMatch[1]
		}

		matches = append(matches, ixbrlMatch{
			Name:       nameMatch[1],
			ContextRef: ctxMatch[1],
			Value:      value,
			Sign:       sign,
		})
	}

	// Map to DB fields
	for _, mapping := range ixbrlFieldMappings {
		for _, m := range matches {
			// Tag name match: check if the name ends with any of the mapping tags
			tagMatched := false
			for _, tagName := range mapping.TagNames {
				if strings.HasSuffix(m.Name, tagName) {
					tagMatched = true
					break
				}
			}
			if !tagMatched {
				continue
			}

			// Context match: CY_END matches "FY1CurrentYearEndDate", "CY_END", etc.
			// CY matches "FY1CurrentYear", "CY", etc.
			contextMatch := false
			contextUpper := strings.ToUpper(m.ContextRef)
			if mapping.Context == "CY_END" {
				contextMatch = strings.Contains(contextUpper, "END") ||
					strings.Contains(contextUpper, "INSTANT") ||
					strings.Contains(contextUpper, "CY_END")
			} else if mapping.Context == "CY" {
				// CY matches period contexts but NOT instant/end contexts
				contextMatch = !strings.Contains(contextUpper, "END") &&
					!strings.Contains(contextUpper, "INSTANT") &&
					!strings.Contains(contextUpper, "PY")
			}

			if !contextMatch {
				continue
			}

			// Parse the value, applying sign attribute if present
			parsed := parseIXBRLValue(m.Value, mapping.FieldType, m.Sign)
			if parsed != nil {
				result[mapping.DBField] = parsed
				break // Take first match for each field
			}
		}
	}

	return result
}

// parseIXBRLValue cleans and parses an iXBRL text value.
// Values may contain commas, spaces, negative signs, or be wrapped in parentheses.
// The sign parameter comes from the sign="" attribute on the ix:nonFraction tag.
// When sign="-", the displayed value is positive but the actual value is negative.
func parseIXBRLValue(raw, fieldType, sign string) interface{} {
	// Clean: remove commas, spaces, trim
	cleaned := strings.ReplaceAll(raw, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.TrimSpace(cleaned)

	// Handle ixt2:zerodash format — a lone "-" means zero
	if cleaned == "-" {
		return nil
	}

	// Handle negative: (123) → -123
	if strings.HasPrefix(cleaned, "(") && strings.HasSuffix(cleaned, ")") {
		cleaned = "-" + cleaned[1:len(cleaned)-1]
	}

	if cleaned == "" {
		return nil
	}

	switch fieldType {
	case "money":
		f, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return nil
		}
		// Apply sign attribute: sign="-" means negate
		if sign == "-" && f > 0 {
			f = -f
		}
		return f
	case "integer":
		i, err := strconv.Atoi(cleaned)
		if err != nil {
			// Try float then truncate
			f, err := strconv.ParseFloat(cleaned, 64)
			if err != nil {
				return nil
			}
			if sign == "-" && f > 0 {
				f = -f
			}
			return int(f)
		}
		if sign == "-" && i > 0 {
			i = -i
		}
		return i
	}
	return nil
}

// storeAccountsData updates the companies_house_data row with financial fields.
func storeAccountsData(ctx context.Context, db *sql.DB, businessID string, data map[string]interface{}) error {
	// Parse accounts_date to sql.NullTime for safe handling
	var accountsDate sql.NullTime
	if dateStr, ok := data["accounts_date"].(string); ok && dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			accountsDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	_, err := db.ExecContext(ctx, `
		UPDATE business_intel.companies_house_data
		SET accounts_date = COALESCE($2, accounts_date),
			accounts_type = COALESCE(NULLIF($3, ''), accounts_type),
			net_worth_gbp = COALESCE($4, net_worth_gbp),
			total_assets_gbp = COALESCE($5, total_assets_gbp),
			employee_count = COALESCE($6, employee_count),
			turnover_gbp = COALESCE($7, turnover_gbp),
			profit_loss_gbp = COALESCE($8, profit_loss_gbp),
			updated_at = NOW()
		WHERE business_id = $1`,
		businessID,
		accountsDate,
		stringVal(data["accounts_type"]),
		nullFloat64(data["net_worth_gbp"]),
		nullFloat64(data["total_assets_gbp"]),
		nullIntAccounts(data["employee_count"]),
		nullFloat64(data["turnover_gbp"]),
		nullFloat64(data["profit_loss_gbp"]),
	)
	return err
}

// markAccountsFetched marks the ch_vet_companies row as accounts-fetched.
func markAccountsFetched(ctx context.Context, db *sql.DB, companyNumber, businessID string, data map[string]interface{}, note string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE business_intel.ch_vet_companies
		SET accounts_fetched = true,
			accounts_fetched_at = NOW(),
			updated_at = NOW()
		WHERE company_number = $1`,
		companyNumber)
	return err
}

// stringVal safely extracts a string from an interface{}.
func stringVal(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// nullFloat64 converts an interface{} to sql.NullFloat64 for DB writes.
// Handles float64 and int values. Returns null for nil or unparseable values.
func nullFloat64(v interface{}) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	switch val := v.(type) {
	case float64:
		return sql.NullFloat64{Float64: val, Valid: true}
	case int:
		return sql.NullFloat64{Float64: float64(val), Valid: true}
	case int64:
		return sql.NullFloat64{Float64: float64(val), Valid: true}
	}
	return sql.NullFloat64{}
}

// nullInt converts an interface{} to sql.NullInt64 for DB writes.
// Note: this shadows the existing nullInt in companies_house_actions.go
// which takes (int, bool). This version takes interface{} for convenience
// in the accounts context where values come from a map.
// If there's a naming collision, rename to nullIntFromMap.
func nullIntAccounts(v interface{}) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	switch val := v.(type) {
	case int:
		return sql.NullInt64{Int64: int64(val), Valid: true}
	case int64:
		return sql.NullInt64{Int64: val, Valid: true}
	case float64:
		return sql.NullInt64{Int64: int64(val), Valid: true}
	}
	return sql.NullInt64{}
}
