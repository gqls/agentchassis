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
//   - Usage (battle-tested components score higher, with diminishing returns) — DERIVED
//     from page_components bindings, never a stored counter; see ComponentUsageSitesSQL
//     and bugs_open/378 for why the stored usage_count column was abandoned
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

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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

// ============================================================================
// How proven is this component — ONE definition, derived, never maintained
// ============================================================================

// ComponentUsageSitesSQL is THE definition of "how proven is this component" for
// every consumer in this estate. It is a scalar sub-select intended to be spliced
// into a query whose FROM is (unaliased) content_components.
//
// It counts DISTINCT SITES the component is bound to, from page_components — the
// durable record of the binding — and is deliberately NOT a stored counter.
//
// WHY DERIVED (bugs_open/378). The column it replaces, content_components.usage_count,
// was incremented by exactly one helper reachable from exactly one of the THREE
// resolution paths in plan_sections_action.go's section loop (stored component_id,
// name/function match, section_type selector). A component bound to a live page by
// either of the other two was never counted, while the same column was read as a
// merit signal — so it recorded which ROUTE found a component, not whether it is any
// good. It also over-counted: the increment fired before planSection decided
// ready/deferred/skipped and again on every re-plan, so [MEASURED 2026-08-24] the two
// largest values in the column were both components with ZERO page bindings
// (testimonials-modern: 12 counts, born 2026-08-23, never bound; and a retired
// _pre_037 backup copy at 20).
//
// A maintained counter was rejected on a census, not on taste: there were SEVEN
// INSERT INTO page_components sites as of 2026-08-24 (save_page_sections,
// deploy_tool, create_tool_component, create_report_page, rebuild_blog_listing,
// adopt_verbatim, cmd/webdesignport/import). That is seven places to forget and an
// eighth next month. Derived at read, there is no counter to drift.
//
// WHY DISTINCT SITES rather than raw binding rows [MEASURED 2026-08-24, section level]:
// raw bindings run max 414 / median 1, so normalising them yields a near-binary signal
// that mostly reports "is this on a big site". Distinct sites run max 27 / p90 9 — a
// real spread, and it says what "battle-tested" is supposed to mean: proven across
// different sites, not repeated down one site's pages. It also dilutes a known data
// defect: bugs_open/357's mis-bound rows (22 rows declaring `hero` while storing a whole
// tool page, a population still being minted) collapse to just 3 of hero's 27 sites.
//
// build_status='removed' is excluded — a removed binding is not a use.
//
// ⚠ SWITCHING THIS TO THE PROVENANCE STAMP LATER (register CLC-026). A strictly more
// honest per-row signal exists: page_components.component_version_id, which proves the
// component produced those bytes. It is NOT used here yet, and the reason is not doubt
// about the mechanism — nothing backfills it, so at today's coverage it would see 39 of
// 151 active section components against 108 for bindings, and it would measure how
// RECENTLY a component was rebuilt rather than how proven it is. That is precisely the
// defect above, one epoch over. The condition for switching is NOT a coverage
// percentage: coverage approaches but never reaches 100% by design, because a page that
// is never rebuilt is never stamped, so the last population to convert is the long-stable
// component — exactly what a merit signal must credit. Switch when the unstamped
// remainder is small enough to be a stated exception rather than the majority. When that
// day comes, this constant is the only edit.
// var, not const, since 2026-08-26: the tombstone exclusion is
// datahelpers.NotRemoved — NULL-safe to match the assembler's serving
// predicate (council 89dcc04a; a NULL-status row is served, so it IS a use).
var ComponentUsageSitesSQL = `(
			SELECT count(DISTINCT p.site_id)
			  FROM page_components pc
			  JOIN pages p ON p.id = pc.page_id
			 WHERE pc.component_id = content_components.id
			   AND ` + datahelpers.NotRemoved("pc") + `
		)`

