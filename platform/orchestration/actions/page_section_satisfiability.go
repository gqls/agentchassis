// FILE: platform/orchestration/actions/page_section_satisfiability.go
//
// One shared, read-only answer to the question every `needs_page` producer asks
// at emit time and the review queue's drain asks again later: WOULD
// PAGE-BUILD-HANDLER FIND ANYTHING TO BUILD FOR THIS PAGE?
//
// WHY IT IS SHARED (bugs_closed/177 → bugs_open/187). 177 fixed one emitter by
// mirroring the handler's own resolution chain read-only at emit time. 187
// measured the same shape under two more — image-build-handler's re-render emit
// and page-rerender's writer escalation — and found the items unsatisfiable at
// the moment they were minted: the handler resolves no sections, no-ops, and the
// WDS-004/149 routing parks the row in needs_human_review, where nothing drains
// it. The 177 close-out's architecture seat named the trigger for this file:
// "a THIRD copy of the satisfiability-mirror would be the moment to extract one
// shared resolver."
//
// WHAT IT MIRRORS. declaredPageSections walks
// load_page_sections_from_spec_action.go's fallbacks 1 to 3 —
// site_plan_sections for the CURRENT plan, then the site_specs `site_plan`
// aspect, then pages.sections — in that order, first non-empty source winning.
// The ORDER is the contract: an answer taken from a different source order is an
// answer to a different question. If the loader gains a source or reorders its
// own, this must follow, or the guard starts refusing items the handler could
// have built.
//
// WHAT IT DELIBERATELY DOES NOT RESOLVE. The loader's fallback 4 (same-role
// sibling layout synthesis) is not resolved here — only its GATE is.
// pageInCurrentPlan answers the membership question that licenses synthesis, and
// a plan member counts as SATISFIABLE even when it declares no sections at all,
// because the handler may synthesise a layout for it. Reproducing the synthesis
// itself would mean re-deriving the modal sibling layout and would still be a
// guess about what the handler will do minutes later; asking whether it is
// ALLOWED to run is cheap and cannot be wrong in the expensive direction. Every
// ambiguity, including a membership query that fails outright, therefore
// resolves in the direction of EMITTING.
//
// THIS FILE MUST NEVER WRITE. The loader syncs a served result back into
// pages.sections; nothing here does, and nothing here may. These functions run
// at emit time — judging a page whose shape is still being decided — and at
// revalidation time, judging whether a parked item may be closed. A resolver
// that writes to the table it is judging changes the answer for everything
// downstream of it, and a revalidator that writes has stopped being a read.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// declaredPageSections reads the sections a page declares, in the order
// page-build-handler will read them, and returns the first non-empty source
// along with the name the loader gives it ("site_plan_tables", "site_specs",
// "pages_table", "none").
//
// Every failure returns what has been resolved so far rather than an error: a
// source that cannot be read is a source that declares nothing, and the caller's
// only decision is whether the handler would have anything to build.
func declaredPageSections(ctx context.Context, db *sql.DB, logger *zap.Logger, siteID uuid.UUID, pageName string) ([]string, string) {
	var sections []string

	planRows, err := db.QueryContext(ctx, `
		SELECT sps.component_name
		FROM site_plan_sections sps
		JOIN site_plans sp ON sp.id = sps.plan_id
		WHERE sp.site_id = $1 AND sp.is_current = true AND sps.page_name = $2
		ORDER BY sps.ordering
	`, siteID, pageName)
	if err != nil {
		logger.Warn("declaredPageSections: site_plan_sections lookup failed", zap.Error(err))
	} else {
		for planRows.Next() {
			var component string
			if scanErr := planRows.Scan(&component); scanErr != nil {
				logger.Warn("declaredPageSections: site_plan_sections scan failed", zap.Error(scanErr))
				continue
			}
			if component != "" {
				sections = append(sections, component)
			}
		}
		planRows.Close()
	}
	if len(sections) > 0 {
		return sections, "site_plan_tables"
	}

	// The older planner generation's store. Five sites still carry a current
	// site_plan aspect and are served here, their table lookup above simply
	// missing.
	var planDataJSON []byte
	if err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true
	`, siteID).Scan(&planDataJSON); err == nil && planDataJSON != nil {
		var planData map[string]interface{}
		if json.Unmarshal(planDataJSON, &planData) == nil {
			if pages, ok := planData["pages"].([]interface{}); ok {
				for _, pageRaw := range pages {
					page, ok := pageRaw.(map[string]interface{})
					if !ok {
						continue
					}
					if name, _ := page["name"].(string); name == pageName {
						if declared, ok := page["sections"].([]interface{}); ok {
							for _, s := range declared {
								if sName, ok := s.(string); ok && sName != "" {
									sections = append(sections, sName)
								}
							}
						}
						break
					}
				}
			}
		}
	}
	if len(sections) > 0 {
		return sections, "site_specs"
	}

	var sectionsJSON []byte
	if err := db.QueryRowContext(ctx, `
		SELECT sections FROM pages
		WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&sectionsJSON); err == nil && sectionsJSON != nil {
		if unmarshalErr := json.Unmarshal(sectionsJSON, &sections); unmarshalErr != nil {
			logger.Warn("declaredPageSections: pages.sections is not an array of strings",
				zap.Error(unmarshalErr))
			sections = nil
		}
	}
	if len(sections) > 0 {
		return sections, "pages_table"
	}

	return nil, "none"
}

