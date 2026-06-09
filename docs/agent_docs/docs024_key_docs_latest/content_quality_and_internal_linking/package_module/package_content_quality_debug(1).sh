#!/bin/bash
#
# package_content_quality_debug.sh   (was: package_page_build_debug.sh)
#   Self-contained packager for the CONTENT-QUALITY / HERO-CTA debug context.
#   Bundles the content-generation / render / CTA-resolution / site-spec slice
#   of agent-chassis into one context file for an AI assistant, COPIES the
#   working + guideline docs alongside (like the thunder packager), and
#   (optionally) appends a read-only live capture aimed at the CTA defects plus
#   a slim verify of the deploy-pending sections-durability changes.
#
#   Standalone extract of package_module.sh: carries its own wrapper.
#
# Usage:  ./package_content_quality_debug.sh [-o output_dir] [-e env] [-d domain] [--no-live]
# Example: ./package_content_quality_debug.sh -d gamesdesign.co.uk
#          ./package_content_quality_debug.sh --no-live        # code+docs only
#
# Output:  <output_dir>/<environment>_content-quality-debug_context.txt
#          <output_dir>/*.md   (docs copied in)
#
# ---------------------------------------------------------------------
# SCOPE (next task — gamesdesign content quality, lead item first):
#   - Hero CTAs wrong site-wide: every hero -> /contact.html & /services.html;
#     text<->destination mismatch; /services.html is a phantom page.
#     (hero content_component CTA fields; build_render_context;
#      prepare_link_context/available_pages; LLM-vs-template CTA resolution)
#   - Guide copy tool-flavoured (possible embedded interactive demos for guides)
#   - Polish: empty "Browse All" hrefs (*_index_url specs unpopulated /
#     inconsistent sources), brand-suffix card titles, empty footer/contact.
#
# ALSO (verify-on-resume): the deploy-pending sections-durability changes
#   2b / S1 / S2 / Fix A — slim live checks included below.
#
# REUSE-DISCOVERY: keeps registry.go (every registered action), types.go,
#   helpers.go, datahelpers/, input_contracts/, discovery_checks/ so we don't
#   reinvent an existing CTA/link resolver. STEP ZERO before writing.
#
# NOTE ON CODE PATHS: a few content/render action filenames below are
#   best-effort (marked "verify"). write_file silently skips a missing path,
#   so an unconfirmed name costs nothing but a gap — confirm against the tree.
# ---------------------------------------------------------------------

set -e

# --- Self-locating logic ---------------------------------------------
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT="${PROJECT_ROOT:-$(realpath "$SCRIPT_DIR/../../")}"
if [ ! -f "$PROJECT_ROOT/go.mod" ] && [ -f "$PWD/go.mod" ]; then
  PROJECT_ROOT="$PWD"
fi
cd "$PROJECT_ROOT"

# --- Configuration ---------------------------------------------------
DEFAULT_OUTPUT_DIR="$SCRIPT_DIR/docs/agent_docs/docs024_key_docs_latest/adoption/docubundle/package_module/output_contexts"
ENVIRONMENT="production"               # only affects the output filename
COMPONENT_NAME="content-quality-debug" # fixed; this script packages one module
DOMAIN="gamesdesign.co.uk"             # site for the live capture
WITH_LIVE=true                         # append read-only live capture if kubectl present
# Where markdown docs live (hint for the copy step). Empty/wrong => find-by-basename.
DOCS_SEARCH_ROOT="${DOCS_SEARCH_ROOT:-$PROJECT_ROOT}"

EXCLUDE_FILES=(
  "platform/orchestration/datahelpers/content_search.go"
  "platform/orchestration/datahelpers/deep_search.go"
  "platform/orchestration/datahelpers/file_extractor.go"
  "platform/orchestration/datahelpers/web_architecture_helpers.go"
  "platform/orchestration/datahelpers/debug_collected_data.go"
  "platform/orchestration/datahelpers/duplicate_logger.go"
  "platform/orchestration/datahelpers/sql_helpers.go"
  "platform/orchestration/datahelpers/action_inputs_example.md"
)

# --- Argument parsing ------------------------------------------------
OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do
  case $1 in
    -o | --output)      shift; OUTPUT_DIR="$1" ;;
    -e | --environment) shift; ENVIRONMENT="$1" ;;
    -d | --domain)      shift; DOMAIN="$1" ;;
    --no-live)          WITH_LIVE=false ;;
    -h | --help)
      echo "Usage: $0 [-o output_dir] [-e environment] [-d domain] [--no-live]"
      echo "Packages the content-quality / hero-CTA code context + docs, plus a"
      echo "read-only live capture (CTA defects + durability verify) unless --no-live."
      exit 0
      ;;
  esac
  shift
