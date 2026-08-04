package actions

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tests for the content_data transport-envelope guard (bugs_open/190).
//
// The mutation that must break each test is named ON the test. A test whose
// named mutation still passes is not testing what its name says — and this
// guard is unusually exposed to that, because its happy path is "change
// nothing", which every broken version of it also does.
//
// The fixtures are the two LIVE rows wherever possible, not invented shapes:
// finetuning.uk's {content,result,type} superset and gaswholesalers.com's
// prose_around payload are the cases that actually exist, and each one refutes
// a plausible implementation.

// --- The NO-OP negative control -------------------------------------------
//
// The guard's dominant case is legitimate content, and the damage a bad
// predicate does is invisible: it rewrites a map nobody was looking at. So this
// asserts BYTE identity through json.Marshal (Go sorts object keys, so the
// encoding is deterministic), not merely "no error".
//
// MUTATION THAT MUST BREAK IT: key the predicate on the presence of `type`
// alone, or on `result` alone — either turns the first two fixtures into
// envelopes and rewrites them.
func TestLegitimateContentDataPassesByteIdentical(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]interface{}
	}{
		{
			// Has `type`, but no STRING `result`. A section legitimately
			// describing itself as a text block.
			name: "type text but no result",
			in: map[string]interface{}{
				"type":     "text",
				"headline": "Wholesale Fuel Supply",
				"body":     "<p>Real content.</p>",
			},
		},
		{
			// Has a string `result`, but is not type text. This is the shape a
			// calculator/tool component stores for a displayed value.
			name: "string result but not type text",
			in: map[string]interface{}{
				"result": "42",
				"label":  "Answer",
				"type":   "calculator",
			},
		},
		{
			// The JSON path's envelope: result is an OBJECT, not a string. It
			// never reaches storage intact, and the guard must not claim it.
			name: "json envelope with object result",
			in: map[string]interface{}{
				"type":   "json",
				"result": map[string]interface{}{"headline": "x"},
			},
		},
		{
			name: "ordinary article map",
			in: map[string]interface{}{
				"content":  "<h2>What Is an AI Data Risk Checker?</h2><p>Long article body.</p>",
				"heading":  "AI Data Risk",
				"sections": []interface{}{"a", "b"},
			},
		},
		{
			name: "empty map",
			in:   map[string]interface{}{},
		},
		{
			name: "nil map",
			in:   nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("marshalling fixture: %v", err)
			}

			out, changed, err := normalizeContentDataEnvelope(tc.in)
			if err != nil {
				t.Fatalf("legitimate content_data was refused: %v", err)
			}
			if changed {
				t.Fatalf("legitimate content_data was reported as changed")
			}

			after, err := json.Marshal(out)
			if err != nil {
				t.Fatalf("marshalling result: %v", err)
			}
			if string(before) != string(after) {
				t.Errorf("legitimate content_data was rewritten:\n before: %s\n  after: %s", before, after)
			}
		})
	}
}

// --- Decode, but ONLY on the lossless tiers --------------------------------

// MUTATION THAT MUST BREAK IT: restrict decoding to ProvenanceClean only (the
// `repaired` sub-case then fails), or drop the decode entirely.
func TestEnvelopeDecodedWhenLossless(t *testing.T) {
	t.Run("clean", func(t *testing.T) {
		in := map[string]interface{}{
			"type":   "text",
			"result": `{"headline":"Real headline","body":"Real body"}`,
		}
		out, changed, err := normalizeContentDataEnvelope(in)
		if err != nil {
			t.Fatalf("a cleanly-parseable envelope should decode, got: %v", err)
		}
		if !changed {
			t.Fatal("decode was not reported as a change")
		}
		if _, stillThere := out["type"]; stillThere {
			t.Error("envelope key `type` survived the decode")
		}
		if _, stillThere := out["result"]; stillThere {
			t.Error("envelope key `result` survived the decode")
		}
		if out["headline"] != "Real headline" || out["body"] != "Real body" {
			t.Errorf("decoded payload not stored: %#v", out)
		}
	})

	t.Run("repaired: raw newline inside a string value", func(t *testing.T) {
		// The escaping-only malformation ParseLLMJSON exists to repair. It
		// discards no bytes, so it is on the storable side of the line.
		in := map[string]interface{}{
			"type":   "text",
			"result": "{\"headline\":\"Line one\nLine two\"}",
		}
		out, changed, err := normalizeContentDataEnvelope(in)
		if err != nil {
			t.Fatalf("an escaping-repairable envelope should decode, got: %v", err)
		}
		if !changed {
			t.Fatal("decode was not reported as a change")
		}
		if h, _ := out["headline"].(string); !strings.Contains(h, "Line one") {
			t.Errorf("repaired payload not stored: %#v", out)
		}
	})
}

