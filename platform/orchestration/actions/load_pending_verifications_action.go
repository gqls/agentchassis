// FILE: platform/orchestration/actions/load_pending_verifications.go
//
// LoadPendingVerificationsAction queries businesses with verification_status
// IN ('pending', 'seed_import') and returns them as a list for loop iteration.
//
// This replaces the dispatch_verifiers pattern — instead of fire-and-forget
// messages, the pipeline loop calls a spawned verifier agent per business.
//
// Workflow config:
//
//	"load_pending_businesses": {
//	    "action": "load_pending_verifications",
//	    "config": {
//	        "verify_limit": 100,
//	        "vertical_slug": "veterinary"
//	    },
//	    "output_field": "pending_businesses",
//	    "next_step": "check_pending"
//	}
//
// Output shape:
//
//	{
//	    "businesses": [{"id": "uuid", "name": "..."}, ...],
//	    "count": 42,
//	    "vertical_slug": "veterinary"
//	}
//
// Registration:
//   "load_pending_verifications": LoadPendingVerificationsAction,

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadPendingVerificationsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"verify_limit", "vertical_slug"},
	Defaults: map[string]interface{}{
		"verify_limit":  100,
		"vertical_slug": "veterinary",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_pending_verifications", LoadPendingVerificationsInputSpec)
}

func LoadPendingVerificationsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadPendingVerificationsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadPendingVerificationsInputSpec,
		params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	limit := inputs.GetInt("verify_limit", 100)
	verticalSlug := inputs.Get("vertical_slug")
	if verticalSlug == "" {
		verticalSlug = "veterinary"
	}

	rows, err := params.DB.QueryContext(ctx, `
		SELECT b.id, b.name
		FROM business_intel.businesses b
		JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
		WHERE bv.slug = $1
		  AND b.verification_status IN ('pending', 'seed_import')
		  AND b.is_active = true
		ORDER BY b.created_at ASC
		LIMIT $2`, verticalSlug, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending businesses: %w", err)
	}
	defer rows.Close()

	var businesses []map[string]interface{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			params.Logger.Warn("LoadPendingVerificationsAction: scan error", zap.Error(err))
			continue
		}
		businesses = append(businesses, map[string]interface{}{
			"id":   id,
			"name": name,
		})
	}

	if businesses == nil {
		businesses = []map[string]interface{}{}
	}

	params.Logger.Info("LoadPendingVerificationsAction: loaded",
		zap.Int("count", len(businesses)),
		zap.String("vertical_slug", verticalSlug))

	return map[string]interface{}{
		"businesses":    businesses,
		"count":         len(businesses),
		"vertical_slug": verticalSlug,
		"loaded_at":     time.Now().UTC().Format(time.RFC3339),
	}, nil
}
