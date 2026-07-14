# Register — site-snapshots-and-revert

4 concepts, consolidated from 12 raw extractions (6 unique blocks, each duplicated
once in the source cluster file) across units U01_docs024_numbered_core,
U12_docs024_archives, U17a_docs019_archive_discussions_and_main,
U19_sql_tables_components, U23_docs_root_vonc. The "dated-backup reversibility"
raw block (U23) described three sibling disciplines at once and was split across
two entries: its site-snapshot invocation detail folds into SNAP-001, its
agent-snapshot invocation detail folds into SNAP-004.

### SNAP-001 — Site snapshots: point-in-time capture and revert (site_snapshots / take_site_snapshot / revert_site_to_snapshot)
- **status:** deployed
- **status-evidence:** Migration 085 shipped `take_site_snapshot`/`revert_site_to_snapshot` as plpgsql functions, iterated twice in-file with column-name fixes ("indicating it was actually run and debugged against the live schema"); real invocations are recorded in production runbooks (a pre-migration snapshot `044a0b57` taken 2026-06-23).
- **what:** Full site state is captured into one self-contained JSONB row (survives row deletions): site record key fields, all current `site_specs`, all pages with their `page_components` (content_data + rendered_html), nav groups/items, `site_components`; `git_commit_sha` links the DB snapshot to the deployed files. `revert_site_to_snapshot` always takes a `pre_revert` safety snapshot first, then supersedes specs and delete-and-reinserts pages/components/nav/site_components and restores site fields — explicitly NOT a git revert, and it does not touch global `content_components` templates. Triggers: deploy, manual, pre_edit, scheduled — though the recommended auto-triggers (post-deploy, pre-propagation, nightly) were not yet wired as of the base doc. A lighter targeted-rollback sibling exists for single-aspect changes: flip `site_specs.is_current` to a previous version + create rebuild work items, relying on git history for per-work-item revert of already-deployed HTML. An Admin API and three workflow actions exist for triggering snapshots/reverts. In practice this is invoked alongside two other reversibility disciplines for the same "always leave a way back" principle: `snapshot_agent`/`revert_agent` for agent definitions (see SNAP-004), and ad-hoc dated `CREATE TABLE _<site>_<what>_backup_<date> AS SELECT ...` tables taken before risky direct-SQL row edits (with the explicit rule never to reuse an old backup name, since `CREATE TABLE IF NOT EXISTS` silently no-ops while looking fresh; restore is UPDATE-in-place keyed on id, not delete+insert).
- **sources:** docs/agent_docs/sql_for_tables/031_site_snapshots.sql; 002(4)#Site Rollback Pattern; 014 (full); docs/RUNBOOK_vonc_migrations(14).md#reference-snapshot; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-dated-backups; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§5-T0
- **relations:** component_versions population (SNAP-002); Milestone-tagged site-spec history (SNAP-003, the superseded predecessor design); Snapshots and revert for agents (SNAP-004); page_component_history (finer grain); deployment-github (file-side counterpart)
- **verify-later:** site_snapshots rows in prod; whether a post-deploy snapshot step was ever added; snapshot triggers actually firing on deploy; v_site_snapshots contents

### SNAP-002 — component_versions population and change_source provenance
- **status:** deployed
- **status-evidence:** 014 documents the April 2026 `change_source` column addition; 001(5) records two years of silently-lost history pre-fix (version_note→change_description drift).
- **what:** Three best-effort writers populate `component_versions`: `StoreGeneratedComponentAction` create (v1) and regen (MAX+1, snapshot BEFORE update), and `UpdateComponentHTMLAction` (tool improvements). `change_source` records the originating work-item source. Enforced unique on (component_id, version_number). Hard lesson banked here: best-effort operations need active monitoring — silent best-effort turned out to be "silent no-effort" for two years before the fix.
- **sources:** 014#Populating component_versions; 001(5) bug 18 second case; 026(2)
- **relations:** Site snapshots (SNAP-001); component regeneration flow
- **verify-later:** component_versions row counts by changed_by

### SNAP-003 — Milestone-tagged site-spec history with inline git-snapshot function (superseded design)
- **status:** partial
- **status-evidence:** The archive `site_specs` schema carries `milestone`/`superseded_by` columns and a `CommitSpecSnapshot` Go function called inline; the live doc drops both columns entirely and replaces inline snapshotting with a work-item-triggered `snapshot-agent`.
- **stage2-verified (2026-07-14):** superseded → partial — Old CommitSpecSnapshot fn / milestone,superseded_by columns: 0 grep hits anywhere in repo (archive-only) — confirmed gone. Replacement claim is split: page_component_history IS real+wired (platform/orchestration/actions/save_component_history_action.go:142, save_page_sections_action.go:437, registry.go:592). But the...
- **what:** The original design kept unbounded site-spec history in the DB, labelled key rows with a `milestone` string (`initial_research`, `post_build`, `rebrand_q2`...), and relied on a bare Go function invoked directly by completing actions to write a `.site-spec.json` git checkpoint. Content-level rollback used `content_snapshot` on `page_components`. This whole history/rollback substrate was replaced by a decoupled model: `site_specs` now prunes to last-5-per-aspect, `page_component_history` is a dedicated append-only table for component rollback, and snapshotting became an ordinary dispatched work item (`needs_snapshot` → `snapshot-agent`) rather than an inline side-effect call. The legacy `page_components.content_snapshot`/`schema_snapshot` columns were also dropped.
- **sources:** old/older1/005_build_expand_plan.md#"Table: site_specs",#"Git Spec Snapshots"; docs024_key_docs_latest/P1_build_expand_plan.md#"Removing legacy columns",#"Snapshots"
- **relations:** superseded by Site snapshots (SNAP-001) + page_component_history; content-governance locking model
- **verify-later:** confirm in DB whether page_components.content_snapshot/schema_snapshot columns still exist

### SNAP-004 — Snapshots and revert for agent definitions (snapshot_agent/revert_agent)
- **status:** deployed
- **status-evidence:** 016 v2_44 §6.1: "the convention is to call snapshot_agent('<agent-type>') first … revert_agent finds the most recent unrestored snapshot"; confirmed in live practice by vonc's migration runbook, which explicitly calls out never hand-rolling an `agent_definitions_backup` row.
- **what:** Before patching an agent's `default_config`, the convention is to call `snapshot_agent(type, reason)`; roll back with `revert_agent(type)`. Snapshots are rows in `agent_definitions_backup`, kept as an audit trail — never hand-rolled directly. A legacy pre-migration pattern stored snapshots in `agent_definitions` itself (`is_snapshot`/`version+1000`), which was the source of several patch/revert footguns and is now deprecated in favour of the dedicated backup table.
- **sources:** WM/016_debugging_guide_v2_44.md#6.1,#9; docs/RUNBOOK_vonc_migrations(14).md#reference-snapshot; docs/016b_debugging_guide_merged(3).md#key-schema-gotchas
- **relations:** Site snapshots (SNAP-001, same reversibility philosophy applied to a different subject); deprecate-not-delete; component_versions history (SNAP-002)
- **verify-later:** snapshot_agent/revert_agent functions; agent_definitions_backup
