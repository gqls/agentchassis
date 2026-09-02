// logo_clause_parity_test.go — bugs_open/417.
//
// The fallback builder and the generation choke point must state the SAME text
// rule. The bug was precisely that they did not: the rule lived here alone, and
// the path every real site takes never saw it. Sharing one constant is what
// makes drift impossible; this test is what makes the sharing load-bearing.
//
// MUTATION THAT MUST BREAK THIS: replace the shared constant at either site
// with a fresh string literal saying the same thing in its own words.
package discovery_checks

import (
	"strings"
	"testing"
)

func TestLogoFallbackPromptCarriesTheSharedTextRule(t *testing.T) {
	got := composeBrandImagePrompt(siteBrandFacts{
		Name: "Boxing Online", Domain: "boxingonline.com", Industry: "sport",
	}, "logo")

	if !strings.Contains(got, LogoTextFreeClause) {
		t.Fatalf("the fallback logo prompt no longer carries LogoTextFreeClause verbatim — "+
			"the two prompt paths have drifted, which is bugs_open/417's mechanism:\n%s", got)
	}
}

// A hero prompt must NOT acquire the logo text rule: heroes are deliberately
// composed with clear space for overlaid headline text.
func TestHeroPromptDoesNotCarryTheLogoTextRule(t *testing.T) {
	got := composeBrandImagePrompt(siteBrandFacts{
		Name: "Boxing Online", Domain: "boxingonline.com",
	}, "hero")

	if strings.Contains(got, LogoTextFreeSentinel) {
		t.Fatalf("hero prompt carries the logo text rule:\n%s", got)
	}
}

func TestLogoWordmarkClauseNamesTheExactString(t *testing.T) {
	got := LogoWordmarkClause("farmerinsurance")
	if !strings.Contains(got, `"farmerinsurance"`) {
		t.Fatalf("the clause must quote the exact string, so an unnamed wordmark is "+
			"not expressible at all:\n%s", got)
	}
}

// TestLogoBackgroundKeyClauseIsSelfConsistent — bugs_open/424. The clause text
// and the structured KeyGround value handed to the adapter (dynamic_adapter.go)
// are both derived from LogoBackgroundKeyHex; this pins that they still agree,
// so a future reword of the clause cannot silently change the colour without
// also changing what the adapter mattes for.
func TestLogoBackgroundKeyClauseIsSelfConsistent(t *testing.T) {
	if !strings.Contains(LogoBackgroundKeyClause, LogoBackgroundKeySentinel) {
		t.Fatalf("clause does not contain its own sentinel — idempotence in "+
			"applyLogoBackgroundPolicy relies on this substring match:\n%s", LogoBackgroundKeyClause)
	}
	if !strings.Contains(LogoBackgroundKeyClause, LogoBackgroundKeyHex) {
		t.Fatalf("clause does not name the key hex the adapter is told to matte for:\n%s",
			LogoBackgroundKeyClause)
	}
}

// TestHeroPromptDoesNotCarryTheBackgroundKeyRule — heroes are photographic
// content, deliberately NOT flat-vector marks, and must never be keyed
// against a solid colour. Nothing currently wires this clause into hero
// prompts (it applies only at the kind=logo choke point in
// generate_image_actions.go), so this pins that absence as a guard against a
// future session widening the gate carelessly.
func TestHeroPromptDoesNotCarryTheBackgroundKeyRule(t *testing.T) {
	got := composeBrandImagePrompt(siteBrandFacts{
		Name: "Boxing Online", Domain: "boxingonline.com",
	}, "hero")
	if strings.Contains(got, LogoBackgroundKeySentinel) {
		t.Fatalf("hero prompt carries the logo background-key rule:\n%s", got)
	}
}
