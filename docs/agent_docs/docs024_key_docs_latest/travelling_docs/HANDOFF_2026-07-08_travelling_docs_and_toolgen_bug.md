# HANDOFF — travelling-docs rollout · tool-generation bug · gamesdesign recreation
**Written:** 2026-07-08 · **updated same day: the blocker is now ROOT-CAUSED — see §3.** For continuation in a fresh chat.
**Companion docs (read alongside):** `RUNBOOK_travelling_docs.md` (rev 28, the living tracker), `RUNNING_NOTES_travelling_docs.md` (chronological log), `016b_debugging_guide_7_3_.md` (internal v5, durable patterns), `PLAN_travelling_docs.md` (feature spec), `PLAN_tool_acceptance_runner.md` (Tier-4 plan).

---

## 0. First actions in the new chat (in order)
1. Re-paste the working constitution (the reuse-first / every-agent-is-an-orchestrator / snapshot-before-update / logger.Info / check-schema-before-SQL rules).
2. Read this doc, then the RUNBOOK position line + §3a/§3b/TASK sections.
3. The blocker is **diagnosed** (§3): a schema drift, fix drafted at `drafts/0NN_add_component_provenance.sql`. Apply it (reuse-check first), re-run `085`, and the Task-3 proof closes. §5/§6 are kept as the record of how it was found.

---

## 1. Project orientation (one paragraph)
An AI agent-orchestration platform: a Go "chassis" (`gqls/agentchassis`) runs agents whose workflows live in Postgres (`clients_db`) at `agent_definitions.default_config->'workflow'`; agents talk over Kafka (namespace `kafka`), pods run in namespace `ai-persona-system`, and built sites deploy via a GitHub repo to Backblaze. Every agent is an orchestrator owning a workflow of steps that call Go actions; complexity lives in the actions, workflows stay simple, no sub-workflows (spawn sub-agents instead). We have been building **travelling docs**: a Postgres-backed PLAN + NOTES system so every tool and pipeline carries its own intent (PLAN) and history (NOTES), written *by the agents themselves* rather than by hand.

---

