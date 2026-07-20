// FILE: platform/orchestration/datahelpers/ctafields.go
//
// Schema-derived CTA field pairing — the single definition of "this input_schema
// field is a button destination", shared by resolve_internal_links (build-time
// resolution), rerender_page_sections (repair/observe), and any future check.
// Exists for the same reason links.go does: three consumers previously answered
// "which fields are CTAs" from a hand-maintained map (ctaFieldNames) that four
// migrations (091, 096, 097b, 098) patched with the same lesson, and that covers
// 5 of the 33 component functions which actually declare CTA url fields.
//
// STAGED ROLLOUT (council trail 2525f980): in the current stage the derivation
// is OBSERVE-ONLY — callers log the delta against ctaFieldNames but the map
// still decides every write. Precedence inverts only after a real build's delta
// log has been reviewed, via a further council round. doc_notes rows under
// subject_keys resolve_internal_links / rerender_page_sections carry this
// context; read them before changing precedence.
//
// Derivation rule (corrected per doc_notes 2026-07-20, commit b6e374fc2):
//   - a field ending "_url" is a CTA destination iff a sibling field of the
//     SAME stem exists in one of THREE forms: "<stem>_label", "<stem>_text",
//     or the bare "<stem>" (hero.secondary_cta_url pairs with secondary_cta —
//     3 of ctaFieldNames' own 10 fields pair bare-stem);
//   - EXCEPT when the url field's own source is site_assets.* — those are
//     images (logo_url has sibling logo_text but is a logo, not a button).

package datahelpers

import (
	"encoding/json"
	"sort"
	"strings"
)

// CTAField is one derived button-destination field and its paired label.
type CTAField struct {
	URLField   string // e.g. "cta_primary_url"
	LabelField string // e.g. "cta_primary_label", "cta_text", or bare "secondary_cta"
	Source     string // the url field's declared source, so callers can reason about authority
}

// ParseInputSchemaValue normalises a component's input_schema, which arrives as
// a JSON string after a DB round-trip or an already-parsed map in-process.
// One parse for all callers (extracted so loadSingleComponentSchema's inline
// branching can converge onto it at the precedence flip).
func ParseInputSchemaValue(v interface{}) map[string]interface{} {
	switch s := v.(type) {
	case map[string]interface{}:
		return s
	case string:
		if s == "" {
			return nil
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(s), &m); err == nil {
			return m
		}
	}
	return nil
}

// ctaSiblingSuffixes are the label forms a url field's stem may pair with, in
// match-preference order. The empty suffix is the bare stem.
var ctaSiblingSuffixes = []string{"_label", "_text", ""}

// queryResolvedPrefixes are the source classes plan_sections re-resolves from
// the DB on every render (diagnosis c84940fb) — the classes whose values can
// silently overwrite a resolver-written destination. site_assets.* is
// deliberately absent: those are image fields, never button destinations.
var queryResolvedPrefixes = []string{"site_specs.", "pages.", "config."}

func fieldSource(raw interface{}) string {
	spec, _ := raw.(map[string]interface{})
	if spec == nil {
		return ""
	}
	src, _ := spec["source"].(string)
	return src
}

// DeriveCTAURLFields returns the CTA destination fields a schema declares,
// per the derivation rule above. Output is sorted by URLField for determinism.
func DeriveCTAURLFields(inputSchema map[string]interface{}) []CTAField {
	fields, _ := inputSchema["fields"].(map[string]interface{})
	if len(fields) == 0 {
		return nil
	}
	var out []CTAField
	for name, raw := range fields {
		if !strings.HasSuffix(name, "_url") {
			continue
		}
		if strings.HasPrefix(fieldSource(raw), "site_assets.") {
			continue // image/logo guard: logo_url + logo_text must not derive
		}
		stem := strings.TrimSuffix(name, "_url")
		label := ""
		for _, suf := range ctaSiblingSuffixes {
			if _, ok := fields[stem+suf]; ok {
				label = stem + suf
				break
			}
		}
		if label == "" {
			continue
		}
		out = append(out, CTAField{URLField: name, LabelField: label, Source: fieldSource(raw)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URLField < out[j].URLField })
	return out
}

// UncoveredCTAURLFields returns url fields the derivation CANNOT see — a
// query-resolved source with no sibling in any form. This is ctaFieldNames'
// old silent failure mode made loud: callers log each at Warn. Warn-only in
// the observe stage; the precedence flip must replace this with a consumed
// detection path (a work-item type with a named handler) before it ships.
func UncoveredCTAURLFields(inputSchema map[string]interface{}) []string {
	fields, _ := inputSchema["fields"].(map[string]interface{})
	if len(fields) == 0 {
		return nil
	}
	derived := make(map[string]bool)
	for _, cf := range DeriveCTAURLFields(inputSchema) {
		derived[cf.URLField] = true
	}
	var out []string
	for name, raw := range fields {
		if !strings.HasSuffix(name, "_url") || derived[name] {
			continue
		}
		src := fieldSource(raw)
		for _, p := range queryResolvedPrefixes {
			if strings.HasPrefix(src, p) {
				out = append(out, name)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}
