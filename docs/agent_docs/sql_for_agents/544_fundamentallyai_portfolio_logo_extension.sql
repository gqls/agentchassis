-- 544_fundamentallyai_portfolio_logo_extension.sql
--
-- bugs_open/235's residual, and the last live consequence of a defect that was
-- fixed at source twelve days ago.
--
-- WHAT 235 WAS. `image-build-handler`'s brand-update branch carried a STATIC
-- `"purpose": "hero"` on its `store_imagery_brand_asset` step, so every asset it
-- stored was typed a hero — including logos. `DeployedAssetPath` then takes the
-- FILENAME from asset_key and the EXTENSION plus resize class from purpose, so a
-- logo item shipped as `logo.jpg` at hero encoding (1600x900, q85) instead of
-- `logo.png` (400x400, q90). That source defect is FIXED AND LIVE: migration 360,
-- applied 2026-08-09, replaced the static with `purpose_field:
-- "input_data.spec.purpose"`. Verified in the live config 2026-08-21 — the step
-- carries no static `purpose` key at all now.
--
-- WHAT WAS NEVER FINISHED. The artefacts that defect had already produced. On
-- 2026-08-11 a session patched the affected portfolio card URLs at
-- `page_components.content_data`, and the brochure_component_library lane recorded
-- in the same hour that the patch DID NOT SURVIVE a full rebuild: its
-- `needs_page:index` regeneration deployed at 11:26:58Z and the served index went
-- back to referencing two `.jpg` logos. Their addendum named the fix — "the durable
-- fix needs to land where the resolver reads the logo URL from (the portfolio
-- source), not in the resolved copy" — and nobody acted on it for ten days.
--
-- WHERE THE SOURCE ACTUALLY IS, and why the earlier patch could not hold. It is
-- THIS ROW: `site_specs`, aspect `portfolio`, `is_current`, fundamentallyai.com,
-- authored 2026-07-22. Its `data->'projects'` holds three entries with hard-coded
-- `logo_url` strings, two of them `.jpg`. Nothing in Go reads `aspect='portfolio'`
-- — the spec reaches the content writer as LLM context via the generic current-spec
-- load. That is precisely why a `content_data` patch holds through assembly and
-- rerender (neither re-consults the spec) but dies at the first `needs_page`
-- regeneration (which does). Patching the resolved copy was treating the output.
--
-- WHY THE SPEC SAYS `.jpg` AT ALL. It was authored 2026-07-22, before migration 360
-- (08-09). At that time relojistas.com and idea.uk GENUINELY served `logo.jpg`,
-- because 235's brand-update branch had stored their logos as heroes. This row is a
-- fossil of the fixed defect, not a second instance of it — so this migration is
-- cleanup, and recurrence is already prevented at source.
--
-- BLAST RADIUS [MEASURED] 2026-08-21 — ONE ROW, ONE SITE:
--   SELECT s.domain, ss.aspect FROM site_specs ss JOIN sites s ON ss.site_id=s.id
--    WHERE ss.is_current AND ss.data::text LIKE '%logo.jpg%';
--   -> fundamentallyai.com | portfolio        (exactly one row, fleet-wide)
--
-- THE REPLACEMENT TARGETS EXIST — probed live before writing this, because a spec
-- pointing at a missing file serves a BROKEN image, which is worse than a
-- wrong-format one:
--   https://relojistas.com/assets/images/logo.png -> 200, 32,212 B
--   https://idea.uk/assets/images/logo.png        -> 200, 146,681 B
--
-- WHAT THIS DOES NOT DO. It does not delete the stale `logo.jpg` objects on those
-- two sites. 235 gates that on the owner's word, and the blocking condition was
-- fundamentallyai's index still referencing them — which this migration clears, so
-- the gate becomes open rather than satisfied. Deletion stays a human decision.
--
-- SHAPE-PRESERVING BY CONSTRUCTION. The update rebuilds `projects` element by
-- element with `jsonb_set` on `logo_url` alone, so all 8 keys per entry and the
-- array order survive. It deliberately does NOT rewrite the whole `data` blob,
-- which would silently drop anything added to the spec since 2026-07-22.
--
-- leopardessconsulting.co.uk is left untouched on purpose: it already carries
-- `.png` and is the in-row control — if it moves, the predicate is wrong.

