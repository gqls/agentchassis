// Called after site_specs are populated (e.g. after domain research completes).

// SyncSiteIdentityAction reads identity and briefing from site_specs and
// populates the sites table columns (company_name, tagline, email, phone).
// This ensures loadSiteDataFull and RenderSiteComponentsAction get the data
// without needing their own site_specs fallback logic.
//
// Config:
//   site_id (required, path) — e.g. "site_record.site_id"
//
// Used by: build pipeline, after domain research / briefing agents complete.
// Should be added as a step in the build flow after site_specs are populated.

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var SyncSiteIdentityInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{},
	Defaults: map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("sync_site_identity", SyncSiteIdentityInputSpec)
}

func SyncSiteIdentityAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "sync_site_identity"))
	logger.Info("SyncSiteIdentityAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		SyncSiteIdentityInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Read identity and briefing from site_specs
	rows, err := params.DB.QueryContext(ctx, `
		SELECT aspect, data
		FROM site_specs
		WHERE site_id = $1
		  AND aspect IN ('identity', 'briefing')
		  AND is_current = true
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query site_specs: %w", err)
	}
	defer rows.Close()

	var companyName, tagline, email, phone string

	for rows.Next() {
		var aspect string
		var dataJSON []byte
		if err := rows.Scan(&aspect, &dataJSON); err != nil {
			continue
		}
		var data map[string]interface{}
		if json.Unmarshal(dataJSON, &data) != nil {
			continue
		}

		switch aspect {
		case "identity":
			if v, _ := data["company_name"].(string); v != "" && companyName == "" {
				companyName = v
			}
			if v, _ := data["tagline"].(string); v != "" && tagline == "" {
				tagline = v
			}
			if v, _ := data["industry"].(string); v != "" {
				// Also useful — store in company_name fallback if empty
				if companyName == "" {
					// Don't use industry as company name
				}
			}
			// Extract contact from identity.contact nested object
			if contact, ok := data["contact"].(map[string]interface{}); ok {
				if v, _ := contact["email"].(string); v != "" && email == "" {
					email = v
				}
				if v, _ := contact["phone"].(string); v != "" && phone == "" {
					phone = v
				}
			}

		case "briefing":
			if v, _ := data["company_name"].(string); v != "" && companyName == "" {
				companyName = v
			}
			if v, _ := data["tagline"].(string); v != "" && tagline == "" {
				tagline = v
			}
			if v, _ := data["contact_email"].(string); v != "" && email == "" {
				email = v
			}
			if v, _ := data["contact_phone"].(string); v != "" && phone == "" {
				phone = v
			}
		}
	}

	// Update sites columns — only set non-empty fields, don't overwrite existing
	result, err := params.DB.ExecContext(ctx, `
		UPDATE sites SET
			company_name = CASE WHEN COALESCE(company_name, '') = '' AND $2 != '' THEN $2 ELSE company_name END,
			tagline      = CASE WHEN COALESCE(tagline, '')      = '' AND $3 != '' THEN $3 ELSE tagline END,
			email        = CASE WHEN COALESCE(email, '')        = '' AND $4 != '' THEN $4 ELSE email END,
			phone        = CASE WHEN COALESCE(phone, '')        = '' AND $5 != '' THEN $5 ELSE phone END,
			updated_at   = now()
		WHERE id = $1
	`, siteID, companyName, tagline, email, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to update sites columns: %w", err)
	}

	affected, _ := result.RowsAffected()

	logger.Info("SyncSiteIdentityAction: Complete",
		zap.String("site_id", siteID.String()),
		zap.String("company_name", companyName),
		zap.String("email", email),
		zap.Int64("rows_affected", affected),
	)

	return map[string]interface{}{
		"site_id":      siteID.String(),
		"company_name": companyName,
		"tagline":      tagline,
		"email":        email,
		"phone":        phone,
		"synced":       true,
	}, nil
}
