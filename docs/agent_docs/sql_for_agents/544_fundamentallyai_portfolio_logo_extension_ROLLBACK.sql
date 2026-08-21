-- 544_fundamentallyai_portfolio_logo_extension_ROLLBACK.sql
--
-- Reverses 544: puts the two portfolio logo_url values back to `.jpg` in
-- fundamentallyai.com's current `portfolio` site_spec.
--
-- ⚠ WHAT ROLLING BACK COSTS, and why you almost certainly do not want it. The
-- `.jpg` objects those URLs point at are hero-encoded (1600x900, q85) renders of a
-- LOGO — the artefact bugs_open/235 exists to describe. Restoring them re-arms the
-- defect at its source: the next `needs_page` regeneration of fundamentallyai.com's
-- index will re-emit two `.jpg` logo references into the portfolio-showcase
-- component and redeploy them. Nothing will 404 and nothing will alarm, which is
-- exactly how this survived unnoticed for a month.
--
-- The only honest reason to run this is if the `.png` targets stop serving. Check
-- FIRST, because that — not the extension — would be the real fault:
--   curl -sI https://relojistas.com/assets/images/logo.png
--   curl -sI https://idea.uk/assets/images/logo.png
-- If those 200, roll back nothing; fix whatever else is wrong.
--
-- This reverses the SPEC only. It does not touch `page_components` — a page already
-- regenerated under 544 keeps its `.png` references until its next regeneration.
-- The two stores are independent, which is the whole lesson of 235's residual.

BEGIN;

UPDATE site_specs ss
SET data = jsonb_set(
        ss.data,
        '{projects}',
        (
            SELECT jsonb_agg(
                       CASE
                           WHEN elem->>'domain' IN ('relojistas.com', 'idea.uk')
                                AND elem->>'logo_url' LIKE '%/assets/images/logo.png'
                               THEN jsonb_set(
                                        elem,
                                        '{logo_url}',
                                        to_jsonb(
                                            regexp_replace(elem->>'logo_url', '\.png$', '.jpg')
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

-- Narrower than 544's forward predicate ON PURPOSE: it names the two domains, so a
-- rollback cannot reach leopardessconsulting.co.uk, which was `.png` before 544 and
-- must stay `.png` after any revert.
DO $$
DECLARE
    v_data       jsonb;
    v_n_jpg      int;
    v_leopardess text;
BEGIN
    SELECT ss.data INTO v_data
    FROM site_specs ss JOIN sites s ON ss.site_id = s.id
    WHERE s.domain = 'fundamentallyai.com'
      AND ss.aspect = 'portfolio'
      AND ss.is_current;

    IF v_data IS NULL THEN
        RAISE EXCEPTION '235/544 ROLLBACK verify: no current portfolio spec for fundamentallyai.com';
    END IF;

    IF jsonb_array_length(v_data->'projects') <> 3 THEN
        RAISE EXCEPTION '235/544 ROLLBACK verify: projects count is %, expected 3',
            jsonb_array_length(v_data->'projects');
    END IF;

    SELECT count(*) FILTER (WHERE elem->>'logo_url' LIKE '%.jpg')
      INTO v_n_jpg
    FROM jsonb_array_elements(v_data->'projects') elem;

    IF v_n_jpg <> 2 THEN
        RAISE EXCEPTION '235/544 ROLLBACK verify: % logo_url end .jpg, expected 2', v_n_jpg;
    END IF;

    SELECT elem->>'logo_url' INTO v_leopardess
    FROM jsonb_array_elements(v_data->'projects') elem
    WHERE elem->>'domain' = 'leopardessconsulting.co.uk';

    IF v_leopardess IS DISTINCT FROM 'https://leopardessconsulting.co.uk/assets/images/logo.png' THEN
        RAISE EXCEPTION '235/544 ROLLBACK verify: control row moved (leopardess = %)', v_leopardess;
    END IF;

    RAISE NOTICE '235/544 ROLLBACK: reverted — 2 .jpg restored, control unmoved. The 235 defect is re-armed at this spec.';
END $$;

COMMIT;
