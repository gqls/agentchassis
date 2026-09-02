// FILE: platform/orchestration/actions/feed_event_extraction_actions.go
//
// bugs_open/427 fix candidate #1 (news_feed_ingestion lane). Extraction step
// downstream of feed-triage: given items already scored 'relevant', an LLM
// step (wired in the feed-triage workflow config, not in this file) decides
// whether any of them confirm a specific, dated real-world event, and the
// resulting candidates are registered as evidence_base facts by
// VerifyAndRegisterCitationsAction (evidence_citations.go) — reused UNCHANGED,
// same idiom as verify_and_register_directory_claims (directory_claims.go).
// This file adds nothing to how a citation is verified; it only supplies the
// two mechanical steps either side of the LLM call:
//
//   load_feed_items_for_event_extraction — loads items not yet considered
//   mark_feed_items_event_extracted — marks them considered, so a non-event
//     article isn't re-sent to the LLM every triage cycle forever
//
// event_extracted_at is deliberately a separate column from processed_at
// (set by apply_feed_scores, meaning "triaged") — conflating the two would
// make a reader of processed_at unable to tell which stage actually ran.
// Migration: docs/agent_docs/sql_for_agents/684_content_feed_items_event_extraction_column.sql.

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// LoadFeedItemsForEventExtractionAction
// ---------------------------------------------------------------------------

var LoadFeedItemsForEventExtractionInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"max_items"},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_feed_items_for_event_extraction", LoadFeedItemsForEventExtractionInputSpec)
	datahelpers.RegisterActionInputSpec("mark_feed_items_event_extracted", MarkFeedItemsEventExtractedInputSpec)
}

// LoadFeedItemsForEventExtractionAction queries content_feed_items WHERE
// status = 'relevant' AND event_extracted_at IS NULL — items feed-triage has
// already scored as relevant to the site, that this pass has not yet
// considered for a dated-event extraction. Mirrors
// LoadFeedItemsForTriageAction's shape (feed_triage_actions.go) so the same
// prompt-authoring conventions apply downstream.
func LoadFeedItemsForEventExtractionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_feed_items_for_event_extraction"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadFeedItemsForEventExtractionInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID := inputs.Get("site_id")
	if _, err := uuid.Parse(siteID); err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	maxItems := datahelpers.GetIntField(params.StepConfig.Config, "max_items", 50)

	rows, err := params.DB.QueryContext(ctx, `
		SELECT
			cfi.id::text,
			COALESCE(cfi.source_title, '') as source_title,
			COALESCE(cfi.source_summary, '') as source_summary,
			COALESCE(cfi.source_content, '') as source_content,
			COALESCE(cfi.source_url, '') as source_url,
			COALESCE(cfi.source_published_at::text, '') as published_at,
			COALESCE(cs.name, 'unknown') as source_name,
			COALESCE(cfi.topics::text, '[]') as topics
		FROM content_feed_items cfi
		LEFT JOIN content_sources cs ON cs.id = cfi.source_id
		WHERE cfi.site_id = $1
		  AND cfi.status = 'relevant'
		  AND cfi.event_extracted_at IS NULL
		ORDER BY cfi.source_published_at DESC NULLS LAST
		LIMIT $2
	`, siteID, maxItems)
	if err != nil {
		return nil, fmt.Errorf("query feed items: %w", err)
	}
	defer rows.Close()

	var items []map[string]interface{}
	var ids []string
	// offered counts rows the CURSOR yielded, against len(items) kept
	// (bugs_open/410's pinned count, loadStoredSections' worked example).
	offered := 0
	for rows.Next() {
		offered++
		var id, title, summary, content, url, publishedAt, sourceName, topics string
		if err := rows.Scan(&id, &title, &summary, &content, &url, &publishedAt, &sourceName, &topics); err != nil {
			// scan-loss:accepted: counted — ScanShortfall below refuses the
			// partial result rather than silently marking fewer items
			// considered (mark_feed_items_event_extracted would then stamp
			// event_extracted_at on a smaller set than the cursor actually
			// offered, and the dropped row would never be reconsidered).
			logger.Warn("LoadFeedItemsForEventExtraction: row scan failed", zap.Error(err))
			continue
		}
		items = append(items, map[string]interface{}{
			"id":             id,
			"source_title":   title,
			"source_summary": summary,
			"source_content": content,
			"source_url":     url,
			"published_at":   publishedAt,
			"source_name":    sourceName,
			"topics":         topics,
		})
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := datahelpers.ScanShortfall(offered, len(items), "load_feed_items_for_event_extraction"); err != nil {
		return nil, err
	}

	logger.Info("LoadFeedItemsForEventExtraction: loaded items",
		zap.Int("count", len(items)),
		zap.String("site_id", siteID),
	)

	return map[string]interface{}{
		"items":   items,
		"ids":     ids,
		"count":   len(items),
		"site_id": siteID,
	}, nil
}

// ---------------------------------------------------------------------------
// MarkFeedItemsEventExtractedAction
// ---------------------------------------------------------------------------

var MarkFeedItemsEventExtractedInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id", "ids"},
}

// MarkFeedItemsEventExtractedAction stamps event_extracted_at = now() on every
// item id considered by the extraction pass, whether or not it yielded a
// candidate fact. This is the idempotency mechanism: without it, every
// triage cycle would re-spend LLM budget re-examining the same non-event
// articles forever. Deliberately does NOT touch `status` — the render path
// and everything else reading content_feed_items.status is unaffected.
func MarkFeedItemsEventExtractedAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "mark_feed_items_event_extracted"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		MarkFeedItemsEventExtractedInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID := inputs.Get("site_id")
	if _, err := uuid.Parse(siteID); err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	idsPath, _ := params.StepConfig.Config["ids"].(string)
	if idsPath == "" {
		return nil, fmt.Errorf("config 'ids' (dot-path to the considered item id array) is required")
	}
	rawIDs := datahelpers.ExtractNestedField(params.CollectedData, idsPath)
	idList, ok := rawIDs.([]interface{})
	if !ok {
		return nil, fmt.Errorf("ids at %q is not an array (got %T)", idsPath, rawIDs)
	}

	var ids []string
	for _, v := range idList {
		if s, ok := v.(string); ok && s != "" {
			ids = append(ids, s)
		}
	}
	if len(ids) == 0 {
		return map[string]interface{}{"site_id": siteID, "marked": 0}, nil
	}

	result, err := params.DB.ExecContext(ctx, `
		UPDATE content_feed_items
		   SET event_extracted_at = $1
		 WHERE site_id = $2
		   AND id = ANY($3::uuid[])
		   AND event_extracted_at IS NULL
	`, time.Now().UTC(), siteID, toPGTextArrayLiteral(ids))
	if err != nil {
		return nil, fmt.Errorf("mark items event-extracted: %w", err)
	}
	marked, _ := result.RowsAffected()

	logger.Info("MarkFeedItemsEventExtracted: complete",
		zap.String("site_id", siteID),
		zap.Int("requested", len(ids)),
		zap.Int64("marked", marked),
	)

	return map[string]interface{}{
		"site_id":   siteID,
		"requested": len(ids),
		"marked":    marked,
	}, nil
}
