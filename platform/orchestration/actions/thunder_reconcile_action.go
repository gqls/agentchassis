// FILE: platform/orchestration/actions/thunder_reconcile_action.go
//
// reconcile_thunder_instances — the orphan sweep's step 2. Compares
// Thunder's own account view (step 1, dispatch_thunder_list) against the
// thunder_instances table and files a site_work_items row for every
// instance that is billing at Thunder while invisible to our automation.
//
// Why this exists: thunder_instances is written by provision_action.go
// AFTER the vendor instance is up, so a crash in that window — or any
// hand-created instance — leaves a box billing at Thunder with no row
// here. Every automated check (the reaper included) reads only the table,
// so such an instance bills forever until a human notices. The manual net
// was RUNBOOK §1b ("AM I BEING BILLED RIGHT NOW?"); this action is that
// check, automated, with a durable trail.
//
// Findings (classification lives in classifyThunderReconcile, pure and
// unit-tested):
//   - orphan_no_row       vendor instance, no thunder_instances row at all.
//                         Filed as a work item, severity high.
//   - orphan_terminal_row vendor instance whose only rows are terminal —
//                         our records say it is gone, Thunder still lists
//                         it. Filed, severity high.
//   - ghost_row           live thunder_instances row with no vendor
//                         instance behind it. NOT filed — it costs nothing
//                         (there is nothing billing), self-heals through
//                         the decommission path (Thunder delete 404s to
//                         success), and a mid-decommission snapshot can
//                         produce one transiently. Reported in the result
//                         and logged at Warn.
//
// The provision window is absorbed by a grace period on orphan_no_row:
// a vendor instance younger than grace_minutes (default 30 — well past
// the longest WaitForRunning) is counted in_grace, not filed. A vendor
// instance with NO parseable createdAt is flagged despite the grace —
// an unknown-age instance must not be able to hide behind it.
//
// Work items go to the system.internal site (the platform-level site row
// that needs_diagnosis items already use), item_type thunder_orphan,
// item_key "thunder_orphan:<vendor id>", written through insertWorkItem —
// the shared door (load_work_item_actions.go), which carries the targeted
// idx_swi_dedup conflict clause via workItemTerminalStatuses, within-cycle
// suppression, and the two-strike unresolved labelling. Hand-rolling this
// INSERT is exactly what a prior council round objected to at three other
// call sites (bugs_open/091, workItem.parentItemID's own comment tells the
// story). Re-observing an already-filed orphan is dedup-suppressed while
// the item is open; an orphan closed twice in seven days and re-observed
// is born unresolved, which the RFC_010 revalidation path can see.
//
// There is deliberately NO automated remediation here: decommissioning
// needs a thunder_instances row, which is exactly what an orphan lacks,
// and killing vendor instances on the strength of a list comparison is
// authority this read-and-report action does not want. The work item's
// summary tells the human what to do.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// thunderInstanceTerminalStatuses mirrors the terminal set in
// internal/adapters/thunder/store.MarkDecommissioning (rows in these
// statuses are "already done"). A row in any of these states asserts the
// vendor instance no longer exists — which is what makes a vendor
// instance matching ONLY such rows an orphan.
var thunderInstanceTerminalStatuses = map[string]bool{
	"decommissioned": true,
	"failed":         true,
	"reaped":         true,
}

// thunderVendorInstance is the subset of Thunder's list entry the
// reconcile needs, parsed defensively from the awaited response JSON.
type thunderVendorInstance struct {
	ID        string
	UUID      string
	Name      string
	Status    string // Thunder returns UPPERCASE ("RUNNING")
	GpuType   string
	CreatedAt time.Time // zero = not supplied / unparseable
}

// thunderDBRow is the subset of a thunder_instances row the reconcile needs.
type thunderDBRow struct {
	RowID             string
	ThunderInstanceID string
	Status            string
}

// thunderReconcileFinding is one mismatch, either direction.
type thunderReconcileFinding struct {
	Kind       string `json:"kind"` // orphan_no_row | orphan_terminal_row | ghost_row
	VendorID   string `json:"vendor_id,omitempty"`
	VendorUUID string `json:"vendor_uuid,omitempty"`
	Status     string `json:"status"` // vendor status for orphans, row status for ghosts
	GpuType    string `json:"gpu_type,omitempty"`
	RowID      string `json:"row_id,omitempty"`
	AgeUnknown bool   `json:"age_unknown,omitempty"`
	AgeMinutes int    `json:"age_minutes,omitempty"`
	Summary    string `json:"summary"`
}

