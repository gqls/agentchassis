-- ROLLBACK for 515 — remove the `page_type?` declaration from page-build-handler's
-- plan_sections step, returning `page_type` to the whole-tree search.
--
-- WHEN YOU WOULD RUN THIS: if section planning visibly degrades on pages whose
-- own record is absent from the tree — i.e. the component selector starts
-- choosing worse components because page_type became the page NAME instead of a
-- real type. The observable is component CHOICE, not an error: absence is a
-- handled state (plan_sections_action.go:972-975), so nothing will fail loudly.
--
-- NOTE it restores the SEARCH, not the conflict rows; the rows follow because the
-- search is what writes them. And it restores the wrong-value path this file
-- exists to close (18 of 31 runs, where the only candidates are other pages'
-- page_type) — so prefer fixing the path over rolling this back.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','plan_sections','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '515 ROLLBACK: no live page-build-handler plan_sections step';
    END IF;
    IF NOT (cfg ? 'page_type?') THEN
        RAISE EXCEPTION '515 ROLLBACK: no page_type? key — 515 is not applied, or already rolled back';
    END IF;

    UPDATE agent_definitions
       SET default_config = default_config
             #- ARRAY['workflow','steps','plan_sections','config','page_type?'],
           updated_at = NOW()
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

DO $$
BEGIN
    IF (SELECT default_config #> ARRAY['workflow','steps','plan_sections','config']
          FROM agent_definitions
         WHERE type = 'page-build-handler' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) ? 'page_type?' THEN
        RAISE EXCEPTION '515 ROLLBACK VERIFY FAILED: page_type? still present';
    END IF;
    RAISE NOTICE '515 ROLLBACK OK: page_type? removed';
END $$;

COMMIT;
