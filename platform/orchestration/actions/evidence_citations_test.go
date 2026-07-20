// FILE: platform/orchestration/actions/evidence_citations_test.go
//
// The network half of citation verification, tested offline against a local
// httptest server. The property under test is the failure CLASSIFICATION —
// found vs lost vs unknown — because each routes differently (bump / drift /
// error) and confusing loss with unknown would either raise false alarms on
// every paywall or silently trust facts nobody can check.

package actions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func testCitation(url string) *datahelpers.Citation {
	return &datahelpers.Citation{
		Publisher: "International Gas Union", Title: "World LNG Report 2025",
		URL: url, Quote: "global LNG trade reached 411 million tonnes in 2024",
	}
}

func TestVerifyCitationLiveClassification(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/supports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><p>Against expectations, global LNG trade reached
			411&nbsp;million tonnes in 2024 — a record.</p></body></html>`))
	})
	mux.HandleFunc("/reworded", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><p>Trade volumes were strong in 2024.</p></body></html>`))
	})
	mux.HandleFunc("/paywalled", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "subscribe", http.StatusForbidden)
	})
	mux.HandleFunc("/report.pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.7 global LNG trade reached 411 million tonnes in 2024"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Quote present (with entity + dash presentation differences) → found.
	out := verifyCitationLive(t.Context(), testCitation(srv.URL+"/supports"))
	if !out.Found {
		t.Errorf("supporting page must verify, got %+v", out)
	}

	// 200 but the quote is gone → citation_lost (drift), NOT an error.
	out = verifyCitationLive(t.Context(), testCitation(srv.URL+"/reworded"))
	if out.Found || out.FailClass != "citation_lost" {
		t.Errorf("reworded page must classify citation_lost, got %+v", out)
	}

	// 403 → fetch_error (unknown), never citation_lost: a paywall going up is
	// not evidence the fact is wrong.
	out = verifyCitationLive(t.Context(), testCitation(srv.URL+"/paywalled"))
	if out.Found || out.FailClass != "fetch_error" {
		t.Errorf("paywall must classify fetch_error, got %+v", out)
	}

	// PDF → refused as unsupported content, even though the bytes contain the
	// quote — half-reading a binary as text would fake verification.
	out = verifyCitationLive(t.Context(), testCitation(srv.URL+"/report.pdf"))
	if out.Found || out.FailClass != "fetch_error" || !strings.Contains(out.FailDetail, "unsupported content type") {
		t.Errorf("pdf must be refused as fetch_error/unsupported, got %+v", out)
	}

	// Non-http URL refused before any fetch.
	out = verifyCitationLive(t.Context(), testCitation("file:///etc/passwd"))
	if out.Found || out.FailClass != "fetch_error" || !strings.Contains(out.FailDetail, "must be http(s)") {
		t.Errorf("non-http scheme must be refused, got %+v", out)
	}
}

func TestCitationDateStale(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2026-07-20")
	cases := []struct {
		name                  string
		published, verifiedAt string
		days                  float64
		want                  bool
	}{
		{"no policy, never stale", "2020-01", "2020-01-01", 0, false},
		{"fresh source", "2026-06", "", 400, false},
		{"aged source", "2024-06", "", 400, true},
		{"ages from PUBLICATION even if recently re-verified", "2024-06", "2026-07-19", 400, true},
		{"year-only date parses", "2023", "", 400, true},
		{"unparseable published falls back to verified_at", "mid-2025", "2026-07-01", 400, false},
		{"no usable date at all — cannot age", "n/a", "", 400, false},
	}
	for _, c := range cases {
		if got := citationDateStale(c.published, c.verifiedAt, c.days, now); got != c.want {
			t.Errorf("%s: citationDateStale(%q,%q,%v) = %v, want %v",
				c.name, c.published, c.verifiedAt, c.days, got, c.want)
		}
	}
}

func TestCitationFactID(t *testing.T) {
	a := citationFactID("https://x.org/r", "reached 411 MT")
	// Deterministic, and stable across presentation variants of the same quote
	// — so a re-run cannot double-register one citation.
	if a != citationFactID("https://x.org/r", "reached  411&nbsp;MT") {
		t.Error("id must be stable across quote presentation variants")
	}
	if a == citationFactID("https://x.org/other", "reached 411 MT") {
		t.Error("different urls must yield different ids")
	}
	if !strings.HasPrefix(a, "CIT-") {
		t.Errorf("id shape: %s", a)
	}
}
