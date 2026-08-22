// FILE: platform/aiservice/truncation.go
//
// TruncatedError — a completion that was CUT at the provider's output cap.
//
// The platform already detected truncation correctly (stop_reason=max_tokens /
// done_reason=length) and returned a hard error. What it did NOT do was keep the
// partial text: the provider clients returned "" alongside the error, so the
// bytes the model had already produced were destroyed at the transport layer and
// no caller could ever recover from a truncation, however recoverable it was.
//
// bugs_open/019 is what that costs. A council seat that overruns its cap fails
// its step, the step's error_step routes the whole round to complete_invalid, and
// every other seat's review is lost — for the councils, review_editquality runs
// first, so in practice the round dies before any seat is read. Nine such runs in
// ten days. The case file proposed salvaging the truncated review with
// repairTruncatedJSON; that was impossible as written, because the partial never
// left the provider client.
//
// So: carry the partial WITH the error. This is deliberately additive —
// GenerateText's signature is unchanged and every existing caller does
// `if err != nil { ... }` and is unaffected. Only a caller that explicitly asks
// (errors.As for *TruncatedError) can salvage, so tolerating a truncation stays
// an opt-in decision made at the step that understands the consequences, never a
// silent platform-wide downgrade of an error into a success.
//
// NOTE this is NOT a reason to raise output caps and call the class fixed:
// experience-planner/compose truncated at a 32,000-token cap, 4x the council
// seats' 8,000. Whatever the number, the step that writes most approaches it on
// the work most worth doing. Losing the round is the part with no upside.
package aiservice

import (
	"errors"
	"fmt"
)

// TruncatedError reports a completion cut short by the provider's output cap,
// carrying whatever text the model had produced before it was cut.
//
// Partial may be empty — a cap small enough to truncate before the first content
// block leaves nothing to salvage — so callers must handle that case rather than
// assume a non-empty partial.
type TruncatedError struct {
	// Partial is the text produced before the cut. Possibly empty; never valid
	// JSON in the general case, since it is by definition an unfinished document.
	Partial string
	// OutputTokens is what the provider reported producing (0 if unreported).
	OutputTokens int
	// Reason is the provider's own stop signal, kept verbatim for the log
	// ("stop_reason=max_tokens", "done_reason=length").
	Reason string
	// Provider names which client produced this, so a mixed-provider deployment
	// can tell where a truncation came from.
	Provider string
}

func (e *TruncatedError) Error() string {
	return fmt.Sprintf("response truncated: %s (output_tokens=%d reached the configured cap, %d chars recovered); raise max_tokens or shorten the prompt",
		e.Reason, e.OutputTokens, len(e.Partial))
}

// IsTruncated reports whether err is (or wraps) a TruncatedError, and returns it
// so the caller can reach the partial. Use this rather than string-matching the
// message: the wording has already changed once, and ai_errors.go's substring
// matching on error text is the fragility this is meant to avoid inheriting.
func IsTruncated(err error) (*TruncatedError, bool) {
	var te *TruncatedError
	if errors.As(err, &te) {
		return te, true
	}
	return nil, false
}

// ── THE USAGE-COUNT BACKSTOP, AND WHY IT IS A THREE-STATE ────────────────────
//
// Everything above is the STRUCTURAL signal: the provider says it stopped at the
// cap, we raise a typed error, and a caller doing `if err != nil` is already
// protected without arithmetic. Prefer it. This block is for the two cases where
// that signal is not available:
//
//  1. a provider that returns a cut answer as a 200 with no error (the reason
//     rewrite_negations keeps a backstop at all — bugs_open/305, council round 2,
//     llm_reliability seat, HIGH);
//  2. reading llm_call_log AFTER the fact, where there is no error object at all,
//     only two integers.
//
// CLAUDE.md's rule is "output_tokens == max_tokens means the completion was CUT".
// Written as a bare boolean — `outTok >= maxTok` — that rule has a hole big
// enough to lose the whole check through: when usage is UNREPORTED, outTok is 0,
// the comparison is FALSE, and "we could not tell" is silently returned as "the
// answer was complete". The two are not the same claim and must not share a
// return value.
//
// Hence three states, and hence the ordering: **TruncationUnknown is the zero
// value**, so a caller that forgets a case, or reads a zeroed struct, gets the
// state that forces a decision rather than the one that says "fine".

// TruncationState is what the usage counts can tell you about whether a
// completion was cut — including that they can tell you nothing.
type TruncationState int

const (
	// TruncationUnknown means the counts cannot answer the question: no ceiling
	// was recorded, or the provider reported no output usage. NOT "fine" — the
	// caller must decide what to do with a completion it cannot vouch for.
	// Deliberately the zero value.
	TruncationUnknown TruncationState = iota
	// TruncationBelow means usage was reported and finished below the ceiling.
	TruncationBelow
	// TruncationAtCeiling means usage was reported and reached or passed the
	// ceiling: the completion was CUT, whatever the text looks like.
	TruncationAtCeiling
)

func (s TruncationState) String() string {
	switch s {
	case TruncationBelow:
		return "below_ceiling"
	case TruncationAtCeiling:
		return "at_ceiling"
	default:
		return "unknown"
	}
}

// Truncated reports whether this state is a known cut. It is deliberately NOT
// the negation of "fine": TruncationUnknown returns false here AND false from
// Complete, because neither claim is available.
func (s TruncationState) Truncated() bool { return s == TruncationAtCeiling }

// Complete reports whether the completion is known to have finished below the
// ceiling. Unknown returns false — see Truncated.
func (s TruncationState) Complete() bool { return s == TruncationBelow }

// ClassifyTruncation turns a reported (outputTokens, maxTokens) pair into the
// three states above. Both are taken as plain ints because that is what every
// caller has: options["__usage_output_tokens"] in flight, and
// llm_call_log.output_tokens / max_tokens after the fact.
//
// A zero or negative maxTokens means no ceiling was recorded, and a zero or
// negative outputTokens means usage was not reported — 0 tokens produced is not
// a distinguishable answer, and treating it as "below the ceiling" is exactly
// the hole this exists to close.
func ClassifyTruncation(outputTokens, maxTokens int) TruncationState {
	if maxTokens <= 0 || outputTokens <= 0 {
		return TruncationUnknown
	}
	if outputTokens >= maxTokens {
		return TruncationAtCeiling
	}
	return TruncationBelow
}
