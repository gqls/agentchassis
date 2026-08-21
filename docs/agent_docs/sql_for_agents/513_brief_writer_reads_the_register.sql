-- 513 — the brief writer READS the positioning register.
--
-- WHY HERE AND NOT IN THE CLASSIFIER. `RFC_037`'s ruling was that the register is
-- fed to the classifier as advisory input. Working it through
-- (`PLAN_2026-08-19_one_flow_three_brief_sources.md` addendum §C) changed the
-- answer, and the deciding argument came from a LATER owner decision rather than
-- from architecture:
--
--   *"I'd like to briefly look at each one. I may have a few words of direction on
--    many of them that I'd like to add to or change."*
--
-- The owner reads every BRIEF. The brief is the human-readable artefact he
-- reviews; a classifier's prompt input is not. **Positioning that lands in the
-- brief is positioning he can see and correct. Positioning fed straight to the
-- classifier is invisible to him** — it would shape the site with no point at
-- which a human could disagree with it. For an estate whose whole premise is that
-- 1,500 domains are 1,500 different businesses, putting the differentiation where
-- the owner can edit it is worth more than putting it one step earlier.
--
-- Three supporting reasons: REACH (the classifier is one agent; the brief is
-- inherited by the strategist, briefing agent and planner through the specs
-- derived from it, which is the gap RFC_037 §4 itself names); RISK (RFC_037's
-- change adds an input to `classify_and_extract`, a shared seam every fleet site
-- passes through — this is a new agent nothing depends on yet); and ONE READER
-- (two consumers of one table with different shapes is the drift class
-- `099_SYNC_gate_roster.py` exists to prevent).
--
-- **What that gives up, stated because it is a real loss:** the classifier no
-- longer sees its siblings AS siblings, so a BINDING collision check still needs
-- RFC_037. That RFC should stay open as the home for the binding half — and at
-- 1,500 briefs, review-on-collision is the only sampling rule that scales, so it
-- will be wanted.
--
-- ── THE TWO THINGS THIS MIGRATION MUST GET RIGHT ─────────────────────────────
--
-- 1. **INERT FOR A DOMAIN WITH NO ENTRY.** RFC_037 §5 question 4 requires it, and
--    it is most of the estate today: 189 rows exist against ~1,500 domains, and
--    the 50 test domains are deliberately entry-free. The query returns zero rows
--    and the prompt is told explicitly what to do with nothing. It must never
--    fail closed, and it must never invent a position.
--
-- 2. **`register_entry` MUST BE IN `write_brief`'s `input_fields`.** A template
--    variable without its `input_fields` entry renders EMPTY and errors nothing
--    (LANDMINES) — the migration would apply, the prompt would look right, and the
--    register would silently reach nothing. The verify block below asserts the
--    field is listed, which is the only cheap protection against that.
--
-- SIBLINGS ARE DERIVED, NOT READ FROM THE PARSED `neighbours` COLUMN. At 44
-- entries neighbours were hand-named; at 1,500 they cannot be (RFC_037's open
-- question). Deriving them from `family` is a rule that scales, and it is bounded
-- by LIMIT 12 because a large family would otherwise flood the prompt — a real
-- risk once the estate is loaded.
--
-- Rollback: 513_brief_writer_reads_the_register_ROLLBACK.sql

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        jsonb_set(
          default_config,
          '{workflow,steps,read_register}',
          jsonb_build_object(
            'action', 'query_database',
            'description', 'Read this domain''s positioning entry and its family siblings. Returns ZERO rows for a domain with no entry, which is most of the estate — the prompt is told what to do with nothing.',
            'config', jsonb_build_object(
              'query',
              'SELECT r.entry_code, r.family, r.proposition, r.audience, r.stage, r.mode, r.stance, '
              || 'r.attribution, r.raw_md, '
              || '(SELECT jsonb_agg(jsonb_build_object(''domain'', n.domain, ''entry'', n.entry_code, '
              || '''proposition'', n.proposition, ''audience'', n.audience, ''mode'', n.mode) ORDER BY n.entry_code) '
              || 'FROM (SELECT * FROM positioning_register s WHERE s.family = r.family '
              || 'AND lower(s.domain) <> lower(r.domain) AND s.is_primary '
              || 'ORDER BY s.entry_code LIMIT 12) n) AS siblings '
              || 'FROM positioning_register r WHERE lower(r.domain) = lower($1)',
              'params', jsonb_build_array('input_data.domain'),
              'output_format', 'array'),
            'output_field', 'register_entry',
            'next_step', 'search_web'),
          true),
        '{workflow,steps,read_specs,next_step}',
        '"read_register"'::jsonb, true),
      '{workflow,steps,write_brief,config,input_fields}',
      '["input_data","site_specs","register_entry","search_results","scrape_results","prepared_urls"]'::jsonb,
      true),
    updated_at = now()
WHERE type = 'brief-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- The prompt gains a positioning section, inserted on a verbatim anchor so a
-- reworded prompt cannot be silently left unchanged (the verify block checks it
-- took).
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,write_brief,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,write_brief,config,prompt_template}',
        '## Research: what this subject actually contains',
'## Its place in the portfolio — READ THIS BEFORE YOU DECIDE ANYTHING

{{.register_entry}}

**If that is empty, this domain has no recorded position.** Say so in `open_questions` and write the best brief the subject supports. Do NOT invent a position for it, and do NOT assume it must differ from anything — most of the estate has no entry yet, and an absent entry means unknown, not unconstrained.

**If it is present**, it is the portfolio''s own reasoning about why this domain exists as a separate business, and it OUTRANKS your reading of the domain name. Two duties follow:

1. **Stay inside the proposition.** The `proposition`, `audience` and `mode` are the position this site is meant to occupy. Write toward them. Where `raw_md` is present it is the authoritative text and carries the argument the typed fields only summarise — read it.
2. **Do not drift into a sibling.** `siblings` are other sites in the same family, each with its own proposition and audience. Your brief must not propose content that is plainly THEIR job. If the obvious thing to write is a sibling''s subject, that is the signal to find this site''s own angle, not to write theirs — two of our sites competing for one search result is the exact failure the register exists to prevent. Where you deliberately stay off a sibling''s ground, say so in `differentiation`.

**If the entry and the subject genuinely conflict** — the research says the domain is about one thing and the entry says another — do not quietly pick one. Write the brief to the entry and put the conflict in `open_questions`. A person reads every one of these and that is exactly the kind of thing they need to see.

## Research: what this subject actually contains'
      )),
      true),
    updated_at = now()
WHERE type = 'brief-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,write_brief,config,prompt_template}'
      LIKE '%## Research: what this subject actually contains%';

DO $$
DECLARE
    n_steps int; n_fields int; nxt text; has_section int; n_params int;
BEGIN
    SELECT count(*) INTO n_steps FROM agent_definitions,
         LATERAL jsonb_object_keys(default_config->'workflow'->'steps') k
     WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_steps <> 9 THEN RAISE EXCEPTION 'expected 9 steps after adding read_register, found %', n_steps; END IF;

    SELECT default_config #>> '{workflow,steps,read_specs,next_step}' INTO nxt
      FROM agent_definitions WHERE type='brief-writer' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF nxt <> 'read_register' THEN
      RAISE EXCEPTION 'read_specs still chains to % — the new step is orphaned and would never run', nxt;
    END IF;

    -- THE ONE THAT MATTERS: a template variable without its input_fields entry
    -- renders EMPTY and errors nothing. Assert it is listed.
    SELECT count(*) INTO n_fields FROM agent_definitions,
         LATERAL jsonb_array_elements_text(default_config #> '{workflow,steps,write_brief,config,input_fields}') f
     WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND f = 'register_entry';
    IF n_fields <> 1 THEN
      RAISE EXCEPTION 'register_entry is NOT in write_brief.input_fields — the prompt would render it empty and say nothing';
    END IF;

    SELECT count(*) INTO has_section FROM agent_definitions
     WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,write_brief,config,prompt_template}'
           LIKE '%Its place in the portfolio%';
    IF has_section <> 1 THEN RAISE EXCEPTION 'the positioning section did not reach the prompt'; END IF;

    SELECT jsonb_array_length(default_config #> '{workflow,steps,read_register,config,params}') INTO n_params
      FROM agent_definitions WHERE type='brief-writer' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_params <> 1 THEN RAISE EXCEPTION 'read_register expects 1 param ($1 = domain), found %', n_params; END IF;

    RAISE NOTICE '513 OK — brief-writer reads the register: 9 steps, read_specs->read_register, register_entry in input_fields, prompt section present';
END $$;

COMMIT;
