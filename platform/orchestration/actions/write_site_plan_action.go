// FILE: platform/orchestration/actions/write_site_plan_action.go
//
// WriteSitePlanAction is the plan-builder's terminal step. It takes the
// LLM's structure output plus optional design_direction and
// content_strategy partials from CollectedData, runs each page through
// the role validator and the canonicaliser (which now produces both
// canonical name and canonical URL — see doc 030 Phase 1
// consolidation), and writes one site_plans row + N site_plan_pages
// rows + M site_plan_partials rows in a single transaction.
//
// The action does NOT emit work items. That responsibility belongs to
// the reconciler (doc 030 step 5). Plan-builder's job ends with
// "the new plan is durably current"; the reconciler decides what to
// build from it.
//
// CollectedData keys read:
//   - page_plan / site_plan  (extracted by existing extractPagesFromPlan)
//   - design_direction       (optional, written as site_plan_partials)
//   - content_strategy       (optional, written as site_plan_partials)
//   - site_id                (declared in WriteSitePlanInputSpec)
//
// Returns:
//
//	{
//	  plan_id: <uuid>,
//	  superseded_plan_id: <uuid string, empty if first plan for site>,
//	  pages_written: <int>,
//	  partials_written: <int>,
//	  role_corrections: <int>,
//	}

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// WriteSitePlanInputSpec declares the inputs this action expects.
// Fields documented in the file-level comment above.
//
// Note: design_direction and content_strategy are NOT declared here
// because they're free-form jsonb partials read directly from
// CollectedData. ActionInputSpec is for typed scalars; partials are
// pass-through blobs.
var WriteSitePlanInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_site_plan", WriteSitePlanInputSpec)
}

