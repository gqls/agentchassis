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
//
// A THIRD malformation was added 2026-07-26 (bugs_open/088): the answer is
// COMPLETE but not alone — prose around it, a code fence around it, or the model
// visibly correcting itself and emitting the object twice. Nothing was truncated
// and nothing was mis-escaped; the response simply was not bare, so the whole
// thing was discarded and a page build died reporting "likely LLM truncation".
// extractCompleteJSONValue recovers those, and every rule in it is measured
// against the stored corpus rather than assumed — including the two rules that
// look obvious and are wrong, which are documented at the function so nobody
// re-introduces them.

package actions

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// Provenance values reported by ParseLLMJSONWithProvenance. Anything other than
// ProvenanceClean means the model wrapped its answer in something it was told not
// to emit, and the platform recovered it — the caller should record that, because
// an unmeasured recovery is how this class stayed invisible for four months.
const (
	ProvenanceClean     = "clean"           // parsed as sent
	ProvenanceRepaired  = "repaired"        // escaping-only repair (raw control chars / quotes)
	ProvenanceFenced    = "fenced_block"    // the answer was inside a ``` block, with prose around it
	ProvenanceUnwrapped = "prose_around"    // exactly one complete value, prose before and/or after it
	ProvenanceReemitted = "reemitted_value" // the model emitted its answer, then re-emitted the SAME SHAPE
)

// ParseLLMJSON parses an LLM response as JSON. Wrapper kept for the existing
// callers; see ParseLLMJSONWithProvenance for which recovery was used.
func ParseLLMJSON(s string) (value interface{}, repaired bool, err error) {
	value, prov, err := ParseLLMJSONWithProvenance(s)
	return value, prov != ProvenanceClean, err
}

// ParseLLMJSONWithProvenance parses an LLM response as JSON and reports HOW it
// got there.
//
// Three tiers, tried in order, each strictly narrower than "guess something":
//
//  1. parse as sent;
//  2. repair the escaping-only malformations models emit constantly (raw control
//     characters and unescaped quotes inside string values);
//  3. RECOVER a complete value the model buried in commentary — see
//     extractCompleteJSONValue.
//
// An error means the payload holds no COMPLETE JSON value (genuine prose, or a
// truncated document) and the caller must treat it as plain text.
//
// TIER 3 CANNOT WEAKEN THE TRUNCATION GUARD, and that is the property to preserve
// in any future edit: it only ever returns a value that decoded *completely*, and
// it refuses outright when anything after that value still looks like the start of
// another one — the signature of a document cut mid-write. Measured against 5,844
// stored responses (bugs_open/088): of the 647 today's parser rejects, 537 hold no
// complete value at all and every one of them is still rejected after this change.
func ParseLLMJSONWithProvenance(s string) (value interface{}, provenance string, err error) {
	if err := json.Unmarshal([]byte(s), &value); err == nil {
		return value, ProvenanceClean, nil
	}

	trimmed := strings.TrimSpace(s)

	// Tier 2 — escaping-only repair. Only on something structurally trying to be
	// a JSON object or array, so prose that merely failed to parse falls through.
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		fixed := repairJSONStringLiterals(trimmed)
		if err := json.Unmarshal([]byte(fixed), &value); err == nil {
			return value, ProvenanceRepaired, nil
		}
	}

	// Tier 3 — the answer is complete but not alone.
	if v, prov, ok := extractCompleteJSONValue(trimmed); ok {
		return v, prov, nil
	}

	return nil, "", fmt.Errorf("no complete JSON value in response")
}

