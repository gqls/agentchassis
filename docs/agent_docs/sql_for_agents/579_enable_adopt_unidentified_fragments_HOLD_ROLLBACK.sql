-- ROLLBACK for 579 — DISARM `adopt_unidentified_fragments` on all six live steps.
--
-- The exact inverse: removes the one key from each of the six paths 579 set. With
-- the key absent the code takes its default (OFF) and `save_page_sections` behaves
-- byte-identically to how it behaved before phase 2 — there is no third state.
--
-- ⚠ WHAT DISARMING DOES **NOT** DO, and this is the part to understand before
-- running it: it does not un-adopt rows already adopted. Any `page_components` row
-- already bound to `adopted-fragment` STAYS bound, keeps its `content_data.body`
-- and keeps its provenance stamp. That is deliberate and it is safe:
--
--   * those rows are TRUE — the component provably reproduces their bytes, which
--     was checked by a byte-identity round trip before the binding was written;
--   * they are regenerable, which the rows they replaced were not;
--   * `rendered_html` was never touched, so nothing about what the page SERVES
--     changes either way.
--
-- What disarming DOES change for them is that Layer 2 stops carrying their
-- identity forward, so the NEXT rebuild will re-impose the page plan's identity —
-- i.e. they drift back toward the `hero` mislabelling over time, one rebuild each.
-- That is the correct behaviour for a disarm (it is the pre-phase-2 world) but it
-- is a slow reversion rather than an instant one, so do not expect the population
-- query to move the moment this runs.
--
-- If the intent is to remove the component entirely, disarm FIRST, then re-type
-- the bound rows, and only then run 577's rollback — which refuses while any row
-- is still bound, precisely so the order cannot be got wrong.

BEGIN;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,save_sections,config,adopt_unidentified_fragments}'
 WHERE type IN ('page-build-handler', 'tool-recreation-handler', 'page-rerender')
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,adopt_unidentified_fragments}'
 WHERE type IN ('pageflow-builder', 'page-rebuild')
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,adopt_unidentified_fragments}'
 WHERE type = 'site-work-orchestrator'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE still_armed int;
BEGIN
    SELECT count(*) INTO still_armed
      FROM agent_definitions ad,
           LATERAL jsonb_path_query(ad.default_config, 'strict $.**') AS step(value)
     WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND jsonb_typeof(step.value)='object'
       AND step.value->>'action' = 'save_page_sections'
       AND (step.value->'config'->>'adopt_unidentified_fragments') = 'true';
    IF still_armed <> 0 THEN
        RAISE EXCEPTION 'disarm incomplete: % save_page_sections step(s) still armed', still_armed;
    END IF;
END $$;

COMMIT;
