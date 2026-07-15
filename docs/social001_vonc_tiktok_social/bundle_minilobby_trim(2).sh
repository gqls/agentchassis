#!/usr/bin/env bash
# bundle_minilobby_trim.sh — primed cmd/bundle request for the provocation-card
# mini-lobby trim (vonc.com / Spark).   v2, 2026-07-09 (resolver rewritten)
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
# only ASSEMBLES pre-rendered components. Meanwhile fix_component_template_action
# already implements a `remove_element` fix_type, but its header says it does NOT
# touch page_components rendered_html content because "content changes go through
# the section-editor workflow". So the supported path is genuinely unclear, and
# guessing means either a lost edit or a needless LLM rebuild. That is a code
# question → a bundle.
#
# READ-ONLY: bundle runs dbcontext (\d, capped SELECTs, existing-log reads) then
# the pure assembler. It triggers nothing — no builds, no spawns, no writes.
#
# USAGE
#   bash bundle_minilobby_trim.sh                 # gather + assemble
#   DRY=1 bash bundle_minilobby_trim.sh           # print the commands, run nothing
#   ROOT=/path/to/agentchassis bash bundle_minilobby_trim.sh
#   CK=relative/path/to/contextkit bash bundle_minilobby_trim.sh
#   PIN_RENDER_SNIPPETS=platform/.../x.go bash bundle_minilobby_trim.sh   # override a resolver
#
# NB cmd/bundle lives in the CONTEXTKIT tree, not at the repo root. Per
# RUNBOOK_thin_slice: run from the repo root, invoke ./$CK/cmd/bundle, and pass
# -constitution $CK/thin_slice_constitution.md. -root stays the repo root so the
# assembler resolves -scope/-doc paths against the chassis.
#
# PREREQ: /tmp/analysis_repo.json exists (the analyser JSON).

set -euo pipefail

ROOT="${ROOT:-$HOME/projects/agentchassis}"
CK="${CK:-docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit}"
ANALYSIS="${ANALYSIS:-/tmp/analysis_repo.json}"
CONSTITUTION="${CONSTITUTION:-$CK/thin_slice_constitution.md}"
OUT="${OUT:-/tmp/bundle_minilobby_trim.md}"
PSQL='kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db'

