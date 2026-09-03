// FILE: platform/orchestration/actions/discovery_checks/check_section_source_drift.go
//
// Detects pages whose section list disagrees across the THREE stores that
// load_page_sections_from_spec reads in priority order:
//   1. site_plan_sections table (site_plans family) — AUTHORITATIVE
//   2. site_specs.site_plan aspect JSON (older planner generation)
//   3. pages.sections (materialised cache / what assembly & most edits touch)
//
// The build resolves the highest-priority source that has the page and SYNCS
// it DOWN over pages.sections. So if someone edits only pages.sections (or the
// aspect) while a higher source still holds a different list, the next rebuild
// silently reverts the edit and resurrects the old layout. That exact trap bit
// the robot-hands product-detail component swap (2026-07-15): migration 153
// updated pages.sections + the aspect, but not the authoritative table, so the
// rebuild brought the deleted components back.
//
// This check computes, per page, the EFFECTIVE authoritative list (table if
// present, else aspect if present) and compares it to pages.sections. A
// persistent mismatch means the page has not been rebuilt since the sources
// diverged — a latent revert waiting for the next build.
//
// Ordered comparison: "sections" is a layout, so [a,b] != [b,a] is real drift.
//
// LOCKED LIVE ROWS ARE NOT DRIFT (2026-08-15, bugs_open/285, register
// LOCK-008). The loader now merges the page's human-locked page_components
// rows into the list it assembles and syncs THAT into pages.sections — so on
// a page carrying a locked section the plan does not name, the authoritative
// table and the cache legitimately differ by exactly that section, and the
// next rebuild will NOT revert anything. Both sides of the comparison are
// therefore viewed through the same merge (datahelpers.MergeLockedPageSlots,
// the loader's own function): a cache that pre-dates the fix (raw plan) and a
// cache written after it (plan + locked) both compare equal to the merged
// authoritative list; a genuine edit to one store still differs. Comparing
// raw lists here would have filed one `section_source_drift` item per fixed
// page (13 fleet-wide the day this shipped) — the check would have been the
// noise the fix removed.
//
// ════════════════════════════════════════════════════════════════════════════
// 2026-09-03, bugs_open/469 — THIS CHECK NOW CLOSES ITS OWN ITEMS, AND THE
// HARD PART IS THAT CLOSING THEM NAIVELY WOULD BE WORSE THAN NEVER CLOSING.
// ════════════════════════════════════════════════════════════════════════════
//
// Until now the check was flag-only and nothing on the estate ever closed one
// of its items. Three properties compounded (469 §3):
//
//   - nothing closed them, so they accumulated — six open, the oldest 37 days;
//   - the item's `spec` is frozen at filing time and reads as CURRENT, so a
//     reader triaging the backlog learns nothing about today;
//   - an open item SUPPRESSES re-filing (idx_swi_dedup is UNIQUE on
//     (site_id, item_key) over non-terminal statuses), so the detector went
//     BLIND on exactly the pages it had already flagged.
//
// The sequence that followed is the bug: drift detected → flagged → no handler
// → the build wins → the stores agree again → the item now describes a state
// that no longer exists and LOOKS RESOLVED. On robot-hands.com/gripper-catalog
// that ran for 37 days and a section a human had deliberately added was
// destroyed — the very component migration 154 was written to rescue in July.
//
// ── THE SHARP EDGE, and it is the whole design ──────────────────────────────
//
// For THIS check, "the finding no longer reproduces" and "the damage
// completed" are THE SAME OBSERVATION. The stores agree again precisely
// BECAUSE the build synced the authority down over the cache. So a retraction
// on agreement alone would close, automatically and fleet-wide, exactly the
// cases that most need a human — it would convert a silence into a
// certificate. Migration 753 ran this classification by hand on 2026-09-03 and
// three of six pages had resolved AGAINST the cache.
//
// So the closer here does not ask "do the stores agree?" It asks WHAT WAS
// LOST, and it cannot retract a lossy resolution without filing the receipt in
// the same transaction (ResolvedFinding.Receipt — the coupling lives in
// resolveWorkItems, not here, so it protects every caller of that seam).
//
// ── WHY `lost` AND NOT `direction` IS THE PREDICATE ─────────────────────────
//
// `direction` (753's three-way equality: cache_held / authority_won /
// third_list) is the LABEL, recorded on every close because LANDMINES tells
// readers to query `result->>'direction'`. It is not the TEST, because it
// cannot separate two cases with opposite meanings:
//
//   - oufe.com/contact — the cache had LOST `contact-info` and the authority
//     RESTORED it. `authority_won`, and nothing was destroyed.
//   - robot-hands.com/gripper-catalog — the authority DELETED
//     `gripper-spec-sheet`. Also `authority_won`, and a live section is gone.
//
// The receipt therefore fires on `lost != ∅`: names the frozen cache held that
// the frozen authority did not, which are absent from the page today. That is
// a statement about consequences, and it grades both cases correctly.
//
// ── DETECTION-TIME GRADING: the only part that acts BEFORE the loss ─────────
//
// Severity was `medium` for every inequality. A divergence whose sync-down
// will DROP a section the page currently carries is a different object from
// one that only ADDS, and the summary now names what the next build will
// destroy rather than leaving a reader to diff two lists by eye.
//
// Registration: automatic via init() -> Register(&SectionSourceDriftCheck{})
// Enable: add "section_source_drift" to completeness-discovery-agent's
//   {workflow,steps,run_checks,config,checks} array. (Live there since
//   migration 155; confirmed against agent_definitions 2026-09-03.)

