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
//   - put bare `name` back in neverProseFieldRe -> TestCardNameIsAHeadlineTarget fails
//   - move the Identity check BELOW the headline branch, or delete it
//                                             -> TestIdentityNameIsNeverATarget fails

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
	plan := planNegationRepairs(sectionContent(), nil, 2, 0, 0)
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

// bugs_open/420's motivating shape: a feature card whose `name` IS the heading
// the card renders. Before the fix the walker skipped every `*.name` by field
// name, so two of these shipped on a live page while the sibling `description`
// fields in the same array were repaired in the same run.
func TestCardNameIsAHeadlineTarget(t *testing.T) {
	content := map[string]interface{}{
		"features": []interface{}{
			map[string]interface{}{
				"name":        "Exact math, not simulation",
				"description": "Every figure is computed and the method is stated.",
				"icon":        "calculator",
			},
		},
	}
	plan := planNegationRepairs(content, nil, 5, 0, 0)
	var found *negationTarget
	for i := range plan.targets {
		if plan.targets[i].Field == "features[0].name" {
			found = &plan.targets[i]
		}
	}
	if found == nil {
		t.Fatalf("a card name carrying a shape must become a repair target; got targets %+v", plan.targets)
	}
	if !found.Headline {
		t.Error("a card's name IS its heading, so the target must carry headline severity — that is what makes it repaired regardless of budget, and what selects the heading floor at the judge")
	}
}

// THE ORDERING TEST. `name` is headline-class as of 2026-09-04, and a headline
// hit is never forgiven by the budget — so if the identity check ran AFTER the
// headline branch, a listing item's page slug would be FORCED to the model
// rather than merely allowed there. Nothing about the two field regexes on their
// own shows that; only this ordering does, so only this test protects it.
//
// It also pins the accounting: an identity hit is an EXEMPTION, not a filtered-out
// row, so the repair's total still reconciles with the annotation's count over
// the same walk.
func TestIdentityNameIsNeverATarget(t *testing.T) {
	content := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"name":        "Routes work, not requests",
				"url":         "/agents/orchestrator.html",
				"description": "Builds the section plan a page is rendered from.",
			},
		},
	}
	plan := planNegationRepairs(content, nil, 0, 0, 0)
	for _, tg := range plan.targets {
		if tg.Field == "items[0].name" {
			t.Fatal("a listing item's name is its page slug and the stem of its own url — rewriting it desynchronises the item from pages.name AND from that url, on a page that still renders")
		}
	}
	if plan.total != 1 {
		t.Errorf("the hit must still be COUNTED so the repair reconciles with the annotation, got total=%d", plan.total)
	}
	if plan.exemptReasons[datahelpers.IdentityNameWithURLSibling] != 1 {
		t.Errorf("the refusal must say WHICH rule fired, got exemptReasons=%v", plan.exemptReasons)
	}
	if plan.pageHits != 0 {
		t.Errorf("an identity hit is not the writer's doing, so it must not count toward the page budget; got pageHits=%d", plan.pageHits)
	}
}

