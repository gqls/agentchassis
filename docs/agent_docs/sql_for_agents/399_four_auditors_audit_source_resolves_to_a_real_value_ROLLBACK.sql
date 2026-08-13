-- 399 ROLLBACK — restore the four auditors' write step to its pre-migration
-- shape: audit_source back to the unresolvable literal, set_audit_source step
-- removed, predecessor next_step repointed at the write step directly.
--
-- Safe at any time: this only ever restores audit_source to its ORIGINAL
-- (broken, design-audit-defaulting) value — it does not depend on any work
-- having run since 399 applied.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
--     < 399_four_auditors_audit_source_resolves_to_a_real_value_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('site-review-agent', '399 ROLLBACK: pre-revert');
SELECT snapshot_agent('visual-design-auditor', '399 ROLLBACK: pre-revert');
SELECT snapshot_agent('brief-fidelity-auditor', '399 ROLLBACK: pre-revert');
SELECT snapshot_agent('content-quality-auditor', '399 ROLLBACK: pre-revert');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           (default_config #- '{workflow,steps,set_audit_source}'),
           '{workflow,steps,run_strategic_review,next_step}',
           '"write_strategic_findings"'::jsonb, true),
         '{workflow,steps,write_strategic_findings,config,audit_source}',
         '"site-review"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'site-review-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,set_audit_source}' IS NOT NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           (default_config #- '{workflow,steps,set_audit_source}'),
           '{workflow,steps,run_visual_llm_audit,next_step}',
           '"write_findings"'::jsonb, true),
         '{workflow,steps,write_findings,config,audit_source}',
         '"visual-design-audit"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'visual-design-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,set_audit_source}' IS NOT NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           (default_config #- '{workflow,steps,set_audit_source}'),
           '{workflow,steps,run_fidelity_audit,next_step}',
           '"write_findings"'::jsonb, true),
         '{workflow,steps,write_findings,config,audit_source}',
         '"brief-fidelity-audit"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'brief-fidelity-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,set_audit_source}' IS NOT NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           (default_config #- '{workflow,steps,set_audit_source}'),
           '{workflow,steps,run_content_llm_audit,next_step}',
           '"write_findings"'::jsonb, true),
         '{workflow,steps,write_findings,config,audit_source}',
         '"content-quality-audit"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'content-quality-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,set_audit_source}' IS NOT NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config INTO cfg FROM agent_definitions WHERE type = 'site-review-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF cfg #> '{workflow,steps,set_audit_source}' IS NOT NULL THEN RAISE EXCEPTION '399 ROLLBACK: site-review-agent still has set_audit_source'; END IF;
    IF cfg #>> '{workflow,steps,write_strategic_findings,config,audit_source}' IS DISTINCT FROM 'site-review' THEN RAISE EXCEPTION '399 ROLLBACK: site-review-agent audit_source not restored'; END IF;

    SELECT default_config INTO cfg FROM agent_definitions WHERE type = 'visual-design-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF cfg #> '{workflow,steps,set_audit_source}' IS NOT NULL THEN RAISE EXCEPTION '399 ROLLBACK: visual-design-auditor still has set_audit_source'; END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,audit_source}' IS DISTINCT FROM 'visual-design-audit' THEN RAISE EXCEPTION '399 ROLLBACK: visual-design-auditor audit_source not restored'; END IF;

    SELECT default_config INTO cfg FROM agent_definitions WHERE type = 'brief-fidelity-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF cfg #> '{workflow,steps,set_audit_source}' IS NOT NULL THEN RAISE EXCEPTION '399 ROLLBACK: brief-fidelity-auditor still has set_audit_source'; END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,audit_source}' IS DISTINCT FROM 'brief-fidelity-audit' THEN RAISE EXCEPTION '399 ROLLBACK: brief-fidelity-auditor audit_source not restored'; END IF;

    SELECT default_config INTO cfg FROM agent_definitions WHERE type = 'content-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF cfg #> '{workflow,steps,set_audit_source}' IS NOT NULL THEN RAISE EXCEPTION '399 ROLLBACK: content-quality-auditor still has set_audit_source'; END IF;
    IF cfg #>> '{workflow,steps,write_findings,config,audit_source}' IS DISTINCT FROM 'content-quality-audit' THEN RAISE EXCEPTION '399 ROLLBACK: content-quality-auditor audit_source not restored'; END IF;

    RAISE NOTICE '399 ROLLBACK: all four auditors restored to pre-migration shape';
END $$;

COMMIT;
