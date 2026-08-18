-- ROLLBACK for 471 — restore the finance directory researcher's narrower intake.
--
-- Note what this does NOT undo: any directory_entities / directory_claims rows the
-- widened runs registered in the meantime. Those are verified facts and stay —
-- narrowing intake again does not make an already-cited fact wrong. If the intent
-- is to remove them, archive the entities by hand and say why.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(jsonb_set(jsonb_set(
        default_config,
        '{workflow,steps,search_web,config,num_results}',   '10'::jsonb, true),
        '{workflow,steps,prepare_urls,config,max_scrapes}',  '4'::jsonb, true),
        '{workflow,steps,prepare_urls,config,max_snippets}', '5'::jsonb, true),
    updated_at = now()
WHERE type = 'finance-directory-researcher'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_claims,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,extract_claims,config,prompt_template}',
            'A page that ENUMERATES many firms and states a fact about each is a strong source and historically the highest-yield one — a trade-body member list, a regulator or government list of authorised firms, or a specialist broker''s provider round-up. What is weak is a page whose substance is RANKINGS or PRICING rather than firm facts. Prefer, in order: the regulator or a government list, a trade-body member list, the firm''s own regulatory page, then a specialist round-up.',
            'Third-party listicles are weak — prefer the register or the firm.'
        )),
        true),
    updated_at = now()
WHERE type = 'finance-directory-researcher'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DELETE FROM scheduled_tasks
WHERE name IN ('mortgage-lender-directory-discovery-bsa',
               'mortgage-lender-directory-discovery-adverse',
               'mortgage-lender-directory-discovery-btl',
               'mortgage-lender-directory-discovery-fscs');

DELETE FROM schema_migrations WHERE filename = '471_widen_finance_directory_discovery.sql';

DO $$
DECLARE n int; p int;
BEGIN
    SELECT (default_config #>> '{workflow,steps,prepare_urls,config,max_scrapes}')::int INTO n
      FROM agent_definitions WHERE type='finance-directory-researcher' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n IS DISTINCT FROM 4 THEN RAISE EXCEPTION 'rollback: max_scrapes is %, expected 4', n; END IF;
    SELECT count(*) INTO p FROM agent_definitions
     WHERE type='finance-directory-researcher' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,extract_claims,config,prompt_template}' LIKE '%listicles are weak%';
    IF p <> 1 THEN RAISE EXCEPTION 'rollback: prompt not restored (matching rows: %)', p; END IF;
    RAISE NOTICE '471 rollback OK';
END $$;

COMMIT;
