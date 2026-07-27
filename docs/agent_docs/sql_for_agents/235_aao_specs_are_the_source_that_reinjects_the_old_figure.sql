-- 235_aao_specs_are_the_source_that_reinjects_the_old_figure.sql
--
-- ══ THE FINDING THIS FILE EXISTS FOR ═══════════════════════════════════════
-- `site_specs` prose is not merely an INSTRUCTION to a writer. It is a live
-- SOURCE that a re-render resolves fields from, so it **overwrites page
-- corrections**. Proven today, by accident, and it is the most useful thing
-- this lane has learned in a while.
--
-- Sequence, all on 2026-07-27:
--
--   12:2x  231/232 replaced "30+ agent types" with "170+ agent types" in
--          ai-agent-orchestration.com's page_components.content_data. Their
--          post-conditions asserted ZERO stale figures site-wide, and passed.
--   12:37  the /about page re-rendered (233), reason=section_data_resolved.
--   12:4x  content_data for about/leadership-team contained "30+ agent types"
--          AGAIN, and the deployed HTML with it.
--
-- The re-rendered bio and the `identity` spec are byte-identical:
--
--   PAGE content_data: …gle-model approach. The framework now coordinates 30+ agent types on Kubernetes and Kafka, processing over a thousand or…
--   SPEC identity:     …gle-model approach. The framework now coordinates 30+ agent types on Kubernetes and Kafka, processing over a thousand or…
--
-- `rerender_page_sections` re-renders "from stored content_data **+ fresh
-- resolved fields**". The leadership bio is a resolved field; its source is the
-- spec; the spec still said 30+. So the render faithfully restored the old
-- claim over the correction.
--
-- ── WHY THIS MATTERS BEYOND ONE SENTENCE ───────────────────────────────────
-- The fabricated-stats lane's owed item (c) has been recorded as "a number in a
-- briefing/identity spec is an instruction to the writer and nothing refreshes
-- it". That is too mild, and this is the correction: **a stale figure in a spec
-- is re-injected into published content on every re-render, silently, over the
-- top of any repair.** Fixing pages without fixing specs is not a partial fix,
-- it is a temporary one — and the temporary window can be minutes.
--
-- It also explains a shape the fleet keeps hitting: a correction that "did not
-- stick" and gets re-diagnosed as a caching or deploy problem, when in fact
-- something re-derived it correctly from a stale source.
--
-- ══ WHAT THIS CHANGES ══════════════════════════════════════════════════════
-- The same owner ruling as 231 (use the real agent figure), applied at the
-- SOURCE this time: identity, briefing, strategy, portfolio.
--   "30+ agent types"        -> "170+ agent types"        (live: 174 distinct)
--   "over 70 specialised AI agents" -> "170+ agents"      (live: 175 active)
--   "thousands of concurrent instances" -> removed        (never measured; the
--                            site's own writer_block has always forbidden it)
--
-- `site_plan` has none of these. **"8 departments" is deliberately NOT touched**
-- — ai-agent-orchestration registers it as a fact (its own internal taxonomy)
-- while leopardess bans the identical phrase as an invented fabrication, and
-- the database has no department concept at all. That contradiction needs an
-- owner ruling, not a migration.
--
--
-- TWO VARIANTS THE FIRST ATTEMPT MISSED, and the post-condition CAUGHT them,
-- rolling the whole transaction back rather than shipping a partial fix:
--   * site_plan  "the 70+ agent organisation"  — singular "agent", so the
--     plural pattern did not match it;
--   * portfolio  "across potentially thousands of concurrent agent instances"
--     — the removal pattern expected a verb (with|managing|handling|processing)
--     and this one reads "across potentially".
-- This is the difference a post-condition written from a DIFFERENT angle than
-- the change makes: 231's assertion was built from its own list of shapes and
-- passed over two defects; this one asserts the absence of the BANNED PATTERNS
-- and so had no way to agree with a change that had missed a variant.
--
-- ── ANCHORED, SINGLE-PASS REGEXES (the lesson from 232) ────────────────────
-- 231 used a chain of replace() calls in which one pattern matched another's
-- REPLACEMENT ("170+ agents" contains "70+ agent") and produced "1170+".
-- Every pattern here is anchored with Postgres's \m word-start, so "70" inside
-- "170" can never match and no step can feed the next.

BEGIN;

DO $fix$
DECLARE
    v_site uuid;
    n int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'ai-agent-orchestration.com';
    IF v_site IS NULL THEN RAISE EXCEPTION '235: no site row'; END IF;

    WITH upd AS (
        UPDATE site_specs ss
        SET data = regexp_replace(
                   regexp_replace(
                   regexp_replace(
                   regexp_replace(
                     regexp_replace(ss.data::text,
                       '\m(30|thirty)\s*\+?\s*(distinct )?agent types', '170+ agent types', 'gi'),
                       '\m[Oo]ver 70 specialised AI agents', '170+ agents', 'g'),
                       '\m70\s*\+?\s*agents', '170+ agents', 'g'),
                       '\m70\s*\+?\s*agent\M', '170+ agent', 'g'),
                       ',?\s*(with|managing|handling|processing|across( potentially)?)\s+thousands of concurrent( agent)? instances( active at peak)?', '', 'g'
                   )::jsonb,
            updated_at = now()
        WHERE ss.site_id = v_site AND ss.is_current
          AND ss.aspect IN ('identity','briefing','strategy','portfolio','site_plan')
          AND (ss.data::text ~* '\m(30|thirty)\s*\+?\s*(distinct )?agent types'
            OR ss.data::text ~  '\m[Oo]ver 70 specialised AI agents'
            OR ss.data::text ~  '\m70\s*\+?\s*agents'
            -- singular too: site_plan says "the 70+ agent organisation". The
            -- first run added this to the SET and not to the WHERE, so the row
            -- was never selected and the post-condition failed the whole
            -- transaction. A row filter narrower than the transform is a
            -- silent no-op on exactly the rows you most meant to catch.
            OR ss.data::text ~  '\m70\s*\+?\s*agent\M'
            OR ss.data::text ~* 'thousands of concurrent( agent)? instances')
        RETURNING 1
    )
    SELECT count(*) INTO n FROM upd;

    IF n = 0 THEN
        RAISE EXCEPTION '235: no spec rows matched — already fixed, or the wording moved. Re-survey.';
    END IF;
    RAISE NOTICE '235: rewrote % spec row(s)', n;
END $fix$;

-- ── Post-conditions ────────────────────────────────────────────────────────
-- Deliberately NOT written from the same list of shapes as the change: these
-- assert the ABSENCE of the banned patterns AND the survival of the cascade
-- artefact check, which is the pair 231 got wrong.
DO $post$
DECLARE
    v_30 int; v_70 int; v_conc int; v_cascade int; v_aspects int;
BEGIN
    SELECT
      count(*) FILTER (WHERE ss.data::text ~* '\m(30|thirty)\s*\+?\s*(distinct )?agent types'),
      count(*) FILTER (WHERE ss.data::text ~* '\m(70|seventy)\s*\+?\s*(specialised |specialized )?(ai )?agents?\M'),
      count(*) FILTER (WHERE ss.data::text ~* 'thousands of concurrent( agent)? instances'),
      count(*) FILTER (WHERE ss.data::text ~ '1170'),
      count(*)
      INTO v_30, v_70, v_conc, v_cascade, v_aspects
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
    WHERE s.domain='ai-agent-orchestration.com' AND ss.is_current
      AND ss.aspect IN ('identity','briefing','strategy','portfolio','site_plan');

    IF v_30      <> 0 THEN RAISE EXCEPTION '235: % spec(s) still carry a 30+ agent-types claim', v_30; END IF;
    IF v_70      <> 0 THEN RAISE EXCEPTION '235: % spec(s) still carry a 70-agents claim', v_70; END IF;
    IF v_conc    <> 0 THEN RAISE EXCEPTION '235: % spec(s) still carry a concurrent-instances claim', v_conc; END IF;
    IF v_cascade <> 0 THEN RAISE EXCEPTION '235: % spec(s) contain the 1170 cascade artefact — the anchoring failed', v_cascade; END IF;
    IF v_aspects  = 0 THEN RAISE EXCEPTION '235: no spec rows found at all — the post-condition is vacuous'; END IF;

    RAISE NOTICE '235 OK: % spec rows checked, none carries a superseded agent figure', v_aspects;
END $post$;

COMMIT;

-- ── AFTER THIS: the page must be re-rendered AGAIN ─────────────────────────
-- about/leadership-team's content_data still holds the re-injected "30+".
-- Correct content_data, then re-render, in that order — otherwise the spec (now
-- fixed) and the page (still stale) simply disagree until something re-resolves.
