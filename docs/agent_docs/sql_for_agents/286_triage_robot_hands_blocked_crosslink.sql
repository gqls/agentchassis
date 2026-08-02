-- 286 — triage of the work-item pair that stalled the fleet (bugs_closed/176 follow-up)
--
-- Owner-directed 2026-08-02: "yes, please triage that blocked item".
-- DATA triage of exactly 2 rows on robot-hands.com. No schema, no config, no code.
--
-- ── WHAT WAS FOUND ────────────────────────────────────────────────────────
--
-- 0733a7a4  needs_content_page  "Write content for tool page: Gripper Safety
--                                Factor Calculator"
--           status needs_human_review, attempt 1/3, created_by tool-generator
--           error: "page-build-handler no-op: no sections ready to build (empty
--           spec sections, or all sections deferred for missing data)"
--
-- 93f2a3b7  content_rewrite     "Add Gripper Safety Factor Calculator tool
--                                reference to how-to-specify-a-gripper page"
--           status triaged, depends_on = {0733a7a4}
--
-- **0733a7a4 is SPURIOUS.** tool-generator built the page and its component and
-- then, 45 MILLISECONDS later, raised an item asking for content to be written
-- for that page:
--     content_components 769c0b80   updated 12:27:28.312336
--     page_components    (1 slot)   updated 12:27:28.342761
--     pages 1f939069                created 12:27:28.326869
--     item  0733a7a4                created 12:27:28.387981   <- 45ms after
-- The page is `deployed`, `deployed_at` set, carrying a 10,336-char tool
-- component. page-build-handler found no sections to build because THERE ARE
-- NONE TO BUILD — and correctly refused to report success (routing to
-- needs_human_review rather than silently completing is the WDS-004 fix working
-- as designed; nothing is broken in the handler).
--
-- **One slot IS the finished shape of a tool page.** Measured fleet-wide
-- 2026-08-02: of 141 tool pages, **116 have exactly 1 slot**, 14 have 0 (never
-- built), 11 have 2+. All five DEPLOYED robot-hands tool pages have exactly 1.
-- So the item asks for work the platform does not do for tool pages at all.
--
-- **It is systematic, not a stray row.** EVERY `tool_content:%` item ever
-- created — **8 of 8**, from 2026-07-14 to 2026-07-31 — sits in
-- needs_human_review with the byte-identical no-op error, each at attempt 1.
-- Not one has ever completed. Filed as `bugs_open/177`.
--
-- **93f2a3b7 is REAL outstanding work.** Verified: none of the three
-- page_components on /how-to-specify-a-gripper.html contains the string
-- "gripper-safety-factor-calculator", so the crosslink genuinely does not exist.
-- Its 5 predecessors all failed at validate_content (0 of 5), and it is the
-- FIRST tool_crosslink ever to carry a depends_on — so the dependency is new
-- behaviour, and it chained real work to a class of item that has never once
-- succeeded.
--
-- ── WHY THE DISPOSITIONS ARE WHAT THEY ARE ────────────────────────────────
--
-- 0733a7a4 → **wont_fix**, NOT complete.
--   `complete` would be the more convenient choice because it is the ONLY thing
--   that releases a dependency (see below) — and that is exactly why it must not
--   be used here. The item's stated work was never performed by anything; the
--   page was already finished before the item existed. Marking it complete would
--   assert a success that did not happen, which is the silent-completion
--   pathology WDS-004 exists to prevent, committed by hand. `wont_fix` states
--   the true thing: this item should never have been raised.
--
-- 93f2a3b7 → **depends_on cleared**, item left `triaged` to be attempted on its
--   own merits.
--   ⚠ **A DEPENDENCY CAN ONLY BE RELEASED BY complete/verified.** The loader's
--   clause is `dep_id NOT IN (SELECT id ... WHERE status IN ('complete',
--   'verified'))`, so wont_fix / rejected / cancelled / failed ALL leave the
--   dependent blocked for ever. There is no "dismissed" state a blocker can
--   reach. Hence clearing the array is the only honest lever available: it
--   neither fabricates a success nor abandons real work.
--   Expected outcome is NOT assumed: on its predecessors' record it may well
--   fail validate_content and land in needs_human_review. That is an acceptable
--   and INFORMATIVE result — it fails loudly against a live page's validator
--   rather than sitting invisible behind a dependency that can never clear.
--   It cannot damage the live page: a validation failure routes the item, it
--   does not publish.
--
-- ── NOT DONE HERE, DELIBERATELY ───────────────────────────────────────────
-- The other 7 `tool_content:%` rows are left untouched. They block nothing (only
-- 0733a7a4 ever had a dependent), so there is no urgency, and sweeping 7 rows on
-- 4 sites is a different action from triaging the one that stalled the fleet.
-- Their disposition belongs with the fix for `bugs_open/177`.

BEGIN;

