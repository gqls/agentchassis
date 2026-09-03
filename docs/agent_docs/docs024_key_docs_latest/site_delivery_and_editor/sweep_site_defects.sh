#!/usr/bin/env bash
# sweep_site_defects.sh — run the mechanisable checks of
# docs024_key_docs_latest/SITE_DEFECT_CATEGORIES.md against ONE site.
#
#   ./sweep_site_defects.sh <domain>            # e.g. boxingonline.com
#
# WHY THIS EXISTS. The checklist is prose with ~30 checks; before this, each
# pre-delivery sweep re-typed them, and the §0 disciplines (serving host, a
# must-be-present control on every probe, pages enumerated from the DB, the
# artefact dated) were re-remembered rather than executed. This runs the ones
# that are mechanisable and PRINTS the ones that are not, so a "clean sweep"
# cannot silently mean "I ran the easy half".
#
# §0 DISCIPLINES, BUILT IN (each is a real defence, not decoration):
#  - serving host resolved from sites.publish_project / publish_target, never
#    the customer domain (a parked domain 200s every path);
#  - an invented-path 404 control on every fetch pass, and a must-be-present
#    string control per page — a control of 0 marks the page BLIND, never clean;
#  - pages enumerated from `pages WHERE deployed_at IS NOT NULL`;
#  - served last-modified printed beside pages.deployed_at (the 420
#    discriminator: older served = mirror lag, wait; newer = look upstream);
#  - Postgres regex uses \y, never \b (which is BACKSPACE and returns a clean 0);
#  - every inner `kubectl exec -i` gets </dev/null so it cannot eat the loop's
#    stdin and report one row, cleanly.
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

SITE_ID=$(q "SELECT id FROM sites WHERE domain='${DOMAIN}';" | tr -d ' ')
[ -n "$SITE_ID" ] || { echo "no site row for ${DOMAIN}" >&2; exit 2; }
HOST=$(q "SELECT coalesce(nullif(publish_project,''), domain) FROM sites WHERE id='${SITE_ID}';" | tr -d ' ')
BASE="https://${HOST}"
echo "=== sweep ${DOMAIN} — site ${SITE_ID} — serving host ${HOST} — $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

CTL=$(curl -s -o /dev/null -m 20 -w '%{http_code}' "${BASE}/zzz-not-real-${RANDOM}.html")
echo "[control] invented path -> ${CTL}  $([ "$CTL" = 404 ] && echo OK || echo '⚠ NOT 404 — parked/catch-all host, every result below is suspect')"

TMP=$(mktemp -d); trap 'rm -rf "$TMP"' EXIT
mapfile -t URLS < <(q "SELECT url FROM pages WHERE site_id='${SITE_ID}' AND deployed_at IS NOT NULL AND status='active' ORDER BY url;")
echo "[pages] ${#URLS[@]} deployed+active pages enumerated from the DB"
for u in "${URLS[@]}"; do
  f="$TMP/$(echo "$u" | tr '/' '_')"
  code=$(curl -s -m 25 -o "$f" -w '%{http_code}' "${BASE}${u}?cb=$RANDOM$RANDOM")
  lm=$(curl -s -m 25 -o /dev/null -w '%header{last-modified}' "${BASE}${u}?cb=$RANDOM$RANDOM")
  dep=$(q "SELECT to_char(deployed_at,'YYYY-MM-DD HH24:MI:SS') FROM pages WHERE site_id='${SITE_ID}' AND url='${u}';")
  present=$(grep -c '</html>' "$f" 2>/dev/null || echo 0)
  printf '%s %-46s %s  served=%s  deployed_at=%s  control(</html>)=%s\n' \
    "$([ "$code" = 200 ] && [ "$present" -gt 0 ] && echo ' ' || echo '⚠')" "$u" "$code" "${lm:-none}" "$dep" "$present"
done

