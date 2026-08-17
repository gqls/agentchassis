// FILE: facts.go — live site facts from the cluster's facts relay.
//
// OPT-IN: FACTS_URL unset means none of this runs and the bot behaves exactly
// as before (compiled-in systemPromptFacts). Set FACTS_URL + FACTS_TOKEN and
// the system prompt's facts section is rendered from the site's own
// evidence_base rows, fetched over WireGuard from the in-cluster relay
// (core-manager /api/v1/site-facts/:domain) — the fix for the drift landmine
// where the £75 deposit was live in the database while this binary's
// compiled-in facts still promised a full refund.
//
// THE FALLBACK CHAIN IS DELIBERATELY NOT "fall back to the compiled-in copy".
// Once the operator has opted into live facts, the compiled string is exactly
// the stale copy this mechanism exists to retire — silently reviving it on a
// fetch failure would reintroduce the drift in the one situation (relay down,
// nobody watching) where it would live longest. Instead:
//
//	live fetch  →  last-good copy persisted on disk  →  REFUSE TO START.
//
// Refusing follows main.go's own stated pattern ("a silently-broken intake
// bot is worse than one that never started"): systemd restarts us, each
// restart retries the fetch, and the failure is loud in journalctl. A running
// service that loses the relay AFTER startup keeps serving its last-good
// facts and retries in the background — facts change on owner-decision
// timescales (days), so hours of staleness during an outage is acceptable;
// silent unbounded staleness is not.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// siteFact is one evidence_base fact as the relay serves it. Claim is the
// owner-attested sentence — the bullets the bot may state are these, verbatim,
// which is MORE faithful than the hand-paraphrased bullets they replace.
type siteFact struct {
	ID    string `json:"id"`
	Claim string `json:"claim"`
}

type factsResponse struct {
	Domain string     `json:"domain"`
	Facts  []siteFact `json:"facts"`
}

// promptFrame is everything around the facts: behaviour instructions, not
// facts, so it stays compiled-in and owner-reviewed like all bot copy. The
// facts bullets are interpolated between Intro and Conduct.
//
// The intro's SITE IDENTITY is a parameter (SITE_DOMAIN + SITE_DESCRIPTION,
// PLAN_2026-08-11 step 5: one binary, several sites on one box). It is NOT
// defaulted: a second site's instance falling back to another site's identity
// is the worst failure this file can produce — the bot would introduce itself
// as a different business and no error would ever say so. main.go refuses to
// start live mode without both.
func renderPromptIntro(domain, description string) string {
	return "You are the intake assistant for " + domain + ", " + description + `.

Facts you may state, and the ONLY facts you may state as numbers or commitments — never invent, round, or approximate anything beyond these:`
}

const promptConduct = `Your job: have a short, plain conversation. Ask what the site is for and what domain they'd want it on. Do not assume the visitor runs a business: sites here are built for anyone who wants one, and asking "what business are you in?" of someone building a personal, community or project site reads as not listening. Do not ask for anything else unless they offer it. Do not invent services, features, or numbers beyond the facts above. Do not promise anything about timing, price, or process that isn't stated above. If asked something you don't know, say so plainly and point at the contact details. Write in plain, direct British English — short sentences, no agency-marketing language, no em dashes. This is a first conversation, not a sales pitch: restraint reads as confidence here.`

func renderSystemPrompt(domain, description string, facts []siteFact) string {
	var b strings.Builder
	b.WriteString(renderPromptIntro(domain, description))
	b.WriteString("\n")
	for _, f := range facts {
		claim := strings.TrimSpace(f.Claim)
		if claim == "" {
			continue // an empty claim renders an empty bullet the model reads as permission
		}
		b.WriteString("- ")
		b.WriteString(claim)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(promptConduct)
	return b.String()
}

// factsProvider hands the current system prompt to handleChat and refreshes
// it in the background. Zero-value is never used — construct via newFactsProvider.
type factsProvider struct {
	mu     sync.RWMutex
	prompt string

	// domain and description parameterise the prompt intro; domain is ALSO
	// the cross-check against what the relay says it served (see fetchFacts).
	domain      string
	description string
}

func (p *factsProvider) SystemPrompt() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.prompt
}