// extractCompleteJSONValue recovers a COMPLETE JSON value from a response that
// also contains prose, a code fence, or a second copy of the answer.
//
// WHY THIS EXISTS
// ---------------
// bugs_open/088: a writer returned a complete object, then "Wait — I must scan for
// em dashes before returning. Found one in the headline. Rewriting now.", then a
// corrected object. json.Unmarshal fails on the trailing content, the raw-text
// envelope swallows the lot, and the required-field gate kills the whole page build
// at iteration 0 — reporting it as "likely LLM truncation", which it is not.
//
// EVERY RULE HERE IS DRAWN FROM THE STORED CORPUS, NOT FROM TASTE. Measured over
// 5,844 llm_call_log responses (2026-03-25 → 2026-07-26) that today's parser was
// given, of which it rejects 647:
//
//	537  hold no complete value at all — truncated or not JSON. STILL REJECTED.
//	 64  hold exactly one parseable fenced block. ZERO hold more than one, so
//	     "the fenced block" is unambiguous and needs no first/last policy.
//	 46  hold exactly one complete top-level value with prose around it.
//	 26  hold two or more complete top-level values — and this is where the
//	     obvious rule is WRONG: only 4 are the same answer re-emitted. 17 are
//	     DIFFERENT objects (a hero's {headline,subheadline}, then {content,heading},
//	     then {headline,subheadline,testimonials} — the model answering for several
//	     sections at once). "Take the last value" would have handed the hero section
//	     a testimonials object and called it success.
//
// So a multi-value response is only accepted when every value is an object with an
// IDENTICAL key set, which is what a re-emission looks like and what a
// several-answers-in-one-response does not.
func extractCompleteJSONValue(s string) (interface{}, string, bool) {
	// A MARKDOWN DOCUMENT is never unwrapped. Several agents are told to return
	// markdown that CONTAINS a fenced JSON block — experience-planner's compose
	// step says so in as many words: "Output the whole plan as markdown … the five
	// sections, the ```criteria fence, and then one final line exactly:
	// <!-- END EXPERIENCE_PLAN -->". For those the whole document IS the answer and
	// the fence is a part of it, so recovering the fence would replace a plan with
	// one of its own sub-blocks — the same fragment-for-whole substitution this
	// function exists to prevent, arrived at from the other direction.
	//
	// This is not a guess about intent: of the 93 responses an unguarded version of
	// this function recovered, 59 began with a markdown heading, and every one of
	// those belonged to experience-planner / tool-generator / generic — agents whose
	// steps ask for markdown. None of page-content-writer's did.
	if strings.HasPrefix(s, "#") {
		return nil, "", false
	}

	// A fenced block: an explicit "this is the payload" marker from the model, and
	// unambiguous in the corpus (no response ever held more than one that parses),
	// so there is no first-or-last policy to get wrong. Prose outside it is then
	// irrelevant.
	if blocks := fencedJSONBlocks(s); len(blocks) == 1 {
		return blocks[0], ProvenanceFenced, true
	}

	values, tail := scanTopLevelJSONValues(s)
	if len(values) == 0 {
		return nil, "", false
	}
	// Anything after the last complete value that still opens a JSON value is a
	// document cut mid-write. Refuse: returning the earlier value would ship a
	// SUPERSEDED answer and hide the truncation, which is strictly worse than
	// failing loud (bugs_closed/005, bug 026).
	if strings.ContainsAny(tail, "{[") {
		return nil, "", false
	}

	if len(values) == 1 {
		return values[0], ProvenanceUnwrapped, true
	}

	first, ok := jsonObjectKeySignature(values[0])
	if !ok {
		return nil, "", false
	}
	for _, v := range values[1:] {
		sig, ok := jsonObjectKeySignature(v)
		if !ok || sig != first {
			return nil, "", false
		}
	}
	// Same shape emitted more than once: the model corrected itself in the open.
	// Its LAST word is the answer it meant to give — in 088 the first copy carried
	// the em dash its own instructions forbade and the second did not.
	return values[len(values)-1], ProvenanceReemitted, true
}

// scanTopLevelJSONValues decodes consecutive TOP-LEVEL JSON values, skipping any
// prose between them, and returns the text remaining after the last one.
//
// It NEVER descends into a value that failed to decode. An earlier draft advanced
// one byte on failure and retried, which walked INTO a truncated array and reported
// each surviving element as a "top-level value" — 19 of them in one live response.
// Combined with "take the last value" that would have returned a single array
// element in place of a cut document: a fragment shipped as a whole answer, which
// is precisely bug 026. On a failed decode this stops.
func scanTopLevelJSONValues(s string) (values []interface{}, tailAfterLast string) {
	pos := strings.IndexAny(s, "{[")
	if pos < 0 {
		return nil, s
	}
	for {
		dec := json.NewDecoder(strings.NewReader(s[pos:]))
		var v interface{}
		if err := dec.Decode(&v); err != nil {
			return values, s[pos:]
		}
		values = append(values, v)
		pos += int(dec.InputOffset())
		if pos >= len(s) {
			return values, ""
		}
		next := strings.IndexAny(s[pos:], "{[")
		if next < 0 {
			return values, s[pos:]
		}
		pos += next
	}
}

// fencedJSONBlocks returns the parsed contents of every ``` block that is itself
// valid JSON. Parsing is via json.Unmarshal plus the escaping repair — NOT via
// ParseLLMJSONWithProvenance, so a fence can never re-enter this recovery path.
func fencedJSONBlocks(s string) []interface{} {
	var out []interface{}
	rest := s
	for {
		i := strings.Index(rest, "```")
		if i < 0 {
			return out
		}
		rest = rest[i+3:]
		// Drop the language tag ("json", "html", …) up to the end of its line.
		if nl := strings.IndexByte(rest, '\n'); nl >= 0 && nl <= 16 {
			rest = rest[nl+1:]
		}
		body := rest
		if j := strings.Index(rest, "```"); j >= 0 {
			body = rest[:j]
			rest = rest[j+3:]
		} else {
			rest = ""
		}
		body = strings.TrimSpace(body)
		if body == "" {
			if rest == "" {
				return out
			}
			continue
		}
		var v interface{}
		if err := json.Unmarshal([]byte(body), &v); err == nil {
			out = append(out, v)
		} else if strings.HasPrefix(body, "{") || strings.HasPrefix(body, "[") {
			if err := json.Unmarshal([]byte(repairJSONStringLiterals(body)), &v); err == nil {
				out = append(out, v)
			}
		}
		if rest == "" {
			return out
		}
	}
}

// jsonObjectKeySignature returns a stable signature of an object's key set, and
// false for anything that is not an object. Two values with the same signature are
// the same answer emitted twice; two with different signatures are two different
// answers, and the platform must not choose between them.
func jsonObjectKeySignature(v interface{}) (string, bool) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return "", false
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, "\x00"), true
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
