// FILE: platform/orchestration/actions/discovery_checks/check_missing_structure.go
//
// Flags deployed pages that are missing rendered header, footer, or head.
// This happens when pages were built before site_components were rendered,
// or when the assemble step didn't inject them.
//
// Creates one needs_rerender work item per site (not per page) with
// refresh_site_components=true, which triggers a full reassembly.

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&MissingStructureCheck{}) }

type MissingStructureCheck struct{}

func (c *MissingStructureCheck) Name() string { return "missing_structure" }

func (c *MissingStructureCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	pages, err := findPagesWithMissingStructure(dctx)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return &CheckResult{}, nil
	}

	// Build findings list
	pageNames := make([]string, len(pages))
	for i, p := range pages {
		pageNames[i] = p.Name
	}

	dctx.Logger.Warn("Found pages with missing structure",
		zap.Int("count", len(pages)),
		zap.Strings("pages", pageNames))

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":  "missing_structure",
			"count":  len(pages),
			"pages":  pages,
			"detail": "Deployed pages missing rendered header/footer — need full reassembly",
		}},
	}

	// Create ONE rerender work item for the whole site (not per page).
	// The rerender-pages agent processes all pages in one pass.
	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":                   "missing_structure",
		"refresh_site_components": true,
		"affected_page_count":     len(pages),
		"affected_pages":          pageNames,
		"reason":                  "Pages deployed without header/footer — likely built before site_components were rendered",
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "needs_rerender",
		Severity:     "high",
		Summary:      fmt.Sprintf("%d pages missing header/footer — need reassembly", len(pages)),
		SpecJSON:     string(specJSON),
		Priority:     30, // high priority — structural issue visible to users
		HandlerAgent: "rerender-pages",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      "missing_structure:rerender",
		BatchID:      dctx.BatchID,
	})

	return result, nil
}

type missingStructureFinding struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	PageID        string `json:"page_id"`
	MissingHeader bool   `json:"missing_header"`
	MissingFooter bool   `json:"missing_footer"`
	MissingHead   bool   `json:"missing_head"`
}

func findPagesWithMissingStructure(dctx DiscoveryCheckContext) ([]missingStructureFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name, p.url,
		       p.rendered_header IS NULL as missing_header,
		       p.rendered_footer IS NULL as missing_footer,
		       p.rendered_head IS NULL as missing_head
		FROM pages p
		WHERE p.site_id = $1
		  AND p.status IN ('active', 'deployed')
		  AND (p.rendered_header IS NULL OR p.rendered_footer IS NULL)
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("missing_structure query failed: %w", err)
	}
	defer rows.Close()

	var findings []missingStructureFinding
	for rows.Next() {
		var f missingStructureFinding
		if err := rows.Scan(&f.PageID, &f.Name, &f.URL,
			&f.MissingHeader, &f.MissingFooter, &f.MissingHead); err != nil {
			dctx.Logger.Warn("missing_structure: scan error", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
