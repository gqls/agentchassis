
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec-supersede rollback pattern (and full snapshot revert as its big brother)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 002(4) rollback steps; 014 documents deployed snapshot system (migration 085)
- **what:** Targeted rollback = flip site_specs is_current to a previous aspect version + create rebuild work items; git history gives per-work-item revert of deployed HTML. Full point-in-time revert = site_snapshots (JSONB capture of site record/specs/pages+components/nav/site_components + git SHA), take_site_snapshot / revert_site_to_snapshot (always takes a pre_revert safety snapshot; does NOT git-revert or touch global content_components/research_results). Admin API + three workflow actions exist; recommended auto-triggers (post-deploy, pre-propagation, nightly) not yet wired.
- **sources:** 002(4)#Site Rollback Pattern; 014 full
- **relations:** component_versions; agent snapshots (different concern)
- **verify-later:** site_snapshots rows in prod; whether post-deploy snapshot step was added

<!-- SOURCE: U01_docs024_numbered_core.md -->
### component_versions population and change_source provenance
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 014 (April 2026 change_source column); 001(5) records two years of silently-lost history pre-fix (version_note→change_description drift)
- **what:** Three best-effort writers: StoreGeneratedComponentAction create (v1) and regen (MAX+1, snapshot BEFORE update), UpdateComponentHTMLAction (tool improvements). change_source records originating work-item source. Unique (component_id, version_number). Lesson: best-effort operations need active monitoring — silent best-effort was "silent no-effort" for two years.
- **sources:** 014#Populating component_versions; 001(5) bug 18 second case; 026(2)
- **relations:** component regeneration flow; snapshots
- **verify-later:** component_versions row counts by changed_by

<!-- SOURCE: U12_docs024_archives.md -->
### Milestone-tagged site-spec history with inline git-snapshot function
- **category:** site-snapshots-and-revert
- **status-signal:** superseded
- **status-evidence:** Archive `site_specs` schema carries `milestone`, `superseded_by` columns and a `CommitSpecSnapshot` Go function called inline; live doc drops `milestone`/`superseded_by` entirely, replaces inline snapshotting with a work-item-triggered `snapshot-agent`, adds a bounded "last 5 rows" pruning policy, and drops the legacy `page_components.content_snapshot`/`schema_snapshot` columns.
- **what:** The original design kept unbounded site-spec history in the DB, labelled key rows with a `milestone` string (`initial_research`, `post_build`, `rebrand_q2`...), and relied on a bare Go function invoked directly by completing actions to write a `.site-spec.json` git checkpoint. Content-level rollback used `content_snapshot` on `page_components`. This whole history/rollback substrate was replaced by a decoupled model: `site_specs` prunes to last-5-per-aspect, `page_component_history` is a dedicated append-only table for component rollback, and snapshotting became an ordinary dispatched work item (`needs_snapshot` → `snapshot-agent`) rather than an inline side-effect call.
- **sources:** old/older1/005_build_expand_plan.md#"Table: site_specs", #"Git Spec Snapshots"; docs024_key_docs_latest/P1_build_expand_plan.md#"Removing legacy columns", #"Snapshots"
- **relations:** superseded by snapshot-agent + page_component_history; content-governance locking model
- **verify-later:** confirm in DB whether `page_components.content_snapshot`/`schema_snapshot` columns still exist.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Snapshots and revert (snapshot_agent/revert_agent)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.1 "the convention is to call snapshot_agent('<agent-type>') first … revert_agent finds the most recent unrestored snapshot"
- **what:** Before patching an agent's `default_config`, call `snapshot_agent(type, reason)`; roll back with `revert_agent(type)`. Snapshots are rows in `agent_definitions_backup` kept as an audit trail. A legacy pre-migration pattern stored snapshots in `agent_definitions` itself (is_snapshot/version+1000), the source of several patch/revert footguns.
- **sources:** WM/016_debugging_guide_v2_44.md#6.1, WM/016_debugging_guide_v2_44.md#9
- **relations:** deprecate-not-delete; component_versions history; debugging guide
- **verify-later:** snapshot_agent/revert_agent functions; agent_definitions_backup

