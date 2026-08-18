// FILE: facts_test.go — the facts provider's contract, especially the
// fallback chain (live → disk cache → refuse), because the failure modes are
// where this either fixes the drift landmine or quietly reintroduces it.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func factsRelayStub(t *testing.T, token string, facts []siteFact, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Facts-Token") != token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_ = json.NewEncoder(w).Encode(factsResponse{Domain: "webdesign.uk", Facts: facts})
	}))
}

func TestRenderedPromptCarriesEveryClaimAndTheFixedFrame(t *testing.T) {
	facts := []siteFact{
		{ID: "price_total", Claim: "The website build costs one thousand two hundred pounds."},
		{ID: "deposit", Claim: "A £75 deposit is non-refundable."},
		{ID: "empty_one", Claim: "   "}, // must be skipped, not rendered as an empty bullet
	}
	prompt := renderSystemPrompt("webdesign.uk", "a service that builds complete websites for small and medium UK businesses", facts)

	for _, want := range []string{
		"- The website build costs one thousand two hundred pounds.",
		"- A £75 deposit is non-refundable.",
		"intake assistant for webdesign.uk", // frame intro survived
		"restraint reads as confidence",     // frame conduct survived
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
	if strings.Contains(prompt, "- \n") || strings.Contains(prompt, "-  ") {
		t.Error("an empty claim rendered as an empty bullet — that reads as permission to improvise")
	}
}

func TestStartupFetchesPersistsAndServes(t *testing.T) {
	facts := []siteFact{{ID: "a", Claim: "Fact A."}}
	srv := factsRelayStub(t, "tok", facts, http.StatusOK)
	defer srv.Close()

	cache := filepath.Join(t.TempDir(), "facts-lastgood.json")
	p, err := newFactsProvider(srv.URL, "tok", "webdesign.uk", "a test site", cache, time.Hour)
	if err != nil {
		t.Fatalf("provider init: %v", err)
	}
	if !strings.Contains(p.SystemPrompt(), "- Fact A.") {
		t.Error("live fetch did not reach the prompt")
	}
	// The successful fetch must be on disk for the next cold start.
	raw, err := os.ReadFile(cache)
	if err != nil {
		t.Fatalf("last-good cache not written: %v", err)
	}
	var cached []siteFact
	if json.Unmarshal(raw, &cached) != nil || len(cached) != 1 || cached[0].Claim != "Fact A." {
		t.Errorf("cache content wrong: %s", raw)
	}
}

func TestRelayDownFallsBackToDiskCache(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "facts-lastgood.json")
	seed := []siteFact{{ID: "cached", Claim: "The cached fact."}}
	raw, _ := json.Marshal(seed)
	if err := os.WriteFile(cache, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	// A relay that answers 500 to everything.
	srv := factsRelayStub(t, "tok", nil, http.StatusInternalServerError)
	defer srv.Close()

	p, err := newFactsProvider(srv.URL, "tok", "webdesign.uk", "a test site", cache, time.Hour)
	if err != nil {
		t.Fatalf("cache fallback should have succeeded: %v", err)
	}
	if !strings.Contains(p.SystemPrompt(), "- The cached fact.") {
		t.Error("cached facts did not reach the prompt")
	}
}

// THE REFUSAL CASE — the design decision under test. Relay down + no cache
// must be a startup error, and specifically must NOT fall back to the
// compiled-in systemPromptFacts, which is exactly the stale copy this
// mechanism retires.
func TestRelayDownAndNoCacheRefusesToStart(t *testing.T) {
	srv := factsRelayStub(t, "tok", nil, http.StatusInternalServerError)
	defer srv.Close()

	_, err := newFactsProvider(srv.URL, "tok", "webdesign.uk", "a test site", filepath.Join(t.TempDir(), "absent.json"), time.Hour)
	if err == nil {
		t.Fatal("no relay + no cache must refuse, not improvise a prompt from somewhere")
	}
}

// Zero facts from a healthy relay is a misconfiguration, not an update — a
// prompt with an empty facts section licenses the model to invent numbers.
func TestZeroFactsIsAnErrorNotAnUpdate(t *testing.T) {
	srv := factsRelayStub(t, "tok", []siteFact{}, http.StatusOK)
	defer srv.Close()

	if _, err := fetchFacts(srv.URL, "tok", "webdesign.uk"); err == nil {
		t.Fatal("zero facts must be an error")
	}
}

func TestWrongTokenSurfacesAs401Error(t *testing.T) {
	srv := factsRelayStub(t, "right-token", []siteFact{{ID: "a", Claim: "x"}}, http.StatusOK)
	defer srv.Close()

	_, err := fetchFacts(srv.URL, "wrong-token", "webdesign.uk")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("want an error naming 401 (its fix differs from a 404's), got %v", err)
	}
}

