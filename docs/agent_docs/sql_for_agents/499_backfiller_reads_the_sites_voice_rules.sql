-- 499 — tell the meta-description writer what the site's voice gate BANS, before
--       it writes (bugs_open/320, SEO-004)
--
-- WHY. The gate in `save_page_meta_description` works — proven in production on the
-- very first scheduled run, 2026-08-20 06:52Z. `leopardessconsulting.co.uk`'s only
-- fillable page got this from the writer:
--
--     "Read evidence-based articles on AI adoption, trust gaps, and governance
--      across healthcare, finance, hiring and more."
--
-- and the action refused it, with the site's own rule quoted back:
--
--     reason: voice_tell
--     detail: banned_phrase: owner 2026-07-18: overused; say what is
--             checked/verified instead ("trust")
--
-- That is exactly the owner's condition for waiving the review pass, working. **This
-- migration does not weaken it.** The gate stays; the writer is simply told the rules
-- before it writes instead of after.
--
-- ── THE DEFECT THIS CLOSES: A PERMANENT HOURLY RETRY ────────────────────────
--
-- The backfiller is now on an hourly schedule (`498`). A page whose generated copy
-- keeps tripping the gate is never filled AND is re-selected every hour, paying for
-- one LLM call each time, for ever. Nothing about that is visible as a failure: the
-- orchestration COMPLETEs, the scheduled task stamps a clean run, and the page stays
-- blank. It is the same shape as the silent no-op this lane already fixed twice —
-- a green record over work that did not happen.
--
-- The population is bounded and known: `[MEASURED 2026-08-20]` **9 sites carry an
-- enabled voice gate**, with banned-phrase counts of 14 (leopardessconsulting), 10
-- (oufe), and 1 each for noted, relojistas, vetcomparison, loancash, lendzy,
-- gamesdesign and webdesign.uk. On the other ~18 sites this step is a no-op.
--
-- ── WHY THE PROMPT AND NOT A RETRY LOOP ─────────────────────────────────────
--
-- A retry loop would ask the same model the same question and hope. Telling it the
-- constraint is the fix that removes the cause. It also puts the rules where a
-- reviewer of the OUTPUT can see what the writer was working to, rather than in a
-- gate that only speaks when it refuses.
--
-- ⚠ THE RULES ARE REGEXES, and they are handed to the model AS REGEXES with their
-- reasons attached, deliberately un-prettified. Translating `\bharness(ing)?\b` into
-- "don't say harness" in SQL would be a second, hand-maintained rendering of a list
-- that already exists — the drift surface this estate keeps paying for. The reason
-- string is what carries the human meaning, and it is already written for a human
-- ("hype verb", "unless the stack is named in the same sentence").
--
-- A site with no voice spec produces an empty list and the prompt block is skipped by
-- the template's own `{{if}}`. Opt-in stays opt-in.
--
-- ROLLBACK: 499_backfiller_reads_the_sites_voice_rules_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('meta-description-backfiller', '499_voice_rules: pre-update');

