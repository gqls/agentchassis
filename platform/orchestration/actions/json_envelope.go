// FILE: platform/orchestration/actions/json_envelope.go
//
// Lenient parsing of LLM JSON responses, and the schema-required-field check
// that stops a missing content field from being rendered as a silent blank.
//
// WHY THIS EXISTS
// ---------------
// Models asked for "a JSON object with these keys" sometimes emit string values
// containing LITERAL newlines or unescaped double quotes (from HTML attributes)
// rather than the escaping RFC 8259 requires. That document is invalid JSON, so
// encoding/json rejects it and ExecuteLLMPromptAction used to treat the rejection
// as "the model returned prose", storing {"type":"text","result":"<raw JSON>"}.
//
// That envelope then flowed into page_components.content_data verbatim, so a
// component template asking for {{.content}} found no `content` key and — under
// missingkey=zero — rendered an EMPTY section, silently. Nine live article
// bodies were blanked that way and five more leaked raw JSON onto the page.
// See docs/agent_docs/docs024_key_docs_latest/HANDOFF_2026-07-14_article_body_json_envelope.md.
//
// ParseLLMJSON repairs the escaping-only malformations so a response that was
// only invalid because of raw control characters parses into the object the
// model meant. A TRUNCATED response (the dominant real cause — the writer's
// generate_content step was capped at max_tokens=2000 and cut articles off
// mid-sentence) is genuinely incomplete and still fails to parse; that must NOT
// be silently salvaged — it falls through to the text path, where
// missingRequiredLLMFields makes the renderer refuse to ship a blank section.

package actions

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// ParseLLMJSON parses an LLM response as JSON, repairing the one malformation
// models produce constantly: raw control characters inside string literals.
//
// Returns the parsed value and whether a repair was needed to get it. An error
// means the payload is not recoverable as JSON (genuine prose, or a truncated
// document) and the caller should treat it as plain text.
func ParseLLMJSON(s string) (value interface{}, repaired bool, err error) {
	if err := json.Unmarshal([]byte(s), &value); err == nil {
		return value, false, nil
	}

	// Only attempt a repair on something structurally trying to be a JSON
	// object or array. Prose that merely failed to parse keeps falling through
	// to the text path.
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
// literals — structural whitespace, braces, commas — are copied untouched, so
// this cannot change the shape of a document, only rescue its string values.
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
			b.WriteRune('�')

		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// missingRequiredLLMFields returns the names of input_schema fields that are
// declared source:"llm" AND required:true but are absent (or empty) in content.
//
// The v2 schema shape is input_schema.fields.<name>.{source, required, type}.
// This is the check that turns "a required content field never arrived" from a
// silently-empty render into a loud, actionable failure — the article-body
// component declares exactly one such field, `content`.
func missingRequiredLLMFields(inputSchema map[string]interface{}, content map[string]interface{}) []string {
	fields, ok := inputSchema["fields"].(map[string]interface{})
	if !ok {
		return nil
	}

	var missing []string
	for name, defRaw := range fields {
		def, ok := defRaw.(map[string]interface{})
		if !ok {
			continue
		}
		source, _ := def["source"].(string)
		required, _ := def["required"].(bool)
		if source != "llm" || !required {
			continue
		}
		if isEmptyContentValue(content[name]) {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

// isEmptyContentValue reports whether a content field carries no usable value —
// nil, a blank/whitespace string, or an empty collection.
func isEmptyContentValue(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(t) == ""
	case []interface{}:
		return len(t) == 0
	case map[string]interface{}:
		return len(t) == 0
	}
	return false
}
