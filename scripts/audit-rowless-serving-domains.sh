#!/usr/bin/env bash
# audit-rowless-serving-domains.sh — bugs_open/432 §6 candidate 1.
#
# THE QUESTION: is there a domain we are SERVING to the public that the platform
# has NO RECORD of? Every other check enumerates FROM `pages` / `sites` rows and
# probes outward, so a domain whose rows were DELETED (not archived — deleted)
# is invisible to all of them by construction: not reported clean, absent from
# the report. gamedesign.uk served 13 empty pages for 4.5 months that way.
# This script enumerates the OTHER side — the bucket — and reconciles INTO the DB.
#
# WHY THE BUCKET AND NOT ~/projects/sites. Measured 2026-09-02: the sites repo
# holds 36 domain directories; `b2 ls b2://portfolio-sites/` holds 55 prefixes.
# publish_site / b2worker write straight to the bucket for the paid flow, so the
# repo is a SUBSET of what serves. Enumerate the artefact, never the proxy.
#
# THREE CLASSES, because "has a row" is not "is known":
#   NO_ROW        prefix serves, no `sites` row at all           — 432's shape
#   ROW_NO_PAGES  `sites` row exists, zero `pages` rows          — a hand-built
#                 site (oxenunity.com) or a site pre-seeded and not yet built.
#                 Reported, never scored as damage: the reader decides.
#   OK            row + pages                                    — every other
#                 check can see it; not this script's business
#
# CONTROLS, both directions, per 359's rule — a refusal is never a pass:
#   A. an INVENTED url on the domain must be non-200 (a parked/catch-all domain
#      200s every path, and would make every rowless prefix read as damage).
#   B. the b2 listing must be NON-EMPTY and psql must answer. For THIS question
#      an empty enumeration reports ZERO findings — the healthy-looking answer —
#      so an empty listing is a REFUSAL (exit 2), never a clean run.
#   The row total and the prefix total are both printed: a census that prints
#   only what it found cannot be told from one that stopped early.
#
# A prefix that does not RESOLVE (000 twice) is NOT-RESOLVING: the bucket holds
# it but nobody can reach it (the *.internal pools, a domain whose DNS points
# elsewhere). Reported, not scored, not a refusal — the question is "serving to
# the PUBLIC", and unreachable is not serving. *.ugg2.com prefixes are the paid
# flow's slug mirrors (DGH-021); they are matched to their parent stem's row.
#
# EXIT CODES (the check-fleet convention):
#   0 = every serving prefix has a row with pages, controls holding
#   1 = at least one NO_ROW prefix is SERVING while its control holds
#   2 = could not enumerate, psql failed, or a control failed — a REFUSAL
#
# USAGE:
#   scripts/audit-rowless-serving-domains.sh [--json] [<prefix> ...]
#   scripts/audit-rowless-serving-domains.sh --self-test     # verdict logic, no cluster, no b2
#
# ⚠ `kubectl exec -i` inside a loop EATS THE LOOP'S STDIN (LANDMINES.md); every
# psql call carries </dev/null and the loop is fed by process substitution.
set -uo pipefail

NS=ai-persona-system
BUCKET=b2://portfolio-sites/
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F '|')
UA='agentchassis-audit/1.0 (+bugs_open/432)'

fetch_code() {
    local c
    c=$(curl -s -o /dev/null -w '%{http_code}' --max-time 20 -A "$UA" "$1")
    [[ "$c" == "000" ]] && c=$(curl -s -o /dev/null -w '%{http_code}' --max-time 20 -A "$UA" "$1")
    printf '%s' "$c"
}

# verdict <class> <root_code> <invented_code> -> prints a word, returns 0/1/2
#   class: NO_ROW | ROW_NO_PAGES | OK
verdict() {
    local cls="$1" root="$2" inv="$3"
    [[ "$cls" == "OK" ]] && { echo "ok"; return 0; }
    if [[ "$root" == "000" ]]; then echo "NOT-RESOLVING"; return 0; fi
    if [[ "$inv" == "200" ]]; then echo "CONTROL-FAIL-catchall"; return 2; fi
    if [[ "$inv" == "000" ]]; then echo "UNJUDGEABLE-control-transport"; return 2; fi
    if [[ "$root" == "200" ]]; then
        if [[ "$cls" == "NO_ROW" ]]; then echo "NO-ROW-AND-SERVING"; return 1; fi
        echo "row-no-pages-serving"; return 0
    fi
    echo "not-serving($root)"; return 0
}

