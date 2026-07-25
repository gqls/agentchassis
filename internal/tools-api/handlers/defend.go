package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type defendRequest struct {
	RoundID     string `json:"round_id"`
	DefenceText string `json:"defence_text"`
}

type defendAIResponse struct {
	Verdict string `json:"verdict"`
	Reasons string `json:"reasons"`
}

// DefendHandler handles POST /defend.
// It loads the round, calls the AI for an even-handed verdict and reasons,
// validates the response is non-empty, persists it, and returns 200 JSON.
// If the AI response is malformed or has empty fields, it returns 502 and
// does NOT persist (fail-loud, no silent-blank persistence).
func DefendHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req defendRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(req.RoundID) == "" || strings.TrimSpace(req.DefenceText) == "" {
			httperr.JSONError(c, http.StatusBadRequest, "round_id and defence_text are required")
			return
		}

		ctx := context.Background()

		round, err := store.GetRound(ctx, pool, req.RoundID)
		if err != nil {
			// Only a genuinely missing row is a 404 — any other failure
			// (DB outage, scan defect) must surface as a 500, not be
			// disguised as the caller's mistake.
			if errors.Is(err, pgx.ErrNoRows) {
				httperr.JSONError(c, http.StatusNotFound, "round not found")
			} else {
				httperr.JSONError(c, http.StatusInternalServerError, "round lookup failed")
			}
			return
		}

		prompt := `You are an impartial judge in the Gauntlet debate system.

The debate provocation is:
` + string(round.Provocation) + `

The user stated their position:
` + round.PositionText + `

The opponent's counter-position and challenge:
` + string(round.Counter) + `

The user's defence:
` + req.DefenceText + `

Your task: judge this debate round honestly and even-handedly. A strong defence wins — do not favour the opponent by default. Assess the quality of the arguments on their merits.

Reply with ONLY a JSON object and no prose wrapper, markdown, or explanation. The object must have exactly these two fields:
{"verdict":"your verdict here (e.g. user wins, opponent wins, draw)","reasons":"your reasoning here"}`

		client, err := aiservice.NewAnthropicClient(ctx, map[string]interface{}{"model": cfg.Model})
		if err != nil {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet judge unavailable")
			return
		}

		text, err := client.GenerateText(ctx, prompt, map[string]interface{}{})
		if err != nil {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet judge unavailable")
			return
		}

		var aiResp defendAIResponse
		if err := json.Unmarshal([]byte(text), &aiResp); err != nil {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet judge response was invalid")
			return
		}
		if strings.TrimSpace(aiResp.Verdict) == "" || strings.TrimSpace(aiResp.Reasons) == "" {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet judge response was invalid")
			return
		}

		verdictJSON, err := json.Marshal(aiResp)
		if err != nil {
			httperr.JSONError(c, http.StatusInternalServerError, "internal error")
			return
		}

		if err := store.UpdateRoundVerdict(ctx, pool, req.RoundID, req.DefenceText, json.RawMessage(verdictJSON)); err != nil {
			httperr.JSONError(c, http.StatusInternalServerError, "failed to persist verdict")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"verdict": aiResp.Verdict,
			"reasons": aiResp.Reasons,
		})
	}
}
