-- ROLLBACK for 572 — remove `recurrence_expected` from the five build-chain
-- `create_work_item` steps, restoring the pre-572 state in which the key is absent.
--
-- NOTE THE DIFFERENCE BETWEEN ABSENT AND FALSE. `create_work_item` reads the key with
-- `config["recurrence_expected"].(bool)`, so an absent key and an explicit `false` behave
-- identically at runtime — but they are NOT the same to the offline census
-- (`config-key-audit --undeclared-recurrence`), which reports a MISSING declaration as a
-- finding and an explicit value of either kind as clean. Rolling back to `false` would
-- silence the census on exactly the steps it exists to watch. So this removes the key.
--
-- What rolling back restores, stated plainly so nobody applies it casually: a domain
-- re-submitted within three hours of its previous stage item being created will once again
-- queue nothing and report success (bugs_open/326). Prefer the Go-side kill switch
-- (DISABLE_ANTI_CHURN_DEFERRAL) if the problem is the deferral rather than this
-- classification.

BEGIN;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,create_research_item,config,recurrence_expected}',
       updated_at = NOW()
 WHERE type = 'domain-submitter'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,create_next_item,config,recurrence_expected}',
       updated_at = NOW()
 WHERE type IN ('domain-research-classifier','vertical-exemplar-researcher',
                'domain-strategist','build-briefing-agent')
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

DO $$
DECLARE still_declared INT;
BEGIN
    SELECT count(*) INTO still_declared
    FROM agent_definitions d
    CROSS JOIN LATERAL (
        SELECT CASE d.type WHEN 'domain-submitter' THEN 'create_research_item'
                           ELSE 'create_next_item' END AS s
    ) x
    WHERE d.type IN ('domain-submitter','domain-research-classifier',
                     'vertical-exemplar-researcher','domain-strategist','build-briefing-agent')
      AND d.is_active AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
      AND d.default_config->'workflow'->'steps'->x.s->'config' ? 'recurrence_expected';

    IF still_declared <> 0 THEN
        RAISE EXCEPTION '572 ROLLBACK FAILED: % steps still carry the key', still_declared;
    END IF;
    RAISE NOTICE '572 ROLLBACK OK: key removed from all 5 build-chain steps';
END $$;

COMMIT;
