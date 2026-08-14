# 152 — `assets.url` is never rewritten off its presigned URL, and two call sites fetch it directly

**Filed:** 2026-07-29 by the leopardessconsulting workstream, while cleaning up missing
images on that site (`docs/leopardessconsulting/RUNNING_NOTES.md`, session 2026-07-29).
**Severity:** Medium. Nothing is broken today on any site checked. It is a timer: every
image generation on the platform writes a URL that goes dead in 7 days, and two real
call sites read that column and fetch it.
**Class:** a documented landmine (RUNBOOK) that was never promoted to a filed, fixable
defect, so it kept recurring per-site instead of being fixed once.
**Status:** OPEN, unowned.

---

## The defect

`deploy_image_asset` only rewrites `assets.url` from the provider's presigned S3 URL to
the deployed git path **when called with `asset_id`** (RUNBOOK O5 landmine 6, already
known). Measured on leopardessconsulting.co.uk 2026-07-29: **every one of 13 active
hero/infographic asset rows** — including four generated minutes earlier in the same
session — carried a presigned `https://s3.us-east-005.backblazeb2.com/...&X-Amz-
Expires=604800&...` URL. Probed one directly:

```
$ curl -s https://s3.us-east-005.backblazeb2.com/.../20260718/....png?X-Amz-...
<Error><Code>UnauthorizedAccess</Code>
<Message>Request has expired given timestamp: '20260718T095833Z' and expiration: 604800</Message></Error>
```

So this is not a one-off from an old batch — it is standing behavior of the current
deploy path. Every image generated on any site enters a 7-day countdown in this column
regardless of when it was made.

## Why this is a live risk, not just stale data

