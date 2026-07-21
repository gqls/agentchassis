# PLAN — dormant-agents capability inventory (bugs_open/044)

**Started:** 2026-07-21. **Bug:** `bugs_open/044` — nothing detects a capability
that exists but nothing routes work to; active agents that have never run are
undetectable because nothing fails.

## The problem, restated

`bugs_closed/002` D: `section-editor` (shipped 2026-02-19, 3 production runs) was
declared **nonexistent** through a handoff and two sign-offs. The platform has no
inventory of its own capabilities and no detector for unused ones, so knowledge
of what exists decays into folklore at the pace of session turnover. This is the
**producer-side** mirror of `bugs_open/033` (work with no worker); here it is a
**worker with no work**. Both are invisible because nothing errors.

## Scope — the detector half only (044 §"The two halves")

044 separates two halves:

1. **The detector** (THIS WORK). A fleet sweep emitting one work item per
   never-run-in-N-days active agent, with an age floor and the mirrored-agent
   caveat handled.
2. **`is_active` hygiene** (NOT this work — explicit owner call in 044). ~34
   retired-but-active rows; 5 types with >1 active row silently shadowed by
   `FindByType`'s `ORDER BY version DESC LIMIT 1`. Deactivating retired agents is
   a judgement about intent. We SURFACE it in the report (duplicate-active-rows
   flagged, era-grouped) but never deactivate anything.

## Design decisions (and why)

- **Home: a new `diagnose_*` action, not a 50th discovery check.** 044 [INFERRED]
  the right home is "a fleet-level sweep beside the immune system's existing
  triage / silent-check." Confirmed by reading the code: all 49 discovery checks
  take a `site_id` (`DiscoveryCheckContext`); "which agents never run" is not a
  per-site question. So `diagnose_dormant_agents_action.go` mirrors
  `diagnose_silent_check_action.go` exactly (deterministic, no LLM, emits inert
  items + one doc_note per sweep, ships dry_run, manual trigger).

- **Detection method: the STEP FINGERPRINT** (044's validated method). A step key
  belonging to exactly one agent, looked for among the top-level step keys ever
  seen in a real `orchestration_states.workflow_plan`. `owner_agent_type` is
  DELIBERATELY NOT used — 95k+ of ~106k rows carry `'generic'` (the dispatch
  path, not the agent); counting that way produced the wrong "110" (WRONG_CALLS
  2026-07-20).

- **Age floor (default 14d).** Without it every fresh seed is flagged
  (evidence-researcher, seeded the day 044 was filed, is new not dormant). The
  floor gates EMISSION only; the report shows under-floor agents in their own
  section so nothing is hidden.

- **Mirrored-agent blind spot: never flagged, always listed.** An agent with no
  unique step (council-gate, whose 099 mirror copies fix-proposer's steps) is
  unmeasurable; by construction it cannot enter the never-observed set. Listed in
  the report as unmeasured. `orchestration_name` was evaluated as a second signal
  and REJECTED — live it is `generic-orchestrate-<ts>`, it does not name the
  agent (see NOTES 2026-07-21).

- **Emission: INERT `dormant_agent` items.** `status='dormant'` (bespoke, inert),
  `pipeline='maintenance'`, anchored to the `system.internal` pseudo-site
  (`triageSystemSiteID`). Nothing claims or dispatches it (claim requires
  `status IN ('triaged','approved')`); triage never escalates it (it gathers only
  `failed`/`deferred`/`capability_gap`); it does NOT land in the 033
  needs_human_review queue. `status='dormant'` is inside `idx_swi_dedup`, so
  `ON CONFLICT DO NOTHING` dedups cleanly (unlike silent-check's `failed` items,
  which need explicit NOT EXISTS). Closed to `'complete'` once the agent runs.

- **Ships dry_run=true, manual trigger.** Same posture as silent-check — the
  first run writes only the report; the owner reviews, then flips dry_run.

## Live measurement at design time (2026-07-21)

| figure | value |
|---|---|
| active non-snapshot agents with a workflow | 155 |
| …measurable (≥1 unique fingerprint step) | 123 |
| …mirrored-agent blind spot (no unique step) | 32 |
| …never observed running | 77 |
| …past the 14d age floor (eligible to emit) | 70 |
| …under the floor (reported only) | 7 |
| types with >1 active row (hygiene shadowing) | 3 |

(044 measured 156/122/57 on 2026-07-20; the never-count grew to 77 as more agents
were seeded and the orchestration window advanced. See NOTES.)

## Deliverables

- `platform/orchestration/actions/diagnose_dormant_agents_action.go` — the action.
- `platform/orchestration/actions/diagnose_dormant_agents_test.go` — pure-fn tests.
- `registry.go` — register `diagnose_dormant_agents`.
- `docs024.../dormant_agents_inventory/seed_diagnosis_dormant_agents.sql` — the
  `diagnosis-dormant-agents` agent def (image-first, dry_run).
- These five living docs.

## Sequencing (image-first)

The Go action is inert until a chassis image carrying it is rebuilt and rolled.
1. Commit (rides the next `git archive HEAD` build).
2. Build + roll chassis (owner/next build cycle).
3. Verify in-pod: `grep -ac diagnose_dormant_agents /proc/1/exe` ≥ 1.
4. Apply the seed (creates `diagnosis-dormant-agents`, dry_run).
5. Trigger a manual sweep; read the report note; review; flip dry_run.

**Until step 2 lands, 044 stays OPEN** (the bar is fixed AND live).

## Open questions for the owner (surfaced, not decided here)

- The `is_active` hygiene half (retire the ~34 legacy rows; collapse the 3 dupes).
- The default emit cap (10) vs the ~70 past-floor findings — coverage is capped,
  not complete, at the default; raise once the report is reviewed.
