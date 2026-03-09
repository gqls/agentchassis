// FILE: platform/orchestration/actions/validate_page_content_action.go
//
// Validates and triages generated page content before deployment.
// Three responsibilities:
//
//   1. TRIAGE: Finds DATA_NEEDED and SKIP markers from the content writer.
//      Strips those sections from the HTML. Creates needs_human_review
//      work items for DATA_NEEDED sections so humans can provide missing data.
//
//   2. VALIDATE: Checks remaining HTML for placeholder text, unrendered
//      templates, cross-site contamination. Returns error if blockers found.
//
//   3. CLEAN: Returns the cleaned HTML (markers stripped) for save_sections.
//
// Returns error only for content that should NOT deploy (placeholders,
// contamination). DATA_NEEDED sections are handled gracefully — stripped
// and queued for human input.
//
// Registration:
//   "validate_page_content": {
//       Handler:     ValidatePageContentAction,
//       Category:    "site",
//       Description: "Validate and triage page content before deployment",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ValidatePageContentInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"html_field", "domain", "company_name", "site_id", "page_name", "work_item_id"},
	Defaults:   map[string]interface{}{"html_field": "page_content.response.page_html"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("validate_page_content", ValidatePageContentInputSpec)
}

// ============================================================================
// Marker patterns from content writer
// ============================================================================

var dataNeedRegex = regexp.MustCompile(`<!--\s*DATA_NEEDED:\s*([^|]+)\s*\|\s*([^-]*?)-->`)
var skipRegex = regexp.MustCompile(`<!--\s*SKIP:\s*([^|]+)\s*\|\s*([^-]*?)-->`)

// Also catch the surrounding section if possible
// Matches: <section...DATA_NEEDED...></section> or just the comment
var sectionWithMarkerRegex = regexp.MustCompile(
	`<section[^>]*>[^<]*<!--\s*(DATA_NEEDED|SKIP):[^>]*-->[^<]*</section>`)

// ============================================================================
// Placeholder patterns
// ============================================================================

var placeholderPatterns = []struct {
	Pattern string
	Label   string
}{
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
}

var templateVarRegex = regexp.MustCompile(`\{\{[\s]*[\.\w]+[\s]*\}\}`)
var templateBlockRegex = regexp.MustCompile(`\{\{[\s]*(range|if|with|end|else|template|block|define)[\s]`)

// ============================================================================
// Issue types
// ============================================================================

type contentIssue struct {
	Category string `json:"category"`
	Detail   string `json:"detail"`
	Snippet  string `json:"snippet"`
}

// ============================================================================
// Main action
// ============================================================================

func ValidatePageContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "validate_page_content"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, ValidatePageContentInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	// Get the HTML to validate
	htmlField := inputs.Get("html_field")
	if htmlField == "" {
		htmlField = "page_content.response.page_html"
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

	domain := inputs.Get("domain")
	companyName := inputs.Get("company_name")
	siteID := inputs.Get("site_id")
	pageName := inputs.Get("page_name")
	workItemID := inputs.Get("work_item_id")

	// ── Step 1: Extract and handle DATA_NEEDED / SKIP markers ──
	cleanHTML, dataNeeded, skipped := extractMarkers(htmlStr, logger)

	// Create work items for DATA_NEEDED sections
	if params.DB != nil && siteID != "" && len(dataNeeded) > 0 {
		createDataNeededItems(ctx, params.DB, siteID, domain, pageName, workItemID, dataNeeded, logger)
	}

	logger.Info("ValidatePageContentAction: markers processed",
		zap.Int("data_needed_sections", len(dataNeeded)),
		zap.Int("skipped_sections", len(skipped)),
		zap.Int("original_length", len(htmlStr)),
		zap.Int("clean_length", len(cleanHTML)))

	// ── Step 2: Validate the cleaned HTML ──
	var issues []contentIssue

	placeholderIssues := checkPlaceholders(cleanHTML)
	issues = append(issues, placeholderIssues...)

	templateIssues := checkTemplates(cleanHTML)
	issues = append(issues, templateIssues...)

	if domain != "" {
		contaminationIssues := checkCrossSiteContamination(cleanHTML, domain, companyName)
		issues = append(issues, contaminationIssues...)
	}

	shortIssues := checkContentLength(cleanHTML)
	issues = append(issues, shortIssues...)

	// Count blockers
	blockers := 0
	warnings := 0
	for _, issue := range issues {
		switch issue.Category {
		case "placeholder", "template", "contamination":
			blockers++
		default:
			warnings++
		}
	}

	issuesMaps := make([]map[string]string, len(issues))
	for i, issue := range issues {
		issuesMaps[i] = map[string]string{
			"category": issue.Category,
			"detail":   issue.Detail,
			"snippet":  issue.Snippet,
		}
	}

	logger.Info("ValidatePageContentAction: validation complete",
		zap.Bool("valid", blockers == 0),
		zap.Int("blockers", blockers),
		zap.Int("warnings", warnings),
		zap.String("domain", domain))

	if blockers > 0 {
		for _, issue := range issues {
			if issue.Category == "placeholder" || issue.Category == "template" || issue.Category == "contamination" {
				logger.Warn("ValidatePageContentAction: blocker",
					zap.String("category", issue.Category),
					zap.String("detail", issue.Detail),
					zap.String("snippet", issue.Snippet))
			}
		}
		return nil, fmt.Errorf("content validation failed: %d blockers found", blockers)
	}

	// ── Step 3: Return cleaned HTML for downstream save_sections ──
	return map[string]interface{}{
		"valid":             true,
		"clean_html":        cleanHTML,
		"blockers":          blockers,
		"warnings":          warnings,
		"issues":            issuesMaps,
		"data_needed_count": len(dataNeeded),
		"skipped_count":     len(skipped),
		"data_needed":       dataNeeded,
		"skipped":           skipped,
	}, nil
}

// ============================================================================
// Marker extraction — finds DATA_NEEDED and SKIP, strips them from HTML
// ============================================================================

type markerInfo struct {
	SectionName string `json:"section_name"`
	Reason      string `json:"reason"`
}

func extractMarkers(html string, logger *zap.Logger) (cleanHTML string, dataNeeded []markerInfo, skipped []markerInfo) {
	cleanHTML = html

	// Extract DATA_NEEDED markers
	dnMatches := dataNeedRegex.FindAllStringSubmatch(html, -1)
	for _, match := range dnMatches {
		if len(match) >= 3 {
			dataNeeded = append(dataNeeded, markerInfo{
				SectionName: strings.TrimSpace(match[1]),
				Reason:      strings.TrimSpace(match[2]),
			})
			logger.Info("ValidatePageContentAction: DATA_NEEDED section found",
				zap.String("section", strings.TrimSpace(match[1])),
				zap.String("reason", strings.TrimSpace(match[2])))
		}
	}

	// Extract SKIP markers
	skipMatches := skipRegex.FindAllStringSubmatch(html, -1)
	for _, match := range skipMatches {
		if len(match) >= 3 {
			skipped = append(skipped, markerInfo{
				SectionName: strings.TrimSpace(match[1]),
				Reason:      strings.TrimSpace(match[2]),
			})
		}
	}

	// Strip sections that contain markers (try section-level first)
	cleanHTML = sectionWithMarkerRegex.ReplaceAllString(cleanHTML, "")

	// Strip any remaining standalone markers
	cleanHTML = dataNeedRegex.ReplaceAllString(cleanHTML, "")
	cleanHTML = skipRegex.ReplaceAllString(cleanHTML, "")

	// Clean up any double whitespace left behind
	doubleNewline := regexp.MustCompile(`\n\s*\n\s*\n`)
	cleanHTML = doubleNewline.ReplaceAllString(cleanHTML, "\n\n")

	return cleanHTML, dataNeeded, skipped
}

// ============================================================================
// Create work items for DATA_NEEDED sections
// ============================================================================

