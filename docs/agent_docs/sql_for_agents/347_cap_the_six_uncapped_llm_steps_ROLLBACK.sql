-- 347_cap_the_six_uncapped_llm_steps_ROLLBACK.sql — hand-run companion, NOT a
-- migration (uppercase suffix: the runner excludes it from --apply).
--
-- Deletes exactly the keys 347 set, VALUE-MATCHED so a cap anyone has since
-- re-chosen (different number) survives, as does the pre-existing
-- chief-strategist 8192. Restores the fall-to-2048 behaviour; the ai_actions
-- WARN will fire again for these steps. Config is live immediately.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('brand-designer',               '347 ROLLBACK: pre-delete');
SELECT snapshot_agent('chief-strategist',             '347 ROLLBACK: pre-delete');
SELECT snapshot_agent('content-creator',              '347 ROLLBACK: pre-delete');
SELECT snapshot_agent('domain-analyst',               '347 ROLLBACK: pre-delete');
SELECT snapshot_agent('provocation-gate-calibration', '347 ROLLBACK: pre-delete');
SELECT snapshot_agent('site-architect',               '347 ROLLBACK: pre-delete');

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,analyze_brand,config,ai_service,max_tokens}',
       updated_at = now()
 WHERE type = 'brand-designer'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND (default_config #>> '{workflow,steps,analyze_brand,config,ai_service,max_tokens}')::numeric = 8000;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,generate_build_plan,config,ai_service,max_tokens}',
       updated_at = now()
 WHERE type = 'chief-strategist'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND (default_config #>> '{workflow,steps,generate_build_plan,config,ai_service,max_tokens}')::numeric = 16000;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,create_content,config,ai_service,max_tokens}',
       updated_at = now()
 WHERE type = 'content-creator'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND (default_config #>> '{workflow,steps,create_content,config,ai_service,max_tokens}')::numeric = 16000;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,analyze,config,ai_service,max_tokens}',
       updated_at = now()
 WHERE type = 'domain-analyst'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND (default_config #>> '{workflow,steps,analyze,config,ai_service,max_tokens}')::numeric = 8000;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,gate,config,ai_service,max_tokens}',
       updated_at = now()
 WHERE type = 'provocation-gate-calibration'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND (default_config #>> '{workflow,steps,gate,config,ai_service,max_tokens}')::numeric = 8000;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,design,config,ai_service,max_tokens}',
       updated_at = now()
 WHERE type = 'site-architect'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND (default_config #>> '{workflow,steps,design,config,ai_service,max_tokens}')::numeric = 32000;

-- Post-condition: none of the six (type, step) pairs carries the 347 value,
-- and the pre-existing chief-strategist 8192 is intact.
DO $verify$
DECLARE
    v_count int;
BEGIN
    SELECT count(*) INTO v_count
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
       AND ((ad.type = 'brand-designer'               AND s.key = 'analyze_brand'       AND (s.value#>>'{config,ai_service,max_tokens}')::numeric = 8000)
         OR (ad.type = 'chief-strategist'             AND s.key = 'generate_build_plan' AND (s.value#>>'{config,ai_service,max_tokens}')::numeric = 16000)
         OR (ad.type = 'content-creator'              AND s.key = 'create_content'      AND (s.value#>>'{config,ai_service,max_tokens}')::numeric = 16000)
         OR (ad.type = 'domain-analyst'               AND s.key = 'analyze'             AND (s.value#>>'{config,ai_service,max_tokens}')::numeric = 8000)
         OR (ad.type = 'provocation-gate-calibration' AND s.key = 'gate'                AND (s.value#>>'{config,ai_service,max_tokens}')::numeric = 8000)
         OR (ad.type = 'site-architect'               AND s.key = 'design'              AND (s.value#>>'{config,ai_service,max_tokens}')::numeric = 32000));
    IF v_count <> 0 THEN
        RAISE EXCEPTION '347 ROLLBACK: % step(s) still carry the 347 caps', v_count;
    END IF;

    SELECT count(*) INTO v_count
      FROM agent_definitions
     WHERE type = 'chief-strategist'
       AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
       AND (default_config #>> '{workflow,steps,generate_build_plan,config,ai_service,max_tokens}')::numeric = 8192;
    IF v_count <> 1 THEN
        RAISE EXCEPTION '347 ROLLBACK: the pre-existing 8192 cap was disturbed (found % rows)', v_count;
    END IF;
END;
$verify$;

COMMIT;
