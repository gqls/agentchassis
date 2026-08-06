# PLAN 2026-08-06 — asset source identity is READ from the row, never reconstructed (bugs 152 + 155)

**Lane:** bugfix_152_155_asset_source_identity. **Bugs:** `bugs_open/152` (assets.url
never rewritten off its presigned URL — now morphed, see below) and `bugs_open/155`
(deploy resolves the source by PURPOSE, not asset_id — diagnosis-loop **CONFIRMED**
2026-07-31, correlation `0dd9aee4`). `bugs_closed/179` explicitly handed this seam
over: *"the same 'identity reconstructed rather than read' root … should probably be
designed with them rather than separately."*

## The claim, and how it is verified (owner ruling 2026-07-31 declaration)

The cross-cutting mechanism of 155 was through the diagnosis loop already
(CONFIRMED, iteration 1, cited the same lines). The **additional** structural claim
this plan rests on — *the deploy-time url flip (IMG-053 Edit F) strands every
`assets.url` reader, because none of the four readers consults `storage_path`* — is
filed on **first-hand verification substituting for a 090 run, declared per the
2026-07-31 ruling**: all four readers read at HEAD with line citations (below), the
write side read at HEAD, and a live census whose result could have come out
otherwise (it showed 49 rows already stranded and 5 active logo rows primed to fail
the next brand-head derivation). No open work item or live session touches these
files (checked 2026-08-06: transcripts of all 12 active sessions, `who-owns.py`,
`site_work_items` queue).

## What is true at HEAD (2026-08-06, all cited)

Four readers reconstruct a generated asset's **source object** identity instead of
reading it from the asset row:

1. `deploy_image_asset_action.go:406-453` `resolveStorageURIFromAsset` —
   Priority-1 reads `sites.content_data->>'{purpose}_uri'`, a **site-wide,
   last-write-wins cache** keyed by purpose alone (written by
   `v3_site_actions.go:2747` and `generate_image_actions.go:976`). Wrong bytes for
   every asset except the last-stored of its purpose. Live: robot-hands has **23**
   active `hero` assets, dartsonline **20** `icon` — multi-asset purposes are the
   norm now. Priority-2 parses `assets.url` as a presigned URL.
2. `derive_brand_head_assets_action.go:94,135` — reads `a.url`, parses presigned →
   errors on any other form.
3. `derive_card_asset_action.go:165,290,301` — same, with an error message that
   *names* `storage_path` and still doesn't read it.
4. `imagery_style_guide.go:431-450` `resolveReferenceAssetURIs` — reads `url` only,
   and **skips silently** when it doesn't parse: style anchors quietly vanish.

Meanwhile the deploy path itself **flips `url` to a local web path** on every
asset_id deploy (`deploy_image_asset_action.go:357-381`, IMG-053 "Edit F"),
preserving the source into `storage_path` — which nothing reads back.

Live census 2026-08-06 (`assets`): 205 presigned-only rows (readable by parse);
**107 flipped rows with `storage_path`** (source recoverable, no reader looks);
**49 flipped rows with no `storage_path`** (stranded); 2 presigned+storage_path.
`storage_path` forms: 104 https-object-URL, 1 bare key, 4 local (derived
artefacts). Five sites' **active logo rows** carry a non-presigned `url` (four are
the unrendered template literal `/assets/images/input-data.asset-key.jpg`) — the
next `derive_brand_head_assets` on webdesign.co.uk / gaswholesalers.com /
finetuning.uk / vetcomparison.uk / leopardessconsulting.co.uk fails at
`:136 "could not derive storage key from logo url"`; four of the five are
recoverable via their populated `storage_path` today.

The DB-side `{purpose}_uri` cache has **exactly one reader** — the buggy
Priority-1. Measured three ways, each disconfirmable: Go grep for `_uri` key
construction (only the writers + Priority-1), regex over live active
`agent_definitions` configs (`~ '(hero|logo|icon|content_hero)_uri'` → 0 rows;
positive control: `%_uri%` finds 7 agents, all `s3_uri`/`image_uri`/`dataset_uri`),
grep over `sql_for_agents/` + `internal/core-manager/` (0). The in-run
`collected_data[{purpose}_uri]` copy has a real reader (`findStorageURI`
Priority-2, legacy pageflow) and is a **strict top-level path lookup**
(`data_helpers.go:1199-1234` — dot-path with `.response` unwrap only, not
recursive), so it cannot accidentally see the DB value through a loaded
site_record.

## The fix — one derivation for "which stored object is this row's source"

Mirrors 168's move for the *destination* path (one shared derivation in
`platform/storage`, writer and readers resolve identically) on the *source* side.