package discovery_checks

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&SectionSourceDriftCheck{}) }

type SectionSourceDriftCheck struct{}

func (c *SectionSourceDriftCheck) Name() string { return "section_source_drift" }

// ⚠ THESE CONSTANTS ARE DELIBERATELY NOT USED AT THE ItemType FIELD SITES.
//
// verifier_coverage_test.go's SENSOR half scans this package's SOURCE for a
// quoted item type in an ItemType field, so that a new type fails the build the
// moment it is written. A constant is invisible to that scan, and the escape
// hatch (computedItemTypeSites) would make BOTH of this file's types invisible
// to CLASSIFICATION too — a hole in the guard exactly where it claims coverage.
// So the field sites spell the string, these constants serve the keys and
// queries, and TestSectionDriftItemTypeConstantsMatchTheStrings pins the two
// together so they cannot drift.
//
// ⚠⚠ AND DO NOT WRITE THAT FIELD NAME FOLLOWED BY A QUOTED LOWERCASE WORD IN A
// COMMENT IN THIS PACKAGE. A source-scanning test makes your COMMENTS
// load-bearing: the first draft of this very comment quoted the pattern it was
// describing, and the guard duly reported that this file produces an item type
// called "literal". The example is now described, not spelled.
const (
	sectionDriftItemType = "section_source_drift"
	// sectionCompositionLostItemType records a divergence that RESOLVED BY
	// DESTROYING something. It is filed only by this check's retraction path,
	// and only ever alongside the retraction it justifies.
	sectionCompositionLostItemType = "section_composition_lost"
)

// maxSectionDriftFlagsPerPass bounds noise on a badly-drifted site.
//
// ⚠ THE RETRACTION PASS MUST NOT LIVE BEHIND THIS CAP, and it does not — it
// runs before the filing loop. The recorded landmine (check_archived_page_
// still_serving.go): "a monotonic check's early return makes its new
// retraction INERT on exactly the sites that need it", because the site with
// nothing to file is the site whose stale items most need closing.
const maxSectionDriftFlagsPerPass = 25

