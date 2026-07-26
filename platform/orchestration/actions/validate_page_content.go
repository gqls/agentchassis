// FILE: platform/orchestration/actions/validate_page_content.go
//
// ValidatePageContentAction checks generated page content for issues
// before it gets saved and deployed.
//
// Checks performed:
//   1. Placeholder text — "To be added", "Lorem ipsum", "[insert", etc.
//   2. Unrendered templates — {{.field}}, {{range}}, {{if}}
//   3. Cross-site contamination — wrong domain/company name in content
//   4. Broken internal links — hrefs pointing to non-existent pages. These are
//      REPAIRED here rather than merely reported (bugs_open/079): a target that
//      exists at its .html form has its href rewritten to the stored pages.url;
//      anything else has the <a> removed and its text kept. Opt out per call
//      with repair_internal_links:false.
//   5. Hallucinated emails — email addresses that don't match site's contact
//   6. Content length — suspiciously short content
//   7. LLM meta-commentary — model prose about its own task/inability
//      persisted as page content (robot-hands 2026-07-14: a section stored
//      "The data schema for this section requires product array data…"
//      as its rendered copy)
//   8. Claims vs evidence base (opt-in per site) — when the site has a
//      site_specs 'evidence_base' aspect: banned-claim patterns (the site's
//      own audited-out fabrications; blocker — these are KNOWN falsehoods)
//      and unregistered numbers (numbers asserted as facts about the
//      business that no evidence_base fact supports; error — extraction has
//      false positives, and error already routes to a human). Sites without
//      an evidence_base skip both silently. Scans parse assertion TEXT
//      NODES, never raw HTML (an email/number in a placeholder= attribute or
//      <code> sample is not an assertion). See SPEC_claims_verification.
//
// Returns:
//   - valid: bool (false if any blockers or errors found)
//   - clean_html: HTML with stray comments stripped AND dead internal links
//     repaired — this is what save_sections persists, so a repair here is what
//     ships
//   - issues: array of issues with category, severity, detail
//   - blocker/warning/error counts
//   - links_rewritten / links_unlinked / link_repairs: what the repair pass did
//
// Blockers (prevent deployment): placeholder, template, contamination,
//	meta-commentary, banned claim, placeholder email
//
// Errors (prevent deployment): invalid/unregistered email, unregistered number
//
// Warnings (do NOT prevent deployment): short_content, and the link findings —
// which are warnings BECAUSE the repair pass has already removed the defect from
// clean_html, not because a dead link is acceptable. See validateInternalLinks.
//
// CORRECTED 2026-07-26: this legend previously read "Errors (prevent
// deployment): broken_link" and had never been true — broken_link has always
// been filed at warning severity, and warnings have never been counted by
// `valid`. That gap between the documented and actual policy is bugs_open/079.
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
	"encoding/json"
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
	// Note: bare "placeholder" was previously in this list but produced
	// false positives on legitimate HTML attributes (e.g.
	// <input placeholder="...">) and CSS class names. The bracketed
	// markers above and the labelled placeholder phrases below remain;
	// they're unambiguous.
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
	{"human review required", "human review marker"},
	{"review required", "review marker"},
	{"name needed", "placeholder prompt"},
	{"title needed", "placeholder prompt"},
	{"photo needed", "placeholder prompt"},
	{"name needed", "generic placeholder"},
	{"content needed", "placeholder prompt"},
	{"bio needed", "placeholder prompt"},
	{"details needed", "placeholder prompt"},
	{"image needed", "placeholder prompt"},
	{"tbd", "to be determined"}, // without bracket prefix
	{"not provided", "missing data marker"},
	{"not yet available", "placeholder"},
	{"details pending", "placeholder"},
	{"human review required", "human review marker"},
	{"description needed", "placeholder prompt"},
	{"information needed", "placeholder prompt"},
	{"text needed", "placeholder prompt"},
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
	checkClaims := configBoolOrDefault(config, "check_claims", true)
	// repair_internal_links is the reversal lever for bugs_open/079's fleet-wide
	// content change. It defaults ON — an off-by-default repair would be inert
	// and the bug would still be live — but DB config is live-immediately, so
	// the behaviour can be withdrawn fleet-wide without waiting for an image
	// roll. Repair is additionally gated on check_internal_links: the repair
	// acts on what that check found, so it cannot run without it.
	repairLinks := configBoolOrDefault(config, "repair_internal_links", true)
	checkStatClaims := configBoolOrDefault(config, "check_stat_claims", true)
	checkStatUnits := configBoolOrDefault(config, "check_stat_units", true)

	// ── Run all checks ──
	var issues []ValidationIssue

	// 1. Placeholder text
	issues = append(issues, checkPlaceholderPatterns(htmlStr)...)

	// 2. Unrendered templates
	issues = append(issues, checkUnrenderedTemplates(htmlStr)...)

	// 3. Cross-site contamination
	if domain != "" {
		// A portfolio / meta site may legitimately name another of our sites
		// (owner-approved case studies). Load this site's opt-in allowlist so
		// those references are not misread as contamination (bugs_open/055).
		var allowedRefs map[string]bool
		if params.DB != nil && siteIDStr != "" {
			if siteID, err := uuid.Parse(siteIDStr); err == nil {
				allowedRefs = loadAllowedReferenceDomains(ctx, params.DB, siteID, logger)
			}
		}
		issues = append(issues, checkDomainContamination(htmlStr, domain, companyName, allowedRefs)...)
	}

	// 4. Broken internal links.
	// The page set is loaded ONCE here and reused by the repair pass below, so
	// detection and repair can never disagree about what a real page is.
	// pageIndexOK false means the query failed — in that case we neither flag
	// nor repair. Flagging against a half-loaded set produces a phantom warning
	// for every link on the page; repairing against one would strip them all.
	var pageIndex datahelpers.PageURLIndex
	pageIndexOK := false
	if checkLinks && params.DB != nil && siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			pageIndex, pageIndexOK = loadValidPagePaths(ctx, params.DB, siteID, logger)
			if pageIndexOK {
				issues = append(issues, validateInternalLinks(htmlStr, pageIndex)...)
			}
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

	// 7. LLM meta-commentary persisted as content
	issues = append(issues, checkMetaCommentary(htmlStr)...)

	// 8. Claims vs evidence base (opt-in: no-op when the site has none)
	if checkClaims && params.DB != nil && siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			if eb := loadEvidenceBase(ctx, params.DB, siteID, logger); eb != nil {
				blocks := datahelpers.ExtractAssertionText(htmlStr)
				issues = append(issues, checkBannedClaims(blocks, eb)...)
				issues = append(issues, checkUnregisteredNumbers(blocks, eb)...)
			}
		}
	}

	// 9. Stat fields vs evidence base, over content_data rather than rendered
	//    HTML (bugs_open/043 candidates 2 and 4). Check 8 cannot see a stat
	//    card: the value and its label are separate block elements, so the
	//    number's claim window is the bare figure. See
	//    validate_page_content_stats.go and datahelpers/claims_stats.go.
	//    Silent no-op for callers whose workflow produces no sections_metadata.
	if checkStatClaims || checkStatUnits {
		issues = append(issues, runStatChecks(ctx, params.DB, params.CollectedData,
			config, siteIDStr, checkStatClaims, checkStatUnits, logger)...)
	}

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

		// Persist the structured issue list to agent_error_log so post-mortem
		// debugging doesn't require pod-log access. The chassis will write its
		// own generic "step failed" row after this action returns; this insert
		// adds the concrete blocker/error details that the chassis row lacks.
		// Failure to write the log is itself only a Warn — we never want
		// logging-failure to mask the actual validation failure.
		writeValidationFailureLog(ctx, params, siteIDStr, domain,
			issues, blockerCount, errorCount, logger)

		return nil, fmt.Errorf("content validation failed: %d blockers, %d errors", blockerCount, errorCount)
	}

	// ── Clean HTML — strip stray comments ──
	commentRegex := regexp.MustCompile(`<!--[\s\S]*?-->`)
	cleanHTML := commentRegex.ReplaceAllString(htmlStr, "")
	doubleNewline := regexp.MustCompile(`\n\s*\n\s*\n`)
	cleanHTML = doubleNewline.ReplaceAllString(cleanHTML, "\n\n")

	// ── Repair the dead internal links check 4 just found (bugs_open/079) ──
	// Deliberately AFTER every check has run against the original htmlStr, so
	// no other check's input is changed by this pass; and against cleanHTML, so
	// the repair lands in the string save_sections actually persists.
	var repairs []datahelpers.LinkRepair
	if repairLinks && pageIndexOK {
		cleanHTML, repairs = datahelpers.RepairPageLinks(cleanHTML, pageIndex)
	}
	rewritten, unlinked := countLinkRepairs(repairs)
	annotateLinkRepairs(issues, repairs)

	if len(repairs) > 0 {
		logger.Info("ValidatePageContentAction: repaired dead internal links",
			zap.Int("rewritten", rewritten),
			zap.Int("unlinked", unlinked),
			zap.String("domain", domain))
		// A pod log line is not a record. Persist what we DID, not only what we
		// saw — bugs_open/071 gap 3: on the success path this action wrote
		// nothing durable, and collected_data is pruned at ~24h.
		writeLinkRepairLog(ctx, params, siteIDStr, domain, repairs, rewritten, unlinked, logger)
	}

	// Build issues list for output — after the repair pass, so each link issue
	// carries what was done about it, not just that it was seen.
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

	repairMaps := make([]map[string]string, 0, len(repairs))
	for _, r := range repairs {
		repairMaps = append(repairMaps, map[string]string{
			"href":     r.Href,
			"new_href": r.NewHref,
			"action":   r.Action,
		})
	}

	return map[string]interface{}{
		"valid":           valid,
		"clean_html":      cleanHTML,
		"blockers":        blockerCount,
		"errors":          errorCount,
		"warnings":        warningCount,
		"issue_count":     len(issues),
		"issues":          issuesMaps,
		"checked_links":   countInternalLinks(htmlStr),
		"checked_emails":  countEmailAddresses(htmlStr),
		"links_rewritten": rewritten,
		"links_unlinked":  unlinked,
		"link_repairs":    repairMaps,
	}, nil
}

