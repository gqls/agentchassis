// FILE: platform/orchestration/actions/rewrite_negations_action_test.go
//
// The repair split so it is testable without an AI client: planNegationRepairs
// (what would be sent), negationRepairPrompt (what it says), matchTarget and the
// splice (what comes back). The judging itself is AcceptNegationRewrite, tested
// in datahelpers.
//
// Mutation checks (by hand, recorded in the lane NOTES):
//   - count exempt hits against the budget    -> TestExemptHitsDoNotConsumeTheBudget fails
//   - make headline hits obey the budget      -> TestHeadlineHitIsAlwaysATarget fails
//   - key matchTarget on the FIELD name       -> TestMatchTargetIgnoresRenamedField fails
//   - drop the per-page carry in CollectedData -> TestBudgetIsPerPageNotPerSection fails

package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func sectionContent() map[string]interface{} {
	return map[string]interface{}{
		"headline":   "The registry shows you what's possible, not what survives production.",
		"intro":      "This list is pulled from the production registry, not from provider marketing pages.",
		"body":       "<p>We run 1,600 orchestrations a day.</p><p>Talk to the people who run the pipeline, not the ones who catalogued it.</p>",
		"cta_url":    "/contact.html",
		"disclaimer": "These figures are an estimate, not financial advice.",
	}
}

func TestPlanClassifiesEveryHitOnce(t *testing.T) {
	plan := planNegationRepairs(sectionContent(), nil, 2, 0)
	if plan.total < 4 {
		t.Fatalf("expected at least four hits in this content, got %d", plan.total)
	}
	if plan.exemptCount != 1 || plan.exemptReasons["regulatory"] != 1 {
		t.Errorf("the disclaimer must be exempt as regulatory, got %d exempt %v", plan.exemptCount, plan.exemptReasons)
	}
	for _, tg := range plan.targets {
		if tg.Field == "disclaimer" {
			t.Error("a regulatory negation must never become a repair target")
		}
		if tg.Field == "cta_url" {
			t.Error("a URL field must never be walked, let alone repaired")
		}
	}
}

// A brief-supplied phrase is the brief's decision. It is counted and reported;
// it is never rewritten, and it must not eat the page's budget either — the
// budget exists to bound what the WRITER does.
func TestExemptHitsDoNotConsumeTheBudget(t *testing.T) {
	content := map[string]interface{}{
		"hero":  "Multi-agent systems deployed to production in days, not months, on Kubernetes.",
		"intro": "This list is pulled from the production registry, not from provider marketing pages.",
		"outro": "We show the pipeline running, not a slide about it.",
	}
	brief := []string{"Tagline: 'Multi-agent systems deployed to production in days, not months' — use it in the homepage hero."}

	plan := planNegationRepairs(content, brief, 2, 0)
	if plan.exemptCount != 1 {
		t.Fatalf("the supplied tagline must be exempt, got %d (%v)", plan.exemptCount, plan.exemptReasons)
	}
	if plan.withinBudget != 2 || len(plan.targets) != 0 {
		t.Errorf("two non-exempt hits with a budget of 2 must both pass: withinBudget=%d targets=%d",
			plan.withinBudget, len(plan.targets))
	}
	if plan.pageHits != 2 {
		t.Errorf("only NON-exempt hits count toward the page: got %d, want 2", plan.pageHits)
	}
}

func TestHeadlineHitIsAlwaysATarget(t *testing.T) {
	content := map[string]interface{}{
		"headline": "The registry shows you what's possible, not what survives production.",
	}
	// Budget of 5 — far more than the single hit — and it is STILL repaired.
	plan := planNegationRepairs(content, nil, 5, 0)
	if len(plan.targets) != 1 || !plan.targets[0].Headline {
		t.Errorf("a headline hit must be a target regardless of budget, got %+v", plan.targets)
	}
}

