# 106 — the concept register has no staleness detector, and 67% of workstreams postdate its freeze

**Filed** 2026-07-27 from the oufe.com workstream, as a contribution to the
concept-register workstream rather than a competing effort.
**Severity** medium — nothing is broken. The register is accurate about what it
covers; the defect is that it silently stopped covering new work, and it is the
instrument sessions are told to consult before concluding something does not
exist.
**Status** OPEN.

## Measurement

Extraction froze **2026-07-13** (`register/000_concept_index.md`: 1,633 concepts,
107 category files, ~4,111 source documents).

```bash
for d in docs/agent_docs/docs024_key_docs_latest/*/; do
  git log --reverse --format=%ad --date=short -- "$d" | head -1
done | awk '$1>"2026-07-13"' | wc -l
```

2026-07-27: **51 of 76** workstream directories were created after the freeze —
**67%**. Among them, entire subsystems: `claims_verification`, `experience_loop`,
`cta_link_integrity`, `work_item_completion_integrity`, `news_feed_pooling`,
`dispatch_queue_serialisation`, `gauntlet_dead_cta`, `model_directory_pipeline`,
`durable_write_guard`, `chassis_replica_scaling`, and around forty more.

Until today, `grep -rl evidence_base register/` returned **nothing**. The entire
claims-verification layer — V0 through V5, the evidence register, the banned-claim
scanner, the citation verifier — had no entry. (Added as
`register/claims-verification.md`, 12 concepts, 2026-07-27.)

## This is a known, recurring class — patched twice by luck

The index records both previous instances in its own words:

> **2026-07-16 addition:** a coordination pass with the fixloop workstream … found
> a genuine gap — that whole subsystem shipped after extraction froze on
> 2026-07-13, so none of it was in the register.

and a second for MDL-038/039 on 2026-07-17. Both were found because somebody
happened to be working next to the hole. Claims-verification is the third, found
the same way, eleven days later and much larger.

**Three instances of one failure mode, each detected by coincidence, is the
signature of a missing detector rather than three unlucky events.**

## Why this matters beyond tidiness

The register is the estate's index of *design artefacts*, and design artefacts are
where **dormant capability** lives — a field that is declared and never read, a
decision deferred with a trigger nobody watches, a precedent in a sibling
subsystem. None of those leave a trail in code: an unused field has no call sites,
a deferred spec question is in neither `bugs_open/` nor `features_open/`, a
disabled sweep is runtime-indistinguishable from a check never built.

So code search is systematically biased toward what is *running* and blind to what
is *available but dormant* — and the register is precisely the instrument meant to
close that gap.

On 2026-07-26 a session concluded "nothing in the estate looks for this", wrote it
into a live council seat's standing instructions, and was about to build a
redundant subsystem. Every one of the four things it missed was a design artefact
of a subsystem **absent from the register**. The register could not have helped,
because it had a hole exactly where it was needed. Full account in
`WRONG_CALLS.md` (2026-07-26 entry and its 2026-07-27 addendum).

## Fix candidates, ordered by what closes the door

1. **A coverage sensor, on the model of `verifier_coverage_test.go`.** That test
   already solves this shape for a different registry: a **SENSOR** that scans
   source for every `ItemType` literal, and a **RATCHET** — a hand-refreshed list
   of what is known — with the test failing when they diverge. The register
   equivalent enumerates observable subsystems (workstream directories under
   `docs024_key_docs_latest/`, active `agent_definitions.type`, registered action
   names, `sql_for_agents/NNN_*.sql`) and reports any not referenced by a register
   entry. Makes the drift **visible on a cadence instead of by coincidence**, and
   deliberately reports rather than blocks — breadth of uncovered work is a
   backlog, not an error.
2. **A freshness stamp per register file** — `covers-through: <date>` — so a
   reader can see at a glance that an entry predates the subsystem they are asking
   about. Cheap, honest, and it fixes the *misleading* half of the problem without
   fixing the coverage half. A register that says "I stopped looking on 13 July" is
   far safer than one that silently implies completeness.
3. **Periodic re-extraction.** Thorough and expensive; it repeats the whole
   stage-1/stage-2 cost and would go stale again the following week. Worst
   ratio of the three unless 1 shows the backlog is genuinely large.

Recommend 2 immediately (minutes, and it removes the false impression of
completeness) and 1 as the durable fix.

## The general form, which is the same bug as `bugs_open/104`

**A freeze with no watcher becomes permanent, exactly as a deferral with no
watcher becomes policy.** 104 is a decision that said "revisit at two sites" and
was never revisited at eight. This is an extraction that said "as of 13 July" and
was never revisited at fifty-one workstreams. In both cases the original
judgement was correct when made, nothing was wrong with the work, and the damage
came entirely from there being no mechanism to notice the precondition had moved.

