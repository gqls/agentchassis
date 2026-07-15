# HANDOFF — gamesdesign.co.uk build, guide-list + skinner-box (2026-06-06)

Resume point for a fresh chat. Site: **gamesdesign.co.uk** (adoption clone of gamedesign.uk). Platform: Go "agent-chassis", Kafka saga agents, Postgres `clients_db`. Build cascade: adoption → build-site-planner → validate → write → sync → composition → page-build → rerender/deploy (git → Backblaze S3).

---

## Standing rules (carry into the new chat)
- **Snapshot before any DB change**: `CREATE TABLE IF NOT EXISTS <tbl>_bak_<tag> AS SELECT … WHERE …` (short names; NAMEDATALEN 63). Note: `CREATE TABLE IF NOT EXISTS … AS SELECT` is a no-op on re-run, so don't trust a re-run's snapshot/printed counts.
- **Check schema before writing SQL. Request a FRESH dump** (`\d <table>`) rather than trusting the project snapshots (`schemas_all`/`schemas_some` are stale `\d` lists) — stale dumps caused several wrong guesses.
- Reuse existing funcs/actions; don't create parallel ones. Complexity in Go actions, keep workflows/prompts thin. No `logger.Debug`. Don't rename vars/fields.
- **Don't conclude from partial signals.** Verify the decisive fact before prescribing. (This session: "guides built late", "wire reconcile_section_data", "page_type/status mismatch", and "just re-fire the rebuild" were all wrong until checked.)
- **`site_id` changes on every teardown** → always resolve via `(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')`.
- Division of labour: assistant reads files in `/mnt/project` + `/mnt/user-data/uploads`; **user runs all SQL, kubectl, builds, deploys**. No Go toolchain in assistant env (validate by brace/paren balance only). Namespaces: `ai-persona-system` (app) and `kafka`.

---

## RESOLVED this session
1. **Adoption convergence** (reconcile no-op) — RESOLVED earlier (Part 14n): the `[]map[string]interface{}` vs `[]interface{}` type assertion in `ValidateSitePlanAction` was fixed; clean re-adoption converged (5 guides as `role=guide`, zero bare siblings).
2. **guide-list empty on guides hub AND root index** — RESOLVED (Part 14p). Root cause: `guide-list_pre_037`'s `cta_url` field was `required=true` with no `on_missing` (defaults to `skip_field`), and `plan_sections`' required-field switch has no `skip_field` case → it fell to default-defer, holding the whole section. Siblings (`tool-list`/`game-list`) had `required=false`. Fix applied (DB, `content_components` is source of truth): set `cta_url.required=false` on **`guide-list_pre_037`** and **`blog-listing_pre_037`** (the other deviant). The list query (`pages_where_type:guide`) was never the problem.
3. **index render gap** — WORKED AROUND (Part 14q). A `needs_page` rebuild of an already-`build_status='deployed'` page completed with no error but never rebuilt the planned-but-missing `guide-list` component. Reset `pages.build_status='needs_rebuild'` + re-queue `needs_page` → full render; index + 4 guides now render the guide-list. Mechanism inferred to live in the spawned `page-rerender` agent (NOT code-verified).

Operational fact confirmed: **a manually-inserted `needs_page` / `needs_content_page` work item IS claimed by `build-dispatch-loop`** (status `triaged → claimed → complete`), so single-page (re)builds can be hand-triggered.

---

## OPEN — immediate thread: guide-skinner-box won't build

**State:** deployed git `guides/` has no `skinner-box/` directory; `pages.build_status='planned'`, `pages.sections=[]`, **zero** `site_plan_sections` rows for `guide-skinner-box`, **zero** `page_components`.

**How it got here (confirmed):** the original `needs_content_page` (2026-06-05 17:33) died `Claim timed out — handler pod likely died` and was marked **complete** (silent-completion mode 3). Content was never written and the page was left in `site_plan_pages` with **no** `site_plan_sections`. Nothing reconciles that.

**Why every rebuild since completes empty in ~90s (confirmed via the `page-build-handler` workflow def):** `load_spec_sections → plan_sections → check_has_ready_sections` (condition `section_plan.ready_count > 0`). With no sections, `ready_count=0` → ELSE → `complete_error`, which is a `complete_workflow` (SUCCESS) action with message *"Content writer skipped — page has no sections defined."* The content-writer is never spawned. (90s ≪ the 1200s the writer would take.)

