// FILE: platform/orchestration/actions/check_endpoint_health_action.go
//
// Periodic health check for AI endpoints. Pings each endpoint in
// ai_endpoint_health where check_mode='active' and the check interval
// has elapsed. Updates healthy/last_checked/error for each.
//
// For Ollama endpoints: GET {url}/api/tags — 200 = healthy.
// For Claude: POST a cheap 1-token haiku request (~$0.000003/check).
//
// Triggered by kafka-scheduler as a fire_message=true task.
// Task name: "ai-endpoint-health-check"
//
// Registration:
//   "check_endpoint_health": {
//       Handler:     CheckEndpointHealthAction,
//       Category:    "system",
//       Description: "Ping AI endpoints and update health table",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CheckEndpointHealthInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"task_name"},
	Defaults:   map[string]interface{}{"task_name": "ai-endpoint-health-check"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("check_endpoint_health", CheckEndpointHealthInputSpec)
}

type endpointToCheck struct {
	URL      string
	Name     string
	PingPath string
}

func CheckEndpointHealthAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "check_endpoint_health"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Load endpoints due for checking
	rows, err := params.DB.QueryContext(ctx, `
		SELECT endpoint_url, name, COALESCE(ping_path, '/api/tags')
		FROM ai_endpoint_health
		WHERE check_mode = 'active'
		  AND (
		    last_checked IS NULL
		    OR last_checked + (check_interval_seconds || ' seconds')::interval <= NOW()
		  )
	`)
	if err != nil {
		return nil, fmt.Errorf("query endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []endpointToCheck
	for rows.Next() {
		var e endpointToCheck
		if err := rows.Scan(&e.URL, &e.Name, &e.PingPath); err != nil {
			logger.Warn("Failed to scan endpoint", zap.Error(err))
			continue
		}
		endpoints = append(endpoints, e)
	}

	if len(endpoints) == 0 {
		logger.Info("No endpoints due for health check")
		markTaskComplete(ctx, params.DB, config, logger)
		return map[string]interface{}{"checked": 0, "reason": "none_due"}, nil
	}

	checked := 0
	changed := 0

	for _, ep := range endpoints {
		var healthy bool
		var errMsg string

		if ep.PingPath == "claude_ping" {
			healthy, errMsg = pingClaude(ep.URL, logger)
		} else {
			healthy, errMsg = pingOllama(ep.URL, ep.PingPath, logger)
		}

		// Update health table
		var wasHealthy bool
		params.DB.QueryRowContext(ctx,
			`SELECT healthy FROM ai_endpoint_health WHERE endpoint_url = $1`,
			ep.URL).Scan(&wasHealthy)

		_, updateErr := params.DB.ExecContext(ctx, `
			UPDATE ai_endpoint_health
			SET healthy = $2,
			    last_checked = NOW(),
			    last_healthy = CASE WHEN $2 THEN NOW() ELSE last_healthy END,
			    error = $3,
			    updated_at = NOW()
			WHERE endpoint_url = $1
		`, ep.URL, healthy, nullIfEmpty(errMsg))

		if updateErr != nil {
			logger.Warn("Failed to update endpoint health",
				zap.String("endpoint", ep.Name), zap.Error(updateErr))
		}

		if wasHealthy != healthy {
			changed++
			logger.Info("Endpoint health changed",
				zap.String("endpoint", ep.Name),
				zap.Bool("was_healthy", wasHealthy),
				zap.Bool("now_healthy", healthy),
				zap.String("error", errMsg))
		} else {
			logger.Info("Endpoint health unchanged",
				zap.String("endpoint", ep.Name),
				zap.Bool("healthy", healthy))
		}

		checked++
	}

	// Notify scheduler
	markTaskComplete(ctx, params.DB, config, logger)

	return map[string]interface{}{
		"checked": checked,
		"changed": changed,
	}, nil
}

// pingOllama does GET {baseURL}{pingPath} and checks for 200.
func pingOllama(baseURL string, pingPath string, logger *zap.Logger) (bool, string) {
	url := strings.TrimRight(baseURL, "/") + pingPath
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // drain

	if resp.StatusCode == 200 {
		return true, ""
	}
	return false, fmt.Sprintf("status %d", resp.StatusCode)
}

// pingClaude sends a minimal 1-token haiku request.
// Cost: ~$0.000003 per check. At hourly = ~$0.002/month.
func pingClaude(baseURL string, logger *zap.Logger) (bool, string) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return false, "ANTHROPIC_API_KEY not set"
	}

	body := `{"model":"claude-haiku-4-5-20251001","max_tokens":1,"messages":[{"role":"user","content":"1"}]}`
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(body))
	if err != nil {
		return false, err.Error()
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // drain

	switch resp.StatusCode {
	case 200:
		return true, ""
	case 402:
		return false, "credits exhausted"
	case 401:
		return false, "authentication failed"
	case 529:
		return true, "" // overloaded but reachable — credits valid
	default:
		return true, "" // any non-auth error means API is reachable
	}
}

// markTaskComplete updates the scheduled_tasks table.
func markTaskComplete(ctx context.Context, db *sql.DB, config map[string]interface{}, logger *zap.Logger) {
	taskName := "ai-endpoint-health-check"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, err := db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)
	if err != nil {
		logger.Warn("Failed to mark task complete", zap.String("task", taskName), zap.Error(err))
	}
}
