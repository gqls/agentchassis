-- 348_cap_the_truncating_and_the_uncapped_steps.sql
--
-- bugs_open/205 follow-through, part 2. Three groups, one theme: a cap is an
-- insurance ceiling, not a budget — you pay for tokens generated, so a
-- too-small cap buys nothing and costs a silent truncation.
--
-- GROUP A — steps that have DEMONSTRABLY truncated and are still at the cap
-- they truncated at. Counted from `error_message ILIKE '%stop_reason=max_tokens%'`
-- (the ONLY reliable signal — a truncated row has output_tokens NULL):
--   * tool-auditor/llm_audit 4000 -> 16000. NINE truncations 2026-07-21..08-08,
--     the most recent recovering 15,181 chars at a 4,000-token cap, i.e. it was
--     still producing text when cut. sonnet-4-6.
--   * page-content-writer generate_content 8000 -> 16000. This step is NESTED
--     inside process_sections_loop's sub_workflow (path below). Truncated
--     2026-08-09 recovering only 4,229 chars from an 8,000 budget — on sonnet-5
--     the cap is a THINKING + TEXT budget (bugs_closed/138's cap-120 lesson at
--     production scale), and its healthy runs already peak at 5,713–6,453
--     output tokens, so 8000 had almost no margin.
--
-- GROUP B — the six council seats that truncated at 8000 and are still there,
-- brought level with the four siblings ALREADY at 16000 (review_editquality,
-- review_guidelines, review_prior_art, review_architecture). Written to
-- `fix-proposer` ONLY: CLAUDE.md forbids hand-patching the gate, so
-- 099_SYNC_gate_roster.py mirrors these into council-gate immediately after
-- this file. Mirror drift was verified "(none)" before this change, so the
-- mirror will carry exactly these six steps and nothing else.
--
-- GROUP C — eleven steps that are UNCAPPED and were invisible to the census in
-- bugs_open/205, which required the step to already carry an `ai_service`
-- block (`s.value->'config' ? 'ai_service'`). A step with
-- `action: execute_llm_prompt` and NO ai_service block at all is equally
-- uncapped and equally falls to anthropic.go's hardcoded 2048. The
-- depth-aware, block-agnostic census finds 134 LLM nodes fleet-wide and
-- ELEVEN uncapped, not the 0 this lane reported on 2026-08-09. All eleven are
-- dormant (zero rows in llm_call_log, which reaches back to 2026-03-25) —
-- capped at the fleet mode 8000 because dormant-and-uncapped is exactly how
-- 205 started. Correction recorded in the bug file and WRONG_CALLS.
--
-- NOT DONE HERE, deliberately: feature-designer carries the same review_*
-- seats but has NOT truncated at them (its only truncations were
-- review_editquality, already raised) — no evidence, no change.
--
-- ROLLBACK: 348_..._ROLLBACK.sql, value-matched per path.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent(t, '348_cap_the_truncating_and_the_uncapped_steps.sql: pre-update')
  FROM unnest(ARRAY[
    'tool-auditor','page-content-writer','fix-proposer',
    'content-creator-about','content-creator-contact','content-creator-cta',
    'content-creator-features','content-creator-hero',
    'content-creator-hero-without-research','content-creator-testimonials',
    'content_researcher','simple-content-writer-with-approval','visual-designer'
  ]) AS t;

DO $apply$
DECLARE
    r        record;
    v_rows   int;
    v_total  int := 0;
    v_ai     text[];
