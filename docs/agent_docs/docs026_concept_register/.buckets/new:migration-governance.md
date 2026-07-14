
<!-- SOURCE: U12_docs024_archives.md -->
### NEW:migration-governance — proposed migration runner/ledger for hand-applied agent-def changes (never built)
- **category:** NEW:migration-governance
- **status-signal:** aspirational
- **status-evidence:** "If this ever graduates beyond hand-application, a migration runner (or even a tiny applied_migrations log table) would enforce order and make re-applying an earlier one structurally impossible" — explicitly proposed, never implemented, in any version of the family or the live doc.
- **what:** An idea for formalizing ad-hoc `jsonb_set` migrations applied by hand to `agent_definitions`/launcher defs: a lightweight ledger table or runner tracking which numbered migrations had been applied, preventing accidental reversion.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** procedural stand-in currently in place is the RUNBOOK "2d state check" (manual, not automated)
- **verify-later:** grep codebase/DB for any `applied_migrations`/`schema_migrations`-style table scoped to `agent_definitions` — none expected to exist.

<!-- SOURCE: U12_docs024_archives.md -->
### NEW:migration-governance — proposed migration runner/ledger for hand-applied agent-def changes (never built)
- **category:** NEW:migration-governance
- **status-signal:** aspirational
- **status-evidence:** "If this ever graduates beyond hand-application, a migration runner (or even a tiny applied_migrations log table) would enforce order and make re-applying an earlier one structurally impossible" — explicitly proposed, never implemented, in any version of the family or the live doc.
- **what:** An idea for formalizing ad-hoc `jsonb_set` migrations applied by hand to `agent_definitions`/launcher defs: a lightweight ledger table or runner tracking which numbered migrations had been applied, preventing accidental reversion.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** procedural stand-in currently in place is the RUNBOOK "2d state check" (manual, not automated)
- **verify-later:** grep codebase/DB for any `applied_migrations`/`schema_migrations`-style table scoped to `agent_definitions` — none expected to exist.
