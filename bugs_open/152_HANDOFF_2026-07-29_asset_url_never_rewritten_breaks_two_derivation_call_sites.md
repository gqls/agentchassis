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
