// FILE: platform/orchestration/actions/complete_work_item_no_change.go
//
// Completion gate 1b: a handler that reports it changed NOTHING has not repaired
// anything, so its item must not be stamped 'complete'.
//
// WHY THIS EXISTS, and why it is a gate rather than a verifier (bugs_open/213 D1).
// `dark_section_audit` items are dispatched to color-variable-fixer, whose repair
// transform provably cannot touch them: [MEASURED 2026-08-12] the handler's own
// literal transform (checks.ReplaceHardcodedColors) changes 0 of 61 live bodies —
// 0/16 of the components those items name, 0/16 of their backing templates, and
// 0/29 across the action's own sweep sets on all 15 affected sites. The reason
// generalises past any one instance: that transform's entire output alphabet is
// var(--color-primary) and var(--color-secondary), while every one of the 15 live
// items asks for --section-text / --section-heading / --color-cta-*. A transform
// with two words cannot satisfy a criterion naming five others, on any input.
//
// The observable damage, first-hand and re-runnable. finetuning.uk, item_key
// design-audit_dark_section_audit_index_1368e337-…:
//
//	764fe035  complete  created 2026-08-11 13:21  completed 13:38
//	          result.response.fix_result.total_fixed        = 0
//	          result.response.text_color_result.total_fixed = 0
//	b82b9f1f  detected  created 2026-08-12 13:39   ← same item_key, re-filed by the next audit
//
// Every page_components.updated_at on that page reads 2026-08-11 09:37:58 —
// BEFORE the item was created — so nothing changed either side of the completion.
// The handler stated in its own payload that it had done nothing, the item closed
// green, and the defect was still there the next day. Fleet-wide the same census
// shows color-variable-fixer has NEVER reported a non-zero fix on either of its
// item types (0 of 30 live rows).
//
// WHY NOT A VERIFIER, which was the obvious first design and does not work:
// verifyBeforeComplete's VerifyTarget carries the spec, not the result, and the
// handler's report is an ACTION INPUT that load_work_item_actions.go marshals into
// site_work_items.result only AFTER the gate has run. A verifier querying that
// column would read the row's PREVIOUS value and grade the wrong evidence. So this
// question — "did the handler change anything?" — can only be asked here, beside
// handlerReportedFailure, which reads the same payload for the same reason.
//
// It is also cheaper than the alternatives in a way that matters: it needs no
// browser, no HTTP probe and no page fetch on the completion path, which is the
// standing objection recorded three times in verifier_coverage_test.go against
// every option that computes a rendered property. It cannot confirm a repair — only
// refuse a completion that provably is not one. That is deliberate: the damage in
// bugs_open/213 is the false green, not the missing green, and grading the item's
// own stated acceptance_test is a separate and larger job (that field is free LLM
// prose; 10 of 15 live values name a computed property and 2 contain clauses no
// probe can assess, so it needs a producer-side contract change first).
//
// OPT-IN, WITH THE UNSAFE DEFAULT OFF, per the owner's 2026-08-02 shared-seam
// ruling. An item_type absent from noChangeGates below takes a map miss and this
// file changes nothing about it — byte-identical to today. That is not decoration:
// "the handler changed nothing" is a legitimate SUCCESS for other handlers (an
// idempotent repair finding its work already done, a check-and-confirm step), so
// applying this fleet-wide would block real completions. Whoever adds a type here
// is asserting, for that type, that a zero-change run cannot be a repair — and
// owes the measurement that says so, in the rule's Why.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// noChangeRule declares, for ONE item_type, how to read its handler's own report
// of how much it changed.
//
// CounterPaths are read from the result payload the handler returned. ALL of them
// must resolve to a number and ALL must be zero before completion is blocked — a
// single non-zero counter anywhere means the handler did something, and this gate
// has no business judging whether what it did was enough (that is a verifier's
// question, and for this type there is not yet one that can be written).
type noChangeRule struct {
	// Why the type opts in: the measurement that licenses "a zero-change run
	// cannot be a repair here". Surfaced to whoever reads a blocked item.
	Why string

	// CounterPaths are dotted paths into the handler's result payload, each
	// expected to hold an integer count of artefacts changed.
	CounterPaths []string
}

