package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

// The two payload shapes below are the LIVE ones, copied from the database rather
// than composed, because a fixture written to exercise a rule will exercise it.
// Both come from the 2026-08-12 census in bugs_open/213 §B/§D.

// fixerEnvelope is what 4 of the 14 completed dark_section_audit rows carry: the
// color-variable-fixer response, both repair steps reporting zero.
func fixerEnvelope(fixTotal, textTotal interface{}) map[string]interface{} {
	return map[string]interface{}{
		"response_status":      "complete",
		"response_received_at": "2026-08-11T13:38:18Z",
		"response": map[string]interface{}{
			"fix_result": map[string]interface{}{
				"site_id":         "1368e337-dd1d-4799-bbb3-8221a1b79bcc",
				"total_fixed":     fixTotal,
				"templates_fixed": fixTotal,
				"rendered_fixed":  fixTotal,
				"needs_rerender":  false,
				"details":         nil,
			},
			"text_color_result": map[string]interface{}{
				"site_id":              "1368e337-dd1d-4799-bbb3-8221a1b79bcc",
				"total_fixed":          textTotal,
				"templates_fixed":      textTotal,
				"rendered_fixed":       textTotal,
				"contracts_added":      0,
				"skipped_low_contrast": 0,
				"needs_rerender":       false,
				"details":              nil,
			},
		},
	}
}

// designSystemPayload is what the OTHER 10 carry — a design-system spec with no
// response envelope at all. Why these rows exist is NOT ESTABLISHED
// (bugs_open/213 §D); the gate must abstain on them, not guess.
func designSystemPayload() map[string]interface{} {
	return map[string]interface{}{
		"spacing":      map[string]interface{}{"section_padding": "96px 0", "container_max_width": "1200px"},
		"typography":   map[string]interface{}{"base_size": "16px", "line_height": "1.6"},
		"color_scheme": map[string]interface{}{"text": "#E6EDF3", "primary": "#0D1117"},
		"design_notes": "Color rationale: the near-black blue-tinted background…",
	}
}

// spawnRecordPayload is what 7 of the 11 recorded abstain-completions carried —
// copied from live row a2ef2613-028c-4fd6-893c-4ccb2d7e733b. It is not the
// handler's reply at all: it is the SPAWN record of the agent that was asked to do
// the work (bugs_closed/287, fixed and live on v1.0.1307). Kept as a fixture even
// though its producing defect is closed, because it is the exact shape this gate
// must refuse rather than complete, and a regression would look like this again.
func spawnRecordPayload() map[string]interface{} {
	return map[string]interface{}{
		"role":       "handler",
		"agent_id":   "b4b21baf-44fe-4889-870a-ed6eadf1ac5f",
		"agent_type": "color-variable-fixer",
		"topics": map[string]interface{}{
			"requests":         "job.9f725e7d-d3a4c260-color-variable-fixer-process_item_iter_1_spawn_handler.requests",
			"responses":        "job.9f725e7d-d869e45e-build-dispatch-loop-spawn_dispatch.responses",
			"parent_responses": "job.9f725e7d-d869e45e-build-dispatch-loop-spawn_dispatch.responses",
		},
	}
}

// foreignTriagePayload is the 11th abstain-completion — live row
// 525275da-19e3-4c49-975c-5a6db5398783. A content-gap planner's decision about
// which sections to add to a page, stored as a dark_section_audit item's result.
// The prose values are truncated here; the KEY SET is the shape under test and is
// copied verbatim.
func foreignTriagePayload() map[string]interface{} {
	return map[string]interface{}{
		"approach":  "add_to_page",
		"reasoning": "The index (landing) page already exists and is the natural home…",
		"add_to_page": map[string]interface{}{
			"page_name":        "index",
			"add_sections":     []interface{}{"hero-tool", "brief-explanation", "tool-list"},
			"content_guidance": "hero-tool: Lead with the standard loan repayment calculator…",
		},
		"new_page":        nil,
		"update_spec":     nil,
		"not_actionable":  nil,
		"retype_existing": nil,
	}
}

// withRosterEntry installs a synthetic roster entry for the duration of one test
// and removes it afterwards.
//
// It exists because the two abstain declarations cannot otherwise be reached: the
// only live entry declares unreadableRefuses, so without this the zero-value
// safety property — an undeclared entry must NOT block — would be asserted by
// nothing, which is precisely the sort of untested defence this estate treats as
// absent. Named types, so a reader can see the case is synthetic and not mistake
// it for a live opt-in.
func withRosterEntry(t *testing.T, itemType string, rule noChangeRule) {
	t.Helper()
	if _, exists := noChangeGates[itemType]; exists {
		t.Fatalf("%s is a LIVE roster entry — pick a synthetic name, or this test is rewriting production policy", itemType)
	}
	noChangeGates[itemType] = rule
	t.Cleanup(func() { delete(noChangeGates, itemType) })
}

