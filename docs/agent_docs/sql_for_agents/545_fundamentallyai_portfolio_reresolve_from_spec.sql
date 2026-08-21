-- 545_fundamentallyai_portfolio_reresolve_from_spec.sql
--
-- The second half of bugs_open/235's residual. 544 corrected the SOURCE (the
-- current `portfolio` site_spec). This re-applies the resolution to the copy that
-- is actually served, so the live page stops showing two hero-encoded JPEGs
-- without waiting for the site's next full regeneration.
--
-- WHY THIS IS A RE-RESOLUTION AND NOT A HAND EDIT — the distinction matters,
-- because a hand-typed content value is exactly what this estate does not want.
-- The portfolio-showcase component is a VERBATIM COPY of the spec. Measured
-- 2026-08-21, immediately after 544 and before this file, for all three projects:
--
--   (spec_element - 'logo_url') = (component_element - 'logo_url')  ->  t, t, t
--
-- Every one of the other seven keys (title, domain, live_url, logo_alt,
-- build_time, built_with, description) is byte-identical between the two stores.
-- No LLM authors these fields; `sourceResolver.ensureSpecs`
-- (plan_sections_action.go:279-309) loads every current spec keyed by aspect and
-- the resolver copies them through. So this migration does not INVENT a value —
-- it takes `data->'projects'` from the spec row and writes it into the component,
-- which is precisely what the next regeneration would compute. The two stores
-- converge either way; this only decides whether the live page is correct now or
-- at the site's next rebuild.
--
-- WHY NOT A FULL `needs_page` REGENERATION. That is the other way to re-resolve,
-- and it was rejected on proportionality: it regenerates every section of a live
-- business homepage — rewriting copy the LLM authored — to correct two file
-- extensions. The blast radius of the remedy would exceed the blast radius of the
-- defect. This writes the two values and touches nothing else, and the equality
-- assertion above is what makes that claim checkable rather than hopeful.
--
-- WHY THE EARLIER ATTEMPT FAILED AND THIS ONE DOES NOT. On 2026-08-11 a session
-- patched this same column and the brochure_component_library lane recorded, the
-- same hour, that the next `needs_page` regeneration put the `.jpg` references
-- straight back. That patch was futile because the SOURCE still said `.jpg`, so
-- every regeneration re-injected it. 544 fixed the source first, so the identical
-- write is now durable: regeneration recomputes `.png` from the corrected spec.
-- Order is the whole difference between the two, and 544 must be applied first —
-- the guard below enforces it rather than trusting the filename ordering.
--
-- rendered_html IS NOT TOUCHED HERE. It is regenerated from content_data by a
-- page rerender, which is dispatched after this file applies. Writing HTML
-- directly would bypass the render seam and its link repair.

BEGIN;

-- Refuse to run ahead of 544. Without this, applying 545 alone would write the
-- stale `.jpg` values back over themselves from an uncorrected spec and report
-- success — a no-op that reads exactly like a fix.
DO $$
DECLARE
    v_jpg int;
BEGIN
    SELECT count(*) INTO v_jpg
    FROM site_specs ss JOIN sites s ON ss.site_id = s.id,
         LATERAL jsonb_array_elements(ss.data->'projects') elem
    WHERE s.domain = 'fundamentallyai.com'
      AND ss.aspect = 'portfolio'
      AND ss.is_current
      AND elem->>'logo_url' LIKE '%.jpg';

    IF v_jpg > 0 THEN
        RAISE EXCEPTION '235/545: the portfolio spec still carries % .jpg logo_url — apply 544 first; running this now would re-copy the stale values and look like a fix', v_jpg;
    END IF;
END $$;

-- Re-resolve: take the projects array from the spec, write it to the component.
UPDATE page_components pc
SET content_data = jsonb_set(
        pc.content_data,
        '{projects}',
        (
            SELECT ss.data->'projects'
            FROM site_specs ss
            WHERE ss.site_id = p.site_id
              AND ss.aspect = 'portfolio'
              AND ss.is_current
        )
    ),
    updated_at = now()
FROM pages p, sites s
WHERE pc.page_id = p.id
  AND p.site_id = s.id
  AND s.domain = 'fundamentallyai.com'
  AND pc.slot_name = 'portfolio-showcase';

DO $$
DECLARE
    v_n_jpg      int;
    v_n_png      int;
    v_n_projects int;
    v_mismatch   int;
BEGIN
    SELECT count(*) FILTER (WHERE elem->>'logo_url' LIKE '%.jpg'),
           count(*) FILTER (WHERE elem->>'logo_url' LIKE '%.png'),
           count(*)
      INTO v_n_jpg, v_n_png, v_n_projects
    FROM pages p
         JOIN page_components pc ON pc.page_id = p.id
         JOIN sites s ON p.site_id = s.id,
         LATERAL jsonb_array_elements(pc.content_data->'projects') elem
    WHERE s.domain = 'fundamentallyai.com'
      AND pc.slot_name = 'portfolio-showcase';

    IF v_n_projects <> 3 THEN
        RAISE EXCEPTION '235/545 verify: component has % projects, expected 3', v_n_projects;
    END IF;
    IF v_n_jpg <> 0 THEN
        RAISE EXCEPTION '235/545 verify: % logo_url still end .jpg in the served copy', v_n_jpg;
    END IF;
    IF v_n_png <> 3 THEN
        RAISE EXCEPTION '235/545 verify: % logo_url end .png, expected 3', v_n_png;
    END IF;

    -- the identity that makes this a re-resolution: component == spec, exactly
    SELECT count(*) INTO v_mismatch
    FROM (
        SELECT elem AS c FROM pages p
             JOIN page_components pc ON pc.page_id = p.id
             JOIN sites s ON p.site_id = s.id,
             LATERAL jsonb_array_elements(pc.content_data->'projects') elem
        WHERE s.domain = 'fundamentallyai.com' AND pc.slot_name = 'portfolio-showcase'
    ) comp
    FULL OUTER JOIN (
        SELECT elem AS sp FROM site_specs ss JOIN sites s ON ss.site_id = s.id,
             LATERAL jsonb_array_elements(ss.data->'projects') elem
        WHERE s.domain = 'fundamentallyai.com' AND ss.aspect = 'portfolio' AND ss.is_current
    ) spec ON comp.c = spec.sp
    WHERE comp.c IS NULL OR spec.sp IS NULL;

    IF v_mismatch <> 0 THEN
        RAISE EXCEPTION '235/545 verify: % project element(s) differ between the component and the spec — this write was not a faithful re-resolution', v_mismatch;
    END IF;

    RAISE NOTICE '235/545: verified — 3 projects, 0 .jpg, 3 .png, and the component is now element-for-element identical to the spec';
END $$;

COMMIT;
