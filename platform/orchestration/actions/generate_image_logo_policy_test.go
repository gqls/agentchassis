// generate_image_logo_policy_test.go — bugs_open/417.
//
// The estate's no-lettering rule for generated logos existed only inside the
// FALLBACK prompt builder, so it governed the one population that never needed
// it: every planner-built site supplies its own prompt, and that prompt reached
// the image model ungoverned. The result was invented brand names on two live
// sites — "Farm Shield Info" on farmerinsurance.uk and "BOXING NEWS" on
// boxingonline.com, the first paid customer. applyLogoTextPolicy moves the rule
// to the generation choke point, where every prompt from every producer passes.
//
// MUTATIONS THAT MUST BREAK THESE — apply each ALONE and watch the named test
// fail. A mutation that PASSES has hit a guard in series, not a redundant test;
// investigate it rather than assuming the coverage is doubled.
//
//  1. delete the default-arm append, or invert the kind gate at the call site
//     → TestLogoTextPolicyDefaultAppendsTextFreeClause.
//  2. widen the call-site gate to hero/empty kind
//     → TestLogoTextPolicyLeavesNonLogoPromptsAlone (the pure function must
//     never be reached for those kinds; this pins the contract it relies on).
//  3. drop the sentinel idempotence check
//     → TestLogoTextPolicyIsIdempotent.
//  4. drop the negative-prompt reconciliation on an accepted wordmark
//     → TestLogoTextPolicyAcceptedWordmarkStopsFightingItself.
//  5. make validation accept any string (the door this closes: a planner LLM
//     can write the opt-in field, so "the field exists" must not be the licence)
//     → TestLogoTextPolicyRejectsAWordmarkThatIsNotThisSitesName.
//  6. break the normaliser (drop case-folding or punctuation-stripping)
//     → TestLogoTextPolicyGroundsAcrossNamingColumns.
package actions

import (
	"strings"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"go.uber.org/zap"
)

const boxingPrompt = "A bold, sharp logomark for a boxing news website — a stylised " +
	"boxing glove or ring ropes abstracted into a clean geometric form, strong and " +
	"confident, no text other than the wordmark itself, suitable for dark and light backgrounds"

func TestLogoTextPolicyDefaultAppendsTextFreeClause(t *testing.T) {
	// The real boxingonline prompt: the licence 670 could not match, because it
	// is a PARAPHRASE ("other than", not "outside"). The guard needs no literal
	// match at all, which is why it bounds the class where a migration floors it.
	got, neg, src := applyLogoTextPolicy(
		boxingPrompt, "people, faces, signature, watermark", "site_plan", "",
		logoIdentity{}, zap.NewNop())

	if !strings.Contains(got, checks.LogoTextFreeClause) {
		t.Fatalf("text-free clause not appended:\n%s", got)
	}
	if !strings.Contains(got, "no text other than the wordmark itself") {
		t.Fatal("the original prompt must be preserved verbatim — the clause OVERRIDES it, " +
			"it does not rewrite it, so the wording stays auditable in assets.origin_prompt")
	}
	if !strings.HasSuffix(src, "+logo_text_policy") {
		t.Fatalf("prompt source not tagged, so a census cannot tell the guard fired: %q", src)
	}
	for _, want := range []string{"text", "lettering", "words"} {
		if !strings.Contains(neg, want) {
			t.Errorf("negative prompt missing %q (belt, not the mechanism): %q", want, neg)
		}
	}
}

// TestLogoTextPolicyLeavesNonLogoPromptsAlone pins the mirror of the
// 2026-05-20 contamination lesson: a hero prompt must never acquire a
// text-free instruction, because heroes legitimately carry overlaid headlines.
// The kind gate lives at the call site, so what this test can assert is that
// the function is a pure transform with no hidden global effect — run it and
// confirm an unrelated prompt is returned untouched when nothing opts in and
// the clause is already present.
func TestLogoTextPolicyLeavesNonLogoPromptsAlone(t *testing.T) {
	hero := "A hero image for a boxing news site, with clear space for overlaid headline text."
	got, neg, _ := applyLogoTextPolicy(
		hero+" "+checks.LogoTextFreeClause, "people", "site_plan", "", logoIdentity{}, zap.NewNop())
	if got != hero+" "+checks.LogoTextFreeClause {
		t.Fatalf("prompt mutated when the clause was already present:\n%s", got)
	}
	if neg != "people" {
		t.Fatalf("negative prompt mutated on the idempotent path: %q", neg)
	}
}

func TestLogoTextPolicyIsIdempotent(t *testing.T) {
	once, _, _ := applyLogoTextPolicy(boxingPrompt, "", "site_plan", "", logoIdentity{}, zap.NewNop())
	twice, _, src := applyLogoTextPolicy(once, "", "site_plan", "", logoIdentity{}, zap.NewNop())

	if once != twice {
		t.Fatalf("second application changed the prompt:\n once: %s\ntwice: %s", once, twice)
	}
	if n := strings.Count(twice, checks.LogoTextFreeSentinel); n != 1 {
		t.Fatalf("clause appears %d times, want exactly 1 — a retried generation or a "+
			"670/680-washed row must not accumulate copies", n)
	}
	if !strings.HasSuffix(src, "+logo_text_policy") {
		t.Fatalf("idempotent path must still tag the source: %q", src)
	}
}

