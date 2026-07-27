# Register — new:resilience-self-heal

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

2 concepts, consolidated from 2 raw extractions across units U12.

### RSH-001 — Dual-signal self-heal on missing spec dependency
- **status:** deployed
- **status-evidence:** "validate_composition_inputs both loud-logs AND queues a recovery work item on miss... the two-strike rule marks the item unresolved."
- **what:** A general resilience pattern for a Go action depending on a spec aspect that may not yet exist: emit a loud error log AND queue a recovery work item in the same failure path — the log is a durable dashboard signal, and the queued item is a genuine self-heal mechanism, since if it later runs successfully the originally-dependent item auto-redispatches. Repeated failures of the same recovery item accumulate via a two-strike rule into a terminal `unresolved` state rather than retrying forever.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"1. Status Summary", #"5. Decisions Made This Session" (docs024_archives unit)
- **relations:** site-design-planner Choice B scope; work-item two-strike/wont_fix pattern; composition resolver orphan-rows policy (RSH-002, same composition subsystem)
- **verify-later:** `validate_composition_inputs_action.go` implementation; two-strike rule location in the dispatch loop

### RSH-002 — Composition resolver orphan-rows policy
- **status:** aspirational
- **status-evidence:** "If install fails, those rows become orphans... we extend the existing database-cleanup scheduled task to sweep them. Draft SQL in draft_composition_orphan_cleanup.sql."
- **what:** Because the palette/typography_set resolvers each commit in their own transaction before `install_site_composition` runs, a failed install leaves orphaned rows behind. The accepted design tolerates this — low-cost orphans are allowed to occur and are swept up periodically by an extension to the existing `database-cleanup` scheduled task, rather than adding cross-resolver rollback/transaction coordination.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Orphan policy" (docs024_archives unit)
- **relations:** composition resolution architecture; dual-signal self-heal (RSH-001, same source unit/subsystem)
- **verify-later:** `database-cleanup` scheduled task pre_query — confirm the orphan-sweep CTE was actually merged in
