// FILE: platform/orchestration/actions/write_render_audit_findings_action.go
//
// WriteRenderAuditFindingsAction turns a render audit's browser measurements
// (request_render_audit → browser-runner adapter → awaited response) into
// routed site_work_items. Until this action, the render audit's findings
// stopped in collected_data — a correct measurement nobody consumed, the
// bugs_open/115 / 083 shape. This is the drain half the 256 seed's header
// named as deliberately deferred.
//
// What it FILES, and where it drains:
//
//   - A firm contrast failure (over_image=false) → item_type "contrast_failure"
//     at css-patch-agent. The dedup key contrast_failure:<page-path>#<class> is
//     the de-facto verifier: a repaired pairing stops being re-filed because
//     the NEXT render audit no longer measures it, and an unrepaired one
//     collapses onto the open row.
//   - A broken image whose src resolves to an assets row → item_type
//     "undeployed_asset" at asset-deployer, keyed undeployed_asset:<asset_id>
//     so it CO-DEDUPS with check_undeployed_assets' rows (same item_type, same
//     dedup unit, same handler — the shared namespace is the point, not an
//     accident).
//
// What it deliberately does NOT file, and why (each is COUNTED in the result,
// so "not filed" is visible, never silent):
//
//   - over_image contrast approximations — the adapter's own header calls the
//     backdrop unknown; filing approximations at a fixer manufactures churn.
//   - overflow findings — on THIS path they carry no culprit attribution
//     (URL + widths only; the attributing measurement lives in the Tier-4
//     no_horizontal_overflow check). A responsive_fix item that cannot name a
//     component is undispatchable paperwork.
//   - broken images with NO assets row — that is a source-side planning gap
//     owned by check_content_image_missing / the imageryplan contract; minting
//     a needs_imagery item here without an imagery-plan spec would hand
//     image-build-handler a row it cannot act on.
//   - anything whose culprit markup lives in a LOCKED component (page or
//     chrome, locked_at IS NOT NULL) — a lock is a human's "hands off", and a
//     site-CSS patch aimed at a locked component's class changes the locked
//     thing's appearance by the back door. Counted as skipped_locked.
//
// Items are born status='detected': promotion is improvement-loop's
// triage_findings (the migration-286 single owner), so this action composes
// with the drain rather than bypassing it.
//
// A run that would file more than max_items (default 60) files the worst
// max_items by ratio and reports findings_capped=true with the dropped count —
// the no-silent-caps rule.
//
// Registration:
//   "write_render_audit_findings": {
//       Handler:     WriteRenderAuditFindingsAction,
//       Category:    "quality",
//       Description: "File a render audit's firm findings as routed work items",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var WriteRenderAuditFindingsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"audit_field", "max_items"},
	Defaults:    map[string]interface{}{"audit_field": "render_audit"},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_render_audit_findings", WriteRenderAuditFindingsInputSpec)
}

// Local mirror of the adapter's result shape (internal/adapters/browserrunner/
// render_audit_action.go). Mirrored rather than imported: platform/ does not
// depend on internal/adapters, and the JSON contract is the coupling point.
type renderAuditContrast struct {
	URL       string  `json:"url"`
	Tag       string  `json:"tag"`
	Class     string  `json:"class"`
	Text      string  `json:"text"`
	FG        string  `json:"fg"`
	BG        string  `json:"bg"`
	Ratio     float64 `json:"ratio"`
	Need      float64 `json:"need"`
	FontPx    int     `json:"font_px"`
	OverImage bool    `json:"over_image"`
}

type renderAuditBrokenImage struct {
	URL string `json:"url"`
	Src string `json:"src"`
	Alt string `json:"alt"`
}

type renderAuditPayload struct {
	RunID    string                   `json:"run_id"`
	Contrast []renderAuditContrast    `json:"contrast"`
	Images   []renderAuditBrokenImage `json:"broken_images"`
	Overflow []json.RawMessage        `json:"overflow"`
}

func WriteRenderAuditFindingsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_render_audit_findings"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("write_render_audit_findings: database connection required")
	}

	config := params.StepConfig.Config
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, WriteRenderAuditFindingsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("write_render_audit_findings: input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("write_render_audit_findings: invalid site_id: %w", err)
	}

	auditField := inputs.Get("audit_field")
	if auditField == "" {
		auditField = "render_audit"
	}
	maxItems := datahelpers.GetIntField(config, "max_items", 60)

	payload, err := extractRenderAuditPayload(params.CollectedData, auditField)
	if err != nil {
		// Absent ≠ malformed: an audit that never ran must not read as a clean
		// write. Fail the step so the workflow's error edge sees it.
		return nil, fmt.Errorf("write_render_audit_findings: %w", err)
	}

	lockedHTML, err := loadLockedComponentHTML(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("write_render_audit_findings: locked-component load: %w", err)
	}

	type pending struct {
		item  workItem
		ratio float64
	}
	var toFile []pending
	skippedLocked := 0
	overImage := 0

	for _, c := range payload.Contrast {
		if c.OverImage {
			overImage++
			continue
		}
		if htmlCorpusContainsClass(lockedHTML, c.Class) {
			skippedLocked++
			continue
		}
		pagePath := urlPath(c.URL)
		selector := contrastSelector(c.Tag, c.Class)
		severity := "medium"
		if c.Ratio < 2.0 {
			severity = "high"
		}
		spec := map[string]interface{}{
			"page_url":        c.URL,
			"page_path":       pagePath,
			"selector":        selector,
			"fg":              c.FG,
			"bg":              c.BG,
			"ratio":           c.Ratio,
			"need":            c.Need,
			"font_px":         c.FontPx,
			"text_sample":     truncateString(c.Text, 120),
			"fix_type":        "contrast_fix",
			"current_value":   fmt.Sprintf("%s on %s measures %.2f:1 (needs %.1f:1)", c.FG, c.BG, c.Ratio, c.Need),
			"acceptance_test": fmt.Sprintf("elements matching %s on %s render at or above %.1f:1 contrast on the next render audit", selector, pagePath, c.Need),
			"max_fix_attempts": 2,
			"run_id":          payload.RunID,
		}
		specJSON, mErr := json.Marshal(spec)
		if mErr != nil {
			logger.Warn("write_render_audit_findings: spec marshal failed", zap.Error(mErr))
			continue
		}
		toFile = append(toFile, pending{
			ratio: c.Ratio,
			item: workItem{
				siteID:       siteID,
				source:       "render-audit",
				pipeline:     "build",
				itemType:     "contrast_failure",
				severity:     severity,
				summary:      fmt.Sprintf("Contrast %.2f:1 (needs %.1f:1) for %s on %s", c.Ratio, c.Need, selector, pagePath),
				spec:         string(specJSON),
				priority:     60,
				handlerAgent: "css-patch-agent",
				status:       "detected",
				createdBy:    params.ExecutionContext.Sender.AgentType,
				itemKey:      workItemKey("contrast_failure", pagePath+"#"+selector),
			},
		})
	}

	// Worst first, then the loud cap.
	sort.Slice(toFile, func(i, j int) bool { return toFile[i].ratio < toFile[j].ratio })
	dropped := 0
	if len(toFile) > maxItems {
		dropped = len(toFile) - maxItems
		toFile = toFile[:maxItems]
	}

	unattributedImages := 0
	for _, img := range payload.Images {
		assetID, aErr := lookupAssetBySrc(ctx, params.DB, siteID, img.Src)
		if aErr != nil {
			logger.Warn("write_render_audit_findings: asset lookup failed",
				zap.String("src", img.Src), zap.Error(aErr))
			unattributedImages++
			continue
		}
		if assetID == uuid.Nil {
			unattributedImages++
			continue
		}
		spec := map[string]interface{}{
			"asset_id": assetID.String(),
			"src":      img.Src,
			"page_url": img.URL,
			"reason":   "render_audit_404",
			"run_id":   payload.RunID,
		}
		specJSON, mErr := json.Marshal(spec)
		if mErr != nil {
			continue
		}
		// Same item_type, same key namespace, same handler as
		// check_undeployed_assets — deliberate co-dedup (see header).
		toFile = append(toFile, pending{item: workItem{
			siteID:       siteID,
			source:       "render-audit",
			pipeline:     "build",
			itemType:     "undeployed_asset",
			severity:     "medium",
			summary:      fmt.Sprintf("Image %s serves broken on %s (asset row exists)", img.Src, urlPath(img.URL)),
			spec:         string(specJSON),
			priority:     60,
			handlerAgent: "asset-deployer",
			status:       "detected",
			createdBy:    params.ExecutionContext.Sender.AgentType,
			itemKey:      fmt.Sprintf("undeployed_asset:%s", assetID),
		}})
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("write_render_audit_findings: begin tx: %w", err)
	}
	defer tx.Rollback()

	inserted, deduped := 0, 0
	for _, p := range toFile {
		ok, iErr := insertWorkItem(ctx, tx, p.item, logger)
		if iErr != nil {
			return nil, fmt.Errorf("write_render_audit_findings: insert %s: %w", p.item.itemKey, iErr)
		}
		if ok {
			inserted++
		} else {
			deduped++
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("write_render_audit_findings: commit: %w", err)
	}

	result := map[string]interface{}{
		"inserted":             inserted,
		"deduped":              deduped,
		"skipped_locked":       skippedLocked,
		"over_image_reported":  overImage,
		"overflow_reported":    len(payload.Overflow),
		"unattributed_images":  unattributedImages,
		"findings_capped":      dropped > 0,
		"findings_dropped":     dropped,
		"run_id":               payload.RunID,
	}
	logger.Info("write_render_audit_findings: complete",
		zap.Int("inserted", inserted), zap.Int("deduped", deduped),
		zap.Int("skipped_locked", skippedLocked), zap.Int("dropped", dropped))
	return result, nil
}

