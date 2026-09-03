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

---

## FIX SHIPPED 2026-09-03 (commit `afcf3ebdb`, council `82989388`) — candidate 3 is ANSWERED, and candidate 2 was right to be amended

Picked up by the `bugfix_451_457_433_unowned_queue` lane: this bug was unowned for the fix. It was
split out of `bugs_open/424`'s verification by a session that said plainly *"I have not looked at the
writer and am not claiming which"*, and the 424 lane closed the same day.

### Candidate 3 (which writer?) — answered by census, not by reading code

`[MEASURED 2026-09-03 ~16:0xZ]`, by `asset_key` shape:

| writer | rows | mime_type |
|---|---|---|
| `card_*` → `derive_card_asset_action.go` | 389 | **all populated**, none empty |
| `favicon` / `og_card` → `recordDerivedAsset` | 68 | 66 empty + 2 png |
| everything else → the `StoreAssetAction` path | 961 | **957 empty** + 3 png + 1 jpeg |

**Two writers account for all 1,023 empty rows.** The populated ones are not a different era or a
different origin_type — they are a different *writer*, the only one that encodes its own bytes and
could therefore state the type truthfully. Refreshed totals: **1,023 empty / 390 jpeg / 5 png** of
1,418 (this file's 910/362/5 was 2026-09-02 — the population grows).

### The question this file never asked, and it decides everything else

A row describes **two** artefacts. `url` and `filename` are written by `deploy_image_asset` and
describe the **deployed** artefact; only `storage_path` describes the source object in B2. So the
rule shipped is: **`mime_type` describes the artefact at `assets.url`, recorded by the writer that
publishes those bytes, or NULL.** A consequence to state rather than hide: a row stored but never
deployed keeps NULL, so **the fleet count will not reach zero**, and anyone quoting "still 175
empty" as a failure will be wrong. `[MEASURED 2026-09-03 16:2xZ]` 806 of the empties carry a
deployed url; 175 do not.

### ⚠ Candidate 2's amendment was right, and it was nearly re-made in a NEW place

The 417/420 CONTRIB warns against propagating `uploadImage`'s hardcoded `"image/png"`. Correct. But
I then argued that `deploy_image_asset`'s `processed.ContentType` was *"truthful by construction"*,
because `DownloadOptimizeAndPrepare` derives it from the same `extension` that selects `png.Encode`
vs `jpeg.Encode`. **Both readings were right and the conclusion was false.**
`DownloadAndOptimizeImage` returns the **original, un-re-encoded bytes** when optimisation fails
(logging `"Optimization failed, using original"`), and the content type was derived from the
**purpose**, not the bytes. Propagating it would have filled ~1,000 rows with a confidently wrong
value in a second place. Recorded in `WRONG_CALLS.md` 2026-09-03; the cheap check is *for any "X and
Y cannot disagree" claim, read every path that can produce X.*

### What shipped

`platform/storage/image_format.go` — one place that answers *what format are these bytes, really*,
by **magic bytes** (not `image.DecodeConfig`: Go's image registry is process-global, so a
decode-based answer differs between binaries). **It has NO FALLBACK.** Unrecognised input returns
empty. Every defect in this file began as a plausible default — `mimeFromKey`'s
`default: return "image/png"` is literally commented *"PNG is the safest fallback"*.

`DownloadOptimizeAndPrepare` now sniffs the bytes it is about to publish, and **logs** a disagreement
with the purpose's extension rather than refusing. Refusing is the tidier invariant and would take
sites down: nothing in this repo registers a webp decoder, yet `image/webp` is a possible provider
response, so such an image ships under a `.png` name today and browsers sniff it.

All four writers now obey one rule: `deploy_image_asset` records it beside the `filename`/`url` it
already writes; `derive_card_asset` sniffs its own encoded bytes instead of restating `'image/jpeg'`
**and gained `mime_type` in its `DO UPDATE SET`, which it lacked — so an upsert onto a
StoreAssetAction-born row refreshed every other column and left the type stale forever**;
`recordDerivedAsset` takes the type from the PNG bytes it just encoded; and `StoreAssetAction`
writes explicit NULL **and clears it on upsert**, because that upsert resets `url` to a fresh
presigned *source* url while any existing `mime_type` describes the previously *deployed* artefact.
Clearing is part of the invariant, not an omission from it.

Framework half: `TestEveryAssetsWriterThatSetsURLAlsoSetsMimeType` is **ban-shaped** — the exemption
map is empty, because after this every writer complies. It asserts column-list membership, never
value non-NULL-ness (a source scan cannot judge that, and a NULL from a writer that sniffed and
failed is exactly right), and it reads the **statement**, never a line window — this package logs
`zap.String("mime_type", …)` near several of these writes.

### The gate that licensed the whole design, run first

`TestEveryImagePurposeEncodesToItsDeclaredExtension` feeds every purpose in `ImagePurposes` the
**opposite** format and asserts the output's magic bytes match the extension it deploys under. It
passes for every purpose. That is what makes "read a deployed artefact's format from its purpose"
legitimate — the 417/420 lane's "extension is a lie" finding is about the **source** object, where
nothing re-encodes, and conflating the two is the one mistake that would make a backfill wrong.

### ⚠ Correction to this file's framing: the object-level lie HAS a live consumer

This file (and I) treated the `.png`-named JPEG as a labelling problem with no reader. It is read.
`referenceFetcher.fetchS3` derives a reference image's MIME from the **key extension** via
`mimeFromKey`; `fetchHTTP` takes **B2's response header** — the literal `image/png` `uploadImage`
set. Both feed `banana/provider.go`'s `ReferenceImage.MimeType` and thence Gemini's `inlineData`.
So the platform tells Gemini `image/png` over JPEG bytes on the live style-anchoring path. That
makes fixing it a correctness fix, not housekeeping — **and sniffing in the fetcher retro-fixes
every existing mislabelled object with no re-upload.** Deliberately NOT in this commit; it needs its
own tests.

### Backfill: still NO, and here is the trap inside the obvious predicate

Of the 806 empties carrying a deployed url, the purposes include **`illustration` (32),
`infographic` (7) and one empty** — **none of which is a key of `ImagePurposes`**, so
`GetImageConfig` falls through to `default` silently. A purpose-derived backfill would assign
`image/jpeg` to 40 rows on the strength of a fallback rather than a fact. Any backfill needs an
**enumerated** purpose list (766, not 806) plus a byte spot-check of ≥5 deployed artefacts first.
