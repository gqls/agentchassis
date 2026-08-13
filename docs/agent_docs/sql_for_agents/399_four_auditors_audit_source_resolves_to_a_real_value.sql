-- 399 — four auditors' audit_source now resolves to a real value, not the
-- design-audit default
--
-- Fixes bugs_open/264 candidate 1 (config-only, no roll — see the bug file for
-- the full mechanism). Root cause is bugs_closed/042's unfixed string-literal
-- half: ExtractActionInputs treats every string in step config as a REFERENCE
-- to resolve against collected_data, never as a literal
-- (write_audit_findings_action.go's Strategy 5 comment states this is
-- deliberate — a string that fails to resolve is left alone on purpose, not
-- taken as its own value). Four agent definitions configured audit_source as a
-- plain string ("content-quality-audit" etc.) that can never resolve that way,
-- so all four auditors' findings silently fell back to the "design-audit"
-- default. Measured 2026-08-12: 265 site_work_items rows with
-- audit_source='design-audit' all-history, zero with any of the four
-- configured values (bugs_open/264 §"Measured, live clients_db").
--
-- THE FIX: one new step per agent — a query_database step with no FROM clause
-- (SELECT '<name>'::text AS audit_source) whose output_format:"object"
-- flattens the literal to the top level of a NEW collected_data field
-- (audit_source_literal). The write step's audit_source config then becomes a
-- genuine two-segment dot-path ("audit_source_literal.audit_source"), which
-- Strategy 0 resolves via ExtractNestedField — the same mechanism every
-- correctly-wired config value in this fleet already uses. query_database has
-- no registered ActionInputSpec (it reads "query"/"output_format" straight off
-- StepConfig.Config, confirmed by grep), so this new step needs no config-key
-- registration. No Go change; this migration is config-only and live the
-- moment it commits.
--
-- ORDERING NOTE for the follow-up (candidate 2, tracked in bugs_open/264): a
-- Go change that drops the "design-audit" Defaults entry and makes
-- audit_source Required — closing the door on a fifth author repeating this
-- mistake — MUST NOT roll before this migration is live. An older binary is
-- unaffected either way (Strategy 0 resolution does not depend on the Go
-- change), but rolling the stricter Go code FIRST would hard-fail all four
-- auditors until this migration lands. This migration is the prerequisite and
-- is applied first, config-only, no roll required.
--
-- SURGICAL: jsonb_set only adds the new step key and touches next_step / the
-- one audit_source config value per agent — nothing else in any of the four
-- workflows moves. The guard asserts the new step, the rewired next_step, the
-- updated audit_source, AND that each write step's other config keys and each
-- predecessor step's other keys survived untouched.
--
-- HOW TO VERIFY AFTER APPLYING (per the bug file — do not trust a single
-- design-audit row, that is also what a fully unfixed system produces):
-- run one audit per auditor, then:
--   SELECT spec->>'audit_source', count(*) FROM site_work_items
--   WHERE created_at > '<fix time>' AND spec ? 'audit_source' GROUP BY 1;
-- must return four distinct values, not one.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
--     < 399_four_auditors_audit_source_resolves_to_a_real_value.sql

BEGIN;

SELECT snapshot_agent('site-review-agent', '399_four_auditors_audit_source_resolves_to_a_real_value.sql: pre-update');
SELECT snapshot_agent('visual-design-auditor', '399_four_auditors_audit_source_resolves_to_a_real_value.sql: pre-update');
SELECT snapshot_agent('brief-fidelity-auditor', '399_four_auditors_audit_source_resolves_to_a_real_value.sql: pre-update');
SELECT snapshot_agent('content-quality-auditor', '399_four_auditors_audit_source_resolves_to_a_real_value.sql: pre-update');

