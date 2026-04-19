package actions

// FILE: platform/orchestration/actions/fork_theme_composition.go
//
// Helpers used by ForkThemeFromSiteAction (Phase 5) to produce
// palette + typography_set rows from an adopted site's design_spec
// and to resolve which layout best fits the site's classification.
//
// Called once per adoption fork, inside the fork's transaction.
// All three resolutions happen before the css_themes INSERT so the
// three FKs are known values when the insert runs.
//
// Design: palette is always new (site colours are site-specific);
// typography_set is matched-or-new (fonts are shared more often);
// layout is matched-or-default (pick from existing library by
// classification tag match, fall back to brochure-formal).

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// --- Palette creation ---

// createPaletteForFork inserts a new palettes row from the site's
// design_spec.color_scheme, with origin='adopted', needs_review=true,
// and source lineage populated. Always creates a new row — site
// colours are site-specific and library palette reuse would be lying.
//
// Runs inside the caller's transaction. Returns the new palette's id.
//
// The inserted row's `colours` JSONB contains every string-valued
// entry from color_scheme. Core palette keys (primary, secondary,
// accent, background, surface, text, text_muted, border) are all
// included if present. Non-core keys present in color_scheme are
// ALSO included — specialised slots a site declares propagate, the
// renderer's merge rules will then treat them as theme-owned.
func createPaletteForFork(
	ctx context.Context,
	tx *sql.Tx,
	baseName string,
	displayName string,
	category string,
	industryTags []string,
	designSpec map[string]interface{},
	siteID uuid.UUID,
	domain string,
	parentPaletteID *uuid.UUID,
	logger *zap.Logger,
) (uuid.UUID, error) {

	// Extract the colour map from design_spec.
	coloursMap := extractStringMap(designSpec, "color_scheme")
	if len(coloursMap) == 0 {
		return uuid.Nil, fmt.Errorf(
			"cannot fork palette: design_spec has no color_scheme entries",
		)
	}

	coloursJSON, err := json.Marshal(coloursMap)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal palette colours: %w", err)
	}

	paletteName, err := resolveUniqueNameInTx(
		ctx, tx,
		"palettes", "name",
		"palette-"+baseName,
		logger,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve palette name: %w", err)
	}

	industryTagsArr := industryTagsToTextArray(industryTags)

	var newID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO palettes (
			name, display_name, description, colours,
			category, industry_tags, is_active,
			origin, needs_review,
			forked_from_palette_id, source_site_id, source_domain, forked_at
		) VALUES (
			$1, $2, $3, $4::jsonb,
			$5, $6, true,
			'adopted', true,
			$7, $8, $9, NOW()
		) RETURNING id
	`,
		paletteName,
		displayName,
		fmt.Sprintf("Palette extracted from adopted site %s", domain),
		string(coloursJSON),
		category,
		industryTagsArr,
		parentPaletteID, siteID, domain,
	).Scan(&newID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert palette: %w", err)
	}

	logger.Info("createPaletteForFork: inserted palette",
		zap.String("palette_id", newID.String()),
		zap.String("palette_name", paletteName),
		zap.Int("colour_count", len(coloursMap)),
	)

	return newID, nil
}

// --- Typography set creation or match ---

// resolveTypographySetForFork matches the design_spec's typography
// block against existing typography_sets rows by exact font_family
// string match. If none match, creates a new row.
//
// Returns (id, matched): matched=true means an existing row was
// reused; matched=false means a new row was inserted.
//
// The match is deliberately exact-string on font_family. Normalising
// font stacks is fiddly and often wrong — "'Inter', system-ui" and
// "Inter, -apple-system, BlinkMacSystemFont" likely render the same
// Inter glyphs, but they're different stacks and shouldn't be
// silently merged. The HITL review can consolidate duplicates later.
func resolveTypographySetForFork(
	ctx context.Context,
	tx *sql.Tx,
	baseName string,
	displayName string,
	category string,
	industryTags []string,
	designSpec map[string]interface{},
	siteID uuid.UUID,
	domain string,
	logger *zap.Logger,
) (uuid.UUID, bool, error) {

	typoMap := extractStringMap(designSpec, "typography")
	if len(typoMap) == 0 {
		// Typography is optional in design_spec. Fall back to the
		// default typography_set (sans-modern). The merge rules in the
		// renderer will let design_spec override specific fields if
		// they appear later in the spec pipeline.
		var defaultID uuid.UUID
		err := tx.QueryRowContext(ctx,
			`SELECT id FROM typography_sets
			 WHERE name = 'sans-modern' AND is_active = true
			 LIMIT 1`,
		).Scan(&defaultID)
		if err != nil {
			return uuid.Nil, false, fmt.Errorf(
				"fallback to sans-modern typography_set failed: %w", err,
			)
		}
		logger.Info("resolveTypographySetForFork: no typography in spec, using sans-modern default",
			zap.String("typography_set_id", defaultID.String()),
		)
		return defaultID, true, nil
	}

	// Try to match on font_family alone (the most distinctive field).
	// heading_font defaulting to "inherit" in many design_specs makes
	// matching on both fields too strict.
	fontFamily := typoMap["font_family"]
	if fontFamily != "" {
		var matchID uuid.UUID
		err := tx.QueryRowContext(ctx, `
			SELECT id FROM typography_sets
			WHERE fonts->>'font_family' = $1
			  AND is_active = true
			ORDER BY
				CASE WHEN origin = 'seed' THEN 0 ELSE 1 END,
				created_at ASC
			LIMIT 1
		`, fontFamily).Scan(&matchID)
		if err == nil {
			logger.Info("resolveTypographySetForFork: matched existing typography_set",
				zap.String("typography_set_id", matchID.String()),
				zap.String("font_family", fontFamily),
			)
			return matchID, true, nil
		}
		if err != sql.ErrNoRows {
			return uuid.Nil, false, fmt.Errorf("typography match query failed: %w", err)
		}
	}

	// No match — create a new typography_set row.
	fontsJSON, err := json.Marshal(typoMap)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("marshal fonts: %w", err)
	}

	typoName, err := resolveUniqueNameInTx(
		ctx, tx,
		"typography_sets", "name",
		"typography-"+baseName,
		logger,
	)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("resolve typography_set name: %w", err)
	}

	industryTagsArr := industryTagsToTextArray(industryTags)

	var newID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO typography_sets (
			name, display_name, description, fonts, scale,
			category, industry_tags, is_active,
			origin, needs_review,
			source_site_id, source_domain, forked_at
		) VALUES (
			$1, $2, $3, $4::jsonb, '{}'::jsonb,
			$5, $6, true,
			'adopted', true,
			$7, $8, NOW()
		) RETURNING id
	`,
		typoName,
		displayName,
		fmt.Sprintf("Typography extracted from adopted site %s", domain),
		string(fontsJSON),
		category,
		industryTagsArr,
		siteID, domain,
	).Scan(&newID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("insert typography_set: %w", err)
	}

	logger.Info("resolveTypographySetForFork: inserted new typography_set",
		zap.String("typography_set_id", newID.String()),
		zap.String("typography_set_name", typoName),
		zap.String("font_family", fontFamily),
	)

	return newID, false, nil
}

