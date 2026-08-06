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

---

## Progress 2026-08-06 — fix candidate 1+2 combined, committed (with 152's, one change)

Taken up together with `bugs_open/152` — `bugs_closed/179` had already named the
shared root ("identity reconstructed rather than read") and asked for a joint
design. Re-verified at HEAD first: `resolveStorageURIFromAsset`'s Priority-1
purpose-cache read was unchanged, and the failure population has GROWN — multi
same-purpose assets are now the norm (census 2026-08-06: robot-hands 23 active
`hero` rows, dartsonline 20 `icon`, fundamentallyai 15 `icon`, 12 (site,purpose)
groups >1). dartsonline's `content_data->>'icon_uri'` was last overwritten
2026-08-05, so the wrong-bytes branch was still live and armed.

**What shipped:** Priority-1 is DELETED, not scoped (this file's candidate 2 —
and the "does any caller have purpose but not asset_id?" question is answered:
`resolveStorageURIFromAsset` is only ever reached WITH an asset_id, so the cache
had no legitimate consumer). Resolution is now one query of the asset row itself
(`storage_path`, `url`) through the new shared `storage.AssetSourceRef`
(register IMG-068). The write side (`StoreAssetAction`) stops writing the
`{purpose}_uri` DB cache entirely — measured three ways that its ONLY reader was
the deleted branch (Go grep; regex over live active `agent_definitions` configs,
positive-controlled; grep over `sql_for_agents/` + core-manager) — and instead
records `storage_path` (+ provider) on the row at generation, which is candidate
1's cache-keying fixed by removing the cache rather than re-keying it. The
in-run `collected_data[{purpose}_uri]` copy stays (legacy pageflow reads it
same-workflow; `data_helpers.go:1199` is a strict path lookup, so it cannot leak
across runs). `sql_for_agents/323` backfills `storage_path` on the 205
presigned-only rows so the fallback parse is never load-bearing for long.