// SITE IDENTITY IS A PARAMETER (PLAN_2026-08-11 step 5): the intro must name
// the site the instance was configured for, and nothing else.
func TestPromptIntroNamesTheConfiguredSite(t *testing.T) {
	prompt := renderSystemPrompt("noted.co.uk", "a note-taking app", []siteFact{{ID: "a", Claim: "Fact A."}})
	if !strings.Contains(prompt, "intake assistant for noted.co.uk, a note-taking app.") {
		t.Errorf("intro does not carry the configured identity: %q", prompt[:120])
	}
	if strings.Contains(prompt, "webdesign.uk") {
		t.Error("another site's identity leaked into this instance's prompt")
	}
}

// THE CROSS-CHECK: several instances on one box read env files that differ by
// one line. A FACTS_URL copy-pasted from another site's env would have this
// instance state a different business's prices in this site's name, with every
// fetch reporting success — so the relay's own domain field is checked against
// SITE_DOMAIN and a mismatch is a hard error, not an update.
func TestRelayServingAnotherSitesFactsIsRefused(t *testing.T) {
	// The stub always answers Domain: "webdesign.uk".
	srv := factsRelayStub(t, "tok", []siteFact{{ID: "a", Claim: "£149 for a site."}}, http.StatusOK)
	defer srv.Close()

	_, err := fetchFacts(srv.URL, "tok", "noted.co.uk")
	if err == nil || !strings.Contains(err.Error(), "another site") {
		t.Fatalf("a relay serving webdesign.uk facts to the noted.co.uk instance must be refused, got %v", err)
	}
	// Case-insensitive on the domain: the relay lowercases, an operator may not.
	if _, err := fetchFacts(srv.URL, "tok", "WebDesign.UK"); err != nil {
		t.Fatalf("case difference alone must not refuse: %v", err)
	}
}

// The conduct forbids em dashes. Prompt text is read by the model as an example
// of the behaviour it describes, so a rule stated in a sentence that breaks the
// rule is worse than no rule. This is not hypothetical: the conduct string before
// 2026-08-18 banned em dashes in a sentence containing one.
//
// It is a real test because it FAILS on that previous string, not merely on a
// contrived one. Checked when written.
func TestConductDoesNotBreakItsOwnStyleRule(t *testing.T) {
	if !strings.Contains(promptConduct, "no em dashes") {
		t.Fatal("the em dash rule has gone from the conduct; either restore it or delete this test deliberately")
	}
	for _, dash := range []string{"—", "–"} {
		if i := strings.Index(promptConduct, dash); i >= 0 {
			lo := i - 50
			if lo < 0 {
				lo = 0
			}
			hi := i + 50
			if hi > len(promptConduct) {
				hi = len(promptConduct)
			}
			t.Errorf("the conduct bans em dashes and contains one at %d: %q", i, promptConduct[lo:hi])
		}
	}
}

// The brief-builder is the whole point of the 2026-08-18 change and it is one
// string away from being silently reverted by anyone tidying the prompt. Pin the
// load-bearing clauses: the second job, the reason it matters (which is drawn
// from the register's own one_shot_no_approval fact, not invented), the rule that
// keeps it from becoming an interrogation, and the permission to stop.
func TestConductCarriesTheBriefBuilderAndItsRestraints(t *testing.T) {
	for _, want := range []string{
		"help them work out what to ask for", // the second job exists
		"no approval stage and no revisions", // why it matters, per the register
		"Ask ONE thing at a time",            // not an interrogation
		"take it and stop",                   // the visitor may decline
		"Only write it if they say yes",      // the brief is offered, not imposed
		"under 250 words",                    // fits max_tokens 1024; see claude.go
		"a draft for them to check",          // honest about what it is
	} {
		if !strings.Contains(promptConduct, want) {
			t.Errorf("conduct lost the clause %q", want)
		}
	}
	// The old rule and the brief-builder cannot both hold. If someone restores it,
	// the bot stops eliciting and the change is dead without any test going red.
	if strings.Contains(promptConduct, "Do not ask for anything else unless they offer it") {
		t.Error("the pre-2026-08-18 rule is back; it contradicts the brief-builder above")
	}
}