## 2. State of play — what is DONE (with dates + snapshot ids)
- **Tables live** (`0NN_doc_plans_and_notes.sql`, applied 2026-07-04): `doc_plans` (PLAN, supersede-versioned via `is_current`) + `doc_notes` (append-only). Keyed `(subject_type, subject_key)` = `('tool', <content_components.function>)` or `('pipeline', build|content|design|maintenance)`. Acceptance criteria live in the PLAN body as a fenced ```criteria JSON block.
- **Four Go actions deployed:** `write_doc_plan`, `append_doc_note`, `load_doc_context`, `persist_diagnosis_note`.
- **Stage 3a CLOSED** (diagnosis loop skips persistence with no subject): `persist_diagnosis_note` wired into `diagnose-agent` (`config.error_step="complete"`); `load_runtime.config.error_step` set to its own `next_step` (`assemble_bundle`, derived) so anchorless runs degrade instead of dying. Verified in state form (a run's `ProcessingHistory` + `COMPLETED` + 0-count), because the skip stdout line only survives inside the 3600s idle-reaper window.
- **Stage 3b CLOSED** (positive path): subject threaded through — `diagnose-orchestrator.call_diagnoser.input_mapping` += `subject_type?`/`subject_key?`, and **both** `input_contract.optional` lists += `subject_type`,`subject_key` (the mapping must satisfy the callee's contract). Positive run `cc61fad8` wrote the **first machine-written NOTES row**: `('pipeline','build')`, categories `["diagnosis","unconfirmed-diagnosis"]`, `source='diagnosis-loop'`, stop reason **scope-not-narrowing**, 2026-07-07 11:41.
- **Retroactive snapshots taken:** `diagnose-agent` (`34f4afc8…`), `diagnose-orchestrator` (`e8e96d24…`). Learned: `snapshot_agent(p_agent_type, p_reason DEFAULT NULL) RETURNS uuid` stores snapshots **outside** `agent_definitions` (the inspect showed only the originals), so the `is_snapshot` current-row selector predicate is **not** needed.
- **Pilot PLAN seeded:** `tool-archetype-taster-quiz` (`has_fence`, 2,761 chars, 2026-07-07 12:32) — the format road-test; Stage-5's precondition (>=1 tool PLAN with criteria) is met.
- **Task 3 APPLIED + shape-verified** (`0NN_tool_generator_plan_writing.sql`; snapshot `1bca62f6…`): `tool-generator` now writes a PLAN at every tool creation — `save_tool` -> `compose_plan` -> `write_plan` (`write_doc_plan`) -> `index_plan` (`rag_index` into `tool_docs`) -> `complete`; every new step `config.error_step="complete"` (docs never fail creation); timeout 300->480; and the three inert step-level `error_step`s moved into `config`. **PROOF PENDING** (blocked — see §3).
- **Task 4 APPLIED** (`0NN_fix_agents_note_writing.sql`; snapshots `076455bf…` fixer, `1f3ebb4a…` improver, `8701375f…` recreation): the three fix agents append a NOTES entry at every *successful* fix (`compose_note` -> `append_note` on the success path; `config.error_step` to the terminal step). `component-template-fixer` covers both branches (incl. the no-op else); `tool-improver` keys on `tool_data.function`; `tool-recreation-handler` keys on `input_data.spec.function` (may be absent -> append errors-to-complete, note skipped, run unharmed). All ten of recreation's inert step-level `error_step`s corrected into `config`; guard asserts zero step-level `error_step`s remain across all three. **PROOF PENDING** (needs one successful fix/recreation to write the first `categories ? 'fix'` row).

So: **both halves of the loop exist in production** — PLANs at birth, NOTES at every fix — and what remains is proving each fires on a healthy run, which the blocker currently prevents.

---

## 3. THE BLOCKER — ROOT-CAUSED 2026-07-08 (schema drift, fix drafted)

**Run:** `correlation_id = 00688389-1d83-4574-be15-8467612f4cb0`, orchestration
`9f93a988-ddf6-4d5c-806f-e8b82e194032`, 2026-07-08.

**The evidence that settled it** (`agent_error_log`, 16:14:44, same
orchestration id):
> `step save_tool failed: failed to execute action create_tool_component: failed to create tool component: ERROR: column "source_agent_type" of relation "content_components" does not exist (SQLSTATE 42703)`

**So `generate_tool_html` SUCCEEDED.** The LLM produced the tool HTML; the run
died one step later in `save_tool`. The earlier reading ("the LLM step failed")
came from polling `current_step` every 120s, which sampled the run during
generation and never saw `save_tool` execute. **Lesson banked in 016b v6:**
`agent_error_log` outlives the pod and names the failing step + action — read it
FIRST, filtered by `orchestration_id`; `current_step` is a sampled value, not an
attribution.

**Root cause: the binary shipped ahead of its migration.**
`create_tool_component_action.go` documents its own dependency directly above the
INSERT: *"source_* = creation provenance (NNN_add_component_provenance.sql),
mirroring knowledge_base's pair … apply that migration before this binary
deploys."* Production `content_components` has neither `source_agent_type` nor
`source_orchestration_id`; `knowledge_base` has both (`rag_actions.go` inserts
them). **Latent for ~2 months**: the last `created_from='generated'` tool
component is 2026-05-16, and `component-creator` inserts a different column set
(so it kept working — `provocations-archive-list`, 07-06). Our proof run was the
first caller of this path since the drift landed.

**Task 3 is NOT broken.** `complete_error` is terminal and bypasses
`compose_plan`, so zero `doc_plans` rows was the *correct* outcome of an upstream
failure. Task 3 remains **unproven**, not regressed.

**The fix (3 steps):**
1. **Reuse first** — find the canonical migration before writing one:
   `find ~/projects/agentchassis -name '*provenance*'` and
   `git -C ~/projects/agentchassis grep -l add_component_provenance`.
   If it exists, apply that file (diff it against the draft). If the file the
   comment names was never written, apply the draft and commit it to the repo.
2. Apply **`drafts/0NN_add_component_provenance.sql`** — it copies
   `knowledge_base`'s exact column types via `format_type`/`pg_attribute` +
   `EXECUTE format(... ADD COLUMN IF NOT EXISTS ...)` rather than guessing;
   additive, nullable, idempotent, guarded. No `snapshot_agent()` (data table).
   Verify with the select in the file tail.
3. **Re-run `./drafts/085_TRIGGER_toolgen_gamesdesign_v1.sh` unchanged.** The
   function name `tool-xp-curve-designer` is still free — the failed run inserted
   nothing (the component INSERT was its first statement; no orphan rows, as the
   user's own queries confirmed). Expect: component + page + nav, then
   `compose_plan → write_plan → index_plan`, i.e. a `doc_plans` row with
   `source='tool-generator'` = **the Task-3 proof**, plus the component stamped
   `source_agent_type='tool-generator'`.

**No second drift lurking:** every other INSERT in that action (`pages`,
`page_components`, `site_work_items`) was compared column-by-column against the
bundle's production schema — all match.

**Bonus finding (makes a backlog item precise):** `write_doc_plan_action.go` and
`append_doc_note_action.go` both do `sourceAgent = params.Headers["agent_type"]`,
which is empty in that step context — that is exactly why `source_agent` was
blank on the first machine-written NOTES row. `create_tool_component` uses
`params.ExecutionContext.Sender.AgentType`, which populates. Next chassis build:
fall back to the execution context in the doc actions (and `rag_actions.go`,
which has the same weakness).

## 4. The gamesdesign recreation thread (the second half of the task)
**Decision on record:** gamesdesign.co.uk is the dogfood site (over a new domain) — already in the pipelines with specs, has live tools, and a real broken component to exercise docs + fix + history.

**The economy simulator is a RECREATION case, not an improve case.** The definitive lookup (Q1, page-linked, no name filter) returns only a shared **`hero` section** (`component_level=section`, 2,790 chars, same `component_id` as the guide page) on the game page — **there is no game-body component**. So `tool-improver` has nothing to load; the live interactive page must be **adopted into a component via `tool-recreation-handler`** (analyse -> recreate -> check_completeness -> validate -> save_page_sections -> deploy). This is also the agent whose ten `error_step`s Task 4 just corrected, so a recreation run doubles as a Task-4 exercise (it will append the first `fix`-category NOTES entry).

**Live source of record:** `https://gamesdesign.co.uk/games/economy-simulator/index.html` (uploaded as `index_7_.html`).

