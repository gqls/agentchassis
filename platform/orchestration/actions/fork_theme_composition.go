package actions

// FILE: platform/orchestration/actions/fork_theme_composition.go
//
// Shared composition helpers used by the site-design-planner actions:
//   - resolveLayoutByTags  — weighted, scheme-aware layout match (resolve_composition_layout)
//   - resolveTypographySet — match-or-insert a typography_sets row (resolve_composition_typography)
//   - createPalette        — insert a site-specific palettes row (resolve_composition_palette)
//   - resolveUniqueNameInTx and small utilities
//
// Each runs inside the caller's transaction. Palettes are always new (site
// colours are site-specific); typography_sets are matched-or-new (fonts are
// shared more often); layouts are matched from the library or fall back to
// brochure-formal.
//
// NOTE: ForkThemeFromSiteAction does NOT use these — it builds css_themes +
// style_collections with the palette/typography stored inline as JSON on the
// css_themes row and templates the CSS from the captured design_spec.

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

	Fit       layoutFit // structured evidence for how well the chosen layout actually fits
	IsWeakFit bool      // Fit.TagCoverage < lmMinTagCoverage (a library-gap signal)
}

// layoutFit is the structured record of HOW WELL the chosen layout fits, as
// opposed to merely which one won. Migration 103 specified this in April 2026
// (`lineage.layout_match_score`, "(float 0-1) — tag-overlap score for chosen
// layout") and it was never computed: measured 2026-09-03, 0 of 33
// resolved_composition rows carry the key, and the only surviving trace of a
// score is a prose sentence in `reasoning` that has to be parsed with a regex.
//
// TagCoverage is the normalised score 103 asked for. Note the denominator uses
// the SAME IDF weights as the numerator, including for terms no layout carries:
// weight() gives an unmatched term df=0 → d=1 → the maximum weight. That is
// deliberate. A term the whole library cannot serve is exactly the thing this
// measure must count against the fit, not quietly drop.
type layoutFit struct {
	TagCoverage    float64  // matched weight / total site-term weight, in [0,1]
	SiteTermCount  int      //
	MatchedTerms   []string // sorted canonical site terms the chosen layout carries
	UnmatchedTerms []string // sorted canonical site terms it does not
	Score          float64  // total incl. bonuses — the figure `reasoning` prints
	TagScore       float64  // the tag-overlap half alone
	RunnerUp       string   // next eligible layout, "" if none
	RunnerUpScore  float64  //
	Margin         float64  // Score - RunnerUpScore
	Threshold      float64  // lmMinTagCoverage in force when this was recorded
}

// LibraryGap — the single definition of "the library did not have a good answer
// for this site", shared by the action and its tests so the two cannot drift.
//
// The weak-fit arm is the one added by bugs_open/445. The other two are the
// original behaviour and are unchanged.
func (r *layoutResolution) LibraryGap() bool {
	return r.IsFallback || r.IsSchemeMismatch || r.IsWeakFit
}

// GapReason names which arm fired, most-severe first, for the work item and the
// lineage record. Empty string when there is no gap.
func (r *layoutResolution) GapReason() string {
	switch {
	case r.IsFallback:
		return "fallback"
	case r.IsSchemeMismatch:
		return "scheme_mismatch"
	case r.IsWeakFit:
		return "weak_tag_fit"
	default:
		return ""
	}
}

// Weighted, scheme-aware layout matching. The tunables, synonym map, and
// scoring live here; resolveLayoutByTags (below) is the single entry point.