func TestHandlerReportedNoChange(t *testing.T) {
	// The synthetic entries exercise the two declarations no live type uses. Same
	// CounterPaths as the real entry so the ONLY variable is OnUnreadable.
	withRosterEntry(t, "test_type_abstains", noChangeRule{
		Why:          "synthetic: fixture for the abstain declaration",
		CounterPaths: []string{"response.fix_result.total_fixed"},
		OnUnreadable: unreadableAbstains,
	})
	withRosterEntry(t, "test_type_undeclared", noChangeRule{
		Why:          "synthetic: fixture for the zero value, which must behave as abstain",
		CounterPaths: []string{"response.fix_result.total_fixed"},
		// OnUnreadable deliberately omitted — this IS the zero value under test.
	})

	cases := []struct {
		name           string
		itemType       string
		result         map[string]interface{}
		want           noChangeOutcome
		detailContains string
	}{
		{
			// THE OPT-IN DEFAULT. An item_type that has not asked for this gate takes
			// a map miss, so a payload that WOULD be blocked for dark_section_audit
			// passes untouched. This is the assertion that the owner's 2026-08-02
			// unsafe-default-OFF ruling is satisfied by construction, and it is first
			// deliberately: if it ever fails, every item type in the fleet is affected.
			name:     "unregistered item_type is inert even with all counters zero",
			itemType: "hardcoded_section_colors",
			result:   fixerEnvelope(float64(0), float64(0)),
			want:     noChangePass,
		},
		{
			name:           "opted-in type, both counters zero as float64 (the JSON round-trip form)",
			itemType:       "dark_section_audit",
			result:         fixerEnvelope(float64(0), float64(0)),
			want:           noChangeBlocked,
			detailContains: "handler reported 0 changes at response.fix_result.total_fixed and response.text_color_result.total_fixed",
		},
		{
			// Same data, the form a Go action's own return value arrives in. A type
			// switch missing this would read "counter absent" for a counter that is
			// present and zero — which since 2026-08-18 is no longer a milder verdict
			// for this type but a DIFFERENT block, so the distinction still matters.
			name:     "opted-in type, both counters zero as int",
			itemType: "dark_section_audit",
			result:   fixerEnvelope(0, 0),
			want:     noChangeBlocked,
		},
		{
			name:     "opted-in type, both counters zero as json.Number",
			itemType: "dark_section_audit",
			result:   fixerEnvelope(json.Number("0"), json.Number("0")),
			want:     noChangeBlocked,
		},
		{
			// The gate's whole job is to stay out of the way when work happened. One
			// non-zero counter anywhere is enough, even beside a zero.
			name:     "opted-in type, one counter non-zero — not this gate's business",
			itemType: "dark_section_audit",
			result:   fixerEnvelope(float64(0), float64(3)),
			want:     noChangePass,
		},
		{
			name:     "opted-in type, both counters non-zero",
			itemType: "dark_section_audit",
			result:   fixerEnvelope(float64(2), float64(5)),
			want:     noChangePass,
		},
		{
			// A partial payload must still be judged on what it DOES carry, or a
			// handler could escape the gate by dropping a field.
			name:     "opted-in type, one counter zero and the other absent → still blocked, absence named",
			itemType: "dark_section_audit",
			result: map[string]interface{}{
				"response": map[string]interface{}{
					"fix_result": map[string]interface{}{"total_fixed": float64(0)},
				},
			},
			want:           noChangeBlocked,
			detailContains: "no value present at response.text_color_result.total_fixed",
		},
		{
			// CHANGED 2026-08-18 (bugs_open/302): this case used to expect abstain and
			// completion. dark_section_audit now declares unreadableRefuses, so the
			// live 10-of-14 payload is REFUSED. The shape text is still asserted —
			// naming what WAS in the payload is the useful half either way.
			name:           "opted-in type declaring refuse, payload is not this handler's → BLOCKED, naming the keys found",
			itemType:       "dark_section_audit",
			result:         designSystemPayload(),
			want:           noChangeUnreadableBlocked,
			detailContains: "color_scheme design_notes spacing typography",
		},
		{
			// The live spawn-record shape — 7 of the 11 abstain-completions, and the
			// one that reversed this gate's own refusal on 5 items.
			name:           "opted-in type declaring refuse, payload is a SPAWN RECORD → BLOCKED",
			itemType:       "dark_section_audit",
			result:         spawnRecordPayload(),
			want:           noChangeUnreadableBlocked,
			detailContains: "agent_id agent_type role topics",
		},
		{
			// The live foreign-decision shape — the 11th.
			name:           "opted-in type declaring refuse, payload is another page's triage decision → BLOCKED",
			itemType:       "dark_section_audit",
			result:         foreignTriagePayload(),
			want:           noChangeUnreadableBlocked,
			detailContains: "add_to_page approach",
		},
		{
			name:     "opted-in type declaring refuse, empty payload → BLOCKED",
			itemType: "dark_section_audit",
			result:   map[string]interface{}{},
			want:     noChangeUnreadableBlocked,
		},
		{
			// A path that runs into a non-map mid-way must not panic or be read as a
			// zero. total_fixed is nested under a string here.
			name:     "opted-in type declaring refuse, counter path blocked by a non-map → BLOCKED",
			itemType: "dark_section_audit",
			result:   map[string]interface{}{"response": "complete"},
			want:     noChangeUnreadableBlocked,
		},
		{
			// A counter present but of an unreadable type must never be silently
			// treated as zero — "0" the string is not a measurement. Note the verdict
			// differs from the zero case above: unreadable, not no-change.
			name:     "opted-in type declaring refuse, counter is a string → BLOCKED as unreadable, not as zero",
			itemType: "dark_section_audit",
			result: map[string]interface{}{
				"response": map[string]interface{}{
					"fix_result":        map[string]interface{}{"total_fixed": "0"},
					"text_color_result": map[string]interface{}{"total_fixed": "0"},
				},
			},
			want: noChangeUnreadableBlocked,
		},
		{
			// The declaration is per-type and must actually be read: same unreadable
			// payload, a type that declares abstain, opposite outcome.
			name:     "opted-in type declaring ABSTAIN, unreadable payload → completes, recorded",
			itemType: "test_type_abstains",
			result:   designSystemPayload(),
			want:     noChangeUnreadableAbstained,
		},
		{
			// THE ZERO-VALUE SAFETY PROPERTY. An entry whose OnUnreadable was never
			// set must not block: a roster line added by somebody who did not read the
			// declaration cannot start refusing completions by accident. The roster
			// test is what stops it shipping; this is what stops it biting if it does.
			name:     "opted-in type with OnUnreadable UNDECLARED, unreadable payload → abstains, never blocks",
			itemType: "test_type_undeclared",
			result:   designSystemPayload(),
			want:     noChangeUnreadableAbstained,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			detail, got := handlerReportedNoChange(tc.itemType, tc.result)

			if got != tc.want {
				t.Fatalf("outcome = %v, want %v (detail=%q)", got, tc.want, detail)
			}
			if tc.detailContains != "" && !strings.Contains(detail, tc.detailContains) {
				t.Fatalf("expected text %q in detail, got %q", tc.detailContains, detail)
			}
			// Each blocking arm must carry the evidence that licenses IT, because this
			// string is what an operator reads off the item's error column — and the two
			// licences are different claims (see noChangeRule.UnreadableWhy).
			switch got {
			case noChangeBlocked:
				if !strings.Contains(detail, noChangeGates[tc.itemType].Why) {
					t.Fatalf("no-change block does not carry the rule's Why: %q", detail)
				}
			case noChangeUnreadableBlocked:
				if !strings.Contains(detail, noChangeGates[tc.itemType].UnreadableWhy) {
					t.Fatalf("unreadable block does not carry the rule's UnreadableWhy: %q", detail)
				}
			}
			// A pass must say nothing: a non-empty detail on a pass would be recorded
			// against an item this gate has no opinion about.
			if got == noChangePass && detail != "" {
				t.Fatalf("pass carried a detail: %q", detail)
			}
		})
	}
}

