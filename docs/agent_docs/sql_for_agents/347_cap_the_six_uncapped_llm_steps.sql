-- 347_cap_the_six_uncapped_llm_steps.sql
--
-- bugs_open/205 follow-through (owner decision 2026-08-09): the six remaining
-- LLM steps with no max_tokens at any level get explicit caps, so no active
-- step falls to the anthropic.go hardcoded 2048 fallback.
--
-- WHY THESE NUMBERS (measured 2026-08-09, evidence in
-- docs024_key_docs_latest/bugfix_205_poison_pill_reaper/NOTES + bug file):
--   * A cap is an insurance ceiling, not a budget — tokens are paid per
--     generation, so over-generous costs nothing on success; under-generous is
--     bug 205 itself (silent truncation + retry loops). Rule: ~2x the biggest
--     plausible output, rounded to a fleet-standard value.
--   * site-architect/design 32000: design-class steps measured p95 14,456 /
--     17,352 and max 20,189 output tokens (llm_call_log, feature-designer +
--     pre-07-26 generic rows).
--   * chief-strategist/generate_build_plan 16000: never run; plan ~ design-
--     sized document; largest opus-4-6 cap in the fleet is 16000. The OLDER of
--     the two active rows already carries 8192 chosen 2025-11-15 — NOT touched
--     (see the IS NULL scoping below).
--   * content-creator/create_content 16000 (both active rows): long-form
--     output is the product; haiku-4-5, worst case ~8p/call.
--   * brand-designer/analyze_brand, domain-analyst/analyze 8000: analysis-
--     shaped, fleet mode.
--   * provocation-gate-calibration/gate 8000: verdict-shaped BUT sonnet-5
--     thinks server-side into output_tokens (bugs_open/138's cap-120 lesson:
--     never a small cap on a thinking model). Fleet sonnet-5 mode = 8000
--     (40 steps).
--   * med-price-collector/scrape_prices is OUT OF SCOPE: it never passes
--     ai_actions.go — its action hardcodes num_predict:500 (measured max
--     output ever ~150-200 tokens, ~3x headroom).
--
-- MECHANISM: the resolver (ai_actions.go) reads top-level
-- default_config.max_tokens FIRST (asserted NULL below, else this write would
-- be shadowed), then the root-ai_service/step-ai_service overlay. The write
-- path config.ai_service.max_tokens is the one proven live by the
-- extract_and_reconcile cap (2026-08-08, output 3135 at cap 8000 first try).
-- Every UPDATE is scoped to rows where the path is NULL: caps the uncapped,
-- never overwrites a chosen value.
--
-- ROLLBACK: 347_cap_the_six_uncapped_llm_steps_ROLLBACK.sql deletes exactly
-- the keys this file set (value-matched, so the pre-existing 8192 survives).
-- Config is live immediately; no image roll involved.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('brand-designer',               '347_cap_the_six_uncapped_llm_steps.sql: pre-update');
SELECT snapshot_agent('chief-strategist',             '347_cap_the_six_uncapped_llm_steps.sql: pre-update');
SELECT snapshot_agent('content-creator',              '347_cap_the_six_uncapped_llm_steps.sql: pre-update');
SELECT snapshot_agent('domain-analyst',               '347_cap_the_six_uncapped_llm_steps.sql: pre-update');
SELECT snapshot_agent('provocation-gate-calibration', '347_cap_the_six_uncapped_llm_steps.sql: pre-update');
SELECT snapshot_agent('site-architect',               '347_cap_the_six_uncapped_llm_steps.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,analyze_brand,config,ai_service,max_tokens}', to_jsonb(8000::int), true),
       updated_at = now()
 WHERE type = 'brand-designer'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,analyze_brand,config,ai_service}' IS NOT NULL
   AND default_config #>> '{workflow,steps,analyze_brand,config,ai_service,max_tokens}' IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,generate_build_plan,config,ai_service,max_tokens}', to_jsonb(16000::int), true),
       updated_at = now()
 WHERE type = 'chief-strategist'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,generate_build_plan,config,ai_service}' IS NOT NULL
   AND default_config #>> '{workflow,steps,generate_build_plan,config,ai_service,max_tokens}' IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,create_content,config,ai_service,max_tokens}', to_jsonb(16000::int), true),
       updated_at = now()
 WHERE type = 'content-creator'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,create_content,config,ai_service}' IS NOT NULL
   AND default_config #>> '{workflow,steps,create_content,config,ai_service,max_tokens}' IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,analyze,config,ai_service,max_tokens}', to_jsonb(8000::int), true),
       updated_at = now()
 WHERE type = 'domain-analyst'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,analyze,config,ai_service}' IS NOT NULL
   AND default_config #>> '{workflow,steps,analyze,config,ai_service,max_tokens}' IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,gate,config,ai_service,max_tokens}', to_jsonb(8000::int), true),
       updated_at = now()
 WHERE type = 'provocation-gate-calibration'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,gate,config,ai_service}' IS NOT NULL
   AND default_config #>> '{workflow,steps,gate,config,ai_service,max_tokens}' IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,design,config,ai_service,max_tokens}', to_jsonb(32000::int), true),
       updated_at = now()
 WHERE type = 'site-architect'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,design,config,ai_service}' IS NOT NULL
   AND default_config #>> '{workflow,steps,design,config,ai_service,max_tokens}' IS NULL;

