// FILE: internal/core-manager/admin/topic_cleanup_handler.go
//
// Admin endpoint for cleaning up stale job.* Kafka topics.
//
// Job topics are created per-orchestration for spawned child agents. Once the
// orchestration completes or fails, these topics are no longer needed. This
// handler cross-references Kafka topics against orchestration_states to find
// and delete orphans.
//
// Endpoint:
//   POST /api/v1/admin/system/cleanup-topics
//   Query params:
//     dry_run=true     — list candidates without deleting (default: false)
//     min_age_hours=2  — only delete if orchestration completed > N hours ago (default: 2)
//     batch_size=50    — max topics to delete per call (default: 50)

package admin

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

// HandleCleanupStaleTopics removes orphaned job.* topics from Kafka.
func (h *SystemHandlers) HandleCleanupStaleTopics(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse options
	dryRun := c.DefaultQuery("dry_run", "false") == "true"
	minAgeHours := 2
	if v, err := strconv.Atoi(c.DefaultQuery("min_age_hours", "2")); err == nil && v > 0 {
		minAgeHours = v
	}
	batchSize := 50
	if v, err := strconv.Atoi(c.DefaultQuery("batch_size", "50")); err == nil && v > 0 && v <= 500 {
		batchSize = v
	}

	// Get Kafka brokers
	brokers := kafka.GetBrokers()
	if len(brokers) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no Kafka brokers configured"})
		return
	}

	topicManager := kafka.NewTopicManager(brokers, h.logger)

	// Step 1: List all topics
	allTopics, err := topicManager.ListTopics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list topics: " + err.Error()})
		return
	}

	// Filter to job.* topics
	var jobTopics []string
	for _, t := range allTopics {
		if strings.HasPrefix(t, "job.") {
			jobTopics = append(jobTopics, t)
		}
	}

	h.logger.Info("Topic cleanup: listing",
		zap.Int("job_topics", len(jobTopics)),
		zap.Int("total_topics", len(allTopics)))

	if len(jobTopics) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"job_topics_found": 0,
			"deleted":          0,
			"message":          "no job topics found",
		})
		return
	}

	// Step 2: Get topics referenced by active orchestrations
	activeTopics, err := queryOrchestrationTopics(ctx, h.clientsDB,
		`status IN ('RUNNING', 'AWAITING_RESPONSES', 'PAUSED_FOR_HUMAN_INPUT', 'EXECUTING_STEP')`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query active topics: " + err.Error()})
		return
	}

	// Step 3: Get topics from recently-updated orchestrations (safety buffer)
	recentTopics, err := queryOrchestrationTopics(ctx, h.clientsDB,
		"updated_at > NOW() - ("+strconv.Itoa(minAgeHours)+" || ' hours')::interval")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query recent topics: " + err.Error()})
		return
	}

	// Step 4: Build protected set
	protected := make(map[string]bool)
	for _, t := range activeTopics {
		protected[t] = true
	}
	for _, t := range recentTopics {
		protected[t] = true
	}

	// Step 5: Find deletable topics
	var toDelete []string
	for _, t := range jobTopics {
		if !protected[t] {
			toDelete = append(toDelete, t)
		}
	}

	h.logger.Info("Topic cleanup: analysis",
		zap.Int("job_topics", len(jobTopics)),
		zap.Int("active_protected", len(activeTopics)),
		zap.Int("recent_protected", len(recentTopics)),
		zap.Int("candidates", len(toDelete)),
		zap.Bool("dry_run", dryRun))

	if dryRun {
		sampleTopics := toDelete
		if len(sampleTopics) > 20 {
			sampleTopics = sampleTopics[:20]
		}
		c.JSON(http.StatusOK, gin.H{
			"job_topics_found": len(jobTopics),
			"would_delete":     len(toDelete),
			"active_protected": len(activeTopics),
			"recent_protected": len(recentTopics),
			"dry_run":          true,
			"sample_deletions": sampleTopics,
		})
		return
	}

	// Step 6: Delete in batches
	deleted := 0
	failed := 0
	var deletedTopics []string
	var failedTopics []string

	for i, t := range toDelete {
		if i >= batchSize {
			break
		}

		if err := topicManager.DeleteTopic(ctx, t); err != nil {
			h.logger.Warn("Failed to delete topic",
				zap.String("topic", t),
				zap.Error(err))
			failed++
			failedTopics = append(failedTopics, t)
			continue
		}

		h.logger.Info("Deleted stale topic", zap.String("topic", t))
		deleted++
		deletedTopics = append(deletedTopics, t)
	}

	remaining := len(toDelete) - batchSize
	if remaining < 0 {
		remaining = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"job_topics_found": len(jobTopics),
		"deleted":          deleted,
		"failed":           failed,
		"total_candidates": len(toDelete),
		"remaining":        remaining,
		"active_protected": len(activeTopics),
		"recent_protected": len(recentTopics),
	})
}

// queryOrchestrationTopics returns distinct topics from orchestration_states matching a WHERE clause.
// The whereClause is always hardcoded by callers, never user input.
func queryOrchestrationTopics(ctx context.Context, db *sql.DB, whereClause string) ([]string, error) {
	query := `
		SELECT DISTINCT topic FROM (
			SELECT requests_topic AS topic FROM orchestration_states
			WHERE ` + whereClause + `
			  AND requests_topic IS NOT NULL AND requests_topic != ''
			UNION ALL
			SELECT responses_topic AS topic FROM orchestration_states
			WHERE ` + whereClause + `
			  AND responses_topic IS NOT NULL AND responses_topic != ''
		) topics
	`

	rows, err := db.QueryContext(ctx, query)
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
	return topics, nil
}
