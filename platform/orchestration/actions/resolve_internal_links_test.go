package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestChooseCTATargetsInteractiveFirst(t *testing.T) {
	hubs := []contentHub{
		{Name: "archetypes", Title: "The Archetypes", URL: "/archetypes.html", NavOrder: 10},
		{Name: "provocations", Title: "Provocations", URL: "/provocations.html", NavOrder: 20},
	}
	interactive := []contentHub{
		{Name: "tool-quiz", Title: "Archetype Quiz", URL: "/tools/quiz/index.html", Area: "tools", NavOrder: 40},
		{Name: "tool-gauntlet", Title: "The Gauntlet", URL: "/tools/gauntlet/index.html", Area: "tools", NavOrder: 30},
	}

	primary, secondary := chooseCTATargets("", "index", interactive, hubs)
	if primary.URL != "/tools/gauntlet/index.html" {
		t.Errorf("primary = %q, want the gauntlet (interactive, lowest nav_order)", primary.URL)
	}
	if secondary.URL != "/tools/quiz/index.html" {
		t.Errorf("secondary = %q, want the quiz (interactive beats hubs)", secondary.URL)
	}

	// v1 behaviour preserved: no interactive pages -> hubs by nav_order.
	primary, secondary = chooseCTATargets("", "index", nil, hubs)
	if primary.URL != "/archetypes.html" || secondary.URL != "/provocations.html" {
		t.Errorf("hubs-only = (%q, %q), want archetypes/provocations", primary.URL, secondary.URL)
	}

	// Self-exclusion: a page's own name never becomes its CTA target.
	primary, _ = chooseCTATargets("", "tool-gauntlet", interactive, hubs)
	if primary.URL != "/tools/quiz/index.html" {
		t.Errorf("self-exclusion: primary = %q, want the quiz", primary.URL)
	}

	// No candidates at all -> zero values (gated templates render no button).
	primary, secondary = chooseCTATargets("", "index", nil, nil)
	if primary.URL != "" || secondary.URL != "" {
		t.Errorf("empty = (%q, %q), want empty", primary.URL, secondary.URL)
	}
}

// TestChooseCTATargetsRefusesAnOptedOutPage pins bugs_open/436's lever at the
// wrapper all three callers share — build resolve, rerender recompute, and the
// site HEADER fallback, whose exact call shape (pageName "") the second half
// uses: that caller's output is never persisted, so this test is the only
// place its behaviour change is cheaply visible. The ranking itself is pinned
// in datahelpers/cta_positional_test.go; this asserts the wrapper actually
// routes through it.
func TestChooseCTATargetsRefusesAnOptedOutPage(t *testing.T) {
	interactive := []contentHub{
		{Name: "tool-fossil", Title: "Off-Topic Toy", URL: "/tools/fossil.html", Area: "tools", NavOrder: 1, IneligibleAsCTATarget: true},
		{Name: "tool-on-topic", Title: "On Topic", URL: "/tools/on-topic.html", Area: "tools", NavOrder: 100},
	}
	primary, _ := chooseCTATargets("", "index", interactive, nil)
	if primary.URL != "/tools/on-topic.html" {
		t.Errorf("build-path primary = %q, want the eligible tool", primary.URL)
	}
	// The header fallback's call shape: pageType "", pageName "".
	primary, _ = chooseCTATargets("", "", interactive, nil)
	if primary.URL != "/tools/on-topic.html" {
		t.Errorf("header-fallback primary = %q, want the eligible tool", primary.URL)
	}
}

func TestChooseCTATargetsExcludesUtilityAreas(t *testing.T) {
	hubs := []contentHub{
		{Name: "contact-hub", Title: "Contact", URL: "/contact/index.html", Area: "contact", NavOrder: 1},
		{Name: "guides", Title: "Guides", URL: "/guides/index.html", Area: "guides", NavOrder: 2},
	}
	primary, _ := chooseCTATargets("", "index", nil, hubs)
	if primary.URL != "/guides/index.html" {
		t.Errorf("primary = %q, want guides (contact excluded)", primary.URL)
	}
}