**Still to close** (this file's own verify recipe): after the roll, deploy 2+
same-purpose assets by `asset_id` alone and confirm DISTINCT sha256s each
matching its own `origin_prompt` — the dartsonline founding icons are the
natural re-run. Pod-grep first: `AssetSourceRef` ≥1 AND
`Resolved s3_uri from site content_data via asset_id` = 0, both replicas. Then
correct the LANDMINES entry ("resolves by PURPOSE") visibly, dated. The third
independent defect noted in Evidence (generation/storage mix-up on
`icon_specialist_range`) remains UNTOUCHED and open, deliberately.

**Council APPROVED round 1** (`c055840a-9edc-4f9a-8a4a-b23ac4cad02a`, 8 advisory,
none high; trailer upgraded to `Council-Reviewed:` on `bb53326a8` after reading
it). The objection worth carrying into this file: my "the purpose cache has no
other reader" census was Go-grep + live-config regex + repo grep, and the council
was right that none of those can see a queue-built or external caller. Re-run
wider, it found one — `scripts/initial_messages/180_adoption/081c_direct_asset_
deployer.sh` prints `content_data->>'hero_uri' / 'hero_about_uri' / 'logo_uri'`
as the operator's crib for the `s3_uri` to paste into a hand deploy. **That is
this bug in manual form**: one URI per purpose, so on a multi-asset site it hands
you the wrong file to deploy, and after this fix ships it is stale as well.
Rewritten to select `storage_path`/`url` per `asset_key` from `assets`. The other
hit was this bug's own 090 diagnosis row — not a consumer.

---

## LIVE 2026-08-06 on chassis `v1.0.1259` — the buggy branch is GONE from the binary

Pod-verified on **both** replicas (`agent-chassis-5cf5db5bd8-54xsx`, `-ldx5z`),
four controls in one exec, exactly the retirement test this lane wrote down:

```
AssetSourceRef                                        2   POSITIVE  (the fix is in)
"Resolved s3_uri from site content_data via asset_id" 0   NEGATIVE  (the purpose-cache
                                                                     branch is GONE)
"Resolved source object from asset row"               1   the replacement, present
"AssetSourceRefZZZnotreal"                            0   nonsense control (the grep
                                                                     discriminates)
```

The negative is the one that matters: that string is the deleted branch's own log
line, so its absence is evidence about *this* binary rather than about the tag or
the roll (`bugs_open/153`). Pod start 10:50:29Z, fix commit `1d11827c1` 10:03:46Z
— the build postdates the commit, which is the precondition, not the proof.

**The founding case, re-measured** (dartsonline.com, 20 active `icon` assets):

| resolution path | distinct sources across the 20 icons |
|---|---|
| OLD — `sites.content_data->>'icon_uri'` | **1** |
| NEW — each row's own `storage_path`/`url` | **20** |

That is this bug stated as arithmetic: one input for twenty deploys is exactly why
six of them produced byte-identical files. `[INDUCTION, not the behavioural proof]`
— it measures the INPUTS the live function now receives, and the function itself is
unit-tested (incl. two mutation proofs) and confirmed present in the running binary.
It is not a dispatch.

**What is still owed to CLOSE this file**: one real deploy of 2+ same-purpose assets
by `asset_id` alone (no `s3_uri` in spec), then `sha256sum` the resulting files and
confirm they differ, opening at least one against its own `origin_prompt`. Nothing
short of that is the recipe this file wrote, and `success:true` plus distinct
destination paths were both already true while the bug was shipping identical bytes.

---

## CORRECTION 2026-08-06 — the fix closed ONE arm of two, and the closure test found the other

**Do not close this bug on the v1.0.1259 proof above.** Trying to run this file's own
closure test is what exposed it.

Three `asset_id`-only deploys were dispatched at dartsonline icons (correlations
`380f5cb4`, `b7fd8a68`, `8bdd7c90`; all three orchestrations COMPLETED). **All three
returned `{"deployed": false, "skipped": true, "reason": "no storage URI found for
icon"}`** — a correct, loud skip, but not the test. `asset_id` was present in the
child's `input_data` (verified in `collected_data`) and the action still never saw it:

- the live `asset-deployer` `deploy_asset` step declares
  `input_fields: ["s3_uri","purpose","domain","asset_key"]` — **no `asset_id`**, and
  no `"asset_id": "<path>"` config key either;
- `ExtractActionInputs` **Strategy 1 wins whenever `input_fields` is present** and
  extracts *only* the listed names (`action_inputs.go:441-455`); the recursive
  all-fields hunt is Strategy 2, reached only when `input_fields` is absent.

So `deploy_image_asset` **cannot receive an `asset_id` through `asset-deployer`**, and
that config is unchanged since **2026-02-20** (single row, no snapshots) — i.e. it was
already true on 2026-07-30. Every live `deploy_image_asset` step, fleet-wide:

| agent | step | input_fields | can pass asset_id? |
|---|---|---|---|
| asset-deployer | deploy_asset | `["s3_uri","purpose","domain","asset_key"]` | **no** (Strategy 1) |
| pageflow-builder | deploy_hero_image / deploy_logo_image | *(none)* | yes (Strategy 2) |
| site-work-orchestrator | deploy_hero_image / deploy_logo_image | *(none)* | yes (Strategy 2) |

**What this means for this file's original diagnosis.** The 6 identical icons were
deployed through `asset-deployer`, which cannot reach `resolveStorageURIFromAsset` at
all — so the Priority-1 purpose-cache branch, though genuinely defective and now
deleted, is unlikely to be what produced the observed symptom. The surviving
candidate, in the same function and NOT removed, is `findStorageURI` **Priority 2**:

```go
// deploy_image_asset_action.go, findStorageURI
if uri := datahelpers.ExtractNestedFieldString(collectedData, purpose+"_uri"); uri != "" { … }
```

a top-level `{purpose}_uri` in `collected_data`, written in-run by
`StoreAssetAction` (`v3_site_actions.go:2810`) and `generate_image_actions.go:994`.
**Same defect, same last-write-wins shape, one layer up, and it is consulted BEFORE
the asset_id path.** I kept it deliberately for the legacy pageflow flow and said so
— but I described the wrong-bytes state as "unrepresentable", and for this arm it is
not. `[UNRECOVERABLE]` whether the July runs carried that key: terminal
`orchestration_states` rows are reaped at ~24h.

**Revised closure requirements:**

1. Decide the in-run arm. Either narrow `findStorageURI` Priority 2 (it can only be
   correct when exactly one asset of that purpose is stored per run) or make the
   asset_id path reachable and preferred.
2. **Add `asset_id` to `asset-deployer`'s `deploy_asset` `input_fields`** — a live
   config change, no image needed. Note this also means IMG-053's post-deploy
   `UPDATE assets SET url = …` (gated on `inputs.Get("asset_id")`) has never fired
   through this agent either, which is worth checking against `bugs_open/152`'s
   107-flipped-row population: something else flipped them.
3. Only then re-run this file's sha256-differ test.

**What the v1.0.1259 proof DOES establish, unchanged:** the DB-side purpose cache is
gone from the binary and from the write path, all four readers resolve per-asset, and
migration 323 made 205 rows durable. That is real and it stands. It is one arm.