-- Post-conditions. Counts measured 2026-08-09 pre-apply:
-- brand-designer 1, chief-strategist 1-at-16000 (+1 pre-existing 8192),
-- content-creator 2, domain-analyst 1, provocation-gate-calibration 1,
-- site-architect 1; top-level default_config.max_tokens NULL on all 8 rows.
DO $verify$
DECLARE
    v_count int;
    v_expected record;
BEGIN
    FOR v_expected IN
        SELECT * FROM (VALUES
            ('brand-designer',               'analyze_brand',       8000::numeric, 1),
            ('chief-strategist',             'generate_build_plan', 16000,         1),
            ('content-creator',              'create_content',      16000,         2),
            ('domain-analyst',               'analyze',             8000,          1),
            ('provocation-gate-calibration', 'gate',                8000,          1),
            ('site-architect',               'design',              32000,         1)
        ) AS t(agent_type, step_name, cap, expected_rows)
    LOOP
        SELECT count(*) INTO v_count
          FROM agent_definitions
         WHERE type = v_expected.agent_type
           AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
           AND jsonb_typeof(default_config
                 #> ('{workflow,steps,' || v_expected.step_name || ',config,ai_service,max_tokens}')::text[]) = 'number'
           AND (default_config
                 #>> ('{workflow,steps,' || v_expected.step_name || ',config,ai_service,max_tokens}')::text[])::numeric
               = v_expected.cap;
        IF v_count <> v_expected.expected_rows THEN
            RAISE EXCEPTION '347: %/% expected % row(s) at cap %, found %',
                v_expected.agent_type, v_expected.step_name,
                v_expected.expected_rows, v_expected.cap, v_count;
        END IF;
    END LOOP;

    -- the pre-existing chief-strategist 8192 must have survived untouched
    SELECT count(*) INTO v_count
      FROM agent_definitions
     WHERE type = 'chief-strategist'
       AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
       AND (default_config #>> '{workflow,steps,generate_build_plan,config,ai_service,max_tokens}')::numeric = 8192;
    IF v_count <> 1 THEN
        RAISE EXCEPTION '347: the pre-existing chief-strategist 8192 cap was disturbed (found % rows)', v_count;
    END IF;

    -- no top-level max_tokens may shadow the step caps (resolver reads it first)
    SELECT count(*) INTO v_count
      FROM agent_definitions
     WHERE type IN ('brand-designer','chief-strategist','content-creator',
                    'domain-analyst','provocation-gate-calibration','site-architect')
       AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
       AND default_config ? 'max_tokens';
    IF v_count <> 0 THEN
        RAISE EXCEPTION '347: % row(s) carry a top-level max_tokens that would shadow the step cap', v_count;
    END IF;

    -- the fleet-wide uncapped census (bugs_open/205) must now be empty
    SELECT count(*) INTO v_count
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
       AND s.value->'config' ? 'ai_service'
       AND s.value->'config'->>'max_tokens' IS NULL
       AND s.value->'config'->'ai_service'->>'max_tokens' IS NULL
       AND ad.default_config #>> '{ai_service,max_tokens}' IS NULL;
    IF v_count <> 0 THEN
        RAISE EXCEPTION '347: uncapped-step census still returns % step(s) after apply', v_count;
    END IF;
END;
$verify$;

COMMIT;
