-- 510 — the BRIEF WRITER: generate a comprehensive mission brief for a domain,
--       then HOLD for the owner to read and edit before anything is built.
--
-- WHY (owner rulings, 2026-08-19/20). Flow A is ruled: a site is always built from
-- a brief. But *"for my thousands of domains I will not have time to write a brief
-- for each"*, so the brief needs a third producer beside the owner and a
-- third-party customer. This is that producer.
--
-- THE TWO RULINGS THAT SHAPE THE PROMPT, and they pull in opposite directions
-- from what I first proposed:
--
--   1. *"The spec is aspirational, the plan is achievable."* The brief must NOT be
--      trimmed to what the platform can build today. Constraining it would bake a
--      snapshot of the platform into a document meant to outlive it, and would
--      delete the evidence of what we are missing — which, across ~1,500 briefs,
--      is the best-evidenced roadmap this estate could produce. The PLANNER
--      degrades (it already receives the live component catalogue via its
--      `load_components` step); the brief aims.
--
--   2. *"I'd like to briefly look at each one. I may have a few words of direction
--      on many of them that I'd like to add to or change."* So this is
--      generate-then-EDIT, not generate-then-approve. The brief is persisted as a
--      normal `mission_brief` spec, which means the owner's edit supersedes it in
--      `site_specs` with the generated version preserved underneath — the history
--      of what the machine proposed versus what he changed is kept for free.
--
-- HOW THE HOLD WORKS, and why it is `status` and not `approval_mode`.
-- `site_work_items.approval_mode` is the richer mechanism and the dispatcher
-- honours it (`load_work_item_actions.go:709`: a non-`auto` item is withheld until
-- `status='approved'`). **Proven in both directions 2026-08-20** against the
-- dispatcher's own predicate: `manual`+`triaged` is not returned, and flipping to
-- `approved` returns it. BUT `create_work_item` has no `approval_mode` config key,
-- so an agent cannot set it without a Go change — and Go is inert until a roll.
-- `status` IS a supported key, and the dispatcher only picks
-- `status IN ('triaged','approved')`, so `status='needs_human_review'` holds the
-- item today with no code at all. It is also the estate's existing HITL idiom
-- (`directory_citation_unverified` does exactly this). Adding `approval_mode` to
-- `create_work_item` is the better long-term shape and is recorded as a residual.
--
-- WHAT IT DOES NOT DO, stated so nobody assumes it:
--   * it does not read the positioning register (that is the next piece — the
--     addendum to PLAN_2026-08-19 argues the brief-writer should be the register's
--     reader, replacing RFC_037's classifier input, because the owner reads briefs
--     and so can correct positioning that lands there);
--   * it does not screen a third-party brief for security or reasonability (a
--     separate owner requirement, and a different job: this WRITES a brief, that
--     one VETS an incoming one);
--   * nothing dispatches it automatically. `domain-submitter` is untouched, so a
--     hand-written mission still wins and nothing changes for existing flows. A
--     `needs_brief` item is created deliberately, which matches "one at a time,
--     supervised".
--
-- Rollback: 510_brief_writer_agent_ROLLBACK.sql

BEGIN;