// countLinkRepairs splits the repair list by action.
func countLinkRepairs(repairs []datahelpers.LinkRepair) (rewritten, unlinked int) {
	for _, r := range repairs {
		switch r.Action {
		case datahelpers.LinkRepairRewrite:
			rewritten++
		case datahelpers.LinkRepairUnlink:
			unlinked++
		}
	}
	return rewritten, unlinked
}

// annotateLinkRepairs records on each link issue what the repair pass did about
// it. Type and Severity are deliberately UNCHANGED: bugs_open/071's evidence
// trail and the post-deploy audit both key on "phantom_link", and a fix that
// renames the finding would silently empty every query that looks for it.
func annotateLinkRepairs(issues []ValidationIssue, repairs []datahelpers.LinkRepair) {
	if len(repairs) == 0 {
		return
	}
	byHref := make(map[string]datahelpers.LinkRepair, len(repairs))
	for _, r := range repairs {
		byHref[r.Href] = r
	}
	for i := range issues {
		if issues[i].Category != "link" {
			continue
		}
		r, ok := byHref[issues[i].Value]
		if !ok {
			continue
		}
		switch r.Action {
		case datahelpers.LinkRepairRewrite:
			issues[i].Expected = r.NewHref
			issues[i].Description += fmt.Sprintf(" — href rewritten to %s before save", r.NewHref)
		case datahelpers.LinkRepairUnlink:
			issues[i].Description += " — link removed before save, anchor text kept"
		}
	}
}

