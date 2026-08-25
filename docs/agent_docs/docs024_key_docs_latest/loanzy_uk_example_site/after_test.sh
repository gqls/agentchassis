#!/bin/bash
# after-test for the garden-tools.uk one-shot build.  REWRITTEN 2026-08-23 20:58Z.
#
# ⚠ WHY IT WAS REWRITTEN — the v1 harness printed "(no lines = nothing found)" for THREE checks
# whose page-list query had ERRORED (pages.is_archived does not exist). A failed query produced a
# reassuring message. That is the exact could-not-have-come-out-otherwise failure this lane keeps
# logging, committed by the instrument built to detect it.
#
# Rules this version enforces:
#   1. Every query's row count is captured. ZERO rows is reported as UNKNOWN/EMPTY, never as PASS.
#   2. Page URLs come from pages.url — NOT constructed from pages.name. v1 guessed and guessed wrong
#      (real urls include /brand-directory/index.html and /tools/finder/index.html).
#   3. Every grep prints WHAT IT MATCHED, never just a count. A count cannot be sanity-checked.
#   4. No `grep -i` on a pattern whose discriminating power is capitalisation.
#
# ── PROMOTED into this lane directory 2026-08-25 from the 08-24 session's scratchpad, per that
#    session's own instruction in HANDOFF_2026-08-24 §3. Unchanged except this banner.
#
# ⚠ SECTION (a) 311 — RE-PIN AND CHECK. The eight md5s below were taken 2026-08-23.
#
#    ~~All eight incumbents moved on 2026-08-20 under bugs_open/283, and RFC_032 is rewriting
#    html_template fleet-wide, so they will keep moving. On any run that is NOT immediately
#    preceded by re-pinning them yourself, "*** HTML CHANGED ***" means "the pin is old", NOT
#    "this build collided". Re-pin before a build; ignore section (a) otherwise.~~
#
#    ⚠ CORRECTED 2026-08-25 10:29Z — I WROTE THAT BANNER THIS MORNING AND IT WAS WRONG IN THE
#    DANGEROUS DIRECTION. Caught by the bugs_open/381 lane, who re-pinned at 10:27:16Z; verified
#    here independently. [MEASURED 2026-08-25 10:29Z] ALL EIGHT MATCH THE 08-23 PINS EXACTLY,
#    8 of 8, html_md5 AND schema_md5 — and content_components.updated_at on all eight is
#    2026-08-20, i.e. UNTOUCHED FOR FIVE DAYS. RFC_032 has not reached them.
#
#    So the 08-20 move is TRUE AS HISTORY and FALSE AS CURRENT STATE, and my inference from it
#    inverted the instrument: I told the reader that "*** HTML CHANGED ***" means the pin is
#    stale. With the pins current, a CHANGED line is a REAL COLLISION — the exact signal
#    bugs_open/311 exists to raise — and my banner would have had someone dismiss a true
#    positive as a documentation artefact. A caveat that reads as caution can disarm a check.
#
#    THE OPERATIVE INSTRUCTION: re-pin before a build if you can (it is one query), but if you
#    have not, DO NOT DISMISS A CHANGED LINE — check whether that component moved
#    (SELECT updated_at ... ) before deciding which of the two it is. The pins have been stable
#    since 08-23; treat drift as news, not as noise.
#
set -o pipefail
# Usage: ./after_test.sh [domain]      (default: garden-tools.uk, the 08-24 worked example)
#   PARAMETERISED 2026-08-25 so it can be pointed at a NEW build without editing it. Everything
#   below already read $DOM; only this line was hardcoded.
#   ⚠ Section (a)'s md5 pins are FLEET-WIDE incumbents, not per-site — they stay meaningful on any
#     domain, and stay a dated PIN rather than a baseline (see the banner above).
#   ⚠ The PROMISE-vs-DELIVERY month/step/compare vocabulary was written for a gardening subject.
#     On a different vertical, read the raw tables=/li=/strong= columns and add promise words for
#     that subject rather than trusting a silent run.
DOM=${1:-garden-tools.uk}
echo "### after_test.sh target domain: $DOM   (run at $(date -u '+%Y-%m-%d %H:%M:%SZ'))"
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
q() {
  local out rc
  out=$($PSQL -tA -c "$1" 2>&1); rc=$?
  if [ $rc -ne 0 ] || printf '%s' "$out" | grep -q '^ERROR:'; then
    echo "__QUERY_FAILED__" >&2
    printf '%s\n' "$out" | head -3 >&2
    return 1
  fi
  printf '%s\n' "$out"
}

hr() { echo; echo "=============== $* ==============="; }

