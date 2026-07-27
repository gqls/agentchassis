package handlers

import (
	"log"

	"github.com/gqls/agentchassis/platform/aiservice"
)

// Diagnostic logging for the LLM-backed Gauntlet endpoints (/position, /defend).
//
// WHY THIS FILE EXISTS — bugs_open/083 (gauntlet_engine_503_discards_the_error):
// both handlers previously did
//
//	text, err := client.GenerateText(ctx, prompt, ...)
//	if err != nil {
//	    httperr.JSONError(c, http.StatusServiceUnavailable, "gauntlet judge unavailable")
//	    return   // <-- err discarded: not logged, not wrapped, not counted
//	}
//
// so a 429, a 529, a network timeout, a truncated completion and a malformed
// response all reached the visitor as the same opaque 503, and nothing anywhere
// recorded which had happened. The failure is bursty (measured 2026-07-26: two
// live failures inside one verification window, then 23 consecutive clean calls
// minutes later), so it cannot be reproduced on demand either. That combination —
// no reproduction and no record — is what made it undiagnosable from outside.
//
// These helpers live in one place rather than being hand-copied into each
// handler because two sibling call sites drifting apart is its own defect class
// (see bugs_open/103, where one function composes a description correctly and
// its sibling 100 lines away does not).
//
// Output goes to stdout via the standard library logger, matching
// cmd/tools-api/main.go, so `docker compose logs tools-api` on the island picks
// it up with no new dependency or transport.

// logAIFailure records why an LLM call failed, immediately before the handler
// returns 503.
//
// Truncation is reported distinctly rather than folded into err: it is the one
// failure here that is NOT an upstream fault but our own configured cap, and it
// needs a different fix (raise max_tokens for that call), so the log must let
// you tell it apart without string-matching an error message.
//
// NOTE: 083 §2 records the truncation theory as NOT fitting the evidence —
// successful responses measured ~373 output tokens against a 2048 default. This
// branch is here to settle that question with data, not because it is expected
// to fire. If it never fires, that is itself the finding.
func logAIFailure(endpoint, stage, roundID string, err error) {
	if te, ok := aiservice.IsTruncated(err); ok {
		log.Printf("gauntlet/%s: %s TRUNCATED round_id=%s provider=%s reason=%s output_tokens=%d partial_chars=%d",
			endpoint, stage, roundID, te.Provider, te.Reason, te.OutputTokens, len(te.Partial))
		return
	}
	log.Printf("gauntlet/%s: %s FAILED round_id=%s err=%v", endpoint, stage, roundID, err)
}

// logInternalFailure records a non-LLM server-side failure — a DB lookup, a
// persist, a marshal — immediately before the handler returns 5xx.
//
// Separate from logAIFailure because the truncation check is meaningless for a
// database error, and a helper whose name claims "AI" would misdescribe the call
// site. These paths discarded their errors exactly as the LLM ones did; they are
// 500s rather than 503s, so they sit outside bugs_open/083's title, but they are
// the same defect and the same one-line remedy.
//
// The 400-class returns are deliberately NOT logged: those are caller mistakes,
// not faults, and gin.Logger() already records method, path and status for every
// request, so logging them again would add noise without adding a fact.
func logInternalFailure(endpoint, stage, roundID string, err error) {
	log.Printf("gauntlet/%s: %s FAILED round_id=%s err=%v", endpoint, stage, roundID, err)
}

// logAIBadResponse records a response that arrived intact but could not be used
// — malformed JSON, or valid JSON with the required fields empty.
//
// It records the response's SHAPE, never its text. "Unparseable" alone cannot
// distinguish a model that wrapped its JSON in prose from one that emitted two
// objects (the bugs_closed/088 class) from a genuinely empty completion, and
// those have different fixes — but every one of those questions is structural,
// so aiservice.Fingerprint answers all of them without reproducing a single
// character of what the model wrote, or of what the visitor wrote and the model
// quoted back.
//
// This deliberately supersedes an earlier capped-excerpt version. Council corr
// e004fd81 approved the fix but recorded that logging model text "cannot be
// closed by this council alone"; the owner ruled for a fingerprint on
// 2026-07-27. It is also strictly MORE diagnostic than the excerpt was: a
// 300-char cap could never reveal 088's second object, which begins ~1,500
// chars in, so the excerpt could not detect the case that justified it.
func logAIBadResponse(endpoint, reason, roundID, body string) {
	log.Printf("gauntlet/%s: response UNUSABLE round_id=%s reason=%s %s",
		endpoint, roundID, reason, aiservice.Fingerprint(body))
}