**Fix prescribed (Part 14q) — apply, then VERIFY (do NOT re-fire blind):**
```sql
-- snapshots
CREATE TABLE IF NOT EXISTS sps_bak_skbox AS
SELECT * FROM site_plan_sections
WHERE plan_id=(SELECT id FROM site_plans WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND is_current=true);
CREATE TABLE IF NOT EXISTS pages_bak_skbox AS
SELECT * FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND name='guide-skinner-box';

-- 1) add 2 plan-section rows mirroring guide-rng-design exactly (hero@0, generic-text-block@1)
INSERT INTO site_plan_sections (id, plan_id, page_name, ordering, component_name, created_at)
SELECT gen_random_uuid(), sp.id, 'guide-skinner-box', v.ordering, v.component_name, now()
FROM site_plans sp, (VALUES (0,'hero'),(1,'generic-text-block')) AS v(ordering, component_name)
WHERE sp.site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND sp.is_current=true
ON CONFLICT DO NOTHING;

-- 2) set pages.sections to match
UPDATE pages SET sections='["hero","generic-text-block"]'::jsonb
WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk') AND name='guide-skinner-box';

-- 3) re-issue the content write
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, page_id, priority, handler_agent, status, created_by, spec, item_key)
SELECT s.id,'manual-rebuild','build','needs_content_page','medium','Write content for guide-skinner-box — sections added to plan',
       p.id,90,'page-build-handler','triaged','manual-rebuild',
       jsonb_build_object('mode','recreate','source','adoption','page_name','guide-skinner-box','page_type','guide'),
       'needs_content_page:guide-skinner-box-retry2'
FROM sites s JOIN pages p ON p.site_id=s.id AND p.name='guide-skinner-box'
WHERE s.domain='gamesdesign.co.uk' ON CONFLICT DO NOTHING;
```

**Verify (the decisive fork — see Part 14r):**
- Section rows present? (`site_plan_sections` for the page) and `pages.sections` set?
- Retry item `needs_content_page:guide-skinner-box-retry2` status; `page_components` for the page; `build_status`.
- **If rows ABSENT** → the INSERT/UPDATE wasn't run; run it then re-issue.
- **If rows PRESENT but the item completes ~90s/empty AGAIN** → the relational `site_plan_sections` + `pages.sections` are NOT the source `plan_sections` reads for readiness. **Next root to chase:** request fresh `load_page_sections_from_spec` + `plan_sections` action code from the chassis (not on disk) and find the actual source it resolves "sections" from + how `ready_count` is computed. Its stated primary `site_specs.site_plan` does **not** exist as an aspect for this site (aspects: briefing, classification, content_direction, design_intent, design_reference, identity, resolved_composition, site_archetype, strategy, structure; `structure.data.pages` is a flat name list with no per-page sections).
- **Success looks like:** item runs >90s, `page_components` non-empty, `build_status='deployed'`, skinner-box card gains its description.

(Reference working sibling — `guide-rng-design` `site_plan_sections`: `hero`@0, `generic-text-block`@1; `component_version_id`/`palette_id`/`layout_id`/`typography_set_id` all NULL; current plan_id observed `c96b501c-…` but resolve via `is_current`.)

---

## OPEN — structural debt surfaced this session (note, fix deliberately deferred)
1. **Render decision off `build_status`** — the `page-rerender` agent appears to skip rebuilding planned-but-unrendered sections on a `deployed` page. Proper fix: drive render off a planned-vs-rendered diff (does every planned section have a current `page_component`?), not `build_status`. Until then the `build_status='needs_rebuild'` reset is the workaround.
2. **Sectionless page completes as SUCCESS** — `check_has_ready_sections` ELSE → `complete_error` → `complete_workflow` (success message "Content writer skipped — page has no sections defined"). A sectionless page should be a distinct non-terminal/flagged state, never `complete`.
3. **No reconciliation of "page in plan, zero sections"** — same family as the empty-guide-list and the render short-circuit: a partial failure that no later pass repairs. A planner/convergence guard should re-plan or flag a sectionless page.
4. **Silent-completion modes** (`FOCUS_page_build_handler_silent_completion.md`): claim-timeout (mode 3) marks items `complete` instead of resetting to `triaged`+retry; `validate_content` routes inconsistently; reaper auto-completes on lost responses. These are what leave the partial-failure residue (skinner-box is a direct consequence).

