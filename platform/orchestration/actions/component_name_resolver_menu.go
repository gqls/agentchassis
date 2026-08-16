package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// ============================================================================
// The ACCEPT half of a component-menu widening (bugs_open/282)
// ============================================================================
//
// A planner step is SHOWN a menu of components (a `query_database` step whose
// output_field the LLM prompt renders — e.g. build-site-planner's
// `load_components` -> `available_components`). validate_site_plan then RESOLVES
// each proposed section name against reality and DROPS what it cannot resolve
// (`loadComponentNameResolver`, v3_site_actions.go).
//
// Those are two halves of one contract — what may be OFFERED and what will be
// ACCEPTED — and until this file they were maintained in two places, in two
// languages: the offer as a SQL string in agent_definitions, the acceptance as
// a Go query with `component_level IN ('section','element')` hardcoded. Migration
// 407 widened the offer (tool-level components, per-site opt-in) and nothing
// widened the acceptance, so the planner placed twelve calculators and validate
// ate every one of them, silently, with only a Warn per drop. That is
// bugs_open/282.
//
// THE FIX IS NOT A SECOND COPY OF THE PREDICATE. Mirroring 407's SQL into Go
// would have been a third hand-maintained copy of a string that had ALREADY
// drifted — migration 419 added a `requires-backend` gate to the same query,
// and 419 guards its own apply by asserting 407's exact text. Instead the
// acceptance surface consumes the offer's OUTPUT: the very rows the planner was
// shown, already in CollectedData at validate time. One list, one source. Any
// future gate on the menu flows through automatically, because there is nothing
// to keep in step.
//
// UNION, NOT INTERSECTION. addMenu ADDS the menu's components to the resolver's
// section/element base; it never removes. Intersection (drop what the menu
// withheld) would be a tightening on every site — a separate decision with its
// own blast radius (see bugs_open/276 for the requires-backend case).
//
// OPT-IN, UNSAFE DEFAULT OFF. The step must name the menu field
// (`menu_field` in validate_site_plan's step config). Absent key = today's
// behaviour exactly, so the Go and the config half are safe to land in either
// order — which they must be, since Go rides the next image roll while
// agent_definitions config is live on apply.

// menuRowsFrom reads a component menu from a workflow's collected data.
//
// The live shape is a `query_database` step with output_format "array": the
// coordinator stores []interface{} of map[string]interface{} rows under the
// step's output_field. Anything else — absent, a string, a map, an empty array —
// returns ok=false, and the caller falls back to the section/element base rather
// than silently narrowing what it will accept.
func menuRowsFrom(collected map[string]interface{}, path string) ([]interface{}, bool) {
	if collected == nil || path == "" {
		return nil, false
	}
	raw := datahelpers.ExtractNestedField(collected, path)
	rows, ok := raw.([]interface{})
	if !ok || len(rows) == 0 {
		return nil, false
	}
	return rows, true
}

// addMenu adds the components of a planner's own menu to the resolver's valid
// set, so a name the planner was OFFERED can never be dropped by the surface
// that validates its proposal.
//
// Each row is a menu row as the menu query selects it: "function" is the
// canonical identity; "name" and "display_name" feed the same lower-cased
// lookup maps `loadComponentNameResolver` builds, so a planner that echoes a
// display name resolves exactly as it would for a section-level component.
// Rows that are not objects, or carry no "function", are ignored — a menu query
// that selects different columns degrades to no-op rather than to a resolver
// that accepts "".
//
// Existing base entries WIN on the name/display maps: the DB base is loaded
// from content_components directly and is the stronger statement of identity.
// Returns the number of functions the menu added beyond that base — the count
// is the observable that says whether this arm did anything at all.
func (r *componentNameResolver) addMenu(rows []interface{}) int {
	if r == nil {
		return 0
	}
	added := 0
	for _, raw := range rows {
		row, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		fn, _ := row["function"].(string)
		fn = strings.TrimSpace(fn)
		if fn == "" {
			continue
		}
		if !r.validFunctions[fn] {
			r.validFunctions[fn] = true
			if r.menuOnly == nil {
				r.menuOnly = make(map[string]bool)
			}
			r.menuOnly[fn] = true
			added++
		}
		if name, _ := row["name"].(string); name != "" {
			if _, exists := r.nameToFunc[strings.ToLower(name)]; !exists {
				r.nameToFunc[strings.ToLower(name)] = fn
			}
		}
		if display, _ := row["display_name"].(string); display != "" {
			if _, exists := r.displayToFunc[strings.ToLower(display)]; !exists {
				r.displayToFunc[strings.ToLower(display)] = fn
			}
		}
	}
	return added
}

