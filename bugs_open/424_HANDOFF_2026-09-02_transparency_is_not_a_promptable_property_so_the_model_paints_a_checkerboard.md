# 422 — transparency is NOT a promptable property, so the model paints the PICTURE of transparency: a grey-and-white checkerboard, served as the logo

**Filed 2026-09-02 by the `bugfix_417_420` lane. Artefact evidence and the diagnosis are the
`site_delivery_and_editor` lane's** (they downloaded, looked, and ran the chunk scan); **the code
evidence that it is a capability gap rather than a prompt defect is this lane's.** Diagnosis loop
NOT run — substituted equivalent first-hand verification per the 2026-07-31 ruling: every claim
below is a read of the served file, a live row, or the code, each quoted.

**Filed as its own file rather than folded into 417 or 421 — deliberately. Third mechanism, third
layer, third fix site:**
- `bugs_open/417` — the prompt **PERMITTED** something it should not (input licence). Fixed.
- `bugs_open/421` — the returned file's **STRUCTURE** was never examined (output acceptance). Open.
- **this** — the pipeline **CANNOT PRODUCE** the property being asked for (capability gap). No
  prompt and no config change can close it.

## What happened

OWNER RULING 2026-09-02: *"the background behind a logo shouldn't be part of the logo"* → the
regeneration prompt asked for *"a fully TRANSPARENT background (PNG alpha), no ground colour, no
panel, no backdrop of any kind"*.

The model returned an image of a **grey-and-white checkerboard** behind the mark — the
**UI REPRESENTATION** of transparency, painted as pixels.

Served asset, verified by the delivery lane: sha256 `1abcf69c08ab4462`, 243,080 B, **PNG colour
type 2 (RGB)**, and a chunk scan showing **no `tRNS` chunk**. There is no alpha mechanism in the
file at all, so the checkerboard is paint, not a viewer artefact. (Their note on how nearly this
was misread is worth keeping: a viewer draws a checkerboard *for* real alpha, so the visual is
ambiguous by construction — **only the chunk scan settles it**.)

## Why it is a capability gap, not a prompt defect `[MEASURED 2026-09-02, this lane]`

Alpha is a **file-format capability**, negotiated with the provider or added by post-processing.
It is not a property a model can be asked for in prose. Confirmed across the codebase:

- `internal/adapters/imagegenerator/banana/provider.go` — **no** `output_format`, no alpha, no RGBA
  and no mime negotiation anywhere; `MimeType` is only ever read back from what arrived.
- **No background-removal or post-processing step exists** anywhere in `platform/` or `internal/`
  (`background_remov|remove_background|rembg|alpha_channel|make_transparent` → zero hits).
- `kindDefaults["logo"]` carries only a negative prompt (*"…complex background"*) — **there is no
  ground/background knob** to set.

**So the owner's ruling cannot be satisfied by any prompt revision, and further regeneration
attempts asking for transparency will keep returning pictures of transparency.** This needs
imagery-pipeline work: request an alpha-capable output format from the provider, or add a
background-removal step after generation.

## The interim problem, and why there is no clean revert

The site now serves the checkerboard logo on every page — **visibly worse** than the 08-31 mark it
replaced, which had a baked dark ground and was invisible-safe against the dark header.

`[MEASURED 2026-09-02]` **There is no revert seam for a generated asset:**
- the `assets` table keeps **no version history** — the store path UPSERTs, so a regeneration
  overwrites `origin_prompt`/`url` in place (see the LANDMINES entry: `created_at` still reads the
  original generation);
- `retract_asset_files` **deletes** orphan files; it does not restore a prior version;
- nothing under `site-snapshots-and-revert` covers assets.

So the delivery lane holds the 08-31 bytes as a banked baseline with **no sanctioned path to put
them back**, and was right to refuse a hand-upload (owner ruling 2026-08-04: everything goes
through the framework).

**The framework-legal interim is a regeneration naming an explicit SOLID GROUND COLOUR matching the
header** — which *is* promptable, and is proven achievable because the 08-31 generation produced
exactly that. It is not a third blind attempt: we now know what to ask for and what not to.

## Fix candidates (ordered by what closes the door)

1. **Alpha at the provider** — request an alpha-capable format and carry it through the adapter,
   the store path (`mime_type`) and the deploy. Closes the class: "transparent" becomes a thing the
   pipeline *does*, not a thing it *asks for*.
2. **A post-generation background-removal step** for `kind=logo`. Independent of provider support;
   costs a processing step and a new failure mode (a mark clipped by an over-eager matte).
