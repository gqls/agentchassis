-- 540 — content-feed-orchestrator's `complete` step exposes `commit_sha` at a
--       CANONICAL top-level key, converting `output_fields` to
--       `result_mapping`. RFC_029 §10.13 step 5's real gate (bdl/`commit_sha`).
--       CONFIG ONLY. Found by the `staged-component-build` lane's `537` wire
--       guard on its first dry run — a TENTH handler neither this lane's
--       structural census nor its earlier empirical cross-check caught.
--
-- ============================================================================
-- WHY BOTH EARLIER METHODS MISSED THIS ONE, AND WHY THE WIRE'S GUARD DID NOT
-- ============================================================================
-- This lane's structural census (git_commit step ∩ live handler_agent, used to
-- find 519-528/534) DID see content-feed-orchestrator's two git_commit steps —
-- but it was scoped OUT deliberately at the time, on the grounds that its
-- handler_agent volume was negligible (1 item, all-time, per that day's read)
-- and its two-commit shape (below) made it a genuine judgement call rather
-- than a mechanical one, so it was set aside rather than decided in a hurry.
--
-- The peer lane's later empirical cross-check (site_work_items.result ?
-- 'commit_sha') ALSO missed it, for the opposite reason: it has never
-- RECORDED a commit_sha in that column (zero occurrences), because nothing
-- has wired it there yet — "who records one today" cannot see a handler that
-- is correctly silent today but CAN produce a commit.
--
-- The wire's own apply-guard (migration 537, `staged-component-build`) asks a
-- third, better question — "does this handler's OWN `orchestration_states`
-- tree ever carry a `commit_sha`" — and that is what caught it: [MEASURED]
-- 3 of 3 orchestrations in the last 30 days carry one.
--
-- ============================================================================
-- WHY THIS ONE IS A DIFFERENT SHAPE OF JUDGEMENT CALL FROM 523/527
-- ============================================================================
-- section-editor and webdesign-agent each have ONE commit that is "this item's
-- own deliverable" and a SECOND, later call that is a downstream side-effect
-- (a deployment trigger) — the choice there was mechanical once you know which
-- is which. Here there is no side-effect call: **both `commit_news` and
-- `commit_rss` are genuine, independent deliverables of the SAME item**, each
-- separately conditional on whether that feed had anything to publish
-- (`check_has_news`/`check_has_rss` gate on `*_render_result.item_count`, and
-- either can be skipped independently).
--
-- [MEASURED 2026-08-21, all 3 orchestrations in the 30-day window]:
--
-- | run | news_commit_result | rss_commit_result |
-- |---|---|---|
-- | 1 | sha `277f6965…` | absent (0 rss items) |
-- | 2 | sha `08b96f49…` | sha `b5c1059a…` — TWO DIFFERENT REAL COMMITS |
-- | 3 | sha `1870d438…` | absent (0 rss items) |
--
-- **`news_commit_result` is present in 3/3; `rss_commit_result` in 1/3.** News
-- is the primary, near-always-present deliverable; RSS is a secondary one that
-- is frequently skipped. This migration maps `commit_sha` to
-- `news_commit_result.response.data.commit_sha` on that basis.
--
-- ============================================================================
-- THE DISCLOSED LOSS — stated plainly, not hidden
-- ============================================================================
-- **On a run where BOTH feeds commit (row 2 above), `commit_sha` will report
-- only the NEWS sha; the RSS sha is real, live, and NOT represented in this
-- field.** No canonical single-field choice avoids this — the item genuinely
-- has two deliverables and `commit_sha` is a single value. This is a real,
-- acknowledged information loss for the rarer dual-commit case, traded for
-- covering the far more common single-commit case correctly. If a future
-- consumer needs BOTH shas, `rss_commit_result.response.data.commit_sha`
-- remains reachable directly (rss_commit_result is still exposed, identity-
-- mapped, unchanged) — it is simply not what bare `commit_sha` means here.
--
-- ============================================================================
-- `complete_no_sources` — untouched, correctly
-- ============================================================================
-- Reached when there are no sources at all (upstream of both commits);
-- exposes only `["seed_result"]` today and never has a commit to report.
-- Absence of `commit_sha` there is correct; this file does not touch it.
--
-- Same mechanism as migration 519 (output_fields cannot rename a nested path;
-- the fix is a result_mapping CONVERSION) — not repeated here.
--
-- ROLLBACK: 540_content_feed_orchestrator_exposes_commit_sha_canonically_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('content-feed-orchestrator', '540_content_feed_orchestrator_exposes_commit_sha_canonically: pre-update');

