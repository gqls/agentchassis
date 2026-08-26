-- ROLLBACK for 642 — restore the unconditional 3-day-only due test on
-- sitemap-refresh-rotation, removing the early-due (changed-since-render +
-- 30-minute quiet) branch and its comment block.
--
-- Leaves 622's has-deployed-pages guard, the locked_at guard and the
-- claimed-build guard untouched — they predate 642.

BEGIN;

DO $$
DECLARE
    v_pq text;
    v_new constant text :=
       E'    -- 642: the deploy-path half. A site whose pages changed since its last\n'
        '    -- render is due EARLY once quiet for 30 minutes (every pages-writer\n'
        '    -- bumps updated_at, so this covers deploy, retraction and expiry\n'
        '    -- without naming any visibility column — see 622 on why it must not).\n'
        '    -- The 3-day line is the FLOOR and stays UNCONDITIONAL: the quiet test\n'
        '    -- gates only the early branch, so a site edited for ever still gets\n'
        '    -- its floor refresh. Never-selected is the failure mode to refuse.\n'
        '    AND (\n'
        '      COALESCE(r.last_selected_at, ''-infinity''::timestamptz) < now() - interval ''3 days''\n'
        '      OR (EXISTS (SELECT 1 FROM pages pu\n'
        '                   WHERE pu.site_id = s.id\n'
        '                     AND pu.updated_at > r.last_selected_at)\n'
        '          AND NOT EXISTS (SELECT 1 FROM pages pq\n'
        '                   WHERE pq.site_id = s.id\n'
        '                     AND pq.updated_at > now() - interval ''30 minutes''))\n'
        '    )';
    v_old constant text :=
        '    AND COALESCE(r.last_selected_at, ''-infinity''::timestamptz) < now() - interval ''3 days''';
BEGIN
    SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';

    -- The whole inserted block must occur exactly once (occurrence, not row).
    IF (length(v_pq) - length(replace(v_pq, v_new, ''))) / length(v_new) <> 1 THEN
        RAISE EXCEPTION '642 ROLLBACK pre-flight: the 642 block does not occur exactly once — the row has drifted since 642 applied; re-read it and edit by hand';
    END IF;

    UPDATE scheduled_tasks
       SET pre_query = replace(pre_query, v_new, v_old),
           updated_at = now()
     WHERE name = 'sitemap-refresh-rotation';

    SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';

    IF v_pq LIKE '%interval ''30 minutes''%' OR v_pq LIKE '%pu.updated_at%' THEN
        RAISE EXCEPTION '642 ROLLBACK: the early-due branch is still present after the reverse replace';
    END IF;
    IF (length(v_pq) - length(replace(v_pq, v_old, ''))) / length(v_old) <> 1 THEN
        RAISE EXCEPTION '642 ROLLBACK: the plain 3-day line does not occur exactly once after restore';
    END IF;
    IF (length(v_pq) - length(replace(v_pq, 'pg.deployed_at IS NOT NULL', ''))) / length('pg.deployed_at IS NOT NULL') <> 1 THEN
        RAISE EXCEPTION '642 ROLLBACK: 622''s guard was disturbed — it must survive a 642 rollback';
    END IF;
    IF v_pq NOT LIKE '%locked_at IS NULL%' THEN
        RAISE EXCEPTION '642 ROLLBACK: the locked_at guard is gone';
    END IF;

    RAISE NOTICE '642 ROLLBACK OK — rotation is 3-day-floor-only again; 622 and locked_at guards intact.';
END $$;

COMMIT;