func TestCTAExcludedDestination(t *testing.T) {
	cases := map[string]bool{
		"/contact.html":              true, // top-level page — firstPathSegment blind spot
		"/contact/index.html":        true,
		"/about.html":                true,
		"/legal/terms.html":          true,
		"/tools/gauntlet/index.html": false,
		"/guides/index.html":         false,
		"/":                          false,
	}
	for url, want := range cases {
		if got := ctaExcludedDestination(url); got != want {
			t.Errorf("ctaExcludedDestination(%q) = %v, want %v", url, got, want)
		}
	}
}

func TestCTATargetTitleField(t *testing.T) {
	cases := map[string]string{
		"cta_url":           "cta_target_title",
		"primary_cta_url":   "primary_cta_target_title",
		"secondary_cta_url": "secondary_cta_target_title",
	}
	for field, want := range cases {
		if got := ctaTargetTitleField(field); got != want {
			t.Errorf("ctaTargetTitleField(%q) = %q, want %q", field, got, want)
		}
	}
}

func TestApplyCTARecompute(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{
		"/index.html", "/contact.html", "/archetypes.html",
		"/tools/gauntlet/index.html", "/tools/quiz/index.html",
	})
	gauntlet := contentHub{Name: "tool-gauntlet", Title: "The Gauntlet", URL: "/tools/gauntlet/index.html"}
	pageURL := "/index.html"

	run := func(storedURL string) map[string]interface{} {
		resolved := map[string]interface{}{}
		stored := map[string]interface{}{}
		if storedURL != "" {
			stored["cta_url"] = storedURL
		}
		applyCTARecompute(resolved, stored, "cta_url", gauntlet, valid, pageURL, "", nil, "")
		return resolved
	}

	// Authored link to a real, sensible page: KEPT (nothing written).
	if got := run("/archetypes.html"); len(got) != 0 {
		t.Errorf("authored valid link overwritten: %v", got)
	}
	// Utility-area destination (/contact.html), valid and not self: KEPT.
	//
	// INVERTED 2026-08-17 for bugs_open/248, slug
	// cta_recompute_clobbers_authored_contact_links. This case used to assert
	// REPLACED, which enshrined the defect as the spec: a stored VALID
	// /contact.html cannot have come from this resolver (rank() will not pick a
	// utility page and candidatesFromHubs will not offer one), so it was
	// authored, and overwriting it destroyed working contact buttons on 13 live
	// components. Written, not merely left alone — see keep #1.
	if got := run("/contact.html"); got["cta_url"] != "/contact.html" {
		t.Errorf("authored utility destination not kept: %v", got)
	}
	// Phantom: REPLACED.
	if got := run("/services.html"); got["cta_url"] != gauntlet.URL {
		t.Errorf("phantom not replaced: %v", got)
	}
	// Circular (links back to its own page): REPLACED.
	if got := run("/index.html"); got["cta_url"] != gauntlet.URL {
		t.Errorf("circular link not replaced: %v", got)
	}
	// Missing entirely: WRITTEN, with the companion target title.
	got := run("")
	if got["cta_url"] != gauntlet.URL || got["cta_target_title"] != "The Gauntlet" {
		t.Errorf("missing field not filled with url+title: %v", got)
	}

	// No valid target to offer: stored value left alone even when bad.
	resolved := map[string]interface{}{}
	applyCTARecompute(resolved, map[string]interface{}{"cta_url": "/services.html"},
		"cta_url", contentHub{}, valid, pageURL, "", nil, "")
	if len(resolved) != 0 {
		t.Errorf("wrote a value with no valid target: %v", resolved)
	}

	// Empty field name (a single-URL component's absent secondary slot):
	// no-op, no panic, nothing written under "".
	resolved = map[string]interface{}{}
	applyCTARecompute(resolved, map[string]interface{}{}, "", gauntlet, valid, pageURL, "", nil, "")
	if len(resolved) != 0 {
		t.Errorf("empty field name wrote a value: %v", resolved)
	}
}

