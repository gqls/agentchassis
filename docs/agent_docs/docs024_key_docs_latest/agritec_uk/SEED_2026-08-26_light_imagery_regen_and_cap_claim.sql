-- SEED_2026-08-26_light_imagery_regen_and_cap_claim.sql
--
-- Two dispatches from the 2026-08-25 handoff's priority list, revised against the
-- 2026-08-26 morning queue (the design-discovery rotation visited overnight —
-- bugs_open/401 / peer message from the webdesign-tool-rebuilds lane):
--
-- 1. LIGHT-PALETTE IMAGERY REGENERATION (handoff §3.1). All 17 assets were
--    generated 2026-08-24 under the DARK imagery_style_guide; the guide went
--    light on 2026-08-25, the assets did not. generate_image reads the guide
--    LIVE per run (generate_image_actions.go, getImageryStyleGuideForSite), so
--    re-firing the same items regenerates light. Both framework emitters skip
--    plan rows with an active asset (imageryplan.HasActiveAsset), so the
--    re-request must be operator-inserted. Safe because StoreAssetAction is an
--    UPSERT on (site_id, asset_key) WHERE status='active'
--    (v3_site_actions.go:3359): the existing rows update in place, same served
--    URLs, no window with a missing asset; only LOCKED assets refuse, and none
--    of the 17 is locked (checked 2026-08-26). idx_swi_dedup permits the
--    re-insert because all 17 predecessors are terminal ('complete'); the Go
--    anti-churn brake does not see raw inserts (load_work_item_actions.go names
--    raw-INSERT writers as bypassing that door, with claim as the backstop).
--    spec.check stays 'emit_imagery_items', byte-compatible with what
--    image-build-handler has already processed 17 times for this site;
--    provenance is carried by source/created_by.
--    The LOGO plan prompt said "suitable for dark backgrounds", and logos get
--    NO direction from the style guide (directionForKind excludes them), so the
--    prompt is amended at the AUTHORITY — site_plan_imagery, not the derived
--    work-item spec (the pages.sections lesson, RUNBOOK §12) — and the logo
--    item's spec.prompt is taken from the amended plan row.
--    NOT touched: the two OPEN needs_imagery items filed by design-discovery on
--    2026-08-26 00:24 (content heroes, content_hero_* keys — disjoint from
--    these 17), and the rotation's own acceptance_run/audit_tool items.
--    The handoff's acceptance_run dispatch (§3.2) is DROPPED from this seed:
--    design-discovery filed the identical key at 2026-08-26 00:24:47 and it is
--    open at 'triaged' — inserting it again would violate idx_swi_dedup.
--
-- 2. THE GARBLED AGREEMENT-CAP CLAIM (handoff §3.3, claims_unverified item
--    claims_llm_agritec.uk). The stacking explainer asserts "Defra and the
--    Rural Payments Agency have capped SFI26 at 100,000 agreements per year"
--    and then elaborates the wrong mechanism explicitly ("a national limit on
--    the number of agreements accepted, not a limit on what any one agreement
--    can contain") — which is BACKWARDS relative to the registered fact
--    CIT-86c4010f7cdf820d: a £100,000 VALUE cap per agreement year, enforced
--    at the application service. No register fact supports a national
--    agreement-volume cap, and the owner rulings of 2026-08-21/24 (no
--    unsourced figure anywhere; cite everything) decide the direction. Fix
--    route: content_rewrite, mode=edit_live, handler page-build-handler — the
--    same route the bugfix_391_cta_relevance lane ran fleet-wide on
--    2026-08-25. The claims_unverified item itself is deliberately LEFT OPEN:
--    the revalidator closes it once the copy no longer carries the claim
--    (observed arm 'resolved_all_gates_passed' on gamesdesign.co.uk,
--    2026-08-18).
--
-- Applied by hand via psql (lane convention). Sequential statements in ONE
-- transaction; DO/RAISE guards assert the pre-state (RUNBOOK §10).

BEGIN;

-- Guard 1: the logo plan row exists, is unlocked, and still carries the dark
-- wording this seed rewrites.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id = spi.plan_id
    JOIN sites s ON s.id = sp.site_id
   WHERE s.domain = 'agritec.uk' AND sp.is_current
     AND spi.scope = 'site' AND spi.key = 'logo'
     AND spi.locked_at IS NULL
     AND spi.prompt LIKE '%suitable for dark backgrounds%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'pre-state: expected exactly 1 unlocked dark-worded logo plan row, found %', n;
  END IF;
END $$;

-- Guard 2: the 17 originals are terminal and none of their keys is open
-- (another session or the loop could have re-queued them since this was read).
DO $$
DECLARE n_complete int; n_open int;
BEGIN
  SELECT count(*) FILTER (WHERE wi.status = 'complete'),
         count(*) FILTER (WHERE wi.status NOT IN
           ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled'))
    INTO n_complete, n_open
    FROM site_work_items wi
    JOIN sites s ON s.id = wi.site_id
   WHERE s.domain = 'agritec.uk'
     AND wi.item_type = 'needs_imagery'
     AND wi.item_key LIKE 'needs_imagery:%'
     AND wi.item_key NOT LIKE '%content_hero%';
  IF n_complete <> 17 OR n_open <> 0 THEN
    RAISE EXCEPTION 'pre-state: expected 17 complete / 0 open plan-imagery items, found % / %',
      n_complete, n_open;
  END IF;
END $$;

-- Guard 3: the content_rewrite key is not open, and the defect is still present
-- in the article-body component (the rewrite corrects something that exists).
DO $$
DECLARE n_open int; n_defect int;
BEGIN
  SELECT count(*) INTO n_open
    FROM site_work_items wi
    JOIN sites s ON s.id = wi.site_id
   WHERE s.domain = 'agritec.uk'
     AND wi.item_key = 'content_rewrite:stacking-agricultural-scheme-actions:agreement-cap'
     AND wi.status NOT IN
       ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  SELECT count(*) INTO n_defect
    FROM pages p
    JOIN sites s ON s.id = p.site_id
    JOIN page_components pc ON pc.page_id = p.id
   WHERE s.domain = 'agritec.uk'
     AND p.name = 'stacking-agricultural-scheme-actions'
     AND pc.content_data::text LIKE '%100,000 agreements%';
  IF n_open <> 0 OR n_defect < 1 THEN
    RAISE EXCEPTION 'pre-state: rewrite key open=% (want 0), defect present in % component(s) (want >=1)',
      n_open, n_defect;
  END IF;
END $$;

-- 1a. Amend the logo prompt at the authority.
UPDATE site_plan_imagery spi
   SET prompt = replace(spi.prompt, 'suitable for dark backgrounds', 'suitable for light backgrounds'),
       source = 'manual'
  FROM site_plans sp, sites s
 WHERE sp.id = spi.plan_id AND s.id = sp.site_id
   AND s.domain = 'agritec.uk' AND sp.is_current
   AND spi.scope = 'site' AND spi.key = 'logo';

-- 1b. Re-request all 17, copying the framework's own rows (severity, priority,
-- spec, keys) so nothing downstream meets a novel shape. The logo item's
-- spec.prompt is taken from the amended plan row.
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, batch_id, pipeline, approval_mode)
SELECT wi.site_id,
       'operator',
       'needs_imagery',
       wi.severity,
       'Regenerate under the LIGHT imagery_style_guide (asset generated 2026-08-24 under the dark guide): '
         || (wi.spec->>'asset_key'),
       CASE WHEN wi.item_key = 'needs_imagery:site:-:logo'
            THEN jsonb_set(wi.spec, '{prompt}',
                   to_jsonb((SELECT spi.prompt
                               FROM site_plan_imagery spi
                               JOIN site_plans sp ON sp.id = spi.plan_id
                              WHERE sp.site_id = wi.site_id AND sp.is_current
                                AND spi.scope = 'site' AND spi.key = 'logo')))
            ELSE wi.spec
       END,
       wi.priority,
       'image-build-handler',
       'triaged',
       'agritek-session-2026-08-26',
       wi.item_key,
       'af3e9ffa-5c16-40e2-afcc-3b77b9bf5874'::uuid,
       wi.pipeline,
       wi.approval_mode
  FROM site_work_items wi
  JOIN sites s ON s.id = wi.site_id
 WHERE s.domain = 'agritec.uk'
   AND wi.item_type = 'needs_imagery'
   AND wi.status = 'complete';

