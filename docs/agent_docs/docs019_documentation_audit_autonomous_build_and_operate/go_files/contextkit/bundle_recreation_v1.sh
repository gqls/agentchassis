#!/usr/bin/env bash
# bundle_recreation_v1.sh — build the context bundle for the NEXT chat.
# Modelled on example_bundle.txt. Resolves file paths FIRST (last time both
# guessed paths missed), prints them for eyeballing, then runs cmd/bundle.
#
# Usage:
#   cd ~/projects/agentchassis
#   # regenerate the repo analysis exactly as before, then:
#   bash bundle_recreation_v1.sh
#
# Output: /tmp/bundle_recreation.md
set -uo pipefail

ROOT="${ROOT:-$HOME/projects/agentchassis}"
ACT="$ROOT/platform/orchestration/actions"
cd "$ROOT"

echo "=== resolving scope paths (eyeball these before the bundle runs) ==="
resolve() {  # resolve <symbol-or-token> <fallback-glob>
  local hit
  hit=$(grep -rl "$1" "$ACT"/*.go 2>/dev/null | grep -v _test | grep -v registry.go | head -1)
  [ -n "$hit" ] && echo "${hit#$ROOT/}" || echo ""
}

LOAD_EXISTING="platform/orchestration/actions/load_existing_content_action.go"   # confirmed by grep
APPLY_ADOPTION="platform/orchestration/actions/apply_adoption_plan_action.go"    # confirmed by grep
CREATE_TOOL="platform/orchestration/actions/create_tool_component_action.go"     # confirmed (uploaded)
RAG="platform/orchestration/actions/rag_actions.go"                              # confirmed (bundle)
WRITE_PLAN="platform/orchestration/actions/write_doc_plan_action.go"             # confirmed (bundle)
APPEND_NOTE="platform/orchestration/actions/append_doc_note_action.go"           # confirmed (bundle)
COMPLETENESS="platform/orchestration/actions/check_tool_completeness_action.go"  # confirmed (uploaded)

SAVE_SECTIONS=$(resolve '"save_page_sections"')
VALIDATE_PAGE=$(resolve '"validate_page_content"')
UPDATE_STATUS=$(resolve '"update_page_status"')
TOOL_HEALTH=$(resolve '"check_tool_health"')
LLM_ACTION=$(grep -rl 'func.*ExecuteLLMPrompt\|"execute_llm_prompt"' "$ACT"/*.go 2>/dev/null | grep -v _test | grep -v registry.go | head -1)
LLM_ACTION="${LLM_ACTION#$ROOT/}"
COORD=$(ls platform/orchestration/coordinator.go 2>/dev/null || echo "")

for v in LOAD_EXISTING APPLY_ADOPTION CREATE_TOOL RAG WRITE_PLAN APPEND_NOTE COMPLETENESS \
         SAVE_SECTIONS VALIDATE_PAGE UPDATE_STATUS TOOL_HEALTH LLM_ACTION COORD; do
  printf '  %-14s %s\n' "$v" "${!v:-<not found>}"
done
echo
echo "If any line says <not found>, fix it before continuing (or drop that -scope)."
echo "Press Enter to build the bundle, Ctrl-C to abort."; read -r _

SCOPES=()
add_scope() { [ -n "${1:-}" ] && [ -f "$ROOT/$1" ] && SCOPES+=(-scope "$1"); }
add_scope "$LOAD_EXISTING"
add_scope "$APPLY_ADOPTION"
add_scope "$SAVE_SECTIONS"
add_scope "$VALIDATE_PAGE"
add_scope "$UPDATE_STATUS"
add_scope "$COMPLETENESS"
add_scope "$CREATE_TOOL"
add_scope "$WRITE_PLAN"
add_scope "$APPEND_NOTE"
add_scope "$RAG"
add_scope "$TOOL_HEALTH"
add_scope "$LLM_ACTION"
[ -n "$COORD" ] && SCOPES+=(-scope "$COORD")

DOCS=()
for d in 007_adoption_pipeline_v4.md \
         005_tool_pipeline_1_.md \
         020_tool_lifecycle_2_.md \
         003_contracts_and_standards_7_.md \
         001_development_guide_5_.md \
         019_tool_library_2_.md; do
  p="docs/agent_docs/docs024_key_docs_latest/$d"
  [ -f "$ROOT/$p" ] && DOCS+=(-doc "$p") || echo "WARN: missing doc $p"
done

go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root "$ROOT" \
  -constitution thin_slice_constitution.md -step debug \
  -task "Finish the travelling-docs rollout by recreating a broken game through the system, then clear two chassis-hygiene faults. CONTEXT: travelling docs = doc_plans (PLAN, supersede-versioned by is_current) + doc_notes (append-only), keyed (subject_type, subject_key) where tool -> content_components.function and pipeline -> build|content|design|maintenance. Acceptance criteria live in a fenced criteria JSON block inside the PLAN body. Task 3 (tool-generator writes a PLAN at every tool creation) is APPLIED and PROVEN: run 1923badd produced a doc_plans row source=tool-generator for tool-xp-curve-designer. Task 4 (component-template-fixer, tool-improver, tool-recreation-handler each append a NOTES entry after a successful fix) is APPLIED but UNPROVEN - no machine-written fix note exists yet. IMMEDIATE WORK, in order. (1) Two agent_definitions fixes for tool-recreation-handler, in ONE snapshot-prefixed migration (snapshot_agent(type,reason) first, always): its input_contract declares required [site_id, domain] and optional [page_name, page_id, sections] but the workflow reads input_data.spec.mode, input_data.spec.interactive_features and input_data.spec.function, so add spec to optional; and its append_note step currently uses subject_type tool with subject_key_field input_data.spec.function, which would create a dangling doc because tool-recreation-handler never calls create_tool_component - it ends save_page_sections then update_page_status then deploy_page - so re-subject the note to pipeline/build with note_site_id_field site_record.site_id and drop subject_key_field, mirroring component-template-fixer. (2) Trigger tool-recreation-handler on gamesdesign.co.uk page game-economy-simulator with spec.mode=recreate (apply_adoption_plan_action.go sets exactly this value and load_existing_content checks for it; the crawl is in research_results with result_type adoption_crawl for full markdown plus rawHTML and adoption_page for per-page markdown). Verify first that research_results still holds the rawHtml for that page, because pages.sections is empty (30 bytes of jsonb), page_components has only a shared hero, and the 32KB game body currently exists only as deployed HTML in the sites repo. (3) The recreation must FIX two real bugs in the economy simulator, which become the recreated tool's first acceptance criteria: a logic bug where the tick reads the raw Player-Influx slider index as the players-per-tick rate (popInflux = parseInt(el.slPop.value), slider min=0 max=5), so even Flood adds only about 5 players per tick, roughly 20x too weak (the live screenshot shows 191 players = 50 + 5x28 ticks); and a display bug where the Players chart dataset is bound to yAxisID yGold on a 0 to 100k scale, so around 200 players renders flat on the baseline. The fix maps slider indices to real rates such as 0,1,5,15,40,100 and gives Players its own axis. Pass these as spec.interactive_features so the recreate_tool LLM step produces corrected code. A successful run must leave a doc_notes row with category fix - that is the Task-4 proof. (4) Chassis hygiene: the shared generic chassis pod was OOMKilled (exit 137) 23 seconds into a rag_index step after 2 days 19 hours of uptime; collected_data for that run was only 192 kB with 2 __raw_message__ copies, so state growth is NOT the cause and a slow leak is the leading hypothesis. The coordinator logs the ENTIRE params object at Info on every action (log lines beginning DEBUGaa: params sent to action handler and DEBUGaa: in buildActionParams), which serialises all collected data twice per action and prints headers - remove that logging. rag_index chunks content then calls GenerateEmbedding per chunk against the ollama-adapter with no context deadline; the embedder is healthy (api/tags lists nomic-embed-text and api/embeddings returns a vector) but add a context.WithTimeout so a stall degrades into the existing non-fatal store-without-embeddings path, and consider an action-level deadline so no action outlives workflow timeout_seconds. index_plan is currently BYPASSED in tool-generator (write_plan.next_step = complete) and must be re-enabled once the deadline ships. (5) Stage 5, the static acceptance checker: validate each criteria selector's ANCHOR (its leftmost id or class token) against the component html_template, never the whole path, because criteria are runtime assertions evaluated against the rendered DOM. Worked example: the composer invented #xpTableBody and #statsStrip (neither exists) while #tableWrap is a real but EMPTY div that JS fills, so #tableWrap tr is valid at runtime and unfindable statically. Static checks confirm, never refute; unverifiable checks get dropped or marked with an id ending -EDIT, which checkers skip. Also add the select interaction verb to the Tier-4 vocabulary. HARD-WON RULES: error_step must live inside step Config and name an existing step, and inside loops these values are iteration-prefixed so they must name substeps of that loop; agent_error_log is the FIRST read when a step fails (it outlives the pod, carries step_name and action and error_message; its orchestration_id column is TEXT so do not cast to uuid); a row stuck at EXECUTING_STEP usually means the worker died, so check pod RESTARTS and describe pod Last State before blaming the action; prompt templates receive a text-format LLM output as a BARE STRING (use {{.generated_html}}) and a json-format output as a map (use {{.tool_analysis.result | toJSON}}), while action CONFIG paths keep the .result suffix; snapshot agents with snapshot_agent(type, reason) before every agent_definitions update; check DB schemas before writing SQL; Postgres caps bounded regex repetition at 255 so prefer strpos/substr in guards; an aborted transaction ignores every later command including BEGIN, so ROLLBACK first and run migrations with psql -f rather than pasting. CONSTRAINTS: every agent is an orchestrator; reuse existing functions and structs before creating new ones; keep workflows simple with complexity in Go actions; no sub-workflows, spawn sub-agents instead; use logger.Info not Debug; namespaces are ai-persona-system and kafka; deploys go to the gqls/agentchassis repo via GitHub Actions to Backblaze." \
  "${SCOPES[@]}" \
  -include platform/orchestration/actions/registry.go \
  "${DOCS[@]}" \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables agent_definitions,doc_plans,doc_notes,content_components,pages,page_components,site_work_items,research_results,sites,site_specs,orchestration_states,agent_error_log,knowledge_base \
  -runtime-site gamesdesign.co.uk -runtime-page game-economy-simulator \
  -out /tmp/bundle_recreation.md

echo
echo "Bundle written to /tmp/bundle_recreation.md"
echo "Check it contains: load_existing_content_action.go, the save/validate/status actions,"
echo "the doc actions, rag_actions.go, and a 'Recent errors (agent_error_log)' section."