// THE USAGE TERM IS NO LONGER PART OF THE SCORE, and that is a deliberate decision
// rather than an omission (bugs_open/378). The selector's SELECT list still reports the
// derived figure — it is honest now, and worth logging — but nothing scores on it.
//
// The obvious fix was to keep the term and feed it the corrected number. That was built
// first and then withdrawn on its own measurement. Simulated over the 4,888 contested
// (section_type, site_type, page_type) contexts the live library actually presents:
//
//   removing the term entirely ...........   0 winners change
//   feeding it the corrected number ....... 3,246 winners change, across 3 section_types
//     features                    features              -> differentiators-section
//     tool-archetype-taster-quiz  archetype-taster-quiz -> tool-archetype-taster-quiz
//     hero                        case-studies-hero     -> contact-hero
//
// The old term changed nothing because every candidate in every contest read 0 — the 12
// components carrying a count were all the sole candidate for their section_type. So
// "repair the term" is not the conservative option here; it is by far the larger
// behavioural change, and it is one this lane has no evidence is an improvement.
//
// The deeper reason is that a working version of this term is a PREFERENTIAL ATTACHMENT
// loop: selected -> count rises -> scores higher -> selected again. bugs_open/107 ("every
// site gets the same homepage skeleton") is the open complaint about exactly that outcome,
// and it cites this file. Repairing the term would have switched that loop on — and it
// switches on in favour of whichever component is already the incumbent, which is the one
// thing 107 must not have. A term that cannot be made accurate without making the estate
// more homogeneous is not a term worth keeping in the score.
//
// What remains derived, and where it still matters: load_existing_component_action.go
// orders the CONTRACT row by ComponentUsageSitesSQL. That is a different question — which
// stored row IS this section_type's component — where "the one most sites actually use" is
// the right answer and no homogeneity loop exists, because it selects a row to enforce, not
// a component to put on a page.

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
			` + ComponentUsageSitesSQL + ` as usage_count,
			avg_quality_score,
			(
				CASE WHEN suitable_site_types @> to_jsonb($2::text) THEN 0.35 ELSE 0.05 END
				+ CASE WHEN suitable_page_types @> to_jsonb($3::text) THEN 0.15 ELSE 0.0 END
				+ COALESCE(avg_quality_score, 0.3) * 0.3
				+ CASE WHEN COALESCE(jsonb_array_length(suitable_site_types), 0) BETWEEN 1 AND 3
				       THEN 0.1 ELSE 0.02 END
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
			`+ComponentUsageSitesSQL+` as usage_count,
			avg_quality_score,
			(
				CASE WHEN suitable_site_types @> to_jsonb($%d::text) THEN 0.35 ELSE 0.05 END
				+ CASE WHEN suitable_page_types @> to_jsonb($%d::text) THEN 0.15 ELSE 0.0 END
				+ COALESCE(avg_quality_score, 0.3) * 0.3
				+ CASE WHEN COALESCE(jsonb_array_length(suitable_site_types), 0) BETWEEN 1 AND 3
				       THEN 0.1 ELSE 0.02 END
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

	// Backstop guard (bugs_open/041): never raise needs_new_component for a
	// section that already resolves under kebab-normalisation. The library stores
	// kebab-case (content_components.function); a snake_case/CamelCase request
	// that reached here is a naming mismatch, not a missing component. Raising the
	// item asks a creator to rebuild a component that already exists — that is
	// exactly the 10 failed items across 4 sites this bug catalogued.
	// loadSectionComponents + loadComponentSchemas now resolve these upstream, so
	// this branch should no longer be reached for such names; this is the last
	// line of defence for any future caller that bypasses them.
	if norm := NormalizeComponentFunction(sectionType); norm != sectionType {
		var resolves bool
		if qErr := db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM content_components
				WHERE is_active = true
				  AND (lower(function) = lower($1)
				       OR lower(name) = lower($1)
				       OR lower(section_type) = lower($1))
			)`, norm).Scan(&resolves); qErr != nil {
			logger.Warn("component_selector: normalisation guard check failed, proceeding",
				zap.String("section_type", sectionType),
				zap.Error(qErr))
		} else if resolves {
			logger.Warn("component_selector: NOT raising needs_new_component — section resolves under kebab-normalisation (naming mismatch, not a missing component)",
				zap.String("requested", sectionType),
				zap.String("normalised", norm))
			return nil
		}
	}

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