BEGIN;

UPDATE site_specs ss
SET data = jsonb_set(
        ss.data,
        '{projects}',
        (
            SELECT jsonb_agg(
                       CASE
                           WHEN elem->>'logo_url' LIKE '%/assets/images/logo.jpg'
                               THEN jsonb_set(
                                        elem,
                                        '{logo_url}',
                                        to_jsonb(
                                            regexp_replace(elem->>'logo_url', '\.jpg$', '.png')
                                        )
                                    )
                           ELSE elem
                       END
                       ORDER BY ord
                   )
            FROM jsonb_array_elements(ss.data->'projects') WITH ORDINALITY t(elem, ord)
        )
    ),
    updated_at = now()
FROM sites s
WHERE ss.site_id = s.id
  AND s.domain = 'fundamentallyai.com'
  AND ss.aspect = 'portfolio'
  AND ss.is_current;

-- Verify as a DO block that RAISEs, never as a trailing SELECT: `ON_ERROR_STOP`
-- ignores a non-empty result set, so a block of SELECTs cannot stop the COMMIT
-- (LANDMINES.md, the 2026-08-02 RFC_006 entry).
DO $$
DECLARE
    v_data       jsonb;
    v_n_projects int;
    v_n_jpg      int;
    v_n_png      int;
    v_leopardess text;
    v_keys_min   int;
BEGIN
    SELECT ss.data INTO v_data
    FROM site_specs ss JOIN sites s ON ss.site_id = s.id
    WHERE s.domain = 'fundamentallyai.com'
      AND ss.aspect = 'portfolio'
      AND ss.is_current;

    IF v_data IS NULL THEN
        RAISE EXCEPTION '235/544 verify: no current portfolio spec for fundamentallyai.com — predicate matched nothing';
    END IF;

    -- the array is intact: still three projects, in order
    SELECT jsonb_array_length(v_data->'projects') INTO v_n_projects;
    IF v_n_projects <> 3 THEN
        RAISE EXCEPTION '235/544 verify: projects count is %, expected 3 — the rebuild lost or duplicated an entry', v_n_projects;
    END IF;

    -- no entry lost keys: every project still carries its full 8-key shape
    SELECT min(cnt) INTO v_keys_min
    FROM (
        SELECT (SELECT count(*) FROM jsonb_object_keys(elem)) AS cnt
        FROM jsonb_array_elements(v_data->'projects') elem
    ) k;
    IF v_keys_min < 8 THEN
        RAISE EXCEPTION '235/544 verify: a project entry has only % keys, expected >= 8 — jsonb_set dropped fields', v_keys_min;
    END IF;

    -- the change itself: zero .jpg logos left, three .png
    SELECT count(*) FILTER (WHERE elem->>'logo_url' LIKE '%.jpg'),
           count(*) FILTER (WHERE elem->>'logo_url' LIKE '%.png')
      INTO v_n_jpg, v_n_png
    FROM jsonb_array_elements(v_data->'projects') elem;

    IF v_n_jpg <> 0 THEN
        RAISE EXCEPTION '235/544 verify: % logo_url still ends .jpg', v_n_jpg;
    END IF;
    IF v_n_png <> 3 THEN
        RAISE EXCEPTION '235/544 verify: % logo_url end .png, expected 3', v_n_png;
    END IF;

    -- the control: leopardess was already correct and must be byte-identical
    SELECT elem->>'logo_url' INTO v_leopardess
    FROM jsonb_array_elements(v_data->'projects') elem
    WHERE elem->>'domain' = 'leopardessconsulting.co.uk';

    IF v_leopardess IS DISTINCT FROM 'https://leopardessconsulting.co.uk/assets/images/logo.png' THEN
        RAISE EXCEPTION '235/544 verify: control row moved (leopardess = %) — the predicate is wider than intended', v_leopardess;
    END IF;

    RAISE NOTICE '235/544: verified — 3 projects, 8+ keys each, 0 .jpg, 3 .png, control unmoved';
END $$;

COMMIT;
