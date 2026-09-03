# PLAN 2026-09-03 — bugs_open/361: make the render-check ratchet component-scoped

## The problem, in one line

`component-render-check` has been RED for **25 consecutive days** (`lastSuccessfulTime =
2026-08-09T06:55:21Z`, read from the live CronJob 2026-09-03) because its ratchet is a
**key-level set difference against a component-level-blind baseline** — so every component
born after the baseline was cut manufactures "NEW" findings that are not regressions.

## Why this lane exists

`bugs_open/361` filed 2026-08-22, **OPEN, UNOWNED**. Originating lane
(`bugfix_140_contact_info_fabrication`) last commit **2026-08-07** — quiet 27 days.
`scripts/who-owns.py 361` and a re-read of the tree confirm nobody is working it.

## What I measured first-hand before designing anything `[MEASURED 2026-09-03]`

Instrument: dumped the live `content_components` (497 active) to a JSON fixture through
`kubectl exec … | gzip | base64` (the plain stream truncates — 6.5 MB, the documented flake),
then ran the tool offline with `--json <fixture> --compare`.

- The offline run reports **478 NEW**, which **matches this morning's cluster run exactly**
  (doc_notes 2026-09-03). So the instrument agrees with the live job.
- The bug has grown since filing: NEW **227 → 478**, active components **282 → 497**.
- Splitting today's 478 NEW by whether the finding's component owns ANY baseline key:
  - **18 findings across 5 components** are in components the baseline KNOWS — real
    regression candidates (`blog-listing_pre_037`, `social_proof`,
    `tool-ab-test-calculator_pre_037`, `tool-equity-release_pre_037`,
    `tool-gas-unit-converter-gaswholesalers-com`).
  - **460 findings across 62 components** are in components with ZERO baseline keys —
    unbaselined growth, not regressions.

## The hole in the bug file's own fix candidate 1 — found here, corrects the bug file

361 §4 candidate 1 says *"a finding whose component owns **zero keys** in the baseline is not
NEW — it is unbaselined"*. **That is not sufficient, and the artefact proves it:**
`baseline.json`'s own note reads *"1023 findings across **139 analysed components**"*, while
its keys span only **115 distinct components**. So **24 components were analysed and CLEAN**
at baseline-cut time.

A keys-derived "covered set" cannot see those 24. Consequence: **a component that was analysed
and clean at baseline time and later REGRESSES would be classified `unbaselined` and would NOT
fail the run** — precisely the regression a ratchet exists to catch.

`[MEASURED]` 0 of them regress today, so the hole is **latent, not live damage**.

> **CORRECTED 2026-09-03, same day, by a Fable review — two claims in the sentence that stood
> here were wrong, and both flattered the design:**
> 1. ~~"analysed-and-clean is the healthy majority"~~ — **false at cut time: 24 clean of 139
>    analysed, i.e. 17%.** The design conclusion is unaffected (the hole is real at any width),
>    but the figure was invented rather than measured, in a document whose whole argument is
>    that figures must be measured.
> 2. ~~"grows every time the baseline is regenerated"~~ — true only of **candidate 1** (a
>    keys-derived covered set). With coverage actually recorded, **regeneration CLOSES it.**
>    Left as written, it read as though the shipped fix carries a growing hole.
>
> **And the review found a SECOND blind spot of the same shape that this plan missed entirely:
> STATIC templates.** A template with no actions is skipped before `checked++`, so a covered set
> derived from "analysed" cannot see one — and a static component later rewritten to render a
> hole is precisely this check's stated signal. That population is **27 today, 37 at cut** —
> *larger* than the 24 this fix was designed around. Fixed in `d716c837a`; coverage is now
> collected before the static skip. See NOTES.

**So the fix is not "scope the ratchet by component" — it is "make the baseline record what it
COVERED, not just what it FOUND".** That is what makes the bad state unrepresentable, which is
CLAUDE.md's ordering rule for fix candidates.

## Design

1. `baselineFile` gains `components []string` — the analysed-component set (in CANONICAL
   names) at write time.
2. `writeBaseline` populates it.
3. `loadBaseline` returns the key set, the covered set, and whether the baseline is legacy.
4. **Legacy fallback, loud:** a baseline with no `components` field (today's embedded one)
   derives its covered set from the keys AND says so in stdout and in the `doc_notes` body,
   naming the exact blind spot. Keeps the tool working with today's artefact; the hole closes
   the moment anyone regenerates.
5. The ratchet becomes three buckets:
   - **regression** — component IS covered, key is not in the baseline → **fails the run**
   - **unbaselined** — component is NOT covered → own bucket, own count, **does not fail**
   - **uncovered** — unchanged, still fails
6. Summary line and `doc_notes` body carry all three counts, so the unbaselined debt stays
   visible daily even though it no longer fails the run.

## The cost, stated rather than hidden

A brand-new component that renders holes will **not** fail this job. That debt belongs to
birth-time gating (CGV-029 / the component birth path), not to a regression ratchet. It is
reported every day with its count so it cannot go quiet.

## Verification — both arms, by mutation (361 §6)

A fix verified only on the "unbaselined does not fail" arm is a fix that turned the check off.
Both arms are proved:
- a **baselined** component that gains a key → exit **1**
- an **unbaselined** component that produces a key → exit **0**

## Sequencing note

`component-render-check` is in `RELEASE_IMAGES` (makefile:95), so it rebuilds from committed
HEAD in a fleet release. A chassis build was announced for ~2026-09-03 13:00 local, so this
lands committed before it to ride that roll.