cd "$ROOT"
[ -f "$ANALYSIS" ] || { echo "missing analyser JSON: $ANALYSIS" >&2; exit 1; }
[ -d "$CK/cmd/bundle" ] || { echo "cmd/bundle not found at $ROOT/$CK/cmd/bundle — set CK=<relative path>" >&2; exit 1; }
[ -f "$CONSTITUTION" ] || { echo "constitution not found: $ROOT/$CONSTITUTION" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Step 0 — resolvers. A wrong -scope silently UNDER-SCOPES the bundle (no error),
# so each path is either verified to exist or resolved from the registry.
#
# resolve_by_key: two-stage, and the reason v1 failed. Some action files repeat
# their registration key in a header comment; others (e.g. the js-snippets
# renderer) do not. So: find the key ANYWHERE (registry.go included) → read the
# `Handler:` symbol from that entry → find the file that DEFINES that function.
# ---------------------------------------------------------------------------
resolve_by_key() {
  local label="$1" key="$2" handler hits n reg
  reg=$(grep -rn "\"$key\"" --include='*.go' . 2>/dev/null | head -5)
  if [ -z "$reg" ]; then
    echo "RESOLVER: '$label' — action key \"$key\" not found anywhere under $PWD" >&2
    return 1
  fi
  handler=$(grep -rA4 "\"$key\"" --include='*.go' . 2>/dev/null \
            | grep -oE 'Handler:[[:space:]]*[A-Za-z0-9_]+' | head -1 \
            | sed -E 's/Handler:[[:space:]]*//')
  if [ -z "$handler" ]; then
    { echo "RESOLVER: '$label' — key found but no 'Handler:' symbol near it. Sites:"
      printf '%s\n' "$reg"; } >&2
    return 1
  fi
  hits=$(grep -rlE "^func $handler\(" --include='*.go' . 2>/dev/null | sed 's|^\./||' | sort -u)
  n=$(printf '%s\n' "$hits" | grep -c . || true)
  if [ "$n" -eq 1 ]; then printf '%s|%s' "$hits" "$handler"; return 0; fi
  { echo "RESOLVER: '$label' — handler $handler defined in $n file(s):"
    printf '  %s\n' $hits
    echo "  → pin it:  PIN_<VAR>=<path> bash $0"; } >&2
  return 1
}

# scope_for KEY "path|Handler" -> a -scope value.
# A file dedicated to one action (<key>_action.go) is scoped WHOLE — its helpers
# matter. A SHARED file (e.g. v3_site_actions.go, which also holds
# UpdateSiteDefaults/CompilePageSections) is scoped BY SYMBOL so the bundle is
# not diluted with unrelated actions (B4a: irrelevant content costs attention).
scope_for() {
  local key="$1" pair="$2" path handler base
  path="${pair%%|*}"; handler="${pair##*|}"
  base=$(basename "$path" .go); base="${base%_action}"
  if [ "$base" = "$key" ]; then printf '%s' "$path"; else printf '%s:%s' "$path" "$handler"; fi
}

must_exist() {  # known path, verified rather than assumed
  [ -f "$1" ] || { echo "MISSING expected file: $1 (repo layout changed — pin by hand)" >&2; exit 1; }
  printf '%s' "$1"
}

A=platform/orchestration/actions

FIX_TEMPLATE=$(must_exist "$A/fix_component_template_action.go")
STORE_COMPONENT=$(must_exist "$A/store_generated_component_action.go")
RERENDER_SINGLE=$(must_exist "$A/rerender_single_page_action.go")
REGISTRY=$(must_exist "$A/registry.go")

RENDER_COMPONENT="${PIN_RENDER_COMPONENT:-$(scope_for render_component            "$(resolve_by_key 'render_component action'           'render_component')")}"
RERENDER_SECTIONS="${PIN_RERENDER_SECTIONS:-$(scope_for rerender_page_sections    "$(resolve_by_key 'rerender_page_sections action'   'rerender_page_sections')")}"
RENDER_SNIPPETS="${PIN_RENDER_SNIPPETS:-$(scope_for render_js_snippets_for_site   "$(resolve_by_key 'render_js_snippets_for_site action'  'render_js_snippets_for_site')")}"

# save_page_sections: usually a known file; fall back to the registry if renamed.
if [ -f "$A/save_page_sections_action.go" ]; then
  SAVE_SECTIONS="$A/save_page_sections_action.go"
else
  SAVE_SECTIONS="${PIN_SAVE_SECTIONS:-$(scope_for save_page_sections "$(resolve_by_key 'save_page_sections action' 'save_page_sections')")}"
fi

# Section editor: named by fix_component_template's header, but it may be a
# WORKFLOW (agent_definitions row) rather than a file. Soft-resolve: warn, continue.
SECTION_EDIT_FILES=$(grep -rlE 'component_swap' --include='*.go' . 2>/dev/null | sed 's|^\./||' | sort -u | head -2 || true)
if [ -z "$SECTION_EDIT_FILES" ]; then
  echo "NOTE: no Go file mentions component_swap — the section-editor is likely a workflow" >&2
  echo "      (agent_definitions). The bundle will rely on doc 002's 'Section Editor' section." >&2
fi

# Docs directory + version-tolerant picks (filenames carry a version suffix).
D=$(ls -d docs/agent_docs/docs024_key_docs_latest 2>/dev/null || true)
[ -n "$D" ] || { echo "docs dir not found — pin D by hand" >&2; exit 1; }
pick() { ls "$D"/$1 2>/dev/null | head -1; }
DOC_CONTRACTS=$(pick '003_contracts_and_standards*.md')   # source-of-truth principle; the TWO re-render paths
DOC_ARCH=$(pick '002_system_architecture*.md')            # Section Editor; JS Management (js_snippets)
DOC_DYNAMIC=$(pick '022_dynamic_applications*.md')        # Tier-1 client-side data = the runtime-fill contract
DOC_DEBUG_B=$(pick '016b_debugging_guide*.md')            # silent-noop / marker / assemble-only entries
DOC_LIFECYCLE=$(pick '020_tool_lifecycle*.md')            # component create/fork/improve propagation

echo "=== resolved ==="
printf '  %-24s %s\n' \
  fix_component_template "$FIX_TEMPLATE" \
  store_generated_comp   "$STORE_COMPONENT" \
  rerender_single_page   "$RERENDER_SINGLE" \
  render_component       "$RENDER_COMPONENT" \
  rerender_page_sections "$RERENDER_SECTIONS" \
  render_js_snippets     "$RENDER_SNIPPETS" \
  save_page_sections     "$SAVE_SECTIONS"
[ -n "$SECTION_EDIT_FILES" ] && printf '  %-24s %s\n' section_editor "$(echo $SECTION_EDIT_FILES | tr '\n' ' ')"
echo ""

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

  # the reuse candidate: it already has a remove_element fix_type
  -scope "$FIX_TEMPLATE"

  # how a template change marks/propagates to dependents (symbols: the file is ~49KB)
  -scope "$STORE_COMPONENT:StoreGeneratedComponentAction"
  -scope "$STORE_COMPONENT:markPagesForRebuild"
  -scope "$STORE_COMPONENT:markPagesPendingRebuild"

  # the three candidate render/assemble paths
  -scope "$RENDER_COMPONENT"      # template + content_data -> rendered_html
  -scope "$RERENDER_SECTIONS"     # light path (no LLM) — the 003 mechanism
  -scope "$RERENDER_SINGLE"       # assemble-only; injects cc.js_content; carries our data-runtime-fill exemption

  # how a snippet change reaches the deployed bundle
  -scope "$RENDER_SNIPPETS"

  # how rendered sections are persisted (slot_name / ordering / content_data)
  -scope "$SAVE_SECTIONS"

  # wiring
  -include "$REGISTRY"

  # authored context
  -doc "$DOC_CONTRACTS"
  -doc "$DOC_ARCH"
  -doc "$DOC_DYNAMIC"
  -doc "$DOC_DEBUG_B"
  -doc "$DOC_LIFECYCLE"

  # read-only DB evidence
  -psql "$PSQL"
  -schema-tables content_components,page_components,pages,js_snippets,component_versions,site_work_items,sites,site_specs
  -runtime-site vonc.com -runtime-page index

  -out "$OUT"
)

