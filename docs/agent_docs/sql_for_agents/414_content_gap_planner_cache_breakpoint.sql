-- 414_content_gap_planner_cache_breakpoint.sql
--
-- Inserts the LCO-008 cache breakpoint into the `plan_gaps` prompt template of
-- content-gap-planner, so Anthropic prompt caching fires on the agent the
-- 2026-08-15 cacheTTL measurement identified as the single best candidate in
-- the fleet.
--
-- WHY THIS IS A ONE-LINE INSERT AND 377 WAS A REORDER. Migration 377 had to
-- HOIST council-gate's shared evidence block to the front, because each seat's
-- persona sat above it and pushed the shared body to a different offset —
-- measured common prefix ZERO. content-gap-planner needs no move: its template
-- already emits the shared material (site identity, content direction, existing
-- pages, available components) BEFORE the one per-call block
-- (`## Content Gap to Address`). The boundary is already in the right place;
-- all that is missing is the marker that tells the client where it is.
--
-- THE MEASUREMENT, taken 2026-08-15 against llm_call_log, 3 days, 404 calls:
--
--   input tokens   1,730,777   (vs 178,363 output — 91% input, uncached)
--   cache reads             0   no marker, so the client sends the prompt whole
--   avg prompt         14,910 chars / 4,284 input tokens
--
-- Position of '## Content Gap to Address' in the RENDERED prompt, i.e. exactly
-- how much becomes cacheable, grouped by (boundary, total length):
--
--   boundary  total   calls   distinct prefixes at boundary
--     11,481  15,448    149   1
--     10,107  14,074    116   1
--     11,447  15,414     83   1
--     10,191  14,158     20   1
--     10,103  14,070     17   1
--     11,875  15,842      8   1
--
-- 393 of 404 calls sit in six groups, and DISTINCT PREFIXES AT THE BOUNDARY IS
-- 1 IN EVERY GROUP — the text above the boundary is byte-identical across every
-- call in a group, which is the precise precondition Anthropic prefix caching
-- requires. The remaining ~11 calls are singletons (a site seen once); each
-- pays a write it never reads back, and that cost is included in the arithmetic
-- below rather than excluded to flatter it.
--
-- WHY IT PAYS AT 1h AND WOULD NOT HAVE AT 5m. A write costs 2x base input at
-- the 1h TTL, a read 0.1x, so break-even is a ~53% hit rate. Measured
-- same-prefix call gaps for this agent: 1.0% within 5 minutes, 99.8% within an
-- hour. At the old 5-minute TTL this marker would have cost ~24% MORE than not
-- caching at all — which is why it was deliberately NOT added until the TTL
-- flip went live on v1.0.1301 (stamp 0115f2b45). This migration is the second
-- half of that change and is inert without it.
--
-- Expected effect on the ~74% of each prompt that sits above the boundary:
-- 6 writes + 387 reads + 11 unpaired writes, against 404 full-rate sends —
-- roughly an 82% reduction on the cacheable portion, ~1.05M input tokens per
-- three days.
--
-- ⚠ THE FAILURE MODE IS SILENT AND LOOKS LIKE SUCCESS. A single per-call byte
-- above the marker (a timestamp, a UUID, a gap description) makes every call
-- write a fresh entry and read none — strictly worse than no caching. The guard
-- below asserts that NOTHING derived from `.input_data` appears above the
-- marker; that is the one post-condition that actually protects the spend.
--
-- IN-FLIGHT RUNS: SAFE. An orchestration executes from the copy of the step
-- config captured in orchestration_states.workflow_plan, so this edit cannot
-- reach a run already under way (established and measured under migration 377).
-- Runs created after this migration pick up the marker.
--
-- VERIFY AFTER APPLYING — a zero here is the failure mode, not the absence of
-- one. Assert a non-zero READ on the second and later calls:
--
--   SELECT created_at, input_tokens,
--          coalesce(cache_creation_input_tokens,0) AS writes,
--          coalesce(cache_read_input_tokens,0)     AS reads
--   FROM llm_call_log
--   WHERE agent_type='content-gap-planner' AND created_at > now() - interval '2 hours'
--   ORDER BY created_at;
--
-- ROLLBACK: run 414_content_gap_planner_cache_breakpoint_ROLLBACK.sql, which
-- deletes the marker substring. The snapshot below is the belt-and-braces copy.

SELECT snapshot_agent('content-gap-planner',
                      '414_content_gap_planner_cache_breakpoint.sql: pre-update');

BEGIN;

