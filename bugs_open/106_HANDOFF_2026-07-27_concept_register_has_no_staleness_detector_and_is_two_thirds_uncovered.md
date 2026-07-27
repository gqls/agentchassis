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