**The two defects — these become the recreated tool's first acceptance criteria:**
1. **Logic:** the tick reads the raw Player-Influx slider *index* as the players-per-tick rate. The slider is `min=0 max=5`; even "Flood" adds only ~5 players/tick, roughly 20x too weak to notice. (The screenshot's 191 players = 50 + 5x~28 ticks confirms players *are* added and gold generation *does* scale — just far too weakly to see.)
2. **Display:** the Players chart dataset is bound to `yAxisID: 'yGold'` (0-100k scale), so ~200 players renders flat on the baseline — the "green line stays on 0."

**Fix shape:** map slider indices to real rates (e.g. `[0,1,5,15,40,100]`) and give Players its own axis. **Tier-4 acceptance check:** set Player Influx to max -> expect the Active Players figure and the green series to rise visibly within N ticks. (Honest scope note: the chassis diagnosis loop analyses the Go repo and cannot see in-page JS; the Tier-4 acceptance runner — see `PLAN_tool_acceptance_runner.md` — is the layer that would catch this class of bug.)

---

## 5. DATA COLLECTED (2026-07-08) — kept as the record of how the cause was found
Outcome: probe **D** proved the bug latent (last generated tool = 2026-05-16);
the bundle's own `runtime.md` section carried the decisive `agent_error_log`
row. Pod logs (B) were unavailable — the pod had been reaped — which is exactly
why `agent_error_log` is now the first read. **Add this query to the standard
set** (it is better than everything below):
```sql
SELECT occurred_at, agent_type, step_name, action, error_code, severity,
       left(error_message,400) AS err, context
FROM agent_error_log
WHERE orchestration_id = '<uuid>'::uuid ORDER BY occurred_at;
```
The original checklist follows, for reference.

**A. The failing run's own evidence (highest value):**
```sql
SELECT status, current_step,
       jsonb_pretty(collected_data) AS collected,
       substring(COALESCE(error,''),1,800) AS err
FROM orchestration_states
WHERE correlation_id = '00688389-1d83-4574-be15-8467612f4cb0'::uuid
ORDER BY created_at;
-- if there is a processing_history / history jsonb column, add jsonb_pretty() of it too.
```

**B. The tool-generator pod log for this run** (hyphen label key; attribute by the correlation id — pods are shared):
```bash
kubectl -n ai-persona-system logs -l agent-type=tool-generator --tail=4000 | grep 00688389
kubectl -n ai-persona-system logs -l agent-type=tool-generator --tail=4000 \
  | grep -iE 'error|panic|timeout|refus|invalid|empty|token|rate|overload|template' | tail -40
```

**C. The current stored workflow** (confirm structure is intact post-migration):
```sql
SELECT jsonb_pretty(default_config->'workflow')
FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL
ORDER BY version DESC LIMIT 1;
```

**D. Is the failure systemic / latent?** (has generation worked at all recently?):
```sql
SELECT function, created_from, created_at FROM content_components
WHERE created_from IN ('generated','tool') ORDER BY created_at DESC LIMIT 10;
```

**E. Pod env + image** (model/key sanity):
```bash
P=$(kubectl -n ai-persona-system get pods -o name | grep tool-generator | head -1)
[ -n "$P" ] && kubectl -n ai-persona-system describe "$P" | grep -iE -A2 'ANTHROPIC_API_KEY|Image:'
```

**F. Model-string cross-check** (which agents use it; has any run recently succeeded?):
```sql
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%claude-sonnet-4-6%' AND deleted_at IS NULL ORDER BY type;
```

---

## 6. THE BUNDLE COMMAND (filled for this task + bug)
Built around `execute_llm_prompt` (the fulcrum for the bug and for compose_plan/compose_note/recreate). Whole-file scopes where the failure could be anywhere in the file; symbol scopes where one function is the target.

**Step 0 — regenerate the repo analysis** exactly as you did for the example run (the analyzer step that produces `/tmp/analysis_repo.json`).

**Result 2026-07-08:** the bundle rendered fine and contained the needed file
(`create_tool_component_action.go`) plus the schema and the `agent_error_log`
extract. **Both Step-0b resolvers misfired** — `LLM_ACTION` resolved to
`ai_errors.go` (first alphabetical match, wrong) and `AI_CLIENT` to a stray
`.txt` file — a reminder to eyeball the echoes. It did not matter, because the
fault was in an explicitly-scoped file. If a future bundle targets the LLM path,
resolve those two by hand.

**Step 0b — resolve the two paths I am least sure of, and eyeball them before running:**
```bash
cd ~/projects/agentchassis
LLM_ACTION=$(grep -rl 'execute_llm_prompt' platform/orchestration/actions/*.go | grep -v registry | head -1)
AI_CLIENT=$(grep -rlEi 'api\.anthropic\.com|anthropic' pkg platform internal 2>/dev/null | grep -v _test | grep -vi 'actions/registry' | head -1)
echo "LLM_ACTION=$LLM_ACTION"      # expect the execute_llm_prompt action file
echo "AI_CLIENT=$AI_CLIENT"        # expect the Anthropic HTTP client / ai service
# Also confirm the doc-action filenames exist (adjust the -scope lines if named differently):
grep -rl 'write_doc_plan\|append_doc_note\|load_doc_context\|persist_diagnosis_note' platform/orchestration/actions/*.go
```

**The command** (the `-doc` prefix and `-psql` form are copied from your example; verify the doc-action and read_site_spec/ensure_site_record filenames against Step 0b):
```bash
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "Diagnose and fix a tool-generation failure, then complete the travelling-docs rollout and recreate a broken game tool. IMMEDIATE BUG: the tool-generator agent's generate_tool_html step (action execute_llm_prompt, model claude-sonnet-4-6, ~8000 tokens) failed on run correlation_id 00688389-1d83-4574-be15-8467612f4cb0 (creating tool tool-xp-curve-designer on gamesdesign.co.uk) and routed to complete_error after ~60-96s; no content_components row and no page were created, therefore correctly no doc_plans row (complete_error bypasses the new compose_plan/write_plan steps, so 0 PLANs is expected, not a regression). Find the root cause of the execute_llm_prompt failure and rank then confirm among: a call-level timeout shorter than the step runtime; empty/invalid/refusal LLM output rejected by the action's validation; a Go text/template execution error from the prompt referencing nil site_specs sub-fields such as .site_specs.specs.identity.* possibly under missingkey=error; an API/auth/model-availability error such as the model string no longer being valid; or an output-size/fence issue. The step's error_step was correctly moved into config by an already-applied migration, so the routing that fired is the corrected behaviour catching a real generation error underneath - do not treat the routing as the bug, and prefer a structural fix in the action over a workflow workaround. execute_llm_prompt is shared by generate_tool_html, the recreation steps analyze_tool and recreate_tool (Opus), and the newly-added compose_plan (Task 3) and compose_note (Task 4) steps, so a systemic fault affects all of them; the doc steps contain errors to their terminal step so they degrade gracefully, but generation must succeed. TASK CONTEXT: travelling docs is a Postgres-backed PLAN+NOTES system (doc_plans supersede-versioned by is_current, doc_notes append-only, keyed by subject_type and subject_key) that is live end-to-end for diagnoses; Task 3 (tool-generator writes a PLAN at every tool creation) and Task 4 (component-template-fixer, tool-improver and tool-recreation-handler each append a NOTES entry at every successful fix) are applied and shape-verified but unproven on a healthy run - once generation works, re-running the tool-generator trigger must produce both the tool and its auto-PLAN with source tool-generator, and a subsequent fix or recreation must produce a doc_notes row with category fix; confirm the composer output obeys its template (five standard checks verbatim, interaction selectors copied from real HTML or omitted, criteria fence parses, under 3000 chars). Then RECREATE the broken economy-simulator game at gamesdesign.co.uk/games/economy-simulator/index.html: it is a recreation case (its page has no body component, only a shared hero section linked via page_components), so it flows through tool-recreation-handler (analyse, recreate, check_completeness, validate, save_page_sections, deploy), not tool-improver. Its two defects become the recreated tool's first acceptance criteria: a logic bug where the tick reads the raw Player-Influx slider index (min 0 max 5) as the players-per-tick rate so even Flood adds only 5 per tick, about 20x too weak; and a display bug where the Players chart dataset is bound to yAxisID yGold on a 0 to 100k scale so it renders flat on the baseline. The fix maps slider indices to real rates and gives Players its own axis, and the Tier-4 acceptance check is: set influx to max, expect the Active Players figure and the green series to rise visibly within N ticks. CONSTRAINTS: every agent is an orchestrator; reuse existing functions and structs before creating new; keep workflows simple with complexity in Go actions; no sub-workflows, spawn sub-agents instead; error_step lives in step Config and must name an existing step, and inside loops values are iteration-prefixed so name substeps only; snapshot agents via snapshot_agent(type,reason) before any agent_definitions update; check DB schemas before writing SQL; use logger.Info not Debug; kubectl namespaces are ai-persona-system and kafka; deploys go to the gqls/agentchassis repo via GitHub Actions to Backblaze." \
  -scope "$LLM_ACTION" \
  -scope "$AI_CLIENT" \
  -scope platform/orchestration/actions/create_tool_component_action.go:CreateToolComponentAction \
  -scope platform/orchestration/actions/check_tool_completeness_action.go \
  -scope platform/orchestration/actions/read_site_spec_action.go \
  -scope platform/orchestration/actions/ensure_site_record_action.go \
  -scope platform/orchestration/actions/write_doc_plan_action.go \
  -scope platform/orchestration/actions/append_doc_note_action.go \
  -scope platform/orchestration/actions/load_doc_context_action.go \
  -scope platform/orchestration/actions/rag_actions.go \
  -include platform/orchestration/actions/registry.go \
  -doc docs/agent_docs/docs024_key_docs_latest/005_tool_pipeline_1_.md \
  -doc docs/agent_docs/docs024_key_docs_latest/020_tool_lifecycle_2_.md \
  -doc docs/agent_docs/docs024_key_docs_latest/003_contracts_and_standards_7_.md \
  -doc docs/agent_docs/docs024_key_docs_latest/023_llm_quality_testing.md \
  -doc docs/agent_docs/docs024_key_docs_latest/001_development_guide_5_.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables agent_definitions,doc_plans,doc_notes,content_components,pages,page_components,sites,site_specs,orchestration_states \
  -runtime-site gamesdesign.co.uk -runtime-page game-economy-simulator \
  -out /tmp/bundle_toolgen_llm.md
```

**Path notes (verify before running):**
- `$LLM_ACTION` / `$AI_CLIENT` come from Step 0b — the single most important scopes; make sure the echoes look right.
- `create_tool_component_action.go`, `check_tool_completeness_action.go`, `rag_actions.go` match filenames seen earlier in this project; the doc-action names (`write_doc_plan_action.go` etc.) and `read_site_spec_action.go` / `ensure_site_record_action.go` are best-guesses — adjust from Step 0b's grep.
- `-runtime-page game-economy-simulator` uses the page *name*; if the bundler expects a url or slug, use `/games/economy-simulator/index.html` or `games/economy-simulator`.

**Optional additions if you want to prep the recreation in the same bundle** (append these `-scope` lines) — the recreation LLM steps already ride on `$LLM_ACTION`, so only the non-LLM save/validate actions are extra:
```
  -scope platform/orchestration/actions/validate_page_content_action.go \
  -scope platform/orchestration/actions/save_page_sections_action.go \
  -scope platform/orchestration/actions/save_tool_training_data_action.go \
  -scope platform/orchestration/actions/update_page_status_action.go
```
Add `-doc docs/agent_docs/docs024_key_docs_latest/019_tool_library_2_.md` and `-doc docs/agent_docs/docs024_key_docs_latest/011_database_and_infrastructure.md` if you want the KB-write and infra/timeout context too.

---

## 7. Key facts, schema, gotchas (so nothing is relearned)
**Agent definitions / workflows:**
- Workflow lives at `default_config->'workflow'`. Select the current row as `max(version) WHERE deleted_at IS NULL`; update in place per approval. Address agents by `type`, not `name`. There is no `processing_mode` column.
- **`snapshot_agent(p_agent_type, p_reason DEFAULT NULL) RETURNS uuid`** — snapshot before every agent update; stores **outside** `agent_definitions`; no `is_snapshot` selector predicate needed. (`\df *agent*` likely reveals a companion restore function when first needed.)
- **`error_step` must be inside `step.Config`** (dev guide §16) — step-level is silently ignored — **and must name an existing step** (an unknown target converts a recoverable step failure into a fatal one). Pattern: derive the target from the step's own `next_step` so success and failure converge. **Inside a loop's substeps, `error_step`/`then_step`/`fallback_step` are iteration-prefixed at expansion — name substeps of that loop only; `continue_on_error: true` is the iteration-scoped alternative.** Dormant step-level instances still exist in `tool-auditor` (Stage-5's concern).

**Runs / logs:**
- Pod **label** key is `agent-type` (hyphen); log **fields** say `agent_type` (underscore); the underscore selector matches zero pods. `-l agent-type=<agent>` spans all live pods of that type (idle reaper 3600s), so attribute every line by orchestration id / pod / timestamp — "the plan table is ground truth, the rest is weather."
- **Two failure envelopes.** Step-level failure: header `status: complete` + `body.status: failed`; the parent `orchestration_states` row shows `COMPLETED` with the child's error in its `error` column. Workflow-start failure: `status: error_unrecoverable`, body `CHILD_ORCHESTRATION_FAILED`. Check the body, never the header alone. A `COMPLETED` row with a non-empty `error` = a forwarded child failure.
- **`complete_error` is terminal and bypasses downstream steps** — 0 PLAN/NOTE rows can be the *correct* outcome, not a fault (this is exactly the current blocker's shape).
- Gate-evidence capture window = the 3600s reaper; past it, a post-completion state dump (`ProcessingHistory` showing the step executed + the terminal status + the 0-count) is the accepted substitute.

**Schema (verified 2026-07-08):**
- `content_components` has **no `site_id` column** — components attach to sites only via `page_components` / `site_components`. `idx_cc_tool_function_unique` = `UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND is_active=true`, so **duplicate `function` rows are forks/inactive versions**. `name` is UNIQUE. `created_from` CHECK allows `{manual, generated, adopted, tool, forked}`. `component_level` default `section`.
- The game page (`game-economy-simulator`, site `e33263f4-74f8-494f-b191-546845dbbddf`) has only a shared `hero` section linked -> recreation case.

**Travelling docs:**
- `source_agent` comes back **empty** on machine-persisted rows (the `agent_type` header is absent in that step context; provenance is carried by `source`). Backlog: a fallback population in the doc actions.
- The pilot PLAN's `EDIT:` markers are fill-later blanks (valid as seeded); fills arrive by **supersede** (flip `is_current`, insert new — never edit history in place). Acceptance checks whose id ends `-EDIT` are skipped by verification until real selectors replace placeholders.

---

## 8. Open threads / backlog
1. **Blocker (root-caused):** apply `drafts/0NN_add_component_provenance.sql` (after the reuse check), re-run `085` — a clean run yields the tool *and* the auto-PLAN (Task-3 proof) in one pass. Then paste the first auto-PLAN body for a one-time composer review (template conformance).
2. **Recreation:** draft a `tool-recreation-handler` trigger for the economy-simulator page (scope to the page; set `spec.mode` and `spec.function` so the note tail has a subject). A successful run writes the first `fix`-category NOTES row (Task-4 proof) and produces a corrected game whose PLAN carries the influx/axis acceptance check.
3. **Refinement on record:** stamp `function` into recreation item specs at creation, so `tool-recreation-handler`'s note tail always has a subject (currently skip-on-absence).
4. **Chassis build (next code drop):** ship `doc_actions_helpers_test.go`; soften `diagnose_load_runtime` for the no-anchor case (`{runtime_evidence:"", skipped:true, reason}` instead of a hard error); **`source_agent` fallback — replace `params.Headers["agent_type"]` with a fallback to `params.ExecutionContext.Sender.AgentType` in `write_doc_plan_action.go`, `append_doc_note_action.go`, `rag_actions.go`** (root cause now known). Consider a pre-deploy check that greps a diff for new column names and asserts each exists in production (this incident's class).
5. **Stage 5 (Tier-2 static check):** read `check_tool_health.go` first, then extend it with static contract-presence checks against the PLAN's criteria.
6. **Stage 6 (Tier-4 runner):** `browser-runner-adapter` (Playwright desktop+mobile, 035-conformant envelope, topic `system.adapter.browser-runner.requests`) — see `PLAN_tool_acceptance_runner.md`. This is the layer that would catch the game bug.
7. **Standing opens:** KB `tool_docs` write per doc 019; `deploy_tool_to_site` `source_*` stamp; `rag_index` `source_type` parameterisation (currently reads `scrape`).

---

## 9. File inventory
**Living docs (`/mnt/user-data/outputs/`):** `RUNBOOK_travelling_docs.md` (rev 28), `RUNNING_NOTES_travelling_docs.md`, `PLAN_travelling_docs.md`, `PLAN_tool_acceptance_runner.md`, `016b_debugging_guide_7_3_.md`, and this handoff.

**Drafts (`/mnt/user-data/outputs/drafts/`), all applied unless noted:**
- `0NN_doc_plans_and_notes.sql` (tables), `verify_before_migration.sql`
- `write_doc_plan_action.go`, `append_doc_note_action.go`, `load_doc_context_action.go`, `persist_diagnosis_note_action.go`, `doc_actions_helpers_test.go` (test un-shipped)
- `0NN_wire_persist_diagnosis_note.sql`, `0NN_diagnose_load_runtime_error_step.sql`, `0NN_fix_load_runtime_error_step_target.sql`, `0NN_wire_diagnosis_subject_threading.sql`
- `0NN_agent_definition_snapshots.sql` — **SUPERSEDED, do not apply** (use `snapshot_agent()`)
- `pilot_PLAN_tool-archetype-taster-quiz.sql`
- `0NN_tool_generator_plan_writing.sql` (Task 3), `0NN_fix_agents_note_writing.sql` (Task 4)
- `0NN_add_component_provenance.sql` — **NOT YET APPLIED**; unblocks tool creation (schema drift; reuse-check the repo first)
- `084_TRIGGER_diagnose_v1.sh` (diagnosis trigger; `SUBJECT_TYPE`/`SUBJECT_KEY` env, same-line prefix or export), `085_TRIGGER_toolgen_gamesdesign_v1.sh` (tool-creation trigger; real side effects on the live site)

**Uploads referenced this session:** `index_7_.html` (economy-simulator source of record), the guideline docs (001/002/003/005/011/016/019/020/023/035), the three fix-agent workflow dumps, `create_tool_component_action.go`, `check_tool_completeness_action.go`, `rag_actions.go`, `example_bundle.txt` (the bundle-command template this doc's §6 is modelled on).
