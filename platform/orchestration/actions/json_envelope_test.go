package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseLLMJSON_RepairsLiveEnvelopes runs the repair against the 14 real
// LLM responses that blanked or leaked live article bodies (captured from
// page_components.content_data->>'result' on 2026-07-14). Every one of them is
// invalid JSON as sent — literal newlines inside the "content" string — and
// every one must come back as an object carrying usable article HTML.
func TestParseLLMJSON_RepairsLiveEnvelopes(t *testing.T) {
	paths, err := filepath.Glob("testdata/envelopes/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 14 {
		t.Fatalf("expected 14 captured envelopes, found %d", len(paths))
	}

	for _, p := range paths {
		t.Run(filepath.Base(p), func(t *testing.T) {
			raw, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}

			// Precondition: this is the malformed shape. If encoding/json ever
			// accepts it, the fixture no longer represents the defect.
			var probe interface{}
			if json.Unmarshal(raw, &probe) == nil {
				t.Fatal("fixture parses as valid JSON — it no longer reproduces the defect")
			}

			value, repaired, err := ParseLLMJSON(string(raw))
			if err != nil {
				t.Fatalf("repair failed: %v", err)
			}
			if !repaired {
				t.Fatal("expected repaired=true for a malformed envelope")
			}

			obj, ok := value.(map[string]interface{})
			if !ok {
				t.Fatalf("expected a JSON object, got %T", value)
			}

			content, ok := obj["content"].(string)
			if !ok {
				t.Fatalf("no string `content` key; keys = %v", keysOf(obj))
			}
			if !strings.Contains(content, "<p>") && !strings.Contains(content, "<h2>") {
				t.Fatalf("recovered content carries no article HTML (len=%d)", len(content))
			}
			t.Logf("recovered %d bytes of article HTML", len(content))
		})
	}
}

// TestParseLLMJSON_LeavesValidJSONAlone guards the healthy path: a well-formed
// response must parse without being flagged as repaired.
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

// TestParseLLMJSON_RejectsProse ensures genuine prose still fails, so the
// text fallback in ExecuteLLMPromptAction keeps working.
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

// TestUnwrapLLMEnvelope covers the shapes RenderComponentAction must handle:
// a plain content map, a {type,result} envelope holding a parsed object, and
// the defect shape — an envelope whose result is a malformed JSON string.
func TestUnwrapLLMEnvelope(t *testing.T) {
	t.Run("plain content map passes through", func(t *testing.T) {
		in := map[string]interface{}{"content": "<p>hi</p>", "headline": "x"}
		got := UnwrapLLMEnvelope(in)
		if got["content"] != "<p>hi</p>" || got["headline"] != "x" {
			t.Fatalf("content map altered: %v", got)
		}
	})

	t.Run("envelope with parsed result is unwrapped", func(t *testing.T) {
		in := map[string]interface{}{
			"type":   "json",
			"result": map[string]interface{}{"content": "<p>hi</p>"},
		}
		if got := UnwrapLLMEnvelope(in); got["content"] != "<p>hi</p>" {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("envelope with malformed JSON string is repaired and unwrapped", func(t *testing.T) {
		in := map[string]interface{}{
			"type":   "text",
			"result": "{\n  \"content\": \"<h2>Title</h2>\nreal newline here\"\n}",
		}
		got := UnwrapLLMEnvelope(in)
		if got == nil {
			t.Fatal("envelope not unwrapped — this is the live defect")
		}
		if _, isEnvelope := got["result"]; isEnvelope {
			t.Fatal("returned the envelope itself as content_data — the defect")
		}
		if !strings.Contains(got["content"].(string), "<h2>Title</h2>") {
			t.Fatalf("content not recovered: %v", got)
		}
	})

	t.Run("envelope wrapping prose yields no content map", func(t *testing.T) {
		in := map[string]interface{}{"type": "text", "result": "just prose"}
		if got := UnwrapLLMEnvelope(in); got != nil {
			t.Fatalf("prose must not become content_data, got %v", got)
		}
	})

	t.Run("content map that legitimately has a result field is not mistaken for an envelope", func(t *testing.T) {
		in := map[string]interface{}{"result": "the outcome", "headline": "h"}
		got := UnwrapLLMEnvelope(in)
		if got["headline"] != "h" {
			t.Fatalf("real content map was unwrapped as an envelope: %v", got)
		}
	})
}

func keysOf(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
