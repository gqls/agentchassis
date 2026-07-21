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

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

// TestProbeGETStatus exercises the actual probe branch (not just the filter):
// only 405 is the method_mismatch signal, and a transport failure must surface
// as an ERROR — never as a status the caller could mistake for "probed clean".
// A live backend stands these up as: /audience-check 405 (POST-only handler),
// a real page 200, a bogus path 404 (see the live probe in bugs_open/017).
func TestProbeGETStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/audience-check", "/request":
			w.WriteHeader(http.StatusMethodNotAllowed) // 405 — POST-only handler
		case "/subscribe":
			w.WriteHeader(http.StatusBadRequest) // 400 — a GET-answering route, NOT our class
		case "/tools":
			w.WriteHeader(http.StatusOK) // 200 — a real page
		default:
			w.WriteHeader(http.StatusNotFound) // 404 — phantom, not our class
		}
	}))
	defer srv.Close()

	cases := []struct {
		path       string
		wantStatus int
		wantFlag   bool // does the check's 405-only rule flag it?
	}{
		{"/audience-check", http.StatusMethodNotAllowed, true},
		{"/request", http.StatusMethodNotAllowed, true},
		{"/subscribe", http.StatusBadRequest, false},
		{"/tools", http.StatusOK, false},
		{"/nope", http.StatusNotFound, false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			status, err := probeGETStatus(context.Background(), srv.URL+tc.path)
			if err != nil {
				t.Fatalf("probeGETStatus(%s) unexpected error: %v", tc.path, err)
			}
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
			if got := status == http.StatusMethodNotAllowed; got != tc.wantFlag {
				t.Errorf("405-flag = %v, want %v", got, tc.wantFlag)
			}
		})
	}

	// A transport failure must return an error, NOT a swallowed status==0 that
	// the loop would read as "not a 405 => clean". This is the bug_historian
	// objection made a test: the check must not go silently blind on a probe
	// that could not run.
	srv.Close() // server is now down: the GET cannot complete
	if _, err := probeGETStatus(context.Background(), srv.URL+"/audience-check"); err == nil {
		t.Fatal("probeGETStatus against a downed server returned nil error — a failed probe must surface, not read as clean")
	}
}
