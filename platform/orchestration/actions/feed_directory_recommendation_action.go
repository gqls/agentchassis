// FILE: platform/orchestration/actions/feed_directory_recommendation_action.go
//
// EvaluateDirectoryFeaturesAction — post-classification enrichment: decide
// whether a site's vertical should carry a provider DIRECTORY page (Phase B,
// 2026-08-13), and write the recommendation into the classification spec.
//
// A deliberate, close mirror of EvaluateNewsFeedAction
// (feed_news_recommendation_action.go), which is the proven precedent for
// "a deterministic, no-LLM enrichment writes a content_features flag the
// planner then reads": same fuzzy signal matching, same deep-merge, same
// supersede-then-insert spec write, same no-match-means-no-write rule.
// Kept as a separate action rather than a second map inside the news one
// because the two flags are consumed by different machinery (news: the feed
// pipeline; directory: the directoryCheckProfiles/render_directory family)
// and enabling one on an agent must never implicitly enable the other.
//
// WHY THE SPEC KEY IS PER KIND (content_features.<SpecKey>, e.g.
// mortgage_lender_directory) rather than one shared content_features.directory:
// the directory registry is global and each kind renders the SAME list on
// every opted-in site (render_directory_action.go's header), so which KIND a
// site carries is exactly the per-site decision this flag records — two flags
// for two kinds on one site must be able to coexist.
//
// The verticalDirectoryMap below covers ONLY Phase B's starter kinds
// (mortgage-lender, savings-provider, health-insurer — owner ruling
// 2026-08-13). Directories carry NON-PRICE facts only (owner ruling:
// regulator status, product types, underwriter, established year — never
// APR/rates/premiums); that policy lives in the researcher prompt and the
// concept-register entry, not here, but it is why the Reasons below never
// mention rates.
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

var EvaluateDirectoryFeaturesInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("evaluate_directory_features", EvaluateDirectoryFeaturesInputSpec)
}

// verticalDirectoryConfig holds the directory recommendation for an industry
// vertical. Kind names a directoryCheckProfiles/directoryPublishProfiles
// entry; SpecKey is that profile's content_features key. The two travel
// together by construction — see check_directory.go's profile table.
type verticalDirectoryConfig struct {
	Recommended  bool   `json:"recommended"`
	Reason       string `json:"reason"`
	Kind         string `json:"kind"`
	SpecKey      string `json:"spec_key"`
	SeparatePage bool   `json:"separate_page"`
}

// verticalDirectoryMap maps industry/site-type signals to a directory
// recommendation. Signal vocabulary deliberately overlaps verticalNewsMap's
// (classification writes the same industry/site_type/category strings both
// actions read).
var verticalDirectoryMap = map[string]verticalDirectoryConfig{
	"mortgage": {
		Recommended:  true,
		Reason:       "Mortgage sites gain authority from a cited, verified directory of UK lenders",
		Kind:         "mortgage-lender",
		SpecKey:      "mortgage_lender_directory",
		SeparatePage: true,
	},
	"savings": {
		Recommended:  true,
		Reason:       "Savings sites gain authority from a cited, verified directory of UK savings providers",
		Kind:         "savings-provider",
		SpecKey:      "savings_provider_directory",
		SeparatePage: true,
	},
	"banking": {
		Recommended:  true,
		Reason:       "Banking sites gain authority from a cited, verified directory of UK savings providers",
		Kind:         "savings-provider",
		SpecKey:      "savings_provider_directory",
		SeparatePage: true,
	},
	"health insurance": {
		Recommended:  true,
		Reason:       "Health-insurance sites gain authority from a cited, verified directory of UK health insurers",
		Kind:         "health-insurer",
		SpecKey:      "health_insurer_directory",
		SeparatePage: true,
	},
	"insurance": {
		Recommended:  true,
		Reason:       "Insurance sites gain authority from a cited, verified directory of UK health insurers (the one insurer kind built so far; more kinds follow)",
		Kind:         "health-insurer",
		SpecKey:      "health_insurer_directory",
		SeparatePage: true,
	},
	// "finance" alone is deliberately NOT recommended: it is too generic to
	// pick a single provider class, and a wrong directory on a site is worse
	// than none. A site classified merely "finance" gets no directory until a
	// sharper vertical signal lands (the news action recommends on "finance"
	// because news has no per-kind choice to get wrong — this one does).
	"finance": {
		Recommended: false,
		Reason:      "'finance' alone is too generic to choose a provider class; no directory until a sharper vertical signal (mortgage/savings/insurance) is present",
	},
}

