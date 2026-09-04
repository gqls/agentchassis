-- 769_..._ROLLBACK.sql — puts every budget back in the place 769 moved it from.
--
-- ⚠ THE TEN AGENTS ARE NAMED BY ID, NOT SELECTED BY SHAPE. After 769 they are indistinguishable
-- from the three agents that legitimately declared a root `ai_service.max_tokens` all along
-- (feed-triage and two others). A shape-based rollback would sweep those three into the bare
-- spelling and INVENT the very defect 769 removed. The values are the ones measured 2026-09-04.
--
-- Restoring is a live behaviour change in its own right: the ten agents' single LLM step drops
-- from its declared 8000 back to 500-2000, and site-adoption-agent's four steps go back to the
-- root 16000. Only run this if 769 itself is judged wrong, not to "undo a change".

BEGIN;

-- (1) the ten agent-level keys, back to the bare root spelling with their original values.
UPDATE agent_definitions AS a
   SET default_config =
         jsonb_set(
           jsonb_set(a.default_config, '{ai_service}', (a.default_config->'ai_service') - 'max_tokens'),
           '{max_tokens}', to_jsonb(v.value), true),
       updated_at = now()
  FROM (VALUES
      ('535f8d1b-5d9b-42b3-a4b7-9a1432421fef'::uuid, 2000),
      ('92b207b3-bf5e-4a4f-9fec-7b67d79f7678'::uuid, 1500),
      ('ec74d095-fee7-418c-bfdc-72b31eb7b72d'::uuid, 2000),
      ('c6432aa9-4ec3-418f-ae9a-d51df4dc627d'::uuid, 2000),
      ('358d7d8d-379f-4fb4-8d1a-15a9d36dce09'::uuid, 2000),
      ('d8b5e2c1-7539-4965-987b-25a4834ccf38'::uuid, 2000),
      ('cb3a90d6-85a8-4650-b3a7-1943df9d0714'::uuid, 1500),
      ('39ab6388-d7b8-4286-8bff-9a80cf18ca63'::uuid, 2000),
      ('43153077-5b48-4638-9b7c-0ae089ff50e0'::uuid, 1500),
      ('411211f5-7da2-4320-8f6f-0194ea23848c'::uuid, 500)
    ) AS v(id, value)
 WHERE a.id = v.id;

-- (2) site-adoption-agent's four step keys, back outside the ai_service block.
UPDATE agent_definitions a
   SET default_config = jsonb_set(a.default_config, '{workflow,steps}', (
         SELECT jsonb_object_agg(s.key,
                  CASE WHEN s.key IN ('analyze_site','derive_content_direction','classify_archetype','generate_design_intent')
                        AND s.value->'config'->'ai_service' ? 'max_tokens'
                       THEN jsonb_set(s.value, '{config}',
                              (s.value->'config'
                               || jsonb_build_object('max_tokens', s.value->'config'->'ai_service'->'max_tokens'))
                              || jsonb_build_object('ai_service', (s.value->'config'->'ai_service') - 'max_tokens'))
                       ELSE s.value END)
           FROM jsonb_each(a.default_config->'workflow'->'steps') s)),
       updated_at = now()
 WHERE a.type = 'site-adoption-agent'
   AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND default_config ? 'max_tokens';
  IF n <> 10 THEN
    RAISE EXCEPTION '769 ROLLBACK VERIFY FAILED: % agents carry a bare root max_tokens, expected the original 10', n;
  END IF;
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND default_config->'ai_service' ? 'max_tokens';
  IF n <> 3 THEN
    RAISE EXCEPTION '769 ROLLBACK VERIFY FAILED: % agents declare root ai_service.max_tokens, expected the original 3', n;
  END IF;
  RAISE NOTICE '769 ROLLBACK OK';
END $$;

COMMIT;