-- ── site-review-agent ──────────────────────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,run_strategic_review,next_step}',
             '"set_audit_source"'::jsonb, true),
           '{workflow,steps,set_audit_source}',
           jsonb_build_object(
             'action', 'query_database',
             'config', jsonb_build_object(
               'query', 'SELECT ''site-review''::text AS audit_source',
               'output_format', 'object'),
             'next_step', 'write_strategic_findings',
             'output_field', 'audit_source_literal'),
           true),
         '{workflow,steps,write_strategic_findings,config,audit_source}',
         '"audit_source_literal.audit_source"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'site-review-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── visual-design-auditor ──────────────────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,run_visual_llm_audit,next_step}',
             '"set_audit_source"'::jsonb, true),
           '{workflow,steps,set_audit_source}',
           jsonb_build_object(
             'action', 'query_database',
             'config', jsonb_build_object(
               'query', 'SELECT ''visual-design-audit''::text AS audit_source',
               'output_format', 'object'),
             'next_step', 'write_findings',
             'output_field', 'audit_source_literal'),
           true),
         '{workflow,steps,write_findings,config,audit_source}',
         '"audit_source_literal.audit_source"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'visual-design-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── brief-fidelity-auditor ─────────────────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,run_fidelity_audit,next_step}',
             '"set_audit_source"'::jsonb, true),
           '{workflow,steps,set_audit_source}',
           jsonb_build_object(
             'action', 'query_database',
             'config', jsonb_build_object(
               'query', 'SELECT ''brief-fidelity-audit''::text AS audit_source',
               'output_format', 'object'),
             'next_step', 'write_findings',
             'output_field', 'audit_source_literal'),
           true),
         '{workflow,steps,write_findings,config,audit_source}',
         '"audit_source_literal.audit_source"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'brief-fidelity-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── content-quality-auditor ────────────────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,run_content_llm_audit,next_step}',
             '"set_audit_source"'::jsonb, true),
           '{workflow,steps,set_audit_source}',
           jsonb_build_object(
             'action', 'query_database',
             'config', jsonb_build_object(
               'query', 'SELECT ''content-quality-audit''::text AS audit_source',
               'output_format', 'object'),
             'next_step', 'write_findings',
             'output_field', 'audit_source_literal'),
           true),
         '{workflow,steps,write_findings,config,audit_source}',
         '"audit_source_literal.audit_source"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'content-quality-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    -- --- site-review-agent
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'site-review-agent' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg #>> '{workflow,steps,run_strategic_review,next_step}' IS DISTINCT FROM 'set_audit_source' THEN
        RAISE EXCEPTION '399: site-review-agent run_strategic_review.next_step not rewired';
    END IF;
    IF cfg #>> '{workflow,steps,run_strategic_review,output_field}' IS DISTINCT FROM 'strategic_review' THEN
        RAISE EXCEPTION '399: site-review-agent run_strategic_review lost its own output_field';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,config,query}' IS DISTINCT FROM 'SELECT ''site-review''::text AS audit_source' THEN
        RAISE EXCEPTION '399: site-review-agent set_audit_source query wrong: %', cfg #>> '{workflow,steps,set_audit_source,config,query}';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,next_step}' IS DISTINCT FROM 'write_strategic_findings' THEN
        RAISE EXCEPTION '399: site-review-agent set_audit_source does not chain to write_strategic_findings';
    END IF;
    IF cfg #>> '{workflow,steps,write_strategic_findings,config,audit_source}' IS DISTINCT FROM 'audit_source_literal.audit_source' THEN
        RAISE EXCEPTION '399: site-review-agent write_strategic_findings.audit_source not repointed';
    END IF;
    IF cfg #>> '{workflow,steps,write_strategic_findings,config,site_id}' IS DISTINCT FROM 'site_record.site_id' THEN
        RAISE EXCEPTION '399: site-review-agent write_strategic_findings lost site_id';
    END IF;
    IF cfg #>> '{workflow,steps,write_strategic_findings,config,findings_field}' IS DISTINCT FROM 'strategic_review.result' THEN
        RAISE EXCEPTION '399: site-review-agent write_strategic_findings lost findings_field';
    END IF;

    -- --- visual-design-auditor
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'visual-design-auditor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg #>> '{workflow,steps,run_visual_llm_audit,next_step}' IS DISTINCT FROM 'set_audit_source' THEN
        RAISE EXCEPTION '399: visual-design-auditor run_visual_llm_audit.next_step not rewired';
    END IF;
    IF cfg #>> '{workflow,steps,run_visual_llm_audit,output_field}' IS DISTINCT FROM 'visual_audit' THEN
        RAISE EXCEPTION '399: visual-design-auditor run_visual_llm_audit lost its own output_field';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,config,query}' IS DISTINCT FROM 'SELECT ''visual-design-audit''::text AS audit_source' THEN
        RAISE EXCEPTION '399: visual-design-auditor set_audit_source query wrong: %', cfg #>> '{workflow,steps,set_audit_source,config,query}';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,next_step}' IS DISTINCT FROM 'write_findings' THEN
        RAISE EXCEPTION '399: visual-design-auditor set_audit_source does not chain to write_findings';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,audit_source}' IS DISTINCT FROM 'audit_source_literal.audit_source' THEN
        RAISE EXCEPTION '399: visual-design-auditor write_findings.audit_source not repointed';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,site_id}' IS DISTINCT FROM 'site_record.site_id' THEN
        RAISE EXCEPTION '399: visual-design-auditor write_findings lost site_id';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,findings_field}' IS DISTINCT FROM 'visual_audit.result' THEN
        RAISE EXCEPTION '399: visual-design-auditor write_findings lost findings_field';
    END IF;

    -- --- brief-fidelity-auditor
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'brief-fidelity-auditor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg #>> '{workflow,steps,run_fidelity_audit,next_step}' IS DISTINCT FROM 'set_audit_source' THEN
        RAISE EXCEPTION '399: brief-fidelity-auditor run_fidelity_audit.next_step not rewired';
    END IF;
    IF cfg #>> '{workflow,steps,run_fidelity_audit,output_field}' IS DISTINCT FROM 'fidelity_audit' THEN
        RAISE EXCEPTION '399: brief-fidelity-auditor run_fidelity_audit lost its own output_field';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,config,query}' IS DISTINCT FROM 'SELECT ''brief-fidelity-audit''::text AS audit_source' THEN
        RAISE EXCEPTION '399: brief-fidelity-auditor set_audit_source query wrong: %', cfg #>> '{workflow,steps,set_audit_source,config,query}';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,next_step}' IS DISTINCT FROM 'write_findings' THEN
        RAISE EXCEPTION '399: brief-fidelity-auditor set_audit_source does not chain to write_findings';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,audit_source}' IS DISTINCT FROM 'audit_source_literal.audit_source' THEN
        RAISE EXCEPTION '399: brief-fidelity-auditor write_findings.audit_source not repointed';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,site_id}' IS DISTINCT FROM 'site_record.site_id' THEN
        RAISE EXCEPTION '399: brief-fidelity-auditor write_findings lost site_id';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,findings_field}' IS DISTINCT FROM 'fidelity_audit.result' THEN
        RAISE EXCEPTION '399: brief-fidelity-auditor write_findings lost findings_field';
    END IF;

    -- --- content-quality-auditor
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'content-quality-auditor' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg #>> '{workflow,steps,run_content_llm_audit,next_step}' IS DISTINCT FROM 'set_audit_source' THEN
        RAISE EXCEPTION '399: content-quality-auditor run_content_llm_audit.next_step not rewired';
    END IF;
    IF cfg #>> '{workflow,steps,run_content_llm_audit,output_field}' IS DISTINCT FROM 'content_audit' THEN
        RAISE EXCEPTION '399: content-quality-auditor run_content_llm_audit lost its own output_field';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,config,query}' IS DISTINCT FROM 'SELECT ''content-quality-audit''::text AS audit_source' THEN
        RAISE EXCEPTION '399: content-quality-auditor set_audit_source query wrong: %', cfg #>> '{workflow,steps,set_audit_source,config,query}';
    END IF;
    IF cfg #>> '{workflow,steps,set_audit_source,next_step}' IS DISTINCT FROM 'write_findings' THEN
        RAISE EXCEPTION '399: content-quality-auditor set_audit_source does not chain to write_findings';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,audit_source}' IS DISTINCT FROM 'audit_source_literal.audit_source' THEN
        RAISE EXCEPTION '399: content-quality-auditor write_findings.audit_source not repointed';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,site_id}' IS DISTINCT FROM 'site_record.site_id' THEN
        RAISE EXCEPTION '399: content-quality-auditor write_findings lost site_id';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,findings_field}' IS DISTINCT FROM 'content_audit.result' THEN
        RAISE EXCEPTION '399: content-quality-auditor write_findings lost findings_field';
    END IF;

    RAISE NOTICE '399: all four auditors now resolve audit_source via audit_source_literal.audit_source';
END $$;

COMMIT;