INSERT INTO agent_definitions (type, display_name, category, description, default_config, is_active)
SELECT
  'brief-writer',
  'Brief Writer',
  'specialist',
  'Generates a comprehensive, deliberately aspirational mission brief for a domain, then holds for the owner to read and edit before the build starts.',
  jsonb_build_object(
    'input_schema', jsonb_build_object('required', jsonb_build_array('site_id', 'domain')),
    'workflow', jsonb_build_object(
      'start_step', 'read_specs',
      'steps', jsonb_build_object(

        'read_specs', jsonb_build_object(
          'action', 'read_site_spec',
          'description', 'Load whatever already exists for this site (a thin third-party brief, an earlier submission). The writer ENRICHES rather than overwrites when there is something there.',
          'config', jsonb_build_object('site_id', 'input_data.site_id'),
          'output_field', 'site_specs',
          'next_step', 'search_web'),

        'search_web', jsonb_build_object(
          'action', 'web_search',
          'description', 'What does this subject actually contain? A brief written from the domain string alone is a guess; this is what makes it specific.',
          'config', jsonb_build_object('query_from', 'input_data.research_query', 'num_results', 12),
          'output_field', 'search_results',
          'next_step', 'prepare_urls'),

        'prepare_urls', jsonb_build_object(
          'action', 'prepare_urls',
          'description', 'Pick the strongest sources. No prefer_domains: for a brief we want the shape of the subject, not authority on any one fact.',
          'config', jsonb_build_object(
            'max_scrapes', 5, 'max_snippets', 10,
            'exclude_domains', jsonb_build_array('pinterest.com','facebook.com','twitter.com','instagram.com','reddit.com','tiktok.com','linkedin.com')),
          'output_field', 'prepared_urls',
          'next_step', 'scrape_pages'),

        'scrape_pages', jsonb_build_object(
          'action', 'batch_webscrape',
          'description', 'markdown-only and main-content-only: bugs_closed/062 — a batch reply over the broker max is dropped and the caller starves.',
          'config', jsonb_build_object(
            'urls_field', 'prepared_urls.urls_to_scrape',
            'scrape_config', jsonb_build_object('formats', jsonb_build_array('markdown'), 'only_main_content', true, 'capture_screenshot', false),
            'timeout_seconds', 240),
          'output_field', 'scrape_results',
          'next_step', 'write_brief'),

        'write_brief', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Write the brief. Deliberately aspirational — the planner degrades, the brief aims.',
          'config', jsonb_build_object(
            'ai_service', jsonb_build_object('provider','anthropic','model','claude-sonnet-4-6','max_tokens',8000,'api_key_env_var','ANTHROPIC_API_KEY'),
            'input_fields', jsonb_build_array('input_data','site_specs','search_results','scrape_results','prepared_urls'),
            'output_format', 'json',
            'prompt_template',
'You are writing the MISSION BRIEF for a website that does not exist yet. The brief is the founding document: every later agent — classifier, strategist, planner, writers — reads it and builds from it.

## The domain
{{.input_data.domain}}

## What the owner asked for, if anything
{{.input_data.direction}}

## Anything already recorded for this site
{{.site_specs}}

## Research: what this subject actually contains
{{.scrape_results}}

Supporting snippets:
{{.prepared_urls.snippet_context}}

## THE ONE RULE THAT SHAPES EVERYTHING ELSE

**The spec is aspirational; the plan is achievable.** Describe the site this domain SHOULD be, richly and specifically. Do NOT trim your ambition to what you imagine a platform can build — a later planning step decides what is buildable today and records what it could not do. Under-reaching here is the expensive mistake, because it silently becomes the ceiling. Over-reaching costs nothing: it is written down and either built later or not.

So: be generous and concrete. A brief that says "informational content about the topic" is a failure. A brief that names eight specific things a reader would come for, and what each one is, is what this is for.

## BE SPECIFIC TO THIS SUBJECT

Everything must come from the research above or from what the domain name genuinely implies. Do not write a brief that would fit any domain with the nouns swapped — that is the failure mode here, and it is what makes 1,500 sites collapse into one repeated 1,500 times. If the research is thin, say so in `research_quality` and lower your confidence rather than padding with generalities.

## WHAT A RICH SITE CAN CONTAIN — draw on these where they genuinely fit, and ignore the ones that do not

- **guides and explainers** — the durable questions a newcomer to this subject has
- **editorial / opinion** — a stance, a house view, something a person would disagree with
- **research and analysis** — original comparison, aggregation, or measurement
- **a directory** — a list of the named providers, products, brands, venues, models or organisations in this field, with a fact per entry
- **tools** — anything with inputs and a computed answer: a calculator, a selector, a scorer, a checker, a converter, a generator, a quiz. Interactive things are among the strongest reasons to visit a site.
- **games** — playable, if the subject supports one
- **news** — only where the subject genuinely changes often enough to sustain it
- **data and reference tables** — the numbers people look up

Pick what suits THIS subject. A subject with no news should not get a news section, and saying so is more useful than inventing one.

## MUST-NOTS

State what this site must never do or claim. Be concrete. If the subject is regulated (money, credit, insurance, mortgages, investments, health, legal advice), say plainly that the site is NOT a regulated firm and must not present as one, must not advise, arrange, broker or introduce, and must not assert an authorisation number or a provider panel. That is a legal position and no site adopts it without explicit instruction.

## Return ONLY valid JSON

{
  "proposition": "one paragraph: what this site is, for whom, and why it deserves to exist",
  "audience": {"primary": "specific — who, at what moment, with what question", "secondary": "or null"},
  "reader_intent": ["the 4-8 things a visitor actually wants, most common first"],
  "stance": "the site''s point of view and voice, in a sentence — what it is FOR and what it is against",
  "content_plan": [
    {"kind": "guide|editorial|research|directory|tool|game|news|data", "name": "specific name", "what": "1-2 sentences: what it contains and why a reader wants it", "priority": "core|valuable|aspirational"}
  ],
  "directory_opportunity": "if a directory of named things fits, say WHAT the entries would be and what fact each would carry; else null",
  "tool_opportunities": [{"name": "...", "input": "what the reader gives it", "output": "what it computes or decides", "why": "why that is worth using"}],
  "differentiation": "what makes this site distinct from the obvious competitors the research surfaced — name them",
  "must_nots": ["concrete prohibitions"],
  "regulated_subject": true/false,
  "research_quality": "thin|adequate|good — and one clause saying why",
  "confidence": 0.0,
  "open_questions": ["what a human should decide that you could not"]
}

Mark `priority` honestly: `core` is what the site is pointless without, `valuable` earns its place, `aspirational` is the reach. The planner will build core first, and `open_questions` is where you put anything you would rather a person settled — the owner reads every one of these.'),
          'output_field', 'brief',
          'next_step', 'persist_brief'),

        'persist_brief', jsonb_build_object(
          'action', 'write_site_spec',
          'description', 'Persist as mission_brief — the SAME aspect a hand-written brief uses, so nothing downstream changes and the owner''s later edit supersedes this with the generated version preserved underneath.',
          'config', jsonb_build_object(
            'aspect', 'mission_brief',
            'site_id', 'input_data.site_id',
            'spec_data', 'brief.result',
            'source', 'brief-writer',
            'source_agent', 'brief-writer',
            'source_item_id', 'input_data.work_item_id'),
          'output_field', 'brief_stored',
          'next_step', 'create_review_item'),

        'create_review_item', jsonb_build_object(
          'action', 'create_work_item',
          'description', 'HOLD for the owner. status=needs_human_review keeps it out of the dispatcher (which selects status IN (triaged,approved)) — proven 2026-08-20. Approving means "the brief is right"; the operator then flips this to triaged to release the build.',
          'config', jsonb_build_object(
            'site_id', 'input_data.site_id',
            'source', 'brief-writer',
            'item_pipeline', 'build',
            'item_type', 'needs_brief_review',
            'severity', 'medium',
            'priority', 5,
            'handler_agent', 'human-review',
            'status', 'needs_human_review',
            'item_key_prefix', 'brief_review',
            'summary', 'Read the generated brief before the build starts — add or change any direction, then release',
            'spec_literal', jsonb_build_object(
              'check', 'generated_brief_review',
              'aspect', 'mission_brief',
              'how_to_read', 'SELECT jsonb_pretty(data) FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain=''<domain>'' AND ss.aspect=''mission_brief'' AND ss.is_current;',
              'how_to_edit', 'Write a NEW mission_brief spec (supersede the current one). site_specs keeps the generated version underneath, so the diff between what the machine proposed and what you changed survives.',
              'how_to_release', 'Create the needs_domain_research item for this site (handler domain-research-classifier), or flip this item to triaged if a held research item already exists.')),
          'output_field', 'review_item',
          'next_step', 'complete'),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Brief written and held for review',
          'config', jsonb_build_object('result_from', 'review_item'))
      )
    )
  ),
  true
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'brief-writer' AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL);

