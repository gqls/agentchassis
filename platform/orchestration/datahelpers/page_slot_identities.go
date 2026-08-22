// FILE: platform/orchestration/datahelpers/page_slot_identities.go
//
// ONE answer to "what is this page actually made of, as the page itself records
// it?" — the stored slot identities on `page_components` (bugs_closed/204).
//
// THE DEFECT THIS SERVES. On a decomposed site a page's composition is a list of
// POSITIONAL slot names — `prose-0`, `tool-1` — which are neither a
// `content_components.name` nor a `.function`. The component each slot really is
// lives on the row, in `page_components.component_id`. Every surface that maps a
// section name to a component by consulting the component catalogue alone is
// therefore blind on those sites, and the estate has now fixed that blindness
// three times, separately, in three private loaders:
//   - `bugs_closed/182` — the re-render path (`rerender_page_sections`);
//   - commit `13252f714` — the build path (`plan_sections`), whose loader this
//     file now IS (moved here verbatim in behaviour, prefix-parameterised);
//   - the site-plan validation path — the third, and the one that DELETES rather
//     than defers, which is what made a shared home worth the move.
// A fourth private copy is how the third one came to exist. LOCK-008 is the
// precedent for the shape and for this file's location.
//
// WHAT LIVES HERE AND WHY HERE. The loader plus two PURE derivations, because the
// two questions callers ask are genuinely different judgements:
//   - `SlotIDMap` — "which component is at this slot?" It must pick one id, so it
//     needs a conflict rule: a slot_name repeated with DIFFERENT component_ids is
//     dropped from the map (an ambiguous carry source is no carry source).
//   - `SlotNameSet` — "does this page carry a slot under this name at all?" It
//     picks nothing, so a repeat is not a conflict and no rule is needed. Keeping
//     it separate is deliberate: reusing the id map for membership would silently
//     inherit the conflict rule and answer "no" for a page that plainly does
//     carry the slot, which is the destructive direction.
// They are separated from the loader so a caller can test the judgement without a
// database, and so `SlotNameSet` cannot acquire the id map's rule by accident.
//
// THE PREDICATE IS DELIBERATELY NARROW, and this is a decision, not an omission.
// There is no `build_status <> 'removed'` membership filter, unlike
// LockedPageSlotsSQL's. Adding one here would change `plan_sections`' behaviour in
// what is meant to be a pure extraction, and for the membership caller a removed
// row's slot name is still a name `pages.sections` lists today — keeping it
// preserves the status quo rather than trimming the page's record as a side
// effect of a refactor. Revisit it as its own change, with its own measurement.
//
// Consumers, told not merely measured (owner ruling 2026-07-29 §3):
//   - `plan_sections` (`plan_sections_action.go`) — `SlotIDMap`, to resolve a
//     positional slot to the component it stores (Path 0).
//   - `ValidateSitePlanAction`'s `validate_components` arm (`v3_site_actions.go`)
//     — `SlotNameSet`, to keep a proposed name the page already carries instead of
//     deleting it.
//   - `applyAddToPage` / `applyRetypeExisting` (`apply_gap_plan_action.go`) — the
//     same membership judgement, per page.
//   - `applyNewPage` deliberately does NOT consume this: a genuinely new page has
//     no stored rows, so a positional name proposed for it points at nothing and
//     dropping it is correct. Named here so the omission reads as a decision.
//
// The log and error strings are load-bearing beyond this file: `bugs_closed/204`'s
// closure evidence pod-greps "load page slot identities" and "slot_name repeats
// with different component_ids" to prove a binary carries the fix. Changing them
// silently retires somebody's verification.

package datahelpers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PageSlotRow is one `page_components` row reduced to the identities these
// judgements need: which page, which slot name, which component.
type PageSlotRow struct {
	PageName    string // pages.name
	Slot        string // page_components.slot_name ('' if NULL)
	ComponentID string // page_components.component_id ('' if NULL)
}

// pageSlotPredicate is the FROM/JOIN/WHERE the two projections share. Single
// sourced so the two reads can never disagree about WHICH rows a page's stored
// identity is made of — the drift class this file exists to end. $2 = '' means
// every page on the site.
const pageSlotPredicate = `
	FROM page_components pc
	JOIN pages p ON p.id = pc.page_id
	WHERE p.site_id = $1
	  AND ($2 = '' OR p.name = $2)`

// PageSlotIdentitiesSQL reads ONE page's slot identities. Its projection is
// byte-identical to the query commit 13252f714 shipped in plan_sections — the
// move to this package deliberately changed nothing a caller or a test mock can
// observe. Exported so a test can pin exactly that, which is the same reason
// LockedPageSlotsSQL is exported.
var PageSlotIdentitiesSQL = `
	SELECT COALESCE(pc.slot_name, ''), COALESCE(pc.component_id::text, '')` +
	pageSlotPredicate + `
	ORDER BY pc.position ASC`

