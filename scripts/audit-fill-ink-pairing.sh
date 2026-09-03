#!/bin/bash
# Read-only. Which component templates ink a PALETTE FILL with the PAGE GROUND
# instead of that fill's own paired ink?
#
# bugs_open/458. The renderer emits paired inks -- --color-primary-text is the
# ink that goes ON a primary fill; --color-primary-ink is primary re-tinted to
# be legible AS text on the page. A template that fills with var(--color-primary)
# and inks it with var(--color-background) reads as idiomatic CSS and is a
# guaranteed failure on any palette whose primary sits near its ground: 1.04:1
# on ai-agent-orchestration.com, where 9 of 59 palettes score under 3.0:1 for
# that pairing and 7 under 1.25:1 `[MEASURED 2026-09-03]`.
#
# WHY THIS EXISTS AT ALL, and it is the whole point (council round 2,
# bug_historian): migration 732 teaches the pairing rule to the two tool-writing
# prompts, and a prompt rule is TAUGHT, not enforced -- an LLM re-decides it
# every run. Nothing else in the pipeline can see this defect:
#   - check_stylesheet_gutted sees a token's ABSENCE from a stylesheet. The
#     wrong pairing uses valid, PRESENT vars, so it is invisible there.
#   - render_audit.py measures the states a page PAINTS. An .error-msg or
#     :hover rule is never composited, so its colours are never sampled
#     (LANDMINES, "a render-time contrast audit only measures the states the
#     page PAINTS"). One such rule sat at 1.00:1 through two clean audits.
# Reading the TEMPLATE is the only view that sees every rule in every state.
#
# Usage: scripts/audit-fill-ink-pairing.sh [--json] [--census] [--since DAYS]
#   Default: components created in the last 7 days carrying the shape -- i.e.
#   NEW instances, which is what tells you a prompt rule has stopped holding.
#   --census: the full fleet split (tool vs non-tool, by month), no findings.
#   --since N: widen or narrow the default window.
# Exit:  0 = no findings (or census-only) - 1 = findings - 2 = could not determine
#
# NOTE on exit codes (LANDMINES.md, `go run` collapses the child's exit status):
# a refusal is discriminated by EMPTY OUTPUT where JSON belongs, never by
# branching on exit code 2.

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
JSON_OUT=0
CENSUS=0
SINCE_DAYS=7

while [[ $# -gt 0 ]]; do
    case "$1" in
        --json)   JSON_OUT=1; shift ;;
        --census) CENSUS=1;   shift ;;
        --since)  SINCE_DAYS="${2:-7}"; shift 2 ;;
        *) echo "unknown argument: $1" >&2; exit 2 ;;
    esac
done

psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -At -F'|' -c "$1" 2>/dev/null
}

# The predicate, in one place so the report and the census cannot disagree.
#   fill      : the template paints a background with a palette fill
#   groundInk : and colours text with the page's own ground
#   paired    : but does NOT use that fill's paired ink anywhere
# A template using the paired ink is left alone even if it also uses a ground
# colour elsewhere -- a mixed template is an authoring choice, not the defect.
PREDICATE="html_template ~ 'background:\\s*var\\(--color-(primary|accent)\\)'
       AND html_template ~ 'color:\\s*var\\(--color-(background|surface)\\)'
       AND html_template !~ '--color-(primary|accent)-(text|ink)'"

if [[ "$CENSUS" == "1" ]]; then
    rows=$(psql_q "
      WITH c AS (
        SELECT date_trunc('month', created_at)::date AS month,
               (category ILIKE '%tool%' OR function ILIKE '%tool%' OR name ILIKE 'tool-%') AS is_tool,
               ($PREDICATE) AS defect
        FROM content_components WHERE is_active AND forked_from IS NULL)
      SELECT month, is_tool, count(*) FILTER (WHERE defect), count(*)
      FROM c GROUP BY 1,2 ORDER BY 1,2;")
    if [[ -z "$rows" ]]; then echo "could not reach the database" >&2; exit 2; fi
    printf '%-12s %-6s %8s %8s\n' MONTH TOOL DEFECT TOTAL
    echo "$rows" | awk -F'|' '{printf "%-12s %-6s %8s %8s\n",$1,($2=="t"?"yes":"no"),$3,$4}'
    exit 0
fi

rows=$(psql_q "
  SELECT name, COALESCE(NULLIF(category,''),function), created_at::date
  FROM content_components
  WHERE is_active AND forked_from IS NULL
    AND created_at > now() - interval '$SINCE_DAYS days'
    AND ($PREDICATE)
  ORDER BY created_at DESC, name;")

# A refusal and a clean result must not look alike: probe reachability
# separately rather than reading an empty rowset as 'nothing wrong'.
probe=$(psql_q "SELECT count(*) FROM content_components WHERE is_active;")
if [[ -z "$probe" ]]; then
    if [[ "$JSON_OUT" == "1" ]]; then :; else echo "could not reach the database" >&2; fi
    exit 2
fi

n=0
[[ -n "$rows" ]] && n=$(echo "$rows" | wc -l | tr -d ' ')

if [[ "$JSON_OUT" == "1" ]]; then
    printf '{"window_days":%s,"findings":%s,"components":[' "$SINCE_DAYS" "$n"
    first=1
    while IFS='|' read -r nm cat dt; do
        [[ -z "$nm" ]] && continue
        [[ $first -eq 0 ]] && printf ','
        printf '{"name":"%s","category":"%s","created":"%s"}' "$nm" "$cat" "$dt"
        first=0
    done <<< "$rows"
    printf ']}\n'
else
    if [[ "$n" == "0" ]]; then
        echo "no component created in the last $SINCE_DAYS days inks a palette fill with the page ground."
        echo "(the shape is still present in older rows -- run --census for the standing total)"
    else
        echo "$n component(s) created in the last $SINCE_DAYS days ink a palette fill with the page ground:"
        echo "$rows" | awk -F'|' '{printf "  %-46s %-22s %s\n",$1,$2,$3}'
        echo
        echo "Each fills with var(--color-primary|accent) and colours text var(--color-background|surface),"
        echo "using no paired ink. Repair: text ON a fill is var(--color-primary-text, #fff);"
        echo "a palette colour used AS text is var(--color-primary-ink, var(--color-primary))."
        echo "If these post-date migration 732, the prompt rule is not holding -- that is the finding."
    fi
fi

[[ "$n" == "0" ]] && exit 0
exit 1
