// FILE: platform/livespec/rerender_reasons.go
//
// ONE DEFINITION of the re-render SECTIONS-BRANCH routing vocabulary —
// bugs_open/404.
//
// THE DEFECT THIS ENDS. `page-rerender`'s `check_rerender_mode` step is an
// allow-list of reasons that route to `rerender_sections`; anything else takes
// `else_step: render_page`, which re-ships the stored HTML verbatim. The Go
// action that STAMPS reasons onto items carried its own hardcoded copy of that
// list, and the fixer that files items carries a third copy in raw SQL. On
// 2026-08-18 two lanes appended a value each to the live gate — `template_changed`
// (migration 460) and `literal_markdown` (473) — and neither touched Go. So the
// gate knew five and the Go reader knew three.
//
// ── WHY THAT IS DANGEROUS RATHER THAN UNTIDY ───────────────────────────────
//
// **Every reader that does not know a value fails toward `assemble`.** The gate's
// `else_step` is `render_page`; the Go reader's unknown reason leaves the item
// carrying no reason at all, which the gate then reads as assemble. A vocabulary
// whose readers failed toward RE-RESOLVE would announce itself — you would get too
// many re-renders and notice. Failing toward assemble means the estate's own
// preferred, safe, cheap mode is also its silent-failure mode: the work completes,
// the status is green, and nothing changed. `bugs_open/283` §13 had to measure it
// end to end (111 items completed, page DEPLOYED, served bytes unchanged) before
// anyone believed it.
//
// ── WHAT THIS GOVERNS, AND WHAT IT DOES NOT ────────────────────────────────
//
// This is the ROUTING half only. `[MEASURED 2026-08-26]` there are **sixteen**
// distinct strings in `spec.reason` across live and archived `page_rerender`
// items, and four of them are free prose — whole sentences, a bug reference, an
// operator's note to themselves. **`spec.reason` is two fields wearing one name:**
// a routing key the gate branches on, and an annotation humans write for humans.
// One Go producer does it deliberately too — `adopt_verbatim.go` stamps
// `verbatim_adoption_deploy` on items that are SUPPOSED to assemble, because
// verbatim adoption re-ships stored HTML by design.
//
// That is why an undeclared reason is WARNED and not REFUSED at the item creator:
// refusing would be new authority on a shared seam with a live legitimate
// counter-example (owner ruling 2026-08-02 §2). Splitting annotation from routing
// key is the real repair and is RFC-scope; this file is the half that can ship now.
//
// ── WHY THE DEFINITION LIVES IN `livespec` ─────────────────────────────────
//
// Because the other half of the loop is a LIVE DATABASE OBJECT, and this package
// exists for exactly that (bugs_open/363). A Go guard that asserts a property of a
// live object by reading the MIGRATION FILE cannot work: a migration is
// append-only history frozen by its checksum, while the live object is the
// accumulation of every migration since plus lawful direct edits. So the list
// lives here, the `Declarations` entries beside it say what the live objects must
// contain, and the daily auditor compares those declarations against the live
// objects. Go ↔ declaration by construction; declaration ↔ live every morning.
//
// `ClaimedItemTimeoutExclusions` + `ClaimedItemTimeoutExclusionClause()` in
// livespec.go is the worked precedent for a Go list generating the fragment a
// declaration asserts. This is the same shape, for the same reason.
//
// ⚠ DB CONFIG CANNOT IMPORT GO, so the live gate is ASSERTED against this list,
// never generated from it at runtime. A migration author appending a sixth value
// PASTES `CheckRerenderModeConditionClause()`'s output rather than hand-writing a
// disjunct — and if they hand-write it anyway, the corpus lint in
// rerender_reasons_test.go fails at commit and the auditor fails the next morning.

package livespec

import "strings"

// RerenderSectionReason is one value of the sections-branch routing vocabulary.
//
// It is a struct rather than a bare string because the two gates in
// `create_rerender_items` are DIFFERENT TESTS and each reason belongs in them on
// the merits of what that reason MEANS. Encoding the judgement here is the point:
// leaving it to be re-derived at the call site is how the call site drifted.
type RerenderSectionReason struct {
	Name string

	// ComponentScoped: the fan-out narrows to the pages carrying the triggering
	// component. Requires a component_id — without one there is nothing to scope
	// BY, and the caller's own page list stands.
	ComponentScoped bool

	// StampAlways: stamp the reason onto the item even when the fan-out is not
	// component-scoped.
	//
	// ⚠ DELIBERATELY NOT SET for image_landed and section_data_resolved. Those two
	// carry REB-001's designed degrade — a reason without a component_id falls
	// back to assemble-only — and preserving it byte-for-byte is not this bug's to
	// change. Setting StampAlways on either would alter behaviour for the two
	// reasons that are working correctly, in a change about the two that are not.
	StampAlways bool
}

