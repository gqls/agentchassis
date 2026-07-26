// FILE: platform/orchestration/actions/revalidate_review_queue_action.go
//
// bugs_open/033 — the human-review queue has no drain. 370 items sit at
// needs_human_review (2026-07-25); not one has ever been actioned through the
// admin surface, and the surface has been visible and reachable since
// 2026-07-20. The owner ruled it a QUEUE, not a bin (2026-07-22), and ruled the
// approach: "split it — auto-drain what can be, queue the rest" (2026-07-20).
// This action is the auto-drain half.
//
// THE PROBLEM IT SOLVES IS NOT THE BACKLOG SIZE. It is that 321 of the 370
// parked items describe a page that has been REDEPLOYED since the item was
// filed, and nothing re-checks them — so a ghost and a live finding are
// byte-identical in site_work_items. Measured example: leopardessconsulting.co.uk
// /how-we-work carried two unresolved_cta items parked 2026-07-10 saying the
// hero and call-to-action had no destination for cta_url/secondary_cta_url; the
// page redeployed 2026-07-18 with every one of those fields populated. A human
// working that queue spends their attention on findings about page states that
// stopped existing months ago.
//
// WHAT IT DOES
//
//	For each parked item it looks up a REVALIDATOR by item_type and re-evaluates
//	the ORIGINAL finding against currently-deployed state. Three verdicts:
//
//	  resolved    — the condition is provably no longer true of the live page.
//	                Close it: status='complete', resolution_path='auto:revalidated',
//	                evidence into result.revalidation.
//	  still_holds — the condition is provably still true. NO status change; the
//	                item is stamped so a human reading the queue can see it was
//	                re-confirmed, and when.
//	  unknown     — cannot be determined (no revalidator for the type, no
//	                deployed component, component renders from something other
//	                than content_data). Stamped with the reason; left alone.
//
// WHY AUTO-CLOSING IS SAFE, AND WHY IT IS ALLOWED TO BE WRONG
//
//	Every terminal status is excluded from the idx_swi_dedup predicate, so
//	closing an item RELEASES ITS DEDUP KEY. If a 'resolved' verdict is wrong,
//	the producing check raises it again, fresh and correctly dated. A wrong
//	close costs one re-raise; it cannot lose a finding. That is what makes this
//	a reversible sweep rather than a bulk delete.
//
//	> QUALIFIED 2026-07-25 after the council gate (corr ccba9c51, bug_historian,
//	> medium) pressed on this claim — correctly: it was asserted, not measured,
//	> and it is the whole safety case. Verified since, and it does NOT hold
//	> unconditionally.
//	>
//	> All three producers re-insert with ON CONFLICT DO NOTHING on a
//	> deterministic item_key, so a terminal row does not block a re-raise
//	> (resolve_internal_links_action.go:257 for unresolved_cta; plan_sections'
//	> createDeferredItems for needs_section_data; RunDiscoveryChecks ->
//	> insertWorkItem for required_fields_missing). What does NOT hold is
//	> "the check will run again": all three fire on a PAGE BUILD or a discovery
//	> pass over that site, not on a timer. A page that is never rebuilt never
//	> re-raises, and a wrong close on such a page IS a silent loss.
//	>
//	> Also worth knowing before trusting the re-raise: across the platform's
//	> whole history only 8 items of these three types have EVER reached a
//	> terminal status (7 needs_section_data + 1 required_fields_missing; zero of
//	> the 70 unresolved_cta rows). So the re-raise path has essentially never
//	> been exercised for these types. Treat it as reasoned, not observed.
//	>
//	> The mitigation that does hold unconditionally is the audit trail: every
//	> close records the exact fields it judged populated in result.revalidation
//	> and stamps resolution_path='auto:revalidated', so a wrong close is
//	> individually identifiable and reversible by SQL (see the RUNBOOK) whether
//	> or not anything re-raises.
//
//	The asymmetry is deliberate: 'resolved' demands POSITIVE evidence that the
//	finding no longer holds. Every ambiguity resolves to 'unknown' and stays
//	queued. Guessing in the other direction would be the same failure the whole
//	bug is about.
//
// WHAT IT DELIBERATELY DOES NOT TOUCH
//   - updated_at, on anything it does not close. reconcile_superseded_reviews
//     computes "parked since" as GREATEST(created_at, updated_at); bumping
//     updated_at on every sweep would push that forward and hide genuinely
//     superseded pairs from it. The sweep timestamp lives in
//     result.revalidation.at instead. (General updated_at maintenance is
//     bugs_open/035's.)
//   - cta_names_unknown_destination. That check is mid-flight in the
//     cta_link_integrity workstream / bugs_open/023, which already knows 18 of
//     them are false positives of the excluded-area branch. Two threads on one
//     check is what that costs.
//   - The 78 items carrying an error and a machine handler_agent — failures
//     parked by FailWorkItemAction's status_override branch, which does not
//     increment attempt_count so they neither retry nor age out. Real defect,
//     open owner decision (033 D2), not this sweep's.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RevalidateReviewQueueInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{"site_id", "item_type", "max_items", "dry_run"},
	Defaults: map[string]interface{}{
		"max_items": 50,
		"dry_run":   true, // report-only unless a caller explicitly opts in
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("revalidate_review_queue", RevalidateReviewQueueInputSpec)
}