# HTML ONLY. A binary file in $TMP (the logo we fetch for 4.5) makes grep report
# "binary file matches" and STOP COUNTING — a silent false-clean on every text
# check downstream. Keep fetched assets out of the corpus by extension.
body_all() { for f in "$TMP"/*; do case "$f" in *.png|*.jpg|*.jpeg|*.webp|*.ico) continue;; esac; [ -f "$f" ] && cat "$f"; done 2>/dev/null; }
n_all() { body_all | grep -oiE "$1" | wc -l; }

echo; echo "--- 1.1 meta-copy (first-person-plural policy prose), per page with context"
for f in "$TMP"/*; do case "$f" in *.png|*.jpg|*.jpeg|*.webp|*.ico) continue;; esac; [ -f "$f" ] || continue
  n=$(grep -oiE "we write|we'd rather|we cover|gets checked|we update" "$f" | wc -l)
  [ "$n" -gt 0 ] && printf '      %-46s %s\n' "$(basename "$f" | tr '_' '/')" "$n"
done || true
body_all | grep -ohiE ".{45}(we write|we'd rather|we cover|gets checked|we update).{45}" 2>/dev/null | sort -u | sed 's/^/        /' | head -6 || true
echo "--- 1.4 raw source residue"; echo "    markdown links '](http': $(n_all '\]\(http')   ellipses: $(n_all '\.\.\.')"
echo "--- 1.5 AI tells (context printed: 'honest'/'plainly' are weak needles and need a reader — checklist 1.2's rule)"
body_all | grep -ohiE '.{40}(plainly|honest|starting point, not the final word|before your [a-z]+ have to).{40}' 2>/dev/null | sort -u | sed 's/^/      /' | head -8 || true
echo "    control-labelled fields over 120 chars (DB):"
q "SELECT '      '||cc.name||'.'||k||' = '||length(v)||' chars' FROM page_components pc JOIN pages p ON p.id=pc.page_id LEFT JOIN content_components cc ON cc.id=pc.component_id, jsonb_each_text(pc.content_data) AS e(k,v) WHERE p.site_id='${SITE_ID}' AND (k ~ '(button|label|_cta$|^cta)') AND length(v) > 120 ORDER BY length(v) DESC;" || true
echo "--- 2.2 index-role pages and their listable-item counts (zero is a finding)"
q "SELECT '      '||p.url||' -> '||coalesce(max(coalesce(jsonb_array_length(pc.content_data->'articles'), jsonb_array_length(pc.content_data->'items'))),0)||' items (max over '||count(pc.id)||' listing component(s))' FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id AND (pc.content_data ? 'articles' OR pc.content_data ? 'items') WHERE p.site_id='${SITE_ID}' AND p.deployed_at IS NOT NULL AND p.page_type ~ '(index|directory)' GROUP BY p.url ORDER BY p.url;" || true
echo "--- 2.3 nav: distinct hrefs vs distinct labels in <header>"
for f in "$TMP"/_index.html; do [ -f "$f" ] || continue
  h=$(sed -n '/<header/,/<\/header>/p' "$f"); echo "    hrefs $(printf '%s' "$h" | grep -oE 'href="[^"]+"' | sort -u | wc -l)  labels $(printf '%s' "$h" | grep -oE '>[A-Za-z][^<]{1,30}</a>' | sort -u | wc -l)"; done
echo "--- 2.4 orphans (zero inbound links from any other deployed page)"
for u in "${URLS[@]}"; do
  [ "$u" = "/index.html" ] && continue
  inb=$(grep -l "href=\"${u}\"" "$TMP"/* 2>/dev/null | grep -v "$TMP/$(echo "$u" | tr '/' '_')" | wc -l)
  [ "$inb" = 0 ] && echo "    ⚠ ORPHAN (0 inbound): $u"
done
echo "--- 2.6 nav reach per page-type family"
q "SELECT '      '||page_type||': '||count(*)||' deployed' FROM pages WHERE site_id='${SITE_ID}' AND deployed_at IS NOT NULL GROUP BY page_type ORDER BY page_type;" || true
# ⚠ Count BOTH: a family absent from <header> but linked from footer/body on
# every page IS reachable — the checklist asks "does at least one served nav
# label reach it (directly or via a hub)". Header-only was measured on
# boxingonline 2026-09-03 and would have reported guides+blog unreachable while
# 19 of 20 pages link them.
for fam in tool guide blog-post; do
  pat=$(q "SELECT string_agg(DISTINCT split_part(ltrim(url,'/'),'/',1), '|') FROM pages WHERE site_id='${SITE_ID}' AND page_type='${fam}' AND deployed_at IS NOT NULL;")
  [ -n "${pat:-}" ] || continue
  hdr=$(sed -n '/<header/,/<\/header>/p' "$TMP"/_index.html 2>/dev/null | grep -cE "href=\"/(${pat})" || true)
  any=$(grep -lE "href=\"/(${pat})" "$TMP"/* 2>/dev/null | wc -l)
  flag=' '; [ "$any" = 0 ] && flag='⚠'
  echo "      ${flag} ${fam}: header links ${hdr}, pages linking it anywhere ${any}/${#URLS[@]} $([ "$hdr" = 0 ] && [ "$any" -gt 0 ] && echo '(reachable, but not from the header)')"
done
echo "--- 3.1 tools: STORED component markup (never the served page — chrome adds a menu button)"
q "SELECT '      '||p.url||': inputs '||(SELECT count(*) FROM regexp_matches(pc.rendered_html,'<input','g'))||' selects '||(SELECT count(*) FROM regexp_matches(pc.rendered_html,'<select','g'))||' textareas '||(SELECT count(*) FROM regexp_matches(pc.rendered_html,'<textarea','g')) FROM pages p JOIN page_components pc ON pc.page_id=p.id JOIN content_components cc ON cc.id=pc.component_id WHERE p.site_id='${SITE_ID}' AND p.page_type='tool' AND cc.component_level='tool' ORDER BY p.url;" || true
echo "    stored-newer-than-deployed (repair written, not served — pair with served bytes):"
q "SELECT '      ⚠ '||p.url||' components '||to_char(max(pc.updated_at),'HH24:MI')||' > deployed '||to_char(p.deployed_at,'HH24:MI') FROM pages p JOIN page_components pc ON pc.page_id=p.id WHERE p.site_id='${SITE_ID}' AND p.deployed_at IS NOT NULL AND p.build_status='deployed' GROUP BY p.url, p.deployed_at HAVING max(pc.updated_at) > p.deployed_at + interval '2 minutes';" || true
echo "--- 3.3 evidence_base facts"; echo "    $(q "SELECT coalesce(jsonb_array_length(data->'facts'),0)||' facts' FROM site_specs WHERE site_id='${SITE_ID}' AND aspect='evidence_base' AND is_current;" || echo 'no evidence_base row')"
echo "--- 4.1/4.2 imagery: BOTH encodings, per page"
for u in "${URLS[@]}"; do f="$TMP/$(echo "$u" | tr '/' '_')"; [ -f "$f" ] || continue
  printf '      %-46s <img> %-3s css-url %s\n' "$u" "$(grep -o '<img[^>]*>' "$f" | wc -l)" "$(grep -oE 'background-image:[^;}]*url\([^)]*\)' "$f" | wc -l)"; done
echo "    distinct hero urls: $(body_all | grep -oE "url\('?[^')]+'?\)" | sort -u | wc -l)"
echo "--- 4.5 logo alpha (colour type 4/6 OR tRNS — test both)"
curl -s -m 25 -o "$TMP/../logo_probe.png" "${BASE}/assets/images/logo.png?cb=$RANDOM" && python3 - "$TMP/../logo_probe.png" <<'PY' || echo "      logo fetch failed"
import struct,sys
d=open(sys.argv[1],'rb').read()
if d[:8]!=b'\x89PNG\r\n\x1a\n': print(f"      NOT a PNG (magic {d[:4].hex()}) — 4.4/4.5 both suspect"); raise SystemExit
w,h,bd,ct=struct.unpack('>IIBB',d[16:26]); trns=b'tRNS' in d
print(f"      {w}x{h} depth={bd} colour_type={ct} tRNS={trns} -> {'OK (real alpha)' if ct in (4,6) or trns else '⚠ 4.5 baked background'}")
PY
echo "--- 5.1 empty slot elements"; echo "    hits: $(body_all | grep -oE 'class="[^"]*__(excerpt|category|date|meta)"></(p|span|div)>' | wc -l)"
echo "--- 5.3 unsubstituted placeholders"
# ⚠ THREE false-positive arms, each measured on boxingonline 2026-09-03 — do not
# loosen them back: (a) `placehold` as a substring matches the legitimate HTML
# attribute `placeholder="e.g. …"` (5 hits, 0 defects), so require the .co host
# or a standalone token; (b) `example` matches ordinary prose ("the clearest
# example"), so the token is case-SENSITIVE; (c) an ALL_CAPS_UNDERSCORE token is
# only unsubstituted if nothing DECLARES it — `var TOTAL_QUESTIONS = 10` is a
# working JS constant and was reported as a placeholder until this check asked.
echo "    placehold.co / hard tokens: $(body_all | grep -oE 'placehold\.co|\bEXAMPLE\b|\bTODO\b|\bLorem ipsum\b|_ADDRESS\b' | wc -l)"
ALLCAPS=$(body_all | grep -oE '\b[A-Z]{3,}_[A-Z_]{2,}\b' | sort -u)
for t in $ALLCAPS; do
  # ⚠ NOT `grep -q`: under `set -o pipefail` grep -q exits on the first match,
  # the upstream `cat` takes SIGPIPE (141), and the PIPELINE reports failure even
  # though it matched — so an if-guard inverts and this check warned on every
  # token, including `var TOTAL_QUESTIONS = 10`. Count instead; -c reads to EOF.
  ndecl=$(body_all | grep -cE "(var|let|const|window\.)[[:space:]]*${t}[[:space:]]*=" || true)
  if [ "${ndecl:-0}" -gt 0 ]; then echo "      ${t}: DECLARED in-document (${ndecl} declaration line(s)) — a working constant, not a placeholder";
  else echo "      ⚠ ${t}: no declaration found in the served bytes — candidate unsubstituted token"; fi
done
echo "--- 6.1 ordering party's email on the site (must be 0)"
# ⚠ An EMPTY needle makes `grep -Fc ""` meaningless and the check reads clean
# while measuring nothing — this is the owner's item 0, so it declares itself
# BLIND instead. Try both known homes for the ordering identity.
OE=$(q "SELECT coalesce(direction->>'customer_email','') FROM build_queue WHERE lower(domain)=lower('${DOMAIN}') ORDER BY created_at DESC LIMIT 1;" | tr -d ' ')
[ -n "$OE" ] || OE=$(q "SELECT coalesce(direction->>'email','') FROM build_queue WHERE lower(domain)=lower('${DOMAIN}') ORDER BY created_at DESC LIMIT 1;" | tr -d ' ')
if [ -n "$OE" ]; then
  echo "    served hits for the order email (must be 0): $(body_all | grep -Fci "$OE")   [needle length $(printf '%s' "$OE" | wc -c), control: a known-present string 'Boxing' -> $(body_all | grep -Fci 'Boxing')]"
else
  echo "    ⚠ BLIND — no ordering email found in build_queue.direction for ${DOMAIN}; this check measured NOTHING. Find the needle before reading a zero here."
fi
echo "    sites.email: $(q "SELECT coalesce(nullif(email,''),'(empty)') FROM sites WHERE id='${SITE_ID}';")"
echo "    any mailto/tel on served pages: $(n_all 'mailto:|tel:')"
echo "--- 6.2 forms that submit nowhere"
q "SELECT '      form_action='||coalesce(fa,'(null)')||' x'||n FROM (SELECT pc.content_data->>'form_action' AS fa, count(*) AS n FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE p.site_id='${SITE_ID}' AND pc.content_data ? 'form_action' GROUP BY 1) t ORDER BY n DESC;" || true
echo "--- 6.3 retracted pages: archived in DB, still serving? (404 AND zero inbound is the close)"
while read -r au; do [ -n "$au" ] || continue
  c=$(curl -s -o /dev/null -m 20 -w '%{http_code}' "${BASE}${au}?cb=$RANDOM"); inb=$(grep -l "href=\"${au}\"" "$TMP"/* 2>/dev/null | wc -l)
  echo "      ${au}: served ${c} (want 404), inbound links ${inb} (want 0)"
done < <(q "SELECT url FROM pages WHERE site_id='${SITE_ID}' AND status='archived';")
echo "--- 10.2 brief-echo headings on index pages"; echo "    hits: $(n_all 'what (they|we) avoid|what gets included|how the entries are written|what the pieces do')"
echo "--- 10.3 CSS-url heroes that resolve to a non-200"
body_all | grep -oE "url\('?[^')]+'?\)" | grep -oE '/[^'"'"')]+' | sort -u | while read -r p; do
  c=$(curl -s -o /dev/null -m 20 -w '%{http_code}' "${BASE}${p}"); [ "$c" = 200 ] || echo "      ⚠ ${c} ${p}"; done
echo "--- 10.4 interactive elements"; echo "    canvas/iframe/video/data-tool: $(n_all '<canvas|<iframe|<video|data-game|data-tool')"
echo; echo "=== OWED, not mechanisable (the checklist says so): 1.2, 1.3, 3.2, 4.3, 7.1, 7.2, 7.3, 9.1-9.4, 10.1, 10.5 ==="
echo "=== §8: a complete work item is not a repaired artefact — every line above is read at the served page or the row, not at a status ==="
