# Context pack — gamesdesign.co.uk adoption: guide-skinner-box won't build (fresh thread)

Starting context for the open build bug. Site **gamesdesign.co.uk** (adoption clone of gamedesign.uk); Go agent-chassis, Kafka saga agents, Postgres `clients_db`. Build cascade: adoption → build-site-planner → validate → write → sync → composition → page-build → rerender/deploy (git → Backblaze).

---

## State + next action

**Open thread:** `guide-skinner-box` won't build. The page is in the plan with **zero sections** (`pages.build_status='planned'`, `pages.sections=[]`, zero `site_plan_sections` rows, zero `page_components`). It got there when the original `needs_content_page` died ("claim timed out — handler pod likely died") and was marked **complete** with content never written, and nothing reconciles a sectionless page. Every rebuild since completes empty in ~90s because `page-build-handler` runs `load_spec_sections → plan_sections → check_has_ready_sections` (condition `ready_count > 0`); with no sections, `ready_count=0` → ELSE → `complete_error`, which is a **success-labelled** `complete_workflow` ("Content writer skipped — page has no sections defined"). The writer is never spawned.

**Next action — apply the prescribed fix, then VERIFY (do not re-fire blind):**
1. Snapshot, then add 2 plan-section rows mirroring a working sibling (`guide-rng-design`: `hero`@0, `generic-text-block`@1), set `pages.sections` to match, and re-issue a `needs_content_page` item. (Full SQL is in the source handoff `HANDOFF_2026-06-06_guide_list_and_skinner_box.md`.)
2. **The decisive fork:**
   - If the section rows are **absent** after the INSERT → the SQL didn't run; run it, re-issue.
   - If the rows are **present but the item still completes ~90s/empty** → the relational `site_plan_sections` + `pages.sections` are **not** the source `plan_sections` reads for readiness. Pull **fresh** `load_page_sections_from_spec` + `plan_sections` code (below) and find the actual source it resolves "sections" from and how `ready_count` is computed. Note: its claimed primary `site_specs.site_plan` does **not** exist as an aspect for this site (aspects: briefing, classification, content_direction, design_intent, design_reference, identity, resolved_composition, site_archetype, strategy, structure; `structure.data.pages` is a flat name list with no per-page sections).
   - Success looks like: item runs >90s, `page_components` non-empty, `build_status='deployed'`, skinner-box card gains its description.

## Standing rules (the constitution)

Snapshot before any DB change (short bak-table names; `CREATE TABLE IF NOT EXISTS … AS SELECT` is a no-op on re-run, so don't trust a re-run's counts). **Check schema with a fresh `\d` before SQL** (`schemas_all`/`schemas_some` are stale). Reuse existing funcs/actions; complexity in Go, workflows/prompts thin; no `logger.Debug`; don't rename vars/fields. **Don't conclude from partial signals — verify the decisive fact before prescribing.** `site_id` changes on every teardown → always resolve via `(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')`. Namespaces `ai-persona-system`, `kafka`.

## Attach — code

**Pull FRESH from the chassis (these were NOT on disk last session and are the next root):**
`load_page_sections_from_spec` (the decisive one — actual "sections" source), `plan_sections` (how `ready_count` is computed), `save_page_sections`, `page-content-writer` (agent def), `page-rerender`/`page_renderer` (def + its assemble-and-deploy action — the `build_status='deployed'` render short-circuit lives here), `build-dispatch-loop` + the claim-timeout path (marks `complete` instead of resetting to `triaged`), `validate_page_content` routing, the reaper auto-complete, and the terminal-status actions `complete_workflow`/`update_page_status`/`fail_work_item`. Best single artifact: a **current** dump of the chassis actions/agents.

**On disk (re-attach):** `v3_site_actions.go`, `write_site_plan_action.go`, `queryresolve.go`, `reconcile_section_data_action.go`, `site_db_actions.go`, `page_canonical.go`, `site_spec_actions.go`, `spawn_actions.go`, `coordinator.go`, `registry.go`, `action_inputs.go`, `safe_unmarshal.go`, `data_helpers.go`, and the `check_*.go` set if working the structural debt.

## Attach — docs

`HANDOFF_2026-06-06_guide_list_and_skinner_box.md` (start here), `running_notes_14_sync_fix_and_adoption_rerun.md`, `016_debugging_guide_v2_33.md`, `FOCUS_adoption_faithfulness_via_locks.md`, and `FOCUS_page_build_handler_silent_completion.md` (the silent-completion modes — directly relevant).

## Pull — schema (fresh `\d`)

```
dbcontext -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema site_work_items,site_plan_sections,site_plan_pages,site_plans,pages,page_components,site_specs,content_components,agent_definitions,agent_error_log
```
(Add `orchestration_states`,`awaited_requests` only if chasing the dispatcher/reaper.)

## Pull — live data (the decisive trio + health)

Resolve the id once, then (each as a `dbcontext -rows "…"`):
- Section rows + `pages.sections` for `guide-skinner-box` (did the fix take?).
- The retry work item status/attempts/error.
- `page_components` for the page (any rendered components yet?).
- Current plan layout (page → ordering → component_name).
- Work-item queue health (item_type, status, count).
- Rendered state of `index`/`guides-index`/`tools-index`/`games-index`.
- Recent `agent_error_log` for the site (last ~40).

(Exact SQL for all of these is in `NEXT_CHAT_INPUTS_2026-06-06.md` §4.)

## Capture — runtime

```
kubectl -n ai-persona-system get pods | grep -iE 'dispatch|build|rerender|content|pipeline|sweep'
kubectl -n ai-persona-system get cronjobs
```
(Is the pipeline alive to claim work items? A manually-inserted `needs_content_page`/`needs_page` IS claimed by `build-dispatch-loop`.)

## Structural debt (note; fix deliberately as a set, not piecemeal)

The theme: **"complete" is being used to mean "we stopped", not "the work succeeded."** (1) render decision off `build_status` instead of a planned-vs-rendered diff; (2) sectionless page completes as SUCCESS; (3) no reconciliation of "page in plan, zero sections"; (4) the silent-completion modes (claim-timeout → complete, validate routing, reaper auto-complete). Worth fixing together with positive-evidence checks (`pageHasComponents`, `pageIsDeployed`) before marking complete. The skinner-box failure is a direct consequence.

## Minimum set to start fast

The HANDOFF + **fresh `load_page_sections_from_spec` and `plan_sections`** code + the skinner-box trio queries + `\d site_plan_sections`/`pages`/`site_work_items`. Enough to settle the one fork: are the section rows being read, or not.
