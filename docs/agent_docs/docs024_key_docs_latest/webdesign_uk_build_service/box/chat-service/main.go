package main

// main.go — entrypoint. Fails loudly at startup rather than starting broken:
// no API key, no contact details to fail closed WITH, or a store that can't
// initialize all stop the process before it binds a port. A silently-broken
// intake bot is worse than one that never started (idea.uk/main.go's own
// pattern: log.Fatal on config problems, never limp along).

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := env("PORT", "8081")

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		log.Fatal("ANTHROPIC_API_KEY not set — refusing to start")
	}

	contactEmail := env("CONTACT_EMAIL", "")
	contactPhone := env("CONTACT_PHONE", "")
	if contactEmail == "" && contactPhone == "" {
		log.Fatal("neither CONTACT_EMAIL nor CONTACT_PHONE set — the daily spend " +
			"ceiling and turn cap fail closed TO the contact details, so without " +
			"them there is nothing safe to fail closed to; refusing to start")
	}
	contactLine := "Thanks for your patience. Please reach us directly:"
	if contactEmail != "" {
		contactLine += " " + contactEmail
	}
	if contactPhone != "" {
		contactLine += " or " + contactPhone
	}

	// Engineering defaults, NOT owner-confirmed business figures — unlike the
	// £1,200/14-day/2-rounds facts above, these two are safety valves and can
	// be tuned without a copy-style sign-off. Still worth a second look before
	// go-live: see RUNBOOK "sizing the daily ceiling".
	maxTurns, err := strconv.Atoi(env("MAX_TURNS_PER_CONVERSATION", "20"))
	if err != nil || maxTurns < 1 {
		log.Fatalf("MAX_TURNS_PER_CONVERSATION invalid: %v", env("MAX_TURNS_PER_CONVERSATION", "20"))
	}
	dailyCeiling, err := strconv.ParseFloat(env("DAILY_SPEND_CEILING_USD", "10.00"), 64)
	if err != nil || dailyCeiling <= 0 {
		log.Fatalf("DAILY_SPEND_CEILING_USD invalid: %v", env("DAILY_SPEND_CEILING_USD", "10.00"))
	}

	dataDir := env("DATA_DIR", "/var/lib/webdesign-chat")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		log.Fatalf("cannot create DATA_DIR %s: %v", dataDir, err)
	}
	store, err := NewStore(
		dataDir+"/state.json",
		dataDir+"/requests.jsonl",
		dataDir+"/transcripts.jsonl",
	)
	if err != nil {
		log.Fatalf("store init failed: %v", err)
	}

	// System prompt source. Legacy (FACTS_URL unset): the compiled-in
	// systemPromptFacts, byte-identical behaviour to every build before the
	// facts relay existed. Opted in (FACTS_URL + FACTS_TOKEN set): live facts
	// from the cluster relay, refreshed every 5 minutes, last-good cache in
	// DATA_DIR, and REFUSAL to start when neither is available — see facts.go
	// for why falling back to the compiled copy is deliberately not offered.
	systemPrompt := func() string { return systemPromptFacts }
	if factsURL := os.Getenv("FACTS_URL"); factsURL != "" {
		factsToken := os.Getenv("FACTS_TOKEN")
		if factsToken == "" {
			log.Fatal("FACTS_URL is set but FACTS_TOKEN is not — refusing to start half-configured")
		}
		// Live mode renders the prompt intro from the site's own identity
		// (PLAN_2026-08-11 step 5: one binary, several sites on one box).
		// No default: an instance falling back to another site's identity
		// would introduce itself as a different business and nothing would
		// ever error. SITE_DOMAIN is also cross-checked against the domain
		// the relay says it served — see facts.go.
		siteDomain := os.Getenv("SITE_DOMAIN")
		siteDescription := os.Getenv("SITE_DESCRIPTION")
		if siteDomain == "" || siteDescription == "" {
			log.Fatal("FACTS_URL is set but SITE_DOMAIN/SITE_DESCRIPTION are not — live mode renders " +
				"the prompt intro from them; refusing to start half-configured")
		}
		provider, err := newFactsProvider(factsURL, factsToken, siteDomain, siteDescription,
			dataDir+"/facts-lastgood.json", 5*time.Minute)
		if err != nil {
			log.Fatalf("facts provider init failed (relay unreachable and no last-good cache): %v", err)
		}
		systemPrompt = provider.SystemPrompt
		log.Printf("facts: live mode, site=%s, relay=%s", siteDomain, factsURL)
	}

	cs := &chatServer{
		store:           store,
		ipLimiter:       newChatIPLimiter(),
		maxTurns:        maxTurns,
		dailyCeilingUSD: dailyCeiling,
		contactLine:     contactLine,
		systemPrompt:    systemPrompt,
	}
	hs := &healthServer{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat", cs.handleChat)
	mux.HandleFunc("/health", hs.handleHealth)

	// BIND_ADDR lets an instance bind loopback only — the noted.co.uk nginx
	// config on this same box names the historical *:8081 bind as the pattern
	// NOT to copy (ufw as the only control). Default keeps the historical
	// all-interfaces bind so a binary swap alone changes nothing for the
	// running webdesign instance; new instances set BIND_ADDR=127.0.0.1:<port>.
	addr := env("BIND_ADDR", ":"+port)
	log.Printf("sitechat on %s (max_turns=%d, daily_ceiling=$%.2f)", addr, maxTurns, dailyCeiling)
	log.Fatal(http.ListenAndServe(addr, mux))
}
