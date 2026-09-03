// FILE: platform/livespec/rerender_reason_stamping.go
//
// ONE definition of "how a page_rerender spec carries its reason" —
// bugs_open/440 phase 2, RFC_062, register REB-008.
//
// WHY THIS EXISTS RATHER THAN A KEY ADDED AT EACH SITE. The routing/annotation
// split needs every producer to write `routing_reason` beside `reason`. Doing
// that by hand at each call site would re-create, one level along, the exact
// defect this lane is closing: a vocabulary judgement copied to N sites, which
// then drift (`bugs_open/404` is that story — the gate knew five values and Go
// knew three). The pair is defined here, once, and the sites call it.
//
// THE RULE THE HELPERS ENCODE, and it is phase 1b's lockstep rule generalised:
// `reason` is always written as given (it is the free-prose ANNOTATION and is
// never validated — owner ruling D4, 2026-09-03); `routing_reason` is written
// ONLY when the value is in the vocabulary. An out-of-vocabulary reason
// therefore yields NO routing key, so RFC_062 phase 3 can never refuse an item
// the estate's own producers minted (REB-008's no-bad-producer constraint).
//
// ⚠ SCOPE: `item_type='page_rerender'` specs ONLY. `[MEASURED 2026-09-03]` four
// nearby Go sites write an in-vocabulary reason onto a DIFFERENT item type —
// `render_directory_action.go` and `reconcile_section_data_action.go` file
// `needs_page`, `store_generated_component_action.go` files `needs_rerender`
// (whose reason is propagated INTO page_rerender items by the item creator,
// which stamps there), and `discovery_checks/check_literal_markdown.go` files
// `literal_markdown`. Stamping a routing key on those would put a
// page-rerender routing decision on an item no rerender gate ever reads. Check
// the item type before calling these helpers; a blanket sweep gets it wrong.

package livespec

import (
	"encoding/json"
	"strings"
)

// RerenderReasonFields returns the reason fields a page_rerender spec carries
// for the given reason: always `reason`, plus `routing_reason` when (and only
// when) the value is in the sections-rerender vocabulary. An empty reason
// yields no fields at all — an item with no reason is the assemble-only case
// and must stay exactly that.
func RerenderReasonFields(reason string) map[string]string {
	if reason == "" {
		return nil
	}
	out := map[string]string{"reason": reason}
	if _, known := RerenderSectionReasonByName(reason); known {
		out[RoutingReasonSpecKey] = reason
	}
	return out
}

// StampRerenderReason writes those fields into a spec map being built as a Go
// map, and REPORTS whether the value was in the vocabulary. Nil-safe on an
// empty reason so callers need no branch of their own.
//
// ⚠ THE BOOL IS DELIBERATE, and it is the sibling's rule applied here
// (RerenderSectionReasonByName: "the bool is the whole point: 'not in the
// vocabulary' is a state a caller must be able to see and report, not one it
// should silently treat as absent"). Without it this helper answers an
// out-of-vocabulary reason with silence — no routing key and no signal — which
// is the shape the council's bug_historian seat objected to at round c7dab2c1
// and which 016b §9 records as "a mistyped routing key produces silence in
// every gate at once".
//
// A caller passing a COMPILE-TIME CONSTANT from this package may ignore it:
// the constant cannot be out-of-vocabulary, and a test in this package pins
// that. A caller passing a VARIABLE must not — report it the way
// create_rerender_items does (warn, name the vocabulary, and let the item
// assemble), because a reason that reached that call site from config or from
// a database row is exactly how this vocabulary drifted in the first place
// (bugs_open/404).
func StampRerenderReason(spec map[string]interface{}, reason string) (known bool) {
	fields := RerenderReasonFields(reason)
	for k, v := range fields {
		spec[k] = v
	}
	_, known = fields[RoutingReasonSpecKey]
	return known
}

// RerenderReasonJSONPrefix renders the same fields as the OPENING of a JSON
// object body — the fields plus a TRAILING COMMA — for the specs still built as
// string literals (`fmt.Sprintf` with a backticked JSON template). Returns ""
// for an empty reason, or if marshalling somehow fails.
//
// ⚠ THE TRAILING COMMA IS THE SAFETY PROPERTY, not an oddity. Callers write
// `{%s"page_name":%q}`, so BOTH states compose into valid JSON: with fields,
// `{"reason":"x","routing_reason":"x","page_name":"y"}`; without,
// `{"page_name":"y"}`. The comma-separated alternative (`{%s,"page_name":%q}`)
// emits `{,"page_name":"y"}` the first time anyone passes a variable that is
// empty — invalid JSON, written into a text column, discovered whenever
// something next tries to parse that spec. A helper must not have a shape that
// only works while every caller passes a constant.
//
// It is DERIVED from RerenderReasonFields by marshalling and stripping the
// braces, not written out a second time by hand: two hand-written renderings of
// one rule is the drift this file exists to prevent, and
// TestRerenderReasonJSONPrefixMatchesTheMapForm pins the two together.
func RerenderReasonJSONPrefix(reason string) string {
	fields := RerenderReasonFields(reason)
	if len(fields) == 0 {
		return ""
	}
	b, err := json.Marshal(fields)
	if err != nil {
		// Unreachable for map[string]string, and silently emitting a fragment
		// that does not parse would corrupt the caller's spec — so emit
		// nothing and let the caller's own spec stay valid JSON.
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(string(b), "{"), "}") + ","
}