// Verdicts.
const (
	revalidationResolved   = "resolved"
	revalidationStillHolds = "still_holds"
	revalidationUnknown    = "unknown"
)

// parkedReviewItem is one row of the needs_human_review queue.
type parkedReviewItem struct {
	ID       string
	SiteID   uuid.UUID
	ItemType string
	ItemKey  string
	Spec     map[string]interface{}
}

// revalidationVerdict is one revalidator's answer about one parked item.
type revalidationVerdict struct {
	Verdict  string                 `json:"verdict"`
	Reason   string                 `json:"reason"`
	Evidence map[string]interface{} `json:"evidence,omitempty"`
}

// reviewRevalidator re-evaluates one parked item against currently-deployed state.
type reviewRevalidator func(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict

// reviewRevalidators maps item_type → revalidator. An item_type absent from this
// map yields 'unknown' and is counted in uncovered_types, so the coverage gap is
// reported rather than silently read as "nothing to drain".
//
// All three v1 entries reduce to the same question — are the fields this item
// names now populated on the currently-deployed component? — differing only in
// which spec key carries the field list.
var reviewRevalidators = map[string]reviewRevalidator{
	// spec.missing: ["cta_url", "secondary_cta_url"]
	"unresolved_cta": revalidateNamedFields("missing"),
	// spec.missing_fields: ["headline"]
	"required_fields_missing": revalidateNamedFields("missing_fields"),
	// spec.missing: [{"field":"email","source":"site_specs.identity.email"}, …]
	"needs_section_data": revalidateNamedFields("missing"),
}

func RevalidateReviewQueueAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "revalidate_review_queue"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("revalidate_review_queue: no DB handle")
	}

	config := params.StepConfig.Config
	maxItems := datahelpers.GetIntField(config, "max_items", 50)
	if maxItems <= 0 {
		maxItems = 50
	}
	dryRun := true
	if d, ok := config["dry_run"].(bool); ok {
		dryRun = d
	}
	siteFilter, _ := config["site_id"].(string)
	typeFilter, _ := config["item_type"].(string)

	items, err := loadParkedReviewItems(ctx, params.DB, siteFilter, typeFilter, maxItems)
	if err != nil {
		return nil, err
	}

	sweptAt := time.Now().UTC().Format(time.RFC3339)
	counts := map[string]int{revalidationResolved: 0, revalidationStillHolds: 0, revalidationUnknown: 0}
	uncovered := map[string]int{}
	closed := 0
	reports := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		revalidator, covered := reviewRevalidators[item.ItemType]
		var verdict revalidationVerdict
		if !covered {
			uncovered[item.ItemType]++
			verdict = revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("no revalidator registered for item_type %q", item.ItemType),
			}
		} else {
			verdict = revalidator(ctx, params.DB, item, logger)
		}

		record := map[string]interface{}{
			"verdict":   verdict.Verdict,
			"reason":    verdict.Reason,
			"at":        sweptAt,
			"item_type": item.ItemType,
		}
		if verdict.Evidence != nil {
			record["evidence"] = verdict.Evidence
		}

		counts[verdict.Verdict]++
		reports = append(reports, map[string]interface{}{
			"item_id":   item.ID,
			"item_key":  item.ItemKey,
			"item_type": item.ItemType,
			"verdict":   verdict.Verdict,
			"reason":    verdict.Reason,
			"evidence":  verdict.Evidence,
		})

		if verdict.Verdict == revalidationResolved {
			logger.Info("review item auto-resolved — the finding no longer holds on the deployed page",
				zap.String("item_id", item.ID),
				zap.String("item_key", item.ItemKey),
				zap.String("item_type", item.ItemType),
				zap.String("reason", verdict.Reason),
				zap.Bool("dry_run", dryRun))
		}
		if dryRun {
			continue
		}
		didClose, wErr := applyRevalidation(ctx, params.DB, item, verdict, record)
		if wErr != nil {
			logger.Warn("could not persist revalidation; item left as-is for the next sweep",
				zap.String("item_id", item.ID), zap.Error(wErr))
			continue
		}
		if didClose {
			closed++
		}
	}

	logger.Info("revalidate_review_queue: sweep done",
		zap.Int("scanned", len(items)),
		zap.Int("resolved", counts[revalidationResolved]),
		zap.Int("still_holds", counts[revalidationStillHolds]),
		zap.Int("unknown", counts[revalidationUnknown]),
		zap.Int("closed", closed),
		zap.Bool("dry_run", dryRun))

	return map[string]interface{}{
		"success":         true,
		"scanned":         len(items),
		"resolved":        counts[revalidationResolved],
		"still_holds":     counts[revalidationStillHolds],
		"unknown":         counts[revalidationUnknown],
		"closed":          closed,
		"dry_run":         dryRun,
		"capped_at":       maxItems,
		"uncovered_types": uncovered,
		"items":           reports,
	}, nil
}

