// FILE: platform/orchestration/actions/feed_triage_actions.go
//
// Actions for the feed-triage agent. Scores ingested items for relevance
// and credibility to the site's vertical, values, and audience using the
// site spec.
//
// Actions:
//   - apply_feed_scores: reads LLM scores, updates content_feed_items rows
//     with relevance_score, credibility, source_attribution, and status
//   - load_feed_items_for_triage: loads unscored items with source metadata

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// ApplyFeedScoresAction
// ---------------------------------------------------------------------------

var ApplyFeedScoresInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"scores_field", "relevance_threshold"},
}

func init() {
	datahelpers.RegisterActionInputSpec("apply_feed_scores", ApplyFeedScoresInputSpec)
	datahelpers.RegisterActionInputSpec("load_feed_items_for_triage", LoadFeedItemsForTriageInputSpec)
}

// ApplyFeedScoresAction reads LLM-produced scores from collected_data and
// updates content_feed_items rows with relevance_score, credibility,
// source_attribution, status, and topics.
//
// Expected input shape (from execute_llm_prompt with response_format: json):
//
//	scores.result = [
//	    {
//	        "id": "uuid",
//	        "score": 85,
//	        "credibility": "high",
//	        "credibility_reason": "Reuters wire service article with direct quotes",
//	        "source_attribution": {
//	            "original_source": "Reuters",
//	            "found_via": "OilPrice RSS",
//	            "source_tier": "tier1_news"
//	        },
//	        "reason": "directly relevant to gas wholesale market",
//	        "topics": ["gas prices", "supply"],
//	        "flagged": false
//	    },
//	    ...
//	]
//
// Status thresholds (configurable via relevance_threshold, default 50):
//
//	score >= threshold  → status = 'relevant'   (displays on site)
//	score 20..threshold → status = 'review'     (held for manual check)
//	score < 20          → status = 'rejected'   (never displays)
//
// Items flagged with "flagged": true are always set to 'rejected' regardless
// of score (for values/legal conflicts or low-credibility unverified claims).
func ApplyFeedScoresAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "apply_feed_scores"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ApplyFeedScoresInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID := inputs.Get("site_id")
	if _, err := uuid.Parse(siteID); err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Read threshold from config or site settings
	threshold := 50.0
	if t, ok := params.StepConfig.Config["relevance_threshold"].(float64); ok && t > 0 {
		threshold = t
	}
	// Override from site settings if present
	if siteThreshold := datahelpers.ExtractNestedField(params.CollectedData,
		"site_spec.data.maintenance_profile.content_feed.relevance_threshold"); siteThreshold != nil {
		if t, ok := siteThreshold.(float64); ok && t > 0 {
			threshold = t
		}
	}

	// Find scores in collected_data
	scoresField := "scores"
	if sf, ok := params.StepConfig.Config["scores_field"].(string); ok && sf != "" {
		scoresField = sf
	}

	// execute_llm_prompt returns {result: <parsed>, type: "json"}
	// So scores are at <scoresField>.result
	scoresRaw := datahelpers.ExtractNestedField(params.CollectedData, scoresField+".result")
	if scoresRaw == nil {
		// Fallback: try the field directly (in case caller puts scores there directly)
		scoresRaw = datahelpers.ExtractNestedField(params.CollectedData, scoresField)
	}

	scoresArray, ok := scoresRaw.([]interface{})
	if !ok || len(scoresArray) == 0 {
		logger.Warn("ApplyFeedScoresAction: no scores array found",
			zap.String("scores_field", scoresField),
		)
		return map[string]interface{}{
			"site_id":  siteID,
			"applied":  0,
			"relevant": 0,
			"review":   0,
			"rejected": 0,
			"error":    "no scores found in collected_data",
		}, nil
	}

	// Parse and apply each score
	applied, relevant, review, rejected := 0, 0, 0, 0

	for _, raw := range scoresArray {
		scoreMap, ok := raw.(map[string]interface{})
		if !ok {
			logger.Warn("ApplyFeedScoresAction: score entry is not a map, skipping")
			continue
		}

		itemID, _ := scoreMap["id"].(string)
		if itemID == "" {
			continue
		}

		// Parse score — handle both float64 (JSON default) and string
		score := 0.0
		switch v := scoreMap["score"].(type) {
		case float64:
			score = v
		case int:
			score = float64(v)
		case json.Number:
			score, _ = v.Float64()
		}

		reason, _ := scoreMap["reason"].(string)

		// Parse topics
		var topicsJSON []byte
		if topics, ok := scoreMap["topics"].([]interface{}); ok {
			topicsJSON, _ = json.Marshal(topics)
		} else {
			topicsJSON = []byte("[]")
		}

		// Parse credibility fields
		credibility, _ := scoreMap["credibility"].(string)
		credibilityReason, _ := scoreMap["credibility_reason"].(string)

		// Parse source_attribution — store as JSONB
		var attributionJSON []byte
		if attr, ok := scoreMap["source_attribution"].(map[string]interface{}); ok {
			attributionJSON, _ = json.Marshal(attr)
		} else {
			attributionJSON = []byte("{}")
		}

		// Check for explicit flag (values/legal conflict, low credibility)
		flagged, _ := scoreMap["flagged"].(bool)

		// Determine status
		status := "review"
		if flagged {
			status = "rejected"
			rejected++
		} else if score >= threshold {
			status = "relevant"
			relevant++
		} else if score < 20 {
			status = "rejected"
			rejected++
		} else {
			review++
		}

		// Update the row with relevance + credibility + attribution
		result, err := params.DB.ExecContext(ctx, `
			UPDATE content_feed_items 
			SET relevance_score = $1,
			    status = $2,
			    topics = $3,
			    processed_at = $4,
			    credibility = $5,
			    credibility_reason = $6,
			    source_attribution = $7
			WHERE id = $8 
			  AND site_id = $9
			  AND status = 'ingested'
		`, score, status, topicsJSON, time.Now().UTC(),
			nullIfEmpty(credibility), nullIfEmpty(credibilityReason), attributionJSON,
			itemID, siteID)

		if err != nil {
			logger.Warn("ApplyFeedScoresAction: failed to update item",
				zap.String("item_id", itemID),
				zap.Error(err),
			)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			applied++
			logger.Info("ApplyFeedScoresAction: scored item",
				zap.String("item_id", itemID),
				zap.Float64("score", score),
				zap.String("status", status),
				zap.String("credibility", credibility),
				zap.String("reason", datahelpers.TruncateString(reason, 80)),
			)
		}
	}

	logger.Info("ApplyFeedScoresAction: complete",
		zap.Int("applied", applied),
		zap.Int("relevant", relevant),
		zap.Int("review", review),
		zap.Int("rejected", rejected),
		zap.Float64("threshold", threshold),
	)

	return map[string]interface{}{
		"site_id":   siteID,
		"applied":   applied,
		"relevant":  relevant,
		"review":    review,
		"rejected":  rejected,
		"threshold": threshold,
	}, nil
}

