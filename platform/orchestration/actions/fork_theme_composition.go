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

// Goal: make the typography resolver reusable by site-design-planner the
// same way we refactored the layout resolver.
//
// The existing `resolveTypographySetForFork` does three things:
//   1. Extracts typography info from a design_spec map
//   2. Tries to match an existing typography_set by font_family
//   3. Inserts a new typography_sets row if no match
//
// site-design-planner has a different source for the typography signal
// (design_reference or design_intent specs, or mission hints), but wants
// the same match-or-insert behaviour. Splitting #1 off lets both callers
// share #2 and #3.
//
// New shape:
//   resolveTypographySet(ctx, tx, fonts, baseName, displayName, category,
//                        industryTags, siteID, domain, logger)
//      → uuid, matched bool, error
//
//   fonts is a plain map[string]string with keys like "font_family",
//   "heading_font", "body_font". Callers build this however they want.
//
//   resolveTypographySetForFork stays as a thin wrapper for the fork path:
//   extracts fonts from design_spec, delegates to resolveTypographySet.
//
// Behaviour is identical for the fork path. The fork action call site
// changes zero characters.
// ============================================================================

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
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
//
// now just a thin wrapper
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

	coloursMap := extractStringMap(designSpec, "color_scheme")

	return createPalette(
		ctx, tx,
		coloursMap,
		baseName, displayName,
		category, industryTags,
		siteID, domain,
		parentPaletteID,
		"adopted", true, // fork path always adopted, review-gated
		logger,
	)
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
//
// now just a thin wrapper
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

	return resolveTypographySet(
		ctx, tx,
		typoMap,
		baseName, displayName,
		category, industryTags,
		siteID, domain,
		logger,
	)
}

// --- Layout resolution ---

// layoutResolution is the result of picking a layout for a forked theme.
// Carries the id plus the reasoning so the HITL work item can surface
// "we picked X because Y" to the reviewer.
type layoutResolution struct {
	LayoutID         uuid.UUID
	LayoutName       string
	Reason           string   // human-readable justification
	Candidates       []string // top few candidate names, for reviewer to switch to
	IsFallback       bool     // true when no layout matched and brochure-formal was used
	Scheme           string   // chosen layout's scheme (light/dark/neutral/""), for the audit trail
	IsSchemeMismatch bool     // true when only an opposite-scheme layout fit (a library-gap signal)
}

// resolveLayoutByTags / resolveLayoutByTagsWeighted — weighted, scheme-aware
// layout matching. See the model note in resolveLayoutByTagsWeighted below.
// resolveLayoutByTags is kept as a backwards-compatible shim (no scheme
// constraint) for any caller that still uses the old 5-arg signature.
// --- tunables (one place to adjust scoring) ---
const (
	lmCategoryMatchBonus = 0.75 // layout.category matches a site term
	lmDescWordBonus      = 0.15 // per distinct site term found in the description
	lmDescBonusCap       = 0.90 // max description contribution
	lmSameSchemeBonus    = 0.50 // nudge toward an exactly-matching scheme
)

// canonicalTag folds known synonyms/variants to one token so the matcher
// treats them as equal. Starter controlled vocabulary — extend as tags appear.
func canonicalTag(t string) string {
	c := strings.ToLower(strings.TrimSpace(t))
	switch c {
	case "games", "game", "gaming":
		return "game-design"
	case "game-dev", "gamedev":
		return "game-development"
	case "editorial", "publication", "magazine", "editorial-content":
		return "editorial-publication"
	case "dev-tools", "developer-tool", "devtools":
		return "developer-tools"
	case "calculator", "calculators":
		return "calculators"
	case "tool", "tooling":
		return "tools"
	case "maker", "makers", "builder-tools":
		return "maker-tools"
	case "founder", "founders", "startup-tools":
		return "founder-tools"
	case "ai", "llm", "genai":
		return "ai-platform"
	case "social-network", "community", "community-platform":
		return "social-platform"
	default:
		return c
	}
}

func canonicalSet(tags []string) map[string]struct{} {
	out := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		if c := canonicalTag(t); c != "" && c != "general" {
			out[c] = struct{}{}
		}
	}
	return out
}

// deriveSchemeFromDesignIntent maps design_intent/classification style fields to
// "light"/"dark"/"" (unknown -> no constraint). Conservative: only on a clear cue.
func deriveSchemeFromDesignIntent(styleDirection, suggestedStyle string) string {
	s := strings.ToLower(styleDirection + " " + suggestedStyle)
	switch {
	case strings.Contains(s, "dark"):
		return "dark"
	case strings.Contains(s, "light"), strings.Contains(s, "warm"),
		strings.Contains(s, "editorial"), strings.Contains(s, "paper"),
		strings.Contains(s, "bright"), strings.Contains(s, "soft"):
		return "light"
	default:
		return ""
	}
}

