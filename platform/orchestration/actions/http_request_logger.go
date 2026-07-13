// FILE: platform/orchestration/actions/http_request_logger.go
// Centralised HTTP request logging for outbound API calls.
// Follows the same fire-and-forget pattern as llm_call_logger.go.
//
// Usage:
//
//	LogHTTPRequest(params.DB, params.Logger, HTTPRequestLogParams{
//	    AgentType:  params.AgentType,
//	    ActionName: "ch_fetch_accounts",
//	    Method:     "GET",
//	    URL:        fullURL,
//	    StatusCode: resp.StatusCode,
//	    ...
//	})
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// HTTPRequestLogParams captures what we want to log about an outbound HTTP call.
type HTTPRequestLogParams struct {
	// Who made the call
	AgentType       string
	AgentID         interface{} // may be string or nil from headers
	StepName        string
	OrchestrationID string
	CorrelationID   string
	ActionName      string // e.g. "ch_fetch_accounts", "ch_detail_fetch"

	// What was called
	Method string
	URL    string

	// Response
	StatusCode    int
	ResponseBytes int
	ContentType   string
	LatencyMs     int
	Success       bool
	ErrorMessage  string

	// Extra context
	Metadata map[string]interface{} // e.g. {"company_number": "12345678"}
}

// LogHTTPRequest logs an outbound HTTP request asynchronously.
// Fire-and-forget: runs in a goroutine with a 5s timeout.
// Failures are logged but never propagate to the caller.
func LogHTTPRequest(db *sql.DB, logger *zap.Logger, p HTTPRequestLogParams) {
	if db == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		domain, pathStr := extractDomainAndPath(p.URL)

		agentID := ""
		if s, ok := p.AgentID.(string); ok {
			agentID = s
		}

		metadataJSON := []byte("{}")
		if p.Metadata != nil {
			if b, err := json.Marshal(p.Metadata); err == nil {
				metadataJSON = b
			}
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO http_request_log (
				agent_type, agent_id, step_name, orchestration_id, correlation_id,
				action_name, method, url, domain, path,
				status_code, response_bytes, content_type, latency_ms,
				success, error_message, metadata
			) VALUES (
				$1, NULLIF($2, ''), $3, $4, $5,
				$6, $7, $8, $9, $10,
				$11, $12, NULLIF($13, ''), $14,
				$15, NULLIF($16, ''), $17::jsonb
			)`,
			p.AgentType,
			agentID,
			p.StepName,
			p.OrchestrationID,
			p.CorrelationID,
			p.ActionName,
			p.Method,
			p.URL,
			domain,
			pathStr,
			p.StatusCode,
			p.ResponseBytes,
			p.ContentType,
			p.LatencyMs,
			p.Success,
			p.ErrorMessage,
			metadataJSON,
		)
		if err != nil {
			logger.Warn("Failed to log HTTP request",
				zap.Error(err),
				zap.String("url", p.URL))
		}
	}()
}

// extractDomainAndPath pulls the domain and path from a URL string.
func extractDomainAndPath(rawURL string) (string, string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", rawURL
	}
	return parsed.Host, parsed.Path
}