// NOTE on what this test no longer needs to assert. It used to end with a case
// proving "both noChange and unknownShape set" could not happen — a defence
// against a three-valued return that could express a contradictory state. The
// single noChangeOutcome value makes that state unrepresentable, so the assertion
// was deleted rather than kept as decoration.

// TestNoChangeGatesRosterCarriesItsEvidence guards the roster itself rather than
// the logic. Adding a type here is an assertion about somebody else's handler —
// that a zero-change run cannot be a repair for it — and the estate's rule is that
// such an assertion arrives with the measurement that licenses it.
func TestNoChangeGatesRosterCarriesItsEvidence(t *testing.T) {
	if len(noChangeGates) == 0 {
		t.Fatal("roster is empty — this gate is inert; if that is deliberate, delete the file rather than leaving a mechanism nothing drives")
	}
	for itemType, rule := range noChangeGates {
		if strings.TrimSpace(rule.Why) == "" {
			t.Errorf("%s: no Why — an operator reading a blocked item gets no reason, and no reviewer can check the claim", itemType)
		}
		if len(rule.CounterPaths) == 0 {
			t.Errorf("%s: no CounterPaths — the gate can never fire, so the entry only looks like protection", itemType)
		}
		for _, p := range rule.CounterPaths {
			if !strings.Contains(p, ".") {
				t.Errorf("%s: counter path %q is a bare key; the handler's counts live under its response envelope", itemType, p)
			}
		}
		// THE FORCING FUNCTION (bugs_open/302). Before this, an entry that said
		// nothing about unreadable payloads silently got the completing behaviour —
		// so the author of the NEXT repair type inherited a waiver they never chose
		// and could not see. Leaving the field unset is now unshippable rather than
		// a default, which is the only version of this that survives an author who
		// did not read the file. Synthetic entries in the logic test are installed
		// and removed inside their own test, so they cannot reach this one.
		if rule.OnUnreadable == unreadableUndeclared {
			t.Errorf("%s: OnUnreadable is undeclared — say what an UNREADABLE handler payload means for this "+
				"type (unreadableRefuses, or unreadableAbstains to keep the pre-2026-08-18 behaviour as a "+
				"stated choice); the zero value is deliberately not a policy", itemType)
		}
		// A refusal is new blocking authority on a shared completion path, so it
		// costs evidence — the same bar the Why above is held to, and for the same
		// reason: a reviewer must be able to check the claim, and an operator
		// reading a blocked item must be told what licensed the block.
		if rule.OnUnreadable == unreadableRefuses && strings.TrimSpace(rule.UnreadableWhy) == "" {
			t.Errorf("%s: declares unreadableRefuses with no UnreadableWhy — state the measurement showing "+
				"that an unreadable payload for this type is not a repair", itemType)
		}
	}
}

