#!/usr/bin/env bash
# Guarded whole-site serving check for loancalculator.co.uk.
#
# Why the guard: a fetch taken inside a deploy window returns a B2 error blob at
# HTTP 200. Every grep against that blob reads clean, so "0 occurrences of the
# bad phrase" is indistinguishable from success. HTTP 200 alone proves nothing —
# each page must also be big enough to be a page and start with a DOCTYPE.
#
# Usage: ./check_site_serving.sh [pages_file]
#   pages_file defaults to a live query for the site's active page URLs.
#
# Exit 0 only if every page is 200 AND >2000 bytes AND starts with <!DOCTYPE.
set -uo pipefail

SITE_ID='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
BASE='https://loancalculator.co.uk'
MIN_BYTES=2000

if [ $# -ge 1 ] && [ -f "$1" ]; then
  PAGES=$(cat "$1")
else
  PAGES=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -c \
    "SELECT url FROM pages WHERE site_id='$SITE_ID' AND status='active' ORDER BY url;" 2>/dev/null)
fi

total=0; ok=0; bad=0
while IFS= read -r path; do
  [ -z "$path" ] && continue
  total=$((total+1))
  tmp=$(mktemp)
  code=$(curl -s -o "$tmp" -w '%{http_code}' "$BASE$path")
  bytes=$(wc -c < "$tmp")
  first=$(head -c 9 "$tmp")
  if [ "$code" = "200" ] && [ "$bytes" -ge "$MIN_BYTES" ] && [ "$first" = "<!DOCTYPE" ]; then
    ok=$((ok+1))
  else
    bad=$((bad+1))
    echo "FAIL $path  http=$code bytes=$bytes first='$first'"
  fi
  rm -f "$tmp"
done <<< "$PAGES"

echo "serving: $ok/$total pass, $bad fail (guard: 200 + >=${MIN_BYTES}B + DOCTYPE)"
[ "$bad" -eq 0 ]