-- ── STEP 1 — PRE-FLIGHT: assert both rows are exactly as triaged ──────────
-- Expect 1 row, and every boolean TRUE. Any false → the rows moved under us
-- (another session, or a handler); ROLLBACK and re-read before proceeding.
SELECT
  (SELECT status FROM site_work_items WHERE id='0733a7a4-929e-412b-81e8-68c9dbbf5d41')
     = 'needs_human_review'                        AS blocker_still_in_review,
  (SELECT status FROM site_work_items WHERE id='93f2a3b7-4713-493d-aa8b-b828fa9b9126')
     = 'triaged'                                   AS dependent_still_triaged,
  (SELECT depends_on FROM site_work_items WHERE id='93f2a3b7-4713-493d-aa8b-b828fa9b9126')
     = '{0733a7a4-929e-412b-81e8-68c9dbbf5d41}'::uuid[] AS dependency_unchanged,
  (SELECT build_status FROM pages WHERE id='1f939069-aa79-47c0-9511-f270a9a02c3c')
     = 'deployed'                                  AS tool_page_deployed,
  (SELECT count(*) FROM page_components WHERE page_id='1f939069-aa79-47c0-9511-f270a9a02c3c')
     = 1                                           AS tool_page_has_its_one_slot;

-- ── STEP 2 — the spurious blocker, closed honestly ────────────────────────
UPDATE site_work_items
SET status = 'wont_fix',
    completed_at = now(),
    updated_at = now(),
    resolution_path = 'triage-286',
    error = 'wont_fix (triage 2026-08-02, bugs_closed/176): spurious item. tool-generator '
         || 'built this tool page and its component and raised this "write content" item 45ms '
         || 'later. The page is deployed with its single tool slot — the finished shape of 116 '
         || 'of 141 tool pages fleet-wide — so there were never any sections to build. '
         || 'page-build-handler''s no-op was correct and its refusal to report success was '
         || 'correct. 8 of 8 tool_content items have failed identically since 2026-07-14; the '
         || 'generator defect is bugs_open/177. NOT marked complete: no work was ever done, and '
         || 'complete would assert a success that did not happen. Original handler error was: '
         || COALESCE(error, '(none)')
WHERE id = '0733a7a4-929e-412b-81e8-68c9dbbf5d41'
  AND status = 'needs_human_review';   -- idempotent + refuses a moved row

-- ── STEP 3 — release the dependent, without fabricating a success ─────────
UPDATE site_work_items
SET depends_on = NULL,
    updated_at = now(),
    resolution_path = 'triage-286',
    suggested_action = 'Dependency on 0733a7a4 cleared by triage 2026-08-02: that item is '
                    || 'wont_fix and a blocker can only be released by complete/verified, so it '
                    || 'would have blocked this item for ever. The crosslink itself is REAL — '
                    || 'verified that no page_component of /how-to-specify-a-gripper.html '
                    || 'references gripper-safety-factor-calculator. Note 5 of 5 previous '
                    || 'tool_crosslink items failed at validate_content; if this one does too, '
                    || 'that is the next thing to investigate, not this dependency.'
WHERE id = '93f2a3b7-4713-493d-aa8b-b828fa9b9126'
  AND status = 'triaged'
  AND depends_on IS NOT NULL;   -- idempotent + refuses a moved row

-- ── STEP 4 — VERIFY BEFORE COMMIT ─────────────────────────────────────────
-- Expect: blocker wont_fix; dependent triaged with depends_on NULL.
SELECT left(id::text,8) AS item, item_type, status, depends_on, completed_at
FROM site_work_items
WHERE id IN ('0733a7a4-929e-412b-81e8-68c9dbbf5d41','93f2a3b7-4713-493d-aa8b-b828fa9b9126')
ORDER BY created_at;

COMMIT;

-- ── ROLLBACK ──
-- UPDATE site_work_items SET status='needs_human_review', completed_at=NULL,
--        resolution_path=NULL
--  WHERE id='0733a7a4-929e-412b-81e8-68c9dbbf5d41';
-- UPDATE site_work_items
--    SET depends_on='{0733a7a4-929e-412b-81e8-68c9dbbf5d41}'::uuid[],
--        resolution_path=NULL, suggested_action=NULL
--  WHERE id='93f2a3b7-4713-493d-aa8b-b828fa9b9126';
-- (The original handler error text is preserved inside the new error string.)

-- ── VERIFY AT THE ARTEFACT ──
-- robot-hands.com should now become a normal dispatch participant: it has one
-- eligible item again and, under 285, the selector will only pick it when the
-- loader can actually load it. Watch for the crosslink being attempted:
--
--   SELECT status, attempt_count, left(error,120) FROM site_work_items
--   WHERE id='93f2a3b7-4713-493d-aa8b-b828fa9b9126';
--
-- SUCCESS looks like: the /how-to-specify-a-gripper.html page gaining a link to
-- /tools/gripper-safety-factor-calculator/index.html. Check the ARTEFACT, not the
-- status (WDS-004):
--   SELECT slot_name FROM page_components
--   WHERE page_id='5a385981-c2fd-4edb-bc4d-927b93177281'
--     AND rendered_html LIKE '%gripper-safety-factor-calculator%';
