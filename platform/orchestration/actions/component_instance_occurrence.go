package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// PER-INSTANCE ELEMENT-ID OCCURRENCE, for the two render paths that can only
// see ONE section at a time. bugs_closed/283 / RFC_032 step 3.
//
// WHAT THE PROBLEM WAS. A repeatable component's element ids are namespaced with
// {{.InstanceID}}, whose value is InstanceToken(function, occurrence) — the
// occurrence being how many sections with the SAME function precede this one on
// the page, in position order. That rule is canonical (InstanceCounter.Next,
// component_instance_scope.go) and any path holding the whole page applies it
// correctly. Two paths do not hold the whole page — RenderComponentAction (the
// build / content_rewrite route) and the section editor — and both supplied a
// constant occurrence 0 to every instance. So ANY per-section render re-collided
// every multi-instance page it touched: three of the twelve pages repaired on
// 2026-08-23 were re-collided within hours by an unrelated lane's backfill.
//
// WHAT CHANGED. Neither path invents a second rule; both now feed the SAME rule
// its real INPUT.
//
//   - The build path is always a loop iteration (measured 2026-08-24: exactly one
//     active agent definition anywhere in the fleet runs a render_component step,
//     page-content-writer, and both of its render steps sit inside
//     process_sections_loop). Loop expansion tells each injected step which
//     iteration it is and parks every item in CollectedData, so the occurrence is
//     counted from the items ALREADY RENDERED IN THIS PASS. That is the same
//     arithmetic InstanceCounter does, over the same list, one section at a time.
//     It is therefore correct on a page's FIRST build, when no page_components
//     rows exist yet to be counted — the case a stored-row lookup cannot reach.
//
//   - The editor path has no loop but does have the stored row it is editing,
//     with its page id and 1-based position, so it counts same-function
//     predecessors in the database, position-exact.
//
// WHY READING THE LOOP KEYS IS NOT REACHING INTO LOOP INTERNALS. loop_item_index
// and loop_name arrive on this step's OWN config, by the same channel as every
// other config key, and they are declared platform surface: datahelpers'
// frameworkStepConfigKeys lists them as "injected by loop expansion" and the
// config-key audit treats them as never-unknown. An action in this very package
// (LoopCompleteAction) has read loop_name and the item keys for months. The one
// genuinely fragile part — the item key's spelling — is now single-sourced as
// datahelpers.LoopItemKey, and a cross-package test drives the REAL expander and
// asserts this file's reader agrees with it.
//
// THE FALLBACK IS THE SAFETY CASE, AND IT IS UNIVERSAL. Every branch that cannot
// derive an occurrence binds occurrence 0 — today's behaviour exactly — and
// nothing here can fail a render. The derivation is an INPUT-IMPROVER, never a
// gate: a lookup error yields a warning and occurrence 0, never an error return.
// Against the constant-0 status quo the change is monotonically non-worse within
// the set of sections a path renders: under constant 0 every same-function pair
// already collides, so a derived assignment can only ever MERGE pairs that were
// already merged, never create a new collision. (The one cross-path exception is
// the errs-high editor case documented on storedPredecessorCount.)
//
// ⚠ THE OLD BINDER IS STILL PRESENT AND IS NO LONGER CALLED BY PRODUCTION CODE.
// BindSingleSectionInstanceToken (component_instance_scope.go) is retained
// deliberately, not by oversight: that file is being edited concurrently by the
// lane finishing RFC_032's Half B, and moving a function out from under it would
// mint exactly the same-file passenger that cost this lane a guardian veto on
// 2026-08-23. It is retired in a follow-up commit once that lane is clean. Until
// then, a census of "single-section binders" honestly returns two.

// SectionPlacement is what a single-section render path can know about WHICH
// placement it is rendering. The zero value means "nothing is known", which
// yields occurrence 0 — precisely the behaviour that shipped before this file.
type SectionPlacement struct {
	// --- build path (a page-content-writer loop iteration) ---

	// LoopKnown is true only when BOTH loop_item_index and loop_name resolved
	// from the step's own config. It is what distinguishes "this is iteration 0,
	// so there are genuinely no predecessors" from "there is no loop context, so
	// fall back" — two states that both count zero predecessors and must not be
	// conflated in the logs.
	LoopKnown bool
	// PriorFunctions holds the resolved component function of every loop item
	// BEFORE this one, in loop order.
	PriorFunctions []string

	// --- editor path (the stored row being edited) ---

	PageID   string
	Position int    // 1-based, as page_components.position is written (i+1)
	RowID    string // breaks a position tie under the canonical (position, id) walk
}

