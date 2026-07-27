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

-- CORRECTED 2026-07-27, after v1 of this file was REFUSED BY ITS OWN GUARD.
-- v1 assumed the Voice & Style block was the final section and truncated the
-- template at the anchor. It is not: the block sits at char 272 of 16,150 and
-- ends at "## Company Context". The guard caught it ("template is only 289
-- chars - too much was cut") and rolled back, leaving the row untouched.
--
-- That refusal also surfaced a SECOND defect, in the earlier v4 apply: the three
-- new rules were appended to the end of the TEMPLATE, not the end of the BLOCK.
-- They sit ~11,500 chars away, after the JSON output instructions, and one of
-- them still says "the word-weight rule above" - a reference that no longer
-- resolves anywhere near it. The v4 guard passed because it asserted the strings
-- were PRESENT, not that they were POSITIONED. An assertion that checks presence
-- and not position passes a misplacement.
--
-- This file therefore does two things: swap the block for the placeholder, and
-- delete the orphaned tail. Both are redundant once the placeholder resolves,
-- because the canonical row holds the complete v4 rules.

WITH cur AS (
  SELECT a.id,
         a.default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}' #>> '{}' AS tmpl
  FROM agent_definitions a
  WHERE a.type='page-content-writer' AND a.is_active
    AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
), cut AS (
  SELECT id,
    -- 1. block -> placeholder: keep everything before the block, insert the
    --    placeholder, then resume at the heading that follows the block.
    substring(tmpl from 1 for position('Voice & Style (how the copy must READ' in tmpl) - 1)
      || E'{{.voice_style}}\n\n'
      || substring(tmpl from position('## Company Context' in tmpl))
    AS tmpl
  FROM cur
), detailed AS (
  SELECT id,
    -- 2. drop the orphaned tail appended by the v4 apply.
    CASE WHEN position('- Say why it matters, not just what is true' in tmpl) > 0
         THEN rtrim(substring(tmpl from 1 for position('- Say why it matters, not just what is true' in tmpl) - 1))
         ELSE tmpl END AS tmpl
  FROM cut
)
UPDATE agent_definitions a
SET default_config = jsonb_set(
      a.default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(d.tmpl)),
    updated_at = now()
FROM detailed d
WHERE a.id = d.id;

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
  -- Everything after the block must survive. These are the sections that
  -- followed it; losing one means the anchor matched in the wrong place.
  IF position('## Company Context' in t) = 0 THEN
    RAISE EXCEPTION 'Company Context section lost. ROLLING BACK.';
  END IF;
  IF position('## STRICT RULES:' in t) = 0 THEN
    RAISE EXCEPTION 'STRICT RULES section lost. ROLLING BACK.';
  END IF;
  -- The orphaned v4 tail must be gone, or the rules exist twice again.
  IF t LIKE '%Do not restate your opening%' THEN
    RAISE EXCEPTION 'the orphaned v4 tail is still present. ROLLING BACK.';
  END IF;
  -- POSITION, not just presence: the placeholder must sit where the block did,
  -- near the top, not wherever a stray match put it. This is the assertion the
  -- v4 apply lacked.
  IF position('{{.voice_style}}' in t) > 500 THEN
    RAISE EXCEPTION 'placeholder is at char %, not near the top where the block was. ROLLING BACK.', position('{{.voice_style}}' in t);
  END IF;
  RAISE NOTICE 'OK: literal replaced by placeholder, % chars remain.', length(t);
END $$;

COMMIT;

-- AFTER APPLYING: rebuild one page and read the copy. The placeholder resolving
-- to nothing looks identical to it resolving correctly, right up until you read
-- the output - so check the artefact, not the status.