hr "(a) 311 NO COLLATERAL — all EIGHT incumbents, vs THIS session's pre-run pins"
$PSQL -c "
WITH baseline(id8, html_md5, schema_md5) AS (VALUES
  ('7d8b0503','1de725368744680ef052ab1da2b4dc94','8e2cfe0afb1863b178390d6a048409b0'),
  ('9cbfe279','a591c07c6da83d77aea7bc7d29819257','3bba8e7d9d13338ea0370971f9ef487c'),
  ('824e3309','67e3d20d83ddad4b0cff54b2e4a98559','dd8f9863c84f8a5a7ec3e99154241f43'),
  ('2cf33f06','07aa4a2ba7a7778b736e8fadb6cff8b3','a805b2af699f1c28a9d7833ff35405e6'),
  ('b7a499f4','12bf5cc88fbd8138769f78502702ab7a','fd2a6336dd159833892afdad62863f19'),
  ('70b72b3e','c42b9a8c843638d660509ca883eb7e9f','b7a1e6090d00f0bc1f17178d9ade3a45'),
  ('b420389f','a9dea7cd35372bd6c0bd70cee8140d06','a5790bcfeb1d46da94cb8ef3d9fc5fdc'),
  ('b89f91e1','a453a6565489c348ad6a9156a8af812f','8265ae5a931b735305b1fe007b148acb'))
SELECT b.id8, cc.function,
       CASE WHEN md5(cc.html_template)=b.html_md5 THEN 'UNCHANGED' ELSE '*** HTML CHANGED ***' END AS html,
       CASE WHEN md5(cc.input_schema::text)=b.schema_md5 THEN 'UNCHANGED' ELSE '*** SCHEMA CHANGED ***' END AS schema
FROM baseline b JOIN content_components cc ON left(cc.id::text,8)=b.id8 ORDER BY cc.function;"
n=$(q "SELECT count(*) FROM content_components WHERE left(id::text,8) IN ('7d8b0503','9cbfe279','824e3309','2cf33f06','b7a499f4','70b72b3e','b420389f','b89f91e1');")
echo ">> incumbents matched: $n (MUST be 8 — anything less means the check silently skipped rows)"

hr "(a) 311 DIVERSION — did the guard actually get EXERCISED?"
scoped=$(q "SELECT count(*) FROM content_components WHERE function LIKE '%garden-tools-uk%';")
div=$(q "SELECT count(*) FROM agent_error_log WHERE error_code='COMPONENT_COLLISION_DIVERTED' AND occurred_at > '2026-08-23 17:17:00';")
echo "scoped components for this site : $scoped"
echo "COMPONENT_COLLISION_DIVERTED rows: $div"
if [ "$scoped" = "0" ] && [ "$div" = "0" ]; then
  echo ">> NOT EXERCISED. Report as 'the guard never ran', NOT as 'no collision occurred'."
  echo "   Cause here: the only tool page (tool-finder) was owner-gated to human review, so the"
  echo "   component-creation path was never reached. A clean result proves nothing about the guard."
fi

hr "(b) 260 template-syntax leak (ceiling, not a count — validate_content caps at 10/detector)"
for t in '{{end}}' '{{if' '{{.label}}' '{{range'; do
  printf '%-12s stored_html_hits: %s\n' "$t" "$(q "SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id WHERE s.domain='$DOM' AND pc.rendered_html LIKE '%$t%';")"
done

hr "PAGES (schema-correct: status, url)"
$PSQL -c "SELECT p.name, p.page_type, p.build_status, p.status, p.url FROM pages p JOIN sites s ON s.id=p.site_id
          WHERE s.domain='$DOM' ORDER BY p.build_status DESC, p.name;"

hr "THE ARTEFACT — every page fetched at its REAL url"
URLS=$(q "SELECT p.url FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='$DOM' AND p.status='active' ORDER BY p.name;")
URLQ=$?
cnt=$(printf '%s\n' "$URLS" | grep -c . )
if [ "$URLQ" -ne 0 ]; then echo "!! URL QUERY FAILED — everything below is UNKNOWN, not a pass."; cnt=0; fi
echo ">> page rows returned: $cnt"
if [ "$cnt" = "0" ]; then
  echo "!! QUERY RETURNED NOTHING — this is UNKNOWN, not a pass. Fix the query before reading on."