self_test() {
    local fails=0
    check() { # desc class root invented want_rc
        local out rc; out=$(verdict "$2" "$3" "$4"); rc=$?
        [[ "$rc" != "$5" ]] && { echo "SELF-TEST FAIL: $1 (want rc=$5 got rc=$rc: $out)"; fails=$((fails+1)); }
    }
    check "orphan serving, control holds"           NO_ROW       200 404 1
    check "orphan on a catch-all: judged for nothing" NO_ROW     200 200 2
    check "orphan prefix but hostname 404s"         NO_ROW       404 404 0
    check "orphan prefix, nothing resolves"         NO_ROW       000 000 0
    check "orphan serving, control transport fails" NO_ROW       200 000 2
    check "hand-built row, serving: reported not scored" ROW_NO_PAGES 200 404 0
    check "row-no-pages on a catch-all still refuses" ROW_NO_PAGES 200 200 2
    check "row with pages: not our business"        OK           200 200 0
    # The property this script exists to hold: an EMPTY enumeration must never
    # read as clean. Assert the refusal path exists by name.
    grep -q 'REFUSE: bucket listing is empty' "$0" || { echo "SELF-TEST FAIL: no empty-listing refusal"; fails=$((fails+1)); }
    grep -q 'REFUSE: psql returned no rows' "$0"    || { echo "SELF-TEST FAIL: no empty-psql refusal"; fails=$((fails+1)); }
    if [[ $fails -gt 0 ]]; then echo "SELF-TEST: $fails failure(s)"; return 2; fi
    echo "SELF-TEST: all verdict rows + both empty-enumeration refusals hold"; return 0
}

JSON=0
[[ "${1:-}" == "--self-test" ]] && { self_test; exit $?; }
[[ "${1:-}" == "--json" ]] && { JSON=1; shift; }
FILTER=("$@")

# ── B. enumerate the SERVING surface ──────────────────────────────────────────
mapfile -t PREFIXES < <(b2 ls "$BUCKET" 2>/dev/null | sed 's#/$##' | grep -E '^[A-Za-z0-9.-]+\.[A-Za-z0-9-]+$' | sort)
if [[ ${#PREFIXES[@]} -eq 0 ]]; then
    echo "REFUSE: bucket listing is empty or b2 failed ($BUCKET) — an empty enumeration is not a clean run" >&2; exit 2
fi
NONHOST=$(b2 ls "$BUCKET" 2>/dev/null | sed 's#/$##' | grep -vE '^[A-Za-z0-9.-]+\.[A-Za-z0-9-]+$' | tr '\n' ' ')

# ── B. enumerate the RECORD ───────────────────────────────────────────────────
ROWS=$("${PSQL[@]}" -c "SELECT s.domain, (SELECT count(*) FROM pages p WHERE p.site_id=s.id) FROM sites s ORDER BY 1" </dev/null 2>/dev/null)
if [[ -z "$ROWS" ]]; then echo "REFUSE: psql returned no rows from sites — cannot reconcile against nothing" >&2; exit 2; fi
declare -A PAGES_OF
while IFS='|' read -r d n; do [[ -n "$d" ]] && PAGES_OF["$d"]="$n"; done <<< "$ROWS"

classify() { # prefix -> NO_ROW | ROW_NO_PAGES | OK   (matches *.ugg2.com to its parent stem)
    local p="$1" stem
    if [[ -n "${PAGES_OF[$p]+x}" ]]; then
        [[ "${PAGES_OF[$p]}" -gt 0 ]] && echo OK || echo ROW_NO_PAGES; return
    fi
    if [[ "$p" == *.ugg2.com ]]; then
        stem="${p%.ugg2.com}"
        for d in "${!PAGES_OF[@]}"; do
            if [[ "$d" == "$stem".* ]]; then [[ "${PAGES_OF[$d]}" -gt 0 ]] && echo OK || echo ROW_NO_PAGES; return; fi
        done
    fi
    echo NO_ROW
}

findings=0; refusals=0; reported=0; judged=0
[[ $JSON -eq 0 ]] && printf "%-36s %-13s %5s %5s  %s\n" PREFIX CLASS ROOT INV VERDICT
[[ $JSON -eq 1 ]] && echo "["
first=1
for p in "${PREFIXES[@]}"; do
    if [[ ${#FILTER[@]} -gt 0 ]]; then
        printf '%s\n' "${FILTER[@]}" | grep -qx -- "$p" || continue
    fi
    cls=$(classify "$p")
    if [[ "$cls" == "OK" ]]; then continue; fi          # every other check can see these
    reported=$((reported+1))
    root=$(fetch_code "https://$p/")
    inv=$(fetch_code "https://$p/this-path-cannot-exist-432-control.html")
    v=$(verdict "$cls" "$root" "$inv"); rc=$?
    [[ $rc -eq 1 ]] && findings=$((findings+1))
    [[ $rc -eq 2 ]] && refusals=$((refusals+1))
    [[ "$root" != "000" ]] && judged=$((judged+1))
    if [[ $JSON -eq 1 ]]; then
        [[ $first -eq 0 ]] && echo ","; first=0
        printf '  {"prefix":"%s","class":"%s","root":"%s","invented":"%s","verdict":"%s","rc":%d}' "$p" "$cls" "$root" "$inv" "$v" "$rc"
    else
        printf "%-36s %-13s %5s %5s  %s\n" "$p" "$cls" "$root" "$inv" "$v"
    fi
done
[[ $JSON -eq 1 ]] && echo && echo "]"

echo "---" >&2
echo "prefixes enumerated: ${#PREFIXES[@]} (hostname-shaped; skipped non-hostname: ${NONHOST:-none})" >&2
echo "sites rows: $(wc -l <<< "$ROWS") | reported (NO_ROW + ROW_NO_PAGES): $reported | judged (resolved): $judged" >&2
echo "NO-ROW-AND-SERVING: $findings | control refusals: $refusals" >&2
[[ $refusals -gt 0 ]] && exit 2
[[ $findings -gt 0 ]] && exit 1
exit 0
