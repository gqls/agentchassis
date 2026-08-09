package main

// health.go — GET /health. The nginx config comment is explicit: this must
// exercise nginx AND the engine behind it, not be a static 200 — a static
// pass would be a lie-shaped pass (check_backend_unreachable's contract).
// So this handler proves the ENGINE is actually functional: the state file
// is writable (a real disk round-trip, not just "the process is scheduled"),
// and the API key is present (existence, not validity — a live Anthropic
// call on every health poll would spend money on every check interval,
// which is not what a liveness probe is for).

import (
	"encoding/json"
	"net/http"
	"os"
)

type healthServer struct {
	store *Store
}

func (hs *healthServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	problems := []string{}

	// Real disk round-trip through the same Store the chat handler uses —
	// catches a full disk, a permissions problem, or a wedged mutex.
	if _, err := hs.store.GetOrCreateConversation("__health_check__", "127.0.0.1"); err != nil {
		problems = append(problems, "store: "+err.Error())
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		problems = append(problems, "ANTHROPIC_API_KEY not set")
	}

	w.Header().Set("Content-Type", "application/json")
	if len(problems) > 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "unhealthy", "problems": problems})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}
