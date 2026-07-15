#!/usr/bin/env bash
# bundle_minilobby_trim.sh — primed cmd/bundle request for the provocation-card
# mini-lobby trim (vonc.com / Spark).
#
# WHY A BUNDLE AND NOT JUST SQL
# -----------------------------
# The trim looks like string surgery, but docs 003/002 say HTML patching was
# REJECTED as an edit mechanism ("content_data is always the source of truth …
# if we only patched rendered_html, the edit would be lost on the next
# re-render"). 003 names two re-render paths — the full path (needs_page →
# page-content-writer, LLM) and the light path (rerender_page_sections behind a
# page_rerender item, no LLM, re-renders each section from stored content_data
# through RenderComponentAction) — and neither is rerender_single_page, which
# only ASSEMBLES pre-rendered components. Meanwhile
# fix_component_template_action already implements a `remove_element` fix_type,
# but its header says it does NOT touch page_components rendered_html content
# because "content changes go through the section-editor workflow".
# So the supported path for this change is genuinely unclear, and guessing it
# means either a lost edit or a needless LLM rebuild. That is a code question →
# a bundle.
#
# READ-ONLY: bundle runs dbcontext (\d, capped SELECTs, existing-log reads) then
# the pure assembler. It triggers nothing — no builds, no spawns, no writes.
#
# USAGE
#   bash bundle_minilobby_trim.sh              # gather + assemble
#   DRY=1 bash bundle_minilobby_trim.sh        # print the commands, run nothing
#   ROOT=/path/to/agentchassis bash bundle_minilobby_trim.sh
#
# PREREQ: /tmp/analysis_repo.json exists (the analyser JSON). Regenerate it the
# way you normally do if it is stale.

set -euo pipefail

ROOT="${ROOT:-$HOME/projects/agentchassis}"
ANALYSIS="${ANALYSIS:-/tmp/analysis_repo.json}"
CONSTITUTION="${CONSTITUTION:-thin_slice_constitution.md}"
OUT="${OUT:-/tmp/bundle_minilobby_trim.md}"
PSQL='kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db'

