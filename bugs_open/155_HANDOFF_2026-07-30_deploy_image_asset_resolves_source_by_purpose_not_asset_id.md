# 155 — `deploy_image_asset` resolves the source image by PURPOSE, not `asset_id` — a site with 2+ same-purpose assets silently gets the wrong file

**Filed:** 2026-07-30 by the dartsonline_traffic workstream, while deploying 7
newly-generated icon assets to dartsonline.com.
**Severity:** High. Silent data corruption — `success:true`, no error anywhere, wrong
image served with a green light.
**Status:** OPEN, unowned.
**Diagnosis-loop verification (2026-07-31):** run via `090_TRIGGER_needs_diagnosis_v1.sh`
(`RUN_CORRELATION_ID=0dd9aee4-2982-4b36-9857-0b037c40851e`, item_key
`needs_diagnosis:deploy-image-asset-purpose-not-assetid`) — **CONFIRMED**, iteration 1.
The loop independently re-read `resolveStorageURIFromAsset` and `StoreAssetAction` and
cited the same lines this file does (`uriKey := purpose + "_uri"`; the
`content_data->>$2 ... WHERE a.id = $1` query; `updateContentDataField(purpose+"_uri", ...)`).
**Caveat, stated rather than glossed over:** its "state" evidence tier cites this
session's own retry work-item summaries (e.g. "bypasses purpose-keyed shared-lookup
bug") — text written by the same investigation, not independently discovered. The
code re-read is the genuinely independent part; the state evidence is corroboration,
not a blind replication. Filed per RFC_005 §3.1 (apply the existing diagnosis-loop
discipline to a durable, cross-cutting claim before treating it as settled).

---

## Symptom

Deployed 7 distinct icon assets (all generated 2026-07-29 for dartsonline.com, all
`purpose='icon'`, 7 distinct `asset_id`/`asset_key`/prompts/source images) via
hand-triaged `site_work_items` → `asset-deployer` → `deploy_image_asset`, spec =
`{asset_id, purpose, asset_key, domain}` (no `s3_uri` supplied). 6 of 7 deploys reported
`response.data.success: true` and wrote to 6 **distinct** destination paths
(`/assets/images/icon-<key>.jpg`, one per `asset_key`) — but all 6 downloaded files were
**byte-identical**: sha256 `e647f9fb0dcaa609cafc21c35f718ed03abbaada06a8e1ae5f5ba2f8da22b0a1`,
201615 bytes, 1408×768 (a widescreen HERO-shaped photo, not any of the 6 flat
line-icons that were actually requested).

## Root cause

`platform/orchestration/actions/deploy_image_asset_action.go`,
`resolveStorageURIFromAsset` (lines 300–347). Its Priority-1 branch:

```go
query1 := `
    SELECT s.content_data->>$2
    FROM assets a
    JOIN sites s ON a.site_id = s.id
    WHERE a.id = $1
`
```

resolves the source image for a requested `asset_id` by reading
`sites.content_data->>'{purpose}_uri'` — a single, **site-wide, purpose-keyed** field.
The `WHERE a.id = $1` join only confirms the requested asset belongs to the right site;
the *value actually returned* ignores which specific asset was asked for.

That field is written by `StoreAssetAction`
(`platform/orchestration/actions/v3_site_actions.go:2730`,
`updateContentDataField(ctx, params.DB, *siteID, purpose+"_uri", storageURI, ...)`) —
every time ANY asset of a given purpose is generated, it **overwrites**
`content_data->>'{purpose}_uri'` with that asset's own URI. Last-write-wins, site-wide,
per purpose. Confirmed live: `sites.content_data->>'icon_uri'` for dartsonline.com held
exactly `icon_specialist_range`'s own source object (generated 18:36:47Z, the last of
the 7 icons stored that session) — so every asset_id-only icon deploy on this site,
regardless of which of the 7 it actually named, fetched that one file.

