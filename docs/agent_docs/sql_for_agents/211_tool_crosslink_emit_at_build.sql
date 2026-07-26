-- 211_tool_crosslink_emit_at_build.sql — bugs_open/029: stop emitting tool
-- cross-link items at SUGGESTION time; emit them from the tool BUILD paths.
--
-- The defect: a tool page's URL cannot be constructed from the tool's function
-- name. tool-suggester's create_cross_links step called
-- create_tool_cross_link_items, which built "/tools/{function}.html" and baked
-- it into both the rewrite instruction AND its acceptance test. The platform
-- produces three different URL shapes for tool pages
-- (/tools/x.html with the tool- prefix stripped, /tools/tool-x.html with it
-- kept, /tools/x/index.html from CanonicalisePage), so the constructed URL was
-- wrong on all three: 0 of 27 emitted items across 4 sites resolved to a real
-- page, including tools that WERE built. page-content-writer obeyed the
-- fabricated URL because page-build-handler maps
-- rewrite_guidance? -> input_data.spec.suggestion and the item's own
-- acceptance_test demanded that URL. Visitors got 404s.
--
-- The fix is in the chassis binary (emitToolCrossLinkItems in
-- create_tool_cross_link_items.go, called from deploy_tool_action.go and
-- create_tool_component_action.go): the emitter now takes the page's REAL
-- pages.url, which only the build path knows, and gates each item behind the
-- page actually going live.
--
-- This migration does the config half:
--   1. tool-suggester: delete the create_cross_links step, repoint
--      create_items_loop -> complete, drop cross_links_created from the
--      complete step's output_fields.
--   2. tool-deployer  (deploy_tool step): pass related_pages from the work
--      item spec into the action.
--   3. tool-generator (save_tool step):   same.
--
-- ORDERING — this file is safe to apply BEFORE the image roll, and doing so is
-- preferable: part 1 stops the phantom links immediately (the only thing lost
-- meanwhile is cross-linking that is currently 100% broken), and parts 2/3 are
-- inert on the deployed binary, which has no related_pages input and ignores an
-- unknown config key. Parts 2/3 activate when the image carrying
-- emitToolCrossLinkItems ships. (The 147/162 "safe now, active later" pattern.)
--
-- ROLLBACK: restore the three rows from the snapshots taken below, e.g.
--   UPDATE agent_definitions a SET default_config = s.default_config
--   FROM agent_definitions s
--   WHERE a.type = s.type AND a.is_active AND s.is_snapshot
--     AND s.snapshot_reason LIKE '211_tool_crosslink_emit_at_build%';
-- Re-adding create_cross_links restores the fabrication — do not roll back
-- part 1 without also reverting the Go change.

BEGIN;

SELECT snapshot_agent('tool-suggester', '211_tool_crosslink_emit_at_build: pre-update');
SELECT snapshot_agent('tool-deployer', '211_tool_crosslink_emit_at_build: pre-update');
SELECT snapshot_agent('tool-generator', '211_tool_crosslink_emit_at_build: pre-update');

-- ═══════════════════════════════════════════════════════════════════════
-- Part 1: tool-suggester stops emitting cross-links
-- ═══════════════════════════════════════════════════════════════════════

DO $$
DECLARE
  cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type = 'tool-suggester' AND is_active AND COALESCE(is_snapshot,false) = false;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '211: no active tool-suggester';
  END IF;

  -- Guard: we are rerouting the chain we inspected. Either the step is still
  -- wired (first run) or this file already ran (idempotent re-run).
  IF cfg #>> '{workflow,steps,create_items_loop,next_step}' NOT IN ('create_cross_links', 'complete') THEN
    RAISE EXCEPTION '211: create_items_loop.next_step is %, expected create_cross_links or complete — re-inspect before rerouting',
      cfg #>> '{workflow,steps,create_items_loop,next_step}';
  END IF;

  cfg := jsonb_set(cfg, '{workflow,steps,create_items_loop,next_step}', '"complete"'::jsonb);
  cfg := cfg #- '{workflow,steps,create_cross_links}';

  IF cfg #> '{workflow,steps,complete,config,output_fields}' IS NOT NULL THEN
    cfg := jsonb_set(cfg, '{workflow,steps,complete,config,output_fields}',
                     '["evaluation", "items_created"]'::jsonb);
  END IF;

  UPDATE agent_definitions SET default_config = cfg, updated_at = NOW()
  WHERE type = 'tool-suggester' AND is_active AND COALESCE(is_snapshot,false) = false;
END $$;

-- ═══════════════════════════════════════════════════════════════════════
-- Part 2: tool-deployer passes related_pages into deploy_tool_to_site
-- ═══════════════════════════════════════════════════════════════════════

DO $$
DECLARE
  cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type = 'tool-deployer' AND is_active AND COALESCE(is_snapshot,false) = false;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '211: no active tool-deployer';
  END IF;
  IF cfg #>> '{workflow,steps,deploy_tool,action}' <> 'deploy_tool_to_site' THEN
    RAISE EXCEPTION '211: tool-deployer deploy_tool step is %, expected deploy_tool_to_site',
      cfg #>> '{workflow,steps,deploy_tool,action}';
  END IF;

  cfg := jsonb_set(cfg, '{workflow,steps,deploy_tool,config,related_pages}',
                   '"input_data.spec.related_pages"'::jsonb);

  UPDATE agent_definitions SET default_config = cfg, updated_at = NOW()
  WHERE type = 'tool-deployer' AND is_active AND COALESCE(is_snapshot,false) = false;
END $$;

-- ═══════════════════════════════════════════════════════════════════════
-- Part 3: tool-generator passes related_pages into create_tool_component
-- ═══════════════════════════════════════════════════════════════════════