// The gaswholesalers.com case, and the most important test in this file.
//
// Its payload DOES recover mechanically — via prose_around, to 7 tier_* keys
// totalling 131 characters — while the real section copy sits in the markdown
// tail the recovery correctly refuses to guess at. Storing that "recovery"
// would silently replace a live page's content with 131 characters.
//
// So this test is the codified promise that no automated repair ever fires at
// that row: the guard must REFUSE, not decode.
//
// MUTATION THAT MUST BREAK IT: accept any successful parse regardless of
// provenance — the single most natural way to write this guard, and it would
// destroy the page.
func TestLossyProvenanceIsRefusedNotDecoded(t *testing.T) {
	in := map[string]interface{}{
		"type": "text",
		"result": `{"tier_1":"a","tier_2":"b"}

## Wholesale Fuel Supply Programmes

The real section copy lives out here in the markdown tail, and it is much
longer than the fragment above.`,
	}

	// Precondition: the payload really is mechanically recoverable, or this
	// test would pass for the wrong reason (a refusal caused by a parse
	// failure, not by the provenance policy).
	if _, provenance, err := ParseLLMJSONWithProvenance(in["result"].(string)); err != nil {
		t.Fatalf("fixture is not recoverable at all, so it cannot test the provenance rule: %v", err)
	} else if provenance == ProvenanceClean || provenance == ProvenanceRepaired {
		t.Fatalf("fixture recovered via %q, which is a LOSSLESS tier — it cannot test the lossy rule", provenance)
	}

	if _, _, err := normalizeContentDataEnvelope(in); err == nil {
		t.Fatal("a lossy recovery was accepted — this would silently replace live page content")
	}
}

// MUTATION THAT MUST BREAK IT: none needed beyond the parser — but the guard
// must not mistake genuine prose for something storable.
func TestUnparseableEnvelopeIsRefused(t *testing.T) {
	in := map[string]interface{}{
		"type":   "text",
		"result": "I'm sorry, I can't produce that section without more detail about the audience.",
	}
	if _, _, err := normalizeContentDataEnvelope(in); err == nil {
		t.Fatal("an envelope holding plain prose was accepted")
	}
}

// This guard drops every __-prefixed transport marker on a lossless decode,
// including __truncated — which makes it look like a truncation-marker consumer
// to truncation_guard_test.go's source scan. It is exempted there rather than
// registered, and THIS is the check that the exemption's reasoning is sound: a
// truncated payload can never reach the decode branch, so dropping the marker
// cannot lose a warning that would otherwise have been acted on.
//
// Both outcomes must be a refusal. A truncated JSON string never reaches its
// closing quote, so it either fails to parse outright, or tier 3 recovers
// something via a LOSSY provenance — and this guard refuses on lossy.
//
// MUTATION THAT MUST BREAK IT: accept any successful parse regardless of
// provenance (the same mutation TestLossyProvenanceIsRefusedNotDecoded catches)
// — after which a truncated article would be stored as its own fragment.
func TestTruncatedEnvelopeCannotReachTheDecodeBranch(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]interface{}
	}{
		{
			name: "cut mid string value",
			in: map[string]interface{}{
				"type":   "text",
				"result": `{"content":"This article begins normally and then is cut off mid-sen`,
			},
		},
		{
			name: "cut mid string, producer flagged it",
			in: map[string]interface{}{
				"type":        "text",
				"result":      `{"content":"Another article cut at max_tokens`,
				"__truncated": true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, changed, err := normalizeContentDataEnvelope(tc.in)
			if err == nil {
				t.Fatalf("a truncated envelope was ACCEPTED (changed=%v, out=%#v) — the exemption in "+
					"truncationMarkerExemptions claims this is impossible", changed, out)
			}
		})
	}
}

