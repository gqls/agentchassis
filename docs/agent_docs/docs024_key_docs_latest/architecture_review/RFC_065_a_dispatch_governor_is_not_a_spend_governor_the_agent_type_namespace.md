# RFC 065 — A dispatch governor is not a spend governor: the agent-type namespace (D4b)

**Status:** stage A LIVE (migration 751, 2026-09-03 17:12:21Z); architecture round **APPROVED**
(council corr `dc6d2a54-bd73-4827-8267-49c5500467ac`, 2026-09-03 17:43Z, "approved with 4
advisory objections — none high-severity", 6 abstentions); **stage B REDESIGNED on the
guardian's advisory and NOT yet written.** Filed at the architecture seat's request so the
reasoning is citable rather than buried in migration 751's header. Owner ruling verbatim
(2026-09-03): **"extend it, reducing council spend is a fairly easy save if it comes to the
crunch."** Lane: `dispatch_throughput`. Register: **AGOV-013**.

## 1. The problem, measured

The D4 spend governor went live 2026-09-03 10:14:32Z and was proven end to end the same
morning by an owner-authorised induced shed. It works. It also cannot defend a budget, because
it sheds **work items** and can only touch spend that descends from a dispatch loop.

Attribution by **orchestration lineage**, last 24 h, $319.67 `[MEASURED 2026-09-03]`:

| lineage | share |
|---|---|
| no dispatch-loop ancestor at all | **69.4%** ($221.89) — `council-gate` alone $198.38 = **62%** |
| `build-dispatch-loop` (governed) | 27.6% ($88.16) |
| `diagnose-dispatch-loop` (ungoverned by design) | 3.0% ($9.62) |

Control that makes this trustworthy: **all 4,620 calls' `orchestration_id`s resolve** — nothing
was binned as "outside" by a broken join. Steady, not a spike: council was 69.9% of spend on
09-02 and 62.1% on 09-03; 09-01 ran no councils and cost $54.80 all day. CLAUDE.md §council had
already recorded council as "~85% of fleet LLM spend" before migration 377's caching. The two
facts had never been put side by side.

**⚠ The wrong method, recorded so nobody repeats it:** splitting spend on
`llm_call_log.work_item_id` gives a confident **93.7%** "outside" — false. That column is
propagated one generation and lost (2,278 of 2,278 grandchild-of-a-loop calls carry none;
`page-content-writer`, $112 MTD, is 0.0% populated). It measures where a key is *carried*, not
where the governor *reaches* (`WRONG_CALLS.md` 2026-09-03).

## 2. What was decided, and why

1. **The level comparison is factored into ONE function**, `governor_admits_class(class,
   llm_bearing)`; `governor_admits(item_type)` and `governor_admits_agent(agent_type)` are
   one-line callers differing only in which map they read. The r1 architecture objection on D4
   (corr `8f4bb57d`) was a predicate hand-copied four times; a second namespace is exactly the
   pressure that would re-copy it, so it was factored *before* the namespace was added.
2. **`governor_admits(item_type)` was rewritten LIVE** — the enforcing body the dispatch
   selector, the Go loader and the claim backstop consult. Signature unchanged; selector md5
   `fcbe8821a2a56512911955735796460e` untouched (asserted in the verify). Equivalence was
   **proven**: the old body kept as `governor_admits_legacy`, compared across 4 levels × 3
   classes × 2 bearings + unmapped with the governor enabled, a discriminating control (a
   known-shed cell must read false on BOTH bodies), two induced mutations each caught at its
   own arm, then the copy dropped. Chained apply→rollback proven in one transaction.
3. **A SEPARATE `governor_agent_class_map`**, not more rows in the item-type map: different
   namespaces, and a shared PK would collide silently the day one string appeared in both.
4. **UNMAPPED AGENT TYPES ARE ADMITTED** — the opposite default to work items, deliberately.
   An unmapped item type is a queue row we chose to file; an unmapped agent type is every
   agent in the estate, and defaulting those to shed turns a typo into a fleet outage. Driven
   at every level in the verify; mutation-proven.
5. **`council-gate` seeded at `research` (L3)** and **the LEVEL recorded as an open owner
   question**, not absorbed into a default. "Crunch" reads as *late*. L1 would remove advisory
   review for roughly the second half of each month at current burn — a cultural choice, not a
   config default. The architecture seat concurred: do not arm the flag until the owner
   confirms the level.

## 3. The council round, and what it changed

Approved, with four advisory objections that improved the design. Dispositions:

- **guardian (medium) — "why the shared choke point?"** Stage B was sketched as a gate in
  `processor.go executeWorkflow`, the single admission point for EVERY agent. The guardian
  asked why a narrower seam — a leading step inside council-gate's own workflow — was not
  considered. **It was not, and it is better.** council-gate's workflow already speaks
  `query_database` and `conditional`; a leading `gate_spend_governor` step calling
  `governor_admits_agent('council-gate')` routed to a `complete_withheld` terminal is
  **config-only stage B**: no Go, no roll, no shared choke point, per-agent opt-in by
  construction (only a workflow that carries the step is governed), and **the orchestration
  row itself is the observable** (`current_step = complete_withheld`, found by the same
  `fix_correlation_id` query the 097 runbook already prescribes). **Adopted.**
- **bug_historian (medium) — a failed withheld-row write is fail-CLOSED with no trace.** True
  of the Go sketch: read errors failed open, write errors dropped the run silently. Under the
  config-only design the orchestration row exists *before* the decision, so there is nothing to
  lose on a failed write; a durable `append_doc_note` after the gate adds the long-lived record
  (orchestration rows evaporate in ~24 h). **Dissolved by the redesign; the rule is carried:
  no record, no refusal.**
- **reuse_agent (medium) — `agent_error_log` already exists; why a third audit table?** It
  does: 61,333 rows, 54 codes, written by `validation_drop.go` and `produce_retry.go`
  `[MEASURED 2026-09-03]`. Under the redesign **`governor_withheld_runs` and its view are no
  longer needed** and should be dropped by stage B's migration rather than left as dormant
  surface (the architecture seat's own concern). **Adopted.** (low) A shared Go helper exists —
  `workItemNotGovernorShedSQL` — and all three Go-side callers go through it; had stage B
  stayed in Go it would have mirrored that helper. Moot under the redesign. (low) A single
  `governor_admits_ns(namespace, key, default)` one layer up — noted, not adopted: the
  asymmetric default is the point and reads more clearly as two named functions.
- **guardian (medium) — `099_SYNC_gate_roster.py --apply` regenerates all 17 council-gate
  prompts and configs and would erase the activation.** True of BOTH designs (a flag or a
  workflow step, either lives in the row 099 rewrites). 099 `--apply` is suspended
  (LANDMINES, migration 377), but the guard is owed: **stage B must ship a daily parity check
  that council-gate's `start_step` is the governor gate**, in the 657-VERIFY shape, so a
  regenerated roster reads as drift rather than silently ungoverning the council. **Adopted
  as a stage-B requirement.**
- **editquality (medium) — a sketch is not an edit; risk 6 ("self-applying") is false
  today.** Both fair. Stage B gets its own lighter round with the real config change; risk 6
  is qualified as post-stage-B.
- **debug_historian (medium) — withheld / DB-error / queued are indistinguishable to a
  polling session unless it knows to join.** Under the redesign the 097 runbook's existing
  "find your run" query shows `complete_withheld` directly; the RUNBOOK also carries the
  doc_note lookup. **Runbook line owed with stage B.**
- **architecture (approve, low) — file an RFC; dormant surface if stage B slips; don't arm
  until the level is confirmed.** This file; the drop in stage B; agreed.
- **prior_art (approve) — attach the caller census.** Go: only `work_items_common.go` (the
  helper); config: 1 row (`build-pipeline-trigger`); tasks: 0. council-gate active rows: **1**
  (version 2) — no duplicate-row hazard for the activation.

## 4. Stage B as now designed (config-only, NOT yet written)

A migration on the live `council-gate` `agent_definitions` row, md5-guarded like 674:
1. new `start_step = gate_spend_governor` — `query_database`:
   `SELECT governor_admits_agent('council-gate') AS admitted`, `output_field: governor`;
2. `route_spend_governor` — `conditional`: `governor.admitted == true` → `load_schema_hint`
   (the old start step, unchanged) else → `note_withheld`;
3. `note_withheld` — `append_doc_note`, `subject_key 'spend-governor'`, categories
   `['spend-governor','withheld-run']`, body carrying the correlation id and level;
4. `complete_withheld` — `complete_workflow` with a `success_message` that says, in words a
   polling session will read, *withheld by the spend governor at level N — not queued, do not
   retry; re-trigger when the level drops*.
Plus: drop `governor_withheld_runs` + view (§3, reuse_agent); the daily start-step parity check
(§3, guardian); the 097 runbook line (§3, debug_historian). Verify: drive the level to 3 in a
rolled-back transaction and prove the route flips; fleet negative control that no other agent
row carries the step. **Not armed until the owner confirms the level (§2.5).**

## 5. Consumers told (owner ruling 2026-07-29 §3)

`governor_admits(item_type)`'s three consumers — the dispatch selector text, the Go loader and
the claim backstop via `workItemNotGovernorShedSQL` — have an **unchanged guarantee**: same
signature, same truth table on every cell, selector md5 untouched, proven not asserted.
`council-gate` is the first agent-type consumer and is told by this RFC and the round. Any
further agent type mapped into `governor_agent_class_map` is its own row **and its own review**,
per AGOV-013's standing gate.

## 6. Open

- **Owner:** the council-gate shed LEVEL (§2.5) — one UPDATE; and whether the rest of the 69%
  (`landmine-verifier` $7.94/24 h, the auditors) is worth mapping at all.
- **Lane:** stage B per §4; `bugs_open/459` (the level-change alarm never fires) remains the
  next fix after it.
