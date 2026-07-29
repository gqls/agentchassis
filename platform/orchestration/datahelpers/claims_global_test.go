// FILE: platform/orchestration/datahelpers/claims_global_test.go
//
// Regression suite for the fleet-wide banned-claim set (bugs_open/104).
//
// The pass-list here is the point. The oufe pattern set was tested carefully
// before it went live — 10 fabrication shapes blocked, 13 legitimate sentences
// passed — and it still carried a false-positive class, because no test sentence
// NEGATED one of its own patterns. Every negated fixture below is real copy taken
// from a live site during the 2026-07-28 fleet dry run, and each one would have
// failed a page build.
//
// The fixtures are VERBATIM, extracted from the components' stored rendered_html
// rather than retyped from claimscan's output — its snippets are truncated with
// ellipses, and reconstructing a sentence from one produces a plausible quote
// that the site never published. Two of these were wrong that way before being
// checked against the source.

package datahelpers

import (
	"regexp"
	"strings"
	"testing"
)

func globalFindings(t *testing.T, sentence string) []ClaimFinding {
	t.Helper()
	return ScanAllBannedClaims(ExtractAssertionText("<p>"+sentence+"</p>"), nil)
}

// ---------------------------------------------------------------------------
// The set must be wired at all. A silently empty global set and a working one
// are indistinguishable from every other test in this file.
// ---------------------------------------------------------------------------

func TestGlobalSetIsWired(t *testing.T) {
	if n := GlobalBannedClaimCount(); n != 10 {
		t.Fatalf("expected 10 fleet-wide patterns, got %d — if this changed deliberately, "+
			"re-run the fleet dry run (see claims_global.go header) before editing this number", n)
	}
}

// ---------------------------------------------------------------------------
// MUST BLOCK — one canonical fabrication per pattern.
// ---------------------------------------------------------------------------

func TestGlobalBlocksOverclaims(t *testing.T) {
	mustBlock := []string{
		"A claim without a source does not appear here.",
		"Prices do not appear here.",
		"If we cannot verify a figure, it doesn't appear.",
		"Every claim on this site is verified.",
		"You can rely on our analysis.",
		"Our reporting is always accurate.",
		"Guaranteed accurate pricing on every line.",
		"Our method is not a disclaimer.",
		"We are never wrong about a specification.",
	}
	for _, s := range mustBlock {
		if f := globalFindings(t, s); len(f) == 0 {
			t.Errorf("fleet-wide set did NOT block an overclaim: %q", s)
		}
	}
}

// The shape bugs_open/104 § "How to verify a fix" names explicitly, on a site
// with no register of its own — which is the whole purpose of the fleet-wide set.
func TestGlobalAppliesWithNoRegisterAtAll(t *testing.T) {
	f := ScanAllBannedClaims(
		ExtractAssertionText("<p>Every claim on this site is verified.</p>"), nil)
	if len(f) != 1 {
		t.Fatalf("a site with no evidence_base must still be protected: got %d findings, want 1", len(f))
	}
	if f[0].Check != "banned_claim" {
		t.Errorf("check = %q, want banned_claim", f[0].Check)
	}
}

// ---------------------------------------------------------------------------
// MUST PASS — negated forms. Real copy from the 2026-07-28 fleet dry run; each
// of these was a FALSE POSITIVE that would have failed a live page build.
// ---------------------------------------------------------------------------

func TestGlobalDoesNotBlockNegatedDisclosure(t *testing.T) {
	mustPass := []struct{ site, sentence string }{
		{"robot-hands.com",
			"Where manufacturer data has not been independently verified, that is stated explicitly."},
		{"robot-hands.com",
			"When a figure cannot be independently verified, it is marked as unverified rather than carried forward as fact."},
		{"vonc.com",
			"Competitor characterisations reflect general platform mechanics as of 2026 and are Spark's own assessment, not independently verified."},
		// CORRECTED 2026-07-29 — a fourth fixture was removed from this list.
		// It was robot-hands' "…with specifications drawn directly from
		// manufacturer datasheets and, where available, independently verified
		// test data", which I filed here on 07-28 as one of the four negated
		// false positives. It contains NO negation. The 07-28 dry run produced 4
		// FINDINGS from 3 distinct sentences (vonc's counted twice, on two
		// components) — I read "4 findings" as "4 sentences" and recruited the
		// nearest other sentence containing the phrase to make up the number.
		// It is an ASSERTION and now sits in the must-block list below, where
		// the fleet dry run says it belongs. It passed here for eight hours only
		// because the pattern that matches it had been removed from the set.
	}
	for _, c := range mustPass {
		if f := globalFindings(t, c.sentence); len(f) != 0 {
			t.Errorf("FALSE POSITIVE on live %s copy — %q matched %q (%s)",
				c.site, c.sentence, f[0].Pattern, f[0].Reason)
		}
	}
}

