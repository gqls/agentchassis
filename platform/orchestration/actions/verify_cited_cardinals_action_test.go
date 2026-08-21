// FILE: platform/orchestration/actions/verify_cited_cardinals_action_test.go
//
// Every fixture below is REAL: the prose and the premise fields are the exact
// strings live on the estate on 2026-08-21, read from site_specs. That matters
// more than usual here, because the whole design question was which shapes are
// quantity CLAIMS and which are ordinary English — and that is not a question
// you can answer from invented examples. The negative controls are the ones
// that nearly killed the gate: two of the three legitimate specifics on this
// estate are spelled-out numerals, so a rule that bans word numerals outright
// passes the bug and destroys the artefact.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ---- the live corpus, verbatim ------------------------------------------

const (
	// leopardessconsulting.co.uk — the defect (bugs_open/335). "eight" is
	// nowhere in the cited field; the true site count was 23.
	leopardessTrustThreshold = "High — the buyer is a CTO or VP Engineering authorising a significant technical engagement, so they need to verify that Leopardess has built and operated real production systems on the named infrastructure before initiating contact; the site addresses this by naming specific technologies, quoting verifiable metrics, and pointing to live deployed systems rather than case study narratives alone."
	leopardessBadPoint       = "Your agent system will run on Kubernetes, Kafka, and Postgres — the same stack that runs eight live sites built by this team — delivering working orchestration infrastructure in days, not months."

	// webdesign.co.uk — a legitimate word numeral, present verbatim.
	webdesignValueProp = "A free, client-side toolbox and learning library for front-end practitioners — sixty-three single-purpose tools and thirty paired articles, no account required, nothing uploaded, nothing stored."
	webdesignGoodPoint = "You can run any of the sixty-three tools here right now, in your browser, without an account and without anything leaving your machine."

	// robot-hands.com — a legitimate word numeral, and a legitimate digit
	// range whose dash differs between premise ("2-3") and point ("2–3").
	robotHandsValueProp = "Robot-Hands.com is the vendor-neutral technical reference platform where industrial automation engineers compare gripper technologies across six actuation types, calculate application parameters with documented tools, and run MatchMatrix to specify the right end-effector — faster and with more confidence than navigating manufacturer datasheets alone."
	robotHandsGoodPoint = "Compare gripper technologies across six actuation types — pneumatic, electric, magnetic, vacuum, soft-robotic, adhesive — with consistent benchmarking, then calculate application parameters and run MatchMatrix to confirm your selection."
	robotHandsRecurring = "Engineers return because the catalog grows (new gripper models and manufacturers added continuously), tool outputs change as they input new application parameters for new projects, and the Learning Center publishes 2-3 new technical articles per month covering technology comparisons and application engineering scenarios directly relevant to live specification decisions."
	robotHandsDashPoint = "The catalog grows continuously, tool outputs change as you input new application parameters, and the Learning Center publishes 2–3 technical articles per month — so the platform remains useful across every project, not just the one you arrived with."
)

func cardinalParams(t *testing.T, obj map[string]interface{}, source interface{}, extra map[string]interface{}) ActionParams {
	t.Helper()
	cfg := map[string]interface{}{
		"object_field": "analysis.ordering",
		"items_key":    "lead_with",
		"source_field": "premise.strategy",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "gate"},
		CollectedData: map[string]interface{}{
			"analysis": map[string]interface{}{"ordering": obj},
			"premise":  map[string]interface{}{"strategy": source},
		},
		StepConfig: models.Step{Config: cfg},
	}
}

func point(rank int, field, text string) map[string]interface{} {
	return map[string]interface{}{"rank": rank, "from_field": field, "point": text}
}

func ordering(points ...map[string]interface{}) map[string]interface{} {
	items := make([]interface{}, 0, len(points))
	for _, p := range points {
		items = append(items, p)
	}
	return map[string]interface{}{
		"reader_goal": "a goal", "lead_with": items, "degraded": false,
	}
}

// ---- the defect ----------------------------------------------------------

// The whole reason this action exists. If this test ever passes clean, the
// gate has stopped doing its job.
func TestRejectsTheLeopardessDefect(t *testing.T) {
	p := cardinalParams(t,
		ordering(point(1, "trust_threshold", leopardessBadPoint)),
		map[string]interface{}{"trust_threshold": leopardessTrustThreshold}, nil)

	out, err := VerifyCitedCardinalsAction(context.Background(), p)
	if err == nil {
		t.Fatalf("the rank-1 false claim was accepted; got %+v", out)
	}
	if !strings.Contains(err.Error(), "trust_threshold") || !strings.Contains(err.Error(), "8") {
		t.Errorf("error must name the cited field and the unsourced value, got: %v", err)
	}
}

