-- 231_aao_real_agent_figure_and_banned_claims.sql
-- OWNER RULING 2026-07-27: use the real agent figure on ai-agent-orchestration.com
-- instead of "70+ agents" / "30+ agent types".
--
-- Three things, all one theme: make the published copy and the site's own rails agree.
--
-- ══ 1. THE FIGURE ══════════════════════════════════════════════════════════
-- Live, re-derived 2026-07-27 by running the register's OWN declared queries:
--
--   aao-agent-definitions  stored 175  live 175  (count(*) ... WHERE is_active)
--   aao-agent-types        stored 174  live 174  (count(DISTINCT type) ...)
--
-- So "70+ agents" understated the fleet by more than half, and "30+ agent
-- types" by more than five times. Both were TRUE (they are `gte` lower bounds,
-- which is why no checker ever objected) and both were badly misleading.
--
-- ── WHY THE COPY SAYS "170+" AND NOT "175" ─────────────────────────────────
-- Because the count moves, and it moved DURING this session: measured 176
-- active at 11:07 and 175 at 12:2x — one definition removed inside the hour.
--
-- That matters more than the arithmetic. `aao-agent-definitions` is a `gte`
-- fact, so a published figure is supported only while it is <= the registered
-- value. Freeze "175" into a hero headline and the first deactivation makes the
-- page's own build gate reject it as unsupported — the page becomes unbuildable
-- for a sentence that was true when written, which is bugs_closed/073's shape.
-- "170+" is the same real magnitude with ~5 units of headroom against exactly
-- the drift that was observed. The exact figure keeps living in the register,
-- where a query refreshes it (`evidence-freshness`, last run 2026-07-26 18:34)
-- rather than a human remembering to.
--
-- ══ 2. A CLAIM THE SITE'S OWN WRITER_BLOCK ALREADY FORBADE ═════════════════
-- Found while doing the above. aao's writer_block NEVER-STATE list says:
--
--   "concurrent-instance counts ('thousands of concurrent instances' is not
--    measured)"
--
-- and the site publishes it in FOUR sentences, on /about and
-- /enterprise-reference-deployment. It is removed here.
--
-- ══ 3. WHY IT SURVIVED — THE STRUCTURAL FINDING ════════════════════════════
-- aao has **7 facts and ZERO banned_claims**:
--
--   SELECT jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb))
--   FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--   WHERE aspect='evidence_base' AND is_current AND s.domain='ai-agent-orchestration.com';
--   -- 0
--
-- The writer_block's NEVER-STATE section and banned_claims[] are TWO
-- hand-maintained lists of the same prohibitions, and only the first was
-- written. So the PREVENTIVE rail forbids six categories and the DETECTIVE
-- rail catches none of them: nothing could ever flag the claim, on any path,
-- however many times the page was rebuilt or audited. Same drift class as the
-- council-roster duplication CLAUDE.md warns about — two lists that must stay
-- identical and no mechanism keeping them so.
--
-- This file seeds banned_claims FROM that NEVER-STATE list, closing the gap for
-- this site. The general defect (nothing derives one list from the other) is
-- recorded in the lane's NOTES rather than fixed here — it needs Go, not SQL.
--
-- ── EXPECTED, NOT A BUG: these bans fire immediately on the LIVE pages ──────
-- This edits content_data only; rendered_html still carries "70+ agents" and
-- "thousands of concurrent instances" until the pages are re-rendered. The
-- post-deploy audit will therefore raise banned_claim findings against the
-- deployed HTML. That is correct: it is the platform telling us these two pages
-- are stale, and it self-clears on re-render.

BEGIN;

