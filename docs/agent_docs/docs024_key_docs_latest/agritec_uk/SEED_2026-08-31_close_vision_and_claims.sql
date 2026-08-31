-- SEED_2026-08-31_close_vision_and_claims.sql
--
-- Two operator rulings on agritec.uk needs_human_review items.
--
-- 1. vision_finding on the SFI26 stacker (2026-08-26 12:25, tool-acceptance
--    vision pass): every Subtotal cell shows an em-dash. RULING: AS DESIGNED.
--    The dash is the tool's pre-calculation empty state — the vision pass
--    screenshots the page WITHOUT pressing "Calculate scenario", so it saw the
--    honest not-yet-calculated state. The considered alternative (pre-filling
--    "£0.00") was rejected because it asserts a computed value that has not
--    been computed — this site's whole discipline is not publishing figures
--    that nothing produced. The 8/9-passing acceptance run plus the audit_fix
--    of 2026-08-26 15:29 (the AHW2/CAHL2 blank-input edge) cover the
--    computation itself.
--
-- 2. claims_unverified (claims_llm_agritec.uk, 2026-08-24): the garbled
--    agreement-cap assertion. The copy was corrected by the content_rewrite of
--    2026-08-26 (complete 12:34) and VERIFIED AT THE SERVED PAGE 2026-08-30
--    and again 2026-08-31: "100,000 agreements" appears 0 times; the section
--    states the registered £100,000 per-agreement value cap
--    (CIT-86c4010f7cdf820d) correctly. The item cannot self-close: the
--    revalidator's 2026-08-30 pass returned arm 'spec_no_page_id' (the claims
--    audit writes no page_id into its spec, deliberately unparsed), verdict
--    'unknown' — so a manual close with the artefact evidence is the intended
--    path for this item shape.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE s.domain='agritec.uk' AND wi.status='needs_human_review'
     AND wi.item_key IN ('vision_finding:tool-sfi26-revenue-stacker:0a538b4a-803c-4f82-b298-d916f893fe8e',
                         'claims_llm_agritec.uk');
  IF n <> 2 THEN
    RAISE EXCEPTION 'pre-state: expected exactly the 2 items open, found %', n;
  END IF;
END $$;

UPDATE site_work_items wi
   SET status='complete', completed_at=now(), handled_by='agritek-session-2026-08-31',
       result = wi.result || jsonb_build_object(
         'ruling', 'as_designed',
         'reason', 'the em-dash is the pre-calculation empty state; the vision pass does not press Calculate. Pre-filling £0.00 would assert an uncomputed figure. Computation correctness is covered by the 8/9 acceptance run + the 08-26 audit_fix (AHW2/CAHL2 blank edge).',
         'ruled_by', 'agritek lane session, 2026-08-31')
 WHERE wi.item_key='vision_finding:tool-sfi26-revenue-stacker:0a538b4a-803c-4f82-b298-d916f893fe8e'
   AND wi.status='needs_human_review';

UPDATE site_work_items wi
   SET status='complete', completed_at=now(), handled_by='agritek-session-2026-08-31',
       result = wi.result || jsonb_build_object(
         'ruling', 'fixed_and_artefact_verified',
         'reason', 'copy corrected by content_rewrite:stacking-agricultural-scheme-actions:agreement-cap (complete 2026-08-26 12:34); served page verified 2026-08-30 and 2026-08-31: "100,000 agreements" x0, the £100,000 per-agreement value cap stated correctly per CIT-86c4010f7cdf820d. Manual close because the revalidator returns spec_no_page_id/unknown for this item shape (its spec carries no page_id).',
         'ruled_by', 'agritek lane session, 2026-08-31')
 WHERE wi.item_key='claims_llm_agritec.uk'
   AND wi.status='needs_human_review';

COMMIT;
