// FILE: platform/orchestration/actions/discovery_checks/check_image_url_404.go
//
// Discovery check: rendered HTML references an /assets/images/* path that
// doesn't correspond to any assets row for this site. This is the symptom
// of a component template that hardcodes an image reference for which no
// asset was ever generated, OR a stale reference from a regenerate that
// removed the asset.
//
// This is the lightweight DB-only version (per PLAN section 1.3): we don't
// check whether the file is actually in git, only whether an assets row
// exists. The full HTTP/git version would catch deployment failures too,
// but is deferred until we have git-adapter integration on the discovery
// path.
//
// READ THIS BEFORE TRUSTING THE NAME OR THE SILENCE (bugs_open/128, measured
// 2026-07-29). The paragraph above understates the limitation in a way that
// matters. What this check actually compares is a rendered path's BASENAME
// against the set of active asset PURPOSES for the site (loadKnownAssetPurposes)
// — and purposes are names like "hero"/"icon"/"logo", not paths. So:
//
//   - Owning ONE asset of a purpose, at any URL including an S3 one, makes every
//     rendered path sharing that purpose's name unreportable. The `root` fallback
//     below widens that to the whole `hero-*`, `icon-*`, `logo-*` prefix space.
//   - Measured fleet-wide: **79 of 95** distinct rendered image paths across 13
//     sites are masked this way — 83% of the surface — and SIX of the masked
//     paths were serving live 404s at the time of measurement.
//   - So a FINDING here means "no asset purpose matches this basename", which is
//     neither HTTP status nor really registration; and ABSENCE means almost
//     nothing at all. Do not read a clean run as "no broken images".
//
// A path-based predicate was measured and REFUTED as the fix: only 9 of those 79
// paths have an assets row whose url/filename carries the basename, while 73 of
// the 79 serve 200 — nothing in the DB records which static files were deployed,
// so swapping the predicate would flag ~70 working images. See
// check_image_url_404_masking_test.go, which pins this behaviour deliberately and
// carries the trap: unmasking activates the knownPurposeMapping routing below,
// which has NEVER fired in production (0 of 10 items ever created under an
// image_url_404 key were needs_hero_image/needs_logo).
//
// Routing: the path's basename (without extension) is matched against
// known purpose names. If a known purpose, route to image-build-handler.
// Otherwise emit a flag-only finding — a human or a follow-up audit needs
// to decide whether the reference itself is wrong (template bug) or the
// missing asset is wrong (generation gap).

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

func init() { Register(&ImageURL404Check{}) }

type ImageURL404Check struct{}

func (c *ImageURL404Check) Name() string { return "image_url_404" }

// imagePathRefPattern matches references like /assets/images/<basename>.<ext>
// or /assets/images/<basename>-<variant>.<ext>. The non-greedy match on
// the basename keeps multi-segment paths whole for downstream interpretation.
var imagePathRefPattern = regexp.MustCompile(
	`/assets/images/([a-zA-Z0-9_\-]+)(?:-[a-zA-Z0-9_\-]+)?\.(?:jpg|jpeg|png|webp|svg|gif)`,
)

// knownPurposeMapping maps an image basename to the purpose-and-handler
// combination used to repair it. Anything not in this map is a flag-only
// finding (no automatic regeneration).
var knownPurposeMapping = map[string]struct {
	purpose  string
	itemType string
	priority int
}{
	"hero": {"hero", "needs_hero_image", 60},
	"logo": {"logo", "needs_logo", 65},
}