// ---------------------------------------------------------------------------
// THE GUARD, and why the test above is not enough on its own.
//
// From 2026-07-28 to 2026-07-29 TestGlobalDoesNotBlockNegatedDisclosure passed
// for a reason that had nothing to do with negation: the only pattern that
// matched those sentences had been REMOVED from the set. A pass from a check
// that cannot fail outlives the blindness — it gets quoted later as "negated
// copy is handled", which was not true and was never tested.
//
// So these assert the mechanism, not the outcome: the pattern DOES match the
// sentence, and the guard is what drops it. If someone deletes the guard, the
// test above starts failing; if someone deletes the pattern, THIS one does.
// ---------------------------------------------------------------------------

func TestNegatedCopyIsSuppressedByTheGuardNotByAbsence(t *testing.T) {
	// Verbatim live copy, same source as the fixtures above.
	negated := []string{
		"Where manufacturer data has not been independently verified, that is stated explicitly.",
		"When a figure cannot be independently verified, it is marked as unverified rather than carried forward as fact.",
		"Competitor characterisations reflect general platform mechanics as of 2026 and are Spark's own assessment, not independently verified.",
	}
	for _, s := range negated {
		blocks := ExtractAssertionText("<p>" + s + "</p>")
		findings, suppressed := ScanAllBannedClaimsWithSuppressed(blocks, nil)
		if len(findings) != 0 {
			t.Errorf("guard failed to suppress a negated claim: %q raised %q", s, findings[0].Pattern)
			continue
		}
		if len(suppressed) == 0 {
			t.Errorf("VACUOUS PASS: nothing in the fleet-wide set matches %q at all, so the "+
				"guard was never exercised. Either the external-verification pattern was "+
				"removed, or this fixture no longer contains the phrase it was chosen for.", s)
		}
	}
}

// The unguarded scan must still fire, so "the guard suppressed it" is a claim
// about the guard and not about a pattern that quietly stopped matching. This is
// the pre-2026-07-29 behaviour, pinned.
func TestUnguardedScanStillFiresOnNegatedCopy(t *testing.T) {
	blocks := ExtractAssertionText(
		"<p>Where manufacturer data has not been independently verified, that is stated explicitly.</p>")
	if f := globalEvidence().ScanBannedClaimsIgnoringNegation(blocks); len(f) == 0 {
		t.Fatal("ScanBannedClaimsIgnoringNegation raised nothing — it is supposed to reproduce " +
			"the false positive this workstream exists to fix, so the guard's effect stays visible")
	}
}

// The two live overclaims the guard made visible. They were in the corpus the
// whole time, hidden behind the false positives that got the pattern excluded —
// so they are fixtures now, and a build of those components is expected to fail.
func TestGlobalBlocksTheLiveExternalVerificationOverclaims(t *testing.T) {
	mustBlock := []struct{ site, sentence string }{
		{"robot-hands.com/gripper-catalog",
			"Grip force, stroke, cycle time, and IP rating pulled from manufacturer datasheets and independently verified."},
		// VERBATIM, em-dash clause included. An abbreviated quote is a different
		// claim, and the abbreviation would remove the ", where available," that
		// makes this one a hedged assertion rather than a flat one — the very
		// thing being tested.
		{"robot-hands.com/how-it-works",
			"The catalog indexes gripper models across six actuation technologies — pneumatic, electric, vacuum, magnetic, soft-robotic, and adhesive — with specifications drawn directly from manufacturer datasheets and, where available, independently verified test data."},
	}
	for _, c := range mustBlock {
		if f := globalFindings(t, c.sentence); len(f) == 0 {
			t.Errorf("live overclaim on %s no longer blocked: %q", c.site, c.sentence)
		}
	}
}

// The guard is clause-local on purpose. A negation in a DIFFERENT clause must
// not launder an overclaim — otherwise "we do not use AI, and every claim here
// is verified" becomes the evasion.
func TestGuardIsClauseLocal(t *testing.T) {
	stillBlocked := []string{
		"We do not use AI, and every claim on this site is verified.",
		"We never hide our sources. Our reporting is always accurate.",
		"Sources are not paywalled; every figure is independently verified.",
	}
	for _, s := range stillBlocked {
		if f := globalFindings(t, s); len(f) == 0 {
			t.Errorf("guard laundered an overclaim across a clause boundary: %q", s)
		}
	}
}

