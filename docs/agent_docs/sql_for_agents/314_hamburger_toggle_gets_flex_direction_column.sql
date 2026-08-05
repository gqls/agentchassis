-- 314 — bug 200: the mobile hamburger renders as ONE bar on every layout.
--
-- Every layouts.css_template sets `.mobile-menu-toggle { display: inline-flex; }`
-- in its mobile media query with no flex-direction, so the button is a ROW flex
-- container and the three <span> bars fuse side-by-side into one ~72px line
-- (observed bar width = exactly 3 span-widths — bugs_open/200 §2).
--
-- This is a SURGICAL replace on the live rows, NOT a seed-driver re-run:
-- five rows (brochure-bold, media-grid, docs-sidebar, high-energy,
-- affiliate-hub) carry 2026-07-02 changes the seed files never got, and
-- tool-portal-light has no seed in the layouts dir at all — the driver's own
-- "re-running is safe" header is WRONG for those six (measured 2026-08-05,
-- drift_check in bugs_open/200's trail). The seed files receive the same edit
-- separately so a future driver run cannot regress this.
--
-- Pre-measured 2026-08-05: the exact bare rule appears EXACTLY ONCE in each of
-- the 18 live rows, so replace() touches one site per row and nothing else.
-- The verify block was INDUCED before the apply (run against the pre-fix state
-- it raised '200: 18 rows still carry...'), so it is known to be able to fail.

BEGIN;

UPDATE layouts
SET css_template = replace(
        css_template,
        '.mobile-menu-toggle { display: inline-flex; }',
        '.mobile-menu-toggle { display: inline-flex; flex-direction: column; justify-content: center; align-items: center; }'),
    updated_at = NOW()
WHERE css_template LIKE '%.mobile-menu-toggle { display: inline-flex; }%';

DO $$
DECLARE bad int; fixed int;
BEGIN
    SELECT count(*) INTO bad FROM layouts
     WHERE css_template LIKE '%.mobile-menu-toggle { display: inline-flex; }%';
    SELECT count(*) INTO fixed FROM layouts
     WHERE css_template LIKE '%.mobile-menu-toggle { display: inline-flex; flex-direction: column; justify-content: center; align-items: center; }%';
    IF bad <> 0 THEN
        RAISE EXCEPTION '200: % rows still carry the bare inline-flex toggle rule', bad;
    END IF;
    IF fixed <> 18 THEN
        RAISE EXCEPTION '200: expected 18 fixed rows, found %', fixed;
    END IF;
END $$;

COMMIT;
