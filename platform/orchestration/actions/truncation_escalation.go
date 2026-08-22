// FILE: platform/orchestration/actions/truncation_escalation.go
//
// ONE ESCALATED RETRY WHEN THE OUTPUT CAP CUT THE RESPONSE (bugs_open/337).
//
// The failure this ends: a step whose configured `ai_service.max_tokens` is
// below what its task genuinely produces fails DETERMINISTICALLY — the provider
// stops at the cap, the transport correctly refuses the fragment (the
// bugs_open/012 lesson, working as intended), and the work-item ladder then
// retries the identical call to max_attempts. bugs_open/337's worked case:
// `component-creator/generate_template` at 16,000 tokens cut the
// `loans-credit-health-check` section at 46,441-48,817 chars NINE times across
// three items — nine full-price generations, zero components, two live pages
// shipped hollow.
//
// The mechanism: a step may declare `ai_service.max_tokens_ceiling`. When a
// call comes back truncated (stop_reason=max_tokens / done_reason=length, via
// the typed aiservice.TruncatedError) AND the ceiling exceeds the cap the cut
// call was actually sent with, execute_llm_prompt retries ONCE with
// max_tokens raised to the ceiling. That is the whole authority: one retry,
// bounded by a number an operator chose, and only on the one error class where
// an identical retry is known-futile but a taller one is not.
//
// WHY THIS IS NOT "RAISE THE CAP AND CALL IT FIXED" (aiservice/truncation.go:
// "whatever the number, the step that writes most approaches it on the work
// most worth doing"). The routine `max_tokens` stays the cost control and stays
// SIZED FROM MEASUREMENT; the ceiling is a bounded second attempt that keeps
// the failure loud if even it is exceeded. Contrast the bugs_open/119 re-ask,
// which deliberately does NOT raise max_tokens: that path serves a JUDGEMENT
// that can be re-asked shorter. A 47k-char document cannot be asked shorter —
// its length IS the work product — so for a writer step the only honest retry
// is a taller one.
//
// OPT-IN, WITH THE UNSAFE SIDE AS THE DEFAULT (owner ruling 2026-08-02 §2,
// RFC_010; the same shape as aiservice/max_tokens.go rule 2). A step whose
// config does not name the key runs the identical code path as before, byte
// for byte. The opt-in signal is the key itself, in the same `ai_service`
// block as the budget it bounds, resolved through the same root->step->runtime
// overlay (bugs_open/009), so a step override behaves like every other
// ai_service key.
//
// INTERPLAY, in the order the error path runs:
//  1. Escalation fires FIRST, before tolerate_truncation: a complete answer
//     at a taller cap beats keeping a fragment. The cut first call gets its
//     own llm_call_log row (success=false, "ESCALATED (bugs_open/337: ...)"
//     prefix) — one call, one forensic row, the rule every caller follows.
//  2. If the escalated call ALSO truncates, the existing machinery sees the
//     SECOND error verbatim: tolerate_truncation (with its bugs_open/076
//     consumer guard) may salvage the taller partial, and the hard-fail
//     message then honestly reports the ceiling as the cap that was hit.
//  3. Any non-truncation error from the escalated call flows down the
//     existing ladder (isAIUnavailable, model errors, 5xx retries) unchanged.
//
// Cost, measured before shipping (llm_call_log, 2026-08-22): the estate-wide
// truncation-failure rate is a handful of calls per week outside the two steps
// this bug names, so the second call is paid almost never — while the failure
// mode it removes cost three full generations per item and the page.
package actions

import (
	"encoding/json"

	"github.com/gqls/agentchassis/platform/aiservice"
)

// truncationEscalationCeiling reads `max_tokens_ceiling` from a resolved
// ai_service config block.
//
// It accepts int, int64, float64 and json.Number for the same reason
// aiservice.configMaxTokens does: the SAME KEY arrives as different Go types
// depending on the path (jsonb decodes to float64, viper-read YAML to int),
// and a float64-only read silently drops the YAML case.
//
// A value must be > 0; zero and negatives are treated as unset. Returns
// (0, false) when the key is absent or unusable, so a caller can distinguish
// "operator chose nothing" (escalation stays off) from "operator chose a
// bound".
func truncationEscalationCeiling(aiServiceConfig map[string]interface{}) (int, bool) {
	if aiServiceConfig == nil {
		return 0, false
	}
	switch v := aiServiceConfig["max_tokens_ceiling"].(type) {
	case int:
		if v > 0 {
			return v, true
		}
	case int64:
		if v > 0 {
			return int(v), true
		}
	case float64:
		if v > 0 {
			return int(v), true
		}
	case json.Number:
		if n, err := v.Int64(); err == nil && n > 0 {
			return int(n), true
		}
	}
	return 0, false
}

// truncationEscalationApplies is the whole escalation decision, pure so it can
// be tested without a provider: escalate exactly when the error is a typed
// truncation AND an operator declared a ceiling AND that ceiling is strictly
// taller than the cap the cut call was actually sent with (`sentMaxTokens`,
// from options["__sent_max_tokens"] — the wire number, not the config's,
// because the two can differ under overlay or the provider fallback).
//
// A ceiling at or below the sent cap refuses to escalate rather than retrying
// at the same height: an identical retry on a deterministic cut is the exact
// waste bugs_open/337 measured (nine cap-hits, zero successes), and refusing
// keeps a misconfiguration (ceiling <= max_tokens) inert instead of making it
// a silent doubled spend.
func truncationEscalationApplies(err error, aiServiceConfig map[string]interface{}, sentMaxTokens int) (int, bool) {
	if _, isTrunc := aiservice.IsTruncated(err); !isTrunc {
		return 0, false
	}
	ceiling, ok := truncationEscalationCeiling(aiServiceConfig)
	if !ok || ceiling <= sentMaxTokens {
		return 0, false
	}
	return ceiling, true
}
