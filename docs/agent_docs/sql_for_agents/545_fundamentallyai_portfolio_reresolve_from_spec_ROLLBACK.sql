-- 545_fundamentallyai_portfolio_reresolve_from_spec_ROLLBACK.sql
--
-- 545 is a DERIVATION, not an authored value: it copies `data->'projects'` from
-- fundamentallyai.com's current `portfolio` site_spec into the portfolio-showcase
-- component. So there is nothing here to invert on its own terms — re-running 545
-- against a given spec always produces the same component content, and running it
-- twice changes nothing the second time.
--
-- THEREFORE, TO REVERT: roll back the SOURCE, then re-derive.
--
--   1. psql -f 544_fundamentallyai_portfolio_logo_extension_ROLLBACK.sql
--   2. psql -f 545_fundamentallyai_portfolio_reresolve_from_spec.sql
--
-- Step 2 is not a typo. 545's own ordering guard refuses to run while the spec
-- still carries `.jpg`, so after step 1 that guard is what you must defeat — and
-- you cannot, which is deliberate. Use the statement below instead, which is the
-- only supported way to put the `.jpg` values back into the served copy.
--
-- ⚠ READ 544's ROLLBACK HEADER FIRST. Restoring `.jpg` re-arms bugs_open/235's
-- artefact: the portfolio then shows two hero-encoded (1600x900 q85) renders of a
-- LOGO. Nothing 404s and nothing alarms, which is how it went unnoticed for a
-- month. The only honest reason to do this is if the `.png` targets stop serving,
-- and that would be a different fault — check before reverting:
--   curl -sI https://relojistas.com/assets/images/logo.png
--   curl -sI https://idea.uk/assets/images/logo.png
--
-- After this, dispatch a page rerender for fundamentallyai.com/index.html — this
-- writes content_data only, and rendered_html is regenerated from it by the render
-- seam, never by hand.

BEGIN;

UPDATE page_components pc
SET content_data = jsonb_set(
        pc.content_data,
        '{projects}',
        (
            SELECT jsonb_agg(
                       CASE
                           WHEN elem->>'domain' IN ('relojistas.com', 'idea.uk')
                                AND elem->>'logo_url' LIKE '%/assets/images/logo.png'
                               THEN jsonb_set(
                                        elem,
                                        '{logo_url}',
                                        to_jsonb(regexp_replace(elem->>'logo_url', '\.png$', '.jpg'))
                                    )
                           ELSE elem
                       END
                       ORDER BY ord
                   )
            FROM jsonb_array_elements(pc.content_data->'projects') WITH ORDINALITY t(elem, ord)
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
    v_n_projects int;
BEGIN
    SELECT count(*) FILTER (WHERE elem->>'logo_url' LIKE '%.jpg'), count(*)
      INTO v_n_jpg, v_n_projects
    FROM pages p
         JOIN page_components pc ON pc.page_id = p.id
         JOIN sites s ON p.site_id = s.id,
         LATERAL jsonb_array_elements(pc.content_data->'projects') elem
    WHERE s.domain = 'fundamentallyai.com'
      AND pc.slot_name = 'portfolio-showcase';

    IF v_n_projects <> 3 THEN
        RAISE EXCEPTION '235/545 ROLLBACK verify: component has % projects, expected 3', v_n_projects;
    END IF;
    IF v_n_jpg <> 2 THEN
        RAISE EXCEPTION '235/545 ROLLBACK verify: % logo_url end .jpg, expected 2', v_n_jpg;
    END IF;

    RAISE NOTICE '235/545 ROLLBACK: 2 .jpg restored in the served copy. The 235 artefact is re-armed; dispatch a rerender to push it live.';
END $$;

COMMIT;
