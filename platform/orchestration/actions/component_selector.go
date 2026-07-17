// FILE: platform/orchestration/actions/component_selector.go
//
// ComponentSelector resolves section_type names to concrete content_components.
//
// Used by plan_sections_action.go when a section name doesn't match any
// existing component function directly. The selector queries by section_type,
// scores candidates against the site context, and returns the best match.
//
// The scoring considers:
//   - Site type relevance (does the component declare this site type as suitable?)
//   - Page type relevance (does the component suit this page type?)
//   - Quality score (auditor feedback, NULL = unproven gets a neutral score)
//   - Usage count (battle-tested components score higher, with diminishing returns)
//   - Specificity bonus (components targeting fewer site types are more specialised)
//
// If no component matches, returns nil — the caller creates a needs_new_component
// work item for the component-creator handler.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ComponentCandidate represents a scored component from the selector query.
type ComponentCandidate struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Function        string   `json:"function"`
	DisplayName     string   `json:"display_name"`
	SectionType     string   `json:"section_type"`
	Category        string   `json:"category"`
	IsDarkSection   bool     `json:"is_dark_section"`
	Score           float64  `json:"score"`
	UsageCount      int      `json:"usage_count"`
	AvgQualityScore *float64 `json:"avg_quality_score"`
}

// SelectorContext provides the site and page context for scoring.
type SelectorContext struct {
	SiteType string // e.g. "brochure", "interactive-platform"
	PageType string // e.g. "landing", "about", "services"
	PageName string // e.g. "index", "about"
}

// SelectComponentByType queries content_components for the best match for a
// given section_type within the provided context.
//
// Returns the best candidate, or nil if no components match.
// The caller should check for nil and handle the "no component" case
// (typically by creating a needs_new_component work item).
func SelectComponentByType(
	ctx context.Context,
	db *sql.DB,
	sectionType string,
	selCtx SelectorContext,
	logger *zap.Logger,
) (*ComponentCandidate, error) {

	candidates, err := queryCandidates(ctx, db, sectionType, selCtx, logger)
	if err != nil {
		return nil, fmt.Errorf("selector query failed for section_type %q: %w", sectionType, err)
	}

	if len(candidates) == 0 {
		logger.Info("component_selector: no candidates found",
			zap.String("section_type", sectionType),
			zap.String("site_type", selCtx.SiteType),
			zap.String("page_type", selCtx.PageType))
		return nil, nil
	}

	// Best candidate is first (query orders by score DESC)
	best := candidates[0]

	logger.Info("component_selector: selected component",
		zap.String("section_type", sectionType),
		zap.String("selected_function", best.Function),
		zap.Float64("score", best.Score),
		zap.Int("candidates_found", len(candidates)),
		zap.String("site_type", selCtx.SiteType))

	return &best, nil
}

// SelectComponentsByTypes resolves multiple section_types in one pass.
// Returns a map of section_type → best candidate (nil value means no match).
// This is more efficient than calling SelectComponentByType in a loop
// because it batches the DB queries.
func SelectComponentsByTypes(
	ctx context.Context,
	db *sql.DB,
	sectionTypes []string,
	selCtx SelectorContext,
	logger *zap.Logger,
) (map[string]*ComponentCandidate, error) {

	result := make(map[string]*ComponentCandidate, len(sectionTypes))

	// Query all candidates in one go
	allCandidates, err := queryCandidatesBatch(ctx, db, sectionTypes, selCtx, logger)
	if err != nil {
		return nil, fmt.Errorf("batch selector query failed: %w", err)
	}

	// Group by section_type, pick best per type
	for _, sectionType := range sectionTypes {
		candidates := allCandidates[sectionType]
		if len(candidates) == 0 {
			result[sectionType] = nil
			logger.Info("component_selector: no candidates for section_type",
				zap.String("section_type", sectionType))
		} else {
			result[sectionType] = &candidates[0]
			logger.Info("component_selector: selected",
				zap.String("section_type", sectionType),
				zap.String("function", candidates[0].Function),
				zap.Float64("score", candidates[0].Score))
		}
	}

	return result, nil
}

