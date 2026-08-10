package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/namecheck"
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

		// RFC_020 §5.2 — refuse to make PUBLIC a round that makes a checkable
		// allegation about an apparently-named third party.
		//
		// It runs here, before PublishRound, because this is the only place a round
		// stops being one person's private argument and becomes a permanent public
		// record. Playing, scoring and reading your own round are all untouched:
		// the visitor keeps everything except the share link.
		//
		// BOTH HALVES ARE CHECKED, and the model's half is not trusted for being
		// ours. The verdict and counter are the SERVICE's text (RFC_020 §1.4), which
		// makes them the part we are most clearly the author of, not the least.
		//
		// FAILS CLOSED: if the round cannot be read, it is not published. A failure
		// to publish costs one share; publishing unchecked is the incident.
		round, rerr := store.GetRound(c.Request.Context(), pool, body.RoundID)
		if rerr != nil {
			if errors.Is(rerr, pgx.ErrNoRows) {
				httperr.JSONError(c, http.StatusNotFound, "round not found")
				return
			}
			logInternalFailure("publish", "get_round_for_namecheck", body.RoundID, rerr)
			httperr.JSONError(c, http.StatusInternalServerError, "could not publish round")
			return
		}
		// GetRound is NOT site-scoped (PublishRound is). Without this, a round_id
		// belonging to another site would reach the check and could answer 422
		// instead of 404 — which distinguishes "exists on another site AND contains
		// an allegation" from every other outcome, i.e. an oracle this handler did
		// not have before the check was added. Answer exactly what PublishRound
		// would have.
		if round.SiteID != siteID {
			httperr.JSONError(c, http.StatusNotFound, "round not found")
			return
		}
		if findings := namecheck.ScanAll(
			round.PositionText, round.DefenceText,
			string(round.Counter), string(round.Verdict),
		); len(findings) > 0 {
			// The reasons are logged, not returned. A caller who could see WHICH
			// term tripped it could tune prose against the check until it passed,
			// which is the one way to make this worse than nothing.
			logPublishRefusal(body.RoundID, findings)
			httperr.JSONError(c, http.StatusUnprocessableEntity,
				"this round cannot be shared publicly because it appears to make a "+
					"factual claim about a named person or business. Your round is "+
					"unaffected and you can still read it.")
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
		// bugs_open/232: the only per-round URL in existence, serving raw
		// visitor UGC + an AI verdict. Set before any body write so it
		// covers every outcome below, including the 404/error paths.
		// Defence in depth alongside the round.html page's own noindex meta
		// (platform/orchestration/actions/rerender_single_page_action.go) --
		// this stops the JSON endpoint itself being indexed; it does not
		// and cannot protect the HTML page, which is a different URL.
		c.Header("X-Robots-Tag", "noindex, nofollow")

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