// ============================================================================
// Structured failure logging — agent_error_log
// ============================================================================
//
// When validation produces blockers, the action returns a generic error
// message. The chassis catches that and writes a row to agent_error_log,
// but with no detail beyond "1 blockers, 0 errors". For post-mortem
// debugging we want the actual blocker descriptions in the database —
// not just in pod logs that may have rotated by the time we look.
//
// writeValidationFailureLog inserts a sibling row before the action
// returns its error. The row's severity is 'warning' (the chassis row
// is the canonical 'error') and its error_code is dedicated to this
// detail-write so queries can join on it. The context JSONB carries the
// full structured issue list.
//
// This is best-effort: if the insert fails, we log the secondary error
// at Warn and proceed. Never let logging-failure mask the validation
// failure itself.

const validationDetailErrorCode = "CONTENT_VALIDATION_BLOCKER_DETAIL"

func writeValidationFailureLog(
	ctx context.Context,
	params ActionParams,
	siteIDStr string,
	domain string,
	issues []ValidationIssue,
	blockerCount, errorCount int,
	logger *zap.Logger,
) {
	if params.DB == nil {
		return
	}

	// Build the issues payload — only blockers and errors. Warnings are
	// not why the validation failed, so they don't belong in the failure
	// detail row. (They're still in the action's return map on success.)
	failureIssues := make([]map[string]string, 0, blockerCount+errorCount)
	for _, issue := range issues {
		if issue.Severity != "blocker" && issue.Severity != "error" {
			continue
		}
		failureIssues = append(failureIssues, map[string]string{
			"type":        issue.Type,
			"category":    issue.Category,
			"severity":    issue.Severity,
			"location":    issue.Location,
			"value":       issue.Value,
			"description": issue.Description,
		})
	}

	contextData := map[string]interface{}{
		"blocker_count": blockerCount,
		"error_count":   errorCount,
		"issues":        failureIssues,
		"page_name":     datahelpers.ExtractNestedFieldString(params.CollectedData, "page_record.name"),
	}
	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		logger.Warn("ValidatePageContentAction: failed to marshal failure context", zap.Error(err))
		return
	}

	// site_id is uuid; only include if parseable.
	var siteIDArg interface{}
	if siteIDStr != "" {
		if id, err := uuid.Parse(siteIDStr); err == nil {
			siteIDArg = id
		}
	}

	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action,
		     error_message, error_code, severity, context)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, 'warning', $8::jsonb)
	`,
		siteIDArg,
		domain,
		"page-build-handler", // best-effort; the action runs under this agent
		"validate_content",
		"validate_page_content",
		fmt.Sprintf("Validation produced %d blocker(s) and %d error(s); see context.issues for detail",
			blockerCount, errorCount),
		validationDetailErrorCode,
		string(contextJSON),
	)
	if err != nil {
		logger.Warn("ValidatePageContentAction: failed to write structured failure log",
			zap.Error(err))
		return
	}

	logger.Info("ValidatePageContentAction: wrote structured failure log",
		zap.Int("blocker_count", blockerCount),
		zap.Int("error_count", errorCount))
}

// linkRepairErrorCode is DELIBERATELY distinct from validationDetailErrorCode.
// The two rows answer different questions — "why did this build fail" versus
// "what did the gate change on a build that succeeded" — and existing queries
// filtering on the blocker code must keep returning exactly what they returned
// before this change.
const linkRepairErrorCode = "CONTENT_LINK_REPAIR_DETAIL"

// writeLinkRepairLog persists what the repair pass DID, on the success path.
//
// bugs_open/071 gap 3: a page whose only findings were warnings wrote nothing
// durable at all. The issue list survived solely in collected_data, which
// database-cleanup prunes at ~24h, and in one pod-log line carrying the COUNT
// but not the hrefs — so a day later, the fact that the platform knew about
// eight specific broken links was unrecoverable.
//
// This writes a work RECORD, not a work ITEM, and that is a deliberate choice.
// A site_work_items row would promise a repair that nothing performs:
// bugs_open/083 measured phantom_internal_link detected 22 times and fixed zero
// times ever, because the only thing that promotes 'detected' rows lives in the
// disabled improvement-sweep; and bugs_open/077 warns specifically against
// filing items whose handler has no remit. The link is already repaired by the
// time this runs — what is owed is an auditable account, not a queue entry.
//
// Best-effort, like its sibling: a logging failure must never fail a build whose
// content is already correct.
func writeLinkRepairLog(
	ctx context.Context,
	params ActionParams,
	siteIDStr string,
	domain string,
	repairs []datahelpers.LinkRepair,
	rewritten, unlinked int,
	logger *zap.Logger,
) {
	if params.DB == nil || len(repairs) == 0 {
		return
	}

	repairMaps := make([]map[string]string, 0, len(repairs))
	for _, r := range repairs {
		repairMaps = append(repairMaps, map[string]string{
			"href":     r.Href,
			"new_href": r.NewHref,
			"action":   r.Action,
		})
	}

	contextData := map[string]interface{}{
		"rewritten": rewritten,
		"unlinked":  unlinked,
		"repairs":   repairMaps,
		"page_name": datahelpers.ExtractNestedFieldString(params.CollectedData, "page_record.name"),
		"page_url":  datahelpers.ExtractNestedFieldString(params.CollectedData, "page_record.url"),
	}
	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		logger.Warn("ValidatePageContentAction: failed to marshal link repair context", zap.Error(err))
		return
	}

	var siteIDArg interface{}
	if siteIDStr != "" {
		if id, err := uuid.Parse(siteIDStr); err == nil {
			siteIDArg = id
		}
	}

	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action,
		     error_message, error_code, severity, context)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, 'warning', $8::jsonb)
	`,
		siteIDArg,
		domain,
		"page-build-handler", // best-effort; the action runs under this agent
		"validate_content",
		"validate_page_content",
		fmt.Sprintf("Repaired %d dead internal link(s) before save: %d href(s) rewritten, %d link(s) removed; see context.repairs",
			len(repairs), rewritten, unlinked),
		linkRepairErrorCode,
		string(contextJSON),
	)
	if err != nil {
		logger.Warn("ValidatePageContentAction: failed to write link repair log", zap.Error(err))
		return
	}

	logger.Info("ValidatePageContentAction: wrote link repair log",
		zap.Int("rewritten", rewritten),
		zap.Int("unlinked", unlinked))
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

