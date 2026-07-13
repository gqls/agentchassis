# HANDOFF — 2026-06-09: sections-durability shipped (code written), next = gamesdesign content quality

Resume point for a fresh chat. Site: **gamesdesign.co.uk** (adoption clone of gamedesign.uk). Platform: Go agent-chassis, Kafka saga agents, Postgres `clients_db`. Namespaces: `ai-persona-system` (app), `kafka`. Deploy: code → GitHub → GitHub Actions → Backblaze S3 for the *site*; the *chassis* is a container image rolled in `ai-persona-system`.

---

## Standing rules (carry forward — unchanged)
- **Snapshot before any DB change** (`CREATE TABLE IF NOT EXISTS <tbl>_bak_<tag> AS SELECT …`; short names, NAMEDATALEN 63; re-run of `IF NOT EXISTS … AS SELECT` is a no-op, don't trust re-run counts).
- **Check schema before writing SQL — request a FRESH `\d`** rather than trusting snapshots.
- **Reuse before creating** (STEP ZERO: search `agent_definitions`, registry, `discovery_checks/`, existing actions). Complexity in Go actions; workflows/prompts thin. No `logger.Debug`. Don't rename vars/fields (note explicitly if you must).
- **Don't conclude from partial signals** — verify the decisive fact before prescribing. Don't take a 0-row SQL result as decisive before checking the query isn't at fault.
- **`site_id` changes on every teardown** → resolve via `(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')`.
- Every agent is an orchestrator. Spawn sub-agents (own workflow) rather than sub-workflows in SQL. Keep workflow var names in sync with action expectations.
- Division of labour: assistant reads `/mnt/project` + `/mnt/user-data/uploads`; **user runs all SQL, kubectl, builds, deploys**. No Go toolchain in assistant env (validate by brace/paren balance only).
- No "final/last fix"; avoid "perfect/critical/excellent"; pragmatic, no congratulation.

---

## RESOLVED / WRITTEN this session (2026-06-08 → 09)

**guide-skinner-box: built and deployed.** Root was a single-page sectionless residue: the page was in the current plan with zero `site_plan_sections` and empty `pages.sections`, so the build read zero sections and short-circuited. Manual fix applied (set `pages.sections=["hero","generic-text-block"]` + mirror `site_plan_sections` + re-issue `needs_content_page`); verified `page_components`=2, `build_status='deployed'`.

**Decisive code facts established (from `production_page-build-debug_context.txt`):**
- `load_page_sections_from_spec` reads `site_specs.site_plan` (absent for this site) → falls back to `pages.sections`. **`pages.sections` is the build-read field; `site_plan_sections` is NOT on the build path.**
- The page-emission writers are each correct, including the convergence union (`normaliseRealisedToPlanPage` carries sections). A page arrives sectionless only when the value it faithfully carries is already empty. **Do NOT patch the union.**

**Durability code WRITTEN this session (NOT yet deployed) — see runbook below:**
- **2b** — `load_page_sections_from_spec_action.go`: final fallback synthesises a layout from a **same-role sibling** when `site_specs.site_plan` and `pages.sections` are both empty; writes to `pages.sections`; WARN-logged; modal-layout pick.
- **S1** — `discovery_checks/check_sectionless_pages.go` (new check, self-registers): flags planned pages with empty `pages.sections` that a same-role sibling can fix, re-triggers `needs_content_page` to `page-build-handler`. Enable by adding `"sectionless_pages"` to completeness-discovery-agent `default_config {workflow,steps,run_checks,config,checks}`.
- **S2** — page-build-handler workflow def: new `mark_no_sections` step (`fail_work_item` → `needs_human_review`); repoint `check_has_ready_sections.else_step` from `complete_error` → `mark_no_sections`.
- **Fix A** — `load_work_item_actions.go` `CompleteWorkItemAction`: guard so it won't overwrite a deliberate flagged/terminal status (`needs_human_review`, `failed`, …). **Prerequisite for S2** (and makes the existing `mark_needs_review` effective). Confirmed needed: the 2026-06-06 skinner-box retry hit `complete_error` and ended `complete`, proving `complete_error → dispatch mark_complete` fires; the unconditional `complete_work_item` would clobber a just-set flag.

**FOCUS silent-completion modes — RE-VERIFIED as already fixed in current code** (April doc was stale):
- Mode 1 (reaper lost-response): the `claimed-item-timeout` scheduled task now auto-completes ONLY with positive evidence (`page_components` updated_at > claimed_at; `deployed_at > claimed_at`); else resets.
- Mode 3 (claim-timeout): reaper resets to `triaged`/`failed` with `attempt_count+1`, not complete.
- Mode 2 (validate_content): errors on blockers → `mark_needs_review` → `needs_human_review`.
- Monitor query = 0 rows → no current silent-completion residue.
- `FOCUS_page_build_handler_silent_completion.md` updated with all of the above; Fix A marked applied; Fix B (`complete_error` reads as success on genuine-failure paths) deferred, low urgency.

---

## IMMEDIATE on resume — deploy + verify the pending changes
Use `RUNBOOK_section_sectionless_durability.md` (the standalone section; merge into the real page-build deploy runbook if one exists — note `RUNBOOK_phase_b_c_d_deploy_5_.md` is the *training/checkpoint* runbook, wrong domain). Order:
1. Confirm fresh `\d site_plan_pages` (`role`), `\d site_plan_sections` (`page_name,ordering,component_name`).
2. Place 3 Go files (2b, S1, Fix A) + roll the chassis image in `ai-persona-system`. **Fix A must ship with S2.**
3. Apply S2 (page-build-handler workflow-def UPDATE) + enable S1 (`sectionless_pages` into the checks array). Snapshot agent_definitions first.
4. Verify: blank a test guide's sections → re-trigger → WARN `SYNTHESISED layout from same-role sibling` + page builds; a no-sibling sectionless page lands `needs_human_review` not `complete` (S2+Fix A).

---

## NEXT TASK — gamesdesign content quality (the main feature work)
Source of record: `CATALOGUE_gamesdesign_post_sync_fix_defects` (latest version). Lead item first:

1. **Hero CTAs wrong site-wide (highest value — every page).** Every hero links both buttons to `/contact.html` and `/services.html`; CTA text↔destination mismatch (e.g. "Browse Tools" → `/contact.html`); **`/services.html` is a phantom page** (no such page). Investigate hero-component CTA resolution: the `hero` `content_components.input_schema` CTA fields (`cta_url`/`cta_text` source/fallback), `build_render_context` / `prepare_link_context` (internal-link/available-pages resolution), and where the LLM vs template sets the CTA. Decide: data-driven CTA resolution from real pages, not hardcoded defaults.
2. **Guide copy is tool-flavoured.** Guide heroes/CTAs describe simulators ("Launch RNG Simulator"). User is open to giving guides real embedded interactive demos if simple — check whether a guide page can host an interactive component the way tool pages do (see `019_tool_library`/`020_tool_lifecycle`).
3. **Polish batch.** Empty `href=""` on "Browse All X" buttons across the three list hubs (the `*_index_url` site_specs are unpopulated, and sources are inconsistent: `identity.*_index_url` for tool/game vs `navigation.*`/`blog.*` for guide/blog); `- GameDesign.uk` / `| GameDesign.uk` brand-suffix in card titles; one empty tool description (TTK card); empty footer brand-tagline and contact mailto/phone.

Approach per standing rules: read `CATALOGUE` + `003_contracts_and_standards` (component schema) first; STEP ZERO for any existing CTA/link resolver before writing; complexity in Go.

---

## LATER / parked
- **Fix B** (complete_error success-labelling on genuine-failure paths) — `FOCUS_page_build_handler_silent_completion.md`. Low urgency (monitor=0).
- **Render-off-`build_status` debt** — page-rerender skips planned-but-unrendered sections on a `deployed` page; proper fix = planned-vs-rendered `page_component` diff (current workaround: `build_status='needs_rebuild'` reset).
- **Tools/games behavioural QA loop** — `PLAN_tools_games_behavioral_qa_loop.md`.

---

## Key references
- **Build path:** page-build-handler workflow (`ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections → plan_sections → check_has_ready_sections → spawn/call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender → deploy_page → complete`). `complete_error` is a SUCCESS-labelled `complete_workflow` (Fix B target).
- **Dispatch:** build-dispatch-loop (`load_items → claim → spawn_handler → call_handler → mark_complete`); `mark_complete` = `complete_work_item` (Fix A guards it). Reaper = `claimed-item-timeout` scheduled_task (positive-evidence + reset).
- **Content/CTA:** page-content-writer workflow (`spawn_research → load_site_specs → prepare_link_context → build_render_context → process_sections_loop[render/generate] → compile_page`). It only produces content; persistence/deploy live in the page-build-handler wrapper.
- **Tables:** `pages(name,page_type,status,build_status,slug,url,sections jsonb,deployed_at)`; `site_plan_sections(plan_id,page_name,ordering,component_name)`; `site_plan_pages(plan_id,name,role,slug,url)`; `site_plans(is_current)`; `site_specs(aspect,data jsonb,is_current)`; `content_components(name,input_schema jsonb)`; `site_components(site_id,slot_name,component_id,rendered_html)`; `page_components(page_id,slot_name,component_id,rendered_html,updated_at)`; `site_work_items(... status,item_key,spec,handler_agent,attempt_count,max_attempts,claimed_at,completed_at,error)`; `workItemTerminalStatuses = complete,failed,verified,rejected,wont_fix,unresolved` (note: `needs_human_review` is NON-terminal).
- **Artifacts this session (outputs):** `running_notes_15_skinner_box_and_adoption_sections.md`, `load_page_sections_from_spec_action.go` (2b), `check_sectionless_pages.go` (S1), `load_work_item_actions.go` (Fix A), `RUNBOOK_section_sectionless_durability.md`, `FOCUS_page_build_handler_silent_completion.md` (updated).
