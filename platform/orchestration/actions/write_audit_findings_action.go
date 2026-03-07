// FILE: platform/orchestration/actions/write_audit_findings_action.go
//
// WriteAuditFindingsAction takes structured findings from an LLM audit
// (design-audit-agent or site-review-agent) and creates site_work_items.
//
// Input: findings JSON array from LLM response, either as a string to parse
// or as a pre-parsed []interface{}.
//
// Each finding must have: category, severity, description, suggestion.
// Optional: page, fix_type, affected_component.
//
// The action maps finding categories to handler agents via a configurable
// routing map. Unknown categories default to "component-template-fixer".
//
// Registration:
//   "write_audit_findings": {
//       Handler:     WriteAuditFindingsAction,
//       Category:    "site",
//       Description: "Convert LLM audit findings into site_work_items",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
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

// Default routing: finding category → handler agent
var defaultFindingRouting = map[string]string{
	// Design audit categories
	"colour":        "webdesign-agent",
	"color":         "webdesign-agent",
	"spacing":       "component-template-fixer",
	"typography":    "webdesign-agent",
	"header_footer": "site-component-linker",
	"dark_section":  "color-variable-fixer",
	"responsive":    "component-template-fixer",
	// Site review categories
	"content_rewrite":  "page-build-handler",
	"tone":             "page-build-handler",
	"content":          "page-build-handler",
	"gap":              "page-build-handler",
	"cta":              "component-template-fixer",
	"differentiation":  "page-build-handler",
	"structure":        "page-build-handler",
	"nav_restructure":  "site-component-linker",
	"contact_mismatch": "site-metadata-fixer",
}

// Category → work item type mapping
var defaultItemTypeMapping = map[string]string{
	"colour":           "needs_design_review",
	"color":            "needs_design_review",
	"spacing":          "spacing_fix",
	"typography":       "needs_design_review",
	"header_footer":    "header_footer_fix",
	"dark_section":     "hardcoded_section_colors",
	"responsive":       "responsive_fix",
	"content_rewrite":  "content_rewrite",
	"tone":             "tone_shift",
	"content":          "content_rewrite",
	"gap":              "needs_content_page",
	"cta":              "cta_improvement",
	"differentiation":  "content_rewrite",
	"structure":        "nav_restructure",
	"nav_restructure":  "nav_restructure",
	"contact_mismatch": "contact_info_mismatch",
}

type auditFinding struct {
	Category          string `json:"category"`
	Severity          string `json:"severity"`
	Description       string `json:"description"`
	Suggestion        string `json:"suggestion"`
	Page              string `json:"page"`
	FixType           string `json:"fix_type"`
	AffectedComponent string `json:"affected_component"`
	WorkItemType      string `json:"work_item_type"`
}

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

	// Load custom routing from config if provided
	routing := defaultFindingRouting
	if customRouting, ok := config["handler_routing"].(map[string]interface{}); ok {
		routing = make(map[string]string)
		for k, v := range defaultFindingRouting {
			routing[k] = v
		}
		for k, v := range customRouting {
			if vs, ok := v.(string); ok {
				routing[k] = vs
			}
		}
	}

	// Extract findings from LLM response
	findingsField := "audit_result.result"
	if f, ok := config["findings_field"].(string); ok && f != "" {
		findingsField = f
	}

	findingsRaw := datahelpers.ExtractNestedField(params.CollectedData, findingsField)
	if findingsRaw == nil {
		// Try common alternative paths
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

	// Parse findings — could be a JSON string or a pre-parsed array
	var findings []auditFinding

	switch v := findingsRaw.(type) {
	case string:
		// Strip markdown fences if present
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)

		if err := json.Unmarshal([]byte(cleaned), &findings); err != nil {
			// Try parsing as object with findings array
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
		// Pre-parsed array — convert each element
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
					WorkItemType:      getStringFromMap(m, "work_item_type"),
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

	// Load existing blocked items for this site to skip findings that match
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
	if len(blockedKeys) > 0 {
		logger.Info("WriteAuditFindingsAction: Loaded blocked items for filtering",
			zap.Int("blocked_count", len(blockedKeys)))
	}

	// Insert work items
	batchID := uuid.New()
	created := 0
	skipped := 0
	skippedBlocked := 0

	for _, f := range findings {
		// Determine handler agent
		handler := routing[f.Category]
		if handler == "" {
			handler = "component-template-fixer" // default
		}

		// Determine item type
		itemType := f.WorkItemType
		if itemType == "" {
			itemType = defaultItemTypeMapping[f.Category]
		}
		if itemType == "" {
			itemType = "audit_finding_" + f.Category
		}

		// Determine severity
		severity := f.Severity
		if severity == "" {
			severity = "medium"
		}

		// Build spec
		spec := map[string]interface{}{
			"audit_source":       auditSource,
			"category":           f.Category,
			"description":        f.Description,
			"suggestion":         f.Suggestion,
			"affected_component": f.AffectedComponent,
		}
		if f.FixType != "" {
			spec["fix_type"] = f.FixType
		}
		if f.Page != "" {
			spec["page_name"] = f.Page
		}
		specJSON, _ := json.Marshal(spec)

		// Dedup key
		dedupKey := fmt.Sprintf("%s_%s_%s_%s", auditSource, itemType, f.Page, siteID)

		// Priority: high=10, medium=30, low=60
		priority := 30
		switch severity {
		case "high":
			priority = 10
		case "low":
			priority = 60
		}

		// Skip findings that match existing blocked items
		if blockedKeys[dedupKey] {
			skippedBlocked++
			continue
		}

		// Also check if a blocked item exists with a broader key pattern
		// (e.g. blocked by handler_agent, not by specific audit source)
		var isBlocked bool
		params.DB.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND status = 'blocked'
				  AND item_type = $2
				  AND (page_id IS NULL OR page_id = (
				      SELECT id FROM pages WHERE site_id = $1 AND name = $3 LIMIT 1
				  ))
			)
		`, siteID, itemType, f.Page).Scan(&isBlocked)

		if isBlocked {
			skippedBlocked++
			continue
		}

		// Insert with dedup — skip if pending item already exists
		var exists bool
		params.DB.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND item_key = $2
				  AND status IN ('detected', 'triaged', 'claimed', 'blocked')
			)
		`, siteID, dedupKey).Scan(&exists)

		if exists {
			skipped++
			continue
		}

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, domain, item_type, severity, summary,
				spec, priority, handler_agent, status, created_by,
				item_key, batch_id
			) VALUES ($1, $2, 'build', $3, $4, $5, $6::jsonb, $7, $8, 'detected', $9, $10, $11)
		`, siteID, "discovery", itemType, severity, f.Description,
			string(specJSON), priority, handler, auditSource, dedupKey, batchID)

		if err != nil {
			logger.Warn("Failed to insert finding work item",
				zap.String("category", f.Category),
				zap.Error(err))
			continue
		}
		created++
	}

	logger.Info("WriteAuditFindingsAction: Complete",
		zap.Int("created", created),
		zap.Int("skipped_duplicates", skipped),
		zap.Int("skipped_blocked", skippedBlocked),
		zap.Int("total_findings", len(findings)))

	return map[string]interface{}{
		"items_created":         created,
		"items_skipped":         skipped,
		"items_skipped_blocked": skippedBlocked,
		"total_findings":        len(findings),
		"batch_id":              batchID.String(),
		"audit_source":          auditSource,
	}, nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
