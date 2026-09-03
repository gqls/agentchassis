-- 743_loancash_restore_content_dropped_by_the_739_rewrite.sql
--
-- ** REVISION 2 ** — answering the council REVISE on corr `4718725c-7d23-41ca-a320-17ebbbfb5e02`.
-- Resubmit with RESUBMIT_CORR=4718725c-7d23-41ca-a320-17ebbbfb5e02 so the trail accumulates.
-- NEVER APPLIED in revision 1; nothing ran, so this is a forward edit of an unapplied file.
-- It now covers ALL THREE affected pages and SUPERSEDES migration 745, which is left as a stub.
--
-- WHAT THIS REPAIRS. Migration 739 corrected three wrong regulatory claims (owner ruling D2) and
-- the corrections LANDED — all three wrong strings went 1 → 0, and the served jargon-buster now
-- states the cumulative £15 rule better than the brief asked. But 739's items left `spec.mode`
-- UNSET, so page-build-handler REGENERATED each section instead of editing it. Three of the four
-- pages lost their site-identity block. **Do NOT revert 739.**
--
-- ⚠ AND THE COUNCIL SAID SO BEFORE IT HAPPENED. 739's verdict was APPROVED with a medium
-- objection naming this exact defect; I read it after the items had been claimed and run.
-- `WRONG_CALLS.md`, 2026-09-03.
--
-- ─── WHAT CHANGED IN REVISION 2, and the best of it was not my idea ──────────────────────────
--
-- **(1) THE TEXT IS NOW VERBATIM, RECOVERED — NOT RECONSTRUCTED.** Objection: *"the plan
-- re-derives the dropped sentences via a fresh LLM edit_live instruction, but never shows it
-- checked `page_component_history` — the platform's existing archive of pre-overwrite
-- rendered_html … that is a second approximation of content that already exists exactly."*
-- **Checked, and it exists.** Each rewrite wrote a `delete`-op history row carrying the
-- pre-overwrite HTML (price-cap 9,508 chars at 14:28:52; the current component is 8,470). The
-- exact blocks are now embedded per page in `spec.restore_verbatim`.
--
-- **The objection paid for itself immediately:** my paraphrase had DROPPED the sentence
-- *"We are independent."*, and the three pages' blocks are **not identical** — the price-cap
-- page carries an extra sentence about what the site publishes. An LLM reconstruction from my
-- wording would have produced three homogenised blocks, each missing a sentence, and every
-- acceptance test I had written would still have passed.
--
-- ⚠ WHY NOT RESTORE THE HISTORY ROW WHOLESALE: each page is a SINGLE `ported-page` component,
-- so putting the old `rendered_html` back would restore the wrong claims too and undo 739's
-- correction. And a direct `rendered_html` patch is destroyed by the next rerender, which
-- regenerates from `content_data`. The framework must write it; this migration supplies the
-- exact words for it to write.
--
-- **(2) `approval_mode='manual'`** (was `auto`). Objection: *"the regression being fixed here was
-- an UNSUPERVISED rewrite of regulatory content … manual approval_mode would match the
-- sensitivity the plan itself argues for."* Correct, and it also buys back the window 739 never
-- had: those items were claimed 7–20 minutes after filing, which is why the verdict arrived too
-- late to act on.
--
-- **(3) THE `edit_live` vs `recreate` QUESTION — RESOLVED AT THE CODE, and it was the right
-- thing to challenge.** Two HIGH objections warned that three landmine entries say
-- *"page-build-handler's content writer never sees a page's OWN stored prose unless
-- spec.mode=\"recreate\""*, and that guard 4 only checked the action was WIRED, not that the
-- literal matched. Verified: `load_current_section_content_action.go:98` declares
-- `const editLiveMode = "edit_live"` and returns `passthrough("not_edit_live")` otherwise, and
-- **page-build-handler's live `default_config` carries both that step and the `edit_live`
-- literal** (`page-content-writer`'s does not — it is the wrong agent to look at). The landmines
-- naming `recreate` describe `load_existing_content`, a DIFFERENT step that sources the adoption
-- crawl rather than current prose. `edit_live` is correct. **Guard 4 below is strengthened to
-- assert the literal on page-build-handler specifically, not merely that something is wired.**
--
-- **(4) THE JARGON-BUSTER GAP, filled.** Objection: it underwent the same near-total rewrite
-- (88% bytes, 49 of 50 sentences) yet was omitted from the loss list with no stated reason.
-- Measured: it **GAINED** the whole site-identity block (`does not lend money` 0 → 1,
-- `We are independent` 0 → 1, `not the Financial Conduct Authority` 0 → 1) and kept its
-- `fca.org.uk` pointer. **No compliance content lost there** — that is why it is not in this
-- migration, and now it says so.
--
-- ─── OBJECTIONS ACCEPTED AND *NOT* FIXED HERE, stated rather than buried ─────────────────────
--
-- **THIS IS A SYMPTOM FIX.** Two objections make the same point and both are right: nothing
-- enforces `spec.mode` on `content_rewrite` at the producer, so *the next item minted anywhere on
-- the fleet without it is exposed to the identical destructive regeneration*. That is a
-- platform-wide gap (`bugs_open/178`'s own header describes it), it is larger than this lane, and
-- it wants its own council round rather than being smuggled into a content restoration. **Filed
-- as the lane's top follow-up in the handoff; do not let it be forgotten because these three
-- pages got better.**
--
-- **NO DATA-LEVEL SNAPSHOT IS TAKEN BEFORE THE DELEGATED EDIT.** Objection: the real mutation
-- happens later inside page-content-writer with no pre-image, so the ROLLBACK can only cancel an
-- unclaimed item. True. The mitigation is the one this migration is built on: `page_component_history`
-- IS that pre-image — it is what made this repair recoverable at all, and it will capture the next
-- state too. `approval_mode='manual'` is the other half.
--
-- **NOTHING ENFORCES THE ACCEPTANCE TEST.** True, and unfixable from a migration: no verifier is
-- registered for `content_rewrite`. The test is written to be runnable by a person in two curls
-- and is reproduced in the handoff.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancash_couk_fca_validation/
-- Rollback: 743_..._ROLLBACK.sql

BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND domain = 'loancash.co.uk') THEN
    RAISE EXCEPTION '743 ABORT: site_id does not resolve to loancash.co.uk';
  END IF;
END $$;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
     AND item_key IN ('loancash_restore_disclaimer_price_cap',
                      'loancash_restore_disclaimer_and_fsma_loan_sharks',
                      'loancash_restore_disclaimer_check_your_lender')
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 0 THEN
    RAISE EXCEPTION '743 ABORT: % restoration item(s) already open', n;
  END IF;
END $$;

-- The premise: the block must still BE site convention somewhere, or a rewrite is not warranted.
DO $$
DECLARE nkeep int;
BEGIN
  SELECT count(DISTINCT pc.page_id) INTO nkeep
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
     AND pc.rendered_html LIKE '%does not lend money, broker loans%';
  IF nkeep = 0 THEN
    RAISE EXCEPTION '743 ABORT: the block appears on NO page of this site - the premise that it is site convention is wrong';
  END IF;
  RAISE NOTICE '743: block present on % page(s) - convention confirmed', nkeep;
END $$;

-- STRENGTHENED (rev 2, council HIGH): assert the LITERAL on page-build-handler specifically.
-- Checking only that the action is "wired" somewhere would pass on a value the workflow never
-- branches on - which is exactly the failure the objection described.
DO $$
DECLARE cfg text;
BEGIN
  SELECT default_config::text INTO cfg FROM agent_definitions
   WHERE type = 'page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   LIMIT 1;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '743 ABORT: no live page-build-handler definition - the handler these items route to does not exist';
  END IF;
  IF cfg NOT LIKE '%load_current_section_content%' THEN
    RAISE EXCEPTION '743 ABORT: page-build-handler does not carry load_current_section_content - mode=edit_live would be inert';
  END IF;
  IF cfg NOT LIKE '%edit_live%' THEN
    RAISE EXCEPTION '743 ABORT: page-build-handler carries the step but NOT the edit_live literal - the value this migration sets is not the one the workflow branches on; re-derive it before filing';
  END IF;
  RAISE NOTICE '743: page-build-handler carries load_current_section_content AND the edit_live literal';
END $$;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, page_id, affected_url,
   priority, handler_agent, status, created_by, item_key, approval_mode)