// IncrementUsageCount bumps the usage_count for a component after it's been
// assigned to a page. Called by plan_sections after selection succeeds.
func IncrementUsageCount(ctx context.Context, db *sql.DB, componentID string, logger *zap.Logger) {
	_, err := db.ExecContext(ctx, `
		UPDATE content_components
		SET usage_count = COALESCE(usage_count, 0) + 1,
		    updated_at = NOW()
		WHERE id = $1
	`, componentID)
	if err != nil {
		logger.Warn("component_selector: failed to increment usage_count",
			zap.String("component_id", componentID),
			zap.Error(err))
	}
}

// ============================================================================
// Query and scoring
// ============================================================================

// queryCandidates finds and scores components for a single section_type.
func queryCandidates(
	ctx context.Context,
	db *sql.DB,
	sectionType string,
	selCtx SelectorContext,
	logger *zap.Logger,
) ([]ComponentCandidate, error) {

	// Always pass all three parameters — empty string scores low, which is correct.
	siteType := selCtx.SiteType
	pageType := selCtx.PageType

	query := `
		SELECT
			id::text,
			name,
			function,
			COALESCE(display_name, name) as display_name,
			section_type,
			COALESCE(category, '') as category,
			COALESCE(is_dark_section, false) as is_dark_section,
			COALESCE(usage_count, 0) as usage_count,
			avg_quality_score,
			(
				CASE WHEN suitable_site_types @> to_jsonb($2::text) THEN 0.35 ELSE 0.05 END
				+ CASE WHEN suitable_page_types @> to_jsonb($3::text) THEN 0.15 ELSE 0.0 END
				+ COALESCE(avg_quality_score, 0.3) * 0.3
				+ CASE WHEN COALESCE(jsonb_array_length(suitable_site_types), 0) BETWEEN 1 AND 3
				       THEN 0.1 ELSE 0.02 END
				+ LEAST(COALESCE(usage_count, 0)::float / 50.0, 1.0) * 0.1
			) as score
		FROM content_components
		WHERE section_type = $1
		  AND component_level = 'section'
		  AND is_active = true
		  AND forked_from IS NULL
		ORDER BY score DESC
		LIMIT 5
	`

	return executeAndScan(ctx, db, query, []interface{}{sectionType, siteType, pageType}, logger)
}

// queryCandidatesBatch finds candidates for multiple section_types at once.
func queryCandidatesBatch(
	ctx context.Context,
	db *sql.DB,
	sectionTypes []string,
	selCtx SelectorContext,
	logger *zap.Logger,
) (map[string][]ComponentCandidate, error) {

	if len(sectionTypes) == 0 {
		return make(map[string][]ComponentCandidate), nil
	}

	// Build IN clause for section_types
	placeholders := make([]string, len(sectionTypes))
	args := make([]interface{}, len(sectionTypes))
	for i, st := range sectionTypes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = st
	}

	argIdx := len(sectionTypes) + 1

	query := fmt.Sprintf(`
		SELECT
			id::text,
			name,
			function,
			COALESCE(display_name, name) as display_name,
			section_type,
			COALESCE(category, '') as category,
			COALESCE(is_dark_section, false) as is_dark_section,
			COALESCE(usage_count, 0) as usage_count,
			avg_quality_score,
			(
				CASE WHEN suitable_site_types @> to_jsonb($%d::text) THEN 0.35 ELSE 0.05 END
				+ CASE WHEN suitable_page_types @> to_jsonb($%d::text) THEN 0.15 ELSE 0.0 END
				+ COALESCE(avg_quality_score, 0.3) * 0.3
				+ CASE WHEN COALESCE(jsonb_array_length(suitable_site_types), 0) BETWEEN 1 AND 3
				       THEN 0.1 ELSE 0.02 END
				+ LEAST(COALESCE(usage_count, 0)::float / 50.0, 1.0) * 0.1
			) as score
		FROM content_components
		WHERE section_type IN (%s)
		  AND component_level = 'section'
		  AND is_active = true
		  AND forked_from IS NULL
		ORDER BY section_type, score DESC
	`, argIdx, argIdx+1, strings.Join(placeholders, ","))

	// Add site_type and page_type as the last args
	if selCtx.SiteType != "" {
		args = append(args, selCtx.SiteType)
	} else {
		args = append(args, "")
	}
	if selCtx.PageType != "" {
		args = append(args, selCtx.PageType)
	} else {
		args = append(args, "")
	}

	candidates, err := executeAndScan(ctx, db, query, args, logger)
	if err != nil {
		return nil, err
	}

	// Group by section_type
	grouped := make(map[string][]ComponentCandidate)
	for _, c := range candidates {
		grouped[c.SectionType] = append(grouped[c.SectionType], c)
	}

	return grouped, nil
}

