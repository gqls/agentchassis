// FILE: platform/orchestration/datahelpers/claims_banned_pattern_escaping_test.go
//
// A banned_claims pattern that is DOUBLE-ESCAPED compiles cleanly and matches
// nothing — for ever, silently, while every structural check on it passes.
//
// claims.go:284-289 compiles each pattern with regexp.Compile("(?i)"+p) and
// falls back to regexp.QuoteMeta only when compilation ERRORS. `\\bguaranteed`
// (a literal backslash, then 'b') is a PERFECTLY VALID regex — it just never
// matches English prose — so it compiles, the fallback never fires, and the
// guard is loaded, listed and counted while being inert.
//
// This happened for real on 2026-08-17: the remortgagecalculator.uk pilot seed
// wrote its patterns inside a dollar-quoted SQL string (bytes passed
// LITERALLY) that is then parsed as JSON (which DOES unescape), so `\\\\b` in
// the file stored `\\b`. All six patterns were inert. The seed's own verify
// block asserted `jsonb_array_length(banned_claims) = 6` and passed — a count
// comes out identical whether the guards work or not.
//
// So this test asserts BEHAVIOUR, not shape: strings that must be caught, and
// one that must not. It is the authoritative check for that site's patterns,
// because the seeding layer cannot run it (Postgres regex and Go RE2 disagree
// on \b — PG ARE spells word boundary \y, and \b is backspace — so a SQL probe
// is a check in the wrong engine).
package datahelpers

import (
	"regexp"
	"testing"
)

// compileLikeProduction mirrors loadEvidenceBase's compile step exactly,
// including the "(?i)" prefix and the QuoteMeta fallback. If claims.go changes
// how it compiles, this test should be updated to match it — the point is to
// test what production does, never what we wish it did.
func compileLikeProduction(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(pattern))
	}
	return re
}

// remortgagePilotBannedPatterns are the patterns as they must be STORED (single
// backslash for a word boundary, a real £). Kept in sync by hand with
// portfolio_positioning/SEED_2026-08-17b_…_fix_banned_claim_escaping.sql; the
// SQL file's own verify block asserts the stored rows carry no double
// backslash, so the two halves fail in different ways and cannot both drift
// silently.
var remortgagePilotBannedPatterns = []string{
	`\bguaranteed (acceptance|approval|rate|saving)\b`,
	`\b(best|cheapest|lowest) (rate|deal)s? (available|on the market|in the uk)\b`,
	`\bsave (up to )?£[0-9,]+`,
	`\b(we are|we're) (fca|financially) (regulated|authorised)\b`,
	`\b(you (will|would) save|this will save you)\b`,
	`\b[0-9]+(\.[0-9]+)?% (apr|apcr|rate)\b`,
}

func TestRemortgagePilotBannedClaims_ActuallyMatch(t *testing.T) {
	res := make([]*regexp.Regexp, 0, len(remortgagePilotBannedPatterns))
	for _, p := range remortgagePilotBannedPatterns {
		res = append(res, compileLikeProduction(t, p))
	}

	anyMatch := func(s string) bool {
		for _, re := range res {
			if re.MatchString(s) {
				return true
			}
		}
		return false
	}

	mustCatch := []struct{ name, text string }{
		{"guarantee", "We offer guaranteed acceptance for all applicants"},
		{"market superlative", "The cheapest rates available in the UK today"},
		{"saving figure", "Save up to £4,200 a year by remortgaging"},
		{"implies regulated", "We are FCA regulated and here to help"},
		{"outcome promise", "You will save money by switching now"},
		{"literal rate in prose", "Fixed at 4.29% APR for two years"},
	}
	for _, tc := range mustCatch {
		t.Run("catches/"+tc.name, func(t *testing.T) {
			if !anyMatch(tc.text) {
				t.Fatalf("no banned pattern matched %q — this is the double-escaping "+
					"failure mode: the patterns compile and match nothing", tc.text)
			}
		})
	}

	// The negative side matters as much: a guard that matches everything is as
	// useless as one that matches nothing, and only the pair distinguishes a
	// working regex from a broken one.
	mustAllow := []struct{ name, text string }{
		{"plain guidance", "Your fixed rate ends soon; check your lender's terms"},
		{"hypothetical example", "If your balance were £200,000 over 20 years, the tool shows the difference"},
		{"non-price lender fact", "Nationwide Building Society is authorised by the Prudential Regulation Authority"},
		{"uncertainty", "This depends on your lender and we have not verified it"},
	}
	for _, tc := range mustAllow {
		t.Run("allows/"+tc.name, func(t *testing.T) {
			if anyMatch(tc.text) {
				t.Fatalf("a banned pattern matched legitimate copy %q — over-broad guard", tc.text)
			}
		})
	}
}

// TestDoubleEscapedPatternIsSilentlyInert pins the MECHANISM, so the next
// person who sees a passing count on a seeded evidence_base knows what a count
// cannot tell them. It asserts the failure mode directly: the corrupted form
// compiles without error (so the QuoteMeta fallback never fires) and matches
// the text it was written to catch — not at all.
func TestDoubleEscapedPatternIsSilentlyInert(t *testing.T) {
	good := `\bguaranteed (acceptance|approval)\b`
	bad := `\\bguaranteed (acceptance|approval)\\b` // what double-escaping stores
	text := "We offer guaranteed acceptance for all applicants"

	if _, err := regexp.Compile("(?i)" + bad); err != nil {
		t.Fatalf("expected the corrupted pattern to COMPILE (that is why it is silent), got: %v", err)
	}
	if !compileLikeProduction(t, good).MatchString(text) {
		t.Fatal("the correct pattern must match the banned text")
	}
	if compileLikeProduction(t, bad).MatchString(text) {
		t.Fatal("the corrupted pattern unexpectedly matched — the premise of this test is wrong, re-derive it")
	}
}
