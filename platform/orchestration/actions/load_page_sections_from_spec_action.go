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
