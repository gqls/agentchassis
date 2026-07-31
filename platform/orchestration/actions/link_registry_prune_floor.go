// FILE: platform/orchestration/actions/link_registry_prune_floor.go
//
// A COMPLETENESS floor for extract_and_sync_links' reconciliation delete —
// bugs_open/165 site C, applying the rule shipped by bugs_closed/135
// (prune_floor.go, register CTXA-025).
//
// THE DEFECT. syncLinksToDB deletes every link_registry row for the source page
// and re-inserts whatever the extractor returned — `DELETE FROM link_registry
// WHERE source_page_id = $1`, then a per-link INSERT loop. The population comes
// from goquery parsing rendered HTML, which returns a short-but-nonzero result
// with no error whenever it is handed truncated or partially-rendered markup.
// Nothing checked that the run saw the whole page before deleting the rest.
//
// THE STAKES ARE LOWER THAN A's AND THE MECHANISM IS IDENTICAL. link_registry is
// regeneratable from the rendered HTML, so a short run costs a stale registry
// rather than lost customer-facing content. It is guarded anyway because the
// council's `bug_historian` seat objected to exactly the pattern of fixing one
// call site of a shared judgement and leaving the siblings heuristic — 016b §9,
// and case file 021_HANDOFF_durable_write_guard_covers_one_path_only.
//
// ONE COHORT, AND THE PARTITION IS DELIBERATELY DEFERRED. Sites A and B each
// partitioned after reading the live distribution, and each found the partition
// their own bug file proposed was wrong on the data. Here there is no
// distribution to read: link_registry holds ZERO rows, all-history, fleet-wide
// (verified 2026-07-31 — `SELECT count(*), max(created_at) FROM link_registry`
// returns 0 and NULL; the table has no retention job, so this is not a window
// artefact). The insert half has evidently never run to completion on any site.
//
// Guessing a partition from the shape of the SQL — per `link_type`, per `scope`
// — is precisely what bugs_open/165's own PLAN decision 3 warns against ("how a
// guard that fires on legitimate edits gets built, and a guard that cries wolf
// gets deleted by the first person it blocks"). So this ships the single
// unpartitioned cohort, which is sound at any distribution, and records what to
// measure when the corpus exists:
//
//	-- when link_registry is non-empty, before adding a partition:
//	SELECT link_type, scope, count(*) AS rows,
//	       count(DISTINCT source_page_id) AS pages,
//	       round(avg(n),2) AS avg_rows_per_page
//	FROM (SELECT link_type, scope, source_page_id, count(*) OVER (PARTITION BY source_page_id, link_type) AS n
//	      FROM link_registry) x GROUP BY 1,2 ORDER BY rows DESC;
//
// If most (page, link_type) groups hold one row, a per-link_type cohort is site
// A's rejected per-slot_name shape again and must not be added.
//
// INERT TODAY, BY CONSTRUCTION, AND THAT IS NOT AN ARGUMENT AGAINST IT. With the
// table empty every cohort reads Stored=0, which prune_floor treats as fully
// confirmed (a class appearing for the first time must never be able to refuse a
// prune), so the floor allows every prune until the table has rows. It arms
// itself the moment the insert half starts working — which is the state this
// guard exists for, and the state nobody will remember to add a guard in.
//
// NOT IN SCOPE, and named so it is not inherited as a finding: WHY the table is
// empty. ExtractAndSyncLinksAction returns a success-shaped
// {"links_extracted": N, "persisted": false} when site_id does not resolve, and
// the action runs on exactly one agent (multipage-website-builder) with 0
// orchestrations in the retained window — so "the exposure fires" and "the agent
// never runs" are indistinguishable from here. That is bugs_open/092's territory
// (the extractSiteID caller audit), not this floor's. [UNDETERMINED]

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// linkRegistryFloorKey is the step-config knob that resolves this floor. Same
// spelling as every other consumer — the floor is one concept across every
// reconciliation delete, and step config is per-step so the values cannot collide.
const linkRegistryFloorKey = "prune_floor_ratio"

// linkRegistryCompleteness holds one page's measurement, kept as a struct so the
// raw numbers are reported on a PASSING run too. A bare "links_persisted: 3" is
// the alarm presented as output — candidate (3) of bugs_open/135.
type linkRegistryCompleteness struct {
	LinksToWrite int // rows this run will insert
	LinksStored  int // rows the DELETE will remove (its exact complement)
	Cohorts      []pruneCohort
}

