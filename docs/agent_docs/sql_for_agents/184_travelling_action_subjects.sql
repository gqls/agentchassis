-- 184_travelling_action_subjects.sql
--
-- Enable travelling PLAN+NOTES subjects for ACTIONS, then seat the three
-- fix-loop actions the feature stage loop now shares.
--
-- Answers the delta-2 council-gate objection (submission 5a65ec4c,
-- tooling_provenance): diagnose_read_repo_files, diagnose_prepare_fix_commit
-- and diagnose_build_gate are pre-existing tools the LIVE fix loop depends on;
-- delta 2 added seams to all three. The standing rule says such subjects carry a
-- travelling PLAN+NOTES (subject_type='action', keyed by action name). That was
-- NOT actually possible: the doc_plans/doc_notes subject_type CHECK allowed only
-- tool/pipeline/experience, so the convention the seat cites had no schema
-- support (and zero action rows existed). This migration adds 'action' to both
-- constraints and seats the three subjects — the highest-value judgement surface,
-- because a regression in these actions breaks fix-implementer, not just the new
-- feature path.
--
-- Idempotent: constraint re-add is guarded; row inserts use NOT EXISTS.

BEGIN;

-- 1. Allow 'action' as a subject_type on both tables.
ALTER TABLE doc_plans DROP CONSTRAINT IF EXISTS doc_plans_subject_type_check;
ALTER TABLE doc_plans ADD CONSTRAINT doc_plans_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text,'pipeline'::text,'experience'::text,'action'::text]));

ALTER TABLE doc_notes DROP CONSTRAINT IF EXISTS doc_notes_subject_type_check;
ALTER TABLE doc_notes ADD CONSTRAINT doc_notes_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text,'pipeline'::text,'experience'::text,'action'::text]));

-- 2. PLANS — standing design + invariants per shared action.
INSERT INTO doc_plans (subject_type, subject_key, body, source, source_agent, created_by)
SELECT 'action', v.k, v.body, 'council-gate-5a65ec4c', 'feature-builder-thread', 'fixloop-feature-builder-2'
FROM (VALUES
  ('diagnose_read_repo_files',
   E'SHARED ACTION — consumers: fix-implementer AND feature-implementer (verified live 2026-07-18: no third pipeline calls it).\n\nPurpose: read the plan''s modify/add files via the GitHub contents API at a chosen ref; the read token is supplied to the dedicated pod by the spawn gate.\n\nSeam added for the feature stage loop (delta 2, c19b5d097): optional `ref_field` resolves the base ref from collected_data so a per-stage router can point stage N at the feat/* branch carrying stage N-1''s commits. INVARIANT: `ref_field` UNSET keeps the single-plan (fix-implementer) path byte-identical. `ref_field` CONFIGURED-BUT-RESOLVED-EMPTY is a HARD ERROR (9c94cc842), never a silent fall back to main — falling back would read the wrong tree, the platform''s worst-known failure shape.'),
  ('diagnose_prepare_fix_commit',
   E'SHARED ACTION — consumers: fix-implementer AND feature-implementer (verified live 2026-07-18: no third pipeline).\n\nPurpose: allowlist-validate the generated whole-file bodies and prepare the git-adapter commit payload. Since fc38c6058 it also gofmts generated Go at commit-prep so cosmetic misalignment does not burn a build-gate run (bugs_open/013).\n\nSeams added for the feature stage loop (delta 2, c19b5d097): optional `branch_field` (commit to a routed branch instead of deriving fix/<short>), `commit_message_field`, and `expected_symbols_field` enforcement (each declared symbol must appear verbatim in a produced file; missingExpectedSymbols rejects otherwise). INVARIANTS: all three UNSET keep the single-plan path byte-identical; `branch_field`/`commit_message_field` CONFIGURED-BUT-EMPTY are HARD ERRORS (9c94cc842). The expected_symbols false-reject risk (a symbol defined in an earlier stage''s file) is mitigated at the SOURCE — feature-designer PATCH_018 steers the designer to name only symbols the stage''s OWN files introduce — NOT by weakening this gate.'),
  ('diagnose_build_gate',
   E'SHARED ACTION — consumers: fix-implementer AND feature-implementer (verified live 2026-07-18).\n\nPurpose: a k8s Job that clones the fix branch and runs gofmt -l on changed files + a targeted go build; red gate => no PR, branch + log left for inspection.\n\nSeam added for the feature stage loop (delta 2, c19b5d097, E2): optional `test_packages_field` end-gate mode runs `go test` over the union of packages containing the plan''s edited .go files, DERIVED in Go from the committed file list, never declared by the model (D6). INVARIANT: `test_packages_field` UNSET keeps the build-only gate byte-identical; CONFIGURED-BUT-RESOLVED-TO-NO-PACKAGES is a HARD ERROR (9c94cc842) — a silent build-only gate would forfeit the exact D6 guarantee the mode exists to provide. buildGateScript gained one trailing parameter; all call sites (3 test + 1 production + the definition) accounted for.')
) AS v(k, body)
WHERE NOT EXISTS (
  SELECT 1 FROM doc_plans p WHERE p.subject_type='action' AND p.subject_key=v.k AND p.is_current
);

-- 3. NOTES — the delta-2 change entry per action.
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, source_agent, created_by)
SELECT 'action', v.k, v.body,
       '["fixloop","feature-builder","provenance","council-gate"]'::jsonb,
       'council-gate-5a65ec4c', 'feature-builder-thread', 'fixloop-feature-builder-2'
FROM (VALUES
  ('diagnose_read_repo_files',
   E'2026-07-21: added `ref_field` seam for the feature stage loop (delta 2). Reviewed by the council gate (5a65ec4c); the configured-but-empty silent-fallback flagged by bug_historian/guardian is fixed fail-loud (9c94cc842). Before changing this action, keep the single-plan path byte-identical and the empty-resolution hard-error.'),
  ('diagnose_prepare_fix_commit',
   E'2026-07-21: added `branch_field` / `commit_message_field` / `expected_symbols_field` seams (delta 2). Council gate 5a65ec4c raised the silent-fallback class high — fixed fail-loud (9c94cc842). expected_symbols false-reject mitigated at the designer source (PATCH_018), gate intact. Also carries the bugs_open/013 gofmt-at-commit-prep behaviour (fc38c6058) — do not re-commit ungofmted generated Go.'),
  ('diagnose_build_gate',
   E'2026-07-21: added `test_packages_field` end-gate mode (delta 2, E2). Council gate 5a65ec4c: build-only silent no-op fixed fail-loud (9c94cc842); packages are DERIVED in Go from committed .go edits, never model-declared (D6). One trailing buildGateScript parameter added; all call sites updated.')
) AS v(k, body)
WHERE NOT EXISTS (
  SELECT 1 FROM doc_notes n WHERE n.subject_type='action' AND n.subject_key=v.k
    AND n.source='council-gate-5a65ec4c'
);

COMMIT;

-- Verify
SELECT 'plans' AS kind, subject_key FROM doc_plans WHERE subject_type='action' AND is_current
UNION ALL
SELECT 'notes', subject_key FROM doc_notes WHERE subject_type='action' AND source='council-gate-5a65ec4c'
ORDER BY 1, 2;