// --- The superset case: finetuning.uk's live shape -------------------------

// The live row is {content, result, type} — THREE keys. bugs_open/190's own
// verification recipe says to key on "the exact two-key shape", which is
// silent on exactly this row.
//
// MUTATION THAT MUST BREAK IT: predicate on len(m) == 2, or on an exact
// {type,result} key set.
func TestSupersetEnvelopeIsDetected(t *testing.T) {
	in := map[string]interface{}{
		"type":    "text",
		"result":  `{"heading":"AI Data Risk"}`,
		"content": "<h2>What Is an AI Data Risk Checker?</h2>",
	}
	if !isLLMTransportEnvelope(in) {
		t.Fatal("a three-key envelope was not detected — this is the live finetuning.uk shape")
	}

	out, changed, err := normalizeContentDataEnvelope(in)
	if err != nil {
		t.Fatalf("superset envelope should decode: %v", err)
	}
	if !changed {
		t.Fatal("decode was not reported as a change")
	}
	if out["content"] != "<h2>What Is an AI Data Risk Checker?</h2>" {
		t.Errorf("the real sibling key was lost: %#v", out)
	}
	if out["heading"] != "AI Data Risk" {
		t.Errorf("the decoded payload was lost: %#v", out)
	}
	if _, stillThere := out["result"]; stillThere {
		t.Error("envelope key `result` survived")
	}
}

// MUTATION THAT MUST BREAK IT: merge blindly (last write wins) instead of
// comparing. Two candidate values for one field is a choice, not a merge.
func TestSupersetWithConflictingValuesIsRefused(t *testing.T) {
	in := map[string]interface{}{
		"type":    "text",
		"result":  `{"content":"the decoded version"}`,
		"content": "the stored sibling version",
	}
	if _, _, err := normalizeContentDataEnvelope(in); err == nil {
		t.Fatal("a superset whose two candidate values DISAGREE was silently merged")
	}
}

// Agreement is not a conflict — the guard must not refuse a row just because a
// key appears on both sides with the same value.
//
// MUTATION THAT MUST BREAK IT: refuse on any overlapping key.
func TestSupersetWithAgreeingValuesIsMerged(t *testing.T) {
	in := map[string]interface{}{
		"type":    "text",
		"result":  `{"content":"same on both sides","extra":"new"}`,
		"content": "same on both sides",
	}
	out, changed, err := normalizeContentDataEnvelope(in)
	if err != nil {
		t.Fatalf("agreeing values should merge, not refuse: %v", err)
	}
	if !changed || out["content"] != "same on both sides" || out["extra"] != "new" {
		t.Errorf("merge result wrong: %#v", out)
	}
}

// The platform markers the modern text path adds describe the TRANSPORT, not
// the content, so a decode drops them with type and result.
//
// MUTATION THAT MUST BREAK IT: preserve all non-envelope siblings
// indiscriminately.
func TestPlatformMarkersAreDroppedOnDecode(t *testing.T) {
	in := map[string]interface{}{
		"type":                  "text",
		"result":                `{"headline":"x"}`,
		"__json_contract_unmet": true,
		"__truncated":           false,
	}
	out, _, err := normalizeContentDataEnvelope(in)
	if err != nil {
		t.Fatalf("unexpected refusal: %v", err)
	}
	for k := range out {
		if strings.HasPrefix(k, "__") {
			t.Errorf("platform marker %q survived the decode", k)
		}
	}
}

// --- Double-wrap defence (bugs_closed/054 precedent) -----------------------

// MUTATION THAT MUST BREAK IT: unwrap exactly once and return.
func TestDoubleWrappedEnvelopeIsUnwrapped(t *testing.T) {
	inner := `{"type":"text","result":"{\"headline\":\"deep\"}"}`
	in := map[string]interface{}{"type": "text", "result": inner}

	out, changed, err := normalizeContentDataEnvelope(in)
	if err != nil {
		t.Fatalf("a double-wrapped envelope should unwrap: %v", err)
	}
	if !changed {
		t.Fatal("decode was not reported as a change")
	}
	if out["headline"] != "deep" {
		t.Errorf("inner payload not reached: %#v", out)
	}
	if isLLMTransportEnvelope(out) {
		t.Error("result is STILL an envelope after normalisation")
	}
}

