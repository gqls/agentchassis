// FILE: platform/orchestration/actions/discovery_checks/check_image_url_404.go
//
// Discovery check: a deployed page (or the site chrome) references an
// /assets/images/* path that NO active asset of this site would be deployed to.
// That is the symptom of a component template hardcoding an image reference for
// which no asset was ever generated, or a stale reference left behind when the
// asset it named was regenerated under a different key.
//
// HOW THE QUESTION IS ANSWERED, and why it can be answered without HTTP
// (rewritten 2026-07-31, bugs_open/128).
//
// storage.DeployedWebPath(asset_key, purpose) is the platform's single source of
// truth for "the web path a generated asset is committed to and served from". The
// writers all resolve through it — plan_sections_action, render_site_components_action,
// emit_sprite_css_action, derive_card_asset_action, queryresolve — and
// deploy_image_asset_action commits to exactly that path via the shared
// storage.AssetKeyFilename. So the set of paths this site's assets occupy is
// computable from the assets table, and this check is simply the INVERSE of the
// render-time resolver: a finding means "nothing this site owns lands here".
//
// That keeps the promise recorded in verifier_coverage_test.go:171 — no outbound
// HTTP on the discovery or completion path — while still answering the question
// the check is named for.
//
// WHAT THIS REPLACED, and the measurement that condemned it. Until 2026-07-31 the
// check compared a rendered path's BASENAME against the set of active asset
// PURPOSES ("hero", "icon", "logo") and skipped on a match, or on a match of the
// prefix before the first hyphen. Purposes are not paths. Owning one hero asset —
// at any URL, including an S3 one never served from the site — made every rendered
// `hero*` path unreportable. Measured over all 127 distinct rendered image paths
// on 13 live sites with HTTP status as ground truth:
//
//	                                    reports a WORKING image | SILENT on a broken one
//	  purpose/prefix skip (old)                              21 | 6
//	  DeployedWebPath (this file)                             1 | 0
//
// The six it could not see were /assets/images/hero.jpg on dartsonline,
// gamesdesign, idea.uk, oufe, relojistas and vonc — broken on every one.
//
// THE ONE RESIDUAL FALSE POSITIVE, stated rather than hidden. A file committed to
// the site repo by no asset row — webdesign.co.uk's legacy /assets/images/hero.jpg,
// 455KB, serving 200 — is reported here, because the database genuinely does not
// know it exists. That is 1 of 127, and it is arguably a true finding of a
// different thing: a served file no pipeline maintains and any repo reconciliation
// would delete.
//
// SURFACES SCANNED. page_components (deployed, unlocked) AND site_components — the
// head/header chrome, which appears on EVERY page and was never scanned before
// (bugs_open/128 defect 3; it is how idea.uk shipped a 404 favicon and og-card
// site-wide without a single finding).
//
// WHAT THIS CHECK DOES NOT OWN, so a reader knows where the neighbouring answer is:
//
//   - "an asset row exists but its file was never deployed" → check_undeployed_assets.
//     This check is silent there by design: the path resolves as far as the DB is
//     concerned. gaswholesalers.com's logo.png is the live example.
//   - "a page references the documented FALLBACK path and no asset of that purpose
//     exists, so build one" → check_placeholder_image_in_use, which owns that
//     repair and routes it to image-build-handler. This check used to carry a
//     duplicate of that branch (same paths, same purposes, same needs_hero_image /
//     needs_logo item types, same handler, same precondition, both enabled on
//     design-discovery-agent, neither ever fired). The duplicate is gone; a
//     finding here is always flag-only.
//   - "a component asks for an image the pipeline cannot supply" →
//     check_image_source_unsatisfiable, which reads input_schema and catches the
//     CAUSE. The empty-src emission below catches the SYMPTOM in rendered HTML,
//     which the schema-side check cannot see when the empty string is hardcoded in
//     Go rather than resolved from a schema.
//
// Every emission is item_type image_url_404 with no handler agent: a stale
// reference is repaired by removing or repointing it, which no image generator can
// decide. spec.kind distinguishes the two shapes ("unbacked_path", "empty_src").

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

