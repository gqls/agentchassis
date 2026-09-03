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

// RFC_060 §3d/Q6: the CONC 6.7.23-vs-6.7.17 fixture (a whole page, several
// rules, one genuinely mis-attributable quote), served over httptest rather
// than the live FCA Handbook — offline, deterministic, and it lets a wrong
// attribution be asserted as a FAILURE rather than merely observed once.
const conc67HandbookFixture = `<html><body><p>CONC 6.7 Post contract: business practices</p>
<p>CONC 6.7.17 01/04/2014 R</p>
<p>In CONC 6.7.18 R to CONC 6.7.23 R "refinance" means to extend or vary a high-cost
short-term credit agreement or to enter into a further such agreement.</p>
<p>CONC 6.7.18 01/04/2014 R</p>
<p>A firm must not exercise forbearance in a way that disguises problem debt.</p>
<p>CONC 6.7.23 01/04/2014 R</p>
<p>A firm must not refinance high-cost short-term credit (other than by exercising
forbearance) on more than two occasions.</p>
</body></html>`

func TestVerifyCitationLiveForRule_CorrectAttributionVerifies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(conc67HandbookFixture))
	}))
	defer srv.Close()

	cit := &datahelpers.Citation{URL: srv.URL,
		Quote: "must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions"}
	out := verifyCitationLiveForRule(t.Context(), cit, "CONC 6.7.23")
	if !out.Found {
		t.Fatalf("correctly-attributed citation must verify, got %+v", out)
	}
}

// This is the load-bearing test in this file: the SAME quote, the SAME
// fetch, attributed to the WRONG rule — exactly the live defect this
// mechanism exists to catch. verifyCitationLive (the un-ruled function)
// would pass this, because the quote genuinely IS on the page.
func TestVerifyCitationLiveForRule_WrongAttributionFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(conc67HandbookFixture))
	}))
	defer srv.Close()

	cit := &datahelpers.Citation{URL: srv.URL,
		Quote: "must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions"}

	// Control: the whole-page function DOES pass this — proving the fixture
	// is realistic and the failure below is CitationRuleSpan's doing, not an
	// artefact of a broken fixture.
	whole := verifyCitationLive(t.Context(), cit)
	if !whole.Found {
		t.Fatalf("control failed: whole-page verifyCitationLive should find this quote (it IS on the page); "+
			"got %+v — the fixture itself is wrong, fix it before trusting the rule-scoped result below", whole)
	}

	out := verifyCitationLiveForRule(t.Context(), cit, "CONC 6.7.17")
	if out.Found {
		t.Fatalf("the quote belongs to CONC 6.7.23, not 6.7.17 — this MUST fail, and the whole-page control " +
			"above proves it isn't failing for the wrong reason (a broken fetch, an empty quote, etc.)")
	}
	if out.FailClass != "citation_lost" {
		t.Fatalf("expected FailClass=citation_lost (a human-reviewable finding), got %+v", out)
	}
}

func TestVerifyCitationLiveForRule_EmptyRuleIDMatchesWholePageBehaviour(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(conc67HandbookFixture))
	}))
	defer srv.Close()

	cit := &datahelpers.Citation{URL: srv.URL,
		Quote: "must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions"}
	out := verifyCitationLiveForRule(t.Context(), cit, "")
	if !out.Found {
		t.Fatalf("empty ruleID must fall back to whole-page matching (found on this page), got %+v", out)
	}
}

func TestVerifyCitationLiveForRule_NonChapterPageFallsBackToWholePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><p>Consumer Credit Act 1974 Section 97. The debtor may request a settlement figure.</p></body></html>`))
	}))
	defer srv.Close()

	cit := &datahelpers.Citation{URL: srv.URL, Quote: "the debtor may request a settlement figure"}
	// A legislation.gov.uk-shaped page (no CONC-style headings) with a rule
	// name that doesn't match the FCA pattern at all — must fall through to
	// ordinary whole-page verification rather than reporting a false failure.
	out := verifyCitationLiveForRule(t.Context(), cit, "Consumer Credit Act 1974 s.97")
	if !out.Found {
		t.Fatalf("a non-chapter page must fall back to whole-page matching, got %+v", out)
	}
}

// RFC_060 §3e/Q7. cfChallengeFixture is trimmed from the REAL response
// captured live against maps.org.uk on 2026-09-03 (curl, HTTP 200) — the
// three markers below (title, _cf_chl_opt, challenge-error-text) are the
// actual bytes Cloudflare served, not an invented approximation.
const cfChallengeFixture = `<!DOCTYPE html><html lang="en-US"><head><title>Just a moment...</title>` +
	`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"></head><body>` +
	`<div class="main-wrapper"><noscript><div class="h2"><span id="challenge-error-text">` +
	`Enable JavaScript and cookies to continue</span></div></noscript></div>` +
	`<script>window._cf_chl_opt = {cvId: '3', cZone: 'maps.org.uk'};</script></body></html>`

func TestBotChallengeReasonDetectsTheRealCloudflarePage(t *testing.T) {
	if reason := botChallengeReason(cfChallengeFixture); reason == "" {
		t.Fatalf("the real captured Cloudflare challenge markup was not detected")
	}
}

func TestBotChallengeReasonEachMarkerAloneIsSufficient(t *testing.T) {
	cases := map[string]string{
		"title only":      `<html><head><title>Just a moment...</title></head><body></body></html>`,
		"cf_chl_opt only": `<html><body><script>window._cf_chl_opt = {};</script></body></html>`,
		"noscript only":   `<html><body><noscript><span id="challenge-error-text">x</span></noscript></body></html>`,
	}
	for name, html := range cases {
		if botChallengeReason(html) == "" {
			t.Errorf("%s: marker alone should be sufficient, got no match", name)
		}
	}
}

// The discriminating control: ordinary content — including a page that
// happens to mention "moment" or "challenge" in prose — must NOT match.
func TestBotChallengeReasonCleanPageIsNotFlagged(t *testing.T) {
	ordinary := `<html><head><title>World LNG Report 2025</title></head><body>` +
		`<p>For a moment, prices dipped, but the CONC 5A cost cap remains a challenge for lenders.</p>` +
		`</body></html>`
	if reason := botChallengeReason(ordinary); reason != "" {
		t.Fatalf("ordinary prose containing the words 'moment' and 'challenge' was flagged: %q", reason)
	}
}

// TestFetchCitationDocument_BotChallengeBecomesFetchErrorNotCitationLost is
// the load-bearing, end-to-end test: a real citation, served the real
// captured challenge page at HTTP 200, going through the actual production
// path (verifyCitationLive, not a unit test of the detector in isolation).
// Before this fix the outcome would be citation_lost (drift) — the loanzy
// lane's own finding, "passes a human skim and then classifies as
// citation_lost every day for ever, caused by the host, not the quote."
func TestFetchCitationDocument_BotChallengeBecomesFetchErrorNotCitationLost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK) // Cloudflare serves the challenge at 200 — status alone says nothing
		w.Write([]byte(cfChallengeFixture))
	}))
	defer srv.Close()

	cit := testCitation(srv.URL)
	out := verifyCitationLive(t.Context(), cit)
	if out.Found {
		t.Fatalf("a challenge page must never verify a quote, got %+v", out)
	}
	if out.FailClass != "fetch_error" {
		t.Fatalf("a challenge page must classify fetch_error (unknown), matching the existing 403 treatment — "+
			"NOT citation_lost, which would wrongly read as the FACT having drifted rather than the HOST being "+
			"unreachable unattended. Got %+v", out)
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
