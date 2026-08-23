-- 575 — enable outbound suppression of links to never-shipped pages (bugs_open/328)
--
-- ⚠ _HOLD: DO NOT APPLY UNTIL THE GO CHANGE HAS ROLLED.
--
-- This is the CONFIG half of bugs_open/328. The Go half is inert until an image
-- is built from committed HEAD and rolled; this file is live the moment it
-- applies. Applying it early is harmless (an unknown config key is simply
-- unread by the old binary) but it would make the fix look live when it is not,
-- and "did my fix ship" would then have two answers. Held until the roll, then
-- applied by hand — a migration's guard checks DRIFT, not ORDER, so a banner
-- cannot hold it; the _HOLD suffix can.
--
-- WHAT IT TURNS ON. `suppress_unshipped_links: true` on the step configs of the
-- two OUTBOUND seams — the last point before a page's HTML leaves for deploy.
-- With the flag absent (today, and on every step this file does not touch) the
-- seam returns the html byte-identical, which is the owner's RFC_010 §2 rule:
-- new authority on a shared seam ships as an opt-in field whose UNSAFE side —
-- here, the writer that removes an anchor — is the default-OFF one.
--
-- WHAT THE SEAM DOES, in one line a reviewer of this file can check: it drops
-- in-body anchors whose target page this site's own `pages` table says would
-- 404 and is not arriving; the authored href is untouched in content_data, so
-- the anchor returns by itself on the first render after the target ships.
--
-- THE STEPS, and why these six paths are the whole live surface (enumerated
-- 2026-08-23, recursively, because three of them are nested in sub_workflows and
-- a top-level jsonb_each MISSES them):
--
--   pageflow-builder        workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page   -> ON
--   page-rebuild            workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page   -> ON
--   site-work-orchestrator  workflow,steps,build_items_loop,config,sub_workflow,steps,assemble_page   -> ON
--   page-rerender           workflow,steps,render_page                                                -> ON
--   report-builder          workflow,steps,render_page                                                -> deliberately LEFT OFF
--   page-rerender           workflow,steps,rerender_sections                                          -> not a seam; see below
--
-- page-rerender.render_page carries the initial build too: page-build-handler's
-- deploy_page step is a call_agent with target_role 'page_renderer', which is
-- the page-rerender agent. So one key covers build, rerender and rebuild.
--
-- rerender_sections is NOT a seam and needs no key: its action
-- (rerender_page_sections) writes sections and always flows onward — through
-- save_sections to render_page — which is where assembly and the repair passes
-- happen. Named here so the next reader does not have to re-derive that it was
-- considered.
--
-- report-builder is left OFF because report pages are not the measured harm and
-- their link shapes were not part of the 2026-08-23 census. Turning it on later
-- is one UPDATE; turning it on now would be an unmeasured claim.
--
-- ROLLBACK: 575_enable_suppress_unshipped_links_HOLD_ROLLBACK.sql — the exact
-- inverse (#- of the one key on each of the four paths).

BEGIN;

-- $pre$ — refuse a re-run and refuse a drifted target. A verify block of plain
-- SELECTs CANNOT stop the COMMIT (ON_ERROR_STOP ignores a non-empty result), so
-- every assertion here is a DO/RAISE.
DO $pre$
DECLARE
    v_missing text;
BEGIN
    SELECT string_agg(t.type || ':' || t.path, ', ')
      INTO v_missing
      FROM (VALUES
            ('pageflow-builder',       '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page}'),
            ('page-rebuild',           '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page}'),
            ('site-work-orchestrator', '{workflow,steps,build_items_loop,config,sub_workflow,steps,assemble_page}'),
            ('page-rerender',          '{workflow,steps,render_page}')
           ) AS t(type, path)
     WHERE NOT EXISTS (
             SELECT 1 FROM agent_definitions a
              WHERE a.type = t.type AND a.is_active
                AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
                AND a.default_config #> t.path::text[] IS NOT NULL
           );
    IF v_missing IS NOT NULL THEN
        RAISE EXCEPTION '575 PRE: step path missing or agent absent for: % — the workflow shape has drifted since 2026-08-23; re-enumerate before applying', v_missing;
    END IF;

    IF EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
           AND default_config::text LIKE '%suppress_unshipped_links%'
    ) THEN
        RAISE EXCEPTION '575 PRE: suppress_unshipped_links is already present somewhere live — already applied? Read it back before re-running';
    END IF;
END
$pre$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,suppress_unshipped_links}',
        'true'::jsonb, true),
       updated_at = now()
 WHERE type IN ('pageflow-builder', 'page-rebuild')
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_items_loop,config,sub_workflow,steps,assemble_page,config,suppress_unshipped_links}',
        'true'::jsonb, true),
       updated_at = now()
 WHERE type = 'site-work-orchestrator'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
        '{workflow,steps,render_page,config,suppress_unshipped_links}',
        'true'::jsonb, true),
       updated_at = now()
 WHERE type = 'page-rerender'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- $post$ — read the keys back and RAISE on any mismatch, including the one that
-- would matter most: report-builder must NOT have acquired the key.
DO $post$
DECLARE
    v_on  int;
    v_off int;
BEGIN
    SELECT count(*) INTO v_on
      FROM (VALUES
            ('pageflow-builder',       '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,suppress_unshipped_links}'),
            ('page-rebuild',           '{workflow,steps,build_pages_loop,config,sub_workflow,steps,assemble_page,config,suppress_unshipped_links}'),
            ('site-work-orchestrator', '{workflow,steps,build_items_loop,config,sub_workflow,steps,assemble_page,config,suppress_unshipped_links}'),
            ('page-rerender',          '{workflow,steps,render_page,config,suppress_unshipped_links}')
           ) AS t(type, path)
      JOIN agent_definitions a
        ON a.type = t.type AND a.is_active
       AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
     WHERE a.default_config #> t.path::text[] = 'true'::jsonb;

    IF v_on <> 4 THEN
        RAISE EXCEPTION '575 POST FAILED: expected 4 steps carrying suppress_unshipped_links=true, found %', v_on;
    END IF;

    SELECT count(*) INTO v_off
      FROM agent_definitions
     WHERE type = 'report-builder' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config::text LIKE '%suppress_unshipped_links%';

    IF v_off <> 0 THEN
        RAISE EXCEPTION '575 POST FAILED: report-builder acquired suppress_unshipped_links; it is deliberately left OFF';
    END IF;
END
$post$;

COMMIT;