// sectionSourceQuerier is the read surface the three section-source loaders
// need. It is an interface rather than *sql.DB so the same loaders can run
// inside a transaction — which is what RFC_064's proposed `apply_page_composition`
// needs in order to re-run the drift predicate as a postcondition BEFORE it
// commits, rather than discovering afterwards that it wrote a state the build
// would resolve differently.
//
// Both *sql.DB and *sql.Tx satisfy it. Deliberately unexported for now: an
// exported helper with no caller reads as a finished refactor to the next
// person, and this estate has a landmine on exactly that. The RFC_064 lane's
// own commit is what exports it, at which point it has a caller.
type sectionSourceQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (c *SectionSourceDriftCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	tableSections, err := loadPlanTableSections(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	aspectSections, err := loadAspectSections(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	cacheSections, cachePageIDs, err := loadPagesCacheSections(dctx.Ctx, dctx.DB, dctx.SiteID, dctx.Logger)
	if err != nil {
		return nil, err
	}
	// Locked live rows per page — the loader merges these into its list and
	// into the cache, so the comparison must see them on both sides. Loud on
	// failure: comparing raw lists would flag every locked page as drift.
	lockedByPage, err := datahelpers.LoadLockedPageSlotsForSite(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("section_source_drift: %w", err)
	}
	// Live slot names per page, for the consequence grade. Loud on failure for
	// the same reason every other load here is: a silently-empty map would
	// grade every drop as harmless, which is the direction that loses data.
	slotRows, err := datahelpers.LoadPageSlotRowsForSite(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("section_source_drift: live slot identities: %w", err)
	}
	liveSlots := map[string][]string{}
	for _, r := range slotRows {
		liveSlots[r.PageName] = append(liveSlots[r.PageName], r.Slot)
	}

	result := &CheckResult{}

	// ── PASS 1: RETRACTIONS, before the filing loop and before the cap. ──────
	// A failure to read the open items is NOT fatal to the whole check: the
	// filing half is still worth running, and a check that returns an error
	// retracts nothing anyway (the runner skips Resolved on error). Degrading
	// loudly here, rather than failing, keeps a transient work-item read
	// problem from also blinding detection.
	if err := c.appendRetractions(dctx, result, tableSections, aspectSections,
		cacheSections, cachePageIDs, lockedByPage); err != nil {
		dctx.Logger.Warn("section_source_drift: retraction pass degraded — filing continues, nothing retracted this run",
			zap.Error(err))
	}

	// ── PASS 2: FILING. ─────────────────────────────────────────────────────
	emitted := 0

	// Iterate the cache set — every deployed page has a pages row. A page with
	// no cache entry but a plan entry is the sectionless case (owned by
	// check_sectionless_pages), not drift.
	for pageName, cache := range cacheSections {
		authoritative, authSource, ok := effectiveAuthoritative(pageName, tableSections, aspectSections)
		if !ok {
			// No higher source — pages.sections is authoritative for this page;
			// nothing can silently override it. Not drift.
			continue
		}

		// Both sides through the loader's own merge: a locked live section the
		// plan does not name is membership the rebuild will KEEP, not drift.
		mergedAuth, mergedCache := mergeBothSides(authoritative, cache, lockedByPage[pageName])
		if orderedListsEqual(mergedAuth, mergedCache) {
			continue
		}

		if emitted >= maxSectionDriftFlagsPerPass {
			dctx.Logger.Info("section_source_drift: per-pass cap reached",
				zap.Int("cap", maxSectionDriftFlagsPerPass))
			break
		}

		// ── The consequence grade (469). What will the next build DO? ───────
		wouldDrop := sectionMultisetDiff(mergedCache, mergedAuth)
		wouldAdd := sectionMultisetDiff(mergedAuth, mergedCache)
		wouldDropPresent := namesPresentOnPage(wouldDrop, liveSlots[pageName])
		reorderedOnly := len(wouldDrop) == 0 && len(wouldAdd) == 0

		severity := "medium"
		if len(wouldDropPresent) > 0 {
			severity = "high"
		}

		finding := map[string]interface{}{
			"check":                "section_source_drift",
			"page":                 pageName,
			"authoritative_source": authSource,
			"authoritative":        authoritative,
			"pages_sections":       cache,
			"would_drop":           wouldDrop,
			"would_add":            wouldAdd,
			"would_drop_present":   wouldDropPresent,
			"reordered_only":       reorderedOnly,
			"severity":             severity,
		}
		result.Findings = append(result.Findings, finding)

		specMap := map[string]interface{}{
			"check":                "section_source_drift",
			"page_name":            pageName,
			"authoritative_source": authSource,
			"authoritative":        authoritative,
			"pages_sections":       cache,
			// The consequence, computed at filing so a triager does not have to
			// diff two lists by eye — and so the retraction pass can tell later
			// whether a lost section had been on the page at the time.
			"would_drop":         wouldDrop,
			"would_add":          wouldAdd,
			"would_drop_present": wouldDropPresent,
			"reordered_only":     reorderedOnly,
			"reason": "the authoritative section source disagrees with pages.sections; " +
				"the next rebuild will overwrite pages.sections with the authoritative list, " +
				"reverting any edit made only to pages.sections (or only to the aspect)",
			"fix": "Correct the AUTHORITATIVE store, never pages.sections alone — writing the " +
				"cache is what causes this (migrations 153 and 719/727/728 both did it, and " +
				"both were reverted by the next build). Template: migration 750 " +
				"(rename in place at a fixed `ordering`; NOT 154's delete-renumber-insert, " +
				"because `ordering` is a positional join key for assigned_fact_ids, subject, " +
				"page_components.position and site_plan_imagery.scope_ref). If the page is " +
				"already built_from_plan_version = the current plan, a corrected plan will " +
				"NOT render — that is RFC_064 §7 q2, open with the owner.",
		}
		spec, mErr := json.Marshal(specMap)
		if mErr != nil {
			continue
		}

		summary := fmt.Sprintf("Section-list drift on page '%s': %s has [%s] but pages.sections has [%s]",
			pageName, authSource, strings.Join(authoritative, ", "), strings.Join(cache, ", "))
		// Lead with the consequence when there is one: what the next build will
		// DESTROY is the part a human triages on, and it was previously
		// recoverable only by diffing the two lists in the sentence above.
		if len(wouldDropPresent) > 0 {
			summary = fmt.Sprintf("Next build will DROP [%s] from page '%s' (on the page now); %s",
				strings.Join(wouldDropPresent, ", "), pageName, summary)
		} else if len(wouldDrop) > 0 {
			summary = fmt.Sprintf("Next build will drop [%s] from page '%s' (not currently on the page); %s",
				strings.Join(wouldDrop, ", "), pageName, summary)
		}

		item := WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "section_source_drift", // literal, not the const — see the pin test
			Severity: severity,
			Summary:  summary,
			SpecJSON: string(spec),
			Priority: 130,
			// Flag-only: no handler auto-aligns planning sources; a human picks
			// the intended layout and updates all sources (cf. migration 154).
			HandlerAgent: "",
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      sectionDriftItemKey(pageName),
			BatchID:      dctx.BatchID,
		}
		if id, ok := cachePageIDs[pageName]; ok {
			pid := id
			item.PageID = &pid
		}
		result.WorkItems = append(result.WorkItems, item)
		emitted++
	}

	if emitted > 0 {
		dctx.Logger.Info("section_source_drift: flagged pages with divergent section sources",
			zap.Int("count", emitted))
	}
	return result, nil
}

func sectionDriftItemKey(pageName string) string {
	return sectionDriftItemType + ":" + pageName
}

// effectiveAuthoritative returns the list that will WIN for this page, and its
// store, or ok=false when no store outranks pages.sections.
func effectiveAuthoritative(pageName string, table, aspect map[string][]string) ([]string, string, bool) {
	if t, ok := table[pageName]; ok {
		return t, "site_plan_sections", true
	}
	if a, ok := aspect[pageName]; ok {
		return a, "site_specs.site_plan", true
	}
	return nil, "", false
}

// mergeBothSides puts the authoritative list and the cache through the LOADER'S
// OWN merge, so "already in the list" means the same thing here as it does in
// the code whose behaviour this check predicts.
func mergeBothSides(authoritative, cache []string, locked []datahelpers.LockedPageSlot) (mergedAuth, mergedCache []string) {
	mergedAuth, _, _ = datahelpers.MergeLockedPageSlots(authoritative, locked)
	mergedCache, _, _ = datahelpers.MergeLockedPageSlots(cache, locked)
	return mergedAuth, mergedCache
}

// sectionMultisetDiff returns the entries of `have` that `want` does not match
// one-for-one, in have's order.
//
// MULTISET, NOT SET, and the difference is load-bearing. One live page's plan
// names `generic-text-block` SIX times (measured 2026-09-03, an
// illustrated-text-block swap). A set difference between a list naming it six
// times and one naming it five reports "nothing dropped" — the exact
// under-report that lets a destroyed section pass as benign.
func sectionMultisetDiff(have, want []string) []string {
	budget := make(map[string]int, len(want))
	for _, s := range want {
		budget[s]++
	}
	out := []string{}
	for _, s := range have {
		if budget[s] > 0 {
			budget[s]--
			continue
		}
		out = append(out, s)
	}
	return out
}

// namesPresentOnPage returns the subset of `names` that a live page_components
// row on this page carries, by slot name.
//
// TWO MEASURED LIMITS, both stated because a grade nobody can read the terms of
// is a claim rather than evidence [MEASURED 2026-09-03, live page_components,
// 3,333 rows]:
//
//   - The projection carries no build_status, so a row already marked `removed`
//     still matches. That OVER-grades to `high`, which is the safe direction,
//     and `removed` is 56 rows of 3,333 (1.7%).
//   - It matches by NAME, so this says "the page carries a row under that slot
//     name", not "that section is deployed". The spec key is therefore
//     `would_drop_present`, not `..._deployed` — the check has not measured
//     deployment and must not say it has.
//
// slot_name is never NULL on the live table (0 of 3,333), so the name match has
// no silent under-report through a missing slot.
func namesPresentOnPage(names []string, liveSlots []string) []string {
	if len(names) == 0 || len(liveSlots) == 0 {
		return []string{}
	}
	have := make(map[string]bool, len(liveSlots))
	for _, s := range liveSlots {
		if s == "" {
			continue
		}
		have[s] = true
		have[datahelpers.NormalizeComponentFunction(s)] = true
	}
	out := []string{}
	for _, n := range names {
		if have[n] || have[datahelpers.NormalizeComponentFunction(n)] {
			out = append(out, n)
		}
	}
	return out
}

// ─── The retraction half ────────────────────────────────────────────────────

// driftItemSnapshot is one open item's FROZEN filing-time evidence.
//
// The frozen lists are not stale data to be refreshed — they are the ONLY
// surviving record of what the page's composition was before the build ran, and
// `lost` is computed FROM them. (refreshOnConflict exists on the write seam and
// is deliberately not used here: refreshing this spec would erase the baseline.)
type driftItemSnapshot struct {
	ItemKey       string
	ItemID        string
	PageName      string
	AuthSource    string
	Authoritative []string
	PagesSections []string
	// WouldDropPresent is present only on items filed after 2026-09-03; older
	// items have none, and the receipt says so rather than implying zero.
	WouldDropPresent []string
	HasConsequence   bool
	FiledAt          string
}

// driftResolution is one open item re-observed against today's stores.
type driftResolution struct {
	item      driftItemSnapshot
	pageID    *uuid.UUID
	direction string   // cache_held | authority_won | third_list — 753's vocabulary
	serving   string   // which store serves the page today
	agreed    []string // the merged list both stores now hold
	lost      []string // THE PREDICATE: frozen-cache-only names, absent today
	gained    []string
	reordered bool
}

// classifyDriftResolution is pure and is the whole judgement. Kept separate
// from every query so it is table-testable and so a mutation to it fails a unit
// test rather than needing a database.
func classifyDriftResolution(it driftItemSnapshot, pageID *uuid.UUID,
	rawCacheToday, mergedToday []string, serving string) driftResolution {

	r := driftResolution{item: it, pageID: pageID, agreed: mergedToday, serving: serving}

	// The LABEL, byte-compatible with migration 753 so `result->>'direction'`
	// answers the same question however a row was closed.
	switch {
	case orderedListsEqual(rawCacheToday, it.PagesSections):
		r.direction = "cache_held"
	case orderedListsEqual(rawCacheToday, it.Authoritative):
		r.direction = "authority_won"
	default:
		r.direction = "third_list"
	}

	// The TEST. What the cache held and the authority did not — minus whatever
	// is on the page today.
	//
	// Subtracting the MERGED list rather than the raw cache is not a detail: a
	// human-locked row the plan never names sits in the frozen cache and not in
	// the frozen authority, so it lands in `cacheOnly` — and it is still on the
	// page. Against the raw list it would read as lost, and this check would
	// file a false destruction receipt on every locked page on the estate.
	cacheOnly := sectionMultisetDiff(it.PagesSections, it.Authoritative)
	r.lost = sectionMultisetDiff(cacheOnly, mergedToday)
	r.gained = sectionMultisetDiff(mergedToday, it.PagesSections)
	r.reordered = len(r.lost) == 0 && len(r.gained) == 0 &&
		!orderedListsEqual(mergedToday, it.PagesSections)
	return r
}

// evidence is what lands in the closed row's `result`, beside resolved_by and
// reason. Shaped to answer, from the row alone, the question a reader of a
// closed drift item actually has: which side won, and did anything die.
func (r driftResolution) evidence() map[string]interface{} {
	return map[string]interface{}{
		"direction":       r.direction,
		"serving_source":  r.serving,
		"agreed_list":     r.agreed,
		"lost_sections":   r.lost,
		"gained_sections": r.gained,
		"reordered":       r.reordered,
		"closed_by_check": "section_source_drift",
	}
}

// sectionCompositionLostKey makes distinct losses distinct findings while an
// identical repeat dedups — the page_divergence_overwritten key design.
//
// The digest is over the ORDERED lost multiset, so losing [a] and later [a,b]
// are two findings, and gripper-spec-sheet's real history (added, destroyed,
// rescued by hand, destroyed again) files a fresh receipt the second time once
// a human has closed the first.
func sectionCompositionLostKey(pageName string, lost []string) string {
	sum := sha256.Sum256([]byte(strings.Join(lost, "\x1f")))
	return fmt.Sprintf("%s:%s:%s", sectionCompositionLostItemType, pageName, hex.EncodeToString(sum[:])[:12])
}

// retraction is THE ONLY place in this file that builds a ResolvedFinding, and
// receipt() is the only place that names sectionCompositionLostItemType.
//
// That is the structural pairing, and it is deliberately one function rather
// than a rule in a comment: a caller cannot obtain the retraction for a lossy
// resolution without the receipt attached, because the same value computed
// both. A source-pin test asserts the file holds exactly one of each literal,
// so a second constructor fails the build rather than review.
func (r driftResolution) retraction(dctx DiscoveryCheckContext, historyRows map[string]int) ResolvedFinding {
	f := ResolvedFinding{
		ItemType: "section_source_drift", // literal, NOT the const: see the pin test
		ItemKey:  r.item.ItemKey,
		Reason: fmt.Sprintf("re-read: %s and pages.sections agree again on page %q (direction=%s, lost=[%s], gained=[%s])",
			r.serving, r.item.PageName, r.direction,
			strings.Join(r.lost, ", "), strings.Join(r.gained, ", ")),
		Evidence: r.evidence(),
	}
	if len(r.lost) == 0 {
		return f
	}
	rc := r.receipt(dctx, historyRows)
	f.Receipt = &rc
	f.Evidence["receipt_item_key"] = rc.ItemKey
	f.Evidence["receipt_item_type"] = rc.ItemType
	return f
}

// receipt builds the durable record of a completed composition loss.
//
// IT COPIES ITS EVIDENCE, IT DOES NOT POINT AT IT. The drift row this receipt
// justifies is being closed in the same transaction, and site_work_items is a
// ROLLING WINDOW — a closed row is archived out of the table a reader would
// query. A receipt holding only `drift_item_id` would resolve to nothing
// exactly when someone finally came to read it.
func (r driftResolution) receipt(dctx DiscoveryCheckContext, historyRows map[string]int) WorkItemSpec {
	// `high` when there is evidence the lost section had really been on the
	// page: either the filing-time consequence grade said so, or the artefact
	// archive holds rows for it.
	severity := "medium"
	for _, name := range r.lost {
		if historyRows[name] > 0 || containsString(r.item.WouldDropPresent, name) {
			severity = "high"
			break
		}
	}

	spec := map[string]interface{}{
		"check":                          "section_source_drift",
		"page_name":                      r.item.PageName,
		"drift_item_id":                  r.item.ItemID,
		"drift_item_key":                 r.item.ItemKey,
		"drift_filed_at":                 r.item.FiledAt,
		"direction":                      r.direction,
		"serving_source_today":           r.serving,
		"authoritative_source_at_filing": r.item.AuthSource,
		"authoritative_at_filing":        r.item.Authoritative,
		"pages_sections_at_filing":       r.item.PagesSections,
		"agreed_list_today":              r.agreed,
		"lost_sections":                  r.lost,
		"gained_sections":                r.gained,
		"reordered":                      r.reordered,
		"history_rows_for_lost_sections": historyRows,
		"consequence_recorded_at_filing": r.item.HasConsequence,
		"would_drop_present_at_filing":   r.item.WouldDropPresent,
		"recovery": "WHAT SURVIVES AND WHAT DOES NOT. The lost section's BYTES are recoverable " +
			"when it had rendered: page_component_history holds every deleted page_components row " +
			"(migration 357's trigger pair, from ~2026-08-09). " +
			"SELECT slot_name, position, rendered_html, divergence, created_at " +
			"FROM page_component_history WHERE page_id = <page_id> " +
			"AND source = 'artefact_archive_trigger' AND op = 'delete' " +
			"AND slot_name = ANY(<lost_sections>) ORDER BY created_at DESC. " +
			"⚠ Every rebuild deletes every row, so a delete row proves nothing on its own — " +
			"read it for the CONTENT, not as proof of a loss. " +
			"THE LIST IS ARCHIVED NOWHERE: `pages_sections_at_filing` above is the only surviving " +
			"record that the page held these sections in this order. " +
			"REPAIR: correct the AUTHORITATIVE store, never pages.sections alone — that is what " +
			"caused this. Template: migration 750 (rename in place at a fixed `ordering`; NOT " +
			"154's delete-renumber-insert, because `ordering` is a positional join key). If the " +
			"section must be INSERTED rather than renamed, or the page is already " +
			"built_from_plan_version = the current plan, a corrected plan will not render and the " +
			"stamp question is RFC_064 §7 q2, open with the owner. " +
			"IF THE REMOVAL WAS INTENDED, close this item — that is a valid outcome and the reason " +
			"no handler is attached (bugs_open/469 §5.1).",
	}
	specJSON, _ := json.Marshal(spec)

	item := WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: "content",
		ItemType: "section_composition_lost", // literal, NOT the const: see the pin test
		Severity: severity,
		Summary: fmt.Sprintf("Composition LOST on page '%s': [%s] was in pages.sections when drift was filed (%s) and is gone; %s won",
			r.item.PageName, strings.Join(r.lost, ", "), r.item.FiledAt, r.serving),
		SpecJSON: string(specJSON),
		Priority: 30,
		// No handler, on purpose: a machine cannot tell a deliberate removal
		// from this bug's completion (469 §5.1). Both are valid endings and only
		// someone who knows what the page is for can say which.
		HandlerAgent: "",
		Status:       "needs_human_review",
		CreatedBy:    dctx.AgentType,
		ItemKey:      sectionCompositionLostKey(r.item.PageName, r.lost),
		BatchID:      dctx.BatchID,
		// A second loss on the same page is a SECOND EVENT, not a repeat — this
		// component was destroyed, rescued by hand in July, and destroyed again.
		// Without this the anti-churn brake drops or brands the re-file.
		RecurrenceExpected: true,
	}
	if r.pageID != nil {
		pid := *r.pageID
		item.PageID = &pid
	}
	return item
}

