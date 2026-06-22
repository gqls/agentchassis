// MERGE TARGET: platform/orchestration/actions/fork_theme_composition.go
// =====================================================================
// HOW TO MERGE (three files; all package `actions`)
//
// A) fork_theme_composition.go
//    1. Imports: add "math" and "sort" to the existing import block
//       (it already has context, database/sql, encoding/json, fmt, strings,
//        github.com/google/uuid, go.uber.org/zap — nothing else new needed).
//    2. layoutResolution struct: add two fields —
//         Scheme           string // chosen layout's scheme ("" if unknown)
//         IsSchemeMismatch bool   // true if we had to cross schemes (gap signal)
//    3. DELETE the old resolveLayoutByTags(...) and the old fallbackLayout(...)
//       and paste everything BELOW this header in their place (or drop this
//       file in as a new file in the same package once the old two funcs are
//       removed). The old call signature is preserved as a shim, so the fork
//       caller (fork_theme_from_site) compiles and behaves exactly as before
//       (it passes no scheme -> "" -> no scheme constraint).
//
// B) resolve_composition_layout_action.go — change the one caller:
//    1. Replace the call:
//         resolution, err := resolveLayoutByTags(ctx, tx, category, industryTags, logger)
//       with, just above it deriving the scheme, then:
//         siteScheme := deriveSiteScheme(ctx, params, siteID, logger)
//         resolution, err := resolveLayoutByTagsWeighted(ctx, tx, category, industryTags, siteScheme, logger)
//       (update the error string to resolveLayoutByTagsWeighted).
//    2. Queue the library-growth item on a scheme GAP as well as a hard
//       fallback. Change `if resolution.IsFallback {` to
//         `if resolution.IsFallback || resolution.IsSchemeMismatch {`
//       and pass the accurate reason/applied layout/scheme into
//       queueLayoutCandidateReview (see C).
//    3. Add to the returned map:
//         "scheme":             resolution.Scheme,
//         "is_scheme_mismatch": resolution.IsSchemeMismatch,
//    4. Add the deriveSiteScheme helper (bottom of B):
//
//         func deriveSiteScheme(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) string {
//             var styleDirection, suggestedStyle string
//             if di, found, err := loadCurrentSpecData(ctx, params.DB, siteID, "design_intent"); err == nil && found {
//                 styleDirection, _ = di["style_direction"].(string)
//             }
//             if cl, found, err := loadCurrentSpecData(ctx, params.DB, siteID, "classification"); err == nil && found {
//                 suggestedStyle, _ = cl["suggested_style"].(string)
//             }
//             s := deriveSchemeFromDesignIntent(styleDirection, suggestedStyle)
//             logger.Info("deriveSiteScheme",
//                 zap.String("style_direction", styleDirection),
//                 zap.String("suggested_style", suggestedStyle),
//                 zap.String("scheme", s))
//             return s
//         }
//
// C) resolve_composition_layout_action.go — widen queueLayoutCandidateReview
//    so the work item is honest for the scheme-gap case (a cross-scheme layout
//    was applied, not brochure-formal). New signature + spec:
//
//      func queueLayoutCandidateReview(ctx context.Context, db *sql.DB,
//          siteID uuid.UUID, domain, siteScheme, reason, appliedLayout string,
//          siteTags, candidatesConsidered []string, logger *zap.Logger) (*uuid.UUID, error) {
//          ...
//          spec := map[string]interface{}{
//              "domain":                domain,
//              "reason":                reason,
//              "site_scheme":           siteScheme,
//              "queued_by":             "site-design-planner:resolve_composition_layout",
//              "site_tags":             siteTags,
//              "candidates_considered": candidatesConsidered,
//              "applied_layout":        appliedLayout,
//          }
//          ... (rest unchanged) ...
//      }
//    and the call becomes:
//      queueLayoutCandidateReview(ctx, params.DB, siteID, domain, siteScheme,
//          resolution.Reason, resolution.LayoutName, siteTagsForOutput, resolution.Candidates, logger)
//
// D) Requires the layouts.scheme column —
//    migration_layouts_scheme_and_light_tool_portal.sql.
//
// Everything below is the replacement logic. Reads no driver-specific array
// type (scans industry_tags via array_to_json), so it is portable.
// =====================================================================

package actions

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
//   scheme   near-hard constraint: a light site won't take a dark layout while
//            any non-dark (same/neutral/unknown) layout fits, and vice-versa.
//   tags     IDF-weighted (rare tag = specific = high weight); synonyms folded.
//   category small bonus on exact match.
//   desc     small bonus per site term found in the layout's description.
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