// extractRenderAuditPayload unwraps collected_data[field], descending into the
// coordinator's ".response" wrapper when present (coordinator.go stores an
// awaited agent response under output_field.response).
func extractRenderAuditPayload(collected map[string]interface{}, field string) (*renderAuditPayload, error) {
	raw, ok := collected[field]
	if !ok || raw == nil {
		return nil, fmt.Errorf("no %q in collected data — the render audit has not run", field)
	}
	obj, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%q is not an object", field)
	}
	if _, direct := obj["contrast"]; !direct {
		inner, hasResp := obj["response"].(map[string]interface{})
		if !hasResp {
			return nil, fmt.Errorf("%q carries neither a result nor a .response — the audit is still awaited or failed", field)
		}
		obj = inner
	}
	buf, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("re-marshal %q: %w", field, err)
	}
	var payload renderAuditPayload
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", field, err)
	}
	return &payload, nil
}

// loadLockedComponentHTML returns the rendered_html of every locked component
// (page and chrome) for the site, for culprit-lock checks.
func loadLockedComponentHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(pc.rendered_html, '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1 AND pc.locked_at IS NOT NULL
		UNION ALL
		SELECT COALESCE(sc.rendered_html, '')
		FROM site_components sc
		WHERE sc.site_id = $1 AND sc.locked_at IS NOT NULL
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var html string
		if err := rows.Scan(&html); err != nil {
			return nil, err
		}
		if html != "" {
			out = append(out, html)
		}
	}
	return out, rows.Err()
}

// htmlCorpusContainsClass reports whether any locked component's markup uses
// any of the finding's class tokens. Conservative on purpose: a false skip
// costs one unfiled finding; a false file edits a locked component's look.
func htmlCorpusContainsClass(corpus []string, class string) bool {
	for _, token := range strings.Fields(class) {
		if token == "" {
			continue
		}
		for _, html := range corpus {
			if strings.Contains(html, token) {
				return true
			}
		}
	}
	return false
}

// lookupAssetBySrc resolves an <img src> to the site's assets row by basename,
// against storage_path and url. Returns uuid.Nil when nothing matches. Matching
// is by basename because assets.url may be stale (bugs_open/152) while the
// stable serving path keeps the storage basename.
func lookupAssetBySrc(ctx context.Context, db *sql.DB, siteID uuid.UUID, src string) (uuid.UUID, error) {
	base := path.Base(urlPath(src))
	if base == "" || base == "." || base == "/" {
		return uuid.Nil, nil
	}
	var id uuid.UUID
	err := db.QueryRowContext(ctx, `
		SELECT id FROM assets
		WHERE site_id = $1
		  AND status = 'active'
		  AND (storage_path LIKE '%' || $2 OR url LIKE '%' || $2)
		ORDER BY created_at DESC
		LIMIT 1
	`, siteID, base).Scan(&id)
	if err == sql.ErrNoRows {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func contrastSelector(tag, class string) string {
	tokens := strings.Fields(class)
	if len(tokens) > 0 {
		return tag + "." + tokens[0]
	}
	if tag != "" {
		return tag
	}
	return "*"
}

func urlPath(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" {
		return raw
	}
	return u.Path
}

func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
