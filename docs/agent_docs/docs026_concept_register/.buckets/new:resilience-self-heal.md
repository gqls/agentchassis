
<!-- SOURCE: U12_docs024_archives.md -->
### Dual-signal self-heal on missing spec dependency
- **category:** NEW:resilience-self-heal
- **status-signal:** deployed
- **status-evidence:** "validate_composition_inputs both loud-logs AND queues a recovery work item on miss... the two-strike rule marks the item unresolved."
- **what:** General resilience pattern for a Go action depending on a spec aspect that may not yet exist: emit a loud error log AND queue a recovery work item that is both a durable dashboard signal and a self-heal mechanism — if it runs successfully, the dependent item auto-redispatches. Repeated failures accumulate via the two-strike rule into `unresolved`.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"1. Status Summary", #"5. Decisions Made This Session"
- **relations:** site-design-planner Choice B scope; work item two-strike/wont_fix pattern
- **verify-later:** `validate_composition_inputs_action.go` implementation; two-strike rule location in dispatch loop.

<!-- SOURCE: U12_docs024_archives.md -->
### Composition resolver orphan-rows policy
- **category:** NEW:resilience-self-heal
- **status-signal:** aspirational
- **status-evidence:** "If install fails, those rows become orphans... we extend the existing database-cleanup scheduled task to sweep them. Draft SQL in draft_composition_orphan_cleanup.sql."
- **what:** Because palette/typography_set resolvers each commit in their own transaction before `install_site_composition` runs, a failed install leaves orphaned rows. Accepted design: let low-cost orphans occur and sweep them periodically via an extension to the existing `database-cleanup` scheduled task, rather than cross-resolver rollback.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Orphan policy"
- **relations:** composition resolution architecture
- **verify-later:** `database-cleanup` scheduled task pre_query — confirm orphan-sweep CTE merged in.

## Proposed new categories
- **NEW:migration-governance** — governance for hand-applied, DB-object-level migrations to `agent_definitions`/launcher defs (idempotency verification, ordering, ledger). Currently only a manual runbook step ("2d state check"); no automated ledger exists.
- **NEW:resilience-self-heal** — cross-cutting pattern for actions with missing upstream dependencies (spec aspects, composition rows) that combine loud logging with a self-healing recovery work item, tolerating cheap orphaned state rather than enforcing cross-action rollback.

<!-- SOURCE: U12_docs024_archives.md -->
### Dual-signal self-heal on missing spec dependency
- **category:** NEW:resilience-self-heal
- **status-signal:** deployed
- **status-evidence:** "validate_composition_inputs both loud-logs AND queues a recovery work item on miss... the two-strike rule marks the item unresolved."
- **what:** General resilience pattern for a Go action depending on a spec aspect that may not yet exist: emit a loud error log AND queue a recovery work item that is both a durable dashboard signal and a self-heal mechanism — if it runs successfully, the dependent item auto-redispatches. Repeated failures accumulate via the two-strike rule into `unresolved`.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"1. Status Summary", #"5. Decisions Made This Session"
- **relations:** site-design-planner Choice B scope; work item two-strike/wont_fix pattern
- **verify-later:** `validate_composition_inputs_action.go` implementation; two-strike rule location in dispatch loop.

<!-- SOURCE: U12_docs024_archives.md -->
### Composition resolver orphan-rows policy
- **category:** NEW:resilience-self-heal
- **status-signal:** aspirational
- **status-evidence:** "If install fails, those rows become orphans... we extend the existing database-cleanup scheduled task to sweep them. Draft SQL in draft_composition_orphan_cleanup.sql."
- **what:** Because palette/typography_set resolvers each commit in their own transaction before `install_site_composition` runs, a failed install leaves orphaned rows. Accepted design: let low-cost orphans occur and sweep them periodically via an extension to the existing `database-cleanup` scheduled task, rather than cross-resolver rollback.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Orphan policy"
- **relations:** composition resolution architecture
- **verify-later:** `database-cleanup` scheduled task pre_query — confirm orphan-sweep CTE merged in.

## Proposed new categories
- **NEW:migration-governance** — governance for hand-applied, DB-object-level migrations to `agent_definitions`/launcher defs (idempotency verification, ordering, ledger). Currently only a manual runbook step ("2d state check"); no automated ledger exists.
- **NEW:resilience-self-heal** — cross-cutting pattern for actions with missing upstream dependencies (spec aspects, composition rows) that combine loud logging with a self-healing recovery work item, tolerating cheap orphaned state rather than enforcing cross-action rollback.
