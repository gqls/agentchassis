-- 768 ROLLBACK — remove the two copywriter directory check names from completeness-discovery-agent.
BEGIN;
UPDATE agent_definitions ad
   SET default_config = jsonb_set(ad.default_config,
         ARRAY['workflow','steps', (SELECT s.k FROM jsonb_each(ad.default_config->'workflow'->'steps') s(k,v) WHERE s.v->'config' ? 'checks' LIMIT 1), 'config','checks'],
         (SELECT jsonb_agg(e) FROM jsonb_array_elements(
            (SELECT s.v->'config'->'checks' FROM jsonb_each(ad.default_config->'workflow'->'steps') s(k,v) WHERE s.v->'config' ? 'checks' LIMIT 1)) e
           WHERE e::text NOT IN ('"missing_copywriter_directory_section"','"missing_copywriter_directory_page"'))),
       updated_at = NOW()
 WHERE ad.type='completeness-discovery-agent' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
DO $v$
DECLARE arr jsonb;
BEGIN
  SELECT s.v->'config'->'checks' INTO arr FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s(k,v)
   WHERE ad.type='completeness-discovery-agent' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL AND s.v->'config' ? 'checks';
  IF arr ? 'missing_copywriter_directory_section' OR arr ? 'missing_copywriter_directory_page' THEN
    RAISE EXCEPTION '768 ROLLBACK VERIFY: check names still present'; END IF;
  IF NOT (arr ? 'missing_model_directory_section') THEN RAISE EXCEPTION '768 ROLLBACK VERIFY: siblings lost'; END IF;
END $v$;
COMMIT;
