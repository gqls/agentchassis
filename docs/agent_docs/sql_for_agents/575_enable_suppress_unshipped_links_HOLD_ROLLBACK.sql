-- ROLLBACK for 575 — disable outbound link suppression (bugs_open/328)
--
-- The exact inverse: removes the one key from each of the four step configs.
-- It does NOT restore a whole config, deliberately — 575 added exactly one key
-- per path and nothing else, so #- is a precise inverse and a snapshot restore
-- would clobber anything another session has changed on those rows since.
--
-- After this runs the seams read the key as absent, which is the default-OFF
-- path: the outbound html is returned byte-identical and the platform is back to
-- the pre-328 behaviour.

BEGIN;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,suppress_unshipped_links}',
       updated_at = now()
 WHERE type IN ('pageflow-builder', 'page-rebuild')
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,build_items_loop,config,sub_workflow,steps,assemble_page,config,suppress_unshipped_links}',
       updated_at = now()
 WHERE type = 'site-work-orchestrator'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,render_page,config,suppress_unshipped_links}',
       updated_at = now()
 WHERE type = 'page-rerender'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $post$
DECLARE
    v_left int;
BEGIN
    SELECT count(*) INTO v_left
      FROM agent_definitions
     WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config::text LIKE '%suppress_unshipped_links%';
    IF v_left <> 0 THEN
        RAISE EXCEPTION '575 ROLLBACK FAILED: % live agent(s) still carry suppress_unshipped_links', v_left;
    END IF;
END
$post$;

COMMIT;
