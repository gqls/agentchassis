// FILE: internal/core-manager/handlers/sitefacts.go
//
// Read-only site-facts relay: serves one site's evidence_base FACTS as JSON,
// keyed by domain, so a box-hosted service (the webdesign.uk chat intake, and
// any future requires-backend tool) can hold live DB-backed facts instead of
// compiling them in. This is the seam that closes the systemPromptFacts drift
// landmine (webdesign_uk_build_service NOTES 2026-08-10: the £75 deposit was
// live in evidence_base while the deployed bot's compiled-in facts still said
// "full refund" — two sources of truth, one of them stale, no error anywhere).
//
// DELIBERATE LIMITS, each one load-bearing:
//
//   - FACTS ONLY, never writer_block. The facts array is attested, customer-
//     visible content (price, deposit, timescales — all already published on
//     the site itself). writer_block is internal prompt text for the site's
//     content writer; serving it here would leak authoring instructions to
//     whatever holds the token, and nothing consuming this endpoint needs it.
//
//   - AUTH IS A STATIC HEADER TOKEN (X-Facts-Token), the agentbootstrap.go
//     pattern, not a JWT: the caller is a headless systemd service on a VM,
//     not a user. The endpoint is mounted OUTSIDE the AuthMiddleware group for
//     the same reason /api/v1/agents/bootstrap is. Empty env = endpoint
//     refuses everything (fail closed), so an undeployed secret cannot become
//     an open endpoint.
//
//   - NO PUBLIC EXPOSURE. This handler is intended to be reached over the
//     cluster's existing WireGuard transport (the VM joins as a peer) —
//     ClusterIP only, no Ingress resource, no tunnel. If someone later fronts
//     core-manager with an ingress, this endpoint's token is the only thing
//     between the internet and the facts of every site; that is why the token
//     is required even though today's network path already restricts callers.
//
//   - READ-ONLY BY CONSTRUCTION: one SELECT, no write path, no side effects
//     beyond a log line. There is nothing to lock, nothing to race.
package handlers

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SiteFactsHandler serves evidence_base facts for box-hosted consumers.
type SiteFactsHandler struct {
	clientsDB *sql.DB
	logger    *zap.Logger
}

// NewSiteFactsHandler creates the handler.
func NewSiteFactsHandler(clientsDB *sql.DB, logger *zap.Logger) *SiteFactsHandler {
	return &SiteFactsHandler{clientsDB: clientsDB, logger: logger}
}

// HandleGetSiteFacts is GET /api/v1/site-facts/:domain.
//
// 200: {"domain": ..., "site_id": ..., "facts": [...], "verified_at_max": ...}
// 401: missing/wrong token, or no token configured server-side (fail closed)
// 404: unknown domain, or the site has no current evidence_base row — the
//
//	two cases are deliberately indistinguishable to the caller, so the
//	endpoint cannot be used to enumerate which domains exist.
func (h *SiteFactsHandler) HandleGetSiteFacts(c *gin.Context) {
	expected := os.Getenv("SITE_FACTS_TOKEN")
	if expected == "" {
		// No token deployed = nobody gets in, rather than everybody. An
		// operator who forgot the env var sees 401s in the caller's logs and
		// this warning in ours — both point at the same missing secret.
		h.logger.Warn("site-facts request refused: SITE_FACTS_TOKEN not configured")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	got := c.GetHeader("X-Facts-Token")
	// Constant-time compare: this token is long-lived and the endpoint is
	// cheap to hammer, which is exactly the shape timing attacks like.
	if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
		h.logger.Warn("site-facts request refused: bad token",
			zap.String("domain", c.Param("domain")))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	domain := strings.ToLower(strings.TrimSpace(c.Param("domain")))
	if domain == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var siteID string
	var factsJSON []byte
	err := h.clientsDB.QueryRowContext(c.Request.Context(), `
		SELECT s.id::text, ss.data->'facts'
		FROM sites s
		JOIN site_specs ss ON ss.site_id = s.id
		WHERE lower(s.domain) = $1
		  AND ss.aspect = 'evidence_base'
		  AND ss.is_current = true
		  AND ss.data ? 'facts'
	`, domain).Scan(&siteID, &factsJSON)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		h.logger.Error("site-facts query failed", zap.String("domain", domain), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Re-marshal through RawMessage rather than string-splicing, so the
	// response is valid JSON or a 500 — never a truncated fragment.
	var facts json.RawMessage = factsJSON
	c.JSON(http.StatusOK, gin.H{
		"domain":  domain,
		"site_id": siteID,
		"facts":   facts,
	})
}
