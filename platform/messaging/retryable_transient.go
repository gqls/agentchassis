// FILE: platform/messaging/retryable_transient.go
//
// The RETRYABLE arm of failure classification, beside its permanent twin in
// validation_drop.go (bugs_open/197; the twin's history is bugs 034 → 195).
//
// THE CLASSIFICATION CORE NOW LIVES IN platform/errors (transient_failure.go),
// moved 2026-08-08 (bugs_open/217) — see validation_drop.go's header for why.
// The names below are re-exports aliasing that one implementation: same backing
// slice, same functions, one list. The census, the case-fold argument, the
// accepted over-matches and the rejected needles all travelled with the code —
// read them THERE before changing anything, and change the list only through
// the pinning test beside these aliases (it pins the shared slice itself).
//
// Every processing failure is classified twice, in sequence: first
// MatchedPermanentFailure asks "is this statically broken — drop it, record
// it, never retry?"; anything not permanent then reaches the transient
// question, "can a retry succeed?". The answer becomes the wire status
// (error_recoverable / error_unrecoverable), which the coordinator routes to
// a re-dispatch or to terminal failure.

package messaging

import (
	"github.com/gqls/agentchassis/platform/errors"
)

// RetryableTransientCodes re-exports the typed transient code list. The
// admission arguments live with the list: platform/errors/transient_failure.go.
var RetryableTransientCodes = errors.RetryableTransientCodes

// transientErrorNeedles aliases the shared prose-fallback list so the
// list-order pinning test in this package keeps pinning the real thing.
var transientErrorNeedles = errors.TransientErrorNeedles

// MatchedTransientFailure delegates to the shared classifier in
// platform/errors — explicit Retryable flag first, typed code list second,
// case-folded prose fallback last.
func MatchedTransientFailure(err error) string {
	return errors.MatchedTransientFailure(err)
}

// RetryDisposition delegates to the shared sequenced classifier — permanent
// first, then transient, else terminal. The ORDER is the contract
// (bugs_open/207); see transient_failure.go for why it exists as one function
// rather than as guidance.
func RetryDisposition(err error) (recoverable bool, matched string) {
	return errors.RetryDisposition(err)
}
