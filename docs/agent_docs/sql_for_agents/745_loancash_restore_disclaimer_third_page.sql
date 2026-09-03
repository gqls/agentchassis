-- 745_loancash_restore_disclaimer_third_page.sql
--
-- Files the THIRD restoration item foreseen by migration 743's own header.
--
-- 743 said: *"/guides/check-your-lender-is-authorised.html WAS STILL `claimed` WHEN THIS WAS
-- WRITTEN. It carries 739's defective spec, so expect the same wholesale rewrite there. Re-run
-- the served-bytes diff for that page afterwards and file a third restoration if it lost the
-- disclaimer too."* It completed at **14:43:43 UTC** and it did.
--
-- `[MEASURED 2026-09-03 14:53 UTC, served bytes]` the disclaimer phrase "does not lend money,
-- broker loans" now counts **0** on `the-payday-loan-price-cap`, `loan-sharks-and-illegal-lending`
-- and `check-your-lender-is-authorised`, and **1** on `jargon-buster` (whose rewrite ADDED one)
-- and on every page 739 did not touch. So three of the four rewritten pages lost it. 743 repairs
-- the first two; this repairs the third.
--
-- **739's correction on this page DID land** — the wrong reference "the affordability checks set
-- out under CONC 5A" is gone (served count 0). This is a restoration of collateral, not a revert,
-- and the acceptance test explicitly requires that the correction STAY gone.
--
-- Carries `spec.mode='edit_live'` — the control 739 lacked — plus the anti-fabrication constraint
-- bound to the 738 register, `source='manual'`, and `page_id`/`affected_url` derived from `pages`
-- by SELECT rather than typed. All four fixes the 739 council round asked for.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancash_couk_fca_validation/
-- Rollback: 745_..._ROLLBACK.sql

BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND domain = 'loancash.co.uk') THEN
    RAISE EXCEPTION '745 ABORT: site_id does not resolve to loancash.co.uk';
  END IF;
END $$;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND item_key = 'loancash_restore_disclaimer_check_your_lender'
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 0 THEN
    RAISE EXCEPTION '745 ABORT: the restoration item is already open';
  END IF;
END $$;

-- The disclaimer must still be site convention somewhere, or the premise is wrong.
DO $$
DECLARE nkeep int;
BEGIN
  SELECT count(DISTINCT pc.page_id) INTO nkeep
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
     AND pc.rendered_html LIKE '%does not lend money, broker loans%';
  IF nkeep = 0 THEN
    RAISE EXCEPTION '745 ABORT: the disclaimer appears on NO page of this site - the premise that it is site convention is wrong';
  END IF;
  RAISE NOTICE '745: disclaimer present on % page(s) - convention confirmed', nkeep;
END $$;

-- mode=edit_live must still be wired, or this fix is inert while reading as applied.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE default_config::text LIKE '%load_current_section_content%'
       AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  ) THEN
    RAISE EXCEPTION '745 ABORT: no live agent wires load_current_section_content - mode=edit_live would be inert';
  END IF;
END $$;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, page_id, affected_url,
   priority, handler_agent, status, created_by, item_key, approval_mode)
VALUES (
  'ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'manual', 'build', 'content_rewrite', 'high',
  'Restore the site-identity disclaimer block dropped by the 739 rewrite on check-your-lender-is-authorised (the third of three pages that lost it; 739''s CONC 5.2A correction landed and must stay)',
  '{"origin":"regression_repair","category":"content","page_name":"guide-check-your-lender-is-authorised","mode":"edit_live","mode_why":"REQUIRED. This page is the FOURTH and last of migration 739''s repair items, and it completed at 14:43:43 UTC with 739''s defective spec (spec.mode UNSET). load_current_section_content_action.go (bugs_open/178) attaches a section''s current rendered_html only when spec.mode=''edit_live''; without it page-content-writer gets the guidance text and nothing to work from and fabricates a full replacement section. Migration 743''s header PREDICTED this page would lose its disclaimer for the same reason and said to file a third restoration if it did. It did. This item must EDIT, not regenerate.","current_value":"The served page is the post-739 rewrite. The regulatory correction 739 commissioned LANDED - the wrong reference ''the affordability checks set out under CONC 5A'' is gone (served count 0) - but the wholesale regeneration dropped the site-identity disclaimer block, which the page carried before (served count 1 -> 0).","suggestion":"ADD BACK the site-identity disclaimer block, in the site''s existing voice and position, and change nothing else on the page: \"LoanCash.co.uk does not lend money, broker loans, or take applications, and never will. Nothing here is financial or legal advice. We are not the Financial Conduct Authority and are not affiliated with it.\" It appeared on 14 of this site''s 30 pages before the 739 rewrites and still appears on the untouched ones (e.g. /guides/if-you-cant-pay.html), so match those pages'' placement and wording rather than inventing a new form. Do NOT re-add the wrong ''CONC 5A'' affordability reference that 739 correctly removed - affordability is CONC 5.2A (registered fact FCA-CONC-5-2A-4, migration 738).","constraint":"SMALLEST TRUE CHANGE, and this item is a RESTORATION - the risk here is a second rewrite, not an insufficient one. Do NOT rewrite, reorder, condense or ''improve'' any prose already on the page. Do NOT introduce any new figure, statistic, date, example or named source that is not already an evidenced fact in this site''s evidence_base register (migration 738, 19 facts) - that register is the CLOSED SET of figures this site may assert. Add the named material; leave everything else byte-identical.","affected_component":"Guide body - the closing site-identity block","acceptance_test":"Fetch https://loancash.co.uk/guides/check-your-lender-is-authorised.html and confirm the served bytes contain \"does not lend money, broker loans\" (count must be >= 1; it is currently 0) AND still contain NO occurrence of \"under CONC 5A\" (currently 0 - 739''s correction must not be undone) AND that visible text length has GROWN rather than shrunk. Then diff the visible text against the current served version: additions only, no sentence removed or reworded. Use an invented-URL 404 as a fetch control.","authority":"Regression caused by migration 739 under OWNER RULING D2 (RFC_060 §3g), same mechanism as the two items migration 743 files. 739''s council round (APPROVED, corr 93897fb5) named the cause; the verdict was read after the items had run. WRONG_CALLS carries the lesson.","beware":"A page_rerender re-ships stored content_data and completes successfully - it cannot restore anything. And no verifier is registered for content_rewrite (complete_work_item_acceptance_predicate.go), so a ''complete'' status is not evidence the bytes changed. Verify at the served bytes.","filed_by_lane":"loancash_couk_fca_validation lane (migration 745)"}'::jsonb,
  (SELECT id  FROM pages WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url = '/guides/check-your-lender-is-authorised.html'),
  (SELECT url FROM pages WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url = '/guides/check-your-lender-is-authorised.html'),
  15, 'page-build-handler', 'triaged',
  'loancash_couk_fca_validation lane (migration 745)',
  'loancash_restore_disclaimer_check_your_lender', 'auto'
);

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items w JOIN pages p ON p.id = w.page_id
   WHERE w.created_by = 'loancash_couk_fca_validation lane (migration 745)'
     AND w.spec->>'mode' = 'edit_live'
     AND w.spec->>'constraint' LIKE '%SMALLEST TRUE CHANGE%'
     AND w.source = 'manual'
     AND w.affected_url = p.url
     AND p.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74';
  IF n <> 1 THEN
    RAISE EXCEPTION '745 VERIFY: expected 1 item with mode=edit_live, the constraint, source=manual and affected_url matching pages.url, found %', n;
  END IF;
  RAISE NOTICE '745 OK: third restoration item filed WITH mode=edit_live';
END $$;

COMMIT;
