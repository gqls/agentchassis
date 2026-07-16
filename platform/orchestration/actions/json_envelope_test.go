package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseLLMJSON_LiveEnvelopeDistribution runs the parser against the 14 real
// LLM responses that blanked or leaked live article bodies (captured from
// page_components.content_data->>'result' on 2026-07-14) and asserts the actual
// failure distribution.
//
// Two axes matter, and they are independent: whether a fixture is JSON-complete
// (its raw bytes end with a real closing `"}`) and whether ParseLLMJSON can parse
// it. 12 of the 14 are genuinely TRUNCATED — cut off mid-word by the writer's
// generate_content max_tokens=2000 ceiling — and must stay unparseable forever;
// no repair can complete a sentence the model never finished. The other 2 ARE
// complete: one only had raw newlines (repaired since 2026-07-15), and one
// (finetuning.uk tool-ai-data-risk-checker-guide, recovered manually 2026-07-16
// — see aaa_fails_to_mend/005 §Follow-up) is complete but ends its article in a
// contact link (`<a href="mailto:...">`), so its only fault is the unescaped
// attribute quotes that repairJSONStringLiterals now handles.
func TestParseLLMJSON_LiveEnvelopeDistribution(t *testing.T) {
	paths, err := filepath.Glob("testdata/envelopes/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 14 {
		t.Fatalf("expected 14 captured envelopes, found %d", len(paths))
	}

	// The two fixtures whose raw bytes are genuinely complete (end in a real
	// closing `"}`) — everything else is truncated regardless of which parse
	// error encoding/json happens to report.
	wantRepairable := map[string]bool{
		"f8412ad3-7a62-49ba-ab73-82304bb45e74.json": true, // newline-only
		"850e356d-936b-426d-9a54-8b89efbf91ec.json": true, // embedded quotes, contact link
	}

	var repairedCount, unparseableCount int
	for _, p := range paths {
		name := filepath.Base(p)
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}

		// Precondition: every fixture is invalid JSON as sent. If encoding/json
		// ever accepts one, it no longer reproduces the defect.
		var probe interface{}
		if json.Unmarshal(raw, &probe) == nil {
			t.Fatalf("%s parses as valid JSON — no longer reproduces the defect", name)
		}

		value, repaired, err := ParseLLMJSON(string(raw))
		switch {
		case err != nil:
			unparseableCount++
			if wantRepairable[name] {
				t.Errorf("%s: expected repairable (complete document), got unparseable: %v", name, err)
			}
		case repaired:
			repairedCount++
			if !wantRepairable[name] {
				t.Errorf("%s: repaired but this fixture is genuinely truncated — must not silently accept a partial article", name)
			}
			// A repaired envelope must yield the object the model meant.
			obj, ok := value.(map[string]interface{})
			if !ok {
				t.Fatalf("%s: repaired value is %T, want object", name, value)
			}
			if c, _ := obj["content"].(string); !strings.Contains(c, "<") {
				t.Fatalf("%s: repaired content carries no HTML", name)
			}
		default:
			t.Fatalf("%s: parsed clean but fixture was invalid JSON", name)
		}
	}

	// Locks in the real distribution: 2 repairable (complete documents), 12
	// truncated (correctly, permanently unparseable).
	if repairedCount != 2 || unparseableCount != 12 {
		t.Fatalf("distribution changed: repaired=%d unparseable=%d (want 2 / 12)", repairedCount, unparseableCount)
	}
}

// TestParseLLMJSON_LeavesValidJSONAlone guards the healthy path: a well-formed
// response parses without being flagged as repaired.
func TestParseLLMJSON_LeavesValidJSONAlone(t *testing.T) {
	v, repaired, err := ParseLLMJSON(`{"content":"<p>Already escaped\nfine</p>"}`)
	if err != nil {
		t.Fatal(err)
	}
	if repaired {
		t.Fatal("valid JSON should not be reported as repaired")
	}
	if got := v.(map[string]interface{})["content"]; got != "<p>Already escaped\nfine</p>" {
		t.Fatalf("value altered by parse: %q", got)
	}
}

// TestParseLLMJSON_RepairsUnescapedNewlines proves the escaping-only repair: a
// document whose only fault is raw newlines inside a string parses after repair.
func TestParseLLMJSON_RepairsUnescapedNewlines(t *testing.T) {
	in := "{\n  \"content\": \"<h2>Title</h2>\nreal newline in the value\"\n}"
	v, repaired, err := ParseLLMJSON(in)
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if !repaired {
		t.Fatal("expected repaired=true")
	}
	if c := v.(map[string]interface{})["content"].(string); c != "<h2>Title</h2>\nreal newline in the value" {
		t.Fatalf("content = %q", c)
	}
}

// TestParseLLMJSON_RejectsProse ensures genuine prose still fails, so the text
// fallback in ExecuteLLMPromptAction keeps working.
func TestParseLLMJSON_RejectsProse(t *testing.T) {
	for _, s := range []string{
		"Here is the content you asked for.",
		"",
		"<h2>Just some HTML</h2>",
	} {
		if _, _, err := ParseLLMJSON(s); err == nil {
			t.Fatalf("expected prose %q to fail JSON parsing", s)
		}
	}
}