cd "$ROOT"
[ -f "$ANALYSIS" ] || { echo "missing analyser JSON: $ANALYSIS" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Step 0 — resolvers for the paths that cannot be asserted from outside the repo.
# A wrong -scope silently UNDER-SCOPES the bundle (no error), so each resolver
# must match exactly one file or the script stops and prints the candidates.
# Pattern used: the action's own registration key, which appears in the action
# file's header comment; registry.go is excluded so the key's registration site
# does not count as a second hit.
# ---------------------------------------------------------------------------
resolve_one() {
  local label="$1" pattern="$2" hits n
  hits=$(grep -rlE "$pattern" --include='*.go' platform/ 2>/dev/null \
          | grep -v '/registry\.go$' | sort -u || true)
  n=$(printf '%s\n' "$hits" | grep -c . || true)
  if [ "$n" -eq 1 ]; then
    printf '%s' "$hits"
  else
    { echo "RESOLVER: '$label' matched $n files (need exactly 1):"
      printf '  %s\n' $hits
      echo "  → pin the path by hand in this script and re-run"; } >&2
    return 1
  fi
}

RENDER_COMPONENT=$(resolve_one 'render_component action'            '"render_component"')
RERENDER_SECTIONS=$(resolve_one 'rerender_page_sections action'     '"rerender_page_sections"')
RENDER_SNIPPETS=$(resolve_one 'render_js_snippets_for_site action'  '"render_js_snippets_for_site"')
SECTION_EDIT=$(resolve_one 'section-editor content edit action'     'component_swap')

# Docs directory + version-tolerant doc picks (filenames carry a version suffix).
D=$(ls -d docs/agent_docs/docs024_key_docs_latest 2>/dev/null || true)
[ -n "$D" ] || { echo "docs dir not found — pin D by hand" >&2; exit 1; }
pick() { ls "$D"/$1 2>/dev/null | head -1; }
DOC_CONTRACTS=$(pick '003_contracts_and_standards*.md')   # source-of-truth principle; the TWO re-render paths; function linkage
DOC_ARCH=$(pick '002_system_architecture*.md')            # Section Editor; JS Management (js_snippets)
DOC_DYNAMIC=$(pick '022_dynamic_applications*.md')        # Tier-1 client-side data = the runtime-fill contract
DOC_DEBUG_B=$(pick '016b_debugging_guide*.md')            # our silent-noop / marker / assemble-only entries
DOC_LIFECYCLE=$(pick '020_tool_lifecycle*.md')            # component create/fork/improve propagation

echo "resolved:"
printf '  %s\n' "$RENDER_COMPONENT" "$RERENDER_SECTIONS" "$RENDER_SNIPPETS" "$SECTION_EDIT" \
                "$DOC_CONTRACTS" "$DOC_ARCH" "$DOC_DYNAMIC" "$DOC_DEBUG_B" "$DOC_LIFECYCLE"

TASK="On vonc.com the provocation-card component (6163ff14-9f94-4962-aa19-d2718eabdeb1) must lose its \
mini-lobby — the div.pc-card-grid block and its four article.pc-card children — plus the now-orphaned \
pc-container 'grid-template-columns: 1fr 1fr' media-query rule, and its provocation-card-loader \
js_snippet must lose the data.lobby fill; settle the SUPPORTED path end to end: whether \
fix_component_template's remove_element (or store_generated_component, or the section-editor \
component_swap/content_edit) is the sanctioned way to change content_components.html_template rather \
than a direct SQL UPDATE; which action re-renders page_components.rendered_html from the CHANGED \
template — rerender_page_sections (light path, from stored content_data via RenderComponentAction) \
versus needs_page (full LLM rebuild) versus rerender_single_page (which only assembles pre-rendered \
components) — and specifically what happens for a Mode-B runtime-fill section whose content_data is \
empty or NULL, given 003's rule that a NULL-content_data section escalates the light path to a full \
rebuild; which work-item type and agent raises each path (page_rerender vs needs_page vs \
needs_rerender as raised by store_generated_component's markPagesPendingRebuild); whether \
data-runtime-fill and other attributes hand-added to rendered_html survive a re-render from the \
template; and how a changed js_snippets.js_content reaches /assets/js/snippets.js (is there an action, \
or is direct SQL the only writer, with render_js_snippets_for_site as the only reader)."

ARGS=(
  -analysis "$ANALYSIS"
  -root "$ROOT"
  -constitution "$CONSTITUTION"
  -step debug
  -task "$TASK"

  # --- the reuse candidate: it already has a remove_element fix_type ---
  -scope platform/orchestration/actions/fix_component_template_action.go

  # --- how a template change marks/propagates to dependents (symbols: the file is large) ---
  -scope platform/orchestration/actions/store_generated_component_action.go:StoreGeneratedComponentAction
  -scope platform/orchestration/actions/store_generated_component_action.go:markPagesForRebuild
  -scope platform/orchestration/actions/store_generated_component_action.go:markPagesPendingRebuild

  # --- the three candidate render/assemble paths ---
  -scope "$RENDER_COMPONENT"                                               # template + content_data -> rendered_html
  -scope "$RERENDER_SECTIONS"                                              # light path (no LLM), the 003 mechanism
  -scope platform/orchestration/actions/rerender_single_page_action.go     # assemble-only; also injects cc.js_content; carries our data-runtime-fill exemption

  # --- the section-editor path named by fix_component_template's own header ---
  -scope "$SECTION_EDIT"

  # --- how a snippet change reaches the deployed bundle ---
  -scope "$RENDER_SNIPPETS"

  # --- how rendered sections are persisted (slot_name / ordering / content_data) ---
  -scope platform/orchestration/actions/save_page_sections_action.go

  # --- wiring ---
  -include platform/orchestration/actions/registry.go

  # --- authored context ---
  -doc "$DOC_CONTRACTS"
  -doc "$DOC_ARCH"
  -doc "$DOC_DYNAMIC"
  -doc "$DOC_DEBUG_B"
  -doc "$DOC_LIFECYCLE"

  # --- read-only DB evidence ---
  -psql "$PSQL"
  -schema-tables content_components,page_components,pages,js_snippets,component_versions,site_work_items,sites,site_specs
  -runtime-site vonc.com -runtime-page index

  -out "$OUT"
)

if [ "${DRY:-0}" = "1" ]; then
  ARGS+=(-dry-run)
fi

go run ./cmd/bundle "${ARGS[@]}"
echo "wrote $OUT"

# ---------------------------------------------------------------------------
# NOT in the bundle, because bundles carry code/docs/schema/runtime but not the
# specific row state this change depends on. Collect these alongside it:
#
#   -- does provocation-card's index instance have content_data at all?
#   --   NULL/empty => 003 says the light path escalates to a full rebuild.
#   SELECT pc.slot_name, pc.schema_mode,
#          (pc.content_data IS NULL)            AS content_data_null,
#          COALESCE(jsonb_typeof(pc.content_data),'-') AS cd_type,
#          COALESCE(pc.content_data::text,'')   AS cd,
#          LENGTH(pc.rendered_html)             AS rendered_len
#   FROM page_components pc
#   WHERE pc.page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'
#     AND pc.component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
#
#   -- does the component declare any schema fields (Mode-B => none)?
#   SELECT function, schema_field_count, render_mode,
#          (html_template LIKE '%data-runtime-fill%') AS marker_in_template,
#          (html_template LIKE '%pc-card-grid%')      AS grid_in_template,
#          LENGTH(js_content)                          AS cc_js_len
#   FROM content_components
#   WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
#
#   -- is it still vonc-only? (forked_from IS NULL = shared library row)
#   SELECT COUNT(DISTINCT p.site_id) AS sites
#   FROM page_components pc JOIN pages p ON p.id = pc.page_id
#   WHERE pc.component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
#
#   -- which item types have actually run against this page, and what did they do?
#   SELECT item_type, item_key, status, completed_at
#   FROM site_work_items
#   WHERE page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'
#   ORDER BY created_at DESC LIMIT 10;
#
# Optional extra scopes if the verdict asks for them:
#   -scope platform/orchestration/actions/discovery_checks/check_component_standards.go
#   -neighbour callgraph -max-neighbour 1
# ---------------------------------------------------------------------------