<!-- SOURCE: U19_sql_tables_components.md -->
### Site snapshots: point-in-time capture and revert
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 085 migration with take_site_snapshot / revert_site_to_snapshot plpgsql functions, iterated twice in-file with column-name fixes — indicating it was actually run and debugged against the live schema.
- **what:** Full site state captured into one self-contained JSONB row (survives row deletions): site record key fields, all current site_specs, all pages with their page_components (content_data + rendered_html), nav groups/items, site_components; git_commit_sha links DB state to deployed files. Revert takes a safety pre_revert snapshot first, then supersedes specs, delete-and-reinserts pages/components/nav/site_components and restores site fields — explicitly NOT a git revert and does not touch global content_components templates. Triggers: deploy, manual, pre_edit, scheduled.
- **sources:** docs/agent_docs/sql_for_tables/031_site_snapshots.sql
- **relations:** page_component_history (finer grain); agent snapshot/revert (same philosophy for agents); deployment-github (file-side counterpart).
- **verify-later:** snapshot triggers actually firing on deploy; v_site_snapshots contents.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Site snapshots + dated-backup reversibility discipline
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** Pre-migration snapshot 044a0b57 taken 2026-06-23; take_site_snapshot call pattern in the migrations runbook; dated backup tables (_vonc_pc_backup_20260704/09 etc.) created before every risky UPDATE.
- **what:** Every significant change is preceded by reversibility: `take_site_snapshot(site_id, name, ..., 'manual')` for site state; `snapshot_agent('<type>','<reason>')`/`revert_agent` for agent definitions (never a hand-rolled agent_definitions_backup); ad-hoc dated `CREATE TABLE _<site>_<what>_backup_<date> AS SELECT ...` before direct row edits, with the explicit rule never to reuse an old backup name (CREATE TABLE IF NOT EXISTS silently no-ops while looking fresh); restore is UPDATE-in-place keyed on id, not delete+insert.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-snapshot; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-dated-backups; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§5-T0; docs/016b_debugging_guide_merged(3).md#key-schema-gotchas
- **relations:** debugging doctrine; direct-SQL-bypasses-guards caveat
- **verify-later:** take_site_snapshot / snapshot_agent SQL functions (doc 014)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec-supersede rollback pattern (and full snapshot revert as its big brother)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 002(4) rollback steps; 014 documents deployed snapshot system (migration 085)
- **what:** Targeted rollback = flip site_specs is_current to a previous aspect version + create rebuild work items; git history gives per-work-item revert of deployed HTML. Full point-in-time revert = site_snapshots (JSONB capture of site record/specs/pages+components/nav/site_components + git SHA), take_site_snapshot / revert_site_to_snapshot (always takes a pre_revert safety snapshot; does NOT git-revert or touch global content_components/research_results). Admin API + three workflow actions exist; recommended auto-triggers (post-deploy, pre-propagation, nightly) not yet wired.
- **sources:** 002(4)#Site Rollback Pattern; 014 full
- **relations:** component_versions; agent snapshots (different concern)
- **verify-later:** site_snapshots rows in prod; whether post-deploy snapshot step was added

<!-- SOURCE: U01_docs024_numbered_core.md -->
### component_versions population and change_source provenance
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 014 (April 2026 change_source column); 001(5) records two years of silently-lost history pre-fix (version_note→change_description drift)
- **what:** Three best-effort writers: StoreGeneratedComponentAction create (v1) and regen (MAX+1, snapshot BEFORE update), UpdateComponentHTMLAction (tool improvements). change_source records originating work-item source. Unique (component_id, version_number). Lesson: best-effort operations need active monitoring — silent best-effort was "silent no-effort" for two years.
- **sources:** 014#Populating component_versions; 001(5) bug 18 second case; 026(2)
- **relations:** component regeneration flow; snapshots
- **verify-later:** component_versions row counts by changed_by

