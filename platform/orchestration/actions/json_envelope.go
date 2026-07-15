// FILE: platform/orchestration/actions/json_envelope.go
//
// Lenient parsing of LLM JSON responses.
//
// WHY THIS EXISTS
// ---------------
// Models asked for "a JSON object with these keys" routinely emit string values
// that contain LITERAL newlines/tabs rather than the escaped \n / \t that
// RFC 8259 requires inside a string. That document is invalid JSON, so
// encoding/json rejects it. Before this file, ExecuteLLMPromptAction treated
// that rejection as "the model returned prose" and fell back to
// {"type":"text","result":"<the whole raw JSON document as a string>"}.
//
// That envelope then flowed into page_components.content_data verbatim, so a
// component template asking for {{.content}} found no `content` key and — under
// missingkey=zero — rendered an EMPTY section, silently. Nine live article
// bodies were blanked that way and five more leaked the raw JSON onto the page.
// See docs/agent_docs/docs024_key_docs_latest/HANDOFF_2026-07-14_article_body_json_envelope.md.
//
// The repair below rescues exactly that case: it escapes control characters that
// appear INSIDE JSON string literals and leaves every structural byte alone, so
// a document that was only invalid because of unescaped newlines parses as the
// object the model meant to send. Anything that is genuinely not JSON still
// fails, and the text fallback still applies.

package actions

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ParseLLMJSON parses an LLM response as JSON, repairing the one malformation
// models produce constantly: raw control characters inside string literals.
//
// Returns the parsed value and whether a repair was needed to get it. An error
// means the payload is not recoverable as JSON and the caller should treat it as
// plain text.
func ParseLLMJSON(s string) (value interface{}, repaired bool, err error) {
	if err := json.Unmarshal([]byte(s), &value); err == nil {
		return value, false, nil
	}

	// Only attempt a repair on something that is structurally trying to be a
	// JSON object or array. Prose that merely failed to parse must keep falling
	// through to the text path.
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return nil, false, fmt.Errorf("not a JSON object or array")
	}

	fixed := escapeControlCharsInJSONStrings(trimmed)
	if err := json.Unmarshal([]byte(fixed), &value); err != nil {
		return nil, false, fmt.Errorf("unrecoverable after control-char repair: %w", err)
	}
	return value, true, nil
}

// escapeControlCharsInJSONStrings walks a JSON document and escapes control
// characters (< 0x20) that appear inside string literals. Bytes outside string
// literals — the structural whitespace, braces, commas — are copied untouched,
// so this cannot change the shape of a document, only rescue its string values.
//
// Escape state is tracked properly: a backslash inside a string escapes the next
// byte, so a legitimate \" does not end the string and a literal \\ before a
// quote does.
func escapeControlCharsInJSONStrings(s string) string {
	var b strings.Builder
	b.Grow(len(s) + len(s)/16)

	inString := false
	escaped := false

	for _, r := range s {
		switch {
		case !inString:
			b.WriteRune(r)
			if r == '"' {
				inString = true
			}

		case escaped:
			// Previous byte was a backslash — this rune is part of that escape
			// sequence, whatever it is. Copy it and clear the flag.
			b.WriteRune(r)
			escaped = false

		case r == '\\':
			b.WriteRune(r)
			escaped = true

		case r == '"':
			b.WriteRune(r)
			inString = false

		case r < 0x20:
			// The bug we are here to fix: a raw control character inside a
			// string literal. Emit the escape the model should have sent.
			switch r {
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\t':
				b.WriteString(`\t`)
			case '\b':
				b.WriteString(`\b`)
			case '\f':
				b.WriteString(`\f`)
			default:
				fmt.Fprintf(&b, `\u%04x`, r)
			}

		case r == utf8.RuneError:
			// Invalid UTF-8 byte. encoding/json would reject the document;
			// substitute the replacement character so the rest still parses.
			b.WriteRune('�')

		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// UnwrapLLMEnvelope returns the content map an LLM step meant to produce.
//
// It accepts either shape:
//   - the parsed object itself                  → returned as-is
//   - an {"type":..., "result":...} step envelope → the result is unwrapped, and
//     if result is a STRING holding a JSON document (the malformed-JSON case),
//     that document is parsed leniently and its object returned
//
// Returns nil when there is no content map to be had, so callers can distinguish
// "no content" from "content that happens to be the envelope" — persisting the
// envelope as content_data is precisely the defect this guards against.
func UnwrapLLMEnvelope(v interface{}) map[string]interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		// A step envelope is exactly {type, result} — a content map never is.
		if !isLLMStepEnvelope(t) {
			return t
		}
		return UnwrapLLMEnvelope(t["result"])

	case string:
		parsed, _, err := ParseLLMJSON(t)
		if err != nil {
			return nil
		}
		if m, ok := parsed.(map[string]interface{}); ok {
			return m
		}
		return nil
	}
	return nil
}

// isLLMStepEnvelope reports whether m is the {"type","result"} wrapper that
// ExecuteLLMPromptAction returns, rather than a map of content fields. The
// wrapper carries a `result` key and nothing beyond `type` alongside it.
func isLLMStepEnvelope(m map[string]interface{}) bool {
	if _, hasResult := m["result"]; !hasResult {
		return false
	}
	for k := range m {
		if k != "result" && k != "type" {
			return false
		}
	}
	return true
}