// EvaluateDirectoryFeaturesAction checks whether the site's vertical should
// carry a provider directory and writes the recommendation to the
// classification spec via deep merge. Deterministic, zero LLM calls.
func EvaluateDirectoryFeaturesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "evaluate_directory_features"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		EvaluateDirectoryFeaturesInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	var specDataJSON []byte
	err = params.DB.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, siteID).Scan(&specDataJSON)
	if err == sql.ErrNoRows {
		logger.Info("EvaluateDirectoryFeaturesAction: no classification spec yet, skipping")
		return map[string]interface{}{
			"recommended": false,
			"reason":      "no classification spec available",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query classification spec: %w", err)
	}

	var specData map[string]interface{}
	if err := json.Unmarshal(specDataJSON, &specData); err != nil {
		return nil, fmt.Errorf("unmarshal spec: %w", err)
	}

	industry, _ := specData["industry"].(string)
	siteType, _ := specData["site_type"].(string)
	category, _ := specData["category"].(string)

	var domain string
	_ = params.DB.QueryRowContext(ctx, `
		SELECT domain FROM sites WHERE id = $1
	`, siteID).Scan(&domain)

	config := matchVerticalDirectory(industry, siteType, category, domain, logger)

	if config == nil || !config.Recommended {
		// No match, or an explicit not-recommended entry (the "finance" case):
		// either way, NO WRITE — same rule as the news action. A spec that
		// never mentions the key is indistinguishable from one that was never
		// evaluated, which is the safe default for an opt-in flag.
		reason := "no matching vertical profile for a provider directory"
		if config != nil {
			reason = config.Reason
		}
		logger.Info("EvaluateDirectoryFeaturesAction: no directory recommended",
			zap.String("industry", industry),
			zap.String("site_type", siteType),
			zap.String("reason", reason))
		return map[string]interface{}{
			"recommended": false,
			"reason":      reason,
			"industry":    industry,
			"site_type":   siteType,
		}, nil
	}

	directorySpec := map[string]interface{}{
		"content_features": map[string]interface{}{
			config.SpecKey: map[string]interface{}{
				"recommended":   config.Recommended,
				"reason":        config.Reason,
				"kind":          config.Kind,
				"separate_page": config.SeparatePage,
			},
		},
	}

	// deepMergeNewsFeed is a generic recursive map merge despite its name —
	// reused rather than duplicated (it has no news-specific behaviour).
	mergedSpec := deepMergeNewsFeed(specData, directorySpec)
	mergedJSON, err := json.Marshal(mergedSpec)
	if err != nil {
		return nil, fmt.Errorf("marshal merged spec: %w", err)
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE site_specs
		SET is_current = false, superseded_at = NOW()
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("supersede old spec: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by)
		VALUES ($1, 'classification', $2::jsonb, 'enrichment', 'evaluate_directory_features', true, 'evaluate_directory_features')
	`, siteID, string(mergedJSON))
	if err != nil {
		return nil, fmt.Errorf("insert enriched spec: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit enriched spec: %w", err)
	}

	logger.Info("EvaluateDirectoryFeaturesAction: recommendation written",
		zap.String("site_id", siteID.String()),
		zap.String("kind", config.Kind),
		zap.String("spec_key", config.SpecKey))

	return map[string]interface{}{
		"recommended":   true,
		"reason":        config.Reason,
		"kind":          config.Kind,
		"spec_key":      config.SpecKey,
		"separate_page": config.SeparatePage,
	}, nil
}

// matchVerticalDirectory mirrors matchVerticalNews's matching order exactly:
// exact signal match first, then partial (substring) match, plus
// domain-derived signals. An explicit not-recommended entry (e.g. "finance")
// is returned as a match so its reason reaches the caller's log — the caller
// treats it as no-write.
func matchVerticalDirectory(industry, siteType, category, domain string, logger *zap.Logger) *verticalDirectoryConfig {
	signals := []string{
		strings.ToLower(industry),
		strings.ToLower(siteType),
		strings.ToLower(category),
	}

	// Domain-derived signal. Append AT MOST ONE, chosen by the same rule the
	// partial matcher below uses — longest key wins, lexicographic tie-break.
	//
	// It cannot append every match: `for k := range map` is randomised, this
	// map deliberately mixes recommending and NOT-recommending entries, and the
	// dispatch loop below returns on the FIRST exact match — so a domain
	// containing two opposite keywords resolved differently run to run. That is
	// the identical defect the partial-match arm already guards against; it
	// lived one level up, in the loop that BUILDS the signal list.
	//
	// Reachable, not hypothetical: `mortgage-refinance.co.uk` (portfolio
	// register M4, the Phase C pilot's own family) contains "mortgage"
	// (recommended) AND — inside "refinance" — "finance" (deliberately not
	// recommended). Reproduced on iteration 1 of
	// TestMatchVerticalDirectory_DomainSignalIsDeterministic before this fix.
	//
	// Longest-wins is also the RIGHT answer here, not merely a stable one: the
	// specific provider class ("mortgage") is what a remortgage/refinance site
	// needs, and "finance" is not-recommended precisely because it is too
	// generic to choose one.
	domainLower := strings.ToLower(domain)
	bestDomainKey := ""
	for keyword := range verticalDirectoryMap {
		if !strings.Contains(domainLower, strings.ReplaceAll(keyword, " ", "")) {
			continue
		}
		if len(keyword) > len(bestDomainKey) ||
			(len(keyword) == len(bestDomainKey) && keyword < bestDomainKey) {
			bestDomainKey = keyword
		}
	}
	if bestDomainKey != "" {
		signals = append(signals, bestDomainKey)
	}

	for _, signal := range signals {
		if signal == "" {
			continue
		}
		if config, ok := verticalDirectoryMap[signal]; ok {
			logger.Info("matchVerticalDirectory: matched",
				zap.String("signal", signal),
				zap.Bool("recommended", config.Recommended))
			return &config
		}
		// Partial match — UNLIKE matchVerticalNews, this must not range the
		// map directly: map iteration order is random, and this map (unlike
		// the news one) mixes recommending and non-recommending entries, so a
		// signal containing two keys (e.g. "insurance finance") would flip
		// between opposite outcomes run to run. Pick the LONGEST matching key
		// (most specific wins: "health insurance" beats "insurance"), with a
		// lexicographic tie-break for determinism.
		best := ""
		for key := range verticalDirectoryMap {
			if !strings.Contains(signal, key) {
				continue
			}
			if len(key) > len(best) || (len(key) == len(best) && key < best) {
				best = key
			}
		}
		if best != "" {
			config := verticalDirectoryMap[best]
			logger.Info("matchVerticalDirectory: partial match",
				zap.String("signal", signal),
				zap.String("matched_key", best),
				zap.Bool("recommended", config.Recommended))
			return &config
		}
	}

	return nil
}
