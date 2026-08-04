-- 300 — bugs_open/187: retire the needs_page rows that were unsatisfiable at birth.
--
-- Scope: ONLY rows whose target page resolves NO sections anywhere in
-- page-build-handler's chain AND is in no current plan (measured per-row in the
-- bug file's triage, 2026-08-03), plus the one row whose page is archived.
-- wont_fix, NOT complete: no work happened, and complete would release any
-- dependent on a lie (the 297 precedent from bugs_closed/177).
--
-- Deliberately NOT touched:
--   * reconcile_site_plan rows for pages with 0 sections + 0 plan rows
--     (directory-index, practice, guides-index, brand-detail,
--     platform-log-index) — bugs_closed/015's shape: REAL gaps the item is
--     correctly surfacing. They stay parked for the queue's owner.
--   * Rows whose ask is satisfied or satisfiable (tungsten-guide, board-setup,
--     cases-index, thames-water, grip-styles, brands-index, shop-index,
--     password-entropy) — the needs_page revalidator judges those WITH an
--     audit trail once the image rolls; hand-closing here would skip the
--     evidence the close is supposed to carry.
--   * The four sectionless tool pages that ARE current-plan members
--     (tool-grip-force-friction-calculator, tool-gripper-cycle-time-estimator,
--     tool-gripper-payload-calculator, tool-matchmatrix — measured 2026-08-03):
--     plan membership makes them synthesis-eligible, so "unsatisfiable" is not
--     provable; the fact they still parked means synthesis found no same-role
--     donor — a plan-shape question for the queue's owner (TL-009 territory),
--     not a sweep's call.
--
-- Apply AFTER the guarded image is rolled (the emitters keep minting until
-- then; a pre-roll sweep invites a same-key re-mint racing the sweep).
--
-- Verify block: expects EXACTLY the rows this file names, and aborts the
-- transaction otherwise (DO/RAISE — a bare SELECT cannot stop a COMMIT).

BEGIN;

WITH parked AS (
  SELECT w.id, w.site_id, split_part(w.item_key,':',2) AS page_name
  FROM site_work_items w
  WHERE w.item_type = 'needs_page'
    AND w.status = 'needs_human_review'
    AND w.error LIKE '%no sections ready to build%'
    AND w.source IN ('image-build-handler','page-rerender')
),
unsatisfiable AS (
  SELECT k.id
  FROM parked k
  JOIN pages p ON p.site_id = k.site_id AND p.name = k.page_name
  WHERE (
          p.status = 'archived'
       OR (
            COALESCE(jsonb_array_length(p.sections),0) = 0
        AND NOT EXISTS (
              SELECT 1 FROM site_plan_sections sps
              JOIN site_plans pl ON pl.id = sps.plan_id AND pl.is_current
              WHERE pl.site_id = k.site_id AND sps.page_name = k.page_name)
        AND NOT EXISTS (
              SELECT 1 FROM site_plan_pages spp
              JOIN site_plans pl2 ON pl2.id = spp.plan_id AND pl2.is_current
              WHERE pl2.site_id = k.site_id AND spp.name = k.page_name)
          )
        )
)
UPDATE site_work_items w
SET status = 'wont_fix',
    resolution_path = 'manual:187_sweep',
    result = COALESCE(w.result,'{}'::jsonb) || jsonb_build_object(
      'sweep', jsonb_build_object(
        'bug', 'bugs_open/187',
        'reason', 'unsatisfiable at birth: page resolves no sections in the handler chain and is in no current plan (or page archived); emit-side guard shipped',
        'original_error', w.error,
        'swept_at', now()::text)),
    updated_at = now()
FROM unsatisfiable u
WHERE w.id = u.id;

-- Verify: the update must have retired every image-build-handler/page-rerender
-- parked row EXCEPT the satisfiable/plan-member ones named above (expected
-- leftovers: brands-index, password-entropy, shop-index, and the four
-- plan-member tool pages). Abort on any other outcome.
-- (First draft asserted 3 leftovers — corrected after measuring plan
-- MEMBERSHIP, not just plan section rows: 4 sectionless tool pages are
-- current-plan members and must not be swept. NOTES 2026-08-03.)
DO $$
DECLARE
  remaining integer;
  leftovers text;
BEGIN
  SELECT count(*), COALESCE(string_agg(DISTINCT split_part(item_key,':',2), ',' ORDER BY split_part(item_key,':',2)), '')
  INTO remaining, leftovers
  FROM site_work_items
  WHERE item_type='needs_page' AND status='needs_human_review'
    AND error LIKE '%no sections ready to build%'
    AND source IN ('image-build-handler','page-rerender');
  IF remaining <> 7 OR leftovers <> 'brands-index,password-entropy,shop-index,tool-grip-force-friction-calculator,tool-gripper-cycle-time-estimator,tool-gripper-payload-calculator,tool-matchmatrix' THEN
    RAISE EXCEPTION '187 sweep verify failed: % rows remain (%), expected 7 — re-run the triage before applying', remaining, leftovers;
  END IF;
END $$;

COMMIT;
