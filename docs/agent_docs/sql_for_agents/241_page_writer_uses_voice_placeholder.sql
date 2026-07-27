-- 241_page_writer_uses_voice_placeholder.sql
--
-- DO NOT APPLY UNTIL THE CHASSIS CARRIES THE INJECTION. This is the second half
-- of bugs_open/121: it removes page-content-writer's literal copy of the house
-- voice and replaces it with the {{.voice_style}} placeholder that the chassis
-- fills from the single `voice_style_block` row (seeded by 240).
--
-- WHY THE GATE IS REAL, NOT CEREMONIAL. The prompt renderer is missingkey=zero:
-- an unresolved {{.voice_style}} renders as NOTHING, with no error and no log
-- line. Applying this against a chassis that predates the injection therefore
-- deletes the house voice from every page build, silently, and the only symptom
-- is that the writing gets worse. There is no failing status to notice.
--
-- GATE — both must pass on the RUNNING pod before applying:
--   POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
--   kubectl exec -n ai-persona-system $POD -- sh -c \
--     'strings /app/agent-chassis | grep -c "voice_style_block"'   # expect > 0
--   kubectl exec -n ai-persona-system $POD -- sh -c \
--     'strings /app/agent-chassis | grep -c "SELECT config->>.text. FROM agent_default_configs"'  # expect > 0
--
-- Then, and only then:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db < <this file>

\set ON_ERROR_STOP on
BEGIN;

DROP TABLE IF EXISTS bak_agent_definitions_pcw_voice_placeholder;
CREATE TABLE bak_agent_definitions_pcw_voice_placeholder AS
SELECT * FROM agent_definitions WHERE type = 'page-content-writer';

-- Replace everything from the "Voice & Style" heading to the end of the block
-- with the placeholder. The block is the final section of the prompt_template,
-- so this is a suffix replacement anchored on the heading.
WITH cur AS (
  SELECT a.id,
         a.default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}' #>> '{}' AS tmpl
  FROM agent_definitions a
  WHERE a.type='page-content-writer' AND a.is_active
    AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
), cut AS (
  SELECT id,
         substring(tmpl from 1 for position('Voice & Style (how the copy must READ' in tmpl) - 1)
           || E'{{.voice_style}}\n' AS tmpl,
         position('Voice & Style (how the copy must READ' in tmpl) AS anchor
  FROM cur
)
UPDATE agent_definitions a
SET default_config = jsonb_set(
      a.default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(c.tmpl)),
    updated_at = now()
FROM cut c
WHERE a.id = c.id AND c.anchor > 0;

DO $$
DECLARE t text;
BEGIN
  SELECT v->'config'->>'prompt_template' INTO t
  FROM agent_definitions a,
       jsonb_each(a.default_config->'workflow'->'steps'->'process_sections_loop'
                  ->'config'->'sub_workflow'->'steps') AS e(k,v)
  WHERE a.type='page-content-writer' AND a.is_active
    AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL AND e.k='generate_content';

  IF t NOT LIKE '%{{.voice_style}}%' THEN
    RAISE EXCEPTION 'placeholder absent - the anchor did not match. ROLLING BACK.';
  END IF;
  IF t LIKE '%No em dashes, anywhere, ever%' THEN
    RAISE EXCEPTION 'the literal block is still present, so this would be a THIRD copy. ROLLING BACK.';
  END IF;
  -- The rest of the prompt must survive. If the anchor matched too early the
  -- template would be gutted, and a short template is the tell.
  IF length(t) < 8000 THEN
    RAISE EXCEPTION 'template is only % chars - too much was cut. ROLLING BACK.', length(t);
  END IF;
  RAISE NOTICE 'OK: literal replaced by placeholder, % chars remain.', length(t);
END $$;

COMMIT;

-- AFTER APPLYING: rebuild one page and read the copy. The placeholder resolving
-- to nothing looks identical to it resolving correctly, right up until you read
-- the output - so check the artefact, not the status.