else
  CB=$(date +%s)
  printf '%s\n' "$URLS" | while read -r u; do
    [ -z "$u" ] && continue
    case "$u" in /*) full="https://$DOM$u" ;; http*) full="$u" ;; *) full="https://$DOM/$u" ;; esac
    body=$(curl -s -m 25 "$full?cb=$CB"); code=$(curl -s -m 25 -o /dev/null -w '%{http_code}' "$full?cb=$CB")
    printf '%-34s http=%-4s bytes=%-6s inputs=%-3s buttons=%-3s cta-anchors=%-3s leak=%s\n' \
      "$u" "$code" "$(printf '%s' "$body" | wc -c)" \
      "$(printf '%s' "$body" | grep -o '<input' | wc -l)" \
      "$(printf '%s' "$body" | grep -o '<button' | wc -l)" \
      "$(printf '%s' "$body" | grep -oiE '<a[^>]*class="[^"]*(cta|btn)[^"]*"' | wc -l)" \
      "$(printf '%s' "$body" | grep -oE '\{\{(end|if|range|\.[a-z])' | wc -l)"
    # claims + invented identity, printing MATCHES not counts
    em=$(printf '%s' "$body" | grep -oE '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}|/cdn-cgi/l/email-protection' | sort -u | tr '\n' ' ')
    ph=$(printf '%s' "$body" | grep -oE '(\+44[0-9 ]{9,13}|\b0[12378][0-9]{8,9}\b)' | sort -u | tr '\n' ' ')
    by=$(printf '%s' "$body" | grep -oE '(written|reviewed|tested|authored) by [A-Z][a-z]+ [A-Z][a-z]+' | sort -u | tr '\n' ' ')
    pa=$(printf '%s' "$body" | grep -oiE 'amazon|rhs shop|thompson ?& ?morgan|crocus|dobies' | sort -u | tr '\n' ' ')
    rh=$(printf '%s' "$body" | grep -o 'RHS' | wc -l)
    [ -n "$em" ] && echo "      ⚠ EMAIL: $em"
    [ -n "$ph" ] && echo "      ⚠ PHONE: $ph"
    [ -n "$by" ] && echo "      ⚠ BYLINE: $by"
    [ -n "$pa" ] && echo "        partner-names: $pa  (boilerplate vs claimed relationship — READ it)"
    [ "$rh" != "0" ] && echo "        RHS mentions: $rh (read in context; 'RHS testing standards' is not an endorsement claim)"
  done
fi

hr "(328) DEAD LINKS — every internal link on every served page"
CB=$(date +%s)
printf '%s\n' "$URLS" | while read -r u; do
  [ -z "$u" ] && continue
  case "$u" in /*) full="https://$DOM$u" ;; *) full="https://$DOM/$u" ;; esac
  curl -s -m 25 "$full?cb=$CB" | grep -oE 'href="/[^"#?]*"' | sed 's/href="//;s/"//' | sort -u | while read -r h; do
    case "$h" in *.css|*.js|*.png|*.jpg|*.svg|*.webp|*.ico) continue ;; esac
    c=$(curl -s -m 20 -o /dev/null -w '%{http_code}' "https://$DOM$h?cb=$CB")
    [ "$c" != "200" ] && echo "  DEAD $c  $h   (linked from $u)"
  done
done
echo ">> any DEAD line above is bugs_open/328 live. Transient during a build; permanent once the"
echo "   target page sits at needs_human_review with nothing scheduled to build it."

hr "PROMISE vs DELIVERY — does each page contain what its own HEADINGS say it does?"
# Added 2026-08-24 after the owner found a page promising "month by month" with no calendar.
# The v2 harness had NO check of this kind: http/bytes/inputs/buttons/leaks/dead-links all passed.
# A page's own <h1>/<h2> is a promise. Check the promise against the markup.
CB=$(date +%s)
printf '%s\n' "$URLS" | while read -r u; do
  [ -z "$u" ] && continue
  case "$u" in /*) full="https://$DOM$u" ;; *) full="https://$DOM/$u" ;; esac
  body=$(curl -s -m 25 "$full?cb=$CB")
  [ -z "$body" ] && continue
  # strip css/js so counts are markup, not stylesheet rules
  m=$(printf '%s' "$body" | perl -0pe 's/<style.*?<\/style>//gs; s/<script.*?<\/script>//gs' 2>/dev/null || printf '%s' "$body")
  heads=$(printf '%s' "$m" | grep -oE '<h[123][^>]*>[^<]{3,90}' | sed 's/<[^>]*>//g')
  tbl=$(printf '%s' "$m" | grep -c '<table')
  li=$(printf '%s' "$m" | grep -o '<li' | wc -l)
  strong=$(printf '%s' "$m" | grep -o '<strong' | wc -l)
  months=$(printf '%s' "$m" | grep -oE '\b(January|February|March|April|May|June|July|August|September|October|November|December)\b' | sort -u | wc -l)
  printf '%-30s tables=%-2s li=%-3s strong=%-3s distinct_months=%-2s\n' "$u" "$tbl" "$li" "$strong" "$months"
  # promise words in headings that imply STRUCTURE the page must then contain
  printf '%s\n' "$heads" | while read -r h; do
    case "$h" in
      *"month by month"*|*"month-by-month"*|*calendar*|*Calendar*)
        [ "$months" -lt 6 ] && echo "      ⚠ PROMISE UNMET: heading says '$(echo "$h"|sed 's/^ *//')' but only $months distinct month names on the page" ;;
      *"step by step"*|*"checklist"*|*"Checklist"*)
        [ "$li" -lt 3 ] && echo "      ⚠ PROMISE UNMET: heading says '$(echo "$h"|sed 's/^ *//')' but only $li list items" ;;
      *compare*|*Compare*|*comparison*|*Comparison*|*"side by side"*)
        [ "$tbl" = "0" ] && echo "      ⚠ PROMISE UNMET: heading says '$(echo "$h"|sed 's/^ *//')' but the page has NO table" ;;
    esac
  done
done
echo ">> A page can be 200, 67KB, leak-free and fully linked and still not contain what it promises."
echo "   Byte count is not a completeness check. Neither is 'it deployed'."
