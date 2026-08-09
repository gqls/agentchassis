// FILE: platform/orchestration/actions/discovery_checks/check_placeholder_image_in_use.go
//
// Discovery check: rendered HTML contains the fallback image path
// (e.g. /assets/images/hero.jpg) AND no assets row exists for the matching
// purpose. The component's input_schema fallback resolved because the asset
// was never generated; the page is silently shipping the system default.
//
// This is the symptom you see when image-build-handler never ran for the
// site, or when its run completed in error. The unfulfilled_image_prompt
// check catches the upstream cause (planner asked, no asset); this check
// catches the downstream symptom (no asset, fallback rendered) — the two
// can fire together but for different reasons:
//   - unfulfilled_image_prompt: planner provided a prompt
//   - placeholder_image_in_use: a page references the fallback
// They produce identical work items in practice; deduping is by ItemKey.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"go.uber.org/zap"
)

func init() { Register(&PlaceholderImageInUseCheck{}) }

type PlaceholderImageInUseCheck struct{}

func (c *PlaceholderImageInUseCheck) Name() string { return "placeholder_image_in_use" }

// placeholderPathMapping maps the documented fallback path to the assets.purpose
// value that, if missing, indicates we're rendering the placeholder for real.
// Paths come from 003_contracts_and_standards_v7.md.
var placeholderPathMapping = []struct {
	path     string // the fallback path that appears in rendered_html
	purpose  string // assets.purpose to check for
	itemType string // work item type for image-build-handler
	priority int
}{
	{"/assets/images/hero.jpg", "hero", "needs_hero_image", 65},
	{"/assets/images/logo.png", "logo", "needs_logo", 70},
}

func (c *PlaceholderImageInUseCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	for _, mapping := range placeholderPathMapping {
		referenced, err := isPathReferencedInPages(dctx, mapping.path)
		if err != nil {
			dctx.Logger.Warn("placeholder_image_in_use: rendered_html scan failed",
				zap.String("path", mapping.path),
				zap.Error(err))
			continue
		}
		if !referenced {
			continue
		}

		hasAsset, err := hasActiveAssetForPurpose(dctx, mapping.purpose)
		if err != nil {
			dctx.Logger.Warn("placeholder_image_in_use: asset existence check failed",
				zap.String("purpose", mapping.purpose),
				zap.Error(err))
			continue
		}
		if hasAsset {
			continue // asset exists; this isn't a placeholder use, the deployed
			// path just happens to match the canonical name for that purpose.
		}

		// Recover the planner's intent if it exists. IT USUALLY DOES NOT:
		// measured 2026-08-09, 34 of 39 sites have no usable image_prompts
		// (33 of them have no current site_plan spec row at all), so the
		// promptless branch below is the NORMAL case, not an edge case.
		//
		// THERE IS NO HANDLER FALLBACK. The comment that used to sit here
		// claimed one — "the handler will fall back to its default prompt
		// template" — and that claim was the bug (bugs_open/210).
		// image-build-handler's call_logo_gen maps "prompt" from
		// input_data.spec.image_prompts.logo as a REQUIRED field, and
		// input_mapping is a strict allow-list (input_mapping.go:122-136;
		// only a "?" suffix on the destination makes a path optional), so a
		// promptless item dies at input extraction one step BEFORE any
		// fallback could run. Census: 2 of 2 promptless items failed, 4 of 4
		// items carrying the key completed.
		promptKey := mapping.purpose
		if mapping.purpose == "hero" {
			promptKey = "hero_home"
		}
		prompts, _ := loadImagePromptsForSite(dctx)
		prompt := prompts[promptKey]

		spec := map[string]interface{}{
			"check":      "placeholder_image_in_use",
			"purpose":    mapping.purpose,
			"path":       mapping.path,
			"prompt_key": promptKey,
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":        "placeholder_image_in_use",
			"purpose":      mapping.purpose,
			"path":         mapping.path,
			"prompt_known": prompt != "",
		})

		if prompt == "" {
			// No prompt, and no automated author for one. A logo's appearance
			// is a human decision on this platform by design, not by omission:
			// imagery_style_guide.go excludes logos from style-guide direction
			// entirely ("logos get nothing — the 2026-05-20 contamination
			// lesson") and generate_image_actions.go states the intended
			// lifecycle as "generated once, human-approved, then locked".
			//
			// So file the finding for a human rather than a generation request
			// that is guaranteed to fail. Same ItemKey as the generation item
			// deliberately: one open question per (site, purpose), never both.
			spec["reason"] = "no planned prompt for this purpose; generation cannot be requested"
			specJSON, _ := json.Marshal(spec)

			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:   dctx.SiteID,
				Source:   "discovery",
				Pipeline: dctx.Pipeline,
				ItemType: mapping.itemType,
				Severity: "high",
				Summary: fmt.Sprintf(
					"Pages reference fallback %s, no asset exists, and no %s prompt was ever planned — needs a human ruling on the brand direction",
					mapping.path, mapping.purpose),
				SpecJSON: string(specJSON),
				Priority: mapping.priority,
				// The EMPTY handler is the canonical spelling of a deliberate
				// no-agent route for a needs_human_review item (migration 217;
				// handler_coverage_test.go:53-57), and it is what the fleet
				// actually does — 433 rows at this status carry no handler
				// against 12 naming "human-review", which is not an
				// agent_definitions type at all. Status is what keeps the row
				// out of dispatch (load_work_item_actions.go:951), so it never
				// reaches claim and is never marked 'blocked'.
				HandlerAgent: "",
				Status:       "needs_human_review",
				CreatedBy:    dctx.AgentType,
				ItemKey:      fmt.Sprintf("placeholder_image_in_use:%s", mapping.purpose),
				BatchID:      dctx.BatchID,
			})
			continue
		}

		spec["image_prompts"] = map[string]interface{}{promptKey: prompt}
		// The flat key too, matching check_unfulfilled_image_prompt's dual
		// format (:101-105). The two Phase-2E branches read spec.prompt; the
		// two legacy branches still read the nested image_prompts.<key>.
		// Writing both means this producer is correct under either, and does
		// not have to be revisited when that convergence is finished.
		spec["prompt"] = prompt
		specJSON, _ := json.Marshal(spec)

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     dctx.Pipeline,
			ItemType:     mapping.itemType,
			Severity:     "high",
			Summary:      fmt.Sprintf("Pages reference fallback %s but no asset exists", mapping.path),
			SpecJSON:     string(specJSON),
			Priority:     mapping.priority,
			HandlerAgent: "image-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("placeholder_image_in_use:%s", mapping.purpose),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

