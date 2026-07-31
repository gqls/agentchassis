# PLAN — `bugs_open/142`: the undeployed_asset detector cannot see a missing artefact

**Started 2026-07-31.** Bug picked from `bugs_open/` after an ownership sweep
(see NOTES §1): one filing commit on 07-29, nobody since, and no live session
holds it.

---

## What the bug is

`platform/orchestration/actions/discovery_checks/check_undeployed_assets.go`
(`findUndeployedAssets`) answers "which of this site's assets were generated but
never deployed?" It gets both halves wrong for the two **brand-head** artefacts
(`favicon`, `og_card`):

1. **The population is the `assets` table**, so a site whose artefact was never
   generated has no row and is structurally unexaminable. A detector whose
   denominator is the artefact table cannot see a *missing* artefact.
2. **The deploy evidence is `page_components.rendered_html`**, but `favicon` and
   `og_card` are referenced from the site HEAD (`injectBrandHeadTags` →
   `site_components`, slot `head`). A present, serving artefact can therefore
   never look deployed, and fires for ever.

## Verified live, 2026-07-31 (not carried forward from the bug file)

Every figure below was re-measured today; the bug file's own numbers had moved
(NOTES §2 records the correction).

| claim | evidence |
|---|---|
| the shipping predicate fires on **96 assets across all 14 sites** | ran the exact Go query as SQL; `favicon`+`og_card` on 12 of 14 |
| those are **false** positives | `og-card.png` + `favicon.png` both **200** on the wire for fundamentallyai, gamesdesign, vonc, dartsonline |
| the check is **blind** to real absence | idea.uk (61 deployed components) and webdesign.co.uk (6) have **no** og_card/favicon/logo asset row and serve **404** for both files |
| idea.uk *advertises* the missing card | 13 of 14 `site_components` heads contain `/assets/images/og-card.png`; idea.uk is one of them, and it 404s |
| the refs live in the head, not in pages | `site_components` slot `head`: 13/14 carry both refs; `page_components`: none do |

## Decisions, and why

**D1 — Fix the brand-head half only; leave the page-asset half alone.**
`icon`/`hero`/`illustration`/`logo` findings are doing real work (dartsonline's
18 icon items on 07-30 were genuine and were serviced). The defect is specific
to artefacts whose reference lives in the chrome.

**D2 — DO NOT "fix" the `LIKE` underscore wildcard.** `purpose` values contain
`_` (`og_card`, `content_hero`, `sprite_sheet`, `hero_home`) and SQL `LIKE`
treats `_` as *any character*, so `content_hero` matches the real filename
`content-hero…`. Measured: **38 of 38 `content_hero` assets read as deployed
unescaped and 0 of 38 escaped.** Escaping the wildcard — the obvious "correct"
fix — would manufacture 38 false findings. The accident is load-bearing.
Recorded as a landmine and pinned by a test.

**D3 — The population for brand-head artefacts is SITES, not assets.** This is
the bug file's fix candidate 1 and the only shape that can represent absence.
Emits `needs_brand_head_assets`, an item type that is **already classified**
(`verifier_coverage_test.go:253`, `catCreation`) and already in `liveItemTypes`,
so it carries no new coverage obligation. Handler `asset-deployer` is live and
is the agent that owns `derive_brand_head_assets`.

**D4 — Deploy evidence for brand-head artefacts is the site head.**
`site_components.rendered_html` where `slot_name='head'`.
⚠ **`site_components.build_status` is `'rendered'`, never `'deployed'`** (all 42
rows, 3 slots × 14 sites). Mirroring the page-component predicate's
`build_status='deployed'` here would make the new check permanently blind —
recorded as a landmine.

**D5 — One declaration of the artefact paths, pinned by a source scan.** The
strings `/assets/images/favicon.png` and `/assets/images/og-card.png` are spelled
as literals in four places in `derive_brand_head_assets_action.go` and again in
`injectBrandHeadTags`. The check needs the same strings, and a sixth hand-copy is
the drift class this repo keeps paying for. So they go in
`platform/storage/url_helpers.go` (`BrandHeadAssetPaths`) beside `ImagePurposes`,
which already documents these two filenames in prose.

> **D5a — the deriver is NOT edited, deliberately.** Session `759437b9` is
> actively working `bugs_open/143` inside
> `derive_brand_head_assets_action.go` (49 transcript hits at 18:48). Two
> sessions editing one file is the one collision no hook can prevent. Instead
> `TestBrandHeadAssetPathsMatchTheDeriver` **scans that file's source** and fails
> if any `/assets/images/*.png` literal it writes is absent from the map. That is
> strictly better than an edit would have been: the deriver keeps its literals,
> and drift becomes a build failure rather than a silent divergence. Adopting the
> map in the deriver is left for its own lane.

**D6 — Findings stay at `status='detected'`.** `bugs_open/083`'s landmine is
explicit: do not "fix" the parked queue by having detectors write `triaged`.
This fix makes the detector correct; it does not make the queue drain. Stated
rather than hidden — see NOTES §5.

## Phasing

1. ✅ Re-verify the bug against live production (done — table above)
2. ✅ File the diagnosis loop (owner ruling 2026-07-31) — intake
   `65ccb69d-120e-41a0-9b09-603bfe37ecd0`, run `fafbb1bf-8a9e-40d1-b47f-aa5e1721f071`
3. Code: `BrandHeadAssetPaths` + rewrite the check's brand-head half
4. Tests: the path-drift sensor, the wildcard pin, the absence case
5. Council gate, commit, build
6. Verify on the pod, then `bugs_open` → `bugs_closed`
