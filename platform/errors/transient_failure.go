// FILE: platform/errors/transient_failure.go
//
// The RETRYABLE arm of failure classification, beside its permanent twin in
// permanent_failure.go (bugs_open/197; the twin's history is bugs 034 → 195).
// RELOCATED here from platform/messaging/retryable_transient.go on 2026-08-08
// (bugs_open/217) — see permanent_failure.go's header for why: the coordinator
// needs RetryDisposition and cannot import messaging (cycle), so the seam moved
// to this stdlib-only leaf; messaging re-exports every moved name, so there is
// still exactly ONE implementation and its callers and pins are unchanged.
// Everything below travelled verbatim, with only package-local qualifiers and
// file-name pointers adjusted.
//
// Every processing failure is classified twice, in sequence: first
// MatchedPermanentFailure asks "is this statically broken — drop it, record
// it, never retry?"; anything not permanent then reaches this file's
// question, "is this transient — can a retry succeed?". The answer becomes
// the wire status (error_recoverable / error_unrecoverable), which the
// coordinator routes to a re-dispatch or to terminal failure. One function
// decides retry-vs-terminal for every agent in the fleet, which is why it
// lives at the shared seam, and not as a private helper in the layer
// that happens to call it today — two lists for one judgement is the drift
// bugs_closed/034 closed, and this file replaces exactly such a pair:
// agentbase's private isRecoverableError (three case-sensitive needles, no
// nil guard) and this package's own IsRecoverable (twelve case-folded
// patterns, ZERO callers — dead, and disagreeing with its live twin on
// case). Both were deleted in the commit that added this.
//
// THE CENSUS THIS LIST IS BUILT FROM (2026-08-06, agent_error_log rows with
// error_code='PROCESSING_FAILED' or classification='transient' — precisely
// the population that reaches this seam, courtesy of bugs_closed/195's
// unconditional recorder):
//
//	population 2,996 · old verdicts: 570 recoverable / 2,426 unrecoverable
//	885 rows contain "deadline exceeded" and were made TERMINAL — transient
//	    by definition; ~30% of the population lost work a retry could save
//	882 rows carried capitalised Timeout/Connection/Temporary and missed on
//	    case alone — bugs_closed/195's capital-letter defect, on the sibling
//	rate-limit/5xx prose: 37 rows, 20 missed
//
// WHY THE FALLBACK IS CASE-FOLDED HERE WHILE THE PERMANENT TWIN'S NEEDLES
// STAY BYTE-PINNED — the costs are asymmetric. A permanent-side over-match
// drops retryable work forever (unbounded loss, so those needles never
// fold). A transient-side over-match costs at most three futile
// redispatches and then the terminal arm anyway: the coordinator caps at
// retry_version >= 3 (coordinator.go, handleRecoverableError), and adapter
// re-executions are capped independently via execution_metadata.retry_count
// (bugs_open/075). Folding buys the 882 case-missed rows for that bounded
// downside. Each needle below is still individually judged against "can
// retrying this succeed?" — the fold is not a shortcut applied blind.
//
// KNOWN, ACCEPTED OVER-MATCHES (measured, not hypothetical):
//   - "connection" also matches "…establish a secure connection…" inside
//     SSL/TLS *certificate* errors (live rows quoted in the census) — a
//     broken TLS config is not healed by retrying, so those rows buy three
//     futile attempts before terminal. Kept because "connection" is the
//     live needle carrying genuine refused/reset recoveries, and the cost
//     is bounded.
//   - "dial tcp" admits NXDOMAIN from a typo'd hostname to the same three
//     futile retries. Accepted for the same reason.
//
// REJECTED NEEDLES, and the reasons are load-bearing — do not "complete"
// this list from the deleted IsRecoverable's without re-arguing them:
//   - bare "502"/"503"/"504": the dominant message shape in the census is
//     nested "workflow failed: Request <uuid> failed: …", and UUIDs are
//     hex — "503" is a valid hex substring, so a bare-digit needle would
//     classify an arbitrary permanent failure as retryable off its own
//     correlation id. Pinned OUT by a test.
//   - "unreachable", "network": ZERO rows in the census population match
//     either (measured 2026-08-06). A needle with no observed input is
//     policy nobody asked for; admit them when evidence arrives.

package errors

import "strings"