// thunderReconcileOutcome is classifyThunderReconcile's full result.
type thunderReconcileOutcome struct {
	VendorBilling int // vendor instances not DELETED
	VendorDeleted int
	DBRows        int
	DBLive        int // rows in a non-terminal status
	Matched       int // vendor instances covered by a live row
	InGrace       int // vendor instances younger than grace, unfiled
	Findings      []thunderReconcileFinding
}

// classifyThunderReconcile is the pure comparison. Vendor statuses are
// compared case-insensitively (Thunder returns UPPERCASE); a vendor
// instance in DELETED status is not billing and counts as absent on both
// sides of the comparison.
func classifyThunderReconcile(
	vendor []thunderVendorInstance,
	rows []thunderDBRow,
	grace time.Duration,
	now time.Time,
) thunderReconcileOutcome {
	out := thunderReconcileOutcome{DBRows: len(rows)}

	rowsByThunderID := make(map[string][]thunderDBRow, len(rows))
	for _, r := range rows {
		rowsByThunderID[r.ThunderInstanceID] = append(rowsByThunderID[r.ThunderInstanceID], r)
		if !thunderInstanceTerminalStatuses[strings.ToLower(r.Status)] {
			out.DBLive++
		}
	}

	// Vendor ids that exist (not DELETED) — the presence set for ghost
	// detection below.
	vendorPresent := make(map[string]bool, len(vendor))

	for _, v := range vendor {
		if strings.EqualFold(strings.TrimSpace(v.Status), "deleted") {
			out.VendorDeleted++
			continue
		}
		out.VendorBilling++
		vendorPresent[v.ID] = true

		var hasLiveRow, hasTerminalRow bool
		for _, r := range rowsByThunderID[v.ID] {
			if thunderInstanceTerminalStatuses[strings.ToLower(r.Status)] {
				hasTerminalRow = true
			} else {
				hasLiveRow = true
			}
		}

		switch {
		case hasLiveRow:
			out.Matched++
		case hasTerminalRow:
			out.Findings = append(out.Findings, thunderReconcileFinding{
				Kind:       "orphan_terminal_row",
				VendorID:   v.ID,
				VendorUUID: v.UUID,
				Status:     v.Status,
				GpuType:    v.GpuType,
				AgeUnknown: v.CreatedAt.IsZero(),
				AgeMinutes: ageMinutes(v.CreatedAt, now),
				Summary: fmt.Sprintf(
					"THUNDER ORPHAN: instance %s (%s, %s) is still listed at Thunder but every thunder_instances row for it is terminal — our records say it is gone while Thunder may still be billing it. Verify at the Thunder console and delete it by hand; then establish which decommission failed to stick (RUNBOOK finetuning_uk_service §1b).",
					v.ID, orUnknown(v.GpuType), v.Status),
			})
		default:
			age := now.Sub(v.CreatedAt)
			if !v.CreatedAt.IsZero() && age < grace {
				// Provision window: provision_action.go INSERTs the row
				// only after the box is up, so a young vendor instance
				// with no row yet is the normal shape of a provision in
				// flight — not an orphan, this time.
				out.InGrace++
				continue
			}
			out.Findings = append(out.Findings, thunderReconcileFinding{
				Kind:       "orphan_no_row",
				VendorID:   v.ID,
				VendorUUID: v.UUID,
				Status:     v.Status,
				GpuType:    v.GpuType,
				AgeUnknown: v.CreatedAt.IsZero(),
				AgeMinutes: ageMinutes(v.CreatedAt, now),
				Summary: fmt.Sprintf(
					"THUNDER ORPHAN: instance %s (%s, %s, age %s) is billing at Thunder with NO thunder_instances row — invisible to the reaper and every automated check. Kill it at the Thunder console, then establish what created it without writing a row (RUNBOOK finetuning_uk_service §1b).",
					v.ID, orUnknown(v.GpuType), v.Status, ageLabel(v.CreatedAt, now)),
			})
		}
	}

	// Ghosts: live rows whose instance does not exist at Thunder.
	for _, r := range rows {
		if thunderInstanceTerminalStatuses[strings.ToLower(r.Status)] {
			continue
		}
		if vendorPresent[r.ThunderInstanceID] {
			continue
		}
		out.Findings = append(out.Findings, thunderReconcileFinding{
			Kind:     "ghost_row",
			RowID:    r.RowID,
			VendorID: r.ThunderInstanceID,
			Status:   r.Status,
			Summary: fmt.Sprintf(
				"thunder_instances row %s (thunder_instance_id %s, status %s) has no instance behind it at Thunder — stale bookkeeping, nothing billing. The decommission path self-heals this (vendor delete 404s to success); if it persists across scans, decommission or delete the row.",
				r.RowID, r.ThunderInstanceID, r.Status),
		})
	}

	return out
}

