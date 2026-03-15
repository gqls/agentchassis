// FILE: platform/orchestration/actions/cleanup_stale_topics_action.go
//
// CleanupStaleTopicsAction removes orphaned job.* topics from Kafka.
//
// Job topics are created per-orchestration for spawned child agents. Once the
// orchestration completes or fails, these topics are no longer needed but are
// never cleaned up. Over time they accumulate into the hundreds.
//
// Logic:
//   1. List all Kafka topics with "job." prefix
//   2. Query orchestration_states for any active orchestration referencing each topic
//   3. If no active orchestration references the topic, and the topic isn't
//      referenced by ANY orchestration updated in the last 2 hours → delete it
//
// This is triggered by the kafka-scheduler as a periodic maintenance task.
//
// Registration:
//   "cleanup_stale_topics": {
//       Handler:     CleanupStaleTopicsAction,
//       Category:    "maintenance",
//       Description: "Remove orphaned job.* Kafka topics",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CleanupStaleTopicsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"min_age_hours", "dry_run", "batch_size"},
	Defaults:   map[string]interface{}{"min_age_hours": 2, "dry_run": false, "batch_size": 50},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("cleanup_stale_topics", CleanupStaleTopicsInputSpec)
}

func CleanupStaleTopicsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "cleanup_stale_topics"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Parse options
	minAgeHours := 2
	if v, ok := config["min_age_hours"].(float64); ok {
		minAgeHours = int(v)
	}
	dryRun := false
	if v, ok := config["dry_run"].(bool); ok {
		dryRun = v
	}
	batchSize := 50
	if v, ok := config["batch_size"].(float64); ok {
		batchSize = int(v)
	}

	// Get Kafka brokers
	brokers := kafka.GetBrokers()
	if len(brokers) == 0 {
		return nil, fmt.Errorf("no Kafka brokers configured")
	}

	topicManager := kafka.NewTopicManager(brokers, logger)

	// Step 1: List all topics
	allTopics, err := topicManager.ListTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	// Filter to job.* topics only
	var jobTopics []string
	for _, t := range allTopics {
		if strings.HasPrefix(t, "job.") {
			jobTopics = append(jobTopics, t)
		}
	}

	logger.Info("Found job topics",
		zap.Int("job_topics", len(jobTopics)),
		zap.Int("total_topics", len(allTopics)))

	if len(jobTopics) == 0 {
		return map[string]interface{}{
			"job_topics_found": 0,
			"deleted":          0,
		}, nil
	}

	// Step 2: Get all topics referenced by active orchestrations
	activeTopics, err := getActiveOrchestrationTopics(ctx, params.DB, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get active topics: %w", err)
	}

	// Step 3: Get topics referenced by recently-updated orchestrations (safety buffer)
	recentTopics, err := getRecentOrchestrationTopics(ctx, params.DB, minAgeHours, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent topics: %w", err)
	}

	// Step 4: Build protected set
	protected := make(map[string]bool)
	for _, t := range activeTopics {
		protected[t] = true
	}
	for _, t := range recentTopics {
		protected[t] = true
	}

	// Also protect system topics (non-job)
	// and any topic containing "system." just in case
	for _, t := range allTopics {
		if !strings.HasPrefix(t, "job.") {
			protected[t] = true
		}
	}

	// Step 5: Find deletable topics
	var toDelete []string
	for _, t := range jobTopics {
		if !protected[t] {
			toDelete = append(toDelete, t)
		}
	}

	logger.Info("Topic cleanup analysis",
		zap.Int("job_topics", len(jobTopics)),
		zap.Int("active_protected", len(activeTopics)),
		zap.Int("recent_protected", len(recentTopics)),
		zap.Int("candidates_for_deletion", len(toDelete)),
		zap.Bool("dry_run", dryRun))

	if dryRun {
		// Log what would be deleted but don't do it
		for _, t := range toDelete {
			logger.Info("Would delete (dry run)", zap.String("topic", t))
		}
		return map[string]interface{}{
			"job_topics_found": len(jobTopics),
			"would_delete":     len(toDelete),
			"active_protected": len(activeTopics),
			"recent_protected": len(recentTopics),
			"dry_run":          true,
		}, nil
	}

	// Step 6: Delete in batches
	deleted := 0
	failed := 0
	for i, t := range toDelete {
		if i >= batchSize {
			logger.Info("Reached batch size limit, remaining topics will be cleaned next run",
				zap.Int("deleted_this_run", deleted),
				zap.Int("remaining", len(toDelete)-i))
			break
		}

		if err := topicManager.DeleteTopic(ctx, t); err != nil {
			logger.Warn("Failed to delete topic",
				zap.String("topic", t),
				zap.Error(err))
			failed++
			continue
		}

		logger.Info("Deleted stale topic", zap.String("topic", t))
		deleted++
	}

	logger.Info("Topic cleanup complete",
		zap.Int("deleted", deleted),
		zap.Int("failed", failed),
		zap.Int("total_candidates", len(toDelete)))

	return map[string]interface{}{
		"job_topics_found": len(jobTopics),
		"deleted":          deleted,
		"failed":           failed,
		"total_candidates": len(toDelete),
		"active_protected": len(activeTopics),
		"recent_protected": len(recentTopics),
	}, nil
}

// getActiveOrchestrationTopics returns all topics referenced by non-terminal orchestrations.
func getActiveOrchestrationTopics(ctx context.Context, db *sql.DB, logger *zap.Logger) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT topic FROM (
			SELECT requests_topic AS topic FROM orchestration_states
			WHERE status IN ('RUNNING', 'AWAITING_RESPONSES', 'PAUSED_FOR_HUMAN_INPUT', 'EXECUTING_STEP')
			  AND requests_topic IS NOT NULL AND requests_topic != ''
			UNION ALL
			SELECT responses_topic AS topic FROM orchestration_states
			WHERE status IN ('RUNNING', 'AWAITING_RESPONSES', 'PAUSED_FOR_HUMAN_INPUT', 'EXECUTING_STEP')
			  AND responses_topic IS NOT NULL AND responses_topic != ''
		) active_topics
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			continue
		}
		topics = append(topics, t)
	}

	logger.Info("Active orchestration topics", zap.Int("count", len(topics)))
	return topics, nil
}

// getRecentOrchestrationTopics returns topics from orchestrations updated within the
// safety window. This prevents deleting topics for orchestrations that just completed
// but might have in-flight messages.
func getRecentOrchestrationTopics(ctx context.Context, db *sql.DB, minAgeHours int, logger *zap.Logger) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT topic FROM (
			SELECT requests_topic AS topic FROM orchestration_states
			WHERE updated_at > NOW() - ($1 || ' hours')::interval
			  AND requests_topic IS NOT NULL AND requests_topic != ''
			UNION ALL
			SELECT responses_topic AS topic FROM orchestration_states
			WHERE updated_at > NOW() - ($1 || ' hours')::interval
			  AND responses_topic IS NOT NULL AND responses_topic != ''
		) recent_topics
	`, minAgeHours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			continue
		}
		topics = append(topics, t)
	}

	logger.Info("Recent orchestration topics (protected)",
		zap.Int("count", len(topics)),
		zap.Int("min_age_hours", minAgeHours))
	return topics, nil
}
