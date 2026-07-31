package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Publication — step 2 of the owner's 2026-07-31 share-card ruling.
//
// A round becomes a public record only because the visitor who argued it pressed
// share. Neither handler here can invent one: PublishHandler refuses a round with
// no verdict, and PublicRoundHandler serves only rows whose published_at is set.
// Nothing backfills, so the public record starts empty — which is the honest
// state, since every round stored before today is our own harness traffic.

// PublishHandler handles POST /api/v1/tools/gauntlet/publish.
// Body: {"round_id": "<uuid>"}. Returns {"slug": "...", "path": "..."}.
//
// Idempotent: pressing share twice returns the SAME slug, because the card the
// visitor already holds carries the first URL.
func PublishHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteID := c.GetString("site_id")

		var body struct {
			RoundID string `json:"round_id"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.RoundID == "" {
			httperr.JSONError(c, http.StatusBadRequest, "round_id is required")
			return
		}

		// Parse the uuid HERE rather than letting Postgres reject it. id is a uuid
		// column, so a malformed value raises 22P02 — which is not pgx.ErrNoRows,
		// so it would fall through to a 500 and be logged as an internal failure
		// for what is simply a bad request. (position/defend have the same shape
		// against store.GetRound; not changed here, but worth knowing.)
		if _, err := uuid.Parse(body.RoundID); err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "round_id is not a valid id")
			return
		}

		slug, err := store.PublishRound(c.Request.Context(), pool, siteID, body.RoundID)
		switch {
		case err == nil:
			// The path, not a full URL: the browser knows its own origin, and
			// hardcoding a domain here would be wrong for any other site that
			// adopts the tool.
			c.JSON(http.StatusOK, gin.H{
				"slug": slug,
				"path": "/tools/gauntlet/round.html?r=" + slug,
			})
		case errors.Is(err, store.ErrRoundNotPublishable):
			// 409, not 404: the round exists, it simply has no verdict yet. The
			// front end must not turn this into "your round is gone".
			httperr.JSONError(c, http.StatusConflict, "round has no verdict yet")
		case errors.Is(err, pgx.ErrNoRows):
			httperr.JSONError(c, http.StatusNotFound, "round not found")
		default:
			logInternalFailure("publish", "publish_round", body.RoundID, err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not publish round")
		}
	}
}

// PublicRoundHandler handles GET /api/v1/tools/gauntlet/round/:slug.
//
// Read-only, no LLM call, no write. It serves store.PublicRound, which is a
// separate type from store.Round precisely so that adding a column to the table
// cannot silently widen what is public — client_ip_hash is not in it.
func PublicRoundHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteID := c.GetString("site_id")
		slug := c.Param("slug")

		// store owns both the alphabet and this check, so the generator and the
		// validator cannot drift apart.
		if !store.ValidSlug(slug) {
			httperr.JSONError(c, http.StatusNotFound, "no such round")
			return
		}

		round, err := store.GetPublishedRound(c.Request.Context(), pool, siteID, slug)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Unpublished and non-existent answer identically on purpose:
				// this endpoint must not reveal that a private round exists.
				httperr.JSONError(c, http.StatusNotFound, "no such round")
				return
			}
			logInternalFailure("public_round", "get_published_round", slug, err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not load round")
			return
		}

		c.JSON(http.StatusOK, round)
	}
}
