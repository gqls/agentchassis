// FILE: platform/orchestration/actions/nav_prune_floor.go
//
// A COMPLETENESS floor for populate_nav_tables' reconciliation delete —
// bugs_open/165 site B, applying the rule shipped by bugs_closed/135
// (prune_floor.go, register CTXA-025).
//
// THE DEFECT. PopulateNavTablesAction deletes EVERY nav row for the site —
// `DELETE FROM site_nav_items WHERE site_id = $1`, then the groups — and
// rebuilds them from the page list it just loaded. That is correct when the run
// saw the whole page corpus and wrong when it did not, and nothing checked
// which. The only existing guard is `len(pages) == 0`, which catches a TOTAL
// read failure and nothing short of it: loadPagesForNav's own scan loop logs a
// warning and `continue`s on a row it cannot scan, so a partial read is
// silent, success-shaped, and indistinguishable from a site that genuinely has
// fewer pages. The outcome is not a broken row — it is a site serving a nav bar
// missing half its links, which reads as "those pages were never in the nav".
//
// WHY NOT PER-GROUP COHORTS — AND THIS IS THE MEASURED CORRECTION TO WHAT
// bugs_open/165 PROPOSED. The bug file suggested "per nav group, plus distinct
// nav items". The per-group half is wrong on this table, for a reason that does
// not apply to site A: classifyPagesForNav RE-HOMES pages between groups as a
// matter of course. Measured 2026-07-31, robot-hands.com holds a `tools` group
// (created by hand SQL on 2026-07-31 12:27 — no Go code writes that group_key)
// containing /tools/gripper-safety-factor-calculator, and the current classifier
// places that same page in `utility`. A per-group cohort scores that legitimate
// re-homing as tools 0/1 = 0% and refuses the rebuild for ever. Group membership
// is a classifier OUTPUT, not an independently-losable class, so it cannot be
// the partition. (Site A rejected its own bug file's suggested partition on the
// data too — see prune_floor.go's consumer list. Measure, then partition.)
//
// THE TWO COHORTS, in genuinely different units:
//
//   - `pages seen` — the pages this run LOADED against the pages that exist
//     under the loader's own predicate. This is the completeness signal proper:
//     it asks "did this run see the corpus?" rather than "did this run write
//     less?", and it is the one that catches the silent scan-failure drop. It is
//     measured with navPageScopeSQL, the SAME predicate the loader's WHERE
//     clause is built from, so the count and the load cannot drift apart. A
//     second spelling would be a second predicate free to diverge from the one
//     being guarded, which is a guard reporting a ratio for rows nobody was
//     going to touch.
//
//   - `nav items` — the rows this run will INSERT against the rows the DELETE
//     will remove (site-wide, the delete's exact complement). This catches the
//     case the page load cannot: a classifier collapse or a mass flag flip that
//     loads every page and then places almost none of them. bugs_open/149 A2 was
//     exactly that shape — a URL-prefix rule that dropped every /tools/ page out
//     of nav entirely while the pages loaded fine.
//
// pages IS A GENUINELY INDEPENDENT SECOND OPINION. This action only ever READS
// pages.in_header / in_footer (the SELECT at loadPagesForNav) and never writes
// them; seven other actions do — apply_gap_plan, apply_adoption_plan,
// create_report_page, html_actions, maintenance_actions, site_db_actions,
// v3_site_actions. So the `pages seen` denominator is not an echo of this run's
// own output, which is the property that makes site A's `pages.sections` cohort
// worth having.
//
// NO RATCHET HERE, unlike site A, and the asymmetry is why this file has two
// cohorts and not three.
//
// CLARIFIED after a council objection (corr c69e935a, editquality seat, round 1)
// which was fair: the `nav items` cohort below IS a
// what-I-will-write vs what-is-stored comparison, i.e. the same SHAPE as site A's
// row cohort, so "no ratchet" needs more than an assertion. The distinction is
// not the shape of the comparison — it is whether the NUMERATOR has any memory of
// the previous damage:
//
//   - page_components (AUTHORED). The numerator is an LLM writer's output. After a
//     truncation cuts a page from 12 rows to 2, the NEXT run's writer is handed the
//     same truncated context and also returns about 2, so the cohort reads 2/2 =
//     100% and the damage IS the new baseline. Only a denominator from a different
//     source (pages.sections) can still see 12. Hence site A's plan cohort.
//   - site_nav_items (DERIVED). The numerator is recomputed from the page corpus by
//     deterministic Go on every rebuild. After a truncation cuts nav from 17 items
//     to 6, the next healthy run classifies the full page list and projects 17
//     again — against 6 stored. The ratio is 17/6 > 1, which passes, and the nav is
//     REPAIRED. The low stored value cannot suppress its own detection, because
//     nothing derives the numerator from it.
//
// So a derived table self-heals and an authored one ratchets, and the test for a
// future consumer is: does this run's numerator come from the same place the
// damage was written to? If yes, add a cohort in a different unit.
//
// The honest limit of this, also from that round (debug_historian, HIGH): a floor
// of this shape guards against NEW loss, and cannot detect loss that already
// happened and stabilised. If a historic defect removed a class of nav link and
// both the classifier and the stored rows now agree it is absent, both sides of
// every cohort here omit it equally and all cohorts read 100%. That is a real
// blind spot and it is NOT what this file claims to fix — the fleet-wide replay
// showing expected == stored on 16 of 16 sites demonstrates internal consistency
// with the CURRENT classifier, not that the nav is correct. Detecting historic
// absence is a different mechanism (a nav-drift CHECK against declared
// membership, which is bugs_open/149's territory), and the landmine on this file
// — nav links under /tools//blog//guides deleted with none put back — is an
// instance of exactly that other shape, not of this one.
//
// MEASURED FALSE-POSITIVE RATE: 0 of 16 sites. On 2026-07-31 the membership rule
// was replayed in SQL against production and the item count a rebuild would
// produce equalled the stored count on all 16 sites with nav (finetuning.uk
// 25=25, ai-agent-orchestration 24=24, gaswholesalers 23=23, leopardess 23=23,
// robot-hands 17=17, gamesdesign 12=12, dartsonline 9=9, fundamentallyai 9=9,
// idea.uk 8=8, vonc 8=8, oufe 8=8, relojistas 7=7, webdesign 5=5,
// vetcomparison 4=4, and both loan-calculator sites 1=1). Every cohort reads
// 100% today, so the floor is inert on the live fleet — which is the point, and
// also why a green run proves nothing about it.
//
// WHY THE REFUSAL FAILS THE WHOLE ACTION, like site A and unlike 135. The delete
// and the insert are one operation inside one transaction: refusing the delete
// but performing the insert would double every nav item. So the action returns
// an error before opening the transaction, leaving the existing good nav in
// place.

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// navPageScopeSQL is the page-scope predicate shared by loadPagesForNav's SELECT
// and the floor's denominator. It exists so the two cannot drift: the guard must
// measure the same population the run was trying to read, or it is judging a
// different corpus from the one at risk. $1 is the site id in both uses.
const navPageScopeSQL = `site_id = $1 AND status IN ('active', 'deployed', 'pending')`

