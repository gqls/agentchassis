# 209 — `deploy_image_asset` still resolves a source by PURPOSE from `collected_data`, and that lookup runs FIRST

**Filed:** 2026-08-06 by the `bugfix_152_155_asset_source_identity` lane, while running
`bugs_open/155`'s closure test. **This is 155's second arm.** 155 closed on its own
recipe (proof below is in `bugs_closed/155`); this file exists because closing it
without naming what survived would have retired the class on one arm's evidence.
**Status:** OPEN, unowned. **Severity:** medium — same wrong-bytes outcome as 155, but
only inside a single build workflow, and no live instance is yet demonstrated.

## The defect

`findStorageURI` (`platform/orchestration/actions/deploy_image_asset_action.go`)
**Priority 2**:

```go
if uri := datahelpers.ExtractNestedFieldString(collectedData, purpose+"_uri"); uri != "" {
    return uri
}
```

A top-level `{purpose}_uri` in `collected_data` — a **purpose-keyed, last-write-wins**
value used as a per-asset source. Identical in shape to the `sites.content_data`
cache that 155 was filed for and that is now deleted, one layer up.

Two writers seed it in-run:
- `v3_site_actions.go:2810` — `params.CollectedData[purpose+"_uri"] = storageURI` (StoreAssetAction)
- `generate_image_actions.go:994` — same key (legacy StoreImageReference)

**It is consulted BEFORE the asset_id path**, so where both are available it wins:
`findStorageURI` runs first, and only if it returns "" does the action fall through to
`resolveStorageURIFromAsset` (the arm fixed in `1d11827c1`).

## Why it survived 155's fix, stated plainly

It was **kept deliberately** and the reason is in that lane's PLAN: the legacy
pageflow deploy step reads this key within the same workflow, and the DB-side cache
(the arm 155 named) had a different, provably-unused reader. What the lane got wrong
was the SCOPE of the claim it wrapped around that decision — "the wrong-bytes state
becomes unrepresentable" — which is true of the DB arm and not of this one. That
overclaim is logged in `WRONG_CALLS.md` (2026-08-06).

## What is NOT yet established — read this before fixing

- `[UNMEASURED]` Whether any live workflow actually stores 2+ same-purpose assets and
  then deploys them **in one run**, which is the only shape that triggers this. A
  single-asset-per-run workflow is correct today and always has been.
- `[UNRECOVERABLE]` Whether this arm produced 155's founding symptom (6 identical
  icons, 2026-07-30). It is the better candidate — `asset-deployer` could not pass
  `asset_id` at all until migration 324 today, so 155's own arm was unreachable
  through that agent — but terminal `orchestration_states` rows are reaped at ~24h,
  so the deciding `collected_data` is gone. **Do not assert it.**
- The obvious suspect to measure first is `image-build-handler`, which stores and
  deploys imagery in one orchestration. Read its workflow before assuming.

## Fix candidates, ordered by what closes the door

1. **Delete Priority 2 and make the asset_id path the only DB-free route**, now that
   `asset-deployer` passes `asset_id` (migration 324) and every deploy has an asset
   row. Requires checking the legacy `pageflow-builder` / `site-work-orchestrator`
   steps, which have NO `input_fields` and so reach every spec field by the recursive
   Strategy 2 — they may already supply `asset_id` without anyone noticing.
2. **Key the in-run value by asset, not purpose** (`asset_uris.<asset_key>`), matching
   what the row-side fix did. Smaller blast radius; keeps a same-run fast path.
3. **Make Priority 2 conditional on there being exactly one same-purpose asset in the
   run.** Weakest — it is a guard against a state the code can still express.

## How to verify a fix

Not a hash comparison at the deploy — this arm needs a **single workflow** that stores
two same-purpose assets and deploys both. Assert the two committed files differ, and
assert it on the workflow that really does it, not on a hand-built one: a synthetic
two-store-one-deploy run proves the branch, not the exposure.

## Related

- `bugs_closed/155` — the same defect, DB-side arm, fixed and behaviourally proven.
- `bugs_open/152` — the source-recording half; `storage.AssetSourceRef` (IMG-068) is
  the shared derivation both arms should end up resolving through.
- LANDMINES: the 155 entry, retired 2026-08-06 — **it covers the DB arm only.**
