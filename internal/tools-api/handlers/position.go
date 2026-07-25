package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/jackc/pgx/v5/pgxpool"
)

type positionRequest struct {
	RoundID      string `json:"round_id"`
	PositionText string `json:"position_text"`
}

type positionAIResponse struct {
	CounterPosition string `json:"counter_position"`
	Challenge       string `json:"challenge"`
}

// PositionHandler handles POST /position.
// It loads the round, calls the AI for a counter_position and challenge,
// validates the response is non-empty, persists it, and returns 200 JSON.
// If the AI response is malformed or has empty fields, it returns 502 and
// does NOT persist (fail-loud, no silent-blank persistence).
func PositionHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req positionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(req.RoundID) == "" || strings.TrimSpace(req.PositionText) == "" {
			httperr.JSONError(c, http.StatusBadRequest, "round_id and position_text are required")
			return
		}

		ctx := context.Background()

		round, err := store.GetRound(ctx, pool, req.RoundID)
		if err != nil {
			httperr.JSONError(c, http.StatusNotFound, "round not found")
			return
		}

		prompt := `You are a debate opponent in the Gauntlet debate system.

The debate provocation is:
` + string(round.Provocation) + `

The user has stated their position:
` + req.PositionText + `

Your task: argue a genuine opposing position and pose a direct challenge to the user's argument.

Reply with ONLY a JSON object and no prose wrapper, markdown, or explanation. The object must have exactly these two fields:
{"counter_position":"your opposing argument here","challenge":"your direct challenge question here"}`

		client, err := aiservice.NewAnthropicClient(ctx, map[string]interface{}{"model": cfg.Model})
		if err != nil {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet opponent unavailable")
			return
		}

		text, err := client.GenerateText(ctx, prompt, map[string]interface{}{})
		if err != nil {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet opponent unavailable")
			return
		}

		var aiResp positionAIResponse
		if err := json.Unmarshal([]byte(text), &aiResp); err != nil {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet opponent response was invalid")
			return
		}
		if strings.TrimSpace(aiResp.CounterPosition) == "" || strings.TrimSpace(aiResp.Challenge) == "" {
			httperr.JSONError(c, http.StatusBadGateway, "gauntlet opponent response was invalid")
			return
		}

		counterJSON, err := json.Marshal(aiResp)
		if err != nil {
			httperr.JSONError(c, http.StatusInternalServerError, "internal error")
			return
		}

		if err := store.UpdateRoundPosition(ctx, pool, req.RoundID, req.PositionText, json.RawMessage(counterJSON)); err != nil {
			httperr.JSONError(c, http.StatusInternalServerError, "failed to persist position")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"counter_position": aiResp.CounterPosition,
			"challenge":        aiResp.Challenge,
		})
	}
}
