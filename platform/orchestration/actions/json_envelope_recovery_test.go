// FILE: platform/orchestration/actions/json_envelope_recovery_test.go
//
// Tests for tier 3 of ParseLLMJSONWithProvenance — recovering a COMPLETE answer
// the model buried in commentary (bugs_open/088).
//
// The negative tests are the load-bearing ones. Two rules look obvious, both are
// wrong, and both were caught by running candidate implementations over 5,844
// stored llm_call_log responses rather than by reasoning:
//
//   - "take the last complete value" — walks into a TRUNCATED array and returns a
//     single element as if it were the whole answer;
//   - "take the last value when there are several" — the model often answers for
//     several sections in one response, so the last value is a different object,
//     not a better copy of the same one.
//
// If a future change makes TestRecoveryNeverSalvagesTruncation or
// TestRecoveryRefusesDifferentShapes fail, the change is reintroducing bug 026.
package actions

import (
	"strings"
	"testing"
)

// The real payload from correlation d9fd6ed2-28e7-4af2-a49a-a749d71bccd3
// (2026-07-26 14:26Z) that killed a model-directory build at iteration 0.
const doubledHeroResponse = `{
  "headline": "Multi-agent systems deployed to production in days, not months — on Kubernetes, Kafka, and Postgres",
  "subheadline": "POCs fail in production because orchestration is harder than the model.",
  "cta_text": "Book a Technical Discovery Call",
  "secondary_cta": "See the Agent Registry"
}

Wait — I must scan for em dashes before returning. Found one in the headline. Rewriting now.

{
  "headline": "Multi-agent systems deployed to production in days, not months. Built on Kubernetes, Kafka, and Postgres.",
  "subheadline": "POCs fail in production because orchestration is harder than the model.",
  "cta_text": "Book a Technical Discovery Call",
  "secondary_cta": "See the Agent Registry"
}`

func TestRecoversSelfCorrectedReemission(t *testing.T) {
	v, prov, err := ParseLLMJSONWithProvenance(doubledHeroResponse)
	if err != nil {
		t.Fatalf("bugs_open/088's own payload must be recovered, got: %v", err)
	}
	if prov != ProvenanceReemitted {
		t.Errorf("provenance: want %q, got %q", ProvenanceReemitted, prov)
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("want an object, got %T", v)
	}
	headline, _ := m["headline"].(string)
	// The model's SECOND word is the answer it meant: the first copy carries the
	// em dash its own instructions forbade. Recovering the first would be a
	// regression dressed as a fix.
	if strings.Contains(headline, "—") {
		t.Errorf("recovered the SUPERSEDED copy (it still contains an em dash): %q", headline)
	}
	if !strings.Contains(headline, "Built on Kubernetes") {
		t.Errorf("want the corrected headline, got %q", headline)
	}
}

func TestRecoversProseAroundASingleValue(t *testing.T) {
	cases := map[string]string{
		"prose before": `Here is the section content you asked for:
{"heading": "About us", "content": "<p>Real copy.</p>"}`,
		"prose after": `{"heading": "About us", "content": "<p>Real copy.</p>"}

Note: I could not find pricing tiers, so I have left that section out. Please supply them.`,
		"prose both sides": `Certainly. Below is the JSON.
{"heading": "About us", "content": "<p>Real copy.</p>"}
Let me know if you would like a different tone.`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			v, prov, err := ParseLLMJSONWithProvenance(in)
			if err != nil {
				t.Fatalf("want recovery, got %v", err)
			}
			if prov != ProvenanceUnwrapped {
				t.Errorf("provenance: want %q, got %q", ProvenanceUnwrapped, prov)
			}
			if m, ok := v.(map[string]interface{}); !ok || m["heading"] != "About us" {
				t.Errorf("wrong value recovered: %#v", v)
			}
		})
	}
}

func TestRecoversFencedBlockWithProseAround(t *testing.T) {
	in := "I have written the section below.\n\n```json\n{\"heading\": \"About us\", \"content\": \"<p>Real copy.</p>\"}\n```\n\nTell me if the tone is wrong."
	v, prov, err := ParseLLMJSONWithProvenance(in)
	if err != nil {
		t.Fatalf("want recovery, got %v", err)
	}
	if prov != ProvenanceFenced {
		t.Errorf("provenance: want %q, got %q", ProvenanceFenced, prov)
	}
	if m, ok := v.(map[string]interface{}); !ok || m["heading"] != "About us" {
		t.Errorf("wrong value recovered: %#v", v)
	}
}