-- Guard 4: exactly 17 were queued.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM site_work_items
   WHERE batch_id = 'af3e9ffa-5c16-40e2-afcc-3b77b9bf5874'::uuid;
  IF n <> 17 THEN
    RAISE EXCEPTION 'post-insert: expected 17 queued regeneration items, found %', n;
  END IF;
END $$;

-- 2. The agreement-cap copy correction.
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, page_id, pipeline, approval_mode)
SELECT s.id,
       'operator',
       'content_rewrite',
       'high',
       'Correct the garbled SFI26 agreement-cap claim on stacking-agricultural-scheme-actions (claims_unverified ruling: the copy inverted registered fact CIT-86c4010f7cdf820d)',
       jsonb_build_object(
         'mode', 'edit_live',
         'source', 'agritek-session-2026-08-26',
         'page_name', 'stacking-agricultural-scheme-actions',
         'original_pipeline', 'build',
         'suggestion',
         'Correct ONE factual error on this page and change nothing else. The section headed "The agreement cap" currently claims that Defra and the Rural Payments Agency have capped SFI26 at 100,000 agreements per year, and describes this as a national limit on the number of agreements accepted rather than a limit on what any one agreement can contain. That is wrong, and it is backwards. The registered fact (gov.uk, SFI26 scheme rules and guidance) is: the SFI26 annual agreement value cap is 100,000 pounds per agreement year, and the application service will prevent submission of applications exceeding this limit. It is a cap on the VALUE of each individual agreement, not a national count of agreements. No evidence supports any national limit on the number of agreements, so no such claim may appear anywhere in the rewritten section. Rewrite the "The agreement cap" section so it states the 100,000-pound per-agreement annual value cap correctly, keeps a visible citation to the gov.uk SFI26 scheme rules and guidance, and relates the cap honestly to this page''s arithmetic: a stack whose annual total exceeds 100,000 pounds cannot all be entered under one agreement, which is why the site''s SFI26 Revenue Stacker models the cap. Leave every other sentence on the page exactly as it is.'
       ),
       90,
       'page-build-handler',
       'triaged',
       'agritek-session-2026-08-26',
       'content_rewrite:stacking-agricultural-scheme-actions:agreement-cap',
       '5ecd80f4-1474-4b6c-aa97-358dc37d2f3c'::uuid,
       'build',
       'auto'
  FROM sites s
 WHERE s.domain = 'agritec.uk';

COMMIT;
