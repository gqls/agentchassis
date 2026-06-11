// FILE: platform/orchestration/actions/discovery_checks/check_empty_sections.go
//
// Detects page sections with empty or near-empty rendered HTML.
// Creates empty_section work items routed to page-build-handler.
//
// CHANGES:
//   - HandlerAgent: "page-build-handler" (was "page-content-writer")
//   - SQL excludes blog/blog-index pages (handled by check_empty_blog.go)
//   - Skips locked components (locked_at IS NULL filter)
//     Human-locked components are intentionally managed — don't flag them.
//
// Registration: automatic via init() → Register(&EmptySectionsCheck{})

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&EmptySectionsCheck{}) }

type EmptySectionsCheck struct{}

func (c *EmptySectionsCheck) Name() string { return "empty_sections" }

func (c *EmptySectionsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	sections, err := findEmptySections(dctx)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":      "empty_sections",
			"count":      len(sections),
			"components": sections,
		}},
	}

	for _, section := range sections {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":              "empty_sections",
			"component_id":       section.ComponentID,
			"page_id":            section.PageID,
			"page_name":          section.PageName,
			"slot_name":          section.SlotName,
			"component_function": section.ComponentFunction,
			"empty_pattern":      section.EmptyPattern,
		})

		var pageIDPtr *uuid.UUID
		if parsed, err := uuid.Parse(section.PageID); err == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "empty_section",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Empty section '%s' on page %s", section.SlotName, section.PageName),
			SpecJSON:     string(specJSON),
			PageID:       pageIDPtr,
			Priority:     100,
			HandlerAgent: "page-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("empty_section:%s:%s", section.PageID, section.SlotName),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

type emptySectionFinding struct {
	ComponentID       string `json:"component_id"`
	PageID            string `json:"page_id"`
	PageName          string `json:"page_name"`
	SlotName          string `json:"slot_name"`
	ComponentFunction string `json:"component_function"`
	HTMLLength        int    `json:"html_length"`
	EmptyPattern      string `json:"empty_pattern"`
}

func findEmptySections(dctx DiscoveryCheckContext) ([]emptySectionFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.id, pc.page_id, p.name, COALESCE(pc.slot_name, ''),
		       COALESCE(cc.function, pc.slot_name, 'unknown'),
		       LENGTH(COALESCE(pc.rendered_html, '')),
		       CASE
		           WHEN pc.rendered_html IS NULL THEN 'null_html'
		           WHEN TRIM(pc.rendered_html) = '' THEN 'empty_html'
		           WHEN LENGTH(pc.rendered_html) < 50 THEN 'minimal_html'
		           WHEN pc.rendered_html ~* '<(h[1-6])[^>]*>\s*</\1>' THEN 'empty_heading'
		           WHEN pc.rendered_html ~* 'class="section[^"]*">\s*</div>' THEN 'empty_container'
		           ELSE 'near_empty'
		       END
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		LEFT JOIN content_components cc ON pc.component_id = cc.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')
		  AND COALESCE(cc.function, '') NOT IN ('header', 'footer', 'head-seo')
		  AND NOT (COALESCE(p.suppressed_sections, '[]'::jsonb) ? COALESCE(pc.slot_name, ''))
		  AND p.name NOT IN ('blog')
		  AND COALESCE(p.page_type, '') NOT IN ('blog-index')
		  AND (
		      pc.rendered_html IS NULL
		      OR TRIM(pc.rendered_html) = ''
		      OR LENGTH(pc.rendered_html) < 50
		      OR pc.rendered_html ~* '<(h[1-6])[^>]*>\s*</\1>'
		  )
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("empty_sections query failed: %w", err)
	}
	defer rows.Close()

	var findings []emptySectionFinding
	for rows.Next() {
		var f emptySectionFinding
		if err := rows.Scan(&f.ComponentID, &f.PageID, &f.PageName, &f.SlotName,
			&f.ComponentFunction, &f.HTMLLength, &f.EmptyPattern); err != nil {
			dctx.Logger.Warn("Failed to scan empty section", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
