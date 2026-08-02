# NOTES — bugfix 168, deployed asset path

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-02 — taking the bug, and checking it was still real

Picked 168 after checking ownership two ways: `scripts/who-owns.py 168` (only the filing
commit touches it) **and** a grep of every `.jsonl` session transcript modified in the last
4 hours for `bugs_open/NNN` references, because `who-owns` reads COMMITS and is blind to a
session mid-fix. 27 sessions were live; the bugs they were holding were 010, 018, 029, 043,
064, 081, 083, 092, 095, 099, 117, 119, 120, 129, 136, 137, 138, 139, 149, 150, 151, 154,
156, 157, 159, 162, 164, 165, 169, 171–175. 168 appeared in none of them.

Bug still valid: `platform/storage/url_helpers.go` unchanged since `d671fb2b2`, and
`DeployedWebPath` still returns `/assets/images/og_card.png`.

## MISSTEP 1 — I nearly inherited the bug file's mechanism without checking it

The filed root cause says an underscore purpose stored with `asset_key = purpose` "yields a
path with an underscore where the deployed file has a hyphen". I was one edit away from
implementing fix candidate 2 (swap `_`→`-` unconditionally) on that basis.

Then I read the writer. `deploy_image_asset_action.go:185` branches on the **identical**
condition:

```go
if assetKey != "" && assetKey != purpose {   // deployer
if assetKey == "" || assetKey == purpose {   // helper (the skip)
```

So for a deployer-published file the deployed name **also** has the underscore. Helper and
deployer agree; candidate 2 would have *broken* that agreement for every future underscore
purpose — introducing the drift it claims to remove. The check that caught it was reading
`DownloadOptimizeAndPrepare` (it returns `BuildAssetPaths(purpose, ext)` verbatim), which
took two minutes.

**Cheap check that generalises:** when a bug file names a mismatch between a reader and a
writer, read the WRITER before believing the direction of the mismatch. A helper documented
as "mirroring" something is a claim about behaviour, not the behaviour.

## MISSTEP 2 — a live-impact hypothesis I talked myself into, then refuted

The queue showed `undeployed_asset: Asset 'og_card' generated but not deployed to site`,
four of them `unresolved after 2 attempts`. I formed a satisfying theory: the repair route
is `asset-deployer` → `deploy_image_asset` → writes `og_card.png` → the head references
`og-card.png` → the 404 persists → hence four permanent unresolved items. That would have
made 168 live-biting rather than latent, and I would have written it up that way.

It is wrong. `check_undeployed_assets.go:256` excludes brand-head purposes from the generic
half — `AND NOT (COALESCE(a.purpose,'') = ANY($2::text[]))` — and routes them to
`needs_brand_head_assets`, whose repair is **re-derivation**, not deployment. So brand-head
never reaches the deployer. The unresolved rows predate the 142 lane's fix.

**What caught it:** reading the SQL of the check I was about to accuse, instead of reasoning
from the item's summary text. The summary said "not deployed to site"; the predicate said
which rows can produce that summary, and og_card is not among them any more.

## The diagnosis loop: REFUTED — and it was right about the thing that mattered

Filed `090` before asserting a structural cause, per the owner ruling of 2026-07-31.
Intake `62ab1470-2edc-46ce-a480-8deea38e4ed0`, run corr
`ae9404bd-dab7-4606-ade3-c439ebda93af`. Verdict **REFUTED**, 2 iterations.

Its correct and useful finding: `injectBrandHeadTags` **hardcodes** `/assets/images/
favicon.png` and `/assets/images/og-card.png` and never calls `DeployedWebPath` or reads
`assets.url`. So there is no render-time failure today — which corroborates the bug file's
own "Low today, latent" severity, independently of me. This is why the fix is framed as
removing a drift mechanism, not as repairing live breakage, and why the code comment says
"corrects a latent disagreement rather than moving live traffic".

**Where the loop itself was wrong, recorded rather than accepted:** it asserted
*"DeployedWebPath's only found call site (queryresolve.go's webPath)"*. There are **six**
(grep below). One of them — `check_image_url_404` — is exactly where the two derivations
DO meet, which is why the 128 lane had to add an `IsBrandHeadPurpose` branch there to avoid
reporting a 404 for the og card and favicon of every site in the fleet. Its refutation rests
on an incomplete census, so I did not treat REFUTED as "no defect here".

```
plan_sections_action.go:304,328,355,393,423   render_site_components_action.go:415
emit_sprite_css_action.go:136                 derive_card_asset_action.go:204
queryresolve/queryresolve.go:294,297          discovery_checks/check_image_url_404.go:426
```

**Genuine side-finding from the loop, not mine to fix:** several active `favicon`/`og_card`
rows carry `assets.url = '/assets/images/input-data.asset-key.jpg'` — an unresolved template
literal run through the `_`→`-` swap. Already documented in `check_undeployed_assets.go` and
owned by `bugs_open/152`. Left alone.

## Measurements, with the queries inline

Fleet census of the skip branch, re-run 2026-08-02 (the bug file measured it 2026-07-31 —
grounding a figure rather than carrying it forward):

```sql
SELECT purpose, COALESCE(asset_key,'<null>') AS asset_key, count(*) AS rows
  FROM assets WHERE status='active'
   AND (asset_key IS NULL OR asset_key='' OR asset_key=purpose)
 GROUP BY 1,2 ORDER BY 3 DESC;
```
```
og_card | og_card | 12      hero | hero | 5
favicon | favicon | 12      logo | logo | 4
```

Identical to 07-31. 267 active rows total; every other underscore purpose
(`content_hero` ×31, `sprite_sheet` ×1) carries a distinct key and takes the swap correctly.
So the risk set outside brand-head is empty **structurally**, not by luck.

Also measured, and it surprised me: **all 24 brand-head rows store a site-relative web path
in `assets.url`**, not a presigned S3 URI — `recordDerivedAsset` writes the path it just
committed. `url` is therefore polymorphic across writers (presigned S3 for generated assets,
web path for derived brand-head ones). Noted, not relied on; the helper's doc comment already
warns callers off `assets.url`.

## Guards proven by mutation, because a green run proves nothing

Three mutations, three confirmed failures, each reverted immediately:

1. Removed the brand-head branch from `DeployedAssetPath` →
   `TestDeployedWebPathExpressesBrandHeadPaths` fails: `= "/assets/images/og_card.png",
   want "/assets/images/og-card.png"`.
2. Broke `RelativeURL` = `"/"+FilePath` in `assetPathsForFilename` →
   `TestDeployedAssetPathFormsAreConsistent` fails on five inputs, plus the two older tests.
3. Made the deployer re-implement the derivation (`storage.BuildAssetPaths` + a reference to
   `AssetKeyFilename`) → all three arms of the source sensor fire, including the
   anti-vacuity arm.

The tautology I deliberately did **not** write: "the deployer's path equals the renderer's
path". Both now call one function, so that assertion cannot fail, and it would keep passing
if someone reintroduced a private copy behind a condition the test does not exercise. What
broke before was a STRUCTURE — two implementations held together by prose — so the sensor
reads the source. That is stated in the test itself, so nobody deletes it as "not a real
test".
