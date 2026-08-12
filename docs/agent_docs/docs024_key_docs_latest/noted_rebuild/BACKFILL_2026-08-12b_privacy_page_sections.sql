-- =============================================================================
-- noted.co.uk — part 2 of the backfill: the privacy page's SECTION plan
-- 2026-08-12. Fixes an incomplete part 1.
-- =============================================================================
--
-- WHAT PART 1 GOT WRONG. BACKFILL_2026-08-12_structure_spec_and_site_plan.sql
-- added `privacy` to site_plan_pages and stopped there. A planned page needs rows
-- in BOTH site_plan_pages AND site_plan_sections. Without sections the build runs,
-- reaches `check_has_ready_sections`, and ends `complete_error` with **no
-- __step_error recorded at all** — and the needs_page work item is still marked
-- `complete`. Measured 2026-08-12 18:48:32: page=privacy, complete_error, no step
-- error, no pages row, item complete. Nothing in agent_error_log either, because
-- that is written by the validator and this never reached validation.
--
-- So the failure signature is: a `complete` work item, no page, and silence
-- everywhere. Worth knowing — it looks nothing like the guide's failure two
-- minutes later, which DID reach validation and DID log its blocker.
--
-- COMPOSITION. Mirrors `about` (the closest existing content page): a hero plus a
-- prose block. Component names are taken from what this site's plan already uses,
-- not invented — `hero` and `generic-text-block` both appear in the current plan
-- (index/how-it-works/migrate), so they are known-good on this site's component set.
--
-- THE CONTENT ITSELF is NOT written here. The owner-approved copy lives in
-- evidence_base.supplied_copy.privacy and the writer_block instructs VERBATIM use.
-- That instruction is an LLM instruction, so it is a hope until checked: after the
-- page builds, DIFF the rendered text against the approved copy. The query is at
-- the bottom.
-- =============================================================================

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM site_plan_pages spp JOIN site_plans sp ON sp.id=spp.plan_id JOIN sites s ON s.id=sp.site_id
  WHERE s.domain='noted.co.uk' AND sp.is_current AND spp.name='privacy';
  IF n <> 1 THEN RAISE EXCEPTION 'privacy is not in the current plan (found %) — run part 1 first', n; END IF;
END $$;

INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
SELECT sp.id, 'privacy', v.ordering, v.component_name
FROM site_plans sp JOIN sites s ON s.id = sp.site_id
CROSS JOIN (VALUES
  (1, 'hero'),
  (2, 'generic-text-block')
) AS v(ordering, component_name)
WHERE s.domain='noted.co.uk' AND sp.is_current
  AND NOT EXISTS (
    SELECT 1 FROM site_plan_sections x
    WHERE x.plan_id = sp.id AND x.page_name = 'privacy' AND x.ordering = v.ordering
  );

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id JOIN sites s ON s.id=sp.site_id
  WHERE s.domain='noted.co.uk' AND sp.is_current AND sps.page_name='privacy';
  IF n <> 2 THEN RAISE EXCEPTION 'expected 2 privacy sections, found %', n; END IF;
  RAISE NOTICE 'privacy section plan backfilled: 2 sections';
END $$;

COMMIT;

-- ---------------------------------------------------------------- verify ----
SELECT sps.page_name, sps.ordering, sps.component_name
FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id JOIN sites s ON s.id=sp.site_id
WHERE s.domain='noted.co.uk' AND sp.is_current AND sps.page_name='privacy'
ORDER BY sps.ordering;

-- AFTER the page builds — does it carry the APPROVED copy, or did the writer
-- paraphrase it? The writer_block says verbatim; this is what checks.
--   SELECT pc.slot_name, left(regexp_replace(pc.rendered_html,'<[^>]+>',' ','g'), 400)
--   FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
--   WHERE s.domain='noted.co.uk' AND p.name='privacy' ORDER BY pc.position;
