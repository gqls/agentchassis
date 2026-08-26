-- 642 — the sitemap follows the deploy: a site whose pages CHANGED since its
--       last render becomes due EARLY, instead of waiting out the 3-day floor.
--
-- THE DEFERRED HALF, NOW DUE. `590` wired render_sitemap into a rotation and
-- deliberately deferred "the deploy path" with this costing: "the deploy path
-- would re-probe a whole site on every page change (135 requests for
-- webdesign.co.uk, every time)". That costing was right about the wrong design.
-- Firing a render PER EDIT is expensive; making an edited site DUE at the
-- selector costs nothing new, because the rotation already has a hard cost cap:
-- ONE site per 1800s tick, so at most 48 renders/day fleet-wide no matter what
-- this clause does. The deploy-path event lands in the selector, where the cap
-- already lives. SEO-002's open question ("new page and retracted page are
-- different events") is answered for both: each is an UPDATE to a pages row.
--
-- WHY `updated_at` ALONE IS THE RIGHT SIGNAL. Surveyed 2026-08-26: every
-- `UPDATE pages` statement in the Go tree bumps `updated_at = NOW()` alongside
-- whatever it changes (deploy stamp in deploy_evidence.go:294, retraction,
-- needs_rebuild, section stores — 12 call sites), and **0 rows fleet-wide have
-- updated_at < deployed_at** [MEASURED 2026-08-26]. So `updated_at` covers a
-- new page's deploy, a redeploy, a retraction, a noindex flip and an
-- expiry-set — the full visibility surface — without this pre_query naming any
-- visibility column (see 622: naming them here is the forbidden mirror).
--
-- It OVER-triggers, deliberately: a pre-deploy content edit also bumps
-- updated_at and buys a render that changes nothing served. Accepted — the
-- mid-build case is already excluded by the claimed-build guard below, the
-- cost is capped by the tick, and the safe failure direction for this selector
-- is render-too-often, never never-render (the 622 argument, again).
--
-- THE QUIET PERIOD, and why it gates ONLY the early branch. Churn arrives in
-- waves: sites with >=1 page updated per day were 2/2/1/0/3 over 2026-08-19..23
-- and then 27 on 08-25, 21 by 09:30 on 08-26 (a fleet rerender wave)
-- [MEASURED 2026-08-26]. Rendering mid-wave lists a half-updated site and buys
-- a second render when the wave ends; requiring 30 quiet minutes first means
-- one render, after it settles. But the quiet requirement MUST NOT gate the
-- 3-day floor: a site edited for ever would then be silently never selected —
-- the exact drift mode 622 was written to refuse. So the floor stays
-- unconditional and the quiet test sits inside the OR's second arm only.
--
-- COST, measured not guessed [all 2026-08-26]: a render is one GET per listable
-- page — fleet average ~26, maximum 147 (webdesign.co.uk). Steady state adds
-- ~2-3 early renders/day on normal churn; a full-fleet wave (31 sites) drains
-- in ~15.5h at the tick cap. An identical re-render pushes an empty-diff
-- commit (the git adapter's no-op skip covers deletions only,
-- github_client.go:250) — accepted: it is rare (needs two same-day renders
-- with no lastmod movement), succeeds, and fails no run.
--
-- THE REJECTED ALTERNATIVE, named so it is not re-derived: "probe only the
-- changed URL and merge into the served sitemap" (the handoff's other option).
-- Rejected, not deferred — it needs a Go change plus reading the live artefact
-- as input, and what it saves is ~26 GETs per render. Wrong trade at this
-- scale; revisit only if the fleet grows to where a full render is expensive.
--
-- LATENCY, before -> after: a published or retracted page waited up to 3 DAYS
-- to be reflected; now 30-60 minutes typical (30 quiet + <=30 tick) plus any
-- backlog queueing, floor unchanged as the safety net. This also answers §3's
-- "is 3 days right?" — 3 days is now a re-probe floor for sites with no DB
-- change (catching serving-side drift), not the delivery latency of an edit.
--
-- THE BEHAVIOURAL PROOF IS BUILT IN AT APPLY TIME. All 31 rotation stamps are
-- currently 2026-08-24/25, so under the OLD rule ZERO sites are due before
-- 2026-08-27 16:02:34 (oldest stamp + 3 days) [MEASURED 2026-08-26]. Any
-- selection before that instant can only be this migration's branch — and 28
-- sites are due-by-change-and-quiet right now, so the next tick should select
-- one. "The rotation still runs" cannot be mistaken for the proof; a selection
-- BEFORE the floor date is.
--
-- ANCHORING, per the 622 lesson (its council round, editquality): pre-flight
-- asserts the anchor's OCCURRENCE COUNT is exactly 1 (replace() is global, and
-- a row count says nothing about occurrences within a row), and the verify
-- block asserts POST-state occurrence counts, which a partial match cannot
-- satisfy.

