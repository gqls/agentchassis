-- 743_loancash_restore_content_dropped_by_the_739_rewrite.sql
--
-- Restores content that migration 739's repair DROPPED. 739 corrected three wrong regulatory
-- claims (owner ruling D2) and the corrections landed accurately — but its items left
-- `spec.mode` UNSET, so page-build-handler REGENERATED each section instead of editing it.
--
-- ⚠ THE COUNCIL SAID THIS WOULD HAPPEN, AND I READ THE VERDICT TOO LATE.
-- 739's verdict was APPROVED (corr `93897fb5-0b73-4b1e-b4aa-c0d2b9d4a87b`) with a medium
-- objection that named the defect exactly: *"page-build-handler's content writer only sees the
-- page's own stored prose when spec.mode is set … If mode isn't set, the writer may regenerate
-- the section without reference to the exact current_value quoted, undermining the whole 'exact
-- required correction' design even though the spec text looks precise."* By the time I read it,
-- `build-dispatch-loop` had already claimed all four items and completed three. The correcting
-- migration I had written could not run: its own guard refused, correctly, because rewriting a
-- spec underneath a running handler is worse than the defect. **Applying before reading the
-- verdict is what cost this.** Logged in `WRONG_CALLS.md`.
--
-- WHAT ACTUALLY HAPPENED, measured at the served bytes against a pre-repair crawl:
--
--   | page | bytes retained | sentences replaced | wrong claim gone? |
--   |---|---|---|---|
--   | the-payday-loan-price-cap    | 84% | **36 of 37** | yes |
--   | jargon-buster                | 88% | **49 of 50** | yes |
--   | loan-sharks-and-illegal-lending | 66% | 47 of ~48 | yes |
--   | check-your-lender-is-authorised | (in flight) | — | not yet |
--
-- ⚠ **BYTE RETENTION HID A TOTAL REWRITE.** 84% and 88% read as mild edits; the sentence-level
-- diff shows near-complete replacement at a similar length. **A length-retention check cannot
-- detect a rewrite** — only an identity diff can. That is the transferable half of this.
--
-- AND THE REWRITES ARE GOOD, which is why this is a narrow restoration and not a revert. The
-- new jargon-buster states the correction better than the brief asked: *"CONC 5A.2.14R(1) makes
-- this a cumulative limit, not a per-instance one: £15 is the most a lender can charge in default
-- fees across the whole agreement, however many payments you miss."* The CPA correction is right
-- on loan-sharks too. **Do not revert 739.** What went is selective:
--
--   1. **THE SITE-IDENTITY DISCLAIMER, dropped from TWO pages** — "LoanCash.co.uk does not lend
--      money, broker loans, or take applications…", plus the not-advice and not-the-FCA lines.
--      `[MEASURED 2026-09-03]` it was on **14 of 30** pages before; it is now **0** on the
--      price-cap and loan-sharks pages, and still **1** on untouched pages such as
--      `/guides/if-you-cant-pay.html`. On a finance information site this is a compliance
--      element, and losing it silently is the worst of the collateral.
--   2. **Three substantive losses on loan-sharks** — the FSMA-2000 criminal-offence framing
--      (which is a REGISTERED fact, ids FSMA-2000-S19/S23), the prohibition on taking a card or
--      passport as security, and the Illegal Money Lending Team's anonymity/free/three-nations
--      detail. Those are the sentences that change what a frightened reader does.
--
-- WHAT THIS MIGRATION DOES: files TWO `content_rewrite` items, **both carrying
-- `spec.mode='edit_live'`** so the writer edits the current page instead of regenerating it, and
-- both carrying the anti-fabrication constraint bound to the 738 register. Also adopts the other
-- three council fixes: `source='manual'` (truthful provenance — a hand-filed item must not wear
-- a pipeline-shaped name), `affected_url` and `page_id` **derived from `pages` by SELECT** rather
-- than typed, and `priority=15` so they precede ordinary build work.
--
-- ⚠ `/guides/check-your-lender-is-authorised.html` WAS STILL `claimed` WHEN THIS WAS WRITTEN.
-- It carries 739's defective spec, so expect the same wholesale rewrite there. **Re-run the
-- served-bytes diff for that page afterwards and file a third restoration if it lost the
-- disclaimer too** — the query is in the lane RUNBOOK and in HANDOFF §8.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancash_couk_fca_validation/
-- Rollback: 743_..._ROLLBACK.sql

