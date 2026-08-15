-- SQL_2026-08-15_215_o2_pair7_cycle_time_merge.sql
-- bugs_open/215 O2 — PAIR 7, robot-hands.com `gripper-cycle-time-estimator`.
-- THE MERGE HALF ONLY. Nothing is archived, retracted or removed from the plan here.
--
-- OWNER RULING 2026-08-13: "MERGE the bare page's prose into `tool-`, then retire bare."
-- OWNER RULING 2026-08-15: merge the EXPLAINER and the FAQ only (option A of three).
--   The hero (88 w) and the closing CTA (88 w) stay behind and die with the bare page:
--   the hero duplicates the tool component's own heading and its button reads
--   "Run the Estimator" while pointing at /contact.html (a defect already live on the
--   bare page today); the CTA's primary button points at
--   /tools/gripper-cycle-time-estimator/index.html, i.e. at the SURVIVOR, so moving it
--   verbatim would make a self-link. Neither is prose. 1,587 of the ~1,700 words move.
--
-- WHAT THIS IS. Not an authoring job and not an LLM call. The two components below are
-- finished, rendered and already deployed on the bare page; this copies the rows onto
-- the survivor verbatim — same component_id, same content_data, same rendered_html.
-- Nothing is written by a session, which is what CLAUDE.md's 2026-08-06 ruling requires:
-- the framework wrote this copy, and it stays exactly as the framework wrote it.
--   [MEASURED 2026-08-15] generic-text-block 449 w / 3,305 b html; faq 1,138 w (8 Q&As)
--   / 8,933 b html. Neither content_data contains a single URL, so nothing repoints and
--   there is no link decision hiding in either. The tool component is deliberately NOT
--   copied — it is the actual duplicate, and the survivor's own variant is richer
--   (16,850 b vs the bare page's 12,520 b).
--
-- WHY A DIRECT INSERT IS THE SANCTIONED ROUTE, not a workaround.
-- The survivor is rebuild_policy='owned'. SavePageSectionsAction HARD-REFUSES an owned
-- page ("a generic section save would clobber it", save_page_sections_action.go:186-196)
-- and apply_section_edit only edits an EXISTING page_components row (content_edit /
-- component_swap) — neither can add a section. owned_page_guard.go:29-36 states the
-- design: the guard sits at assemble_page precisely BECAUSE re-assembly of existing
-- page_components "is deliberately NOT gated — it is how owned pages deploy".
-- So: INSERT the rows, then deploy ASSEMBLE-ONLY. This is 267's pattern
-- (docs/agent_docs/sql_for_agents/267_tool_guide_intro_recovery_waterfall.sql), the
-- worked precedent for adding a prose section to an owned tool page.
--
-- THE SHAPE IS ROUTINE, not novel. [MEASURED 2026-08-15, fleet-wide] 23 active
-- rebuild_policy='owned' pages carry more than one component — 9 on
-- loanandmortgagecalculator.co.uk with a deliberate prose-0 / tool-1 / prose-2
-- interleave, plus oufe.com tool-recovery-waterfall and webdesign.co.uk
-- tool-ab-test-calculator. An earlier reading of this lane's own three robot-hands
-- tool- pages (all single-component) suggested tool pages never carry prose; that is
-- an n=3 inference and it is FALSE fleet-wide.
--
-- TWO DELIBERATE DEVIATIONS FROM 267, both stated so a reviewer can overrule them:
--   1. pages.sections is NOT updated. 267 appended its new slot there. Assembly does
--      not read it: rerender_single_page_action.go:839-845 is
--        SELECT rendered_html, slot_name FROM page_components
--        WHERE page_id=$1 AND build_status IS DISTINCT FROM 'removed' ORDER BY position ASC
--      pages.sections is a planning cache / legacy fallback
--      (load_page_sections_from_spec_action.go:10-22, design_actions.go:335-341).
--      Writing it here would also flip ensure_page_section_layout_action.go:118 from
--      "will write a layout" to "refuses, already non-empty" — a behaviour change this
--      task does not need. The survivor's cache is [] today and the page renders fine.
--   2. The rows are NOT locked. 267 locked 'permanent' because that site's copy is
--      hand-authored and must never be regenerated. This copy is framework-written, so
--      locking it would opt it out of ordinary maintenance for no gain. Owned pages are
--      already excluded from the generic build loops (get_pages_to_build_actions.go:63,
--      bugs_open/208), and a section_data_resolved rerender re-renders from stored
--      content_data with NO LLM call.
--
-- SAFETY GATE CLEARED BEFORE WRITING [MEASURED 2026-08-15]: the survivor's existing
-- component has content_data = '{}', NOT NULL. 049b_deploy_single_page.sh's header
-- warns that if ANY section has NULL content_data the whole page escalates to the
-- content writer and the copy IS regenerated — which here would have rewritten the tool
-- itself. It is '{}', so the assemble-only deploy is safe. deploy_mode is absent, so the
-- page is not on the verbatim path (rerender_single_page_action.go:287-311) and already
-- assembles normally; a second component is exactly the case that code anticipates.
--
-- REPLAY GUARD: NOT EXISTS on (survivor, slot_name). Re-running inserts nothing.
-- REVERT (exact, restores the pre-execution state):
--   DELETE FROM page_components
--    WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
--      AND slot_name IN ('generic-text-block','faq');
--
-- AFTER THIS COMMITS, ONE DISPATCH IS OWED (assemble-only, no reason argument):
--   ./docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh \
--     acc27598-28c6-4950-9ec5-61b1a9f5061d \
--     00ff3af5-dad8-4770-9f70-3edc267a3c92 \
--     robot-hands.com
--   Pass NO 4th argument. A reason takes the rerender_sections pre-pass instead.
--   Then verify at the artefact: the survivor must gain ~1,587 words
--   (32,165 b / 2,129 w  ->  expect ~3,7xx w), the bare page must be UNCHANGED at
--   46,158 b / 3,832 w (it is untouched by this file), and a collateral control
--   (/how-it-works.html, 29,993 b) must stay 200.

BEGIN;

-- ---------------------------------------------------------------- pre-assertions
-- DO/RAISE, not SELECTs: a verify block of SELECTs cannot stop the COMMIT, because
-- ON_ERROR_STOP ignores a non-empty result set.
DO $pre$
DECLARE
  n_survivor_comps int;
  n_source_rows    int;
  survivor_policy  text;
  survivor_status  text;
  n_already        int;
BEGIN
  SELECT COALESCE(rebuild_policy,'generic'), status
    INTO survivor_policy, survivor_status
    FROM pages WHERE id='acc27598-28c6-4950-9ec5-61b1a9f5061d';
  IF survivor_policy IS NULL THEN
    RAISE EXCEPTION 'ABORT: survivor page acc27598 not found';
  END IF;
  IF survivor_policy <> 'owned' THEN
    RAISE EXCEPTION 'ABORT: survivor rebuild_policy is %, expected owned — the premise of this file changed', survivor_policy;
  END IF;
  IF survivor_status <> 'active' THEN
    RAISE EXCEPTION 'ABORT: survivor status is %, expected active', survivor_status;
  END IF;

  SELECT count(*) INTO n_survivor_comps
    FROM page_components WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d';
  IF n_survivor_comps <> 1 THEN
    RAISE EXCEPTION 'ABORT: survivor has % components, expected exactly 1 (the tool). Another session has changed this page', n_survivor_comps;
  END IF;

  -- The tool component must still carry non-NULL content_data, or the deploy that
  -- follows this file escalates to the content writer and regenerates the tool.
  IF EXISTS (SELECT 1 FROM page_components
              WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
                AND content_data IS NULL) THEN
    RAISE EXCEPTION 'ABORT: survivor has a NULL content_data — an assemble-only deploy would regenerate the tool copy';
  END IF;

  SELECT count(*) INTO n_source_rows
    FROM page_components
   WHERE page_id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b'
     AND slot_name IN ('generic-text-block','faq')
     AND content_data IS NOT NULL
     AND COALESCE(rendered_html,'') <> '';
  IF n_source_rows <> 2 THEN
    RAISE EXCEPTION 'ABORT: expected 2 source rows (generic-text-block, faq) with content and html on the bare page, found %', n_source_rows;
  END IF;

  SELECT count(*) INTO n_already
    FROM page_components
   WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
     AND slot_name IN ('generic-text-block','faq');
  IF n_already <> 0 THEN
    RAISE NOTICE 'NOTE: % of the 2 slots already present on the survivor — replay guard will skip them', n_already;
  END IF;
END
$pre$;

-- ---------------------------------------------------------------- the copy
-- Verbatim: same component_id, same content_data, same rendered_html, same
-- build_status as the already-deployed source row. Only page_id and position change.
-- Positions preserve the bare page's reading order (tool 2 -> explainer 3 -> faq 4);
-- the survivor's tool component already sits at position 2.
INSERT INTO page_components (page_id, component_id, position, slot_name,
                             content_data, rendered_html, build_status)
SELECT 'acc27598-28c6-4950-9ec5-61b1a9f5061d'::uuid,
       src.component_id,
       CASE src.slot_name WHEN 'generic-text-block' THEN 3
                          WHEN 'faq'                THEN 4 END,
       src.slot_name,
       src.content_data,
       src.rendered_html,
       src.build_status
FROM page_components src
WHERE src.page_id = 'abae9dc9-8f3b-4e3f-97f7-b31439b29e1b'
  AND src.slot_name IN ('generic-text-block','faq')
  AND NOT EXISTS (
        SELECT 1 FROM page_components dst
         WHERE dst.page_id  = 'acc27598-28c6-4950-9ec5-61b1a9f5061d'
           AND dst.slot_name = src.slot_name);

-- ---------------------------------------------------------------- post-assertions
DO $post$
DECLARE
  n_comps    int;
  n_prose    int;
  bad_html   int;
  bad_order  int;
  loser_n    int;
BEGIN
  SELECT count(*) INTO n_comps
    FROM page_components WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d';
  IF n_comps <> 3 THEN
    RAISE EXCEPTION 'ABORT: survivor has % components after the insert, expected 3', n_comps;
  END IF;

  SELECT count(*) INTO n_prose
    FROM page_components
   WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
     AND slot_name IN ('generic-text-block','faq');
  IF n_prose <> 2 THEN
    RAISE EXCEPTION 'ABORT: expected both prose slots on the survivor, found %', n_prose;
  END IF;

  -- The copy must be byte-identical to the source, or it is not the merge the owner
  -- decided — it is a different page with similar words.
  SELECT count(*) INTO bad_html
    FROM page_components dst
    JOIN page_components src
      ON src.page_id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b'
     AND src.slot_name=dst.slot_name
   WHERE dst.page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
     AND dst.slot_name IN ('generic-text-block','faq')
     AND (dst.rendered_html IS DISTINCT FROM src.rendered_html
          OR dst.content_data IS DISTINCT FROM src.content_data
          OR dst.component_id IS DISTINCT FROM src.component_id);
  IF bad_html <> 0 THEN
    RAISE EXCEPTION 'ABORT: % copied row(s) differ from the source', bad_html;
  END IF;

  -- Assembly is ORDER BY position ASC, so the tool must still come first.
  SELECT count(*) INTO bad_order
    FROM page_components
   WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
     AND ((slot_name='tool-gripper-cycle-time-estimator' AND position <> 2)
       OR (slot_name='generic-text-block'                AND position <> 3)
       OR (slot_name='faq'                               AND position <> 4));
  IF bad_order <> 0 THEN
    RAISE EXCEPTION 'ABORT: positions are wrong on % row(s) — assembly orders by position', bad_order;
  END IF;

  -- The bare page must be untouched: this file is the MERGE half only.
  SELECT count(*) INTO loser_n
    FROM page_components WHERE page_id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b';
  IF loser_n <> 5 THEN
    RAISE EXCEPTION 'ABORT: bare page now has % components, expected its original 5 — it must not be modified here', loser_n;
  END IF;

  RAISE NOTICE 'OK: survivor now carries tool(2) + explainer(3) + faq(4); copies byte-identical; bare page untouched at 5 components';
END
$post$;

COMMIT;

-- VERIFY: three slots in reading order, the tool first, real HTML on each.
SELECT pc.position, pc.slot_name, pc.build_status,
       length(pc.rendered_html) AS html_b,
       length(pc.content_data::text) AS data_b
FROM page_components pc
WHERE pc.page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
ORDER BY pc.position;
