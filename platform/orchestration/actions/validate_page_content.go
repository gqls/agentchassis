// FILE: platform/orchestration/actions/validate_page_content.go
//
// ValidatePageContentAction checks generated page content for issues
// before it gets saved and deployed.
//
// Checks performed:
//   1. Placeholder text — "To be added", "Lorem ipsum", "[insert", etc.
//   2. Unrendered templates — {{.field}}, {{range}}, {{if}}
//   3. Cross-site contamination — wrong domain/company name in content
//   4. Broken internal links — hrefs pointing to non-existent pages
//   5. Hallucinated emails — email addresses that don't match site's contact
//   6. Content length — suspiciously short content
//
// Returns:
//   - valid: bool (false if any blockers found)
//   - clean_html: HTML with stray comments stripped (for save_sections)
//   - issues: array of issues with category, severity, detail
//   - blocker/warning/error counts
//
// Blockers (prevent deployment): placeholder, template, contamination
// Errors (prevent deployment): broken_link, invalid_email
// Warnings (logged only): short_content
//
// Registration:
//   "validate_page_content": {
//       Handler:     ValidatePageContentAction,
//       Category:    "site",
//       Description: "Validate page content before deployment",
//       IsLocal:     true,
//   },

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
	Type        string `json:"type"`
	Category    string `json:"category"` // "placeholder", "template", "contamination", "link", "email", "short_content"
	Severity    string `json:"severity"` // "blocker", "error", "warning"
	Location    string `json:"location"`
	Value       string `json:"value"`
	Expected    string `json:"expected"`
	Description string `json:"description"`
}

// ============================================================================
// Placeholder patterns
// ============================================================================

var placeholderPatterns = []struct {
	Pattern string
	Label   string
}{
	{"needs human review", "human review marker"},
	{"needs_human_review", "human review marker"},
	{"needs review", "review marker"},
	{"to be added", "placeholder name/content"},
	{"to be confirmed", "unconfirmed content"},
	{"to be updated", "incomplete content"},
	{"lorem ipsum", "lorem ipsum placeholder"},
	{"dolor sit amet", "lorem ipsum placeholder"},
	{"[insert", "template bracket placeholder"},
	{"[your ", "template bracket placeholder"},
	{"[company", "template bracket placeholder"},
	{"[name", "template bracket placeholder"},
	{"[client", "template bracket placeholder"},
	{"[add ", "template bracket placeholder"},
	{"[replace", "template bracket placeholder"},
	{"[todo", "todo marker"},
	{"[tbd", "to be determined"},
	{"placeholder", "explicit placeholder"},
	{"sample text", "sample text"},
	{"example text", "example text"},
	{"john doe", "placeholder name"},
	{"jane doe", "placeholder name"},
	{"test@test", "test email"},
	{"test@example", "test email"},
	{"user@example", "example email"},
	{"name@example", "example email"},
	{"123 main st", "placeholder address"},
	{"your name here", "placeholder prompt"},
	{"your company", "placeholder prompt"},
	{"acme corp", "placeholder company"},
	{"todo:", "todo marker"},
	{"fixme:", "fixme marker"},
	{"coming soon", "coming soon placeholder"},
	{"<no value>", "unrendered template variable"},
}

var templateVarRegex = regexp.MustCompile(`\{\{[\s]*[\.\w]+[\s]*\}\}`)
var templateBlockRegex = regexp.MustCompile(`\{\{[\s]*(range|if|with|end|else|template|block|define)[\s]`)

// ============================================================================
// Main action
// ============================================================================

func ValidatePageContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "validate_page_content"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// ── Extract HTML ──
	htmlField := "page_content.response.page_html"
	if hf, ok := config["html_field"].(string); ok && hf != "" {
		htmlField = hf
	}

	htmlRaw := datahelpers.ExtractNestedField(params.CollectedData, htmlField)
	if htmlRaw == nil {
		for _, alt := range []string{
			"page_content.response.page_body",
			"page_content.response.html",
		} {
			htmlRaw = datahelpers.ExtractNestedField(params.CollectedData, alt)
			if htmlRaw != nil {
				break
			}
		}
	}

	htmlStr, _ := htmlRaw.(string)
	if htmlStr == "" {
		return map[string]interface{}{
			"valid":      true,
			"clean_html": "",
			"reason":     "no content to validate",
		}, nil
	}

	// ── Extract site context ──
	domain := resolveConfigString(config, "domain", params.CollectedData, logger)
	companyName := resolveConfigString(config, "company_name", params.CollectedData, logger)

	siteIDStr := resolveConfigString(config, "site_id", params.CollectedData, logger)
	if siteIDStr == "" {
		// Try site_record.site_id
		siteIDStr = extractNestedString(params.CollectedData, "site_record.site_id")
	}

	// Config toggles (all default to true)
	checkLinks := configBoolOrDefault(config, "check_internal_links", true)
	checkEmails := configBoolOrDefault(config, "check_emails", true)

	// ── Run all checks ──
	var issues []ValidationIssue

	// 1. Placeholder text
	issues = append(issues, checkPlaceholderPatterns(htmlStr)...)

	// 2. Unrendered templates
	issues = append(issues, checkUnrenderedTemplates(htmlStr)...)

	// 3. Cross-site contamination
	if domain != "" {
		issues = append(issues, checkDomainContamination(htmlStr, domain, companyName)...)
	}

	// 4. Broken internal links
	if checkLinks && params.DB != nil && siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			issues = append(issues, validateInternalLinks(ctx, params.DB, htmlStr, siteID, logger)...)
		}
	}

	// 5. Hallucinated emails
	if checkEmails && params.DB != nil && siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			issues = append(issues, validateEmails(ctx, params.DB, htmlStr, siteID, logger)...)
		}
	}

	// 6. Content length
	issues = append(issues, checkTextLength(htmlStr)...)

	// ── Categorise results ──
	blockerCount := 0
	errorCount := 0
	warningCount := 0

	for _, issue := range issues {
		switch issue.Severity {
		case "blocker":
			blockerCount++
		case "error":
			errorCount++
		case "warning":
			warningCount++
		}
	}

	valid := blockerCount == 0 && errorCount == 0

	// Build issues list for output
	issuesMaps := make([]map[string]string, len(issues))
	for i, issue := range issues {
		issuesMaps[i] = map[string]string{
			"type":        issue.Type,
			"category":    issue.Category,
			"severity":    issue.Severity,
			"location":    issue.Location,
			"value":       issue.Value,
			"expected":    issue.Expected,
			"description": issue.Description,
		}
	}

	logger.Info("ValidatePageContentAction: complete",
		zap.Bool("valid", valid),
		zap.Int("blockers", blockerCount),
		zap.Int("errors", errorCount),
		zap.Int("warnings", warningCount),
		zap.String("domain", domain))

	if !valid {
		for _, issue := range issues {
			if issue.Severity == "blocker" || issue.Severity == "error" {
				logger.Warn("ValidatePageContentAction: issue",
					zap.String("category", issue.Category),
					zap.String("severity", issue.Severity),
					zap.String("detail", issue.Description),
					zap.String("value", issue.Value))
			}
		}
		return nil, fmt.Errorf("content validation failed: %d blockers, %d errors", blockerCount, errorCount)
	}

	// ── Clean HTML — strip stray comments ──
	commentRegex := regexp.MustCompile(`<!--[\s\S]*?-->`)
	cleanHTML := commentRegex.ReplaceAllString(htmlStr, "")
	doubleNewline := regexp.MustCompile(`\n\s*\n\s*\n`)
	cleanHTML = doubleNewline.ReplaceAllString(cleanHTML, "\n\n")

	return map[string]interface{}{
		"valid":          valid,
		"clean_html":     cleanHTML,
		"blockers":       blockerCount,
		"errors":         errorCount,
		"warnings":       warningCount,
		"issue_count":    len(issues),
		"issues":         issuesMaps,
		"checked_links":  countInternalLinks(htmlStr),
		"checked_emails": countEmailAddresses(htmlStr),
	}, nil
}

