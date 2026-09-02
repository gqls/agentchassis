-- 692_aiao_remove_degraded_duplicate_estimator_placement.sql
--
-- ai-agent-orchestration.com — `/tools/agent-complexity-estimator.html` is serving the
-- SAME TOOL TWICE, and the second copy is a degraded regeneration. Removes the
-- duplicate PLACEMENT (not the component) and restores the page to the state this
-- lane verified on 2026-08-25.
--
-- ── WHAT A VISITOR SEES TODAY (measured 2026-09-02, live HTTP 200, 72,629 bytes) ──
--   TWO `<h2>Agent Architecture Complexity Estimator` headings
--   TWO estimator UIs — "Estimate complexity" and "Generate Architecture Estimate →"
--
-- ── WHY THE NEWER ONE GOES, AND IT IS NOT SENIORITY ────────────────────────────
-- The newer placement is a MUCH REDUCED tool. Measured on the stored artefacts:
--
--   b2b7acbd (2026-04-09, kept)   22,732 bytes   4 fieldsets   4 legends   **12 inputs**
--   9aa63fc0 (2026-08-26, removed) 19,964 bytes  1 fieldset    1 legend    **1 input**
--
-- An estimator with one input where twelve existed is not an equivalent fork; it is a
-- fragment. This is `bugs_open/012`'s family — a regeneration that persists less than
-- it replaced and reports success. **The byte counts are close enough (−12%) that a
-- size check alone would have waved it through; the input count is what shows it.**
--
-- ⚠ AND THE CONTRAST EVIDENCE AGREES. This page measured **0** firm failures on
-- 2026-08-25 after migration `625`. It measures **1** today — `BUTTON#c-tool-agent-
-- complexity-estimator-estimate-button`, `#080B10` on `#0D1117` = 1.04:1 — and that
-- button belongs to the NEW component, which never received `625`'s repoint. So the
-- duplicate also silently re-opened a defect this lane had closed.
--
-- ⚠ NOT A CLASS. Censused fleet-wide: exactly ONE page has two placements in one slot
-- with different `component_id`s created since 2026-08-20 — this one. So this file
-- repairs an incident; it does not paper over a pattern. **Whatever created it is not
-- identified** — there is no migration and no commit anywhere in the repo naming
-- `tool-agent-complexity-estimator-ai-agent-orchestration-com`, so it was written by a
-- live agent process. **If the duplicate returns, that is the finding**, and it should
-- be filed against whatever fork-and-place path produced a 1-input estimator.
--
-- ⚠ THE COMPONENT IS NOT DELETED, only this placement. `a6322da1` keeps its row in
-- `content_components`; it simply stops being rendered on this page. It has exactly
-- ONE placement (this one), so after this file it has none — that is intentional and
-- reversible, and it leaves the evidence intact for whoever diagnoses the producer.
--
-- ⚠ THE md5 DISCRIMINATOR THE LANDMINE PRESCRIBES IS VACUOUS HERE, and that matters.
-- The standing rule for de-duplicating `page_components` is "act only where
-- `count(DISTINCT md5(content_data))` = 1". Both rows carry `content_data = '{}'`, so
-- both md5s are `99914b93…` (the md5 of `{}`) and the rule reports agreement it has not
-- actually established — the same shape the landmine warns about for NULL. **The
-- discriminator used here is the rendered artefact's own structure (fieldsets, legends,
-- inputs), not `content_data`.**
--
-- DOES NOT RE-RENDER. The page must be re-assembled afterwards for the change to
-- reach the site. ⚠ This page is `rebuild_policy='owned'`, so use the owned-page route
-- (`refresh_owned_page_chrome.sh`, assemble mode), NEVER a generic rebuild — see 625.
--
-- ROLLBACK: 692_..._ROLLBACK.sql (re-inserts the exact row from migration_backups)

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '692_aiao_remove_degraded_duplicate_estimator_placement', 'page_components', pc.id::text,
       to_jsonb(pc), 'pre-692 FULL ROW of the degraded duplicate estimator placement'
FROM page_components pc WHERE pc.id = '9aa63fc0-5edf-4768-8e90-cade95e6cf34';

DELETE FROM page_components WHERE id = '9aa63fc0-5edf-4768-8e90-cade95e6cf34';

DO $$
DECLARE b int; n int; inputs int;
BEGIN
  SELECT count(*) INTO b FROM migration_backups
   WHERE migration_name='692_aiao_remove_degraded_duplicate_estimator_placement';
  IF b <> 1 THEN RAISE EXCEPTION '692: expected 1 backup row (the whole row), wrote %', b; END IF;

  -- Exactly one placement left in that slot, and it must be the COMPLETE tool.
  SELECT count(*) INTO n FROM page_components
   WHERE page_id='7c66cb41-4eba-451d-a483-54e4c541c7ba' AND slot_name='tool-agent-complexity-estimator';
  IF n <> 1 THEN RAISE EXCEPTION '692: expected 1 remaining placement in the slot, found %', n; END IF;

  SELECT (SELECT count(*) FROM regexp_matches(pc.rendered_html,'<input','g')) INTO inputs
    FROM page_components pc
   WHERE pc.page_id='7c66cb41-4eba-451d-a483-54e4c541c7ba' AND pc.slot_name='tool-agent-complexity-estimator';
  IF inputs < 12 THEN
    RAISE EXCEPTION '692: the surviving estimator has only % inputs — the WRONG row was removed, roll back now', inputs;
  END IF;

  -- The survivor must still carry 625's repoint, or this file has undone that work.
  SELECT count(*) INTO n FROM page_components pc
   WHERE pc.page_id='7c66cb41-4eba-451d-a483-54e4c541c7ba'
     AND pc.slot_name='tool-agent-complexity-estimator'
     AND pc.rendered_html LIKE '%var(--color-primary-ink, var(--color-primary))%';
  IF n <> 1 THEN RAISE EXCEPTION '692: the surviving estimator does not carry migration 625''s ink repoint'; END IF;

  -- Nothing else on this site may have been deleted.
  SELECT count(*) INTO n FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da';
  IF n < 100 THEN RAISE EXCEPTION '692: only % page_components remain on this site — far too few, something else was deleted', n; END IF;

  RAISE NOTICE '692 OK: degraded duplicate placement removed; the surviving estimator has >=12 inputs and still carries 625. Re-assemble via the OWNED-page route, then verify over HTTP.';
END $$;

COMMIT;
