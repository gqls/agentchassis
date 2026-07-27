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