func (p *factsProvider) set(facts []siteFact) {
	rendered := renderSystemPrompt(p.domain, p.description, facts)
	p.mu.Lock()
	p.prompt = rendered
	p.mu.Unlock()
}

var factsHTTP = &http.Client{Timeout: 10 * time.Second}

// fetchFacts GETs the relay. Non-200 is an error carrying the status so a 401
// (bad/missing token) is distinguishable from a 404 (domain not registered)
// in journalctl — those have different fixes and identical symptoms otherwise.
//
// expectDomain is the instance's own SITE_DOMAIN, checked against what the
// relay SAYS it served: with several instances on one box reading env files
// that differ by one line, a FACTS_URL copy-pasted from another site's env
// would otherwise have this instance state a different business's prices in
// this site's name, with every fetch reporting success.
func fetchFacts(url, token, expectDomain string) ([]siteFact, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Facts-Token", token)
	resp, err := factsHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("facts relay returned %d", resp.StatusCode)
	}
	var fr factsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return nil, fmt.Errorf("facts relay response not decodable: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(fr.Domain), strings.TrimSpace(expectDomain)) {
		return nil, fmt.Errorf("facts relay served domain %q but this instance is %q — refusing another site's facts", fr.Domain, expectDomain)
	}
	if len(fr.Facts) == 0 {
		// Zero facts is indistinguishable from a misconfigured relay, and a
		// prompt with an empty facts section licenses the model to improvise
		// — the one failure mode this whole file exists to prevent.
		return nil, fmt.Errorf("facts relay returned zero facts")
	}
	return fr.Facts, nil
}

// newFactsProvider fetches once (falling back to the on-disk last-good copy),
// persists successful fetches, and starts the background refresher. An error
// here means neither the relay nor a cached copy could supply facts — the
// caller should treat that as fatal, per the header comment.
func newFactsProvider(url, token, domain, description, cachePath string, refreshEvery time.Duration) (*factsProvider, error) {
	p := &factsProvider{domain: domain, description: description}

	facts, err := fetchFacts(url, token, domain)
	if err != nil {
		log.Printf("facts: startup fetch failed (%v), trying last-good cache %s", err, cachePath)
		cached, cacheErr := os.ReadFile(cachePath)
		if cacheErr != nil {
			return nil, fmt.Errorf("facts fetch failed (%v) and no last-good cache (%v)", err, cacheErr)
		}
		if jsonErr := json.Unmarshal(cached, &facts); jsonErr != nil || len(facts) == 0 {
			return nil, fmt.Errorf("facts fetch failed (%v) and cache unreadable (%v)", err, jsonErr)
		}
		log.Printf("facts: serving %d facts from last-good cache", len(facts))
	} else {
		persistFacts(cachePath, facts)
		log.Printf("facts: fetched %d facts from relay", len(facts))
	}
	p.set(facts)

	go func() {
		ticker := time.NewTicker(refreshEvery)
		defer ticker.Stop()
		for range ticker.C {
			fresh, err := fetchFacts(url, token, domain)
			if err != nil {
				// Keep last-good; loud enough to find, quiet enough not to
				// page anyone over a transient blip.
				log.Printf("facts: refresh failed, keeping last-good: %v", err)
				continue
			}
			p.set(fresh)
			persistFacts(cachePath, fresh)
		}
	}()

	return p, nil
}

// persistFacts writes atomically (tmp+rename, the store.go pattern) so a
// crash mid-write can never leave a truncated cache that then poisons the
// next startup's fallback path.
func persistFacts(path string, facts []siteFact) {
	raw, err := json.Marshal(facts)
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		log.Printf("facts: cache write failed: %v", err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		log.Printf("facts: cache rename failed: %v", err)
	}
}