func createDataNeededItems(ctx context.Context, db *sql.DB, siteIDStr, domain, pageName, parentWorkItemID string, needed []markerInfo, logger *zap.Logger) {
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		logger.Warn("createDataNeededItems: invalid site_id", zap.Error(err))
		return
	}

	var parentID *uuid.UUID
	if parentWorkItemID != "" {
		if parsed, err := uuid.Parse(parentWorkItemID); err == nil {
			parentID = &parsed
		}
	}

	for _, item := range needed {
		spec := map[string]interface{}{
			"page_name":    pageName,
			"section_name": item.SectionName,
			"data_needed":  item.Reason,
			"source":       "validate_page_content",
		}
		specJSON, _ := json.Marshal(spec)

		summary := fmt.Sprintf("Section '%s' on %s needs data: %s",
			item.SectionName, pageName, truncateStr(item.Reason, 80))

		itemKey := fmt.Sprintf("data_needed_%s_%s_%s",
			pageName, sanitiseKey(item.SectionName), siteID)

		_, err := db.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, domain, item_type, severity, summary,
				spec, priority, handler_agent, status, created_by,
				item_key, parent_item_id
			) VALUES ($1, 'content-validation', $2, 'needs_section_data', 'medium', $3,
			          $4::jsonb, 50, 'page-build-handler', 'needs_human_review',
			          'validate_page_content', $5, $6)
			ON CONFLICT DO NOTHING
		`, siteID, domain, summary, string(specJSON), itemKey, parentID)

		if err != nil {
			logger.Warn("createDataNeededItems: failed to insert",
				zap.String("section", item.SectionName),
				zap.Error(err))
		} else {
			logger.Info("createDataNeededItems: created HITL work item",
				zap.String("section", item.SectionName),
				zap.String("page", pageName),
				zap.String("status", "needs_human_review"))
		}
	}
}

// ============================================================================
// Validation checks
// ============================================================================

func checkPlaceholders(html string) []contentIssue {
	lower := strings.ToLower(html)
	var issues []contentIssue

	for _, p := range placeholderPatterns {
		idx := strings.Index(lower, p.Pattern)
		if idx >= 0 {
			snippet := extractSnippet(html, idx, 80)
			issues = append(issues, contentIssue{
				Category: "placeholder",
				Detail:   fmt.Sprintf("Found '%s' (%s)", p.Pattern, p.Label),
				Snippet:  snippet,
			})
		}
	}

	return issues
}

func checkTemplates(html string) []contentIssue {
	var issues []contentIssue

	matches := templateVarRegex.FindAllString(html, 10)
	for _, match := range matches {
		issues = append(issues, contentIssue{
			Category: "template",
			Detail:   fmt.Sprintf("Unrendered template variable: %s", match),
			Snippet:  match,
		})
	}

	blockMatches := templateBlockRegex.FindAllString(html, 10)
	for _, match := range blockMatches {
		issues = append(issues, contentIssue{
			Category: "template",
			Detail:   fmt.Sprintf("Unrendered template block: %s", match),
			Snippet:  match,
		})
	}

	return issues
}

func checkCrossSiteContamination(html string, expectedDomain string, expectedCompany string) []contentIssue {
	var issues []contentIssue

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
			issues = append(issues, contentIssue{
				Category: "contamination",
				Detail:   fmt.Sprintf("Found '%s' in content for '%s'", known.Domain, expectedDomain),
				Snippet:  extractSnippet(html, idx, 80),
			})
		}

		if known.Company != "" && expectedCompany != "" &&
			strings.ToLower(known.Company) != strings.ToLower(expectedCompany) {
			companyLower := strings.ToLower(known.Company)
			if strings.Contains(lower, companyLower) {
				idx := strings.Index(lower, companyLower)
				issues = append(issues, contentIssue{
					Category: "contamination",
					Detail:   fmt.Sprintf("Found '%s' in content for '%s'", known.Company, expectedCompany),
					Snippet:  extractSnippet(html, idx, 80),
				})
			}
		}
	}

	return issues
}

func checkContentLength(html string) []contentIssue {
	var issues []contentIssue

	stripped := stripHTMLTags(html)
	textLen := len(strings.TrimSpace(stripped))

	if textLen < 50 {
		issues = append(issues, contentIssue{
			Category: "short_content",
			Detail:   fmt.Sprintf("Content is very short (%d characters of text)", textLen),
			Snippet:  strings.TrimSpace(stripped),
		})
	}

	return issues
}

// ============================================================================
// Helpers
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

func stripHTMLTags(html string) string {
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	return tagRegex.ReplaceAllString(html, " ")
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func sanitiseKey(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '_'
		}
		return -1
	}, s)
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
