#!/usr/bin/env bash
# audit-archived-still-serving.sh — bugs_open/359 §6 candidate 1.
#
# THE QUESTION: `pages.status='archived'` means the platform RETIRED this page.
# It does not mean the page stopped being served — retirement sets `status` and
# leaves the deployed artefact where it is; removing the file is a separate act
# (`retract_page_deployment_action.go`, bugs_closed/098). This script asks, of
# every archived page that was ever deployed, whether it is still answering the
# public.
#
# WHY IT IS A SCRIPT AND NOT A `curl` LOOP. The answer is meaningless without
# two controls per domain, and they guard failures in OPPOSITE directions:
#
#   A. an INVENTED url must be non-200. A parked or catch-all domain answers
#      200 on every path, so without this control every archived page on it
#      reads as damage. (memory: "a parked domain 200s EVERY path"; the trap
#      that reversed an architectural conclusion once already.)
#
#   B. a KNOWN-GOOD `active` + shipped sibling must be 200. If the origin is
#      down, every target 404s and the run reports ZERO — and for THIS question
#      zero is the healthy-looking answer. This is the control specific to a
#      detector whose finding is a 200: its blinded state is a FALSE ALL-CLEAR,
#      the opposite profile from check_asset_reference_404, whose blinded state
#      merely under-reports.
#
# Neither control is optional, and a domain that fails either is reported
# CONTROL-FAIL and judged for nothing. A refusal is never a pass.
#
# EXIT CODES (the check-fleet convention):
#   0 = every archived+deployed page is correctly absent, both controls holding
#   1 = at least one archived page is SERVING while both its controls hold
#   2 = could not run, or a control failed on some domain — a REFUSAL, not a pass
#
# USAGE:
#   scripts/audit-archived-still-serving.sh [--json] [<domain> ...]
#   scripts/audit-archived-still-serving.sh --self-test    # verdict logic, no cluster
#
# ⚠ `kubectl exec -i` inside a `while read` loop EATS THE LOOP'S STDIN
# (LANDMINES.md). Every psql call below carries `</dev/null` and the loop is fed
# by process substitution. The row total is printed and must equal the DB count —
# a census that prints only what it found cannot be told from one that stopped
# early.
set -uo pipefail

NS=ai-persona-system
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F '|')
UA='agentchassis-audit/1.0 (+bugs_open/359)'

fetch_code() { curl -s -o /dev/null -w '%{http_code}' --max-time 25 -A "$UA" "$1"; }

# verdict <target> <invented> <sibling> -> prints a word, returns 0/1/2
#   0 = correctly absent, 1 = archived AND serving, 2 = controls do not hold
verdict() {
    local t="$1" inv="$2" sib="$3"
    if [[ "$inv" == "200" ]]; then
        echo "CONTROL-FAIL-catchall"; return 2
    fi
    if [[ "$sib" != "200" ]]; then
        echo "CONTROL-FAIL-origin"; return 2
    fi
    if [[ "$t" == "200" ]]; then echo "ARCHIVED-AND-SERVING"; return 1; fi
    echo "correctly-absent"; return 0
}

self_test() {
    local fails=0
    check() { # desc target invented sibling want_rc
        local out rc
        out=$(verdict "$2" "$3" "$4"); rc=$?
        if [[ "$rc" != "$5" ]]; then
            echo "SELF-TEST FAIL: $1 (want rc=$5 got rc=$rc: $out)"; fails=$((fails+1))
        fi
    }
    check "retired page really gone"                 404 404 200 0
    check "retired page still serving"               200 404 200 1
    check "catch-all masks a real absence"           404 200 200 2
    check "catch-all masks a real finding"           200 200 200 2
    check "origin down: everything looks absent"     404 404 522 2
    check "origin down: sibling transport failure"   404 404 000 2
    check "target transport failure, controls hold"  000 404 200 0
    check "gone with 410"                            410 404 200 0
    # The property this script exists to hold: a 200 on a catch-all domain must
    # NOT be reported as damage, and a 404 on a dead origin must NOT be reported
    # as health. Both are covered above; assert the script never composes a URL
    # from a page NAME (probe-page-url.sh's rule, four wrong calls behind it).
    if grep -nE 'https://\$\{?domain\}?/\$\{?(name|page)' "$0" >/dev/null; then
        echo "SELF-TEST FAIL: a composed name-URL construction exists in this script"
        fails=$((fails+1))
    fi
    if [[ $fails -gt 0 ]]; then echo "SELF-TEST: $fails failure(s)"; return 2; fi
    echo "SELF-TEST: all verdict rows + the no-composed-URL property hold"
    return 0
}