func init() { Register(&ImageURL404Check{}) }

type ImageURL404Check struct{}

func (c *ImageURL404Check) Name() string { return "image_url_404" }

// imagePathRefPattern matches references like /assets/images/<name>.<ext>.
// The whole name is captured: hyphens are part of the filename, never a
// separator this check may reason about (reasoning about the prefix before the
// first hyphen is exactly what the purpose skip did wrong).
var imagePathRefPattern = regexp.MustCompile(
	`/assets/images/([a-zA-Z0-9_\-]+\.(?:jpg|jpeg|png|webp|svg|gif))`,
)

// emptyImgSrcPattern matches an <img> whose src cannot resolve to an image:
// empty, whitespace-only, or the "#" placeholder. Browsers paint the broken-image
// icon for these, and per the HTML spec an empty src resolves against the current
// document — so the page re-requests ITSELF as an image and any HTTP-based checker
// would score it 200. It has no path, so no path predicate can see it; this is the
// only structural way to catch it.
var emptyImgSrcPattern = regexp.MustCompile(
	`(?is)<img\b[^>]*\bsrc\s*=\s*("\s*"|'\s*'|"#"|'#')`,
)

// maxEmptySrcSamples bounds the sample list carried in the work item spec. The
// count is always exact; only the examples are capped.
const maxEmptySrcSamples = 5

// imageSurface records where a reference was found, so a finding can say which
// surface to repair. Chrome is site-wide: one bad path there is on every page.
type imageSurface struct {
	page   bool
	chrome bool
}

func (s imageSurface) String() string {
	switch {
	case s.page && s.chrome:
		return "page+chrome"
	case s.chrome:
		return "chrome"
	default:
		return "page"
	}
}

func (c *ImageURL404Check) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	references, emptySrc, err := collectImageReferences(dctx)
	if err != nil {
		return nil, err
	}

	result := &CheckResult{}

	if len(references) > 0 {
		deployed, err := loadDeployedAssetPaths(dctx)
		if err != nil {
			return nil, err
		}

		// Sorted so a run's findings and work items are ordered deterministically
		// — two runs over the same data must produce the same sequence.
		paths := make([]string, 0, len(references))
		for p := range references {
			paths = append(paths, p)
		}
		sort.Strings(paths)

		for _, path := range paths {
			if deployed[path] {
				continue
			}
			surface := references[path]
			filename := path[strings.LastIndex(path, "/")+1:]

			spec := map[string]interface{}{
				"check":    "image_url_404",
				"kind":     "unbacked_path",
				"path":     path,
				"filename": filename,
				"surface":  surface.String(),
				// basename retained for readers of older items: the dedup key was
				// the extension-less basename until 2026-07-31.
				"basename": strings.TrimSuffix(filename, "."+extensionOf(filename)),
			}
			specJSON, _ := json.Marshal(spec)

			result.Findings = append(result.Findings, map[string]interface{}{
				"check":   "image_url_404",
				"kind":    "unbacked_path",
				"path":    path,
				"surface": surface.String(),
			})

			summary := fmt.Sprintf("Pages reference %s but no active asset deploys to that path", path)
			if surface.chrome {
				summary = fmt.Sprintf("Site chrome references %s on every page but no active asset deploys to that path", path)
			}

			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:    dctx.SiteID,
				Source:    "discovery",
				Pipeline:  dctx.Pipeline,
				ItemType:  "image_url_404",
				Severity:  severityForSurface(surface),
				Summary:   summary,
				SpecJSON:  string(specJSON),
				Priority:  40,
				Status:    "detected",
				CreatedBy: dctx.AgentType,
				// The extension is part of the key: /assets/images/logo.jpg and
				// /assets/images/logo.png are two files with two HTTP results
				// (fundamentallyai.com serves 200 and 404 for that exact pair).
				// An extension-blind key lets idx_swi_dedup silently drop the
				// second finding — the failure mode of bugs_open/091.
				ItemKey: fmt.Sprintf("image_url_404:%s", filename),
				BatchID: dctx.BatchID,
				// HandlerAgent intentionally empty — flag-only. Repairing a stale
				// reference means removing or repointing it, which no image
				// generator can decide. Generation is check_placeholder_image_in_use's
				// remit, under its own precondition.
			})
		}
	}

	if emptySrc.count > 0 {
		spec := map[string]interface{}{
			"check":   "image_url_404",
			"kind":    "empty_src",
			"count":   emptySrc.count,
			"samples": emptySrc.samples,
		}
		specJSON, _ := json.Marshal(spec)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "image_url_404",
			"kind":  "empty_src",
			"count": emptySrc.count,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:    dctx.SiteID,
			Source:    "discovery",
			Pipeline:  dctx.Pipeline,
			ItemType:  "image_url_404",
			Severity:  "medium",
			Summary:   fmt.Sprintf("%d <img> tags render with no image source (empty or '#' src)", emptySrc.count),
			SpecJSON:  string(specJSON),
			Priority:  40,
			Status:    "detected",
			CreatedBy: dctx.AgentType,
			ItemKey:   "image_url_404:empty-src",
			BatchID:   dctx.BatchID,
		})
	}

	return result, nil
}

