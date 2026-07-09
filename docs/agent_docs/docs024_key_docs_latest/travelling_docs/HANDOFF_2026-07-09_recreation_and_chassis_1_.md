# HANDOFF — travelling docs (Tasks 3 & 4) · game recreation · chassis hygiene
**Written:** 2026-07-09 · supersedes `HANDOFF_2026-07-08_travelling_docs_and_toolgen_bug.md`.
**Companions:** `RUNBOOK_travelling_docs.md` (living tracker, rev 42), `RUNNING_NOTES_travelling_docs.md` (chronological log), `016b_debugging_guide_7_3_.md` (internal v8 — durable patterns), `PLAN_travelling_docs.md`, `PLAN_tool_acceptance_runner.md`.
**Bundle for the new chat:** run `drafts/bundle_recreation_v1.sh` (resolves paths, then builds `/tmp/bundle_recreation.md`).

---

## 0. First actions in the new chat
1. Re-paste the working constitution.
2. Read this doc, then the RUNBOOK position line + `YOUR TASKS`.
3. Run the bundle script (§6) and paste `/tmp/bundle_recreation.md`.
4. Start at §3 TASK A (the two `tool-recreation-handler` fixes). Fetch before drafting; snapshot before updating.

---

## 1. Where the project stands
**Both halves of the travelling-docs loop are in production.**
- **Task 3 — PLANs at birth: APPLIED and PROVEN.** `tool-generator` now runs `save_tool → compose_plan → write_plan → index_plan* → complete`. Run `1923badd` created `tool-xp-curve-designer` on gamesdesign.co.uk and wrote its own PLAN (`doc_plans`, `source='tool-generator'`, fence intact, 2,982 chars). *(`index_plan` is currently bypassed — see §4.)*
- **Task 4 — NOTES at every fix: APPLIED, NOT YET PROVEN.** `component-template-fixer`, `tool-improver`, `tool-recreation-handler` each gained `compose_note → append_note` on their success paths, with `config.error_step` containment so docs can never fail a fix. No machine-written `fix` note exists yet. **Producing one is the next milestone.**
- **Stage 3 (diagnosis loop writes NOTES) closed earlier:** first machine-written note is `('pipeline','build')`, categories `["diagnosis","unconfirmed-diagnosis"]`.
- **Pilot PLAN** seeded by hand for `tool-archetype-taster-quiz`. **Corrected PLAN** for `tool-xp-curve-designer` (v1 retired, v2 current) after the composer invented two selectors — see §5.

