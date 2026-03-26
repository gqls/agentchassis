// FILE: platform/orchestration/actions/dispatch_feed_sources_action.go
//
// DispatchFeedSourcesAction is the core of the content-feed-orchestrator.
// It queries content_sources for due sources and, for each one, produces
// a message to the generic agent's requests topic with a spawn+call
// inline workflow targeting feed-ingester. This reuses the standard
// CLI trigger pattern.
//
// The generic agent receives each message, spawns a feed-ingester K8s job,
// and calls it with the source config. Each ingester runs independently
// with its own logs and orchestration.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DispatchFeedSourcesInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"source_type", "max_dispatches"},
	Defaults: map[string]interface{}{
		"max_dispatches": 10,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("dispatch_feed_sources", DispatchFeedSourcesInputSpec)
}

func DispatchFeedSourcesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "dispatch_feed_sources"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		DispatchFeedSourcesInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	sourceTypeFilter := inputs.Get("source_type")
	maxDispatches := inputs.GetInt("max_dispatches", 10)

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// Check if content feed is enabled for this site
	feedEnabled, err := isFeedEnabled(ctx, params.DB, siteID)
	if err != nil {
		logger.Warn("DispatchFeedSourcesAction: failed to check feed config, proceeding anyway",
			zap.Error(err))
	} else if !feedEnabled {
		logger.Info("DispatchFeedSourcesAction: content feed disabled for site",
			zap.String("site_id", siteIDStr))
		return map[string]interface{}{
			"dispatched": 0,
			"site_id":    siteIDStr,
			"reason":     "content_feed_disabled",
		}, nil
	}

	// Query due sources
	query := `
		SELECT id, source_type, name, config
		FROM content_sources
		WHERE site_id = $1
		  AND is_active = true
		  AND (next_fetch_at IS NULL OR next_fetch_at <= NOW())
	`
	args := []interface{}{siteID}

	if sourceTypeFilter != "" {
		query += " AND source_type = $2"
		args = append(args, sourceTypeFilter)
	}

	query += " ORDER BY next_fetch_at ASC NULLS FIRST LIMIT $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, maxDispatches)

	rows, err := params.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query due sources: %w", err)
	}
	defer rows.Close()

	type sourceRecord struct {
		ID         string
		SourceType string
		Name       string
		Config     json.RawMessage
	}

	var sources []sourceRecord
	for rows.Next() {
		var s sourceRecord
		if err := rows.Scan(&s.ID, &s.SourceType, &s.Name, &s.Config); err != nil {
			logger.Warn("DispatchFeedSourcesAction: scan failed", zap.Error(err))
			continue
		}
		sources = append(sources, s)
	}

	if len(sources) == 0 {
		logger.Info("DispatchFeedSourcesAction: no sources due for fetch",
			zap.String("site_id", siteIDStr))
		return map[string]interface{}{
			"dispatched": 0,
			"site_id":    siteIDStr,
		}, nil
	}

	logger.Info("DispatchFeedSourcesAction: found due sources",
		zap.String("site_id", siteIDStr),
		zap.Int("count", len(sources)),
	)

	// For each source, produce a message to generic agent with spawn+call workflow
	genericRequestsTopic := "system.agent.generic.requests"
	dispatched := 0
	dispatchErrors := 0

	for _, source := range sources {
		var sourceConfig map[string]interface{}
		if err := json.Unmarshal(source.Config, &sourceConfig); err != nil {
			logger.Warn("DispatchFeedSourcesAction: failed to parse source config",
				zap.String("source_id", source.ID),
				zap.Error(err))
			dispatchErrors++
			continue
		}

		correlationID := uuid.New().String()
		orchestrationID := uuid.New().String()
		messageID := uuid.New().String()
		requestID := uuid.New().String()
		timestamp := time.Now().UTC().Format(time.RFC3339)

		// Build the spawn+call inline workflow for the generic agent
		inlineWorkflow := map[string]interface{}{
			"start_step": "spawn_ingester",
			"steps": map[string]interface{}{
				"spawn_ingester": map[string]interface{}{
					"action": "spawn_agent",
					"config": map[string]interface{}{
						"agent_type": "feed-ingester",
						"role":       "ingester-" + source.SourceType,
					},
					"next_step":   "call_ingester",
					"description": "Spawn feed-ingester for " + source.Name,
				},
				"call_ingester": map[string]interface{}{
					"action": "call_agent",
					"config": map[string]interface{}{
						"agent_type": "feed-ingester",
						"input_mapping": map[string]interface{}{
							"site_id":       siteIDStr,
							"source_id":     source.ID,
							"source_type":   source.SourceType,
							"source_name":   source.Name,
							"source_config": sourceConfig,
						},
					},
					"next_step":   "complete",
					"description": "Call feed-ingester with source config",
				},
				"complete": map[string]interface{}{
					"action":      "complete_workflow",
					"description": "Done",
				},
			},
		}

		// Get the parent responses topic for routing
		parentResponsesTopic := params.ExecutionContext.ResponsesTopic
		if parentResponsesTopic == "" {
			parentResponsesTopic = os.Getenv("RESPONSES_TOPIC")
		}
		if parentResponsesTopic == "" {
			parentResponsesTopic = "system.agent.generic.responses"
		}

		message := map[string]interface{}{
			"headers": map[string]interface{}{
				"correlation_id":         correlationID,
				"orchestration_id":       orchestrationID,
				"message_type":           "request",
				"action":                 "orchestrate",
				"client_id":              params.ExecutionContext.ClientID,
				"message_id":             messageID,
				"request_id":             requestID,
				"timestamp":              timestamp,
				"parent_responses_topic": parentResponsesTopic,
				"sender": map[string]interface{}{
					"agent_id":   params.ExecutionContext.FromAgentID,
					"agent_type": params.AgentType,
					"pod_name":   os.Getenv("POD_NAME"),
				},
			},
			"config": map[string]interface{}{
				"workflow":        inlineWorkflow,
				"processing_mode": "orchestrator",
				"timeout_seconds": 300,
			},
			"input_data": map[string]interface{}{
				"site_id":       siteIDStr,
				"source_id":     source.ID,
				"source_type":   source.SourceType,
				"source_name":   source.Name,
				"source_config": sourceConfig,
			},
		}

		msgBytes, err := json.Marshal(message)
		if err != nil {
			logger.Warn("DispatchFeedSourcesAction: failed to marshal message",
				zap.String("source_id", source.ID),
				zap.Error(err))
			dispatchErrors++
			continue
		}

		headers := map[string]string{
			"correlation_id":   correlationID,
			"orchestration_id": orchestrationID,
			"message_type":     "request",
			"action":           "orchestrate",
			"client_id":        params.ExecutionContext.ClientID,
		}

		if err := params.Producer.ProduceWithValidation(
			ctx,
			genericRequestsTopic,
			headers,
			[]byte(correlationID),
			msgBytes,
		); err != nil {
			logger.Warn("DispatchFeedSourcesAction: failed to produce message",
				zap.String("source_id", source.ID),
				zap.String("topic", genericRequestsTopic),
				zap.Error(err))
			dispatchErrors++
			continue
		}

		// Optimistically update next_fetch_at to prevent re-dispatch before completion
		// The feed-ingester's update_source_timestamps will set the real next time
		_, _ = params.DB.ExecContext(ctx, `
			UPDATE content_sources
			SET next_fetch_at = NOW() + fetch_interval,
			    updated_at = NOW()
			WHERE id = $1
		`, source.ID)

		dispatched++
		logger.Info("DispatchFeedSourcesAction: dispatched ingester",
			zap.String("source_id", source.ID),
			zap.String("source_type", source.SourceType),
			zap.String("source_name", source.Name),
			zap.String("correlation_id", correlationID),
		)
	}

	logger.Info("DispatchFeedSourcesAction: dispatch complete",
		zap.String("site_id", siteIDStr),
		zap.Int("dispatched", dispatched),
		zap.Int("errors", dispatchErrors),
		zap.Int("total_sources", len(sources)),
	)

	return map[string]interface{}{
		"dispatched": dispatched,
		"errors":     dispatchErrors,
		"total":      len(sources),
		"site_id":    siteIDStr,
	}, nil
}

// isFeedEnabled checks the site's maintenance_profile for content_feed.enabled
func isFeedEnabled(ctx context.Context, db *sql.DB, siteID uuid.UUID) (bool, error) {
	var settings json.RawMessage
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(settings, '{}'::jsonb) FROM sites WHERE id = $1
	`, siteID).Scan(&settings)
	if err != nil {
		return false, err
	}

	var s map[string]interface{}
	if err := json.Unmarshal(settings, &s); err != nil {
		return false, err
	}

	// Navigate: settings.maintenance_profile.content_feed.enabled
	if mp, ok := s["maintenance_profile"].(map[string]interface{}); ok {
		if cf, ok := mp["content_feed"].(map[string]interface{}); ok {
			if enabled, ok := cf["enabled"].(bool); ok {
				return enabled, nil
			}
		}
	}

	// Default: enabled if content_sources exist for this site
	var hasSource bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM content_sources WHERE site_id = $1 AND is_active = true)
	`, siteID).Scan(&hasSource)
	return hasSource, err
}
