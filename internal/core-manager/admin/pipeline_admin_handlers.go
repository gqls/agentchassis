// FILE: internal/core-manager/admin/pipeline_admin_handlers.go
//
// Admin endpoints for managing data pipelines (scheduled_tasks):
//   - List all pipelines with computed state and progress
//   - Enable/disable pipelines
//   - Force-trigger a pipeline run
//   - Get pipeline-specific stats (CH enrichment, verification, etc.)

package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PipelineAdminHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPipelineAdminHandlers(db *sql.DB, logger *zap.Logger) *PipelineAdminHandlers {
	return &PipelineAdminHandlers{db: db, logger: logger}
}

// ============================================================================
// GET /admin/pipelines
// ============================================================================
// Returns all scheduled tasks with computed state (idle/in_flight/timed_out),
// grouped by concurrency_group.

func (h *PipelineAdminHandlers) HandleListPipelines(c *gin.Context) {
	ctx := c.Request.Context()

	// Optional filter
	group := c.Query("group")

	query := `
		SELECT
			st.id, st.name, st.target_agent_type, st.target_topic,
			st.enabled, st.interval_seconds, st.concurrency_group,
			st.max_concurrent, st.timeout_seconds,
			st.last_triggered_at, st.last_completed_at,
			st.fire_message,
			CASE
				WHEN st.last_triggered_at IS NULL THEN 'never_run'
				WHEN st.last_completed_at >= st.last_triggered_at THEN 'idle'
				WHEN st.last_triggered_at + (st.timeout_seconds || ' seconds')::interval <= NOW() THEN 'timed_out'
				ELSE 'in_flight'
			END AS state,
			CASE
				WHEN st.last_triggered_at IS NULL THEN NULL
				ELSE EXTRACT(EPOCH FROM (COALESCE(st.last_completed_at, NOW()) - st.last_triggered_at))::int
			END AS last_duration_seconds
		FROM scheduled_tasks st
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if group != "" {
		query += fmt.Sprintf(" AND st.concurrency_group = $%d", argIdx)
		args = append(args, group)
		argIdx++
	}

	query += " ORDER BY st.concurrency_group NULLS LAST, st.name"

	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type pipeline struct {
		ID                  string  `json:"id"`
		Name                string  `json:"name"`
		TargetAgentType     string  `json:"target_agent_type"`
		TargetTopic         string  `json:"target_topic"`
		Enabled             bool    `json:"enabled"`
		IntervalSeconds     int     `json:"interval_seconds"`
		ConcurrencyGroup    *string `json:"concurrency_group"`
		MaxConcurrent       int     `json:"max_concurrent"`
		TimeoutSeconds      int     `json:"timeout_seconds"`
		LastTriggeredAt     *string `json:"last_triggered_at"`
		LastCompletedAt     *string `json:"last_completed_at"`
		FireMessage         bool    `json:"fire_message"`
		State               string  `json:"state"`
		LastDurationSeconds *int    `json:"last_duration_seconds"`
	}

	var pipelines []pipeline
	for rows.Next() {
		var p pipeline
		var lastTriggered, lastCompleted sql.NullTime
		var concGroup sql.NullString
		var lastDuration sql.NullInt64

		if err := rows.Scan(
			&p.ID, &p.Name, &p.TargetAgentType, &p.TargetTopic,
			&p.Enabled, &p.IntervalSeconds, &concGroup,
			&p.MaxConcurrent, &p.TimeoutSeconds,
			&lastTriggered, &lastCompleted,
			&p.FireMessage, &p.State, &lastDuration,
		); err != nil {
			h.logger.Warn("Failed to scan pipeline", zap.Error(err))
			continue
		}

		if concGroup.Valid {
			p.ConcurrencyGroup = &concGroup.String
		}
		if lastTriggered.Valid {
			s := lastTriggered.Time.Format("2006-01-02T15:04:05Z")
			p.LastTriggeredAt = &s
		}
		if lastCompleted.Valid {
			s := lastCompleted.Time.Format("2006-01-02T15:04:05Z")
			p.LastCompletedAt = &s
		}
		if lastDuration.Valid {
			d := int(lastDuration.Int64)
			p.LastDurationSeconds = &d
		}

		pipelines = append(pipelines, p)
	}

	if pipelines == nil {
		pipelines = []pipeline{}
	}

	c.JSON(http.StatusOK, gin.H{
		"pipelines": pipelines,
		"count":     len(pipelines),
	})
}

// ============================================================================
// PATCH /admin/pipelines/:name
// ============================================================================
// Toggle enabled, update interval_seconds.

func (h *PipelineAdminHandlers) HandleUpdatePipeline(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	var body struct {
		Enabled         *bool `json:"enabled"`
		IntervalSeconds *int  `json:"interval_seconds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic UPDATE
	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if body.Enabled != nil {
		sets = append(sets, fmt.Sprintf("enabled = $%d", argIdx))
		args = append(args, *body.Enabled)
		argIdx++
	}
	if body.IntervalSeconds != nil {
		if *body.IntervalSeconds < 60 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "interval_seconds must be >= 60"})
			return
		}
		sets = append(sets, fmt.Sprintf("interval_seconds = $%d", argIdx))
		args = append(args, *body.IntervalSeconds)
		argIdx++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to update"})
		return
	}

	query := fmt.Sprintf("UPDATE scheduled_tasks SET %s WHERE name = $%d",
		strings.Join(sets, ", "), argIdx)
	args = append(args, name)

	res, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}

	h.logger.Info("Pipeline updated",
		zap.String("name", name),
		zap.Any("enabled", body.Enabled),
		zap.Any("interval", body.IntervalSeconds))

	c.JSON(http.StatusOK, gin.H{"updated": true, "name": name})
}

