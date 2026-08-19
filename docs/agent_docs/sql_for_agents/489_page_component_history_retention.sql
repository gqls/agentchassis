-- 489_page_component_history_retention.sql
--
-- bugs_closed/229 / register STY-056 open-review (a), now due: the volume watch
-- TRIPPED on 2026-08-19 — page_component_history 63MB (was 30MB on 08-10),
-- ~3.5MB/day against the design's ~0.9MB/day worst-case projection. The owner's
-- candidate-1 ruling deferred pruning "until page-side measurements exist";
-- nine days of them now do, and the driver is unambiguous: delete/machine_made
-- trigger-arm payloads are 75% of trigger rows (4,085 of 5,478 since 08-10).
-- Full design + measurements + reader census:
-- docs024_key_docs_latest/bugfix_229_page_component_archive/PLAN_2026-08-19_history_retention.md
--
-- WHAT THIS DOES. A daily scheduled task (the database-cleanup / pure-pre_query
-- shape, fire_message=false — the SQL is the whole mechanism, no agent) nulls
-- `rendered_html` on trigger-arm rows that are BOTH machine_made AND older than
-- 30 days. Nothing else: the ledger row, its content_data, its digest, its
-- slot/position identity all survive. One doc_notes row per run, ON ZERO TOO
-- (the WFA-013 precedent: a MISSING row means the task did not run, and must
-- never read as "nothing to prune").
--
-- WHAT IS NEVER TOUCHED (the preservation contract):
--   * hand_patched payloads — kept for ever (not reproducible; the whole reason
--     the trigger exists);
--   * unstamped payloads — kept, revisit at a later watch (not provably
--     reproducible; the class shrinks naturally as restamping proceeds);
--   * content_data on every row — the restore recipes' (migs 287/378/431)
--     recovery source;
--   * snapshot-source rows (save_page_sections_overwrite) — open-review (d)'s
--     question, deliberately not answered here;
--   * the chrome sibling site_component_history — 57 rows, no volume problem.
--
-- WHY machine_made IS THE DROPPABLE CLASS: the platform's own accepted policy.
-- The save-path snapshot has always COALESCE-dropped the artefact when
-- content_data exists (14,831/14,863 rows at design time); the trigger arm was
-- built to catch the DIVERGENT class, not to reverse that policy.
-- divergence='machine_made' (outgoing md5 == same-statement stamp) is the
-- council-reviewed marker for "exactly what a stamped writer last wrote".
--
-- THE ENABLEMENT RUNWAY IS STRUCTURAL: the trigger has only existed since
-- 2026-08-09, so the oldest trigger-arm row reaches 30 days on ~2026-09-08.
-- Enabled today, the task no-ops (and says so daily in doc_notes) for ~20 runs
-- before the first byte is nulled. Abort at any time before then for free:
--   UPDATE scheduled_tasks SET enabled=false
--    WHERE name='page-component-history-retention';
--
-- ROLLBACK (489_..._ROLLBACK.sql): removes the task + the index. Already-nulled
-- payloads are NOT recoverable — stated plainly; that is why the predicate
-- takes only the class the platform already declines to keep.

BEGIN;

-- 1. Partial index so the daily scan never walks the whole (growing) table.
--    Matches the retention predicate exactly; rows leave the index as they are
--    pruned, so it stays the size of the un-pruned machine_made backlog.
CREATE INDEX IF NOT EXISTS idx_pch_retention_candidates
    ON page_component_history (created_at)
    WHERE source = 'artefact_archive_trigger'
      AND divergence = 'machine_made'
      AND rendered_html IS NOT NULL;

-- 2. The task. Pure pre_query (fire_message=false), the database-cleanup shape:
--    the scheduler executes the SQL on its tick; the final SELECT always
--    returns a row so the run is marked executed.
INSERT INTO scheduled_tasks
    (name, description, interval_seconds, target_agent_type, input_data,
     pre_query, enabled, fire_message, timeout_seconds)