done

# --- Helper functions (verbatim from package_module.sh) --------------
function write_file() {
  local file_path=$1; local output_file=$2; local list_only=$3
  if [ -f "$file_path" ]; then
    echo "filepath = ./$file_path" >> "$output_file"
    if [ "$list_only" = "true" ]; then
      echo "[File listed only - content not included]" >> "$output_file"
    else
      cat "$file_path" >> "$output_file"
    fi
    echo "-------------------------------------------------" >> "$output_file"
  fi
}

function is_excluded() {
  local f="${1#./}"; local ex
  for ex in "${EXCLUDE_FILES[@]}"; do
    ex="${ex#./}"
    if [ "$f" = "$ex" ]; then return 0; fi
  done
  return 1
}

function write_directory() {
  local dir_path=$1; local output_file=$2
  if [ ! -d "$dir_path" ]; then
    echo "Warning: Directory '$dir_path' not found in '$PWD'. Skipping." >&2
    return
  fi
  dir_path="${dir_path%/}"
  while IFS= read -r -d $'\0' file; do
    if [ "$(realpath "$file" 2>/dev/null)" = "$(realpath "$output_file" 2>/dev/null)" ]; then
      continue
    fi
    if is_excluded "$file"; then continue; fi
    write_file "$file" "$output_file" "false"
  done < <(find "$dir_path" -type f \
    -not -path '*/.git/*' \
    -not -path '*/node_modules/*' \
    -not -path '*/dist/*' \
    -not -path '*/build/*' \
    -not -path '*/vendor/*' \
    -not -path '*/output_contexts/*' \
    -not -name '*.log' \
    -not -name '*.zip' -not -name '*.tar' -not -name '*.gz' \
    -not -name 'go.sum' \
    -not -name '*_test.go' \
    -not -name '.DS_Store' \
    -print0)
}

function process_module_files() {
  local item=$1; local output_file=$2
  if [ -f "$item" ]; then
    write_file "$item" "$output_file" "false"
  elif [ -d "$item" ]; then
    write_directory "$item" "$output_file"
  fi
}

# --- Module definition: content-quality / hero-CTA -------------------
MODULE_DIRS=(
  # Shared types + contracts + helpers (reuse-discovery)
  "platform/orchestration/types/"
  "platform/orchestration/input_contracts/"
  "platform/orchestration/actioncheck/"
  "platform/orchestration/datahelpers/"
  # Discovery checks (existing audits — incl. any CTA/link/empty checks to extend)
  "platform/orchestration/actions/discovery_checks/"
)