// measureLinkRegistryCompleteness counts the rows the DELETE would remove, using
// that delete's own predicate (source_page_id = $1) rather than a hand-rolled
// equivalent. Any other spelling is a second predicate free to drift from the one
// being guarded, and a guard measuring a different population from the one at
// risk reports a ratio for rows nobody was going to delete.
func measureLinkRegistryCompleteness(ctx context.Context, db *sql.DB, pageID uuid.UUID,
	linksToWrite int) (linkRegistryCompleteness, error) {

	m := linkRegistryCompleteness{LinksToWrite: linksToWrite}
	if db == nil {
		return m, fmt.Errorf("no DB handle")
	}

	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM link_registry WHERE source_page_id = $1`, pageID,
	).Scan(&m.LinksStored); err != nil {
		return m, fmt.Errorf("completeness measurement: %w", err)
	}

	m.Cohorts = []pruneCohort{
		{Label: "links", Confirmed: m.LinksToWrite, Stored: m.LinksStored},
	}
	return m, nil
}

// enforceLinkRegistryFloor is the guard, called immediately before syncLinksToDB.
// It returns the numbers for the action's result on a pass, and an error on a
// refusal — at which point the caller must return without deleting or inserting.
//
// FAILS CLOSED on an unmeasurable floor, for the same reason as its siblings: the
// measurement is the only thing standing between a half-blind extractor and an
// emptied registry, so when it cannot be taken the destructive half must not run.
//
// THE CONTRACT IS AN ERROR, IDENTICAL TO SITES A AND B, and getting here took two
// council rounds on corr c69e935a-7134-45c1-81c3-2f1da7831827. Round 2 made this
// one return (detail, bool) and SKIP the sync instead, because measurement showed
// its only live caller is nested — multipage-website-builder, at
// workflow.steps.generate_pages_loop.config.substeps.extract_links — and that loop
// sets no continue_on_error and the substep no error_step, so by coordinator.go's
// routing (shouldContinueLoopOnError false, routeToErrorStepOrFail falling through
// to failWorkflow) an error fails the WHOLE site build over one page's links.
//
// FOUR SEATS RULED THAT REASONING BACKWARDS AND THEY WERE RIGHT. The rationale
// named the real cause — the loop has no per-substep error routing — and then
// stepped around it instead of repairing it (constitution, HIGH). It also turned a
// refusal into a success-shaped return whose only signal is a queue
// (bug_historian, HIGH), invented a second failure semantics for one caller of a
// shared mechanism (architecture, and bug_historian's 093 parallel), and swapped a
// compiler-enforced error for a bool any future caller can ignore.
//
// So the contract is uniform again, and the real gap is filed at its own layer as
// bugs_open/173 (loop substeps cannot carry error routing at substep granularity;
// continue_on_error is all-or-nothing per loop). Deferring it here rather than
// setting continue_on_error on that loop is deliberate: that flag would change
// failure semantics for EVERY substep in the page-generation loop, turning a page
// that genuinely failed to build into a silently skipped one — the same silent-drop
// class, moved rather than removed. It is a workflow-owner's decision, not a rider
// on a bug fix.
//
// The disproportion is real but it is NOT introduced here: that loop already fails
// the whole build on any substep error, and this adds one more possible error to a
// loop that has many. With link_registry empty fleet-wide the floor is inert, so
// nothing is live today either way — which is what makes deferring honest rather
// than negligent.
//
// The surrounding action's OTHER success-shaped failure paths are bugs_open/092's
// business and are left exactly as they were.
func enforceLinkRegistryFloor(ctx context.Context, params ActionParams, siteID, pageID uuid.UUID,
	pageName string, linksToWrite int) (map[string]interface{}, error) {

	floor, fromConfig := pruneFloorFromConfig(params.StepConfig.Config, linkRegistryFloorKey, defaultPruneFloorRatio)
	subject := fmt.Sprintf("page %q", pageName)

	m, err := measureLinkRegistryCompleteness(ctx, params.DB, pageID, linksToWrite)
	if err != nil {
		reason := fmt.Sprintf("extract_and_sync_links: REFUSED for %s — the completeness floor could not be measured (%v), so no link row was deleted or written", subject, err)
		params.Logger.Error(reason)
		_ = emitLinkRegistryRefusal(ctx, params.DB, siteID, pageID, pageName, subject, reason, params.Logger)
		return nil, fmt.Errorf("%s", reason)
	}

	verdict := evaluatePruneFloor(floor, m.Cohorts)
	reason := verdict.Reason("extract_and_sync_links: resync", subject, linkRegistryFloorKey)

	if verdict.Clamped {
		params.Logger.Warn("extract_and_sync_links: prune_floor_ratio out of range — clamped",
			zap.String("page_name", pageName),
			zap.Float64("configured", verdict.Asked),
			zap.Float64("applied", verdict.Floor))
	}

	detail := pruneFloorDetail(verdict, fromConfig, reason, map[string]interface{}{
		"links_to_write": m.LinksToWrite,
		"links_stored":   m.LinksStored,
	})

	switch {
	case !verdict.Allowed:
		params.Logger.Error(reason,
			zap.String("page_name", pageName),
			zap.Int("links_to_write", m.LinksToWrite),
			zap.Int("links_stored", m.LinksStored))
		_ = emitLinkRegistryRefusal(ctx, params.DB, siteID, pageID, pageName, subject, reason, params.Logger)
		return nil, fmt.Errorf("%s", reason)
	case verdict.Disabled:
		params.Logger.Warn(reason, zap.String("page_name", pageName))
	default:
		params.Logger.Info(reason, zap.String("page_name", pageName))
	}
	return detail, nil
}

// emitLinkRegistryRefusal records the refusal where a human will see it, via the
// shared site-scoped surface in prune_floor.go.
func emitLinkRegistryRefusal(ctx context.Context, db *sql.DB, siteID, pageID uuid.UUID,
	pageName, subject, reason string, logger *zap.Logger) error {

	pID := pageID
	return emitPruneRefusalWorkItem(ctx, db, pruneRefusal{
		SiteID:   siteID,
		PageID:   &pID,
		Source:   "extract_and_sync_links",
		Pipeline: "build",
		ItemType: "link_sync_refused_incomplete",
		ItemKey:  fmt.Sprintf("link_sync_refused_incomplete:%s", pageName),
		Subject:  subject,
		Summary:  fmt.Sprintf("Link resync refused: %s extracted too few links to replace what is stored", subject),
		Reason:   reason,
		Fix: "A link extraction returned substantially fewer links than the registry holds for this page, " +
			"so NOTHING was deleted and the stored links still stand (bugs_open/165). Decide: if the page " +
			"genuinely lost that many links, lower prune_floor_ratio on the extract_and_sync_links step (0 " +
			"disables the floor); otherwise find why the extractor saw less — truncated or partially " +
			"rendered HTML reaching goquery is the usual cause.",
	}, logger)
}