// executeAndScan runs the query and scans into ComponentCandidate structs.
func executeAndScan(
	ctx context.Context,
	db *sql.DB,
	query string,
	args []interface{},
	logger *zap.Logger,
) ([]ComponentCandidate, error) {

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("selector query failed: %w", err)
	}
	defer rows.Close()

	var candidates []ComponentCandidate
	for rows.Next() {
		var c ComponentCandidate
		var avgScore sql.NullFloat64

		if err := rows.Scan(
			&c.ID, &c.Name, &c.Function, &c.DisplayName,
			&c.SectionType, &c.Category, &c.IsDarkSection,
			&c.UsageCount, &avgScore, &c.Score,
		); err != nil {
			logger.Warn("component_selector: scan error", zap.Error(err))
			continue
		}

		if avgScore.Valid {
			c.AvgQualityScore = &avgScore.Float64
		}

		candidates = append(candidates, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("selector rows error: %w", err)
	}

	return candidates, nil
}

// ============================================================================
// Work item creation for missing section types
// ============================================================================

// CreateNeedsNewComponentItem creates a site_work_item for a section_type
// that has no matching component in the library. The component-creator
// handler will process this item to generate a new template.
//
// Uses check-first dedup (SELECT EXISTS then INSERT) instead of ON CONFLICT
// to avoid partial unique index matching issues with idx_swi_dedup.
func CreateNeedsNewComponentItem(
	ctx context.Context,
	db *sql.DB,
	siteID string,
	sectionType string,
	pageContext string,
	description string,
	designDirection string,
	siteType string,
	logger *zap.Logger,
) error {

	// item_key ensures dedup — only one work item per section_type per site
	itemKey := fmt.Sprintf("needs_new_component:%s", sectionType)

	// Check if an active item already exists (same logic as idx_swi_dedup;
	// status list from workItemTerminalStatuses so it tracks the index).
	var exists bool
	err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM site_work_items
			WHERE site_id = $1::uuid
			  AND item_key = $2
			  AND status NOT IN (%s)
		)
	`, sqlInList(workItemTerminalStatuses)), siteID, itemKey).Scan(&exists)
	if err != nil {
		logger.Warn("component_selector: dedup check failed, attempting insert anyway",
			zap.String("section_type", sectionType),
			zap.Error(err))
	}

	if exists {
		logger.Info("component_selector: needs_new_component item already exists, skipping",
			zap.String("section_type", sectionType),
			zap.String("item_key", itemKey))
		return nil
	}

	// Build the spec for the component-creator
	spec := map[string]interface{}{
		"section_type":     sectionType,
		"site_type":        siteType,
		"page_context":     pageContext,
		"description":      description,
		"design_direction": designDirection,
	}

	specJSON, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal component spec: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by, item_key
		) VALUES (
			$1::uuid, 'component_selector', 'build', 'needs_new_component', 'medium',
			$2, $3::jsonb, 50, 'component-creator', 'triaged', 'component_selector', $4
		)
	`, siteID,
		fmt.Sprintf("Need component template for section type: %s", sectionType),
		string(specJSON),
		itemKey,
	)

	if err != nil {
		return fmt.Errorf("failed to create needs_new_component work item: %w", err)
	}

	logger.Info("component_selector: created needs_new_component work item",
		zap.String("section_type", sectionType),
		zap.String("site_id", siteID),
		zap.String("item_key", itemKey))

	return nil
}
