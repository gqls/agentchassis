package datahelpers

import "testing"

func TestIsNoopHref(t *testing.T) {
	noop := []string{"#", "#!", "javascript:void(0)", "javascript:void(0);", "javascript:;", "javascript:", " # ", "JAVASCRIPT:VOID(0)"}
	for _, h := range noop {
		if !IsNoopHref(h) {
			t.Errorf("IsNoopHref(%q) = false, want true", h)
		}
	}
	// Real destinations, named fragments, and the empty class (owned by
	// phantom_internal_links) must not be flagged.
	real := []string{"", "#section", "/tools/arena/index.html", "https://example.com", "mailto:x@y.z", "#top"}
	for _, h := range real {
		if IsNoopHref(h) {
			t.Errorf("IsNoopHref(%q) = true, want false", h)
		}
	}
}

func TestDeadControlAnchors(t *testing.T) {
	html := `
		<a href="#" class="btn">Enter the Gauntlet</a>
		<a href="#!">Preview</a>
		<a href="#rules">Rules</a>
		<a href="/tools/arena/index.html">Arena</a>
		<a href="javascript:void(0)"><span>Play</span> now</a>`
	dead := DeadControlAnchors(html)
	if len(dead) != 3 {
		t.Fatalf("got %d dead controls, want 3: %+v", len(dead), dead)
	}
	if dead[0].Text != "Enter the Gauntlet" {
		t.Errorf("first dead control text = %q, want %q", dead[0].Text, "Enter the Gauntlet")
	}
	if dead[2].Text != "Play now" {
		t.Errorf("inner-markup text = %q, want %q", dead[2].Text, "Play now")
	}
}
