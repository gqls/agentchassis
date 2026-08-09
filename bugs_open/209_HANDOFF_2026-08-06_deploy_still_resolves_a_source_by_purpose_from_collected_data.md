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

---

# VERIFICATION 2026-08-08 — `bugfix_209_deploy_purpose_keyed_source` lane

Picked up unowned (the `who-owns.py` hit naming `bugfix_221` is a false positive —
that lane's handoff merely *cites* this file). Full evidence and every misstep:
`docs/agent_docs/docs024_key_docs_latest/bugfix_209_deploy_purpose_keyed_source/`.
No fix written — three findings changed the plan, and two of them change *this file*.

**Method note (per the 2026-07-31 owner ruling):** `090` was not run. Substituted
first-hand verification, stated plainly: a census of **every** live agent definition
carrying the action, branch-level reads of the routing that decides exclusivity,
key enumeration over the live `collected_data` of all 18 recent runs, and the real
Go helpers executed 400× under test. The claims below are measurements, not reads.

## 1. `[MEASURED]` The defect is LATENT — no live workflow can reach it

Upgrades this file's `[UNMEASURED]` to a measured negative. Exactly **three** live
definitions carry `deploy_image_asset`: `asset-deployer`, `pageflow-builder`,
`site-work-orchestrator`. None can produce the two-same-purpose-assets-in-one-run
shape:

- `pageflow-builder` / `site-work-orchestrator` store+deploy **hero and logo** —
  different purposes, so `hero_uri` and `logo_uri` are distinct slots.
- `image-build-handler` (this file's named suspect) **does** have two `store_asset`
  steps both at static `purpose: "hero"` — but they sit on **mutually exclusive
  branches** (`check_item_type_imagery` → `check_imagery_brand_update` vs
  `check_item_type`), so one runs per orchestration. It has **no deploy step**: it
  delegates via `call_asset_deployer` with `s3_uri: asset_stored.image_uri`, i.e.
  the just-stored asset carried **by identity**.
- `asset-deployer`'s `collected_data` holds **no `{purpose}_uri` key at all**
  (enumerated across 18 runs), and `ExtractNestedField` is a strict path walk, not a
  recursive search — so Priority 2 is unreachable there even when the store fails
  and the deliberate `error_step` still deploys. That path degrades to a safe skip.

**Latent, not closed.** Real defect, no reachable instance; the door is open for the
next workflow that stores two same-purpose assets and deploys in-run.

## 2. `[MEASURED]` Fix candidate 1 would CAUSE the failure it prevents — do not take it

Both legacy deploy steps carry **no `input_fields`**, so they resolve through
`ExtractActionInputs` Strategy 2 → `extractSingleField` Strategy 4 → aggressive
recursive search. `findFieldRecursive` walks `for key, val := range m`
(`unified_extractor.go:494`) — **Go randomises map iteration order.**

Running the real helper 400× on identical input with both `hero_stored.asset_id`
and `logo_stored.asset_id` present:

```
hero asset_id : 344/400   <- WRONG asset for the LOGO deploy step
logo asset_id :  56/400
```

**The logo step resolved the hero's `asset_id` in 86% of runs.** Priority 2 is
therefore *load-bearing*, not cruft: for these workflows the two assets differ
exactly by purpose, so the purpose key is the correct discriminator. Deleting it
swaps a correct lookup for an 86%-wrong one.

Today those steps in fact resolve earlier still: both carry `uri_field`
(`hero_result.image_uri` / `logo_result.image_uri`), which the spec's `Deprecated`
map bridges to `s3_uri` before `findStorageURI` is ever called.

## 3. `[MEASURED]` The defect is wider than "Priority 2"

Priorities **3–7** are *also* purpose-keyed (`{purpose}_result.image_uri`,
`{purpose}_stored.asset_url`, …, `deploy_image_asset_action.go:460-495`). "Delete
Priority 2 so the asset_id path is the only DB-free route" does not achieve that —
it falls through to five more purpose-keyed lookups. A real fix must change the
function's **keying**, not one branch.

## Recommended ranking, replacing the one above

1. **Key by asset identity, additively** (this file's candidate 2): write
   `asset_uris.<asset_key>` *alongside* `{purpose}_uri`, reader prefers `asset_key`
   when supplied (`asset-deployer` already passes it; `image-build-handler` passes
   `asset_key?`). Nothing deleted, every caller keeps working. **This is a platform
   seam** — new reserved key namespace on a shared action — so per the 2026-07-28/29
   rulings: concept-register entry *in the same commit*, council submission
   alongside, and the other consumers **told**, not merely measured.
2. **Give the legacy deploy steps explicit `input_fields`** — config-only, live
   immediately, and it fixes the general recursive-resolution hazard rather than
   just this bug.
3. Retire the purpose-keyed priorities only once no caller depends on them.

Candidate 3 (guard on "exactly one same-purpose asset") stays last: a guard against
a state the code can still express.

## Pinned in code

`platform/orchestration/actions/deploy_image_asset_purpose_source_test.go` — four
characterisation tests (all passing) recording the last-write-wins branch, that
distinct purposes do not collide, the 86% instability, and which route supplies the
legacy source today. They exist so a future thread reaching for candidate 1 trips
over the reason Priority 2 was kept.

## Still open

- `[UNVERIFIED]` Are `pageflow-builder` / `site-work-orchestrator` reachable at all?
  No live definition spawns or calls either, and neither ran today — but
  `orchestration_states` retains completed runs only ~24h (13 rows older than 24h,
  **0** older than 7 days), and the longer-retention cross-check was **blind**
  (`llm_call_log` returned 0 for `asset-deployer` on a day it ran 16 times; positive
  control failed, result discarded). **Dormant, not proven dead.** If they are dead,
  retiring them removes the legacy constraint and makes fix 1 clean.
- Does the recursive-`asset_id` instability bite other actions with a
  no-`input_fields` step whose spec field name occurs twice in `collected_data`?
  That is a shared-helper question wider than 209.

### STATUS 2026-08-09 (later) — INTO-LINE APPROVED; PHASE 1 APPLIED AND ROW-VERIFIED

The owner superseded the divergence direction after the cost question surfaced
`bugs_open/231` (the pair's logo steps resolve purpose="hero" — repair and
alignment turned out to be the same edit): **"carry on with the into-line fix …
Phase 1 first."**

**Phase 1 is LIVE**: migration `348_pageflow_swo_deploy_steps_resolve_by_identity.sql`
(applied 09:41:53, ROLLBACK sidecar alongside, induced DO/RAISE verify, scoped
apply, row-verified with the store steps as negative control). All four deploy
steps now resolve purpose/s3_uri/asset_id/domain by Strategy-0 dotted paths from
their own store step's output — deterministic (8/8 harness tests incl. the
store-failure corner: `input_fields` deliberately excludes `s3_uri`, so a failed
store degrades to a safe skip, never a sibling's bytes). This closes, for these
workflows: the 231 shadow, the 86% recursive-`asset_id` hazard, AND their use of
the purpose-keyed scratchpad route (`s3_uri` resolves before `findStorageURI`;
its fallback would now be reached only on a store failure, where it finds the
same per-purpose value the store didn't write — i.e. nothing).

**Still owed on this bug:**
1. **Behavioural proof** — one sacrificial-domain run of each workflow (hero.*
   and logo.* committed, different bytes, `logo_url` serving). Satisfies this
   file's own "verify on the workflow that really does it" bar.
2. **Phase 2** — delete `findStorageURI` + call site (~90 lines, council).
   **Ordering (statable): the image must not roll before 348 is verified at the
   rows** — done above, so Phase 2 is unblocked, but re-verify on pickup (a roll
   re-stamps `updated_at`; compare CONTENT — migration 341's `gate_next_item`
   surviving the 08-09 08:49 stamp is the measured proof that content survives).
3. Phase 3 (optional): retire the `{purpose}_uri` writers — blocked on
   classifying 6-per-definition other references; NOT needed to close this bug.

### OWNER RULING 2026-08-09 — the legacy pair is ALIVE, FROZEN; fix by divergence

*"pageflow-builder and site-work-orchestrator are not dead, but not being worked
on. If we need to diverge from them then we can use new actions and workflows as
necessary."* This retires the "still open" question below and constrains every fix
candidate: the legacy pair's behaviour must survive untouched, and new behaviour
arrives as an opt-in (field or new action), not as a change to what existing
callers get. Proposed shapes + the pending decision: the lane's
`PLAN_2026-08-08_209_purpose_keyed_source.md` §"DECISION PENDING".

### ADDENDUM 2026-08-09 — re-verified post-roll; finding 2's exposure stated precisely

Chassis v1.0.1270 re-applied seeds at 08:49Z (`updated_at` bumped on all four
relevant definitions); the deciding config facts were re-read **by content** and
are byte-identical to the 08-08 census, and no 209-relevant file changed at HEAD.
**The LATENT verdict and all three findings stand on v1.0.1270.**

Precision on finding 2, so it is not read stronger than measured: the 86%-wrong
`asset_id` resolution is the legacy steps' behaviour **when asked** — and today
they are never asked, because their primary route (the `uri_field` bridge →
`s3_uri`) resolves first, and their fallback (the purpose-keyed lookups) resolves
correctly when the primary fails. Candidate 1 leaves the primary intact and
replaces the **fallback**: correct-by-purpose becomes 86%-wrong-by-recursive-search.
Conditional exposure — it fires on the day a `*_result` map goes missing — but a
fallback that deploys the wrong asset's bytes exactly when the primary hiccups is
the defect class this whole family (152/155/209) exists to remove. Ranking unchanged.
