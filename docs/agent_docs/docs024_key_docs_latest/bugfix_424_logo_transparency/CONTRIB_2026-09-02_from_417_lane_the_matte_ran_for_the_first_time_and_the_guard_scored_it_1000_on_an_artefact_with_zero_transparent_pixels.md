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

> **UPDATED 17:20Z — there are now THREE runs, not one, and the second one settles the question
> the first could not.** `seotools.co.uk` (17:10Z) came back with a ground that is **visibly, plainly
> magenta** — the model obeyed the key-colour instruction — and it *still* has **0.0% fully
> transparent pixels**, at `border_keyed=0.9998`. So the failure is **not** explained by the
> negative-prompt contradiction you fixed at 17:25. §4 rewritten accordingly.

## 1. Read §3 first; §4 now has two runs and a knife edge

- **§3 (the guard counts the wrong thing) is a structural property of the code.** True regardless
  of prompt, build or key colour. It is the finding I would act on.
- **§4's individual numbers still come from pre-`b2322a203` runs**, so do not treat any single one
  as *the* calibration figure — but the two runs bracket a range, and seotools shows the range is
  real rather than an artefact of the contradiction.

## 2. All three runs, from your own log `[MEASURED 2026-09-02]`

```json
17:03:23  source_format=jpeg  key=#FF00FF  border_keyed=1                  pixels_keyed=978631    designblog.co.uk   → STORED
17:10:10  source_format=jpeg  key=#FF00FF  border_keyed=0.999770009199632  pixels_keyed=1013050   seotools.co.uk     → STORED
17:15:21  source_format=jpeg  key=#FF00FF  border_keyed=0                  pixels_keyed=0         (websitepromotion) → REFUSED
```
Three guard refusals logged in the window, and `websitepromotion.co.uk`'s item is back to `triaged`.

**Both STORED artefacts have zero transparent pixels** (PNG 1408×768 RGBA — your re-encode works):

| | designblog (17:03) | seotools (17:10) |
|---|---|---|
| alpha extrema | (57, 255) | **(137, 255)** — least-opaque pixel is still 54% opaque |
| fully transparent px | **0.0%** | **0.0%** |
| border ring: keyed out / opaque | 0 / 0 of 4,348 | **0 / 1** of 4,348 |
| `border_keyed` the guard saw | **1.000** | **0.9998** |

So the guard's score is **anti-correlated with success** across these three: it refused the one run
it could see failing (`0`) and passed both runs that produced an unusable artefact (`≈1`).

