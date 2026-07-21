// FILE: platform/orchestration/actions/discovery_checks/check_backend_entry_orphaned_test.go
//
// Pins handlerRouteCandidate — the pure filter that decides which <a href> is
// worth a live GET probe in backend_entry_orphaned. The probe itself (405 => a
// GET link to a POST-only handler) is exercised against the live site, like
// check_backend_unreachable; this table pins the classification so a future edit
// to the filter cannot silently widen the probe set or drop the 017 case.
//
// The load-bearing cases:
//   - /audience-check MUST be a candidate (the exact bugs_open/017 symptom).
//   - a .html page or an asset MUST NOT be — they cannot 405, and probing them
//     is the cost this filter exists to avoid.
//   - a #fragment or ?query MUST be stripped from the request path but MUST NOT
//     disqualify an otherwise-handler-like route.

package discovery_checks

import "testing"

func TestHandlerRouteCandidate(t *testing.T) {
	tests := []struct {
		name     string
		href     string
		wantPath string
		wantOK   bool
	}{
		// The case the check was built for.
		{"extensionless handler route", "/audience-check", "/audience-check", true},
		{"another handler route", "/request", "/request", true},
		{"subscribe handler", "/subscribe", "/subscribe", true},

		// Extension => a page or asset, cannot be a POST-only handler.
		{".html page", "/tools.html", "", false},
		{"nested index.html", "/news/index.html", "", false},
		{"asset by extension", "/logo.png", "", false},
		{"asset by /assets/ prefix", "/assets/images/logo.jpg", "", false},

		// Non-page scopes are excluded by ClassifyLinkScope.
		{"empty href", "", "", false},
		{"bare anchor", "#", "", false},
		{"named fragment on same page", "#request", "", false},
		{"external absolute", "https://example.com/audience-check", "", false},
		{"protocol-relative external", "//example.com/x", "", false},
		{"mailto", "mailto:hello@idea.uk", "", false},

		// The homepage is not a handler route.
		{"root", "/", "", false},

		// Fragment / query are stripped but must not disqualify.
		{"handler route with fragment", "/audience-check#request", "/audience-check", true},
		{"handler route with query", "/audience-check?ref=nav", "/audience-check", true},

		// A .html target with a fragment stays excluded (last segment has a dot).
		{".html with fragment", "/tools.html#audience-check", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotOK := handlerRouteCandidate(tc.href)
			if gotOK != tc.wantOK {
				t.Fatalf("handlerRouteCandidate(%q) ok = %v, want %v", tc.href, gotOK, tc.wantOK)
			}
			if gotPath != tc.wantPath {
				t.Errorf("handlerRouteCandidate(%q) path = %q, want %q", tc.href, gotPath, tc.wantPath)
			}
		})
	}
}
