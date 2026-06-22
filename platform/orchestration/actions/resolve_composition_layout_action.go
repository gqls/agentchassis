// FILE: platform/orchestration/actions/resolve_composition_layout_action.go
//
// ResolveCompositionLayoutAction is site-design-planner's layout picker. It
// runs inside the site-design-planner workflow after ValidateCompositionInputs
// has confirmed identity + classification specs exist.
//
// Separation of concerns:
//   - The matching LOGIC lives in `resolveLayoutByTagsWeighted`
//     (fork_theme_composition.go) — weighted, scheme-aware. The old
//     `resolveLayoutByTags(ctx, tx, category, tags, logger)` is kept there as a
//     backwards-compatible shim for the fork path.
//   - This action is the thin workflow wrapper: extract inputs, derive the
//     light/dark scheme from the design brief, call the matcher, shape the
//     output, and emit the library-growth work item when the library had no
//     same-scheme match (a hard fallback, or a scheme gap).
//
// On fallback (no layout scored above zero) OR a scheme gap (only an
// opposite-scheme layout fit), the action emits a `needs_new_layout_candidate`
// work item. This is the library-growth signal — a reviewer decides whether a
// new layout (e.g. a light variant) is warranted. Same durable-work-item
// pattern as the classifier recovery item in validate_composition_inputs.
//
// Inputs:
//   - site_id (path-resolved, required)
//
// Config literals:
//   - classification_source (optional) — path to classification data in
//     collected_data. Default: "validated_inputs.classification".
//
// Returns:
//   {
//     "layout_id":           "uuid-string",
//     "layout_name":         "tool-portal-light",
//     "reason":              "weighted match: score 3.41 …",
//     "candidates":          ["tool-portal-light", "tool-portal-dark", ...],
//     "is_fallback":         false,
//     "scheme":              "light",   // chosen layout's scheme ("" if unknown)
//     "is_scheme_mismatch":  false,     // true if we had to cross schemes
//     "site_tags":           ["interactive", "tool-portal", "founder-tools"],
//     "review_item_queued":  null | "uuid-string",
//   }
//
// Registration (registry.go) unchanged:
//
//   "resolve_composition_layout": {
//       Handler:     ResolveCompositionLayoutAction,
//       Category:    "site",
//       Description: "Pick a library layout by weighted, scheme-aware match",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ResolveCompositionLayoutInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec(
		"resolve_composition_layout",
		ResolveCompositionLayoutInputSpec,
	)
}

