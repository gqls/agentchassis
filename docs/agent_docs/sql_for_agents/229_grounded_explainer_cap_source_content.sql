-- ============================================================================
-- 229_grounded_explainer_cap_source_content.sql
--
-- The grounded-explainer lane found the right source and then choked on it.
--
-- WHAT HAPPENED (corr 2c5bbf90, oufe.com, "The Thames Water restructuring")
--   Search and selection worked exactly as intended. It reached the primary
--   instrument, not commentary about it:
--     judiciary.uk/…/Kington-S.A.R.L.-Thames-Water-…-judgment.pdf
--       → "Neutral Citation Number: [2025] EWCA Civ …"
--   The scrape succeeded. Then `extract_claims` failed and the run took the
--   `complete_no_sources` branch, producing nothing.
--
--   The cause is size. `{{.scrape_results}}` interpolates every scraped source
--   whole:
--     Thames run  : 584,152 chars of scraped content
--     working run : 320,692 chars (19 citations registered)
--   A full Court of Appeal judgment is not a web page, and the lane fed the
--   entire thing into one prompt.
--
-- THE FIX ALREADY EXISTED, WHICH IS THIS WEEK'S RECURRING LESSON
--   `format_research_content` (research_actions.go:204-320) takes
--   `max_content_per_source` and truncates each source before assembling one
--   LLM-ready block. It is the exact missing step. I listed it in my own
--   research notes when designing this lane — "format_research_content — Format
--   scraped content for LLM context" — and then wired scrape straight into the
--   extractor anyway.
--
--   Third time in a week: the capability was present, unused, and invisible
--   because nothing pointed at it.
--
-- THE HONEST LIMITATION THIS INTRODUCES
--   Capping at 24,000 chars/source means a long judgment is searched for
--   quotable claims across roughly its first fifth. Judgments front-load the
--   neutral citation, parties, issues and often a summary, so the opening is the
--   richest part for citable statements — but a claim buried at paragraph 180
--   will not be found, and the run will not say so.
--
--   That is a real ceiling on this lane, not a solved problem. Chunking a long
--   document and extracting per chunk is the actual answer; it needs a loop this
--   workflow does not have. Recorded rather than papered over.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- 1. new step between scrape and extract -------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,format_sources}',
      jsonb_build_object(
        'action', 'format_research_content',
        'config', jsonb_build_object(
          'scrape_field', 'scrape_results',
          'snippets_field', 'prepared_urls.snippet_context',
          'max_content_per_source', 24000
        ),
        'next_step', 'extract_claims',
        'output_field', 'research_content',
        'description', 'Cap each source before it reaches the extractor. A full judgment is not a web page: an uncapped run fed 584k chars into one prompt and produced nothing.'
      ),
      true
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,scrape_pages,next_step}',
      '"format_sources"'::jsonb, false),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

-- 2. extractor reads the capped block, not the raw scrape ---------------------
UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,extract_claims,config,input_fields}',
      '["research_content","prepared_urls","input_data"]'::jsonb, false),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,extract_claims,config,prompt_template}',
      to_jsonb(replace(
        default_config->'workflow'->'steps'->'extract_claims'->'config'->>'prompt_template',
        'SCRAPED SOURCES (each has a url):
{{.scrape_results}}',
        'SOURCES (each has a url; long documents are truncated, so quote only from what is actually shown here):
{{.research_content}}'
      )), false),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

COMMIT;

-- Verify
--   SELECT default_config->'workflow'->'steps'->'scrape_pages'->>'next_step' AS after_scrape,
--          default_config->'workflow'->'steps'->'format_sources'->'config'->>'max_content_per_source' AS cap,
--          (default_config->'workflow'->'steps'->'extract_claims'->'config'->>'prompt_template'
--             LIKE '%research_content%') AS extractor_reads_capped
--     FROM agent_definitions WHERE type='grounded-explainer' AND deleted_at IS NULL;
--
-- Then re-run the Thames query. Judge it on whether citations register, not on
-- whether the run completes: `complete_no_sources` is a legitimate outcome and
-- looks identical to a silent failure from outside.