// severityForSurface: a bad path in the chrome is on every page of the site, so
// it outranks one bad path on one page.
func severityForSurface(s imageSurface) string {
	if s.chrome {
		return "high"
	}
	return "medium"
}

func extensionOf(filename string) string {
	if i := strings.LastIndex(filename, "."); i >= 0 {
		return filename[i+1:]
	}
	return ""
}

// emptySrcTally is the site-wide count of <img> tags with no usable source,
// plus a bounded sample of the surrounding markup for a human to locate them.
type emptySrcTally struct {
	count   int
	samples []string
}

// collectImageReferences scans BOTH rendered surfaces — deployed unlocked page
// components and the site chrome — returning every distinct /assets/images/*
// path with the surfaces it appeared on, and the empty-src tally.
//
// Locked components are skipped on both surfaces: a human-locked component is
// presumed deliberate, the same convention isPathReferencedInPages follows.
func collectImageReferences(dctx DiscoveryCheckContext) (map[string]imageSurface, emptySrcTally, error) {
	refs := make(map[string]imageSurface)
	var empty emptySrcTally

	scan := func(html string, chrome bool) {
		for _, m := range imagePathRefPattern.FindAllStringSubmatch(html, -1) {
			if len(m) < 2 {
				continue
			}
			path := m[0]
			s := refs[path]
			if chrome {
				s.chrome = true
			} else {
				s.page = true
			}
			refs[path] = s
		}
		for _, m := range emptyImgSrcPattern.FindAllString(html, -1) {
			empty.count++
			if len(empty.samples) < maxEmptySrcSamples {
				empty.samples = append(empty.samples, strings.TrimSpace(m))
			}
		}
	}

	pageRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.rendered_html
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html IS NOT NULL
	`, dctx.SiteID)
	if err != nil {
		return nil, empty, fmt.Errorf("page_components scan failed: %w", err)
	}
	for pageRows.Next() {
		var html string
		if err := pageRows.Scan(&html); err != nil {
			dctx.Logger.Warn("image_url_404: scan page component failed", zap.Error(err))
			continue
		}
		scan(html, false)
	}
	if err := pageRows.Err(); err != nil {
		pageRows.Close()
		return nil, empty, fmt.Errorf("page_components scan failed: %w", err)
	}
	pageRows.Close()

	// The chrome surface (bugs_open/128 defect 3). site_components carries no
	// build_status='deployed' contract comparable to page_components — it is the
	// stored artefact the whole site renders (bugs_open/117) — so every unlocked
	// row is in scope.
	chromeRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.rendered_html
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.locked_at IS NULL
		  AND sc.rendered_html IS NOT NULL
	`, dctx.SiteID)
	if err != nil {
		return nil, empty, fmt.Errorf("site_components scan failed: %w", err)
	}
	defer chromeRows.Close()
	for chromeRows.Next() {
		var html string
		if err := chromeRows.Scan(&html); err != nil {
			dctx.Logger.Warn("image_url_404: scan site component failed", zap.Error(err))
			continue
		}
		scan(html, true)
	}
	return refs, empty, chromeRows.Err()
}