// populateNavFloorKey is the step-config knob that resolves this floor. It
// deliberately shares the name every other consumer uses: the floor is one
// concept across every reconciliation delete, and a second spelling is how an
// operator who has learnt the first one gets surprised. Step config is per-step,
// so there is no collision between actions' values.
const populateNavFloorKey = "prune_floor_ratio"

// navCompleteness holds one site's measurement, kept as a struct so the raw
// numbers can be reported on a PASSING run too. A bare "items: 9" is the alarm
// presented as output — candidate (3) of bugs_open/135 restated for this table.
type navCompleteness struct {
	Domain       string // for the human sentence; falls back to the site id
	PagesLoaded  int    // pages loadPagesForNav actually returned
	PagesStored  int    // pages that exist under navPageScopeSQL
	ItemsToWrite int    // nav items this run will insert
	ItemsStored  int    // nav items the DELETE will remove (its exact complement)
	Cohorts      []pruneCohort
}

// measureNavCompleteness takes both denominators, and the domain for the
// refusal sentence, in one round trip.
func measureNavCompleteness(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	pagesLoaded, itemsToWrite int) (navCompleteness, error) {

	m := navCompleteness{PagesLoaded: pagesLoaded, ItemsToWrite: itemsToWrite}
	if db == nil {
		return m, fmt.Errorf("no DB handle")
	}

	var domain sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT
			(SELECT count(*) FROM pages WHERE `+navPageScopeSQL+`),
			(SELECT count(*) FROM site_nav_items WHERE site_id = $1),
			(SELECT domain FROM sites WHERE id = $1)
	`, siteID).Scan(&m.PagesStored, &m.ItemsStored, &domain)
	if err != nil {
		return m, fmt.Errorf("completeness measurement: %w", err)
	}
	m.Domain = domain.String

	m.Cohorts = []pruneCohort{
		{Label: "pages seen", Confirmed: m.PagesLoaded, Stored: m.PagesStored},
		{Label: "nav items", Confirmed: m.ItemsToWrite, Stored: m.ItemsStored},
	}
	return m, nil
}

// enforceNavPruneFloor is the guard itself, called immediately before the
// transaction that deletes and rebuilds the nav. It returns the numbers for the
// action's result on a pass, and an error on a refusal — at which point the
// caller must return without deleting or inserting anything.
//
// FAILS CLOSED on an unmeasurable floor. The measurement is the only thing
// standing between a half-blind page read and a stripped nav bar, so when it
// cannot be taken the destructive half must not proceed. That costs a failed
// build step, which is loud and recoverable; the alternative is a silently
// truncated nav, which is not.
func enforceNavPruneFloor(ctx context.Context, params ActionParams, siteID uuid.UUID,
	pagesLoaded, itemsToWrite int) (map[string]interface{}, error) {

	floor, fromConfig := pruneFloorFromConfig(params.StepConfig.Config, populateNavFloorKey, defaultPruneFloorRatio)

	m, err := measureNavCompleteness(ctx, params.DB, siteID, pagesLoaded, itemsToWrite)
	if err != nil {
		subject := fmt.Sprintf("site %s", siteID)
		reason := fmt.Sprintf("populate_nav_tables: REFUSED for %s — the completeness floor could not be measured (%v), so no nav row was deleted or written", subject, err)
		params.Logger.Error(reason)
		_ = emitNavPruneRefusal(ctx, params.DB, siteID, subject, reason, params.Logger)
		return nil, fmt.Errorf("%s", reason)
	}

	domain := m.Domain
	if domain == "" {
		domain = siteID.String()
	}
	subject := fmt.Sprintf("site %s", domain)

	verdict := evaluatePruneFloor(floor, m.Cohorts)
	reason := verdict.Reason("populate_nav_tables: rebuild", subject, populateNavFloorKey,
		"the whole rebuild was refused, so nothing was written either and the nav already stored still stands")

	if verdict.Clamped {
		params.Logger.Warn("populate_nav_tables: prune_floor_ratio out of range — clamped",
			zap.String("domain", domain),
			zap.Float64("configured", verdict.Asked),
			zap.Float64("applied", verdict.Floor))
	}

	detail := pruneFloorDetail(verdict, fromConfig, reason, map[string]interface{}{
		"pages_loaded":   m.PagesLoaded,
		"pages_stored":   m.PagesStored,
		"items_to_write": m.ItemsToWrite,
		"items_stored":   m.ItemsStored,
	})

	switch {
	case !verdict.Allowed:
		params.Logger.Error(reason,
			zap.String("domain", domain),
			zap.Int("pages_loaded", m.PagesLoaded),
			zap.Int("pages_stored", m.PagesStored),
			zap.Int("items_to_write", m.ItemsToWrite),
			zap.Int("items_stored", m.ItemsStored))
		_ = emitNavPruneRefusal(ctx, params.DB, siteID, subject, reason, params.Logger)
		return nil, fmt.Errorf("%s", reason)
	case verdict.Disabled:
		params.Logger.Warn(reason, zap.String("domain", domain))
	default:
		params.Logger.Info(reason, zap.String("domain", domain))
	}
	return detail, nil
}

// emitNavPruneRefusal records the refusal where a human will see it, via the
// shared site-scoped surface in prune_floor.go. Site-scoped with no single page
// at fault, so pageID is nil and the dedup key is the site.
func emitNavPruneRefusal(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	subject, reason string, logger *zap.Logger) error {

	return emitPruneRefusalWorkItem(ctx, db, pruneRefusal{
		SiteID:   siteID,
		Source:   "populate_nav_tables",
		Pipeline: "build",
		ItemType: "nav_rebuild_refused_incomplete",
		ItemKey:  fmt.Sprintf("nav_rebuild_refused_incomplete:%s", siteID),
		Subject:  subject,
		Summary:  fmt.Sprintf("Nav rebuild refused: %s saw too little of its page corpus to replace the stored nav", subject),
		Reason:   reason,
		Fix: "A nav rebuild would have replaced the stored navigation with substantially fewer items, " +
			"so NOTHING was deleted and the existing nav still stands (bugs_open/165). Decide: if the site " +
			"genuinely lost that many pages, lower prune_floor_ratio on the populate_nav_tables step (0 " +
			"disables the floor); otherwise find why the page read came back short — a scan failure inside " +
			"loadPagesForNav, pages moved out of active/deployed/pending, or an upstream step that failed " +
			"without erroring.",
	}, logger)
}