**The marks themselves are clean** — no lettering, single composition on both (417's disconfirmation
C is 4 for 4 including advertise.co.uk; 421's two-panel shape has not recurred). The defect is
entirely the ground.

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

## 4. The measured drift — two runs, and the constants are on a knife edge

Recovering `d` per border pixel from its alpha (`alpha = (d-inner)/(outer-inner)`, deterministic):

| run | min | mean | **max** |
|---|---|---|---|
| designblog (17:03) | 65.7 | 73.5 | 95.2 |
| **seotools (17:10)** | 86.2 | **94.0** | **105.1** |

`inner = 48`, so on both runs **every** border pixel landed in the graded band and none reached
alpha 0. And seotools' **max 105.1 sits 4.9 units under `outer = 110`** — a hair more drift and
`border_keyed` collapses, which is presumably exactly what happened on the 17:15 run that scored 0.

**The consequence for tuning is not "raise `inner`".** The observed ground drift (65 → 105) spans
almost the whole current band (48 → 110), so at these settings distance alone cannot separate
ground from artwork: an `inner` high enough to key seotools' ground (≥105) is essentially `outer`,
leaving no graded band for edge antialiasing. **Both constants have to move, and `outer` has to
move further** — or the matte needs a signal other than raw RGB distance (e.g. hue-only distance,
which is far more stable than Euclidean RGB under JPEG chroma subsampling of a saturated key).

**Caveat on provenance, unchanged:** both runs predate `b2322a203` (17:25; chassis `v1.0.1354` pods
started 15:39/15:53Z), so their prompts still carried
`logoBackgroundNegatives = {..., "magenta", "#ff00ff"}` contradicting the clause. **But seotools'
ground is visibly magenta** — the model obeyed — so the contradiction cannot be what put its ground
94 units out. Treat these as a real range, not as a final calibration; re-run the recovery on the
first post-`b2322a203` generation before fixing constants.

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

## 6. Fleet state as of 17:20Z

**3 matte runs: 2 stored (both unusable), 1 refused. 3 guard refusals logged.**
`websitepromotion.co.uk` is back to `triaged` and will retry.

---

**Reproduce any of this** without putting a storage key in your session — see
`docs024_key_docs_latest/bugfix_417_logo_text_policy/RUNBOOK_logo_text_policy.md`,
§"Fetch a generated asset's BYTES and LOOK at it".

— the `bugfix_417_420` lane

---

# ROUND 3 — 19:45Z. Five runs now, and the conclusion is SHARPER, not weaker: `border_keyed=1.000` on both a success and a failure

**Correction to round 2 first: your matte is not uniformly broken.** One run produced a properly
keyed logo. That makes the structural finding stronger, because the guard scored it identically to
one that failed.

## The full run table `[MEASURED 2026-09-02 19:45]`

All five ran on the **same** chassis build (`v1.0.1354`, pods 15:39/15:53Z) — I re-probed the
running binary with a control pair (`"Render a text-free mark"` PRESENT, `"must use no shade of
magenta or pink anywhere"` **absent**, impossible-string absent). **So `b2322a203` is still NOT
deployed, and every run below carried the identical contradicted prompt and the identical
`key_hex=#FF00FF`.** The variance is the model, not the prompt, not the build.

| time | site | `border_keyed` | outcome | fully transparent |
|---|---|---|---|---|
| 17:03 | designblog.co.uk | **1.000** | stored | **0.0%** ❌ |
| 17:10 | seotools.co.uk | 0.9998 | stored | **0.0%** ❌ |
| 17:15 | websitepromotion (try 1) | **0** | **REFUSED** ✅ | — |
| 17:58 | gamedesign.uk | **1.000** | stored | **0.0%** ❌ |
| 18:00 | websitepromotion (try 2) | **1.000** | stored | **87.4%** ✅ |

Verified at the source objects, and for websitepromotion also at the deployed file
(`https://websitepromotion.co.uk/assets/images/logo.png`, 400×218 RGBA, 84.3% transparent).

## Why this is the finding, stated as sharply as it goes

**`border_keyed = 1.000` appears on designblog (0.0% transparent, unusable) and on websitepromotion
(87.4% transparent, correct).** Identical score, opposite outcomes. That is a complete, single-line
demonstration that the statistic is measuring the wrong thing — no threshold argument required, and
no appeal to the prompt contradiction. §3's mechanism explains it exactly: `BorderKeyed` counts
flood *membership* (`dist <= outer`), and both a ground at `d≈5` and a ground at `d≈95` are members.

**The guard is not useless — it is one-sided.** It correctly refused the 17:15 run (`border_keyed=0`,
the model painted nothing near the key). It cannot see the middle band, which is where 3 of 4 stored
artefacts landed. So: keep it, and add the border-*transparency* statistic beside it.

## Revised hit rate, and what it costs today

**1 good of 4 stored (25%), plus 1 correct refusal of 5 attempts.** Three sites are now carrying an
unusable logo that the platform believes is fine. websitepromotion's good one still shows a faint
**magenta fringe** on the mark's edges — `despill` is not fully removing key spill at the graded
edge, which is worth a look once the guard is fixed.

## What round 2 got wrong, so you can discount it correctly

Round 2 said the drift range 65→105 was stable and implied a threshold retune could not work. With
websitepromotion in the sample that is **overstated**: the model *can* land inside `inner=48`, so
the constants are not obviously wrong — the **variance** is the problem, and a guard that cannot
detect the bad tail is what makes the variance expensive. My earlier "seotools rules out the
contradiction excuse" still holds and is now redundant: the contradiction was live for the good run
too.

— the `bugfix_417_420` lane

---

# ROUND 4 — 21:00Z. The fix (`fcbe6071c`) verifies 4/4 against the real artefacts, and is NOT yet live

The 424 lane fixed this within the hour and mutation-proved it (reinstated the old computation,
watched the new test fail at 1.000, restored). `BorderKeyed` now counts border pixels whose
**final alpha is 0** rather than BFS reachability. What follows is the empirical half they flagged
as still owed — *"reasoning from your data, not yet re-verified against a real post-fix run."*

## The new statistic, replayed on four real production artefacts `[MEASURED 2026-09-02 21:00Z]`

`finalAlpha` is computed on exactly the bytes that get uploaded, so the stored PNGs **are** a
faithful input for the new statistic. Border ring = outermost pixel ring; fraction with `alpha==0`;
threshold `keyGroundMinBorderKeyed = 0.95`, unchanged by the fix:

| artefact | new `BorderKeyed` | new guard | correct? |
|---|---|---|---|
| websitepromotion.co.uk (the 87.4%-transparent success) | **0.9993** | PASS | ✅ |
| designblog.co.uk | 0.0000 | REFUSED | ✅ |
| seotools.co.uk | 0.0000 | REFUSED | ✅ |
| gamedesign.uk | 0.0000 | REFUSED | ✅ |

Both halves hold on real data, with `inner`/`outer` untouched — which also confirms round 3's
reading that this was **variance the guard could not see**, not a threshold miscalibration.

**⚠ What this is NOT, stated so nobody quotes it as more than it is:** it replays the *statistic*,
not the code path. Decode → flood → grade never ran. It can catch a discrimination regression and
nothing else; a real post-fix generation is still owed.

**⚠ The margin is now the thing to watch.** websitepromotion passes with only **3 of 4,348** border
pixels non-transparent. That is comfortable here *because the prompt forbids edge-bleed* ("the
artwork must not touch the image edges") — so that prompt clause has quietly become load-bearing
for the guard. A design that legitimately reaches the border would be refused as a matte failure.

## The fix is committed but NOT in the running adapter `[MEASURED 2026-09-02 21:00Z]`

- adapter pods: `v1.0.1355`, started **20:56:52Z**, build provenance `git_commit=0d2feee2ff61d89b3f18588cdd81b569fc2c4ee6`
- `git merge-base --is-ancestor fcbe6071c 0d2feee2f` → **NO**
- control: `git merge-base --is-ancestor 6440ec968 0d2feee2f` → **YES** (the original matting fix is
  in the same stamp, so the comparison is sound, not a broken invocation)

The roll landed ~20 minutes after the fix commit and picked up a build that predates it.

**⚠ A binary grep cannot answer this one.** The fix adds no string literal — its only added quoted
text is inside a comment — so there is no needle and no removed-string control. I nearly reported
"absent" off `borderTransparent`, which is a **local variable** and would never have been in the
binary whatever the truth was. The build-provenance stamp is the only working instrument here, and
this is a clean example of the case CLAUDE.md's "ask the service what it is running" rule exists for.

## Live damage while it waits

designblog.co.uk, seotools.co.uk and gamedesign.uk each serve `/assets/images/logo.png` at 200 and
carry `logo-img` in their served markup — i.e. **all three veiled logos are live now.** Until the
fix rolls, a regeneration reproduces the same class. Left to the 424 lane to decide; not touched.
