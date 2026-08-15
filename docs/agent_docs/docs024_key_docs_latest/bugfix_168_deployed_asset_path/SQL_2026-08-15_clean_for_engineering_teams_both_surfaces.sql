-- SQL_2026-08-15_clean_for_engineering_teams_both_surfaces.sql
--
-- Removes a FALSE claim ("90,790 orchestration state records") from
-- leopardessconsulting.co.uk/for-engineering-teams, component 9ddedb63
-- (slot generic-text-block), on BOTH surfaces in ONE transaction.
--
-- WHY BOTH SURFACES: ScanDeployedClaims reads rendered_html AND content_data,
-- and the claim-granular gate searches html || contentJSON. A one-surface edit
-- leaves the finding standing behind a COMPLETED orchestration.
--
-- THE CLAIM IS GENUINELY FALSE, independent of the test it enables:
--   page says          90,790
--   register fact C4    4,595  (gte, verified 2026-08-15)
--   live table count    4,818  (SELECT count(*) FROM orchestration_states)
-- ~19x overclaim. Minimal deletion of one sentence, no rewrite, no substituted
-- figure, no connective repair needed (owner 2026-08-06: minimal deletion is
-- not writing).
--
-- ⚠ THE PAGE IS DELIBERATELY *NOT* REDEPLOYED AFTERWARDS. That is the point:
-- it makes newest_component_update > deployed_at, which is the input the
-- published gate refuses on (expected arm gate_published_correction_unpublished
-- on the next daily sweep). Redeploy immediately after the observation.
--
-- USAGE — induce the guard BEFORE trusting it:
--   psql -v expect=999 -f this.sql   -> MUST abort (proves the guard fires)
--   psql -v expect=166 -f this.sql   -> commits
-- The GUC dance is required: psql does NOT interpolate :vars inside a
-- dollar-quoted body.

BEGIN;
SET LOCAL app.expect_delta = :'expect';

DO $$
DECLARE
  k CONSTANT text := ' The orchestration state table has passed 90,790 records from our own systems — that is the operational history the platform runs on, and it is readable at any point.';
  cid CONSTANT uuid := '9ddedb63-9c0c-4e73-a62c-b158172a9ebb';
  v_expect      int := current_setting('app.expect_delta')::int;
  v_html_before int; v_json_before int;
  v_html_after  int; v_json_after  int;
  v_rows        int;
  v_left        int;
BEGIN
  SELECT length(rendered_html), length(content_data->>'content')
    INTO v_html_before, v_json_before
    FROM page_components WHERE id = cid;

  UPDATE page_components
     SET rendered_html = replace(rendered_html, k, ''),
         content_data  = jsonb_set(content_data, '{content}',
                           to_jsonb(replace(content_data->>'content', k, ''))),
         updated_at    = now()
   WHERE id = cid;
  GET DIAGNOSTICS v_rows = ROW_COUNT;

  SELECT length(rendered_html), length(content_data->>'content')
    INTO v_html_after, v_json_after
    FROM page_components WHERE id = cid;

  IF v_rows <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 row updated, got %', v_rows;
  END IF;
  IF (v_html_before - v_html_after) <> v_expect THEN
    RAISE EXCEPTION 'rendered_html delta % <> expected % (before % after %)',
      v_html_before - v_html_after, v_expect, v_html_before, v_html_after;
  END IF;
  IF (v_json_before - v_json_after) <> v_expect THEN
    RAISE EXCEPTION 'content_data delta % <> expected % (before % after %)',
      v_json_before - v_json_after, v_expect, v_json_before, v_json_after;
  END IF;

  -- the claim must be GONE from BOTH surfaces, not merely shorter
  SELECT count(*) INTO v_left FROM page_components
   WHERE id = cid
     AND (rendered_html LIKE '%90,790%' OR content_data->>'content' LIKE '%90,790%');
  IF v_left <> 0 THEN
    RAISE EXCEPTION 'claim 90,790 still present on a surface after the edit';
  END IF;

  RAISE NOTICE 'OK rendered_html % -> %, content % -> %, claim absent from both surfaces',
    v_html_before, v_html_after, v_json_before, v_json_after;
END $$;

COMMIT;
