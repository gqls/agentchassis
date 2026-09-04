-- 770_..._ROLLBACK.sql — puts html-developer-chunked's three budgets back outside ai_service.
--
-- Needed only if the round-3 ladder is rolled BACK after 770 has been applied: with the old
-- getMaxTokens, a budget inside ai_service is invisible and all three steps silently take the
-- hardcoded 16000. Restoring the bare spelling is what makes them readable again.
BEGIN;

UPDATE agent_definitions a
   SET default_config = jsonb_set(a.default_config, '{workflow,steps}', (
         SELECT jsonb_object_agg(s.key,
                  CASE WHEN s.value->'config'->'ai_service' ? 'max_tokens'
                       THEN jsonb_set(s.value, '{config}',
                              (s.value->'config'
                               || jsonb_build_object('max_tokens', s.value->'config'->'ai_service'->'max_tokens'))
                              || jsonb_build_object('ai_service', (s.value->'config'->'ai_service') - 'max_tokens'))
                       ELSE s.value END)
           FROM jsonb_each(a.default_config->'workflow'->'steps') s)),
       updated_at = now()
 WHERE a.type='html-developer-chunked'
   AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
   WHERE a.type='html-developer-chunked' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
     AND s.value->'config' ? 'max_tokens';
  IF n <> 3 THEN
    RAISE EXCEPTION '770 ROLLBACK VERIFY FAILED: % bare step budgets, expected 3', n;
  END IF;
  RAISE NOTICE '770 ROLLBACK OK';
END $$;

COMMIT;