// PageSlotIdentitiesForSiteSQL reads EVERY page's slot identities in one pass,
// carrying the page name so the rows can be grouped. Same predicate, wider
// projection: validate_plan judges a whole plan at once and would otherwise issue
// one query per page (the reason LoadLockedPageSlotsForSite exists too).
var PageSlotIdentitiesForSiteSQL = `
	SELECT p.name, COALESCE(pc.slot_name, ''), COALESCE(pc.component_id::text, '')` +
	pageSlotPredicate + `
	ORDER BY p.name ASC, pc.position ASC`

// LoadPageSlotRows returns the named page's stored slot identities in position
// order. PageName is filled in from the argument rather than the row, because the
// per-page projection does not select it.
//
// No rows is NORMAL — an initial build, where the page or its components do not
// exist yet — and returns an empty slice, not an error.
//
// A query error IS returned. Planning against a silently-empty map on a
// decomposed site files junk `needs_new_component` items (two per section,
// measured on the 204 canary), so a loud transient failure is the cheaper
// outcome; a caller that would rather degrade than fail says so at its own site.
func LoadPageSlotRows(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string) ([]PageSlotRow, error) {
	// Argument before dependency: an empty page name on THIS projection would
	// silently read the whole site and label every row "", which is worse than a
	// nil db (a nil db fails; a whole-site read succeeds and lies).
	if pageName == "" {
		return nil, fmt.Errorf("load page slot identities: page name required (use LoadPageSlotRowsForSite for a whole site)")
	}
	if db == nil {
		return nil, fmt.Errorf("load page slot identities: nil db")
	}
	rows, err := db.QueryContext(ctx, PageSlotIdentitiesSQL, siteID, pageName)
	if err != nil {
		return nil, fmt.Errorf("load page slot identities: %w", err)
	}
	defer rows.Close()

	var out []PageSlotRow
	for rows.Next() {
		r := PageSlotRow{PageName: pageName}
		if err := rows.Scan(&r.Slot, &r.ComponentID); err != nil {
			return nil, fmt.Errorf("scan page slot identity: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate page slot identities: %w", err)
	}
	return out, nil
}

// LoadPageSlotRowsForSite returns every page's stored slot identities across the
// whole site, for a caller judging many pages in one pass.
func LoadPageSlotRowsForSite(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]PageSlotRow, error) {
	if db == nil {
		return nil, fmt.Errorf("load page slot identities: nil db")
	}
	rows, err := db.QueryContext(ctx, PageSlotIdentitiesForSiteSQL, siteID, "")
	if err != nil {
		return nil, fmt.Errorf("load page slot identities: %w", err)
	}
	defer rows.Close()

	var out []PageSlotRow
	for rows.Next() {
		var r PageSlotRow
		if err := rows.Scan(&r.PageName, &r.Slot, &r.ComponentID); err != nil {
			return nil, fmt.Errorf("scan page slot identity: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate page slot identities: %w", err)
	}
	return out, nil
}

// SlotIDMap reduces one page's rows to slot_name → component_id.
//
// A slot_name repeated across rows is NORMAL (generic-text-block used 2-3× on one
// page — measured fleet-wide, 11 legitimate pages; see LANDMINES.md
// "Deduplicating page_components…"). Repeats AGREEING on component_id map fine;
// repeats DISAGREEING drop that slot from the map with a warning, so resolution
// falls back to the name path rather than picking a row arbitrarily.
//
// `caller` prefixes the warning so the message a service emits still names the
// action a reader is debugging. Rows carrying no slot name or no component id are
// skipped: neither half is usable as an identity on its own.
func SlotIDMap(rows []PageSlotRow, caller, pageName string, logger *zap.Logger) map[string]string {
	slotIDs := make(map[string]string)
	conflicted := make(map[string]bool)
	for _, r := range rows {
		if r.Slot == "" || r.ComponentID == "" {
			continue
		}
		if existing, ok := slotIDs[r.Slot]; ok && existing != r.ComponentID {
			if !conflicted[r.Slot] {
				conflicted[r.Slot] = true
				if logger != nil {
					logger.Warn(caller+": slot_name repeats with different component_ids — leaving it to name/function resolution",
						zap.String("page", pageName),
						zap.String("slot", r.Slot))
				}
			}
			continue
		}
		slotIDs[r.Slot] = r.ComponentID
	}
	for slot := range conflicted {
		delete(slotIDs, slot)
	}
	return slotIDs
}

// SlotNameSet reduces rows to page name → the set of slot names that page carries.
//
// Membership only. It answers "does this page already have a slot called X?", so
// a repeated slot_name is not a conflict and is NOT dropped — deliberately unlike
// SlotIDMap, because a caller asking membership is deciding whether to KEEP a name
// the page demonstrably has, and answering "no" on a legitimate repeat would
// delete a real section. A row with no slot name contributes nothing; a row whose
// component_id is NULL still contributes, because the page carries the slot
// whether or not the link is intact.
func SlotNameSet(rows []PageSlotRow) map[string]map[string]bool {
	out := make(map[string]map[string]bool)
	for _, r := range rows {
		if r.Slot == "" {
			continue
		}
		if out[r.PageName] == nil {
			out[r.PageName] = make(map[string]bool)
		}
		out[r.PageName][r.Slot] = true
	}
	return out
}
