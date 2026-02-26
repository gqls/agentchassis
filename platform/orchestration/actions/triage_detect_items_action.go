// FILE: platform/orchestration/actions/triage_detected_items_action.go
//
// TriageDetectedItemsAction promotes discovery findings from status='detected'
// to status='triaged' with domain='build' so the dispatch loop picks them up.
//
// Discovery agents write items with different domains (content, design) and
// status='detected'. The dispatch loop filters item_domain='build' and
// status IN ('triaged', 'approved'). This action bridges the gap.
//
// Used by: improvement-loop agent (after all discovery agents complete)
//
// Config (literals):
//   - target_domain: string — domain to set on promoted items (default: "build")
//   - batch_id:      string (path) — optional, only promote items from this batch
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — which site's items to promote

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var TriageDetectedItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("triage_detected_items", TriageDetectedItemsInputSpec)
}

// ============================================================================
// ACTION: triage_detected_items
// ============================================================================

func TriageDetectedItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "triage_detected_items"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("TriageDetectedItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Resolve site_id ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		TriageDetectedItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// --- Config ---
	config := params.StepConfig.Config
	targetDomain := "build"
	if td, ok := config["target_domain"].(string); ok && td != "" {
		targetDomain = td
	}

	// --- Promote detected → triaged ---
	// Sets domain to targetDomain so the dispatch loop (which filters item_domain='build')
	// will pick them up. Preserves the original domain in spec.original_domain for auditing.
	result, err := params.DB.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'triaged',
		    triaged_at = now(),
		    spec = jsonb_set(
		        COALESCE(spec, '{}'::jsonb),
		        '{original_domain}',
		        to_jsonb(domain)
		    ),
		    domain = $2
		WHERE site_id = $1
		  AND status = 'detected'
	`, siteID, targetDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to promote detected items: %w", err)
	}

	promoted, _ := result.RowsAffected()

	logger.Info("TriageDetectedItemsAction: Complete",
		zap.String("site_id", siteIDStr),
		zap.Int64("promoted", promoted),
		zap.String("target_domain", targetDomain),
	)

	return map[string]interface{}{
		"site_id":       siteIDStr,
		"promoted":      promoted,
		"target_domain": targetDomain,
		"has_items":     promoted > 0,
	}, nil
}
