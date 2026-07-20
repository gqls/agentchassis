// FILE: platform/aiservice/refusal.go
//
// RefusalError — a completion the model DECLINED to produce.
//
// This is the last open item of bugs_open/008 (item 5, filed "optional" and
// since promoted by a production case). The Messages API signals a decline with
// `stop_reason: "refusal"` and a content array carrying no text block. The
// Anthropic client decoded neither, so the response fell through every branch to
// the terminal fallback and surfaced as:
//
//	AI call failed with unhandled error: no text content in response (had 1 blocks)
//
// That message names no cause. It reads like a transport fault or a parse bug,
// which is what the next person spends their time on. What actually happened is
// that the model understood the request and said no.
//
// Two consequences follow from a refusal that a truncation does not share, and
// both are why this is a distinct type rather than a better error string:
//
//   - It is NOT retryable. An identical prompt is refused identically, so a
//     retry loop spends its whole budget proving that. TruncatedError has the
//     same property for a different reason (the cap does not move between
//     attempts) — a caller that wants to classify "do not retry" can now do it
//     by type for both, instead of substring-matching prose that has already
//     changed once.
//   - The fix is in the PROMPT, never in the platform. A truncation is answered
//     by raising a cap or shortening the input; a refusal is answered by asking
//     for something else. Erroring without saying so sends the reader looking in
//     the wrong layer, which is precisely what happened on 2026-07-18: the
//     refusal killed an experience-planner council round outright and the
//     recorded symptom pointed at the response parser.
//
// Blocks is carried because a single non-text block is the diagnostic tell, and
// "had 1 blocks" — a count with no types — was the part of the old message that
// wasted the most time. A thinking-only response and a refusal both produce one
// block and mean entirely different things.
package aiservice

import (
	"errors"
	"fmt"
	"strings"
)

// RefusalError reports a completion the model declined to produce, naming the
// provider's own stop signal and the content block types it did return.
//
// There is no partial to carry, unlike TruncatedError: a refusal produces no
// answer text at all, so there is nothing to salvage and no opt-in tolerance to
// offer. A caller that can proceed without this particular completion (an
// advisory council seat, say) should treat the error as an abstention rather
// than attempt recovery.
type RefusalError struct {
	// Reason is the provider's stop signal, verbatim ("stop_reason=refusal").
	Reason string
	// Blocks lists the content block types the response carried, in order.
	// Empty means the content array itself was empty.
	Blocks []string
	// Provider names which client produced this.
	Provider string
	// Model is the model that declined, so a fleet-wide refusal pattern on one
	// model is visible without cross-referencing the call log.
	Model string
}

func (e *RefusalError) Error() string {
	blocks := "no content blocks"
	if len(e.Blocks) > 0 {
		blocks = fmt.Sprintf("%d content block(s): %s", len(e.Blocks), strings.Join(e.Blocks, ", "))
	}
	return fmt.Sprintf("model declined to answer: %s (provider=%s, model=%s, %s); change the prompt — an identical retry is refused identically",
		e.Reason, e.Provider, e.Model, blocks)
}

// IsRefusal reports whether err is (or wraps) a RefusalError, and returns it so
// the caller can reach the stop signal and block types. Use this rather than
// string-matching the message, for the same reason IsTruncated exists.
func IsRefusal(err error) (*RefusalError, bool) {
	var re *RefusalError
	if errors.As(err, &re) {
		return re, true
	}
	return nil, false
}
