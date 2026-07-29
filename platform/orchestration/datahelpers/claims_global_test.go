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
	if n := GlobalBannedClaimCount(); n != 9 {
		t.Fatalf("expected 9 fleet-wide patterns, got %d — if this changed deliberately, "+
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
		{"robot-hands.com",
			"The catalog indexes gripper models across six actuation technologies — pneumatic, electric, vacuum, magnetic, soft-robotic, and adhesive — with specifications drawn directly from manufacturer datasheets and, where available, independently verified test data."},
	}
	for _, c := range mustPass {
		if f := globalFindings(t, c.sentence); len(f) != 0 {
			t.Errorf("FALSE POSITIVE on live %s copy — %q matched %q (%s)",
				c.site, c.sentence, f[0].Pattern, f[0].Reason)
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
