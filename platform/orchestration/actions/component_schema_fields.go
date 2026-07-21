// FILE: platform/orchestration/actions/component_schema_fields.go
//
// One reader of a component's input_schema field set, understood by BOTH the
// generation planner (plan_sections) and the render-time required-field gate
// (missingRequiredLLMFields). See bugs_open/026.
//
// WHY THIS EXISTS
// ---------------
// A component declares its content fields in input_schema. The platform speaks
// the v2 dialect, `{"fields": {"<name>": {"source","required","type",...}}}`.
// An OLDER dialect exists — JSON-Schema, `{"type":"object","required":["x"],
// "properties":{"x":{...}}}` — which pre-dates v2 and can re-appear via a config
// re-seed, a restored site snapshot, or a component-creator run that emits
// JSON-Schema instead of `fields`.
//
// Before this helper, two independent readers each did
// `inputSchema["fields"].(map[string]interface{})` and, on a miss, returned
// "no fields" — the SAME blind spot in both:
//
//   - plan_sections (generation): a missed schema became "all fields from LLM"
//     with an empty field-spec list, so the page-content-writer was never told
//     the component even had the field — it was never generated.
//   - missingRequiredLLMFields (enforcement): a missed schema became nil, so a
//     required field that rendered empty was never caught.
//
// The result (bugs_open/026, shared `news-listing`): a `required` headline was
// neither requested nor enforced, and served as an empty <h1>. Reading only one
// dialect does not fail safe — it fails OPEN: "I can't read this contract"
// becomes "there is no contract". This helper closes that by projecting the
// legacy dialect onto the v2 field view both callers already consume, so an
// old-shape component is UNDERSTOOD, not silently shipped empty.
//
// It deliberately does NOT invent fields for the oldest bare example-value
// schemas (e.g. hero/header/footer: `{"headline":"string",...}`) — those declare
// no machine-readable requiredness, so there is nothing to enforce and the
// caller's existing "no declared fields" path is preserved (returns ok=false).

package actions

// schemaContentFields returns a component's declared content fields as the v2
// `fields` map, regardless of the input_schema dialect.
//
//   - v2 (`fields` present): returned as-is (ok=true, even when empty — the
//     callers already treat an empty field set as "no declared content fields").
//   - legacy JSON-Schema (`properties` present, no `fields`): projected onto the
//     v2 field shape, with the top-level `required[]` array folded into each
//     field's `required` flag.
//   - neither (empty {} or a bare example-value map): ok=false, so the caller
//     keeps its existing no-declared-fields behaviour.
//
// The returned map for the legacy case is freshly built; the v2 case shares the
// caller's map (read-only use only).
func schemaContentFields(inputSchema map[string]interface{}) (map[string]interface{}, bool) {
	if inputSchema == nil {
		return nil, false
	}

	// v2 dialect — the current, common case.
	if f, ok := inputSchema["fields"].(map[string]interface{}); ok {
		return f, true
	}

	// Legacy JSON-Schema dialect: properties + a top-level required[] array.
	props, ok := inputSchema["properties"].(map[string]interface{})
	if !ok || len(props) == 0 {
		return nil, false
	}

	requiredSet := make(map[string]bool)
	if reqArr, ok := inputSchema["required"].([]interface{}); ok {
		for _, r := range reqArr {
			if name, ok := r.(string); ok {
				requiredSet[name] = true
			}
		}
	}

	fields := make(map[string]interface{}, len(props))
	for name, defRaw := range props {
		nf := map[string]interface{}{}
		if def, ok := defRaw.(map[string]interface{}); ok {
			// Carry through the keys both callers read verbatim.
			for _, k := range []string{"source", "on_missing", "fallback", "missing_reason", "items", "min_items"} {
				if v, present := def[k]; present {
					nf[k] = v
				}
			}
			// JSON-Schema uses `minItems`; v2 reads `min_items`.
			if _, present := nf["min_items"]; !present {
				if mi, present := def["minItems"]; present {
					nf["min_items"] = mi
				}
			}
			// Type: map JSON-Schema "string" onto the v2 "text"; pass the rest
			// through (array/number/etc. read the same in both dialects).
			if t, ok := def["type"].(string); ok {
				if t == "string" {
					t = "text"
				}
				nf["type"] = t
			}
			// Guidance: the legacy dialect names it `description`; v2 names it
			// `llm_guidance`. Prefer an explicit v2 key if somehow present.
			if g, ok := def["llm_guidance"].(string); ok && g != "" {
				nf["llm_guidance"] = g
			} else if d, ok := def["description"].(string); ok && d != "" {
				nf["llm_guidance"] = d
			}
		}
		// A property with no explicit source is content the writer must supply —
		// default to "llm" so it is both planned for and enforced. The legacy
		// dialect pre-dates query/renderer sources, so this is the safe default.
		if _, ok := nf["source"]; !ok {
			nf["source"] = "llm"
		}
		if requiredSet[name] {
			nf["required"] = true
		}
		fields[name] = nf
	}
	return fields, true
}
