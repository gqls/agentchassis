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
-- WHAT IT TURNS ON. `suppress_unshipped_links: true` on every live step config of
-- the two OUTBOUND seams — the last point before a page's HTML leaves for deploy.
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
--   report-builder          workflow,steps,render_page                                                -> ON (see below)
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
-- report-builder WAS going to be left OFF as "not the measured harm". The council's
-- bug_historian seat objected (2026-08-23) that this is the estate's most recent
-- recurring shape -- one call site of a shared mechanism guarded while the
-- mechanism stays generic at every unaudited call site -- and that shipping a
-- KNOWN unmeasured gap is the same reasoning that leaves the other path live until
-- someone hits it. So it was MEASURED instead of argued: there are **ZERO** pages
-- of page_type 'report' fleet-wide, and none of the 24 pages currently serving a
-- dead anchor is one. An empty population cannot regress, so turning it ON is free
-- and it removes the "one call site guarded" shape entirely. Cheaper than
-- defending the exception.
--
-- ROLLBACK: 575_enable_suppress_unshipped_links_HOLD_ROLLBACK.sql — the exact
-- inverse (#- of the one key on each of the five paths).

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
            ('page-rerender',          '{workflow,steps,render_page}'),
            ('report-builder',          '{workflow,steps,render_page}')
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
 WHERE type IN ('page-rerender', 'report-builder')
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- $post$ — read the keys back and RAISE on any mismatch, including the arm that
-- matters most: NO live step running a seam action may be left without the key.
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
            ('page-rerender',          '{workflow,steps,render_page,config,suppress_unshipped_links}'),
            ('report-builder',          '{workflow,steps,render_page,config,suppress_unshipped_links}')
           ) AS t(type, path)
      JOIN agent_definitions a
        ON a.type = t.type AND a.is_active
       AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
     WHERE a.default_config #> t.path::text[] = 'true'::jsonb;

    IF v_on <> 5 THEN
        RAISE EXCEPTION '575 POST FAILED: expected 5 steps carrying suppress_unshipped_links=true, found %', v_on;
    END IF;

    -- No seam is left unguarded: assert that every live step running one of the
    -- two seam actions now carries the key. This is the arm that would catch a
    -- SIXTH consumer appearing between the enumeration and the apply.
    SELECT count(*) INTO v_off
      FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
     WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
       AND (s.value->>'action') = 'rerender_single_page'
       AND COALESCE(s.value->'config'->>'suppress_unshipped_links', 'false') <> 'true';

    IF v_off <> 0 THEN
        RAISE EXCEPTION '575 POST FAILED: % live rerender_single_page step(s) do NOT carry suppress_unshipped_links — a seam consumer appeared since the 2026-08-23 enumeration', v_off;
    END IF;
END
$post$;

COMMIT;
