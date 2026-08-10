// FILE: platform/errors/permanent_failure.go
//
// The PERMANENT arm of failure classification, beside its transient twin in
// transient_failure.go. RELOCATED here from platform/messaging/validation_drop.go
// on 2026-08-08 (bugs_open/217): the coordinator's child-orchestration failure
// sender needs RetryDisposition, and platform/messaging imports
// platform/orchestration (the drop recorder calls orchestration.LogAgentError),
// so the classifier had to move to a package both layers can import. This
// package is a stdlib-only leaf and already owns DomainError and the typed
// codes these classifiers read. platform/messaging re-exports every name moved
// here, so its callers and pinning tests are unchanged — there is still exactly
// ONE implementation. The drop RECORDER stayed in messaging; only the
// classification core moved. Everything below this paragraph travelled
// verbatim, with only package-local qualifiers adjusted.
//
// The "validation error" drop path this classifier serves is documented where
// the drop happens (messaging/validation_drop.go, bugs_open/034): two layers
// independently decide that a failed message is a validation error and must
// NOT be retried — MessageProcessor.handleError and agentbase.Agent.processMessage.
// Both decided it by substring match against two lists that had silently
// drifted apart; this ONE list is what ended that.

package errors

import "strings"

// ValidationErrorNeedles are the substrings that mark a processing failure as a
// non-retryable validation error. This is the single source of truth for both
// the messaging and agentbase layers.
//
// It is a substring match against the WHOLE error string, unanchored to any
// validation origin, and that is a known defect (bugs_open/034 fix candidate 2,
// not yet done): "invalid" is also a substring of "invalid character 'w' after
// object key:value pair" (a truncated-LLM parse failure), "invalid connection"
// (a database driver fault), "invalid memory address" (a recovered nil deref)
// and "x509: … invalid" (a TLS failure). Each of those is a real runtime error
// that this list silently reclassifies as "the caller sent us rubbish", and then
// drops without a retry.
//
// Replacing it with a typed sentinel (errors.Is(err, ErrValidation)) is the
// structural fix and is deliberately NOT attempted here — it needs every error
// construction site audited, and doing it blind would change retry behaviour
// fleet-wide. Until then the mitigation is visibility, not correctness: every
// drop records WHICH needle fired, so a misclassification is a queryable row
// rather than an invisible one.
//
// > **bugs_open/195 — that deferred structural fix is now DONE, and the reason
// > it could not wait is that this list missed its own commonest input.** The
// > fleet's most frequent permanent config error renders as
// > "WORKFLOW_INVALID: Invalid workflow configuration (caused by: …)", which
// > matches NONE of the four needles: "is required" loses to the wording
// > ("requires a topic"), and "invalid" loses to the CAPITAL I, because the
// > match is case-sensitive. So the error was classified transient, and
// > recordDroppedValidationError — the durable record 034 built — was never
// > reached, because the branch that calls it was never entered. See
// > MatchedPermanentFailure below: typed first, this list only as a fallback
// > for untyped errors. The needles are deliberately NOT case-folded, which
// > would widen every hazard listed above to its capitalised variants too.
//
// Ordering matters only for which needle gets reported when several match.
var ValidationErrorNeedles = []string{"is required", "validation", "invalid", "missing"}

// NonRetryablePermanentCodes are the typed DomainError codes that mark a
// processing failure PERMANENT: drop it, record it durably, never retry it.
//
// Closed and explicit, deliberately. The tempting shortcut — "classify on
// DomainError.Retryable" — is far too wide: InternalError() sets
// Retryable false, so every genuinely transient internal fault would start
// being dropped fleet-wide. A code earns its place here only when retrying it
// can never succeed because the input is statically wrong.
//
// ErrDispatchUnresolvable earns its place on exactly that test (bugs_open/239):
// a request body that is not JSON, or one naming an agent type with no active
// definition, is statically wrong — the same bytes against a healthy fleet fail
// identically for ever. Its transient twin, ErrDispatchLookupUnavailable, is
// deliberately NOT here and is not on the transient list either: it is built
// AsRetryable at the refusal site, and the author's-intent early return in
// MatchedPermanentFailure answers before any list is consulted.
var NonRetryablePermanentCodes = []ErrorCode{
	ErrWorkflowInvalid,
	ErrValidation,
	ErrDispatchUnresolvable,
}

// MatchedPermanentFailure classifies err as a permanent failure and returns an
// audit token naming WHAT matched, or "" for "not permanent".
//
// The token is the point, and it is inherited from MatchedValidationNeedle
// below: returning WHICH thing matched rather than a bool is what makes an
// unanchored classification auditable after the fact. Typed matches report
// "code:WORKFLOW_INVALID"; the legacy substring fallback reports the bare
// needle, so the two are distinguishable in a query.
//
// Order matters: the typed code is exact and cannot be defeated by rewording,
// capitalisation, or %w wrapping, so it is tried first. The substring fallback
// remains only for errors that carry no DomainError at all.
//
// An error explicitly built AsRetryable is never permanent, whatever its code
// and whatever its prose says — an author who wrote "retry this" outranks both
// the list and the substring fallback. That early return is load-bearing, not
// tidiness: without it an AsRetryable(ErrValidation) error skips the typed
// branch and is then classified permanent by the fallback matching the word
// "validation" in its own message. Structure must not be overridden by prose;
// being overridden by prose is the entire bug this function exists to fix.
//
// Note what is deliberately NOT done: a non-retryable DomainError whose code is
// not on the list still falls through to the substring fallback. Suppressing
// that would be more principled — the error has already declared its type — but
// it would move existing drops back to retry fleet-wide, which is exactly the
// blast radius 034's lane deferred this work to avoid.
//
// So this change ADDS permanent classifications, with exactly ONE removal: the
// de.Retryable early return above. An AsRetryable DomainError whose message
// happens to contain a needle ("validation failed") was classified permanent by
// the old substring-only path and is not any more.
//
// > **CORRECTED 2026-08-05 (code-review F1).** This paragraph used to end "This
// > change only ever ADDS permanent classifications; it removes none", which
// > contradicted the paragraph six lines above that argues FOR the removal. The
// > code is right and unchanged; the claim was wrong. It is latent today —
// > ErrorBuilder.AsRetryable (platform/errors/errors.go:171) and the field write
// > inside it are the only ways to set Retryable=true, and a fleet-wide grep
// > finds no caller of either outside this comment and the tests. The first
// > producer inherits the behaviour, which is why the two must agree now.
func MatchedPermanentFailure(err error) string {
	if err == nil {
		return ""
	}
	if de, ok := AsDomainError(err); ok {
		if de.Retryable {
			return ""
		}
		for _, code := range NonRetryablePermanentCodes {
			if de.Code == code {
				return "code:" + string(code)
			}
		}
	}
	return MatchedValidationNeedle(err.Error())
}

// MatchedValidationNeedle returns the first needle contained in errMsg, or ""
// if none match. Returning the needle rather than a bool is the point: it is
// what makes an unanchored match auditable after the fact.
func MatchedValidationNeedle(errMsg string) string {
	for _, needle := range ValidationErrorNeedles {
		if strings.Contains(errMsg, needle) {
			return needle
		}
	}
	return ""
}
