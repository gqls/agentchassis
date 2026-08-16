// FILE: platform/orchestration/datahelpers/locked_page_sections.go
//
// The page-component lock guard protects the ROW; this file lets the LIST
// honour the same lock (bugs_open/285, register LOCK-008).
//
// THE DEFECT THIS CLOSES. A page's section list is assembled from the site
// plan (load_page_sections_from_spec: site_plan_sections → site_specs aspect →
// pages.sections → sibling synthesis). None of those stores knows about a
// section a human pinned onto the live page with a lock — that fact lives on
// page_components. So every rebuild proposed a list WITHOUT the locked
// section, save_page_sections' 058 guard kept the row and filed a
// `lock_blocked_change … blocked_action=remove` item, and pages.sections went
// on telling every reader the section did not exist. Measured 2026-08-15: 13
// tier-1 pages fleet-wide in that state (webdesign.uk/contact's chat box, 12
// loancalculator.co.uk calculators), 5 fresh remove-blocked items that day.
//
// WHAT LIVES HERE AND WHY HERE. Two things, both shared:
//   - LoadLockedPageSlots*: "which rows on this page may automation NOT
//     rewrite?" — answered with THE GUARD'S OWN predicate, AgentWritableSQLFor
//     (chrome_render_inputs.go), never a re-typed copy. The 058 guard, the
//     save-side floors and this loader must classify a row identically or a
//     lock survives the DELETE and is still missing from the list (or the
//     reverse). LOCK-007's rule: the read predicate is DERIVED from the write
//     predicate.
//   - MergeLockedPageSlots: the pure merge — pair each list entry with at most
//     one locked row (mirroring save_page_sections' matchLockedRow arms, so
//     the loader and the guard agree on what "already in the list" means), then
//     insert every unpaired locked row at its live position.
// They live in datahelpers because two packages need them: the loader
// (actions) and check_section_source_drift (actions/discovery_checks), which
// compares plan-vs-cache and must see the SAME merged shape or it flags every
// fixed page as drift. actions imports discovery_checks imports datahelpers.
//
// MEMBERSHIP vs LOCK. The lock predicate is verbatim. `build_status <>
// 'removed'` is a separate MEMBERSHIP condition — a removed row is not on the
// page, so it does not belong in the list of what the page has (0 locked rows
// were removed when this shipped; the guard's own loader has no such filter
// and would still retain such a row through a DELETE, which is fine — this is
// about the LIST, not row survival).
//
// Consumers, told not merely measured (owner ruling 2026-07-29 §3):
// load_page_sections_from_spec (the merge), check_section_source_drift (both
// sides merged), save_sections_prune_floor (reads pages.sections length minus
// locked rows — becomes exact rather than under-counting). Readers that
// re-implement the tier order and stay lock-blind, by design and named:
// page_section_satisfiability, ensure_page_section_layout, plan_section_counts,
// tool_content_item, flag_page_image_rebuild, check_sectionless_pages,
// check_literal_markdown.

package datahelpers

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// LockedPageSlot is one page_components row automation may not rewrite,
// with the identities the merge pairs on and the position it is inserted at.
type LockedPageSlot struct {
	RowID             string
	PageName          string
	Slot              string // page_components.slot_name ('' if NULL)
	Position          int    // live position (1-based, as save_page_sections writes it)
	ComponentID       string
	ComponentFunction string
	ComponentName     string
	LockType          string
	LockedBy          string
}

// LockedPageSlotsSQL is the one query behind LoadLockedPageSlots and
// LoadLockedPageSlotsForSite ($2 = ” means every page on the site). It is
// exported so a test can pin that the lock arm IS AgentWritableSQLFor("pc.")
// and not a re-typed lookalike (LANDMINES: "the PIN predicate is not the POOL
// predicate").
var LockedPageSlotsSQL = `
	SELECT pc.id::text, p.name, COALESCE(pc.slot_name, ''), pc.position,
	       COALESCE(pc.component_id::text, ''), COALESCE(cc.function, ''), COALESCE(cc.name, ''),
	       COALESCE(pc.lock_type, ''), COALESCE(pc.locked_by, '')
	FROM page_components pc
	JOIN pages p ON p.id = pc.page_id
	LEFT JOIN content_components cc ON cc.id = pc.component_id
	WHERE p.site_id = $1
	  AND ($2 = '' OR p.name = $2)
	  AND COALESCE(pc.build_status, '') <> 'removed'
	  AND NOT ` + AgentWritableSQLFor("pc.") + `
	ORDER BY p.name ASC, pc.position ASC`

// LoadLockedPageSlots returns the named page's non-agent-writable rows in
// position order. No rows is normal (most pages carry no lock). A query error
// is returned to the caller, who decides between best-effort and loud.
func LoadLockedPageSlots(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string) ([]LockedPageSlot, error) {
	if pageName == "" {
		return nil, fmt.Errorf("LoadLockedPageSlots: page name required (use LoadLockedPageSlotsForSite for a whole site)")
	}
	bySite, err := loadLockedPageSlots(ctx, db, siteID, pageName)
	if err != nil {
		return nil, err
	}
	return bySite[pageName], nil
}

// LoadLockedPageSlotsForSite returns every page's non-agent-writable rows on
// the site, keyed by page name — for site-wide readers such as
// check_section_source_drift, which would otherwise issue one query per page.
func LoadLockedPageSlotsForSite(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string][]LockedPageSlot, error) {
	return loadLockedPageSlots(ctx, db, siteID, "")
}

