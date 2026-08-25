#!/usr/bin/env bash
# probe-page-url.sh — curl a page at the URL the platform RECORDS, with the two
# controls that make the answer mean something.
#
# WHY THIS IS A SCRIPT AND NOT A PARAGRAPH. Four sessions composed a page URL by
# hand and read the resulting 404 as damage: 2026-07-27 (/sale vs /sale.html,
# WRONG_CALLS), 2026-08-09 (LANDMINES "A page's served URL is NOT derivable from
# pages.name"), 2026-08-24 (bugs_open/387 — filed "three pages deployed and 404"
# against pages serving 200 at their recorded URL), and 2026-08-25 (the 387
# session itself, mid-investigation of the trap). Two written warnings did not
# stop occurrences three and four; a command might.
#
# WHAT IT DOES, per requested page:
#   1. reads pages.url from the DB (NEVER composes a URL from a name);
#   2. curls https://<domain><url> verbatim;
# and per DOMAIN, in the same run, the two controls:
#   A. an INVENTED url must be non-200 — else the domain is a catch-all and
#      every 200 is meaningless (memory: "a parked domain 200s EVERY path");
#   B. a KNOWN-GOOD sibling (a deployed page NOT among the targets) must be 200
#      — else the domain/hosting is down or the URL FORM itself is broken, and a
#      target's 404 says nothing about the target (the control bug 387's filing
#      lacked: its invented-URL control shared the URL form with its claim).
#
# EXIT CODES (the check-fleet convention): 0 = every target serves 200 with both
# controls holding; 1 = at least one target is NOT serving while both controls
# hold (a real finding); 2 = could not run — DB unreachable, unknown page name,
# or a control failed. 2 is a REFUSAL, never a pass.
#
# USAGE:
#   scripts/probe-page-url.sh <domain> <page-name> [<page-name>...]
#   scripts/probe-page-url.sh <domain> --all          # every active deployed page
#   scripts/probe-page-url.sh --self-test             # verdict logic, no cluster
set -uo pipefail

NS=ai-persona-system
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F ' ')

fetch_code() { # $1 = full url -> echoes http code ("000" on transport failure)
    curl -s -o /dev/null -w '%{http_code}' --max-time 20 \
         -A 'agentchassis-probe-page-url/1.0 (+bugs_open/387)' "$1"
}

# verdict <target_code> <invented_code> <sibling_code> -> line + return code
#   return 0 = serving, 1 = real absence, 2 = controls do not hold
verdict() {
    local t="$1" inv="$2" sib="$3"
    if [[ "$inv" == "200" ]]; then
        echo "CONTROL-FAIL catch-all: the invented URL returns 200, every 200 here is meaningless"
        return 2
    fi
    if [[ "$sib" != "200" ]]; then
        echo "CONTROL-FAIL sibling: a known-good deployed page returns $sib — the domain or URL form is broken, a target's $t says nothing about the target"
        return 2
    fi
    if [[ "$t" == "200" ]]; then echo "SERVING"; return 0; fi
    echo "NOT SERVING ($t) — and both controls hold, so this is a real absence"
    return 1
}

self_test() {
    local fails=0
    check() { # desc t inv sib want
        local out rc
        out=$(verdict "$2" "$3" "$4"); rc=$?
        if [[ "$rc" != "$5" ]]; then echo "SELF-TEST FAIL: $1 (want rc=$5 got rc=$rc: $out)"; fails=$((fails+1)); fi
    }
    check "healthy page"                       200 404 200 0
    check "real absence"                       404 404 200 1
    check "catch-all domain masks everything"  200 200 200 2
    check "catch-all with absent target"       404 200 200 2
    check "sibling down = cannot judge"        404 404 522 2
    check "sibling transport fail"             404 404 000 2
    check "target 000 with controls holding"   000 404 200 1
    # The property the four wrong calls violated: this script must never build a
    # URL out of a page NAME. Assert the only URL constructions are the DB value
    # and the invented control.
    if grep -nE 'https://\$\{?domain\}?/\$\{?(name|page)' "$0" >/dev/null; then
        echo "SELF-TEST FAIL: a composed name-URL construction exists in this script"; fails=$((fails+1))
    fi
    if [[ $fails -gt 0 ]]; then echo "SELF-TEST: $fails failure(s)"; return 2; fi
    echo "SELF-TEST: all verdict rows + the no-composed-URL property hold"; return 0
}