<!-- SOURCE: U12_docs024_archives.md -->
### Milestone-tagged site-spec history with inline git-snapshot function
- **category:** site-snapshots-and-revert
- **status-signal:** superseded
- **status-evidence:** Archive `site_specs` schema carries `milestone`, `superseded_by` columns and a `CommitSpecSnapshot` Go function called inline; live doc drops `milestone`/`superseded_by` entirely, replaces inline snapshotting with a work-item-triggered `snapshot-agent`, adds a bounded "last 5 rows" pruning policy, and drops the legacy `page_components.content_snapshot`/`schema_snapshot` columns.
- **what:** The original design kept unbounded site-spec history in the DB, labelled key rows with a `milestone` string (`initial_research`, `post_build`, `rebrand_q2`...), and relied on a bare Go function invoked directly by completing actions to write a `.site-spec.json` git checkpoint. Content-level rollback used `content_snapshot` on `page_components`. This whole history/rollback substrate was replaced by a decoupled model: `site_specs` prunes to last-5-per-aspect, `page_component_history` is a dedicated append-only table for component rollback, and snapshotting became an ordinary dispatched work item (`needs_snapshot` → `snapshot-agent`) rather than an inline side-effect call.
- **sources:** old/older1/005_build_expand_plan.md#"Table: site_specs", #"Git Spec Snapshots"; docs024_key_docs_latest/P1_build_expand_plan.md#"Removing legacy columns", #"Snapshots"
- **relations:** superseded by snapshot-agent + page_component_history; content-governance locking model
- **verify-later:** confirm in DB whether `page_components.content_snapshot`/`schema_snapshot` columns still exist.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Snapshots and revert (snapshot_agent/revert_agent)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.1 "the convention is to call snapshot_agent('<agent-type>') first … revert_agent finds the most recent unrestored snapshot"
- **what:** Before patching an agent's `default_config`, call `snapshot_agent(type, reason)`; roll back with `revert_agent(type)`. Snapshots are rows in `agent_definitions_backup` kept as an audit trail. A legacy pre-migration pattern stored snapshots in `agent_definitions` itself (is_snapshot/version+1000), the source of several patch/revert footguns.
- **sources:** WM/016_debugging_guide_v2_44.md#6.1, WM/016_debugging_guide_v2_44.md#9
- **relations:** deprecate-not-delete; component_versions history; debugging guide
- **verify-later:** snapshot_agent/revert_agent functions; agent_definitions_backup

<!-- SOURCE: U19_sql_tables_components.md -->
### Site snapshots: point-in-time capture and revert
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 085 migration with take_site_snapshot / revert_site_to_snapshot plpgsql functions, iterated twice in-file with column-name fixes — indicating it was actually run and debugged against the live schema.
- **what:** Full site state captured into one self-contained JSONB row (survives row deletions): site record key fields, all current site_specs, all pages with their page_components (content_data + rendered_html), nav groups/items, site_components; git_commit_sha links DB state to deployed files. Revert takes a safety pre_revert snapshot first, then supersedes specs, delete-and-reinserts pages/components/nav/site_components and restores site fields — explicitly NOT a git revert and does not touch global content_components templates. Triggers: deploy, manual, pre_edit, scheduled.
- **sources:** docs/agent_docs/sql_for_tables/031_site_snapshots.sql
- **relations:** page_component_history (finer grain); agent snapshot/revert (same philosophy for agents); deployment-github (file-side counterpart).
- **verify-later:** snapshot triggers actually firing on deploy; v_site_snapshots contents.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Site snapshots + dated-backup reversibility discipline
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** Pre-migration snapshot 044a0b57 taken 2026-06-23; take_site_snapshot call pattern in the migrations runbook; dated backup tables (_vonc_pc_backup_20260704/09 etc.) created before every risky UPDATE.
- **what:** Every significant change is preceded by reversibility: `take_site_snapshot(site_id, name, ..., 'manual')` for site state; `snapshot_agent('<type>','<reason>')`/`revert_agent` for agent definitions (never a hand-rolled agent_definitions_backup); ad-hoc dated `CREATE TABLE _<site>_<what>_backup_<date> AS SELECT ...` before direct row edits, with the explicit rule never to reuse an old backup name (CREATE TABLE IF NOT EXISTS silently no-ops while looking fresh); restore is UPDATE-in-place keyed on id, not delete+insert.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-snapshot; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-dated-backups; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§5-T0; docs/016b_debugging_guide_merged(3).md#key-schema-gotchas
- **relations:** debugging doctrine; direct-SQL-bypasses-guards caveat
- **verify-later:** take_site_snapshot / snapshot_agent SQL functions (doc 014)
