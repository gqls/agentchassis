// FILE: platform/orchestration/actions/seed_build_queue_action.go
//
// Reads queued entries from build_queue and seeds them into the work-item pipeline.
// For each domain: creates/updates site record, examines direction field to determine
// the first work item, inserts it, marks the queue entry as 'seeded'.
//
// Direction field controls what happens:
//   - null                    → needs_domain_research (full pipeline from scratch)
//   - {"objective": "..."}    → needs_domain_research with objective hint in spec
//   - {"brief_complete": true}→ needs_site_plan (skip research/briefing)
//   - {"adopt_from": "..."}   → needs_site_adoption (clone from reference site)
//   - {"fork_from": "..."}    → needs_site_plan (copies specs from source site)
//
// Workflow config example:
//
//   "seed_queue": {
//       "action": "seed_build_queue",
//       "config": {
//           "max_entries": 20
//       },
//       "output_field": "seed_result",
//       "next_step": "complete"
//   }
//
// Registration (add to registry.go):
//   "seed_build_queue": { Handler: SeedBuildQueueAction, IsLocal: true },

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var SeedBuildQueueInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("seed_build_queue", SeedBuildQueueInputSpec)
}

func SeedBuildQueueAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "seed_build_queue"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config
	maxEntries := datahelpers.GetIntField(config, "max_entries", 20)

	// Read queued entries
	rows, err := params.DB.QueryContext(ctx, `
		SELECT id, domain, direction, priority, batch_id
		FROM build_queue
		WHERE status = 'queued'
		ORDER BY priority ASC, created_at ASC
		LIMIT $1
	`, maxEntries)
	if err != nil {
		return nil, fmt.Errorf("query build_queue: %w", err)
	}
	defer rows.Close()

	type queueEntry struct {
		id        uuid.UUID
		domain    string
		direction map[string]interface{}
		priority  int
		batchID   *uuid.UUID
	}

	var entries []queueEntry
	for rows.Next() {
		var e queueEntry
		var dirJSON []byte
		var batchIDPtr *uuid.UUID

		if err := rows.Scan(&e.id, &e.domain, &dirJSON, &e.priority, &batchIDPtr); err != nil {
			logger.Warn("SeedBuildQueueAction: scan error", zap.Error(err))
			continue
		}

		if len(dirJSON) > 0 && string(dirJSON) != "null" {
			if err := json.Unmarshal(dirJSON, &e.direction); err != nil {
				logger.Warn("SeedBuildQueueAction: invalid direction JSON",
					zap.String("domain", e.domain), zap.Error(err))
			}
		}
		e.batchID = batchIDPtr
		entries = append(entries, e)
	}

	if len(entries) == 0 {
		logger.Info("SeedBuildQueueAction: no queued entries")
		return map[string]interface{}{
			"seeded":  0,
			"skipped": 0,
			"total":   0,
		}, nil
	}

	logger.Info("SeedBuildQueueAction: processing entries",
		zap.Int("count", len(entries)))

	// Get default network for site creation
	networkID, err := getDefaultNetworkID(ctx, params.DB, logger)
	if err != nil {
		logger.Warn("SeedBuildQueueAction: using placeholder network ID", zap.Error(err))
		networkID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	}

	seeded := 0
	skipped := 0

	for _, entry := range entries {
		domain := cleanDomain(entry.domain)

		// Create/update site record (reuses existing helper from ensure_site_record)
		site, err := upsertSite(ctx, params.DB, domain, networkID, logger)
		if err != nil {
			logger.Warn("SeedBuildQueueAction: failed to upsert site",
				zap.String("domain", domain), zap.Error(err))
			skipped++
			continue
		}

		// Determine first work item based on direction
		itemType, handlerAgent, spec := seedDetermineFirstItem(entry.direction, domain, logger)

		specJSON, err := json.Marshal(spec)
		if err != nil {
			logger.Warn("SeedBuildQueueAction: marshal spec error",
				zap.String("domain", domain), zap.Error(err))
			skipped++
			continue
		}

		// Use a transaction for the work item insert + queue status update
		tx, err := params.DB.BeginTx(ctx, nil)
		if err != nil {
			logger.Warn("SeedBuildQueueAction: begin tx error",
				zap.String("domain", domain), zap.Error(err))
			skipped++
			continue
		}

		var batchID uuid.UUID
		if entry.batchID != nil {
			batchID = *entry.batchID
		}

		inserted, err := insertWorkItem(ctx, tx, workItem{
			siteID:       site.ID,
			source:       "seed",
			pipeline:     "build",
			itemType:     itemType,
			severity:     "high",
			summary:      fmt.Sprintf("Seed: %s for %s", itemType, domain),
			spec:         string(specJSON),
			pageID:       nil,
			priority:     entry.priority,
			handlerAgent: handlerAgent,
			status:       "triaged",
			createdBy:    "seed_build_queue",
			itemKey:      fmt.Sprintf("seed_%s_%s", itemType, domain),
			batchID:      batchID,
		}, logger)
		if err != nil {
			tx.Rollback()
			logger.Warn("SeedBuildQueueAction: insert work item failed",
				zap.String("domain", domain), zap.Error(err))
			skipped++
			continue
		}

		// Mark queue entry as seeded
		_, err = tx.ExecContext(ctx, `
			UPDATE build_queue
			SET status = 'seeded', updated_at = now()
			WHERE id = $1
		`, entry.id)
		if err != nil {
			tx.Rollback()
			logger.Warn("SeedBuildQueueAction: failed to mark as seeded",
				zap.String("domain", domain), zap.Error(err))
			skipped++
			continue
		}

		if err := tx.Commit(); err != nil {
			logger.Warn("SeedBuildQueueAction: commit failed",
				zap.String("domain", domain), zap.Error(err))
			skipped++
			continue
		}

		if inserted {
			seeded++
		} else {
			skipped++ // dedup — item_key already existed
		}

		logger.Info("SeedBuildQueueAction: domain seeded",
			zap.String("domain", domain),
			zap.String("site_id", site.ID.String()),
			zap.String("item_type", itemType),
			zap.String("handler_agent", handlerAgent),
			zap.Bool("item_inserted", inserted),
		)
	}

	logger.Info("SeedBuildQueueAction: complete",
		zap.Int("seeded", seeded),
		zap.Int("skipped", skipped),
	)

	return map[string]interface{}{
		"seeded":  seeded,
		"skipped": skipped,
		"total":   len(entries),
	}, nil
}