func (c *ImageURL404Check) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	references, err := collectImagePathReferences(dctx)
	if err != nil {
		return nil, err
	}
	if len(references) == 0 {
		return &CheckResult{}, nil
	}

	knownPurposes, err := loadKnownAssetPurposes(dctx)
	if err != nil {
		return nil, err
	}

	result := &CheckResult{}

	for basename, samplePath := range references {
		// Skip if any active asset for this site has a purpose that matches
		// the basename or the basename's prefix (handles hero-about, hero-services).
		if knownPurposes[basename] {
			continue
		}
		// Permit hero variant naming: hero-<page> resolves to hero
		root := basename
		if idx := strings.Index(basename, "-"); idx > 0 {
			root = basename[:idx]
		}
		if knownPurposes[root] {
			continue
		}

		mapping, recognised := knownPurposeMapping[root]

		spec := map[string]interface{}{
			"check":    "image_url_404",
			"basename": basename,
			"path":     samplePath,
		}
		if recognised {
			spec["purpose"] = mapping.purpose
		}
		specJSON, _ := json.Marshal(spec)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":      "image_url_404",
			"basename":   basename,
			"path":       samplePath,
			"recognised": recognised,
		})

		// Flag-only when we can't confidently say what should be regenerated.
		// The dispatch loop ignores items with no handler_agent, so they
		// stay in detected status as a record without consuming budget.
		if !recognised {
			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:    dctx.SiteID,
				Source:    "discovery",
				Pipeline:  dctx.Pipeline,
				ItemType:  "image_url_404",
				Severity:  "medium",
				Summary:   fmt.Sprintf("Pages reference unknown image %s", samplePath),
				SpecJSON:  string(specJSON),
				Priority:  40,
				Status:    "detected",
				CreatedBy: dctx.AgentType,
				ItemKey:   fmt.Sprintf("image_url_404:%s", basename),
				BatchID:   dctx.BatchID,
				// HandlerAgent intentionally empty — flag-only.
			})
			continue
		}

		// Recognised purpose: route to image-build-handler.
		promptKey := mapping.purpose
		if mapping.purpose == "hero" {
			promptKey = "hero_home"
		}
		prompts, _ := loadImagePromptsForSite(dctx)
		if p := prompts[promptKey]; p != "" {
			spec["image_prompts"] = map[string]interface{}{promptKey: p}
			spec["prompt_key"] = promptKey
			specJSON, _ = json.Marshal(spec)
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     dctx.Pipeline,
			ItemType:     mapping.itemType,
			Severity:     "high",
			Summary:      fmt.Sprintf("Pages reference %s but no %s asset exists", samplePath, mapping.purpose),
			SpecJSON:     string(specJSON),
			Priority:     mapping.priority,
			HandlerAgent: "image-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("image_url_404:%s", basename),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

// collectImagePathReferences scans rendered_html across deployed unlocked
// components, extracting unique /assets/images/* basenames and a sample
// path for each. Returns map[basename]samplePath.
func collectImagePathReferences(dctx DiscoveryCheckContext) (map[string]string, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.rendered_html
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html LIKE '%/assets/images/%'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("page_components scan failed: %w", err)
	}
	defer rows.Close()

	refs := make(map[string]string)
	for rows.Next() {
		var html string
		if err := rows.Scan(&html); err != nil {
			dctx.Logger.Warn("image_url_404: scan row failed", zap.Error(err))
			continue
		}
		matches := imagePathRefPattern.FindAllStringSubmatch(html, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			basename := m[1]
			if _, exists := refs[basename]; exists {
				continue
			}
			// Keep the first full sample path we see for this basename.
			full := m[0]
			refs[basename] = filepath.ToSlash(full)
		}
	}
	return refs, rows.Err()
}

// loadKnownAssetPurposes returns the set of active asset purposes for the
// site, used to determine which path basenames already correspond to a
// real asset.
func loadKnownAssetPurposes(dctx DiscoveryCheckContext) (map[string]bool, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT purpose
		FROM assets
		WHERE site_id = $1
		  AND purpose IS NOT NULL
		  AND status = 'active'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("assets purposes query failed: %w", err)
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			dctx.Logger.Warn("image_url_404: scan purpose failed", zap.Error(err))
			continue
		}
		out[p] = true
	}
	return out, rows.Err()
}
