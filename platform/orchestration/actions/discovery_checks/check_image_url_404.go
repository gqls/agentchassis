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
// decide. spec.kind distinguishes the three shapes ("unbacked_path", "empty_src",
// "bare_token_src").
//
// THE THIRD SHAPE, added 2026-08-03. An <img> whose src is a bare word —
// <img src="cpu"> — is invisible to both predicates above: it is not an
// /assets/images/… path and it is not empty. Measured that day, 31 such tags were
// live across 2 sites from 1 component, every one a broken image, with no finding
// on any surface. See bareTokenImgSrcPattern for the shape and why "no dot" is
// safe on this fleet.

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

// bareTokenImgSrcPattern matches an <img> whose src is a single bare word — no
// slash, no dot, no colon — e.g. <img src="cpu">. It is neither a path (the
// pattern above needs /assets/images/…) nor empty (the pattern above that needs
// "" or "#"), so until 2026-08-03 BOTH predicates were silent on it and the
// browser painted a broken-image icon on a live page with no finding anywhere.
//
// WHAT IT ACTUALLY CATCHES, and why the shape is diagnostic rather than
// incidental. A bare token in a src slot is what an ICON NAME looks like when a
// template renders it into an <img> instead of an icon element. departments-grid
// is the worked case: schema declares `icon: string`, the fleet renders icons as
// <i data-lucide="{{.icon}}"> (see the features component, which shares a page
// with it and works), but departments-grid still carried the <img src> of the
// team-photo component it was forked from. Measured 2026-08-03 over every
// rendered surface in the fleet: 31 occurrences, 2 sites, 1 component, and every
// single one a broken image in a browser.
//
// WHY NO DOT IS A SAFE PREDICATE HERE rather than a guess. Everything this
// platform serves as an image is committed under /assets/images/<key>.<ext> —
// an extension is structural, not conventional. The sites are static (no content
// negotiation, no extensionless image routes), so a dotless, slashless src cannot
// resolve to an image on any of them. Excluding ':' keeps scheme-bearing and
// protocol-relative srcs out; excluding '/' keeps every relative and rooted path
// out; excluding '.' keeps every filename out. What remains is only the bare-word
// shape.
//
// NOTE, because two council seats read the paragraph above as a DEPENDENCY and
// objected on the strength of the DeployedWebPath landmine: this predicate never
// calls storage.DeployedWebPath, and nothing in the bare-token branch consults
// the assets table at all. loadDeployedAssetPaths runs only in the
// len(references) > 0 branch above. The sentence is background for WHY an
// extension is structural on this estate; it is not a resolution step, so the
// helper's brand-head quirk (og_card/favicon, fixed at HEAD in bugs_closed/168)
// cannot reach this shape. A bare word has no extension to get wrong.
//
// AND IT IS NOT A FIFTH CHECK. Four checks share the /assets/images space and
// the landmine keyed to this file warns that widening one silently competes with
// another. Each was read before this was added:
//   - image_url_404 (here)          — a rendered path no active asset deploys to;
//                                     REQUIRES the /assets/images/ prefix and an extension
//   - placeholder_image_in_use      — two LITERAL paths, /assets/images/hero.jpg
//                                     and /assets/images/logo.png
//   - undeployed_assets             — queries the assets table, not rendered HTML
//   - unfulfilled_image_prompt      — reads the planner's prompts, not rendered HTML
// Both of placeholder_image_in_use's paths contain '/' and '.', which this
// character class excludes, so there is NO input on which both fire — that is a
// property of the pattern, not a judgement about intent. This is a third KIND
// inside this check, alongside empty_src, which is already the pathless shape;
// "a rendered <img> that cannot resolve, including the pathless cases" is
// precisely what this file already owns.
//
// Deliberately NOT folded into the empty_src tally: that finding says "this img
// has no source at all", which a human fixes by supplying one. This one says
// "this img has a source that is not a source", which a human fixes by changing
// the TEMPLATE that put it there — a different repair, so a different kind.
// '#' is excluded as well as '/', '.' and ':': a fragment src is not a bare word
// but a placeholder, and emptyImgSrcPattern already owns it. Without that
// exclusion this pattern double-reports every "#" as both shapes — which is how
// TestImageURL404_EmptyImgSrcIsCountedAndReportedOnce caught it.
var bareTokenImgSrcPattern = regexp.MustCompile(
	`(?is)<img\b[^>]*\bsrc\s*=\s*"([^"/.:#\s]+)"`,
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
	references, emptySrc, bareSrc, err := collectImageReferences(dctx)
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

	if bareSrc.count > 0 {
		tokens := bareSrc.sortedTokens()
		spec := map[string]interface{}{
			"check":  "image_url_404",
			"kind":   "bare_token_src",
			"count":  bareSrc.count,
			"tokens": tokens,
		}
		specJSON, _ := json.Marshal(spec)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":  "image_url_404",
			"kind":   "bare_token_src",
			"count":  bareSrc.count,
			"tokens": tokens,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "image_url_404",
			// High: unlike a single stale path, this shape is emitted by a
			// TEMPLATE, so it repeats on every page that uses the component and
			// every site that mounts it.
			Severity: "high",
			Summary: fmt.Sprintf(
				"%d <img> tags have a bare word as src (%s) — a template is rendering an icon name into an image slot",
				bareSrc.count, strings.Join(tokens, ", ")),
			SpecJSON:  string(specJSON),
			Priority:  40,
			Status:    "detected",
			CreatedBy: dctx.AgentType,
			ItemKey:   "image_url_404:bare-token-src",
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

// bareTokenTally is the site-wide record of <img src="<bare word>">. Unlike the
// empty-src tally it keeps the DISTINCT tokens, because the token list is the
// evidence that identifies the cause: a set of lucide icon names in src slots
// names the defect ("a template renders an icon as an image") in a way a bare
// count never could.
type bareTokenTally struct {
	count  int
	tokens map[string]int
}

func (t *bareTokenTally) add(token string) {
	if t.tokens == nil {
		t.tokens = make(map[string]int)
	}
	t.count++
	t.tokens[token]++
}

// sortedTokens returns the distinct tokens in a stable order, so two runs over
// the same data produce the same work item spec.
func (t bareTokenTally) sortedTokens() []string {
	out := make([]string, 0, len(t.tokens))
	for tok := range t.tokens {
		out = append(out, tok)
	}
	sort.Strings(out)
	return out
}

// collectImageReferences scans BOTH rendered surfaces — deployed unlocked page
// components and the site chrome — returning every distinct /assets/images/*
// path with the surfaces it appeared on, and the empty-src tally.
//
// Locked components are skipped on both surfaces: a human-locked component is
// presumed deliberate, the same convention isPathReferencedInPages follows.
func collectImageReferences(dctx DiscoveryCheckContext) (map[string]imageSurface, emptySrcTally, bareTokenTally, error) {
	refs := make(map[string]imageSurface)
	var empty emptySrcTally
	var bare bareTokenTally

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
		for _, m := range bareTokenImgSrcPattern.FindAllStringSubmatch(html, -1) {
			if len(m) < 2 {
				continue
			}
			bare.add(m[1])
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
		return nil, empty, bare, fmt.Errorf("page_components scan failed: %w", err)
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
		return nil, empty, bare, fmt.Errorf("page_components scan failed: %w", err)
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
		return nil, empty, bare, fmt.Errorf("site_components scan failed: %w", err)
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
	return refs, empty, bare, chromeRows.Err()
}

// loadDeployedAssetPaths returns the set of site-relative web paths this site's
// ACTIVE assets are deployed and served at.
//
// It is computed with storage.DeployedWebPath — the same helper every writer
// resolves through — so the check and the renderers cannot drift: if a writer
// starts emitting a different path, this predicate moves with it.
//
// UPDATED 2026-08-02 (bugs_open/168). This function used to carry its own
// `storage.IsBrandHeadPurpose` branch, because the helper could not express
// og-card.png and would have returned og_card.png with an underscore — reporting
// a 404 for the og card and the favicon of every site in the fleet, both
// referenced from the head on every page. That branch is gone: storage.
// DeployedAssetPath consults storage.BrandHeadAssetPaths itself, so both writers
// are answered by one call. Having each consumer learn the exception separately
// was 016b §9 case 7, and this was the call site that nearly got it wrong.
//
// HOW FAR THE DRIFT GOES — audited exhaustively 2026-07-31, because a council
// seat was right to ask whether adopting DeployedWebPath as ground truth inherits
// a defect patched at one call site and left generic everywhere else. It did not,
// and the argument is kept so nobody has to re-derive it:
//
//  1. The divergence has exactly ONE mechanism. AssetKeyFilename's `_`→`-` swap
//     is applied only when assetKey differs from purpose; otherwise the path is
//     purpose+ext verbatim. So a wrong path requires a purpose that CONTAINS an
//     underscore AND an asset stored with assetKey empty or equal to it.
//  2. Over all 267 active asset rows fleet-wide, the rows taking that skip are
//     favicon x12, og_card x12 (both brand-head, both now answered by the helper),
//     hero x5 and logo x4 — and hero/logo have no underscore to mis-render. Every
//     other underscore purpose is stored with a distinct key (content_hero x31,
//     sprite_sheet x1) and therefore takes the swap. The risk set outside brand-head
//     is EMPTY, and it is empty structurally, not by luck. Re-measured 2026-08-02:
//     the same four groups, the same counts.
//  3. deploy_image_asset_action no longer implements this derivation at all — it
//     calls storage.DeployedAssetPath, the same function this check calls. Before
//     2026-08-02 it branched on the identical condition and the two were kept in
//     step by a doc comment; that is what bugs_open/168 replaced with one function.
//  4. The deploy_path override (deploy_image_asset_action.go) could still produce
//     a third spelling, and is invisible from (asset_key, purpose). No Go code sets
//     it and it appears in ZERO orchestrations in history — an unused passthrough,
//     and the one part of this contract that is not closed.
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
		out[storage.DeployedWebPath(assetKey, purpose)] = true
	}
	return out, rows.Err()
}