Everything else this platform relies on — cooldowns, staleness sweeps, claim
timeouts, citation re-verification — has a watcher. Deferred decisions and frozen
indexes are the two classes that do not.

## How to verify a fix

Create a workstream directory with a plan document, and add an agent type, without
touching the register. The coverage report must name both as uncovered on its next
run. Then add a register entry for one of them and confirm it drops off the report
while the other remains — **induce the gap**, because a report that is green on a
register somebody has just hand-patched proves only that the patch happened.

## Post-roll triage 2026-07-27 (~15:55 UTC) — BOTH recommended candidates landed today; one is complete, one is half

Filed this morning; both recommendations were implemented within hours, by the
concept-register side, in two commits:

```
c542c3501  2026-07-27 12:13 UTC  docs(concept-register): add the missing claims-verification entry; file 106 …
7272d59d4  2026-07-27 12:33 UTC  feat(concept-register): covers-through stamps on all 108 files + a coverage sensor
```

**Candidate 2 — DONE, verified.** `109` of `109` files in
`docs026_concept_register/register/` now carry a stamp, e.g.
`> **covers-through: 2026-07-13** · extraction freeze.` The *misleading* half of this
bug is closed: a reader can no longer mistake the register for complete.

**Candidate 1 — BUILT and already earning its keep, but NOT on a cadence.**
`102_CHECK_register_coverage.py` + `102_coverage_ratchet.txt` exist and run.
Executed live during this sweep:

```
register files      : 108      workstreams on disk : 78
post-freeze (2026-07-13) : 53   uncovered : 43   (ratchet accepts 41)
2 NEW since the ratchet:
  2026-07-27  bugfix_066_spawn_image_tag  ← post-freeze
  2026-07-27  gemini_content_provider     ← post-freeze
```

**That output is itself the induced-gap verification this file asks for**, and better
than a synthetic one: two genuinely new workstreams appeared *within hours* of the
ratchet being set, and the sensor named both. The sensor/ratchet shape works.

> **[The remaining half, and it is the load-bearing half.]** This file's candidate 1
> says the point is to make drift visible **"on a cadence instead of by
> coincidence"**. The cadence does not exist. `grep -rn "102_CHECK_register_coverage"`
> across the repo returns hits in exactly two places — the register's own
> `RUNBOOK_concept_register.md` and the ratchet file's own header comment. It is
> **not** in `.githooks/`, **not** in `scripts/pattern-check.py`, **not** in
> `.github/`, and **not** a `scheduled_tasks` row. So the sensor runs when a human
> remembers to run it — which is the same "detected by coincidence" mechanism this
> bug was filed about, moved one step earlier. Three coincidental detections is what
> made this a bug; a fourth tool that must be invoked by coincidence does not
> retire it.

**Therefore: 106 stays OPEN, but its remaining scope is small and specific** — wire
`102_CHECK_register_coverage.py` to something that runs without being remembered.
The natural home is `scripts/pattern-check.py` (advisory, already runs on the commit
path, already carries the `check_append_only_docs` precedent that was added for
exactly this "a check worth automating" reason). Report-don't-block is already the
script's own stated design, so it fits without argument. Estimated **under an hour**,
docs/scripts only, no council round, no image window.

**Separately, on the R2 question the concept-register workstream expects "~07-27"
(today): it is due by the calendar and NOT gradable.** Ran the named gating report:

```
$ ./docs/…/fixloop_eg_dartsonline/101_REPORT_mission_review_findings.sh 7
── mission-review findings, last 7 day(s) ──   findings: 2   sites: 2
── classifier runs in the window (denominator for the objection rate) ──   0
```

Two findings, both dated 2026-07-25, and a **denominator of zero** — the classifier
did not run at all in the window. R2's own instruction is to *"hand-grade a sample of
the findings for false positives first"*, and a sample of 2 against 0 runs cannot
establish an objection rate in either direction. **Waiting another week does not fix
this**; the lane has to be producing classifier runs before the promotion decision is
answerable. That is a finding for the owner, not a verdict: R2 should be recorded as
**blocked on an empty denominator**, not as "not yet due". (R2 belongs to the
concept-register workstream, not to this bug — noted here only because this sweep was
asked to establish whether the verdict was due or done.)

## Related

- `bugs_open/104` — the same watcherless-precondition shape, in the claims layer.
- `bugs_open/105` — `EvidenceFact.Kind`, an example of the dormant capability this
  register exists to make findable.