// ResolveCompositionLayoutAction is the workflow entry point.
func ResolveCompositionLayoutAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "resolve_composition_layout"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ResolveCompositionLayoutInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// Domain for logging and work item summary — best-effort
	var domain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)

	// Extract category + industry_tags from either the prior step's output
	// (validated_inputs.classification, set by validate_composition_inputs)
	// or by reading the spec directly.
	category, industryTags, err := extractClassificationTags(
		ctx, params, siteID, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("extract classification tags: %w", err)
	}

	// Derive the light/dark scheme from the design brief so the matcher won't
	// place a light site on a dark layout (or vice-versa) on tag overlap alone.
	siteScheme := deriveSiteScheme(ctx, params, siteID, logger)

	// Open a short transaction. The shared resolver takes *sql.Tx because it
	// is used by the fully-transactional fork path. Here we use it read-only
	// and commit without writes.
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	resolution, err := resolveLayoutByTagsWeighted(ctx, tx, category, industryTags, siteScheme, logger)
	if err != nil {
		return nil, fmt.Errorf("resolveLayoutByTagsWeighted failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit read-only tx: %w", err)
	}

	siteTagsForOutput := collectNormalisedSiteTags(category, industryTags)

	// Emit a library-growth work item when the library had no same-scheme
	// match: either a hard fallback, or a scheme gap (an opposite-scheme
	// layout was applied). Both mean "the library is missing a good fit".
	var reviewItemQueued interface{}
	if resolution.IsFallback || resolution.IsSchemeMismatch {
		if resolution.IsSchemeMismatch && !resolution.IsFallback {
			logger.Warn("ResolveCompositionLayoutAction: scheme gap — no same-scheme layout fit; applied a cross-scheme match",
				zap.String("site_id", siteID.String()),
				zap.String("domain", domain),
				zap.String("site_scheme", siteScheme),
				zap.String("applied_layout", resolution.LayoutName),
				zap.String("applied_layout_scheme", resolution.Scheme),
				zap.Strings("site_tags", siteTagsForOutput),
				zap.Strings("candidates_considered", resolution.Candidates),
				zap.String("recommendation", fmt.Sprintf("consider adding a %s layout for these tags", schemeLabel(siteScheme))),
			)
		} else {
			logger.Error("ResolveCompositionLayoutAction: no library layout matched — fell back to brochure-formal",
				zap.String("site_id", siteID.String()),
				zap.String("domain", domain),
				zap.String("site_scheme", siteScheme),
				zap.Strings("site_tags", siteTagsForOutput),
				zap.Strings("candidates_considered", resolution.Candidates),
				zap.String("recommendation", "consider adding a new layout to the library"),
			)
		}
		itemID, qerr := queueLayoutCandidateReview(
			ctx, params.DB,
			siteID, domain, siteScheme,
			resolution.Reason, resolution.LayoutName,
			siteTagsForOutput, resolution.Candidates,
			logger,
		)
		if qerr != nil {
			logger.Error("Failed to queue needs_new_layout_candidate item",
				zap.Error(qerr),
				zap.String("site_id", siteID.String()),
			)
		}
		if itemID != nil {
			reviewItemQueued = itemID.String()
		}
	}

	return map[string]interface{}{
		"layout_id":          resolution.LayoutID.String(),
		"layout_name":        resolution.LayoutName,
		"reason":             resolution.Reason,
		"candidates":         resolution.Candidates,
		"is_fallback":        resolution.IsFallback,
		"scheme":             resolution.Scheme,
		"is_scheme_mismatch": resolution.IsSchemeMismatch,
		"site_tags":          siteTagsForOutput,
		"review_item_queued": reviewItemQueued,
	}, nil
}

// deriveSiteScheme reads the light/dark intent from the design brief:
// design_intent.style_direction (primary) + classification.suggested_style
// (secondary). Returns "light" | "dark" | "" (unknown -> no constraint).
// deriveSchemeFromDesignIntent lives in fork_theme_composition.go (same package).
func deriveSiteScheme(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) string {
	var styleDirection, suggestedStyle string
	if di, found, err := loadCurrentSpecData(ctx, params.DB, siteID, "design_intent"); err == nil && found {
		styleDirection, _ = di["style_direction"].(string)
	}
	if cl, found, err := loadCurrentSpecData(ctx, params.DB, siteID, "classification"); err == nil && found {
		suggestedStyle, _ = cl["suggested_style"].(string)
	}
	s := deriveSchemeFromDesignIntent(styleDirection, suggestedStyle)
	logger.Info("deriveSiteScheme",
		zap.String("style_direction", styleDirection),
		zap.String("suggested_style", suggestedStyle),
		zap.String("scheme", s),
	)
	return s
}

// schemeLabel renders "" as "a same-scheme" for log readability.
func schemeLabel(s string) string {
	if s == "" {
		return "matching-scheme"
	}
	return s
}

// extractClassificationTags pulls `category` and `industry_tags` from
// classification data. Prefers validated_inputs.classification (set by the
// earlier validate_composition_inputs step), otherwise reads fresh from
// site_specs.
func extractClassificationTags(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	logger *zap.Logger,
) (string, []string, error) {

	classPath := "validated_inputs.classification"
	if cs, ok := params.StepConfig.Config["classification_source"].(string); ok && cs != "" {
		classPath = cs
	}

	var classData map[string]interface{}
	classRaw := datahelpers.ExtractNestedField(params.CollectedData, classPath)
	if classRaw != nil {
		unwrapped := datahelpers.UnwrapDeep(classRaw, logger)
		if m, ok := unwrapped.(map[string]interface{}); ok {
			classData = m
		}
	}

	if len(classData) == 0 {
		logger.Info("classification data not in collected_data, reading from site_specs",
			zap.String("site_id", siteID.String()),
			zap.String("tried_path", classPath),
		)
		data, found, err := loadCurrentSpecData(ctx, params.DB, siteID, "classification")
		if err != nil {
			return "", nil, fmt.Errorf("read classification spec: %w", err)
		}
		if !found {
			return "", nil, fmt.Errorf(
				"classification spec not found for site %s — "+
					"should have been caught by validate_composition_inputs",
				siteID,
			)
		}
		classData = data
	}

	category, _ := classData["category"].(string)

	var tags []string
	if raw, ok := classData["industry_tags"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, t := range v {
				if s, ok := t.(string); ok {
					tags = append(tags, s)
				}
			}
		case []string:
			tags = append(tags, v...)
		}
	}

	logger.Info("extractClassificationTags",
		zap.String("category", category),
		zap.Strings("industry_tags", tags),
	)
	return category, tags, nil
}

