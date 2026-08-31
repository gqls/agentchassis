-- SEED_2026-08-31b_bind_tools_to_evidence.sql
--
-- Two improve_tool dispatches enforcing the owner's standing rule (2026-08-21:
-- no unsourced figure is published anywhere — not in copy, not as a tool
-- default) on tools the ungated add_tool route shipped 2026-08-27. The
-- evidence phase completed 2026-08-31 (register at 116; NOTES tail carries the
-- two §9 reviews). Route: improve_tool → tool-improver, the shape the
-- framework itself used successfully twice on this site on 08-26 (mobile-fit
-- fix, gas-converter zero-division fix); spec copied from the completed
-- audit_fix row, severity given a real value.
--
-- 1. blue-carbon-estimator: bind presets and scenarios to the newly registered
--    facts, label what is extrapolation. 2. bsf-waste-converter: remove the
--    two uncited "typical trials" numeric ranges (no registered source exists;
--    reword to uncommitted language rather than invent a citation).
--
-- Fences (artifact_check per the SEED_2026-08-25b pattern) are deliberately
-- NOT installed here — they assert copy that does not exist until these items
-- land. Install them after verifying the rewritten pages at the artefact.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE s.domain='agritec.uk'
     AND wi.item_key IN ('bind_evidence_agritec.uk_tool-blue-carbon-estimator',
                         'bind_evidence_agritec.uk_tool-bsf-waste-converter')
     AND wi.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 0 THEN
    RAISE EXCEPTION 'pre-state: expected no open binding items, found %', n;
  END IF;
END $$;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, page_id, component_id, pipeline, approval_mode)
SELECT s.id, 'operator', 'improve_tool', 'high',
       'Bind the Blue Carbon Estimator''s encoded constants to the evidence register: visible sources, labelled presets, anchored scenarios (owner rule 2026-08-21)',
       jsonb_build_object(
         'check', 'tool_auditor',
         'page_id', 'c6240de4-e893-41b7-9671-b093c2a593fe',
         'page_name', 'tool-blue-carbon-estimator',
         'component_id', '7ea442b4-8014-4fcc-9fbb-d7c9e9497fb5',
         'issue',
         'This tool encodes empirical constants with no visible source, on a site whose standing rule is that no unsourced figure is published anywhere. The evidence register now covers them; the tool must say where its numbers come from. Make these changes and no others, and DO NOT change any element id (the acceptance fence keys on the instance-scoped ids). '
         || '(1) Add a visible "Where these figures come from" block near the species presets, with clickable links and a captured-2026-08-31 date, stating: cultivated sugar kelp (Saccharina latissima) dry matter measured between 6.3 and 17.4 per cent of fresh weight across cultivation systems, with 10 per cent dry-to-fresh weight a common reference factor (Frontiers in Marine Science, cultivation trials, https://www.frontiersin.org); kelp carbon content 26 to 32 per cent of dry matter, with 30 per cent the standard assumption in carbon-dioxide-removal analyses (NASEM 2022, A Research Strategy for Ocean-based CDR, https://www.nationalacademies.org); fresh Saccharina moisture 77.5 to 89.8 per cent (peer-reviewed review, https://pmc.ncbi.nlm.nih.gov). '
         || '(2) Beside the species preset selector, state plainly that the presets are indicative starting points and that the sugar kelp and oarweed values sit within the cited kelp ranges, while the sea lettuce (Ulva) and Sargassum values are extrapolations not yet backed by a source on this site — users of those species should enter measured values. '
         || '(3) Where the Conservative / Moderate / Optimistic retention scenarios are presented, add one anchoring sentence: the published synthesis (Krause-Jensen and Duarte 2016) estimates that only about 11 per cent of macroalgal net primary production reaches long-term storage, so the 50 and 90 per cent retention scenarios are hypothetical upper bounds for engineered sinking or burial, not observed system-wide values. '
         || '(4) Do not add any figure that is not in the list above, and every figure added must sit next to its link.'
       ),
       60, 'tool-improver', 'triaged', 'agritek-session-2026-08-31',
       'bind_evidence_agritec.uk_tool-blue-carbon-estimator',
       'c6240de4-e893-41b7-9671-b093c2a593fe'::uuid, '7ea442b4-8014-4fcc-9fbb-d7c9e9497fb5'::uuid,
       'build', 'auto'
  FROM sites s WHERE s.domain='agritec.uk';

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, page_id, component_id, pipeline, approval_mode)
SELECT s.id, 'operator', 'improve_tool', 'medium',
       'Remove the BSF Waste Converter''s two uncited "typical trials" numeric ranges from help copy (owner rule 2026-08-21: no unsourced figure anywhere)',
       jsonb_build_object(
         'check', 'tool_auditor',
         'page_id', '4086bbf7-68ab-485e-a953-281f9940154d',
         'page_name', 'tool-bsf-waste-converter',
         'component_id', 'e8b445d2-db0c-46b7-9c29-ab5c22bf2020',
         'issue',
         'The tool''s input help text asserts two numeric ranges with no source: "Typical trials report 40 to 60%." (dry-matter reduction / conversion efficiency) and "Typical range 20 to 45%." (prepupae yield factor). No registered evidence fact backs either range, and this site''s standing rule is that no unsourced figure is published anywhere. The tool''s design is otherwise correct — every rate is a user input. Fix: REMOVE the two numeric range sentences and replace each with uncommitted language of this shape: "This varies widely with feedstock, temperature and system — enter the value you measure; if unsure, run a small batch and measure it." Do NOT invent a citation, do NOT add any new figure, change nothing else, and do not change any element id (ids are instance-scoped and downstream checks key on them).'
       ),
       60, 'tool-improver', 'triaged', 'agritek-session-2026-08-31',
       'bind_evidence_agritec.uk_tool-bsf-waste-converter',
       '4086bbf7-68ab-485e-a953-281f9940154d'::uuid, 'e8b445d2-db0c-46b7-9c29-ab5c22bf2020'::uuid,
       'build', 'auto'
  FROM sites s WHERE s.domain='agritec.uk';

COMMIT;
