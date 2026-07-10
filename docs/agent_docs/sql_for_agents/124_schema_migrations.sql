-- 124_schema_migrations.sql — applied-migrations tracking (the migrations system).
-- APPLIED 2026-07-10 to clients_db. From this file onward, every numbered file
-- in docs/agent_docs/sql_for_agents/ is applied via scripts/migration/
-- run-migrations.sh, which records it here. Files 001–123 predate the system
-- and were applied ad hoc — they are HISTORY, not pending work; the runner's
-- baseline (124) excludes them.
--
-- Related but distinct, both pre-existing:
--   * snapshot_agent(type, reason) — before-images of agent_definitions rows
--     (standing rule for every agent-updating migration);
--   * migration_backups — manual before-value backups of arbitrary rows.
-- schema_migrations records WHAT ran WHEN; those record what it replaced.

BEGIN;

CREATE TABLE IF NOT EXISTS schema_migrations (
    filename   text PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now(),
    checksum   text,
    applied_by text NOT NULL DEFAULT current_user,
    notes      text
);

COMMENT ON TABLE schema_migrations IS
  'Applied-migration ledger for docs/agent_docs/sql_for_agents/ (baseline 124). Maintained by scripts/migration/run-migrations.sh.';

-- ── Backfill: the travelling-docs arc (2026-07-04 → 2026-07-10), applied
-- manually before this system existed. Dates from RUNBOOK_travelling_docs
-- (approximate to the day unless stamped); checksums are of the repo copies
-- at renumbering time. Idempotent.
INSERT INTO schema_migrations (filename, applied_at, checksum, applied_by, notes) VALUES
  ('125_doc_plans_and_notes.sql',                  '2026-07-04', '341a593d9a36b94c966aad4632ad4514', 'backfill', 'doc_plans + doc_notes tables; applied pre-system, date from RUNBOOK'),
  ('126_wire_persist_diagnosis_note.sql',          '2026-07-06', '7a1988e271a645a85d83f18284f9d847', 'backfill', 're-drafted version (error_step inside config); applied pre-system'),
  ('127_diagnose_load_runtime_error_step.sql',     '2026-07-06', '6efc9a0791c946fc95a9a2604ede7b85', 'backfill', 'applied pre-system'),
  ('128_fix_load_runtime_error_step_target.sql',   '2026-07-06', NULL,                               'backfill', 'ORIGINAL FILE LOST (old workspace); repo file is a reconstruction stub; effect verified live 2026-07-10'),
  ('129_wire_diagnosis_subject_threading.sql',     '2026-07-06', '6005e80c7a2d6fb9f6155380316e6f45', 'backfill', 'applied pre-system'),
  ('130_pilot_plan_tool_archetype_taster_quiz.sql','2026-07-07 12:32+01', '8e40bdc5d0532bd25f6f588b92e194a1', 'backfill', 'data seed (first tool PLAN); applied pre-system'),
  ('131_tool_generator_plan_writing.sql',          '2026-07-07', 'a1f1a8d28e5f3f17dc61a0baec322d05', 'backfill', 'Task 3 wiring; applied pre-system'),
  ('132_fix_agents_note_writing.sql',              '2026-07-07', '9332de2e8ba99988d27ec6107147a43c', 'backfill', 'Task 4 wiring (3 agents); applied pre-system'),
  ('133_add_component_provenance.sql',             '2026-07-08', 'de9cabead96aabb9fa921104a5245d7f', 'backfill', 'applied version (guarded, type-mirroring); design doc remains in docs019'),
  ('134_fix_prompt_template_field_paths.sql',      '2026-07-09', '775fb8bd4bd0c2beefcb91803e39754e', 'backfill', 'applied pre-system'),
  ('135_bypass_index_plan_until_embed_timeout.sql','2026-07-09', 'baee7e4da1dfe9a016524ebce3ceb3ad', 'backfill', 'stopgap; undone by 139 after chassis v1.0.1102'),
  ('136_supersede_xp_curve_plan_selectors.sql',    '2026-07-09', 'a7c0b908711b4f6fa2e9808a0a076740', 'backfill', 'v3 (strpos guards + dynamic evidence); v1/v2 attempts remain in travelling_docs'),
  ('137_recreation_spec_and_note_subject.sql',     '2026-07-09', '2268cada6d06630c67c8dc50414b7376', 'backfill', 'declare spec; re-subject recreation note to pipeline/build'),
  ('138_recreate_tool_carries_spec_features.sql',  '2026-07-09', '0d71543d2498e7b0e0a3363a2e45cfa4', 'backfill', 'recreate_tool prompt renders spec.interactive_features'),
  ('139_reenable_index_plan.sql',                  '2026-07-10', '598553be5668877b5a47b04eaea65506', 'backfill', 'applied 2026-07-10 after v1.0.1102 shipped the embedding deadline')
ON CONFLICT (filename) DO NOTHING;

COMMIT;

-- Verify:
--   SELECT filename, applied_at::date, checksum IS NOT NULL AS has_sum, notes
--   FROM schema_migrations ORDER BY filename;
-- Expect 15 backfilled rows + this file's own row once the runner records it.
