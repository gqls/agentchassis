#!/bin/bash
# ============================================================================
# 082_submit_domain_unified.sh — one entry point for FRESH and ADOPT submissions
# ============================================================================
# Doc 028's "single pipeline": a blank domain and an adopted domain travel the
# same agent graph, differing only in the input material the classifier has and
# the fidelity dial. They CONVERGE at one work item:
#
#     needs_domain_research  ->  domain-research-classifier
#       -> needs_strategy   -> domain-strategist
#       -> needs_briefing   -> build-briefing-agent
#       -> needs_site_plan  -> build-site-planner
#       -> (cascade) needs_composition -> site-design-planner
#       -> (cascade) needs_design      -> webdesign-agent
#       -> needs_content_page x N       -> page-build-handler
#       -> rerender
#
# This script picks the right ENTRY agent and lets both meet at that handoff:
#
#   FRESH  (no --from)  -> domain-submitter
#         creates the site row, stores contact/mission, queues needs_domain_research.
#         (This is exactly "adoption from the handoff point onward" — no crawl.)
#
#   ADOPT  (--from URL) -> site-adoption-orchestrator
#         spawns site-adoption-agent: crawl URL -> fingerprint -> analyse ->
#         classify_archetype -> content_direction -> apply_adoption_plan (creates
#         the destination site + seeds identity/archetype/design_reference/
#         design_intent) -> nav -> generate_design_intent -> queues
#         needs_domain_research. The classifier then read-and-extends.
#
# Both envelopes mirror 080c (the current adoption trigger): same topic, headers,
# responses_topic and kcat pod naming.
#
# Usage:
#   ./082_submit_domain_unified.sh <domain> [options]
#
# Options:
#   --from <url>        ADOPT: crawl this source URL into <domain>. Omit = fresh build.
#   --fidelity <level>  locked | high | medium | low | new      (see NOTE)
#   --email <email>
#   --phone <phone>
#   --mission <text>    Owner mission brief (fresh builds; weights the classifier)
#   --mission-file <p>  Owner mission brief read from a file (use for long briefs; escaped + single-lined)
#
# Examples:
#   ./082_submit_domain_unified.sh idea.uk --email idea-uk@leopardess.uk
#   ./082_submit_domain_unified.sh gamesdesign.co.uk --from https://gamedesign.uk --fidelity high
#   ./082_submit_domain_unified.sh newthing.com --fidelity new --mission "best-in-class ..."
#
# NOTE on --fidelity (doc 028 "Fidelity"):
#   *** UPDATED 2026-07-30: `locked` IS NOW ACTIVE. The rest of the dial is not. ***
#
#   --fidelity locked  ADOPT ONLY, and it changes the build fundamentally: the site
#     is PRESERVED, not recreated. apply_adoption_plan takes a byte-preserving path
#     (adopt_verbatim.go, concept register ADO-037) that keeps each page's CRAWLED
#     URL verbatim, stores the complete crawled document as one 'ported-page'
#     component with rebuild_policy='owned' + content_data.deploy_mode='verbatim',
#     queues one page_rerender each, and SKIPS the classifier handoff entirely — so
#     no strategy/briefing/plan/design cascade runs and nothing restyles the site.
#     rerender_single_page then ships those bytes untouched. Use this to adopt a
#     hand-built site you intend to KEEP (working tools, indexed URLs). It is the
#     wrong choice if you want the site rebuilt in our own style — that is `high`.
#
#   high | medium | low | new  STILL NOT WIRED, exactly as before: recorded in
#     input_data (for a fresh build it lands in the 'submission' spec the classifier
#     reads) but modulating nothing. Behaviour remains IMPLICIT high. The
#     build_policy / adoption_meta spec aspect and per-item deployed/planned status
#     (doc 028 Phases 2-3) are still unbuilt.
#     Defaults: adopt -> high, fresh -> medium.
#     "new" = no adopted baseline, so fidelity instead means confidence tolerance.
#
#   PLUMBING, because this bit was broken and silent (fixed by migration 274,
#   2026-07-30): call_agent's input_mapping is an ALLOW-LIST, not a passthrough, so
#   until 274 the ADOPT path dropped `fidelity` at the spawn boundary — the
#   orchestrator had it, the spawned site-adoption-agent that actually runs
#   apply_adoption_plan never did, and every run silently took the recreate path.
#   If a locked adoption ever behaves like a rebuild again, check that first:
#     SELECT collected_data->'input_data'->>'fidelity' FROM orchestration_states
#     WHERE owner_agent_type='site-adoption-agent' ORDER BY created_at DESC LIMIT 1;
#   Expect 'locked'. NULL means the mapping regressed, and nothing will error.
#
#   A fresh domain still gets a research step regardless: the classifier always
#   runs in full (web research + synthesis), and now also emits a structured
#   design_intent.palette (classifier migration, 2026-06-20). Best-in-class design
#   research (a design-research sub-agent + best-in-class mission default) is a
#   separate TODO and is not part of this trigger.
# ============================================================================
# tel: +44 (0) 7934 524 911
# email info@contactforsales.com

set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain> [--from URL] [--fidelity L] [--email E] [--phone P] [--mission TXT] [--mission-file PATH]}"
shift || true

SOURCE_URL=""
FIDELITY=""
EMAIL=""
PHONE=""
MISSION=""
MISSION_FILE=""