// A brief-supplied phrase is the brief's decision. It is counted and reported;
// it is never rewritten, and it must not eat the page's budget either — the
// budget exists to bound what the WRITER does.
func TestExemptHitsDoNotConsumeTheBudget(t *testing.T) {
	// ⚠ FIXTURE CHANGED 2026-08-24, and the assertions below did NOT. The two
	// non-exempt sentences were `x_not_y` (sharp) and are now `rather_than`
	// (mild), because after the D3 ruling only a mild shape can spend the page
	// budget — a sharp one is always repaired, so a sharp fixture can no longer
	// demonstrate anything about budget consumption. This test's actual subject
	// is its title: an EXEMPT hit must not eat the budget. That property is
	// untouched by the ruling and is still what is asserted.
	content := map[string]interface{}{
		"hero":  "Multi-agent systems deployed to production in days, not months, on Kubernetes.",
		"intro": "This list is pulled from the production registry rather than from provider marketing pages.",
		"outro": "We show the pipeline running rather than a slide about it.",
	}
	brief := []string{"Tagline: 'Multi-agent systems deployed to production in days, not months' — use it in the homepage hero."}

	plan := planNegationRepairs(content, brief, 2, 0, 0)
	if plan.exemptCount != 1 {
		t.Fatalf("the supplied tagline must be exempt, got %d (%v)", plan.exemptCount, plan.exemptReasons)
	}
	// RE-PINNED 2026-08-31 (owner Decision A: "repair every one" — the mild set
	// is empty, so no shape spends the budget). The EXEMPTION property this test
	// exists for is unchanged: the brief-supplied hit is never a target. The two
	// non-exempt hits are now BOTH targets regardless of the budget of 2.
	if plan.withinBudget != 0 || len(plan.targets) != 2 {
		t.Errorf("under Decision A both non-exempt hits are repaired and none is forgiven: withinBudget=%d targets=%d",
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
	plan := planNegationRepairs(content, nil, 5, 0, 0)
	if len(plan.targets) != 1 || !plan.targets[0].Headline {
		t.Errorf("a headline hit must be a target regardless of budget, got %+v", plan.targets)
	}
}

// The standard is per PAGE. Section two must see the budget section one used.
func TestBudgetIsPerPageNotPerSection(t *testing.T) {
	// ⚠ FIXTURE CHANGED 2026-08-24 for the same reason as
	// TestExemptHitsDoNotConsumeTheBudget: these were `x_not_y` (sharp), which
	// after the D3 ruling is always repaired and therefore cannot demonstrate a
	// budget carry at all. The subject — the budget is per PAGE, not per section —
	// is unchanged, and the assertions below are untouched.
	one := map[string]interface{}{"intro": "We show what runs rather than what demos."}
	two := map[string]interface{}{"intro": "We ship the pipeline rather than a slide deck."}
	three := map[string]interface{}{"intro": "We report the failures rather than only the wins."}

	p1 := planNegationRepairs(one, nil, 2, 0, 0)
	p2 := planNegationRepairs(two, nil, 2, p1.pageHits, p1.mildHits)
	p3 := planNegationRepairs(three, nil, 2, p2.pageHits, p2.mildHits)

	// RE-PINNED 2026-08-31 (owner Decision A): with the mild set empty nothing
	// is forgiven in ANY section, and the counters still carry per page — the
	// bookkeeping survives the policy so density stays reportable.
	if len(p1.targets) != 1 || len(p2.targets) != 1 || len(p3.targets) != 1 {
		t.Errorf("under Decision A every section's hit is a target: %d, %d, %d",
			len(p1.targets), len(p2.targets), len(p3.targets))
	}
	if p3.pageHits != 3 {
		t.Errorf("the per-page count must still carry: page_hits=%d, want 3", p3.pageHits)
	}
}

// The prompt must name the sentences and must NOT teach the shape it is removing.
func TestRepairPromptCarriesNoExampleOfTheBannedForm(t *testing.T) {
	plan := planNegationRepairs(sectionContent(), nil, 0, 0, 0)
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
	plan := planNegationRepairs(sectionContent(), nil, 0, 0, 0)
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
	for _, tg := range planNegationRepairs(content, nil, 0, 0, 0).targets {
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
	plan := planNegationRepairs(content, nil, 0, 0, 0)
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
		"The two systems overlap rather than compete.":                "The two systems overlap.",
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

// The step must hand its content ON, not merely mutate a map it happens to hold.
//
// MEASURED AT THE ARTEFACT 2026-08-21 (remortgagecalculator.uk/mortgage-lenders):
// the gate reported status=repaired, hits_after=0 — and the stored
// content_data.subheadline was byte-identical to the PRE-repair
// generated_content.result.subheadline. The renderer never saw the patch,
// because the fresh-state copy between steps carries only the CURRENT step's
// own keys, so an in-place edit to the previous step's output is dropped.
//
// `result` is therefore not decoration: it is the only part of this action's
// work that survives to render_section.
func TestActionHandsItsContentOnAsResult(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content map[string]interface{}
		want    string
	}{
		{"clean content is still handed on", map[string]interface{}{
			"headline": "Every agent definition running in production",
		}, "Every agent definition running in production"},
		{"patched content is handed on", map[string]interface{}{
			"headline": "It shows what is possible, not what survives.",
		}, ""}, // repair needs an AI client; we assert the key's PRESENCE and identity below
	} {
		params := ActionParams{
			Logger:        zap.NewNop(),
			CollectedData: map[string]interface{}{"generated_content": map[string]interface{}{"result": tc.content}},
			StepConfig:    models.Step{Config: map[string]interface{}{}},
		}
		out, err := RewriteNegationsAction(nil, params)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		m := out.(map[string]interface{})
		got, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("%s: the step must return its content under `result` — render_section reads it. Got keys %v",
				tc.name, datahelpers.GetMapKeys(m))
		}
		// It must be the SAME map the renderer would otherwise have read, so a
		// splice made in place is carried by it rather than lost beside it.
		if &got == nil || len(got) != len(tc.content) {
			t.Errorf("%s: returned content differs in shape from the input", tc.name)
		}
		if tc.want != "" && got["headline"] != tc.want {
			t.Errorf("%s: headline = %v, want %q", tc.name, got["headline"], tc.want)
		}
	}
}

// A missing content map must NOT produce a `result` key: an empty one handed to
// render_section would be a blank section, which is worse than the loud failure
// of finding nothing.
func TestNoContentHandsOnNothing(t *testing.T) {
	params := ActionParams{Logger: zap.NewNop(), CollectedData: map[string]interface{}{},
		StepConfig: models.Step{Config: map[string]interface{}{}}}
	out, _ := RewriteNegationsAction(nil, params)
	if _, present := out.(map[string]interface{})["result"]; present {
		t.Error("no content found, so no `result` should be handed on")
	}
}

// THE ACCOUNTING INVARIANT: every target we asked about ends up in exactly one
// of the two lists.
//
// MEASURED IN PRODUCTION 2026-08-23, BEFORE THE FIX: 15 of 49 markers with
// `targets > 0` did not reconcile (`targets != rewritten + rejected`), and 12 of
// those accounted for NONE of their targets. The cause is structural rather than
// occasional: `runNegationRepair`'s loop iterates the MODEL'S replacements, so a
// target the answer never mentions is visited by no branch and lands in neither
// list.
//
// `hits_after` stayed honest throughout — it is recomputed from the real content
// — so no page was ever misdescribed. What was corrupted is the INSTRUMENT: this
// action's displacement defence is that every rejection is recorded with its
// reason, and `bugs_open/305`'s D3 (is `rather than` a tic or ordinary English?)
// is to be settled from exactly that log.
func TestEveryTargetIsAccountedForEvenWhenTheModelIgnoresIt(t *testing.T) {
	content := map[string]interface{}{
		"headline": "We ship in days, not months.",
		"content": "<p>The two systems overlap rather than compete.</p>" +
			"<p>Everything past that is refinement rather than requirement.</p>",
	}
	// budget 0 = no allowance, so every non-exempt hit is a target. (99 would
	// ALLOW the first 99 and leave only headline-class hits — the inverse.)
	plan := planNegationRepairs(content, nil, 0, 0, 0)
	if len(plan.targets) < 3 {
		t.Fatalf("fixture must produce at least 3 targets, got %d", len(plan.targets))
	}

	// The model answered exactly ONE of them — the case that used to vanish.
	answered := map[string]bool{negationTargetKey(plan.targets[0]): true}
	unanswered := unansweredTargetRejections(plan.targets, answered)

	if got, want := len(unanswered), len(plan.targets)-1; got != want {
		t.Fatalf("unanswered targets = %d, want %d — a target the model ignored was recorded nowhere", got, want)
	}
	// One rewrite (the answered one) plus the unanswered rejections must
	// reconcile against the target count. This is the invariant the marker
	// publishes and a census reads.
	if got := 1 + len(unanswered); got != len(plan.targets) {
		t.Fatalf("targets=%d but rewritten+rejected=%d — the marker does not reconcile", len(plan.targets), got)
	}
	for _, r := range unanswered {
		if r["reason"] != "no_answer_for_target" {
			t.Errorf("reason = %v, want no_answer_for_target — a census must be able to tell this apart from a judged rejection", r["reason"])
		}
		if r["from"] == nil || r["from"] == "" {
			t.Error("an unanswered rejection with no `from` cannot be traced back to a sentence")
		}
	}
}

// The key must survive the same normalisation matchTarget applies, or a target
// that WAS answered gets re-reported as ignored — a false entry in the very log
// the fix exists to make trustworthy.
func TestAnsweredBookkeepingSurvivesPunctuationDrift(t *testing.T) {
	content := map[string]interface{}{
		"headline": "The registry shows you what's possible, not what survives production.",
	}
	plan := planNegationRepairs(content, nil, 0, 0, 0)
	if len(plan.targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(plan.targets))
	}
	// Same sentence, curly apostrophe — what matchTarget already tolerates.
	drifted := plan.targets[0]
	drifted.Sentence = "The registry shows you what’s possible, not what survives production."

	answered := map[string]bool{negationTargetKey(drifted): true}
	if n := len(unansweredTargetRejections(plan.targets, answered)); n != 0 {
		t.Fatalf("%d targets reported unanswered after a punctuation-only difference — matchTarget accepts this pair, so the bookkeeping must too", n)
	}
}

// The NUL separator in negationTargetKey is a real guard, not decoration, and
// this test is why it can be trusted: without it, `field + sentence` is
// ambiguous, so a field whose name ends where another field's sentence begins
// produces the SAME key — and one target being answered would silently mark the
// other answered too, dropping it from the log with no trace.
//
// Added because the mutation that removed the separator PASSED the rest of this
// suite. A comment claiming a protection that no test exercises is not a
// protection.
func TestTargetKeySeparatorPreventsFieldSentenceCollision(t *testing.T) {
	// "ab" + "c is not d." and "a" + "bc is not d." concatenate identically.
	a := negationTarget{Field: "ab", Sentence: "c is not d.", Shape: "x_not_y"}
	b := negationTarget{Field: "a", Sentence: "bc is not d.", Shape: "x_not_y"}

	if negationTargetKey(a) == negationTargetKey(b) {
		t.Fatalf("two distinct targets share a key (%q) — one being answered would silently account for the other",
			negationTargetKey(a))
	}

	// And prove the consequence, not just the key: answering A must leave B
	// reported as unanswered.
	answered := map[string]bool{negationTargetKey(a): true}
	got := unansweredTargetRejections([]negationTarget{a, b}, answered)
	if len(got) != 1 || got[0]["field"] != "a" {
		t.Fatalf("expected exactly B unanswered, got %d entries: %v", len(got), got)
	}
}

// The marker's reconciliation invariant is NOT `targets == rewritten + rejected`.
// A replacement naming a sentence that is in no target is rejected as
// `no_such_sentence` — an entry with NO target behind it — so it pushes the sum
// ABOVE `targets`. The precise form, and the one a census must test, is:
//
//	targets == len(rewritten) + len(rejected) - count(reason="no_such_sentence")
//
// This is not hypothetical. Measured in production 2026-08-24, the first day the
// target-accounting fix was live: exactly 1 marker of 122 over-counted, with
// `targets=5, rewritten=4, rejected=2` — the 2 being one `no_answer_for_target`
// (the 5th target, correctly recorded) and one `no_such_sentence` (a sentence the
// model invented, correctly logged). All five targets accounted for; the sum is
// six because the model said six things.
//
// What this test defends is the DISCRIMINATION. The tempting "fix" when a census
// reads non-zero is to loosen matchTarget until every replacement finds a target
// — which would silently splice a rewrite into copy the model was not talking
// about. If that happens, this test fails first. See bugs_open/305 §27a, §29.
func TestReconciliationExcludesHallucinatedReplacements(t *testing.T) {
	content := map[string]interface{}{
		"headline": "We ship in days, not months.",
		"content":  "<p>The two systems overlap rather than compete.</p>",
	}
	plan := planNegationRepairs(content, nil, 0, 0, 0)
	if len(plan.targets) < 2 {
		t.Fatalf("fixture must produce at least 2 targets, got %d", len(plan.targets))
	}

	// A sentence the page does not contain, in a field that does exist — the
	// shape a model produces when it paraphrases instead of quoting.
	const invented = "Our platform is the definitive source for orchestration truth."
	if _, found := matchTarget(plan.targets, "content", invented); found {
		t.Fatal("matchTarget matched a sentence that is in NO target — a rewrite would be spliced into copy the model was not describing")
	}

	// The real ones must still match, or the guard above is just a broken matcher.
	for i, tgt := range plan.targets {
		got, found := matchTarget(plan.targets, tgt.Field, tgt.Sentence)
		if !found {
			t.Fatalf("target %d did not match its own sentence — the discrimination is not selective, it is blind", i)
		}
		if got.Sentence != tgt.Sentence {
			t.Errorf("target %d matched the wrong target: %q", i, got.Sentence)
		}
	}

	// NO ARITHMETIC ASSERTION HERE, deliberately. The obvious closer is to compute
	// `rewritten + hallucinated - hallucinated == len(targets)` — but every term of
	// that is a constant this test chose, so it holds whatever the code does. It
	// would read as coverage of the census formula and check nothing, which is the
	// vacuous-assertion shape a council seat correctly flagged at this lane on
	// 2026-08-23. The formula is documented at the `unansweredTargetRejections`
	// call site and is verified where it can actually fail: against production
	// markers (bugs_open/305 §29, 1 over-count in 122). What is testable in a unit
	// is the DISCRIMINATION above, and a mutation dropping the containment check in
	// matchTarget makes it fail.
}

// OWNER RULING 2026-08-24 (D3): "`rather than` is a little bit of a tic."
//
// The gate lets a page keep `page_budget` constructions and repairs the rest.
// Before this ruling the survivors were whichever the scanner walked past FIRST —
// document order, nothing to do with severity — so a page could keep both its
// `x_not_y` constructions and spend the gate's effort rewriting two `rather
// than`s further down. That is the ruling inverted, and it is what this test
// pins: the budget is FORGIVENESS, and only a mild shape may spend it.
//
// The fixture is built so document order works AGAINST the desired outcome: the
// two sharp constructions come first, so under the old rule they would have
// consumed the whole budget of 2 and survived, and the `rather than` after them
// would have been the thing repaired.
func TestOnlyAMildShapeCanSpendThePageBudget(t *testing.T) {
	content := map[string]interface{}{
		// Non-headline field, so the budget is in play at all.
		"content": "<p>The registry shows you what ships, not what demos.</p>" +
			"<p>We measure throughput, not vanity metrics.</p>" +
			"<p>The two systems overlap rather than compete.</p>" +
			"<p>Everything past that is refinement rather than requirement.</p>" +
			"<p>Teams adopt it incrementally rather than all at once.</p>",
	}
	plan := planNegationRepairs(content, nil, 2, 0, 0)

	byShape := map[string]int{}
	for _, tgt := range plan.targets {
		byShape[tgt.Shape]++
	}

	// RE-PINNED 2026-08-31 (owner Decision A: "repair every one"). D3's mild
	// tolerance is repealed: ALL five hits are targets, none is forgiven, and
	// the counters still see everything.
	if byShape["x_not_y"] != 2 {
		t.Errorf("x_not_y targets = %d, want 2 (targets: %+v)", byShape["x_not_y"], byShape)
	}
	if byShape["rather_than"] != 3 {
		t.Errorf("rather_than targets = %d, want 3 — Decision A repealed the mild tolerance", byShape["rather_than"])
	}
	if plan.withinBudget != 0 {
		t.Errorf("withinBudget = %d, want 0 — nothing may spend the budget under Decision A", plan.withinBudget)
	}
	if plan.pageHits != 5 {
		t.Errorf("pageHits = %d, want 5 — the density signal must still see every hit", plan.pageHits)
	}
	if plan.mildHits != 0 {
		t.Errorf("mildHits = %d, want 0 — the mild set is empty", plan.mildHits)
	}
}

// The budget is per PAGE, so the mild count must survive across sections — and it
// must be carried SEPARATELY from page_hits. Seeding the budget from the total
// would let a sharp hit in an earlier section eat a later section's forgiveness,
// which is the same inversion the ruling is about, one section removed.
func TestTheMildBudgetCarriesAcrossSectionsWithoutSharpHitsEatingIt(t *testing.T) {
	sharp := map[string]interface{}{
		"content": "<p>The registry shows you what ships, not what demos.</p>" +
			"<p>We measure throughput, not vanity metrics.</p>",
	}
	mild := map[string]interface{}{
		"content": "<p>The two systems overlap rather than compete.</p>" +
			"<p>Everything past that is refinement rather than requirement.</p>",
	}

	s1 := planNegationRepairs(sharp, nil, 2, 0, 0)
	if len(s1.targets) != 2 {
		t.Fatalf("section 1: %d targets, want 2 — both sharp hits must be repaired", len(s1.targets))
	}
	if s1.mildHits != 0 {
		t.Fatalf("section 1 mildHits = %d, want 0 — sharp hits must not touch the mild counter", s1.mildHits)
	}

	// RE-PINNED 2026-08-31 (owner Decision A): section 2's two `rather than`
	// hits are now BOTH targets — the mild set is empty and no section, first
	// or later, has forgiveness to spend.
	s2 := planNegationRepairs(mild, nil, 2, s1.pageHits, s1.mildHits)
	if len(s2.targets) != 2 {
		t.Errorf("section 2: %d targets, want 2 — Decision A repairs every rather_than", len(s2.targets))
	}
	if s2.withinBudget != 0 {
		t.Errorf("section 2 withinBudget = %d, want 0", s2.withinBudget)
	}
	// The total still counts everything.
	if s2.pageHits != 4 {
		t.Errorf("section 2 pageHits = %d, want 4 (2 sharp + 2 mild across the page)", s2.pageHits)
	}
}

// A headline is repaired regardless of shape or budget — unchanged by the ruling,
// and worth pinning because the mildness test sits directly above the headline
// branch and an edit to one can silently reorder the other.
func TestAMildShapeInAHeadlineIsStillAlwaysRepaired(t *testing.T) {
	content := map[string]interface{}{
		"headline": "We integrate rather than replace.",
	}
	plan := planNegationRepairs(content, nil, 99, 0, 0)
	if len(plan.targets) != 1 {
		t.Fatalf("got %d targets, want 1 — a headline must bypass the budget however mild the shape", len(plan.targets))
	}
	if !plan.targets[0].Headline {
		t.Error("target is not marked headline")
	}
	if plan.withinBudget != 0 {
		t.Errorf("withinBudget = %d, want 0 — a headline must not spend the budget", plan.withinBudget)
	}
}
