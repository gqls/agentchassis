// FILE: platform/orchestration/actions/reconcile_superseded_reviews_action.go
//
// bugs_open/056 (regeneration_silently_drops_content_that_tripped_a_blocker,
// mechanism corrected by diagnosis corr b361298a): a page can be parked at
// needs_human_review by a validation blocker (item A, validate_content's
// error_step mark_needs_review), then a SIBLING work item for the same page
// (e.g. a page_rerender raised when an image asset lands) reruns the build
// pipeline; generation is non-deterministic, the fresh copy omits the flagged
// element, passes validation, and deploys. complete_work_item's
// preserve-the-flag guard protects only the flagged item ITSELF, so the
// pending human review is silently bypassed and dangles (fundamentallyai
// model-fine-tuning: parked 2026-07-20 21:27, page deployed 07-21 03:41).
//
// This is the STANDALONE reconciler shape the council named (corr 1d8ef2c0,
// guardian): a separately-scheduled scan over (parked needs_human_review item)
// × (page deployed after the item was parked) pairs — NOT a hook inside the
// shared save path, so the universal save/dispatch surface carries no
// review-domain coupling, and every write path that deploys a page is covered
// by construction (the scan reads deployed state, not a particular pipeline).
//
// It never blocks anything and closes nothing: it writes one loud
// agent_error_log row per pair (REVIEW_SUPERSEDED_BY_PASSING_SAVE, with each
// previously-flagged blocker value marked dropped-not-resolved vs
// still-present against the CURRENTLY-DEPLOYED content) and annotates the
// parked item's result under 'superseded_by_passing_save' — which also makes
// the scan idempotent (annotated items are excluded from later sweeps). The
// item's STATUS is untouched: a human owns the needs_human_review flag;
// whether a parked review should HOLD deploys is an owner policy call this
// reconciler only produces the evidence for (bugs_open/033 — the queue has no
// working drain, so a hard hold would be the 029-class wedge).
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ReconcileSupersededReviewsInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{"max_items", "dry_run"},
	Defaults: map[string]interface{}{
		"max_items": 25,    // cap per sweep while confidence builds (mirrors triage/silent-check)
		"dry_run":   false, // when true: report pairs only — no log rows, no annotations
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("reconcile_superseded_reviews", ReconcileSupersededReviewsInputSpec)
}

// supersededPair is one (parked review item, deployed page) hit from the scan.
type supersededPair struct {
	itemID, itemKey, itemType string
	siteID, pageID, pageName  string
	deployedAt                string
}

