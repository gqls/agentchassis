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

// RefuseUnknownRoutingKeyMessageTemplate renders the refusal message for
// RFC_062 phase 3's `needs_human_review` step — owner ruling D1: "the message
// names the bad key AND the vocabulary".
//
// ⚠ IT LIVES HERE, WITH THE VOCABULARY, FOR THE REASON THIS WHOLE FILE EXISTS.
// The message has to list the legal values, so writing it in the migration
// would put a FOURTH hand-maintained copy of RerenderSectionReasons into the
// estate — after the gate condition, the Go reader and the fixer's raw INSERT,
// which is the exact drift bugs_open/404 records (two lanes appended a value to
// the live gate on one day in 2026-08 and neither touched Go). Rendered from
// the list, a sixth value reaches the operator's message for free.
//
// The `{{...}}` is deliberate and is NOT this package's business to evaluate:
// it is `fail_work_item`'s opt-in `error_message_template`, rendered by
// text/template over the run's collected data at refusal time. That is what
// names the OFFENDING VALUE — a static literal can name the field and the
// vocabulary but never the key that was actually wrong, which is the half of
// D1 that needed a code change (fail_work_item_message_template.go).
//
// PASTE the output into the migration; do not hand-write it.
func RefuseUnknownRoutingKeyMessageTemplate() string {
	return "This page_rerender item was REFUSED, not assembled: its spec." +
		RoutingReasonSpecKey + " = '{{.input_data.spec." + RoutingReasonSpecKey +
		"}}' is not in the sections-rerender vocabulary (" +
		strings.Join(RerenderSectionReasonNames(), ", ") +
		"). Before this refusal existed, an unrecognised routing key completed " +
		"GREEN having changed nothing — bugs_open/440. If the value was meant as " +
		"a note for a human, move it to spec.reason, which is free prose and is " +
		"never validated. If it was meant to ROUTE, use a vocabulary value, or " +
		"add one to RerenderSectionReasons in platform/livespec and paste the " +
		"regenerated gate clause and this message into a migration. RFC_062."
}

// RefuseUnknownRoutingKeyMessageFallback is the STATIC message the same step
// configures as `error_message`, used when the template above does not render.
//
// It exists because fail_work_item falls back to the literal rather than
// failing — parking the item with a plainer message beats not parking it — and
// a fallback that was EMPTY would park the item with no explanation at all.
// It names the field and the vocabulary; only the offending value is lost,
// and that value is one field away on the row the reader is already looking at.
func RefuseUnknownRoutingKeyMessageFallback() string {
	return "This page_rerender item was REFUSED, not assembled: its spec." +
		RoutingReasonSpecKey + " is not in the sections-rerender vocabulary (" +
		strings.Join(RerenderSectionReasonNames(), ", ") +
		"). Read the offending value at spec." + RoutingReasonSpecKey +
		" on this row. If it was meant as a note for a human, move it to " +
		"spec.reason, which is free prose and is never validated. " +
		"bugs_open/440 / RFC_062."
}