-- ── verification, as DO/RAISE ────────────────────────────────────────────────
-- A verify block of SELECTs cannot stop a COMMIT: ON_ERROR_STOP ignores a
-- non-empty result set (LANDMINES). Every check RAISEs.
DO $$
DECLARE
    n_agent int; n_steps int; n_fields int; n_hold text; n_aspect text;
BEGIN
    SELECT count(*) INTO n_agent FROM agent_definitions
     WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_agent <> 1 THEN RAISE EXCEPTION 'expected exactly 1 active brief-writer, found %', n_agent; END IF;

    SELECT count(*) INTO n_steps FROM agent_definitions,
         LATERAL jsonb_object_keys(default_config->'workflow'->'steps') k
     WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_steps <> 8 THEN RAISE EXCEPTION 'expected 8 steps, found %', n_steps; END IF;

    -- Every template variable the prompt reads MUST be in input_fields, or it
    -- renders EMPTY and errors nothing (LANDMINES). Assert the list here.
    SELECT count(*) INTO n_fields FROM agent_definitions,
         LATERAL jsonb_array_elements_text(default_config #> '{workflow,steps,write_brief,config,input_fields}') f
     WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND f IN ('input_data','site_specs','search_results','scrape_results','prepared_urls');
    IF n_fields <> 5 THEN RAISE EXCEPTION 'write_brief input_fields incomplete: % of 5 present', n_fields; END IF;

    -- The hold is the whole review design; assert it rather than trust it.
    SELECT default_config #>> '{workflow,steps,create_review_item,config,status}' INTO n_hold
      FROM agent_definitions WHERE type='brief-writer' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_hold <> 'needs_human_review' THEN
      RAISE EXCEPTION 'review item status is %, must be needs_human_review or the build is not held', n_hold;
    END IF;

    SELECT default_config #>> '{workflow,steps,persist_brief,config,aspect}' INTO n_aspect
      FROM agent_definitions WHERE type='brief-writer' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_aspect <> 'mission_brief' THEN
      RAISE EXCEPTION 'brief is written to aspect %, but downstream agents read mission_brief', n_aspect;
    END IF;

    RAISE NOTICE '510 OK — brief-writer seeded: 8 steps, 5 input_fields, hold=%, aspect=%', n_hold, n_aspect;
END $$;

COMMIT;