// loadParkedReviewItems reads the queue, oldest first — the oldest items are the
// ones most likely to be describing a page state that no longer exists.
func loadParkedReviewItems(ctx context.Context, db *sql.DB, siteFilter, typeFilter string, maxItems int) ([]parkedReviewItem, error) {
	query := `
		SELECT id::text, site_id, item_type, COALESCE(item_key, ''), COALESCE(spec, '{}'::jsonb)
		FROM site_work_items
		WHERE status = 'needs_human_review'`
	args := []interface{}{}
	if strings.TrimSpace(siteFilter) != "" {
		siteID, err := uuid.Parse(strings.TrimSpace(siteFilter))
		if err != nil {
			return nil, fmt.Errorf("invalid site_id %q: %w", siteFilter, err)
		}
		args = append(args, siteID)
		query += fmt.Sprintf(" AND site_id = $%d", len(args))
	}
	if strings.TrimSpace(typeFilter) != "" {
		args = append(args, strings.TrimSpace(typeFilter))
		query += fmt.Sprintf(" AND item_type = $%d", len(args))
	}
	args = append(args, maxItems)
	query += fmt.Sprintf(" ORDER BY created_at ASC LIMIT $%d", len(args))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load parked review items: %w", err)
	}
	defer rows.Close()

	var items []parkedReviewItem
	for rows.Next() {
		var it parkedReviewItem
		var specJSON []byte
		if err := rows.Scan(&it.ID, &it.SiteID, &it.ItemType, &it.ItemKey, &specJSON); err != nil {
			return nil, fmt.Errorf("scan parked review item: %w", err)
		}
		if len(specJSON) > 0 {
			_ = json.Unmarshal(specJSON, &it.Spec)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// applyRevalidation writes one verdict. The status guard is what makes the sweep
// safe to run alongside a human working the same queue: if the item left
// needs_human_review between the read and the write, nothing happens.
//
// Note what is NOT set on the non-closing path: updated_at. See the file header.
func applyRevalidation(ctx context.Context, db *sql.DB, item parkedReviewItem,
	verdict revalidationVerdict, record map[string]interface{}) (bool, error) {

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return false, fmt.Errorf("marshal revalidation record: %w", err)
	}

	if verdict.Verdict != revalidationResolved {
		_, err = db.ExecContext(ctx, `
			UPDATE site_work_items
			SET result = COALESCE(result, '{}'::jsonb)
			          || jsonb_build_object('revalidation', $2::jsonb)
			WHERE id = $1::uuid AND status = 'needs_human_review'
		`, item.ID, string(recordJSON))
		return false, err
	}

	res, err := db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status          = 'complete',
		    completed_at    = NOW(),
		    updated_at      = NOW(),
		    resolution_path = 'auto:revalidated',
		    result = COALESCE(result, '{}'::jsonb)
		          || jsonb_build_object('revalidation', $2::jsonb)
		WHERE id = $1::uuid AND status = 'needs_human_review'
	`, item.ID, string(recordJSON))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// revalidateNamedFields builds the revalidator shared by every "these fields are
// missing" finding. specKey names the spec entry holding the field list.
//
// It keys the component lookup on (page_name, slot) and NEVER on
// spec.component_id: page_components.id is not stable across re-renders, so a
// stale component_id reads as "the section is gone" when the section is in fact
// still there under a new row id. Measured 2026-07-25: keyed on component_id,
// 30 of 30 parked needs_section_data items and 11 of 45 required_fields_missing
// items resolve to a component that no longer exists.
func revalidateNamedFields(specKey string) reviewRevalidator {
	return func(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict {
		fields := extractMissingFieldNames(item.Spec[specKey])
		if len(fields) == 0 {
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("spec.%s names no fields, so there is nothing to re-check", specKey),
			}
		}

		pageName := specString(item.Spec, "page_name", "page")
		slotName := specString(item.Spec, "slot_name", "section_name", "component_function")
		if pageName == "" || slotName == "" {
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  "spec names no page and/or no section, so the finding cannot be located on the live site",
			}
		}

		row, err := loadPageComponentBySlotRO(ctx, db, item.SiteID, pageName, NormalizeComponentFunction(slotName))
		if err == sql.ErrNoRows {
			// The section is not on the deployed page. That MIGHT mean the
			// finding is moot, but it might equally be a lookup miss — so it is
			// not positive evidence and the item stays queued.
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("no deployed component matches %s/%s", pageName, slotName),
			}
		}
		if err != nil {
			logger.Warn("revalidate: component lookup failed",
				zap.String("item_id", item.ID), zap.Error(err))
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("component lookup failed for %s/%s", pageName, slotName),
			}
		}
		if len(row.ContentData) == 0 {
			// The component renders from a template, a DERIVED source or a
			// static fallback rather than content_data. content_data cannot
			// answer the question, so we do not pretend it can.
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("component %s/%s carries no content_data; it renders from another source", pageName, row.SlotName),
				Evidence: map[string]interface{}{
					"page_name":         pageName,
					"slot_name":         row.SlotName,
					"page_component_id": row.ID.String(),
					"build_status":      row.BuildStatus,
				},
			}
		}

		perField := map[string]bool{}
		allPopulated := true
		for _, f := range fields {
			populated := fieldPopulated(row.ContentData[f])
			perField[f] = populated
			if !populated {
				allPopulated = false
			}
		}

		evidence := map[string]interface{}{
			"page_name":         pageName,
			"slot_name":         row.SlotName,
			"page_component_id": row.ID.String(),
			"build_status":      row.BuildStatus,
			"fields":            perField,
		}
		if allPopulated {
			return revalidationVerdict{
				Verdict:  revalidationResolved,
				Reason:   fmt.Sprintf("every field this item reports missing (%s) is populated on the deployed component", strings.Join(fields, ", ")),
				Evidence: evidence,
			}
		}
		return revalidationVerdict{
			Verdict:  revalidationStillHolds,
			Reason:   "at least one reported-missing field is still empty on the deployed component",
			Evidence: evidence,
		}
	}
}

// extractMissingFieldNames reads a spec's missing-field list in either shape the
// producers emit: a plain array of names (unresolved_cta, required_fields_missing)
// or an array of objects carrying "field" (needs_section_data). Anything else
// yields no names, which the caller reports as 'unknown' rather than guessing.
func extractMissingFieldNames(raw interface{}) []string {
	list, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var names []string
	for _, entry := range list {
		switch v := entry.(type) {
		case string:
			if name := strings.TrimSpace(v); name != "" {
				names = append(names, name)
			}
		case map[string]interface{}:
			if name, _ := v["field"].(string); strings.TrimSpace(name) != "" {
				names = append(names, strings.TrimSpace(name))
			}
		}
	}
	return names
}

// specString returns the first non-empty string value among the given spec keys.
func specString(spec map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := spec[k].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// fieldPopulated reports whether a content_data value counts as supplied.
// Absent, null, blank string, empty list and empty map are all "still missing" —
// the templates render every one of them as nothing, which is the condition the
// original findings describe.
func fieldPopulated(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(t) != ""
	case []interface{}:
		return len(t) > 0
	case map[string]interface{}:
		return len(t) > 0
	case bool:
		return t
	default:
		return true // numbers and anything else concrete count as supplied
	}
}