JSON=0
[[ "${1:-}" == "--self-test" ]] && { self_test; exit $?; }
[[ "${1:-}" == "--json" ]] && { JSON=1; shift; }

DOMAIN_FILTER=""
if [[ $# -gt 0 ]]; then
    DOMAIN_FILTER="AND s.domain IN ($(printf "'%s'," "$@" | sed 's/,$//'))"
fi

# The population. Both axes, spelled the way the Go helpers spell them:
#   LIFECYCLE  status <> 'active'  — the platform no longer wants it served
#   BUILD      it has shipped      — so there is an artefact that could still serve
# Never `build_status` alone: archiving leaves the build columns untouched, which
# is the whole reason this gap exists (LANDMINES, "Archiving sets `status`…").
ROWS=$("${PSQL[@]}" -c "
    SELECT s.domain, p.name, p.url
    FROM pages p JOIN sites s ON s.id = p.site_id
    WHERE p.status = 'archived'
      AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed')
      AND COALESCE(s.domain,'') <> ''
      $DOMAIN_FILTER
    ORDER BY s.domain, p.url" </dev/null) || { echo "REFUSING: DB unreachable" >&2; exit 2; }

EXPECTED=$(printf '%s\n' "$ROWS" | grep -c . )
if [[ "$EXPECTED" == "0" ]]; then
    echo "no archived+deployed pages match — nothing to audit"
    exit 0
fi

declare -A INV SIBCODE SIBURL
RC=0 SERVING=0 ABSENT=0 REFUSED=0 SEEN=0

[[ $JSON == 0 ]] && printf '%-30s %-46s %-6s %-6s %-6s %s\n' DOMAIN URL CODE INV SIB VERDICT

while IFS='|' read -r domain name url; do
    [[ -n "$domain" ]] || continue
    SEEN=$((SEEN+1))

    if [[ -z "${INV[$domain]:-}" ]]; then
        INV[$domain]=$(fetch_code "https://${domain}/zzz-359-control-not-a-page-$$-${SEEN}.html")
        # Control B: a page the platform still WANTS live and that HAS shipped.
        s=$("${PSQL[@]}" -c "
            SELECT p.url FROM pages p JOIN sites s ON s.id = p.site_id
            WHERE s.domain = '${domain}' AND p.status = 'active'
              AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed')
              AND p.url NOT LIKE '%#%' AND p.url NOT LIKE '%?%'
            ORDER BY p.deployed_at DESC NULLS LAST LIMIT 1" </dev/null)
        if [[ -n "$s" ]]; then
            SIBURL[$domain]="$s"; SIBCODE[$domain]=$(fetch_code "https://${domain}${s}")
        else
            # No live sibling exists. Control B cannot be run, so nothing on this
            # domain can be judged — stated, never quietly treated as holding.
            SIBURL[$domain]="(none)"; SIBCODE[$domain]="---"
        fi
    fi

    code=$(fetch_code "https://${domain}${url}")
    v=$(verdict "$code" "${INV[$domain]}" "${SIBCODE[$domain]}"); vrc=$?
    case $vrc in
        0) ABSENT=$((ABSENT+1)) ;;
        1) SERVING=$((SERVING+1)); [[ $RC -lt 1 ]] && RC=1 ;;
        2) REFUSED=$((REFUSED+1)); RC=2 ;;
    esac

    if [[ $JSON == 1 ]]; then
        printf '{"domain":"%s","page":"%s","url":"%s","code":"%s","invented_control":"%s","sibling_control":"%s","sibling_url":"%s","verdict":"%s"}\n' \
            "$domain" "$name" "$url" "$code" "${INV[$domain]}" "${SIBCODE[$domain]}" "${SIBURL[$domain]}" "$v"
    else
        printf '%-30s %-46s %-6s %-6s %-6s %s\n' "$domain" "$url" "$code" "${INV[$domain]}" "${SIBCODE[$domain]}" "$v"
    fi
done < <(printf '%s\n' "$ROWS")

# The instrument control: a loop that shells out can be truncated silently at
# exit 0. This total is computed by a different query than the loop consumed.
if [[ "$SEEN" != "$EXPECTED" ]]; then
    echo "REFUSING: audited ${SEEN} rows but the population is ${EXPECTED} — the loop was truncated, do not read the counts above" >&2
    exit 2
fi

if [[ $JSON == 0 ]]; then
    echo
    echo "population ${EXPECTED}: ${SERVING} archived AND serving, ${ABSENT} correctly absent, ${REFUSED} unjudgeable (control failed)"
    [[ $REFUSED -gt 0 ]] && echo "⚠ a CONTROL-FAIL row is a REFUSAL, not a pass — those pages were not judged"
fi
exit $RC