// ---------------------------------------------------------------------------
// LoadFeedItemsForTriageAction
// ---------------------------------------------------------------------------
//
// Loads unscored (ingested) items for a site. Returns them in a format
// suitable for the triage LLM prompt. Also loads source metadata so the
// prompt can factor in source reputation.

var LoadFeedItemsForTriageInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"max_items", "min_age_minutes"},
}

// LoadFeedItemsForTriageAction queries content_feed_items WHERE status = 'ingested'
// and returns them formatted for the triage prompt.
//
// Returns:
//
//	{
//	    "items": [{id, source_title, source_summary, source_url, published_at, source_name, source_type}...],
//	    "count": N,
//	    "site_id": "..."
//	}
func LoadFeedItemsForTriageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_feed_items_for_triage"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadFeedItemsForTriageInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID := inputs.Get("site_id")
	if _, err := uuid.Parse(siteID); err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	maxItems := inputs.GetInt("max_items", 50)

	// Query items with source info joined
	rows, err := params.DB.QueryContext(ctx, `
		SELECT 
			cfi.id::text,
			COALESCE(cfi.source_title, '') as source_title,
			COALESCE(cfi.source_summary, '') as source_summary,
			COALESCE(cfi.source_url, '') as source_url,
			COALESCE(cfi.source_published_at::text, '') as published_at,
			COALESCE(cs.name, 'unknown') as source_name,
			COALESCE(cs.source_type, 'unknown') as source_type
		FROM content_feed_items cfi
		LEFT JOIN content_sources cs ON cs.id = cfi.source_id
		WHERE cfi.site_id = $1 
		  AND cfi.status = 'ingested'
		ORDER BY cfi.created_at DESC
		LIMIT $2
	`, siteID, maxItems)
	if err != nil {
		return nil, fmt.Errorf("query feed items: %w", err)
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, title, summary, url, publishedAt, sourceName, sourceType string
		if err := rows.Scan(&id, &title, &summary, &url, &publishedAt, &sourceName, &sourceType); err != nil {
			logger.Warn("LoadFeedItemsForTriage: scan error", zap.Error(err))
			continue
		}
		items = append(items, map[string]interface{}{
			"id":             id,
			"source_title":   title,
			"source_summary": summary,
			"source_url":     url,
			"published_at":   publishedAt,
			"source_name":    sourceName,
			"source_type":    sourceType,
		})
	}

	logger.Info("LoadFeedItemsForTriage: loaded items",
		zap.Int("count", len(items)),
		zap.String("site_id", siteID),
	)

	return map[string]interface{}{
		"items":   items,
		"count":   len(items),
		"site_id": siteID,
	}, nil
}
