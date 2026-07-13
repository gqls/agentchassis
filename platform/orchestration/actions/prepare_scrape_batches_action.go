// FILE: platform/orchestration/actions/prepare_scrape_batches_action.go
//
// Takes a list of discovered URLs and splits them into batches for
// sequential scraping via the webscrape adapter's batch_scrape handler.
// Each batch stays within Kafka message size limits.
//
// Workflow config:
//   "prepare_batches": {
//       "action": "prepare_scrape_batches",
//       "config": {
//           "urls_field": "discovered_urls.links",
//           "batch_size": 5,
//           "scrape_config": {
//               "formats": ["markdown", "rawHtml"],
//               "only_main_content": false
//           }
//       },
//       "output_field": "batch_plan",
//       "next_step": "scrape_loop"
//   }
//
// Output: {"batches": [["url1","url2",...], ...], "batch_count": N, "total_urls": N}
//
// Registration:
//   "prepare_scrape_batches": {
//       Handler:     PrepareScrapeBatchesAction,
//       Category:    "webscrape",
//       Description: "Split discovered URLs into batches for paginated scraping",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var PrepareScrapeBatchesInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"urls_field", "batch_size"},
	Defaults:   map[string]interface{}{"urls_field": "discovered_urls.links", "batch_size": float64(5)},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("prepare_scrape_batches", PrepareScrapeBatchesInputSpec)
}

func PrepareScrapeBatchesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "prepare_scrape_batches"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get the URLs field path
	urlsField := "discovered_urls.links"
	if f, ok := config["urls_field"].(string); ok && f != "" {
		urlsField = f
	}

	// Get batch size
	batchSize := 5
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	// Extract URLs from collected data
	urlsRaw := datahelpers.ExtractNestedField(params.CollectedData, urlsField)
	if urlsRaw == nil {
		return nil, fmt.Errorf("no URLs found at %s", urlsField)
	}

	urlsArray, ok := urlsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("URLs at %s is not an array (got %T)", urlsField, urlsRaw)
	}

	if len(urlsArray) == 0 {
		logger.Warn("PrepareScrapeBatches: empty URL list")
		return map[string]interface{}{
			"batches":     []interface{}{},
			"batch_count": 0,
			"total_urls":  0,
		}, nil
	}

	// Convert to string URLs and deduplicate
	seen := make(map[string]bool)
	var uniqueURLs []string
	for _, u := range urlsArray {
		urlStr, ok := u.(string)
		if !ok || urlStr == "" {
			continue
		}
		if !seen[urlStr] {
			seen[urlStr] = true
			uniqueURLs = append(uniqueURLs, urlStr)
		}
	}

	// Split into batches
	var batches []interface{}
	for i := 0; i < len(uniqueURLs); i += batchSize {
		end := i + batchSize
		if end > len(uniqueURLs) {
			end = len(uniqueURLs)
		}
		batch := make([]interface{}, 0, end-i)
		for _, url := range uniqueURLs[i:end] {
			batch = append(batch, url)
		}
		batches = append(batches, batch)
	}

	logger.Info("PrepareScrapeBatches: batches prepared",
		zap.Int("total_urls", len(uniqueURLs)),
		zap.Int("batch_size", batchSize),
		zap.Int("batch_count", len(batches)),
	)

	return map[string]interface{}{
		"batches":     batches,
		"batch_count": len(batches),
		"total_urls":  len(uniqueURLs),
		"batch_size":  batchSize,
	}, nil
}
