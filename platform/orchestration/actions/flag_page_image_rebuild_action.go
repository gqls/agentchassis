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
// Section-scoped imagery (scope_ref = "<page>:<ordinal>") re-renders its page
// too: plan_sections resolves section assets per page, so a landed section
// image needs the same needs_page emit as a hero. Derive the page from the
// scope_ref prefix and fall through to the page path.
//
// MECHANISM (reuses existing pieces)
//   1. flagPagesForRebuild — sets pages.build_status = 'needs_rebuild'.
//   2. insertWorkItem — emits needs_page → page-build-handler, which re-runs
//      plan_sections (re-resolving the hero) and re-deploys the page.
//   Priority 99 so the re-render runs after this site's imagery (≤98) and after
//   the terminal needs_rerender's reassembly, producing the final correct HTML.
//   item_key is stable per page (page_rerender:<page>, shared with the
//   section-data reconciler) so multiple image
//   completions for one page collapse to a single re-render via idx_swi_dedup.
//
// VERIFIED, AND GUARDED (bugs_open/187, 2026-08-03)
//   The page_name-only spec is fine: page-build-handler does re-derive page
//   context (page_id, type, sections) from the page record by name, which is how
//   reconcile-emitted needs_page items are consumed. What was NOT true is the
//   tacit half of that assumption — that a named page resolves any sections at
//   all. It measured false: 14 of the items this action minted parked for ever
//   in needs_human_review because the page they named declared none, so the
//   handler had nothing to build and no-oped them into the review queue.
//   The emit therefore asks the handler's own question first, at emit time and
//   read-only: pageSectionsSatisfiable (page_section_satisfiability.go) walks
//   load_page_sections_from_spec_action.go's fallbacks 1 to 3 and then the
//   current-plan membership that licenses its sibling synthesis. Unsatisfiable →
//   no item, an INFO log, and `needs_page_emit: skipped_sectionless_page` in the
//   return map. The needs_rebuild flag half stays UNCONDITIONAL: flagging a page
//   costs nothing, releases no dependency and parks nothing.
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
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var FlagPageImageRebuildInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"scope", "scope_ref"},
	// wire_hero_on_landing arms wirePageHeroOnLanding (bugs_open/412 candidate
	// 1) — opt-in, default OFF per the 2026-08-02 §2 ruling; armed for
	// image-build-handler by migration 710 (held until the carrying image
	// rolls). Declared here so the unknown-config-key audit reads it as
	// intentional the day the migration lands.
	ConfigKeys: []string{"wire_hero_on_landing"},
	Defaults:   map[string]interface{}{},
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

	// Section-scoped imagery (scope_ref = "<page>:<ordinal>") re-renders its page
	// too: plan_sections resolves section assets per page, so a landed section
	// image needs the same needs_page emit as a hero. Derive the page from the
	// scope_ref prefix and fall through to the page path.
	if scope == "section" && strings.Contains(pageName, ":") {
		pageName = strings.SplitN(pageName, ":", 2)[0]
		scope = "page"
		logger.Info("flag_page_image_rebuild: section-scoped imagery mapped to its page",
			zap.String("scope_ref", inputs.Get("scope_ref")),
			zap.String("page", pageName))
	}

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

	// 2. Emit needs_page so page-build-handler re-renders through plan_sections —
	//    but only when it would find sections to build. A page that resolves none
	//    and is in no current plan cannot be re-rendered by the handler; the item
	//    would no-op straight into needs_human_review and stay there
	//    (bugs_open/187). The skip is surfaced, not swallowed: an INFO log the
	//    fleet can grep and a disposition in the return map, because a guard that
	//    declines silently is indistinguishable from one that never ran.
	satisfiable, declared, sectionSource := pageSectionsSatisfiable(ctx, params.DB, logger, siteID, pageName)
	if !satisfiable {
		logger.Info("flag_page_image_rebuild: skipped_sectionless_page — the page resolves no sections and is in no current plan, so page-build-handler would park the item; see bugs_open/187",
			zap.String("site_id", siteID.String()),
			zap.String("page", pageName),
			zap.String("sections_source", sectionSource))
		return map[string]interface{}{
			"rebuilt":         false,
			"reason":          "sectionless_page",
			"page_name":       pageName,
			"flagged":         true,
			"needs_page_emit": "skipped_sectionless_page",
			"sections_source": sectionSource,
		}, nil
	}

	batchID := uuid.New()
	spec := fmt.Sprintf(`{%s"page_name":%q}`,
		livespec.RerenderReasonJSONPrefix(livespec.ReasonImageLanded), pageName)
	itemKey := fmt.Sprintf("page_rerender:%s", pageName)

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
		// An ACTION REQUEST — "re-render this page, its image just landed". An image can
		// land again (regeneration), and the second re-render is exactly as legitimate as
		// the first; a completed predecessor is a success, not a strike (bugs_open/326).
		recurrenceExpected: true,
	}, logger); err != nil {
		return nil, fmt.Errorf("emit needs_page re-render: %w", err)
	}
	// 2b. Wire the landed content hero into the page's STORED hero fields, in
	//     the same transaction (bugs_open/412 candidate 1) — opt-in, default
	//     OFF; see wire_page_hero_on_landing.go for the full argument. The
	//     stored value is what the LLM-free rerender serves and what
	//     carryStored preserves when plan_sections resolves nothing, so this is
	//     the deterministic floor under the re-render emitted above.
	wireEmit := "disabled"
	if heroWireArmed(params.StepConfig.Config) {
		wireEmit = wirePageHeroOnLanding(ctx, tx, siteID, pageName, logger)
	}
	// 3. Derive this page's listing card, in the SAME transaction, if the page
	//    now has a content hero and no card yet (bugs_open/114). See
	//    emitContentCardDerive for why this cannot wait for the sweep.
	cardEmit := emitContentCardDerive(ctx, tx, siteID, pageName, batchID, logger)
	// 4. Tell the LISTINGS that render this page (bugs_open/384), in the same
	//    transaction. The needs_page above re-renders the article; the pages
	//    that list it hold its image in a stored array that only a
	//    section_data_resolved re-render refreshes. Unless a card derive was
	//    just raised — then the card supersedes this image in the projection,
	//    and derive_card_asset makes the request when the card lands.
	listEmit := reresolvePageListsAfterPageImage(ctx, tx, siteID, pageName, cardEmit, batchID, logger)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("flag_page_image_rebuild: queued page re-render",
		zap.String("site_id", siteID.String()), zap.String("page", pageName),
		zap.String("sections_source", sectionSource),
		zap.String("hero_wire", wireEmit),
		zap.String("card_derive", cardEmit),
		zap.String("page_list_reresolve", listEmit),
		zap.Int("declared_sections", len(declared)))
	return map[string]interface{}{
		"rebuilt":             true,
		"page_name":           pageName,
		"needs_page_emit":     "raised",
		"sections_source":     sectionSource,
		"hero_wire":           wireEmit,
		"card_derive":         cardEmit,
		"page_list_reresolve": listEmit,
	}, nil
}