MODULE_FILES=(
  # --- Orchestration engine ---
  "platform/orchestration/coordinator.go"
  "platform/orchestration/state.go"
  "platform/orchestration/helpers.go"
  "platform/orchestration/agent_error_log.go"

  # --- Action catalogue + shared (REUSE-DISCOVERY) ---
  "platform/orchestration/actions/registry.go"
  "platform/orchestration/actions/types.go"
  "platform/orchestration/actions/helpers.go"

  # --- Content generation + render + CTA resolution (the lead-item core) ---
  # The page-content-writer loop: render context, link context, per-section
  # render/generate, page compile. CTA url/text resolves somewhere in here +
  # the hero component template + site_specs.
  "platform/orchestration/actions/build_render_context_action.go"      # verify name — render_context (company/email/phone, available_pages)
  "platform/orchestration/actions/prepare_link_context_action.go"      # verify name — available-pages / internal-link constraint (CTA destinations, /services.html phantom)
  "platform/orchestration/actions/render_component_action.go"          # verify name — merges resolved_data + content into component template
  "platform/orchestration/actions/compile_page_sections_action.go"    # verify name — assembles section HTML into page
  "platform/orchestration/actions/execute_llm_prompt_action.go"        # verify name — the content-writer LLM call (CTA text)
  "platform/orchestration/actions/site_spec_actions.go"                # read_site_spec; identity/navigation/blog/content_direction (*_index_url sources)
  "platform/orchestration/actions/validate_page_content.go"            # link/placeholder/contamination checks
  "platform/orchestration/actions/load_page_record_action.go"
  "platform/orchestration/actions/load_existing_content_action.go"

  # --- Hero CTA resolution (the lead item is NOT greenfield — there is an existing
  #     hero resolver per HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler).
  #     reconcile_section_data is the section-data half; the hero-resolver action's
  #     exact filename is TBD from that handoff — ADD IT HERE once confirmed. ---
  "platform/orchestration/actions/reconcile_section_data_action.go"    # section-data reconciler (June-02 work)
  # "platform/orchestration/actions/<hero_resolver>_action.go"          # TODO: confirm name from HANDOFF_2026-06-02

  # --- Section planning / data resolution (CTA fields are component fields) ---
  "platform/orchestration/actions/load_page_sections_from_spec_action.go" # 2b lives here (sibling fallback)
  "platform/orchestration/actions/plan_sections_action.go"                # resolves component input_schema fields (incl. cta_url source/fallback)
  "platform/orchestration/actions/save_page_sections_action.go"
  "platform/orchestration/actions/queryresolve/queryresolve.go"           # resolvePagesWhereType (list items / Browse-All targets)

  # --- Render + deploy (assembly into final HTML/git) ---
  "platform/orchestration/actions/render_site_components_action.go"       # header/footer/head slots (footer brand/contact defect)
  "platform/orchestration/actions/rerender_pages_actions.go"
  "platform/orchestration/actions/rerender_single_page_action.go"
  "platform/orchestration/actions/get_pages_for_rerender_action.go"
  "platform/orchestration/actions/git_deployer_actions.go"

  # --- Site/page DB + plan reference ---
  "platform/orchestration/actions/site_db_actions.go"                     # ensure_site_record; upsertPage (pages.url/slug — phantom-page check)
  "platform/orchestration/actions/v3_site_actions.go"                     # update_page_status; ValidateSitePlanAction; reconcilePlanWithRealised
  "platform/orchestration/actions/write_site_plan_action.go"

  # --- Work item + dispatch + lifecycle (Fix A / S2 context) ---
  "platform/orchestration/actions/load_work_item_actions.go"              # CompleteWorkItemAction (Fix A guard); fail_work_item
  "platform/orchestration/actions/claim_work_item_action.go"
  "platform/orchestration/actions/dispatch_actions.go"
  "platform/orchestration/actions/workflow_actions.go"                    # complete_workflow (complete_error == success: Fix B target)
  "platform/orchestration/actions/conditional_branch_action.go"           # check_has_ready_sections
  "platform/orchestration/actions/spawn_actions.go"
  "platform/orchestration/actions/call_agent.go"
  "platform/orchestration/actions/triage_detect_items_action.go"          # detected -> triaged (S1 path)

  # --- DB ---
  "platform/database/postgres.go"

  # --- Agent implementation (page-content-writer behaviour: prompts, CTA text) ---
  "internal/agents/contentcreator/"

  # --- Entry point ---
  "cmd/agent-chassis/main.go"

  # --- Configs / deployment (runtime context) ---
  "configs/agent-chassis.yaml"
  "deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml"
  "deployments/kustomize/services/agent-chassis/base/deployment.yaml"
)

# --- Package the code ------------------------------------------------
mkdir -p "$OUTPUT_DIR"
OUTPUT_FILE="${OUTPUT_DIR}/${ENVIRONMENT}_${COMPONENT_NAME}_context.txt"
> "$OUTPUT_FILE"

echo "Packaging '$COMPONENT_NAME' for environment '$ENVIRONMENT'"
echo "  project root: $PROJECT_ROOT"
echo "  output file:  $OUTPUT_FILE"

for dir in "${MODULE_DIRS[@]}"; do
  write_directory "$dir" "$OUTPUT_FILE"
done
for item in "${MODULE_FILES[@]}"; do
  process_module_files "$item" "$OUTPUT_FILE"
done

