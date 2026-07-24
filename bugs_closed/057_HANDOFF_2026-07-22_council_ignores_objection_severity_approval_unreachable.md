# 057 — the council ignored objection severity, so an APPROVE was unreachable

**Filed 2026-07-22** (fixloop council-gate thread). **Status: CLOSED 2026-07-24 —
FIXED & LIVE on chassis `v1.0.1149`, BOTH halves now behaviourally verified on
organic rounds** (approve path 2026-07-22; gating floor 2026-07-22→24, §"CLOSED"
below). The dogfood submission this file was holding for (corr `e0a9b843`)
APPROVED rather than gated, so it never supplied the gating example — the
examples arrived organically instead, the first (a guardian hard veto, corr
`1d8ef2c0`) the same evening this file was written.

> **VERIFIED LIVE 2026-07-22.** Pod-grep of `agent-chassis-…` on `v1.0.1149`:
> `objectionGates`, `severityGates`, and the new literals `"advisory objection"` /
> `"gating objection"` all present (positive control `decideCouncil` present).
> **Behavioural proof (the half a pod-grep can't give):** an ORGANIC council round
> decided `approved` with `decided_by` = **"approved with 4 advisory objection(s) —
> none high-severity"** (corr `cbbc7c83`, 14:07) — a round with 4 low/medium
> objections that would have been `revise` under the old rule. The gating floor
> (high/veto/degraded/un-graded still gates) is covered by the unit tests baked
> into the same binary and will show on the next organic high-severity round.
Full diagnosis, sizing and the fix options weighed:
`docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/SUMMARY_2026-07-22_council_approval_rate.md`.

## CLOSED 2026-07-24 — the gating floor verified live, organically

All 25 organic council rounds after the roll (2026-07-22 14:00 → 2026-07-24
15:49, `diagnosis_artifacts` kind=`council_report`) split **13 approved / 9
revise / 3 rejected**, and every decision matches the new rule exactly:

- **Every `revise` was decided by an explicitly high-severity objection** —
  `decided_by` = "gating objection from <seat>" and a per-round check found ≥1
  `severity='high'` objection in all 9 (e.g. corr `c2a9fd27`, 07-23 15:56,
  editquality, 2 high; the `fa4b77cd` run of four contracts-seat rounds).
- **Every `rejected` was a veto**: "hard veto from guardian" (corrs `1d8ef2c0`
  07-22 17:08, `6cdbc374` 07-24 15:17) and "veto from feasibility"
  (`fa4b77cd` 07-23 16:00). The first landed ~3 hours after this file said
  "none organic yet — councils quiet".
- **Every `approved` carried the advisory string** ("approved with N advisory
  objection(s) — none high-severity") — no bare-unanimity round needed.
- **This file's own discriminator query over the post-fix window: 0 of 9 revise
  rounds lacked a high objection** (pre-fix: 59 of 88). No round now gates on
  low/medium nits. (Query as in §Root cause, window `>'2026-07-22 14:00'`, with
  `coalesce(rv->'objections','[]')` — some post-fix reviews omit the array.)

Approval per submission ran ~80% in the two days after the roll (8 of 10 corrs;
fixloop NOTES turn 11), against ~5% the week before. Fix commits `872c830a8` +
`e37ec804f` (harvested Degraded-zero-objections test, `Council-Reviewed:
e0a9b843`) both in HEAD; the "gating objection from"/"advisory objection(s)"
strings in today's rows are emitted by the new code, so the live binary is
proven behaviourally, not just by pod-grep.

## Symptom

Councils almost never emit an `approved` verdict. Measured fleet-wide over 7 days
(`diagnosis_artifacts` kind=`council_report`, `metadata->>'decision'`): **123
revise / 3 rejected / 2 approved** across ~44 distinct submissions — a ~4.5%
approval rate. Threads iterate 5–7 rounds and still land on revise. Consequence:
the council-gate coverage report (`098`) reads ~0 REVIEWED (the `Council-Reviewed`
trailer needs an approve), and PR-mode (enforcement) is unbuildable — it would
block ~95% of bugfixes.

This affects **every** council, not just the gate: fix-proposer, council-gate,
experience-planner and concept-register all share `diagnose_council_decide`.

## Root cause (CONFIRMED from primary evidence)

`decideCouncil` in `platform/orchestration/actions/diagnose_council_decide_action.go`
returned `revise` on the FIRST reviewer whose verdict was `object`, and **never
consulted `councilObjection.Severity`** (`low|medium|high`) — a field that has been
on the wire contract and populated by reviewers all along. So a single low nit
from one of ~16 seats blocked exactly as hard as a genuine flaw, and `approved`
required *unanimous bare approval* across every readable seat.

Evidence it is the rule and not the plans (last 3 days, 88 revise rounds):
- **Objection severity mix: medium 279 / low 165 / high 41 — high is 8.5%.**
- **67% of revise rounds (59/88) carried NO high-severity objection at all** — they
  were blocked entirely by low/medium nits. By worst-objection class: high 29,
  medium 56, low 4.
- 14 rounds were decided by a **lone** objector against ~8.8 approving seats; in 13
  of those 14 the blocking objection was low/medium, not high.

Reproducible query (the discriminator):
```sql
WITH r AS (SELECT body::jsonb b FROM diagnosis_artifacts
  WHERE kind='council_report' AND created_at>now()-interval '3 days'
    AND metadata->>'decision'='revise')
SELECT count(*) FILTER (WHERE NOT EXISTS (
  SELECT 1 FROM jsonb_array_elements(b->'reviews') rv, jsonb_array_elements(rv->'objections') o
  WHERE rv->>'verdict'='object' AND lower(o->>'severity')='high')) AS no_high_rounds,
  count(*) AS total FROM r;   -- expect ~59 / ~88 before the roll
```

## The fix (committed `872c830a8`, inert until image roll)

Wire severity into `decideCouncil` (owner ruling 2026-07-22, "only high/veto
blocks"): a high-severity objection or a veto still gates; **explicitly low/medium
objections are advisory** — recorded in the report's `reviews` and returned to the
proposer, but non-gating. Conservative carve-outs so a minor label cannot hide a
real problem: a **Degraded** (truncated) object still gates (its high objection may
have been cut off), and an **un-graded / unrecognised** severity still gates — only
an *explicitly* low/medium objection is waved through. New helpers `severityGates`
+ `objectionGates`; tests in `diagnose_council_test.go`
(`TestDecideCouncil`, `TestObjectionGates`).

## How to verify AFTER the roll

1. Pod carries the change: `kubectl exec -n ai-persona-system <chassis-pod> -- sh -c
   'strings /app/agent-chassis | grep -c objectionGates'` (expect ≥1; 0 = not rolled).
2. Behavioural: a council round whose objections are all low/medium now decides
   `approved` with `decided_by` = "approved with N advisory objection(s) — none
   high-severity". Watch the approval rate lift on the same 7-day query above.
3. The floor still holds: a round with any high-severity objection, a veto, or a
   Degraded/un-graded object still decides `revise`/`rejected`.

## Second finding, routed elsewhere (do not lose)

The diagnosis loop I filed to grade this **stalled at its `route` step** (orch
`9d86cb45`, corr `fa333384-c3bc-4c7e-8d26-105f25ade755`) and was reaped after 4h —
`route` spawns a code-lookup child and awaits a response that never returned. That
is the spawn-loss / dispatch-queue backlog (`bugs_open/003`, `bugs_open/030`)
biting the diagnosis path itself; a code-seeded diagnosis currently cannot
complete. Belongs to those cases — linked here, not forked.