func containsString(hay []string, needle string) bool {
	for _, h := range hay {
		if h == needle {
			return true
		}
	}
	return false
}

// appendRetractions re-observes this site's open drift items and appends a
// retraction for each one it can positively judge.
//
// IT NEVER INFERS FROM ABSENCE. Every retraction below is a statement about a
// `pages` row that was READ and compared. A page whose row has vanished, or
// whose cache is empty, yields NO observation and its item is left open —
// which is the CheckResult.Resolved contract, and the check_empty_sections
// stance ("a slot with no deployed component is equally 'fixed' and 'silently
// deleted'; refuse both").
func (c *SectionSourceDriftCheck) appendRetractions(
	dctx DiscoveryCheckContext,
	result *CheckResult,
	tableSections, aspectSections, cacheSections map[string][]string,
	cachePageIDs map[string]uuid.UUID,
	lockedByPage map[string][]datahelpers.LockedPageSlot,
) error {

	items, err := loadOpenDriftItems(dctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	// A DEMAND CONTROL for the "no authority today" branch below. "No rows for
	// this page" is only an observation if the read itself worked — and a read
	// that returned nothing for the WHOLE SITE is indistinguishable from a
	// mid-replan window or a failed lookup that degraded to empty.
	authorityReadProductive := len(tableSections) > 0 || len(aspectSections) > 0

	var disclaimed, notObserved, stillDrifted int

	for _, it := range items {
		cache, present := cacheSections[it.PageName]
		if !present {
			// The page row is gone, deleted, or holds an empty list. Not an
			// observation about this item's claim.
			notObserved++
			continue
		}

		authoritative, serving, hasAuthority := effectiveAuthoritative(it.PageName, tableSections, aspectSections)
		if !hasAuthority {
			if !authorityReadProductive {
				notObserved++
				continue
			}
			// pages.sections is now the only store for this page, so nothing can
			// silently override it — the finding's premise is gone. Serving
			// source is the cache itself.
			authoritative, serving = cache, "pages.sections"
		}

		mergedAuth, mergedCache := mergeBothSides(authoritative, cache, lockedByPage[it.PageName])
		if !orderedListsEqual(mergedAuth, mergedCache) {
			stillDrifted++
			continue
		}

		var pageID *uuid.UUID
		if id, ok := cachePageIDs[it.PageName]; ok {
			p := id
			pageID = &p
		}

		res := classifyDriftResolution(it, pageID, cache, mergedCache, serving)

		historyRows := map[string]int{}
		if len(res.lost) > 0 && pageID != nil {
			historyRows = countArchivedSlotRows(dctx, *pageID, res.lost)
		}
		result.Resolved = append(result.Resolved, res.retraction(dctx, historyRows))

		if len(res.lost) > 0 {
			dctx.Logger.Warn("section_source_drift: a divergence resolved by DESTROYING sections — retracting WITH a receipt (bugs_open/469)",
				zap.String("page", res.item.PageName),
				zap.String("direction", res.direction),
				zap.Strings("lost", res.lost),
				zap.Strings("gained", res.gained))
		}
	}

	if disclaimed+notObserved+stillDrifted > 0 || len(result.Resolved) > 0 {
		dctx.Logger.Info("section_source_drift: retraction pass",
			zap.Int("open_items", len(items)),
			zap.Int("retracted", len(result.Resolved)),
			zap.Int("still_drifted", stillDrifted),
			zap.Int("page_not_observed", notObserved))
	}
	return nil
}

// loadOpenDriftItems reads this site's drift items and returns only those whose
// SHAPE says this check's predicate speaks for them.
//
// ── WHY A SHAPE TEST AND NOT A PRODUCER LIST (the GradesFunc principle) ──────
// item_type is not a predicate. Any agent definition can file any item_type
// from DB config with no code change, so a code-side list of producers is
// authoritative-looking and permanently behind live config (bugs_open/213 §5.3).
// The question that IS answerable from the row is "is this an item my predicate
// re-runs?" — and it is answered positively: the key must be this check's own,
// and both frozen lists must be present and well-formed, because `lost` is
// computed from them. Anything else is disclaimed and left open. A malformed
// spec therefore retracts nothing and errors nothing.
//
// ── IT CARRIES NO STATUS VOCABULARY, DELIBERATELY ───────────────────────────
// Same reasoning findResolvedEmptySections states at length: this package
// already holds two hand-rolled copies of the closed-status list and they
// already disagree. resolveWorkItems owns the predicate; this function owns the
// observation. Re-reading an already-closed row costs one no-op UPDATE — and
// the whole estate has filed six of these items, all-history.
func loadOpenDriftItems(dctx DiscoveryCheckContext) ([]driftItemSnapshot, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT id::text, item_key, COALESCE(created_at::text, ''), COALESCE(spec::text, '{}')
		  FROM site_work_items
		 WHERE site_id = $1 AND item_type = $2 AND item_key IS NOT NULL
		 ORDER BY created_at`, dctx.SiteID, sectionDriftItemType)
	if err != nil {
		return nil, fmt.Errorf("section_source_drift: load open items: %w", err)
	}
	defer rows.Close()

	var out []driftItemSnapshot
	for rows.Next() {
		var id, key, createdAt, specJSON string
		if sErr := rows.Scan(&id, &key, &createdAt, &specJSON); sErr != nil {
			return nil, fmt.Errorf("section_source_drift: scan open item: %w", sErr)
		}

		var sp struct {
			PageName         string    `json:"page_name"`
			AuthSource       string    `json:"authoritative_source"`
			Authoritative    *[]string `json:"authoritative"`
			PagesSections    *[]string `json:"pages_sections"`
			WouldDropPresent *[]string `json:"would_drop_present"`
		}
		if json.Unmarshal([]byte(specJSON), &sp) != nil {
			continue // disclaimed: not a shape this predicate speaks for
		}
		if sp.PageName == "" || sp.Authoritative == nil || sp.PagesSections == nil {
			continue
		}
		if key != sectionDriftItemKey(sp.PageName) {
			// The key does not follow this check's own contract, so the row was
			// filed by something else under a shared type. Not ours to close.
			continue
		}

		snap := driftItemSnapshot{
			ItemKey:       key,
			ItemID:        id,
			PageName:      sp.PageName,
			AuthSource:    sp.AuthSource,
			Authoritative: *sp.Authoritative,
			PagesSections: *sp.PagesSections,
			FiledAt:       createdAt,
		}
		if sp.WouldDropPresent != nil {
			snap.WouldDropPresent = *sp.WouldDropPresent
			snap.HasConsequence = true
		}
		out = append(out, snap)
	}
	return out, rows.Err()
}

// countArchivedSlotRows counts the artefact-archive rows that hold each lost
// section's destroyed bytes, so the receipt can say whether the content is
// recoverable rather than leaving a reader to guess.
//
// Best-effort: a failure yields an empty map, which downgrades the receipt's
// severity claim but never blocks the receipt. The receipt is the thing that
// must exist; the count is a convenience on it.
func countArchivedSlotRows(dctx DiscoveryCheckContext, pageID uuid.UUID, lost []string) map[string]int {
	out := map[string]int{}
	for _, name := range lost {
		out[name] = 0
	}
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT slot_name, count(*)
		  FROM page_component_history
		 WHERE page_id = $1
		   AND source = 'artefact_archive_trigger'
		   AND slot_name = ANY($2)
		 GROUP BY slot_name`, pageID, pqStringArray(lost))
	if err != nil {
		dctx.Logger.Warn("section_source_drift: could not count archived rows for the lost sections — receipt filed without them",
			zap.String("page_id", pageID.String()), zap.Error(err))
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var slot string
		var n int
		if rows.Scan(&slot, &n) == nil {
			out[slot] = n
		}
	}
	return out
}