VALUES
    ('page-component-history-retention',
     'Nulls machine_made trigger-arm rendered_html payloads older than 30 days in page_component_history (STY-056 open-review (a); bugs_closed/229). Never touches hand_patched/unstamped payloads, content_data, ledger rows, or snapshot-source rows. Writes one doc_notes row per run, on zero too — a MISSING row means this task did not run.',
     86400,
     'page-component-history-retention',
     '{}'::jsonb,
     $pre$
    -- Freed bytes must be read BEFORE the UPDATE: RETURNING sees the NEW row,
    -- where rendered_html is already NULL.
    WITH candidates AS (
        SELECT id, length(rendered_html) AS freed
        FROM page_component_history
        WHERE source = 'artefact_archive_trigger'
          AND divergence = 'machine_made'
          AND created_at < NOW() - INTERVAL '30 days'
          AND rendered_html IS NOT NULL
        FOR UPDATE
    ),
    pruned AS (
        UPDATE page_component_history h
        SET rendered_html = NULL
        FROM candidates c
        WHERE h.id = c.id
        RETURNING h.id
    ),
    note AS (
        INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
        SELECT 'pipeline',
               'page-component-history-retention',
               'retention run: ' || (SELECT count(*) FROM pruned)
                 || ' machine_made trigger-arm payloads >30d nulled, ~'
                 || COALESCE((SELECT sum(freed) FROM candidates), 0)
                 || ' bytes freed. hand_patched/unstamped/content_data/ledger untouched (STY-056 (a)).',
               '["retention","page-component-archive"]'::jsonb,
               'scheduled_task',
               'page-component-history-retention'
        RETURNING id
    )
    SELECT (SELECT count(*) FROM pruned)::text AS payloads_nulled,
           (SELECT COALESCE(sum(freed),0) FROM candidates)::text AS bytes_freed,
           (SELECT count(*) FROM note)::text AS note_written
    -- Always returns a row so the scheduler marks the task executed
     $pre$,
     true,
     false,
     600)
ON CONFLICT (name) DO NOTHING;

-- 3. Probe (DO/RAISE — a verify block of SELECTs cannot stop the COMMIT).
--    Three synthetic rows on a real page FK, the retention UPDATE verbatim,
--    and the assertion that it prunes EXACTLY the old machine_made payload.
DO $probe$
DECLARE
    v_page_id uuid;
    v_site_id uuid;
    n integer;
    v_html text;
    v_cd jsonb;
BEGIN
    SELECT p.id, p.site_id INTO v_page_id, v_site_id
    FROM pages p LIMIT 1;
    IF v_page_id IS NULL THEN
        RAISE EXCEPTION 'mig489 probe: no pages row available for FK';
    END IF;

    INSERT INTO page_component_history
        (page_id, site_id, content_data, source, rendered_html, divergence, op, slot_name, position, created_at)
    VALUES
        (v_page_id, v_site_id, '{"probe":"old-machine"}',  'artefact_archive_trigger', 'PROBE-OLD-MACHINE',  'machine_made', 'delete', 'mig489_probe', 1, NOW() - INTERVAL '40 days'),
        (v_page_id, v_site_id, '{"probe":"old-patched"}',  'artefact_archive_trigger', 'PROBE-OLD-PATCHED',  'hand_patched', 'delete', 'mig489_probe', 2, NOW() - INTERVAL '40 days'),
        (v_page_id, v_site_id, '{"probe":"new-machine"}',  'artefact_archive_trigger', 'PROBE-NEW-MACHINE',  'machine_made', 'delete', 'mig489_probe', 3, NOW() - INTERVAL '1 day');

    -- The retention predicate, verbatim from the task's pre_query.
    UPDATE page_component_history
    SET rendered_html = NULL
    WHERE source = 'artefact_archive_trigger'
      AND divergence = 'machine_made'
      AND created_at < NOW() - INTERVAL '30 days'
      AND rendered_html IS NOT NULL;

    -- Old machine_made: payload gone, row + content_data intact.
    SELECT rendered_html, content_data INTO v_html, v_cd
    FROM page_component_history
    WHERE slot_name = 'mig489_probe' AND position = 1;
    IF v_html IS NOT NULL OR v_cd IS DISTINCT FROM '{"probe":"old-machine"}'::jsonb THEN
        RAISE EXCEPTION 'mig489 probe: old machine_made row wrong after prune (html %, cd %)', v_html, v_cd;
    END IF;

    -- Old hand_patched: untouched.
    SELECT rendered_html INTO v_html
    FROM page_component_history
    WHERE slot_name = 'mig489_probe' AND position = 2;
    IF v_html IS DISTINCT FROM 'PROBE-OLD-PATCHED' THEN
        RAISE EXCEPTION 'mig489 probe: hand_patched payload was touched (%)', v_html;
    END IF;

    -- Recent machine_made: untouched.
    SELECT rendered_html INTO v_html
    FROM page_component_history
    WHERE slot_name = 'mig489_probe' AND position = 3;
    IF v_html IS DISTINCT FROM 'PROBE-NEW-MACHINE' THEN
        RAISE EXCEPTION 'mig489 probe: recent machine_made payload was touched (%)', v_html;
    END IF;

    -- Self-clean.
    DELETE FROM page_component_history WHERE slot_name = 'mig489_probe';

    -- Task row landed and is enabled.
    SELECT count(*) INTO n FROM scheduled_tasks
    WHERE name = 'page-component-history-retention' AND enabled AND fire_message = false;
    IF n <> 1 THEN
        RAISE EXCEPTION 'mig489 probe: expected 1 enabled fire_message=false task row, found %', n;
    END IF;
END $probe$;

COMMIT;