// PlacementFromLoopStep reads the loop-expansion contract off a render step's
// own config and the orchestration's collected data.
//
// The index is read with placementInt, not a type assertion: it round-trips
// through orchestration_states JSONB between plan time and execution, so it
// arrives as a float64 rather than the int that was written. A bare assertion
// would yield the zero value, which reads as "iteration 0, no predecessors" and
// would put every instance back on occurrence 0 — indistinguishable from the bug
// this file exists to fix.
func PlacementFromLoopStep(config, collected map[string]interface{}) SectionPlacement {
	loopName, _ := config["loop_name"].(string)
	idx := placementInt(config, "loop_item_index", -1)
	if loopName == "" || idx < 0 || collected == nil {
		return SectionPlacement{}
	}

	p := SectionPlacement{LoopKnown: true}
	for j := 0; j < idx; j++ {
		item, ok := collected[datahelpers.LoopItemKey(loopName, j)]
		if !ok {
			// A gap cannot be distinguished from an item that was never stored;
			// counting what IS present is the non-worse choice (see the header's
			// monotonicity note) and the caller logs the shortfall.
			continue
		}
		p.PriorFunctions = append(p.PriorFunctions, functionOfLoopItem(item))
	}
	return p
}

// functionOfLoopItem extracts a loop item's component function.
//
// sectionPlanItem carries the RESOLVED component's function at the top level
// (plan_sections_action.go sets Function from the component it resolved), and
// the nested component map carries it too. Both are read, top level first.
//
// ⚠ NEVER read "name" off the component map. That is the DISPLAY name — live
// data has component.name "Generic Text Block" for function "generic-text-block"
// — so a token derived from it would neither match the canonical walk nor be
// stable across components whose display name is title-cased.
func functionOfLoopItem(item interface{}) string {
	m, ok := item.(map[string]interface{})
	if !ok {
		return ""
	}
	if fn, ok := m["function"].(string); ok && fn != "" {
		return fn
	}
	if comp, ok := m["component"].(map[string]interface{}); ok {
		if fn, ok := comp["function"].(string); ok {
			return fn
		}
	}
	return ""
}

// PlacementFromStoredRow reads the identity of the stored row a section edit is
// about to rewrite, from the page_component map LoadEditContextAction builds.
//
// position goes through placementInt for the same reason as the loop index: the
// map is carried through workflow state as JSON between the load step and the
// edit step, so the int that was written is not the shape that arrives.
func PlacementFromStoredRow(pcData map[string]interface{}) SectionPlacement {
	if pcData == nil {
		return SectionPlacement{}
	}
	return SectionPlacement{
		PageID:   getStringVal(pcData, "page_id"),
		Position: placementInt(pcData, "position", 0),
		RowID:    getStringVal(pcData, "id"),
	}
}

// DeriveAndBindInstanceToken derives this placement's occurrence and binds the
// canonical per-instance token onto the render context.
//
// It never returns an error and never leaves the context unbound: every failure
// branch binds occurrence 0, which is what this call site did unconditionally
// before. A render must not fail because a count could not be taken.
//
// The name ends in "BindInstanceToken" on purpose: scripts/pattern-check.py's
// INSTANCE_BIND_SEAM_RE is what proves every RenderTemplate caller binds a token,
// and matching it keeps that guarantee true at these call sites. The regex also
// names this function explicitly, so the match does not depend on the suffix.
func DeriveAndBindInstanceToken(ctx context.Context, db *sql.DB, rc *RenderContext,
	function string, p SectionPlacement, logger *zap.Logger) {

	occ, source := 0, "none"

	switch {
	case p.LoopKnown:
		source = "loop"
		key := instanceFunctionKey(function)
		for _, fn := range p.PriorFunctions {
			if instanceFunctionKey(fn) == key {
				occ++
			}
		}

	case p.PageID != "" && p.Position > 0 && db != nil:
		source = "stored"
		n, err := storedPredecessorCount(ctx, db, function, p)
		if err != nil {
			// An input-improver that refuses is a gate. Warn and carry on.
			logf(logger).Warn("instance occurrence: stored lookup failed — binding occurrence 0 as before",
				zap.String("component_function", function),
				zap.String("page_id", p.PageID),
				zap.Error(err))
			source = "stored-failed"
		} else {
			occ = n
		}

	default:
		logf(logger).Debug("instance occurrence: no placement context — occurrence 0 as before",
			zap.String("component_function", function))
	}

	logf(logger).Debug("instance occurrence derived",
		zap.String("component_function", function),
		zap.String("source", source),
		zap.Int("occurrence", occ))

	BindInstanceToken(rc, InstanceToken(function, occ))
}

