-- ROLLBACK for 540 — restore content-feed-orchestrator's `complete` step to
-- `output_fields` list mode, removing `result_mapping` and `commit_sha`.
--
-- WHEN YOU WOULD RUN THIS: if the news-vs-rss choice turns out to be wrong for
-- a real downstream consumer (e.g. something specifically needs the RSS
-- commit and cannot get it from rss_commit_result directly for some reason).
-- Restores the EXACT prior list; refuses if result_mapping is not the exact
-- one 540 wrote.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'content-feed-orchestrator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '540 ROLLBACK: no live content-feed-orchestrator complete step to roll back';
    END IF;
    IF NOT (cfg ? 'result_mapping') THEN
        RAISE EXCEPTION '540 ROLLBACK: complete carries no result_mapping — 540 is not applied, or already rolled back';
    END IF;
    IF cfg->'result_mapping' <> jsonb_build_object(
        'seed_result',        'seed_result',
        'dispatch_result',    'dispatch_result',
        'triage_result',      'triage_result',
        'news_render_result', 'news_render_result',
        'news_commit_result', 'news_commit_result',
        'rss_render_result',  'rss_render_result',
        'rss_commit_result',  'rss_commit_result',
        'commit_sha',         'news_commit_result.response.data.commit_sha'
    ) THEN
        RAISE EXCEPTION '540 ROLLBACK: result_mapping is %, not exactly what 540 wrote — someone else owns this now; do not remove it', cfg->'result_mapping';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,result_mapping}',
           '{workflow,steps,complete,config,output_fields}',
           '["seed_result","dispatch_result","triage_result","news_render_result","news_commit_result","rss_render_result","rss_commit_result"]'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'content-feed-orchestrator'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','complete','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'content-feed-orchestrator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'result_mapping' THEN
        RAISE EXCEPTION '540 ROLLBACK VERIFY: result_mapping is still present: %', cfg->'result_mapping';
    END IF;
    IF cfg->'output_fields' <> '["seed_result","dispatch_result","triage_result","news_render_result","news_commit_result","rss_render_result","rss_commit_result"]'::jsonb THEN
        RAISE EXCEPTION '540 ROLLBACK VERIFY: output_fields is %, want the original list', cfg->'output_fields';
    END IF;
    RAISE NOTICE '540 ROLLBACK OK: complete restored to output_fields list mode';
END $$;

COMMIT;