# --- Copy the working + guideline docs alongside the context file ----
# DOCS_FILES: basenames (or repo-relative paths). Located via literal path,
# else find-by-basename under DOCS_SEARCH_ROOT, then copied into the package dir.
#
# >>> CONFIRMED 2026-06-09 from the repo doc listing. find-by-basename fallback
#     still applies, so the docs024 paths needn't be exact. <<<
DOCS_FILES=(
  # --- this session's thread docs ---
  "HANDOFF_2026-06-09_sections_durability_and_content_quality.md"
  "running_notes_15_skinner_box_and_adoption_sections(9).md"
  "FOCUS_page_build_handler_silent_completion.md"
  "RUNBOOK_section_sectionless_durability.md"
  "docs/agent_docs/docs024_key_docs_latest/adoption/HANDOFF_2026-06-06_guide_list_and_skinner_box.md"

  # --- THE authority for the lead item (hero CTAs) — read FIRST; hero work is NOT greenfield ---
  "docs/agent_docs/docs024_key_docs_latest/HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md"

  # --- the gamesdesign defect list (task source of record) ---
  "docs/agent_docs/docs024_key_docs_latest/adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(9).md"

  # --- component generation / regeneration (hero component fix path) ---
  "026_component_regeneration_flow(1).md"
  "FOCUS_llm_reliability_for_component_generation.md"

  # --- design-fidelity background (April — VERIFY against current state; pipeline evolved since) ---
  "FOCUS_design_and_styling_adoption_problems.md"
  "FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md"
  "PLAN_design-note-recommendation-specialists.md"

  # --- guide interactive demos (defect item 2) ---
  "docs/agent_docs/docs024_key_docs_latest/019_tool_library.md"
  "docs/agent_docs/docs024_key_docs_latest/020_tool_lifecycle.md"

  # --- guidelines (the constitution) — confirmed current versions ---
  "docs/agent_docs/docs024_key_docs_latest/001_development_guide(3).md"
  "docs/agent_docs/docs024_key_docs_latest/002_system_architecture.md"
  "docs/agent_docs/docs024_key_docs_latest/003_contracts_and_standards.md"
  "docs/agent_docs/docs024_key_docs_latest/016_debugging_guide_v2.md"
)

echo "Copying docs into the package dir…"
for d in "${DOCS_FILES[@]}"; do
  src=""
  if [ -f "$d" ]; then
    src="$d"
  else
    src=$(find "$DOCS_SEARCH_ROOT" -type f -name "$(basename "$d")" \
          -not -path '*/.git/*' -not -path '*/output_contexts/*' 2>/dev/null | head -1)
  fi
  if [ -n "$src" ] && [ -f "$src" ]; then
    cp -f "$src" "$OUTPUT_DIR/$(basename "$d")"
    echo "  + $(basename "$d")"
  else
    echo "  ! MISSING: $d (confirm filename/version and add manually)" >&2
  fi
done

# --- Optional live capture (read-only) -------------------------------
if [ "$WITH_LIVE" = true ] && command -v kubectl >/dev/null 2>&1; then
  echo "Capturing live data (read-only; disable with --no-live)…"
  set +e
  PG="kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db"

  {
    echo ""
    echo "================================================================="
    echo "LIVE CAPTURE — $(date -u +%Y-%m-%dT%H:%M:%SZ)  domain=$DOMAIN"
    echo "read-only: \\d schema, SELECTs, kubectl get"
    echo "================================================================="
  } >> "$OUTPUT_FILE"

  # 1) SCHEMA
  { echo ""; echo "----- SCHEMA (\\d) -----"; } >> "$OUTPUT_FILE"
  $PG >> "$OUTPUT_FILE" 2>&1 <<'SQL'
\d pages
\d page_components
\d site_components
\d site_specs
\d content_components
\d site_work_items
\d agent_definitions
SQL

  # 2) CONTENT-QUALITY / CTA DIAGNOSTICS (site_id resolved inline)
  { echo ""; echo "----- CTA / CONTENT-QUALITY (domain=$DOMAIN) -----"; } >> "$OUTPUT_FILE"
  $PG >> "$OUTPUT_FILE" 2>&1 <<SQL
\echo '== site id =='
SELECT id, domain FROM sites WHERE domain='$DOMAIN';

\echo '== hero component CTA schema (where cta_url/cta_text resolve from) =='
SELECT name,
       jsonb_pretty(input_schema->'fields'->'cta_url')         AS cta_url_field,
       jsonb_pretty(input_schema->'fields'->'cta_text')        AS cta_text_field,
       jsonb_pretty(input_schema->'fields'->'cta_secondary_url')  AS cta2_url_field,
       jsonb_pretty(input_schema->'fields'->'cta_secondary_text') AS cta2_text_field
FROM content_components WHERE name ILIKE '%hero%';

\echo '== *_index_url + nav/blog spec sources for "Browse All" buttons =='
SELECT aspect, jsonb_pretty(data) AS data
FROM site_specs
WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN') AND is_current=true
  AND aspect IN ('identity','navigation','blog','content_direction');

\echo '== pages: real names/urls (to spot the /services.html phantom + valid CTA targets) =='
SELECT name, page_type, role, slug, url, build_status
FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN') ORDER BY name;

