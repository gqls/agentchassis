// FILE: platform/orchestration/actions/reconcile_section_data_action.go
//
// ReconcileSectionDataAction re-attempts open needs_section_data items whose
// missing data is query-resolvable, and re-renders the affected page once that
// data exists. It closes the loop the original needs_section_data design left
// open, WITHOUT a new LLM agent.
//
// CONTEXT
//   plan_sections defers a section to needs_section_data (status
//   needs_human_review) when a required field can't resolve. For list/grid
//   sections the missing field is query-sourced (query.pages_where_type:...,
//   query.pages_under_section:...) — derived data, not human data. It defers
//   when the listed pages don't exist yet (or, until recently, when the query
//   wasn't implemented). Those items then sit forever because nothing re-plans
//   the page after the data lands; loadOpenSectionDataRequests +
//   closeResolvedDataRequest only close an item on a LATER plan_sections run.
//
// WHAT IT DOES
//   For each open needs_section_data item whose missing fields are ALL
//   query-sourced, it re-runs the query now (queryresolve.Resolve). If the data
//   now exists, it emits a needs_page so page-build-handler re-renders the page
//   through plan_sections, which re-resolves the field and auto-closes the item
//   via closeResolvedDataRequest. Items with any non-query (human) missing
//   field — team, pricing, case studies, contact — are left at
//   needs_human_review untouched (distinguished by spec.missing[].source).
//
//   It does NOT close items itself: re-rendering through plan_sections is the
//   single source of truth for resolution + close, so the page actually gets
//   the data, not just a cleared flag.
//
// WIRING (recommended): run periodically in the improvement loop, or as a step
//   in the finalize/rerender path AFTER pages are built (running it at
//   plan time is too early — the listed pages don't exist yet). It is
//   invocation-agnostic: given site_id it scans and re-triggers. Not wired
//   here — pick the host agent and add a step:
//     "reconcile_section_data": {
//        "action": "reconcile_section_data",
//        "config": { "site_id": "<path to site_id>" },
//        "next_step": "..." }
//
// REGISTRATION (registry.go):
//   "reconcile_section_data": {
//       Handler:     ReconcileSectionDataAction,
//       Category:    "site",
//       Description: "Re-trigger pages whose deferred section data is now query-resolvable",
//       IsLocal:     true,
//   }

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ReconcileSectionDataInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"page_name"},
	Defaults: map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("reconcile_section_data", ReconcileSectionDataInputSpec)
}

// reconcileMissingField mirrors the fields of plan_sections' missingField that
// the reconciler reads from needs_section_data spec.missing[].
type reconcileMissingField struct {
	Field  string `json:"field"`
	Source string `json:"source"`
}

type reconcileSpec struct {
	PageName    string                  `json:"page_name"`
	SectionName string                  `json:"section_name"`
	Missing     []reconcileMissingField `json:"missing"`
}

func ReconcileSectionDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "reconcile_section_data"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		ReconcileSectionDataInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}
	pageFilter := inputs.Get("page_name") // optional scope

	// Load open needs_section_data items for this site.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT id, spec
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'needs_section_data'
		  AND status IN ('needs_human_review', 'triaged')
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("load open section-data items: %w", err)
	}
	defer rows.Close()

	type openItem struct {
		id   uuid.UUID
		spec reconcileSpec
	}
	var open []openItem
	for rows.Next() {
		var id uuid.UUID
		var specJSON []byte
		if err := rows.Scan(&id, &specJSON); err != nil {
			logger.Warn("reconcile_section_data: scan failed", zap.Error(err))
			continue
		}
		var s reconcileSpec
		if err := json.Unmarshal(specJSON, &s); err != nil {
			logger.Warn("reconcile_section_data: spec unmarshal failed",
				zap.String("item_id", id.String()), zap.Error(err))
			continue
		}
		if pageFilter != "" && s.PageName != pageFilter {
			continue
		}
		open = append(open, openItem{id: id, spec: s})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate section-data items: %w", err)
	}

	scanned := len(open)
	resolvablePages := make(map[string]bool) // page_name → needs re-render
	skippedHuman := 0

	for _, item := range open {
		if len(item.spec.Missing) == 0 {
			continue
		}

		allQuery := true
		allResolved := true
		for _, m := range item.spec.Missing {
			if !strings.HasPrefix(m.Source, "query.") {
				// Non-query (human/spec) source — needs human input. Leave it.
				allQuery = false
				break
			}
			queryName := strings.TrimPrefix(m.Source, "query.")
			val, qerr := queryresolve.Resolve(ctx, params.DB, queryresolve.QueryRequest{
				Name:   queryName,
				SiteID: siteID,
			}, logger)
			if qerr != nil || !hasItems(val) {
				allResolved = false
			}
		}

		if !allQuery {
			skippedHuman++
			continue
		}
		if allResolved && item.spec.PageName != "" {
			resolvablePages[item.spec.PageName] = true
		}
	}

	// One re-render per resolvable page. Shared dedup key with
	// flag_page_image_rebuild (page_rerender:<page>) so concurrent triggers
	// collapse into a single re-render that re-resolves both image and section
	// data. plan_sections closes the needs_section_data item on re-render.
	var retriggered []string
	if len(resolvablePages) > 0 {
		tx, err := params.DB.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()

		batchID := uuid.New()
		for page := range resolvablePages {
			spec := fmt.Sprintf(`{"reason":"section_data_resolved","page_name":%q}`, page)
			if _, err := insertWorkItem(ctx, tx, workItem{
				siteID:       siteID,
				source:       "section-data-reconciler",
				pipeline:     "build",
				itemType:     "needs_page",
				severity:     "medium",
				summary:      fmt.Sprintf("Re-render %s — deferred section data now resolvable", page),
				spec:         spec,
				priority:     99,
				handlerAgent: "page-build-handler",
				status:       "triaged",
				createdBy:    "section-data-reconciler",
				itemKey:      fmt.Sprintf("page_rerender:%s", page),
				batchID:      batchID,
			}, logger); err != nil {
				return nil, fmt.Errorf("emit needs_page for %s: %w", page, err)
			}
			retriggered = append(retriggered, page)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
	}

	logger.Info("reconcile_section_data: done",
		zap.Int("scanned", scanned),
		zap.Int("retriggered_pages", len(retriggered)),
		zap.Int("skipped_human", skippedHuman))

	return map[string]interface{}{
		"scanned":           scanned,
		"retriggered_pages": retriggered,
		"skipped_human":     skippedHuman,
	}, nil
}

// hasItems reports whether a queryresolve result is a non-empty list.
// hasItems reports whether a resolved query value is a non-empty list. Built on
// the shared queryResultLen primitive (plan_sections_action.go) so the list-shape
// type-switch lives in exactly one place (bugs_open/054, council R1 reuse seat).
func hasItems(v interface{}) bool {
	n, isList := queryResultLen(v)
	return isList && n > 0
}
