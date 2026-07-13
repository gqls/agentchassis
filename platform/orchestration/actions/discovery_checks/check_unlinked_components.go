// FILE: platform/orchestration/actions/discovery_checks/check_unlinked_components.go
//
// Discovery check: unlinked_page_components
//
// Detects page_components that have NULL component_id but contain a
// data-component="..." attribute in their rendered_html that matches
// a content_components.function. Links them by setting component_id.
//
// This is a self-healing check: it does the fix inline rather than
// creating a work item, because the fix is a simple deterministic
// UPDATE (no LLM, no handler agent needed).
//
// Why this happens:
// - Pages built before enrichSectionsWithComponentIDs was deployed
// - Pages saved via the structured metadata path when the content writer
//   didn't include component_id in sections_metadata
// - Adopted sites imported with raw HTML
//
// Impact of unlinked components:
// - Discovery checks that JOIN on component_id miss sections that exist
//   (e.g. missing_news_section falsely fires for pages that have the HTML)
// - Section editors can't address components by function name
// - CSS snippet loading by function may miss sections
//
// Registration: add "unlinked_page_components" to the completeness-discovery-agent
// checks array. Runs as Layer 1 (algorithmic, no LLM).

package discovery_checks

import (
	"fmt"

	"go.uber.org/zap"
)

func init() {
	Register(&UnlinkedPageComponentsCheck{})
}

type UnlinkedPageComponentsCheck struct{}

func (c *UnlinkedPageComponentsCheck) Name() string { return "unlinked_page_components" }

func (c *UnlinkedPageComponentsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Find and fix unlinked page_components in one query.
	// Extracts data-component="..." from rendered_html, matches to
	// content_components.function, sets component_id.
	//
	// Only processes components for the current site (scoped by site_id).
	// Only matches base components (forked_from IS NULL) to avoid linking
	// to site-specific forks that might not be the right match.
	result, err := dctx.DB.ExecContext(dctx.Ctx, `
		WITH unlinked AS (
			SELECT pc.id as pc_id,
			       (regexp_match(pc.rendered_html, 'data-component="([^"]+)"'))[1] as extracted_function
			FROM page_components pc
			JOIN pages p ON p.id = pc.page_id
			WHERE p.site_id = $1
			  AND pc.component_id IS NULL
			  AND pc.rendered_html IS NOT NULL
			  AND pc.rendered_html LIKE '%data-component="%'
		),
		matched AS (
			SELECT DISTINCT ON (u.pc_id)
			       u.pc_id,
			       u.extracted_function,
			       cc.id as cc_id
			FROM unlinked u
			JOIN content_components cc
			    ON cc.function = u.extracted_function
			    AND cc.is_active = true
			    AND cc.forked_from IS NULL
			WHERE u.extracted_function IS NOT NULL
			ORDER BY u.pc_id, cc.created_at ASC
		)
		UPDATE page_components pc
		SET component_id = m.cc_id,
		    updated_at = NOW()
		FROM matched m
		WHERE pc.id = m.pc_id
	`, dctx.SiteID)

	if err != nil {
		return nil, fmt.Errorf("unlinked_page_components: update failed: %w", err)
	}

	linked, _ := result.RowsAffected()

	if linked == 0 {
		// No unlinked components found — site is clean
		return &CheckResult{}, nil
	}

	dctx.Logger.Info("UnlinkedPageComponentsCheck: linked components",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int64("linked", linked))

	// Return findings for logging/audit but no work items — the fix is already done
	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   "unlinked_page_components",
			"message": fmt.Sprintf("Linked %d page_components to their content_components via data-component attribute", linked),
			"linked":  linked,
		}},
		// No WorkItems — self-healed
	}, nil
}