\echo '== is /services.html a real page? (expected: no row) =='
SELECT name, url FROM pages
WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN')
  AND (url ILIKE '%services%' OR slug ILIKE '%service%' OR name ILIKE '%service%');

\echo '== sample rendered hero HTML per page (see the wrong CTAs; truncated) =='
SELECT p.name, pc.slot_name, left(pc.rendered_html, 1400) AS hero_html
FROM pages p JOIN page_components pc ON pc.page_id=p.id AND pc.slot_name ILIKE '%hero%'
WHERE p.site_id=(SELECT id FROM sites WHERE domain='$DOMAIN')
ORDER BY p.name LIMIT 12;

\echo '== list-hub cards: brand-suffix titles + empty Browse-All hrefs (truncated) =='
SELECT p.name, pc.slot_name, left(pc.rendered_html, 1600) AS html
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id=(SELECT id FROM sites WHERE domain='$DOMAIN')
  AND p.name IN ('guides-index','tools-index','games-index')
  AND (pc.slot_name ILIKE '%list%' OR pc.slot_name ILIKE '%grid%')
ORDER BY p.name;

\echo '== footer/site-level components (empty brand tagline / contact) =='
SELECT slot_name, left(rendered_html, 1500) AS html
FROM site_components
WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN')
  AND slot_name IN ('footer','header','head') ORDER BY slot_name;
SQL

  # 3) DURABILITY VERIFY (the deploy-pending 2b/S1/S2/Fix A changes)
  { echo ""; echo "----- SECTIONS-DURABILITY VERIFY -----"; } >> "$OUTPUT_FILE"
  $PG >> "$OUTPUT_FILE" 2>&1 <<SQL
\echo '== skinner-box still built? =='
SELECT name, build_status, sections FROM pages
WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN') AND name='guide-skinner-box';

\echo '== silent-completion monitor (should be 0 rows) =='
SELECT wi.id, wi.item_type, wi.status, left(coalesce(wi.error,''),80) AS err
FROM site_work_items wi
JOIN pages p ON p.site_id=wi.site_id AND p.name=wi.spec->>'page_name'
WHERE wi.site_id=(SELECT id FROM sites WHERE domain='$DOMAIN')
  AND wi.status='complete'
  AND wi.item_type IN ('needs_content_page','content_rewrite','needs_page')
  AND wi.completed_at > NOW() - INTERVAL '7 days'
  AND NOT EXISTS (SELECT 1 FROM page_components pc
                  WHERE pc.page_id=p.id AND pc.component_id IS NOT NULL
                    AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> '');

\echo '== any flagged needs_human_review items (S2/Fix A working) =='
SELECT item_type, status, left(coalesce(error,''),90) AS err, count(*)
FROM site_work_items
WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN')
  AND status='needs_human_review'
GROUP BY item_type, status, err;

\echo '== sectionless_pages check enabled in completeness-discovery-agent? =='
SELECT default_config #> '{workflow,steps,run_checks,config,checks}' AS checks
FROM agent_definitions
WHERE type='completeness-discovery-agent' AND deleted_at IS NULL;
SQL

  # 4) AGENT WORKFLOWS (page-content-writer is the CTA-text producer)
  { echo ""; echo "----- AGENT WORKFLOWS (agent_definitions.default_config) -----"; } >> "$OUTPUT_FILE"
  $PG -A -t >> "$OUTPUT_FILE" 2>&1 <<'SQL'
SELECT type, default_config
FROM agent_definitions
WHERE type IN ('page-content-writer','page-build-handler','page-rerender','content-quality-auditor','content-gap-planner')
ORDER BY type;
SQL

  # 5) RUNTIME
  { echo ""; echo "----- RUNTIME (kubectl) -----"; } >> "$OUTPUT_FILE"
  echo "\$ kubectl -n ai-persona-system get pods | grep -iE 'dispatch|build|rerender|content|pipeline'" >> "$OUTPUT_FILE"
  kubectl -n ai-persona-system get pods 2>&1 | grep -iE 'dispatch|build|rerender|content|pipeline' >> "$OUTPUT_FILE" 2>&1

  set -e
  echo "  live capture appended."
elif [ "$WITH_LIVE" = true ]; then
  echo "Skipping live capture: kubectl not found on PATH (run with --no-live to silence)."
fi

echo "✅ Done. Context saved to $OUTPUT_FILE"
echo "   Docs   : $(ls -1 "$OUTPUT_DIR"/*.md 2>/dev/null | wc -l | tr -d ' ') markdown file(s) in $OUTPUT_DIR"
FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
echo "📦 File size: $FILE_SIZE"