3. **Interim only — an explicit ground-colour clause** for logos, sourced from the site's header
   colour rather than left to the model. Does not satisfy the ruling; makes the artefact usable
   while 1 or 2 is built. ⚠ Note `composedPaletteDirection` deliberately **excludes logos** from
   palette direction (the 2026-05-20 contamination lesson), so this must be a narrow ground-colour
   clause, not a re-opening of that exclusion.
4. **A revert seam for generated assets** — the absence is a separate defect this incident exposed.
   Worth filing on its own if 1–3 are taken elsewhere.

## Verify
- The served file, not the row: `curl -s <logo url> | sha256sum`, then a PNG chunk scan for
  `tRNS` / colour type. **A visual check alone cannot distinguish painted checkerboard from real
  alpha** — that is this bug's whole trap.
- After any fix: colour type 6 (RGBA) or a `tRNS` chunk present, AND the mark still legible at
  favicon size.

## Related
- `bugs_open/417` (logo text policy — text-free HELD on this same asset, verified by eye).
- `bugs_open/421` (multi-panel design comp — single-composition HELD on this same asset).
- `bugs_open/235`, `bugs_open/322` (brand-asset handling).
- LANDMINES: the asset-upsert `created_at` trap, which bit this verification and was routed around.

## 2026-09-02 — Fix candidate 1 (alpha at provider) REFUTED externally; candidate 2 implemented as keyed-ground matting, code complete, not yet built/rolled

Picked up by a separate thread after the `bugfix_417_420`/webdesign/`site_delivery_and_editor`
lanes had confirmed the diagnosis and shipped the interim solid-ground regeneration. All three were
notified before and during this work; no collision (see NOTES cross-references in
`docs024_key_docs_latest/bugfix_424_logo_transparency/`).

**Candidate 1 (alpha at the provider) is now externally confirmed dead, not merely internally
refuted.** External research, 2026-09-02: Gemini's whole image family (every Nano Banana variant,
2.5 and 3.x) has no alpha-channel output at all — a structural model limitation, not a missing
adapter knob. No provider-side fix is possible.

**Shipped design: keyed-ground matting**, registered as **IMG-076**
(`docs026_concept_register/register/imagery.md`), full detail in
`docs024_key_docs_latest/bugfix_424_logo_transparency/PLAN_2026-09-02_logo_background_transparency.md`.
In one sentence: ask for a fixed, deterministic, non-brand-derived key colour (`#FF00FF`) instead of
"transparent" — provable achievable, since the interim regen already showed the model can paint a
requested flat colour — then remove it mathematically in pure Go (border flood-fill, so only
background *connected to the image border* is erased, protecting an interior element that merely
resembles the key), and refuse to store the result if the model didn't honour the key colour
(`BorderKeyed < 0.95` — fail closed, load-bearing because there is still no revert seam for a
generated asset).

**Code complete, unit-tested (17 new/changed tests across three packages, all passing), NOT yet
built into an image or rolled to any pod.** Touches:
`platform/orchestration/actions/discovery_checks/default_brand_prompt.go` (the clause constants),
`platform/orchestration/actions/generate_image_actions.go` (`applyLogoBackgroundPolicy`, wired into
the SAME choke point as 417's text policy — deliberately, so the clause stays visible in
`assets.origin_prompt`), `internal/adapters/imagegenerator/dynamic_adapter.go` (`KeyGround` field,
the fail-closed guard), `internal/adapters/imagegenerator/keyground.go` (new — `KeyOutBackground`).

**Does NOT fix `bugs_open/421`** — recorded explicitly so a future matting pass over a two-panel
comp is not mistaken for having cleared it; see IMG-076's own "does NOT fix" note.

**Still open before this closes:** council submission (owed alongside the commit — not yet run at
the time of this entry); image build + roll; a real magenta-keyed generation to set the
`[UNMEASURED]` threshold constants from evidence instead of an extrapolated drift figure; the
served-PNG chunk-scan re-run at the correct slug (`boxingonline.ugg2.com`, not the parked `.com`
catch-all — caught by the boxingonline session before it became a wrong verification) checking BOTH
colour type 6 and `tRNS`, never one alone.

Separately filed, not part of this fix: `assets.mime_type` is empty on 910/1,277 rows fleet-wide
(measured by the boxingonline session, 2026-09-02) — a store-step defect this bug's own verification
surfaced but does not cause.

## 2026-09-02 (contd) — council APPROVED, but caught a real bug; fixed; NOT yet in the deployed build

Council review (`d018a48f-bd76-420a-8530-4491681d3bd4`): **APPROVED with 4 advisory objections,
none high-severity.** One objection (editquality MEDIUM) was a confirmed real defect: the negative
prompt forbade the exact colour (`magenta`, `#ff00ff`) the clause simultaneously told the model to
paint as the background — a contradiction that could make the model refuse the key colour entirely,
defeating the mechanism. Fixed same session (`b2322a203`): the foreground/background distinction
now lives in one sentence inside the positive clause instead of a separate, contradicting
negative-prompt term. Full detail, plus the other three (non-code-changing) findings:
`docs024_key_docs_latest/bugfix_424_logo_transparency/NOTES_logo_transparency.md`.