// TestRepairJSONStringLiterals_PreservesStructure checks the repair only ever
// touches bytes inside string literals — escaped quotes and backslashes must
// survive, alongside a real control-character repair in the same document.
func TestRepairJSONStringLiterals_PreservesStructure(t *testing.T) {
	in := "{\n  \"a\": \"line1\nline2\",\n  \"b\": \"quote \\\" and backslash \\\\\"\n}"
	v, repaired, err := ParseLLMJSON(in)
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if !repaired {
		t.Fatal("expected repaired=true")
	}
	m := v.(map[string]interface{})
	if m["a"] != "line1\nline2" {
		t.Fatalf("a = %q, want line1\\nline2", m["a"])
	}
	if m["b"] != `quote " and backslash \` {
		t.Fatalf("b = %q", m["b"])
	}
}

// TestParseLLMJSON_RepairsEmbeddedQuotes is the fix for the follow-up gap found
// 2026-07-16: an article whose HTML ends in a contact link
// (<a href="mailto:...">) puts unescaped attribute quotes straight into the JSON
// string value. The document is otherwise complete — this must parse, not just
// escape control characters.
func TestParseLLMJSON_RepairsEmbeddedQuotes(t *testing.T) {
	in := `{"content":"<p>Reach us at <a href="mailto:hello@example.com">hello@example.com</a> or <a href="tel:+441234567890">call</a>.</p>"}`
	v, repaired, err := ParseLLMJSON(in)
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if !repaired {
		t.Fatal("expected repaired=true")
	}
	c := v.(map[string]interface{})["content"].(string)
	want := `<p>Reach us at <a href="mailto:hello@example.com">hello@example.com</a> or <a href="tel:+441234567890">call</a>.</p>`
	if c != want {
		t.Fatalf("content = %q,\n want %q", c, want)
	}
}

// TestParseLLMJSON_RepairsEmbeddedQuotesInNonLastField proves the repair is
// general — not scoped to a single named field, or to the field being last in
// the object. A schema with multiple LLM fields (headline + body, say) must
// recover correctly even when the embedded-quote field is NOT the final key.
func TestParseLLMJSON_RepairsEmbeddedQuotesInNonLastField(t *testing.T) {
	in := `{"body":"<p>Call <a href="tel:+441234567890">us</a> today.</p>","headline":"Get in touch","cta_text":"Contact"}`
	v, repaired, err := ParseLLMJSON(in)
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if !repaired {
		t.Fatal("expected repaired=true")
	}
	m := v.(map[string]interface{})
	if m["body"] != `<p>Call <a href="tel:+441234567890">us</a> today.</p>` {
		t.Fatalf("body = %q", m["body"])
	}
	if m["headline"] != "Get in touch" {
		t.Fatalf("headline = %q — a field AFTER the embedded-quote field was corrupted", m["headline"])
	}
	if m["cta_text"] != "Contact" {
		t.Fatalf("cta_text = %q — a field AFTER the embedded-quote field was corrupted", m["cta_text"])
	}
}

// TestParseLLMJSON_TruncatedWithEmbeddedQuoteStillFails guards the fail-loud
// property: a document that is BOTH truncated AND has an embedded quote earlier
// on must still be rejected — the embedded-quote repair must never paper over a
// genuinely incomplete article.
func TestParseLLMJSON_TruncatedWithEmbeddedQuoteStillFails(t *testing.T) {
	in := `{"content":"<p>Call <a href="tel:+441234567890">us</a> today. Also this article continues on and`
	if _, _, err := ParseLLMJSON(in); err == nil {
		t.Fatal("a truncated document must stay unparseable even with an embedded quote earlier on")
	}
}

// TestMissingRequiredLLMFields covers the schema-required-field check that makes
// the renderer refuse to ship a blank section.
func TestMissingRequiredLLMFields(t *testing.T) {
	// The real article-body schema: one required, source:llm field.
	articleBody := map[string]interface{}{
		"fields": map[string]interface{}{
			"content": map[string]interface{}{
				"type": "text", "source": "llm", "required": true,
			},
		},
	}

	t.Run("present content passes", func(t *testing.T) {
		if m := missingRequiredLLMFields(articleBody, map[string]interface{}{"content": "<p>hi</p>"}); len(m) != 0 {
			t.Fatalf("unexpected missing: %v", m)
		}
	})

	t.Run("absent content is flagged — the blanking case", func(t *testing.T) {
		// This is exactly the broken envelope: {type,result} present, no content.
		cd := map[string]interface{}{"type": "text", "result": "{...raw...}"}
		if m := missingRequiredLLMFields(articleBody, cd); len(m) != 1 || m[0] != "content" {
			t.Fatalf("want [content], got %v", m)
		}
	})

	t.Run("empty-string content is flagged", func(t *testing.T) {
		if m := missingRequiredLLMFields(articleBody, map[string]interface{}{"content": "   "}); len(m) != 1 {
			t.Fatalf("blank string should be missing, got %v", m)
		}
	})

	t.Run("non-llm required field is ignored", func(t *testing.T) {
		schema := map[string]interface{}{"fields": map[string]interface{}{
			"items": map[string]interface{}{"source": "query.products", "required": true},
		}}
		if m := missingRequiredLLMFields(schema, map[string]interface{}{}); len(m) != 0 {
			t.Fatalf("query-sourced field must not be checked, got %v", m)
		}
	})

	t.Run("optional llm field absent is fine", func(t *testing.T) {
		schema := map[string]interface{}{"fields": map[string]interface{}{
			"subheadline": map[string]interface{}{"source": "llm", "required": false},
		}}
		if m := missingRequiredLLMFields(schema, map[string]interface{}{}); len(m) != 0 {
			t.Fatalf("optional field must not be flagged, got %v", m)
		}
	})

	t.Run("no v2 schema is a no-op", func(t *testing.T) {
		if m := missingRequiredLLMFields(map[string]interface{}{}, map[string]interface{}{}); m != nil {
			t.Fatalf("empty schema should return nil, got %v", m)
		}
	})
}
