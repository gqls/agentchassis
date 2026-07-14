package actions

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// The checker's pure functions are safety-relevant: a short error prefix would
// split one platform defect into per-site triage patterns (burning the
// escalation cap), and an unstable item_key would flood site_work_items on
// every sweep. Test the real functions.

// Triage groups failure patterns by left(error, 140). Every check's fixed
// prefix MUST reach 140 chars so all sites collapse into ONE pattern.
func TestSilentErrorPrefixLongEnoughToGroupAcrossSites(t *testing.T) {
	for _, c := range silentChecks {
		if n := len(silentErrorPrefix(c)); n < 140 {
			t.Fatalf("check %s: error prefix is %d chars; must be >= 140 or triage splits the pattern per site", c.Name, n)
		}
	}
}

func TestSilentErrorTextSameTriagePatternAcrossSites(t *testing.T) {
	c := silentChecks[0]
	a := silentErrorText(c, silentGroup{Domain: "dartsonline.com", Pages: []silentViolation{{PageName: "guides-index"}}})
	b := silentErrorText(c, silentGroup{Domain: "idea.uk", Pages: []silentViolation{{PageName: "news-index"}}})
	if a[:140] != b[:140] {
		t.Fatal("first 140 chars must be identical across sites — that is triage's grouping key")
	}
	if !strings.Contains(a, "site=dartsonline.com") || !strings.Contains(a, "pages=guides-index") {
		t.Fatalf("site detail missing from error text: %s", a)
	}
}

func TestSilentItemKeyStablePerCheckAndSite(t *testing.T) {
	k1 := silentItemKey("nav_linked_never_built", "1a2b3c4d-0000-0000-0000-000000000000")
	k2 := silentItemKey("nav_linked_never_built", "1a2b3c4d-0000-0000-0000-000000000000")
	if k1 != k2 {
		t.Fatalf("same (check, site) must yield same key: %s vs %s", k1, k2)
	}
	if k1 != "silent:nav_linked_never_built:1a2b3c4d" {
		t.Fatalf("key not readable/prefixed: %s", k1)
	}
	if silentItemKey("deployed_zero_components", "1a2b3c4d-0000-0000-0000-000000000000") == k1 {
		t.Fatal("different check must yield a different key")
	}
	if silentItemKey("nav_linked_never_built", "ffffffff-0000-0000-0000-000000000000") == k1 {
		t.Fatal("different site must yield a different key")
	}
}

func TestSilentGroupBySiteDeterministicAndComplete(t *testing.T) {
	vs := []silentViolation{
		{SiteID: "b-site", Domain: "b.com", PageName: "p1"},
		{SiteID: "a-site", Domain: "a.com", PageName: "p2"},
		{SiteID: "b-site", Domain: "b.com", PageName: "p3"},
	}
	groups := silentGroupBySite("nav_linked_never_built", vs)
	if len(groups) != 2 {
		t.Fatalf("want 2 groups, got %d", len(groups))
	}
	if groups[0].SiteID != "a-site" || groups[1].SiteID != "b-site" {
		t.Fatalf("groups must be site-sorted for deterministic capping: %+v", groups)
	}
	if len(groups[1].Pages) != 2 {
		t.Fatalf("b-site must carry both its pages, got %d", len(groups[1].Pages))
	}
}

func TestSilentSpecJSONShape(t *testing.T) {
	c := silentChecks[0]
	g := silentGroup{Domain: "dartsonline.com", Pages: []silentViolation{
		{PageID: "pid-1", PageName: "guides-index", PageType: "section-index", Build: "planned"},
	}}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(silentSpecJSON(c, g, 48)), &m); err != nil {
		t.Fatalf("spec not valid JSON: %v", err)
	}
	for _, k := range []string{"check", "invariant", "site_domain", "pages", "grace_hours", "coverage_rule", "source"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("spec missing key %q", k)
		}
	}
	pages, ok := m["pages"].([]interface{})
	if !ok || len(pages) != 1 {
		t.Fatalf("pages must be a 1-element array: %v", m["pages"])
	}
}

func TestRenderSilentCheckDryRunAndCapNeverSilent(t *testing.T) {
	sections := []silentReportSection{{
		Name: "nav_linked_never_built", Description: "desc", Emits: true,
		Violations: []silentViolation{{Domain: "dartsonline.com", PageName: "guides-index", PageType: "section-index", Build: "planned"}},
	}}
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)

	dry := renderSilentCheck(now, 48, sections, 0, 0, 0, 3, 0, true)
	if !strings.Contains(dry, "DRY RUN") {
		t.Fatal("dry-run report must announce itself")
	}
	if strings.Contains(dry, "Bookkeeping") {
		t.Fatal("dry-run report must not show write bookkeeping — nothing was written")
	}

	live := renderSilentCheck(now, 48, sections, 1, 2, 3, 3, 1, false)
	for _, want := range []string{"Emitted 1", "deduped (already open) 2", "capped 3", "closed as resolved 1", "NOT emitted this sweep"} {
		if !strings.Contains(live, want) {
			t.Fatalf("live report missing %q:\n%s", want, live)
		}
	}
	if !strings.Contains(live, "guides-index") || !strings.Contains(live, "dartsonline.com") {
		t.Fatal("report must name the violating pages and sites")
	}
}