// --- tunables (one place to adjust scoring) ---
const (
	lmCategoryMatchBonus = 0.75 // layout.category matches a site term
	lmDescWordBonus      = 0.15 // per distinct site term found in the description
	lmDescBonusCap       = 0.90 // max description contribution
	lmSameSchemeBonus    = 0.50 // nudge toward an exactly-matching scheme

	// lmMinTagCoverage — below this fraction of the site's own weighted identity,
	// a positively-scored match is STILL a library gap.
	//
	// Why a coverage floor exists at all (bugs_open/445). The three bonuses above
	// are added to `total` INDEPENDENTLY of any tag matching, so `total > 0` — the
	// old sole test for "the library had something" — is satisfiable with
	// tagScore == 0. Measured 2026-09-03: four live sites are recorded by this
	// very code as `tags 0.00` AND lineage layout_source `library_match`
	// (webdesign.uk→brochure-bold 0.75; farmerinsurance/garden-tools/vetcomparison
	// →industry-hub 0.90). A layout matching NONE of a site's tags, recorded as a
	// successful library match, with no needs_new_layout_candidate raised.
	//
	// Why 0.50 and not a rounder-sounding smaller number. The measured coverage
	// distribution over the 33 composed sites has exactly two empty intervals:
	// (10%, 15%) — 5 points wide — and (38%, 62%) — 24 points wide. One tag on a
	// ten-term site moves coverage by roughly 8-10 points, so a cut in the narrow
	// band would flip sites on ordinary classifier variance; only the wide band is
	// stable against it. 0.50 is also the threshold migration 103's own worked
	// example names ("scored utility-tool=0.82 above threshold 0.5"), and reads
	// plainly: more than half the site's declared identity, weighted by
	// specificity, is unaddressed by the layout it was given.
	//
	// This does NOT change which layout is selected — only whether the gap is
	// recorded. See layoutFit.
	lmMinTagCoverage = 0.50
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

// resolveLayoutByTags — weighted, scheme-aware layout match.
//
//	scheme   near-hard constraint: a light site won't take a dark layout while
//	         any non-dark (same/neutral/unknown) layout fits, and vice-versa.
//	tags     IDF-weighted (rare tag = specific = high weight); synonyms folded.
//	category small bonus on exact match.
//	desc     small bonus per site term found in the layout's description.
//
// Zero fit anywhere -> scheme-aware fallback.
func resolveLayoutByTags(
	ctx context.Context,
	tx *sql.Tx,
	category string,
	industryTags []string,
	siteScheme string,
	logger *zap.Logger,
) (*layoutResolution, error) {

	siteTerms := canonicalSet(append([]string{category}, industryTags...))
	if len(siteTerms) == 0 {
		return fallbackLayout(ctx, tx, siteScheme, "no classification tags", nil, siteTerms, logger)
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
		return fallbackLayout(ctx, tx, siteScheme, "no active layouts", nil, siteTerms, logger)
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
		fit := lmBuildFit(best, scored, siteTerms, weight)
		weak := fit.TagCoverage < lmMinTagCoverage
		logger.Info("resolveLayoutByTags: matched (scheme-aware, weighted)",
			zap.String("layout_name", best.row.name),
			zap.String("layout_scheme", best.row.scheme),
			zap.String("site_scheme", siteScheme),
			zap.Float64("score", best.total),
			zap.Float64("tag_score", best.tagScore),
			zap.Float64("tag_coverage", fit.TagCoverage),
			zap.Strings("matched_terms", fit.MatchedTerms),
			zap.Strings("unmatched_terms", fit.UnmatchedTerms),
			zap.Bool("weak_fit", weak),
			zap.Strings("candidates", candidates),
		)
		return &layoutResolution{
			LayoutID:   best.row.id,
			LayoutName: best.row.name,
			Scheme:     best.row.scheme,
			Reason: fmt.Sprintf("weighted match: score %.2f (tags %.2f), layout %q [scheme=%s] vs site scheme %q; candidates %s",
				best.total, best.tagScore, best.row.name, lmSchemeOrDash(best.row.scheme), lmSchemeOrDash(siteScheme), strings.Join(candidates, ", ")) +
				lmFitSummary(fit),
			Candidates: candidates,
			IsFallback: false,
			Fit:        fit,
			IsWeakFit:  weak,
		}, nil
	}

	// Only opposite-scheme layouts fit -> use the best, but FLAG the gap.
	if best := lmFirstWithFit(scored); best != nil {
		fit := lmBuildFit(best, scored, siteTerms, weight)
		logger.Warn("resolveLayoutByTags: only opposite-scheme layouts fit — library gap",
			zap.String("layout_name", best.row.name),
			zap.String("layout_scheme", best.row.scheme),
			zap.String("site_scheme", siteScheme),
			zap.Float64("tag_coverage", fit.TagCoverage),
			zap.Strings("candidates", candidates),
		)
		return &layoutResolution{
			LayoutID:         best.row.id,
			LayoutName:       best.row.name,
			Scheme:           best.row.scheme,
			IsSchemeMismatch: true,
			Reason: fmt.Sprintf("scheme gap: no %s layout fit these tags; applied %q [scheme=%s]. candidates %s",
				lmSchemeOrDash(siteScheme), best.row.name, lmSchemeOrDash(best.row.scheme), strings.Join(candidates, ", ")) +
				lmFitSummary(fit),
			Candidates: candidates,
			IsFallback: false,
			Fit:        fit,
			IsWeakFit:  fit.TagCoverage < lmMinTagCoverage,
		}, nil
	}

	return fallbackLayout(ctx, tx, siteScheme,
		fmt.Sprintf("no layout fit site terms (%s)", strings.Join(lmKeys(siteTerms), ",")), candidates, siteTerms, logger)
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

// lmBuildFit assembles the structured fit evidence for the layout that won.
//
// `weight` is the caller's IDF closure — passed in rather than recomputed so the
// coverage denominator is guaranteed to use the same weights as the tagScore
// numerator the matcher already produced. Computing them twice is how the two
// halves of a ratio silently stop being comparable.
func lmBuildFit(
	best *scoredLayout,
	scored []scoredLayout,
	siteTerms map[string]struct{},
	weight func(string) float64,
) layoutFit {

	fit := layoutFit{
		SiteTermCount: len(siteTerms),
		Score:         best.total,
		TagScore:      best.tagScore,
		Threshold:     lmMinTagCoverage,
	}

	winnerTags := canonicalSet(best.row.tags)
	var totalWeight float64
	for term := range siteTerms {
		totalWeight += weight(term)
		if _, ok := winnerTags[term]; ok {
			fit.MatchedTerms = append(fit.MatchedTerms, term)
		} else {
			fit.UnmatchedTerms = append(fit.UnmatchedTerms, term)
		}
	}
	sort.Strings(fit.MatchedTerms)
	sort.Strings(fit.UnmatchedTerms)

	if totalWeight > 0 {
		fit.TagCoverage = best.tagScore / totalWeight
	}

	// Runner-up: the next layout that would itself have been eligible. Reported
	// so a reviewer can see whether the win was decisive or a coin-toss.
	for i := range scored {
		if &scored[i] == best {
			continue
		}
		if !scored[i].mismatched && scored[i].total > 0 {
			if scored[i].total > fit.RunnerUpScore || fit.RunnerUp == "" {
				if scored[i].row.name != best.row.name {
					fit.RunnerUp = scored[i].row.name
					fit.RunnerUpScore = scored[i].total
					break
				}
			}
		}
	}
	fit.Margin = fit.Score - fit.RunnerUpScore
	return fit
}

// lmFitSummary renders the fit as a clause appended to the existing Reason
// string. The prefix of Reason is deliberately left byte-identical, because
// runbooks and at least one other lane read the score out of that prose today.
func lmFitSummary(f layoutFit) string {
	matched := strings.Join(f.MatchedTerms, ",")
	if matched == "" {
		matched = "none"
	}
	return fmt.Sprintf("; tag coverage %.0f%% (%d/%d terms matched: %s)",
		f.TagCoverage*100, len(f.MatchedTerms), f.SiteTermCount, matched)
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
	siteTerms map[string]struct{},
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

	// A fallback matched nothing by construction, so coverage is 0 and every site
	// term is unmatched. Recorded explicitly rather than left as a zero value: an
	// absent fit and a measured-zero fit read identically downstream otherwise,
	// and the unmatched list is what tells a reviewer WHICH vocabulary the
	// library could not serve.
	fit := layoutFit{
		SiteTermCount:  len(siteTerms),
		UnmatchedTerms: lmKeys(siteTerms),
		Threshold:      lmMinTagCoverage,
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
		Fit:              fit,
		IsWeakFit:        len(siteTerms) > 0, // coverage 0 < threshold whenever there were terms to match
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
