// FILE: platform/orchestration/actions/validate_page_content.go
// ValidatePageContentAction checks for hallucinated links and contact info
// Used by content-reviewer before auto-eval or HITL review

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Type        string `json:"type"`        // "broken_link", "invalid_email", "external_link"
	Severity    string `json:"severity"`    // "error", "warning"
	Location    string `json:"location"`    // Where in the HTML
	Value       string `json:"value"`       // The problematic value
	Expected    string `json:"expected"`    // What it should be (if applicable)
	Description string `json:"description"` // Human-readable description
}

// ValidationResult is the output of validation
type ValidationResult struct {
	Valid         bool              `json:"valid"`
	Issues        []ValidationIssue `json:"issues"`
	IssueCount    int               `json:"issue_count"`
	ErrorCount    int               `json:"error_count"`
	WarningCount  int               `json:"warning_count"`
	CheckedLinks  int               `json:"checked_links"`
	CheckedEmails int               `json:"checked_emails"`
}

// ValidatePageContentAction checks page content for common issues
// Config:
//   - input_fields: array of field names to extract (default: ["page_content", "site_record"])
//   - check_internal_links: bool (default: true)
//   - check_emails: bool (default: true)
//   - check_external_links: bool (default: false, just warns)
//   - allowed_external_domains: array of allowed external domains
//
// Expects in extracted fields:
//   - page_content.response.page_html: the HTML to validate
//   - site_record.site_id: the site UUID
func ValidatePageContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ValidatePageContentAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	config := params.StepConfig.Config

	// Get input_fields (default to what we need)
	inputFields := []string{"page_content", "site_record"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}

	// Extract the fields using standard helper (handles input_data prefix)
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	// Get HTML from page_content.response.page_html
	html := ""
	if pageContent, ok := extracted["page_content"].(map[string]interface{}); ok {
		if response, ok := pageContent["response"].(map[string]interface{}); ok {
			html, _ = response["page_html"].(string)
		}
	}
	if html == "" {
		params.Logger.Error("ValidatePageContentAction: No HTML found",
			zap.Any("extracted_keys", datahelpers.GetMapKeys(extracted)),
		)
		return nil, fmt.Errorf("no HTML content found in page_content.response.page_html")
	}

	// Get site_id from site_record.site_id
	siteIDStr := ""
	if siteRecord, ok := extracted["site_record"].(map[string]interface{}); ok {
		siteIDStr, _ = siteRecord["site_id"].(string)
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid or missing site_id in site_record.site_id: %w", err)
	}

	// Config options
	checkInternalLinks := true
	if v, ok := config["check_internal_links"].(bool); ok {
		checkInternalLinks = v
	}
	checkEmails := true
	if v, ok := config["check_emails"].(bool); ok {
		checkEmails = v
	}

	result := ValidationResult{
		Valid:  true,
		Issues: []ValidationIssue{},
	}

	// Check internal links
	if checkInternalLinks {
		linkIssues := validateInternalLinks(ctx, params.DB, html, siteID, params.Logger)
		result.Issues = append(result.Issues, linkIssues...)
		result.CheckedLinks = countLinks(html)
	}

	// Check email addresses
	if checkEmails {
		emailIssues := validateEmails(ctx, params.DB, html, siteID, params.Logger)
		result.Issues = append(result.Issues, emailIssues...)
		result.CheckedEmails = countEmails(html)
	}

	// Count issues by severity
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			result.ErrorCount++
		} else {
			result.WarningCount++
		}
	}
	result.IssueCount = len(result.Issues)
	result.Valid = result.ErrorCount == 0

	params.Logger.Info("ValidatePageContentAction: Complete",
		zap.Bool("valid", result.Valid),
		zap.Int("issues", result.IssueCount),
		zap.Int("errors", result.ErrorCount),
		zap.Int("warnings", result.WarningCount),
	)

	return result, nil
}

// validateInternalLinks checks all internal hrefs against site pages
func validateInternalLinks(ctx context.Context, db *sql.DB, html string, siteID uuid.UUID, logger *zap.Logger) []ValidationIssue {
	var issues []ValidationIssue

	// Load valid pages for this site
	validPages := loadValidPagePaths(ctx, db, siteID, logger)

	// Extract all href values
	hrefRe := regexp.MustCompile(`href=["']([^"']+)["']`)
	matches := hrefRe.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		href := match[1]

		// Skip external links, anchors, mailto, tel, javascript
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") ||
			strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "tel:") || strings.HasPrefix(href, "javascript:") {
			continue
		}

		// Normalize the path
		normalizedPath := normalizePagePath(href)

		// Check if this page exists
		if !isValidPage(normalizedPath, validPages) {
			issues = append(issues, ValidationIssue{
				Type:        "broken_link",
				Severity:    "error",
				Location:    fmt.Sprintf("href=\"%s\"", href),
				Value:       href,
				Description: fmt.Sprintf("Link to non-existent page: %s", href),
			})
		}
	}

	return issues
}