-- PRE-CONDITIONS. Asserted before the write, because `replace` is not
-- idempotent here: running this twice would insert a second marker, and the
-- client splits on the FIRST one, silently shrinking the cached prefix.
DO $$
DECLARE
    n_defs     integer;
    tmpl       text;
    n_heading  integer;
    n_marker   integer;
BEGIN
    SELECT count(*) INTO n_defs
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF n_defs <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 414: expected exactly 1 live content-gap-planner definition, found %', n_defs;
    END IF;

    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template'
      INTO tmpl
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF tmpl IS NULL THEN
        RAISE EXCEPTION 'MIGRATION 414: plan_gaps has no prompt_template — the step shape has changed';
    END IF;

    n_heading := (length(tmpl) - length(replace(tmpl, '## Content Gap to Address', ''))) / length('## Content Gap to Address');
    IF n_heading <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 414: anchor ''## Content Gap to Address'' occurs % times, expected exactly 1 — the template has been rewritten and the boundary must be re-located by hand', n_heading;
    END IF;

    n_marker := (length(tmpl) - length(replace(tmpl, '<!--CACHE_BREAKPOINT-->', ''))) / length('<!--CACHE_BREAKPOINT-->');
    IF n_marker <> 0 THEN
        RAISE EXCEPTION 'MIGRATION 414: template already carries % cache marker(s) — already applied, or applied by hand', n_marker;
    END IF;

    RAISE NOTICE 'migration 414 pre-conditions OK: 1 definition, 1 anchor, 0 existing markers';
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_gaps,config,prompt_template}',
        to_jsonb(
            replace(
                default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template',
                '## Content Gap to Address',
                E'<!--CACHE_BREAKPOINT-->\n## Content Gap to Address'
            )
        )
    ),
    version    = version + 1,
    updated_at = now()
WHERE type='content-gap-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- POST-CONDITIONS: the transformation is a pure insertion.
DO $$
DECLARE
    tmpl     text;
    n_marker integer;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template'
      INTO tmpl
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    n_marker := (length(tmpl) - length(replace(tmpl, '<!--CACHE_BREAKPOINT-->', ''))) / length('<!--CACHE_BREAKPOINT-->');
    IF n_marker <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 414: % markers after update, expected exactly 1 (the client splits on the first)', n_marker;
    END IF;

    -- Nothing was dropped: every placeholder the step renders must survive.
    IF tmpl NOT LIKE '%{{.site_record.domain}}%'
       OR tmpl NOT LIKE '%{{.available_components}}%' AND tmpl NOT LIKE '%available_components%'
       OR tmpl NOT LIKE '%{{.input_data.spec.description}}%'
       OR tmpl NOT LIKE '%{{.existing_pages.pages}}%' AND tmpl NOT LIKE '%existing_pages%' THEN
        RAISE EXCEPTION 'MIGRATION 414: a placeholder went missing — the insertion damaged the template';
    END IF;

    RAISE NOTICE 'migration 414 OK: marker inserted once, all placeholders intact';
END $$;

-- THE ONE THAT MATTERS MOST, separated so its failure message is unambiguous.
-- Everything above the marker is the cache key. If any per-call value renders
-- there, every call writes a new entry and reads none — a silent 2x on the
-- shared prefix, which is worse than the uncached status quo and shows up in no
-- status field. `.input_data` is this step's per-call payload (the gap
-- description, suggestion and category), so its presence above the marker is
-- exactly the disqualifying condition.
DO $$
DECLARE
    prefix text;
    suffix text;
BEGIN
    SELECT split_part(default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template',
                      '<!--CACHE_BREAKPOINT-->', 1),
           split_part(default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template',
                      '<!--CACHE_BREAKPOINT-->', 2)
      INTO prefix, suffix
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF position('.input_data' in prefix) > 0 THEN
        RAISE EXCEPTION
          'MIGRATION 414: the cacheable prefix references .input_data — every call would write its own cache entry and read none. Worse than no caching, and it would look like success.';
    END IF;

    IF position('.input_data' in suffix) = 0 THEN
        RAISE EXCEPTION
          'MIGRATION 414: the per-call suffix does NOT reference .input_data — the marker is below the varying block, so the cache key changes every call';
    END IF;

    IF position('## Available Section Components' in prefix) = 0 THEN
        RAISE EXCEPTION
          'MIGRATION 414: the components block is not above the marker — the prefix is far shorter than the 1024-token floor this model requires, and a prefix below the floor caches NOTHING while still being charged as a write';
    END IF;

    RAISE NOTICE 'migration 414 OK: cacheable prefix carries the shared blocks and no per-call value (% chars of template, rendering to ~10-12k)', length(prefix);
END $$;

COMMIT;