func loadLockedPageSlots(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string) (map[string][]LockedPageSlot, error) {
	if db == nil {
		return nil, fmt.Errorf("LoadLockedPageSlots: nil db")
	}
	rows, err := db.QueryContext(ctx, LockedPageSlotsSQL, siteID, pageName)
	if err != nil {
		return nil, fmt.Errorf("load locked page slots: %w", err)
	}
	defer rows.Close()

	out := map[string][]LockedPageSlot{}
	for rows.Next() {
		var s LockedPageSlot
		if err := rows.Scan(&s.RowID, &s.PageName, &s.Slot, &s.Position, &s.ComponentID,
			&s.ComponentFunction, &s.ComponentName, &s.LockType, &s.LockedBy); err != nil {
			return nil, fmt.Errorf("scan locked page slot: %w", err)
		}
		out[s.PageName] = append(out[s.PageName], s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locked page slots: %w", err)
	}
	return out, nil
}

// MergedName is the list entry a locked row contributes: its own slot_name
// (the identity save_page_sections' guard matches on), falling back to the
// component function when the row was inserted without one.
func (s LockedPageSlot) MergedName() string {
	if s.Slot != "" {
		return s.Slot
	}
	return s.ComponentFunction
}

// MergeLockedPageSlots returns `list` with every locked row that is not
// already represented in it inserted at the row's live position.
//
// PAIRING mirrors save_page_sections' matchLockedRow, arm for arm, so the two
// judgements of "is this locked row already in the proposal" cannot disagree
// (the disagreement is exactly how bugs_open/189 duplicated a locked
// calculator): for each list entry, in order — (1) exact slot_name over all
// unconsumed rows, (2) kebab-normalised slot_name, (3) the row's component
// function or name (the list carries names not ids, so this stands in for the
// guard's identity arm; identity resolution happens later, in plan_sections
// and enrichSectionsWithComponentIDs). Each row pairs with at most ONE entry,
// so two locked rows rendering the same component (generic-text-block twice)
// against a plan naming it twice pair one-to-one and nothing is inserted;
// against a plan naming it once, the second row is inserted.
//
// INSERTION uses the row's live `position` (1-based): index position-1,
// clamped to the current list length; ascending, so a later row indexes
// against the list its predecessors have already grown. A row an earlier
// guard pass exiled to the tail therefore stays at the tail — this restores
// membership, not history; moving it back is a human edit.
//
// Pure: no I/O, no logging. Returns the merged list, the rows inserted, and
// the index each landed at (aligned with `inserted`), so a caller keeping a
// parallel per-index slice (section_facts) can insert placeholders at the same
// spots.
func MergeLockedPageSlots(list []string, locked []LockedPageSlot) (merged []string, inserted []LockedPageSlot, insertedAt []int) {
	merged = make([]string, len(list))
	copy(merged, list)
	if len(locked) == 0 {
		return merged, nil, nil
	}

	consumed := make([]bool, len(locked))
	pair := func(name string) {
		if name == "" {
			return
		}
		for i, lr := range locked { // arm 1: slot exact
			if !consumed[i] && lr.Slot != "" && lr.Slot == name {
				consumed[i] = true
				return
			}
		}
		if norm := NormalizeComponentFunction(name); norm != "" { // arm 2: slot kebab
			for i, lr := range locked {
				if !consumed[i] && lr.Slot != "" && NormalizeComponentFunction(lr.Slot) == norm {
					consumed[i] = true
					return
				}
			}
		}
		for i, lr := range locked { // arm 3: component function / name
			if !consumed[i] && ((lr.ComponentFunction != "" && lr.ComponentFunction == name) ||
				(lr.ComponentName != "" && lr.ComponentName == name)) {
				consumed[i] = true
				return
			}
		}
	}
	for _, name := range list {
		pair(name)
	}

	// Unpaired rows, ascending by position (the loader's SQL orders them so;
	// re-sort defensively — a stable insertion sort keeps this dependency-free).
	var pending []LockedPageSlot
	for i, lr := range locked {
		if !consumed[i] && lr.MergedName() != "" {
			pending = append(pending, lr)
		}
	}
	for i := 1; i < len(pending); i++ {
		for j := i; j > 0 && pending[j].Position < pending[j-1].Position; j-- {
			pending[j], pending[j-1] = pending[j-1], pending[j]
		}
	}
	for _, lr := range pending {
		at := lr.Position - 1
		if at < 0 {
			at = 0
		}
		if at > len(merged) {
			at = len(merged)
		}
		merged = append(merged, "")
		copy(merged[at+1:], merged[at:])
		merged[at] = lr.MergedName()
		inserted = append(inserted, lr)
		insertedAt = append(insertedAt, at)
	}
	return merged, inserted, insertedAt
}

// ---------------------------------------------------------------------------
// Kebab-case normalisation — moved down from actions/component_validation.go
// so the merge above (datahelpers) and the guard's matchLockedRow (actions)
// normalise with ONE function. actions.NormalizeComponentFunction delegates
// here; behaviour unchanged.
// ---------------------------------------------------------------------------

var kebabCaseRe = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// IsKebabCase reports whether s already satisfies the component-function
// naming contract.
func IsKebabCase(s string) bool { return kebabCaseRe.MatchString(s) }

// NormalizeComponentFunction converts a function name to kebab-case.
//
//	"social_proof"    → "social-proof"
//	"call_to_action"  → "call-to-action"
//	"SocialProof"     → "social-proof"
//	"social-proof"    → "social-proof" (no-op)
//	""                → ""             (no-op)
func NormalizeComponentFunction(function string) string {
	if function == "" {
		return ""
	}
	if kebabCaseRe.MatchString(function) {
		return function
	}
	result := strings.ReplaceAll(function, "_", "-")
	var b strings.Builder
	for i, r := range result {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r + 32) // to lowercase
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