func checkDomainContamination(html string, expectedDomain string, expectedCompany string, allowedRefs map[string]bool) []ValidationIssue {
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

		// Owner-approved cross-reference: a portfolio / meta site (e.g.
		// fundamentallyai.com) may name another of our sites on purpose — as a
		// case study, or as the worked example of the self-correction story.
		// When this site's allowlist names the known domain, the reference is
		// intentional, not contamination; skip BOTH its domain and company
		// checks. Company suppression is deliberate: it is the SAME first-party
		// site, so a rewrite that names the company ("Leopardess Consulting")
		// rather than the bare domain must not re-block the same approved
		// reference. This does not widen blast radius — only the five hardcoded
		// first-party known sites can ever be suppressed. Absent allowlist →
		// nil map → every known site still flagged, so this is fully opt-in and
		// unchanged for sites that have not declared one. See bugs_open/055.
		if allowedRefs[strings.ToLower(known.Domain)] {
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

func validateInternalLinks(html string, validPages datahelpers.PageURLIndex) []ValidationIssue {
	var issues []ValidationIssue

	// Href extraction, scope classification and page-path normalisation are the
	// shared datahelpers definitions — the same ones the post-deploy audit
	// (check_phantom_internal_links) uses, so the gate and the audit agree on
	// what counts as an internal page link and what resolves to a real page.
	for _, href := range datahelpers.ExtractHrefs(html) {
		switch datahelpers.ClassifyLinkScope(href) {
		case datahelpers.LinkScopeEmpty:
			// An empty href renders as a dead link/button (e.g. an unpopulated
			// "Browse All X" CTA). Non-blocking — see the policy note below.
			issues = append(issues, ValidationIssue{
				Type:        "empty_internal_href",
				Category:    "link",
				Severity:    "warning",
				Location:    `href=""`,
				Description: "Empty href — link/button has no destination",
			})
		case datahelpers.LinkScopePage:
			// POLICY (rewritten 2026-07-26, bugs_open/079 + 071 candidate 5).
			//
			// A missing internal target is filed at WARNING, so it does not
			// stop the deploy. That is now true for a different reason than it
			// used to be: RepairPageLinks removes the defect from clean_html
			// before save_sections persists it, so what ships has no dead link
			// in it. The finding is recorded, not deferred.
			//
			// The previous justification — "the improvement loop resolves it" —
			// was FALSE for the whole time it was relied on. That loop has been
			// off since 2026-05 (bugs_open/083: this exact item type was
			// detected 22 times and fixed zero times, ever), so the fail-open
			// deferred to a repairer that does not run, and the page shipped a
			// 404. Two threads read that comment as "handled".
			//
			// If repair_internal_links is turned OFF, warning severity reverts
			// to meaning exactly what it used to: the dead link ships and
			// nothing downstream fixes it. Turn it off only knowing that.
			//
			// We flag only true phantoms: an href with no pages row at all.
			// Planned-but-unbuilt pages have a row (status not
			// deleted/archived) and are tolerated silently.
			if !validPages.Contains(href) {
				issues = append(issues, ValidationIssue{
					Type:        "phantom_link",
					Category:    "link",
					Severity:    "warning",
					Location:    fmt.Sprintf("href=%q", href),
					Value:       href,
					Description: fmt.Sprintf("Link target has no matching page: %s", href),
				})
			}
		}
		// external / anchor / mailto / asset scopes are not page links — skip.
	}

	return issues
}

// isAssetPath / normalizePagePath / isValidPage now live in datahelpers
// (links.go) as IsAssetPath / NormalizePagePath / PageURLSet.Contains, shared
// with the post-deploy phantom-link audit so both agree on classification.

// ============================================================================
// Check 5: Hallucinated emails (from existing code)
// ============================================================================

func validateEmails(ctx context.Context, db *sql.DB, html string, siteID uuid.UUID, logger *zap.Logger) []ValidationIssue {
	return checkEmails(html, loadSiteContactEmail(ctx, db, siteID, logger))
}

// checkEmails is the pure core of check 5, split from the DB load so both
// branches are testable (same shape as checkDomainContamination).
func checkEmails(html, officialEmail string) []ValidationIssue {
	var issues []ValidationIssue

	// Assertion contexts only: text nodes plus mailto: hrefs. An email that
	// exists only in a placeholder= attribute, <code> sample, or script body
	// is an example, not a contact claim — flagging those blocked every build
	// of every page using the shared contact-block (fixed 2026-07-14; the
	// checker itself is fixed here). ExtractAssertionEmails returns
	// lowercased, deduplicated addresses.
	emails := datahelpers.ExtractAssertionEmails(html)

	for _, email := range emails {
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
		} else {
			// No registered contact address means EVERY asserted email is an
			// invention — the site has nothing legitimate to publish. This
			// branch used to fall through, so the sites needing the check
			// most had no protection at all (bugs_open/063: a fabricated
			// mailto deployed and served live for ~4h). Severity error routes
			// the build to review, same as the mismatch case.
			issues = append(issues, ValidationIssue{
				Type:        "invalid_email",
				Category:    "email",
				Severity:    "error",
				Location:    "email address",
				Value:       email,
				Description: fmt.Sprintf("Email '%s' asserted but the site has no registered contact address — no email may be published", email),
			})
		}
	}

	return issues
}

// ============================================================================
// Check 8: Claims vs evidence base (SPEC_claims_verification V1a)
// ============================================================================
//
// Both checks run over assertion text blocks (datahelpers.ExtractAssertionText)
// — never raw HTML — and share their scan engine with the post-deploy audit
// (check_unverified_claims), so the gate and the audit agree by one literal
// implementation on what counts as an asserted claim.

// checkBannedClaims flags the site's own audited-out fabrications. Severity is
// blocker: every pattern in banned_claims is a KNOWN falsehood for this site,
// placed there by a human after an audit ruled the claim out.
func checkBannedClaims(blocks []string, eb *datahelpers.EvidenceBase) []ValidationIssue {
	var issues []ValidationIssue
	for _, f := range eb.ScanBannedClaims(blocks) {
		issues = append(issues, ValidationIssue{
			Type:        "banned_claim",
			Category:    "claims",
			Severity:    "blocker",
			Location:    f.Snippet,
			Value:       f.Matched,
			Description: fmt.Sprintf("Banned claim %q (%s) — %d occurrence(s)", f.Matched, f.Reason, f.Occurrences),
		})
	}
	return issues
}

// checkUnregisteredNumbers flags numbers asserted as facts about the business
// that no evidence_base fact supports. Severity is error — never blocker —
// because number extraction has false positives by design, and error already
// routes the page to a human (mark_needs_review) rather than deploying it.
func checkUnregisteredNumbers(blocks []string, eb *datahelpers.EvidenceBase) []ValidationIssue {
	var issues []ValidationIssue
	for _, f := range eb.ScanUnregisteredNumbers(blocks) {
		issues = append(issues, ValidationIssue{
			Type:        "unregistered_number",
			Category:    "claims",
			Severity:    "error",
			Location:    f.Snippet,
			Value:       f.Matched,
			Description: fmt.Sprintf("Unregistered number %q — %s (%d occurrence(s))", f.Matched, f.Reason, f.Occurrences),
		})
	}
	return issues
}

// loadEvidenceBase reads the site's current evidence_base spec. Returns nil
// when the site has none (the common case — the claims layer is opt-in per
// site) or when the row exists but holds nothing scannable.
func loadEvidenceBase(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) *datahelpers.EvidenceBase {
	var data []byte
	err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, siteID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		logger.Warn("Failed to load evidence_base spec", zap.Error(err))
		return nil
	}
	eb, err := datahelpers.ParseEvidenceBase(data)
	if err != nil {
		logger.Warn("Failed to parse evidence_base spec — claims checks skipped",
			zap.Error(err))
		return nil
	}
	return eb
}