// pqStringArray renders a Go slice as a Postgres text[] literal for `= ANY($n)`.
// The package has no lib/pq array helper imported and one string is cheaper than
// a new dependency on a best-effort path.
func pqStringArray(in []string) string {
	esc := make([]string, 0, len(in))
	for _, s := range in {
		esc = append(esc, `"`+strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)+`"`)
	}
	return "{" + strings.Join(esc, ",") + "}"
}

// ─── Loaders ────────────────────────────────────────────────────────────────

// loadPlanTableSections returns page_name -> ordered component list from the
// current plan's site_plan_sections table.
func loadPlanTableSections(ctx context.Context, q sectionSourceQuerier, siteID uuid.UUID) (map[string][]string, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT sps.page_name, sps.component_name
		FROM site_plan_sections sps
		JOIN site_plans sp ON sp.id = sps.plan_id
		WHERE sp.site_id = $1 AND sp.is_current = true
		ORDER BY sps.page_name, sps.ordering
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("section_source_drift: plan table query failed: %w", err)
	}
	defer rows.Close()

	out := map[string][]string{}
	for rows.Next() {
		var page, comp string
		if scanErr := rows.Scan(&page, &comp); scanErr != nil {
			return nil, fmt.Errorf("section_source_drift: plan table scan failed: %w", scanErr)
		}
		out[page] = append(out[page], comp)
	}
	return out, rows.Err()
}