// seedDetermineFirstItem examines the direction field and returns the item_type,
// handler_agent, and spec for the first work item to create.
func seedDetermineFirstItem(direction map[string]interface{}, domain string, logger *zap.Logger) (string, string, map[string]interface{}) {
	if direction == nil {
		// No direction — start from scratch with domain research
		logger.Info("seedDetermineFirstItem: no direction, starting with domain research",
			zap.String("domain", domain))
		return "needs_domain_research", "domain-research-classifier", map[string]interface{}{
			"domain": domain,
		}
	}

	// Check for adoption
	if adoptFrom, ok := direction["adopt_from"].(string); ok && adoptFrom != "" {
		logger.Info("seedDetermineFirstItem: adoption mode",
			zap.String("domain", domain),
			zap.String("adopt_from", adoptFrom))
		return "needs_site_adoption", "site-adoption-agent", map[string]interface{}{
			"domain":     domain,
			"adopt_from": adoptFrom,
		}
	}

	// Check for fork
	if forkFrom, ok := direction["fork_from"].(string); ok && forkFrom != "" {
		logger.Info("seedDetermineFirstItem: fork mode",
			zap.String("domain", domain),
			zap.String("fork_from", forkFrom))
		return "needs_site_plan", "site-planner", map[string]interface{}{
			"domain":    domain,
			"fork_from": forkFrom,
		}
	}

	// Check for brief_complete (skip research/briefing)
	if bc, ok := direction["brief_complete"].(bool); ok && bc {
		logger.Info("seedDetermineFirstItem: brief complete, skipping to plan",
			zap.String("domain", domain))
		return "needs_site_plan", "site-planner", map[string]interface{}{
			"domain": domain,
		}
	}

	// Has objective but needs research
	if objective, ok := direction["objective"].(string); ok && objective != "" {
		logger.Info("seedDetermineFirstItem: objective provided, starting with research",
			zap.String("domain", domain),
			zap.String("objective", objective))
		return "needs_domain_research", "domain-research-classifier", map[string]interface{}{
			"domain":    domain,
			"objective": objective,
		}
	}

	// Fallback — unknown direction shape, default to research
	logger.Warn("seedDetermineFirstItem: unrecognized direction, defaulting to research",
		zap.String("domain", domain),
		zap.Any("direction", direction))
	return "needs_domain_research", "domain-research-classifier", map[string]interface{}{
		"domain": domain,
	}
}
