// FILE: platform/livespec/rerender_routing_key.go
//
// The ROUTING-KEY half of the spec.reason split — bugs_open/440, RFC_062,
// register REB-008. Phase 1a: INERT BY DESIGN. Zero callers outside its own
// tests as of 2026-09-02 (enumerable: grep -rn RoutingReasonSpecKey), no
// behaviour change anywhere, unsafe side default OFF per the owner ruling of
// 2026-08-02 §2 — which makes this NOT architecture-scope under RFC_022's
// three-condition test. The phase-3 flip (the gate refusing a
// present-but-unknown key) is the guarantee change and goes through RFC_062.
//
// WHY A SECOND KEY RATHER THAN VALIDATING spec.reason. `spec.reason` is two
// fields wearing one name (bugs_open/404's measurement, re-confirmed
// 2026-09-02): a routing key the gate branches on AND a free-prose annotation
// humans write for humans — eleven prose items were minted by migrations the
// day this file was written. You cannot refuse unknown values of a field whose
// legitimate values include arbitrary sentences. After the split, `reason`
// stays free forever, `routing_reason` carries only vocabulary values, and
// "present but unknown" becomes a mistake BY CONSTRUCTION with no legitimate
// counter-example — refusable at last.
//
// The three states, and the design rule each encodes:
//
//	absent  → assemble. Today's safe default, preserved: an annotation-only
//	          item (migration 696's shape) is legal forever.
//	known   → route per the vocabulary, exactly as `reason` routes today.
//	unknown → REFUSE toward review, never silently assemble. An unknown
//	          routing key that completes green is bugs_open/440 in one
//	          sentence, and before the split it was IMPOSSIBLE to refuse.
//
// ⚠ The clause renderers below are PASTED into migrations (DB config cannot
// import Go — see rerender_reasons.go's header for the full argument and the
// declaration-auditor idiom that holds the paste honest daily).
//
// > **RESOLVED 2026-09-03 — this header used to say "CONFIRM that the evaluator
// > treats a MISSING key and '' identically before pasting; getting it wrong
// > inverts the guard for legacy items." It was confirmed, BY EXECUTING THE
// > EVALUATOR, and the assumption was FALSE: a missing key does NOT match `''`.
// > The renderer was carrying that defect and is fixed below. The warning is
// > kept as a correction rather than deleted, because the next person to write a
// > workflow condition that tests emptiness will make the same assumption
// > (LANDMINES, 2026-09-03).**

package livespec

import "strings"

// RoutingReasonSpecKey is the spec key the routing half lives under —
// `spec.routing_reason`. The annotation stays in `spec.reason`, unvalidated.
const RoutingReasonSpecKey = "routing_reason"

// RoutingDecision is what a reader must do with a routing key. It is a closed
// three-state enum rather than (bool, bool) so a caller cannot express
// "unknown but proceed" — the state this seam exists to make unrepresentable.
type RoutingDecision int

const (
	// RoutingAssemble: no routing key — an annotation-only item. Assemble,
	// silently, correctly.
	RoutingAssemble RoutingDecision = iota
	// RoutingSections: a vocabulary key — route to the sections re-render per
	// the reason's own scoping/stamping judgement.
	RoutingSections
	// RoutingRefuse: present but not in the vocabulary. A mistake by
	// construction; the consumer must fail the item toward review with a
	// message naming the vocabulary, never silently assemble.
	RoutingRefuse
)

// ResolveRoutingReason classifies a routing-key value into the three states.
// The zero-value RerenderSectionReason accompanies both non-Sections states.
func ResolveRoutingReason(value string) (RerenderSectionReason, RoutingDecision) {
	if value == "" {
		return RerenderSectionReason{}, RoutingAssemble
	}
	if r, ok := RerenderSectionReasonByName(value); ok {
		return r, RoutingSections
	}
	return RerenderSectionReason{}, RoutingRefuse
}

// TransitionRerenderModeConditionClause renders the compat gate condition for
// the drain window (RFC_062 phase 3): every vocabulary value accepted under
// EITHER key, so in-flight items minted before producers stamp the new key
// keep routing. PASTE the output; do not hand-write it.
func TransitionRerenderModeConditionClause() string {
	parts := make([]string, 0, len(RerenderSectionReasons)*2)
	for _, r := range RerenderSectionReasons {
		parts = append(parts, "input_data.spec."+RoutingReasonSpecKey+" == '"+r.Name+"'")
	}
	for _, r := range RerenderSectionReasons {
		parts = append(parts, "input_data.spec.reason == '"+r.Name+"'")
	}
	return strings.Join(parts, " OR ")
}

// CheckRoutingKnownConditionClause renders the read-door guard for the phase-3
// refusal step: TRUE when the routing key is absent or known (proceed), FALSE
// when present-but-unknown (the step's else branch refuses).
//
// ⚠ THE `== null` DISJUNCT IS LOAD-BEARING AND WAS MISSING FROM THE FIRST CUT.
// The evaluator's own semantics, MEASURED by executing it 2026-09-03 rather
// than read (conditional_branch_action.go: compareValues' nil branch runs
// BEFORE quote-stripping, so a quoted ” never equals nil):
//
//	state             == null   == ''   == 'cta_links_stale'
//	absent (no key)    TRUE     false        false
//	present, ""        false     TRUE        false
//	present, known     false    false         TRUE
//	present, unknown   false    false        false     <- the ONLY refusing state
//
// So a clause carrying `== ”` alone evaluates FALSE for every item that has no
// routing key — i.e. every item minted before phase 2 — and would have sent the
// whole legacy population down the refusal branch to human review on the day
// the gate flipped. Both disjuncts are required: `== null` for absent, `== ”`
// for present-but-empty. The four-state table is pinned by
// TestCheckRoutingKnownConditionClause_CoversEveryEvaluatorState in the actions
// package, where the real evaluator lives.
func CheckRoutingKnownConditionClause() string {
	parts := make([]string, 0, len(RerenderSectionReasons)+2)
	parts = append(parts, "input_data.spec."+RoutingReasonSpecKey+" == null")
	parts = append(parts, "input_data.spec."+RoutingReasonSpecKey+" == ''")
	for _, r := range RerenderSectionReasons {
		parts = append(parts, "input_data.spec."+RoutingReasonSpecKey+" == '"+r.Name+"'")
	}
	return strings.Join(parts, " OR ")
}