// noChangeGates is the opt-in roster. Absent item_type → this file is inert.
//
// Keep the roster small and each entry evidenced. A type here without a
// measurement in its Why is a guess about somebody else's handler.
var noChangeGates = map[string]noChangeRule{
	// The two counters are color-variable-fixer's two repair steps
	// (fix_hardcoded_colors and the forced-text-colour step). Both zero means
	// neither step altered a single template or rendered component.
	"dark_section_audit": {
		Why: "color-variable-fixer's transform emits only var(--color-primary)/var(--color-secondary), " +
			"and every live dark_section_audit item asks for --section-* / --color-cta-* properties it cannot write; " +
			"measured 2026-08-12, the handler's literal transform changes 0 of 61 live bodies and the agent has " +
			"never reported a non-zero fix on any of its 30 rows (bugs_open/213 D1)",
		CounterPaths: []string{
			"response.fix_result.total_fixed",
			"response.text_color_result.total_fixed",
		},
	},
}

// handlerReportedNoChange reports whether the result about to be stored is itself
// a record of the handler having changed nothing.
//
// Returns (detail, noChange, unknownShape), mirroring handlerReportedFailure's
// three-valued contract for the same reason: the third case is neither a pass nor
// a block, and it must be RECORDED rather than swallowed.
//
//   - noChange=true  — every declared counter resolved and every one was zero.
//   - unknownShape≠"" — the type opted in, but NO declared counter could be
//     resolved. The item still completes: a payload this guard cannot read is not
//     evidence of a no-op, and inverting that would block legitimate work on a
//     handler whose response shape simply differs. But it is exactly the drift
//     that makes a guard silently stop guarding, so the caller records it.
//     This case is live TODAY and is why the arm exists rather than being
//     defensive boilerplate: of the 14 completed dark_section_audit items,
//     [MEASURED 2026-08-12] only 4 carry the fixer's response envelope at all;
//     the other 10 carry a payload that is not this handler's (a design-system
//     spec for 9 of them, an unrelated child-page triage decision for the 10th).
//     Why that is so is NOT ESTABLISHED — see bugs_open/213 §D, which records it
//     as an observation and deliberately does not guess. Instrumenting it here
//     turns an unexplained split into a queryable one.
//
// A partially-resolved payload counts as readable: the counters that DID resolve
// are judged, and the missing ones are named in the detail. Requiring all of them
// would let a handler escape the gate by dropping one field.
func handlerReportedNoChange(itemType string, result map[string]interface{}) (string, bool, string) {
	rule, opted := noChangeGates[itemType]
	if !opted {
		return "", false, ""
	}

	var (
		zero    []string
		nonZero []string
		missing []string
	)
	for _, path := range rule.CounterPaths {
		n, ok := lookupNumericPath(result, path)
		switch {
		case !ok:
			missing = append(missing, path)
		case n == 0:
			zero = append(zero, path)
		default:
			nonZero = append(nonZero, fmt.Sprintf("%s=%g", path, n))
		}
	}

	// Any evidence of work done → not this gate's business.
	if len(nonZero) > 0 {
		return "", false, ""
	}

	// Nothing readable → cannot assert a no-op. Complete, but say so.
	if len(zero) == 0 {
		return "", false, fmt.Sprintf(
			"item_type %s opted into the no-change gate but none of its declared counters (%s) "+
				"are present in the handler's result; payload top-level keys were [%s]",
			itemType, strings.Join(rule.CounterPaths, ", "), strings.Join(topLevelKeys(result), " "))
	}

	detail := fmt.Sprintf("handler reported 0 changes at %s", strings.Join(zero, " and "))
	if len(missing) > 0 {
		detail += fmt.Sprintf(" (no value present at %s)", strings.Join(missing, ", "))
	}
	return detail + " — " + rule.Why, true, ""
}

