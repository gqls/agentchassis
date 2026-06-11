// FILE: platform/orchestration/actions/discovery_checks/check_unresolved_sections.go
//
// Finds pages whose sections array references section_types that now have
// matching components in the library, but the page hasn't been rebuilt to
// include them. This catches cases where:
//   - component-creator built a component after plan_sections ran
//   - a component was created manually
//   - a component was created for a different site but applies here too
//
// The check marks affected pages as 'needs_rebuild' directly — no work items
// needed since the normal build pipeline handles the rest.
//
// Registration: automatic via init() → Register(&UnresolvedSectionsCheck{})

package discovery_checks

import (
	"go.uber.org/zap"
)

func init() { Register(&UnresolvedSectionsCheck{}) }

type UnresolvedSectionsCheck struct{}

func (c *UnresolvedSectionsCheck) Name() string { return "unresolved_sections" }

func (c *UnresolvedSectionsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Find pages with sections that reference components which now exist
	// but aren't yet linked via page_components.
	//
	// Logic: for each active deployed page, extract section names from the
	// sections jsonb array. Check if any section name matches a section_type
	// (or function) in content_components that is NOT already in page_components
	// for that page.
	res, err := dctx.DB.ExecContext(dctx.Ctx, `
		UPDATE pages SET build_status = 'needs_rebuild', updated_at = NOW()
		WHERE id IN (
			SELECT DISTINCT p.id
			FROM pages p,
			     jsonb_array_elements_text(p.sections) AS sec(section_name)
			WHERE p.site_id = $1
			  AND p.status = 'active'
			  AND p.build_status = 'deployed'
			  AND EXISTS (
			      SELECT 1 FROM content_components cc
			      WHERE cc.is_active = true
			        AND cc.forked_from IS NULL
			        AND (cc.section_type = sec.section_name OR cc.function = sec.section_name)
			  )
			  AND NOT EXISTS (
			      SELECT 1 FROM page_components pc
			      JOIN content_components cc2 ON pc.component_id = cc2.id
			      WHERE pc.page_id = p.id
			        AND (cc2.section_type = sec.section_name OR cc2.function = sec.section_name)
			  )
		)
	`, dctx.SiteID)

	if err != nil {
		dctx.Logger.Warn("check_unresolved_sections: query failed", zap.Error(err))
		return &CheckResult{}, nil // non-fatal — skip this check
	}

	rows, _ := res.RowsAffected()
	if rows > 0 {
		dctx.Logger.Info("check_unresolved_sections: marked pages for rebuild",
			zap.Int64("pages_marked", rows))
	}

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":        "unresolved_sections",
			"pages_marked": rows,
		}},
	}, nil
}
