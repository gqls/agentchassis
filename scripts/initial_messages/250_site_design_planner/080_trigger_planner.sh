Option 1 — Smallest-possible trigger: one site, one composition item
Picks a single non-locked site that currently has style_collection_id = NULL, queues a needs_composition item for it, lets the dispatch loop handle it. If composition completes, queue a needs_design item with the depends_on gate.

-- Pick a target site (the sweep left several non-locked sites with no collection)
SELECT id, domain, locked_at, style_collection_id, build_status
  FROM sites
 WHERE locked_at IS NULL
   AND style_collection_id IS NULL
 ORDER BY domain;

-- Pick one. For the gamedesign.uk case that's been the test target through
-- both chats, substitute that id below. Otherwise pick any non-locked site.

-- Queue needs_composition for the chosen site
WITH target AS (
    SELECT id FROM sites
     WHERE domain = '<CHOSEN-DOMAIN>'
       AND locked_at IS NULL
)
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT id, 'manual', 'build', 'needs_composition', 'high',
       'Manual trigger: exercise merged composition pipeline',
       '{}'::jsonb, 7, 'site-design-planner', 'triaged',
       'deployment-smoke-test',
       'smoke_composition_' || to_char(NOW(), 'YYYYMMDDHH24MI')
  FROM target
  ON CONFLICT DO NOTHING
  RETURNING id, site_id, item_type, status, priority;


kubectl -n ai-persona-system logs -l app=agent-chassis --since=2m \
  | grep -E 'site-design-planner|validate_composition|resolve_composition|install_site_composition'

  -- Work item status
  SELECT status, attempt_count, LEFT(error, 200) AS error,
         LEFT(result::text, 300) AS result
    FROM site_work_items
   WHERE item_key LIKE 'smoke_composition_%'
   ORDER BY created_at DESC LIMIT 1;

  -- Site should now be linked
  SELECT s.domain, s.style_collection_id,
         sc.name AS collection_name,
         ct.name AS theme_name
    FROM sites s
    LEFT JOIN style_collections sc ON sc.id = s.style_collection_id
    LEFT JOIN css_themes ct ON ct.id = sc.css_theme_id
   WHERE s.domain = '<CHOSEN-DOMAIN>';




  -----------------
page builder workflow
"spawn_handler": {
    "action": "spawn_agent",
    "config": {
        "role": "handler",
        "error_step": "mark_failed",
        "agent_type_field": "current_item.handler_agent"   ← dynamic
    },
    "next_step": "call_handler"
},
"call_handler": {
    "action": "call_agent",
    "config": {
        "target_role": "handler",
        "input_mapping": {
            "spec": "current_item.spec",
            "site_id": "current_item.site_id",
            ...
        },
        "timeout_seconds": 1200
    }
}
-------------------------

Option 2 — Full build path for one site: composition + design gated by depends_on
Extends Option 1 by also queuing needs_design that depends on the composition item. Tests the full gated-dispatch flow that WriteBuildItemsAction normally produces.

-- Do Option 1 first, capture the returned id as <COMPOSITION-ITEM-ID>
-- Then queue a gated needs_design:

WITH target AS (
    SELECT id FROM sites WHERE domain = '<CHOSEN-DOMAIN>'
)
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key,
    depends_on
)
SELECT id, 'manual', 'build', 'needs_design', 'high',
       'Manual trigger: gated design after composition',
       '{}'::jsonb, 8, 'webdesign-agent', 'triaged',
       'deployment-smoke-test',
       'smoke_design_' || to_char(NOW(), 'YYYYMMDDHH24MI'),
       ARRAY['<COMPOSITION-ITEM-ID>']::uuid[]
  FROM target
  ON CONFLICT DO NOTHING
  RETURNING id, depends_on;


  kubectl -n ai-persona-system logs -l app=agent-chassis --since=2m \
    | grep -E 'site-design-planner|webdesign-agent|render_css_from_spec'


------------------------------------

Option 3 — Let the scheduler drive it: normal queue path
No manual INSERTs. Put a domain in build_queue and let build-pipeline-trigger (runs every 120s) seed it. This exercises the full seed_build_queue → WriteBuildItemsAction → dispatch path, which is what real builds use.
Use this option once Option 1 and 2 have passed.

-- For a fresh build from scratch (no prior specs):
INSERT INTO build_queue (
    domain, direction, priority, status, source
) VALUES (
    '<NEW-DOMAIN>.com', NULL, 10, 'queued', 'smoke-test'
);

-- For an adoption:
INSERT INTO build_queue (
    domain, direction, priority, status, source
) VALUES (
    '<DOMAIN-TO-ADOPT>.com',
    '{"adopt_from": "<DOMAIN-TO-ADOPT>.com"}'::jsonb,
    10, 'queued', 'smoke-test'
);

The build-pipeline-trigger scheduled task picks this up at its next tick. Watch:

-- Queue entry should move queued → seeded within ~2min
SELECT id, domain, status, LEFT(error, 200) AS error, updated_at
  FROM build_queue
 WHERE source = 'smoke-test'
 ORDER BY created_at DESC LIMIT 5;

 -- Once seeded, work items should appear for the site
 SELECT wi.item_type, wi.status, wi.priority,
        wi.depends_on IS NOT NULL AS has_deps,
        wi.created_at
   FROM site_work_items wi
   JOIN sites s ON s.id = wi.site_id
  WHERE s.domain = '<DOMAIN>'
    AND wi.source IN ('manual', 'seed_build_queue', 'write_build_items')
  ORDER BY wi.created_at DESC, wi.priority;

  -----------------

  For a fresh build, expect: needs_domain_research → (classifier adds specs) → needs_logo + needs_hero_image + needs_composition → needs_design (with depends_on) → needs_content_page × N → needs_rerender.

  Healthy signs

  site-design-planner pod spawns, completes within ~30s
  Composition items move triaged → claimed → complete
  sites.style_collection_id populates with a new UUID
  css_themes and style_collections get new adopted-origin rows referencing the site
  NO emergency_fallback log lines in agent-chassis logs
  NO logger.Error lines about install_on_site stale config

  Likely failure modes and what to capture
  "validate_composition_inputs failed: identity/classification missing"
  The site doesn't have classification yet. The recovery item (needs_domain_research) should auto-queue. Paste the work item row + the queued recovery item.
  site-design-planner succeeds but renderer emergency-falls-back
  Means composition landed but webdesign-agent didn't see it in site_context. Paste:

  SELECT * FROM sites WHERE domain = '<X>' (confirms collection is linked)
  kubectl logs from the webdesign-agent pod around the render time
  The orchestration_states row for the webdesign-agent run

  Composition item completes but style_collection_id stays NULL
  The guarded UPDATE lost the race or the transaction rolled back silently. Paste:

  The work item's result field (install_result should be there)
  SELECT id, name, origin, source_site_id, created_at FROM style_collections WHERE source_site_id = '<site-id>' ORDER BY created_at DESC LIMIT 5
  Search for "install_site_composition" in pod logs

  install_theme fires (shouldn't exist post-merge)
  The SQL didn't fully delete the step. Paste:

  SELECT default_config -> 'workflow' -> 'steps' ? 'install_theme' FROM agent_definitions WHERE type = 'webdesign-agent' AND is_active = true
  Any relevant kubectl log line

  Stale install_on_site warning in logs
  A workflow config still references the deprecated key. Paste the exact log line — it'll name the step_name.

