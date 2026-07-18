// FILE: platform/orchestration/actions/discovery_checks/content_image_helpers_test.go

package discovery_checks

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestContentImageItemKey(t *testing.T) {
	if got := contentImageItemKey("gripper-payload-calculator-guide"); got != "content_image:gripper-payload-calculator-guide" {
		t.Fatalf("item key = %q", got)
	}
}

// The spec is the contract with asset-deployer's content_card routing
// (spec.mode) and derive_card_asset's inputs (entity_type/entity_id/
// page_name) — pin every field the other side reads.
func TestContentImageSpecJSON(t *testing.T) {
	s, err := contentImageSpecJSON("content_image_missing", "3fbb0f5e-0000-0000-0000-000000000000", "my-guide")
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(s), &spec); err != nil {
		t.Fatalf("spec is not valid JSON: %v", err)
	}
	want := map[string]string{
		"mode":        "content_card",
		"check":       "content_image_missing",
		"entity_type": "page",
		"entity_id":   "3fbb0f5e-0000-0000-0000-000000000000",
		"page_name":   "my-guide",
		"purpose":     "card",
	}
	for k, v := range want {
		if spec[k] != v {
			t.Errorf("spec[%q] = %v, want %q", k, spec[k], v)
		}
	}
}

func TestContentImageMissingCheckRegistered(t *testing.T) {
	c := &ContentImageMissingCheck{}
	if c.Name() != "content_image_missing" {
		t.Fatalf("check name = %q", c.Name())
	}
}

// contentImageAction drives the whole D13 convergence: generate → derive →
// silent. Every branch is a distinct live behaviour, so pin them all.
func TestContentImageAction(t *testing.T) {
	cases := []struct {
		name string
		row  contentImageRow
		want string
	}{
		{"no image at all → generate",
			contentImageRow{}, "generate"},
		{"no hero but a legacy card (site-fallback era) → still generate",
			contentImageRow{CardID: "c1", CardOriginID: "site1"}, "generate"},
		{"plan hero, no card → derive",
			contentImageRow{PlanHeroID: "p1"}, "derive"},
		{"content hero, no card → derive",
			contentImageRow{ContentHeroID: "h1"}, "derive"},
		{"card stale by origin (cut from superseded source) → derive",
			contentImageRow{ContentHeroID: "h1", CardID: "c1", CardOriginID: "site1"}, "derive"},
		{"plan hero preferred over content hero for staleness",
			contentImageRow{PlanHeroID: "p1", ContentHeroID: "h1", CardID: "c1", CardOriginID: "h1"}, "derive"},
		{"fulfilled: card derived from current content hero → silent",
			contentImageRow{ContentHeroID: "h1", CardID: "c1", CardOriginID: "h1"}, ""},
		{"fulfilled: card derived from current plan hero → silent",
			contentImageRow{PlanHeroID: "p1", CardID: "c1", CardOriginID: "p1"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := contentImageAction(tc.row); got != tc.want {
				t.Fatalf("contentImageAction(%+v) = %q, want %q", tc.row, got, tc.want)
			}
		})
	}
}

func TestContentHeroPrompt(t *testing.T) {
	p := contentHeroPrompt("Gripper Cycle Time Guide", "How to estimate throughput.", "Article header image")
	for _, want := range []string{"Article header image", "Gripper Cycle Time Guide", "How to estimate throughput.", "no text"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt %q missing %q", p, want)
		}
	}
	// No description → title-only subject, no dangling separator.
	p2 := contentHeroPrompt("Title Only", "  ", "Article header image")
	if strings.Contains(p2, "—") {
		t.Errorf("prompt with empty description carries a dangling separator: %q", p2)
	}
	// F3: the surface supplies the noun, so a tool page is not described as
	// an article. Empty noun falls back rather than emitting a bare prompt.
	if p3 := contentHeroPrompt("Payload Calculator", "", "Header image for a web-based tool"); !strings.HasPrefix(p3, "Header image for a web-based tool representing: Payload Calculator") {
		t.Errorf("tool-surface prompt = %q", p3)
	}
	if p4 := contentHeroPrompt("X", "", ""); !strings.HasPrefix(p4, "Article header image") {
		t.Errorf("empty subject noun did not fall back: %q", p4)
	}
}

// F3 (2026-07-18): the surface table is the check's extension point, so its
// invariants are worth pinning — a malformed entry silently changes what the
// fleet spends money generating.
func TestContentImageSurfaces(t *testing.T) {
	seenType := map[string]bool{}
	for _, s := range contentImageSurfaces {
		if s.PageType == "" || s.ConsumerLike == "" || s.EligibilitySQL == "" || s.SubjectNoun == "" {
			t.Errorf("surface %+v has an empty field — every field is load-bearing", s)
		}
		if seenType[s.PageType] {
			t.Errorf("duplicate surface for page_type %q — the sweep would double-emit", s.PageType)
		}
		seenType[s.PageType] = true
		// The eligibility fragments require alias `p` and must be AND-able
		// onto an existing WHERE clause.
		if !strings.Contains(s.EligibilitySQL, "p.") || !strings.Contains(s.EligibilitySQL, "AND") {
			t.Errorf("surface %q eligibility is not an AND-able `p`-aliased fragment: %q", s.PageType, s.EligibilitySQL)
		}
		// A consumer pattern without wildcards would never match a LIKE.
		if !strings.HasPrefix(s.ConsumerLike, "%") || !strings.HasSuffix(s.ConsumerLike, "%") {
			t.Errorf("surface %q ConsumerLike must be wildcard-wrapped for LIKE: %q", s.PageType, s.ConsumerLike)
		}
	}
	if !seenType["blog-post"] || !seenType["tool"] {
		t.Error("expected both the blog-post and tool surfaces to be registered")
	}
}