// loadAspectSections returns page_name -> ordered section list from the current
// site_specs.site_plan aspect JSON (pages[].sections).
func loadAspectSections(ctx context.Context, q sectionSourceQuerier, siteID uuid.UUID) (map[string][]string, error) {
	var planJSON []byte
	err := q.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true
	`, siteID).Scan(&planJSON)
	if err != nil || planJSON == nil {
		// No aspect for this site (most sites) — not an error.
		return map[string][]string{}, nil
	}

	var plan struct {
		Pages []struct {
			Name     string   `json:"name"`
			Sections []string `json:"sections"`
		} `json:"pages"`
	}
	if json.Unmarshal(planJSON, &plan) != nil {
		return map[string][]string{}, nil
	}

	out := map[string][]string{}
	for _, p := range plan.Pages {
		if p.Name != "" && p.Sections != nil {
			out[p.Name] = p.Sections
		}
	}
	return out, nil
}

// loadPagesCacheSections returns page_name -> ordered section list from
// pages.sections (the materialised cache), for non-deleted pages, plus the page
// ids so a finding and its receipt can carry the first-class column rather than
// only a spec key.
func loadPagesCacheSections(ctx context.Context, q sectionSourceQuerier, siteID uuid.UUID, logger *zap.Logger) (map[string][]string, map[string]uuid.UUID, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, COALESCE(sections, '[]'::jsonb)::text
		FROM pages
		WHERE site_id = $1 AND COALESCE(status, '') <> 'deleted'
	`, siteID)
	if err != nil {
		return nil, nil, fmt.Errorf("section_source_drift: pages query failed: %w", err)
	}
	defer rows.Close()

	out := map[string][]string{}
	ids := map[string]uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		var name, sectionsJSON string
		if scanErr := rows.Scan(&id, &name, &sectionsJSON); scanErr != nil {
			logger.Warn("section_source_drift: pages scan failed", zap.Error(scanErr))
			continue
		}
		var sections []string
		if json.Unmarshal([]byte(sectionsJSON), &sections) != nil {
			continue
		}
		// Only pages that actually have a cached list can drift; an empty cache
		// with a plan entry is the sectionless case, owned elsewhere.
		if len(sections) > 0 {
			out[name] = sections
			ids[name] = id
		}
	}
	return out, ids, rows.Err()
}

// orderedListsEqual reports whether two section lists are identical in content
// and order. "sections" is a layout, so order is significant.
func orderedListsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
