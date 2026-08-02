#!/usr/bin/env bash
# prepare_work.sh — rebuild the decomposition working directory from scratch.
#
# Everything this produces is DERIVED: the stored pages come from the database,
# the assets from the live site, the manifest from decompose_pages.py. Nothing
# here is authored, so it is safe to delete and re-make at any time — which is
# just as well, because a session scratchpad does not survive the session and
# this was originally a sequence of ad-hoc shell that had to be retyped when it
# went. Scripting it also makes the verification reproducible by someone who is
# not the thread that wrote it.
#
#   ./prepare_work.sh <work-dir>
#   export DECOMP_WORK=<work-dir>; python3 verify_assembled.py
#
# BYTE FIDELITY IS CHECKED, NOT ASSUMED. psql -tA appends one newline, so every
# dumped file is one byte longer than the column; the check accounts for that
# and compares md5 as well as length. Do NOT compare against length() — that
# counts CHARACTERS, so every £ in the page reads as one byte short and 20 of
# the 27 pages look corrupt when they are fine.
set -euo pipefail

WORK="${1:?usage: prepare_work.sh <work-dir>}"
SITE_ID="0162cde4-633e-45e9-8ca6-87a6b2fe1d26"
DOMAIN="loancalculator.co.uk"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LANE="$(dirname "$HERE")"

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 --
      psql -U clients_user -d clients_db)

mkdir -p "$WORK/stored" "$WORK/assets/assets/js" "$WORK/assets/assets/css"

echo "== page list =="
"${PSQL[@]}" -tA -F'|' -c \
  "SELECT p.name, p.url FROM pages p WHERE p.site_id='$SITE_ID' ORDER BY p.url;" \
  </dev/null > "$WORK/pages.txt"
echo "   $(wc -l < "$WORK/pages.txt") pages"

echo "== stored verbatim documents =="
while IFS='|' read -r name _url; do
  "${PSQL[@]}" -tA -c \
    "SELECT pc.rendered_html FROM page_components pc JOIN pages p ON p.id=pc.page_id
     WHERE p.site_id='$SITE_ID' AND p.name='$name'
       AND pc.content_data->>'deploy_mode'='verbatim';" </dev/null > "$WORK/stored/$name.html"
done < "$WORK/pages.txt"

"${PSQL[@]}" -tA -F'|' -c \
  "SELECT p.name, octet_length(pc.rendered_html), md5(pc.rendered_html)
   FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='$SITE_ID' AND pc.content_data->>'deploy_mode'='verbatim'
   ORDER BY p.name;" </dev/null > "$WORK/dbmd5.txt"

bad=0
while IFS='|' read -r n l m; do
  f=$(stat -c%s "$WORK/stored/$n.html")
  head -c $((f-1)) "$WORK/stored/$n.html" > "$WORK/.cmp"
  fm=$(md5sum "$WORK/.cmp" | cut -d' ' -f1)
  if [ "$((f-1))" -ne "$l" ] || [ "$fm" != "$m" ]; then
    echo "   MISMATCH $n file=$((f-1))/$fm db=$l/$m"; bad=1
  fi
done < "$WORK/dbmd5.txt"
rm -f "$WORK/.cmp"
[ $bad -eq 0 ] || { echo "dump is not byte-exact — stop"; exit 1; }
echo "   all $(wc -l < "$WORK/dbmd5.txt") documents byte-exact (octet_length + md5)"

echo "== assets =="
# The DECOMPOSER READS THESE, it does not merely copy them: an external script's
# getElementById targets decide what is tool and what is prose, so a missing file
# is a hard failure rather than a slower page. See decompose_pages.py.
for p in /assets/js/nav.js /assets/js/global.js /assets/js/snippets.js /assets/css/style.css; do
  code=$(curl -s -o "$WORK/assets$p" -w '%{http_code}' "https://$DOMAIN$p")
  [ "$code" = "200" ] || { echo "   $p returned $code — stop"; exit 1; }
  echo "   $p $(stat -c%s "$WORK/assets$p")b"
done

echo "== decompose =="
python3 "$LANE/decompose_pages.py" "$WORK/stored" "$WORK/manifest.json" --assets "$WORK/assets"