// The standard is per PAGE. Section two must see the budget section one used.
func TestBudgetIsPerPageNotPerSection(t *testing.T) {
	one := map[string]interface{}{"intro": "We show what runs, not what demos."}
	two := map[string]interface{}{"intro": "We ship the pipeline, not a slide deck."}
	three := map[string]interface{}{"intro": "We report the failures, not just the wins."}

	p1 := planNegationRepairs(one, nil, 2, 0)
	p2 := planNegationRepairs(two, nil, 2, p1.pageHits)
	p3 := planNegationRepairs(three, nil, 2, p2.pageHits)

	if len(p1.targets) != 0 || len(p2.targets) != 0 {
		t.Errorf("the first two hits on a page are earned: %d, %d targets", len(p1.targets), len(p2.targets))
	}
	if len(p3.targets) != 1 {
		t.Errorf("the third hit on the SAME page must be repaired, got %d targets (page_hits=%d)",
			len(p3.targets), p3.pageHits)
	}
}

// The prompt must name the sentences and must NOT teach the shape it is removing.
func TestRepairPromptCarriesNoExampleOfTheBannedForm(t *testing.T) {
	plan := planNegationRepairs(sectionContent(), nil, 0, 0)
	prompt := negationRepairPrompt(plan.targets)
	if !strings.Contains(prompt, "The registry shows you what's possible") {
		t.Error("the prompt must quote the sentence being rewritten")
	}
	if !strings.Contains(prompt, "field:") || !strings.Contains(prompt, "replacements") {
		t.Error("the prompt must ask for field-addressed replacements")
	}
	// The instruction text itself (everything before the quoted sentences) must
	// carry no example of the construction: "the example is the instruction".
	instructions := prompt[:strings.Index(prompt, "Sentences:")]
	if hits := datahelpers.ScanDefineByNegation(instructions); len(hits) > 0 {
		t.Errorf("the repair instructions must not demonstrate the construction they remove: %+v", hits)
	}
	if strings.Contains(strings.ToLower(instructions), "house voice") {
		t.Error("the repair prompt must not quote the house-voice rule — its own text carries the form")
	}
}

func TestMatchTargetIgnoresRenamedField(t *testing.T) {
	plan := planNegationRepairs(sectionContent(), nil, 0, 0)
	// A model that renamed the field but copied the sentence: still matched.
	if _, ok := matchTarget(plan.targets, "title", "The registry shows you what's possible, not what survives production."); !ok {
		t.Error("the sentence is the identity; a renamed field must still match")
	}
	// A model that echoed a sentence nobody sent: refused, never spliced blind.
	if _, ok := matchTarget(plan.targets, "headline", "Something we never asked about, not this."); ok {
		t.Error("a sentence that was never a target must not match")
	}
	// Whitespace and curly-quote drift must not break the match.
	if _, ok := matchTarget(plan.targets, "headline", "The registry shows you what’s possible,   not what survives production."); !ok {
		t.Error("a re-typed apostrophe must still match")
	}
}

// The splice writes into the very map the renderer reads, and touches only the
// sentence — the rest of the field, markup included, is preserved.
func TestSplicePreservesTheRestOfTheField(t *testing.T) {
	content := sectionContent()
	var target negationTarget
	for _, tg := range planNegationRepairs(content, nil, 0, 0).targets {
		if tg.Field == "body" {
			target = tg
		}
	}
	if target.Field == "" {
		t.Fatal("expected a target in the body field")
	}
	updated := strings.Replace(target.text, target.Sentence, "Talk to the people who run the pipeline.", 1)
	target.set(updated)

	got := content["body"].(string)
	if !strings.Contains(got, "<p>We run 1,600 orchestrations a day.</p>") {
		t.Errorf("the untouched paragraph must survive verbatim: %q", got)
	}
	if !strings.Contains(got, "Talk to the people who run the pipeline.") {
		t.Errorf("the rewrite must be spliced in: %q", got)
	}
	if strings.Contains(got, "not the ones who catalogued it") {
		t.Errorf("the construction must be gone: %q", got)
	}
	if !strings.HasPrefix(got, "<p>") || !strings.HasSuffix(got, "</p>") {
		t.Errorf("markup must be preserved: %q", got)
	}
}