DO $$
DECLARE n int; nxt text; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '499: expected exactly 1 live backfiller, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,ensure_site_record,next_step}',
         default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}'
    INTO nxt, p
    FROM agent_definitions WHERE type='meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF nxt IS DISTINCT FROM 'load_pages_missing_meta' THEN
    RAISE EXCEPTION '499: ensure_site_record.next_step is %, expected load_pages_missing_meta — the chain has changed under me', nxt;
  END IF;
  IF position('## House style' in p) = 0 THEN
    RAISE EXCEPTION '499: the house-style block is missing from the prompt — wrong agent state';
  END IF;
  IF position('voice_rules' in p) > 0 THEN
    RAISE EXCEPTION '499: already applied (prompt already references voice_rules)';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config =
         -- 1. new step, spliced between ensure_site_record and the page query
         jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,load_voice_rules}',
           jsonb_build_object(
             'action', 'query_database',
             'config', jsonb_build_object(
               'query',
                 'SELECT COALESCE(jsonb_agg(jsonb_build_object(''pattern'', ph->>''pattern'', ''reason'', ph->>''reason'')), ''[]''::jsonb) AS rules ' ||
                 'FROM site_specs ss ' ||
                 'LEFT JOIN LATERAL jsonb_array_elements(COALESCE(ss.data#>''{voice_gate,banned_phrases}'', ''[]''::jsonb)) ph ON true ' ||
                 'WHERE ss.site_id = $1 AND ss.aspect = ''voice'' AND ss.is_current = true ' ||
                 '  AND COALESCE((ss.data#>>''{voice_gate,enabled}'')::boolean, false) = true',
               'params', jsonb_build_array('site_record.site_id'),
               'output_format', 'object'
             ),
             'next_step', 'load_pages_missing_meta',
             'output_field', 'voice_rules',
             'description', 'The site''s own banned phrases, so the writer is told the rules BEFORE it writes rather than refused after (bugs_open/320). Empty list on a site with no gate.'
           )
         ),
           '{workflow,steps,ensure_site_record,next_step}',
           to_jsonb('load_voice_rules'::text)
         ),
       updated_at = now()
 WHERE type='meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. the prompt gains the rules block, and voice_rules joins input_fields
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,write_descriptions,config,input_fields}',
           '["pages_missing_meta","site_record","voice_rules"]'::jsonb
         ),
         '{workflow,steps,write_descriptions,config,prompt_template}',
         to_jsonb(
           replace(
             default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}',
             '## House style',
             E'{{if .voice_rules.rules}}## THIS SITE BANS THESE PHRASES — a description containing one is REFUSED and the page stays blank\n' ||
             E'Each line is a regular expression the site checks, with the reason it exists. Do not use the phrase, and do not work around it with a synonym that means the same claim.\n' ||
             E'{{range .voice_rules.rules}}- {{.pattern}}  ({{.reason}})\n' ||
             E'{{end}}\n' ||
             E'{{end}}## House style'
           )
         )
       ),
       updated_at = now()
 WHERE type='meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE cfg jsonb; p text; rules jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cfg#>>'{workflow,steps,ensure_site_record,next_step}' IS DISTINCT FROM 'load_voice_rules' THEN
    RAISE EXCEPTION '499 VERIFY: the chain was not rewired';
  END IF;
  IF cfg#>>'{workflow,steps,load_voice_rules,next_step}' IS DISTINCT FROM 'load_pages_missing_meta' THEN
    RAISE EXCEPTION '499 VERIFY: the new step does not hand on to the page query';
  END IF;
  IF cfg#>>'{workflow,steps,load_voice_rules,config,output_format}' IS DISTINCT FROM 'object' THEN
    RAISE EXCEPTION '499 VERIFY: load_voice_rules must use output_format object — array has no addressable keys (bugs_open/313)';
  END IF;

  p := cfg#>>'{workflow,steps,write_descriptions,config,prompt_template}';
  IF position('{{range .voice_rules.rules}}' in p) = 0 THEN
    RAISE EXCEPTION '499 VERIFY: the prompt does not iterate the rules';
  END IF;
  IF position('## House style' in p) = 0 THEN
    RAISE EXCEPTION '499 VERIFY: the house-style block was consumed rather than preserved';
  END IF;
  IF NOT (cfg#>'{workflow,steps,write_descriptions,config,input_fields}' @> '["voice_rules"]'::jsonb) THEN
    RAISE EXCEPTION '499 VERIFY: voice_rules is not in input_fields, so the template would render nothing';
  END IF;

  -- The query must actually return the shape the template iterates, on a site that
  -- HAS rules. Checked here rather than discovered as a blank block at run time.
  EXECUTE 'SELECT COALESCE(jsonb_agg(jsonb_build_object(''pattern'', ph->>''pattern'', ''reason'', ph->>''reason'')), ''[]''::jsonb) '
       || 'FROM site_specs ss LEFT JOIN LATERAL jsonb_array_elements(COALESCE(ss.data#>''{voice_gate,banned_phrases}'', ''[]''::jsonb)) ph ON true '
       || 'WHERE ss.site_id = (SELECT id FROM sites WHERE domain = ''leopardessconsulting.co.uk'') '
       || '  AND ss.aspect = ''voice'' AND ss.is_current = true'
    INTO rules;
  IF rules IS NULL OR jsonb_array_length(rules) < 10 THEN
    RAISE EXCEPTION '499 VERIFY: the rules query returned % entries for a site known to have 14 — the shape is wrong', COALESCE(jsonb_array_length(rules), -1);
  END IF;

  RAISE NOTICE '499 OK: the writer is now told the site''s banned phrases (% on the check site) before it writes', jsonb_array_length(rules);
END $$;

COMMIT;