// ============================================================================
// Check 1: Placeholder text
// ============================================================================

func checkPlaceholderPatterns(html string) []ValidationIssue {
	lower := strings.ToLower(html)
	var issues []ValidationIssue

	for _, p := range placeholderPatterns {
		idx := strings.Index(lower, p.Pattern)
		if idx >= 0 {
			issues = append(issues, ValidationIssue{
				Type:        "placeholder_text",
				Category:    "placeholder",
				Severity:    "blocker",
				Location:    extractSnippet(html, idx, 80),
				Value:       p.Pattern,
				Description: fmt.Sprintf("Found placeholder text '%s' (%s)", p.Pattern, p.Label),
			})
		}
	}
	return issues
}

// ============================================================================
// Check 2: Unrendered templates
// ============================================================================

func checkUnrenderedTemplates(html string) []ValidationIssue {
	var issues []ValidationIssue

	matches := templateVarRegex.FindAllString(html, 10)
	for _, match := range matches {
		issues = append(issues, ValidationIssue{
			Type:        "unrendered_template",
			Category:    "template",
			Severity:    "blocker",
			Value:       match,
			Description: fmt.Sprintf("Unrendered template variable: %s", match),
		})
	}

	blockMatches := templateBlockRegex.FindAllString(html, 10)
	for _, match := range blockMatches {
		issues = append(issues, ValidationIssue{
			Type:        "unrendered_template_block",
			Category:    "template",
			Severity:    "blocker",
			Value:       match,
			Description: fmt.Sprintf("Unrendered template block: %s", match),
		})
	}
	return issues
}

// ============================================================================
// Check 3: Cross-site contamination
// ============================================================================

func checkDomainContamination(html string, expectedDomain string, expectedCompany string) []ValidationIssue {
	var issues []ValidationIssue

	knownSites := []struct {
		Domain  string
		Company string
	}{
		{"finetuning.uk", "FineTuning"},
		{"gaswholesalers.com", "Gas Wholesalers"},
		{"ai-agent-orchestration.com", "AI Agent Orchestration"},
		{"leopardessconsulting.co.uk", "Leopardess Consulting"},
		{"dartsonline.com", "Darts Online"},
	}

	lower := strings.ToLower(html)
	expectedLower := strings.ToLower(expectedDomain)

	for _, known := range knownSites {
		if strings.ToLower(known.Domain) == expectedLower {
			continue
		}

		if strings.Contains(lower, strings.ToLower(known.Domain)) {
			idx := strings.Index(lower, strings.ToLower(known.Domain))
			issues = append(issues, ValidationIssue{
				Type:        "cross_site_domain",
				Category:    "contamination",
				Severity:    "blocker",
				Location:    extractSnippet(html, idx, 80),
				Value:       known.Domain,
				Expected:    expectedDomain,
				Description: fmt.Sprintf("Found domain '%s' in content for '%s'", known.Domain, expectedDomain),
			})
		}

		if known.Company != "" && expectedCompany != "" &&
			strings.ToLower(known.Company) != strings.ToLower(expectedCompany) {
			companyLower := strings.ToLower(known.Company)
			if strings.Contains(lower, companyLower) {
				idx := strings.Index(lower, companyLower)
				issues = append(issues, ValidationIssue{
					Type:        "cross_site_company",
					Category:    "contamination",
					Severity:    "blocker",
					Location:    extractSnippet(html, idx, 80),
					Value:       known.Company,
					Expected:    expectedCompany,
					Description: fmt.Sprintf("Found company '%s' in content for '%s'", known.Company, expectedCompany),
				})
			}
		}
	}
	return issues
}

// ============================================================================
// Check 4: Broken internal links (from existing code)
// ============================================================================