// The vocabulary's values as named constants.
//
// ⚠ THESE EXIST FOR THE READERS THAT ARE *NOT* create_rerender_items, and adding
// them is what the council's bug_historian and reuse_agent seats asked for at
// round 1 — independently, and correctly.
//
// Two further readers branch on a SINGLE value each, outside the item creator:
//
//	rerender_page_sections_action.go  shouldStripLiteralMarkdown — the strip is
//	                                  double-gated on the step flag AND this reason
//	rerender_page_sections_action.go  the CTA recompute, gated on cta_links_stale
//
// Left as bare string literals they would be two more untethered copies of a
// vocabulary value — which is 016b §9's "one call site of a shared judgement gets
// the rigorous fix; the sibling stays heuristic", and it is this bug's own thesis
// that EVERY reader failing to track the vocabulary is the hazard. Naming them
// here means retiring or renaming a value breaks COMPILATION at every reader
// instead of silently changing what one of them does.
//
// TestReasonConstantsAreExactlyTheVocabulary pins the two directions together, so
// a constant cannot drift from the list or the list from the constants.
const (
	ReasonImageLanded         = "image_landed"
	ReasonSectionDataResolved = "section_data_resolved"
	ReasonCTALinksStale       = "cta_links_stale"
	ReasonTemplateChanged     = "template_changed"
	ReasonLiteralMarkdown     = "literal_markdown"
)

// RerenderSectionReasons is the vocabulary.
//
// ⚠ ORDER IS LOAD-BEARING. CheckRerenderModeConditionClause renders in slice
// order, and the Declaration for the live gate asserts that exact string with
// Min:1 Max:1 — so reordering this slice makes the daily auditor fail until the
// live condition is rewritten to match. That is the same rule
// ClaimedItemTimeoutExclusions carries, for the same reason.
var RerenderSectionReasons = []RerenderSectionReason{
	// The legacy pair. Component-caused, and their reason-without-component_id
	// degrade to assemble is REB-001's design, preserved.
	{Name: ReasonImageLanded, ComponentScoped: true},
	{Name: ReasonSectionDataResolved, ComponentScoped: true},

	// Stamped WITHOUT component scoping: the CTA recompute is cheap and
	// page-scoped, and a site-wide CTA repair has no single triggering component
	// to scope by. (The reason create_rerender_items already gave for this one.)
	{Name: ReasonCTALinksStale, StampAlways: true},

	// Component-caused, so scope when a component_id is given — the fixer's own
	// INSERT scopes by `pc.component_id`. But stamp EITHER WAY: assemble can never
	// deliver a template change at all (rerender_page_sections is the only writer
	// of page_components.rendered_html), so the safe degrade here is
	// over-delivery, which is self-announcing, rather than silent under-delivery.
	{Name: ReasonTemplateChanged, ComponentScoped: true, StampAlways: true},

	// Page-wide by meaning: the repair strips markdown from the stored content of
	// the pages it is given. There is no triggering component, and scoping it to
	// some component's dependents would silently SKIP dirty pages — which would be
	// a new defect of exactly this bug's shape.
	{Name: ReasonLiteralMarkdown, StampAlways: true},
}

// CheckRerenderModeConditionClause renders the live gate's condition from the
// list. PASTE THIS OUTPUT into a migration that changes the vocabulary; do not
// hand-write a sixth disjunct.
func CheckRerenderModeConditionClause() string {
	parts := make([]string, 0, len(RerenderSectionReasons))
	for _, r := range RerenderSectionReasons {
		parts = append(parts, "input_data.spec.reason == '"+r.Name+"'")
	}
	return strings.Join(parts, " OR ")
}

// RerenderSectionReasonByName resolves a reason. The bool is the whole point:
// "not in the vocabulary" is a state a caller must be able to see and report,
// not one it should silently treat as absent.
func RerenderSectionReasonByName(name string) (RerenderSectionReason, bool) {
	for _, r := range RerenderSectionReasons {
		if r.Name == name {
			return r, true
		}
	}
	return RerenderSectionReason{}, false
}

// RerenderSectionReasonNames is for a log line that tells an operator what the
// vocabulary actually is at the moment they got it wrong.
func RerenderSectionReasonNames() []string {
	out := make([]string, 0, len(RerenderSectionReasons))
	for _, r := range RerenderSectionReasons {
		out = append(out, r.Name)
	}
	return out
}
