// FILE: platform/orchestration/datahelpers/component_schema_fields.go
//
// One reader of a component's input_schema field set, shared by every consumer
// that needs the declared content fields — the generation planner, the
// render-time required-field gate, and the post-deploy required-field audit.
// See bugs_closed/026 (the fail-open) and bugs_closed/265 (the dialect made
// unrepresentable at the table, migration 437).
//
// WHY THIS EXISTS
// ---------------
// A component declares its content fields in input_schema. The platform speaks
// the v2 dialect, `{"fields": {"<name>": {"source","required","type",...}}}`.
// An OLDER dialect exists — JSON-Schema, `{"type":"object","required":["x"],
// "properties":{"x":{...}}}` — which pre-dates v2. Since migration 437
// (2026-08-16, bugs_closed/265) content_components REFUSES it at the table:
// CHECK constraint chk_input_schema_no_legacy_dialect rejects any row whose
// input_schema carries a top-level "properties" key, whoever writes it — a seed,
// a script, this binary, the admin UI, a restore from a bak_* table. That is
// where the guarantee lives; everything below is the reader's tolerance, kept
// as defence in depth for a schema that reaches it from somewhere else.
//
// Before this helper, independent readers each did
// `inputSchema["fields"].(map[string]interface{})` and, on a miss, returned
// "no fields" — the SAME blind spot in each. In particular the generation
// planner then treated the component as "all fields from LLM" (so the writer was
// never told the field existed) and the render gate treated it as "no required
// fields" (so an empty required field was never caught). A required `news-listing`
// headline was thereby neither requested nor enforced and served empty
// (bugs_closed/026). Reading one dialect and returning "nothing found" on every
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
// fromLegacy is the fail-loud signal. Do NOT read it as "the dialect is extinct,
// so this cannot happen": a comment here said exactly that (0 of 173, 2026-07-21)
// and four legacy rows were seeded by hand in the three weeks after it — a census
// is stale the day after it is taken (bugs_closed/265). What holds now is a
// CONSTRAINT: content_components refuses a top-level "properties" key
// (chk_input_schema_no_legacy_dialect, migration 437), and the three remaining
// legacy rows were converted by the same migration. So a true here means the
// schema did not come from content_components under that constraint — the
// constraint has been dropped, or the caller read a bak_*/versions table, or an
// in-memory schema is being inspected before a write (which the birth path does
// on purpose, see IsLegacyInputSchemaDialect). Callers on a build/audit path
// should surface it via WarnLegacyDialect rather than absorb it silently. The
// returned map for the legacy case is freshly built; the v2 case shares the
// caller's map (read-only use only).
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

// WarnLegacyDialect is the fleet-wide tripwire for a legacy dialect that reached
// a reader. Call it on a build or audit path when SchemaContentFields reports
// fromLegacy. Nil logger is a no-op (tests, non-logging call sites). `where`
// names the call site (e.g. "plan_sections", "check_required_fields_missing").
//
// What a firing MEANS changed with migration 437 (bugs_closed/265). Before it, this
// was the only detector for a reintroduced dialect and it was a Warn line nobody
// read — four rows passed six call sites for up to 15 days unseen. Now the table
// itself refuses the dialect, so this cannot fire on a content_components row
// unless the constraint chk_input_schema_no_legacy_dialect is gone. That is why
// it is an Error and says so: a firing is evidence about the constraint, not
// about a seed. Check `\d content_components` first.
func WarnLegacyDialect(logger *zap.Logger, where string, component string) {
	if logger == nil {
		return
	}
	logger.Error("SchemaContentFields: projected a LEGACY JSON-Schema input_schema dialect. content_components refuses this dialect by CHECK constraint chk_input_schema_no_legacy_dialect (migration 437) — if this schema came from that table the constraint has been dropped; otherwise the caller is reading a bak_*/component_versions copy or an in-memory schema (bugs_closed/265, bugs_closed/026)",
		zap.String("where", where),
		zap.String("component", component))
}

// IsLegacyInputSchemaDialect reports whether inputSchema would be REFUSED by
// content_components' CHECK constraint chk_input_schema_no_legacy_dialect
// (migration 437): a top-level "properties" key, whatever its value. This is
// deliberately the constraint's own predicate rather than SchemaContentFields'
// fromLegacy, which is narrower (a non-empty properties object AND no fields):
// a schema carrying BOTH keys reads as v2 but the table still refuses it, and
// the birth path must refuse the same set the table refuses or a generation
// dies on SQLSTATE 23514 instead of a message that names the fix. Nil and
// non-object schemas are not legacy (the constraint passes NULL).
func IsLegacyInputSchemaDialect(inputSchema map[string]interface{}) bool {
	if inputSchema == nil {
		return false
	}
	_, has := inputSchema["properties"]
	return has
}

// WarnIfLegacyDialect fires the tripwire iff inputSchema is the legacy dialect.
// A convenience for call sites that need the tripwire but consume the field set
// via a different helper (the render gate's paths call missingRequiredLLMFields,
// not SchemaContentFields directly) — so the render/rerender paths surface a
// reintroduced dialect even though they never run plan_sections. The extra read
// is a pure map lookup that costs nothing on the ordinary v2 shape. Nil logger
// no-ops.
//
// It fires on SchemaContentFields' `fromLegacy` — deliberately NARROWER than
// IsLegacyInputSchemaDialect, which the birth path uses. The two ask different
// questions: this one is "did a reader actually have to PROJECT a legacy
// schema", the other is "would the TABLE refuse this row" (which also catches a
// schema carrying both keys, where the reader is fine but the constraint is
// not). Keep them different; collapsing them would either mute this tripwire or
// let the birth path write a row the table rejects.
func WarnIfLegacyDialect(inputSchema map[string]interface{}, logger *zap.Logger, where string, component string) {
	if _, _, fromLegacy := SchemaContentFields(inputSchema); fromLegacy {
		WarnLegacyDialect(logger, where, component)
	}
}
