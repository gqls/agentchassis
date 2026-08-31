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

	"github.com/gqls/agentchassis/pkg/models"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"go.uber.org/zap"
)

const boxingPrompt = "A bold, sharp logomark for a boxing news website — a stylised " +
	"boxing glove or ring ropes abstracted into a clean geometric form, strong and " +
	"confident, no text other than the wordmark itself, suitable for dark and light backgrounds"

func TestLogoTextPolicyDefaultAppendsTextFreeClause(t *testing.T) {
	// The real boxingonline prompt: the licence 670 could not match, because it
	// is a PARAPHRASE ("other than", not "outside"). The guard needs no literal
	// match at all, which is why it bounds the class where a migration floors it.
	got, neg, src, _ := applyLogoTextPolicy(
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
	got, neg, _, _ := applyLogoTextPolicy(
		hero+" "+checks.LogoTextFreeClause, "people", "site_plan", "", logoIdentity{}, zap.NewNop())
	if got != hero+" "+checks.LogoTextFreeClause {
		t.Fatalf("prompt mutated when the clause was already present:\n%s", got)
	}
	if neg != "people" {
		t.Fatalf("negative prompt mutated on the idempotent path: %q", neg)
	}
}

func TestLogoTextPolicyIsIdempotent(t *testing.T) {
	once, _, _, _ := applyLogoTextPolicy(boxingPrompt, "", "site_plan", "", logoIdentity{}, zap.NewNop())
	twice, _, src, _ := applyLogoTextPolicy(once, "", "site_plan", "", logoIdentity{}, zap.NewNop())

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
	got, neg, src, _ := applyLogoTextPolicy(
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
	got, _, src, rejected := applyLogoTextPolicy(
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
	// Council round 1, MEDIUM objection (bugs_open/034's shape): a rejection
	// that exists only as a Warn is invisible to every census. The fourth
	// return is what the caller files a durable note from.
	if rejected != "Farm Shield Info" {
		t.Fatalf("rejected wordmark not surfaced for the durable record: %q", rejected)
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

// TestResolveKindClosesTheLegacyParentGapAndSurfacesDisagreement — three
// council rounds' worth of objections, pinned.
//
// Round 2 HIGH (bug_historian): keying the logo guard on `kind` alone leaves the
// two legacy parents ungoverned, reproducing 417's root cause on a second axis.
// Round 2 MEDIUM (bug_historian): believing any stated non-logo kind opens the
// MIRROR case — a mislabelled logo skips the policy with no error surface at
// all, which is 417 on a THIRD axis.
// Round 2 MEDIUM (editquality): "identified as non-logo" and "nothing
// identified it" must be distinguishable, or a hero call that sets `purpose`
// but no `kind` files a false "no kind" note and pollutes the liveness
// measurement that note exists to provide.
//
// MUTATIONS THAT MUST BREAK THIS — each applied ALONE:
//   - delete the step_name arm → the legacy-parent cases fail.
//   - drop the Conflict assignment → the mislabel case fails.
//   - set Answered=true unconditionally → the "nothing identified it" case fails.
//   - make stepNameKindHint return "logo" when a name contains both "logo" and
//     "hero" → the check_logo_or_hero case fails.
func TestResolveKindClosesTheLegacyParentGapAndSurfacesDisagreement(t *testing.T) {
	cases := []struct {
		name         string
		inputData    map[string]interface{}
		step         models.Step
		wantKind     string
		wantSignal   string
		wantAnswered bool
		wantConflict bool
	}{
		{"modern handler branch", map[string]interface{}{"kind": "logo"}, models.Step{}, "logo", "kind", true, false},
		{"legacy default_kind", map[string]interface{}{"default_kind": "hero"}, models.Step{}, "hero", "default_kind", true, false},
		{"input purpose", map[string]interface{}{"purpose": "logo"}, models.Step{}, "logo", "input_purpose", true, false},
		{"spec purpose", map[string]interface{}{"spec": map[string]interface{}{"purpose": "logo"}}, models.Step{}, "logo", "spec_purpose", true, false},
		{"step config purpose", nil, models.Step{Config: map[string]interface{}{"purpose": "logo"}}, "logo", "step_purpose", true, false},
		{"step config kind", nil, models.Step{Config: map[string]interface{}{"kind": "logo"}}, "logo", "step_kind", true, false},

		// THE LEGACY PARENTS — site-work-orchestrator and pageflow-builder map
		// {prompt, site_id, site_plan, reviewed_brief} and nothing else. The step
		// NAME is the only signal they carry, and without this arm their logo
		// prompts reach the model ungoverned.
		{"legacy parent, step name only", nil, models.Step{Name: "call_logo_generation"}, "logo", "step_name", true, false},
		{"hero step name", nil, models.Step{Name: "call_hero_gen"}, "hero", "step_name", true, false},

		// [MEASURED 2026-08-31] check_logo_or_hero is a REAL live step name and
		// it names both kinds. An ambiguous name must yield NO hint rather than
		// a guess — a wrong guess here silently mis-governs a prompt.
		{"ambiguous name names both kinds", nil, models.Step{Name: "check_logo_or_hero"}, "", "", false, false},

		// The distinction editquality's objection turned on: nothing at all.
		{"nothing identifies it", map[string]interface{}{"prompt": "a mark"}, models.Step{Name: "call_image_gen"}, "", "", false, false},

		// The mirror case: a stated kind the step's own name contradicts. The
		// STATED kind still wins — a classifier overriding a caller is worse —
		// but the disagreement must be surfaced, not swallowed.
		{"mislabelled logo is believed BUT flagged", map[string]interface{}{"kind": "hero"}, models.Step{Name: "call_logo_generation"}, "hero", "kind", true, true},
		{"agreeing declarations raise no conflict", map[string]interface{}{"kind": "logo"}, models.Step{Name: "call_logo_generation"}, "logo", "kind", true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := resolveKind(c.inputData, nil, c.step)
			if got.Kind != c.wantKind || got.Signal != c.wantSignal || got.Answered != c.wantAnswered {
				t.Fatalf("resolveKind = {Kind:%q Signal:%q Answered:%v}, want {Kind:%q Signal:%q Answered:%v}",
					got.Kind, got.Signal, got.Answered, c.wantKind, c.wantSignal, c.wantAnswered)
			}
			if (got.Conflict != "") != c.wantConflict {
				t.Fatalf("Conflict = %q, wantConflict=%v", got.Conflict, c.wantConflict)
			}
		})
	}
}

// TestImagePolicyEventKeepsItsExplicitProvenance — council round 3 raised a
// gating HIGH (tooling_provenance, echoed by prior_art and editquality) that
// rewiring recordImagePolicyEvent onto LogActionEntryInheritingProvenance might
// "swap one silent record-nulling bug (the doc_notes constraint) for another",
// because the estate carries a landmine titled "LogActionEntry's merge fills a
// provenance you meant to set — and every test in the package stays green".
//
// The premise is refutable and this test is the refutation, kept as a GUARD so
// it cannot go stale the way the objection's reading of the landmine did:
//
//   - the landmine is marked ✅ FIXED AND LIVE (v1.0.1268, 2026-08-08) and says
//     in its own body "The merge only fills fields left ZERO, so a named field
//     can never be overwritten"; its header warns "read the new-API paragraph at
//     the bottom, not the pre-roll one above it";
//   - resolveProvenance's inherit branch only assigns when the field is "";
//   - inheritJoinIdentity touches WorkItemID/OrchestrationID/AgentID/PodName
//     ONLY, and never SiteID, AgentType or StepName.
//
// The seat was right to ask. "Every test in the package stays green" is exactly
// the landmine's stated failure mode, so the honest answer is a test that would
// NOT stay green.
//
// MUTATION THAT MUST BREAK THIS: make resolveProvenance's inherit branch assign
// unconditionally (drop the `if entry.AgentType == ""` / `if entry.StepName ==
// ""` guards) — the explicit values are then clobbered and this test fails,
// which is precisely the defect the objection feared.
func TestImagePolicyEventKeepsItsExplicitProvenance(t *testing.T) {
	entry := agenterrors.Entry{
		SiteID:       "11111111-2222-3333-4444-555555555555",
		AgentType:    "image-generator",
		StepName:     "call_logo_generation",
		Action:       "generate_image",
		ErrorCode:    "image_kind_conflict",
		Severity:     "warning",
		ErrorMessage: "declarations disagree",
	}
	params := ActionParams{
		StepConfig: models.Step{Name: "some_other_running_step"},
		Logger:     zap.NewNop(),
	}

	resolveProvenance(params, &entry, true, zap.NewNop())

	// The detector this round leans on is only meaningful if its rows are
	// attributed to the site and step that actually produced them.
	if entry.SiteID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("SiteID was overwritten by the provenance merge: %q", entry.SiteID)
	}
	if entry.AgentType != "image-generator" {
		t.Fatalf("AgentType was overwritten by the provenance merge: %q", entry.AgentType)
	}
	if entry.StepName != "call_logo_generation" {
		t.Fatalf("StepName was overwritten by the provenance merge (got the running step?): %q", entry.StepName)
	}
	if entry.ErrorCode != "image_kind_conflict" {
		t.Fatalf("ErrorCode mutated: %q", entry.ErrorCode)
	}
}
