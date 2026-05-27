// FILE: platform/orchestration/actions/flag_page_image_rebuild_action.go
//
// FlagPageImageRebuildAction re-renders the page affected by an image asset
// that was just generated + deployed, so the page picks up the now-present
// asset instead of the build-time fallback.
//
// WHY THIS EXISTS
//   Page hero images are produced asynchronously via needs_imagery, usually
//   AFTER the page was first rendered. At first render the hero field
//   (site_assets.hero) was unresolved and fell back to the static
//   /assets/images/hero.jpg. plan_sections now resolves the page's hero from
//   site_plan_imagery + assets (page-aware ensureAssets) — but that only takes
//   effect when the page is RE-rendered through plan_sections. The terminal
//   needs_rerender reassembles stored HTML; it does not re-resolve fields. This
//   action emits the re-render that does.
//
// SCOPE
//   Page-scoped imagery only (hero / section images on a specific page). The
//   needs_imagery spec carries scope + scope_ref (the page name). Site-scoped
//   imagery (logo) lives in the header — a SITE component rendered by
//   render_site_components, a separate path — so flagging a page rebuild would
//   not re-resolve it. Logo wiring is a follow-up in the header render path,
//   deliberately not handled here.
//
// MECHANISM (reuses existing pieces)
//   1. flagPagesForRebuild — sets pages.build_status = 'needs_rebuild'.
//   2. insertWorkItem — emits needs_page → page-build-handler, which re-runs
//      plan_sections (re-resolving the hero) and re-deploys the page.
//   Priority 99 so the re-render runs after this site's imagery (≤98) and after
//   the terminal needs_rerender's reassembly, producing the final correct HTML.
//   item_key is stable per page (img_rebuild:<page>) so multiple image
//   completions for one page collapse to a single re-render via idx_swi_dedup.
//
// VERIFY BEFORE RELYING ON IT
//   This emits needs_page with a page_name-only spec, on the assumption that
//   page-build-handler re-derives page context (page_id, type, sections) from
//   the page record by name — which is how reconcile-emitted needs_page items
//   are consumed. If page-build-handler instead requires richer spec fields,
//   switch this to flag needs_rebuild only and let reconcile re-emit, or copy
//   the reconcile spec shape.
//
// WIRING (image-build-handler workflow — terminal step, after store + deploy):
//   "flag_rebuild": {
//       "action": "flag_page_image_rebuild",
//       "config": {
//           "site_id":    "input_data.site_id",
//           "scope":      "input_data.spec.scope",
//           "scope_ref":  "input_data.spec.scope_ref"
//       },
//       "next_step": "complete"
//   }
//
// REGISTRATION (registry.go):
//   "flag_page_image_rebuild": {
//       Handler:     FlagPageImageRebuildAction,
//       Category:    "site",
//       Description: "Re-render a page after its image asset lands so the hero resolves",
//       IsLocal:     true,
//   }

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var FlagPageImageRebuildInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"scope", "scope_ref"},
	Defaults: map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("flag_page_image_rebuild", FlagPageImageRebuildInputSpec)
}

func FlagPageImageRebuildAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "flag_page_image_rebuild"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		FlagPageImageRebuildInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}

	scope := inputs.Get("scope")
	pageName := inputs.Get("scope_ref")

	// Only page-scoped imagery triggers a page re-render.
	if scope != "page" || pageName == "" {
		logger.Info("flag_page_image_rebuild: not page-scoped, nothing to re-render",
			zap.String("scope", scope), zap.String("scope_ref", pageName))
		return map[string]interface{}{"rebuilt": false, "reason": "not_page_scoped"}, nil
	}

	// 1. Flag the page needs_rebuild (reuse existing helper). Idempotent.
	if _, err := flagPagesForRebuild(ctx, params.DB, siteID, []string{pageName}, logger); err != nil {
		return nil, fmt.Errorf("flag page for rebuild: %w", err)
	}

	// 2. Emit needs_page so page-build-handler re-renders through plan_sections.
	batchID := uuid.New()
	spec := fmt.Sprintf(`{"reason":"image_landed","page_name":%q}`, pageName)
	itemKey := fmt.Sprintf("img_rebuild:%s", pageName)

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "image-build-handler",
		pipeline:     "build",
		itemType:     "needs_page",
		severity:     "medium",
		summary:      fmt.Sprintf("Re-render %s after its image asset landed", pageName),
		spec:         spec,
		priority:     99,
		handlerAgent: "page-build-handler",
		status:       "triaged",
		createdBy:    "image-build-handler",
		itemKey:      itemKey,
		batchID:      batchID,
	}, logger); err != nil {
		return nil, fmt.Errorf("emit needs_page re-render: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("flag_page_image_rebuild: queued page re-render",
		zap.String("site_id", siteID.String()), zap.String("page", pageName))
	return map[string]interface{}{"rebuilt": true, "page_name": pageName}, nil
}
