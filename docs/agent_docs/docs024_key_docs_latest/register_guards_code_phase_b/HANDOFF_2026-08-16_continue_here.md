# HANDOFF 2026-08-16 — continue here

**Lane:** `register_guards_code_phase_b` (`bugs_open/288`, the class behind
`bugs_closed/225`). **State (updated 2026-08-17): the code is council-APPROVED and LIVE — proven at the
binary on both replicas, and the first production sweep ran clean. It has still NEVER
FIRED, because no tool declares anything yet; that one step is an ask sitting with the
mortgagecalculator lane.** Read this first, then NOTES for
the evidence and PLAN for why the design is what it is.

## The one-sentence state

The evidence register can now be told which facts a calculator encodes, and the daily
sweep will name that calculator when a fact moves — **but the code is not in a running
binary, no tool declares anything yet, and therefore the mechanism has never fired on
real data.**

## THREE THINGS OWED, in order

### 1. ~~Read the council verdict~~ DONE — APPROVED at round 3

`cff364b8`, 2026-08-16 16:55Z: **approved**, 13 of 15 seats, 2 advisory objections
(both acted on in `6b3b0510e`), none high-severity. The `editquality` seat that gated
rounds 1 and 2 approves. Commits carry `Council-Reviewed: cff364b8-…`. **Nothing is
owed to the gate.** The remaining two items below are.

<details><summary>How to re-read it (the column is <code>body</code>, not <code>content</code>)</summary>

**Rounds 1 and 2 were both REVISE and both gating objections were right.** Round 1: the
mechanism was blind to its own motivating bug. Round 2: **the round-1 fix was inert**,
because the baseline fell back to previous REGISTER state and the register is
re-verified daily, so the "never reconciled" case could never fire. Both answered; round
3 submitted on the same trail. Expect the same seat (`editquality`) to gate again if
anything is still wrong — it has been right twice.

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id'='cff364b8-9e55-49b6-88aa-bfc1edc2e1a6'
 ORDER BY created_at DESC;
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='cff364b8-9e55-49b6-88aa-bfc1edc2e1a6' AND kind='council_report'
 ORDER BY created_at;          -- TWO rows means round 2 has landed; read the newest