// collectNormalisedSiteTags mirrors the tag-set construction in
// resolveLayoutByTags so the action's output reports exactly what was
// matched against.
func collectNormalisedSiteTags(category string, industryTags []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, 1+len(industryTags))
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			return
		}
		if _, dup := seen[s]; dup {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if category != "" && strings.ToLower(category) != "general" {
		add(category)
	}
	for _, t := range industryTags {
		add(t)
	}
	return out
}

// queueLayoutCandidateReview inserts a needs_new_layout_candidate work item.
// Non-fatal: if the queue fails, the main composition path continues with
// the applied layout.
//
// Widened to record the real situation: `reason` and `appliedLayout` come from
// the resolution (so the item is accurate for both the hard-fallback case —
// applied brochure-formal — and the scheme-gap case — applied a cross-scheme
// layout), plus the derived `siteScheme`.
//
// item_key = "needs_new_layout_candidate" (site-scoped by the partial unique
// index). Repeated misses for the same site consolidate into one active item;
// the two-strike rule surfaces persistently-bad classifications as `unresolved`.
func queueLayoutCandidateReview(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	domain string,
	siteScheme string,
	reason string,
	appliedLayout string,
	siteTags []string,
	candidatesConsidered []string,
	logger *zap.Logger,
) (*uuid.UUID, error) {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	spec := map[string]interface{}{
		"domain":                domain,
		"reason":                reason,
		"site_scheme":           siteScheme,
		"queued_by":             "site-design-planner:resolve_composition_layout",
		"site_tags":             siteTags,
		"candidates_considered": candidatesConsidered,
		"applied_layout":        appliedLayout,
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	scheme := siteScheme
	if scheme == "" {
		scheme = "unknown"
	}

	// handler "hitl-review" + status "needs_human_review" matches the existing
	// tool-auditor → HITL pattern. `needs_human_review` is a first-class status:
	// the dispatch loop skips these items, so the bogus handler doesn't cause the
	// "handler not registered → blocked" flip in ClaimWorkItemAction. A human
	// resolves it via admin API (PATCH /work-items/:id): add a layout and retry,
	// or mark wont_fix.
	item := workItem{
		siteID:       siteID,
		source:       "side_effect",
		pipeline:     "build",
		itemType:     "needs_new_layout_candidate",
		severity:     "low",
		summary:      fmt.Sprintf("Layout gap for %s (scheme=%s) — applied %s; review classification or add a layout", domain, scheme, appliedLayout),
		spec:         string(specJSON),
		priority:     80,
		handlerAgent: "hitl-review",
		status:       "needs_human_review",
		createdBy:    "site-design-planner",
		itemKey:      "needs_new_layout_candidate",
	}

	inserted, err := insertWorkItem(ctx, tx, item, logger)
	if err != nil {
		return nil, fmt.Errorf("insert work item: %w", err)
	}
	if !inserted {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit tx (suppressed): %w", err)
		}
		return nil, nil
	}

	var newItemID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM site_work_items
		WHERE site_id = $1
		  AND item_key = 'needs_new_layout_candidate'
		  AND status NOT IN ('complete','verified','rejected','wont_fix','failed')
		ORDER BY created_at DESC
		LIMIT 1
	`, siteID).Scan(&newItemID)
	if err != nil {
		return nil, fmt.Errorf("fetch inserted item id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	logger.Info("Queued needs_new_layout_candidate review item",
		zap.String("site_id", siteID.String()),
		zap.String("new_item_id", newItemID.String()),
		zap.String("site_scheme", scheme),
		zap.String("applied_layout", appliedLayout),
	)
	return &newItemID, nil
}
