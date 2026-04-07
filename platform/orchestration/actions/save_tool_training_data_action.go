// FILE: platform/orchestration/actions/save_tool_training_data_action.go
//
// Saves the (source_html, functional_spec, recreated_html) triple as a
// research_result for training data export. Each successful tool recreation
// produces one training example that can be exported for fine-tuning
// open-weight code generation models.
//
// Workflow placement: after check_completeness, before validate_tool.
// Only saves when the completeness check passes (has_marker = true).
//
// The stored data includes:
//   - source_html: the original crawled rawHtml
//   - functional_spec: the analyze_tool JSON output
//   - recreated_html: the clean LLM-generated HTML
//   - metadata: model used, tokens, domain, page_name, vertical
//
// Export query for training:
//   SELECT
//     data->'functional_spec' as input,
//     data->>'recreated_html' as output,
//     data->'metadata'->>'model' as model,
//     data->'metadata'->>'vertical' as vertical
//   FROM research_results
//   WHERE result_type = 'tool_recreation_training'
//     AND data->'metadata'->>'complete' = 'true'
//   ORDER BY created_at;
//
// Registration (add to registry.go):
//   "save_tool_training_data": {
//       Handler:     SaveToolTrainingDataAction,
//       Category:    "training",
//       Description: "Save tool recreation triple for model fine-tuning",
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

var SaveToolTrainingDataInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("save_tool_training_data", SaveToolTrainingDataInputSpec)
}

func SaveToolTrainingDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "save_tool_training_data"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return map[string]interface{}{"saved": false, "reason": "no_db"}, nil
	}

	// Extract site and page context
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return map[string]interface{}{"saved": false, "reason": "invalid_site_id"}, nil
	}

	pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "page_record.id")
	pageID, _ := uuid.Parse(pageIDStr)

	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain")
	pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "page_record.name")
	vertical := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_specs.specs.identity.industry")

	// Extract the three pieces of the training triple
	sourceHTML := datahelpers.ExtractNestedFieldString(params.CollectedData, "existing_content.existing_content.raw_html")
	sourceMarkdown := datahelpers.ExtractNestedFieldString(params.CollectedData, "existing_content.raw_markdown")

	// Functional spec — this is the analyze_tool JSON output
	var functionalSpec interface{}
	specRaw := datahelpers.ExtractNestedField(params.CollectedData, "tool_analysis.result")
	if specRaw != nil {
		functionalSpec = specRaw
	}

	// Recreated HTML — the clean output after completeness check
	recreatedHTML := datahelpers.ExtractNestedFieldString(params.CollectedData, "completeness_check.clean_html")
	if recreatedHTML == "" {
		// Fallback to raw output
		recreatedHTML = datahelpers.ExtractNestedFieldString(params.CollectedData, "tool_recreation.result")
	}

	// Completeness info
	isComplete := false
	if completeRaw := datahelpers.ExtractNestedField(params.CollectedData, "completeness_check.complete"); completeRaw != nil {
		if b, ok := completeRaw.(bool); ok {
			isComplete = b
		}
	}

	// Get the model that was used (from the agent definition config)
	model := datahelpers.ExtractNestedFieldString(params.CollectedData, "tool_recreation.type")
	if model == "" {
		model = "unknown" // Will be in llm_call_log anyway
	}

	// Only save if we have at least the functional spec and recreated HTML
	if functionalSpec == nil && recreatedHTML == "" {
		logger.Info("SaveToolTrainingData: insufficient data, skipping",
			zap.String("page_name", pageName))
		return map[string]interface{}{"saved": false, "reason": "insufficient_data"}, nil
	}

	// Build the training data record
	trainingData := map[string]interface{}{
		"source_html":     sourceHTML,
		"source_markdown": sourceMarkdown,
		"functional_spec": functionalSpec,
		"recreated_html":  recreatedHTML,
		"metadata": map[string]interface{}{
			"domain":    domain,
			"page_name": pageName,
			"vertical":  vertical,
			"model":     model,
			"complete":  isComplete,
			"source":    "tool-recreation-handler",
		},
	}

	dataJSON, err := json.Marshal(trainingData)
	if err != nil {
		logger.Warn("SaveToolTrainingData: marshal failed", zap.Error(err))
		return map[string]interface{}{"saved": false, "reason": "marshal_error"}, nil
	}

	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO research_results (
			site_id, page_id, query, topic, result_type,
			data, summary,
			researched_by, research_agent_type
		) VALUES (
			$1, $2, $3, 'tool_recreation_training', 'tool_recreation_training',
			$4::jsonb, $5,
			'tool-recreation-handler', 'tool-recreation-handler'
		)
	`, siteID, pageID,
		fmt.Sprintf("Tool recreation training data: %s on %s", pageName, domain),
		string(dataJSON),
		fmt.Sprintf("Training triple for %s: source(%d) + spec + recreation(%d) complete=%v",
			pageName, len(sourceHTML), len(recreatedHTML), isComplete),
	)

	if err != nil {
		logger.Warn("SaveToolTrainingData: insert failed", zap.Error(err))
		return map[string]interface{}{"saved": false, "reason": err.Error()}, nil
	}

	logger.Info("SaveToolTrainingData: saved",
		zap.String("domain", domain),
		zap.String("page_name", pageName),
		zap.String("vertical", vertical),
		zap.Bool("complete", isComplete),
		zap.Int("source_html_len", len(sourceHTML)),
		zap.Int("recreated_html_len", len(recreatedHTML)),
	)

	return map[string]interface{}{
		"saved":    true,
		"complete": isComplete,
		"domain":   domain,
		"page":     pageName,
		"vertical": vertical,
	}, nil
}