// lookupNumericPath resolves a dotted path to a number.
//
// TRAVERSAL IS REUSED, COERCION IS NOT, and the split is deliberate (council
// reuse_agent seat, correlation 0c8e7f5b round 2, medium). The dotted-path descent
// is datahelpers.ExtractNestedField — the fleet's existing resolver, which is
// strictly better than the hand-rolled loop this replaced: it also indexes arrays
// and auto-unwraps a `.response` wrapper.
//
// What could NOT be reused is datahelpers.ExtractNestedFieldInt, the obvious
// candidate, and the reason is the whole basis of this gate's three-valued
// contract: it returns 0 for a MISSING path and 0 for a path holding zero
// (data_helpers.go:1290-1304). Those are the two outcomes this gate must keep
// apart — "the handler reported no changes" (block) versus "I cannot read this
// payload" (abstain and record). Collapsing them would block the 10-of-14 live
// rows whose payload is not this handler's at all, which is the opposite of what
// the evidence supports. So presence is decided here, from the raw value.
//
// The accepted numeric forms are load-bearing rather than defensive: the result
// map arrives either from a Go action's return value (int) or through a JSON
// round-trip in collected_data (float64, or json.Number under some decoders). A
// switch missing one would read "counter absent" for a counter that is present and
// zero — reporting unknown shape where the data says block. Both numeric arms are
// mutation-proven for exactly that reason.
func lookupNumericPath(m map[string]interface{}, path string) (float64, bool) {
	cur := datahelpers.ExtractNestedField(m, path)
	if cur == nil {
		return 0, false
	}
	switch v := cur.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// topLevelKeys names what WAS in the payload, sorted so the recorded string is
// stable and greppable across runs. Without it the unknown-shape record says only
// "not what I expected", which is the least useful half of the observation.
func topLevelKeys(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// noChangeAbstention is the "opted in but could not read the payload" outcome,
// carried out to the caller rather than recorded where it is detected.
//
// It exists so the item_type travels WITH the observation. The alternative — the
// caller re-reading item_type from the database to record it — would put two
// sources of truth for one fact inside a single completion, and the recorded row
// could then disagree with the gate that produced it.
type noChangeAbstention struct {
	ItemType string
	Shape    string
}

// recordUnknownNoChangeShape persists an unreadable handler payload to
// agent_error_log, for the same reason recordUnknownVerdict does: a zap.Warn lives
// in an ephemeral pod log that does not survive a rollout, so a guard going blind
// would leave no queryable trace. Severity 'warning' — the completion itself was
// legitimate under the conservative rule; what needs attention is that this gate
// could not read the payload it was asked to judge.
//
// Best-effort by design: failing to record must never block a completion the guard
// has already allowed.
func recordUnknownNoChangeShape(ctx context.Context, params ActionParams, itemID uuid.UUID,
	ab noChangeAbstention, logger *zap.Logger) {

	itemType, shape := ab.ItemType, ab.Shape

	logger.Warn("handlerReportedNoChange: declared counters absent from the handler's result — completing, but this gate could not judge it",
		zap.String("item_id", itemID.String()),
		zap.String("item_type", itemType),
		zap.String("shape", shape),
		zap.String("remedy", "confirm which agent produced this payload, then correct the CounterPaths in noChangeGates (complete_work_item_no_change.go)"))

	if params.DB == nil {
		return
	}

	rule := noChangeGates[itemType]
	LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
		WorkItemID:   itemID.String(),
		Action:       "complete_work_item",
		ErrorMessage: "no-change gate could not read the handler's result for item_type '" + itemType + "' — item completed ungraded by this gate: " + shape,
		ErrorCode:    "NO_CHANGE_GATE_UNREADABLE_RESULT",
		Severity:     "warning",
		Context: map[string]interface{}{
			"item_type":         itemType,
			"guard":             "handlerReportedNoChange",
			"declared_counters": rule.CounterPaths,
			"remedy":            "identify the producing agent, then correct CounterPaths in noChangeGates (complete_work_item_no_change.go); see bugs_open/213 §D, where this payload split is recorded as NOT ESTABLISHED",
		},
	}, logger)
}
