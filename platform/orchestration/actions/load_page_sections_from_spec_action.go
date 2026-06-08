// FILE: platform/orchestration/actions/load_page_sections_from_spec_action.go
//
// LoadPageSectionsFromSpecAction reads the section list for a page from
// site_specs.site_plan (the authoritative spec), falling back to
// pages.sections (the legacy source).
//
// This replaces the implicit page_record.sections path in the page-build
// workflow. When site_specs has a site_plan for this site, and the plan
// includes sections for this page, those sections are used. Otherwise
// the pages.sections column is used (backward compatible).
//
// Also syncs: if site_specs sections differ from pages.sections, updates
// the pages table so the two stay in sync going forward.
//
// Fallback 2b (added): if neither site_specs.site_plan nor pages.sections
// has sections for this page, borrow the layout skeleton from a same-role
// sibling in the current plan. "sections" is a layout (an ordered list of
// component types, e.g. ["hero","generic-text-block"]), NOT content — same-
// role pages share a layout by design, and content is still written per page
// by the content writer from this page's own source. This rescues a page
// that reached the build with an empty sections array (e.g. a page unioned
// back into the plan during adoption convergence with no sections to carry),
// which would otherwise see zero sections and short-circuit the build to a
// (silently successful) "no sections defined" completion. Adds a new
// "source" value: "same_role_sibling".
//
// Registration:
//   "load_page_sections_from_spec": {
//       Handler:     LoadPageSectionsFromSpecAction,
//       Category:    "site",
//       Description: "Load page sections from site_specs.site_plan with fallback to pages table",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "load_spec_sections": {
//       "action": "load_page_sections_from_spec",
//       "config": {
//           "site_id": "site_record.site_id",
//           "page_name": "page_record.name",
//           "page_sections_fallback": "page_record.sections"
//       },
//       "output_field": "spec_sections",
//       "next_step": "plan_sections"
//   }

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadPageSectionsFromSpecInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "page_name"},
	Optional: []string{"page_sections_fallback"},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_page_sections_from_spec", LoadPageSectionsFromSpecInputSpec)
}

func LoadPageSectionsFromSpecAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_page_sections_from_spec"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadPageSectionsFromSpecInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	pageName := inputs.Get("page_name")
	if pageName == "" {
		return nil, fmt.Errorf("page_name is required")
	}

	// -----------------------------------------------------------------------
	// 1. Try site_specs.site_plan (authoritative source)
	// -----------------------------------------------------------------------
	var specSections []string
	var specSource string

	var planDataJSON []byte
	err = params.DB.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true
	`, siteID).Scan(&planDataJSON)

	if err == nil && planDataJSON != nil {
		var planData map[string]interface{}
		if json.Unmarshal(planDataJSON, &planData) == nil {
			if pages, ok := planData["pages"].([]interface{}); ok {
				for _, pageRaw := range pages {
					page, ok := pageRaw.(map[string]interface{})
					if !ok {
						continue
					}
					if name, _ := page["name"].(string); name == pageName {
						if sections, ok := page["sections"].([]interface{}); ok {
							for _, s := range sections {
								if sName, ok := s.(string); ok {
									specSections = append(specSections, sName)
								}
							}
						}
						break
					}
				}
			}
		}
	}

	if len(specSections) > 0 {
		specSource = "site_specs"
		logger.Info("LoadPageSectionsFromSpec: using site_specs.site_plan",
			zap.String("page", pageName),
			zap.Int("sections", len(specSections)),
			zap.Strings("section_names", specSections))

		// Sync to pages.sections so both sources agree
		sectionsJSON, _ := json.Marshal(specSections)
		_, syncErr := params.DB.ExecContext(ctx, `
			UPDATE pages SET sections = $1::jsonb, updated_at = NOW()
			WHERE site_id = $2 AND name = $3
			  AND sections::text IS DISTINCT FROM $1
		`, string(sectionsJSON), siteID, pageName)
		if syncErr != nil {
			logger.Warn("LoadPageSectionsFromSpec: sync to pages.sections failed",
				zap.Error(syncErr))
		} else {
			logger.Info("LoadPageSectionsFromSpec: synced sections to pages table")
		}
	}

	// -----------------------------------------------------------------------
	// 2. Fallback to pages.sections (legacy source)
	// -----------------------------------------------------------------------
	if len(specSections) == 0 {
		// Try from collected_data first (page_record.sections)
		fallbackRaw := inputs.GetRaw("page_sections_fallback")
		if fallbackRaw != nil {
			switch v := fallbackRaw.(type) {
			case []interface{}:
				for _, s := range v {
					if name, ok := s.(string); ok {
						specSections = append(specSections, name)
					}
				}
			case []string:
				specSections = v
			}
		}

		// If still empty, query pages table directly
		if len(specSections) == 0 {
			var sectionsJSON []byte
			err = params.DB.QueryRowContext(ctx, `
				SELECT sections FROM pages
				WHERE site_id = $1 AND name = $2
			`, siteID, pageName).Scan(&sectionsJSON)
			if err == nil && sectionsJSON != nil {
				json.Unmarshal(sectionsJSON, &specSections)
			}
		}

		if len(specSections) > 0 {
			specSource = "pages_table"
			logger.Info("LoadPageSectionsFromSpec: using pages.sections fallback",
				zap.String("page", pageName),
				zap.Int("sections", len(specSections)))
		}
	}

	// -----------------------------------------------------------------------
	// 2b. Last-resort fallback: borrow the layout from a same-role sibling.
	//
	// "sections" is a layout skeleton (an ordered list of component types,
	// e.g. ["hero","generic-text-block"]), NOT content. Pages of the same
	// role share a layout by design — every guide on a site uses the same
	// skeleton. Content is written later, per page, by the content writer
	// from THIS page's own source, so nothing is borrowed but the skeleton.
	//
	// Rescues a page that reached the build with no sections in either
	// site_specs.site_plan or pages.sections (e.g. a page unioned back into
	// the plan during adoption convergence with an empty sections array).
	// Last resort only, and logged at WARN so a synthesised layout is always
	// visible in the logs. Persists to pages.sections (the field the build
	// reads) using the same guarded UPDATE as the site_specs path above.
	// -----------------------------------------------------------------------
	if len(specSections) == 0 {
		rows, qErr := params.DB.QueryContext(ctx, `
			WITH cur AS (
				SELECT id AS plan_id FROM site_plans
				WHERE site_id = $1 AND is_current = true
			),
			target AS (
				SELECT spp.role
				FROM site_plan_pages spp
				JOIN cur ON spp.plan_id = cur.plan_id
				WHERE spp.name = $2
			)
			SELECT spp.name,
			       to_jsonb(array_agg(sps.component_name ORDER BY sps.ordering))
			FROM site_plan_pages spp
			JOIN cur ON spp.plan_id = cur.plan_id
			JOIN target ON spp.role IS NOT DISTINCT FROM target.role
			JOIN site_plan_sections sps
			  ON sps.plan_id = spp.plan_id AND sps.page_name = spp.name
			WHERE spp.name <> $2
			GROUP BY spp.name
		`, siteID, pageName)
		if qErr != nil {
			logger.Warn("LoadPageSectionsFromSpec: same-role sibling lookup failed",
				zap.Error(qErr))
		} else {
			// Tally distinct layouts so one outlier sibling cannot skew the
			// choice; pick the layout shared by the most siblings (modal),
			// with a deterministic tie-break (more siblings, then longer
			// layout, then lexicographic by key).
			layoutCounts := map[string]int{}
			layoutSections := map[string][]string{}
			layoutExample := map[string]string{}
			for rows.Next() {
				var sibName string
				var compsJSON []byte
				if scanErr := rows.Scan(&sibName, &compsJSON); scanErr != nil {
					logger.Warn("LoadPageSectionsFromSpec: sibling scan failed",
						zap.Error(scanErr))
					continue
				}
				var comps []string
				if json.Unmarshal(compsJSON, &comps) != nil || len(comps) == 0 {
					continue
				}
				key := fmt.Sprintf("%q", comps)
				layoutCounts[key]++
				if _, seen := layoutSections[key]; !seen {
					layoutSections[key] = comps
					layoutExample[key] = sibName
				}
			}
			rows.Close()

			bestKey := ""
			for key := range layoutCounts {
				if bestKey == "" {
					bestKey = key
					continue
				}
				better := layoutCounts[key] > layoutCounts[bestKey] ||
					(layoutCounts[key] == layoutCounts[bestKey] &&
						len(layoutSections[key]) > len(layoutSections[bestKey])) ||
					(layoutCounts[key] == layoutCounts[bestKey] &&
						len(layoutSections[key]) == len(layoutSections[bestKey]) &&
						key < bestKey)
				if better {
					bestKey = key
				}
			}

			if bestKey != "" {
				specSections = layoutSections[bestKey]
				specSource = "same_role_sibling"
				logger.Warn("LoadPageSectionsFromSpec: SYNTHESISED layout from same-role sibling — page had no sections in site_specs or pages.sections",
					zap.String("page", pageName),
					zap.String("sibling", layoutExample[bestKey]),
					zap.Strings("section_names", specSections))

				// Persist to pages.sections so the build can read it and the
				// page stops being a zero-section dead-end.
				sectionsJSON, _ := json.Marshal(specSections)
				if _, syncErr := params.DB.ExecContext(ctx, `
					UPDATE pages SET sections = $1::jsonb, updated_at = NOW()
					WHERE site_id = $2 AND name = $3
					  AND sections::text IS DISTINCT FROM $1
				`, string(sectionsJSON), siteID, pageName); syncErr != nil {
					logger.Warn("LoadPageSectionsFromSpec: persisting synthesised sections failed",
						zap.Error(syncErr))
				}
			}
		}
	}

	// -----------------------------------------------------------------------
	// 3. Return sections for plan_sections to consume
	// -----------------------------------------------------------------------
	if len(specSections) == 0 {
		logger.Warn("LoadPageSectionsFromSpec: no sections found for page",
			zap.String("page", pageName))
		return map[string]interface{}{
			"sections": []string{},
			"source":   "none",
			"count":    0,
		}, nil
	}

	// Return as interface slice (consistent with how plan_sections reads it)
	sectionsIface := make([]interface{}, len(specSections))
	for i, s := range specSections {
		sectionsIface[i] = s
	}

	return map[string]interface{}{
		"sections": sectionsIface,
		"source":   specSource,
		"count":    len(specSections),
	}, nil
}
