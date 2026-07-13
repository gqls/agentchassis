# Build Pipeline Handoff Summary
## Date: 2026-02-24

## What We're Building
A work-item-driven website build pipeline. Given a domain name, the system researches, classifies, briefs, plans, and builds a multi-page website through a chain of specialist agents coordinated by a dispatch loop.

## Pipeline Flow
```
API/trigger sends domain name
  → seed_build_queue action creates site record + initial work item (needs_domain_research, priority 5)
  → triggers build-dispatch-loop

dispatch loop loads highest-priority pending item, claims it, calls handler:

  1. domain-research-classifier (handles needs_domain_research)
     → web search → scrape → LLM classify → write_site_spec (identity + classification)
     → creates needs_briefing item (priority 10, handler: build-briefing-agent)

  2. build-briefing-agent (handles needs_briefing)  
     → read_site_spec → fetch_agent_questionnaire → LLM answers → write_site_spec (briefing)
     → creates needs_site_plan item (priority 15, handler: site-planner)

  3. site-planner (handles needs_site_plan) — needs handler-mode adaptation
     → would create needs_content_page × N items

  ...and so on through content writing, assembly, deployment
```

## Completed Artifacts

### Block A-B: Database Migrations + Actions
All committed and working:
- **Migrations**: site_specs table, build_queue additions to site_work_items, page_component_history table
- **Actions**: write_site_spec, read_site_spec, save_component_history, seed_build_queue, load_work_items (existing), complete_work_item (existing), fail_work_item (existing)

### Block D: Agent Definitions (SQL files created, not yet applied)

1. **domain-research-classifier** (`/home/claude/domain_research_classifier.sql`)
    - 7-step flat workflow: search → scrape → LLM classify → write identity spec → write classification spec → create next work item → complete
    - Handler for: needs_domain_research items
    - Creates: needs_briefing → build-briefing-agent

2. **build-dispatch-loop** (`/home/claude/build_dispatch_loop.sql`)
    - Flat workflow, NO loops, NO sub_workflows
    - One item per invocation: load → check → claim → spawn handler → call handler → mark complete → check remaining → spawn self if more
    - Self-spawning chain: each invocation is separate orchestration with clean logs
    - Uses `first_item` convenience field from load_work_items (requires patch)

3. **build-briefing-agent** (`/home/claude/build_briefing_agent.sql`)
    - Named differently from existing `briefing-agent` (v1) which intake-orchestrator still uses
    - 5-step: read_site_spec(all) → fetch_agent_questionnaire(dynamic) → LLM answer → write_site_spec(briefing) → create_work_item(needs_site_plan)

### New Go Actions (created, not yet committed)

4. **claim_work_item** (`/home/claude/claim_work_item_action.go`)
    - Atomic claim: `UPDATE ... WHERE status IN ('triaged','approved') RETURNING id`
    - Returns `{claimed: bool, work_item_id, reason}` — safe for concurrent dispatch
    - Registry entry needed: category "site", IsLocal true

5. **load_work_items patch** (`/home/claude/load_work_items_patch.go`)
    - Backward-compatible: adds `first_item` field to output map
    - Needed because FindByPath doesn't support array indexing (items.0.id won't resolve)
    - Existing callers unaffected — they use .items and .has_items

### Registry Additions Still Needed
All Block B+D actions need registry entries in registry.go + local_actions.go:
```go
// registry.go
"write_site_spec":          {Handler: WriteSiteSpecAction, Category: "site", ...}
"read_site_spec":           {Handler: ReadSiteSpecAction, Category: "site", ...}
"save_component_history":   {Handler: SaveComponentHistoryAction, Category: "site", ...}
"seed_build_queue":         {Handler: SeedBuildQueueAction, Category: "site", ...}
"create_work_item":         {Handler: CreateWorkItemAction, Category: "site", ...}
"claim_work_item":          {Handler: ClaimWorkItemAction, Category: "site", ...}
```

## Key Design Decisions

1. **No sub_workflows** — they've been problematic. Dispatch loop uses self-spawning instead.
2. **No loops in dispatch** — one item per invocation, self-spawns for next item. Clean logs per item.
3. **site_specs as shared state** — agents write aspects (identity, classification, briefing) to site_specs table. Downstream agents read what they need.
4. **Work items as pipeline glue** — each handler creates the next work item with handler_agent set. Priority ordering handles sequencing.
5. **claim_work_item for concurrency** — atomic claim prevents double-dispatch if multiple dispatch loops run.
6. **Existing briefing-agent preserved** — new one named `build-briefing-agent` to avoid conflict with intake-orchestrator pipeline.

## Open Investigation (where we stopped)

We were investigating **how the loop action handles array iteration** — specifically how `item_variable` (e.g., `current_page`, `current_item`) gets injected into collectedData per iteration. The question was whether we need the `first_item` patch at all, or if the dispatch loop could use a loop pattern instead.

Key findings so far:
- LoopAction returns an `expansion` map with `loop_action: true`
- `handleLoopExpansion` in SagaCoordinator injects steps named `{loopname}_iter_{N}_{substep}`
- Each injected step gets `loop_iteration`, `loop_item_index`, `loop_var_name` in its config
- There's a `setLoopVariable` function that reads these config values and injects the item
- The loop DOES handle arrays — pageflow-builder iterates `pages_to_build.pages` with `item_variable: "current_page"` and sub_workflow steps reference `current_page.name`, `current_page.id` etc.

**However**: The user explicitly doesn't want sub_workflows in the dispatch loop. The self-spawning flat pattern is the chosen approach. The `first_item` patch is still needed for that pattern.

## Files Location
All work files are in `/home/claude/`:
- domain_research_classifier.sql
- build_dispatch_loop.sql
- build_briefing_agent.sql
- claim_work_item_action.go
- load_work_items_patch.go

Copies also in `/mnt/user-data/outputs/`.

## What's Next
1. Apply the load_work_items patch (add first_item to output)
2. Add all registry entries
3. Commit claim_work_item_action.go
4. Apply the 3 agent definition SQLs to database
5. Test the pipeline end-to-end: trigger seed_build_queue with a domain → watch dispatch loop chain through research → briefing → (site plan)
6. Adapt site-planner for handler mode (reads from site_specs instead of receiving data via input_mapping)
7. Eventually: adapt pageflow-builder's page loop to work as dispatch-loop handlers too

## Key Architecture References
- Agent creation guide: `/mnt/project/001b_development_guide_new_agents_v2.md`
- System architecture: `/mnt/project/002b_system_architecture_v2.md`
- Contracts/standards: `/mnt/project/003_contracts_and_standards.md`
- Existing agent definitions: `/mnt/project/agent_definitions_backup.sql`
- Chassis code: `/mnt/project/production_agent-chassis-full_context.txt`
- API adapters: `/mnt/project/production_agent-adapters_context.txt`
- Previous transcripts: `/mnt/transcripts/journal.txt`