DO $$
DECLARE
  cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type = 'tool-generator' AND is_active AND COALESCE(is_snapshot,false) = false;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '211: no active tool-generator';
  END IF;
  IF cfg #>> '{workflow,steps,save_tool,action}' <> 'create_tool_component' THEN
    RAISE EXCEPTION '211: tool-generator save_tool step is %, expected create_tool_component',
      cfg #>> '{workflow,steps,save_tool,action}';
  END IF;

  cfg := jsonb_set(cfg, '{workflow,steps,save_tool,config,related_pages}',
                   '"input_data.spec.related_pages"'::jsonb);

  UPDATE agent_definitions SET default_config = cfg, updated_at = NOW()
  WHERE type = 'tool-generator' AND is_active AND COALESCE(is_snapshot,false) = false;
END $$;

-- ═══════════════════════════════════════════════════════════════════════
-- Post-conditions (inside the transaction: a failure rolls the file back)
-- ═══════════════════════════════════════════════════════════════════════

DO $$
DECLARE
  sug jsonb;
  dep jsonb;
  gen jsonb;
BEGIN
  SELECT default_config INTO sug FROM agent_definitions
  WHERE type = 'tool-suggester' AND is_active AND COALESCE(is_snapshot,false) = false;
  SELECT default_config INTO dep FROM agent_definitions
  WHERE type = 'tool-deployer' AND is_active AND COALESCE(is_snapshot,false) = false;
  SELECT default_config INTO gen FROM agent_definitions
  WHERE type = 'tool-generator' AND is_active AND COALESCE(is_snapshot,false) = false;

  IF sug #> '{workflow,steps,create_cross_links}' IS NOT NULL THEN
    RAISE EXCEPTION '211 GUARD: create_cross_links still present on tool-suggester';
  END IF;
  IF sug #>> '{workflow,steps,create_items_loop,next_step}' <> 'complete' THEN
    RAISE EXCEPTION '211 GUARD: create_items_loop.next_step is %, expected complete',
      sug #>> '{workflow,steps,create_items_loop,next_step}';
  END IF;
  IF dep #>> '{workflow,steps,deploy_tool,config,related_pages}' <> 'input_data.spec.related_pages' THEN
    RAISE EXCEPTION '211 GUARD: tool-deployer related_pages not wired (got %)',
      dep #>> '{workflow,steps,deploy_tool,config,related_pages}';
  END IF;
  IF gen #>> '{workflow,steps,save_tool,config,related_pages}' <> 'input_data.spec.related_pages' THEN
    RAISE EXCEPTION '211 GUARD: tool-generator related_pages not wired (got %)',
      gen #>> '{workflow,steps,save_tool,config,related_pages}';
  END IF;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## Tool cross-links are emitted by the tool build, not the tool suggestion
Observed: pages linked to /tools/<function>.html for tools that had no such page — 0 of 27 emitted cross-link items across 4 sites resolved to a real page, including tools that were built. leopardessconsulting.co.uk /services.html shipped a clickable 404.
Root cause: create_tool_cross_link_items constructed the URL from the tool function name at SUGGESTION time. Tool pages take three different URL shapes depending on the build path, so the construction was wrong on all three, and the item''s acceptance_test required the fabricated URL, so the writer obeyed it.
Fix: emit from the build paths (deploy_tool_to_site, create_tool_component) using the real pages.url they just created, gated behind the page going live; delete tool-suggester''s create_cross_links step; wire related_pages from the work item spec into both build steps (migration 211 + the chassis image carrying emitToolCrossLinkItems). The suggestion-time action is kept registered but now resolves a real page and emits nothing when there is none.
Verified: config half guarded and applied; Go half activates on the next chassis image; bugs_open/029.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '211_tool_crosslink_emit_at_build'
);

COMMIT;

-- ═══════════════════════════════════════════════════════════════════════
-- Verification (read-only, run after apply)
-- ═══════════════════════════════════════════════════════════════════════

-- 1. Chain and wiring
SELECT
  (SELECT default_config #>> '{workflow,steps,create_items_loop,next_step}'
     FROM agent_definitions WHERE type='tool-suggester' AND is_active) AS suggester_next,
  (SELECT default_config #> '{workflow,steps,create_cross_links}'
     FROM agent_definitions WHERE type='tool-suggester' AND is_active) AS cross_links_step,
  (SELECT default_config #>> '{workflow,steps,deploy_tool,config,related_pages}'
     FROM agent_definitions WHERE type='tool-deployer' AND is_active) AS deployer_related_pages,
  (SELECT default_config #>> '{workflow,steps,save_tool,config,related_pages}'
     FROM agent_definitions WHERE type='tool-generator' AND is_active) AS generator_related_pages;
-- Expected: complete | NULL | input_data.spec.related_pages | input_data.spec.related_pages

-- 2. No NEW phantom cross-link rows after this point. Every cross-link item
--    carries its tool page URL in the spec now, so the sweep is exact:
SELECT s.domain, swi.status, swi.created_at::date,
       swi.spec->>'tool_function' AS tool_function,
       swi.spec->>'tool_page_url' AS spec_url,
       p.url AS matched_page_url, p.build_status
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
LEFT JOIN pages p ON p.site_id = swi.site_id AND p.url = swi.spec->>'tool_page_url'
WHERE swi.item_key LIKE 'tool_crosslink:%'
ORDER BY swi.created_at DESC;
-- Expected for rows created after the image roll: matched_page_url non-NULL.
-- Rows predating the fix have no spec.tool_page_url — they are the existing
-- damage, cleaned up separately (see bugs_open/029).