// loadDeployedAssetPaths returns the set of site-relative web paths this site's
// ACTIVE assets are deployed and served at.
//
// It is computed with storage.DeployedWebPath — the same helper every writer
// resolves through — so the check and the renderers cannot drift: if a writer
// starts emitting a different path, this predicate moves with it.
//
// TWO SOURCES, and the branch between them is load-bearing. Brand-head assets
// (favicon, og_card) do NOT go through DeployedWebPath: their published filenames
// are not derivable from the purpose — og_card publishes as og-card.png with a
// HYPHEN, where DeployedWebPath would return og_card.png with an underscore. That
// is precisely the drift storage.BrandHeadAssetPaths was added to end
// (bugs_open/142), and its own doc comment says callers reasoning about "is this
// deployed?" must branch on storage.IsBrandHeadPurpose. Getting this wrong is not
// a near miss: it would report a 404 for the og card and the favicon on every
// site in the fleet, both of which are referenced from the head on every page.
//
// HOW FAR THAT DRIFT GOES — audited exhaustively 2026-07-31, because a council
// seat was right to ask whether adopting DeployedWebPath as ground truth inherits
// a defect patched at one call site and left generic everywhere else. It does not,
// and here is the whole of the argument, so nobody has to re-derive it:
//
//  1. The divergence has exactly ONE mechanism. DeployedWebPath applies
//     AssetKeyFilename's `_`→`-` swap only when assetKey differs from purpose;
//     otherwise it returns purpose+ext verbatim. So a wrong path requires a purpose
//     that CONTAINS an underscore AND an asset stored with assetKey empty or equal
//     to it.
//  2. Over all 267 active asset rows fleet-wide, the rows taking that skip are
//     favicon x12, og_card x12 (both brand-head, both handled by the branch below),
//     hero x5 and logo x4 — and hero/logo have no underscore to mis-render. Every
//     other underscore purpose is stored with a distinct key (content_hero x31,
//     sprite_sheet x1) and therefore takes the swap. The risk set outside brand-head
//     is EMPTY, and it is empty structurally, not by luck.
//  3. deploy_image_asset_action.go:185-196 branches on the identical condition
//     (`if assetKey != "" && assetKey != purpose`), which is why the helper's doc
//     comment says it mirrors the deployer. The two cannot disagree except where
//     something OTHER than the deployer publishes the file — which is exactly the
//     brand-head pair, published directly by derive_brand_head_assets_action.
//  4. The deploy_path override (deploy_image_asset_action.go:203-220) could produce
//     a third spelling, but no Go code sets it and it appears in ZERO orchestrations
//     in history — it is an unused passthrough.
//  5. Empirically, which is the check that would catch a mechanism nobody thought of:
//     of the 127 rendered paths measured live, 109 matched this predicate and ALL 109
//     serve 200. A path this predicate cannot express would show up as a working file
//     at an unpredicted path; exactly one did, and it has no asset row at all.
func loadDeployedAssetPaths(dctx DiscoveryCheckContext) (map[string]bool, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT COALESCE(asset_key, ''), COALESCE(purpose, '')
		FROM assets
		WHERE site_id = $1
		  AND status = 'active'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("assets query failed: %w", err)
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var assetKey, purpose string
		if err := rows.Scan(&assetKey, &purpose); err != nil {
			dctx.Logger.Warn("image_url_404: scan asset failed", zap.Error(err))
			continue
		}
		if assetKey == "" && purpose == "" {
			// Nothing to derive a path from; such a row cannot back any
			// reference, and DeployedWebPath("","") would yield "/assets/images/.jpg".
			continue
		}
		if storage.IsBrandHeadPurpose(purpose) {
			out[storage.BrandHeadAssetPaths[purpose]] = true
			continue
		}
		out[storage.DeployedWebPath(assetKey, purpose)] = true
	}
	return out, rows.Err()
}
