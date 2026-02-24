This is the minimum to build a site via work items instead of the pageflow-builder monolith. All existing agents and workflows remain untouched.

Files the new chat needs loaded as context:

005_build_expand_market_plan.md (the plan we just wrote)
000_b_schemas (current DB schemas)
001_development_guide_new_agents.md (agent creation guidelines)
002_system_architecture.md (architecture doc)
003_contracts_and_standards.md (contracts)
agent_definitions_backup.sql (current agent definitions for reference)
production_agent-chassis-full_context.txt (Go codebase — action patterns, dispatch loop)
revert_004_005.sql (needs applying first if 004/005 were deployed)


Step order (dependency-driven):
Block A — DB migrations (no Go changes needed, can apply immediately)

site_specs table — new table, no dependencies on anything existing
build_queue table — new table
page_component_history table — new table, references page_components and pages
approval_mode column on site_work_items — ALTER TABLE, default 'auto'
page_spec column on pages — ALTER TABLE, nullable jsonb
Drop content_snapshot / schema_snapshot from page_components — check nothing reads these first, then ALTER TABLE DROP COLUMN

Block B — Core Go actions (needed by everything else)

write_site_spec — read current → deep merge → mark old not current → insert new. Used by every handler that writes specs. Needs to work as both a workflow action and a callable Go function.
read_site_spec — simple query wrapper: read one aspect or all current aspects for a site. Also both action and function.
save_component_history — called before any content_data write to page_components. Copies current value to page_component_history. Integrate into existing UpdateComponentContentAction or whatever writes content_data currently.
check_approval_mode — in the dispatch loop, before spawning handler: if approval_mode = 'hitl' and status is not 'approved', set status to 'pending_review' and skip. Small change to existing dispatch logic.
depends_on ordering in dispatch loop — currently processes by priority only. Need: before dispatching an item, check all UUIDs in its depends_on array are in completed status. If not, skip. Check what the current dispatch loop does with depends_on — it may already handle this.

Block C — Seed pipeline

seed_build_queue action — reads from build_queue, for each domain: calls ensure_site_record, examines direction field, writes initial site_specs if direction contains spec data, inserts first work item based on direction type (null → needs_domain_research, adopt_from → needs_site_adoption, fork_from → copy specs + needs_site_plan, brief_complete → needs_site_plan), marks status = 'seeded'. Needs write_site_spec from Block B.

Block D — New agents (each is an independent agent definition, can be built in parallel)

domain-research-classifier — new agent, not modifying existing site-classifier. Orchestrator workflow: spawn research-agent → call research → classify+profile (Sonnet LLM) → write specs via write_site_spec → complete. Output: identity, strategy, tone, visual_direction, image_guidance written to site_specs. This is the enhanced classifier we designed earlier but as a new agent.
briefing-agent as handler — the existing briefing-agent may need a v2 or adapter mode where it receives a work item context (site_id, reads specs from DB) rather than being called inline by an orchestrator with input_mapping. Check the current briefing agent workflow to see how much changes. It writes refined identity and tone back via write_site_spec.
site-planner handler mode — new agent or new workflow for existing planner. Reads specs from DB (identity, tone, visual_direction, image_guidance, briefing output). Loads available components and styles (existing actions). Runs LLM planning prompt. Output: instead of returning JSON to caller, writes page records to pages table AND creates downstream work items in site_work_items (needs_logo, needs_hero, needs_design, needs_content_page × N, needs_nav, needs_sitemap). Writes structure spec. This needs a new Go action: write_plan_as_work_items.
snapshot-agent — new agent. Simple workflow: read_all_current_specs → commit_spec_file to git. Needs commit_spec_snapshot Go action (read site_specs, marshal JSON, git commit as .site-spec.json).

Block E — Dispatch loop integration

Side-effects snapshot triggers — after certain item types complete (needs_site_plan, needs_design), the side-effects check creates a needs_snapshot work item. Also: when no pending items remain for a site, create needs_snapshot with item_key dedup for build_complete.


What to check in the Go codebase early (saves time):

How does the dispatch loop currently iterate work items? Does it already check depends_on?
What action currently writes page_components.content_data? That's where save_component_history hooks in.
How does ensure_site_record work? The seed action reuses it.
What does the side-effects check currently look like? We're adding snapshot triggers to it.
How do existing handlers complete? They need to write specs via write_site_spec as part of their completion — is there a hook point?


Test scenario for end-to-end validation:

Insert into build_queue: finetuning.uk with {"objective": "AI model fine-tuning consultancy"}
Run seed → creates site record + needs_domain_research work item
Dispatch → domain-research-classifier runs → writes identity, strategy, tone, visual_direction, image_guidance to site_specs → completes → creates needs_snapshot (plan not done yet, but research done)
Dispatch → briefing-agent runs → refines identity/tone → completes
Dispatch → site-planner runs → writes structure spec + page records + downstream work items → completes → side-effect creates needs_snapshot (plan_complete)
Dispatch → snapshot-agent commits .site-spec.json to git
Dispatch → image-generator (logo), image-generator (hero) → complete
Dispatch → webdesign-agent → writes design spec → completes → needs_snapshot (design_complete)
Dispatch → content writers × N → complete (page_component_history populated for future rollback)
Dispatch → nav-agent, sitemap-agent → complete
No pending items → needs_snapshot (build_complete) created → snapshot-agent runs
Site is live, all specs in DB, all milestones in git


Estimated build order for a single developer:
OrderWhatDepends onRough effort1Block A: all migrationsNothingHalf day2write_site_spec + read_site_specMigrationsHalf day3check_approval_mode + depends_on in dispatchRead codebaseHalf day4save_component_historyMigrations, find the write pointHalf day5seed_build_queuewrite_site_spec, ensure_site_recordHalf day6domain-research-classifier agentwrite_site_spec, research-agent exists1-2 days7briefing-agent handler modewrite_site_spec, read_site_spec1 day8site-planner handler mode + write_plan_as_work_itemsread_site_spec, page creation1-2 days9snapshot-agent + commit_spec_snapshotread_site_spec, git actionsHalf day10Side-effects snapshot triggerssnapshot-agent, dispatch loopHalf day11End-to-end testEverything1 day
Roughly 7-9 days of focused work. The migrations and core Go actions (Block A+B) unblock everything else. The agents (Block D) can be built one at a time and tested independently.