# section-editor files, if any exist
for f in $SECTION_EDIT_FILES; do ARGS+=(-scope "$f"); done

[ "${DRY:-0}" = "1" ] && ARGS+=(-dry-run)

go run "./$CK/cmd/bundle" "${ARGS[@]}"
echo "wrote $OUT"

# ---------------------------------------------------------------------------
# NOT in the bundle (it carries code/docs/schema/runtime, not this row state).
# Collect alongside it:
#
#   -- D0. THE DECIDING PROBE: does the index instance have content_data?
#   --     NULL/empty => 003 says the light path escalates to a full rebuild.
#   SELECT pc.slot_name, pc.schema_mode,
#          (pc.content_data IS NULL)                   AS content_data_null,
#          COALESCE(jsonb_typeof(pc.content_data),'-') AS cd_type,
#          COALESCE(pc.content_data::text,'')          AS cd,
#          LENGTH(pc.rendered_html)                    AS rendered_len
#   FROM page_components pc
#   WHERE pc.page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
#     AND pc.component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
#
#   -- component schema / markers / component-level js
#   SELECT function, schema_field_count, render_mode,
#          (html_template LIKE '%data-runtime-fill%') AS marker_in_template,
#          (html_template LIKE '%pc-card-grid%')      AS grid_in_template,
#          LENGTH(COALESCE(js_content,''))            AS cc_js_len
#   FROM content_components WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
#
#   -- still vonc-only? (forked_from IS NULL = shared library row)
#   SELECT COUNT(DISTINCT p.site_id) AS sites
#   FROM page_components pc JOIN pages p ON p.id = pc.page_id
#   WHERE pc.component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
#
#   -- which item types have run against this page, and what did they do?
#   SELECT item_type, item_key, status, completed_at
#   FROM site_work_items
#   WHERE page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
#   ORDER BY created_at DESC LIMIT 10;
#
# Optional extra scopes if the verdict asks:
#   -scope platform/orchestration/actions/discovery_checks/check_component_standards.go
#   -neighbour callgraph -max-neighbour 1
# ---------------------------------------------------------------------------