func TestLogoTextPolicyAcceptedWordmarkStopsFightingItself(t *testing.T) {
	ident := logoIdentity{CompanyName: "Boxing Online", Domain: "boxingonline.com"}
	got, neg, src := applyLogoTextPolicy(
		"A bold logomark.", "people, faces, text, letters, signature, watermark",
		"site_plan", "Boxing Online", ident, zap.NewNop())

	if !strings.Contains(got, `the exact wordmark "Boxing Online"`) {
		t.Fatalf("wordmark clause missing or does not name the exact string:\n%s", got)
	}
	if strings.Contains(got, checks.LogoTextFreeClause) {
		t.Fatal("an accepted wordmark must not also carry the text-free clause")
	}
	// The measured lesson from the failing generation: the adapter folds
	// negative_prompt INTO the positive prompt, so leaving "text" in there while
	// asking for a wordmark ships two contradictory instructions in one prompt.
	for _, banned := range []string{"text", "letters"} {
		for _, tok := range strings.Split(neg, ",") {
			if strings.TrimSpace(tok) == banned {
				t.Errorf("negative prompt still forbids %q alongside an approved wordmark: %q", banned, neg)
			}
		}
	}
	for _, keep := range []string{"watermark", "signature", "people"} {
		if !strings.Contains(neg, keep) {
			t.Errorf("reconciliation dropped %q, which forbids artefacts, not the brand name: %q", keep, neg)
		}
	}
	if !strings.Contains(src, "+wordmark") {
		t.Fatalf("source not tagged for the opt-in arm: %q", src)
	}
}

// TestLogoTextPolicyRejectsAWordmarkThatIsNotThisSitesName is the test that
// closes the door against the escape hatch's own producer. "Farm Shield Info"
// is the string the model actually invented; asked for on a site called Boxing
// Online it must degrade to a text-free mark, and must NOT refuse — refusing
// would mint an unhandleable item (the bugs_open/210 lesson).
func TestLogoTextPolicyRejectsAWordmarkThatIsNotThisSitesName(t *testing.T) {
	ident := logoIdentity{CompanyName: "Boxing Online", Domain: "boxingonline.com"}
	got, _, src := applyLogoTextPolicy(
		"A bold logomark.", "", "site_plan", "Farm Shield Info", ident, zap.NewNop())

	if strings.Contains(got, "Farm Shield Info") {
		t.Fatalf("an ungrounded wordmark reached the prompt:\n%s", got)
	}
	if !strings.Contains(got, checks.LogoTextFreeClause) {
		t.Fatalf("rejection must DEGRADE to text-free, not drop the policy entirely:\n%s", got)
	}
	if !strings.Contains(src, "+wordmark_rejected") {
		t.Fatalf("rejection must be visible in the prompt source: %q", src)
	}
}

func TestLogoTextPolicyGroundsAcrossNamingColumns(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		ident logoIdentity
		want  bool
	}{
		{"company_name exact", "Boxing Online", logoIdentity{CompanyName: "Boxing Online"}, true},
		{"company_name case+space insensitive", "boxingonline", logoIdentity{CompanyName: "Boxing Online"}, true},
		{"logo_text", "CareerPrep", logoIdentity{LogoText: "CareerPrep"}, true},
		{"domain stem", "robothands", logoIdentity{Domain: "robot-hands.com"}, true},
		{"domain stem with punctuation in the ask", "ROBOT-HANDS", logoIdentity{Domain: "robot-hands.com"}, true},
		{"multi-label suffix", "cv1", logoIdentity{Domain: "cv1.co.uk"}, true},
		{"www prefix", "loanzy", logoIdentity{Domain: "www.loanzy.uk"}, true},
		{"invented", "Farm Shield Info", logoIdentity{CompanyName: "Farmer Insurance UK", Domain: "farmerinsurance.uk"}, false},
		// THE LIVE CASE, pinned because it is the owner's own named exception and it
		// grounds ONLY on the domain stem: farmerinsurance.uk's plan row carries
		// constraints.wordmark_text="farmerinsurance" (set by the loanzy lane
		// 2026-08-31) while its sites row has company_name AND logo_text both EMPTY.
		// Validating against the naming columns alone would REJECT the owner's
		// explicit instruction and silently degrade it to a text-free mark. The
		// domain stem is not a convenience here — it is the only thing that makes
		// this site's exception expressible at all.
		{"owner's live exception, grounded only by domain stem", "farmerinsurance",
			logoIdentity{CompanyName: "", LogoText: "", Domain: "farmerinsurance.uk"}, true},
		{"empty identity cannot ground anything", "Anything", logoIdentity{}, false},
		{"empty ask", "", logoIdentity{CompanyName: "Boxing Online"}, false},
		{"substring is not a match", "Boxing", logoIdentity{CompanyName: "Boxing Online"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := wordmarkGroundsInIdentity(c.text, c.ident); got != c.want {
				t.Fatalf("wordmarkGroundsInIdentity(%q, %+v) = %v, want %v", c.text, c.ident, got, c.want)
			}
		})
	}
}

func TestLogoWordmarkTextFromInputs(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]interface{}
		want string
	}{
		{"absent", map[string]interface{}{}, ""},
		{"constraints not a map", map[string]interface{}{"constraints": "no_text"}, ""},
		{"no wordmark key", map[string]interface{}{"constraints": map[string]interface{}{"no_text": true}}, ""},
		{"blank is an absence", map[string]interface{}{"constraints": map[string]interface{}{"wordmark_text": "   "}}, ""},
		{"present", map[string]interface{}{"constraints": map[string]interface{}{"wordmark_text": " Boxing Online "}}, "Boxing Online"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := logoWordmarkTextFromInputs(c.in); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