```
Full objection text: `SELECT body FROM diagnosis_artifacts WHERE correlation_id=… AND
kind='council_report' ORDER BY created_at DESC LIMIT 1;` (the column is **`body`**, not
`content` — that cost a round trip).

</details>

### 2. ~~Prove it at the binary~~ DONE 2026-08-17 — live, and the first sweep was clean

| check | result |
|---|---|
| `fact_drift_review` in `/proc/1/exe`, both replicas | **2** (was 0 pre-roll) |
| `unreconciled_declaration`, both replicas | **1** |
| `stale_attestation` (positive control) | 5 |
| strings unique to the final approved revision `6b3b0510e` | present ×2, so the binary is post-council, not an earlier build |
| `evidence-freshness` sweep | ran **09:04:14Z**, 8 register revisions, **0 errors** |
| `fact_drift` items fleet-wide | **0** |

⚠ **That 0 is not a pass.** It has no demand behind it — 0 of 90 tool PLANs declare
anything, so there is nothing to act on. What the run proves is the NO-OP case: the new
query path ran against all 13 register-bearing sites in production and broke nothing.

**The induced proof is still owed** and is the only thing that separates a live
mechanism from an inert one. Recipe in `RUNBOOK_register_guards_code.md`; the reusable
dispatcher is `./dryrun_fact_drift.sh <site_id>` (dry runs write nothing). It cannot be
run until a fence declares — see item 3.

### 3. The first declaration — ASKED 2026-08-17, awaiting the mortgagecalculator lane

**The owner was asked and directed that this go to the lane first**, so the request now
sits at the TOP of `mortgagecalculator_couk_adoption/HANDOFF_2026-08-16b_continue_here.md`
(a CONTRIB alone had sat unread since 08-16), offering three options: seed it, let us run
a dry-run canary and revert, or decline. **Do not apply it unilaterally** — the site is
theirs, they were active on it on 08-16, and seeding files items into a review queue that
`bugs_open/033` says has no working surface. **Do not hand-edit the `doc_plans` row**
either — their `install_fences.py` rewrites whole bodies on `--apply`.

If they say yes, the full detail is in
`CONTRIB_2026-08-16_phase_b_built_your_stamp_duty_fence_is_the_first_consumer.md`.

LMC's is written too, and is inert until that site has an `evidence_base` row at all
(it has none; the `copy_quality_two_stage` lane holds a candidate pending an owner
decision).

## What the two rounds caught, and why it matters more than the fix

The `editquality` seat found that **the mechanism was blind to its own motivating bug**.
The baseline came purely from prior REGISTER state, so a tool stale on the day it opts
in — against a fact that has not moved since — produced `baseline == current` and
emitted nothing. That is `bugs_closed/225` exactly.

I had a test asserting that silence, with a comment justifying it. The justification was
true; the behaviour was still wrong. **The check I should have written first: does this
fire on 225?** Nothing in thirteen passing tests asked.

Fixed by `unreconciled_declaration` — a first declaration files one low-severity review
item per (fact, tool), self-quieting because the item records the value that becomes the
next baseline.

**Then round 2 found that fix was INERT.** `unreconciled_declaration` fires only when the
baseline is nil, and `baselineFor` fell back to the previous `evidence_base` row — but
the register is re-verified daily, so a previous row essentially always exists carrying
the correct value. Baseline was never nil. I had fixed the symptom and left the cause in
the fallback chain. The repair deleted `previousRow` entirely: the baseline is now the
newest `fact_drift` finding for (fact, tool) and nothing else, and its absence is the
never-reconciled signal.

⚠ **Proving that fix took three mutations.** M11 and M11b both PASSED — one mutated a
path the struct-literal test bypasses, the other read a row without using it (and
`sqlmock.ExpectationsWereMet` reports unfulfilled expectations, not extra queries). Only
M11c (declared, populated AND consulted) failed. **An induced red is evidence only when
the mutation is faithful to the defect** — `WRONG_CALLS.md`, 2026-08-16.

Also fixed: a risk inversion (a confirmed move on a `no_auto_fix` tax tool ranked below
an auto-fixable one — now three bands), the audit-vs-liveness page predicate, and the
hand-written subject-key rule (now reuses exported `discovery_checks.ToolSubjectKeyExpr`,
verified live to resolve both tools).

## Landmines this lane learned the hard way

- **`toolEligibilityWhere` admits NEITHER SDLT tool** (both multi-component). Keying the
  fan-out on it would have produced a permanently-silent check that read green.
- **`p.status='active'` is the liveness predicate and is wrong for an audit** — use
  `NeverDeployedPagePredicateFor`.
- **`go build ./...` at the working tree may fail for another session's reasons.** Build
  against `git archive HEAD` plus your own files.
- **A subagent's report is another doc** — I put its stale figures in a commit message.
  `WRONG_CALLS.md`, 2026-08-16.
- **`diagnosis_artifacts.body`**, not `.content`.

## What this does NOT do (do not let a green read wider)

It answers *did the figure MOVE*, never *is the figure RIGHT* — a tool and a register
wrong in the same direction agree and it is silent. That is Piece 4 (an oracle), behind
its own RFC. It needs a human to author every declaration. It cannot see a page with no
`pages` row (mortgagecalculator's repo-only `stamp-duty.html` twin carried the identical
defect). The prose half of the class — tool-page prose and currency anywhere — is
untouched. All named in `bugs_open/288` §5.

## The five living docs

PLAN (design + the two rejected alternatives) · RUNBOOK (commands, each with its gotcha)
· NOTES (evidence + both missteps + the council round) · README_where_we_are (the
owner's plain-prose log) · **no SUMMARY yet — the milestone is the first live emission,
not the build.**