// ============================================================================
// Check 7: LLM meta-commentary persisted as content
// ============================================================================
//
// A model that cannot fulfil a section brief sometimes writes ABOUT the task
// instead of doing it — and that prose then ships as page copy. Live catch
// (robot-hands, 2026-07-14): product-card-with-cta stored "The data schema
// for this section requires product array data sourced from
// query.affiliate_products. Per the schema definition, this field is marked
// required: true with on_missing: skip_section…" as its rendered content.
// The declared skip_section never fired, and the apology deployed.
//
// Patterns are deliberately narrow: schema/pipeline vocabulary and
// first-person refusals never belong in page copy, whereas broader phrases
// ("this section requires…") risk false-positiving legitimate content. A
// false positive here routes to needs_human_review, not silent breakage —
// but keep the list unambiguous anyway.

var metaCommentaryPatterns = []struct {
	Pattern string
	Label   string
}{
	{"as an ai", "first-person AI disclosure"},
	{"as a language model", "first-person AI disclosure"},
	{"i cannot generate", "refusal prose"},
	{"i can't generate", "refusal prose"},
	{"i am unable to generate", "refusal prose"},
	{"i'm unable to generate", "refusal prose"},
	{"i don't have access to", "refusal prose"},
	{"i do not have access to", "refusal prose"},
	{"the data schema", "schema vocabulary in copy"},
	{"per the schema definition", "schema vocabulary in copy"},
	{"input_schema", "schema vocabulary in copy"},
	{"on_missing", "pipeline vocabulary in copy"},
	{"skip_section", "pipeline vocabulary in copy"},
	{"required: true", "schema vocabulary in copy"},
	{"marked `required", "schema vocabulary in copy"},
}