// validateEmails checks email addresses against site's contact email
func validateEmails(ctx context.Context, db *sql.DB, html string, siteID uuid.UUID, logger *zap.Logger) []ValidationIssue {
	var issues []ValidationIssue

	// Load site's official contact email
	officialEmail := loadSiteContactEmail(ctx, db, siteID, logger)

	// Extract all email addresses from HTML
	// Match both mailto: links and plain email text
	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRe.FindAllString(html, -1)

	// Deduplicate
	seen := make(map[string]bool)
	for _, email := range emails {
		email = strings.ToLower(email)
		if seen[email] {
			continue
		}
		seen[email] = true

		// Skip if it matches official email
		if officialEmail != "" && email == strings.ToLower(officialEmail) {
			continue
		}

		// Skip common placeholder patterns
		if isPlaceholderEmail(email) {
			continue
		}

		// Flag as potentially hallucinated
		issues = append(issues, ValidationIssue{
			Type:        "invalid_email",
			Severity:    "error", // could be Warning because it might be intentional
			Location:    "email address",
			Value:       email,
			Expected:    officialEmail,
			Description: fmt.Sprintf("Email '%s' doesn't match site contact email '%s'", email, officialEmail),
		})
	}

	return issues
}

// loadValidPagePaths returns all valid page paths for a site
func loadValidPagePaths(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) map[string]bool {
	pages := make(map[string]bool)

	rows, err := db.QueryContext(ctx, `
		SELECT url, name FROM pages 
		WHERE site_id = $1 AND status NOT IN ('deleted', 'archived')
	`, siteID)
	if err != nil {
		logger.Warn("Failed to load pages for validation", zap.Error(err))
		return pages
	}
	defer rows.Close()

	for rows.Next() {
		var url, name string
		if err := rows.Scan(&url, &name); err != nil {
			continue
		}

		// Add various forms of the URL
		pages[url] = true
		pages[normalizePagePath(url)] = true

		// Also add common variations
		if url == "/" {
			pages["/index.html"] = true
			pages["index.html"] = true
		}
		if !strings.HasSuffix(url, ".html") && url != "/" {
			pages[url+".html"] = true
			pages[strings.TrimPrefix(url, "/")+".html"] = true
		}
	}

	// Add common static paths that are always valid
	pages["/"] = true
	pages["/index.html"] = true
	pages["index.html"] = true

	logger.Debug("Loaded valid pages",
		zap.Int("count", len(pages)),
	)

	return pages
}

// loadSiteContactEmail returns the site's official contact email
func loadSiteContactEmail(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	var email sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(
			content_data->>'contact_email',
			content_data->'reviewed_brief'->>'contact_email',
			content_data->'brief'->>'contact_email',
			''
		) FROM sites WHERE id = $1
	`, siteID).Scan(&email)

	if err != nil {
		logger.Warn("Failed to load site contact email", zap.Error(err))
		return ""
	}

	return email.String
}

// normalizePagePath normalizes a page path for comparison
func normalizePagePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.ToLower(path)

	// Remove leading slash for comparison
	path = strings.TrimPrefix(path, "/")

	// Add .html if missing (except for root)
	if path != "" && !strings.HasSuffix(path, ".html") && !strings.Contains(path, ".") {
		path = path + ".html"
	}

	return path
}

// isValidPage checks if a path is in the valid pages map
func isValidPage(path string, validPages map[string]bool) bool {
	// Check exact match
	if validPages[path] {
		return true
	}

	// Check with leading slash
	if validPages["/"+path] {
		return true
	}

	// Check without .html
	withoutHtml := strings.TrimSuffix(path, ".html")
	if validPages[withoutHtml] || validPages["/"+withoutHtml] {
		return true
	}

	return false
}

// isPlaceholderEmail checks if an email looks like a placeholder
func isPlaceholderEmail(email string) bool {
	placeholders := []string{
		"example.com",
		"test.com",
		"placeholder",
		"your@email",
		"email@email",
		"name@company",
	}

	email = strings.ToLower(email)
	for _, p := range placeholders {
		if strings.Contains(email, p) {
			return true
		}
	}
	return false
}

// countLinks counts internal links in HTML
func countLinks(html string) int {
	hrefRe := regexp.MustCompile(`href=["']([^"']+)["']`)
	matches := hrefRe.FindAllString(html, -1)
	count := 0
	for _, m := range matches {
		if !strings.Contains(m, "http://") && !strings.Contains(m, "https://") &&
			!strings.Contains(m, "mailto:") && !strings.Contains(m, "tel:") {
			count++
		}
	}
	return count
}

// countEmails counts email addresses in HTML
func countEmails(html string) int {
	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return len(emailRe.FindAllString(html, -1))
}
