# HANDOFF 2026-08-16 — continue here

**Lane:** `register_guards_code_phase_b` (`bugs_open/288`, the class behind
`bugs_closed/225`). **State: the code is built, committed and council round 2 is IN
FLIGHT; nothing is live and nothing has ever fired.** Read this first, then NOTES for
the evidence and PLAN for why the design is what it is.

## The one-sentence state

The evidence register can now be told which facts a calculator encodes, and the daily
sweep will name that calculator when a fact moves — **but the code is not in a running
binary, no tool declares anything yet, and therefore the mechanism has never fired on
real data.**

## THREE THINGS OWED, in order

### 1. Read the round-2 council verdict — the code is already on the shared branch

Round 1 was **REVISE** and the gating objection was right (see §"What round 1 caught").
It was answered in full and resubmitted on the same trail. At the time of writing round
2 is at `review_editquality`.

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

Commits carry `Council-Submitted: cff364b8-…`, which 098 credits automatically on
approval. **Do not write `Council-Reviewed:` unless you have read an approved verdict.**

### 2. After the next chassis roll — prove it at the binary, then INDUCE it

Pre-roll state is **measured**, not assumed (2026-08-16, both replicas):
`fact_drift_review` → **0**, `stale_attestation` (control) → **5**. Re-run exactly that
pair; the first must become non-zero.

```bash
kubectl -n ai-persona-system exec <chassis-pod> -- grep -ac fact_drift_review /proc/1/exe
kubectl -n ai-persona-system exec <chassis-pod> -- grep -ac stale_attestation /proc/1/exe
```

Then the induced proof — **this is the only thing that distinguishes a live mechanism
from an inert one**, and the recipe is in `RUNBOOK_register_guards_code.md`:
seed the mcalc fence → dry run → supersede `sdlt-ftb-relief-cap` 500000→550000 → dry run
must name `stamp-duty` → restore. A dry run that reports nothing after a real change is
the failure.

### 3. The first declaration is another lane's to apply

`CONTRIB_2026-08-16_phase_b_built_your_stamp_duty_fence_is_the_first_consumer.md` is in
`mortgagecalculator_couk_adoption/`. Their site, their `install_fences.py`, their call.
**Do not hand-edit the `doc_plans` row** — their script rewrites whole bodies.

LMC's is written too, and is inert until that site has an `evidence_base` row at all
(it has none; the `copy_quality_two_stage` lane holds a candidate pending an owner
decision).

## What round 1 caught, and why it matters more than the fix

The `editquality` seat found that **the mechanism was blind to its own motivating bug**.
The baseline came purely from prior REGISTER state, so a tool stale on the day it opts
in — against a fact that has not moved since — produced `baseline == current` and
emitted nothing. That is `bugs_closed/225` exactly.

I had a test asserting that silence, with a comment justifying it. The justification was
true; the behaviour was still wrong. **The check I should have written first: does this
fire on 225?** Nothing in thirteen passing tests asked.

Fixed by `unreconciled_declaration` — a first declaration files one low-severity review
item per (fact, tool), self-quieting because the item records the value that becomes the
next baseline. Mutation **M8** reverts to the old behaviour and the new test fails.

Also fixed: a risk inversion (a confirmed move on a `no_auto_fix` tax tool ranked below
an auto-fixable one — now three bands) and the audit-vs-liveness page predicate.

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