// Denials of the banned claim, in the same clause, must pass — including the
// contraction and typographic-apostrophe forms, which are what a renderer emits.
func TestGuardHandlesRealNegationForms(t *testing.T) {
	mustPass := []string{
		"We do not claim every figure is verified.",
		"We can't say that our reporting is always accurate.",
		"We can’t say that our reporting is always accurate.", // curly apostrophe
		"We cannot promise our analysis is always accurate.",
		"We never claim you can rely on this.",
		"We don't guarantee accurate pricing.",
	}
	for _, s := range mustPass {
		if f := globalFindings(t, s); len(f) != 0 {
			t.Errorf("FALSE POSITIVE on a denial: %q matched %q", s, f[0].Pattern)
		}
	}
}

// The three completeness-of-exclusion patterns are themselves NEGATIVE
// constructions ("claims without a source do NOT appear here"). A guard that
// looked anywhere nearby for a cue, rather than strictly BEFORE the match in the
// same clause, would disarm all three — the guard's own worst failure mode, and
// invisible from outside because the gate would simply go quiet.
func TestGuardDoesNotDisarmTheNegativeConstructionPatterns(t *testing.T) {
	stillBlocked := []string{
		"A claim without a source does not appear here.",
		"Prices do not appear here.",
		"If we cannot verify a figure, it doesn't appear.",
		"We are never wrong about a specification.",
	}
	for _, s := range stillBlocked {
		if f := globalFindings(t, s); len(f) == 0 {
			t.Errorf("the negation guard disarmed a pattern whose banned form IS negative: %q", s)
		}
	}
}

// The guard applies to per-site registers too — one rule, not two that drift
// (CLM-004). A site's own audited pattern must behave the same way.
func TestGuardAppliesToPerSitePatternsAsWell(t *testing.T) {
	eb, err := ParseEvidenceBase([]byte(
		`{"audit_doc":"x","facts":[],"banned_claims":[{"pattern":"award.winning","reason":"site-specific"}]}`))
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	if f := ScanAllBannedClaims(ExtractAssertionText("<p>We are an award-winning studio.</p>"), eb); len(f) != 1 {
		t.Fatalf("per-site pattern should fire on the assertion, got %d findings", len(f))
	}
	if f := ScanAllBannedClaims(
		ExtractAssertionText("<p>We are not an award-winning studio.</p>"), eb); len(f) != 0 {
		t.Errorf("the guard must apply to per-site patterns too, got %d findings", len(f))
	}
}

// Occurrence counts are computed from the SURVIVING matches. A block that says
// the thing once and denies it once is one finding, not two — otherwise the
// count reported to a page author is inflated by the sentences they got right.
func TestOccurrenceCountExcludesSuppressedMatches(t *testing.T) {
	blocks := ExtractAssertionText(
		"<p>Every figure is independently verified. Where a figure has not been independently verified, we say so.</p>")
	f := ScanAllBannedClaims(blocks, nil)
	if len(f) == 0 {
		t.Fatal("expected the asserted half to be blocked")
	}
	for _, x := range f {
		if x.Occurrences != 1 {
			t.Errorf("pattern %q reported %d occurrences; the negated one must not be counted",
				x.Pattern, x.Occurrences)
		}
	}
}

// ---------------------------------------------------------------------------
// MUST PASS — legitimate process commitments. A site may describe what it DOES.
// ---------------------------------------------------------------------------

func TestGlobalDoesNotBlockLegitimateProcessCopy(t *testing.T) {
	mustPass := []string{
		"We cite each figure and date it.",
		"The statute is the authoritative text.",
		"We check our sources before we publish.",
		// Owner ruling 2026-07-28: a claim about a site's own delivered work is
		// not an accuracy overclaim. This is why the `every … is verified`
		// pattern is narrowed to claim/content nouns.
		"Every component is verified against production.",
		// The same shape on other process outputs must also pass.
		"Every deployment is checked against staging first.",
	}
	for _, s := range mustPass {
		if f := globalFindings(t, s); len(f) != 0 {
			t.Errorf("FALSE POSITIVE on legitimate copy: %q matched %q (%s)",
				s, f[0].Pattern, f[0].Reason)
		}
	}
}

// ---------------------------------------------------------------------------
// Union semantics with a site's own register.
// ---------------------------------------------------------------------------

