# CONTINUE HERE — 077 is done and live; the feature-builder round is the open thread

**Written:** 2026-07-27 ~11:10 UTC. Supersedes nothing — read
`HANDOFF_2026-07-26_detector_handler_remit_live.md` for the full 077 account; this
file is the *current state* and the *next action*.

---

## TL;DR

`bugs_closed/077` is **finished, live and fully verified** on v1.0.1171, still live
on **v1.0.1172**. Nothing is owed on it.

The live thread is the **first capability gap going through the feature builder**.
Round 1 was council-APPROVED with a medium objection worth designing out; **round 2
is in flight**:

```
FEATURE_CORR (round 2, LIVE)  1a9feed2-b436-42ef-b2c7-ee59bc50dac6
FEATURE_CORR (round 1, done)  c91bb061-250e-4a1a-819f-78c625733956   approved, 5 reviewers, unreadable 0
work item                     7b89fb35-f42c-45d1-b64d-214aff56d918   finetuning.uk, 8/8 residue
```

Check it:

```sql
SELECT status, current_step FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id'='1a9feed2-b436-42ef-b2c7-ee59bc50dac6';
SELECT kind, metadata->>'decision', metadata->>'unreadable'
  FROM diagnosis_artifacts WHERE correlation_id='1a9feed2-b436-42ef-b2c7-ee59bc50dac6';
```

## Verified on v1.0.1172 this morning (11:05Z)

| check | result |
|---|---|
| `capability gap, not a handler failure` | 1 |
| `PartitionByRemit` | 1 |
| `TWO-WORD VOCABULARY` (the corrected brief) | 1 |
| OLD summary `hardcoded hex colors in inline styles instead of CSS variables` | **0** |
| control `unresolved after %d attempts` | 1 |
| migration 222's prompt rule still in the live agent row | **PRESENT** (not re-seed clobbered) |

> **Misstep worth copying the fix of:** my first pass grepped `two-word vocabulary`
> and got **0**, and I was one step from reporting the build as missing my commit.
> The source says `TWO-WORD VOCABULARY` and `grep -F` is case-sensitive. **A pod-grep
> that returns 0 is a claim about your grep until you have proved it is a claim about
> the binary** — which is what the positive control is for, and mine was passing, so
> the contradiction was visible if I had read it.

## The next action, and the decision it needs

Round 2 asked the designer to **design out** this round-1 objection rather than let
the build gate find it (`edit-quality`, verbatim in the work item's `capability`):

> *"s1's gate: build:true claim of independent buildability is not actually
> satisfiable unless the plan specifies a backward-compatible mechanism … As
> written, s1, s2, and s3 look like they must land together to compile."*

Concretely: `ReplaceHardcodedColors` is called by the handler, by the verifier, AND
by the check's own `PartitionByRemit` — three call sites in two packages — so a
signature change in s1 breaks the tree until s3 lands, and **s1 fails its own build
gate**. The revision constraint added to the spec is *every stage must compile ALONE
against the tree as it exists before it*, and *name the mechanism in the stage goal*.

**When round 2 lands:**

1. Read the verdict **and `unreadable`**. An approval carried by seats that could not
   read the plan is not an approval. Round 1 was `5 reviewers, abstained 0,
   unreadable 0` — that is what good looks like.
2. Confirm a `fix_plan` artifact EXISTS. **Do not accept "COMPLETED" as success** —
   the very first designer run COMPLETED, at `complete_refused`, with no artifact and
   `orchestration_states.error` NULL (see `bugs_open/099`).
3. Check each stage's edits are one-file-per-stage and that the goals now state the
   backward-compatible mechanism.
4. **STOP THERE.** Firing `feature-implementer` writes code and opens a PR; the
   pipeline's gates are spec approval (owner) → design approval (council **and**
   owner) → per-stage build gates → PR merge (owner) → seed apply (owner). The owner
   has approved the *spec*, not the *plan*.

## Round 2's plan, and what it will and will NOT fix

It designed the objection out exactly as asked — **a new exported entry point, not a
signature change**, in its own words:

> *"The existing `ReplaceHardcodedColors` keeps its current signature and behaviour
> untouched, so the handler and any other existing caller keep compiling and
> behaving exactly as before. `PartitionByRemit` is extended (not replaced) to
> additionally consult `DeriveColorVariableMap` … This is the only edit to this file
> in this stage."*

(That last sentence is migration `222` visibly working — the designer now states the
rule back.) Symbols: s1 `DeriveColorVariableMap`, `ReplaceHardcodedColorsFromPalette`;
s2 two new tests; s3 **`(none)`**, correctly — a wiring-only stage that introduces no
new symbol takes an empty list under the prompt's rule 8.

**Set expectations before anyone reads a shrinking number as a fix.** Matching is
**exact-string against the site's palette**. finetuning.uk's palette really does
carry `background: #ffffff` and `background_alt: #f8f9fa` (verified on the site's own
row, not the column default — I originally cited the default, which would have been a
different claim). But the residue histogram there is:

```
#fff ×6   #f8f9fa ×5   #e0e0e0 ×3   #fef2f2 ×2   #ecfdf5 ×1   #eee ×1
```

`#fff` is **not** `#ffffff` under exact matching, so the single commonest residue
colour stays out of remit. This feature converts the `#f8f9fa` occurrences and
nothing else. **That is correct, not a shortfall** — "whether a bare `#fff` is worth
variabilising at all" is one of the three open questions the spec deliberately posed
and the plan deliberately declined to answer. The capability gap will SHRINK, not
close, and the remainder is explicitly parked on owner decisions about `#fff`,
semantic tints (`#fef2f2`, `#ecfdf5` — no palette entries exist for them) and inline
`style=""` attributes.

## ROUND 4 — the blocker is resolved. The plan is ready.

`FEATURE_CORR b5097ade-93a7-4e92-bd14-fafbdb2f2680` — **approved, 5 reviewers,
abstained 0, unreadable 0, "1 advisory objection — none high-severity"**, verdicts
**4 approve / 1 object**. Objection trend across the series: **3 → 3 → 2 → 1**.

**The 077-recurrence is fixed, and fixed the right way.** The plan now edits the
verifier in the same file and stage:

> *"`VerifyHardcodedSectionColorsResolved -> hardcodedSectionColoursVerdict ->
> PartitionByRemit` is edited IN THIS SAME FILE to call `DeriveColorVariableMap`
> using the `db` and `target.SiteID` it already holds, and thread that palette into
> `PartitionByRemit`…"*

And it went further than the directive asked: `s1`'s expected symbols include
**`TestVerifierMatchesCheckClassification`** — the "both ends agree" property turned
into a build-enforced test rather than an argument in a risks section. That is the
same shape as this repo's other lockstep guards, and it is what stops a future
thread quietly reintroducing the divergence. `s2` also gained `test=true`.

**What remains is a different and lesser class** — consequences of the DB read the
directive introduced, not the original defect:

| reviewer | sev | concern |
|---|---|---|
| bug_historian | med | check and verifier now read `color_palette` independently at different times; the plan covers the happy path but not what happens if the two reads disagree (palette edited in between) |
| guardian | med | the verify path's I/O profile changes; no caching/batching commitment |
| bug_historian | low | if `DeriveColorVariableMap` returns an EMPTY map (null `style_collection_id`, or keys absent) the handler still mutates `rendered_html` — silent no-op territory |
| reuse_agent | low | no code_pointer names where the `--color-*` emission logic lives, so `DeriveColorVariableMap` may duplicate existing mapping logic |

None is high-severity; all are the sort of thing the per-stage build gates and the
PR review surface. **The transient-disagreement one is arguably correct behaviour**
(a verifier should judge against current state), and the empty-map one is worth
holding as an explicit check when reading the PR.

**RECOMMENDATION: this is ready for the implementer, on the owner's word.** Four
rounds, converging objections, the structural defect closed and test-pinned. A
fifth round would be tidying.

**Pre-flight before firing `feature-implementer`** (from the feature-builder
workstream, not optional): `bugs_open/066` — spawned agent pods pin stale image
tags, so census the image tag before the fire; the merged PR #3 code is NOT in this
working tree (fetch first); and the run refuses a pre-existing `feat/<short-corr>`
branch, which is deliberate.

## THE EARLIER DECISION POINT (as of round 3) — superseded by round 4 above, kept for the trail

Three designer rounds, all APPROVED, all `5 reviewers / abstained 0 / unreadable 0`.
Objections: round 1 **3**, round 2 **3**, round 3 **2**. Converging, not converged.

```
round 1  c91bb061   3 objections   medium: signature change breaks stage independence
round 2  1a9feed2   3 objections   medium: SAME failure moved compile-time -> test-time
                                   medium x3: verifier call site not migrated
round 3  b604f92d   2 objections   test-coupling FIXED (s1 gate build+test, fixture flip
                                   pulled into s1); verifier STILL unresolved
```

**Round 3 fixed the test-coupling properly**: `s1 [build=true test=true]` editing
`check_hardcoded_section_colors.go` AND `check_hardcoded_section_colors_test.go` —
two files, one edit each, so still inside the `222` rule. Two stages, not three.

**What remains is one thing, and it is `bugs_closed/077` about to be re-created by
its own fix.** The plan gives `PartitionByRemit` a variadic optional palette
parameter and argues:

> *"old 1-arg call sites (including VerifyHardcodedSectionColorsResolved …) still
> compile unchanged and, receiving no palette, fall back to exactly today's
> classification, **so no divergence is introduced**"*

That defines divergence away rather than removing it. The CHECK passes a palette
and calls `#f8f9fa` within-remit; the VERIFIER passes none and calls the same
component outside-remit. Two answers about one component — which is 077's shape
exactly. The concrete failure: if the handler does NOT fix the colour, the verifier
sees the component, classifies it out-of-remit, and returns `Resolved: true`,
marking the item verified while the work is undone. It bites precisely in the case
the verifier exists for, and looks fine whenever the handler succeeds (the
component drops out of the population), so a green test will not show it.