BEGIN;

-- GUARD 1: the site_id must resolve to the domain (RUNBOOK_lendzy §8e).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND domain = 'loancash.co.uk') THEN
    RAISE EXCEPTION '743 ABORT: site_id does not resolve to loancash.co.uk';
  END IF;
END $$;

-- GUARD 2: don't double-file.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
     AND item_key IN ('loancash_restore_disclaimer_price_cap',
                      'loancash_restore_disclaimer_and_fsma_loan_sharks')
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 0 THEN
    RAISE EXCEPTION '743 ABORT: % restoration item(s) already open', n;
  END IF;
END $$;

-- GUARD 3: the disclaimer must ACTUALLY be missing. If a later pass already restored it, this
-- migration would commission a no-op rewrite — and a needless rewrite is how content is lost.
-- Verified from the pages this migration does NOT touch: the phrase must still exist somewhere
-- on the site, or the premise (that it is this site's convention) is wrong.
DO $$
DECLARE nkeep int;
BEGIN
  -- NOTE: rendered html lives on page_components, NOT on pages (pages has rendered_header /
  -- rendered_footer / rendered_head only). The first cut of this guard queried a
  -- pages.rendered_html that does not exist and failed loudly — schema first, always.
  SELECT count(DISTINCT pc.page_id) INTO nkeep
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
     AND pc.rendered_html LIKE '%does not lend money, broker loans%';
  IF nkeep = 0 THEN
    RAISE EXCEPTION '743 ABORT: the disclaimer phrase appears on NO page of this site - the premise that it is site convention is wrong, or the column is not where it lives; check at the SERVED bytes before filing';
  END IF;
  RAISE NOTICE '743: disclaimer still present on % page(s) - convention confirmed, proceeding', nkeep;
END $$;

-- GUARD 4: mode=edit_live must still be wired, or the whole point of this migration is inert.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE default_config::text LIKE '%load_current_section_content%'
       AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  ) THEN
    RAISE EXCEPTION '743 ABORT: no live agent wires load_current_section_content - mode=edit_live would be inert, and an inert fix that reads as applied is worse than none';
  END IF;
END $$;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, page_id, affected_url,
   priority, handler_agent, status, created_by, item_key, approval_mode)