// WriteSitePlanAction is the action handler. Registered in the action
// registry as "write_site_plan", IsLocal=true.
func WriteSitePlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_site_plan"))
	logger.Info("WriteSitePlanAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Handle initialization (chassis convention — see SyncPagesToDBAction).
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		WriteSitePlanInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// ── 2. Extract LLM page list ────────────────────────────────────────
	rawPages := extractPagesFromPlan(params.CollectedData, logger)
	if len(rawPages) == 0 {
		return nil, fmt.Errorf("no pages found in page_plan / site_plan")
	}

	// ── 3. Convert raw LLM maps to LLMPlannedPage and run validator ─────
	llmPages := make([]datahelpers.LLMPlannedPage, 0, len(rawPages))
	for _, p := range rawPages {
		llmPages = append(llmPages, datahelpers.LLMPlannedPage{
			Name:          datahelpers.GetStringField(p, "name", ""),
			Role:          firstNonEmptyField(p, "page_type", "type", "role"),
			Slug:          datahelpers.GetStringField(p, "slug", ""),
			URL:           datahelpers.GetStringField(p, "url", ""),
			ParentSection: datahelpers.GetStringField(p, "parent_section", ""),
		})
	}

	validated := datahelpers.ValidateRoles(llmPages)

	// Count corrections for observability — the gamesdesign post-mortem
	// showed 3-of-12 mislabels; we log this every run so we can spot
	// prompt regressions.
	corrections := 0
	for _, v := range validated {
		if v.CorrectedFromRole != "" {
			corrections++
			logger.Info("WriteSitePlanAction: role corrected by validator",
				zap.String("name", v.Name),
				zap.String("from", v.CorrectedFromRole),
				zap.String("to", v.Role))
		}
	}

	// ── 4. Canonicalise each validated page (name, url, page_type) ─────
	// One CanonicalisePage call per page. Phase 1 extended the helper
	// to honour ParentSection in URL synthesis (doc 030); we no longer
	// need a separate URL helper.
	type planPageRow struct {
		Name          string
		Role          string
		Slug          string
		URL           string
		ParentSection string
		InHeader      bool
		InFooter      bool
		NavOrder      int
		PageData      []byte // marshalled JSONB
	}
	planRows := make([]planPageRow, 0, len(validated))

	for i, v := range validated {
		canonicalName, canonicalURL, canonicalType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
			Role:          v.Role,
			Slug:          firstNonEmpty(v.Slug, v.Name),
			ParentSection: v.ParentSection,
		})
		if canonicalName == "" || canonicalURL == "" {
			logger.Warn("WriteSitePlanAction: page failed canonicalisation, skipping",
				zap.String("raw_name", v.Name),
				zap.String("role", v.Role))
			continue
		}
		// canonicalType passes through as the role we persist; it
		// matches v.Role except in the unknown-role case where the
		// helper preserves the raw string.
		_ = canonicalType

		// Read remaining fields from the original raw map so we don't
		// lose title / sections / meta_description.
		raw := rawPages[i]

		pageData := buildPageDataJSON(raw)

		planRows = append(planRows, planPageRow{
			Name:          canonicalName,
			Role:          v.Role,
			Slug:          firstNonEmpty(v.Slug, v.Name),
			URL:           canonicalURL,
			ParentSection: v.ParentSection,
			InHeader:      datahelpers.GetBoolField(raw, "in_header", true),
			InFooter:      datahelpers.GetBoolField(raw, "in_footer", true),
			NavOrder:      datahelpers.GetIntField(raw, "nav_order", 100),
			PageData:      pageData,
		})
	}

	if len(planRows) == 0 {
		return nil, fmt.Errorf("no pages survived validation/canonicalisation")
	}

	// ── 5. Single transaction: supersede prior plan, write new plan ────
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 5a. Mark previous current plan as superseded (if any).
	var supersededID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		UPDATE site_plans
		   SET is_current    = false,
		       superseded_at = NOW()
		 WHERE site_id    = $1
		   AND is_current = true
		RETURNING id
	`, siteID).Scan(&supersededID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("supersede previous plan: %w", err)
	}

	// 5b. Insert the new plan row.
	var planID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_plans (site_id, source_agent, created_by)
		VALUES ($1, $2, $3)
		RETURNING id
	`, siteID, "site-planner", "write_site_plan").Scan(&planID)
	if err != nil {
		return nil, fmt.Errorf("insert site_plans: %w", err)
	}

	// 5c. Insert site_plan_pages rows.
	for _, r := range planRows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_plan_pages
			    (plan_id, name, role, slug, url, parent_section,
			     in_header, in_footer, nav_order, page_data)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''),
			        $7, $8, $9, $10::jsonb)
		`,
			planID, r.Name, r.Role, r.Slug, r.URL, r.ParentSection,
			r.InHeader, r.InFooter, r.NavOrder, string(r.PageData))
		if err != nil {
			return nil, fmt.Errorf("insert site_plan_pages for %q: %w", r.Name, err)
		}
	}

	// 5d. Insert eager partials (design_direction, content_strategy).
	partialsWritten := 0
	for _, partialKey := range []string{"design_direction", "content_strategy"} {
		raw, ok := params.CollectedData[partialKey]
		if !ok {
			logger.Info("WriteSitePlanAction: partial absent, skipping",
				zap.String("partial", partialKey))
			continue
		}
		// Tolerate the LLM-response wrapper: if the value is a map with
		// a "result" key, persist that. Otherwise persist the value as-is.
		dataJSON, err := marshalPartial(raw)
		if err != nil {
			logger.Warn("WriteSitePlanAction: partial marshalling failed, skipping",
				zap.String("partial", partialKey),
				zap.Error(err))
			continue
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_plan_partials
			    (plan_id, partial_type, data, source_agent, created_by)
			VALUES ($1, $2, $3::jsonb, $4, $5)
		`, planID, partialKey, string(dataJSON), "site-planner", "write_site_plan")
		if err != nil {
			return nil, fmt.Errorf("insert site_plan_partials/%s: %w", partialKey, err)
		}
		partialsWritten++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	supersededIDStr := ""
	if supersededID != uuid.Nil {
		supersededIDStr = supersededID.String()
	}

	logger.Info("WriteSitePlanAction: Complete",
		zap.String("site_id", siteID.String()),
		zap.String("plan_id", planID.String()),
		zap.String("superseded_plan_id", supersededIDStr),
		zap.Int("pages_written", len(planRows)),
		zap.Int("partials_written", partialsWritten),
		zap.Int("role_corrections", corrections),
	)

	return map[string]interface{}{
		"plan_id":            planID.String(),
		"superseded_plan_id": supersededIDStr,
		"pages_written":      len(planRows),
		"partials_written":   partialsWritten,
		"role_corrections":   corrections,
	}, nil
}

// firstNonEmpty returns the first non-empty string from its arguments.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// firstNonEmptyField reads keys from a map in order, returning the first
// non-empty string value. Used to handle LLM emissions that vary in
// which key carries the role ("page_type" vs "type" vs "role").
func firstNonEmptyField(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v := datahelpers.GetStringField(m, k, ""); v != "" {
			return v
		}
	}
	return ""
}

// buildPageDataJSON extracts the non-structural page fields (title,
// sections, meta_description, nav_label) into a JSONB blob for
// site_plan_pages.page_data. Structural fields (name, role, slug, url,
// parent_section, in_header, in_footer, nav_order) live in their own
// columns and are excluded.
func buildPageDataJSON(raw map[string]interface{}) []byte {
	keep := []string{
		"title",
		"sections",
		"meta_description",
		"nav_label",
	}
	out := make(map[string]interface{}, len(keep))
	for _, k := range keep {
		if v, ok := raw[k]; ok {
			out[k] = v
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return []byte("{}")
	}
	return b
}

// marshalPartial accepts arbitrary CollectedData partial values and
// produces a JSONB-ready []byte. Unwraps a single-level "result"
// wrapper if present (call_agent response format).
func marshalPartial(v interface{}) ([]byte, error) {
	if m, ok := v.(map[string]interface{}); ok {
		if inner, ok := m["result"]; ok {
			return json.Marshal(inner)
		}
	}
	return json.Marshal(v)
}