// A DIGITS-ONLY gate — verify_report_prose's proseNumRe, the nearest precedent
// — cannot see this defect at all. This test pins the reason the word
// vocabulary is not optional: delete it and the motivating case walks through.
func TestDigitsOnlyScanWouldHaveMissedTheDefect(t *testing.T) {
	if got := cardinalDigitRe.FindAllString(leopardessBadPoint, -1); len(got) != 0 {
		t.Fatalf("premise of this test is broken: the defect point does contain digits %v", got)
	}
	if !cardinalsIn(leopardessBadPoint, cardinalPointWordRe, cardinalUnits)["8"] {
		t.Error(`the word-numeral scan must recover "eight" as 8`)
	}
}

// ---- the negative controls ----------------------------------------------

// The bug file proposed gaswholesalers as the negative control. It is the WRONG
// one: none of its six points contains a cardinal at all, so it passes whatever
// the rule is. These three are the points that can actually discriminate.
func TestLegitimateSpecificsSurvive(t *testing.T) {
	for _, tc := range []struct {
		name, field, source, text string
	}{
		{"webdesign word numeral present verbatim", "value_proposition", webdesignValueProp, webdesignGoodPoint},
		{"robot-hands word numeral present verbatim", "value_proposition", robotHandsValueProp, robotHandsGoodPoint},
		{"robot-hands en-dash range vs hyphen premise", "recurring_value", robotHandsRecurring, robotHandsDashPoint},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := cardinalParams(t, ordering(point(1, tc.field, tc.text)),
				map[string]interface{}{tc.field: tc.source}, nil)
			out, err := VerifyCitedCardinalsAction(context.Background(), p)
			if err != nil {
				t.Fatalf("a premise-sourced specific was rejected: %v", err)
			}
			if out.(map[string]interface{})["verified"] != true {
				t.Errorf("expected verified=true, got %+v", out)
			}
		})
	}
}

// "one" and "zero" as article, pronoun and idiom. These four sentences are live
// on the estate and produced five of the six flags when the vocabulary included
// them — the measurement that put them out of it. The source here provably
// carries no cardinal at all, so nothing but the exclusion can be passing them.
func TestIndefiniteOneAndZeroAreNotQuantityClaims(t *testing.T) {
	const numberFreeSource = "Buyers return when the guidance stays current and the tools keep working."
	if len(cardinalsIn(numberFreeSource, cardinalSourceWordRe, cardinalSourceWords)) != 0 {
		t.Fatal("premise of this test is broken: the source fixture contains a cardinal")
	}
	for _, text := range []string{
		"The fuel budget forecaster and supplier comparison calculator are here when you are benchmarking suppliers — and the inquiry form is one click away when you have run the numbers.",
		"When a step fails, the system picks up where it stopped, so a production outage is a recoverable event, not a restart from zero.",
		"No existing platform combines a benchmarked catalog, an auditable methodology, and interactive tools in one workflow.",
		"Gas Wholesalers supplies wholesale gasoline, diesel, and natural gas — if your operation falls into one of those categories, you are in the right place.",
	} {
		p := cardinalParams(t, ordering(point(1, "recurring_value", text)),
			map[string]interface{}{"recurring_value": numberFreeSource}, nil)
		if _, err := VerifyCitedCardinalsAction(context.Background(), p); err != nil {
			t.Errorf("ordinary English rejected as a quantity claim: %q\n  %v", text, err)
		}
	}
}

// ---- mutation tests: prove the guard bites ------------------------------

func TestMutatedQuantitiesAreCaught(t *testing.T) {
	for _, tc := range []struct {
		name, field, source, text, want string
	}{
		{"inflated word numeral", "value_proposition", webdesignValueProp,
			"You can run any of the sixty-four tools here right now.", "64"},
		{"inflated digit numeral", "value_proposition", webdesignValueProp,
			"You can run any of the 64 tools here right now.", "64"},
		{"inflated range", "recurring_value", robotHandsRecurring,
			"The Learning Center publishes 5–6 technical articles per month.", "5"},
		{"the defect in digit form", "trust_threshold", leopardessTrustThreshold,
			"the same stack that runs 8 live sites built by this team", "8"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := cardinalParams(t, ordering(point(1, tc.field, tc.text)),
				map[string]interface{}{tc.field: tc.source}, nil)
			_, err := VerifyCitedCardinalsAction(context.Background(), p)
			if err == nil {
				t.Fatalf("mutant accepted: %q", tc.text)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error should name %s, got: %v", tc.want, err)
			}
		})
	}
}

// A quantity stated in the other form than the premise used is style, not
// fabrication: the premise did state it.
func TestCrossFormMatchIsAllowed(t *testing.T) {
	p := cardinalParams(t,
		ordering(point(1, "value_proposition", "You can run any of the 63 tools here right now.")),
		map[string]interface{}{"value_proposition": webdesignValueProp}, nil)
	if _, err := VerifyCitedCardinalsAction(context.Background(), p); err != nil {
		t.Errorf(`"63" should trace to the premise's "sixty-three": %v`, err)
	}
}

// ---- citation that names nothing ----------------------------------------