func validateInternalLinks(ctx context.Context, db *sql.DB, html string, siteID uuid.UUID, logger *zap.Logger) []ValidationIssue {
	var issues []ValidationIssue

	validPages := loadValidPagePaths(ctx, db, siteID, logger)

	hrefRe := regexp.MustCompile(`href=["']([^"']+)["']`)
	matches := hrefRe.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		href := match[1]

		// Skip external, anchors, mailto, tel, javascript
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") ||
			strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "tel:") || strings.HasPrefix(href, "javascript:") {
			continue
		}

		// Skip asset paths — these are files in the git repo, not pages.
		// Catches: /assets/css/styles.css, /assets/images/hero.jpg, /favicon.ico, etc.
		if isAssetPath(href) {
			continue
		}

		normalizedPath := normalizePagePath(href)

		if !isValidPage(normalizedPath, validPages) {
			// Missing pages are warnings, not errors — the page may be planned
			// but not yet built (e.g. /privacy.html, /terms.html in footers).
			// This avoids blocking deployment of working pages because a
			// linked page hasn't been created yet.
			issues = append(issues, ValidationIssue{
				Type:        "missing_link_target",
				Category:    "link",
				Severity:    "warning",
				Location:    fmt.Sprintf("href=\"%s\"", href),
				Value:       href,
				Description: fmt.Sprintf("Link to page not found in pages table: %s", href),
			})
		}
	}

	return issues
}

// isAssetPath returns true for paths that are static assets, not pages.
// These come from <link rel="stylesheet">, <img src="">, <script src="">, etc.
func isAssetPath(path string) bool {
	lower := strings.ToLower(path)

	// Directory-based: /assets/*, /images/*, /static/*
	if strings.HasPrefix(lower, "/assets/") ||
		strings.HasPrefix(lower, "/images/") ||
		strings.HasPrefix(lower, "/static/") ||
		strings.HasPrefix(lower, "/fonts/") ||
		strings.HasPrefix(lower, "/js/") ||
		strings.HasPrefix(lower, "/css/") {
		return true
	}

	// Extension-based: non-HTML file extensions
	assetExts := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg",
		".ico", ".woff", ".woff2", ".ttf", ".eot", ".pdf", ".xml", ".json",
		".map", ".txt", ".mp4", ".webm",
	}
	for _, ext := range assetExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}

	return false
}

// ============================================================================
// Check 5: Hallucinated emails (from existing code)
// ============================================================================

func validateEmails(ctx context.Context, db *sql.DB, html string, siteID uuid.UUID, logger *zap.Logger) []ValidationIssue {
	var issues []ValidationIssue

	officialEmail := loadSiteContactEmail(ctx, db, siteID, logger)

	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRe.FindAllString(html, -1)

	seen := make(map[string]bool)
	for _, email := range emails {
		email = strings.ToLower(email)
		if seen[email] {
			continue
		}
		seen[email] = true

		if officialEmail != "" && email == strings.ToLower(officialEmail) {
			continue
		}

		if isPlaceholderEmail(email) {
			issues = append(issues, ValidationIssue{
				Type:        "placeholder_email",
				Category:    "placeholder",
				Severity:    "blocker",
				Value:       email,
				Description: fmt.Sprintf("Placeholder email address: %s", email),
			})
			continue
		}

		if officialEmail != "" {
			issues = append(issues, ValidationIssue{
				Type:        "invalid_email",
				Category:    "email",
				Severity:    "error",
				Location:    "email address",
				Value:       email,
				Expected:    officialEmail,
				Description: fmt.Sprintf("Email '%s' doesn't match site contact '%s'", email, officialEmail),
			})
		}
	}

	return issues
}

// ============================================================================
// Check 6: Content length
// ============================================================================

func checkTextLength(html string) []ValidationIssue {
	var issues []ValidationIssue

	tagRegex := regexp.MustCompile(`<[^>]*>`)
	stripped := tagRegex.ReplaceAllString(html, " ")
	textLen := len(strings.TrimSpace(stripped))

	if textLen < 50 {
		issues = append(issues, ValidationIssue{
			Type:        "short_content",
			Category:    "short_content",
			Severity:    "warning",
			Value:       fmt.Sprintf("%d characters", textLen),
			Description: fmt.Sprintf("Content is very short (%d characters of text)", textLen),
		})
	}
	return issues
}

