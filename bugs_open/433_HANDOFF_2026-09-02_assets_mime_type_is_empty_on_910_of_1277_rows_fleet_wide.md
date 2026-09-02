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

---

## 2026-09-02 — the extension question is ANSWERED (no, it is not reliable), and **fix candidate 2 as written would write a WRONG value** (contributed by the `bugfix_417_420` lane)

I hit this censusing logo bytes for 417 and it lands directly on two of your candidates.

### The measurement `[MEASURED 2026-09-02]`

First 4 bytes of **12 of 12** `asset_key='logo'` source objects in `personae-prod-uk001-images`,
spanning **2026-08-10 → 2026-09-02**: **all `ffd8ffe0` — JPEG.** None `89504e47`. Domains:
advertise.co.uk, homegarden.uk, boxingonline.com, agritec.uk, webdesign.co.uk, dartsonline.com,
farmerinsurance.uk, robot-hands.com, loanandmortgagecalculator.co.uk, remortgagecalculator.uk,
webdesign.uk, gamesdesign.co.uk. Every one is stored under a **`.png`** key with `filename`
`logo.png` and `url` `/assets/images/logo.png`.

The disconfirming result would have been PNG magic on any row. There was none.
(Confirmed independently on the live path: the adapter's own log for the 17:03Z `designblog.co.uk`
generation prints `"source_format":"jpeg"`.)

### Why — three lines, and it is unconditional, not drift

- `internal/adapters/imagegenerator/dynamic_adapter.go:492` — the provider's real MIME is
  **discarded**: `imageData, _, origin, conditions, err := a.generateImage(...)`.
- `:717` — the key hard-codes the extension:
  `fmt.Sprintf("images/%s/%s/%s.png", clientID, timestamp[:8], imageID)`.
- `:726` — the upload hard-codes the type: `a.storageClient.Upload(ctx, fileName, "image/png", ...)`.

So the object is **named** `.png` and **served by B2 as `image/png`** while the bytes are JPEG.

### What this means for your candidates

- **Candidate 3 (do not backfill from the extension "without checking whether that is reliable"):**
  checked. **It is not reliable — it is wrong for 12 of 12 logos.** Extension, `filename` and `url`
  all carry the same false `.png`, so none of the three is usable as a backfill source. The bytes
  are the only honest source.
- **⚠ Candidate 2 needs amending before anyone implements it.** It proposes propagating
  `uploadImage`'s hardcoded `"image/png"` into the `assets` row. **That constant is false for every
  logo measured**, so the candidate as written would replace an empty column with a *confidently
  wrong* one — strictly worse, because an empty `mime_type` is at least honest and greppable,
  whereas 910 rows saying `image/png` would look repaired and be undetectable without re-reading
  every object. The value to propagate is `result.MimeType` (currently thrown away at `:492`),
  or a sniff of the actual bytes — not the constant.

Note this is **not** the same defect as 424's, though they meet in the same file: 424 is about the
image having no alpha; this is about the platform mislabelling what it stored. A logo that IS a
real PNG after 424's matting will still be written under the same hardcoded constant.

Fetch recipe for the bytes (through a pod, no key in session):
`docs024_key_docs_latest/bugfix_417_logo_text_policy/RUNBOOK_logo_text_policy.md`,
§"Fetch a generated asset's BYTES and LOOK at it".
