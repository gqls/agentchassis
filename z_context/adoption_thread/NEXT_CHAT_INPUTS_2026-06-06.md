# Inputs to gather for the next chat — gamesdesign.co.uk

Everything to attach / capture so a fresh chat can resume the build work (open thread: **guide-skinner-box won't build**) without re-deriving context.

---

## 1. Docs to re-attach as project knowledge (all already in this session's outputs)
- **`HANDOFF_2026-06-06_guide_list_and_skinner_box.md`** ← start here
- `running_notes_14_sync_fix_and_adoption_rerun.md` (working log, Parts 14a–14r)
- `016_debugging_guide_v2_33.md` (failure patterns)
- `FOCUS_adoption_faithfulness_via_locks.md` (convergence subsystem)
- `PLAN_tools_games_behavioral_qa_loop.md` (tools/games QA loop, future)

Optional, only if continuing the structural-debt work:
- `FOCUS_page_build_handler_silent_completion.md` (the silent-completion modes — directly relevant to skinner-box)
- `020_tool_lifecycle.md`, `019_tool_library.md`, `004_improvement_loop.md` (for the QA-loop plan)
- `029_*`, `030_*`, `031_*` plan/reconciler/locks docs (convergence/plan lineage)

---

## 2. Go / chassis code — what to ensure is attached
**(A) Pull FRESH from the chassis — these were NOT in the project files this session and are the next-root for skinner-box + the structural debt:**
- `load_page_sections_from_spec` (action) — **the decisive one**: what source it actually resolves "sections" from (claims `site_specs.site_plan`, which doesn't exist as an aspect for this site)
- `plan_sections` (action) — how `section_plan.ready_count` is computed
- `save_page_sections` (action)
- `page-content-writer` (agent definition) — the content specialist `call_content_writer` spawns
- `page-rerender` / `page_renderer` (agent definition + its assemble-and-deploy action) — the `build_status='deployed'` render short-circuit lives here
- The dispatch/reaper code: `build-dispatch-loop` + the **claim-timeout** path (mode 3: marks `complete` instead of resetting to `triaged`), `validate_page_content` routing (mode 2), and the reaper auto-complete (mode 1)
- The terminal-status actions: `complete_workflow`, `update_page_status`, `fail_work_item`

Best single artifact: a **current** dump of the chassis actions/agents (the up-to-date equivalent of `production_agent-chassis-actions-current_context.txt`) — it contains most of the above. The copy from this session is partial/possibly stale; get a fresh one.

**(B) Already in the project files — re-attach (reference, mostly for resolved convergence/list work):**
- `v3_site_actions.go` (ValidateSitePlanAction, reconcilePlanWithRealised, itemStemOf, normaliseRealisedToPlanPage)
- `write_site_plan_action.go`, `queryresolve.go` (resolvePagesWhereType), `reconcile_section_data_action.go`
- `site_db_actions.go` (upsertPage), `page_canonical.go`, `site_spec_actions.go`, `spawn_actions.go`, `coordinator.go`, `registry.go`, `action_inputs.go`, `safe_unmarshal.go`, `data_helpers.go`
- The `check_*.go` files (only if working the QA-loop plan)

---

## 3. Schema — capture FRESH `\d` (project snapshots `schemas_all`/`schemas_some` are stale)
```
\d site_work_items
\d site_plan_sections
\d site_plan_pages
\d site_plans
\d pages
\d page_components
\d site_specs
\d content_components
\d agent_definitions
\d agent_error_log
```
(Add `\d orchestration_states` and `\d awaited_requests` only if chasing the dispatcher/reaper.)

---

## 4. Live table content — snapshot current state into the new chat
First resolve the id once (it changes every teardown):
```sql
SELECT id FROM sites WHERE domain='gamesdesign.co.uk';
```

**Skinner-box (the open thread) — the decisive trio:**
```sql
-- a) did the section rows + pages.sections take?
SELECT page_name, ordering, component_name
FROM site_plan_sections
WHERE plan_id=(SELECT id FROM site_plans WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND is_current=true)
  AND page_name='guide-skinner-box' ORDER BY ordering;
SELECT name, page_type, status, build_status, sections
FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND name='guide-skinner-box';

-- b) the retry work item
SELECT created_at, item_type, status, claimed_by, attempt_count||'/'||max_attempts AS attempts, left(coalesce(error,''),90) AS error, item_key
FROM site_work_items
WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND (spec->>'page_name')='guide-skinner-box' ORDER BY created_at DESC;

-- c) does it have any rendered components yet?
SELECT slot_name, length(rendered_html) AS html_len, updated_at
FROM page_components
WHERE page_id=(SELECT id FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND name='guide-skinner-box')
ORDER BY slot_name;
```

**Current plan (so the new chat sees the section layout):**
```sql
SELECT spp.name AS page, sps.ordering, sps.component_name
FROM site_plans sp
JOIN site_plan_pages spp ON spp.plan_id=sp.id
LEFT JOIN site_plan_sections sps ON sps.plan_id=sp.id AND sps.page_name=spp.name
WHERE sp.site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND sp.is_current=true
ORDER BY spp.name, sps.ordering;
```

**Work-item queue (health snapshot):**
```sql
SELECT item_type, status, count(*) AS n
FROM site_work_items
WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
GROUP BY item_type, status ORDER BY status, item_type;
```

**Rendered state of index + hubs:**
```sql
SELECT p.name, pc.slot_name, length(pc.rendered_html) AS html_len, pc.updated_at
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND p.name IN ('index','guides-index','tools-index','games-index')
ORDER BY p.name, pc.slot_name;
```

**Confirm the cta_url fix persisted (resolved, but verify after any re-adoption):**
```sql
SELECT name, input_schema->'fields'->'cta_url'->>'required' AS cta_required
FROM content_components WHERE name ILIKE '%-list%' ORDER BY name;
```

**The handler workflows (the page-build-handler def is the key reference; grab page-rerender + content-writer too):**
```sql
SELECT type, default_config
FROM agent_definitions
WHERE type IN ('page-build-handler','page-rerender','page-content-writer');
```

**Recent errors for the site (skinner-box / build failures):**
```sql
SELECT occurred_at, agent_type, step_name, action, left(error_message,160) AS error, pod_name
FROM agent_error_log
WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
ORDER BY occurred_at DESC LIMIT 40;
```

---

## 5. Cluster / runtime (is the pipeline alive to pick up work items?)
```bash
kubectl -n ai-persona-system get pods | grep -iE 'dispatch|pipeline|trigger|sweep|build|rerender|content'
kubectl -n ai-persona-system get cronjobs
kubectl -n kafka get pods            # cluster: personae-kafka-cluster-...
```

---

## Minimum set to start fast
If you only grab a few things: the **HANDOFF** doc + **fresh `load_page_sections_from_spec` and `plan_sections`** code + the **skinner-box trio** queries (§4a–c) + `\d site_plan_sections`/`\d pages`/`\d site_work_items`. That's enough to settle the one open fork (are the section rows being read, or not) and decide the next move.