// TestBlockedCompletionReasonDistinguishesNoChange asserts the fourth blocking
// cause gets its own sentence and reason code. The three existing causes are
// asserted alongside it deliberately: this function's whole history is a message
// that recorded a finding the verifier never made, and the way that regresses is
// one arm being made to answer for another.
func TestBlockedCompletionReasonDistinguishesNoChange(t *testing.T) {
	cases := []struct {
		name       string
		payload    map[string]interface{}
		wantReason string
		wantInMsg  string
	}{
		{
			name:       "no-change",
			payload:    map[string]interface{}{"status": "handler_reported_no_change", "detail": "handler reported 0 changes at X"},
			wantReason: "handler_reported_no_change",
			wantInMsg:  "reported it changed nothing",
		},
		{
			name:       "out of scope",
			payload:    map[string]interface{}{"status": "out_of_scope", "detail": "spec carries no check key"},
			wantReason: "verifier_scope_mismatch",
			wantInMsg:  "does not grade this item",
		},
		{
			name:       "verifier error",
			payload:    map[string]interface{}{"status": "error", "error": "dial tcp: refused"},
			wantReason: "verification_unavailable",
			wantInMsg:  "could not run",
		},
		{
			name:       "defect persists",
			payload:    map[string]interface{}{"status": "defect_persists", "detail": "3 components still match"},
			wantReason: "verification_failed",
			wantInMsg:  "still present",
		},
		{
			// Fifth cause (bugs_open/302). Distinct from "no-change" above: there the
			// handler said plainly that it changed nothing; here nothing was readable
			// at all, so no gate graded anything. The seen map below is what stops
			// this arm being answered by another's sentence.
			name:       "handler result unreadable",
			payload:    map[string]interface{}{"status": "handler_result_unreadable", "detail": "payload top-level keys were [agent_id agent_type role topics]"},
			wantReason: "handler_result_unreadable",
			wantInMsg:  "unreadable to the no-change gate",
		},
	}

	seen := map[string]string{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, reason := blockedCompletionReason(tc.payload)
			if reason != tc.wantReason {
				t.Errorf("reason = %q, want %q", reason, tc.wantReason)
			}
			if !strings.Contains(msg, tc.wantInMsg) {
				t.Errorf("message %q does not contain %q", msg, tc.wantInMsg)
			}
			if prev, dup := seen[reason]; dup {
				t.Errorf("reason code %q already used by %q — the codes are what census queries group by", reason, prev)
			}
			seen[reason] = tc.name
			// The detail must survive into the message: an operator reading the error
			// column gets this string and nothing else.
			if d, _ := tc.payload["detail"].(string); d != "" && !strings.Contains(msg, d) {
				t.Errorf("message dropped the detail %q: %q", d, msg)
			}
		})
	}
}