These are a coherent mini-project: **"complete" is being used as "we stopped" rather than "the work succeeded."** Worth fixing together with positive-evidence checks (`pageHasComponents`, `pageIsDeployed`) before marking complete.

---

## OPEN — gamesdesign content quality (separate from the build mechanics above)
- **Hero CTAs wrong site-wide**: every hero links both buttons to `/contact.html` and `/services.html` regardless of page; text↔destination mismatch (e.g. "Browse Tools" → /contact.html); **`/services.html` is a phantom page**. Hero-component CTA resolution. High value (every page).
- **Guide copy is tool-flavoured**: guide heroes/CTAs describe simulators ("Launch RNG Simulator", "Open the PRD Calculator"). User is **open to giving guides real embedded interactive demos** if simple — worth checking whether a guide page can host an interactive component the way tool pages do, rather than just rewriting copy.
- **Polish batch**: empty `href=""` on the "Browse All X" buttons across all three list hubs (the `*_index_url` site_specs they source from are unpopulated; sources are inconsistent too — `identity.*_index_url` for tool/game vs `navigation.*`/`blog.*` for guide/blog); `- GameDesign.uk` / `| GameDesign.uk` source-brand suffix in card titles; one empty tool description (TTK card); empty footer brand-tagline and contact mailto/phone.

---

## FUTURE — tools/games behavioral QA loop
`PLAN_tools_games_behavioral_qa_loop.md` (this session) — a standalone QA/maintenance loop that builds out the planned-but-unbuilt Tier 3 (headless behavioral testing) and adds a games lifecycle. Motivated by real defects: Jelly Invaders degrades over time (skips rows / erratic), P2P host replies don't reach a mobile client (relay-matrix failure), and untested mobile/variants/cross-browser. Phased; first cut Phase 0+1.

---

## Docs (all in outputs; bring forward as project knowledge)
- `running_notes_14_sync_fix_and_adoption_rerun.md` — full working log, Parts 14a–14r (14n convergence, 14p cta_url, 14q index gap + skinner-box, 14r end state).
- `016_debugging_guide_v2_33.md` — failure patterns incl. the two new ones (rebuild-of-deployed-page doesn't refresh components; sectionless page completes as success) and the earlier list-section-deferral + convergence ones.
- `FOCUS_adoption_faithfulness_via_locks.md` — convergence subsystem, marked verified-resolved.
- `PLAN_tools_games_behavioral_qa_loop.md` — the QA-loop plan.

## Key references
- **page-build-handler workflow** (agent_definitions): `ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections → plan_sections → check_has_ready_sections → spawn_content_writer → call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender_agent → deploy_page → complete`. `complete_error` is a `complete_workflow` (success-labelled) — the smell. `call_content_writer` timeout 1200s; whole workflow 1800s.
- **Tables**: `site_plan_sections(plan_id,page_name,ordering,component_name,…; UNIQUE(plan_id,page_name,ordering))`; `site_plans(is_current)`; `site_plan_pages(plan_id,name,role,slug,url)`; `pages(name,page_type,status,build_status,sections jsonb; NO lock cols)`; `site_specs(aspect,data jsonb,is_current; UNIQUE(site_id,aspect) where is_current)`; `content_components(name UNIQUE, input_schema jsonb → fields.<f>.{source,required,on_missing,fallback})`; `site_work_items` (cols incl site_id,source,pipeline,item_type,severity,summary,page_id,priority,handler_agent,status,created_by,spec,item_key; partial UNIQUE(site_id,item_key) where item_key NOT NULL and status not terminal); `page_components(page_id,slot_name,component_id,rendered_html,updated_at,…)`; `agent_error_log(site_id,work_item_id,agent_type,step_name,error_message,context,…)`.
- **Manual rebuild item shapes**: re-render existing components → `item_type='needs_page'`, `handler_agent='page-build-handler'`, spec `{reason,page_name}`. Generate content → `item_type='needs_content_page'`, same handler, spec `{mode:'recreate',source:'adoption',page_name,page_type}`. Both `status='triaged'`, `ON CONFLICT DO NOTHING`.
