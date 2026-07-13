// FILE: platform/orchestration/actions/promote_candidates.go
//
// PromoteCandidatesAction moves discovery_candidates with status='pending'
// into business_intel.businesses with verification_status='pending'.
//
// Deduplicates by website_url against existing businesses.
// Dismisses candidates without a website_url (they need enrichment first).
// Sets candidate status to 'promoted' and links back via promoted_business_id.
//
// This is the "promote" step in the vet-pipeline-orchestrator workflow.
//
// Workflow config:
//
//	"promote": {
//	    "action": "promote_candidates",
//	    "config": {
//	        "input_fields": ["promote_limit", "vertical_slug"]
//	    },
//	    "output_field": "promote_result",
//	    "next_step": "dispatch_verifiers"
//	}
//
// Registration:
//   "promote_candidates": PromoteCandidatesAction,

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var PromoteCandidatesInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"vertical_slug", "business_type"},
	Optional: []string{"promote_limit", "country"},
	Defaults: map[string]interface{}{
		"promote_limit": 500,
		"country":       "GB",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("promote_candidates", PromoteCandidatesInputSpec)
}

func PromoteCandidatesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PromoteCandidatesAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		PromoteCandidatesInputSpec,
		params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	limit := inputs.GetInt("promote_limit", 500)
	verticalSlug := inputs.Get("vertical_slug")
	businessType := inputs.Get("business_type")
	country := inputs.Get("country")
	if country == "" {
		country = "GB"
	}

	// Look up vertical_id for the given slug
	var verticalID string
	err = params.DB.QueryRowContext(ctx,
		`SELECT id FROM business_intel.business_verticals WHERE slug = $1`,
		verticalSlug,
	).Scan(&verticalID)
	if err != nil {
		return nil, fmt.Errorf("failed to find vertical '%s': %w", verticalSlug, err)
	}

	// Load pending candidates that have a website_url
	rows, err := params.DB.QueryContext(ctx, `
		SELECT id, name, website_url, address_snippet, postcode,
		       phone, detected_group, is_independent
		FROM business_intel.discovery_candidates
		WHERE status = 'pending'
		  AND website_url IS NOT NULL
		  AND website_url != ''
		ORDER BY created_at ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query candidates: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		ID             string
		Name           string
		WebsiteURL     string
		AddressSnippet sql.NullString
		Postcode       sql.NullString
		Phone          sql.NullString
		DetectedGroup  sql.NullString
		IsIndependent  sql.NullBool
	}

	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.ID, &c.Name, &c.WebsiteURL,
			&c.AddressSnippet, &c.Postcode, &c.Phone,
			&c.DetectedGroup, &c.IsIndependent); err != nil {
			params.Logger.Warn("PromoteCandidatesAction: scan error", zap.Error(err))
			continue
		}
		candidates = append(candidates, c)
	}

	// Also mark candidates without website_url as needs_enrichment
	needsEnrichmentRes, err := params.DB.ExecContext(ctx, `
		UPDATE business_intel.discovery_candidates
		SET status = 'needs_enrichment', reviewed_at = NOW()
		WHERE status = 'pending'
		  AND (website_url IS NULL OR website_url = '')`)
	needsEnrichment := int64(0)
	if err == nil {
		needsEnrichment, _ = needsEnrichmentRes.RowsAffected()
	}

	params.Logger.Info("PromoteCandidatesAction: loaded candidates",
		zap.Int("with_url", len(candidates)),
		zap.Int64("needs_enrichment", needsEnrichment))

	promoted := 0
	dismissed := 0
	dupErrors := 0

	for _, c := range candidates {
		// Deduplicate: check if a business already has this website_url
		// Use extractRootURL (shared from scan_discovery_candidates.go) for
		// consistent domain-level matching with process_area_sweep
		rootURL := extractRootURL(c.WebsiteURL)

		var existingID string
		err := params.DB.QueryRowContext(ctx, `
			SELECT id FROM business_intel.businesses
			WHERE website_url ILIKE $1
			   OR website_url ILIKE $2
			LIMIT 1`,
			rootURL+"%",
			"www."+strings.TrimPrefix(rootURL, "https://")+"%",
		).Scan(&existingID)

		if err != nil && err != sql.ErrNoRows {
			params.Logger.Warn("PromoteCandidatesAction: DB error checking dup",
				zap.String("candidate_id", c.ID),
				zap.String("url", c.WebsiteURL),
				zap.Error(err))
			dupErrors++
			continue
		}

		if err == nil {
			// Already exists — dismiss candidate, link to matched business
			_, _ = params.DB.ExecContext(ctx, `
				UPDATE business_intel.discovery_candidates
				SET status = 'matched',
				    matched_business_id = $2,
				    match_method = 'website_url',
				    match_confidence = 0.9,
				    reviewed_at = NOW()
				WHERE id = $1`, c.ID, existingID)
			dismissed++
			continue
		}

		// Insert new business + create collection_task + mark candidate
		// All in one transaction to avoid prepared statement conflicts
		var isIndep interface{}
		if c.IsIndependent.Valid {
			isIndep = c.IsIndependent.Bool
		}

		tx, txErr := params.DB.BeginTx(ctx, nil)
		if txErr != nil {
			params.Logger.Warn("PromoteCandidatesAction: tx begin failed",
				zap.String("candidate_id", c.ID),
				zap.Error(txErr))
			dupErrors++
			continue
		}

		var newBusinessID string
		err = tx.QueryRowContext(ctx, `
			INSERT INTO business_intel.businesses
				(name, website_url, postcode, phone,
				 group_name, is_independent,
				 vertical_id, business_type,
				 verification_status, is_active, country,
				 created_at, updated_at)
			VALUES ($1, $2, $3, $4,
				$5, $6,
				$7, $8,
				'pending', TRUE, $9,
				NOW(), NOW())
			RETURNING id`,
			c.Name, c.WebsiteURL,
			nullIfEmpty(c.Postcode.String), nullIfEmpty(c.Phone.String),
			nullIfEmpty(c.DetectedGroup.String), isIndep,
			verticalID, businessType, country,
		).Scan(&newBusinessID)

		if err != nil {
			tx.Rollback()
			params.Logger.Warn("PromoteCandidatesAction: insert failed",
				zap.String("candidate_id", c.ID),
				zap.String("name", c.Name),
				zap.Error(err))
			dupErrors++
			continue
		}

		// Create collection_task so batch-processor picks this up
		_, err = tx.ExecContext(ctx, `
			INSERT INTO business_intel.collection_tasks
				(business_id, task_type, vertical_id, priority, status,
				 created_at, updated_at)
			VALUES ($1, 'initial_verification', $2, 5, 'pending', NOW(), NOW())
			ON CONFLICT DO NOTHING`,
			newBusinessID, verticalID)
		if err != nil {
			tx.Rollback()
			params.Logger.Warn("PromoteCandidatesAction: task insert failed",
				zap.String("business_id", newBusinessID),
				zap.Error(err))
			dupErrors++
			continue
		}

		// Mark candidate as promoted
		_, err = tx.ExecContext(ctx, `
			UPDATE business_intel.discovery_candidates
			SET status = 'promoted',
			    promoted_business_id = $2,
			    reviewed_at = NOW()
			WHERE id = $1`, c.ID, newBusinessID)
		if err != nil {
			tx.Rollback()
			params.Logger.Warn("PromoteCandidatesAction: candidate update failed",
				zap.String("candidate_id", c.ID),
				zap.Error(err))
			dupErrors++
			continue
		}

		if err := tx.Commit(); err != nil {
			params.Logger.Warn("PromoteCandidatesAction: commit failed",
				zap.String("candidate_id", c.ID),
				zap.Error(err))
			dupErrors++
			continue
		}

		promoted++

		params.Logger.Info("PromoteCandidatesAction: promoted",
			zap.String("business_id", newBusinessID),
			zap.String("name", c.Name),
			zap.String("url", c.WebsiteURL))
	}

	params.Logger.Info("PromoteCandidatesAction: complete",
		zap.Int("promoted", promoted),
		zap.Int("dismissed_matched", dismissed),
		zap.Int64("needs_enrichment", needsEnrichment),
		zap.Int("errors", dupErrors))

	return map[string]interface{}{
		"promoted":         promoted,
		"dismissed":        dismissed,
		"needs_enrichment": needsEnrichment,
		"errors":           dupErrors,
		"total_checked":    len(candidates),
		"vertical_slug":    verticalSlug,
		"completed_at":     time.Now().UTC().Format(time.RFC3339),
	}, nil
}
