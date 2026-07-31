# PLAN 2026-07-31 — `image_url_404`: compare PATHS, not purposes

Workstream opened 2026-07-31 to fix `bugs_open/128`. The bug was filed 07-28 and
diagnosed 07-29 by another session, which refuted the filed hypothesis and left a
pinning test. This lane picks it up, re-measures live, and fixes it.

## What the check is supposed to answer

*"Does this deployed page reference an image that will not display?"*

## What it actually did (the defect, re-measured live 2026-07-31)

It skipped any rendered `/assets/images/<name>.<ext>` whose `<name>` — or the
prefix before its first hyphen — matched an **active asset PURPOSE** for the
site. Purposes are `hero`, `icon`, `logo`. Paths are files. Comparing one to the
other means: **owning one hero asset anywhere (including an S3 URL that is not
served from the site at all) makes every rendered `hero*` path unreportable.**

Measured today over all 127 distinct rendered image paths on 13 live sites, with
HTTP status as ground truth:

| predicate | reports a WORKING image | reports a BROKEN one | SILENT on a broken one |
|---|---|---|---|
| purpose/prefix skip (as shipped) | **21** | 11 | **6** |
| `storage.DeployedWebPath` (this fix) | **1** | **17** | **0** |

The six it was silent about are `/assets/images/hero.jpg` on dartsonline,
gamesdesign, idea.uk, oufe, relojistas and vonc — a broken image on six live
sites that no sweep could ever surface.

## The fix, and why it is a reuse rather than a new predicate

`storage.DeployedWebPath(asset_key, purpose)` **already exists** and is already
the platform's single source of truth for "the web path a generated asset is
committed to and served from". Five writers use it — `plan_sections_action`,
`render_site_components_action`, `emit_sprite_css_action`, `derive_card_asset_action`,
`queryresolve` — and `deploy_image_asset_action` commits to exactly that path via
the shared `storage.AssetKeyFilename`.

So the check becomes the **exact inverse of the render-time resolver**: a finding
is "this page references a path that no active asset of this site would be
deployed to". Nothing new is invented, and the check cannot drift from the
writers without the writers changing first.

## Four changes, in the order they close the door

1. **Predicate: paths, not purposes.** Replace `loadKnownAssetPurposes` with
   `loadDeployedAssetPaths`, built through `storage.DeployedWebPath` (plus
   `storage.BrandHeadAssetPaths` for favicon/og-card, whose published filenames
   are not derivable from the purpose — `bugs_open/142` established that map).
2. **Scan the chrome surface too** (defect 3 in the bug file). `site_components`
   was never scanned, so `head`/`header` references — the ones on *every* page —
   were invisible. Measured: +2 real findings (idea.uk's `favicon.png` and
   `og-card.png`, both 404 on every page), 0 false positives.
3. **Delete the recognised-purpose routing branch.** It is an exact functional
   duplicate of `check_placeholder_image_in_use` — same two paths, same purposes,
   same item types (`needs_hero_image`/`needs_logo`), same handler, same
   precondition, both enabled on the same agent. Neither has ever fired. Removing
   it deletes a duplicate, not a capability, and it is what keeps this fix from
   activating a never-fired fleet-wide auto-regeneration path.
4. **Report an `<img>` with no source.** An empty `src` has no path to reason
   about, so every path-based predicate is blind to it — including this fix. Three
   are live right now (ai-agent-orchestration ×2, finetuning ×1). Detected
   structurally, never probed.

## Decisions and their reasons

- **The check is NOT renamed.** `bugs_open/128` candidate 2 asked for
  `image_asset_unregistered`. The name is looked up in `agent_definitions`
  `default_config` (`design-discovery-agent`'s `run_checks`), so renaming the
  registered check silently disables it until a config migration lands. And with
  the predicate fixed, a finding now genuinely does mean "this path will 404" —
  the name stops being a lie because the check learned to answer it, which is the
  better of the two ways to close a name/behaviour gap.
- **The item key gains the extension** (`image_url_404:logo.png`, was
  `image_url_404:logo`). Two paths that differ only by extension are two
  findings with two different HTTP results — fundamentallyai serves `logo.jpg`
  (200) and `logo.png` (404) today. An extension-blind key lets `idx_swi_dedup`
  silently drop the second, which is `bugs_open/091`'s failure mode. Cost, stated:
  six existing `detected` rows keep their old key and will not dedup against the
  new ones. They are listed in NOTES; they are already unclearable (no handler,
  `bugs_open/083`).
- **No new item type.** Everything this check emits stays `image_url_404`, so no
  addition to the shared work-item vocabulary and no change to
  `verifier_coverage_test`'s classification map.
- **Still DB-only, still no outbound HTTP.** `verifier_coverage_test.go:171`
  records a standing objection to putting an HTTP probe on the completion path.
  This fix keeps that promise: `DeployedWebPath` answers the question structurally.

## How it will be verified

Against the acceptance triple the 07-29 diagnosis specified (restated because two
of the original three URLs had been repaired out from under it), plus the fourth
case the leopardess contribution added. See NOTES for the live re-measurement.
