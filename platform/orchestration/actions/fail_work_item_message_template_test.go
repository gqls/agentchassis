// FILE: platform/orchestration/actions/fail_work_item_message_template_test.go
//
// bugs_open/440 / RFC_062 phase 3, owner ruling D1. Two things are pinned here:
// that the opt-in refusal template actually names the offending value AND the
// vocabulary, and that arming this action's config detector did not start
// telling a lie about the two keys that were already there.
//
// MUTATION CHECKS — written as instructions, not as a claim, deliberately.
// Both were RUN on 2026-09-03 and observed failing with the messages quoted
// below; the imperative form is what stays honest if a future edit makes one of
// them pass. (The first draft of this header said "each verified to fail before
// this file was committed" and was written BEFORE either had been run — a
// checked-looking sentence a reviewer cannot distinguish from a checked one.
// WRONG_CALLS 2026-09-03.)
//
//   - Drop "error_message" from FailWorkItemInputSpec.ConfigKeys and
//     TestFailWorkItemConfigKeysCoverTheKeysItActuallyReads MUST fail: the
//     detector then calls a key the action DOES read "silently ignored".
//     Observed: `the detector calls [error_message] unknown, but fail_work_item
//     reads them straight from StepConfig.Config`.
//   - Delete the <no value> guard in renderFailWorkItemMessage and
//     TestRenderFailWorkItemMessage_LoudOnPresentButNilKey MUST fail.
//     Observed: `a present-but-nil routing_reason rendered without complaint`.

package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// specWith builds the collected_data shape a page_rerender run carries, with
// `routing_reason` PRESENT and set to the given value.
//
// ⚠ It always sets the key, including for nil. The first cut of this helper
// skipped the assignment when the value was nil, so the "present but nil" test
// below silently exercised the ABSENT case instead and passed for the wrong
// reason — the two states are distinct at every layer of this bug (they are
// distinct in the evaluator too: see rerender_routing_gate_clause_test.go's
// four-state table) and a helper that collapses them cannot test either.
func specWith(routingReason interface{}) map[string]interface{} {
	return collectedWithSpec(map[string]interface{}{
		"page_name":      "about",
		"routing_reason": routingReason,
	})
}

// specWithoutRoutingKey is the legacy population's shape: no routing key at all.
func specWithoutRoutingKey() map[string]interface{} {
	return collectedWithSpec(map[string]interface{}{"page_name": "about"})
}

func collectedWithSpec(spec map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"spec":         spec,
			"work_item_id": "11111111-1111-1111-1111-111111111111",
		},
	}
}

// TestRefusalMessageNamesTheBadKeyAndTheVocabulary is owner ruling D1, as a test.
//
// D1: "a refused item routes to needs_human_review — message names the bad key
// AND the vocabulary". Both halves are asserted against a value that is NOT in
// the vocabulary, which is the only state that reaches this step.
func TestRefusalMessageNamesTheBadKeyAndTheVocabulary(t *testing.T) {
	const badKey = "tool_retirement" // a real one: 16 live items, bugs_open/440

	got, err := renderFailWorkItemMessage(
		livespec.RefuseUnknownRoutingKeyMessageTemplate(), specWith(badKey))
	if err != nil {
		t.Fatalf("the refusal message must render for the state that reaches it: %v", err)
	}

	if !strings.Contains(got, badKey) {
		t.Errorf("D1 requires the message to name the BAD KEY. A static literal cannot, "+
			"which is why error_message_template exists. Rendered: %s", got)
	}
	for _, name := range livespec.RerenderSectionReasonNames() {
		if !strings.Contains(got, name) {
			t.Errorf("D1 requires the message to name the VOCABULARY; %q is missing. "+
				"It is rendered from RerenderSectionReasonNames precisely so a sixth "+
				"value cannot be forgotten here. Rendered: %s", name, got)
		}
	}
	if strings.Contains(got, "{{") {
		t.Errorf("the template shipped un-rendered — the operator would read Go template "+
			"source instead of the offending key. Rendered: %s", got)
	}
	// The fallback must never be empty, or a failed render parks the item with
	// no explanation at all.
	if fb := livespec.RefuseUnknownRoutingKeyMessageFallback(); strings.TrimSpace(fb) == "" {
		t.Error("the static fallback is empty; a failed template render would park the item mute")
	}
}

// TestRenderFailWorkItemMessage_LoudOnPresentButNilKey pins failure mode 2.
//
// text/template's missingkey=error does NOT catch a key that is PRESENT and
// nil — it renders "<no value>". A refusal reading routing_reason = '<no value>'
// sends a human hunting for a key that is not the problem, so it must be an
// error and fall back, not a shrug.
func TestRenderFailWorkItemMessage_LoudOnPresentButNilKey(t *testing.T) {
	_, err := renderFailWorkItemMessage(
		livespec.RefuseUnknownRoutingKeyMessageTemplate(), specWith(nil))
	if err == nil {
		t.Fatal("a present-but-nil routing_reason rendered without complaint — the message " +
			"would report '<no value>' as the offending key")
	}
	if !strings.Contains(err.Error(), "<no value>") {
		t.Errorf("the error must name what it saw, so an operator can tell this case from a "+
			"parse failure; got: %v", err)
	}

	// The ABSENT case is a DIFFERENT mechanism — missingkey=error, not the
	// <no value> guard — and both must be loud. Asserted here so that deleting
	// either one cannot be masked by the other.
	_, err = renderFailWorkItemMessage(
		livespec.RefuseUnknownRoutingKeyMessageTemplate(), specWithoutRoutingKey())
	if err == nil {
		t.Fatal("an absent routing_reason rendered without complaint")
	}
	if strings.Contains(err.Error(), "<no value>") {
		t.Errorf("the absent case must be caught by missingkey=error, not by the <no value> "+
			"guard — if it reaches the guard, missingkey is no longer set; got: %v", err)
	}
}

