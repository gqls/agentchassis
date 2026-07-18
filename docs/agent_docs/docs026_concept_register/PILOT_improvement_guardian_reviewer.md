# PILOT — "Improvement-Loop Guardian" council reviewer (stage 3, seat #7; candidate #5)

**Status: LIVE as of 2026-07-18.** Applied via
`fixloop_eg_dartsonline/0NN_fix_proposer_v14_improvement_guardian.sql` — gated
+ surgical (the settled pattern). Council is now **9 reviewers** (2 always-on +
7 gated specialists).

---

## 1. Why this seat

Candidate #5. The improvement loop (discovery → triage → fix → rerender) is a
genuine "master workflow" in FIX-036's framing, and its guardrails were earned
from a real incident: the audit→fix→re-audit loop once ran **unbounded** —
845+ findings across 4 domains in ~10 days, consuming most of the token budget
(`IMP-027`, the "triage drain" fix). A fix that touches this machinery could
quietly reopen that class of failure.

## 2. Charter — the contracts it defends

| Contract | Grounding |
|---|---|
| Bounded passes (≥3 → `complete_clean`); passed sections get `locked_at` and are skipped; unlock is always manual | `IMP-003`, `IMP-027` |
| Discovery checks enabled ONLY via the agent's `checks` config array; a check is enabled only once its handler agent exists; findings insert at `detected` (unclaimable, observe-only capable) | `IMP-004` |
| Checks append `WorkItemSpecs`; the RUNNER inserts with dedup — plugins never insert their own rows | `IMP-004` |
| Verification via the finding's `acceptance_test` (cheap targeted call), never a full re-audit; `depends_on` sequencing prevents overlapping fixes | `IMP-027` |
| Sweep cadence lives in `scheduled_tasks` under the shared dispatch concurrency group; discovery never floods a backed-up queue | `IMP-0xx` (cadence concept) |

**Judges:** (a) pass-cap bypass / auto-unlock / re-auditing locked sections;
(b) enabling a check with no handler, or claimable-status findings; (c) checks
inserting their own rows; (d) full re-audits replacing acceptance tests, or
broken sequencing/concurrency discipline. Doesn't touch the loop → approve.

**Verdicts: `approve | object`, no `veto`.**

## 3. Relevance footprint (gated)

```
"improvement": ["improvement","discovery_check","run_discovery_checks","write_audit_findings","triage_detected","audit_pass","complete_clean","locked_at","acceptance_test","needs_rerender","maintenance_queue"]
```
