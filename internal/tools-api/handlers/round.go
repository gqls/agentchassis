package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// provocCache holds a fetched today provocation for a short TTL.
type provocCache struct {
	value     json.RawMessage
	fetchedAt time.Time
}

const provocTTL = 5 * time.Minute

var (
	provocMu    sync.Mutex
	provocStore = map[string]provocCache{}
)

// hashIP returns a short hex digest of the raw IP string so that no
// plaintext address is stored in the database.
func hashIP(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", sum[:16])
}

// FetchProvocation retrieves the 'today' object from
// https://{domain}/data/provocations.json. If the key is absent or its
// value is null/empty the function returns a non-nil error — fail loud
// per bug_historian pattern #7 rather than returning a blank provocation.
// Successful results are cached in-memory per domain for provocTTL.
func FetchProvocation(domain string) (json.RawMessage, error) {
	provocMu.Lock()
	if cached, ok := provocStore[domain]; ok && time.Since(cached.fetchedAt) < provocTTL {
		provocMu.Unlock()
		return cached.value, nil
	}
	provocMu.Unlock()

	url := fmt.Sprintf("https://%s/data/provocations.json", domain)
	resp, err := http.Get(url) //nolint:gosec // domain is from verified sites table
	if err != nil {
		return nil, fmt.Errorf("fetch provocations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch provocations: upstream status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fetch provocations: read body: %w", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("fetch provocations: unmarshal: %w", err)
	}

	today, ok := payload["today"]
	if !ok {
		return nil, fmt.Errorf("fetch provocations: missing 'today' key in response")
	}
	// Treat JSON null or a zero-length value as absent.
	if len(today) == 0 || string(today) == "null" {
		return nil, fmt.Errorf("fetch provocations: 'today' key is null or empty")
	}

	provocMu.Lock()
	provocStore[domain] = provocCache{value: today, fetchedAt: time.Now()}
	provocMu.Unlock()

	return today, nil
}

// RoundHandler handles POST /api/v1/tools/gauntlet/round.
// It sources the provocation server-side (never client-supplied), persists
// a gauntlet_rounds row, and returns {round_id, provocation}.
func RoundHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.GetString("site_domain")
		siteID := c.GetString("site_id")

		prov, err := FetchProvocation(domain)
		if err != nil {
			// 503, not 502 — Cloudflare REPLACES an origin 502's body with its
			// own HTML error page, so the JSON error shape never reaches the
			// browser and the front end cannot tell this apart from a network
			// failure. Commit b498df16b made exactly this change for /position
			// and /defend; /round was missed and kept the 502 until 2026-07-27.
			// (bugs_open/083, taken on by gauntlet_dead_cta.)
			logAIFailure("round", "fetch_provocation", "", err)
			httperr.JSONError(c, http.StatusServiceUnavailable, "provocation unavailable")
			return
		}

		ipHash := hashIP(c.ClientIP())

		round, err := store.CreateRound(c.Request.Context(), pool, siteID, prov, ipHash)
		if err != nil {
			logInternalFailure("round", "create_round", "", err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not create round")
			return
		}

		c.JSON(200, gin.H{
			"round_id":    round.ID,
			"provocation": prov,
		})
	}
}