func ageMinutes(created, now time.Time) int {
	if created.IsZero() {
		return 0
	}
	return int(now.Sub(created).Minutes())
}

func ageLabel(created, now time.Time) string {
	if created.IsZero() {
		return "UNKNOWN"
	}
	return now.Sub(created).Round(time.Minute).String()
}

func orUnknown(s string) string {
	if s == "" {
		return "gpu unknown"
	}
	return s
}

// ReconcileThunderInstancesAction reads step 1's awaited list response,
// loads thunder_instances, classifies, files work items for orphans, and
// returns the full outcome as the step result.
func ReconcileThunderInstancesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	// ── Step config ──
	listField := "thunder_list"
	if v, ok := params.StepConfig.Config["list_field"].(string); ok && v != "" {
		listField = v
	}
	grace := 30 * time.Minute
	if v, ok := params.StepConfig.Config["grace_minutes"].(float64); ok && v > 0 {
		grace = time.Duration(v) * time.Minute
	}

	// ── Read the vendor list from the awaited response ──
	vendor, err := extractThunderVendorList(params.CollectedData, listField)
	if err != nil {
		return nil, fmt.Errorf("reconcile_thunder_instances: %w", err)
	}

	// ── Load thunder_instances ──
	dbRows, err := loadThunderDBRows(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("reconcile_thunder_instances: load thunder_instances: %w", err)
	}

	outcome := classifyThunderReconcile(vendor, dbRows, grace, time.Now().UTC())

	// ── File work items for orphans; log everything ──
	filed, deduped := 0, 0
	for _, f := range outcome.Findings {
		switch f.Kind {
		case "orphan_no_row", "orphan_terminal_row":
			params.Logger.Error("THUNDER ORPHAN DETECTED",
				zap.String("kind", f.Kind),
				zap.String("vendor_id", f.VendorID),
				zap.String("vendor_status", f.Status),
				zap.String("gpu_type", f.GpuType),
				zap.Int("age_minutes", f.AgeMinutes),
				zap.Bool("age_unknown", f.AgeUnknown),
			)
			inserted, fileErr := fileThunderOrphanItem(ctx, params, f)
			if fileErr != nil {
				// Filing failed but the finding is still in the result and
				// the Error log above — degrade to reporting, don't lose
				// the scan.
				params.Logger.Error("Failed to file thunder_orphan work item",
					zap.String("vendor_id", f.VendorID), zap.Error(fileErr))
				continue
			}
			if inserted {
				filed++
			} else {
				deduped++
			}
		case "ghost_row":
			params.Logger.Warn("Thunder ghost row — live row with no vendor instance",
				zap.String("row_id", f.RowID),
				zap.String("thunder_instance_id", f.VendorID),
				zap.String("row_status", f.Status),
			)
		}
	}

	params.Logger.Info("Thunder reconcile complete",
		zap.Int("vendor_billing", outcome.VendorBilling),
		zap.Int("vendor_deleted", outcome.VendorDeleted),
		zap.Int("db_rows", outcome.DBRows),
		zap.Int("db_live", outcome.DBLive),
		zap.Int("matched", outcome.Matched),
		zap.Int("in_grace", outcome.InGrace),
		zap.Int("findings", len(outcome.Findings)),
		zap.Int("work_items_filed", filed),
		zap.Int("work_items_deduped", deduped),
	)

	return map[string]interface{}{
		"vendor_billing":     outcome.VendorBilling,
		"vendor_deleted":     outcome.VendorDeleted,
		"db_rows":            outcome.DBRows,
		"db_live":            outcome.DBLive,
		"matched":            outcome.Matched,
		"in_grace":           outcome.InGrace,
		"findings":           outcome.Findings,
		"work_items_filed":   filed,
		"work_items_deduped": deduped,
		"clean":              len(outcome.Findings) == 0,
	}, nil
}

