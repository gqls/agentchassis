-- SQL_p4_fix_tool_count.sql — webdesign.co.uk, phase 4
--
-- Correcting a wrong number that I put there myself.
--
-- The mission brief said "roughly sixty-four small single-purpose web tools".
-- That figure was written before the catalogue was actually built. The real
-- count is 63: 55 tools carried over from website-design.com plus 8 from
-- websitedesign.com (9 minus the client-side LLM builder the owner asked to
-- skip). Verified against port/catalogue.json, which is generated from the
-- source tree rather than typed:
--     jq '.tools | length' catalogue.json  ->  63
--
-- Eight specs picked the figure up from the brief and repeated it in prose the
-- site will actually show: identity.about_us, strategy.value_proposition, and
-- the briefing the page planner reads. Left alone, the home page would open by
-- advertising a tool that does not exist — an invented statistic, and one this
-- project has now produced twice from the same root cause (a count typed by
-- hand instead of derived). The about page no longer has this problem because
-- its figure is substituted from the catalogue at transform time.
--
-- Run BEFORE releasing needs_site_plan, so the planner never sees the wrong
-- number. Every current spec is superseded and re-inserted with the phrase
-- replaced, preserving all other content.

\set ON_ERROR_STOP on

BEGIN;

-- The INSERT must select FROM the UPDATE's own RETURNING, not from a separate
-- SELECT with the UPDATE sitting in an unreferenced sibling CTE. All CTEs in a
-- statement see the same snapshot and an unreferenced one has no ordering
-- guarantee, so the first version of this file inserted the new rows while the
-- old ones were still is_current and died on idx_site_specs_current
-- (site_id, aspect) WHERE is_current. Making the INSERT consume the UPDATE's
-- output is what forces the supersede to happen first. This is exactly the
-- shape robot_hands/SQL_2026-07-17_r1b uses, and now it is clear why.
WITH old AS (
    UPDATE site_specs ss
       SET is_current = false, superseded_at = now()
     WHERE ss.site_id = (SELECT id FROM sites WHERE domain = 'webdesign.co.uk')
       AND ss.is_current
       AND ss.data::text ILIKE '%sixty-four%'
    RETURNING ss.site_id, ss.aspect, ss.data, ss.source, ss.source_agent
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT
    t.site_id,
    t.aspect,
    replace(
      replace(t.data::text, 'sixty-four', 'sixty-three'),
      'Sixty-four', 'Sixty-three'
    )::jsonb,
    t.source,
    t.source_agent,
    'Corrected tool count 64 -> 63, 2026-07-25. The figure originated in my own mission brief, written before the catalogue existed, and propagated into every downstream spec. Real count verified from port/catalogue.json (55 from website-design.com + 8 from websitedesign.com, the 9th being the skipped LLM builder). See webdesign_couk/SQL_p4_fix_tool_count.sql.',
    true,
    'webdesign-couk-standup'
FROM old t;

DO $verify$
DECLARE v_site uuid; v_bad int; v_good int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT count(*) INTO v_bad FROM site_specs
     WHERE site_id = v_site AND is_current AND data::text ILIKE '%sixty-four%';
    IF v_bad > 0 THEN
        RAISE EXCEPTION 'still % current spec(s) claiming sixty-four tools', v_bad;
    END IF;

    SELECT count(*) INTO v_good FROM site_specs
     WHERE site_id = v_site AND is_current AND data::text ILIKE '%sixty-three%';
    IF v_good < 5 THEN
        RAISE EXCEPTION 'expected the corrected phrase in several specs, found %', v_good;
    END IF;

    -- One current row per aspect, as idx_site_specs_current requires.
    IF EXISTS (
        SELECT 1 FROM site_specs WHERE site_id = v_site AND is_current
         GROUP BY aspect HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'an aspect ended up with more than one current row';
    END IF;

    RAISE NOTICE 'tool count corrected to sixty-three across % spec(s)', v_good;
END
$verify$;

COMMIT;
