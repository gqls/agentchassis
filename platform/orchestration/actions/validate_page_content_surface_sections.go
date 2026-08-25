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
// statistics (bugs_open/364 measured 20 such findings, precision ZERO) and
// silencing the site's own claims to buy that silence (what the 2026-08-22
// interim did, knowingly). Neither is the right answer; the grain was.
//
// WHY SPLITTING THE SCAN IS SAFE, and this was measured before it was written.
// ExtractAssertionText splits on block-level elements and every component is a
// block-level element, so no assertion block ever spans a component boundary.
// Verified by extracting whole-page and per-component and diffing the normalised
// block multisets. First pass 2026-08-25: 4 pages / 21 components, exact match.
// The council's guardian seat called that sample thin for a fleet-wide refusal
// gate and was right to, so it was re-run over the WHOLE live corpus the same
// day — **775 pages / 2,042 components, export asserted 2042/2042 against the
// DB, ZERO pages where the two differ**. So the text the scanner reads is
// identical either way; this changes which surface each block is judged against,
// and nothing else.
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
// very thing being policed; the identity comes from the framework's own section
// metadata. See datahelpers.thirdPartyDataComponents.

// pageSection is one component's bytes plus the identity the surface needs.
type pageSection struct {
	ComponentFunction string
	HTML              string
}

// collectPageSections reads per-section HTML and component identity out of
// sections_metadata. ok=false means "this caller cannot answer at component
// grain" — never "there are no sections", which is why the two are distinct.
//
// IT REUSES extractSectionsFromMetadata RATHER THAN PARSING THE SHAPE ITSELF,
// and that is deliberate (council round 1, correlation 3ed2b792 — the
// reuse_agent, prior_art_librarian and constitution seats all objected to the
// first cut, which hand-parsed the same jsonb). That function is the canonical
// reader: it already resolves slot identity before component identity
// (bugs_open/189's fix), and the carry-key contract in section_metadata_keys.go
// is written against it. A second parser here would be a second thing to update
// when the shape changes — and the shape changing silently under a reader is
// exactly what bugs_open/357 cost a day of production for.
//
// ⚠ AN UNREADABLE SECTION MUST NOT BECOME AN UNSCANNED SECTION. This is the hole
// the bug_historian seat found in the first cut (medium), which did
// `if html == "" { continue }` and so collapsed two states into one silent skip:
// a genuinely empty section (fine), and a section that HAS html under a key the
// reader does not know (a shape drift). In the second case that component was
// dropped from the scan while the page was still reported as scanned — a partial,
// silent loss of coverage on the ONLY gate that refuses a page.
//
// The guard is now a count rather than a shape test, which is why it is strong:
// if the canonical reader returns fewer sections than the metadata array held,
// SOMETHING was dropped — a non-map entry, an empty html, a drifted key, any of
// them — and we refuse component grain for the whole page. The caller then scans
// the page as one string, which still reads those bytes. That converts a silent
// coverage hole into, at worst, the pre-Phase-2 false positives: loud, and
// already understood.
func collectPageSections(collected map[string]interface{}, metaField string, logger *zap.Logger) ([]pageSection, bool) {
	raw := datahelpers.ExtractNestedField(collected, metaField)
	if raw == nil {
		return nil, false
	}
	items, ok := raw.([]interface{})
	if !ok {
		logger.Debug("validate_page_content: sections_metadata is not a list — "+
			"prose number scan stays at page grain",
			zap.String("field", metaField))
		return nil, false
	}
	if len(items) == 0 {
		return nil, false
	}

	sections := extractSectionsFromMetadata(raw, logger)
	if len(sections) != len(items) {
		logger.Warn("validate_page_content: sections_metadata carried section(s) the canonical reader "+
			"could not resolve — falling back to the whole-page prose number scan rather than scanning "+
			"a subset of the page (bugs_open/364 Phase 2)",
			zap.Int("sections_in_metadata", len(items)),
			zap.Int("sections_resolved", len(sections)),
			zap.String("field", metaField))
		return nil, false
	}

	out := make([]pageSection, 0, len(sections))
	for i := range sections {
		out = append(out, pageSection{
			// ComponentName is the resolved slot identity — stored_slot_name, then
			// component_function, then component_name (save_page_sections_action.go
			// :1406-1443). An unresolved identity is the empty string, which reads
			// as UNKNOWN at the surface and is therefore SCANNED.
			//
			// ⚠ It is the INSTANCE's slot, which is USUALLY but not always the
			// component's registered function: measured 2026-08-25, 106 of 2,033
			// live rows differ (`prose-0` vs `ported-prose`, `call_to_action` vs
			// `call-to-action`, `FAQ Section` vs `faq`). That divergence fails SAFE
			// here — a declared member whose slot differs simply is not matched, so
			// the component is scanned — but it is why a new entry in
			// thirdPartyDataComponents must be verified against live `slot_name`
			// values, not against `content_components.function`.
			ComponentFunction: sections[i].ComponentName,
			HTML:              sections[i].HTML,
		})
	}
	return out, true
}