**Snapshots taken** (via the platform's own `snapshot_agent(p_agent_type, p_reason)`, which stores outside `agent_definitions`): `diagnose-agent`, `diagnose-orchestrator`, `tool-generator` (`1bca62f6`), `tool-improver` (`1f3ebb4a`), `component-template-fixer` (`076455bf`), `tool-recreation-handler` (`8701375f`).

---

## 2. What was fixed on the way here (do not re-litigate)
| Fault | Root cause | Resolution |
|---|---|---|
| Tool creation failed at `save_tool` | `create_tool_component` inserts `source_agent_type`/`source_orchestration_id`; production `content_components` had neither. Binary shipped ahead of its migration; latent since 2026-05-16 (no tool created in ~2 months). | `0NN_add_component_provenance.sql` applied — columns **mirrored** from `knowledge_base` (`varchar(100)` / `varchar(255)`, **not** uuid). **Canonical file exists but is parked in `docs/agent_docs/docs019_.../NNN_add_component_provenance.sql`** — never renumbered into the migrations path. Move it there. |
| `compose_plan` failed with `TEMPLATE_FIELD_ERROR` | Prompt used `{{.generated_html.result}}`. `execute_llm_prompt` with `output_format: text` gives the template a **bare string**; with `json` it gives a map. Action **config** paths are a different resolver and keep `.result`. | `0NN_fix_prompt_template_field_paths.sql` applied: `{{.generated_html}}`; the three note templates hardened to whole-object `{{.X | toJSON}}` (the live precedent in `recreate_tool`). |
| First auto-PLAN asserted non-existent selectors | Composer copied a real selector for the control it *acts* on (`#curveType`) and invented the ones it *asserts* on (`#xpTableBody`, `#statsStrip`). | `0NN_supersede_xp_curve_plan_selectors.sql` applied: PLAN superseded, correction recorded as a `doc_notes` entry. **The remedy is a check, not a sterner prompt** — see §5. |
| `index_plan` "hung" for 44 min | **Not a hang.** The chassis pod was `OOMKilled` (exit 137) ~23 s into the step; a dead pod writes nothing, so the row froze at `EXECUTING_STEP`. Embedder is healthy (`/api/tags` lists `nomic-embed-text`; `/api/embeddings` returns a vector). | `0NN_bypass_index_plan_until_embed_timeout.sql` applied (`write_plan.next_step = complete`; `index_plan` left defined + annotated). Real cause still open — see §4. |

---

## 3. TASK A (do first) — two `tool-recreation-handler` fixes, one migration
Both found by reading the definition, not by a failed run. Snapshot first (`SELECT snapshot_agent('tool-recreation-handler', '<file>: pre-update');`).

**(i) `spec` is undeclared.** `input_contract` = `required [site_id, domain]`, `optional [page_name, page_id, sections]`. The workflow reads `input_data.spec.mode`, `input_data.spec.interactive_features`, and (via the Task-4 tail) `input_data.spec.function`. Add `spec` to `optional`. *(Same class as the Stage-3b finding: an input the workflow depends on must be declared.)*

**(ii) The NOTES subject is wrong for this agent.** `tool-recreation-handler` ends `save_page_sections → update_page_status → deploy_page` and **never calls `create_tool_component`** — so `('tool', spec.function)` would key a doc to a function no component owns (a dangling doc). Recreation is site-scoped page work, exactly like `component-template-fixer`. **Re-subject `append_note` to `('pipeline','build')` with `note_site_id_field: site_record.site_id`; drop `subject_key_field`.** This also retires the "recreation items must carry `spec.function`" backlog item (for notes).

Verify after apply: contract shows `spec`; `append_note.config.subject_key = 'build'`, `subject_type = 'pipeline'`, `note_site_id_field` set, no `subject_key_field`.

---

## 4. TASK B — recreate the economy simulator (the Task-4 proof)
**Target:** `https://gamesdesign.co.uk/games/economy-simulator/index.html` (source uploaded previously as `index_7_.html`). Page `game-economy-simulator`, site `e33263f4-74f8-494f-b191-546845dbbddf`.

**Where the body lives (established by query, not assumption):** `page_components` → only a shared `hero`; `pages.sections` → **empty** (30 bytes of jsonb); `site_plan_sections` → site-plan *structure*, not HTML. The 32 KB body exists only as **deployed HTML in the sites repo**, with its source in the adoption crawl. Per doc 007: `research_results` holds `result_type='adoption_crawl'` (full markdown + rawHTML) and `result_type='adoption_page'` (per-page markdown). The site was adopted from `gamedesign.uk`; `economy-simulator` is in its adopted page list.

**`spec.mode` = `"recreate"`** — confirmed: `apply_adoption_plan_action.go:625` sets exactly that value with the comment *"load_existing_content checks for this value"*; `load_existing_content_action.go` declares `Optional: ["page_id","mode"]`.

**Verify before triggering:**
```sql
SELECT result_type, count(*) FROM research_results
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'::uuid GROUP BY 1;
-- and confirm the economy-simulator page's rawHtml/markdown is present
```
If the crawl content is gone, `load_existing_content` error-routes to `load_related_context` (contained by design) and the LLM would rebuild from the spec alone — losing the game's existing behaviour. Prefer feeding it the real content.

**The two bugs to fix (these become the recreated tool's first acceptance criteria):**
1. **Logic:** the tick uses the raw Player-Influx slider **index** as the players-per-tick rate (`popInflux = parseInt(el.slPop.value)`, slider `min=0 max=5`). Even "Flood" adds ~5 players/tick — ~20× too weak to notice. The live screenshot's 191 players = 50 + 5×~28 ticks, so players *are* added and gold *does* scale.
2. **Display:** the Players dataset is bound to `yAxisID: 'yGold'` (0–100k autoscale), so ~200 players renders flat on the baseline — "the green line stays on 0".

**Fix shape:** map slider indices to real rates (e.g. `[0,1,5,15,40,100]`) and give Players its own axis. Pass both as `spec.interactive_features` so `recreate_tool` (Opus, 64k tokens) produces corrected code.
**Tier-4 criterion to write into the PLAN:** set influx to max → the Active Players figure and the green series rise visibly within N ticks.

**Success = the first machine-written fix note:**
```sql
SELECT subject_type, subject_key, site_id, categories, source, left(body,140) AS head, created_at
FROM doc_notes WHERE categories ? 'fix' ORDER BY created_at DESC LIMIT 5;
```
Trigger envelope: mirror `drafts/084_TRIGGER_diagnose_v1.sh` / `085_TRIGGER_toolgen_gamesdesign_v1.sh` (kcat → `system.agent.generic.requests`, `action=orchestrate`, `config.agent_type=tool-recreation-handler`). **Env vars must be a same-line prefix or exported** — `VAR=x; script` does not reach the child (this cost two runs).

---

## 5. Standing design decisions worth keeping
- **Stage 5 (Tier-2 static checker) — the anchor rule.** Validate a criteria selector's **anchor** (leftmost id/class), never the whole path. `#tableWrap` exists ⇒ `#tableWrap tr` passes (rows are JS-built; Tier-4 asserts them for real). `#xpTableBody` exists nowhere ⇒ fail ⇒ drop the check or mark its id `-EDIT` (checkers skip `-EDIT`). **Static checks confirm, never refute**; never delete a check merely because the DOM is built at runtime. Same logic belongs in `tool-auditor`. Also: the composer emitted `"action":"select"` with a `value` — add that verb to the Tier-4 vocabulary.
- **Provenance:** `content_components.source_agent_type` stamped **`generic`**, because the shared chassis pod is the sender. `ExecutionContext.Sender.AgentType` is therefore no better than `Headers["agent_type"]` (which is empty). **Use the config-declared `plan_source` / `note_source`.** The planned "source_agent fallback" is **dropped**.
- **Docs never fail the work:** every doc step carries `config.error_step` to the terminal step. Corollary learned the hard way: containment covers *errors*, not *crashes* or *stalls*.

---

## 6. TASK C — chassis hygiene (parallel track)
- **OOM:** shared `generic` chassis pod, `OOMKilled` exit 137 after **2d19h** uptime, ~23 s into `rag_index`. `collected_data` was only **192 kB** with **2** `__raw_message__` copies ⇒ **state growth is not the cause**; a slow leak is the leading hypothesis. Old pod is gone, so `--previous` logs are lost (**capture crash logs immediately**).
  Next: `describe pod … | grep -A4 Limits`, `top pod` baseline then again after a day, `get events --field-selector reason=OOMKilling`, then pprof.
- **Remove the `DEBUGaa` logging.** The coordinator logs the entire params object at Info on every action (`DEBUGaa: params sent to action handler`, `DEBUGaa: in buildActionParams …`) — it serialises all collected data twice per action and prints headers.
- **`rag_index` needs a deadline.** It chunks content then calls `GenerateEmbedding` per chunk against `ollama-adapter:11434` with no `context.WithTimeout`. Add one (15 s) so a stall degrades into the existing non-fatal "store without embeddings" path; consider an action-level deadline so no action outlives `timeout_seconds`. **Then re-enable `index_plan`** (`write_plan.next_step = 'index_plan'`; the UPDATE is in the bypass migration's tail).
- **Loose end:** the new tool page `tool-xp-curve-designer` is `build_status = 'planned'` — it is not deployed yet. Check the build/rerender pipeline picked up its work item.
- **Unrelated but burning:** `github-actions-runner-…-lhg9l` is in `CrashLoopBackOff` with **3,213** restarts — `StartError`: runc *"expected cgroupsPath to be of format slice:prefix:name for systemd cgroups"*. A node cgroup-driver mismatch, not an app bug. Deploys survive on the healthy replica.

---

## 7. The bundle (run this, paste the output)
`drafts/bundle_recreation_v1.sh` — resolves scope paths, prints them for eyeballing, then runs `cmd/bundle`. It scopes: `load_existing_content_action.go`, `apply_adoption_plan_action.go`, the save/validate/status actions, `check_tool_completeness_action.go`, `create_tool_component_action.go`, the four doc actions, `rag_actions.go`, `check_tool_health.go`, the LLM action, and `coordinator.go`; docs `007_adoption_pipeline_v4`, `005_tool_pipeline`, `020_tool_lifecycle`, `003_contracts_and_standards`, `001_development_guide`, `019_tool_library`; schema for `agent_definitions, doc_plans, doc_notes, content_components, pages, page_components, site_work_items, research_results, sites, site_specs, orchestration_states, agent_error_log, knowledge_base`; runtime `gamesdesign.co.uk` / `game-economy-simulator`.

*Regenerate `/tmp/analysis_repo.json` first, exactly as for the previous bundle.*

**Path-resolution facts learned 2026-07-09** (the script now handles these):
- `execute_llm_prompt` lives in **`platform/orchestration/actions/ai_actions.go`**
  (not a `*_action.go` file). This is the shared action behind `generate_tool_html`,
  `compose_plan`, `compose_note`, `analyze_tool` and `recreate_tool`.
- `validate_page_content.go` has **no `_action` suffix** — file naming is not
  consistent, so resolve actions via their registration in `registry.go` (action
  name → constructor → file), not by filename convention.
- `save_page_sections`, `update_page_status`, `check_tool_health` did **not**
  resolve on the first pass. The script now prints grep candidates for each.
  `check_tool_health` may not exist yet at all — Stage 5 (the static acceptance
  checker) is unbuilt, so treat its absence as expected and confirm before
  planning against it.

---

## 8. Debugging rules banked this arc (016b v8)
- **`agent_error_log` is the FIRST read.** It outlives the pod and carries `step_name`, `action`, `error_message`, `error_code`, `context`. Filter by `orchestration_id` — **which is `text`, not `uuid`** (casting raises `operator does not exist: text = uuid`).
- **`EXECUTING_STEP` forever usually means the worker died**, not that the step is slow. `since_s` then measures time since the crash. Check `RESTARTS`, `describe pod … Last State`, and `logs --previous`. Probe suspected-stalled dependencies with a bound (`curl -m 5`).
- **`current_step` from polling is a sample, not an attribution.** A 120 s poll skipped `save_tool` entirely and made an LLM step look guilty.
- **`complete_error` is terminal** — a downstream artefact missing (0 PLAN rows) can be the *correct* consequence of an upstream failure, not a regression.
- **`error_step` lives inside `step.Config`** and must name an existing step; inside loops the values are iteration-prefixed, so they must name substeps of that loop.
- **Prompt templates vs config paths:** `text` output → bare string (`{{.X}}`); `json` output → map (`{{.X.result | toJSON}}`); action config keeps `.result`.
- **Postgres:** bounded regex repetition caps at 255 (`RE_DUP_MAX`) — prefer `strpos`/`substr` in guards. An aborted transaction ignores every later command **including `BEGIN`** — `ROLLBACK;` first. Run migrations with `psql -f`, not by pasting (pasting mangles comments and dollar-quoted bodies).
- **Snapshot before every `agent_definitions` update:** `SELECT snapshot_agent('<type>','<migration>: pre-update');`

---

## 9. File inventory (`/mnt/user-data/outputs/`)
**Living docs:** `RUNBOOK_travelling_docs.md` (rev 42), `RUNNING_NOTES_travelling_docs.md`, `PLAN_travelling_docs.md`, `PLAN_tool_acceptance_runner.md`, `016b_debugging_guide_7_3_.md`, this handoff.

**Drafts — applied:** `0NN_doc_plans_and_notes.sql`, the four doc-action `.go` files, `0NN_wire_persist_diagnosis_note.sql`, `0NN_diagnose_load_runtime_error_step.sql`, `0NN_fix_load_runtime_error_step_target.sql`, `0NN_wire_diagnosis_subject_threading.sql`, `pilot_PLAN_tool-archetype-taster-quiz.sql`, `0NN_tool_generator_plan_writing.sql`, `0NN_fix_agents_note_writing.sql`, `0NN_add_component_provenance.sql`, `0NN_fix_prompt_template_field_paths.sql`, `0NN_bypass_index_plan_until_embed_timeout.sql`, `0NN_supersede_xp_curve_plan_selectors.sql`.
**Drafts — not applied:** `0NN_agent_definition_snapshots.sql` (**SUPERSEDED — use `snapshot_agent()`**), `doc_actions_helpers_test.go` (unshipped), `preflight_toolgen_columns.sql` (reusable).
**Triggers:** `084_TRIGGER_diagnose_v1.sh` (`SUBJECT_TYPE`/`SUBJECT_KEY` env), `085_TRIGGER_toolgen_gamesdesign_v1.sh` (real side effects on the live site), `bundle_recreation_v1.sh` (new).
**Still to write:** `086_TRIGGER_recreate_economy_simulator.sh`.
