#!/usr/bin/env bash
# check_gtm_state.sh — one command for the three questions people keep asking about Google
# tracking on this estate, each answered at the place that can actually answer it.
#
#   (default)  Is a GA4 tag PUBLISHED in the container?   → reads the LIVE gtm.js. Nothing else can
#              answer this: the snippet being on every page proves nothing (it was on 24 sites for
#              weeks with 0 tags, recording nothing — 039_REFERENCE §3).
#   --sites    Which deployed sites actually SERVE the snippet?  → one curl per domain, redirects
#              followed. ⚠ every run is our own traffic and Cloudflare counts it (039 §1).
#   --db       Which sites carry it DURABLY?  → the four-bucket census: spec key vs stored head
#              artefact. "artefact only" = lost on the next chrome render (bugs_open/397).
#   --all      all three.
#
# Exit 0 always — it is a report. Read the VERDICT lines.
set -uo pipefail
CONTAINER="${CONTAINER:-GTM-PQ3WCTBD}"
NS="${NS:-ai-persona-system}"
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F $'\t')
DO_SITES=0; DO_DB=0
for a in "$@"; do case "$a" in --sites) DO_SITES=1;; --db) DO_DB=1;; --all) DO_SITES=1; DO_DB=1;; *) echo "unknown flag $a" >&2; exit 2;; esac; done

TMP=$(mktemp -d "${TMPDIR:-/tmp}/gtmstate.XXXXXX"); trap 'rm -rf "$TMP"' EXIT

echo "== 1. container $CONTAINER, live at googletagmanager.com  [$(date -u '+%Y-%m-%d %H:%MZ')]"
code=$(curl -s --max-time 25 -o "$TMP/gtm.js" -w '%{http_code}' "https://www.googletagmanager.com/gtm.js?id=$CONTAINER")
if [ "$code" != "200" ]; then echo "   VERDICT: container fetch failed (http=$code) — cannot say anything about GA4"; else
  ver=$(grep -oE '"version":"[0-9]+"' "$TMP/gtm.js" | head -1 | grep -oE '[0-9]+')
  # Count tag templates FILE-WIDE, not inside an extracted "tags":[...] blob: the 2026-09 container
  # format dropped the ],"predicates" anchor the old extraction relied on, so the blob came back
  # empty and a PUBLISHED container read as 0 tags (caught 2026-09-02, minutes after the owner's
  # first publish — the false verdict said his publish had not worked). Template code only appears
  # in gtm.js when a tag of that type exists, so the file-wide count is the honest signal.
  googtag=$(grep -o '"function":"__googtag"' "$TMP/gtm.js" | wc -l)
  gaawe=$(grep -o '"function":"__gaawe"' "$TMP/gtm.js" | wc -l)
  ntags=$((googtag + gaawe))
  gids=$(grep -oE 'G-[A-Z0-9]{6,12}' "$TMP/gtm.js" | sort -u | tr '\n' ' ')
  printf '   version=%s tags=%s google_tag=%s ga4_event_tags=%s measurement_ids=[%s]\n' "${ver:-?}" "$ntags" "$googtag" "$gaawe" "${gids% }"
  if [ "$ntags" -gt 0 ] && [ -n "$gids" ]; then
    echo "   VERDICT: GA4 PUBLISHED — tag(s) live, reporting to ${gids% }. Realtime should show a visit within seconds."
    [ "$googtag" -eq 0 ] && echo "   ⚠ no 'Google Tag' (__googtag) in the container — a GA4 EVENT tag alone does not send page views (039 §3)."
  else
    echo "   VERDICT: GA4 NOT PUBLISHED — version $ver has $ntags tag(s); the snippet loads on every site and records NOTHING."
    echo "            Save is not enough: Workspace → Submit → Publish (039 §3, apis.uk HANDOFF 2026-08-25 §4a)."
  fi
fi

if [ "$DO_SITES" = 1 ]; then
  echo; echo "== 2. served: one curl per deployed/active site with a head component (this IS our own traffic)"
  "${PSQL[@]}" -c "SELECT s.domain FROM sites s WHERE s.status IN ('deployed','active') AND EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id=s.id AND sc.slot_name='head') ORDER BY 1" > "$TMP/domains.txt" 2>/dev/null || { echo "   (DB unreachable — cannot list domains)"; DO_SITES=0; }
fi
if [ "$DO_SITES" = 1 ]; then
  n=$(grep -c . "$TMP/domains.txt")
  # mapfile, not a while-read loop: LANDMINES 2026-08-24, a stdin-forwarding call inside the loop eats the list
  mapfile -t DOMS < "$TMP/domains.txt"
  printf '%s\n' "${DOMS[@]}" | xargs -P 6 -I{} sh -c 'b=$(curl -sL --max-time 25 -w "\n__HTTP=%{http_code} __URL=%{url_effective}" "https://{}/"); c=$(printf "%s" "$b" | grep -c googletagmanager); h=$(printf "%s" "$b" | tail -1); printf "%-34s gtm=%s %s\n" "{}" "$c" "$h"' | sort > "$TMP/served.txt"
  cat "$TMP/served.txt" | sed 's/^/   /'
  got=$(grep -c . "$TMP/served.txt"); serving=$(grep -c 'gtm=[1-9]' "$TMP/served.txt" || true)
  echo "   VERDICT: $serving of $got sites serve the snippet (list=$n, checked=$got — these MUST match)"
fi

if [ "$DO_DB" = 1 ]; then
  echo; echo "== 3. durable or not: spec key (what the template reads) vs stored head artefact (what is served)"
  "${PSQL[@]}" <<SQL 2>/dev/null | sed 's/^/   /' || echo "   (DB unreachable)"
WITH x AS (
  SELECT s.domain, s.status,
         EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id=s.id AND ss.aspect='site_config' AND ss.is_current
                    AND ss.data->'analytics'->>'gtm_container_id' = '$CONTAINER') AS spec_key,
         COALESCE((SELECT sc.rendered_html LIKE '%$CONTAINER%' FROM site_components sc WHERE sc.site_id=s.id AND sc.slot_name='head'), false) AS head_artefact,
         EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id=s.id AND sc.slot_name='head') AS has_head
    FROM sites s WHERE s.status IN ('deployed','active'))
SELECT CASE WHEN spec_key AND head_artefact THEN 'A durable (spec+artefact)'
            WHEN NOT spec_key AND head_artefact THEN 'B ARTEFACT ONLY - reverts on next chrome render'
            WHEN spec_key AND NOT head_artefact THEN 'C spec only - needs a render'
            WHEN NOT has_head THEN 'E no head component'
            ELSE 'D neither (untagged)' END AS bucket,
       count(*), string_agg(domain, ' ' ORDER BY domain)
  FROM x GROUP BY 1 ORDER BY 1;
SQL
  echo "   (bucket B is bugs_open/397; the fix is sql/c2_gtm_spec_key_for_artefact_only_sites.sql — owner-gated, it triggers a per-site rebuild)"
fi
