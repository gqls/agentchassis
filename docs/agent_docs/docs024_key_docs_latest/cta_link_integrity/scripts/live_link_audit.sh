#!/bin/bash
# Live audit: fetch every active page on the affected sites, extract internal
# hrefs from the SHIPPED html (not page_components.rendered_html, which is
# demonstrably stale), and record which ones 404.
set -u
cd "$(dirname "$0")"

PAGES=pages.txt
: > live_anchors.txt
: > target_status.txt

total=0
while IFS='|' read -r domain url; do
  [ -z "${domain:-}" ] && continue
  total=$((total+1))
  body=$(curl -s --max-time 20 "https://${domain}${url}")
  # extract internal hrefs, skip anchors/external/mailto/assets
  echo "$body" | grep -oE 'href="/[^"#][^"]*"' \
    | sed 's/^href="//; s/"$//' \
    | grep -vE '\.(css|js|png|jpg|jpeg|svg|webp|ico|woff2?|xml|json|txt|pdf)$' \
    | sort -u \
    | while read -r h; do echo "${domain}|${url}|${h}"; done >> live_anchors.txt
done < "$PAGES"

echo "pages fetched: $total"
echo "live anchor rows: $(wc -l < live_anchors.txt)"

# unique targets -> status
cut -d'|' -f1,3 live_anchors.txt | sort -u | while IFS='|' read -r domain h; do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 15 "https://${domain}${h}")
  echo "${code}|${domain}|${h}" >> target_status.txt
done

echo "=== unique targets by status ==="
cut -d'|' -f1 target_status.txt | sort | uniq -c | sort -rn