// TestRecoveryNeverSalvagesTruncation is the guard on bug 026 / bugs_closed/005.
// Every case here is INCOMPLETE and must stay rejected, however much valid JSON
// it contains.
func TestRecoveryNeverSalvagesTruncation(t *testing.T) {
	cases := map[string]string{
		// cut mid-string: no complete value at all
		"single object cut mid-string": `{"heading": "About us", "content": "<p>Real copy that stops here`,

		// THE ONE THAT KILLED THE OBVIOUS RULE. A truncated ARRAY whose earlier
		// elements are individually complete. "Take the last complete value" would
		// return the second element — a fragment shipped as the whole answer.
		"truncated array of complete elements": `[
  {"id": "a", "score": 72},
  {"id": "b", "score": 81},
  {"id": "c", "score": 6`,

		// a complete answer followed by a SECOND one that was cut off: returning
		// the first would ship a superseded answer AND hide the truncation.
		"complete object then a truncated second": `{"headline": "First attempt", "sub": "x"}

Actually, let me redo that.

{"headline": "Second attempt", "sub": "y`,

		"prose only":     `I'm sorry, I don't have enough information to write this section.`,
		"fence, but cut": "```json\n{\"heading\": \"About us\", \"content\": \"<p>cut here",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			v, prov, err := ParseLLMJSONWithProvenance(in)
			if err == nil {
				t.Fatalf("TRUNCATION SALVAGED — this reintroduces bug 026.\n  provenance=%q\n  value=%#v", prov, v)
			}
		})
	}
}

// TestRecoveryRefusesDifferentShapes: when a response holds several DIFFERENT
// complete objects the platform must not choose between them. Measured shape from
// the corpus: a writer answering for several sections at once. Picking the last
// would have handed a hero section a testimonials object and reported success.
func TestRecoveryRefusesDifferentShapes(t *testing.T) {
	in := `{"headline": "Design with Authority", "subheadline": "Precision tools"}
{"heading": "About", "content": "<p>Something else entirely.</p>"}
{"headline": "Voices", "subheadline": "What clients say", "testimonials": []}`
	v, prov, err := ParseLLMJSONWithProvenance(in)
	if err == nil {
		t.Fatalf("must refuse to choose between different objects; got provenance=%q value=%#v", prov, v)
	}
}

// TestRecoveryLeavesMarkdownDocumentsAlone: several steps are told to return
// markdown CONTAINING a fenced JSON block (experience-planner's compose: "Output
// the whole plan as markdown … the ```criteria fence … <!-- END EXPERIENCE_PLAN
// -->"). For those the document IS the answer. Recovering the fence would replace
// a whole plan with one of its sub-blocks — 59 of the 93 responses an unguarded
// version recovered were exactly this.
func TestRecoveryLeavesMarkdownDocumentsAlone(t *testing.T) {
	in := "# EXPERIENCE_PLAN — gauntlet\n\n## 1. Intent\nSome narrative.\n\n## 5. Criteria\n\n```criteria\n{\"must_pass\": [\"a\", \"b\"]}\n```\n\n<!-- END EXPERIENCE_PLAN -->"
	if v, prov, err := ParseLLMJSONWithProvenance(in); err == nil {
		t.Fatalf("a markdown plan must stay text, or the plan is replaced by its own criteria block; got provenance=%q value=%#v", prov, v)
	}
}

// The existing tiers must be untouched by the new one.
func TestCleanAndRepairedProvenanceUnchanged(t *testing.T) {
	if v, prov, err := ParseLLMJSONWithProvenance(`{"a": 1}`); err != nil || prov != ProvenanceClean || v == nil {
		t.Errorf("clean parse: prov=%q err=%v", prov, err)
	}
	// raw newline inside a string value — the escaping-only repair, tier 2
	if _, prov, err := ParseLLMJSONWithProvenance("{\"a\": \"line one\nline two\"}"); err != nil || prov != ProvenanceRepaired {
		t.Errorf("repair tier: prov=%q err=%v", prov, err)
	}
	// the wrapper keeps its old contract for every existing caller
	if _, repaired, err := ParseLLMJSON(`{"a": 1}`); err != nil || repaired {
		t.Errorf("ParseLLMJSON wrapper changed behaviour on a clean parse: repaired=%v err=%v", repaired, err)
	}
}