// instanceFunctionKey is InstanceCounter.Next's key equality, spelled once here
// so the derived occurrence and the canonical walk cannot disagree about which
// two functions are "the same" (case and surrounding whitespace both differ in
// live data).
func instanceFunctionKey(function string) string {
	return strings.ToLower(strings.TrimSpace(function))
}

// logf returns a usable logger even when the caller passed none, so a nil logger
// cannot turn a diagnostic into a panic on a live render path.
func logf(logger *zap.Logger) *zap.Logger {
	if logger == nil {
		return zap.NewNop()
	}
	return logger
}

// storedPredecessorCount counts the same-function sections stored BEFORE this
// placement, under the canonical walk's ordering.
//
// The predicates mirror loadStoredSections exactly — non-'removed' rows of this
// page, ordered by (position, id) — and the function comparison mirrors
// InstanceCounter.Next's lower/trim key. The (position, id) tie arm matches the
// ORDER BY tie-break added to loadStoredSections in the same change; without both
// halves a tie would be ordered arbitrarily by Postgres and the two derivations
// could disagree on a page that has one.
//
// ⚠ KNOWN AND DELIBERATE DIVERGENCE, in the errs-high direction. The canonical
// walk advances its counter only for sections that RESOLVED a component; a
// carried section (component missing, template invalid) does not advance it. This
// query cannot see template validity, so where an earlier same-function section
// is unresolvable the count comes out one HIGH. On the build path that is merely
// byte drift against a later full-page rerender — a distinct token, never a
// collision. On the editor path it can hand the edited row the token the next
// canonical instance already holds, i.e. swap which partner it collides with
// rather than adding a collision. It needs an invalid-template same-function
// predecessor to happen at all, it is no worse in collision COUNT than the
// constant 0 it replaces (which also collides with exactly one stored partner),
// and DetectInstanceCollisions reports either shape at assembly.
func storedPredecessorCount(ctx context.Context, db *sql.DB, function string, p SectionPlacement) (int, error) {
	const withTie = `
		SELECT count(*)
		FROM page_components pc
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE pc.page_id = $1
		  AND pc.build_status IS DISTINCT FROM 'removed'
		  AND lower(btrim(cc.function)) = lower(btrim($2))
		  AND (pc.position < $3 OR (pc.position = $3 AND pc.id < $4))`

	const withoutTie = `
		SELECT count(*)
		FROM page_components pc
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE pc.page_id = $1
		  AND pc.build_status IS DISTINCT FROM 'removed'
		  AND lower(btrim(cc.function)) = lower(btrim($2))
		  AND pc.position < $3`

	var n int
	var err error
	if p.RowID != "" {
		err = db.QueryRowContext(ctx, withTie, p.PageID, function, p.Position, p.RowID).Scan(&n)
	} else {
		err = db.QueryRowContext(ctx, withoutTie, p.PageID, function, p.Position).Scan(&n)
	}
	if err != nil {
		return 0, fmt.Errorf("counting same-function predecessors: %w", err)
	}
	return n, nil
}

// placementInt reads an integer that has travelled through workflow state.
//
// THREE SHAPES, and only the first is what was written. datahelpers.GetIntField
// covers int and float64 — float64 being what a JSON round-trip through
// orchestration_states produces — and this adds json.Number, which is what the
// decoders configured with UseNumber produce. All three must be handled at one
// place, because the failure is silent and identical in every case: an
// unrecognised shape yields the default, the default means "no context", and
// "no context" means occurrence 0 on every instance, which is indistinguishable
// from the defect this file exists to fix.
func placementInt(m map[string]interface{}, key string, defaultVal int) int {
	if m == nil {
		return defaultVal
	}
	if n, ok := jsonNumberInt(m[key]); ok {
		return n
	}
	return datahelpers.GetIntField(m, key, defaultVal)
}

// jsonNumberInt reads the json.Number shape that datahelpers.GetIntField does
// not cover. Split out so placementInt reads as the three-shape census it is.
func jsonNumberInt(v interface{}) (int, bool) {
	if n, ok := v.(json.Number); ok {
		if i, err := n.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}