Three reviewers said the same independently — and `edit-quality` found the part I
missed:

> *"The plan's risks section asserts 'VerifyHardcodedSectionColorsResolved is not on
> any code_pointer in this spec, so it cannot be edited directly' — but the spec's
> fifth code_pointer …"*

**The designer's stated reason for not fixing it is factually wrong.** The verifier
lives in `check_hardcoded_section_colors.go`, which is code_pointer #1 AND #5 AND
**a file s1 already edits**. So the fix is small and in-scope: have
`hardcodedSectionColoursVerdict` pass the same derived palette.

**Options, with a recommendation.**

1. **RECOMMENDED — one narrowly-scoped round 4** saying only: *the verifier IS on an
   editable path (pointer #1 and #5) and s1 already edits that file; update
   `hardcodedSectionColoursVerdict` to pass the derived palette so both ends
   classify identically. Change nothing else.* The remaining objection is a factual
   error, not a design disagreement, so this should close in one pass.
2. Accept round 3 and fire the implementer, with the verifier divergence recorded as
   a known defect to fix immediately afterwards. Cheaper now, ships a known 077
   recurrence.
3. Stop; the plan is on record and can be picked up any time.

**A fourth round was NOT fired without asking**, because this thread said it would
stop rather than iterate indefinitely, and because three rounds of "approved with
advisory objections" suggests that is the steady state rather than a failure.

## Landmines this thread paid for (all filed, none theoretical)

- **`kubectl run -i --rm | kcat -P` silently publishes NOTHING, ~4 times in 5** —
  `exit 0`, a printed correlation id, no trace. Affects shipped triggers incl.
  `097_TRIGGER_council_review_v1.sh:121` and the feature-designer trigger. Put the
  payload in the container COMMAND and `&& echo PUBLISH_OK`; working copy at
  `scripts/initial_messages/290_design_discovery/082_fire_design_discovery_any_site.sh`.
  **`--command` is load-bearing** — the image ENTRYPOINT is kcat, so without it your
  `sh -c` becomes kcat's arguments.
- **A run that FAILED a step can read `COMPLETED` / `error` NULL / `final_result`
  NULL.** The reason lives only in `collected_data->>'__step_error'`. Require the
  ARTIFACT, not the status.
- **`bugs_open/099`** — the designer's prompt never stated the validator's
  one-edit-per-file-per-stage rule, so a completed good design was discarded.
  Migration 222 fixes candidate 1 and is verified on the failing branch; **099 stays
  OPEN** for candidate 2 (route the failure into `repropose`), because a designer
  hitting any of the validator's dozen other rules still loses its design silently.
- **A superset proves zero and can NEVER disprove it.** I nearly published a
  "correction" to another thread's measurement on that error; the shipped code later
  confirmed *their* number. `WRONG_CALLS.md` 2026-07-26.
- **`detected` is not terminal**, so it holds an `idx_swi_dedup` slot and suppresses
  re-files — which is why robot-hands.com is useless as a verification control.
- **Check the chassis pod `startTime` BEFORE publishing anything that spends
  credits.** One `kubectl get pods`; the ~300s post-restart drop window is real.

## The technique that is worth reusing, not just the result

A `capability_gap` spec carrying a **measured histogram** and **named open questions
posed AS questions** produced a conservative plan; the plausible-sounding
description I originally wrote would have produced the naive one. I only found this
by measuring the residue instead of trusting my own `capability` string: every
out-of-remit colour on finetuning.uk was a LIGHT SURFACE colour, so "teach the fixer
more hex patterns" would have repainted every white card in the brand colour. The
council then validated the framing in its own words — *"the deliberate exclusion of
#eee, semantic tints, and inline style='' correctly honors the spec's 'open
questions for design, not assumptions' framing"*.

**The plan is only as good as the spec's `code_pointers`**, and those come from
whoever files the gap. Round 1's s2 — *"flip the pinned remit test cases the widening
is supposed to flip"* — exists because a pointer said those `withinRemit:false`
fixtures pin the current remit deliberately. Without it the designer would have
edited the test green.

## Still open elsewhere, not this thread's

- `bugs_open/099` candidate 2 (above).
- The other capability gaps as builder intake: `forced-text-color-fixer`
  (`handler_missing`, action already registered — a seed plus a remit decision, and
  **seeding it without partitioning its check re-creates 077**) and
  `site-metadata-fixer` (`handler_missing`, no action exists — a real build).
  Group them: `SELECT spec->>'builder_needed', count(*) FROM site_work_items
  WHERE item_type='capability_gap' AND status='deferred' GROUP BY 1;`
- `bugs_open/083` — 98 rows stuck at `detected`. 077 cleared its blocker for the
  four checks fixed here **only**; the other ~18 item types are untouched.
