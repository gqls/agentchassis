// FILE: platform/orchestration/actions/load_page_sections_from_spec_action.go
//
// LoadPageSectionsFromSpecAction reads the section list for a page.
//
// Source priority (2026-07-06 — the site_plans table family is authoritative;
// see PLAN_dynamic_sections_and_loaders "Plan storage" decision):
//   1. site_plans tables — site_plan_sections for the CURRENT plan, by
//      page_name, ordered. The normalised, versioned, constrained store the
//      newer planner generation writes. When it serves, sections are synced
//      to pages.sections (materialised cache) exactly as the spec path does.
//   2. site_specs.site_plan aspect — the older planner generation's store;
//      five sites currently carry a current aspect and are served here
//      unchanged (their table lookup simply misses).
//   3. pages.sections (legacy fallback / materialised cache).
//   4. Same-role sibling layout synthesis (last resort, WARN-logged) — note
//      this path already read the site_plans tables before this change.
//   5. (2026-08-15, bugs_open/285, register LOCK-008) MERGE, not a tier: once
//      any tier has served, the page's LOCKED live rows — page_components rows
//      automation may not rewrite, classified by the 058 write guard's OWN
//      predicate (datahelpers.AgentWritableSQLFor) — are inserted into the
//      list at their live position when the list does not already carry them
//      (pairing mirrors save_page_sections' matchLockedRow; see
//      datahelpers.MergeLockedPageSlots). Before this, a section a human had
//      pinned to the page and the plan did not know about was proposed for
//      removal on EVERY rebuild; the guard kept the row and filed
//      `lock_blocked_change … remove` each pass, while pages.sections went on
//      saying the section did not exist. Measured 2026-08-15: 13 pages
//      fleet-wide, 5 fresh remove-blocked items that day. Deliberately NOT
//      applied when no tier served: a locked-only list is neither the plan nor
//      the page, and a rebuild on it would delete the page's unlocked
//      siblings — that page keeps today's "no sections" outcome.
//
// This replaces the implicit page_record.sections path in the page-build
// workflow.
//
// Also syncs: after the merge, ONE guarded UPDATE writes the final list into
// pages.sections when it differs (jsonb-compared — the previous per-tier
// `sections::text IS DISTINCT FROM $1` was always true, because jsonb text
// prints `", "` where Go marshals `","`, so every build bumped
// pages.updated_at). The materialised cache therefore carries the merged list,
// which is what check_section_source_drift now compares through the same
// merge, and what save_sections_prune_floor sizes its plan cohort from.
//
// Fallback 4 (was 2b): if no source has sections for this page, borrow the
// layout skeleton from a same-role sibling in the current plan. "sections" is
// a layout (an ordered list of component types, e.g.
// ["hero","generic-text-block"]), NOT content — same-role pages share a
// layout by design, and content is still written per page by the content
// writer from this page's own source. This rescues a page that reached the
// build with an empty sections array (e.g. a page unioned back into the plan
// during adoption convergence with no sections to carry), which would
// otherwise see zero sections and short-circuit the build to a (silently
// successful) "no sections defined" completion. "source" values:
// "site_plan_tables" (new, 2026-07-06), "site_specs", "pages_table",
// "same_role_sibling", "none". Result also carries "locked_sections_merged"
// (the slot names the merge inserted, [] when none) and "locked_merge_count".
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
	CheckConfig: true,
	Required:    []string{"site_id", "page_name"},
	Optional:    []string{"page_sections_fallback"},
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
	// 1. site_plans tables (authoritative): site_plan_sections for the CURRENT
	//    plan, by page_name, ordered. Written by the newer planner generation;
	//    the same store fallback 4 (sibling synthesis) already reads.
	// -----------------------------------------------------------------------
	var specSections []string
	var specSectionFacts []interface{} // aligned with specSections; nil entry = unscoped
	var specSource string

	planRows, tblErr := params.DB.QueryContext(ctx, `
		SELECT sps.component_name, sps.assigned_fact_ids
		FROM site_plan_sections sps
		JOIN site_plans sp ON sp.id = sps.plan_id
		WHERE sp.site_id = $1 AND sp.is_current = true AND sps.page_name = $2
		ORDER BY sps.ordering
	`, siteID, pageName)
	if tblErr != nil {
		logger.Warn("LoadPageSectionsFromSpec: site_plan_sections lookup failed",
			zap.Error(tblErr))
	} else {
		for planRows.Next() {
			var comp string
			var factsRaw []byte
			if scanErr := planRows.Scan(&comp, &factsRaw); scanErr != nil {
				logger.Warn("LoadPageSectionsFromSpec: site_plan_sections scan failed",
					zap.Error(scanErr))
				continue
			}
			if comp != "" {
				specSections = append(specSections, comp)
				// assigned_fact_ids: NULL = unscoped (nil entry); a JSON
				// array (possibly empty) = plan-time fact scoping for this
				// section (bugs_open/151 candidate 1). Appended in the same
				// branch as the name so the two lists cannot misalign.
				var factsEntry interface{}
				if len(factsRaw) > 0 {
					var ids []interface{}
					if jerr := json.Unmarshal(factsRaw, &ids); jerr == nil {
						factsEntry = ids
					} else {
						logger.Warn("LoadPageSectionsFromSpec: assigned_fact_ids unparseable, treating section as unscoped",
							zap.String("component", comp), zap.Error(jerr))
					}
				}
				specSectionFacts = append(specSectionFacts, factsEntry)
			}
		}
		planRows.Close()
	}

	if len(specSections) > 0 {
		specSource = "site_plan_tables"
		logger.Info("LoadPageSectionsFromSpec: using site_plans tables (authoritative)",
			zap.String("page", pageName),
			zap.Int("sections", len(specSections)),
			zap.Strings("section_names", specSections))
		// pages.sections sync happens ONCE, after the locked-row merge below.
	}

	// -----------------------------------------------------------------------
	// 2. site_specs.site_plan aspect (older planner generation; five sites
	//    currently live here — their table lookup above simply misses)
	// -----------------------------------------------------------------------
	var planDataJSON []byte
	if len(specSections) == 0 {
		err = params.DB.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true
	`, siteID).Scan(&planDataJSON)
	}

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

	// NOTE (2026-07-06): condition gained `specSource == ""` so this block only
	// fires when the ASPECT served — Step 1 (site_plan_tables) now also fills
	// specSections and must not be relabelled/re-synced here.
	if specSource == "" && len(specSections) > 0 {
		specSource = "site_specs"
		logger.Info("LoadPageSectionsFromSpec: using site_specs.site_plan",
			zap.String("page", pageName),
			zap.Int("sections", len(specSections)),
			zap.Strings("section_names", specSections))
		// pages.sections sync happens ONCE, after the locked-row merge below.
	}

	// -----------------------------------------------------------------------
	// 3. Fallback to pages.sections (legacy source / materialised cache)
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
	// 4 (was 2b). Last-resort fallback: borrow the layout from a same-role sibling.
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
				// Persisted to pages.sections by the single sync after the
				// locked-row merge below, so the build can read it and the page
				// stops being a zero-section dead-end.
			}
		}
	}

	// -----------------------------------------------------------------------
	// 5. Return sections for plan_sections to consume
	// -----------------------------------------------------------------------
	if len(specSections) == 0 {
		logger.Warn("LoadPageSectionsFromSpec: no sections found for page",
			zap.String("page", pageName))
		return map[string]interface{}{
			"sections":               []string{},
			"source":                 "none",
			"count":                  0,
			"locked_sections_merged": []string{},
			"locked_merge_count":     0,
		}, nil
	}

	// -----------------------------------------------------------------------
	// 5a. Merge the page's LOCKED live rows into the list (bugs_open/285).
	// A tier served, so the list is a real statement of the page; rows a
	// human pinned to the page that no tier knows about join it here at
	// their live position. Same predicate as the 058 write guard; pairing
	// mirrors its matchLockedRow so "already in the list" means the same
	// thing in both places. Best-effort on query failure, exactly as
	// loadActiveLockedRows is: the guard still protects the row itself, only
	// the list (and this build's cache write) stays lock-blind, and the
	// failure is logged where the 058 preload's is.
	// -----------------------------------------------------------------------
	lockedMerged := []string{}
	if lockedRows, lockErr := datahelpers.LoadLockedPageSlots(ctx, params.DB, siteID, pageName); lockErr != nil {
		logger.Warn("LoadPageSectionsFromSpec: locked-row preload failed — list assembled lock-blind this run (bugs_open/285)",
			zap.String("page", pageName), zap.Error(lockErr))
	} else if len(lockedRows) > 0 {
		merged, inserted, insertedAt := datahelpers.MergeLockedPageSlots(specSections, lockedRows)
		if len(inserted) > 0 {
			// Keep the tier-1 facts slice index-aligned: a merged section has
			// no plan-time fact assignment (nil = unscoped), and the
			// len(specSectionFacts) == len(specSections) guard below would
			// otherwise drop the WHOLE payload silently.
			if specSource == "site_plan_tables" && len(specSectionFacts) == len(specSections) {
				for _, at := range insertedAt {
					specSectionFacts = append(specSectionFacts, nil)
					copy(specSectionFacts[at+1:], specSectionFacts[at:])
					specSectionFacts[at] = nil
				}
			}
			for _, lr := range inserted {
				lockedMerged = append(lockedMerged, lr.MergedName())
			}
			logger.Info("LoadPageSectionsFromSpec: merged human-locked live sections the plan does not name (bugs_open/285)",
				zap.String("page", pageName),
				zap.String("source", specSource),
				zap.Strings("merged_slots", lockedMerged),
				zap.Strings("section_names", merged))
			specSections = merged
		}
	}

	// -----------------------------------------------------------------------
	// 5b. Sync the FINAL list to pages.sections (materialised cache), once.
	// jsonb-compared so an unchanged list is a genuine no-op (the old
	// `sections::text IS DISTINCT FROM $1` guard was always true — jsonb text
	// prints `", "`, json.Marshal `","` — so every build rewrote the row and
	// bumped updated_at). Tier 3 (pages_table) reaches this too: a no-op
	// unless the merge added something, in which case the cache it was read
	// from was lying and is corrected.
	// -----------------------------------------------------------------------
	if sectionsJSON, mErr := json.Marshal(specSections); mErr == nil {
		if _, syncErr := params.DB.ExecContext(ctx, `
			UPDATE pages SET sections = $1::jsonb, updated_at = NOW()
			WHERE site_id = $2 AND name = $3
			  AND sections IS DISTINCT FROM $1::jsonb
		`, string(sectionsJSON), siteID, pageName); syncErr != nil {
			logger.Warn("LoadPageSectionsFromSpec: sync to pages.sections failed",
				zap.String("source", specSource), zap.Error(syncErr))
		} else {
			logger.Info("LoadPageSectionsFromSpec: synced sections to pages table",
				zap.String("source", specSource))
		}
	}

	// Return as interface slice (consistent with how plan_sections reads it)
	sectionsIface := make([]interface{}, len(specSections))
	for i, s := range specSections {
		sectionsIface[i] = s
	}

	result := map[string]interface{}{
		"sections":               sectionsIface,
		"source":                 specSource,
		"count":                  len(specSections),
		"locked_sections_merged": lockedMerged,
		"locked_merge_count":     len(lockedMerged),
	}
	// section_facts is emitted ONLY when the authoritative tier served the
	// sections: it is read from site_plan_sections rows, so index-alignment
	// with any other tier's list would be a guess, not a fact. Fallback-tier
	// sections are simply unscoped.
	if specSource == "site_plan_tables" && len(specSectionFacts) == len(specSections) {
		result["section_facts"] = specSectionFacts
	}
	return result, nil
}
