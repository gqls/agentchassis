// FILE: platform/orchestration/datahelpers/component_schema_fields.go
//
// One reader of a component's input_schema field set, shared by every consumer
// that needs the declared content fields — the generation planner, the
// render-time required-field gate, and the post-deploy required-field audit.
// See bugs_open/026.
//
// WHY THIS EXISTS
// ---------------
// A component declares its content fields in input_schema. The platform speaks
// the v2 dialect, `{"fields": {"<name>": {"source","required","type",...}}}`.
// An OLDER dialect exists — JSON-Schema, `{"type":"object","required":["x"],
// "properties":{"x":{...}}}` — which pre-dates v2 and can re-appear via a config
// re-seed, a restored site snapshot, or a component-creator run that emits
// JSON-Schema.
//
// Before this helper, independent readers each did
// `inputSchema["fields"].(map[string]interface{})` and, on a miss, returned
// "no fields" — the SAME blind spot in each. In particular the generation
// planner then treated the component as "all fields from LLM" (so the writer was
// never told the field existed) and the render gate treated it as "no required
// fields" (so an empty required field was never caught). A required `news-listing`
// headline was thereby neither requested nor enforced and served empty
// (bugs_open/026). Reading one dialect and returning "nothing found" on every
// other dialect does not fail safe — it fails OPEN: "I can't read this contract"
// becomes "there is no contract".
//
// This helper projects the legacy dialect onto the v2 field view its callers
// already consume, so an old-shape component is UNDERSTOOD, not silently shipped
// empty. It lives in datahelpers so BOTH the actions package (planner, gate) and
// the discovery_checks package (post-deploy audit) share one implementation.
//
// It deliberately does NOT invent fields for the oldest bare example-value
// schemas (e.g. hero/header/footer: `{"headline":"string",...}`) — those declare
// no machine-readable requiredness, so there is nothing to enforce and the
// caller's existing "no declared fields" path is preserved.

package datahelpers

import "go.uber.org/zap"

// SchemaContentFields returns a component's declared content fields as the v2
// `fields` map, regardless of the input_schema dialect, plus a fromLegacy flag.
//
//   - v2 (`fields` present): returned as-is; ok=true, fromLegacy=false (even when
//     empty — callers already treat an empty field set as "no declared fields").
//   - legacy JSON-Schema (`properties` present, no `fields`): projected onto the
//     v2 field shape with the top-level `required[]` folded into each field's
//     `required` flag; ok=true, fromLegacy=true.
//   - neither (empty {} or a bare example-value map): ok=false, fromLegacy=false,
//     so the caller keeps its existing no-declared-fields behaviour.
//
// fromLegacy is the fail-loud signal: the legacy dialect is extinct fleet-wide
// (0 of 173 as at 2026-07-21), so a true here means a regression reintroduced it
// — callers on a build/audit path should surface it via WarnLegacyDialect rather
// than absorb it silently. The returned map for the legacy case is freshly built;
// the v2 case shares the caller's map (read-only use only).
func SchemaContentFields(inputSchema map[string]interface{}) (map[string]interface{}, bool, bool) {
	if inputSchema == nil {
		return nil, false, false
	}

	// v2 dialect — the current, common case.
	if f, ok := inputSchema["fields"].(map[string]interface{}); ok {
		return f, true, false
	}

	// Legacy JSON-Schema dialect: properties + a top-level required[] array.
	props, ok := inputSchema["properties"].(map[string]interface{})
	if !ok || len(props) == 0 {
		return nil, false, false
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
			for _, k := range []string{"source", "on_missing", "fallback", "missing_reason", "items", "min_items"} {
				if v, present := def[k]; present {
					nf[k] = v
				}
			}
			if _, present := nf["min_items"]; !present {
				if mi, present := def["minItems"]; present {
					nf["min_items"] = mi
				}
			}
			if t, ok := def["type"].(string); ok {
				if t == "string" {
					t = "text"
				}
				nf["type"] = t
			}
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
	return fields, true, true
}

// WarnLegacyDialect is the fleet-wide tripwire for a reintroduced legacy dialect.
// Call it on a build or audit path when SchemaContentFields reports fromLegacy,
// so a re-seed/snapshot-restore/component-creator regression that revives the
// extinct dialect is visible (a log line, per bugs_open/026's council review)
// rather than silently absorbed by the now-tolerant reader. Nil logger is a no-op
// (tests, non-logging call sites). `where` names the call site (e.g.
// "plan_sections", "check_required_fields_missing").
func WarnLegacyDialect(logger *zap.Logger, where string, component string) {
	if logger == nil {
		return
	}
	logger.Warn("SchemaContentFields: projected a LEGACY JSON-Schema input_schema dialect — this dialect should be extinct fleet-wide; a re-seed, snapshot restore, or component-creator run may have reintroduced it (bugs_open/026)",
		zap.String("where", where),
		zap.String("component", component))
}
