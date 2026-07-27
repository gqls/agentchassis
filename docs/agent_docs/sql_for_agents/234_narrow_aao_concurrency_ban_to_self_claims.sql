-- 234_narrow_aao_concurrency_ban_to_self_claims.sql
--
-- CORRECTS a banned_claims pattern I seeded in 231, ~40 minutes earlier.
-- Its own false positive found it, on correct prose, which is the useful part.
--
-- ══ WHAT WENT WRONG ════════════════════════════════════════════════════════
-- 231 banned the bare phrase `thousands of concurrent`, taken from aao's
-- writer_block NEVER-STATE entry:
--
--   "concurrent-instance counts ('thousands of concurrent instances' is not measured)"
--
-- I dropped the word that was doing the work. The prohibition is about
-- **instances** — a claim about OUR OWN scale, which nothing measures. The
-- pattern I wrote catches any sentence containing the phrase, including a
-- generic statement about what production systems in general require:
--
--   "They need to process thousands of concurrent workflows without becoming
--    a bottleneck."
--   -- aao/why-most-ai-agent-frameworks-fail-at-the-orchestration-layer,
--      article-body. This is an ARTICLE describing an industry requirement.
--      It claims nothing about us. It is correct prose and must stay.
--
-- The three sentences 231/232 legitimately removed were all self-claims and all
-- said "instances": "handling thousands of concurrent instances", "managing
-- thousands of concurrent agent instances in production", "with thousands of
-- concurrent instances active at peak".
--
-- ══ WHY THIS IS THE FIX, RATHER THAN EDITING THE ARTICLE ═══════════════════
-- Two independent reasons, and the second is the stronger one.
--
-- 1. The sentence is true and is not a claim about us. Editing correct writing
--    to satisfy an over-broad rule is how a checker starts costing more than it
--    returns.
-- 2. **That component has NULL `content_data`** — it is rendered HTML only. So
--    "just fix the copy" would mean editing `rendered_html` directly, with no
--    stored source to keep in step, and the edit would be silently discarded
--    the first time anything rebuilt the page. Narrowing the rule is the only
--    repair that stays repaired.
--
-- ══ THE MEASUREMENT THIS TURNED UP, WHICH IS BIGGER THAN THE BUG ═══════════
--   SELECT count(*) AS components, count(DISTINCT p.site_id) AS sites,
--          count(DISTINCT p.id) AS pages
--   FROM page_components pc JOIN pages p ON p.id = pc.page_id
--   WHERE COALESCE(pc.rendered_html,'') <> ''
--     AND (pc.content_data IS NULL OR pc.content_data::text = '{}');
--   --  201 |  8 |  79
--
-- **201 components, across 8 sites and 79 pages, have published HTML and no
-- content_data at all.** They cannot be corrected by a content_data migration,
-- cannot be re-rendered (there is nothing to render from), and are invisible to
-- the stat audit added in bugs_open/093, which reads content_data by
-- construction. They are reachable ONLY by the HTML-side scans that predate all
-- of this work. Recorded on `093`: it is the concrete argument for why the
-- rendered_html scan must stay, and a limit on any "fix it in content_data"
-- strategy fleet-wide.

BEGIN;

DO $fix$
DECLARE
    v_idx int;
    n int;
BEGIN
    -- Locate the pattern by VALUE rather than by array position: 231 wrote it at
    -- index 0, but another session may have reordered the array since.
    SELECT ord - 1 INTO v_idx
    FROM site_specs ss
    JOIN sites s ON s.id = ss.site_id,
    LATERAL jsonb_array_elements(ss.data->'banned_claims') WITH ORDINALITY AS e(b, ord)
    WHERE ss.aspect='evidence_base' AND ss.is_current
      AND s.domain='ai-agent-orchestration.com'
      AND e.b->>'pattern' = 'thousands of concurrent';

    IF v_idx IS NULL THEN
        RAISE EXCEPTION '234: the over-broad pattern is not present — already fixed, or the array changed. Re-survey rather than forcing.';
    END IF;

    UPDATE site_specs ss
    SET data = jsonb_set(
          jsonb_set(ss.data, ARRAY['banned_claims', v_idx::text, 'pattern'],
                    '"thousands of concurrent( agent)? instances"'::jsonb),
          ARRAY['banned_claims', v_idx::text, 'reason'],
          to_jsonb('2026-07-27: concurrent-INSTANCE counts are NOT MEASURED — a claim about our own scale. '
                || 'Narrowed from the bare phrase "thousands of concurrent" on the day it was seeded: that '
                || 'version also matched an article sentence about what production systems in general '
                || 'require ("process thousands of concurrent workflows"), which claims nothing about us '
                || 'and is correct prose. The word "instances" is the prohibition.'::text)),
        updated_at = now()
    FROM sites s
    WHERE s.id = ss.site_id AND ss.aspect='evidence_base' AND ss.is_current
      AND s.domain='ai-agent-orchestration.com';

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN
        RAISE EXCEPTION '234: expected to update 1 evidence_base row, updated %', n;
    END IF;
END $fix$;

-- ── Post-conditions ────────────────────────────────────────────────────────
DO $post$
DECLARE
    v_bans int; v_facts int; v_broad int; v_narrow int;
BEGIN
    SELECT jsonb_array_length(ss.data->'banned_claims'), jsonb_array_length(ss.data->'facts'),
           count(*) FILTER (WHERE b->>'pattern' = 'thousands of concurrent'),
           count(*) FILTER (WHERE b->>'pattern' = 'thousands of concurrent( agent)? instances')
      INTO v_bans, v_facts, v_broad, v_narrow
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
    LATERAL jsonb_array_elements(ss.data->'banned_claims') b
    WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='ai-agent-orchestration.com'
    GROUP BY 1,2;

    IF v_broad  <> 0 THEN RAISE EXCEPTION '234: the over-broad pattern is still present'; END IF;
    IF v_narrow <> 1 THEN RAISE EXCEPTION '234: the narrowed pattern was not written (found %)', v_narrow; END IF;
    IF v_bans   <> 8 THEN RAISE EXCEPTION '234: banned_claims count changed to % — expected 8', v_bans; END IF;
    IF v_facts  <> 7 THEN RAISE EXCEPTION '234: facts changed to % — expected 7', v_facts; END IF;

    RAISE NOTICE '234 OK: ban narrowed to self-claims; % banned_claims, % facts intact', v_bans, v_facts;
END $post$;

COMMIT;
