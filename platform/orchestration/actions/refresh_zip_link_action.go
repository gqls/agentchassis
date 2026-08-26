// FILE: platform/orchestration/actions/refresh_zip_link_action.go
//
// The ZIP link refresher's write half: re-stamp the stored presign on a site's
// LIVE zip_download tokens, so the customer's durable /d/ link keeps working
// past the presign's 7-day SigV4 ceiling (the token lasts LiveLinkWindow; the
// URL inside it does not — DGH-018, DECISION_2026-08-21b step 3).
//
// The presign itself comes from zip-deliverer, the only process allowed to mint
// one (bugs_open/245): the zip-link-refresher agent (seed 652) runs the proven
// spawn->call pair and hands this action the fresh URL. This action is pure DB.
//
// Workflow config (the zip-link-refresher seed):
//
//   "refresh": {
//     "action": "refresh_zip_link",
//     "config": {
//       "site_id":        "input_data.site_id",
//       "presigned_url":  "zip_result.presigned_url",
//       "expiry_minutes": "zip_result.expiry_minutes"
//     },
//     "output_field": "refresh_result",
//     "next_step": "complete"
//   }

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

var RefreshZipLinkInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "presigned_url"},
	Optional: []string{"expiry_minutes"},
	Defaults: map[string]interface{}{},
}

// RefreshZipLinkAction re-stamps stored_url on every LIVE zip_download token of
// one site. One statement; the WHERE is the safety:
//
//   - purpose = 'zip_download'  — a confirm token must never grow a URL;
//   - revoked_at IS NULL        — revocation is permanent, a refresh must not
//     resurrect a link an operator killed;
//   - expires_at > now          — a dead TOKEN stays dead; refreshing the URL
//     inside an expired token would quietly extend a customer window the
//     handover stamp closed.
//
// ZERO rows updated is a WARN, not an error: the scheduler's pre_query selected
// this site while a token was live, and the token can legitimately expire in
// the gap. Failing the run would make a benign race look like a broken
// refresher every time it happens.
func RefreshZipLinkAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "refresh_zip_link"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, RefreshZipLinkInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}
	presignedURL := inputs.Get("presigned_url")
	if presignedURL == "" {
		// An empty URL would "refresh" every live token into the stale state —
		// the exact link-born-broken shape MintZipToken refuses at mint.
		return nil, fmt.Errorf("presigned_url resolved empty: refusing to blank the stored URL on live tokens")
	}

	mins := 60 * 24 * 7 // the SigV4 ceiling, zip-deliverer's own default
	if m := inputs.Get("expiry_minutes"); m != "" {
		if _, err := fmt.Sscanf(m, "%d", &mins); err != nil {
			return nil, fmt.Errorf("expiry_minutes %q is not a number: %w", m, err)
		}
	}
	now := time.Now().UTC()

	res, err := params.DB.ExecContext(ctx, `
		UPDATE customer_access_tokens
		   SET stored_url            = $2,
		       stored_url_expires_at = $3
		 WHERE site_id    = $1
		   AND purpose    = 'zip_download'
		   AND revoked_at IS NULL
		   AND expires_at > $4
	`, siteID, presignedURL, now.Add(time.Duration(mins)*time.Minute), now)
	if err != nil {
		return nil, fmt.Errorf("refresh zip link: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		logger.Warn("zip link refresh matched no live tokens (benign if the token expired since the pre_query; investigate if it recurs for one site)",
			zap.String("site_id", siteID.String()))
	} else {
		logger.Info("zip link refreshed", zap.String("site_id", siteID.String()), zap.Int64("tokens", n))
	}
	return map[string]interface{}{
		"refreshed":         n,
		"stored_url_expiry": now.Add(time.Duration(mins) * time.Minute).Format(time.RFC3339),
	}, nil
}
