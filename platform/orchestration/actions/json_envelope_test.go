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
// failure distribution: the dominant cause is TRUNCATION (the writer's
// generate_content step was capped at max_tokens=2000), which is genuinely
// incomplete JSON and must stay unparseable so the forward path fails loud
// rather than salvaging half an article. Only escaping-only malformations
// (unescaped newlines) are repaired in place.
func TestParseLLMJSON_LiveEnvelopeDistribution(t *testing.T) {
	paths, err := filepath.Glob("testdata/envelopes/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 14 {
		t.Fatalf("expected 14 captured envelopes, found %d", len(paths))
	}

	var repairedCount, unparseableCount int
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}

		// Precondition: every fixture is invalid JSON as sent. If encoding/json
		// ever accepts one, it no longer reproduces the defect.
		var probe interface{}
		if json.Unmarshal(raw, &probe) == nil {
			t.Fatalf("%s parses as valid JSON — no longer reproduces the defect", filepath.Base(p))
		}

		value, repaired, err := ParseLLMJSON(string(raw))
		switch {
		case err != nil:
			unparseableCount++ // truncated / embedded-quote — correctly left for the loud-failure path
		case repaired:
			repairedCount++
			// A repaired envelope must yield the object the model meant.
			obj, ok := value.(map[string]interface{})
			if !ok {
				t.Fatalf("%s: repaired value is %T, want object", filepath.Base(p), value)
			}
			if c, _ := obj["content"].(string); !strings.Contains(c, "<") {
				t.Fatalf("%s: repaired content carries no HTML", filepath.Base(p))
			}
		default:
			t.Fatalf("%s: parsed clean but fixture was invalid JSON", filepath.Base(p))
		}
	}

	// The captured set: 1 escaping-only (repairable), 13 truncated/embedded-quote
	// (correctly unparseable). Locks in that we never silently accept a truncated
	// article as if it were complete.
	if repairedCount != 1 || unparseableCount != 13 {
		t.Fatalf("distribution changed: repaired=%d unparseable=%d (want 1 / 13)", repairedCount, unparseableCount)
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

// TestEscapeControlChars_PreservesStructure checks the repair only ever touches
// bytes inside string literals — escaped quotes and backslashes must survive.
func TestEscapeControlChars_PreservesStructure(t *testing.T) {
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