// resolvedViaMenu reports whether a resolved function owes its acceptance to the
// menu rather than to the section/element base. It exists so a run can SAY that
// this arm fired — the tell that distinguishes "the fix works" from "the planner
// happened to propose nothing tool-level this time", which otherwise look
// identical in the logs and in site_plan_sections.
func (r *componentNameResolver) resolvedViaMenu(fn string) bool {
	return r != nil && r.menuOnly[fn]
}

// ============================================================================
// The residual: a dropped name is still a SILENT loss (bugs_open/282, council
// round 1, bug_historian's gating objection)
// ============================================================================
//
// Accepting the menu fixes the class of drop that motivated 282, for the
// callers that opt in. It does nothing for the REST of the drop surface: a
// typo, a renamed function, a deleted component, or a caller that never opts
// in still loses its section with a single Warn and no durable trace — which is
// byte-identical to the failure this whole lane exists to remove, and is this
// estate's most-repeated shape (a plan/rebuild pass silently loses content and
// nobody learns until the damage is visible).
//
// So every drop now files a durable finding, whether or not a menu was
// configured. It uses the door this action already uses for exactly this
// purpose — LogActionFindings, whose own header gives the standing reason: the
// chassis logs rotate sub-second, so a log line is not a record. Severity is
// warning, matching recordRecomposeOutcomes: a drop IS a legal outcome of the
// mechanism (an unresolvable name SHOULD be removed rather than persisted into
// a plan nothing can build) — what it must not be is invisible.

// droppedSectionName is one section the resolver could not resolve, with the
// page it was proposed for.
type droppedSectionName struct {
	Page string
	Name string
}

// recordDroppedSectionNames persists one durable row per section name validate
// removed from the plan. menuField is reported so a reader can tell the two
// cases apart without reading the agent config: a drop with no menu configured
// invites "should this planner be reading its own menu?" (the 282 fix), while a
// drop WITH a menu configured means the name was in neither the component base
// nor anything the planner was offered — a typo, a rename, or a hallucination.
// droppedRemedy is the operator-facing "what do I do about this" for a dropped
// name. The two cases are diagnostically different and must not share a
// sentence: with no menu configured the name may be one the planner was
// legitimately OFFERED (282's shape, and the reader should suspect the config
// before the name); with a menu configured it resolved to neither the component
// base nor anything the planner saw, so it is a typo, a rename, or invention.
func droppedRemedy(menuField string) string {
	if menuField == "" {
		return "this planner's validate step does not read its own component menu (no menu_field configured), so a component the planner WAS offered can still be dropped here — see bugs_open/282 and register PLAN-050 before assuming the name is wrong"
	}
	return "the planner proposed a section this site's component set does not contain: check for a renamed or deactivated component, or a planner prompt naming a component that does not exist"
}

// droppedFindings builds the durable rows for a set of drops. Pure, so the
// record's shape can be asserted without a database — the findings door itself
// is exercised by the actions that already use it.
func droppedFindings(drops []droppedSectionName, menuField string) []agenterrors.Finding {
	findings := make([]agenterrors.Finding, 0, len(drops))
	remedy := droppedRemedy(menuField)
	for _, d := range drops {
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "PLAN_SECTION_NAME_DROPPED",
			Severity:  "warning",
			Message: fmt.Sprintf("section %q proposed for page %q was dropped from the plan: it resolves to no active component",
				d.Name, d.Page),
			Context: map[string]interface{}{
				"page":       d.Page,
				"section":    d.Name,
				"menu_field": menuField,
				"remedy":     remedy,
			},
		})
	}
	return findings
}

// recordDroppedSectionNames persists one durable row per section name validate
// removed from the plan. menuField is reported so a reader can tell the two
// cases apart without reading the agent config.
//
// It RETURNS the number of rows attempted, and that return is not decoration:
// "a clean plan writes nothing" is a negative, and a database mock cannot prove
// a negative — it fails an *expected* call that never came, never an
// *unexpected* call that did. Returning the count moves the claim somewhere a
// test can actually assert it. (Found by mutation: deleting the empty-drops
// guard left every test passing.)
func recordDroppedSectionNames(ctx context.Context, params ActionParams, drops []droppedSectionName, menuField string) int {
	if len(drops) == 0 {
		return 0
	}
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	attempted, _ := LogActionFindings(ctx, params, siteID, "", "validate_plan", droppedFindings(drops, menuField), params.Logger)
	return attempted
}