- `docs026_concept_register/register/claims-verification.md` — the entry added
  today; 12 concepts, none from the original extraction.
- `docs026_concept_register/PLAN_2026-07-20_direction_reach_and_drift_guard.md` —
  a drift guard already exists for *direction documents* (constitution/mission,
  `100_CHECK_direction_integrity.py`). This is the same idea applied to register
  coverage, and that plan is the precedent for how such a guard is introduced here.

---

# CLOSED 2026-07-28 — the sensor now runs without being remembered

**Fixed by** the `bugfix_106_register_coverage_cadence` thread. Docs/scripts only:
no image, no migration, no council round (out of the gate's `platform|internal|pkg`
scope). Working docs:
`docs024_key_docs_latest/bugfix_106_register_coverage_cadence/`.

## What was left, and what shipped

This file's own post-roll triage narrowed the remaining scope to one thing, and
was right:

> *"A fourth tool that must be invoked by coincidence does not retire it."*

Shipped: **`check_register_coverage`**, the 9th check in
`scripts/pattern-check.py` — advisory, already run by `.githooks/pre-commit`.
Registered as **OPP-004**.

**Trigger: a commit that CREATES a workstream directory the register has never
heard of.** Chosen over a cron deliberately — a cron reports drift up to a week
late, to nobody in particular; this reports it the instant the gap appears, to the
one person who can close it in ten seconds. The message names both silencing
routes (register entry, or the ratchet).

Three properties that keep it off the wallpaper pile:

- **Only NEW directories fire.** The 43 uncovered workstreams on the ratchet are
  accepted backlog; flagging active work on them every commit is how a check dies.
- **It imports the sensor rather than reimplementing `is_covered()`.** One matching
  rule, one implementation — two hand-maintained copies is the `idx_swi_dedup` ↔
  `workItemTerminalStatuses` drift class. Guarded: if the sensor moves, the check
  returns silently rather than breaking commits.
- **Advisory, never blocks**, consistent with the file's stated design.

## Measured before inclusion, per `pattern-check.py`'s own bar

```
1,500 commits scanned    fires: 4    rate: 0.27%    false positives: 0
```

Quieter than every existing check (README 0.7%, SUMMARY 2.0%, twin ~2%), which is
correct for a population of "commits creating a brand-new workstream".

**A very low rate and a dead check look identical from the number**, so all four
fires were inspected: `memory_index`, `bugs_sweep_2026_07`,
`bugfix_066_spawn_image_tag`, `gemini_content_provider`. All genuine — and the
last two are **exactly the pair this file's triage records the sensor finding by
hand on 2026-07-27**. Same gaps, now caught at creation rather than days later by
someone who happened to run the tool. That is the closest thing to a controlled
comparison this bug could have.

## Verified by inducing the gap, as this file demands

> *"induce the gap, because a report that is green on a register somebody has just
> hand-patched proves only that the patch happened."*

| arm | setup | result |
|---|---|---|
| 1 | two new uncovered workstreams staged | **both fire** |
| 2 | one added to `102_coverage_ratchet.txt` | **only the other fires** |
| 3 | the other given a register entry instead | **it goes quiet; the first still fires** |

Negative control: silent, 40 ms, on a commit touching no workstream directory.

**And it was demonstrated on itself.** Creating this fix's own workstream directory
tripped the new check; adding OPP-004 to the register silenced it. The full
intended loop, exercised on the commit that shipped it.

## Residual — recorded, deliberately not fixed

**The register can be complete in coverage and stale in CONTENT, and nothing
detects that.** Two live instances hit the same day:

1. `SCH-012` carried `verify-later: … (should still be false unless deliberately
   turned on)`. It had been **true** for weeks, and that stale expectation is part
   of why `bugs_closed/124` went unnoticed. **A `verify-later` that states an
   expected answer rather than a question reads as reassurance, and nobody re-runs
   it.**
2. The `psql -t -A` command-tag trap had been found and fixed by two threads, each
   privately in its own script comment, neither anywhere findable — so a third
   shipped it into a claim guard (016b §9).

Not folded in, on purpose: the sensor asks only whether a subsystem is
*represented*, never whether an entry is *accurate*. Those are different questions
and conflating them is how a coverage check becomes an audit nobody runs — this
file says so itself. **Sample of two; not filing a bug on it.** If a third instance
appears, that is the signature this file taught us to read.

**Also not taken:** widening the sensor's inputs (agent types, action names,
migration files) — a separate change needing its own fire-rate measurement. And
the R2 / mission-review question, which belongs to the concept-register workstream
and is blocked on an empty denominator.

**Status: CLOSED — the detector runs on a cadence that does not depend on anyone
remembering it.**