VALUES
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'manual', 'build', 'content_rewrite', 'high',
   'Restore the site-identity disclaimer block dropped by the 739 rewrite (regenerated instead of edited - spec.mode was unset)',
   '{"origin":"regression_repair","category":"content","page_name":"guide-the-payday-loan-price-cap","mode":"edit_live","mode_why":"REQUIRED. Migration 739''s items left spec.mode UNSET, and that is why this repair is needed: load_current_section_content_action.go (bugs_open/178) only attaches a section''s current rendered_html when spec.mode=''edit_live''. Without it page-content-writer gets the guidance text and NOTHING to work from and fabricates a full replacement section - measured elsewhere at 4439 -> 1806 chars, and measured HERE at 36 of 37 and 49 of 50 sentences replaced on two pages. This item must EDIT what is on the page now, adding back what is named below and changing nothing else.","current_value":"The served page is the post-739 rewrite. It is substantively CORRECT - the regulatory correction 739 commissioned landed accurately - but the wholesale regeneration dropped the material named in `suggestion`.","suggestion":"ADD BACK the following, in the site''s existing voice and position, and change nothing else on the page. The disclaimer block, verbatim in substance: \"LoanCash.co.uk does not lend money, broker loans, or take applications - and never will. This site publishes plain-English explanations of the FCA rules that govern high-cost credit, and points readers toward free tools and free, independent complaint routes. Nothing here is financial or legal advice. We are not the Financial Conduct Authority and are not affiliated with it.\" It appeared on 14 of this site''s 30 pages before the rewrite and still appears on the untouched ones (e.g. /guides/if-you-cant-pay.html), so match those pages'' placement and wording rather than inventing a new form. ","constraint":"SMALLEST TRUE CHANGE, and this item is a RESTORATION - the risk here is a second rewrite, not an insufficient one. Do NOT rewrite, reorder, condense or ''improve'' any prose that is already on the page. Do NOT introduce any new figure, statistic, date, example or named source that is not already an evidenced fact in this site''s evidence_base register (migration 738, 19 facts) - that register is the CLOSED SET of figures this site may assert. Add the named material; leave everything else byte-identical.","affected_component":"Guide body - the closing site-identity block (and, on loan-sharks, three substantive sections)","acceptance_test":"Fetch https://loancash.co.uk/guides/the-payday-loan-price-cap.html and confirm the served bytes contain the phrase \"does not lend money, broker loans\" (count must be >= 1; it is currently 0) AND that the page''s visible text length has GROWN relative to its current value rather than shrunk. Then diff the visible text against the current served version and confirm no existing sentence was removed or reworded - only additions. Use an invented-URL 404 as a fetch control.","authority":"Regression caused by migration 739 under OWNER RULING D2 (RFC_060 §3g). The council round on 739 (APPROVED, corr 93897fb5) objected on exactly this ground - ''if mode isn''t set, the writer may regenerate the section without reference to the exact current_value quoted'' - and the objection was read AFTER the items had already been claimed and run. WRONG_CALLS carries the lesson.","beware":"A page_rerender re-ships stored content_data and completes successfully - it cannot restore anything and must not be read as this repair. And no verifier is registered for content_rewrite (complete_work_item_acceptance_predicate.go), so a ''complete'' status is not evidence the bytes changed. Verify at the served bytes with the acceptance test above.","filed_by_lane":"loancash_couk_fca_validation lane (migration 743)"}'::jsonb,
   (SELECT id FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/the-payday-loan-price-cap.html'),
   (SELECT url FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/the-payday-loan-price-cap.html'),
   15, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 743)', 'loancash_restore_disclaimer_price_cap', 'auto'),
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'manual', 'build', 'content_rewrite', 'high',
   'Restore the disclaimer block AND three substantive losses from the 739 rewrite: the FSMA criminal-offence framing, the security-taking prohibition, and the Illegal Money Lending Team detail',
   '{"origin":"regression_repair","category":"content","page_name":"guide-loan-sharks-and-illegal-lending","mode":"edit_live","mode_why":"REQUIRED. Migration 739''s items left spec.mode UNSET, and that is why this repair is needed: load_current_section_content_action.go (bugs_open/178) only attaches a section''s current rendered_html when spec.mode=''edit_live''. Without it page-content-writer gets the guidance text and NOTHING to work from and fabricates a full replacement section - measured elsewhere at 4439 -> 1806 chars, and measured HERE at 36 of 37 and 49 of 50 sentences replaced on two pages. This item must EDIT what is on the page now, adding back what is named below and changing nothing else.","current_value":"The served page is the post-739 rewrite. It is substantively CORRECT - the regulatory correction 739 commissioned landed accurately - but the wholesale regeneration dropped the material named in `suggestion`.","suggestion":"ADD BACK the following, in the site''s existing voice and position, and change nothing else on the page. The disclaimer block, verbatim in substance: \"LoanCash.co.uk does not lend money, broker loans, or take applications, and never will. Nothing here is financial or legal advice. We are not the Financial Conduct Authority and are not affiliated with it.\" It appeared on 14 of this site''s 30 pages before the rewrite and still appears on the untouched ones (e.g. /guides/if-you-cant-pay.html), so match those pages'' placement and wording rather than inventing a new form. ALSO RESTORE these three, which the rewrite dropped and which are the page''s most load-bearing content: (a) THE CRIMINAL-OFFENCE FRAMING - that lending money for profit without FCA authorisation is a criminal offence under the Financial Services and Markets Act 2000, not merely a breach of good practice. This is a REGISTERED FACT in this site''s own evidence_base (ids FSMA-2000-S19 and FSMA-2000-S23, migration 738), so it is evidenced and may be stated plainly. It was the page''s opening argument and its removal weakens the whole piece. (b) THE SECURITY PROHIBITION - that an illegal lender cannot lawfully take a bank card, passport, benefit book or driving licence as security, and that threatening a borrower or their family is a criminal offence regardless of whether money changed hands. (c) THE REPORTING DETAIL - that the Illegal Money Lending Team operates in England, Wales and Scotland, that reports can be made ANONYMOUSLY, and that the service is FREE. The rewrite kept a general ''can be reported'' sentence and dropped all three of those specifics, which are the ones that change what a frightened reader will actually do.","constraint":"SMALLEST TRUE CHANGE, and this item is a RESTORATION - the risk here is a second rewrite, not an insufficient one. Do NOT rewrite, reorder, condense or ''improve'' any prose that is already on the page. Do NOT introduce any new figure, statistic, date, example or named source that is not already an evidenced fact in this site''s evidence_base register (migration 738, 19 facts) - that register is the CLOSED SET of figures this site may assert. Add the named material; leave everything else byte-identical.","affected_component":"Guide body - the closing site-identity block (and, on loan-sharks, three substantive sections)","acceptance_test":"Fetch https://loancash.co.uk/guides/loan-sharks-and-illegal-lending.html and confirm the served bytes contain the phrase \"does not lend money, broker loans\" (count must be >= 1; it is currently 0) AND that the page''s visible text length has GROWN relative to its current value rather than shrunk. Then diff the visible text against the current served version and confirm no existing sentence was removed or reworded - only additions. Use an invented-URL 404 as a fetch control.","authority":"Regression caused by migration 739 under OWNER RULING D2 (RFC_060 §3g). The council round on 739 (APPROVED, corr 93897fb5) objected on exactly this ground - ''if mode isn''t set, the writer may regenerate the section without reference to the exact current_value quoted'' - and the objection was read AFTER the items had already been claimed and run. WRONG_CALLS carries the lesson.","beware":"A page_rerender re-ships stored content_data and completes successfully - it cannot restore anything and must not be read as this repair. And no verifier is registered for content_rewrite (complete_work_item_acceptance_predicate.go), so a ''complete'' status is not evidence the bytes changed. Verify at the served bytes with the acceptance test above.","filed_by_lane":"loancash_couk_fca_validation lane (migration 743)"}'::jsonb,
   (SELECT id FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/loan-sharks-and-illegal-lending.html'),
   (SELECT url FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/loan-sharks-and-illegal-lending.html'),
   15, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 743)', 'loancash_restore_disclaimer_and_fsma_loan_sharks', 'auto');

-- VERIFY as DO/RAISE.
DO $$
DECLARE n int; nmode int; nurl int; nsrc int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)';
  IF n <> 2 THEN
    RAISE EXCEPTION '743 VERIFY: expected 2 restoration items, found %', n;
  END IF;

  SELECT count(*) INTO nmode FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)'
     AND spec->>'mode' = 'edit_live'
     AND spec->>'constraint' LIKE '%SMALLEST TRUE CHANGE%';
  IF nmode <> 2 THEN
    RAISE EXCEPTION '743 VERIFY: expected 2 items with mode=edit_live AND the anti-fabrication constraint, found %', nmode;
  END IF;

  -- the drift check the council asked for: affected_url must EQUAL pages.url, and page_id resolve
  SELECT count(*) INTO nurl FROM site_work_items w JOIN pages p ON p.id = w.page_id
   WHERE w.created_by = 'loancash_couk_fca_validation lane (migration 743)'
     AND w.affected_url = p.url AND p.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74';
  IF nurl <> 2 THEN
    RAISE EXCEPTION '743 VERIFY: expected both affected_url to equal their page''s pages.url on this site, found %', nurl;
  END IF;

  SELECT count(*) INTO nsrc FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)' AND source = 'manual';
  IF nsrc <> 2 THEN
    RAISE EXCEPTION '743 VERIFY: expected source=manual on both (truthful provenance), found %', nsrc;
  END IF;

  RAISE NOTICE '743 OK: 2 restoration items filed WITH mode=edit_live - the disclaimer on two pages, plus the FSMA framing, the security prohibition and the IMLT detail on loan-sharks';
END $$;

COMMIT;