func TestCitingAFieldThatDoesNotExistIsReported(t *testing.T) {
	p := cardinalParams(t,
		ordering(point(1, "no_such_field", leopardessBadPoint)),
		map[string]interface{}{"trust_threshold": leopardessTrustThreshold}, nil)
	_, err := VerifyCitedCardinalsAction(context.Background(), p)
	if err == nil {
		t.Fatal("a cardinal citing a non-existent field must not pass")
	}
	if !strings.Contains(err.Error(), "not a field of the source") {
		t.Errorf("the message should say the cited field is absent, got: %v", err)
	}
}

// ---- drop mode -----------------------------------------------------------

func TestDropModeRemovesOnlyTheOffendingItem(t *testing.T) {
	p := cardinalParams(t,
		ordering(
			point(1, "trust_threshold", leopardessBadPoint),
			point(2, "value_proposition", webdesignGoodPoint),
		),
		map[string]interface{}{
			"trust_threshold":   leopardessTrustThreshold,
			"value_proposition": webdesignValueProp,
		},
		map[string]interface{}{"on_violation": "drop"})

	out, err := VerifyCitedCardinalsAction(context.Background(), p)
	if err != nil {
		t.Fatalf("drop mode must not fail the run: %v", err)
	}
	res := out.(map[string]interface{})
	if res["dropped"] != 1 {
		t.Errorf("expected 1 dropped, got %v", res["dropped"])
	}
	obj := res["object"].(map[string]interface{})
	kept := obj["lead_with"].([]interface{})
	if len(kept) != 1 {
		t.Fatalf("expected 1 surviving point, got %d", len(kept))
	}
	if got := kept[0].(map[string]interface{})["point"]; got != webdesignGoodPoint {
		t.Errorf("the wrong point survived: %v", got)
	}
	// The drop must be legible in the artefact itself, or it is exactly the
	// silent-mutation failure this estate keeps paying for.
	dropped, ok := obj["dropped_unsourced"].([]citedCardinalViolation)
	if !ok || len(dropped) != 1 {
		t.Fatalf("the removal was not recorded in the written object: %+v", obj["dropped_unsourced"])
	}
	if dropped[0].CitedField != "trust_threshold" || len(dropped[0].Unsourced) == 0 {
		t.Errorf("the record does not identify what was removed: %+v", dropped[0])
	}
}

// Drop mode must not produce a husk: an ordering with no points reads as
// complete and carries nothing.
func TestDropModeRefusesToEmptyTheArtefact(t *testing.T) {
	p := cardinalParams(t,
		ordering(point(1, "trust_threshold", leopardessBadPoint)),
		map[string]interface{}{"trust_threshold": leopardessTrustThreshold},
		map[string]interface{}{"on_violation": "drop"})

	if _, err := VerifyCitedCardinalsAction(context.Background(), p); err == nil {
		t.Fatal("dropping every item must fail rather than write an empty artefact")
	} else if !strings.Contains(err.Error(), "refusing to write an empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---- the source shape the live workflow actually supplies ----------------

// load_premise selects the strategy spec as `data::text`, so the source arrives
// as a JSON-encoded STRING, not an object. A gate that only handled objects
// would find no cited field, flag every cardinal, and look like it was working.
func TestJSONEncodedStringSourceIsParsed(t *testing.T) {
	p := cardinalParams(t,
		ordering(point(1, "value_proposition", webdesignGoodPoint)),
		`{"value_proposition":`+quoteJSON(webdesignValueProp)+`}`, nil)
	if _, err := VerifyCitedCardinalsAction(context.Background(), p); err != nil {
		t.Errorf("a JSON-string source must be parsed, not treated as opaque: %v", err)
	}
}

// A nested object under the cited field is flattened, not skipped: the model is
// shown the whole premise, so a value nested inside the cited field was
// genuinely available to it (robot-hands' competitive_position is one).
func TestNestedObjectSourceIsFlattened(t *testing.T) {
	p := cardinalParams(t,
		ordering(point(1, "competitive_position", "We benchmark across six actuation types.")),
		map[string]interface{}{
			"competitive_position": map[string]interface{}{
				"defensible_moat": "cross-technology depth across six actuation types",
			},
		}, nil)
	if _, err := VerifyCitedCardinalsAction(context.Background(), p); err != nil {
		t.Errorf("a nested cited field must be scanned, not skipped: %v", err)
	}
}

func TestConfigIsValidated(t *testing.T) {
	p := cardinalParams(t, ordering(point(1, "x", "text")), map[string]interface{}{},
		map[string]interface{}{"on_violation": "sometimes"})
	if _, err := VerifyCitedCardinalsAction(context.Background(), p); err == nil ||
		!strings.Contains(err.Error(), "on_violation") {
		t.Errorf("an unknown on_violation mode must be refused, got: %v", err)
	}
}

func quoteJSON(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return `"` + r.Replace(s) + `"`
}