// extractThunderVendorList digs the instances array out of
// CollectedData[listField].response (the shape an awaited adapter response
// is stored under) and parses each entry into thunderVendorInstance.
func extractThunderVendorList(collected map[string]interface{}, listField string) ([]thunderVendorInstance, error) {
	stepOut, ok := collected[listField].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("collected_data.%s missing or not an object — did the dispatch_thunder_list step run with output_field %q?", listField, listField)
	}
	resp, ok := stepOut["response"].(map[string]interface{})
	if !ok {
		// Tolerate the body being stored unwrapped.
		resp = stepOut
	}
	if success, ok := resp["success"].(bool); ok && !success {
		return nil, fmt.Errorf("list_instances response has success=false: %v", resp["detail"])
	}
	rawList, ok := resp["instances"].([]interface{})
	if !ok {
		// mapKeys is the package-wide helper from apply_adoption_plan_action.go.
		return nil, fmt.Errorf("list_instances response carries no instances array (keys: %v)", mapKeys(resp))
	}

	out := make([]thunderVendorInstance, 0, len(rawList))
	for _, raw := range rawList {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		v := thunderVendorInstance{}
		v.ID, _ = m["id"].(string)
		v.UUID, _ = m["uuid"].(string)
		v.Name, _ = m["name"].(string)
		v.Status, _ = m["status"].(string)
		v.GpuType, _ = m["gpuType"].(string)
		if ts, ok := m["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				v.CreatedAt = t
			}
		}
		out = append(out, v)
	}
	return out, nil
}

// loadThunderDBRows reads every thunder_instances row (id, vendor id,
// status). All-rows is deliberate: terminal rows are what distinguishes
// orphan_terminal_row from orphan_no_row.
func loadThunderDBRows(ctx context.Context, params ActionParams) ([]thunderDBRow, error) {
	rows, err := params.DB.QueryContext(ctx,
		`SELECT id::text, thunder_instance_id, status FROM thunder_instances`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []thunderDBRow
	for rows.Next() {
		var r thunderDBRow
		if err := rows.Scan(&r.RowID, &r.ThunderInstanceID, &r.Status); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// fileThunderOrphanItem writes one thunder_orphan work item against the
// system.internal site, through insertWorkItem — the shared door. Returns
// (true, nil) on insert, (false, nil) when suppressed (open-item dedup, the
// within-cycle window, or a two-strike relabel — all the shared helper's
// business, not this caller's).
func fileThunderOrphanItem(ctx context.Context, params ActionParams, f thunderReconcileFinding) (bool, error) {
	var siteIDStr string
	err := params.DB.QueryRowContext(ctx,
		`SELECT id::text FROM sites WHERE domain = 'system.internal'`).Scan(&siteIDStr)
	if err != nil {
		return false, fmt.Errorf("resolve system.internal site: %w", err)
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return false, fmt.Errorf("system.internal site id %q: %w", siteIDStr, err)
	}

	spec, err := json.Marshal(f)
	if err != nil {
		return false, fmt.Errorf("marshal finding spec: %w", err)
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin filing tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "thunder-orphan-scan",
		pipeline:     "maintenance",
		itemType:     "thunder_orphan",
		severity:     "high",
		summary:      f.Summary,
		spec:         string(spec),
		priority:     20, // urgent to a human reader; no automated consumer exists
		handlerAgent: "",
		status:       "detected",
		createdBy:    "reconcile_thunder_instances",
		itemKey:      "thunder_orphan:" + f.VendorID,
		// recurrenceExpected stays false on purpose: a re-observed orphan
		// after two closes IS a strike — the condition persists and the
		// unresolved label is the honest one.
	}, params.Logger)
	if err != nil {
		return false, err
	}
	return inserted, tx.Commit()
}
