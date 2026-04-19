// FILE: platform/orchestration/actions/resolve_composition_layout_action.go
//
// ResolveCompositionLayoutAction is site-design-planner's layout picker. It
// runs inside the site-design-planner workflow after ValidateCompositionInputs
// has confirmed identity + classification specs exist.
//
// Separation of concerns:
//   - The tag-overlap matching LOGIC lives in `resolveLayoutByTags`
//     (fork_theme_composition.go) — shared with fork_theme_from_site.
//   - This action is the thin workflow wrapper: extract inputs, call the
//     shared helper, shape the output, and emit the library-growth work
//     item on fallback.
//
// On fallback (no layout scored above zero overlap), the action emits a
// `needs_new_layout_candidate` work item. This is the library-growth
// signal — when a site's classification didn't match any seeded layout,
// a reviewer should see it and decide whether a new layout is warranted.
// Same pattern as the classifier recovery item in validate_composition_inputs:
// loud log + durable work item via insertWorkItem.
//
// Inputs:
//   - site_id (path-resolved, required)
//
// Config literals:
//   - classification_source (optional) — path to classification data in
//     collected_data. Default: "validated_inputs.classification" (the
//     output field from validate_composition_inputs). Falls back to a
//     direct site_specs read if the path yields nothing.
//
// Returns:
//   {
//     "layout_id":           "uuid-string",
//     "layout_name":         "brochure-formal",
//     "reason":              "best tag overlap (2 match) …",
//     "candidates":          ["utility-tool", "docs-sidebar", "brochure-formal"],
//     "is_fallback":         false,
//     "site_tags":           ["corporate", "consulting"],
//     "review_item_queued":  null | "uuid-string",
//   }
//
// Registration (add to registry.go):
//
//   "resolve_composition_layout": {
//       Handler:     ResolveCompositionLayoutAction,
//       Category:    "site",
//       Description: "Pick a library layout by tag-overlap against classification",
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

	// Open a short transaction. The shared resolver takes *sql.Tx because it
	// is used by the fully-transactional fork path. Here we use it read-only
	// and commit without writes.
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	resolution, err := resolveLayoutByTags(ctx, tx, category, industryTags, logger)
	if err != nil {
		return nil, fmt.Errorf("resolveLayoutByTags failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit read-only tx: %w", err)
	}

	siteTagsForOutput := collectNormalisedSiteTags(category, industryTags)

	// On fallback, emit a library-growth work item so a reviewer sees
	// that the library didn't have a good match.
	var reviewItemQueued interface{}
	if resolution.IsFallback {
		logger.Error("ResolveCompositionLayoutAction: no library layout matched — fell back to brochure-formal",
			zap.String("site_id", siteID.String()),
			zap.String("domain", domain),
			zap.Strings("site_tags", siteTagsForOutput),
			zap.Strings("candidates_considered", resolution.Candidates),
			zap.String("recommendation", "consider adding a new layout to the library"),
		)
		itemID, qerr := queueLayoutCandidateReview(
			ctx, params.DB,
			siteID, domain,
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
		"site_tags":          siteTagsForOutput,
		"review_item_queued": reviewItemQueued,
	}, nil
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
// the fallback layout.
//
// item_key = "needs_new_layout_candidate" (site-scoped by the partial
// unique index). Repeated fallbacks for the same site consolidate into
// one active item; the two-strike rule surfaces persistently-bad
// classifications as `unresolved` for investigation.
func queueLayoutCandidateReview(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	domain string,
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
		"reason":                "no library layout matched site classification",
		"queued_by":             "site-design-planner:resolve_composition_layout",
		"site_tags":             siteTags,
		"candidates_considered": candidatesConsidered,
		"fallback_applied":      "brochure-formal",
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	item := workItem{
		siteID:       siteID,
		source:       "side_effect",
		pipeline:     "build",
		itemType:     "needs_new_layout_candidate",
		severity:     "low",
		summary:      fmt.Sprintf("No library layout matched %s — review classification or add a new layout", domain),
		spec:         string(specJSON),
		priority:     80,
		handlerAgent: "human-review",
		status:       "triaged",
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
	)
	return &newItemID, nil
}