while [ $# -gt 0 ]; do
  case "$1" in
    --from)         SOURCE_URL="${2:?missing URL}";     shift 2 ;;
    --fidelity)     FIDELITY="${2:?missing level}";     shift 2 ;;
    --email)        EMAIL="${2:?missing email}";        shift 2 ;;
    --phone)        PHONE="${2:?missing phone}";        shift 2 ;;
    --mission)      MISSION="${2:?missing text}";       shift 2 ;;
    --mission-file) MISSION_FILE="${2:?missing path}";  shift 2 ;;
    *) echo "Unknown option: $1" >&2; exit 2 ;;
  esac
done

# --mission-file wins if given. Read the file and make it JSON-safe for the string-concat
# embedding below: escape backslashes then double-quotes, and fold newlines/tabs to spaces so
# the brief becomes a single JSON string value (no jq dependency).
if [ -n "$MISSION_FILE" ]; then
  if [ ! -f "$MISSION_FILE" ]; then echo "mission file not found: $MISSION_FILE" >&2; exit 2; fi
  MISSION="$(sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' "$MISSION_FILE" | tr '\n\t' '  ' | sed 's/  */ /g')"
fi

# Mode + default fidelity (recorded only — see NOTE)
if [ -n "$SOURCE_URL" ]; then
  MODE="adopt"; AGENT="site-adoption-orchestrator"; FIDELITY="${FIDELITY:-high}"
else
  MODE="fresh"; AGENT="domain-submitter";           FIDELITY="${FIDELITY:-medium}"
fi

# Build input_data (string concat, matching 077 — no jq dependency).
# --mission text is embedded as-is (keep it free of " and \); --mission-file is escaped above.
if [ "$MODE" = "adopt" ]; then
  INPUT_DATA="{\"target_url\":\"$SOURCE_URL\",\"destination_domain\":\"$DOMAIN\",\"fidelity\":\"$FIDELITY\"}"
else
  INPUT_DATA="{\"domain\":\"$DOMAIN\",\"fidelity\":\"$FIDELITY\""
  if [ -n "$EMAIL" ];   then INPUT_DATA="${INPUT_DATA},\"email\":\"$EMAIL\""; fi
  if [ -n "$PHONE" ];   then INPUT_DATA="${INPUT_DATA},\"phone\":\"$PHONE\""; fi
  if [ -n "$MISSION" ]; then INPUT_DATA="${INPUT_DATA},\"mission_brief\":{\"text\":\"$MISSION\"}"; fi
  INPUT_DATA="${INPUT_DATA}}"
fi

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Submit (${MODE}): ${DOMAIN}"
echo "  Agent:         ${AGENT}"
[ "$MODE" = "adopt" ] && echo "  Source crawl:  ${SOURCE_URL}"
[ -n "$EMAIL" ]       && echo "  Email:         ${EMAIL}"
echo "  Fidelity:      ${FIDELITY}   (RECORDED ONLY — see NOTE in header)"
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-submit-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=submit-${DOMAIN}-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"$AGENT"},"input_data":$INPUT_DATA}
JSON

echo ""
echo "=== Monitor (psql below = your shortcut / kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db) ==="
echo ""
if [ "$MODE" = "adopt" ]; then
  echo "Find the spawned adopter pod (~10-20s after trigger):"
  echo "  kubectl -n ai-persona-system get pods -l agent_type=site-adoption-agent --sort-by=.metadata.creationTimestamp | tail -3"
  echo "  kubectl -n ai-persona-system logs -f <adopter-pod-name>"
  echo ""
  echo "Provenance (every adopted_from should be the SOURCE host):"
  echo "  SELECT aspect, data->>'adopted_from' AS adopted_from FROM site_specs"
  echo "  WHERE site_id=(SELECT id FROM sites WHERE domain='${DOMAIN}') AND source='adoption' AND is_current=true ORDER BY aspect;"
  echo ""
fi
echo "Orchestration state:"
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;"
echo ""
echo "Site row:"
echo "  SELECT id, domain, status, build_status, style_collection_id FROM sites WHERE domain='${DOMAIN}';"
echo ""
echo "The spec cascade as agents fill it (adoption seeds, classifier supersedes, then strategist/briefing):"
echo "  SELECT aspect, source_agent, is_current FROM site_specs ss JOIN sites s ON s.id=ss.site_id"
echo "  WHERE s.domain='${DOMAIN}' ORDER BY ss.created_at;"
echo ""
echo "Confirm the classifier now emits a STRUCTURED palette (the migration):"
echo "  SELECT data->'palette'->'reference_values' AS palette, data->'typography'->'reference_values' AS typo"
echo "  FROM site_specs ss JOIN sites s ON s.id=ss.site_id"
echo "  WHERE s.domain='${DOMAIN}' AND ss.aspect='design_intent' AND ss.source_agent='domain-research-classifier' AND ss.is_current=true;"
echo ""
echo "Build pipeline work items:"
echo "  SELECT item_type, wi.status, handler_agent, LEFT(summary,60) FROM site_work_items wi JOIN sites s ON s.id=wi.site_id"
echo "  WHERE s.domain='${DOMAIN}' AND pipeline='build' ORDER BY priority;"
echo ""
echo "Anything blocked:"
echo "  SELECT item_type, handler_agent, LEFT(error,80) FROM site_work_items wi JOIN sites s ON s.id=wi.site_id"
echo "  WHERE s.domain='${DOMAIN}' AND status='blocked';"
