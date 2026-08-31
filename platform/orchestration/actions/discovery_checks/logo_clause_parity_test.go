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
