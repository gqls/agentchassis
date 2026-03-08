// FILE: platform/orchestration/actions/write_audit_findings_action.go
//
// WriteAuditFindingsAction takes structured findings from an LLM audit
// and creates properly classified site_work_items.
//
// The LLM audit produces raw observations: category, description, severity,
// affected pages. This action does the classification:
//
//   1. Loads existing pages for the site (one query, cached for all findings)
//   2. For each finding, determines:
//      - Does the referenced page exist? → content fix on existing page
//      - Is the page name a placeholder ("new page needed", "site-wide")? → classify by category
//      - Is it a gap that needs a new page? → route to content-gap-planner
//      - Is it a metadata/config issue? → route to spec updater
//      - Is it a design issue? → route to design handler
//   3. Creates work items with correct item_type, handler_agent, and page_id
//
// Registration:
//   "write_audit_findings": {
//       Handler:     WriteAuditFindingsAction,
//       Category:    "site",
//       Description: "Classify and store LLM audit findings as work items",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var WriteAuditFindingsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"findings_field", "audit_source"},
	Defaults:   map[string]interface{}{"audit_source": "design-audit"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_audit_findings", WriteAuditFindingsInputSpec)
}

// ============================================================================
// Classification rules — deterministic routing based on category + page existence
// ============================================================================

// Design categories route to design/CSS handlers regardless of page existence
var designCategories = map[string]struct{}{
	"colour": {}, "color": {}, "spacing": {}, "typography": {},
	"header_footer": {}, "dark_section": {}, "responsive": {},
}

// Categories that indicate site-wide config/metadata issues, not page content
var metadataCategories = map[string]struct{}{
	"metadata": {}, "config": {}, "identity": {}, "spec": {},
	"contact_mismatch": {},
}

// Categories that indicate CTA/component-level fixes on existing pages
var componentCategories = map[string]struct{}{
	"cta": {}, "nav_restructure": {},
}

// Page names that are not real pages — they're audit shorthand
var placeholderPageNames = map[string]bool{
	"new page needed": true,
	"new page":        true,
	"site-wide":       true,
	"general":         true,
	"all pages":       true,
	"multiple pages":  true,
	"":                true,
}

// Normalise common page name aliases
var pageNameAliases = map[string]string{
	"homepage": "index",
	"home":     "index",
}

// Design category → handler agent
var designRouting = map[string]string{
	"colour":        "webdesign-agent",
	"color":         "webdesign-agent",
	"spacing":       "component-template-fixer",
	"typography":    "webdesign-agent",
	"header_footer": "site-component-linker",
	"dark_section":  "color-variable-fixer",
	"responsive":    "component-template-fixer",
}

// Design category → item type
var designItemTypes = map[string]string{
	"colour":        "needs_design_review",
	"color":         "needs_design_review",
	"spacing":       "spacing_fix",
	"typography":    "needs_design_review",
	"header_footer": "header_footer_fix",
	"dark_section":  "hardcoded_section_colors",
	"responsive":    "responsive_fix",
}

// ============================================================================
// Page info cache — loaded once per invocation
// ============================================================================

type pageInfo struct {
	ID       uuid.UUID
	Name     string
	PageType string
	Sections string
}

func loadSitePages(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]pageInfo, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, COALESCE(page_type, ''), COALESCE(sections::text, '[]')
		FROM pages
		WHERE site_id = $1 AND status = 'active'
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make(map[string]pageInfo)
	for rows.Next() {
		var p pageInfo
		if err := rows.Scan(&p.ID, &p.Name, &p.PageType, &p.Sections); err != nil {
			continue
		}
		pages[p.Name] = p
	}
	return pages, nil
}

// ============================================================================
// Finding struct — what the LLM produces
// ============================================================================

type auditFinding struct {
	Category          string   `json:"category"`
	Severity          string   `json:"severity"`
	Description       string   `json:"description"`
	Suggestion        string   `json:"suggestion"`
	Page              string   `json:"page"`
	AffectedPages     []string `json:"affected_pages"`
	FixType           string   `json:"fix_type"`
	AffectedComponent string   `json:"affected_component"`
}

// ============================================================================
// Classified result — what we insert as a work item
// ============================================================================