type layoutRow struct {
	id       uuid.UUID
	name     string
	category string
	tags     []string
	scheme   string // "", "light", "dark", "neutral"
	desc     string
}

type scoredLayout struct {
	row        layoutRow
	tagScore   float64
	total      float64
	mismatched bool
}

// resolveLayoutByTags — backwards-compatible shim. Existing callers (the fork
// path) keep working unchanged; no scheme constraint is applied for them.
func resolveLayoutByTags(
	ctx context.Context,
	tx *sql.Tx,
	category string,
	industryTags []string,
	logger *zap.Logger,
) (*layoutResolution, error) {
	return resolveLayoutByTagsWeighted(ctx, tx, category, industryTags, "", logger)
}

// resolveLayoutByTagsWeighted — weighted, scheme-aware layout match.
//
//	scheme   near-hard constraint: a light site won't take a dark layout while
//	         any non-dark (same/neutral/unknown) layout fits, and vice-versa.
//	tags     IDF-weighted (rare tag = specific = high weight); synonyms folded.
//	category small bonus on exact match.
//	desc     small bonus per site term found in the layout's description.
//
// Zero fit anywhere -> scheme-aware fallback.
func resolveLayoutByTagsWeighted(
	ctx context.Context,
	tx *sql.Tx,
	category string,
	industryTags []string,
	siteScheme string,
	logger *zap.Logger,
) (*layoutResolution, error) {

	siteTerms := canonicalSet(append([]string{category}, industryTags...))
	if len(siteTerms) == 0 {
		return fallbackLayout(ctx, tx, siteScheme, "no classification tags", nil, logger)
	}
	siteCategory := canonicalTag(category)

	// Fetch every active layout (small table). industry_tags via array_to_json
	// so we don't depend on a specific driver's array scanning.
	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, COALESCE(category,''),
		       COALESCE(array_to_json(industry_tags)::text,'[]'),
		       COALESCE(scheme,''), COALESCE(description,'')
		FROM layouts
		WHERE is_active = true
	`)
	if err != nil {
		return nil, fmt.Errorf("layout fetch failed: %w", err)
	}
	defer rows.Close()

	var layouts []layoutRow
	df := map[string]int{}
	for rows.Next() {
		var lr layoutRow
		var tagsJSON string
		if err := rows.Scan(&lr.id, &lr.name, &lr.category, &tagsJSON, &lr.scheme, &lr.desc); err != nil {
			return nil, fmt.Errorf("layout row scan: %w", err)
		}
		_ = json.Unmarshal([]byte(tagsJSON), &lr.tags)
		layouts = append(layouts, lr)
		for c := range canonicalSet(lr.tags) {
			df[c]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("layout rows iteration: %w", err)
	}
	if len(layouts) == 0 {
		return fallbackLayout(ctx, tx, siteScheme, "no active layouts", nil, logger)
	}
	N := float64(len(layouts))
	weight := func(tag string) float64 {
		d := df[tag]
		if d == 0 {
			d = 1
		}
		return math.Log(1.0 + N/float64(d))
	}

	scored := make([]scoredLayout, 0, len(layouts))
	for _, lr := range layouts {
		layoutTags := canonicalSet(lr.tags)

		var tagScore float64
		for term := range siteTerms {
			if _, ok := layoutTags[term]; ok {
				tagScore += weight(term)
			}
		}

		bonus := 0.0
		if lr.category != "" && (canonicalTag(lr.category) == siteCategory || lmHasTerm(siteTerms, canonicalTag(lr.category))) {
			bonus += lmCategoryMatchBonus
		}

		descLower := " " + strings.ToLower(lr.desc) + " "
		descBonus := 0.0
		for term := range siteTerms {
			if term == "" {
				continue
			}
			if strings.Contains(descLower, " "+term+" ") ||
				strings.Contains(descLower, " "+strings.ReplaceAll(term, "-", " ")+" ") {
				descBonus += lmDescWordBonus
			}
		}
		if descBonus > lmDescBonusCap {
			descBonus = lmDescBonusCap
		}
		bonus += descBonus

		mismatched := false
		if siteScheme != "" && lr.scheme != "" && lr.scheme != "neutral" {
			if lr.scheme == siteScheme {
				bonus += lmSameSchemeBonus
			} else {
				mismatched = true
			}
		}

		scored = append(scored, scoredLayout{row: lr, tagScore: tagScore, total: tagScore + bonus, mismatched: mismatched})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].total != scored[j].total {
			return scored[i].total > scored[j].total
		}
		if scored[i].tagScore != scored[j].tagScore {
			return scored[i].tagScore > scored[j].tagScore
		}
		return scored[i].row.name < scored[j].row.name
	})

	// Candidates = plain layout names (unchanged format for consumers).
	candidates := make([]string, 0, 5)
	for i := 0; i < len(scored) && i < 5; i++ {
		candidates = append(candidates, scored[i].row.name)
	}

	// Prefer the best same-scheme / neutral / unknown layout with positive fit.
	if best := lmFirstEligible(scored); best != nil {
		logger.Info("resolveLayoutByTags: matched (scheme-aware, weighted)",
			zap.String("layout_name", best.row.name),
			zap.String("layout_scheme", best.row.scheme),
			zap.String("site_scheme", siteScheme),
			zap.Float64("score", best.total),
			zap.Float64("tag_score", best.tagScore),
			zap.Strings("candidates", candidates),
		)
		return &layoutResolution{
			LayoutID:   best.row.id,
			LayoutName: best.row.name,
			Scheme:     best.row.scheme,
			Reason: fmt.Sprintf("weighted match: score %.2f (tags %.2f), layout %q [scheme=%s] vs site scheme %q; candidates %s",
				best.total, best.tagScore, best.row.name, lmSchemeOrDash(best.row.scheme), lmSchemeOrDash(siteScheme), strings.Join(candidates, ", ")),
			Candidates: candidates,
			IsFallback: false,
		}, nil
	}

	// Only opposite-scheme layouts fit -> use the best, but FLAG the gap.
	if best := lmFirstWithFit(scored); best != nil {
		logger.Warn("resolveLayoutByTags: only opposite-scheme layouts fit — library gap",
			zap.String("layout_name", best.row.name),
			zap.String("layout_scheme", best.row.scheme),
			zap.String("site_scheme", siteScheme),
			zap.Strings("candidates", candidates),
		)
		return &layoutResolution{
			LayoutID:         best.row.id,
			LayoutName:       best.row.name,
			Scheme:           best.row.scheme,
			IsSchemeMismatch: true,
			Reason: fmt.Sprintf("scheme gap: no %s layout fit these tags; applied %q [scheme=%s]. candidates %s",
				lmSchemeOrDash(siteScheme), best.row.name, lmSchemeOrDash(best.row.scheme), strings.Join(candidates, ", ")),
			Candidates: candidates,
			IsFallback: false,
		}, nil
	}

	return fallbackLayout(ctx, tx, siteScheme,
		fmt.Sprintf("no layout fit site terms (%s)", strings.Join(lmKeys(siteTerms), ",")), candidates, logger)
}

func lmFirstEligible(s []scoredLayout) *scoredLayout {
	for i := range s {
		if !s[i].mismatched && s[i].total > 0 {
			return &s[i]
		}
	}
	return nil
}
func lmFirstWithFit(s []scoredLayout) *scoredLayout {
	for i := range s {
		if s[i].total > 0 {
			return &s[i]
		}
	}
	return nil
}
func lmHasTerm(set map[string]struct{}, t string) bool { _, ok := set[t]; return ok }
func lmKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func lmSchemeOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// fallbackLayout — scheme-aware. brochure-formal (light) is the default; if you
// later seed a dark generic fallback, pick it here for dark sites. The mismatch
// (if any) is recorded so the action can flag the gap.
func fallbackLayout(
	ctx context.Context,
	tx *sql.Tx,
	siteScheme string,
	reason string,
	candidates []string,
	logger *zap.Logger,
) (*layoutResolution, error) {

	fallbackName := "brochure-formal"

	var fb struct {
		id     uuid.UUID
		name   string
		scheme string
	}
	err := tx.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(scheme,'') FROM layouts
		WHERE name = $1 AND is_active = true
		LIMIT 1
	`, fallbackName).Scan(&fb.id, &fb.name, &fb.scheme)
	if err != nil {
		return nil, fmt.Errorf("fallback layout %q not found: %w — was Phase 1 seed loaded?", fallbackName, err)
	}

	logger.Info("fallbackLayout", zap.String("layout", fb.name), zap.String("site_scheme", siteScheme), zap.String("reason", reason))
	return &layoutResolution{
		LayoutID:         fb.id,
		LayoutName:       fb.name,
		Scheme:           fb.scheme,
		IsSchemeMismatch: siteScheme != "" && fb.scheme != "" && fb.scheme != siteScheme,
		Reason:           "fallback — " + reason,
		Candidates:       candidates,
		IsFallback:       true,
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

// resolveTypographySet picks a typography_set row by font_family exact match,
// or inserts a new one if none matches.
//
// Used by:
//   - resolveTypographySetForFork   (fork path — extracts fonts from design_spec)
//   - resolve_composition_typography (site-design-planner — extracts from design
//     specs or falls through to layout default)
//
// The match is deliberately exact-string on font_family. Normalising font
// stacks is fiddly — "'Inter', system-ui" and "Inter, -apple-system, ..."
// likely render the same glyphs but are different stacks and shouldn't be
// silently merged. HITL review can consolidate duplicates later.
//
// If fonts is empty or missing font_family, returns the sans-modern default.
//
// Returns (id, matched, err) where matched=true means an existing row was
// reused; matched=false means a new row was inserted.
func resolveTypographySet(
	ctx context.Context,
	tx *sql.Tx,
	fonts map[string]string,
	baseName string,
	displayName string,
	category string,
	industryTags []string,
	siteID uuid.UUID,
	domain string,
	logger *zap.Logger,
) (uuid.UUID, bool, error) {

	// Empty or missing font_family → fallback to sans-modern.
	fontFamily := ""
	if fonts != nil {
		fontFamily = strings.TrimSpace(fonts["font_family"])
	}

	if len(fonts) == 0 || fontFamily == "" {
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
		logger.Info("resolveTypographySet: no font signal, using sans-modern default",
			zap.String("typography_set_id", defaultID.String()),
		)
		return defaultID, true, nil
	}

	// Match on font_family — the most distinctive field.
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
		logger.Info("resolveTypographySet: matched existing typography_set",
			zap.String("typography_set_id", matchID.String()),
			zap.String("font_family", fontFamily),
		)
		return matchID, true, nil
	}
	if err != sql.ErrNoRows {
		return uuid.Nil, false, fmt.Errorf("typography match query failed: %w", err)
	}

	// No match — create a new typography_set row.
	fontsJSON, err := json.Marshal(fonts)
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
		fmt.Sprintf("Typography extracted from %s", domain),
		string(fontsJSON),
		category,
		industryTagsArr,
		siteID, domain,
	).Scan(&newID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("insert typography_set: %w", err)
	}

	logger.Info("resolveTypographySet: inserted new typography_set",
		zap.String("typography_set_id", newID.String()),
		zap.String("typography_set_name", typoName),
		zap.String("font_family", fontFamily),
	)
	return newID, false, nil
}

