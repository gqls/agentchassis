# Register — new:migration-governance

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 1 raw extraction across units U12.

### MIGG-001 — Proposed migration runner/ledger for hand-applied agent-def changes (never built)
- **status:** aspirational
- **status-evidence:** "If this ever graduates beyond hand-application, a migration runner (or even a tiny applied_migrations log table) would enforce order and make re-applying an earlier one structurally impossible" — explicitly proposed, never implemented, in any version of the guide family seen.
- **what:** An explicit proposal to formalize the ad-hoc `jsonb_set` migrations currently applied by hand to `agent_definitions`/launcher defs: a lightweight ledger table (or small runner) tracking which numbered migrations have already been applied, which would structurally prevent the recurring "re-running an earlier migration reverts later ones" incident (see register/debugging.md DBG-010 for the concrete incident this proposal responds to). The only procedural stand-in currently in place is the manual RUNBOOK "2d state check" — not automated.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md (docs024_archives unit)
- **relations:** register/debugging.md DBG-010 (hand-applied agent-def migrations have no ledger — the incident this proposal exists to prevent)
- **verify-later:** grep codebase/DB for any `applied_migrations`/`schema_migrations`-style table scoped to `agent_definitions` — none expected to exist