BEGIN
    FOR r IN
        SELECT * FROM (VALUES
            -- GROUP A: demonstrated truncations
            ('tool-auditor',                          ARRAY['workflow','steps','llm_audit','config'],                  16000),
            ('page-content-writer',                   ARRAY['workflow','steps','process_sections_loop','config',
                                                            'sub_workflow','steps','generate_content','config'],       16000),
            -- GROUP B: council seats that truncated at 8000 (fix-proposer only; mirror follows)
            ('fix-proposer',   ARRAY['workflow','steps','review_guardian','config'],             16000),
            ('fix-proposer',   ARRAY['workflow','steps','review_render_guardian','config'],      16000),
            ('fix-proposer',   ARRAY['workflow','steps','review_llm_reliability','config'],      16000),
            ('fix-proposer',   ARRAY['workflow','steps','review_bug_historian','config'],        16000),
            ('fix-proposer',   ARRAY['workflow','steps','review_tooling_provenance','config'],   16000),
            ('fix-proposer',   ARRAY['workflow','steps','review_improvement_guardian','config'], 16000),
            -- GROUP C: uncapped, no ai_service block at all
            ('content-creator-about',                 ARRAY['workflow','steps','generate_about_content','config'],   8000),
            ('content-creator-contact',               ARRAY['workflow','steps','generate_contact_content','config'], 8000),
            ('content-creator-contact',               ARRAY['workflow','steps','generate_content','config'],         8000),
            ('content-creator-cta',                   ARRAY['workflow','steps','generate_content','config'],         8000),
            ('content-creator-features',              ARRAY['workflow','steps','generate_content','config'],         8000),
            ('content-creator-hero',                  ARRAY['workflow','steps','generate_hero_content','config'],    8000),
            ('content-creator-hero-without-research', ARRAY['workflow','steps','generate_hero_content','config'],    8000),
            ('content-creator-testimonials',          ARRAY['workflow','steps','generate_content','config'],         8000),
            ('content_researcher',                    ARRAY['workflow','steps','process','config'],                  8000),
            ('simple-content-writer-with-approval',   ARRAY['workflow','steps','generate_draft','config'],           8000),
            ('visual-designer',                       ARRAY['workflow','steps','design','config'],                   8000)
        ) AS t(agent_type, cfg_path, cap)
    LOOP
        -- NOTE the explicit ::text cast. `text[] || 'literal'` makes Postgres
        -- parse the literal as an ARRAY literal and fail with "malformed array
        -- literal" — it cannot tell append-element from concat-array.
        v_ai := r.cfg_path || 'ai_service'::text;

        -- Merge into any existing ai_service block rather than replacing it:
        -- GROUP C's steps have no block at all, GROUP A/B's carry a model that
        -- must survive. create_missing=true creates `ai_service` under an
        -- existing `config`.
        UPDATE agent_definitions
           SET default_config = jsonb_set(default_config, v_ai,
                 COALESCE(default_config #> v_ai, '{}'::jsonb)
                   || jsonb_build_object('max_tokens', r.cap), true),
               updated_at = now()
         WHERE type = r.agent_type
           AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
           AND default_config #> r.cfg_path IS NOT NULL
           AND COALESCE((default_config #>> (v_ai || 'max_tokens'::text))::int, 0) <> r.cap;

        GET DIAGNOSTICS v_rows = ROW_COUNT;
        v_total := v_total + v_rows;
        RAISE NOTICE '348: %/% -> % (% row(s))', r.agent_type, r.cfg_path[3], r.cap, v_rows;

        -- every active row carrying this step must now hold the cap as a NUMBER
        -- (the resolver type-asserts float64; a string silently falls through)
        SELECT count(*) INTO v_rows
          FROM agent_definitions
         WHERE type = r.agent_type
           AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
           AND default_config #> r.cfg_path IS NOT NULL
           AND (jsonb_typeof(default_config #> (v_ai || 'max_tokens'::text)) <> 'number'
                OR (default_config #>> (v_ai || 'max_tokens'::text))::int <> r.cap);
        IF v_rows <> 0 THEN
            RAISE EXCEPTION '348: %/% left % row(s) without a numeric cap of %',
                r.agent_type, r.cfg_path[3], v_rows, r.cap;
        END IF;
    END LOOP;

    RAISE NOTICE '348: % row-updates applied', v_total;
END;
$apply$;

DO $verify$
DECLARE v_count int;
BEGIN
    -- THE census, depth-aware and block-agnostic: no execute_llm_prompt node at
    -- ANY nesting depth may lack a cap. This is the check the 2026-08-09 version
    -- should have been; it returned 0 then only because it required an
    -- ai_service block to exist.
    SELECT count(*) INTO v_count FROM (
      SELECT jsonb_path_query(ad.default_config,
               '$.** ? (@.action == "execute_llm_prompt")') AS node,
             ad.default_config #>> '{ai_service,max_tokens}' AS root_cap
        FROM agent_definitions ad
       WHERE ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL) q
     WHERE node->'config'->'ai_service'->>'max_tokens' IS NULL
       AND node->'config'->>'max_tokens' IS NULL
       AND root_cap IS NULL;
    IF v_count <> 0 THEN
        RAISE EXCEPTION '348: % LLM step(s) still uncapped at some depth', v_count;
    END IF;

    -- nothing may sit at the 2048-adjacent danger zone by accident: assert the
    -- two GROUP A steps specifically, by value
    SELECT count(*) INTO v_count FROM agent_definitions
     WHERE type = 'tool-auditor' AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
       AND (default_config #>> '{workflow,steps,llm_audit,config,ai_service,max_tokens}')::int = 16000;
    IF v_count < 1 THEN RAISE EXCEPTION '348: tool-auditor/llm_audit not at 16000'; END IF;

    SELECT count(*) INTO v_count FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
       AND (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,ai_service,max_tokens}')::int = 16000;
    IF v_count < 1 THEN RAISE EXCEPTION '348: page-content-writer nested generate_content not at 16000'; END IF;

    -- GROUP B is written to fix-proposer ONLY; council-gate is the mirror's job
    SELECT count(*) INTO v_count
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type = 'fix-proposer' AND ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
       AND s.key IN ('review_guardian','review_render_guardian','review_llm_reliability',
                     'review_bug_historian','review_tooling_provenance','review_improvement_guardian')
       AND (s.value->'config'->'ai_service'->>'max_tokens')::int = 16000;
    IF v_count <> 6 THEN
        RAISE EXCEPTION '348: expected 6 fix-proposer seats at 16000, found %', v_count;
    END IF;
END;
$verify$;

COMMIT;

-- AFTER THIS FILE, RUN THE MIRROR (not optional — GROUP B is inert in the gate
-- until it runs, and hand-patching council-gate is forbidden):
--   python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py           # dry run: expect exactly the 6 seats as drift
--   python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py --apply