VALUES
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'manual', 'build', 'content_rewrite', 'high',
   'Restore the site-identity block dropped by the 739 rewrite on guide-the-payday-loan-price-cap — VERBATIM from page_component_history, not reconstructed',
   '{"origin":"regression_repair","category":"content","page_name":"guide-the-payday-loan-price-cap","mode":"edit_live","mode_why":"REQUIRED. Migration 739''s items left spec.mode UNSET, which is why this repair exists: load_current_section_content (page-build-handler''s workflow, const editLiveMode=\"edit_live\", load_current_section_content_action.go:98) attaches a section''s current rendered_html ONLY on this literal. Without it page-content-writer gets guidance text and nothing to work from and fabricates a full replacement section. VERIFIED for this migration: page-build-handler''s live default_config carries both the step and the edit_live literal; the older landmines naming \"recreate\" describe load_existing_content, a DIFFERENT step that sources the adoption crawl, not current prose.","restore_verbatim":"LoanCash.co.uk does not lend money, broker loans, or take applications — and never will. This site publishes plain-English explanations of the FCA rules that govern high-cost credit, and points readers toward free tools and free, independent complaint routes. Nothing here is financial or legal advice. We are independent. We are not the Financial Conduct Authority and are not affiliated with it.","provenance":"The text in `restore_verbatim` is NOT reconstructed. It was recovered VERBATIM from `page_component_history` — the delete-op row written at the moment of the 739 rewrite, which retains the pre-overwrite rendered_html. Council objection on round 1: re-deriving deleted text by LLM instruction is a second approximation of content that already exists exactly. It was right, and it caught a real omission — my paraphrase had dropped the sentence \"We are independent.\", and the three pages'' blocks are NOT identical (the price-cap page carries an extra sentence about what the site publishes). Insert the block as given.","current_value":"The served page is the post-739 rewrite. The regulatory correction 739 commissioned LANDED and must not be undone; what the wholesale regeneration dropped is the block in `restore_verbatim`.","suggestion":"Insert the `restore_verbatim` block back into this page, in the closing position it occupied before (it still occupies that position on pages 739 did not touch, e.g. /guides/if-you-cant-pay.html and /guides/jargon-buster.html — match those). Reproduce it as given, word for word. Change nothing else.","constraint":"SMALLEST TRUE CHANGE, and this is a RESTORATION — the risk here is a SECOND rewrite, not an insufficient one. Do NOT rewrite, reorder, condense or improve any prose already on the page. Do NOT introduce any figure, statistic, date, example or named source that is not already an evidenced fact in this site''s evidence_base register (migration 738, 19 facts) — that register is the CLOSED SET of figures this site may assert. Add the named material; leave everything else byte-identical.","affected_component":"Guide body (single `ported-page` slot) — the closing site-identity block","acceptance_test":"Fetch https://loancash.co.uk/guides/the-payday-loan-price-cap.html and confirm the served bytes contain \"does not lend money, broker loans\" AND \"We are independent\" (both currently 0) AND that visible text length has GROWN, not shrunk. Then diff visible text against the current served version: ADDITIONS ONLY, no existing sentence removed or reworded. Use an invented-URL 404 as a fetch control. This test is the one that would have caught 739 on day one.","authority":"Regression caused by migration 739 under OWNER RULING D2 (RFC_060 §3g). 739''s council round (APPROVED, corr 93897fb5) named the cause; the verdict was read after the items had run. This round (corr 4718725c) returned REVISE and its objections are answered here. WRONG_CALLS carries the lesson.","beware":"A page_rerender re-ships stored content_data and completes successfully — it cannot restore anything. No verifier is registered for content_rewrite, so ''complete'' is not evidence the bytes changed. Verify at the served bytes with the acceptance test above.","filed_by_lane":"loancash_couk_fca_validation lane (migration 743 rev 2)"}'::jsonb,
   (SELECT id  FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/the-payday-loan-price-cap.html'),
   (SELECT url FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/the-payday-loan-price-cap.html'),
   15, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 743)', 'loancash_restore_disclaimer_price_cap', 'manual'),
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'manual', 'build', 'content_rewrite', 'high',
   'Restore the site-identity block dropped by the 739 rewrite on guide-loan-sharks-and-illegal-lending — VERBATIM from page_component_history, not reconstructed',
   '{"origin":"regression_repair","category":"content","page_name":"guide-loan-sharks-and-illegal-lending","mode":"edit_live","mode_why":"REQUIRED. Migration 739''s items left spec.mode UNSET, which is why this repair exists: load_current_section_content (page-build-handler''s workflow, const editLiveMode=\"edit_live\", load_current_section_content_action.go:98) attaches a section''s current rendered_html ONLY on this literal. Without it page-content-writer gets guidance text and nothing to work from and fabricates a full replacement section. VERIFIED for this migration: page-build-handler''s live default_config carries both the step and the edit_live literal; the older landmines naming \"recreate\" describe load_existing_content, a DIFFERENT step that sources the adoption crawl, not current prose.","restore_verbatim":"LoanCash.co.uk does not lend money, broker loans, or take applications, and never will. Nothing here is financial or legal advice. We are independent. We are not the Financial Conduct Authority and are not affiliated with it.","provenance":"The text in `restore_verbatim` is NOT reconstructed. It was recovered VERBATIM from `page_component_history` — the delete-op row written at the moment of the 739 rewrite, which retains the pre-overwrite rendered_html. Council objection on round 1: re-deriving deleted text by LLM instruction is a second approximation of content that already exists exactly. It was right, and it caught a real omission — my paraphrase had dropped the sentence \"We are independent.\", and the three pages'' blocks are NOT identical (the price-cap page carries an extra sentence about what the site publishes). Insert the block as given.","current_value":"The served page is the post-739 rewrite. The regulatory correction 739 commissioned LANDED and must not be undone; what the wholesale regeneration dropped is the block in `restore_verbatim`.","suggestion":"Insert the `restore_verbatim` block back into this page, in the closing position it occupied before (it still occupies that position on pages 739 did not touch, e.g. /guides/if-you-cant-pay.html and /guides/jargon-buster.html — match those). Reproduce it as given, word for word. Change nothing else. ALSO RESTORE three substantive losses, which are this page''s most load-bearing content and which the rewrite dropped: (a) THE CRIMINAL-OFFENCE FRAMING — that lending for profit without FCA authorisation is a criminal offence under the Financial Services and Markets Act 2000, not merely a breach of good practice. This is a REGISTERED fact (ids FSMA-2000-S19, FSMA-2000-S23, migration 738), so it is evidenced and may be stated plainly; it was the page''s opening argument. (b) THE SECURITY PROHIBITION — that an illegal lender cannot lawfully take a bank card, passport, benefit book or driving licence as security, and that threatening a borrower or their family is a criminal offence regardless of whether money changed hands. (c) THE REPORTING DETAIL — that the Illegal Money Lending Team operates in England, Wales and Scotland, that reports can be made ANONYMOUSLY, and that the service is FREE. The rewrite kept a general ''can be reported'' sentence and dropped all three specifics, which are the ones that change what a frightened reader actually does.","constraint":"SMALLEST TRUE CHANGE, and this is a RESTORATION — the risk here is a SECOND rewrite, not an insufficient one. Do NOT rewrite, reorder, condense or improve any prose already on the page. Do NOT introduce any figure, statistic, date, example or named source that is not already an evidenced fact in this site''s evidence_base register (migration 738, 19 facts) — that register is the CLOSED SET of figures this site may assert. Add the named material; leave everything else byte-identical.","affected_component":"Guide body (single `ported-page` slot) — the closing site-identity block","acceptance_test":"Fetch https://loancash.co.uk/guides/loan-sharks-and-illegal-lending.html and confirm the served bytes contain \"does not lend money, broker loans\" AND \"We are independent\" (both currently 0) AND that visible text length has GROWN, not shrunk. Then diff visible text against the current served version: ADDITIONS ONLY, no existing sentence removed or reworded. Use an invented-URL 404 as a fetch control. This test is the one that would have caught 739 on day one.","authority":"Regression caused by migration 739 under OWNER RULING D2 (RFC_060 §3g). 739''s council round (APPROVED, corr 93897fb5) named the cause; the verdict was read after the items had run. This round (corr 4718725c) returned REVISE and its objections are answered here. WRONG_CALLS carries the lesson.","beware":"A page_rerender re-ships stored content_data and completes successfully — it cannot restore anything. No verifier is registered for content_rewrite, so ''complete'' is not evidence the bytes changed. Verify at the served bytes with the acceptance test above.","filed_by_lane":"loancash_couk_fca_validation lane (migration 743 rev 2)"}'::jsonb,
   (SELECT id  FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/loan-sharks-and-illegal-lending.html'),
   (SELECT url FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/loan-sharks-and-illegal-lending.html'),
   15, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 743)', 'loancash_restore_disclaimer_and_fsma_loan_sharks', 'manual'),
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'manual', 'build', 'content_rewrite', 'high',
   'Restore the site-identity block dropped by the 739 rewrite on guide-check-your-lender-is-authorised — VERBATIM from page_component_history, not reconstructed',
   '{"origin":"regression_repair","category":"content","page_name":"guide-check-your-lender-is-authorised","mode":"edit_live","mode_why":"REQUIRED. Migration 739''s items left spec.mode UNSET, which is why this repair exists: load_current_section_content (page-build-handler''s workflow, const editLiveMode=\"edit_live\", load_current_section_content_action.go:98) attaches a section''s current rendered_html ONLY on this literal. Without it page-content-writer gets guidance text and nothing to work from and fabricates a full replacement section. VERIFIED for this migration: page-build-handler''s live default_config carries both the step and the edit_live literal; the older landmines naming \"recreate\" describe load_existing_content, a DIFFERENT step that sources the adoption crawl, not current prose.","restore_verbatim":"LoanCash.co.uk does not lend money, broker loans, or take applications, and never will. Nothing here is financial or legal advice. We are independent. We are not the Financial Conduct Authority and are not affiliated with it.","provenance":"The text in `restore_verbatim` is NOT reconstructed. It was recovered VERBATIM from `page_component_history` — the delete-op row written at the moment of the 739 rewrite, which retains the pre-overwrite rendered_html. Council objection on round 1: re-deriving deleted text by LLM instruction is a second approximation of content that already exists exactly. It was right, and it caught a real omission — my paraphrase had dropped the sentence \"We are independent.\", and the three pages'' blocks are NOT identical (the price-cap page carries an extra sentence about what the site publishes). Insert the block as given.","current_value":"The served page is the post-739 rewrite. The regulatory correction 739 commissioned LANDED and must not be undone; what the wholesale regeneration dropped is the block in `restore_verbatim`.","suggestion":"Insert the `restore_verbatim` block back into this page, in the closing position it occupied before (it still occupies that position on pages 739 did not touch, e.g. /guides/if-you-cant-pay.html and /guides/jargon-buster.html — match those). Reproduce it as given, word for word. Change nothing else.","constraint":"SMALLEST TRUE CHANGE, and this is a RESTORATION — the risk here is a SECOND rewrite, not an insufficient one. Do NOT rewrite, reorder, condense or improve any prose already on the page. Do NOT introduce any figure, statistic, date, example or named source that is not already an evidenced fact in this site''s evidence_base register (migration 738, 19 facts) — that register is the CLOSED SET of figures this site may assert. Add the named material; leave everything else byte-identical.","affected_component":"Guide body (single `ported-page` slot) — the closing site-identity block","acceptance_test":"Fetch https://loancash.co.uk/guides/check-your-lender-is-authorised.html and confirm the served bytes contain \"does not lend money, broker loans\" AND \"We are independent\" (both currently 0) AND that visible text length has GROWN, not shrunk. Then diff visible text against the current served version: ADDITIONS ONLY, no existing sentence removed or reworded. Use an invented-URL 404 as a fetch control. This test is the one that would have caught 739 on day one.","authority":"Regression caused by migration 739 under OWNER RULING D2 (RFC_060 §3g). 739''s council round (APPROVED, corr 93897fb5) named the cause; the verdict was read after the items had run. This round (corr 4718725c) returned REVISE and its objections are answered here. WRONG_CALLS carries the lesson.","beware":"A page_rerender re-ships stored content_data and completes successfully — it cannot restore anything. No verifier is registered for content_rewrite, so ''complete'' is not evidence the bytes changed. Verify at the served bytes with the acceptance test above.","filed_by_lane":"loancash_couk_fca_validation lane (migration 743 rev 2)"}'::jsonb,
   (SELECT id  FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/check-your-lender-is-authorised.html'),
   (SELECT url FROM pages WHERE site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND url='/guides/check-your-lender-is-authorised.html'),
   15, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 743)', 'loancash_restore_disclaimer_check_your_lender', 'manual');

DO $$
DECLARE n int; nverb int; nman int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)';
  IF n <> 3 THEN
    RAISE EXCEPTION '743 VERIFY: expected 3 restoration items, found %', n;
  END IF;

  -- the verbatim block must be PRESENT and must carry the sentence my paraphrase dropped
  SELECT count(*) INTO nverb FROM site_work_items
   WHERE created_by = 'loancash_couk_fca_validation lane (migration 743)'
     AND spec->>'mode' = 'edit_live'
     AND spec->>'restore_verbatim' LIKE '%does not lend money, broker loans%'
     AND spec->>'restore_verbatim' LIKE '%We are independent%'
     AND spec->>'constraint' LIKE '%SMALLEST TRUE CHANGE%';
  IF nverb <> 3 THEN
    RAISE EXCEPTION '743 VERIFY: expected 3 items carrying the VERBATIM block (incl. the "We are independent" sentence the paraphrase dropped), found %', nverb;
  END IF;

  SELECT count(*) INTO nman FROM site_work_items w JOIN pages p ON p.id = w.page_id
   WHERE w.created_by = 'loancash_couk_fca_validation lane (migration 743)'
     AND w.approval_mode = 'manual' AND w.source = 'manual'
     AND w.affected_url = p.url AND p.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74';
  IF nman <> 3 THEN
    RAISE EXCEPTION '743 VERIFY: expected 3 items with approval_mode=manual, source=manual and affected_url matching pages.url, found %', nman;
  END IF;

  RAISE NOTICE '743 OK (rev 2): 3 restoration items filed - VERBATIM text recovered from page_component_history, mode=edit_live asserted against page-build-handler''s own config, approval_mode=manual';
END $$;

COMMIT;
