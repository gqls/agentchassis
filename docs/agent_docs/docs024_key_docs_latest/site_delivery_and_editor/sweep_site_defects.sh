#!/usr/bin/env bash
# sweep_site_defects.sh — run the mechanisable checks of
# docs024_key_docs_latest/SITE_DEFECT_CATEGORIES.md against ONE site.
#
#   ./sweep_site_defects.sh <domain>          e.g. boxingonline.com
#   exit 0 = no findings · 1 = findings · 2 = usage/lookup failure
#   ⚠ a BLIND check always prints and always makes the exit non-zero
#
# WHY THIS EXISTS. The checklist is prose with ~30 checks; before this, each
# pre-delivery sweep re-typed them and re-remembered §0's disciplines.
#
# §0 DISCIPLINES, EXECUTED (each is a real defence, not decoration):
#  - serving host from sites.publish_project/publish_target, never the customer
#    domain (a parked domain 200s every path);
#  - an invented-path 404 control per run, a `</html>` control per page — a
#    control that fails marks the page BLIND, never clean;
#  - pages enumerated from `pages WHERE deployed_at IS NOT NULL`;
#  - served last-modified printed beside pages.deployed_at (the 420
#    discriminator: served older = mirror lag, wait; newer = look upstream);
#  - Postgres regex uses \y, never \b (BACKSPACE, returns a clean 0);
#  - every inner `kubectl exec -i` gets </dev/null so it cannot eat the loop's
#    stdin and report one row, cleanly.
#
# ── SIX DEFECTS THIS SCRIPT SHIPPED ON 2026-09-03 AND WHY EACH ARM IS AS IT IS.
# Its first run produced four false positives and one false clean. Do not
# "simplify" these back:
#  (a) a fetched logo.png in the page corpus made grep say "binary file matches"
#      and STOP counting — silently blinding every text check. The corpus is
#      HTML only, and fetched assets are written OUTSIDE it.
#  (b) the order-email needle came back EMPTY and `grep -Fc ""` reported a tidy
#      0 — a clean report from a check measuring nothing, on the owner's item 0.
#      A needle is now proven non-empty or the check declares itself BLIND.
#  (c) `placehold` as a substring matched five legitimate `placeholder="e.g. …"`
#      attributes. Tokens are case-sensitive and host-anchored.
#  (d) `grep -c` counts LINES; on single-line HTML every per-page count was
#      1-per-file. Occurrences use `grep -o | wc -l`.
#  (e) `grep -q` inside a `set -o pipefail` script makes a MATCHING pipeline
#      report failure (grep exits early, the producer takes SIGPIPE, pipefail
#      promotes it) — an if-guard inverts and warns on everything. Count instead.
#      LANDMINES: "set -o pipefail + grep -q makes a MATCHING pipeline report
#      FAILURE".
#  (f) §3.1 INNER-JOINED tool pages to their tool components, so a tool page with
#      NO tool component — the worst case — was absent from the report rather
#      than flagged. Every check that enumerates a population now LEFT JOINs it,
#      and prints `none` rather than nothing when it finds nothing.
#
# NOT MECHANISABLE — printed as OWED, by the checklist's own admission:
#   1.2 title promises a dated checkable thing (refused as a regex on evidence)
#   1.3 the site holds the material and the page does not use it
#   3.2 empty-set page with no empty state · 7.1/7.2/7.3 plan & brief fidelity
#   9.1-9.4 cross-site sameness (needs a cohort) · 10.1/10.5 vertical temperature
#   4.3 invented brand lettering (LOOK at the image)
set -uo pipefail
DOMAIN="${1:?Usage: $0 <domain>}"
PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)
q() { "${PSQL[@]}" "$1" </dev/null; }
FINDINGS=0; BLIND=0
finding() { FINDINGS=$((FINDINGS+1)); echo "      ⚠ $*"; }
blind()   { BLIND=$((BLIND+1));       echo "      ⚠ BLIND — $*"; }
# print a query's rows, or an explicit `none` — a check that prints nothing when
# clean is indistinguishable from a check that did not run.
rows() { local out; out=$(q "$1"); if [ -n "$out" ]; then printf '%s\n' "$out"; else echo "      none"; fi; }
count_of() { local n; n=$(body_all | grep -oiE "$1" | wc -l); echo "$n"; }