**A fresh chassis build was deployed 2026-09-02 (tag v1.0.1354) — verified at the artefact (build
provenance + binary probe with positive/negative controls, both services) to carry the ORIGINAL
matting fix but NOT the magenta-contradiction fix**, which was committed after that build's pods
started. **Do not trigger a real `kind=logo` generation against the currently-running build** — see
RUNBOOK for the exact commands to re-verify before testing.

Detailed handoff for a fresh session: `docs024_key_docs_latest/bugfix_424_logo_transparency/
HANDOFF_2026-09-02_continue_here.md`.

---

## 2026-09-02 (contd) — the matte RAN, and the fail-closed guard scored it 1.000 on an artefact with ZERO transparent pixels (contributed by the `bugfix_417_420` lane)

First production run: `designblog.co.uk`, 17:03:23Z, triggered by the **queue**, not by a person
(three `needs_imagery:site:-:logo` items were dispatched automatically at 16:10–16:15Z — the
"do not trigger a real logo generation" warning above does not bind the fleet).

Adapter's own log: `source_format":"jpeg","key_hex":"#FF00FF","border_keyed":1,"pixels_keyed":978631`.
Stored artefact: PNG 1408×768 RGBA, **alpha extrema (57,255), 0.0% fully transparent pixels**, and
of the 4,348 border-ring pixels **0 keyed out, 0 opaque, 4,348 graded**. The ground is veiled at
~35–50% opacity, not keyed.

**The structural finding — independent of prompt, build and thresholds:** `stats.BorderKeyed`
counts border-flood MEMBERSHIP (`dist <= outer`, `keyground.go:104,131,149`), but a pixel only
becomes transparent at `dist <= inner` (`keyground.go:176`). A ground at `d = 109` therefore scores
`BorderKeyed = 1.000` and comes out 98% opaque. **The fail-closed guard at `dynamic_adapter.go:683`
cannot observe the failure mode it exists to catch.** Fix shape: add a border-*transparency*
statistic (final alpha 0) and gate on that; keep `BorderKeyed` as the "was there a ground at all"
signal.

**UPDATED 17:20Z — three runs now, and the second one rules out the obvious excuse.**
`seotools.co.uk` (17:10Z) scored `border_keyed=0.9998` on a ground that is **visibly, plainly
magenta** — the model obeyed the key colour — and its artefact still has **0.0% fully transparent
pixels** (alpha extrema (137,255); 0 of 4,348 border pixels keyed out). The 17:15Z run scored `0`
and was correctly refused. **So the guard's score is anti-correlated with success: it refuses what
it can see failing and passes what it cannot.** Both runs predate `b2322a203`, but the
contradiction cannot explain seotools — its ground IS magenta.

**Measured ground drift from `#FF00FF`, recovered per border pixel:** designblog min 65.7 / mean
73.5 / max 95.2; seotools min 86.2 / mean **94.0** / max **105.1**. With `inner=48` nothing reaches
alpha 0, and seotools' max sits **4.9 units under `outer=110`** — a knife edge. The drift range
(65→105) spans almost the whole band (48→110), so **raising `inner` alone cannot work**: an `inner`
high enough to key seotools is essentially `outer`, leaving no graded edge. Both constants must
move, `outer` further — or the matte needs a hue-based distance rather than Euclidean RGB, which is
what JPEG chroma subsampling degrades most on a saturated key.

Full evidence, the recovered per-pixel distances, and the 12-of-12 JPEG census:
`docs024_key_docs_latest/bugfix_424_logo_transparency/CONTRIB_2026-09-02_from_417_lane_the_matte_ran_for_the_first_time_and_the_guard_scored_it_1000_on_an_artefact_with_zero_transparent_pixels.md`

**UPDATED 19:45Z — five runs, and the sharpest form of the finding: `border_keyed=1.000` appears on
BOTH a failure and a success.** designblog 17:03 scored 1.000 → 0.0% transparent (unusable);
websitepromotion 18:00 scored 1.000 → 87.4% transparent (correct). Identical guard score, opposite
outcomes, same chassis build (`v1.0.1354` — re-probed with a control pair; `b2322a203` still NOT
deployed), same `key_hex`, same contradicted prompt. **The matte is not uniformly broken — it is
high-variance, and the guard is blind to the bad tail:** 1 good of 4 stored, plus 1 correct refusal
(`border_keyed=0`, 17:15) of 5 attempts. Three sites now carry an unusable logo the platform
believes is fine. Round 2's "the drift range is stable" is thereby **overstated** — the model can
land inside `inner`; it is the variance, not the constants, that the guard fails to police.