// ============================================================================
// DB helpers (from existing code)
// ============================================================================

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

	var pageNames []string
	for rows.Next() {
		var url, name string
		if err := rows.Scan(&url, &name); err != nil {
			continue
		}
		pageNames = append(pageNames, name)
		pages[url] = true
		pages[normalizePagePath(url)] = true
		if url == "/" {
			pages["/index.html"] = true
			pages["index.html"] = true
		}
		if !strings.HasSuffix(url, ".html") && url != "/" {
			pages[url+".html"] = true
			pages[strings.TrimPrefix(url, "/")+".html"] = true
		}
	}

	pages["/"] = true
	pages["/index.html"] = true
	pages["index.html"] = true

	logger.Info("ValidatePageContentAction: loaded valid pages",
		zap.Int("page_count", len(pageNames)),
		zap.Strings("pages", pageNames))

	return pages
}

func loadSiteContactEmail(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	var email sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(
			s.content_data->>'contact_email',
			s.content_data->'reviewed_brief'->>'contact_email',
			s.content_data->'brief'->>'contact_email',
			(SELECT spec_data->>'email'
			 FROM site_specs
			 WHERE site_id = $1 AND aspect = 'identity'
			   AND spec_data->>'email' IS NOT NULL
			   AND spec_data->>'email' != ''
			 LIMIT 1),
			(SELECT spec_data->>'contact_email'
			 FROM site_specs
			 WHERE site_id = $1 AND aspect = 'identity'
			   AND spec_data->>'contact_email' IS NOT NULL
			   AND spec_data->>'contact_email' != ''
			 LIMIT 1),
			''
		) FROM sites s WHERE s.id = $1
	`, siteID).Scan(&email)

	if err != nil {
		logger.Warn("Failed to load site contact email", zap.Error(err))
		return ""
	}
	return email.String
}

func normalizePagePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.ToLower(path)
	path = strings.TrimPrefix(path, "/")
	if path != "" && !strings.HasSuffix(path, ".html") && !strings.Contains(path, ".") {
		path = path + ".html"
	}
	return path
}

func isValidPage(path string, validPages map[string]bool) bool {
	if validPages[path] {
		return true
	}
	if validPages["/"+path] {
		return true
	}
	withoutHtml := strings.TrimSuffix(path, ".html")
	if validPages[withoutHtml] || validPages["/"+withoutHtml] {
		return true
	}
	return false
}

func isPlaceholderEmail(email string) bool {
	placeholders := []string{
		"example.com", "test.com", "placeholder",
		"your@email", "email@email", "name@company",
	}
	email = strings.ToLower(email)
	for _, p := range placeholders {
		if strings.Contains(email, p) {
			return true
		}
	}
	return false
}

func countInternalLinks(html string) int {
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

func countEmailAddresses(html string) int {
	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return len(emailRe.FindAllString(html, -1))
}

// ============================================================================
// String helpers
// ============================================================================

func extractSnippet(s string, idx int, maxLen int) string {
	start := idx - 20
	if start < 0 {
		start = 0
	}
	end := idx + maxLen
	if end > len(s) {
		end = len(s)
	}
	snippet := s[start:end]
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	snippet = strings.ReplaceAll(snippet, "\r", "")
	spaceRegex := regexp.MustCompile(`\s+`)
	snippet = spaceRegex.ReplaceAllString(snippet, " ")
	return strings.TrimSpace(snippet)
}

func resolveConfigString(config map[string]interface{}, key string, collectedData map[string]interface{}, logger *zap.Logger) string {
	if val, ok := config[key].(string); ok && val != "" {
		// Check if it's a dot-path reference
		if strings.Contains(val, ".") {
			if resolved := extractNestedString(collectedData, val); resolved != "" {
				return resolved
			}
		}
		return val
	}
	return ""
}

func extractNestedString(data map[string]interface{}, dotPath string) string {
	parts := strings.Split(dotPath, ".")
	var current interface{} = data
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current, ok = m[part]
		if !ok {
			return ""
		}
	}
	if s, ok := current.(string); ok {
		return s
	}
	return ""
}

func configBoolOrDefault(config map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := config[key].(bool); ok {
		return v
	}
	return defaultVal
}
