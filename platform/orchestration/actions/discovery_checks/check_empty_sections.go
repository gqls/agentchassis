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
//   - Runtime-fill guard (2026-07-10): a section marked data-runtime-fill is
//     DELIBERATELY empty at build time — a browser-side loader fills it (same
//     exemption as sectionHasVisibleContent in rerender_single_page_action.go).
//     Its empty <h2> etc. are the shell the loader populates; routing it to
//     page-build-handler would bake build-time copy into the shell. First live
//     catch: vonc.com's provocation-card and lobby-grid, flagged as
//     'empty_heading' the first pass after enabling the new checks. Reported as
//     findings, never emitted.
//
// Registration: automatic via init() → Register(&EmptySectionsCheck{})

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() {
	Register(&EmptySectionsCheck{})
	RegisterVerifier("empty_section", VerifyEmptySectionResolved)
}

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
		// Runtime-fill shell: deliberately empty at build time; a loader fills it
		// in the browser. Surface it, never send it to the writer.
		if section.IsRuntimeFill {
			dctx.Logger.Info("empty_sections: runtime-fill shell exempt from rebuild",
				zap.String("page", section.PageName),
				zap.String("slot", section.SlotName),
				zap.String("pattern", section.EmptyPattern))
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":     "empty_sections",
				"page":      section.PageName,
				"slot_name": section.SlotName,
				"skipped":   true,
				"reason":    "data-runtime-fill shell — empty by design, filled client-side",
			})
			continue
		}

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
	IsRuntimeFill     bool   `json:"is_runtime_fill"`
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
		       END,
		       COALESCE(pc.rendered_html, '') LIKE '%data-runtime-fill%'
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
			&f.ComponentFunction, &f.HTMLLength, &f.EmptyPattern, &f.IsRuntimeFill); err != nil {
			dctx.Logger.Warn("Failed to scan empty section", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// ============================================================================
// Completion-time verifier (item_type "empty_section")
// ============================================================================

// emptyHeadingRe mirrors the SQL detection pattern '<(h[1-6])[^>]*>\s*</\1>'
// in findEmptySections. Go's RE2 has no backreferences, so the closing tag
// matches any h1-h6 — marginally broader than the SQL, and anything the
// broader form catches is still an empty heading shell.
var emptyHeadingRe = regexp.MustCompile(`(?i)<h[1-6][^>]*>\s*</h[1-6]>`)

// VerifyEmptySectionResolved re-runs the empty-section predicate for the
// single component named in the item spec. Registered for item_type
// "empty_section" so CompleteWorkItemAction can refuse to stamp 'complete'
// while the section still renders empty.
func VerifyEmptySectionResolved(ctx context.Context, db *sql.DB, spec map[string]interface{}, logger *zap.Logger) (VerifyResult, error) {
	componentID, _ := spec["component_id"].(string)
	if componentID == "" {
		return VerifyResult{}, fmt.Errorf("empty_section spec has no component_id")
	}
	if _, err := uuid.Parse(componentID); err != nil {
		return VerifyResult{}, fmt.Errorf("empty_section spec component_id %q: %w", componentID, err)
	}

	var html sql.NullString
	err := db.QueryRowContext(ctx,
		`SELECT rendered_html FROM page_components WHERE id = $1`, componentID,
	).Scan(&html)
	if err == sql.ErrNoRows {
		// AMBIGUOUS — never report success here (bugs_open/032).
		//
		// A missing page_components row is equally the signature of a genuine fix
		// (or a deliberate removal) and of this platform's most repeated failure:
		// a rebuild silently deleting the component. The two are indistinguishable
		// from this query, so reporting Resolved:true recorded content-loss
		// incidents as verified fixes — the exact blind spot that caused the
		// incidents this gate exists to catch (bugs_open/012, /021).
		//
		// An error rather than Resolved:false because the gate's documented policy
		// is to fail OPEN on verifier error (complete_work_item_verification.go:14):
		// the item still completes and nothing wedges, but result._verification
		// records that verification could not be made. A false success becomes a
		// visible unknown. Resolved:false would burn an attempt and, at
		// max_attempts, strand a legitimately-removed component's item in 'failed'.
		//
		// Deliberately left open: if the page still EXPECTS this component (a
		// plan_sections entry, a slot reference), absence is not ambiguous at all
		// — it is deletion, and Resolved:false would be the honest answer. That is
		// a better verdict and a bigger change; it belongs to the
		// empty_sections_loop_integrity thread, and this floor does not preclude it.
		return VerifyResult{}, fmt.Errorf(
			"cannot verify: component %s no longer exists (genuinely fixed or silently deleted — indistinguishable here)",
			componentID)
	}
	if err != nil {
		return VerifyResult{}, err
	}

	return emptySectionVerdict(html.String), nil
}

// emptySectionVerdict classifies rendered HTML with the same patterns as
// findEmptySections' SQL WHERE clause (null/empty, minimal, empty-heading)
// plus the data-runtime-fill exemption. Pure — unit tested.
func emptySectionVerdict(html string) VerifyResult {
	if strings.TrimSpace(html) == "" {
		return VerifyResult{Resolved: false, Detail: "rendered_html is still null/empty"}
	}
	if strings.Contains(html, "data-runtime-fill") {
		return VerifyResult{Resolved: true, Detail: "runtime-fill shell — empty by design, filled client-side"}
	}
	if len(html) < 50 {
		return VerifyResult{Resolved: false, Detail: "rendered_html still minimal (<50 chars)"}
	}
	if emptyHeadingRe.MatchString(html) {
		return VerifyResult{Resolved: false, Detail: "empty heading shell still present in rendered_html"}
	}
	return VerifyResult{Resolved: true, Detail: "section has rendered content"}
}