SITE_ID=$(q "SELECT id FROM sites WHERE domain='${DOMAIN}';" | tr -d ' ')
[ -n "$SITE_ID" ] || { echo "no site row for ${DOMAIN}" >&2; exit 2; }
HOST=$(q "SELECT coalesce(nullif(publish_project,''), domain) FROM sites WHERE id='${SITE_ID}';" | tr -d ' ')
BASE="https://${HOST}"
echo "=== sweep ${DOMAIN} — site ${SITE_ID} — serving host ${HOST} — $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

CTL=$(curl -s -o /dev/null -m 20 -w '%{http_code}' "${BASE}/zzz-not-real-${RANDOM}.html")
if [ "$CTL" = 404 ]; then echo "[control] invented path -> 404 OK"
else echo "[control] invented path -> ${CTL}"; blind "the host answers ${CTL} for an invented path (parked/catch-all) — EVERY served result below is meaningless"; fi

TMP=$(mktemp -d); ASSETS=$(mktemp -d); trap 'rm -rf "$TMP" "$ASSETS"' EXIT
# (a) HTML-ONLY corpus. A binary file here makes grep report "binary file
# matches" and stop counting; fetched assets go to $ASSETS, never $TMP.
body_all() { for f in "$TMP"/*; do [ -f "$f" ] && cat "$f"; done 2>/dev/null; }

mapfile -t URLS < <(q "SELECT url FROM pages WHERE site_id='${SITE_ID}' AND deployed_at IS NOT NULL AND status='active' ORDER BY url;")
[ "${#URLS[@]}" -gt 0 ] || { blind "no deployed+active pages in the DB for this site — nothing to sweep"; exit 2; }
echo "[pages] ${#URLS[@]} deployed+active pages enumerated from the DB"
for u in "${URLS[@]}"; do
  f="$TMP/$(echo "$u" | tr '/' '_')"
  code=$(curl -s -m 25 -o "$f" -w '%{http_code}' "${BASE}${u}?cb=$RANDOM$RANDOM")
  lm=$(curl -s -m 25 -o /dev/null -w '%header{last-modified}' "${BASE}${u}?cb=$RANDOM$RANDOM")
  dep=$(q "SELECT to_char(deployed_at,'YYYY-MM-DD HH24:MI:SS') FROM pages WHERE site_id='${SITE_ID}' AND url='${u}';")
  present=$(grep -o '</html>' "$f" 2>/dev/null | wc -l)          # (d) occurrences
  flag=' '
  if [ "$code" != 200 ]; then flag='⚠'; FINDINGS=$((FINDINGS+1));
  elif [ "$present" -eq 0 ]; then flag='⚠'; BLIND=$((BLIND+1)); fi
  printf '%s %-46s %s  served=%s  deployed_at=%s  control(</html>)=%s\n' "$flag" "$u" "$code" "${lm:-none}" "$dep" "$present"
done

echo; echo "--- 1.1 meta-copy (first-person-plural policy prose), per page + context"
mc_total=0
for f in "$TMP"/*; do [ -f "$f" ] || continue
  n=$(grep -oiE "we write|we'd rather|we cover|gets checked|we update" "$f" | wc -l)
  [ "$n" -gt 0 ] && { mc_total=$((mc_total+n)); printf '      %-46s %s\n' "$(basename "$f" | tr '_' '/')" "$n"; }
done
if [ "$mc_total" -gt 0 ]; then FINDINGS=$((FINDINGS+1))
  body_all | grep -ohiE ".{45}(we write|we'd rather|we cover|gets checked|we update).{45}" 2>/dev/null | sort -u | sed 's/^/        /' | head -6
else echo "      none"; fi

echo "--- 1.4 raw source residue"
md=$(count_of '\]\(http'); el=$(count_of '\.\.\.')
echo "      markdown links '](http': ${md}   ellipses: ${el}"
[ "$md" -gt 0 ] && finding "literal markdown on served pages (1.4 / bugs_open/332)"

echo "--- 1.5 AI tells — CONTEXT printed, because 'honest'/'plainly' are weak needles and need a reader (checklist 1.2's rule)"
tl=$(count_of 'plainly|honest|starting point, not the final word|before your [a-z]+ have to')
if [ "$tl" -gt 0 ]; then echo "      ${tl} hit(s), judge each:"; body_all | grep -ohiE '.{40}(plainly|honest|starting point, not the final word|before your [a-z]+ have to).{40}' 2>/dev/null | sort -u | sed 's/^/        /' | head -8
else echo "      none"; fi
echo "    control-labelled fields over ~120 chars (DB):"
rows "SELECT '      ⚠ '||coalesce(cc.name,'(no component)')||'.'||k||' = '||length(v)||' chars on '||p.url FROM page_components pc JOIN pages p ON p.id=pc.page_id LEFT JOIN content_components cc ON cc.id=pc.component_id, jsonb_each_text(pc.content_data) AS e(k,v) WHERE p.site_id='${SITE_ID}' AND (k ~ '(button|label|_cta\$|^cta)') AND length(v) > 120 ORDER BY length(v) DESC;"

echo "--- 2.2 index-role pages and their listable-item counts (zero is a finding on its own)"
rows "SELECT '      '||CASE WHEN coalesce(max(coalesce(jsonb_array_length(pc.content_data->'articles'), jsonb_array_length(pc.content_data->'items'))),0)=0 THEN '⚠ ' ELSE '' END||p.url||' -> '||coalesce(max(coalesce(jsonb_array_length(pc.content_data->'articles'), jsonb_array_length(pc.content_data->'items'))),0)||' items (max over '||count(pc.id)||' listing component(s))' FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id AND (pc.content_data ? 'articles' OR pc.content_data ? 'items') WHERE p.site_id='${SITE_ID}' AND p.deployed_at IS NOT NULL AND p.page_type ~ '(index|directory)' GROUP BY p.url ORDER BY p.url;"
z=$(q "SELECT count(*) FROM (SELECT p.url, coalesce(max(coalesce(jsonb_array_length(pc.content_data->'articles'), jsonb_array_length(pc.content_data->'items'))),0) n FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id AND (pc.content_data ? 'articles' OR pc.content_data ? 'items') WHERE p.site_id='${SITE_ID}' AND p.deployed_at IS NOT NULL AND p.page_type ~ '(index|directory)' GROUP BY p.url) t WHERE n=0;")
[ "${z:-0}" -gt 0 ] && finding "${z} index-role page(s) list ZERO items (2.2 / bugs_open/444)"

echo "--- 2.3 nav: distinct hrefs vs distinct labels in <header>"
if [ -f "$TMP/_index.html" ]; then h=$(sed -n '/<header/,/<\/header>/p' "$TMP/_index.html")
  nh=$(printf '%s' "$h" | grep -oE 'href="[^"]+"' | sort -u | wc -l); nl=$(printf '%s' "$h" | grep -oE '>[A-Za-z][^<]{1,30}</a>' | sort -u | wc -l)
  echo "      hrefs ${nh}  labels ${nl}"; [ "$nh" != "$nl" ] && finding "header href/label counts differ (2.3 — one page under two labels, or two pages under one)"
else blind "no /index.html in the corpus — 2.3 not run"; fi

echo "--- 2.4 orphans (no inbound link from any other deployed page; absolute AND relative forms)"
orph=0
for u in "${URLS[@]}"; do
  [ "$u" = "/index.html" ] && continue
  self="$TMP/$(echo "$u" | tr '/' '_')"; b=$(basename "$u")
  # (relative-href blindness) match href="/full/path", href="path.html" and href="./path.html"
  inb=$(grep -lE "href=\"(${u}|\./)?[^\"]*${b}\"" "$TMP"/* 2>/dev/null | grep -v "^${self}$" | wc -l)
  [ "$inb" = 0 ] && { orph=$((orph+1)); finding "ORPHAN (0 inbound): ${u}"; }
done
[ "$orph" = 0 ] && echo "      none"

echo "--- 2.6 nav reach per page-type family (header AND anywhere — footer/body links are reach)"
rows "SELECT '      '||page_type||': '||count(*)||' deployed' FROM pages WHERE site_id='${SITE_ID}' AND deployed_at IS NOT NULL GROUP BY page_type ORDER BY page_type;"
for fam in $(q "SELECT DISTINCT page_type FROM pages WHERE site_id='${SITE_ID}' AND deployed_at IS NOT NULL AND page_type NOT IN ('landing');"); do
  pat=$(q "SELECT string_agg(DISTINCT split_part(ltrim(url,'/'),'/',1), '|') FROM pages WHERE site_id='${SITE_ID}' AND page_type='${fam}' AND deployed_at IS NOT NULL;")
  [ -n "${pat:-}" ] || continue
  hdr=$(sed -n '/<header/,/<\/header>/p' "$TMP/_index.html" 2>/dev/null | grep -oE "href=\"/(${pat})" | wc -l)
  any=$(grep -lE "href=\"/(${pat})" "$TMP"/* 2>/dev/null | wc -l)
  if [ "$any" = 0 ]; then finding "${fam}: NO page links it (2.6 — a live family with no nav path)"
  else printf '      %-14s header %s, linked from %s/%s pages%s\n' "$fam" "$hdr" "$any" "${#URLS[@]}" "$([ "$hdr" = 0 ] && echo ' (reachable, but not from the header)')"; fi
done

echo "--- 3.1 tools: STORED component markup (never the served page — chrome adds a menu button)"
# (f) LEFT JOIN: a tool page with NO tool-level component is the WORST case and
# must appear, flagged — an inner join deleted it from the report.
rows "SELECT '      '||CASE WHEN t.n_tool=0 THEN '⚠ ' ELSE '' END||p.url||': tool components '||t.n_tool||', inputs '||t.inp||' selects '||t.sel||' textareas '||t.ta FROM pages p JOIN LATERAL (SELECT count(*) FILTER (WHERE cc.component_level='tool') AS n_tool, coalesce(sum((SELECT count(*) FROM regexp_matches(pc.rendered_html,'<input','g'))),0) AS inp, coalesce(sum((SELECT count(*) FROM regexp_matches(pc.rendered_html,'<select','g'))),0) AS sel, coalesce(sum((SELECT count(*) FROM regexp_matches(pc.rendered_html,'<textarea','g'))),0) AS ta FROM page_components pc LEFT JOIN content_components cc ON cc.id=pc.component_id WHERE pc.page_id=p.id) t ON true WHERE p.site_id='${SITE_ID}' AND p.page_type='tool' AND p.deployed_at IS NOT NULL ORDER BY p.url;"
nt=$(q "SELECT count(*) FROM pages p WHERE p.site_id='${SITE_ID}' AND p.page_type='tool' AND p.deployed_at IS NOT NULL AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE pc.page_id=p.id AND cc.component_level='tool');")
[ "${nt:-0}" -gt 0 ] && finding "${nt} tool page(s) carry NO tool component at all (3.1 — a tool URL that is not a tool)"
echo "    stored-newer-than-deployed (repair written, not served — ⚠ triage only, ~7/11 precision; confirm at the served bytes):"
rows "SELECT '      ⚠ '||p.url||' components '||to_char(max(pc.updated_at),'HH24:MI')||' > deployed '||to_char(p.deployed_at,'HH24:MI') FROM pages p JOIN page_components pc ON pc.page_id=p.id WHERE p.site_id='${SITE_ID}' AND p.deployed_at IS NOT NULL AND p.build_status='deployed' GROUP BY p.url, p.deployed_at HAVING max(pc.updated_at) > p.deployed_at + interval '2 minutes';"

echo "--- 3.3 evidence_base facts"
nf=$(q "SELECT coalesce(jsonb_array_length(data->'facts'),0) FROM site_specs WHERE site_id='${SITE_ID}' AND aspect='evidence_base' AND is_current;")
if [ -z "${nf:-}" ]; then finding "NO evidence_base row at all (3.3 / bugs_open/427)"; else echo "      ${nf} facts"; [ "$nf" -le 5 ] && finding "evidence_base holds ${nf} facts (<=5 — the fleet's own 'empty corpus' line, 3.3 / bugs_open/427)"; fi

echo "--- 4.1/4.2 imagery: BOTH encodings per page (an <img> count alone is half-blind — heroes are CSS backgrounds)"
logo_only=0
for u in "${URLS[@]}"; do f="$TMP/$(echo "$u" | tr '/' '_')"; [ -f "$f" ] || continue
  im=$(grep -o '<img[^>]*>' "$f" | wc -l); cu=$(grep -oE 'background-image:[^;}]*url\([^)]*\)' "$f" | wc -l)
  [ "$im" -le 1 ] && [ "$cu" -eq 0 ] && logo_only=$((logo_only+1))
  printf '      %-46s <img> %-3s css-url %s\n' "$u" "$im" "$cu"; done
nh=$(body_all | grep -ohE "url\('?[^')]+'?\)" | sort -u | wc -l)
echo "      distinct hero/css-url files: ${nh}"
[ "$logo_only" -gt 0 ] && finding "${logo_only} page(s) carry at most the logo and no CSS hero (4.2 — 'everything except the logo')"
[ "$nh" = 1 ] && finding "ONE distinct hero file across every page that has one (4.2 companion / 9.4 within-site sameness)"

echo "--- 4.5 logo alpha (colour type 4/6 OR tRNS — either alone gives a false negative)"
LOGO=$(q "SELECT url FROM assets WHERE site_id='${SITE_ID}' AND purpose='logo' ORDER BY updated_at DESC LIMIT 1;")
LOGO="${LOGO:-/assets/images/logo.png}"; echo "      probing ${LOGO} (from assets.url)"
if curl -s -m 25 -f -o "$ASSETS/logo" "${BASE}${LOGO}?cb=$RANDOM"; then
  python3 - "$ASSETS/logo" <<'PY' || true
import struct,sys
d=open(sys.argv[1],'rb').read()
if d[:8]!=b'\x89PNG\r\n\x1a\n':
    print(f"      ⚠ NOT a PNG (magic {d[:4].hex()}) — 4.4 shape claim and 4.5 both unresolved"); raise SystemExit(3)
w,h,bd,ct=struct.unpack('>IIBB',d[16:26]); trns=b'tRNS' in d
ok = ct in (4,6) or trns
print(f"      {w}x{h} depth={bd} colour_type={ct} tRNS={trns} -> {'OK (real alpha)' if ok else '⚠ 4.5 baked background'}")
raise SystemExit(0 if ok else 3)
PY
  [ "$?" = 3 ] && finding "logo has no alpha channel (4.5 / bugs_closed/424)"
else blind "logo fetch failed at ${LOGO} — 4.5 not run"; fi

echo "--- 5.1 empty slot elements (a component that renders an empty element cannot tell missing from blank)"
es=0
for f in "$TMP"/*; do [ -f "$f" ] || continue
  n=$(grep -oE 'class="[^"]*__(excerpt|category|date|meta|deck|summary)"></(p|span|div)>' "$f" | wc -l)
  [ "$n" -gt 0 ] && { es=$((es+n)); finding "$(basename "$f" | tr '_' '/'): ${n} empty slot element(s) (5.1 / bugs_open/425)"; }
done
[ "$es" = 0 ] && echo "      none"

echo "--- 5.3 unsubstituted placeholders"
hard=$(body_all | grep -oE 'placehold\.co|\bEXAMPLE\b|\bTODO\b|\bLorem ipsum\b|_ADDRESS\b' | wc -l)
echo "      placehold.co / hard tokens: ${hard}"
[ "$hard" -gt 0 ] && finding "${hard} hard placeholder token(s) served (5.3)"
for t in $(body_all | grep -oE '\b[A-Z]{3,}_[A-Z_]{2,}\b' | sort -u); do
  # (e) NOT grep -q under pipefail — it would invert and warn on every token.
  ndecl=$(body_all | grep -cE "(var|let|const|window\.)[[:space:]]*${t}[[:space:]]*=" || true)
  if [ "${ndecl:-0}" -gt 0 ]; then echo "      ${t}: DECLARED in-document (${ndecl} line(s)) — a working constant, not a placeholder"
  else finding "${t}: no declaration in the served bytes — candidate unsubstituted token"; fi
done

echo "--- 6.1 ordering party's email on the site (must be 0)"
OE=$(q "SELECT coalesce(direction->>'customer_email','') FROM build_queue WHERE lower(domain)=lower('${DOMAIN}') ORDER BY created_at DESC LIMIT 1;" | tr -d ' ')
[ -n "$OE" ] || OE=$(q "SELECT coalesce(direction->>'email','') FROM build_queue WHERE lower(domain)=lower('${DOMAIN}') ORDER BY created_at DESC LIMIT 1;" | tr -d ' ')
if [ -n "$OE" ]; then
  hits=$(body_all | grep -Fci "$OE"); ctl=$(body_all | grep -Fci "$(printf '%s' "$DOMAIN" | cut -d. -f1)")
  echo "      order-email hits: ${hits}   [needle ${#OE} chars; positive control (domain stem) -> ${ctl}]"
  [ "$hits" -gt 0 ] && finding "the ordering party's email is on the served site (6.1 / bugs_open/420)"
  [ "$ctl" = 0 ] && blind "the positive control found nothing — the corpus may be empty, so the 0 above means nothing"
else blind "no ordering email in build_queue.direction for ${DOMAIN} — 6.1 measured NOTHING (an empty needle makes grep -Fc return a tidy 0)"; fi
echo "      sites.email: $(q "SELECT coalesce(nullif(email,''),'(empty)') FROM sites WHERE id='${SITE_ID}';")"
mt=$(count_of 'mailto:|tel:'); echo "      mailto:/tel: on served pages: ${mt}"

echo "--- 6.2 forms that submit nowhere"
rows "SELECT '      '||CASE WHEN coalesce(fa,'')IN('','#contact','#') THEN '⚠ ' ELSE '' END||'form_action='||coalesce(fa,'(null)')||' x'||n FROM (SELECT pc.content_data->>'form_action' AS fa, count(*) AS n FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE p.site_id='${SITE_ID}' AND pc.content_data ? 'form_action' GROUP BY 1) t ORDER BY n DESC;"
nfa=$(q "SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE p.site_id='${SITE_ID}' AND pc.content_data ? 'form_action' AND coalesce(pc.content_data->>'form_action','') IN ('','#contact','#');")
[ "${nfa:-0}" -gt 0 ] && finding "${nfa} form(s) submit nowhere (6.2)"

echo "--- 6.3 retracted pages: deletion closes on 404 AND zero inbound, never one of the two"
arch=$(q "SELECT url FROM pages WHERE site_id='${SITE_ID}' AND status='archived';")
if [ -z "$arch" ]; then echo "      none archived"
else while read -r au; do [ -n "$au" ] || continue
    c=$(curl -s -o /dev/null -m 20 -w '%{http_code}' "${BASE}${au}?cb=$RANDOM"); b=$(basename "$au")
    inb=$(grep -lE "href=\"(${au}|\./)?[^\"]*${b}\"" "$TMP"/* 2>/dev/null | wc -l)
    if [ "$c" = 404 ] && [ "$inb" = 0 ]; then echo "      ${au}: 404 and 0 inbound — closed"
    else finding "${au}: served ${c} (want 404), inbound ${inb} (want 0) (6.3 / bugs_open/429)"; fi
  done <<< "$arch"; fi

echo "--- 10.2 brief-echo headings on index pages"
be=$(count_of 'what (they|we) avoid|what gets included|how the entries are written|what the pieces do')
echo "      hits: ${be}"; [ "$be" -gt 0 ] && finding "brief-echo headings served (10.2 / bugs_open/444)"

echo "--- 10.3 CSS-url heroes that resolve to a non-200"
nb=0
while read -r p; do [ -n "$p" ] || continue
  c=$(curl -s -o /dev/null -m 20 -w '%{http_code}' "${BASE}${p}")
  [ "$c" = 200 ] || { nb=$((nb+1)); finding "${c} ${p} — a hero painting a gradient over nothing (10.3)"; }
done < <(body_all | grep -oE "url\('?[^')]+'?\)" | grep -oE '/[^'"'"')]+' | sort -u)
[ "$nb" = 0 ] && echo "      all CSS-url assets resolve 200"

echo "--- 10.4 interactive elements"; echo "      canvas/iframe/video/data-tool: $(count_of '<canvas|<iframe|<video|data-game|data-tool')"

echo
echo "=== OWED, not mechanisable (the checklist says so): 1.2, 1.3, 3.2, 4.3, 7.1, 7.2, 7.3, 9.1-9.4, 10.1, 10.5 ==="
echo "=== §8: a complete work item is not a repaired artefact — every line above is read at the served page or the row, never at a status ==="
echo "=== RESULT: ${FINDINGS} finding(s), ${BLIND} blind check(s) ==="
[ $((FINDINGS+BLIND)) -eq 0 ] && exit 0 || exit 1