DO $$
DECLARE
    n    int;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'content-feed-orchestrator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '540: expected exactly 1 live content-feed-orchestrator row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','complete'] INTO step
      FROM agent_definitions
     WHERE type = 'content-feed-orchestrator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '540: content-feed-orchestrator has no complete step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'complete_workflow' THEN
        RAISE EXCEPTION '540: complete runs %, not complete_workflow', step->>'action';
    END IF;
    IF step->'config' ? 'result_mapping' THEN
        RAISE EXCEPTION '540: complete ALREADY carries result_mapping (%) — already applied or superseded; do not overwrite', step->'config'->'result_mapping';
    END IF;
    IF step->'config'->'output_fields' <> '["seed_result","dispatch_result","triage_result","news_render_result","news_commit_result","rss_render_result","rss_commit_result"]'::jsonb THEN
        RAISE EXCEPTION '540: complete.output_fields is %, want exactly the seven-entry list — re-derive the mapping against the current list', step->'config'->'output_fields';
    END IF;

    IF (SELECT default_config #>> ARRAY['workflow','steps','commit_news','action']
          FROM agent_definitions WHERE type='content-feed-orchestrator' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'git_commit' THEN
        RAISE EXCEPTION '540: commit_news no longer runs git_commit — re-measure before applying';
    END IF;
    IF (SELECT default_config #>> ARRAY['workflow','steps','commit_news','output_field']
          FROM agent_definitions WHERE type='content-feed-orchestrator' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'news_commit_result' THEN
        RAISE EXCEPTION '540: commit_news''s output_field is no longer news_commit_result — re-measure the mapped path before applying';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config #- '{workflow,steps,complete,config,output_fields}',
           '{workflow,steps,complete,config,result_mapping}',
           jsonb_build_object(
               'seed_result',        'seed_result',
               'dispatch_result',    'dispatch_result',
               'triage_result',      'triage_result',
               'news_render_result', 'news_render_result',
               'news_commit_result', 'news_commit_result',
               'rss_render_result',  'rss_render_result',
               'rss_commit_result',  'rss_commit_result',
               'commit_sha',         'news_commit_result.response.data.commit_sha'
           ),
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

    IF cfg ? 'output_fields' THEN
        RAISE EXCEPTION '540 VERIFY: output_fields still present after conversion: %', cfg->'output_fields';
    END IF;
    IF cfg->'result_mapping'->>'seed_result' IS DISTINCT FROM 'seed_result' THEN RAISE EXCEPTION '540 VERIFY: seed_result mapping wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'dispatch_result' IS DISTINCT FROM 'dispatch_result' THEN RAISE EXCEPTION '540 VERIFY: dispatch_result mapping wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'triage_result' IS DISTINCT FROM 'triage_result' THEN RAISE EXCEPTION '540 VERIFY: triage_result mapping wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'news_render_result' IS DISTINCT FROM 'news_render_result' THEN RAISE EXCEPTION '540 VERIFY: news_render_result mapping wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'news_commit_result' IS DISTINCT FROM 'news_commit_result' THEN RAISE EXCEPTION '540 VERIFY: news_commit_result mapping wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'rss_render_result' IS DISTINCT FROM 'rss_render_result' THEN RAISE EXCEPTION '540 VERIFY: rss_render_result mapping wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'rss_commit_result' IS DISTINCT FROM 'rss_commit_result' THEN RAISE EXCEPTION '540 VERIFY: rss_commit_result mapping wrong: %', cfg->'result_mapping'; END IF;
    IF cfg->'result_mapping'->>'commit_sha' IS DISTINCT FROM 'news_commit_result.response.data.commit_sha' THEN
        RAISE EXCEPTION '540 VERIFY: commit_sha mapping missing or wrong: %', cfg->'result_mapping';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(cfg->'result_mapping')) <> 8 THEN
        RAISE EXCEPTION '540 VERIFY: result_mapping has an unexpected key count: %', cfg->'result_mapping';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM orchestration_states
         WHERE owner_agent_type = 'content-feed-orchestrator'
           AND collected_data #> '{news_commit_result,response,data}' ? 'commit_sha'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION '540 VERIFY: no live orchestration confirms the mapped path resolves — the migration may be correct but is unconfirmed against real data';
    END IF;

    RAISE NOTICE '540 OK: content-feed-orchestrator.complete exposes all 7 prior fields (unchanged) and commit_sha (new, sourced from news_commit_result — the RSS commit, when it also runs, is not represented in this field; see the file header) via result_mapping';
END $$;

COMMIT;