// TestCTAFieldNamesContract pins the covered components and their url field
// names to the live content_components.input_schema field names — a typo here
// silently turns a component's recompute into a no-op (the wrong key is never
// found in stored content_data and the write lands on a key no template reads).
func TestCTAFieldNamesContract(t *testing.T) {
	want := map[string][2]string{
		"hero":                   {"cta_url", "secondary_cta_url"},
		"call-to-action":         {"primary_cta_url", "secondary_cta_url"},
		"archetype-grid":         {"cta_url", ""},
		"archetype-combinations": {"cta_primary_url", "cta_secondary_url"},
		"gauntlet-cta":           {"cta_primary_url", "cta_secondary_url"},
		"content-block-about":    {"cta_url", ""},
	}
	if len(ctaFieldNames) != len(want) {
		t.Errorf("ctaFieldNames has %d entries, want %d — update this contract test with the schema evidence for any new component", len(ctaFieldNames), len(want))
	}
	for fn, fields := range want {
		got, ok := ctaFieldNames[fn]
		if !ok {
			t.Errorf("ctaFieldNames missing %q", fn)
			continue
		}
		if got != fields {
			t.Errorf("ctaFieldNames[%q] = %v, want %v", fn, got, fields)
		}
	}
}

// TestSetCTAFieldEmptyField — the build-time writer must also skip an absent
// secondary slot: nothing written AND nothing reported unresolved (an
// unresolved entry for field "" would create a bogus HITL work item).
func TestSetCTAFieldEmptyField(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{"/tools/gauntlet/index.html"})
	gauntlet := contentHub{Name: "tool-gauntlet", Title: "The Gauntlet", URL: "/tools/gauntlet/index.html"}

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, nil, "", gauntlet, valid, "archetype-grid", "archetype-grid", "secondary", &unresolved, "", nil, "")
	if len(resolved) != 0 {
		t.Errorf("empty field name wrote a value: %v", resolved)
	}
	if len(unresolved) != 0 {
		t.Errorf("empty field name reported unresolved: %v", unresolved)
	}
}

// TestChooseCTATargetsAllOptedOut pins the near-empty-after-filter behaviour
// the council's round-1 gating objection asked for (corr 9faa2a23): opting out
// every candidate must yield ZERO-VALUE targets — the long-documented "no
// sensible target" degrade ("Zero-value contentHub => no sensible target"; the
// gated template renders no button; setCTAField reports the slot unresolved
// and a needs_human_review item is filed) — never a panic and never an
// arbitrary survivor. One eligible survivor fills primary only.
func TestChooseCTATargetsAllOptedOut(t *testing.T) {
	interactive := []contentHub{
		{Name: "tool-a", Title: "A", URL: "/tools/a.html", Area: "tools", NavOrder: 1, IneligibleAsCTATarget: true},
		{Name: "tool-b", Title: "B", URL: "/tools/b.html", Area: "tools", NavOrder: 100, IneligibleAsCTATarget: true},
	}
	hubs := []contentHub{
		{Name: "guides", Title: "Guides", URL: "/guides/index.html", Area: "guides", NavOrder: 10, IneligibleAsCTATarget: true},
	}
	primary, secondary := chooseCTATargets("", "index", interactive, hubs)
	if primary.URL != "" || secondary.URL != "" {
		t.Errorf("all-opted-out must return zero-value targets, got (%q, %q)", primary.URL, secondary.URL)
	}
	// The header fallback's call shape degrades identically (its caller then
	// renders no button via its own primary.URL != "" guard).
	primary, secondary = chooseCTATargets("", "", interactive, hubs)
	if primary.URL != "" || secondary.URL != "" {
		t.Errorf("header form: all-opted-out must return zero values, got (%q, %q)", primary.URL, secondary.URL)
	}
	// One survivor: primary set, secondary zero — [1] is len-guarded.
	hubs[0].IneligibleAsCTATarget = false
	primary, secondary = chooseCTATargets("", "index", interactive, hubs)
	if primary.URL != "/guides/index.html" || secondary.URL != "" {
		t.Errorf("single survivor: got (%q, %q), want (/guides/index.html, \"\")", primary.URL, secondary.URL)
	}
}
