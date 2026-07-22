// FILE: platform/orchestration/actions/json_envelope.go
//
// Lenient parsing of LLM JSON responses, and the schema-required-field check
// that stops a missing content field from being rendered as a silent blank.
//
// WHY THIS EXISTS
// ---------------
// Models asked for "a JSON object with these keys" sometimes emit string values
// containing LITERAL newlines or unescaped double quotes (from HTML attributes,
// e.g. an article body ending in <a href="mailto:...">) rather than the escaping
// RFC 8259 requires. That document is invalid JSON, so encoding/json rejects it
// and ExecuteLLMPromptAction used to treat the rejection as "the model returned
// prose", storing {"type":"text","result":"<raw JSON>"}.
//
// That envelope then flowed into page_components.content_data verbatim, so a
// component template asking for {{.content}} found no `content` key and — under
// missingkey=zero — rendered an EMPTY section, silently. Nine live article
// bodies were blanked that way and five more leaked raw JSON onto the page.
// See docs/agent_docs/docs024_key_docs_latest/HANDOFF_2026-07-14_article_body_json_envelope.md.
//
// ParseLLMJSON repairs both escaping-only malformations: raw control characters,
// and raw (unescaped) double quotes inside string values — the second is what an
// HTML attribute like href="..." produces, and control-char escaping alone can't
// tell an embedded quote from the string's real terminator. See
// repairJSONStringLiterals for how that distinction is made. Recovered
// 2026-07-16 after one live page (finetuning.uk tool-ai-data-risk-checker-guide)
// needed a manual one-off repair: its article ends in a contact link, so its
// response was COMPLETE but had this exact malformation, and kept failing every
// automated regeneration attempt (see aaa_fails_to_mend/005 §Follow-up).
//
// A TRUNCATED response (the dominant real cause — the writer's generate_content
// step was capped at max_tokens=2000 and cut articles off mid-sentence) is
// genuinely incomplete and still fails to parse after both repairs: a truncated
// string never reaches a real closing quote, so it just runs to end-of-input
// unterminated. That must NOT be silently salvaged — it falls through to the
// text path, where missingRequiredLLMFields makes the renderer refuse to ship a
// blank section.

package actions

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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

	fixed := repairJSONStringLiterals(trimmed)
	if err := json.Unmarshal([]byte(fixed), &value); err != nil {
		return nil, false, fmt.Errorf("unrecoverable after string-literal repair: %w", err)
	}
	return value, true, nil
}

// repairJSONStringLiterals walks a JSON document and fixes the two malformations
// inside string literals that models produce constantly: raw control characters,
// and raw (unescaped) double quotes — the latter from HTML attributes like
// href="...". Bytes outside string literals — structural whitespace, braces,
// commas — are copied untouched, so this cannot change the shape of a document,
// only rescue its string values.
//
// The hard part is telling a quote that ENDS a string from a quote that is just
// literal content sitting inside one: `"content": "<a href="mailto:x">Contact</a>"`
// has five quotes after the opening one, and only the LAST is the real
// terminator. The heuristic: a quote is a real JSON string terminator only if,
// skipping any whitespace, the next byte is a JSON structural character — `,`
// `}` `]` (ends a value) or `:` (ends a key). Anything else following the quote
// — a letter, `>`, `/`, another `"` — means it's content, so it gets escaped and
// the string continues. This is exactly the ambiguity a genuinely TRUNCATED
// response can't fake: a truncated string runs out of input entirely before ever
// reaching a quote followed by real JSON structure, so it stays unterminated and
// json.Unmarshal still rejects it — the fail-loud path for truncation is intact.
//
// Escape state is tracked properly: a backslash inside a string escapes the next
// byte, so a legitimate \" does not end the string and a literal \\ before a
// quote does.
func repairJSONStringLiterals(s string) string {
	var b strings.Builder
	b.Grow(len(s) + len(s)/16)

	inString := false
	escaped := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
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
			if jsonStructureFollows(runes, i+1) {
				b.WriteRune(r)
				inString = false
			} else {
				// A literal quote sitting inside the string's content (an HTML
				// attribute, a quoted word) — escape it so the string continues.
				b.WriteString(`\"`)
			}

		case r < 0x20:
			// A raw control character inside a string literal. Emit the escape
			// the model should have sent.
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

// jsonStructureFollows reports whether, skipping whitespace, the rune at or
// after index i is a JSON structural character that can legally follow a string
// (`,` `}` `]` after a value, `:` after a key). End-of-input returns false: a
// quote with nothing real JSON structure after it is either mid-truncation or
// itself literal content, and either way should NOT be treated as the string's
// terminator — letting a truncated document's dangling string stay unterminated
// is what keeps json.Unmarshal rejecting it.
func jsonStructureFollows(runes []rune, i int) bool {
	for i < len(runes) {
		switch runes[i] {
		case ' ', '\t', '\n', '\r':
			i++
			continue
		case ',', '}', ']', ':':
			return true
		default:
			return false
		}
	}
	return false
}

// missingRequiredLLMFields returns the names of input_schema fields that are
// declared source:"llm" AND required:true but are absent (or empty) in content.
//
// The v2 schema shape is input_schema.fields.<name>.{source, required, type}.
// This is the check that turns "a required content field never arrived" from a
// silently-empty render into a loud, actionable failure — the article-body
// component declares exactly one such field, `content`.
//
// The field set is read via datahelpers.SchemaContentFields, so a component
// authored in the legacy JSON-Schema dialect (`properties`+`required[]`, no
// `fields`) is enforced too rather than silently passing with zero required
// fields — the fail-open that let bugs_open/026's required `news-listing`
// headline ship empty. The fail-loud tripwire for a reintroduced dialect is
// fired at the two gate CALL SITES (RenderComponentAction, rerender_page_sections)
// via datahelpers.WarnIfLegacyDialect — those paths reach the gate on a
// re-render/redeploy WITHOUT a fresh plan_sections pass, so they, not this pure
// reader, are where the render-side tripwire belongs (they hold the logger).
func missingRequiredLLMFields(inputSchema map[string]interface{}, content map[string]interface{}) []string {
	fields, ok, _ := datahelpers.SchemaContentFields(inputSchema)
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
