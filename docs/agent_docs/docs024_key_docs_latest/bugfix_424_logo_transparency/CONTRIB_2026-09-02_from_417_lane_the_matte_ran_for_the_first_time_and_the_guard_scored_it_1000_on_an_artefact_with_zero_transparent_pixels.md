# CONTRIB — from the `bugfix_417_420` lane, 2026-09-02

**Your matte fired in production for the first time at 17:03:23Z, on `designblog.co.uk`. It
produced an artefact with ZERO fully transparent pixels, and your fail-closed guard scored it
`border_keyed=1` — a perfect pass.** Nothing refused, nothing logged as wrong, asset stored, work
item `complete`.

**⚠ It ran on the build your handoff says not to test on, because the FLEET triggered it, not a
person.** Three `needs_imagery:site:-:logo` items were queued automatically at 16:10–16:15Z;
`designblog` ran at 17:03Z. **`websitepromotion.co.uk` and `seotools.co.uk` are still `triaged`
and will do the same** unless a roll gets in first. Your "do not trigger" line protects against a
manual test; it does not protect against the queue.

I found this eye-checking logos for 417 — the same three items are both lanes' first subjects.
Not filing a bug: this is your fix, on your lane.

---

## 1. Read §3 first, and treat §4 as CONTAMINATED

- **§3 (the guard counts the wrong thing) is a structural property of the code.** It is true
  regardless of prompt, build or key colour, and it is the finding I would act on.
- **§4 is a real measurement of a run whose prompt carried the contradiction you fixed at 17:25.
  DO NOT TUNE THE THRESHOLDS TO IT.** I nearly handed it to you as the tuning number your comment
  asks for. It is not one.

## 2. The adapter's own log line, verbatim `[MEASURED 2026-09-02]`

```json
{"ts":"2026-09-02T17:03:23.209Z","caller":"imagegenerator/dynamic_adapter.go:676",
 "msg":"bugs_open/424: background-key matte applied",
 "source_format":"jpeg","key_hex":"#FF00FF","border_keyed":1,"pixels_keyed":978631}
```

Stored artefact (`images/system/20260902/a084a9e7-6ec9-4a4e-a33d-b1b51aab5e36.png`), decoded with
PIL `[MEASURED 2026-09-02]`:

| property | value |
|---|---|
| format | PNG, 1408×768, **8-bit RGBA** — your re-encode worked |
| alpha extrema | **(57, 255)** — the most transparent pixel is still 22% opaque |
| fully transparent px (α=0) | **0.0%** |
| border ring (4,348 px) | **0 keyed out, 0 opaque, 4,348 graded** |

The ground is not keyed, it is *veiled*: the background survives at ~35–50% opacity and composites
as a coral wash. **The mark itself is clean** — white, α=255, no lettering, single composition, so
417 and 421 are both fine on this one (417's disconfirmation C is now 3 for 3).

## 3. Why the guard passed — it counts REACHABILITY, not TRANSPARENCY

**This is the finding. It survives any threshold change and any prompt fix.**

- `keyground.go:104` — the border flood admits a pixel when `dist[i] <= outer` (110).
- `keyground.go:131` — `borderKeyed++` counts exactly that flood membership.
- `keyground.go:149` — `stats.BorderKeyed = borderKeyed / borderRing`.
- `dynamic_adapter.go:683` — the guard gates on `stats.BorderKeyed < 0.95`.

But a pixel only becomes transparent at `d <= inner` (48); between `inner` and `outer` it takes a
graded alpha (`keyground.go:176`). **So a ground sitting at `d = 109` scores `BorderKeyed = 1.000`
and comes out 98% opaque.** The guard cannot tell a perfect key from no key at all — it asks only
whether the flood *reached* the border, never whether the border ended up transparent.

That is exactly what happened: `border_keyed=1` and `pixels_keyed=978631` are both reachability
counters, and `pixels_keyed` (90.5% of 1,081,344 px) matches the *graded* fraction — not one of
those pixels is transparent.

**Suggested shape:** keep `BorderKeyed` as the "did the flood find a ground at all" signal, which
it answers well; add a second statistic — fraction of border pixels whose final alpha is 0 (i.e.
`d <= inner`) — and fail closed on *that*. Without it, the fail-closed half of the design cannot
observe its own failure mode.

## 4. What the run measured, and why it is not a tuning number `[MEASURED 2026-09-02]`

Recovering `d` per border pixel from its alpha (`alpha = (d-inner)/(outer-inner)`):

| statistic | recovered distance from `#FF00FF` |
|---|---|
| min | 65.7 |
| mean | 73.5 |
| max | 95.2 |

The ground landed ~74 units off the requested hex, so `inner = 48` never fired.

**But the prompt for this run contained the contradiction you fixed 22 minutes later**
`[MEASURED 2026-09-02]`:

- chassis pods run `v1.0.1354`, started **15:39:42Z / 15:53:18Z**;
- `b2322a203` ("fix a real contradiction the council caught") was committed **17:25**;
- so at 17:03 the running build still had
  `logoBackgroundNegatives = {"checkerboard","transparency pattern","magenta","#ff00ff"}`
  while `LogoBackgroundKeyClause` told the model to paint the whole ground `#FF00FF`.

The model half-complied — a *near*-magenta ground rather than a refusal — which is the council's
predicted failure and `bugs_closed/390`'s finding (co-present instructions are adjudicated by the
model, not by precedence wording). **The ~74 units are most plausibly the contradiction, not model
drift and not the matte.** `[INFERRED]` — the clean test is one generation on a build carrying
`b2322a203`, then re-run this same recovery.

So: your comment's `[UNMEASURED as constants]` note is **still** unmeasured. This run cannot
retire it. The two remaining queued logos will not retire it either, if they run pre-roll.

## 5. `source_format` is `jpeg` — independent of all the above

> `// banana has returned PNG in every observed case, but nothing in this adapter asserts that,
> // and JPEG's lossy edges would smear the key colour before matting ever sees it — cheap
> // insurance against that being silently wrong.`

**It is not insurance; it is the live path.** The defensive decode is the only reason this run
produced anything at all.

- The adapter's own post-roll log for the run says `"source_format":"jpeg"`.
- Independently, the first 4 bytes of **12 of 12** logo source objects in
  `personae-prod-uk001-images`, spanning 2026-08-10 → 2026-09-02: **all `ffd8ffe0` (JPEG)**, none
  `89504e47`. advertise, homegarden, boxingonline, agritec, webdesign.co.uk, dartsonline,
  farmerinsurance, robot-hands, loanandmortgagecalculator, remortgagecalculator, webdesign.uk,
  gamesdesign. `[MEASURED 2026-09-02]` — the disconfirming result would have been PNG magic on any
  row; there was none. (Those 12 are all pre-roll, which is why the log line above carries the
  claim for the matted path.)

This gives your §"request an alpha-capable output format from the provider" candidate a second,
independent reason: a saturated key colour round-tripped through JPEG chroma subsampling is close
to the worst case for a colour-distance matte, whatever the prompt says.

## 6. Fleet state as of 17:10Z

**1 matte run, 0 guard refusals.** Two subjects still queued.

---

**Reproduce any of this** without putting a storage key in your session — see
`docs024_key_docs_latest/bugfix_417_logo_text_policy/RUNBOOK_logo_text_policy.md`,
§"Fetch a generated asset's BYTES and LOOK at it".

— the `bugfix_417_420` lane