// --- Layout resolution ---

// layoutResolution is the result of picking a layout for a forked theme.
// Carries the id plus the reasoning so the HITL work item can surface
// "we picked X because Y" to the reviewer.
type layoutResolution struct {
	LayoutID   uuid.UUID
	LayoutName string
	Reason     string   // human-readable justification
	Candidates []string // top few candidate names, for reviewer to switch to
	IsFallback bool     // true when no layout matched and brochure-formal was used
}

// resolveLayoutByTags picks a layout from the library by matching the
// given tag set against layouts.industry_tags. Falls back to
// brochure-formal on no match.
//
// Used by:
//   - fork_theme_from_site (picks layout for a forked theme)
//   - site-design-planner  (picks layout for a fresh composition)
//
// Selection is deliberately conservative. The initial library has 15
// layouts with curated industry_tags (Phase 1 seeding). A match against
// any site-supplied tag is a strong signal. On ambiguity (multiple
// layouts match), the highest-overlap layout wins; on tie, alphabetical
// by layout name for determinism.
//
// The returned resolution carries the full candidate list so callers
// (HITL work items, library-growth audit, etc.) can see what was
// considered.
func resolveLayoutByTags(
	ctx context.Context,
	tx *sql.Tx,
	category string,
	industryTags []string,
	logger *zap.Logger,
) (*layoutResolution, error) {

	// Build the full tag set: category + industryTags, lowercase, deduped.
	tagSet := make(map[string]struct{})
	if category != "" && category != "general" {
		tagSet[strings.ToLower(category)] = struct{}{}
	}
	for _, t := range industryTags {
		tt := strings.ToLower(strings.TrimSpace(t))
		if tt != "" {
			tagSet[tt] = struct{}{}
		}
	}

	if len(tagSet) == 0 {
		return fallbackLayout(ctx, tx, "no classification tags", nil, logger)
	}

	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}

	// Score each active layout by tag overlap.
	rows, err := tx.QueryContext(ctx, `
		SELECT
			l.id,
			l.name,
			(
				SELECT COUNT(*)
				FROM unnest(l.industry_tags) t
				WHERE t = ANY($1::text[])
			) AS overlap_count
		FROM layouts l
		WHERE l.is_active = true
		ORDER BY overlap_count DESC, l.name ASC
		LIMIT 5
	`, tags)
	if err != nil {
		return nil, fmt.Errorf("layout match query failed: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		id      uuid.UUID
		name    string
		overlap int
	}
	var top []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.name, &c.overlap); err != nil {
			return nil, fmt.Errorf("layout row scan: %w", err)
		}
		top = append(top, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("layout rows iteration: %w", err)
	}

	if len(top) == 0 || top[0].overlap == 0 {
		var candidates []string
		for _, c := range top {
			candidates = append(candidates, c.name)
		}
		return fallbackLayout(
			ctx, tx,
			fmt.Sprintf("no layout matched site tags (%s)", strings.Join(tags, ",")),
			candidates,
			logger,
		)
	}

	picked := top[0]
	reason := fmt.Sprintf(
		"best tag overlap (%d match) between site tags [%s] and layout %q tags",
		picked.overlap,
		strings.Join(tags, ","),
		picked.name,
	)

	var candidates []string
	for _, c := range top {
		candidates = append(candidates, c.name)
	}

	logger.Info("resolveLayoutByTags: matched layout",
		zap.String("layout_id", picked.id.String()),
		zap.String("layout_name", picked.name),
		zap.Int("overlap_count", picked.overlap),
		zap.Strings("site_tags", tags),
		zap.Strings("candidates", candidates),
	)

	return &layoutResolution{
		LayoutID:   picked.id,
		LayoutName: picked.name,
		Reason:     reason,
		Candidates: candidates,
		IsFallback: false,
	}, nil
}