// RetryableTransientCodes are the typed DomainError codes that mark a
// failure transient by the CODE'S OWN MEANING: another attempt can succeed
// without anything being fixed first. Closed and deliberately small — this
// branch is future-proofing (nothing in production constructs these at this
// seam today), so every entry must be defensible alone:
//   - ErrAgentTimeout: the attempt ran out of time; a fresh attempt has a
//     fresh budget.
//   - ErrAgentOverloaded: capacity returns; HTTPStatus already maps it 503.
//   - ErrRateLimited: succeeds after backoff; RetryAfter exists for this.
//
// NOT admitted: ErrExternalService / ErrAIServiceError (a 401 or bad key
// renders under both, and retrying cannot fix a wrong credential);
// ErrWorkflowTimeout (the overall budget is spent — re-running the step
// re-spends it).
var RetryableTransientCodes = []ErrorCode{
	ErrAgentTimeout,
	ErrAgentOverloaded,
	ErrRateLimited,
}

// TransientErrorNeedles is the substring fallback for errors that carry no
// DomainError. Lowercase — matchedTransientNeedle folds the message once —
// and first-match-wins, so order matters only for which needle is reported.
// Every entry is judged in this file's header; change the list only through
// the pinning test beside it (the pin lives in platform/messaging, which
// aliases this very slice — same backing array, one list).
var TransientErrorNeedles = []string{
	"deadline exceeded", // context.DeadlineExceeded prose — the census's 885-row headline
	"timeout",           // i/o timeout, gateway timeout, request timeout
	"connection",        // refused/reset/closed (accepts the SSL over-match above)
	"temporary",         // says so on the tin
	"dial tcp",          // Go net dial failures (accepts the NXDOMAIN over-match above)
	"too many requests", // 429 prose
	"rate limit",        // prefix of "rate limited" too
	"service unavailable",
	"bad gateway",
}

// MatchedTransientFailure classifies err as a transient failure — one a
// retry can plausibly cure — and returns an audit token naming WHAT
// matched, or "" for "not transient" (the caller's terminal arm).
//
// The token convention is inherited from MatchedPermanentFailure and is the
// point: returning WHICH thing matched, not a bool, is what makes an
// unanchored classification auditable after the fact, and the prefix says
// which branch answered — "retryable:<code>" for the author's explicit
// AsRetryable intent, "code:<code>" for the typed list, a bare needle for
// the prose fallback.
//
// Order: the author's explicit Retryable flag outranks everything; the
// typed code list is exact and survives rewording, capitalisation and %w
// wrapping; prose is only consulted for errors carrying no DomainError —
// with one deliberate exception. A DomainError with Retryable=false and a
// code NOT on the list still falls through to the needles, because
// InternalError() defaults Retryable=false while routinely WRAPPING
// a transient cause — classifying on the flag alone would make every
// wrapped timeout terminal, which is this bug re-created one level up
// (mirrors NonRetryablePermanentCodes' argument in permanent_failure.go).
func MatchedTransientFailure(err error) string {
	if err == nil {
		return ""
	}
	if de, ok := AsDomainError(err); ok {
		if de.Retryable {
			return "retryable:" + string(de.Code)
		}
		for _, code := range RetryableTransientCodes {
			if de.Code == code {
				return "code:" + string(code)
			}
		}
	}
	return matchedTransientNeedle(err.Error())
}

// RetryDisposition applies the two-stage classification in its documented
// order — permanent first (MatchedPermanentFailure: drop, record, never
// retry), then transient (MatchedTransientFailure: a retry can plausibly
// cure), else terminal — and returns the wire disposition plus the audit
// token naming what decided it ("permanent:<token>", a transient token,
// or "" for unclassified-terminal).
//
// The ORDER is the contract, and it is why this exists as one function
// rather than as guidance (bugs_open/207). The two questions are asked from
// call sites that cannot all guarantee the sequence themselves:
// handleError's validation branch sends its response through the same
// sender as the transient branch, and a message like "invalid connection"
// (a real DB-driver fault, quoted in permanent_failure.go's header) carries
// BOTH a permanent and a transient needle. Consulting the transient list
// alone would have this seam tell the coordinator error_recoverable for a
// failure the processor has just dropped as validation-never-retry —
// two contradictory dispositions for one failure. Permanent-first makes
// that unrepresentable, and matches the guard agentbase enforces by call
// order (processMessage classifies permanent and drops before
// handleProcessingError ever runs).
func RetryDisposition(err error) (recoverable bool, matched string) {
	if match := MatchedPermanentFailure(err); match != "" {
		return false, "permanent:" + match
	}
	if match := MatchedTransientFailure(err); match != "" {
		return true, match
	}
	return false, ""
}

// matchedTransientNeedle returns the first transient needle contained in
// the case-folded message, or "".
func matchedTransientNeedle(msg string) string {
	folded := strings.ToLower(msg)
	for _, needle := range TransientErrorNeedles {
		if strings.Contains(folded, needle) {
			return needle
		}
	}
	return ""
}