The rendering path itself is fine: `plan_sections_action.go`'s `ensureAssets` explicitly
resolves the hero via `storage.DeployedWebPath(assetKey, purpose)`, **not** `assets.url`,
with a comment saying so ("NOT assets.url, a presigned S3 URL that expires and is
per-generation"). That is why every page on leopardess renders correctly today despite
every row being presigned.

But two call sites read `assets.url` and fetch its content directly:

- `derive_brand_head_assets_action.go:94` — `SELECT a.url, si.domain ... WHERE
  a.asset_key = 'logo'` — derives favicon/og-card from the logo asset.
- `derive_card_asset_action.go`'s `findCardSourceHero` (:216, :227) — `SELECT a.id,
  a.url, a.asset_key ...` — derives a card thumbnail from the page's hero asset. This is
  the Phase I3 card-derivation path leopardess's blog-thumbnail gap needs (see this
  session's task list; deferred rather than built, so not yet exercised against these
  rows, but it is the mechanism that would consume them).

Whenever either fires against a row older than 7 days, the fetch gets the 401 shown
above instead of image bytes. leopardess's logo happens to be safe (its row already held
a local path, hand-set — `leopardess-mark`), but nothing structural stops a fresh logo
row from being presigned-only like every hero/infographic row was.

## Fix candidates

1. **Always pass `asset_id` to `deploy_image_asset`.** If every call site already has
   the row (it must, to know what to deploy), this may be a one-line change per caller
   — worth checking whether any caller currently omits it by choice or by oversight.
2. **Rewrite `assets.url` at read time in the two derivation call sites**, mirroring
   `plan_sections_action.go`'s own workaround: resolve via `storage.DeployedWebPath`
   instead of trusting the stored `url` column, the same fix already applied to
   rendering.
3. **Stop writing a fetchable URL to `assets.url` at all** once the deploy step knows
   the local path — write the local path immediately, or a null, rather than a value
   that is correct for exactly one week.

Any of these closes the door for every site, not just this one; (1) is the smallest.

## What was done for leopardess (contained, not a fix for the class)

Directly rewrote `assets.url` to the already-verified-200 local path for all 12
currently-active rows (12 `UPDATE`s, one `SELECT` per row confirming 200 before
writing), and retired one orphaned row (`hero_case_studies`: wrong-provider SDXL
generation from 2026-07-17, wired to no `site_plan_imagery` row, referenced in no
page's `content_data` — same shape as `bugs_open/114`'s "generated and never wired"
class, on a smaller scale). This buys leopardess time; it does not touch the call
sites above, so the same defect will recur on the next generation run, here and on
every other site.

## Related, not duplicate

- `bugs_open/143` — same file (`derive_card_asset_action.go`), different defect
  (lock-check runs after the git commit, not before). Independent of this one; a fix
  for either does not fix the other.
- `bugs_open/114` — the broader "generated, deployed, never wired" pipeline-integrity
  class. This bug is narrower and mechanical: a specific column that is wrong for a
  specific, predictable, and fixable reason.

## How to verify a fix

Generate an image on any site, confirm `assets.url` is a `/assets/images/...` path (not
`s3...X-Amz...`) immediately after deploy — no need to wait 7 days. Then re-run
`derive_brand_head_assets` / `derive_card_asset` against a row that WOULD have been
7+ days old under the old behavior and confirm the fetch succeeds.

---

## Progress 2026-08-06 — the bug MORPHED, then got its structural fix (with 155, one commit)

**Re-verified at HEAD before acting, because the ground moved since filing.** The
"fetch the expired URL, get a 401" symptom as filed is GONE: none of the readers
fetches `assets.url` any more — all of them PARSE it to a storage key and download
with the client's own credentials (`derive_brand_head_assets_action.go:135`,
`derive_card_asset_action.go:165`, both via `presignedURLToS3URI` → 
`ExtractKeyFromS3URI`), so an expired signature stopped mattering. Meanwhile
IMG-053's "Edit F" landed the deploy-time rewrite this file's candidate 3 asked
for: `deploy_image_asset_action.go:357-381` flips `url` to the deployed local
path and preserves the source into `storage_path`.

**The live defect that replaced it: the flip strands every url reader, because
NOTHING reads `storage_path` back.** Census 2026-08-06: 107 flipped rows carry a
recoverable `storage_path` no reader consults; 49 flipped rows (hand-repairs,
incl. this file's own leopardess workaround, and derived-card upserts) have
neither. Five sites' ACTIVE logo rows carry a non-presigned `url` (four of them
the `input-data.asset-key.jpg` template literal) — the next brand-head derivation
on webdesign.co.uk / gaswholesalers.com / finetuning.uk / vetcomparison.uk /
leopardessconsulting.co.uk fails at "could not derive storage key from logo url".
A fourth reader was found in the same sweep: `resolveReferenceAssetURIs`
(`imagery_style_guide.go`) reads `url` only and skips SILENTLY — style anchors
vanish the moment the referenced asset is deployed.

**Declared per the owner ruling of 2026-07-31**: this updated mechanism was
established by first-hand verification substituting for a 090 run — all four
readers and both writers read at HEAD with line citations, plus a live census
whose result could have come out otherwise. 155's shared-cache half was already
loop-CONFIRMED (corr `0dd9aee4`).

**The fix (committed with 155's, one coherent change):** `storage.AssetSourceRef`
— one derivation of "which stored object is this row's source" (storage_path
first, url parse fallback; returns s3:// or bare key, never https/local, "" =
fail loud) — resolved through by all four readers; `StoreAssetAction` writes
`storage_path` at generation; `sql_for_agents/323` backfills the 205 remaining
presigned-only rows (w9_04 fleet-wide). Full plan + evidence:
`docs024_key_docs_latest/bugfix_152_155_asset_source_identity/PLAN_2026-08-06_asset_source_identity.md`.
Register: IMG-068.

**Fix-candidate disposition against the original list:** candidate 3 (write the
durable value) = StoreAssetAction now records `storage_path` at birth and deploy
still flips `url`; candidate 2 (readers resolve durably) = all four via
`AssetSourceRef`; candidate 1 ("always pass asset_id") was a TRAP — 155 showed
passing `asset_id` routed INTO the wrong-bytes purpose-cache branch; that branch
is now deleted.

**Still to close:** image roll + pod-grep (positive `AssetSourceRef` ≥1, negative
`Resolved s3_uri from site content_data via asset_id` = 0, both replicas); then
this file's own verify recipe (fresh deploy → `url` local AND `storage_path`
names the source; a derivation against that row succeeds).

**Council APPROVED round 1** — `c055840a-9edc-4f9a-8a4a-b23ac4cad02a`, 8 advisory
objections, none high. Two mediums found real things and are discharged in
`bb53326a8`: (a) my reader census could not see queue-built or external callers,
and there WAS one — `scripts/initial_messages/180_adoption/081c_direct_asset_
deployer.sh` hands an operator one URI per *purpose* to paste into a deploy,
which is bug 155 by hand; rewritten to read the asset row. (b) the migration
number was not mine — a concurrent session committed its own `321` inside my
window, so it is now `323` (applied under the old name; the file records that,
its pre-flight count and its rollback statement). Migration `323` APPLIED
2026-08-06: 205 rows, presigned-with-no-`storage_path` now **0** fleet-wide, and
all five at-risk logo rows resolve via `storage_path`.

---

## LIVE 2026-08-06 on chassis `v1.0.1259`

Pod-verified both replicas with positive, negative and nonsense controls (the table
is in `bugs_open/155`'s live section — same binary, same exec). All four readers now
resolve through `storage.AssetSourceRef`, so a deployed asset's `url` being a local
web path no longer strands anything that has a `storage_path`.

Fleet state after migration `323`: **0** rows presigned-with-no-`storage_path`, and
all five previously at-risk logo rows (webdesign.co.uk, gaswholesalers.com,
finetuning.uk, vetcomparison.uk, leopardessconsulting.co.uk) resolve via
`storage_path` — so the next `derive_brand_head_assets` on each will find its source
instead of erroring, which it would have done before today.

**Owed before this closes**: a real derivation run against one of those five logo
rows (favicon/og-card re-derived and served), which is the artefact-level proof;
the 49 genuinely stranded rows stay stranded by design and are NOT a blocker — their
source is unrecorded and inventing one would be worse than failing loud.


## Recurrence data — leopardessconsulting.co.uk (2026-08-14, services-restore session)

Two ACTIVE rows on this site carry presigned S3 urls (`X-Amz`), both created AFTER the
152+155 fix rolled (v1.0.1259, 2026-08-06):

- `hero_case_studies` — re-created 2026-08-08 21:34Z, active
- `content_hero_tool_automation_savings_estimator` — 2026-08-11 20:24Z, active

(Five older retired rows are also presigned, 2026-01-28 → 07-17.) Confirms the
creation-side recurrence on this site, as the 2026-08-12 measurement filed. A
counter-example from the same site today: two icon assets created 2026-08-14 via
scope-less `needs_imagery` → image-build-handler were born with clean `/assets/images/`
urls — so not every creation path writes presigned; the recurring writer is on the
hero/content-hero path(s), not the icon path.
