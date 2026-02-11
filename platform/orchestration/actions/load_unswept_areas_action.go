// FILE: platform/orchestration/actions/load_unswept_areas.go
//
// LoadUnsweptAreasAction loads postcode districts that haven't been swept
// yet (or least recently swept) from business_intel.search_areas.
//
// This is the "load" action for the area-sweep-orchestrator agent.
// Agents own their domain — this agent loads its own data.
//
// Workflow config:
//
//	"load_areas": {
//	    "action": "load_unswept_areas",
//	    "config": {
//	        "input_fields": ["limit", "country", "business_type", "area_code"]
//	    },
//	    "output_field": "unswept_areas",
//	    "next_step": "dispatch_discoverers"
//	}
//
// Registration:
//   "load_unswept_areas": LoadUnsweptAreasAction,

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadUnsweptAreasInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"country", "business_type"},
	Optional: []string{"limit", "area_code"},
	Defaults: map[string]interface{}{
		"limit": 50,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_unswept_areas", LoadUnsweptAreasInputSpec)
}

func LoadUnsweptAreasAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadUnsweptAreasAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadUnsweptAreasInputSpec,
		params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	limit := inputs.GetInt("limit", 50)
	country := inputs.Get("country")
	businessType := inputs.Get("business_type")
	areaCode := inputs.Get("area_code") // optional filter e.g. "BT" for Belfast only

	/*if country == "" {
		country = "GB"
	}
	if businessType == "" {
		businessType = "veterinary practice"
	}*/

	params.Logger.Info("LoadUnsweptAreasAction: querying",
		zap.Int("limit", limit),
		zap.String("country", country),
		zap.String("area_code", areaCode))

	// Build query — optionally filter by area_code
	// limit <= 0 means "all matching rows"
	query := `
		SELECT id, district_code, area_name
		FROM business_intel.search_areas
		WHERE country = $1
	`
	args := []interface{}{country}
	argIdx := 2

	if areaCode != "" {
		query += fmt.Sprintf(` AND area_code = $%d`, argIdx)
		args = append(args, areaCode)
		argIdx++
	}

	query += ` ORDER BY sweep_count ASC, last_swept_at ASC NULLS FIRST, district_code ASC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIdx)
		args = append(args, limit)
	}

	rows, err := params.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query search_areas: %w", err)
	}
	defer rows.Close()

	var areas []map[string]interface{}
	for rows.Next() {
		var id, districtCode, areaName string
		if err := rows.Scan(&id, &districtCode, &areaName); err != nil {
			params.Logger.Warn("Failed to scan row", zap.Error(err))
			continue
		}
		areas = append(areas, map[string]interface{}{
			"search_area_id": id,
			"district_code":  districtCode,
			"area_name":      areaName,
		})
	}

	// Summary stats
	var totalAreas, neverSwept, totalCandidates int
	params.DB.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE sweep_count = 0),
			COALESCE(SUM(candidates_found), 0)
		FROM business_intel.search_areas
		WHERE country = $1
	`, country).Scan(&totalAreas, &neverSwept, &totalCandidates)

	params.Logger.Info("LoadUnsweptAreasAction: loaded",
		zap.Int("areas_to_sweep", len(areas)),
		zap.Int("total_areas", totalAreas),
		zap.Int("never_swept", neverSwept))

	return map[string]interface{}{
		"areas":            areas,
		"count":            len(areas),
		"country":          country,
		"business_type":    businessType,
		"total_areas":      totalAreas,
		"never_swept":      neverSwept,
		"total_candidates": totalCandidates,
		"loaded_at":        time.Now().UTC().Format(time.RFC3339),
	}, nil
}