[[ "${1:-}" == "--self-test" ]] && { self_test; exit $?; }

[[ $# -ge 2 ]] || { echo "usage: $0 <domain> <page-name>... | <domain> --all | --self-test" >&2; exit 2; }
DOMAIN="$1"; shift

# Resolve targets from the DB. Never compose.
if [[ "$1" == "--all" ]]; then
    ROWS=$("${PSQL[@]}" -c "SELECT p.name, p.url FROM pages p JOIN sites s ON s.id=p.site_id
        WHERE s.domain='${DOMAIN}' AND p.status='active' AND p.build_status='deployed'
        ORDER BY p.url") || { echo "REFUSING: DB unreachable" >&2; exit 2; }
else
    NAMES=$(printf "'%s'," "$@"); NAMES=${NAMES%,}
    ROWS=$("${PSQL[@]}" -c "SELECT p.name, p.url FROM pages p JOIN sites s ON s.id=p.site_id
        WHERE s.domain='${DOMAIN}' AND p.name IN (${NAMES})") || { echo "REFUSING: DB unreachable" >&2; exit 2; }
    # every requested name must resolve — an unknown name must not read as a 404
    for want in "$@"; do
        grep -q "^${want} " <<<"$ROWS" || {
            echo "REFUSING: no pages row named '${want}' on ${DOMAIN} — a guessed URL's 404 is indistinguishable from the defect you are hunting" >&2
            "${PSQL[@]}" -c "SELECT '  near match: '||p.name||' -> '||p.url FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='${DOMAIN}' AND p.name ILIKE '%'||left('${want}',12)||'%' LIMIT 5" >&2 || true
            exit 2; }
    done
fi
[[ -n "$ROWS" ]] || { echo "REFUSING: no matching pages rows on ${DOMAIN}" >&2; exit 2; }

# Controls, once per run.
INV_CODE=$(fetch_code "https://${DOMAIN}/zzz-control-not-a-page-${RANDOM}.html")
TARGET_NAMES=$(awk '{print $1}' <<<"$ROWS")
SIB=$("${PSQL[@]}" -c "SELECT p.url FROM pages p JOIN sites s ON s.id=p.site_id
    WHERE s.domain='${DOMAIN}' AND p.status='active' AND p.build_status='deployed'
      AND p.name NOT IN ($(printf "'%s'," $TARGET_NAMES | sed 's/,$//'))
    ORDER BY p.deployed_at DESC NULLS LAST LIMIT 1")
if [[ -z "$SIB" ]]; then
    # every deployed page is a target (--all, or a tiny site): the oldest target
    # doubles as the sibling — weaker, stated.
    SIB=$(awk '{print $2; exit}' <<<"$ROWS")
    echo "note: no non-target deployed page exists; using ${SIB} as the sibling control (weaker)"
fi
SIB_CODE=$(fetch_code "https://${DOMAIN}${SIB}")

RC=0
while read -r name url; do
    [[ -n "$name" ]] || continue
    code=$(fetch_code "https://${DOMAIN}${url}")
    line=$(verdict "$code" "$INV_CODE" "$SIB_CODE"); vrc=$?
    printf '%-28s %-50s %s %s\n' "$name" "$url" "$code" "$line"
    [[ $vrc -gt $RC ]] && RC=$vrc
done <<<"$ROWS"
echo "controls: invented=${INV_CODE} (want non-200)  sibling ${SIB}=${SIB_CODE} (want 200)"
exit $RC
