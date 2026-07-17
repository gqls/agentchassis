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
	p := contentHeroPrompt("Gripper Cycle Time Guide", "How to estimate throughput.")
	for _, want := range []string{"Gripper Cycle Time Guide", "How to estimate throughput.", "no text"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt %q missing %q", p, want)
		}
	}
	// No description → title-only subject, no dangling separator.
	p2 := contentHeroPrompt("Title Only", "  ")
	if strings.Contains(p2, "—") {
		t.Errorf("prompt with empty description carries a dangling separator: %q", p2)
	}
}
