#!/bin/bash
# servegrade.sh <slug> <ported-slot-file> <completed_at 'YYYY-MM-DD HH:MM:SS'> [negatives...]
# GATES on http=200 and on a non-empty, genuinely-later last-modified. Refuses rather than reports.
slug=$1; ported=$2; completed=$3; shift 3
URL="https://webdesign.co.uk/tools/$slug/index.html"
D=$(mktemp -d)
code=$(curl -s -o $D/p.html -D $D/h -w '%{http_code}' "$URL?cb=$(date +%s%N)")
echo "url=$URL"
echo "http=$code bytes=$(wc -c < $D/p.html)"
if [ "$code" != "200" ]; then
  echo "REFUSING TO GRADE: http=$code. Before concluding damage, curl the RECORDED pages.url and a same-form SIBLING."
  exit 1
fi
lm=$(grep -i '^last-modified:' $D/h | sed 's/^[^:]*: //' | tr -d '\r')
if [ -z "$lm" ]; then echo "REFUSING TO GRADE: no last-modified header (cannot prove freshness; an empty value would parse as NOW)"; exit 1; fi
lme=$(date -u -d "$lm" +%s 2>/dev/null) || { echo "REFUSING: unparseable last-modified '$lm'"; exit 1; }
cte=$(date -d "$completed UTC" +%s)   # completed_at comes from Postgres in UTC; -u does NOT make -d parse as UTC
echo "last-modified=$lm"
if [ "$lme" -le "$cte" ]; then
  echo "REFUSING TO GRADE: artefact ($lm) NOT newer than completed_at ($completed) - stale copy"
  cat <<'SKIPHELP'
  Two causes, and they need OPPOSITE responses - do not just re-poll:
  (a) S3 has not published yet          -> wait and re-run (measured lag 11-97s)
  (b) the rerender SILENTLY SKIPPED     -> the item reads complete and the page was never
      reassembled, so waiting is futile. Since bugs_open/408 (chassis >= 6e2d4a039) a
      content_field resolving on no path SKIPS instead of crashing the pod.
      Distinguish them before re-polling - non-null skip_reason means (b):
        SELECT collected_data->'assembled_page'->>'skip_reason'
        FROM orchestration_states
        WHERE collected_data->'input_data'->>'work_item_id' = 'THE-RERENDER-ITEM-ID'
        ORDER BY created_at DESC LIMIT 1;
SKIPHELP
  exit 1
fi
echo "freshness OK: artefact is $((lme-cte))s after completed_at"
echo "--- NEGATIVES (served must be 0; ported must be >=1 or the control is worthless) ---"
fail=0
for pat in "$@"; do
  s=$(grep -o -F -- "$pat" $D/p.html|wc -l); p=$(grep -o -F -- "$pat" "$ported"|wc -l)
  if [ "$p" -lt 1 ]; then v="WORTHLESS(ported=0)"; else [ "$s" -eq 0 ] && v="pass" || { v="FAIL"; fail=1; }; fi
  printf '  %-26s served=%-3s ported=%-3s %s\n' "$pat" "$s" "$p" "$v"
done
echo "--- POSITIVES ---"
printf '  %-26s %s (want >=1)\n' "addEventListener" "$(grep -c addEventListener $D/p.html)"
printf '  %-26s %s (want >=1)\n' "<script" "$(grep -o '<script' $D/p.html|wc -l)"
printf '  %-26s %s (want 0)\n'   "raw {{. template tags" "$(grep -o '{{\.' $D/p.html|wc -l)"
printf '  %-26s %s (want >=1)\n' "scoped instance id" "$(grep -o "id=\"c-tool-$slug[^\"]*\"" $D/p.html|wc -l)"
rm -rf $D
[ $fail -eq 0 ] && echo "RESULT: PASS" || echo "RESULT: FAIL"
