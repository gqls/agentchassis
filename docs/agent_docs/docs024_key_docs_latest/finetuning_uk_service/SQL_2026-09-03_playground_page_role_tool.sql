-- SQL_2026-09-03_playground_page_role_tool.sql — finetuning.uk /playground.html takes the "tool"
-- page role, so the generated chat widget can attach to it (Phase P step 3, path (a)).
--
-- WHY: the first add_tool run (item 74b725b6, orch f5da1e98, 19:53Z) was REFUSED at save_tool by
-- UpsertPageForRole's hasShipped arm: "page playground is live as page_type=content
-- (build_status=deployed); tool-generator wants to write it as tool. Re-typing a live page changes
-- what it serves, so it is not done automatically (bugs_open/175)". It filed the decision as
-- mistyped_deployed_page 2a2725cc with this exact resolution. The decision is the owner's direction
-- (2026-09-03: "the site is the tool"; path (a)); the page IS the tool's page.
-- What a 'tool' page_type changes (grep 2026-09-03 20:00Z, non-test Go): tool-eligibility and
-- tool-health/recreation discovery checks now include it (they flag a tool page with no widget — a
-- transient true positive until the widget lands); owned_page_guard and cta_positional read it;
-- page-build-handler and page-rebuild do not branch on it. Reversible: the previous value is below.
-- PREVIOUS: page_type='content' (page 4234fe60-4332-4f4a-b256-ba420716a884, updated_at 2026-09-03 19:20:58Z).
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE v_page uuid := '4234fe60-4332-4f4a-b256-ba420716a884'; n int;
BEGIN
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
    AND name='playground' AND page_type='content' AND status='active' AND build_status='deployed';
  IF n<>1 THEN RAISE EXCEPTION 'pre-flight: /playground.html is not the deployed content page this file expects (%)', n; END IF;
  UPDATE pages SET page_type='tool', updated_at=NOW() WHERE id=v_page;
  UPDATE site_work_items SET status='complete', completed_at=NOW(),
         result = COALESCE(result,'{}'::jsonb) || jsonb_build_object('decision','page re-typed to tool by the finetuning lane on the owner''s path (a) direction, 2026-09-03')
   WHERE id='2a2725cc-cf28-48b8-a760-a7f419b25ff1' AND item_type='mistyped_deployed_page' AND status='needs_human_review';
  IF NOT FOUND THEN RAISE EXCEPTION 'the mistyped_deployed_page decision item 2a2725cc is not open'; END IF;
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND page_type='tool';
  IF n<>1 THEN RAISE EXCEPTION 'post: page_type not tool'; END IF;
  RAISE NOTICE '/playground.html now holds the tool role; decision item 2a2725cc closed';
END $$;
COMMIT;
