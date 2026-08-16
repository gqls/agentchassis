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