// TestRenderFailWorkItemMessage_LoudOnMissingAndMalformed pins failure mode 1.
func TestRenderFailWorkItemMessage_LoudOnMissingAndMalformed(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
		data map[string]interface{}
		why  string
	}{
		{
			name: "does not parse",
			tmpl: "refused: {{.input_data.spec.routing_reason",
			data: specWith("tool_retirement"),
			why:  "an unclosed action must not ship as literal text in a refusal message",
		},
		{
			name: "names a path that is not there",
			tmpl: "refused: {{.input_data.spec.no_such_field}}",
			data: specWith("tool_retirement"),
			why:  "missingkey=error is set so an absent path is an error, not '<no value>'",
		},
		{
			name: "renders empty",
			tmpl: "{{if false}}unreachable{{end}}",
			data: specWith("tool_retirement"),
			why:  "an empty message parks the item with no explanation",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := renderFailWorkItemMessage(tc.tmpl, tc.data); err == nil {
				t.Errorf("expected a loud failure — %s", tc.why)
			}
		})
	}
}

// TestFailWorkItemConfigKeysCoverTheKeysItActuallyReads is the disconfirmable
// half, and the reason all three keys were declared in one move.
//
// checksConfig() is `CheckConfig || len(ConfigKeys) > 0`, so before this change
// the detector was OFF for fail_work_item (checked=false) and `error_message` /
// `status_override` were declared nowhere. Declaring only the NEW key would arm
// UnknownConfigKeys against those two and emit "silently ignored at execution"
// about keys the action demonstrably reads. The live step shape below is the one
// `[MEASURED 2026-09-03]` on six agents; the bogus key is the negative control,
// without which this test would pass on a detector that had been switched off
// again.
func TestFailWorkItemConfigKeysCoverTheKeysItActuallyReads(t *testing.T) {
	liveShape := map[string]interface{}{
		"work_item_id":    "input_data.work_item_id",
		"error_message":   "Content validation failed — needs human review",
		"status_override": "needs_human_review",
	}

	unknown, checked := datahelpers.UnknownConfigKeys("fail_work_item", liveShape)
	if !checked {
		t.Fatal("fail_work_item's config detector is OFF (checked=false) — an empty result " +
			"from an unarmed detector is not a pass. Declaring ConfigKeys is what arms it")
	}
	if len(unknown) != 0 {
		t.Errorf("the detector calls %v unknown, but fail_work_item reads them straight from "+
			"StepConfig.Config — the warning it emits ('silently ignored at execution') "+
			"would be FALSE for all seven live steps", unknown)
	}

	// The refusal step RFC_062 phase 3 adds, template key included.
	withTemplate := map[string]interface{}{
		"work_item_id":           "input_data.work_item_id",
		"error_message":          livespec.RefuseUnknownRoutingKeyMessageFallback(),
		"error_message_template": livespec.RefuseUnknownRoutingKeyMessageTemplate(),
		"status_override":        "needs_human_review",
	}
	if unknown, _ := datahelpers.UnknownConfigKeys("fail_work_item", withTemplate); len(unknown) != 0 {
		t.Errorf("phase 3's own refusal step carries unrecognised keys %v", unknown)
	}

	// NEGATIVE CONTROL: a key the action really does not read must still be
	// reported, or the assertions above are vacuous.
	bogus := map[string]interface{}{
		"work_item_id":     "input_data.work_item_id",
		"error_messsage":   "typo, three s's",
		"no_such_settingx": true,
	}
	unknown, checked = datahelpers.UnknownConfigKeys("fail_work_item", bogus)
	if !checked {
		t.Fatal("negative control: detector unarmed")
	}
	if len(unknown) != 2 {
		t.Errorf("negative control: expected both bogus keys reported, got %v — a detector "+
			"that reports nothing here would make the clean readings above meaningless", unknown)
	}
}

// TestErrorMessageTemplateIsAbsentFromEveryLiveStep records the RFC_022
// exemption's third condition as a runnable claim rather than a sentence.
//
// It cannot query the fleet from `go test`, so it pins the ENUMERATION QUERY
// instead — the thing a reviewer needs in order to re-check it, and the thing
// the owner ruling of 2026-07-29 §1 says must be enumerated rather than
// asserted. `[MEASURED 2026-09-03]` it returned 7 steps / 6 agents / 0 with
// `{{`.
func TestErrorMessageTemplateIsAbsentFromEveryLiveStep(t *testing.T) {
	const enumeration = `SELECT count(*) FILTER (WHERE s.value->'config' ? 'error_message_template') ` +
		`FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s ` +
		`WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL ` +
		`AND s.value->>'action'='fail_work_item'`

	if !strings.Contains(enumeration, "error_message_template") {
		t.Fatal("the enumeration no longer names the key it is supposed to count")
	}
	if !strings.Contains(enumeration, "is_snapshot") {
		t.Error("a fleet enumeration that does not exclude snapshots counts dead rows, and " +
			"would report live consumers that are not live")
	}
}
