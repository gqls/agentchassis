// FILE: platform/orchestration/actions/resolve_composition_layout_action.go
//
// ResolveCompositionLayoutAction is site-design-planner's layout picker. It
// runs inside the site-design-planner workflow after ValidateCompositionInputs
// has confirmed identity + classification specs exist.
//
// Separation of concerns:
//   - The matching LOGIC lives in `resolveLayoutByTags`
//     (fork_theme_composition.go) — weighted, scheme-aware (tag IDF + scheme
//     constraint + category/description). It gained a `siteScheme` parameter in
//     this change; this action is its single caller.
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
// Classification/identity are read via the shared `readClassificationFromContext`
// cascade (resolve_composition_helpers.go) — the same one install_site_composition,
// resolve_composition_typography and resolve_composition_palette already use.
// Fixed 2026-09-02 (bugs_open/113's tail): this action used to reimplement its
// own narrower extraction (`classData["category"]`/`classData["industry_tags"]`
// only, no identity fallback), the one caller of the four that lacked it, and a
// site whose classifier output has no `category`/`industry_tags` — several do —
// resolved with zero signal even when identity data was present and unused.
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

	// Extract category + industry_tags via the same shared cascade the other
	// three composition resolvers use (install_site_composition, typography,
	// palette): prefer the prior step's output (validated_inputs.classification)
	// or a direct spec read, falling back to identity.industry/sub_industry +
	// site_type when the classifier's current output shape carries neither
	// `category` nor `industry_tags` (bugs_open/113's tail: ai-agent-orchestration.com's
	// classification spec has always lacked both — a legacy shape predating the
	// current classifier — so this resolver alone, of the four, returned zero
	// signal and fell through to brochure-formal, though identity.industry =
	// "Technology Services" was sitting right there unused).
	category, industryTags := readClassificationFromContext(ctx, params, siteID, logger)

	// Derive the light/dark scheme from the design brief so the matcher won't
	// place a light site on a dark layout (or vice-versa) on tag overlap alone.
	siteScheme := deriveSiteScheme(ctx, params, siteID, logger)

	// Theme-kit rung: unlike palette/typography (which the kit steers via
	// design_intent.reference_values — see apply_theme_kit_action.go's doc
	// comment), layout resolution never consults design_intent at all, so a
	// kit's layout choice needs an explicit short-circuit here. Human-set
	// signals (mission/design_intent driving the tag matcher below) are not
	// consulted for layout today, so this rung has nothing above it to defer
	// to — a themed site's layout is exactly what the kit named, no scoring.
	if kit, ok, kerr := loadSiteThemeKitDefaults(ctx, params.DB, siteID); kerr == nil && ok && kit.LayoutID.Valid {
		var kitLayoutName, kitLayoutScheme string
		var kitLayoutActive bool
		lerr := params.DB.QueryRowContext(ctx,
			`SELECT name, COALESCE(scheme, ''), is_active FROM layouts WHERE id = $1`,
			kit.LayoutID.UUID,
		).Scan(&kitLayoutName, &kitLayoutScheme, &kitLayoutActive)
		if lerr == nil && kitLayoutActive {
			logger.Info("ResolveCompositionLayoutAction: theme-kit default",
				zap.String("site_id", siteID.String()),
				zap.String("theme_kit", kit.ThemeKitName),
				zap.String("layout_id", kit.LayoutID.UUID.String()),
				zap.String("layout_name", kitLayoutName),
			)
			// Report the scheme comparison honestly rather than hardcoding
			// false: siteScheme is computed above and a kit CAN name a layout
			// whose scheme contradicts the site's brief (the kit is a human's
			// general choice; the brief is about this site). Honouring the kit
			// is deliberate — but claiming the two agreed when they did not
			// would put a false value in the same field the tag-matching path
			// uses truthfully, and downstream reads it as "no mismatch here".
			kitSchemeMismatch := siteScheme != "" && kitLayoutScheme != "" &&
				siteScheme != kitLayoutScheme
			if kitSchemeMismatch {
				logger.Warn("ResolveCompositionLayoutAction: theme-kit layout contradicts the site's derived scheme — honouring the kit",
					zap.String("site_id", siteID.String()),
					zap.String("theme_kit", kit.ThemeKitName),
					zap.String("site_scheme", siteScheme),
					zap.String("kit_layout_scheme", kitLayoutScheme),
				)
			}
			return map[string]interface{}{
				"layout_id":          kit.LayoutID.UUID.String(),
				"layout_name":        kitLayoutName,
				"reason":             fmt.Sprintf("theme_kit default: %s", kit.ThemeKitName),
				"candidates":         []string{kitLayoutName},
				"is_fallback":        false,
				"scheme":             kitLayoutScheme,
				"is_scheme_mismatch": kitSchemeMismatch,
				"site_tags":          collectNormalisedSiteTags(category, industryTags),
				"review_item_queued": nil,
				"source":             "theme_kit_default",
			}, nil
		}
		if lerr != nil {
			logger.Warn("ResolveCompositionLayoutAction: theme-kit layout lookup failed, falling back to tag match",
				zap.Error(lerr), zap.String("theme_kit", kit.ThemeKitName))
		} else {
			logger.Warn("ResolveCompositionLayoutAction: theme-kit's layout is no longer active, falling back to tag match",
				zap.String("theme_kit", kit.ThemeKitName), zap.String("layout_id", kit.LayoutID.UUID.String()))
		}
	}

	// Open a short transaction. The shared resolver takes *sql.Tx because it
	// is used by the fully-transactional fork path. Here we use it read-only
	// and commit without writes.
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	resolution, err := resolveLayoutByTags(ctx, tx, category, industryTags, siteScheme, logger)
	if err != nil {
		return nil, fmt.Errorf("resolveLayoutByTags failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit read-only tx: %w", err)
	}

	siteTagsForOutput := collectNormalisedSiteTags(category, industryTags)

	// Emit a library-growth work item when the library had no good answer:
	// a hard fallback, a scheme gap (an opposite-scheme layout was applied),
	// or — added by bugs_open/445 — a WEAK TAG FIT: a layout that scored above
	// zero on bonuses while matching little or none of the site's own tags.
	//
	// The weak arm exists because the old two-arm test could not fire on the
	// case that actually happens. `IsFallback` requires the TOTAL score to be
	// zero across the whole library, and the category/description/scheme
	// bonuses are added to that total independently of any tag matching — so a
	// layout matching NONE of a site's tags still scores above zero and the
	// library was recorded as having answered. Measured 2026-09-03: four live
	// sites recorded `tags 0.00` with lineage `library_match`, and exactly two
	// needs_new_layout_candidate items exist across 63,007 work items ever
	// written (29,657 live ∪ 33,350 archived), BOTH from the degenerate
	// no-tags-at-all arm. The mechanism had never once assessed the library and
	// reported it short.
	var reviewItemQueued interface{}
	if resolution.LibraryGap() {
		if resolution.IsWeakFit && !resolution.IsFallback && !resolution.IsSchemeMismatch {
			logger.Warn("ResolveCompositionLayoutAction: weak tag fit — the applied layout addresses little of this site's classification",
				zap.String("site_id", siteID.String()),
				zap.String("domain", domain),
				zap.String("applied_layout", resolution.LayoutName),
				zap.Float64("tag_coverage", resolution.Fit.TagCoverage),
				zap.Float64("threshold", resolution.Fit.Threshold),
				zap.Strings("matched_terms", resolution.Fit.MatchedTerms),
				zap.Strings("unmatched_terms", resolution.Fit.UnmatchedTerms),
				zap.String("runner_up", resolution.Fit.RunnerUp),
				zap.Strings("site_tags", siteTagsForOutput),
				zap.String("recommendation", "extend a layout's industry_tags, correct the classification, or add a layout for this shape"),
			)
		} else if resolution.IsSchemeMismatch && !resolution.IsFallback {
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
			resolution.GapReason(), resolution.Fit,
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

	// `source` is now reported on EVERY branch, not just the theme-kit one.
	// Until this change the tag-match and fallback paths emitted no `source`
	// key at all, so install_site_composition had to INFER layout_source from
	// is_fallback — an inference that cannot represent a scheme gap or a weak
	// fit, and silently recorded both as a clean `library_match`.
	source := "library_match"
	if resolution.IsFallback {
		source = "library_fallback"
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
		"source":             source,

		// Structured fit evidence — migration 103's `layout_match_score` and the
		// context a reviewer needs to act on it without re-running the matcher.
		"library_gap":     resolution.LibraryGap(),
		"gap_reason":      resolution.GapReason(),
		"tag_coverage":    resolution.Fit.TagCoverage,
		"tag_score":       resolution.Fit.TagScore,
		"score":           resolution.Fit.Score,
		"matched_terms":   resolution.Fit.MatchedTerms,
		"unmatched_terms": resolution.Fit.UnmatchedTerms,
		"runner_up":       resolution.Fit.RunnerUp,
		"runner_up_score": resolution.Fit.RunnerUpScore,
		"margin":          resolution.Fit.Margin,
		"fit_threshold":   resolution.Fit.Threshold,
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
	gapReason string,
	fit layoutFit,
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

		// Which arm fired, and the evidence to act on it. `unmatched_terms` is
		// the operative field for a reviewer: it names the vocabulary the
		// library could not serve, which is what distinguishes "we need a new
		// layout for this shape" from "an existing layout needs these tags".
		"gap_reason":      gapReason,
		"tag_coverage":    fit.TagCoverage,
		"coverage_pct":    fmt.Sprintf("%.0f%%", fit.TagCoverage*100),
		"threshold":       fit.Threshold,
		"matched_terms":   fit.MatchedTerms,
		"unmatched_terms": fit.UnmatchedTerms,
		"runner_up":       fit.RunnerUp,
		"margin":          fit.Margin,
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	scheme := siteScheme
	if scheme == "" {
		scheme = "unknown"
	}

	// Parked for a human: status "needs_human_review" + an EMPTY handler — the
	// canonical HITL idiom (migration 217; fleet census 544 rows '' vs 22 at any
	// pseudo-handler name, refresh_evidence_fact_drift.go). The dispatch loop
	// never selects this status, and with no handler named, CHECK
	// swi_no_handlerless_promotable (migration 443) structurally refuses any
	// later promotion into the dispatch queue. A human resolves via admin API
	// (PATCH /work-items/:id): add a layout and retry, or mark wont_fix.
	//
	// This used to say `handlerAgent: "hitl-review"` "matches the existing
	// tool-auditor → HITL pattern" — bugs_open/291: that agent has NEVER existed
	// (an April 2026 convention whose handler was never built), and tool-auditor
	// was not even setting this status, so the "existing pattern" was the bug.
	item := workItem{
		siteID:   siteID,
		source:   "side_effect",
		pipeline: "build",
		itemType: "needs_new_layout_candidate",
		severity: "low",
		// The coverage percentage is in the SUMMARY, not only the spec, so a
		// reviewer can triage the queue by number without opening each row —
		// and so a cluster of sites leaning on the same thin match is visible
		// as a pattern in a list view rather than only in the JSON.
		summary: fmt.Sprintf(
			"Layout gap for %s (%s, scheme=%s) — applied %s at %.0f%% tag coverage (%d/%d terms); extend a layout's tags, fix the classification, or add a layout",
			domain, gapReason, scheme, appliedLayout,
			fit.TagCoverage*100, len(fit.MatchedTerms), fit.SiteTermCount),
		spec:         string(specJSON),
		priority:     80,
		handlerAgent: "",
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