// fallbackLayout resolves to brochure-formal. Called when no
// classification tags matched or when the tag set was empty.
func fallbackLayout(
	ctx context.Context,
	tx *sql.Tx,
	reason string,
	candidates []string,
	logger *zap.Logger,
) (*layoutResolution, error) {

	var fb struct {
		id   uuid.UUID
		name string
	}
	err := tx.QueryRowContext(ctx, `
		SELECT id, name FROM layouts
		WHERE name = 'brochure-formal' AND is_active = true
		LIMIT 1
	`).Scan(&fb.id, &fb.name)
	if err != nil {
		return nil, fmt.Errorf(
			"fallback layout brochure-formal not found: %w — was Phase 1 seed loaded?",
			err,
		)
	}

	logger.Info("fallbackLayout: using brochure-formal",
		zap.String("reason", reason),
		zap.String("layout_id", fb.id.String()),
	)

	return &layoutResolution{
		LayoutID:   fb.id,
		LayoutName: fb.name,
		Reason:     "fallback — " + reason,
		Candidates: candidates,
		IsFallback: true,
	}, nil
}

// --- Shared utilities ---

// resolveUniqueNameInTx finds a name that doesn't exist yet in the
// given table, suffixing a short domain-safe token if the base
// collides. Used by both palette and typography_set insert paths.
//
// Operates inside the caller's transaction so a name resolved here
// remains unique through to COMMIT (no race with concurrent forks).
func resolveUniqueNameInTx(
	ctx context.Context,
	tx *sql.Tx,
	tableName, columnName, baseName string,
	logger *zap.Logger,
) (string, error) {

	// Parameterising identifiers (table/column names) in Postgres
	// requires format here — these come from trusted caller code,
	// not user input. We guard with a whitelist anyway.
	if !isValidIdentifier(tableName) || !isValidIdentifier(columnName) {
		return "", fmt.Errorf(
			"invalid table/column: %q.%q",
			tableName, columnName,
		)
	}

	// Try the base name first.
	var exists bool
	err := tx.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)`,
			tableName, columnName),
		baseName,
	).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("existence check on %s.%s: %w",
			tableName, columnName, err)
	}
	if !exists {
		return baseName, nil
	}

	// Suffix with a short random-ish token. UUID8-from-Time would
	// drift under rapid re-runs; use a trimmed UUID for simplicity.
	suffix := uuid.New().String()[:8]
	resolved := baseName + "-" + suffix
	logger.Info("resolveUniqueNameInTx: name collision, suffixing",
		zap.String("table", tableName),
		zap.String("base", baseName),
		zap.String("resolved", resolved),
	)
	return resolved, nil
}

// isValidIdentifier whitelists table/column names used by the helpers.
// Accepts lowercase letters, digits, and underscores only. Length-bounded.
func isValidIdentifier(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' {
			return false
		}
	}
	return true
}

// industryTagsToTextArray converts a []string to a value suitable for
// Postgres's text[] column type when passed as a parameter. The pq
// driver used across this codebase recognises []string natively for
// text[] columns, so we just return the slice — this function exists
// mainly for call-site clarity and for the empty-slice case.
func industryTagsToTextArray(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