1. **`platform/storage/url_helpers.go` — `AssetSourceRef(storagePath, url string) string`.**
   Returns an `s3://bucket/key` URI or a bare object key; never an https URL,
   never a local path; `""` when the row identifies no object. Order:
   `storage_path` as s3:// → as-is; as https object URL → converted; as bare key
   (no scheme, no leading `/`, contains `/`) → as-is; else `url` as
   presigned/https → converted, as s3:// → as-is; else `""`.
   Contract notes: bucket is dropped downstream by `ExtractKeyFromS3URI`
   exactly as every current path does (single-bucket estate — unchanged
   behaviour); a local `/assets/…` path is rejected here so the
   `ExtractKeyFromS3URI` pass-through trap cannot fire.
2. **Tests** — table tests over the five live row shapes + the traps (local path,
   template literal, empty both, s3-in-url).
3. **`resolveStorageURIFromAsset`** — one query
   `SELECT storage_path, url FROM assets WHERE id=$1` → `AssetSourceRef`.
   Priority-1 (purpose cache) and the url-only parse are **deleted** — the
   wrong-bytes state becomes unrepresentable; an unresolvable row skips loudly
   with the existing `{deployed:false, skipped:true, reason}` shape. The local
   `presignedURLToS3URI` duplicate is deleted once its last caller moves (edit 6).
4. **`derive_brand_head_assets_action.go`** — select `a.storage_path` too, resolve
   `storage.ExtractKeyFromS3URI(storage.AssetSourceRef(sp, url))`, keep fail-loud.
5. **`derive_card_asset_action.go`** — `findCardSourceHero` returns storage_path;
   same resolution at `:165`.
6. **`imagery_style_guide.go`** — `resolveReferenceAssetURIs` selects
   storage_path+url, uses `AssetSourceRef`, keeps only `IsS3URI` results (the
   adapter's ReferenceFetcher needs a full URI; a bare key stays skipped, as
   today).
7. **`v3_site_actions.go` StoreAssetAction** — hoist the existing storageURI
   resolution above the INSERT; write `storage_path` (s3:// form) +
   `storage_provider` at birth, in the upsert (`COALESCE(EXCLUDED.storage_path,
   assets.storage_path)`) and the fallback insert; **delete the
   `updateContentDataField(purpose+"_uri", …)` DB write** (sole reader deleted in
   edit 3; keep the `collected_data` copy for the legacy in-run path; the `_url`
   sibling untouched — templates read it).
8. **`docs/agent_docs/sql_for_agents/321_backfill_assets_storage_path.sql`** —
   fleet-wide mirror of idea.uk's `w9_04`:
   `SET storage_path = COALESCE(storage_path, split_part(url,'?',1))` where `url`
   is presigned — so the remaining 205 rows become durable **before** their url
   ever flips or is hand-repaired. Live-immediately, write-only, no reader
   depends on ordering vs the image.

## Deliberately NOT done

- No lock guard on deploy's `UPDATE assets` (LANDMINES `deploy_image_asset has no
  asset-lock check anywhere and must NOT be given one` — publication, not
  replacement).
- No change to `generate_image_actions.go`'s legacy StoreImageReference path (no
  assets row, collected_data-only; pageflow sites age out per IMG register).
- No repair of the 49 stranded rows' `storage_path` (their source is genuinely
  unrecorded; 45 are derived artefacts whose source is another row; fail-loud is
  the honest state, and the 4 rescuable logos resolve via storage_path already).
- No `AssetPaths` construction anywhere (`TestAssetPathsAreOnlyConstructedInStorage`).

## Consumers, named (owner ruling 2026-07-29 §3)

- **image-build-handler / asset-deployer** (asset_id deploys): guarantee changes
  from "success may deploy another asset's bytes" to "the named asset's bytes or a
  loud skip".
- **131 og-card lane / brand-head derivation**: derivation stops depending on the
  logo's url never having been deployed or hand-repaired.
- **imagery style anchors (IMG-0xx reference images)**: anchors survive the
  referenced asset's deployment instead of silently dropping.
- **pageflow-builder (legacy)**: untouched — collected_data resolution only.
- **Ops habit change**: `spec.s3_uri` hand-derivation (the 155 workaround) becomes
  unnecessary; asset_id alone is now correct.

## Verify

- Unit: all five live row shapes resolve/fail as specified.
- Post-roll pod-grep: `AssetSourceRef` ≥1 (positive) AND
  `"Resolved s3_uri from site content_data via asset_id"` = 0 (negative — the
  removed Priority-1 log line), same exec, both replicas.
- Induce 155's own recipe: two same-purpose assets deployed by asset_id alone →
  sha256s differ. (The founding dartsonline icons are re-runnable candidates.)
- 152's recipe: after any fresh deploy, `assets.url` is local AND `storage_path`
  identifies the source; then a derivation against that row succeeds.