This mechanism can only ever be correct for a site with **at most one** asset of a given
purpose (one `logo`, one `hero`) — presumably the case it was written for — and breaks
silently the moment a site has 2+ same-purpose assets, which is now routine (the
imagery-block prompt guidance explicitly tells the planner to emit N separate icon
entries per section rather than one composite image — "over-decompose rather than
under-decompose").

## Evidence

- 6 deploys (dartsonline.com, `purpose='icon'`, 2026-07-30): all
  `deploy_result.response.data.success = true`, 6 distinct `file_path` values, **all six
  downloads sha256-identical**: `e647f9fb0dcaa609...da22b0a1`.
- `sites.content_data->>'icon_uri'` (dartsonline.com) =
  `s3://personae-prod-uk001-images/images/system/20260729/ec0d186d-12ca-4b61-abad-2b831ff221f3.png`
  — exactly `icon_specialist_range`'s own `storage_path`.
- Bypassing the bug by supplying `spec.s3_uri` directly (a genuine `s3://bucket/key`,
  derived from each asset's own `storage_path`, not its presigned `url` — see
  `bugs_open/152`, that column is separately unreliable) produced 6 further deploys,
  each with a **distinct** sha256, each visually confirmed (read as an image, not just
  hashed) to be the correct, on-prompt flat icon.
- No error, warning, or work-item field anywhere surfaces the mismatch —
  `status='complete'`, `deployed=true`. Matches the platform's own standing pattern: "a
  `complete` work item is not a repaired artefact."
- Separately found on the same site, same session: the asset row for
  `icon_specialist_range` itself has a mismatched source at generation time — its
  `origin_prompt` describes a flat calipers-on-navy icon, but its own stored image is
  the same 1408×768 photorealistic dart-throwing hero scene. That is a **third**,
  independent defect (a generation/storage mix-up, not a deploy-time resolution bug) —
  noted here only because it was found via the same investigation; do not fold it into
  this bug's fix or its verification.

## Fix candidates

1. **Key the `content_data` cache by `asset_key` (or `asset_id`), not bare purpose** —
   e.g. `content_data->'asset_uris'->>asset_key`. Smallest correct fix; requires
   updating the write side (`v3_site_actions.go:2730`) and the read side
   (`deploy_image_asset_action.go:311–327`) together. No migration/backfill needed — the
   field is a transient generation-time cache, always re-derivable from
   `assets.storage_path`.
2. **Drop Priority-1 entirely; always resolve via the asset row itself** (Priority 2,
   the existing `PresignedURLToS3URI` conversion of `assets.url`/`storage_path`, already
   correctly keyed by `asset_id`). Priority 1 appears to exist only to save a query, or
   to serve some caller that has `purpose` but not yet `asset_id` — worth checking
   whether such a caller exists before removing the branch outright.
3. **Require every caller to pass `asset_id`, and make Priority-1 asset_id-aware**
   (keep the cache, scope it) — the least invasive combination of 1+2, only useful if
   some caller genuinely cannot supply `asset_id`.

Any of 1/2 closes the door fleet-wide, not just for this site.

## Related, not duplicate

- **`bugs_open/152`** — `assets.url` is never rewritten off its presigned URL unless
  the deploy call passes `asset_id`; same file (`deploy_image_asset_action.go`),
  different bug: 152 is about a *stale column* read directly by two OTHER call sites
  (`derive_brand_head_assets_action.go`, `derive_card_asset_action.go`); this bug is
  about the *shared, purpose-keyed cache* read by `resolveStorageURIFromAsset` itself,
  inside `deploy_image_asset`. A fix for one does not fix the other.
  **Read this bug before applying 152's fix candidate (1) ("always pass asset_id")** —
  passing `asset_id` is exactly what routes a call into this bug's buggy Priority-1
  branch, so applying 152's fix blindly would make this bug fire *more* often, not less.
- **`bugs_open/142`** — the `undeployed_asset` detector (upstream: decides *whether/when*
  to fire a work item at all). This bug is downstream, in the handler that moves bytes
  once a work item is triaged. Independent; fixing either does not touch the other.

## How to verify a fix

On any site with 2+ active assets sharing one `purpose` (or generate a second
same-purpose asset on any site), deploy each by `asset_id` alone (no `s3_uri` in spec)
and confirm the resulting files are **distinct** (sha256) and each visually matches its
own `origin_prompt`/`asset_key` — not just that each reports `success:true`.

## Contained workaround used this session (not a fix for the class)

For dartsonline.com's 6 affected icons, re-deployed each with an explicit, correctly
formatted `s3_uri` in spec (bypasses `resolveStorageURIFromAsset` entirely) — see
`docs/agent_docs/docs024_key_docs_latest/dartsonline_traffic/SQL_2026-07-30w_retry_six_icons_correct_s3_uri.sql`.
This does not touch the shared function; the next multi-same-purpose-asset site to go
through the normal asset_id-only path will hit the same bug.