// ReconcileSupersededReviewsAction runs one reconciliation sweep.
func ReconcileSupersededReviewsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "reconcile_superseded_reviews"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("reconcile_superseded_reviews: no DB handle")
	}

	maxItems := datahelpers.GetIntField(config, "max_items", 25)
	if maxItems <= 0 {
		maxItems = 25
	}
	dryRun := false
	if d, ok := config["dry_run"].(bool); ok {
		dryRun = d
	}

	// The pair scan. page_id is the first-class join where the item carries it;
	// spec->>'page_name' is the fallback (both live parked items in the 056
	// incident carry page_name; neither carries page_id). "Parked since" is the
	// later of created_at/updated_at because updated_at maintenance is not
	// guaranteed (bugs_open/035). The result-key exclusion is the idempotence:
	// an annotated pair never re-fires.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT wi.id, wi.item_key, wi.item_type, wi.site_id,
		       p.id, p.name, p.deployed_at::text
		FROM site_work_items wi
		JOIN pages p
		  ON p.site_id = wi.site_id
		 AND (p.id = wi.page_id OR (wi.page_id IS NULL AND p.name = wi.spec->>'page_name'))
		WHERE wi.status = 'needs_human_review'
		  AND p.deployed_at IS NOT NULL
		  AND p.deployed_at > GREATEST(wi.created_at, COALESCE(wi.updated_at, wi.created_at))
		  AND NOT (COALESCE(wi.result, '{}'::jsonb) ? 'superseded_by_passing_save')
		ORDER BY p.deployed_at DESC
		LIMIT $1
	`, maxItems)
	if err != nil {
		return nil, fmt.Errorf("pair scan: %w", err)
	}
	var pairs []supersededPair
	for rows.Next() {
		var p supersededPair
		if scanErr := rows.Scan(&p.itemID, &p.itemKey, &p.itemType, &p.siteID, &p.pageID, &p.pageName, &p.deployedAt); scanErr != nil {
			logger.Warn("pair scan row failed", zap.Error(scanErr))
			continue
		}
		pairs = append(pairs, p)
	}
	rows.Close()

	annotated := 0
	pairReports := make([]map[string]interface{}, 0, len(pairs))
	for _, pair := range pairs {
		findings := gatherFlaggedFindings(ctx, params.DB, pair, logger)
		dropped := 0
		for _, f := range findings {
			if present, _ := f["present_in_new_content"].(bool); !present {
				dropped++
			}
		}

		reconContext := map[string]interface{}{
			"page_name":   pair.pageName,
			"page_id":     pair.pageID,
			"deployed_at": pair.deployedAt,
			"parked_item": map[string]string{"id": pair.itemID, "item_key": pair.itemKey, "item_type": pair.itemType},
			"flagged":     findings,
		}
		reconJSON, mErr := json.Marshal(reconContext)
		if mErr != nil {
			reconJSON = []byte("{}")
		}
		pairReports = append(pairReports, reconContext)

		logger.Warn("REVIEW SUPERSEDED — page deployed past a parked needs_human_review",
			zap.String("page_name", pair.pageName),
			zap.String("item_key", pair.itemKey),
			zap.Int("flagged_values", len(findings)),
			zap.Int("flagged_values_dropped", dropped),
			zap.Bool("dry_run", dryRun),
		)
		if dryRun {
			continue
		}

		msg := fmt.Sprintf(
			"page %q deployed at %s while work item %q sat parked at needs_human_review — the pending review has been superseded",
			pair.pageName, pair.deployedAt, pair.itemKey)
		if dropped > 0 {
			msg += fmt.Sprintf("; %d previously-flagged element(s) are ABSENT from the deployed content (dropped, not resolved)", dropped)
		}
		var orchestrationID, agentType interface{}
		if params.ExecutionContext != nil {
			orchestrationID = params.ExecutionContext.OrchestrationID
			agentType = params.ExecutionContext.Sender.AgentType
		}
		if _, iErr := params.DB.ExecContext(ctx, `
			INSERT INTO agent_error_log (
				site_id, work_item_id, orchestration_id, agent_type,
				action, error_message, error_code, severity, context
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
		`, pair.siteID, pair.itemID, orchestrationID, agentType,
			"reconcile_superseded_reviews", msg,
			"REVIEW_SUPERSEDED_BY_PASSING_SAVE", "warning", string(reconJSON),
		); iErr != nil {
			logger.Warn("could not persist supersession record; item left unannotated so a later sweep retries",
				zap.String("item_id", pair.itemID), zap.Error(iErr))
			continue
		}
		if _, uErr := params.DB.ExecContext(ctx, `
			UPDATE site_work_items
			SET result = COALESCE(result, '{}'::jsonb)
			          || jsonb_build_object('superseded_by_passing_save', $2::jsonb)
			WHERE id = $1
		`, pair.itemID, string(reconJSON)); uErr != nil {
			logger.Warn("could not annotate parked item (log row already written; sweep will re-log until this lands)",
				zap.String("item_id", pair.itemID), zap.Error(uErr))
			continue
		}
		annotated++
	}

	return map[string]interface{}{
		"success":     true,
		"pairs_found": len(pairs),
		"annotated":   annotated,
		"dry_run":     dryRun,
		"capped_at":   maxItems,
		"pairs":       pairReports,
	}, nil
}

// gatherFlaggedFindings collects the flagged blocker values for a pair — from
// EVERY CONTENT_VALIDATION_BLOCKER_DETAIL row for the page (bounded, newest
// first; not just the latest — a page can accumulate distinct blocker
// incidents), deduped on type+value — and checks each against the page's
// CURRENTLY-DEPLOYED page_components content. Reading the deployed truth from
// the DB, rather than any pipeline's in-memory content, is what makes the scan
// path-independent.
func gatherFlaggedFindings(ctx context.Context, db *sql.DB, pair supersededPair, logger *zap.Logger) []map[string]interface{} {
	var deployedContent string
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(string_agg(rendered_html, ' '), '') FROM page_components WHERE page_id = $1
	`, pair.pageID).Scan(&deployedContent); err != nil {
		logger.Warn("could not read deployed content; findings will carry no presence verdicts", zap.Error(err))
	}

	rows, err := db.QueryContext(ctx, `
		SELECT context FROM agent_error_log
		WHERE site_id = $1 AND error_code = 'CONTENT_VALIDATION_BLOCKER_DETAIL'
		  AND context->>'page_name' = $2
		ORDER BY occurred_at DESC
		LIMIT 20
	`, pair.siteID, pair.pageName)
	if err != nil {
		logger.Warn("blocker-detail query failed", zap.Error(err))
		return []map[string]interface{}{}
	}
	defer rows.Close()

	merged := []map[string]interface{}{}
	seen := map[string]bool{}
	for rows.Next() {
		var blockerContext []byte
		if rows.Scan(&blockerContext) != nil {
			continue
		}
		for _, f := range flaggedValueFindings(blockerContext, deployedContent) {
			key := fmt.Sprintf("%v|%v", f["type"], f["value"])
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, f)
		}
	}
	return merged
}

// flaggedValueFindings extracts the flagged issue values from one
// CONTENT_VALIDATION_BLOCKER_DETAIL context payload and reports, per value,
// whether it is present in the given content. Empty/absent/unparseable context
// yields an empty slice — the supersession record still gets written, it just
// cannot name the contested elements. Substring matching is conservative: a
// value present in a different, legitimate context reads as still-present
// (the review stays live), never as dropped.
func flaggedValueFindings(blockerContext []byte, content string) []map[string]interface{} {
	findings := []map[string]interface{}{}
	if len(blockerContext) == 0 {
		return findings
	}
	var parsed struct {
		Issues []struct {
			Type     string `json:"type"`
			Value    string `json:"value"`
			Severity string `json:"severity"`
		} `json:"issues"`
	}
	if json.Unmarshal(blockerContext, &parsed) != nil {
		return findings
	}
	lowerContent := strings.ToLower(content)
	for _, issue := range parsed.Issues {
		v := strings.TrimSpace(issue.Value)
		if v == "" {
			continue
		}
		findings = append(findings, map[string]interface{}{
			"type":                   issue.Type,
			"value":                  v,
			"severity":               issue.Severity,
			"present_in_new_content": strings.Contains(lowerContent, strings.ToLower(v)),
		})
	}
	return findings
}