type classifiedFinding struct {
	ItemType     string
	HandlerAgent string
	Severity     string
	Priority     int
	PageID       *uuid.UUID
	PageName     string
	Spec         map[string]interface{}
	DedupKey     string
}

// ============================================================================
// classify — the core algorithmic routing
// ============================================================================

func classifyFinding(f auditFinding, pages map[string]pageInfo, siteID uuid.UUID, auditSource string) classifiedFinding {
	category := strings.ToLower(strings.TrimSpace(f.Category))
	pageName := strings.TrimSpace(f.Page)

	// Normalise page name aliases
	if alias, ok := pageNameAliases[strings.ToLower(pageName)]; ok {
		pageName = alias
	}

	// Use first affected_page if page is empty
	if pageName == "" && len(f.AffectedPages) > 0 {
		pageName = strings.TrimSpace(f.AffectedPages[0])
		if alias, ok := pageNameAliases[strings.ToLower(pageName)]; ok {
			pageName = alias
		}
	}

	severity := f.Severity
	if severity == "" {
		severity = "medium"
	}
	priority := severityToPriority(severity)

	// Build base spec (always included)
	spec := map[string]interface{}{
		"audit_source":    auditSource,
		"category":        f.Category,
		"description":     f.Description,
		"original_domain": "build",
	}
	if f.Suggestion != "" {
		spec["suggestion"] = f.Suggestion
	}
	if f.AffectedComponent != "" {
		spec["affected_component"] = f.AffectedComponent
	}
	if f.FixType != "" {
		spec["fix_type"] = f.FixType
	}

	// ── Rule 1: Design categories → design handlers (page existence irrelevant)
	if _, isDesign := designCategories[category]; isDesign {
		handler := designRouting[category]
		itemType := designItemTypes[category]
		if handler == "" {
			handler = "component-template-fixer"
		}
		if itemType == "" {
			itemType = "needs_design_review"
		}

		var pageID *uuid.UUID
		if pageName != "" {
			if p, exists := pages[pageName]; exists {
				pageID = &p.ID
			}
		}
		spec["page_name"] = pageName

		return classifiedFinding{
			ItemType:     itemType,
			HandlerAgent: handler,
			Severity:     severity,
			Priority:     priority,
			PageID:       pageID,
			PageName:     pageName,
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_%s_%s_%s", auditSource, itemType, pageName, siteID),
		}
	}

	// ── Rule 2: Metadata/config categories → spec update (not a page issue)
	if _, isMeta := metadataCategories[category]; isMeta {
		spec["page_name"] = ""
		return classifiedFinding{
			ItemType:     "needs_spec_update",
			HandlerAgent: "spec-updater",
			Severity:     severity,
			Priority:     priority,
			PageID:       nil,
			PageName:     "",
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_needs_spec_update_%s_%s", auditSource, category, siteID),
		}
	}

	// ── Rule 3: Component-level categories on existing pages
	if _, isComponent := componentCategories[category]; isComponent {
		var pageID *uuid.UUID
		if p, exists := pages[pageName]; exists {
			pageID = &p.ID
		}
		itemType := "cta_improvement"
		if category == "nav_restructure" {
			itemType = "nav_restructure"
		}
		spec["page_name"] = pageName

		return classifiedFinding{
			ItemType:     itemType,
			HandlerAgent: "component-template-fixer",
			Severity:     severity,
			Priority:     priority,
			PageID:       pageID,
			PageName:     pageName,
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_%s_%s_%s", auditSource, itemType, pageName, siteID),
		}
	}

	// ── Rule 4: Content categories — depends on whether the page exists
	isPlaceholder := placeholderPageNames[strings.ToLower(pageName)]

	if !isPlaceholder {
		if p, exists := pages[pageName]; exists {
			// Page exists → content rewrite on existing page
			pageID := p.ID
			itemType := "content_rewrite"
			if category == "tone" {
				itemType = "tone_shift"
			}
			spec["page_name"] = pageName

			return classifiedFinding{
				ItemType:     itemType,
				HandlerAgent: "page-build-handler",
				Severity:     severity,
				Priority:     priority,
				PageID:       &pageID,
				PageName:     pageName,
				Spec:         spec,
				DedupKey:     fmt.Sprintf("%s_%s_%s_%s", auditSource, itemType, pageName, siteID),
			}
		}
	}

	// ── Rule 5: Gap — page doesn't exist or is a placeholder
	// Route to content-gap-planner which decides what to create
	if category == "gap" || category == "content" || category == "differentiation" || category == "structure" {
		spec["page_name"] = pageName

		return classifiedFinding{
			ItemType:     "needs_content_planning",
			HandlerAgent: "content-gap-planner",
			Severity:     severity,
			Priority:     priority + 5,
			PageID:       nil,
			PageName:     "",
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_needs_content_planning_%s_%s", auditSource, sanitiseDedupSegment(f.Description), siteID),
		}
	}

	// ── Rule 6: Tone/content on placeholder page → also goes to planner
	if isPlaceholder && (category == "tone" || category == "content_rewrite") {
		spec["page_name"] = ""

		return classifiedFinding{
			ItemType:     "needs_content_planning",
			HandlerAgent: "content-gap-planner",
			Severity:     severity,
			Priority:     priority + 5,
			PageID:       nil,
			PageName:     "",
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_needs_content_planning_%s_%s", auditSource, sanitiseDedupSegment(f.Description), siteID),
		}
	}

	// ── Fallback: unknown category → content-gap-planner for triage
	spec["page_name"] = pageName
	return classifiedFinding{
		ItemType:     "audit_finding_" + category,
		HandlerAgent: "content-gap-planner",
		Severity:     severity,
		Priority:     priority + 10,
		PageID:       nil,
		PageName:     pageName,
		Spec:         spec,
		DedupKey:     fmt.Sprintf("%s_audit_%s_%s_%s", auditSource, category, sanitiseDedupSegment(pageName), siteID),
	}
}

