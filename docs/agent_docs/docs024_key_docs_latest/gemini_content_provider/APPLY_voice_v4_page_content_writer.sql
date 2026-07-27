-- FILE: APPLY_voice_v4_page_content_writer.sql
--
-- Voice & Style v4 on page-content-writer. Owner-approved 2026-07-27 after a
-- measured 5-run test on gemini-pro-latest (see NOTES): chars 422 -> 637,
-- mean words/sentence 7.6 -> 12.1, sentence-length SD 5.7, em dashes still 0,
-- negative frames still 0.
--
-- WHY RULE 1 IS REPLACED RATHER THAN APPENDED TO. The live rule reads "One idea
-- per sentence. If a sentence chains two or more ideas with commas or dashes,
-- split it." That is the instruction producing the staccato fact-list the owner
-- objected to: it forbids exactly the sentence they asked for ("...easy for a
-- single demo, but getting them to recover takes months"). Appending a softer
-- rule underneath would leave the prompt self-contradictory, and these models
-- follow the literal earlier rule. So the bullet is rewritten in place.
--
-- RUN IT:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db < <this file>

\set ON_ERROR_STOP on
BEGIN;

DROP TABLE IF EXISTS bak_agent_definitions_pcw_voice_v4;
CREATE TABLE bak_agent_definitions_pcw_voice_v4 AS
SELECT * FROM agent_definitions WHERE type = 'page-content-writer';

-- Rewrite bullet 1 and append the three new bullets, in one jsonb_set on the
-- nested generate_content step. Merge semantics are not needed here (we are
-- replacing a single string leaf, not an object with siblings) but the path is
-- the same nested one as the provider flip - NOT a top-level step.
WITH cur AS (
  SELECT a.id,
         a.default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}' #>> '{}' AS tmpl
  FROM agent_definitions a
  WHERE a.type='page-content-writer' AND a.is_active
    AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
), rewritten AS (
  SELECT id,
    replace(
      tmpl,
      '- One idea per sentence. If a sentence chains two or more ideas with commas or dashes, split it.',
      '- One idea per sentence, USUALLY. A sentence may carry two ideas when they are genuinely one thought and a conjunction joins them: "chaining models together is easy for a single demo, but getting them to recover from errors takes months". What this rule bans is the three-clause pile-up and the comma-spliced list, not the word "but". Vary sentence length on purpose. A run of short declaratives in a row reads like a specification being read aloud, which is a worse fault than the one this rule was written to prevent.'
    ) AS tmpl
  FROM cur
)
UPDATE agent_definitions a
SET default_config = jsonb_set(
      a.default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(r.tmpl
        || E'\n- Say why it matters, not just what is true. At least one sentence should give the reader a reason to care that they could not have guessed from the facts alone. Write like someone with a point of view who has done this work, not like a specification being read out.'
        || E'\n- Do not always reach for the most obvious word. Where two words are equally accurate, take the less predictable one, as long as it is still plain English and still the honest word. This is not licence to reach for a grander word: that breaks the word-weight rule above. Reach for a more specific one.'
        || E'\n- Do not restate your opening in different words. If two sentences make the same point with different vocabulary, delete one.')),
    updated_at = now()
FROM rewritten r
WHERE a.id = r.id;

-- Fail LOUD rather than commit a half-applied prompt.
DO $$
DECLARE t text;
BEGIN
  SELECT v->'config'->>'prompt_template' INTO t
  FROM agent_definitions a,
       jsonb_each(a.default_config->'workflow'->'steps'->'process_sections_loop'
                  ->'config'->'sub_workflow'->'steps') AS e(k,v)
  WHERE a.type='page-content-writer' AND a.is_active
    AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL AND e.k='generate_content';

  IF t LIKE '%One idea per sentence. If a sentence chains%' THEN
    RAISE EXCEPTION 'the OLD rule 1 is still present - the replace did not match. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%One idea per sentence, USUALLY%' THEN
    RAISE EXCEPTION 'the new rule 1 is absent. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%reason to care that they could not have guessed%' THEN
    RAISE EXCEPTION 'the why-it-matters rule is absent. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%take the less predictable one%' THEN
    RAISE EXCEPTION 'the word-choice rule is absent. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%Do not restate your opening%' THEN
    RAISE EXCEPTION 'the dedupe rule is absent. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%No em dashes, anywhere, ever%' THEN
    RAISE EXCEPTION 'the em-dash rule was lost. ROLLING BACK.';
  END IF;
  RAISE NOTICE 'OK: rule 1 rewritten, 3 rules appended, em-dash rule intact (% chars).', length(t);
END $$;

COMMIT;
