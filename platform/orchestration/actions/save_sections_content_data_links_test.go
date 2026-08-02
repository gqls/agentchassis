package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func contentAuditIndex() datahelpers.PageURLIndex {
	return datahelpers.NewPageURLIndex([]string{"/index.html", "/about.html", "/tools.html"})
}

// TestAuditSectionContentDataLinksAttributesTheComponent. datahelpers reports a
// path within one content_data map; only this layer knows whose map it was, and
// "which component authors phantoms" is the question the durable record exists
// to answer. A finding with no component name is a finding nobody can route.
func TestAuditSectionContentDataLinksAttributesTheComponent(t *testing.T) {
	sections := []SectionData{
		{
			ComponentName: "info-card-grid",
			ContentData: map[string]interface{}{
				"cards": []interface{}{
					map[string]interface{}{"link_url": "/about"},   // rewritable
					map[string]interface{}{"link_url": "/pricing"}, // phantom
				},
			},
		},
		{
			ComponentName: "hero",
			ContentData:   map[string]interface{}{"cta_url": "/tools.html"}, // clean
		},
	}

	got := auditSectionContentDataLinks(sections, contentAuditIndex(), true)
	if len(got) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(got), got)
	}
	for _, f := range got {
		if f.Component != "info-card-grid" {
			t.Errorf("finding %+v: component = %q, want info-card-grid", f, f.Component)
		}
	}
	if got[0].Action != datahelpers.LinkRepairRewrite || got[0].NewHref != "/about.html" {
		t.Errorf("first finding = %+v, want a rewrite to /about.html", got[0])
	}
	if got[1].Action != datahelpers.ContentDataLinkPhantom {
		t.Errorf("second finding = %+v, want phantom", got[1])
	}

	// The rewrite must land in the map that SavePageSectionsAction is about to
	// marshal — auditing a copy would report a correction nobody persists.
	cards := sections[0].ContentData["cards"].([]interface{})
	if v := cards[0].(map[string]interface{})["link_url"]; v != "/about.html" {
		t.Errorf("rewrite not visible to the caller's section: %v", v)
	}

	rewritten, phantom := countContentDataLinkFindings(got)
	if rewritten != 1 || phantom != 1 {
		t.Errorf("counts = %d/%d, want 1 rewritten / 1 phantom", rewritten, phantom)
	}
}

// TestAuditDegradesOnAnUntrustworthyIndex. Same fail-open rule as
// repairSectionLinks, and it must be provable rather than asserted: an index
// that could not be read means "we do not know this site's pages", so judging
// against it would call real links phantoms and rewrite nothing correctly.
func TestAuditDegradesOnAnUntrustworthyIndex(t *testing.T) {
	newSections := func() []SectionData {
		return []SectionData{{
			ComponentName: "info-card-grid",
			ContentData:   map[string]interface{}{"link_url": "/about"},
		}}
	}

	// indexOK false, index populated: the load failed, the contents are not to
	// be trusted even though they look usable.
	s := newSections()
	if got := auditSectionContentDataLinks(s, contentAuditIndex(), false); got != nil {
		t.Fatalf("indexOK=false must audit nothing, got %+v", got)
	}
	if s[0].ContentData["link_url"] != "/about" {
		t.Fatalf("content_data mutated on the degrade path")
	}

	// indexOK true, index empty: an empty page set is the same claim by another
	// route, and the helper refuses it for the same reason.
	s = newSections()
	if got := auditSectionContentDataLinks(s, datahelpers.NewPageURLIndex(nil), true); got != nil {
		t.Fatalf("empty index must audit nothing, got %+v", got)
	}
	if s[0].ContentData["link_url"] != "/about" {
		t.Fatalf("content_data mutated against an empty index")
	}
}

// TestAuditSkipsRuntimeFillSectionsInLockstepWithTheMarkupPass. The markup
// repair declines a runtime-fill section; if this pass did not decline the same
// section, content_data and rendered_html would disagree about the same link
// until the next re-render — and the stored value is a placeholder the client
// replaces, so there is nothing true to judge.
func TestAuditSkipsRuntimeFillSectionsInLockstepWithTheMarkupPass(t *testing.T) {
	shell := `<div ` + datahelpers.RuntimeFillMarker + `><a href="/about">x</a></div>`
	sections := []SectionData{{
		ComponentName: "info-card-grid",
		HTML:          shell,
		ContentData:   map[string]interface{}{"link_url": "/about"},
	}}
	if got := auditSectionContentDataLinks(sections, contentAuditIndex(), true); got != nil {
		t.Fatalf("a runtime-fill section must be skipped, got %+v", got)
	}
	if sections[0].ContentData["link_url"] != "/about" {
		t.Fatalf("content_data of a runtime-fill section was rewritten")
	}

	// Control: the identical section without the marker IS audited, so the test
	// above is proving the exemption rather than the absence of a link.
	sections[0].HTML = `<div><a href="/about">x</a></div>`
	if got := auditSectionContentDataLinks(sections, contentAuditIndex(), true); len(got) != 1 {
		t.Fatalf("control: expected 1 finding without the marker, got %+v", got)
	}
}

// TestAuditIgnoresSectionsWithNoContentData: the html_field fallback path
// (save_page_sections_action.go:190) produces sections with HTML and no
// structured data. Those are the markup pass's business, not this one's, and a
// nil map must not panic the save.
func TestAuditIgnoresSectionsWithNoContentData(t *testing.T) {
	sections := []SectionData{
		{ComponentName: "generic-text-block", HTML: `<a href="/pricing">x</a>`},
		{ComponentName: "empty-map", ContentData: map[string]interface{}{}},
	}
	if got := auditSectionContentDataLinks(sections, contentAuditIndex(), true); got != nil {
		t.Fatalf("expected no findings, got %+v", got)
	}
}

// TestContentDataLinkErrorCodeIsDistinct — the estate's existing convention,
// which this change nearly shipped without following (caught by the council's
// guardian and reuse_agent seats, corr 40c0c14d, both at low severity on "a
// third code in this family").
//
// It matters in BOTH directions. A code that collides makes "which path caught
// this" unanswerable — the drift bugs_open/097 exists to keep answerable. And a
// code that merely SHARES A PREFIX with a live one is caught by any query using
// LIKE: the estate has two such queries today (`tool_crosslink_not_emitted%`,
// `component_validation_%`), so prefix-disjointness is a real property, not a
// stylistic one. CONTENT_DATA_ vs CONTENT_LINK_ diverge at the ninth character.
func TestContentDataLinkErrorCodeIsDistinct(t *testing.T) {
	taken := []string{
		"CONTENT_LINK_REPAIR_DETAIL", "CONTENT_LINK_REPAIR_SKIPPED",
		"CONTENT_CLAIMS_FLOOR_DETAIL", "CONTENT_VALIDATION_FAILED",
		"CONTENT_VALIDATION_BLOCKER_DETAIL", "TRUNCATION_DEGRADED_REVIEW",
		"VALIDATION_ERROR_DROPPED", "UNKNOWN",
	}
	for _, other := range taken {
		if contentDataLinkErrorCode == other {
			t.Errorf("content_data audit code collides with live code %q", other)
		}
		if strings.HasPrefix(contentDataLinkErrorCode, other) || strings.HasPrefix(other, contentDataLinkErrorCode) {
			t.Errorf("content_data audit code %q shares a prefix with live code %q — a LIKE query would catch both",
				contentDataLinkErrorCode, other)
		}
	}
}