// A per-site pattern and a fleet-wide pattern both apply.
func TestScanAllUnionsPerSiteAndGlobal(t *testing.T) {
	eb := mustParseTestEB(t) // leopardess register: "awards won", "peter grenfell", …
	blocks := ExtractAssertionText(
		"<p>Awards Won: three.</p><p>Our reporting is always accurate.</p>")

	f := ScanAllBannedClaims(blocks, eb)
	if len(f) != 2 {
		t.Fatalf("want 2 findings (1 per-site + 1 fleet-wide), got %d", len(f))
	}
	var sawSite, sawGlobal bool
	for _, x := range f {
		if x.Pattern == "awards won" {
			sawSite = true
		}
		if x.Reason == "self accuracy overclaim. Anchored to self, so 'the statute is authoritative' still passes." {
			sawGlobal = true
		}
	}
	if !sawSite || !sawGlobal {
		t.Errorf("union lost a half: per-site=%v fleet-wide=%v", sawSite, sawGlobal)
	}
}

// A site whose register carries a fleet-wide pattern verbatim — oufe does, via
// migration 226 — must not report the same sentence twice.
func TestScanAllDeduplicatesIdenticalPattern(t *testing.T) {
	dup := globalBannedClaims()[6] // "guaranteed (accurate|correct|…)"
	eb, err := ParseEvidenceBase([]byte(
		`{"audit_doc":"x","facts":[],"banned_claims":[{"pattern":"` + dup.Pattern +
			`","reason":"per-site copy of a fleet-wide pattern"}]}`))
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	f := ScanAllBannedClaims(
		ExtractAssertionText("<p>Guaranteed accurate to the penny.</p>"), eb)
	if len(f) != 1 {
		t.Fatalf("an identical pattern in both sets must report once, got %d findings", len(f))
	}
}

// The numeric scan must NOT be dragged fleet-wide by this change: it stays
// strictly opt-in, because its false-positive rate is why it is never a blocker.
func TestGlobalSetDoesNotArmTheNumberScan(t *testing.T) {
	var nilEB *EvidenceBase
	if f := nilEB.ScanUnregisteredNumbers(
		ExtractAssertionText("<p>We serve 4,200 clients.</p>"), ClaimSurface{}); len(f) != 0 {
		t.Errorf("a site with no register must raise no unregistered-number findings, got %d", len(f))
	}
}

// ---------------------------------------------------------------------------
// The fleet-wide patterns are OUR code, not a site's user-authored config, so a
// malformed one is a programming error rather than a typo to be tolerated.
//
// globalEvidence() keeps ParseEvidenceBase's fallback — an uncompilable pattern
// degrades to a literal case-insensitive substring — because taking the chassis
// down at init over a regex would be far worse than one over-narrow pattern. But
// that fallback is SILENT: it has no logger and no error path, so a typo would
// quietly become a near-inert literal that still looks armed from outside. This
// test is the guard that makes the fallback safe to keep: a malformed fleet-wide
// pattern fails CI instead of shipping as a pattern that matches almost nothing.
//
// Raised by the council's bug_historian seat (round 2, corr 899ed92e).
// ---------------------------------------------------------------------------

func TestEveryGlobalPatternIsAValidRegex(t *testing.T) {
	for i, bc := range globalBannedClaims() {
		if _, err := regexp.Compile("(?i)" + bc.Pattern); err != nil {
			t.Errorf("fleet-wide pattern %d does not compile, so it would silently "+
				"degrade to a literal substring and match almost nothing:\n  pattern: %s\n  error: %v",
				i, bc.Pattern, err)
		}
	}
}

// A pattern that compiles is not necessarily a pattern that MATCHES. An empty or
// whitespace-only pattern compiles happily and then matches every block, which
// would fail every build on every site — the opposite failure to the one above,
// and equally invisible from outside.
func TestNoGlobalPatternIsVacuous(t *testing.T) {
	harmless := ExtractAssertionText("<p>We publish opening hours and a phone number.</p>")
	for i, bc := range globalBannedClaims() {
		if strings.TrimSpace(bc.Pattern) == "" {
			t.Errorf("fleet-wide pattern %d is empty — it would match every block", i)
			continue
		}
		if bc.Reason == "" {
			t.Errorf("fleet-wide pattern %d has no reason; the reason is what the author of a "+
				"blocked page reads: %s", i, bc.Pattern)
		}
		one := &EvidenceBase{BannedClaims: []BannedClaim{{Pattern: bc.Pattern, Reason: bc.Reason}}}
		for j := range one.BannedClaims {
			one.BannedClaims[j].re = regexp.MustCompile("(?i)" + one.BannedClaims[j].Pattern)
		}
		if f := one.ScanBannedClaims(harmless); len(f) != 0 {
			t.Errorf("fleet-wide pattern %d fires on ordinary copy — %q matched %q",
				i, bc.Pattern, f[0].Matched)
		}
	}
}
