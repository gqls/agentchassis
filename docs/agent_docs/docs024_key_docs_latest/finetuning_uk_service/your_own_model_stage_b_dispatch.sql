-- your_own_model_stage_b_dispatch.sql — finetuning.uk /your-own-model.html: the 443 STAGE B
-- rebuild (641 option A4 live 2026-09-03 19:26:46Z), by the RUNBOOK recipe "Rebuild an
-- ALREADY-DEPLOYED page": the brief is the page's ORIGINAL gap_plan item (a38e62c1, 2026-08-24,
-- 6 sections, no mode, no not_dispatchable), copied whole with a `reason` added; item_key kept.
-- The page already carries /playground.html in required_links (owner verdict 2026-09-03) and
-- its six A4 subjects (SQL_2026-09-03_section_subjects_A4_reauthor.sql).
-- Same gates as the technical-details file: G1 (641 live), G2 (deployed, no open needs_content_page),
-- post-conditions (no mode, no not_dispatchable, 6 sections, subjects 6 and A4-shaped).
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE
  v_site uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';
  v_page uuid := 'a8909fc1-f1ff-43fe-842c-5ce364b8b182';   -- /your-own-model.html
  v_orig uuid := 'a38e62c1-4e38-4533-9c27-e2bf893b2280';   -- gap_plan_new_your-own-model_<site>, 2026-08-24
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions WHERE type='page-content-writer' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND default_config::text LIKE '%current_section.subject%';
  IF n <> 1 THEN RAISE EXCEPTION 'G1 FAILED: 641 not on the live writer (%)', n; END IF;
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND site_id=v_site AND build_status='deployed' AND status='active'
    AND jsonb_array_length(section_subjects)=6 AND section_subjects->>0 ~ '^[A-Z]';
  IF n <> 1 THEN RAISE EXCEPTION 'G2 FAILED: page not deployed/active with 6 A4 subjects (%)', n; END IF;
  SELECT count(*) INTO n FROM site_work_items WHERE page_id=v_page AND item_type='needs_content_page'
    AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled','deferred');
  IF n <> 0 THEN RAISE EXCEPTION 'G2 FAILED: % open needs_content_page on the page', n; END IF;
  SELECT count(*) INTO n FROM site_work_items w WHERE w.id=v_orig AND w.page_id=v_page AND w.item_type='needs_content_page'
    AND NOT (w.spec ? 'not_dispatchable') AND NOT (w.spec ? 'mode') AND jsonb_array_length(w.spec->'sections')=6;
  IF n <> 1 THEN RAISE EXCEPTION 'original brief % is not a clean 6-section gap_plan spec', v_orig; END IF;

  UPDATE pages SET build_status='planned', updated_at=NOW() WHERE id=v_page AND build_status='deployed';
  INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary,
                               priority, handler_agent, status, created_by, spec, item_key, batch_id)
  SELECT w.site_id, w.page_id, w.source, w.pipeline, w.item_type, w.severity,
         'your-own-model: 443 Stage B rebuild (641/A4 live) — original 2026-08-24 brief, A4 subjects, playground link required',
         40, 'page-build-handler', 'triaged', 'finetuning_uk_service_lane',
         w.spec || jsonb_build_object('reason', 'bugs_open/443 Stage B, 2026-09-03: first rebuild with 641/A4 (section subject printed as the opening line); brief unchanged from 2026-08-24'),
         w.item_key, gen_random_uuid()
    FROM site_work_items w WHERE w.id=v_orig;

  SELECT count(*) INTO n FROM site_work_items w WHERE w.page_id=v_page AND w.status='triaged' AND w.item_type='needs_content_page'
    AND NOT (w.spec ? 'not_dispatchable') AND NOT (w.spec ? 'mode') AND jsonb_array_length(w.spec->'sections')=6 AND w.spec ? 'reason';
  IF n <> 1 THEN RAISE EXCEPTION 'post: expected 1 clean triaged item, found %', n; END IF;
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND build_status='planned';
  IF n <> 1 THEN RAISE EXCEPTION 'post: page not planned'; END IF;
  RAISE NOTICE 'Stage B item inserted for your-own-model; page planned';
END $$;
COMMIT;
