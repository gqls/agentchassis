-- 348_cap_the_truncating_and_the_uncapped_steps_ROLLBACK.sql — hand-run
-- companion, NOT a migration (uppercase suffix: the runner excludes it).
--
-- Value-matched per path, so a cap anyone has since re-chosen survives.
--
-- GROUP A/B restore the PREVIOUS cap (they had one). GROUP C DELETES the key,
-- since those steps had no ai_service block at all — restoring them means
-- removing the block this file created, and the block is then empty, so it is
-- removed too rather than left as `"ai_service": {}`.
--
-- ⚠ AFTER RUNNING THIS, RE-RUN THE MIRROR or council-gate keeps the 16000
-- seats while fix-proposer drops to 8000 — the exact roster drift the mirror
-- exists to prevent:
--   python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py --apply

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent(t, '348 ROLLBACK: pre-revert')
  FROM unnest(ARRAY[
    'tool-auditor','page-content-writer','fix-proposer',
    'content-creator-about','content-creator-contact','content-creator-cta',
    'content-creator-features','content-creator-hero',
    'content-creator-hero-without-research','content-creator-testimonials',
    'content_researcher','simple-content-writer-with-approval','visual-designer'
  ]) AS t;

DO $revert$
DECLARE
    r      record;
    v_ai   text[];
    v_rows int;
BEGIN
    -- GROUP A + B: restore the prior cap, only where 348's value is still there
    FOR r IN
        SELECT * FROM (VALUES
            ('tool-auditor',        ARRAY['workflow','steps','llm_audit','config'],                   16000,  4000),
            ('page-content-writer', ARRAY['workflow','steps','process_sections_loop','config',
                                          'sub_workflow','steps','generate_content','config'],        16000,  8000),
            ('fix-proposer',        ARRAY['workflow','steps','review_guardian','config'],             16000,  8000),
            ('fix-proposer',        ARRAY['workflow','steps','review_render_guardian','config'],      16000,  8000),
            ('fix-proposer',        ARRAY['workflow','steps','review_llm_reliability','config'],      16000,  8000),
            ('fix-proposer',        ARRAY['workflow','steps','review_bug_historian','config'],        16000,  8000),
            ('fix-proposer',        ARRAY['workflow','steps','review_tooling_provenance','config'],   16000,  8000),
            ('fix-proposer',        ARRAY['workflow','steps','review_improvement_guardian','config'], 16000,  8000)
        ) AS t(agent_type, cfg_path, applied_cap, prior_cap)
    LOOP
        v_ai := r.cfg_path || 'ai_service'::text;
        UPDATE agent_definitions
           SET default_config = jsonb_set(default_config, v_ai || 'max_tokens'::text,
                                          to_jsonb(r.prior_cap), true),
               updated_at = now()
         WHERE type = r.agent_type
           AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
           AND (default_config #>> (v_ai || 'max_tokens'::text))::int = r.applied_cap;
        GET DIAGNOSTICS v_rows = ROW_COUNT;
        RAISE NOTICE '348 ROLLBACK: %/% -> % (% row(s))', r.agent_type, r.cfg_path[3], r.prior_cap, v_rows;
    END LOOP;

    -- GROUP C: remove the ai_service block this file created (only if it is
    -- exactly {"max_tokens": 8000} — anything richer means someone built on it)
    FOR r IN
        SELECT * FROM (VALUES
            ('content-creator-about',                 ARRAY['workflow','steps','generate_about_content','config']),
            ('content-creator-contact',               ARRAY['workflow','steps','generate_contact_content','config']),
            ('content-creator-contact',               ARRAY['workflow','steps','generate_content','config']),
            ('content-creator-cta',                   ARRAY['workflow','steps','generate_content','config']),
            ('content-creator-features',              ARRAY['workflow','steps','generate_content','config']),
            ('content-creator-hero',                  ARRAY['workflow','steps','generate_hero_content','config']),
            ('content-creator-hero-without-research', ARRAY['workflow','steps','generate_hero_content','config']),
            ('content-creator-testimonials',          ARRAY['workflow','steps','generate_content','config']),
            ('content_researcher',                    ARRAY['workflow','steps','process','config']),
            ('simple-content-writer-with-approval',   ARRAY['workflow','steps','generate_draft','config']),
            ('visual-designer',                       ARRAY['workflow','steps','design','config'])
        ) AS t(agent_type, cfg_path)
    LOOP
        v_ai := r.cfg_path || 'ai_service'::text;
        UPDATE agent_definitions
           SET default_config = default_config #- v_ai,
               updated_at = now()
         WHERE type = r.agent_type
           AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
           AND default_config #> v_ai = jsonb_build_object('max_tokens', 8000);
        GET DIAGNOSTICS v_rows = ROW_COUNT;
        RAISE NOTICE '348 ROLLBACK: %/% ai_service removed (% row(s))', r.agent_type, r.cfg_path[3], v_rows;
    END LOOP;
END;
$revert$;

DO $verify$
DECLARE v_count int;
BEGIN
    SELECT count(*) INTO v_count FROM agent_definitions
     WHERE type = 'tool-auditor' AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
       AND (default_config #>> '{workflow,steps,llm_audit,config,ai_service,max_tokens}')::int = 4000;
    IF v_count < 1 THEN RAISE EXCEPTION '348 ROLLBACK: tool-auditor not back at 4000'; END IF;

    SELECT count(*) INTO v_count
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type = 'fix-proposer' AND ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
       AND s.key IN ('review_guardian','review_render_guardian','review_llm_reliability',
                     'review_bug_historian','review_tooling_provenance','review_improvement_guardian')
       AND (s.value->'config'->'ai_service'->>'max_tokens')::int = 8000;
    IF v_count <> 6 THEN
        RAISE EXCEPTION '348 ROLLBACK: expected 6 fix-proposer seats back at 8000, found %', v_count;
    END IF;

    -- and the uncapped population is back to what 348 found (11), proving the
    -- revert really restored the pre-state rather than half of it
    SELECT count(*) INTO v_count FROM (
      SELECT jsonb_path_query(ad.default_config,'$.** ? (@.action == "execute_llm_prompt")') AS node,
             ad.default_config #>> '{ai_service,max_tokens}' AS root_cap
        FROM agent_definitions ad
       WHERE ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL) q
     WHERE node->'config'->'ai_service'->>'max_tokens' IS NULL
       AND node->'config'->>'max_tokens' IS NULL AND root_cap IS NULL;
    IF v_count <> 11 THEN
        RAISE EXCEPTION '348 ROLLBACK: expected 11 uncapped steps restored, found %', v_count;
    END IF;
END;
$verify$;

COMMIT;
