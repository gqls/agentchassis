-- FILE: 304_forbid_markdown_in_text_fields.sql
--
-- bugs_open/184 companion to migration 303 (the detection half). Extends
-- STRICT RULE 9 of page-content-writer's generate_content prompt in place:
-- the live block forbids HTML wrapping in text fields and a markdown wrapper
-- around the JSON envelope (rule 3), but nothing forbids markdown INSIDE a
-- field value — the confirmed gap. Rule 9 extended rather than a rule 19
-- appended: the instruction belongs next to the type it governs, and
-- renumbering 10-18 would mean rewriting the whole block (the scoped-replace
-- mechanism is the owner-approved APPLY_voice_v4 precedent,
-- docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/APPLY_voice_v4_page_content_writer.sql).
--
-- Measured 2026-08-03: content-writer and simple-content-writer-with-approval
-- carry neither STRICT RULES nor the save_page_sections path — not patched.
--
-- RUN IT:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db < <this file>

\set ON_ERROR_STOP on
BEGIN;

DROP TABLE IF EXISTS bak_agent_definitions_pcw_no_markdown;
CREATE TABLE bak_agent_definitions_pcw_no_markdown AS
SELECT * FROM agent_definitions WHERE type = 'page-content-writer';

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
      '9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields.',
      '9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields. Plain string also means NO markdown syntax: no **emphasis**, no `code spans`, no # headings, no bullet markers. Markdown is never rendered — the asterisks, backticks and hashes would reach the site visitor as literal characters. If a phrase needs emphasis, choose words that carry it. This holds for every field type: where rule 10 says HTML, write HTML, markdown syntax is never correct in any output field.'
    ) AS tmpl
  FROM cur
)
UPDATE agent_definitions a
SET default_config = jsonb_set(
      a.default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(r.tmpl)),
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

  IF t NOT LIKE '%Plain string also means NO markdown syntax%' THEN
    RAISE EXCEPTION 'the markdown prohibition is absent — the replace did not match. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%10. For fields of type `rich_text` or `content`%' THEN
    RAISE EXCEPTION 'rule 10 was lost. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%NEVER invent specific statistics%' THEN
    RAISE EXCEPTION 'rule 14 was lost. ROLLING BACK.';
  END IF;
  RAISE NOTICE 'OK: rule 9 extended, rules 10 and 14 intact (% chars).', length(t);
END $$;

COMMIT;