// pageInCurrentPlan reports whether the page is named in the CURRENT plan's
// site_plan_pages — the membership the loader's fallback 4 requires before it
// will synthesise a layout from a same-role sibling. It is the same join the
// loader's `cur`/`target` CTEs make, reduced to the existence question.
//
// FAIL-OPEN BY DESIGN: a query that errors returns true. The only caller is a
// guard deciding whether to skip an emit, and the two ways of being wrong are
// not symmetric — refusing an item the handler could have built loses work
// silently, while raising one it cannot build costs a row in a queue that a
// revalidator now drains. A read that failed is not evidence of absence.
func pageInCurrentPlan(ctx context.Context, db *sql.DB, logger *zap.Logger, siteID uuid.UUID, pageName string) bool {
	var member bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM site_plan_pages spp
			JOIN site_plans sp ON sp.id = spp.plan_id
			WHERE sp.site_id = $1 AND sp.is_current = true AND spp.name = $2
		)
	`, siteID, pageName).Scan(&member)
	if err != nil {
		logger.Warn("pageInCurrentPlan: current-plan membership lookup failed — treating the page as a plan member so the emit is not silently refused",
			zap.String("page", pageName), zap.Error(err))
		return true
	}
	return member
}

// pageSectionsSatisfiable is the emit-time guard both needs_page emitters use.
// It reports whether page-build-handler would have something to build, the
// sections it would see, and the source they came from ("current_plan_member"
// when nothing is declared but synthesis is licensed, "none" when the answer is
// no).
//
// One helper rather than the clause written out twice: the two emitters must ask
// the identical question, and a copy is how the second one drifts.
func pageSectionsSatisfiable(ctx context.Context, db *sql.DB, logger *zap.Logger, siteID uuid.UUID, pageName string) (bool, []string, string) {
	declared, source := declaredPageSections(ctx, db, logger, siteID, pageName)
	if len(declared) > 0 {
		return true, declared, source
	}
	if pageInCurrentPlan(ctx, db, logger, siteID, pageName) {
		return true, nil, "current_plan_member"
	}
	return false, nil, source
}

// revalidateNeedsPage is the review queue's drain for `needs_page`
// (bugs_open/187; registered in reviewRevalidators). The ask an item of this
// type carries is "build this page", so the finding no longer holds only when
// the page is THERE — active, declaring sections, and with a component slot
// built for every section it declares.
//
// SLOT MATCHING, MEASURED 2026-08-03 (live clients_db, read-only). Match is on
// page_components.slot_name, normalised through NormalizeComponentFunction on
// both sides:
//
//   - Over the 1,174 sections declared by active pages that have any components
//     at all, normalised slot_name matches 1,123 (95.7%); content_components
//     .function matches only 1,027 (87.5%), because a page may mount the same
//     component twice under distinct slots — thames-water carries slot_names
//     `evidence-chart-ofwat` and `evidence-timeseries-leakage` over functions
//     `evidence-chart` and `evidence-timeseries`, and matching on function would
//     credit one declaration twice and leave one uncredited.
//   - All 12 slots of the four pages named in 187's triage (tungsten-guide,
//     board-setup, cases-index, thames-water) match their declaration exactly on
//     slot_name.
//   - Normalisation earns its place: 31 live sections entries are written in the
//     underscore dialect (`call_to_action`) while no slot_name is, so raw
//     equality matches 1,106 and normalised matches 1,123.
//   - slot_name is never NULL or blank in any of the 1,200 live rows, so there is
//     no positional-fallback arm to write here (loadPageComponentBySlotRO's
//     exists for a shape this data no longer has).
//
// The residual 4.3% is why the match is exact-name and not a count: a page whose
// declarations and slots merely AGREE IN NUMBER is not evidence that the
// declared sections were built, and this verdict closes items. Mismatches stay
// still_holds, which is the safe direction — a wrong still_holds costs a human
// glance, a wrong resolved closes a live ask.
func revalidateNeedsPage(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict {
	// spec.page_name only. The item_key is NOT read: its prefix differs per
	// producer (`needs_page:<name>` from reconcile_site_plan and page-rerender,
	// `page_rerender:<name>` from image-build-handler), so parsing it would be
	// guessing at the very field the verdict turns on.
	pageName := specString(item.Spec, "page_name")
	if pageName == "" {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  "spec names no page_name, so the ask cannot be located; the item_key is deliberately not parsed for it because its prefix differs per producer",
		}
	}

	var pageID, pageStatus string
	err := db.QueryRowContext(ctx, `
		SELECT p.id::text, COALESCE(p.status, '')
		FROM pages p
		WHERE p.site_id = $1 AND p.name = $2
	`, item.SiteID, pageName).Scan(&pageID, &pageStatus)
	switch {
	case err == sql.ErrNoRows:
		// The page the item asked for does not exist. That is the one arm where
		// "still true" is provable rather than merely unrefuted.
		return revalidationVerdict{
			Verdict: revalidationStillHolds,
			Reason:  fmt.Sprintf("no page named %q exists on this site, so the ask — build this page — is not satisfied", pageName),
			Evidence: map[string]interface{}{
				"page_name": pageName,
				"page_row":  "absent",
			},
		}
	case err != nil:
		logger.Warn("revalidateNeedsPage: page lookup failed",
			zap.String("item_id", item.ID), zap.String("page", pageName), zap.Error(err))
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("page lookup failed for %q", pageName),
		}
	}

	if pageStatus == "archived" {
		// Closing here would assert the ask was satisfied; it was abandoned, which
		// is a different fact and not one this sweep is allowed to record.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("page %q is archived — needs a human call", pageName),
			Evidence: map[string]interface{}{
				"page_name":   pageName,
				"page_status": pageStatus,
			},
		}
	}
	if pageStatus != "active" {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("page %q carries status %q, which this revalidator has no reading for", pageName, pageStatus),
			Evidence: map[string]interface{}{
				"page_name":   pageName,
				"page_status": pageStatus,
			},
		}
	}

	declared, sectionSource := declaredPageSections(ctx, db, logger, item.SiteID, pageName)
	if len(declared) == 0 {
		// 187's own population: the page exists but resolves nothing, so the
		// handler would no-op it again. Not evidence either way, and NOT a close.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("page %q resolves no sections from any source, so there is no positive evidence the ask was satisfied", pageName),
			Evidence: map[string]interface{}{
				"page_name":       pageName,
				"page_status":     pageStatus,
				"sections_source": sectionSource,
			},
		}
	}

	slots, err := builtSlotNames(ctx, db, pageID)
	if err != nil {
		logger.Warn("revalidateNeedsPage: page_components lookup failed",
			zap.String("item_id", item.ID), zap.String("page", pageName), zap.Error(err))
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("component lookup failed for %q", pageName),
		}
	}

	built := make(map[string]string, len(slots))
	for _, slot := range slots {
		built[NormalizeComponentFunction(slot)] = slot
	}

	matched := make([]string, 0, len(declared))
	var missing []string
	for _, section := range declared {
		if slot, ok := built[NormalizeComponentFunction(section)]; ok {
			matched = append(matched, slot)
			continue
		}
		missing = append(missing, section)
	}

	evidence := map[string]interface{}{
		"page_name":         pageName,
		"page_status":       pageStatus,
		"sections_source":   sectionSource,
		"declared_sections": declared,
		"matched_slots":     matched,
		"slot_count":        len(slots),
	}
	if len(missing) > 0 {
		evidence["missing_sections"] = missing
		// "satisfiable now" is load-bearing prose, not decoration: these are the
		// rows a human working the queue should act on, and it is the only signal
		// separating them from the section-less rows above (187, out of scope).
		return revalidationVerdict{
			Verdict: revalidationStillHolds,
			Reason: fmt.Sprintf("satisfiable now — page %q declares %d section(s) from %s but %d have no built component slot: %s",
				pageName, len(declared), sectionSource, len(missing), strings.Join(missing, ", ")),
			Evidence: evidence,
		}
	}

	return revalidationVerdict{
		Verdict: revalidationResolved,
		Reason: fmt.Sprintf("page %q is active and every section it declares (%s) has a built component slot",
			pageName, strings.Join(declared, ", ")),
		Evidence: evidence,
	}
}

// builtSlotNames lists the slots actually mounted on a page, read-only.
func builtSlotNames(ctx context.Context, db *sql.DB, pageID string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(pc.slot_name, '')
		FROM page_components pc
		WHERE pc.page_id = $1::uuid
		ORDER BY pc.position
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []string
	for rows.Next() {
		var slot string
		if scanErr := rows.Scan(&slot); scanErr != nil {
			return nil, scanErr
		}
		if slot != "" {
			slots = append(slots, slot)
		}
	}
	return slots, rows.Err()
}
