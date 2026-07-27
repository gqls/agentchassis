#!/usr/bin/env bash
# Verify the evidence-chart sections on the SERVED pages, against the evidence
# register in the database.
#
# The point of this script is that it shares no logic with the thing it checks.
# It does not read the template, the chart definitions' geometry, or the render
# path. It asks the database "what figure should this chart draw?" and then asks
# the live page "is that figure on you, inside the chart, next to its label?".
# A check that shares its logic with the fix can only agree with it — the
# mistake this workstream made three times on 2026-07-25 (landmine L2).
#
# Usage: verify_evidence_chart_live.sh [domain] [page ...]
#   defaults: fundamentallyai.com index capabilities
#
# Landmine L1: the site serves .html — an extension-less URL 404s.
# Landmine L8: rapid cache-busted probing throttles the origin into 000s and
#              spurious 404s, so each fetch is retried up to 3 times with a pause.

set -uo pipefail
# NOTE: `printf ... | grep -q` is NOT usable here. `grep -q` exits at the first
# match, `printf` then takes SIGPIPE and exits 141, and under `pipefail` the
# pipeline reports 141 — so every check that FOUND its string reported failure
# while every negated check passed. That is a checker that inverts its own
# result on success, which is worse than no checker. Use a here-string.

DOMAIN="${1:-fundamentallyai.com}"
shift || true
PAGES=("${@:-index capabilities}")
if [ "${#PAGES[@]}" -eq 1 ]; then read -r -a PAGES <<< "${PAGES[0]}"; fi

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -F$'\t')

fail=0
note() { printf '%s\n' "$*"; }
check() { # check <ok:0|1> <description>
  if [ "$1" -eq 0 ]; then note "PASS  $2"; else note "FAIL  $2"; fail=$((fail+1)); fi
}

fetch() { # fetch <url> -> body on stdout; retries before condemning (L8)
  local url="$1" body="" code="" i
  for i in 1 2 3; do
    body=$(curl -fsS --max-time 25 "$url" 2>/dev/null) && { printf '%s' "$body"; return 0; }
    code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 25 "$url" 2>/dev/null)
    note "      (attempt $i for $url returned $code; retrying)"
    sleep 3
  done
  return 1
}

note "=== expected, from the evidence register (not from the template) ==="
# chart_id, page, label, value, unit  — one row per point the register says to draw
EXPECTED=$("${PSQL[@]}" -c "
SELECT c->>'id',
       COALESCE(c->'pages'->>0, '*'),
       p->>'label',
       (SELECT f->>'value' FROM jsonb_array_elements(ss.data->'facts') f WHERE f->>'id' = p->>'fact_id'),
       COALESCE(c->>'unit','')
  FROM site_specs ss,
       jsonb_array_elements(ss.data->'charts') c,
       jsonb_array_elements(c->'points') p
 WHERE ss.aspect='evidence_base' AND ss.is_current
   AND ss.site_id=(SELECT id FROM sites WHERE domain='${DOMAIN}');")
printf '%s\n' "$EXPECTED"
[ -n "$EXPECTED" ] || { note "FAIL  the register defines no chart points at all"; exit 1; }

for page in "${PAGES[@]}"; do
  note ""
  note "=== ${DOMAIN}/${page}.html ==="
  html=$(fetch "https://${DOMAIN}/${page}.html") || { check 1 "page fetched"; continue; }
  check 0 "page fetched ($(printf '%s' "$html" | wc -c) bytes)"

  # Asserted only for pages that carry the section — see the `carries` lookup
  # below. Asserting it everywhere reports a FAIL for a page that is correct.
  carries_early=$("${PSQL[@]}" -c "
    SELECT count(*) FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
      JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
     WHERE s.domain='${DOMAIN}' AND p.name='${page}' AND cc.function='evidence-chart';")
  if [ "$carries_early" != "0" ]; then
    grep -q 'data-component="evidence-chart"' <<< "$html"
    check $? "the evidence-chart section is present on the served page"
  fi

  ! grep -q 'ZgotmplZ' <<< "$html"
  check $? "no ZgotmplZ on the page (no value was rejected by the CSS filter)"

  ! grep -qE '\-\-v:[0-9.]*e\+' <<< "$html"
  check $? "no exponent notation in any bar geometry"

  # Does this page actually CARRY the section? Asked of the database, because
  # that — not the register's `pages` key — is what decides today. While
  # bugs_open/085 stands, no page identity reaches a section template, so every
  # chart renders on every page carrying the section and the `pages` key is
  # inert. When 085 is fixed, swap this back to the per-page expectation: the
  # register already holds the right answer.
  carries=$("${PSQL[@]}" -c "
    SELECT count(*) FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
      JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
     WHERE s.domain='${DOMAIN}' AND p.name='${page}' AND cc.function='evidence-chart';")
  note "      (this page carries the section: ${carries} placement(s))"

  # Every point in the register must be on a page that carries the section,
  # with its label and its figure exactly as the register holds them.
  while IFS=$'\t' read -r chart cpage label value unit; do
    [ -n "$chart" ] || continue
    [ "$carries" != "0" ] || continue
    grep -q "data-chart=\"${chart}\"" <<< "$html"
    check $? "chart '${chart}' rendered"
    grep -qF "$label" <<< "$html"
    check $? "  label present: ${label:0:60}"
    # the figure, inside the chart's value span, exactly as the register holds it
    grep -qE "evidence-chart__value\">${value}${unit}<" <<< "$html"
    check $? "  figure ${value}${unit} rendered from fact value"
    grep -qE -- "--v:${value}\.0000" <<< "$html"
    check $? "  bar geometry drawn from the same figure"
  done <<< "$EXPECTED"

  # A page that does NOT carry the section must show no chart at all.
  if [ "$carries" = "0" ]; then
    ! grep -q 'data-chart="' <<< "$html"
    check $? "no chart on a page that does not carry the section"
  fi

  # The section's figures must match the register as it stands NOW, not as it
  # stood when the charts were seeded. refresh_evidence_base re-runs each fact's
  # source.sql and rewrites value/verified_at in place — measured 2026-07-27, the
  # council facts moved 108/37/9 -> 110/38/10 and the live page followed. That is
  # the design working, and it is also why a SQL-sourced fact must carry no
  # hand-written `display`: it would have stayed at the old number.

  # Links: capture EVERY href, then strip the fragment — the anchored-href blind
  # spot that let 21 broken links pass three agreeing checks (L2).
  mapfile -t hrefs < <(grep -oE 'href="(/[^"]*)"' <<< "$html" | sed -E 's/^href="//; s/"$//' | cut -d'#' -f1 | grep -v '^$' | sort -u)
  bad=0
  for h in "${hrefs[@]}"; do
    code=""
    for i in 1 2 3; do
      code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 20 "https://${DOMAIN}${h}")
      [ "$code" = "200" ] && break
      sleep 2
    done
    [ "$code" = "200" ] || { note "      BROKEN ${h} -> ${code}"; bad=$((bad+1)); }
  done
  [ "$bad" -eq 0 ]
  check $? "all ${#hrefs[@]} internal link targets on this page resolve as served"
done

note ""
if [ "$fail" -eq 0 ]; then note "ALL CHECKS PASSED"; else note "${fail} CHECK(S) FAILED"; fi
exit $(( fail > 0 ))
