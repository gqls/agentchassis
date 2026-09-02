# 433 — `assets.mime_type` is empty on 910 of 1,277 rows fleet-wide, and only 5 rows are ever recorded as `image/png`

**Filed 2026-09-02, split out of `bugs_open/424`'s verification** (the boxingonline session, working
that bug's interim logo regeneration, ran the fleet-wide query below and flagged it as
"separate, worth filing on its own rather than as a note"). Diagnosis loop NOT run — this is a
direct, single-query DB measurement, not a claim about root cause; see "Not yet diagnosed" below.

## What was measured

`[MEASURED 2026-09-02]`

```sql
SELECT coalesce(nullif(mime_type,''),'(EMPTY/NULL)'), count(*) FROM assets GROUP BY 1;
  (EMPTY/NULL)   910
  image/jpeg     362
  image/png        5
```

910 of 1,277 assets (71%) carry no `mime_type` at all. Of the 367 that do, only 5 are recorded as
`image/png` — despite every generated-image upload path this estate has (`internal/adapters/
imagegenerator/dynamic_adapter.go`'s `uploadImage`, at minimum) hardcoding `Content-Type: image/png`
on the S3 `PutObject` call. So the column is not merely sparse — for the one asset TYPE most likely
to need its MIME type checked (a generated PNG, where `bugs_open/424`'s own fix needs to distinguish
a matted RGBA PNG from an unmatted RGB one), it is recorded almost never.

Confirmed at one concrete row: the boxingonline.com logo asset (`asset_key='logo'`,
`updated_at` 2026-09-02 10:40:12, regenerated same day) has `mime_type` empty, even though its
upload went through the hardcoded `image/png` path.

## Why this matters

- `assets.mime_type` is a real schema column (`\d assets` confirms `text`, nullable, no default,
  no NOT NULL constraint) — something clearly intended it to be populated per-row.
- Any future code that branches on `mime_type` (a content-type check before serving, a format-aware
  post-processing step, a migration that needs to tell PNG from JPEG without downloading the bytes)
  is branching on a column that is empty on 71% of rows and, within the rows that ARE populated,
  essentially never says PNG even where the upload path guarantees PNG. This is the shape
  `grep-the-config-key-before-calling-it-a-win.md` and `writes-the-field-is-not-reads-the-field.md`
  warn about, from the writer's side: a field that is barely written cannot be trusted by any reader
  that assumes it reflects reality.

## Not yet diagnosed — this file states the measurement, not the mechanism

The boxingonline session was explicit: *"I have not looked at the writer and am not claiming
which."* Candidates, none verified:

1. The image-generation store step (`store_*_asset`, whichever workflow step persists the
   `assets` row after `dynamic_adapter.go`'s `uploadImage` returns) may simply never set
   `mime_type` on INSERT/UPSERT, leaving it at its column default (NULL/empty).
2. Some OTHER asset-writing path (uploads, product images, affiliate custom images — see
   `product_assets`/`affiliate_products.custom_image_id` in the schema) may be the majority
   contributor to the 910 empty rows, in which case the image-generation path specifically might
   be a smaller, more tractable slice than the fleet-wide number suggests.
3. The 362 `image/jpeg` rows and 5 `image/png` rows may come from a DIFFERENT writer than the
   bulk-empty 910 — worth checking whether the populated rows cluster by `origin_type` or
   `asset_type` before assuming one writer explains the whole shape.

## Fix candidates (unordered — diagnosis owed first)

1. Find every INSERT/UPSERT into `assets` (`grep -rn "INSERT INTO assets\|ON CONFLICT.*assets"
   platform/ internal/`) and confirm which do and do not set `mime_type` from the actual upload
   result rather than leaving it unset.
2. For the image-generation path specifically: `uploadImage` (`dynamic_adapter.go`) already knows
   the exact content type it sent to S3 (`"image/png"`, hardcoded) — that value should reach
   whichever step writes the `assets` row, the same way `result.MimeType`/origin already does.
3. A backfill for existing rows is a SEPARATE decision from fixing the writer — do not conflate a
   go-forward fix with retroactively repairing 910 rows, and do not backfill by GUESSING a MIME
   type from a file extension without checking whether that is reliable across `storage_path`/`url`
   conventions first.

## Verify

Re-run the query above after any fix; the writer confirmed should show `mime_type` populated on
its next-written rows without a backfill.

## Related
- `bugs_open/424` (the verification that surfaced this — not the cause of that bug's defect; this
  file exists precisely so the two do not get conflated).
