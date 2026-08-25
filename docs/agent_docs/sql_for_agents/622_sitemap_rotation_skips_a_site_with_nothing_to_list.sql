-- 622 — the sitemap rotation must not SPEND a site's 3-day slot on a site that
--       has nothing to list yet.
--
-- THE DEFECT, MEASURED 2026-08-25, and it is live on two sites right now.
--
-- `590`'s pre_query stamps `site_discovery_rotation` BEFORE firing the Kafka
-- message (fire-and-forget, the standing rotation idiom). If the selected site has
-- no listable pages at that instant, `render_sitemap` correctly returns
-- rendered=false / url_count=0 and `check_has_urls` correctly routes to `complete`
-- without committing — an empty sitemap tells a crawler the site has no pages,
-- which is worse than none.
--
-- But the STAMP has already been written. So the site does not come round again
-- for THREE DAYS, and nothing anywhere says so:
--
--     homegarden.uk   swept 2026-08-25 10:50:18, first page deployed 12:47:09
--                     -> all 20 pages deployed AFTER the sweep. Serves 404.
--     cv1.co.uk       swept 2026-08-25 12:21:48, first page deployed 13:47:04
--                     -> all 3 pages deployed AFTER the sweep. Serves 404.
--
-- Both are fully built, live, and would have had a sitemap within 30 minutes had
-- they simply not been selected early. Their stamps were cleared by hand on
-- 2026-08-25; this stops it recurring.
--
-- THIS IS THE COUNCIL'S `bug_historian` OBJECTION ARRIVING FOR REAL (corr
-- 8a004aab, medium, ACCEPTED-not-fixed): "url_count=0 collapses 'opted out' and
-- 'nothing found' into one silent no-op". It was right, and the live cost is not
-- the missing log line — it is the CONSUMED SLOT.
--
-- WHY THE GUARD IS DELIBERATELY WEAKER THAN THE ACTION'S OWN FILTER.
-- `render_sitemap` lists pages that are status='active' AND noindex IS NOT TRUE
-- AND deployed_at IS NOT NULL AND not expired. This guard tests only
-- `status='active' AND deployed_at IS NOT NULL` — a STRICT SUPERSET of what the
-- action lists.
--
-- That asymmetry is the whole safety argument and must not be "tidied" into a
-- copy of the action's predicate. A guard that exactly mirrored the action would
-- have to be kept in lockstep with it forever, and the failure mode of drift is a
-- site that is silently NEVER SELECTED — strictly worse than the bug being fixed.
-- Because this guard is weaker, drift can only ever make it select a site the
-- action then finds nothing for, which is exactly today's behaviour and no worse.
-- **A guard that cannot be wrong in the dangerous direction beats a guard that is
-- currently exact.**
--
-- It also fixes the observed cases completely: both sites had ZERO rows with
-- deployed_at IS NOT NULL at the moment they were selected.
--
-- NOT FIXED HERE, deliberately: a site whose pages all exist but are ALL noindex
-- or ALL expired still burns its slot silently. That residue needs the
-- machine-readable `skip_reason` (`opted_out` | `no_listable_urls`) the council
-- round designed, which is a Go change. Zero sites are in that state today
-- (checked 2026-08-25: 0 rows fleet-wide with noindex true among deployed active
-- pages), so this migration closes the live case and leaves the latent one named.

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM scheduled_tasks
     WHERE name='sitemap-refresh-rotation'
       AND pre_query LIKE '%AND s.locked_at IS NULL%'
       AND pre_query NOT LIKE '%pg.deployed_at IS NOT NULL%';
    IF n <> 1 THEN
        RAISE EXCEPTION '622 pre-flight: expected exactly 1 un-guarded sitemap-refresh-rotation row, found % — the row has drifted, re-read it before applying', n;
    END IF;
END $$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         '    AND s.locked_at IS NULL',
         E'    AND s.locked_at IS NULL\n'
         '    -- 622: do not spend the site''s 3-day slot on a site with nothing to\n'
         '    -- list. DELIBERATELY WEAKER than render_sitemap''s own filter (which\n'
         '    -- also excludes noindex and expired) so drift can never cause a site\n'
         '    -- to be silently never-selected. Do not tighten it to match.\n'
         '    AND EXISTS (\n'
         '      SELECT 1 FROM pages pg\n'
         '      WHERE pg.site_id = s.id AND pg.status = ''active''\n'
         '        AND pg.deployed_at IS NOT NULL)'),
       updated_at = now()
 WHERE name = 'sitemap-refresh-rotation';

DO $$
DECLARE
    v_pq text; n_sel int;
BEGIN
    SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';

    IF v_pq NOT LIKE '%pg.deployed_at IS NOT NULL%' THEN
        RAISE EXCEPTION '622: the guard was not inserted — the anchor did not match';
    END IF;
    IF v_pq NOT LIKE '%locked_at IS NULL%' THEN
        RAISE EXCEPTION '622: the locked_at guard is gone — an owner HALT would be deployed against';
    END IF;
    -- The guard must not have been tightened into a mirror of the action's filter.
    IF v_pq LIKE '%pg.noindex%' OR v_pq LIKE '%pg.expires_at%' THEN
        RAISE EXCEPTION '622: the guard mirrors render_sitemap''s filter — see the header: drift then causes silent never-selection';
    END IF;

    -- Prove the guarded query still selects real work: with every site currently
    -- stamped inside the 3-day window this may legitimately be 0, so assert only
    -- that it PARSES and runs.
    EXECUTE 'SELECT count(*) FROM sites s WHERE s.status IN (''active'',''deployed'') AND s.locked_at IS NULL '
            'AND EXISTS (SELECT 1 FROM pages pg WHERE pg.site_id = s.id AND pg.status = ''active'' AND pg.deployed_at IS NOT NULL)'
      INTO n_sel;
    RAISE NOTICE '622 OK — guard live; % unlocked live sites currently have at least one deployed active page.', n_sel;
END $$;

COMMIT;