BEGIN;

DO $$
DECLARE
    v_pq text; n int;
BEGIN
    SELECT count(*) INTO n FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';
    IF n <> 1 THEN
        RAISE EXCEPTION '642 pre-flight: expected exactly 1 sitemap-refresh-rotation row, found %', n;
    END IF;

    SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';

    -- The anchor must occur EXACTLY once (occurrence count, not row count).
    IF (length(v_pq) - length(replace(v_pq,
          '    AND COALESCE(r.last_selected_at, ''-infinity''::timestamptz) < now() - interval ''3 days''', '')))
       / length('    AND COALESCE(r.last_selected_at, ''-infinity''::timestamptz) < now() - interval ''3 days''') <> 1 THEN
        RAISE EXCEPTION '642 pre-flight: the 3-day anchor line does not occur exactly once — the row has drifted, re-read it before applying';
    END IF;

    -- Refuse a double apply.
    IF v_pq LIKE '%interval ''30 minutes''%' THEN
        RAISE EXCEPTION '642 pre-flight: pre_query already carries the quiet-period clause — already applied';
    END IF;
END $$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         '    AND COALESCE(r.last_selected_at, ''-infinity''::timestamptz) < now() - interval ''3 days''',
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
         '    )'),
       updated_at = now()
 WHERE name = 'sitemap-refresh-rotation';

DO $$
DECLARE
    v_pq text; n_due_change int; n_due_age int;
BEGIN
    SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';

    -- POST-state occurrence counts (a partial match cannot satisfy these).
    IF (length(v_pq) - length(replace(v_pq, 'interval ''3 days''', ''))) / length('interval ''3 days''') <> 1 THEN
        RAISE EXCEPTION '642: expected exactly 1 occurrence of the 3-day floor after apply';
    END IF;
    IF (length(v_pq) - length(replace(v_pq, 'pu.updated_at > r.last_selected_at', ''))) / length('pu.updated_at > r.last_selected_at') <> 1 THEN
        RAISE EXCEPTION '642: expected exactly 1 occurrence of the changed-since-render clause';
    END IF;
    IF (length(v_pq) - length(replace(v_pq, 'interval ''30 minutes''', ''))) / length('interval ''30 minutes''') <> 1 THEN
        RAISE EXCEPTION '642: expected exactly 1 occurrence of the quiet-period clause';
    END IF;

    -- 622's invariants must survive this edit intact.
    IF (length(v_pq) - length(replace(v_pq, 'pg.deployed_at IS NOT NULL', ''))) / length('pg.deployed_at IS NOT NULL') <> 1 THEN
        RAISE EXCEPTION '642: 622''s has-deployed-pages guard is no longer exactly once';
    END IF;
    IF v_pq NOT LIKE '%locked_at IS NULL%' THEN
        RAISE EXCEPTION '642: the locked_at guard is gone — an owner HALT would be deployed against';
    END IF;
    IF v_pq LIKE '%pg.noindex%' OR v_pq LIKE '%pg.expires_at%' THEN
        RAISE EXCEPTION '642: a visibility column has appeared in the pre_query — see 622: that mirror drifts into silent never-selection';
    END IF;

    -- Live counts, for the apply log. On 2026-08-26 the expectation is
    -- age-due = 0 (all stamps are 08-24/25) and change-due ~28: any selection
    -- before 2026-08-27 16:02:34 is therefore this migration's branch firing.
    EXECUTE '
      SELECT count(*) FILTER (WHERE age_due), count(*) FILTER (WHERE NOT age_due AND change_due AND quiet)
      FROM (
        SELECT COALESCE(r.last_selected_at, ''-infinity''::timestamptz) < now() - interval ''3 days'' AS age_due,
               EXISTS (SELECT 1 FROM pages pu WHERE pu.site_id = s.id AND pu.updated_at > r.last_selected_at) AS change_due,
               NOT EXISTS (SELECT 1 FROM pages pq WHERE pq.site_id = s.id AND pq.updated_at > now() - interval ''30 minutes'') AS quiet
        FROM sites s
        LEFT JOIN site_discovery_rotation r ON r.site_id = s.id AND r.agent_type = ''sitemap-refresh''
        WHERE s.status IN (''active'',''deployed'') AND s.locked_at IS NULL
          AND EXISTS (SELECT 1 FROM pages pg WHERE pg.site_id = s.id AND pg.status = ''active'' AND pg.deployed_at IS NOT NULL)
      ) x'
      INTO n_due_age, n_due_change;

    RAISE NOTICE '642 OK — early-due branch live. Due now: % by age, % by change-and-quiet.', n_due_age, n_due_change;
END $$;

COMMIT;