// createPalette inserts a new palettes row with the caller-supplied
// colours map and metadata. Used by:
//   - createPaletteForFork              (fork path — from design_spec)
//   - resolve_composition_palette_action (composition — from spec cascade)
//
// Palettes are always site-specific — library reuse of palettes across
// sites is almost never correct. Callers should always invoke this to
// create a fresh row, not try to match against existing palettes.
//
// `origin` should be 'adopted' for every caller today. `needsReview`
// lets callers decide whether to gate the palette behind HITL: true
// for fork-to-library, false for direct site composition (the site is
// using its own palette, no library promotion implied).
func createPalette(
	ctx context.Context,
	tx *sql.Tx,
	coloursMap map[string]string,
	baseName string,
	displayName string,
	category string,
	industryTags []string,
	siteID uuid.UUID,
	domain string,
	parentPaletteID *uuid.UUID,
	origin string,
	needsReview bool,
	logger *zap.Logger,
) (uuid.UUID, error) {

	if len(coloursMap) == 0 {
		return uuid.Nil, fmt.Errorf(
			"cannot create palette: no colour entries supplied",
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
			$7, $8,
			$9, $10, $11, NOW()
		) RETURNING id
	`,
		paletteName,
		displayName,
		fmt.Sprintf("Palette for %s", domain),
		string(coloursJSON),
		category,
		industryTagsArr,
		origin, needsReview,
		parentPaletteID, siteID, domain,
	).Scan(&newID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert palette: %w", err)
	}

	logger.Info("createPalette: inserted palette",
		zap.String("palette_id", newID.String()),
		zap.String("palette_name", paletteName),
		zap.Int("colour_count", len(coloursMap)),
		zap.String("origin", origin),
		zap.Bool("needs_review", needsReview),
	)

	return newID, nil
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
