// FILE: platform/orchestration/actions/store_crawl_batch_action.go
//
// Stores a batch of scraped page results into research_results, one row per page.
// Used in the paginated crawl loop — each batch iteration scrapes 5 pages and
// stores them here. Later steps read from research_results instead of collected_data.
//
// Workflow config (inside scrape_pages_loop sub_workflow):
//   "store_batch": {
//       "action": "store_crawl_batch",
//       "config": {
//           "site_id": "site_record.site_id",
//           "batch_field": "batch_result.results"
//       },
//       "output_field": "batch_stored"
//   }
//
// Stores each page as result_type = 'adoption_crawl_page' with data containing:
//   {"url": "...", "markdown": "...", "raw_html": "...", "metadata": {...}}
//
// Registration:
//   "store_crawl_batch": {
//       Handler:     StoreCrawlBatchAction,
//       Category:    "site",
//       Description: "Store batch of scraped pages to research_results for paginated crawl",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var StoreCrawlBatchInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"batch_field"},
	Defaults:   map[string]interface{}{"batch_field": "batch_result.results"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("store_crawl_batch", StoreCrawlBatchInputSpec)
}

func StoreCrawlBatchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "store_crawl_batch"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Extract site_id
	siteIDPath := "site_record.site_id"
	if p, ok := config["site_id"].(string); ok && p != "" {
		siteIDPath = p
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDPath)
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id at %s: %w", siteIDPath, err)
	}

	// Extract batch results
	batchField := "batch_result.results"
	if f, ok := config["batch_field"].(string); ok && f != "" {
		batchField = f
	}

	batchRaw := datahelpers.ExtractNestedField(params.CollectedData, batchField)
	if batchRaw == nil {
		logger.Warn("StoreCrawlBatch: no batch results found",
			zap.String("batch_field", batchField))
		return map[string]interface{}{"stored": 0, "reason": "no_batch_data"}, nil
	}

	batchArray, ok := batchRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("batch results at %s is not an array (got %T)", batchField, batchRaw)
	}

	stored := 0
	skipped := 0

	for _, pageRaw := range batchArray {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			skipped++
			continue
		}

		// Skip failed scrapes
		if success, ok := page["success"].(bool); ok && !success {
			skipped++
			continue
		}

		pageURL, _ := page["url"].(string)
		if pageURL == "" {
			skipped++
			continue
		}

		// Build the stored data — same structure as adoption_page
		// so format_crawl_for_analysis and buildCrawlPageIndex can read it
		pageData := map[string]interface{}{
			"url":      pageURL,
			"markdown": page["markdown"],
			"raw_html": page["raw_html"],
			"metadata": page["metadata"],
			"title":    page["title"],
		}

		dataJSON, err := json.Marshal(pageData)
		if err != nil {
			logger.Warn("StoreCrawlBatch: marshal failed",
				zap.String("url", pageURL), zap.Error(err))
			skipped++
			continue
		}

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO research_results (
				site_id, query, topic, result_type,
				data, summary,
				researched_by, research_agent_type
			) VALUES (
				$1, $2, 'adoption_crawl_page', 'adoption_crawl_page',
				$3::jsonb, $4,
				'site-adoption-agent', 'site-adoption-agent'
			)
		`, siteID, pageURL, string(dataJSON),
			fmt.Sprintf("Crawled page: %s", pageURL),
		)

		if err != nil {
			logger.Warn("StoreCrawlBatch: insert failed",
				zap.String("url", pageURL), zap.Error(err))
			skipped++
			continue
		}

		stored++
	}

	logger.Info("StoreCrawlBatch: complete",
		zap.Int("stored", stored),
		zap.Int("skipped", skipped),
		zap.Int("batch_size", len(batchArray)),
	)

	return map[string]interface{}{
		"stored":  stored,
		"skipped": skipped,
		"total":   len(batchArray),
	}, nil
}
