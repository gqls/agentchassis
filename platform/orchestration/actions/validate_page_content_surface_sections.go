package actions

import (
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// Per-COMPONENT surfaces for the prose number scan (RFC_053, bugs_open/364
// Phase 2).
//
// THE PROBLEM THIS SOLVES. The build gate scans the whole page's HTML as one
// string, so it can only ask a PAGE-grain question: "is this page's body
// editorial?". A tracker or directory page is neither wholly editorial nor
// wholly marketing — it carries a first-person `hero` and `call-to-action`
// wrapped around a listing of OTHER organisations' figures. Answering at page
// grain forces a choice between refusing honest pages over third-party
// statistics (what bugs_open/364 measured: 20 findings, precision ZERO) and
// silencing the site's own claims to buy that silence (what the 2026-08-22
// interim did, knowingly). Neither is the right answer; the grain was.
//
// WHY SPLITTING THE SCAN IS SAFE, and this was measured before it was written.
// ExtractAssertionText splits on block-level elements and every component is a
// block-level element, so no assertion block ever spans a component boundary.
// Verified 2026-08-25 across 4 pages / 21 components on ai-agent-orchestration
// .com by extracting whole-page and per-component and diffing the normalised
// block multisets: 121 vs 121 blocks on adoption-tracker and an exact match on
// all four, zero blocks appearing on either side alone. So the text the scanner
// reads is IDENTICAL either way — this changes which surface each block is
// judged against, and nothing else.
//
// THE FAIL DIRECTION IS THE SAME ONE THE REST OF THIS LAYER TAKES. Only
// page-build-handler supplies sections_metadata (it declares
// require_sections_metadata); the gate's other callers do not. When it is absent
// this returns ok=false and the caller keeps the whole-page, page-type-only scan
// — i.e. exactly today's behaviour, noisy rather than silent. A caller that
// cannot say which component it is reading must never inherit a component-grain
// exemption.
//
// It deliberately does NOT read the component identity out of the rendered HTML.
// That HTML is LLM-generated, so a marker embedded in it could be emitted by the
// very thing being policed; component_function comes from the framework's own
// section metadata. See datahelpers.thirdPartyDataComponents.

// pageSection is one component's bytes plus the identity the surface needs.
type pageSection struct {
	ComponentFunction string
	HTML              string
}

// collectPageSections reads per-section HTML and component identity out of
// sections_metadata. ok=false means "this caller cannot answer at component
// grain" — never "there are no sections", which is why the two are distinct.
func collectPageSections(collected map[string]interface{}, metaField string, logger *zap.Logger) ([]pageSection, bool) {
	raw := datahelpers.ExtractNestedField(collected, metaField)
	if raw == nil {
		return nil, false
	}
	sections, ok := raw.([]interface{})
	if !ok {
		logger.Debug("validate_page_content: sections_metadata is not a list — "+
			"prose number scan stays at page grain",
			zap.String("field", metaField))
		return nil, false
	}

	out := make([]pageSection, 0, len(sections))
	for _, s := range sections {
		sec, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		html := extractHTMLFromSectionMap(sec)
		if html == "" {
			continue
		}
		out = append(out, pageSection{
			ComponentFunction: sectionsMetadataComponentFunction(sec),
			HTML:              html,
		})
	}
	if len(out) == 0 {
		// Metadata was present but carried no usable section. Refuse to claim
		// component grain rather than scanning nothing: a page whose sections we
		// could not read must fall back to the whole-page scan, or this becomes a
		// silent way to skip the check entirely.
		return nil, false
	}
	return out, true
}

// sectionsMetadataComponentFunction resolves the component's registered function
// from a SECTIONS_METADATA element, in the same precedence the save path uses
// (save_page_sections_action.go:1406-1443): stored slot name, then
// component_function, then component_name.
//
// ⚠ NOT the same function as `sectionComponentFunction`
// (resolve_internal_links_action.go:846), and the two must not be unified: that
// one reads a PLANNED section, whose identity lives at `component.function` /
// `component.name` with `name` as the fallback, and it would return "" for every
// element of this shape. Same question, two different carriers — the shapes are
// what differ, not the intent. The name says which carrier this one reads.
//
// An unresolved identity returns "" — which reads as UNKNOWN at the surface and
// is therefore SCANNED. That is the safe direction, and it is why this returns a
// bare string rather than defaulting to something like "unknown-component":
// a placeholder could one day collide with a declared member.
func sectionsMetadataComponentFunction(sec map[string]interface{}) string {
	for _, key := range []string{"stored_slot_name", "component_function", "component_name"} {
		if v, ok := sec[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}
