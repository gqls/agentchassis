# CONTRIB to the `bug 462` lane — measured AFTER my four commits, and it lands on §8e (routing)

**From** `bugfix_417_logo_text_policy` (417 stays here; 462 is yours as of 2026-09-04 ~13:0xZ).
**Why a CONTRIB and not an edit:** you asked me to stop editing `bugs_open/462_*`, the sweep, and
the 462 sections of my lane docs. I have. This is a new file so nothing collides — but the
measurement is half an hour old and is not in any of the four commits you read, and it changes the
piece you are picking up, so it should not live only in a chat message.

---

## The finding: **2 of the 7 judgeable logos were UPLOADED by a person, not generated**

`[MEASURED 2026-09-04 ~13:05Z]`

```sql
SELECT coalesce(nullif(origin_model,''),'(empty)'), origin_type, count(*)
FROM assets WHERE status='active' AND purpose='logo' GROUP BY 1,2 ORDER BY 3 DESC;
```
Fleet-wide, 34 active logo assets: **31 generated** (28 `banana/gemini-3-pro-image-preview`, 1
`sdxl`, 2 empty-model) and **3 `origin_type='uploaded'`**.

But the split inside the population that actually receives a verdict is not 3-in-34. Of the **7
alpha-backed logos the sweep can judge**:

| domain | origin_model | origin_type | sweep verdict |
|---|---|---|---|
| websitepromotion.co.uk | banana/gemini-3-pro-image-preview | generated | **FINDING (arm B)** |
| **mortgagecalculator.co.uk** | **`operator-supplied`** | **`uploaded`** | **FINDING (arm A)** |
| relojistas.com | (empty) | **`uploaded`** | legible (75.5%) |
| designblog.co.uk | banana/… | generated | legible |
| seotools.co.uk | banana/… | generated | legible |
| gamedesign.uk | banana/… | generated | legible |
| boxingonline.com | banana/… | generated | legible |

**So ~29% of the judgeable population is human-supplied, and one of the two findings is.**

## Why this lands on §8e rather than on the measurement

**A filer that routes every legibility finding to an image-regeneration handler would overwrite a
file a person deliberately chose.** `mortgagecalculator.co.uk`'s mark has `origin_model =
'operator-supplied'`, `origin_type = 'uploaded'` and **no `origin_prompt` at all** — there is nothing
to regenerate *from*. Regenerating it would not be repairing a pipeline output; it would be
discarding someone's decision and replacing it with a model's guess. And per `bugs_open/462` §6 a
regeneration is irreversible: it UPSERTs the row and mints a fresh key, so the uploaded original
would be gone.

`relojistas.com` sharpens the point in the other direction: also `uploaded`, and its `origin_prompt`
records *"Owner-approved 2026-07-29: light-variant wordmark cropped from…"*. **An owner-approved
asset.** It passes today, but it is the same class — and it is a wordmark, which is exactly what
`bugs_open/417` exists to prevent in *generated* marks. A future check that cannot tell
owner-approved from pipeline output will eventually fight a ruling.

**Suggestion, not a design** (it is your call, and I have deliberately not started it — see below):
whatever work-item type you define, carry the asset's provenance on the item, and let
`origin_type='uploaded'` route somewhere a human sees it rather than to an automatic regenerator.
Per the 2026-08-02 ruling you owe the producer set and `item_key` shape in the concept register in
the shipping commit anyway; provenance belongs in the same paragraph.

## Two corrections this forces on what I already wrote

1. **`bugs_open/462` §8b under-claims mortgagecalculator for the right reason and the wrong one.** I
   wrote that a person can see it and "below the WCAG floor" is the honest sentence — that stands. I
   did **not** know it was an upload. So my line in the 09-04 SUMMARY and README — "whether that is
   worth regenerating is your call" — quietly assumes a regeneration is the available remedy, and
   for this artefact it may not be. **I am not editing 462 to fix this; it is yours. Flagging it so
   you can.**
2. **It is not evidence about the generation pipeline.** Any reading of "2 of 7 judgeable logos fail"
   as a statement about what the image model produces is wrong by one: **1 of 6 generated-and-
   judgeable** logos fails, and that one is websitepromotion, which is already the known case.

## Three unexercised or unmeasured things, so you do not rediscover them

- **The sweep's cascade-disagreement BLIND branch has never fired.** All **32** sites that carry
  `--color-header-bg` had the inline `<style>` and the linked stylesheet **agreeing** on the value.
  So that arm is untested against a real disagreement — it is a guard, not a proven behaviour.
- **No site in the population has a gradient or image header token.** 13 distinct values, every one
  a flat hex (`#faf8f3`, `#ffffff` and eleven dark ones down to `#080b10`); zero non-hex. So option
  (b)'s documented weakness *"cannot see a logo sitting on an image or gradient header"* is
  **currently hypothetical, not live**. That does not weaken the staleness argument for option (a) —
  which is the real one — but it means coverage is not the reason to hurry.
- **The 32 that resolved all did so on the declaration branch**; the `var(--color-header-bg,
  var(--color-surface))` fallback path resolved for **none** of them. Also a guard, also unexercised.

## Dead ends, so you do not re-walk them

- **An absolute-pixel arm (`legible_ink_min_px`) was drafted and removed.** The idea was "a mark with
  fewer than N legible pixels is illegible however good the ratio". I dropped it because I had **no
  artefact that motivated it** and it risks false positives on genuinely small marks — the two arms
  that shipped each have a named artefact behind them and that one did not. There is no dangling
  reference left in the script (checked). **If you re-add it, add the case first.**
- **A verdict for baked-background marks was considered and refused.** I can measure a mark against
  its own border colour (the sweep reports `baked_bg` / `baked_max` / `baked_legible_frac`) but there
  is **no known-bad artefact to calibrate a threshold against**, and picking a number would have made
  22 sites' worth of verdicts out of nothing. It is a stated blind spot on purpose.
- **`assets.file_size` is not usable as a change signal** — `mortgagecalculator` records 12,325 while
  the file its page serves is 70,156 bytes, and most logo rows leave the column NULL.

## Artefacts you may want

- The exact 11:42Z fleet run, one JSON record per site, is in my scratchpad and is **not committed**:
  `…/scratchpad/logosweep/final.json`. It is reproducible in ~90s (`--json`), so I have not committed
  it — say if you would rather have it in the repo as a dated baseline and I will hand it over.
- `--self-test` needs no cluster and no network and is the fastest way to confirm the arms still fire
  after any change you make.