DO $fix$
DECLARE
    v_site uuid;
    n_content int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'ai-agent-orchestration.com';
    IF v_site IS NULL THEN
        RAISE EXCEPTION '231: no site row for ai-agent-orchestration.com';
    END IF;

    -- ── Content: the figure, and the forbidden claim ───────────────────────
    -- Order matters: 'agents'/'Agents' first, so the bare-'agent' pass only
    -- catches what is left ("The 70+ agent deployment is live").
    WITH upd AS (
        UPDATE page_components pc
        SET content_data = regexp_replace(
                replace(replace(replace(replace(pc.content_data::text,
                    '70+ Agents',      '170+ Agents'),
                    '70+ agents',      '170+ agents'),
                    '70+ agent',       '170+ agent'),
                    '30+ agent types', '170+ agent types'),
                ',?\s*(with|managing|handling)\s+thousands of concurrent( agent)? instances( active at peak)?',
                '', 'g')::jsonb,
            updated_at = now()
        FROM pages p
        WHERE p.id = pc.page_id AND p.site_id = v_site
          AND (pc.content_data::text LIKE '%70+ agent%'
            OR pc.content_data::text LIKE '%70+ Agent%'
            OR pc.content_data::text LIKE '%30+ agent types%'
            OR pc.content_data::text LIKE '%thousands of concurrent%')
        RETURNING 1
    )
    SELECT count(*) INTO n_content FROM upd;

    IF n_content < 7 THEN
        RAISE EXCEPTION '231: expected at least 7 components to change, changed % — re-survey', n_content;
    END IF;
    RAISE NOTICE '231: rewrote % components', n_content;

    -- ── Rails: stop PERMITTING the understatement ──────────────────────────
    UPDATE site_specs
    SET data = jsonb_set(data, '{writer_block}', to_jsonb(
            replace(replace(data->>'writer_block',
              '; "over 70 specialised AI agents" is also true and may stand where already written',
              '. NEVER write "over 70 specialised AI agents" or "70+ agents" — true as a lower bound but understating the fleet by more than half; superseded by owner ruling 2026-07-27. Write "170+ agents".'),
              '; "30+ agent types" is also true and may stand',
              '. NEVER write "30+ agent types" — understates by more than five times; superseded by owner ruling 2026-07-27. Write "170+ agent types".')
        )), updated_at = now()
    WHERE site_id = v_site AND aspect = 'evidence_base' AND is_current
      AND data ? 'writer_block';

    -- ── Rails: give the site the detective half it never had ───────────────
    -- Seeded FROM the writer_block's own NEVER-STATE list, so the two agree.
    -- WORD BOUNDARIES ARE \b, NOT \m..\M. These patterns are compiled by GO
    -- (ParseEvidenceBase: regexp.Compile("(?i)"+p)), not by Postgres, and RE2
    -- rejects \m with 'invalid escape sequence'. That would be silent: on a
    -- compile failure ParseEvidenceBase FALLS BACK TO A LITERAL SUBSTRING, so
    -- the ban would simply never fire and nothing would say so. Caught here by
    -- compiling all eight in Go before seeding them. The boundaries still do
    -- the load-bearing job: without them, /70|30/ matches inside "170+" and
    -- would ban the corrected copy. Both directions are tested.
    UPDATE site_specs
    SET data = jsonb_set(data, '{banned_claims}', $bc$[
          {"pattern": "thousands of concurrent",
           "reason": "2026-07-27: concurrent-instance counts are NOT MEASURED — the writer_block has forbidden this since it was written, and the site published it in four sentences anyway because banned_claims was empty. Any such figure at any value is an invention."},
          {"pattern": "\\b(70|seventy)\\s*\\+?\\s*(specialised |specialized )?(ai )?agents?\\b",
           "reason": "2026-07-27 owner ruling: superseded by the real figure (170+). True as a gte lower bound, which is why no checker objected, and misleading by more than half."},
          {"pattern": "\\b(30|thirty)\\s*\\+?\\s*(distinct )?agent types\\b",
           "reason": "2026-07-27 owner ruling: superseded by the real figure (170+ agent types); understated by more than five times."},
          {"pattern": "departments served",
           "reason": "writer_block NEVER-STATE: the 8 departments are the platform's OWN internal taxonomy, never departments of external clients. This framing is the fabrication audited out of leopardess."},
          {"pattern": "clients served",
           "reason": "writer_block NEVER-STATE: no client count is measured."},
          {"pattern": "satisfaction (rate|score|rating)",
           "reason": "writer_block NEVER-STATE: no satisfaction metric is collected."},
          {"pattern": "awards? won",
           "reason": "writer_block NEVER-STATE: no awards are recorded."},
          {"pattern": "[0-9.]+\\s*%\\s*uptime|uptime[^.]{0,24}[0-9.]+\\s*%",
           "reason": "writer_block NEVER-STATE: uptime is not measured or published."}
        ]$bc$::jsonb), updated_at = now()
    WHERE site_id = v_site AND aspect = 'evidence_base' AND is_current
      AND data ? 'facts';
END $fix$;

-- ── Post-conditions ────────────────────────────────────────────────────────
DO $post$
DECLARE
    v_bad int; v_bans int; v_facts int; v_perm int;
BEGIN
    SELECT count(*) INTO v_bad
    FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
    WHERE s.domain='ai-agent-orchestration.com'
      AND (pc.content_data::text ~ '(^|[^1])70\+\s*[Aa]gent'
        OR pc.content_data::text LIKE '%30+ agent types%'
        OR pc.content_data::text LIKE '%thousands of concurrent%');
    IF v_bad <> 0 THEN
        RAISE EXCEPTION '231: % component(s) still carry a superseded figure or the forbidden concurrency claim', v_bad;
    END IF;

    SELECT jsonb_array_length(ss.data->'banned_claims'), jsonb_array_length(ss.data->'facts'),
           (CASE WHEN ss.data->>'writer_block' LIKE '%is also true and may stand%' THEN 1 ELSE 0 END)
      INTO v_bans, v_facts, v_perm
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
    WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='ai-agent-orchestration.com';

    IF v_bans <> 8 THEN RAISE EXCEPTION '231: expected 8 banned_claims, found %', v_bans; END IF;
    IF v_facts <> 7 THEN RAISE EXCEPTION '231: facts went from 7 to % — the merge clobbered them', v_facts; END IF;
    IF v_perm <> 0 THEN RAISE EXCEPTION '231: the writer_block still PERMITS a superseded figure'; END IF;

    RAISE NOTICE '231 OK: copy on the real figure, concurrency claim gone, % banned_claims seeded, % facts intact', v_bans, v_facts;
END $post$;

COMMIT;