// MUTATION THAT MUST BREAK IT: remove the depth cap — this fixture then
// recurses to the bottom instead of refusing.
func TestPathologicallyNestedEnvelopeIsRefused(t *testing.T) {
	payload := `{"headline":"bottom"}`
	for i := 0; i < maxEnvelopeUnwrapDepth+2; i++ {
		wrapped, err := json.Marshal(map[string]interface{}{"type": "text", "result": payload})
		if err != nil {
			t.Fatalf("building fixture: %v", err)
		}
		payload = string(wrapped)
	}
	var in map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &in); err != nil {
		t.Fatalf("building fixture: %v", err)
	}

	if _, _, err := normalizeContentDataEnvelope(in); err == nil {
		t.Fatal("an envelope nested past the depth cap was accepted")
	}
}

// --- The seam function -----------------------------------------------------

// sanitizeSectionsContentData must mutate the section set IN PLACE, because the
// caller persists that same slice. A version that normalised a copy would pass
// every test above and still write the envelope to the database.
//
// MUTATION THAT MUST BREAK IT: range over `sections` by value
// (`for _, s := range sections`) and assign to s.ContentData.
func TestSeamMutatesSectionsInPlace(t *testing.T) {
	sections := []SectionData{
		{ComponentName: "hero", ContentData: map[string]interface{}{"headline": "untouched"}},
		{ComponentName: "article-body", ContentData: map[string]interface{}{
			"type":   "text",
			"result": `{"content":"decoded body"}`,
		}},
	}

	params := ActionParams{Logger: zap.NewNop()} // DB nil: the log writer no-ops
	if err := sanitizeSectionsContentData(context.Background(), params, uuid.Nil, "how-pricing-works", sections); err != nil {
		t.Fatalf("unexpected refusal: %v", err)
	}

	if sections[0].ContentData["headline"] != "untouched" {
		t.Errorf("a legitimate section was modified: %#v", sections[0].ContentData)
	}
	if got := sections[1].ContentData; got["content"] != "decoded body" {
		t.Errorf("the envelope section was not normalised IN PLACE: %#v", got)
	}
	if _, stillThere := sections[1].ContentData["result"]; stillThere {
		t.Error("envelope key `result` survived in the persisted slice")
	}
}

// A refusal must name the section, so the pod log points at the row to fix
// rather than at the page.
//
// MUTATION THAT MUST BREAK IT: return the bare error from
// normalizeContentDataEnvelope without wrapping it.
func TestSeamRefusalNamesTheSection(t *testing.T) {
	sections := []SectionData{
		{ComponentName: "pricing-tiers", ContentData: map[string]interface{}{
			"type":   "text",
			"result": "not JSON at all, just prose from the model",
		}},
	}

	params := ActionParams{Logger: zap.NewNop()}
	err := sanitizeSectionsContentData(context.Background(), params, uuid.Nil, "how-pricing-works", sections)
	if err == nil {
		t.Fatal("the seam accepted an unrecoverable envelope")
	}
	if !strings.Contains(err.Error(), "pricing-tiers") {
		t.Errorf("refusal does not name the section: %v", err)
	}
	if !strings.Contains(err.Error(), "bugs_open/190") {
		t.Errorf("refusal does not cite the bug, so the next reader cannot find this file: %v", err)
	}
}

// The SQL twin must stay in step with the Go predicate. It cannot be executed
// here, so this asserts the two things a drift would break: that it names the
// same two conditions the Go tests, and that it tests result's TYPE rather than
// its presence.
//
// MUTATION THAT MUST BREAK IT: relax the SQL to `content_data ? 'result'`,
// which is the bug file's original query and matches the object-result form the
// Go predicate deliberately excludes.
func TestSQLPredicateMirrorsTheGoPredicate(t *testing.T) {
	if !strings.Contains(contentDataEnvelopeSQLPredicate, `'type' = 'text'`) {
		t.Error("SQL twin no longer tests type = text")
	}
	if !strings.Contains(contentDataEnvelopeSQLPredicate, `jsonb_typeof(content_data->'result') = 'string'`) {
		t.Error("SQL twin no longer tests that result is a STRING — it would match the json-path envelope too")
	}
}
