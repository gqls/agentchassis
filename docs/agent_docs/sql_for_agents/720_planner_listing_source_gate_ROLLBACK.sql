-- 720_planner_listing_source_gate_ROLLBACK.sql — inverse of 720.
-- Restores rule 3's prior text and removes the enforce_listing_sources flag
-- (key deleted, not set false — absent is the documented default). Refuses if
-- 720's text is not present exactly once (i.e. 720 not applied, or a later
-- migration edited inside the block — resolve THAT first, by hand).

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '720_ROLLBACK REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner',
                         '720_planner_listing_source_gate_ROLLBACK.sql: pre-revert');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  applied text := $RA720$3. Pages with page_type entity-page, tool, blog-index, blog-post may have empty sections arrays. An entity-directory page should name its listing component in sections. A LISTING page — news-index, entity-directory, section-index, or any page whose purpose is a list of items — may only be planned when the site's item source for it exists: a recommended news feed (or seeded sources) for a news page; an opted-in directory kind or a configured business directory for a directory page; child pages in this same plan for a section index. Do NOT plan a glossary, showcase or similar collection page unless the brief names a live producer for its items — a listing page with no item source ships as prose about itself. Validation holds back listing pages whose item source resolves to zero and records each as a capability gap naming what to enable$RA720$;
  original text := $A720$3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays$A720$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '720_ROLLBACK: prompt_template not found'; END IF;

  n := (length(tpl) - length(replace(tpl, applied, ''))) / length(applied);
  IF n <> 1 THEN
    RAISE EXCEPTION '720_ROLLBACK REFUSED: 720 text found % times, expected 1 — not applied, or edited since', n;
  END IF;

  newtpl := replace(tpl, applied, original);

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false)
           #- '{workflow,steps,validate_plan,config,enforce_listing_sources}',
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '720_ROLLBACK: updated % rows, expected exactly 1', n; END IF;
END $do$;

DO $$
DECLARE tpl text; flag text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}',
         default_config #>> '{workflow,steps,validate_plan,config,enforce_listing_sources}'
    INTO tpl, flag
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays' in tpl) = 0 THEN
    RAISE EXCEPTION '720_ROLLBACK VERIFY: original rule 3 not restored';
  END IF;
  IF flag IS NOT NULL THEN
    RAISE EXCEPTION '720_ROLLBACK VERIFY: enforce_listing_sources still present (%)' , flag;
  END IF;
END $$;

COMMIT;