func checkMetaCommentary(html string) []ValidationIssue {
	lower := strings.ToLower(html)
	var issues []ValidationIssue

	for _, p := range metaCommentaryPatterns {
		idx := strings.Index(lower, p.Pattern)
		if idx >= 0 {
			issues = append(issues, ValidationIssue{
				Type:        "meta_commentary",
				Category:    "meta_commentary",
				Severity:    "blocker",
				Location:    extractSnippet(html, idx, 80),
				Value:       p.Pattern,
				Description: fmt.Sprintf("LLM meta-commentary in content: '%s' (%s) — the model wrote about its task instead of doing it", p.Pattern, p.Label),
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

// loadValidPagePaths returns the real page targets for the site, indexed by
// normalised path so a repair can hand back the URL the database actually
// stores. Membership is tested via the shared NormalizePagePath, so the many
// hand-built url variants the old map carried are no longer needed — one normal
// form covers them.
//
// The bool is "this set is TRUSTWORTHY", and it is load-bearing. This function
// used to swallow a query failure and return an empty set, which was survivable
// while the only consequence was a spurious warning per link. It is not
// survivable now: an empty set means every link on the page is a phantom, and
// the repair pass would strip the lot. It also now checks rows.Err(), because a
// mid-iteration failure silently TRUNCATES the set — a partial page list is the
// same hazard wearing a disguise, and it would unlink only some of the page's
// links, which is harder to notice than losing all of them.
func loadValidPagePaths(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (datahelpers.PageURLIndex, bool) {
	rows, err := db.QueryContext(ctx, `
		SELECT url FROM pages
		WHERE site_id = $1 AND status NOT IN ('deleted', 'archived')
	`, siteID)
	if err != nil {
		logger.Warn("Failed to load pages for validation — link check and repair SKIPPED",
			zap.Error(err))
		return nil, false
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		// pages.url is nullable. A NULL is not a link target and not evidence
		// that the list is truncated, so it is skipped rather than treated as a
		// load failure — otherwise one malformed row anywhere on a site would
		// disable link checking for that whole site.
		var url sql.NullString
		if err := rows.Scan(&url); err != nil {
			logger.Warn("Failed to scan page url — link check and repair SKIPPED",
				zap.Error(err))
			return nil, false
		}
		if url.Valid && url.String != "" {
			urls = append(urls, url.String)
		}
	}
	if err := rows.Err(); err != nil {
		logger.Warn("Page list truncated by a row error — link check and repair SKIPPED",
			zap.Error(err))
		return nil, false
	}
	urls = append(urls, "/", "/index.html") // site root is always valid

	logger.Info("ValidatePageContentAction: loaded valid pages",
		zap.Int("page_count", len(urls)))

	return datahelpers.NewPageURLIndex(urls), true
}

func loadSiteContactEmail(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	var email sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(
			NULLIF(s.email, ''),
			NULLIF(s.content_data->>'contact_email', ''),
			NULLIF(s.content_data->'reviewed_brief'->>'contact_email', ''),
			(SELECT NULLIF(ss.data->>'email', '')
			 FROM site_specs ss
			 WHERE ss.site_id = $1 AND ss.aspect = 'identity' AND ss.is_current = true
			 LIMIT 1),
			(SELECT NULLIF(ss.data->>'contact_email', '')
			 FROM site_specs ss
			 WHERE ss.site_id = $1 AND ss.aspect = 'identity' AND ss.is_current = true
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

// loadAllowedReferenceDomains returns the lowercased set of domains this site is
// explicitly allowed to reference in its copy — read from
// sites.content_data->'allowed_reference_domains' (a JSON array of domain
// strings). It is the opt-in escape hatch for portfolio / meta sites whose
// PURPOSE is to name our other sites (e.g. fundamentallyai.com naming
// leopardessconsulting.co.uk as its owner-approved self-correction case study;
// bugs_open/055). Returns nil when the site declares none — so the contamination
// check is unchanged for every site that has not opted in. Data-driven and
// live-editable: adding a domain needs a one-line UPDATE, no image roll.
func loadAllowedReferenceDomains(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) map[string]bool {
	var raw []byte
	err := db.QueryRowContext(ctx, `
		SELECT content_data->'allowed_reference_domains'
		FROM sites WHERE id = $1
	`, siteID).Scan(&raw)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Warn("Failed to load allowed_reference_domains", zap.Error(err))
		}
		return nil
	}
	if len(raw) == 0 {
		return nil // key absent → JSONB null → no allowlist
	}

	var domains []string
	if err := json.Unmarshal(raw, &domains); err != nil {
		logger.Warn("allowed_reference_domains is not a JSON string array — ignored",
			zap.Error(err))
		return nil
	}
	if len(domains) == 0 {
		return nil
	}

	set := make(map[string]bool, len(domains))
	for _, d := range domains {
		if d = strings.ToLower(strings.TrimSpace(d)); d != "" {
			set[d] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	return set
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