## 2026-09-02 21:xx — ACTIVE PRODUCTION INCIDENT: three live sites currently serve a logo the platform believes is transparent and isn't; fix for the guard is written, tested, and council-APPROVED

**Confirmed at the DB, 2026-09-02: `designblog.co.uk`, `seotools.co.uk` and `gamedesign.uk` all
still carry their broken (0.0% actually transparent, 90%+ opaque) logo asset right now.** Nobody
has touched them since the CONTRIB's runs. `websitepromotion.co.uk` carries the one good result
(87.4% transparent, though with a minor despill fringe — separate, unfixed).

**This is not a theoretical risk from a manual test — the fleet's own `needs_imagery:site:-:logo`
queue triggered every one of these runs automatically**, hours before anyone read the CONTRIB. The
earlier "do not trigger a real logo generation against the current build" warning in this file
protected against a person testing it; it did nothing against the autonomous queue, which does not
read handoffs.

**Root cause of the false pass, verified against the code (not taken on the reporting session's
word alone):** `MatteStats.BorderKeyed` (`keyground.go`) was computed from BFS flood-fill
reachability (`dist <= outer`) — "was this border pixel close enough to be eligible for keying" —
not from whether the pixel actually ended up transparent (`dist <= inner`). A ground that landed
anywhere in the wide graded band between the two thresholds scored `BorderKeyed≈1.000` — identical
to a real success — while remaining ~90%+ opaque. The fail-closed guard this statistic exists to
drive (`dynamic_adapter.go`) was therefore fed the wrong number and could not tell a real failure
from a real success; it is why three sites now have a logo the platform certified as fine.

**Fixed and shipped through review same day:** `keyground.go` now tracks each pixel's actual final
alpha through the existing grading pass and computes `BorderKeyed` from that (`alpha==0`), matching
what its own doc comment always claimed it measured. New regression test
`TestKeyOutBackground_GradedBorderIsNotBorderKeyed` reproduces the live-reported shape exactly and
is mutation-proven (confirmed to fail against the reinstated pre-fix computation before confirming
it passes against the fix — the reinstatement was never committed, done via a scratch-backup +
Edit revert/restore since `git stash` is banned on this tree). Commit `fcbe6071c`. Council review
`52bd50a1-3783-4801-868a-31a0ee599e60`: **APPROVED, all reviewers, no objections.**

**Not yet deployed** — needs a roll, same as round 1's fix. **Once it is: the three broken assets
do not self-heal.** They are already stored and their work items are `complete`, not `triaged` —
nothing will re-run them automatically. Someone needs to deliberately reset those three sites'
`needs_imagery:site:-:logo` items after the roll (not before — retrying against the still-unfixed
build would just add a fourth bad result). This session has NOT done that reset itself.

**Reassuring finding buried in the CONTRIB's own round-3 correction, worth keeping**: the drift
range is not as bad as round 2 first suggested — `websitepromotion`'s good run proves the model
CAN land inside `inner=48` — so this looks like a fixable-by-the-guard-alone variance problem, not
necessarily one that also needs `inner`/`outer` retuned. Not yet re-verified against a real
post-fix generation; treat as a working hypothesis, not a settled fact.

Full detail: `docs024_key_docs_latest/bugfix_424_logo_transparency/NOTES_logo_transparency.md` and
the updated `HANDOFF_2026-09-02_continue_here.md`.
Full table in the CONTRIB.

**UPDATED 21:00Z — `fcbe6071c` verifies 4/4 on real artefacts, and is NOT yet live (417 lane).**
Replaying the new `BorderKeyed` (border pixels with final alpha 0) against the four stored
production artefacts, threshold 0.95 unchanged: websitepromotion **0.9993 PASS**; designblog,
seotools, gamedesign **0.0000 REFUSED**. Both halves correct, `inner`/`outer` untouched — so this
was variance the guard could not see, not a threshold problem. ⚠ It replays the STATISTIC, not the
code path; a real post-fix generation is still owed. ⚠ Margin: websitepromotion passes on 3
non-transparent border pixels of 4,348, so the prompt's "artwork must not touch the image edges"
clause is now load-bearing for the guard.
**Not live:** adapter `v1.0.1355` (started 20:56:52Z) stamps `0d2feee2f`, and
`git merge-base --is-ancestor fcbe6071c 0d2feee2f` is NO, while the control `6440ec968` is YES.
No string-literal needle exists for this fix, so only the provenance stamp can answer it.
All three veiled logos are serving live meanwhile. Detail: the CONTRIB, round 4.