func severityToPriority(severity string) int {
	switch severity {
	case "high":
		return 10
	case "low":
		return 60
	default:
		return 30
	}
}

func sanitiseDedupSegment(s string) string {
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
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

// ============================================================================
// Main action
// ============================================================================

func WriteAuditFindingsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_audit_findings"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, WriteAuditFindingsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	auditSource := inputs.Get("audit_source")
	if auditSource == "" {
		auditSource = "design-audit"
	}

	// ── Load existing pages for classification ──
	pages, err := loadSitePages(ctx, params.DB, siteID)
	if err != nil {
		logger.Warn("Failed to load pages for classification, continuing without",
			zap.Error(err))
		pages = make(map[string]pageInfo)
	}
	logger.Info("Loaded site pages for classification",
		zap.Int("page_count", len(pages)),
		zap.String("site_id", siteIDStr))

	// ── Extract findings from LLM response ──
	findingsField := "audit_result.result"
	if f, ok := config["findings_field"].(string); ok && f != "" {
		findingsField = f
	}

	findingsRaw := datahelpers.ExtractNestedField(params.CollectedData, findingsField)
	if findingsRaw == nil {
		for _, alt := range []string{"audit_result.response.result", "audit_result", "review_result.result"} {
			findingsRaw = datahelpers.ExtractNestedField(params.CollectedData, alt)
			if findingsRaw != nil {
				break
			}
		}
	}

	if findingsRaw == nil {
		logger.Warn("No findings found", zap.String("field", findingsField))
		return map[string]interface{}{
			"items_created": 0,
			"reason":        "no findings in " + findingsField,
		}, nil
	}

	// ── Parse findings ──
	var findings []auditFinding

	switch v := findingsRaw.(type) {
	case string:
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)

		if err := json.Unmarshal([]byte(cleaned), &findings); err != nil {
			var wrapper map[string]json.RawMessage
			if err2 := json.Unmarshal([]byte(cleaned), &wrapper); err2 == nil {
				if findingsJSON, ok := wrapper["findings"]; ok {
					json.Unmarshal(findingsJSON, &findings)
				}
			}
			if len(findings) == 0 {
				logger.Warn("Failed to parse findings JSON",
					zap.Error(err),
					zap.String("raw_preview", cleaned[:min(200, len(cleaned))]))
				return map[string]interface{}{
					"items_created": 0,
					"parse_error":   err.Error(),
				}, nil
			}
		}
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				f := auditFinding{
					Category:          getStringFromMap(m, "category"),
					Severity:          getStringFromMap(m, "severity"),
					Description:       getStringFromMap(m, "description"),
					Suggestion:        getStringFromMap(m, "suggestion"),
					Page:              getStringFromMap(m, "page"),
					FixType:           getStringFromMap(m, "fix_type"),
					AffectedComponent: getStringFromMap(m, "affected_component"),
				}
				if ap, ok := m["affected_pages"].([]interface{}); ok {
					for _, p := range ap {
						if ps, ok := p.(string); ok {
							f.AffectedPages = append(f.AffectedPages, ps)
						}
					}
				}
				findings = append(findings, f)
			}
		}
	}

	if len(findings) == 0 {
		return map[string]interface{}{"items_created": 0, "reason": "no valid findings"}, nil
	}

	logger.Info("Parsed audit findings",
		zap.Int("count", len(findings)),
		zap.String("source", auditSource))

	// ── Load blocked items for filtering ──
	blockedKeys := make(map[string]bool)
	blockedRows, err := params.DB.QueryContext(ctx, `
		SELECT item_key FROM site_work_items
		WHERE site_id = $1 AND status = 'blocked' AND item_key IS NOT NULL
	`, siteID)
	if err == nil {
		for blockedRows.Next() {
			var key string
			if blockedRows.Scan(&key) == nil {
				blockedKeys[key] = true
			}
		}
		blockedRows.Close()
	}

	// ── Classify and insert ──
	batchID := uuid.New()
	created := 0
	skipped := 0
	skippedBlocked := 0
	classificationStats := make(map[string]int)

	for _, f := range findings {
		classified := classifyFinding(f, pages, siteID, auditSource)
		classificationStats[classified.ItemType]++

		logger.Info("Classified finding",
			zap.String("category", f.Category),
			zap.String("page", f.Page),
			zap.String("→item_type", classified.ItemType),
			zap.String("→handler", classified.HandlerAgent),
			zap.String("→page_name", classified.PageName),
			zap.Bool("→has_page_id", classified.PageID != nil))

		// Skip blocked
		if blockedKeys[classified.DedupKey] {
			skippedBlocked++
			continue
		}

		// Broader blocked check
		var isBlocked bool
		params.DB.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND status = 'blocked'
				  AND item_type = $2
				  AND ($3::uuid IS NULL OR page_id = $3::uuid)
			)
		`, siteID, classified.ItemType, classified.PageID).Scan(&isBlocked)

		if isBlocked {
			skippedBlocked++
			continue
		}

		// Dedup check
		var exists bool
		params.DB.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND item_key = $2
				  AND status IN ('detected', 'triaged', 'claimed', 'blocked')
			)
		`, siteID, classified.DedupKey).Scan(&exists)

		if exists {
			skipped++
			continue
		}

		specJSON, _ := json.Marshal(classified.Spec)

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, domain, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by,
				item_key, batch_id
			) VALUES ($1, $2, 'build', $3, $4, $5, $6::jsonb, $7, $8, $9, 'detected', $10, $11, $12)
		`, siteID, "discovery", classified.ItemType, classified.Severity,
			f.Description, string(specJSON), classified.PageID, classified.Priority,
			classified.HandlerAgent, auditSource, classified.DedupKey, batchID)

		if err != nil {
			logger.Warn("Failed to insert finding work item",
				zap.String("category", f.Category),
				zap.String("item_type", classified.ItemType),
				zap.Error(err))
			continue
		}
		created++
	}

	logger.Info("WriteAuditFindingsAction: Complete",
		zap.Int("created", created),
		zap.Int("skipped_duplicates", skipped),
		zap.Int("skipped_blocked", skippedBlocked),
		zap.Int("total_findings", len(findings)),
		zap.Any("classification_stats", classificationStats))

	return map[string]interface{}{
		"items_created":         created,
		"items_skipped":         skipped,
		"items_skipped_blocked": skippedBlocked,
		"total_findings":        len(findings),
		"batch_id":              batchID.String(),
		"audit_source":          auditSource,
		"classification_stats":  classificationStats,
	}, nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