// ============================================================================
// POST /admin/pipelines/:name/trigger
// ============================================================================
// Force-trigger by clearing last_triggered_at. Scheduler picks it up next tick.

func (h *PipelineAdminHandlers) HandleTriggerPipeline(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	res, err := h.db.ExecContext(ctx, `
		UPDATE scheduled_tasks
		SET last_triggered_at = NULL,
			updated_at = NOW()
		WHERE name = $1 AND enabled = true
	`, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		// Check if it exists but is disabled
		var exists bool
		h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM scheduled_tasks WHERE name = $1)", name).Scan(&exists)
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "pipeline exists but is disabled — enable it first"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		}
		return
	}

	h.logger.Info("Pipeline triggered", zap.String("name", name))
	c.JSON(http.StatusOK, gin.H{"triggered": true, "name": name})
}

// ============================================================================
// GET /admin/pipelines/stats
// ============================================================================
// Returns aggregate stats for data pipelines: CH enrichment progress,
// verification counts, HTTP request summary, LLM call summary.

func (h *PipelineAdminHandlers) HandlePipelineStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats := gin.H{}

	// CH enrichment progress
	chStats := map[string]interface{}{}
	var total, matched, detailFetched, accountsFetched int
	err := h.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS total,
			COUNT(matched_business_id) AS matched,
			COUNT(*) FILTER (WHERE details_fetched = true) AS detail_fetched,
			COUNT(*) FILTER (WHERE accounts_fetched = true) AS accounts_fetched
		FROM business_intel.ch_vet_companies
		WHERE company_status = 'active'
	`).Scan(&total, &matched, &detailFetched, &accountsFetched)
	if err == nil {
		chStats["total_companies"] = total
		chStats["matched"] = matched
		chStats["detail_fetched"] = detailFetched
		chStats["accounts_fetched"] = accountsFetched
	}

	// Financial data coverage
	var withNetWorth, withEmployees, withTurnover, withProfitLoss int
	err = h.db.QueryRowContext(ctx, `
		SELECT
			COUNT(net_worth_gbp) AS has_net_worth,
			COUNT(employee_count) AS has_employees,
			COUNT(turnover_gbp) AS has_turnover,
			COUNT(profit_loss_gbp) AS has_profit_loss
		FROM business_intel.companies_house_data
		WHERE accounts_date IS NOT NULL
	`).Scan(&withNetWorth, &withEmployees, &withTurnover, &withProfitLoss)
	if err == nil {
		chStats["financial_coverage"] = map[string]int{
			"net_worth":   withNetWorth,
			"employees":   withEmployees,
			"turnover":    withTurnover,
			"profit_loss": withProfitLoss,
		}
	}
	stats["ch_enrichment"] = chStats

	// Business verification progress
	verifyStats := map[string]interface{}{}
	rows, err := h.db.QueryContext(ctx, `
		SELECT verification_status, COUNT(*)
		FROM business_intel.businesses
		WHERE is_active = true
		GROUP BY verification_status
	`)
	if err == nil {
		defer rows.Close()
		statusCounts := map[string]int{}
		for rows.Next() {
			var status string
			var count int
			if rows.Scan(&status, &count) == nil {
				statusCounts[status] = count
			}
		}
		verifyStats["by_status"] = statusCounts
	}
	stats["verification"] = verifyStats

	// HTTP request log summary (last 24h)
	httpStats := []map[string]interface{}{}
	rows2, err := h.db.QueryContext(ctx, `
		SELECT domain, action_name,
			COUNT(*) AS total_calls,
			COUNT(*) FILTER (WHERE success = true) AS successes,
			COUNT(*) FILTER (WHERE success = false) AS failures,
			ROUND(AVG(latency_ms)) AS avg_latency_ms
		FROM http_request_log
		WHERE created_at > NOW() - INTERVAL '24 hours'
		GROUP BY domain, action_name
		ORDER BY total_calls DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var domain, action sql.NullString
			var totalCalls, successes, failures int
			var avgLatency sql.NullInt64
			if rows2.Scan(&domain, &action, &totalCalls, &successes, &failures, &avgLatency) == nil {
				entry := map[string]interface{}{
					"domain":         domain.String,
					"action":         action.String,
					"total_calls":    totalCalls,
					"successes":      successes,
					"failures":       failures,
					"avg_latency_ms": avgLatency.Int64,
				}
				httpStats = append(httpStats, entry)
			}
		}
	}
	stats["http_requests_24h"] = httpStats

	// LLM call summary (last 24h)
	llmStats := []map[string]interface{}{}
	rows3, err := h.db.QueryContext(ctx, `
		SELECT agent_type, model, COUNT(*) AS calls,
			ROUND(AVG(latency_ms)) AS avg_latency_ms,
			SUM(input_tokens) AS total_input_tokens,
			SUM(output_tokens) AS total_output_tokens
		FROM llm_call_log
		WHERE created_at > NOW() - INTERVAL '24 hours'
		GROUP BY agent_type, model
		ORDER BY calls DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var agentType, model sql.NullString
			var calls int
			var avgLatency, inputTokens, outputTokens sql.NullInt64
			if rows3.Scan(&agentType, &model, &calls, &avgLatency, &inputTokens, &outputTokens) == nil {
				llmStats = append(llmStats, map[string]interface{}{
					"agent_type":          agentType.String,
					"model":               model.String,
					"calls":               calls,
					"avg_latency_ms":      avgLatency.Int64,
					"total_input_tokens":  inputTokens.Int64,
					"total_output_tokens": outputTokens.Int64,
				})
			}
		}
	}
	stats["llm_calls_24h"] = llmStats

	c.JSON(http.StatusOK, stats)
}
