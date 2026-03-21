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

// iXBRL tag names mapped to DB fields.
// Context: CY_END = current year-end (balance sheet instant), CY = current year period.
var ixbrlFieldMappings = []struct {
	TagName   string // e.g. "core:NetAssetsLiabilities"
	Context   string // "CY_END" or "CY" (matched as prefix)
	DBField   string // column in companies_house_data
	FieldType string // "money" or "integer"
}{
	{"NetAssetsLiabilities", "CY_END", "net_worth_gbp", "money"},
	{"TotalAssetsLessCurrentLiabilities", "CY_END", "total_assets_gbp", "money"},
	{"AverageNumberEmployeesDuringPeriod", "CY", "employee_count", "integer"},
	{"TurnoverRevenue", "CY", "turnover_gbp", "money"},
	{"ProfitLossForPeriod", "CY", "profit_loss_gbp", "money"},
}

// Regex to extract ix:nonFraction elements from iXBRL.
// Captures: name attribute, contextRef attribute, text content (the value).
var ixbrlPattern = regexp.MustCompile(
	`<ix:nonFraction[^>]*name="([^"]*)"[^>]*contextRef="([^"]*)"[^>]*>([^<]+)</ix:nonFraction>`)

// Also match ix:nonFraction with attributes in different order
var ixbrlPatternAlt = regexp.MustCompile(
	`<ix:nonFraction[^>]*contextRef="([^"]*)"[^>]*name="([^"]*)"[^>]*>([^<]+)</ix:nonFraction>`)

// accountsCompany holds a company ready for accounts fetch
type accountsCompany struct {
	CompanyNumber string
	BusinessID    string
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

	totalFetched := 0
	totalParsed := 0
	totalFailed := 0
	totalNoAccounts := 0

	for _, co := range companies {
		select {
		case <-ctx.Done():
			break
		default:
		}

		result, err := fetchAndParseAccounts(ctx, apiKey, httpClient, co.CompanyNumber, delayMs, params.Logger)
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

// accountsResult holds parsed financial data from iXBRL
type accountsResult struct {
	AccountsDate string
	AccountsType string
	Values       map[string]interface{} // field_name -> value
}

// fetchAndParseAccounts does the three-step fetch: filing history → metadata → iXBRL.
func fetchAndParseAccounts(ctx context.Context, apiKey string, httpClient *http.Client, companyNumber string, delayMs int, logger *zap.Logger) (map[string]interface{}, error) {

	// Step 1: Get latest accounts filing
	filingResp, err := chAPIGet(ctx, apiKey,
		fmt.Sprintf("/company/%s/filing-history?category=accounts&items_per_page=1", companyNumber))
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
	// Note: document metadata URL is on document-api.company-information.service.gov.uk
	// which is a different host than the company API, so we can't use chAPIGet.
	metaResp, err := chDocumentAPIGet(ctx, httpClient, apiKey, documentMetaURL)
	if err != nil {
		return nil, fmt.Errorf("document metadata: %w", err)
	}

	// Find the iXBRL content URL
	resources, _ := metaResp["resources"].(map[string]interface{})
	var contentURL string
	for mimeType, resource := range resources {
		if strings.Contains(mimeType, "xhtml") || strings.Contains(mimeType, "xml") {
			if resMap, ok := resource.(map[string]interface{}); ok {
				contentURL, _ = resMap["content_url"].(string)
				break
			}
		}
	}

	if contentURL == "" {
		// Try PDF as fallback — can't parse, but note it exists
		logger.Info("CHFetchAccounts: no iXBRL available, only PDF",
			zap.String("company_number", companyNumber))
		result := map[string]interface{}{
			"accounts_date": accountsDate,
			"accounts_type": accountsType,
			"format":        "pdf_only",
		}
		return result, nil
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Step 3: Download iXBRL content (follows redirects to S3)
	ixbrlContent, err := downloadDocument(ctx, httpClient, apiKey, contentURL)
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

// chDocumentAPIGet fetches from the CH document API (different host than company API).
// The document metadata URL is a full URL like:
// https://document-api.company-information.service.gov.uk/document/abc123
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
	if !strings.HasPrefix(url, "http") {
		url = "https://document-api.company-information.service.gov.uk" + url
	}

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
	}
	var matches []ixbrlMatch

	// Try primary pattern: name before contextRef
	for _, m := range ixbrlPattern.FindAllStringSubmatch(content, -1) {
		if len(m) >= 4 {
			matches = append(matches, ixbrlMatch{Name: m[1], ContextRef: m[2], Value: m[3]})
		}
	}
	// Try alternate pattern: contextRef before name
	for _, m := range ixbrlPatternAlt.FindAllStringSubmatch(content, -1) {
		if len(m) >= 4 {
			matches = append(matches, ixbrlMatch{Name: m[2], ContextRef: m[1], Value: m[3]})
		}
	}

	// Map to DB fields
	for _, mapping := range ixbrlFieldMappings {
		for _, m := range matches {
			// Tag name match: check if the name ends with the mapping tag
			// e.g. "core:NetAssetsLiabilities" ends with "NetAssetsLiabilities"
			if !strings.HasSuffix(m.Name, mapping.TagName) {
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

			// Parse the value
			parsed := parseIXBRLValue(m.Value, mapping.FieldType)
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
func parseIXBRLValue(raw, fieldType string) interface{} {
	// Clean: remove commas, spaces, trim
	cleaned := strings.ReplaceAll(raw, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.TrimSpace(cleaned)

	// Handle negative: (123) → -123
	if strings.HasPrefix(cleaned, "(") && strings.HasSuffix(cleaned, ")") {
		cleaned = "-" + cleaned[1:len(cleaned)-1]
	}

	if cleaned == "" || cleaned == "-" {
		return nil
	}

	switch fieldType {
	case "money":
		f, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return nil
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
			return int(f)
		}
		return i
	}
	return nil
}

// storeAccountsData updates the companies_house_data row with financial fields.
func storeAccountsData(ctx context.Context, db *sql.DB, businessID string, data map[string]interface{}) error {
	_, err := db.ExecContext(ctx, `
		UPDATE business_intel.companies_house_data
		SET accounts_date = CASE WHEN $2 != '' THEN $2::date ELSE accounts_date END,
			accounts_type = COALESCE(NULLIF($3, ''), accounts_type),
			net_worth_gbp = COALESCE($4, net_worth_gbp),
			total_assets_gbp = COALESCE($5, total_assets_gbp),
			employee_count = COALESCE($6, employee_count),
			turnover_gbp = COALESCE($7, turnover_gbp),
			profit_loss_gbp = COALESCE($8, profit_loss_gbp),
			updated_at = NOW()
		WHERE business_id = $1`,
		businessID,
		stringVal(data["accounts_date"]),
		stringVal(data["accounts_type"]),
		nullFloatFromInterface(data["net_worth_gbp"]),
		nullFloatFromInterface(data["total_assets_gbp"]),
		nullIntFromInterface(data["employee_count"]),
		nullFloatFromInterface(data["turnover_gbp"]),
		nullFloatFromInterface(data["profit_loss_gbp"]),
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