func TestActionReturnsCleanMarkerAndNeverErrorsOnCleanContent(t *testing.T) {
	params := ActionParams{
		Logger: zap.NewNop(),
		CollectedData: map[string]interface{}{
			"generated_content": map[string]interface{}{"result": map[string]interface{}{
				"headline": "Every agent definition running in our production fleet",
				"body":     "<p>We run 1,600 orchestrations a day across 13 live systems.</p>",
			}},
		},
		StepConfig: models.Step{Config: map[string]interface{}{}},
	}
	out, err := RewriteNegationsAction(nil, params)
	if err != nil {
		t.Fatalf("the gate must never fail the step: %v", err)
	}
	m := out.(map[string]interface{})
	if m["status"] != "clean" {
		t.Errorf("expected a clean marker, got %v", m)
	}
	if _, stamped := params.CollectedData[copyGateMarkerKey]; !stamped {
		t.Error("the marker must be stamped on collected data even when clean — an absent marker is how we tell 'ran and found nothing' from 'never ran'")
	}
}

func TestActionIsInertWithoutContent(t *testing.T) {
	params := ActionParams{Logger: zap.NewNop(), CollectedData: map[string]interface{}{},
		StepConfig: models.Step{Config: map[string]interface{}{}}}
	out, err := RewriteNegationsAction(nil, params)
	if err != nil {
		t.Fatalf("missing content must not fail the step: %v", err)
	}
	if out.(map[string]interface{})["status"] != "no_content" {
		t.Errorf("expected no_content, got %v", out)
	}
}

// TWO constructions in ONE field must both survive. Every target of a field
// shares the same captured original, so splicing each against that original and
// writing the whole field back makes the accepted rewrites overwrite each other
// — and the marker still reports them all as rewritten.
//
// MEASURED IN PRODUCTION 2026-08-21 (orch ce002822,
// webdesign.co.uk/tool-social-card-guide): six accepted replacements, all into
// one `content` field, hits_before 8 -> hits_after 7. One net repair out of six,
// reported as six. This test is that page, reduced.
func TestTwoTargetsInOneFieldBothSurvive(t *testing.T) {
	content := map[string]interface{}{
		"content": "<p>The two systems overlap rather than compete.</p>" +
			"<p>Everything past that is refinement rather than requirement.</p>",
	}
	plan := planNegationRepairs(content, nil, 0, 0)
	if len(plan.targets) != 2 {
		t.Fatalf("expected two targets in one field, got %d", len(plan.targets))
	}
	if plan.targets[0].Field != plan.targets[1].Field {
		t.Fatalf("this test is only meaningful when both targets share a field: %q vs %q",
			plan.targets[0].Field, plan.targets[1].Field)
	}

	// Mimic the accept-and-splice loop, which is what the action does once a
	// replacement has passed every check.
	spliced := map[string]string{}
	replacements := map[string]string{
		"The two systems overlap rather than compete.":            "The two systems overlap.",
		"Everything past that is refinement rather than requirement.": "Everything past that is refinement.",
	}
	for _, tg := range plan.targets {
		to, ok := replacements[tg.Sentence]
		if !ok {
			t.Fatalf("no replacement fixture for %q", tg.Sentence)
		}
		base, seen := spliced[tg.Field]
		if !seen {
			base = tg.text
		}
		updated := strings.Replace(base, tg.Sentence, to, 1)
		if updated == base {
			t.Fatalf("splice missed for %q", tg.Sentence)
		}
		spliced[tg.Field] = updated
		tg.set(updated)
	}

	got := content["content"].(string)
	if strings.Contains(got, "rather than") {
		t.Errorf("a construction survived — the splices raced: %q", got)
	}
	for _, want := range []string{"The two systems overlap.", "Everything past that is refinement."} {
		if !strings.Contains(got, want) {
			t.Errorf("rewrite lost: %q missing from %q", want, got)
		}
	}
	if n := len(datahelpers.ScanDefineByNegation(got)); n != 0 {
		t.Errorf("expected 0 constructions after both repairs, got %d", n)
	}
}