// sameOriginPathPattern returns a POSIX regex matching path only where it is
// referenced as a SAME-ORIGIN (root-relative) URL.
//
// A cross-origin reference contains the identical substring and is NOT this
// site rendering its own placeholder:
//
//	<img src="https://partner.example/assets/images/logo.png">
//
// The character immediately before the leading "/" discriminates exactly: a
// host ends in a letter, digit, "." or "-", while a root-relative reference is
// preceded by a quote, "(", "=" or whitespace — src="/a", url('/a'), url(/a),
// srcset="… /a 2x", href=/a.
//
// Measured 2026-08-09: the cross-origin shape was the ONLY
// /assets/images/logo.png match in the entire fleet, and it filed a false
// needs_logo item against fundamentallyai.com for rendering a partner's logo
// from the partner's own domain (bugs_open/210 §6). Same family as
// bugs_closed/128 — a comparison that ignores where a file actually comes from.
func sameOriginPathPattern(path string) string {
	return `['"(=[:space:]]` + regexp.QuoteMeta(path)
}

// isPathReferencedInPages returns true if this site renders the given fallback
// path on either of its two surfaces. Locked rows are skipped on both: a
// human-locked component referencing the fallback is presumably intentional.
//
// SURFACES SCANNED — page_components (deployed, unlocked) AND site_components.
// A LOGO is rendered by the site chrome, not by page content, so scanning only
// page_components left the check that ROUTES a logo repair blind to the surface
// logos live on. bugs_closed/128 found that blindness and fixed it in the
// flag-only sibling (check_image_url_404.go:471-482) while leaving this one
// unfixed — the pair-keyed landmine at LANDMINES.md:1996, firing verbatim.
// site_components has no build_status='deployed' contract comparable to
// page_components (it is the stored artefact the whole site renders,
// bugs_open/117), so every unlocked row is in scope — the sibling's contract.
func isPathReferencedInPages(dctx DiscoveryCheckContext, path string) (bool, error) {
	pattern := sameOriginPathPattern(path)

	var n int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html ~ $2
		LIMIT 1
	`, dctx.SiteID, pattern).Scan(&n)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if n > 0 {
		return true, nil
	}

	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.locked_at IS NULL
		  AND sc.rendered_html IS NOT NULL
		  AND sc.rendered_html ~ $2
		LIMIT 1
	`, dctx.SiteID, pattern).Scan(&n)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return n > 0, nil
}
