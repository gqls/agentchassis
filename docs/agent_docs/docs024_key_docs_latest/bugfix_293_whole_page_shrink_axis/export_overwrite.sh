#!/usr/bin/env bash
# POSITIVE CONTROL for bugs_open/293's pairing rule.
#
# The rule is uniform: an archive row banks the content that was live UNTIL that
# moment, so the write that followed it is (row.rendered_html → the next archive
# row for the same page+slot, or the LIVE row if there is none). export_pairs.sh
# applies it to op='delete' rows, which is the population nobody had paired.
#
# Applied here to op='overwrite' rows it must reproduce a measurement made
# independently, by another lane, before this join existed: 117 pairs, of which
# the visible axis refuses 3 and the tag-stripped axis refuses 1 (the repair)
# — section_visible_text.go's header table. If the rule cannot reproduce a
# number it did not produce, it has no business being trusted on the delete rows.
set -euo pipefail
OUT="${1:?usage: export_overwrite.sh <out.jsonl>}"
: > "$OUT"
CHUNK=10
OFFSET=0
while :; do
  rows=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
WITH ev AS (
  SELECT page_id, slot_name, created_at, op, rendered_html, position,
         lead(rendered_html) OVER w AS next_html
  FROM page_component_history WHERE slot_name IS NOT NULL AND op IS NOT NULL
  WINDOW w AS (PARTITION BY page_id, slot_name ORDER BY created_at)
), uniqname AS (
  SELECT page_id, slot_name FROM page_component_history
  WHERE slot_name IS NOT NULL AND op IS NOT NULL GROUP BY 1,2 HAVING count(DISTINCT position)=1
), livesingle AS (
  SELECT page_id, slot_name, rendered_html FROM page_components pc
  WHERE slot_name IS NOT NULL
    AND (SELECT count(*) FROM page_components x WHERE x.page_id=pc.page_id AND x.slot_name=pc.slot_name)=1
)
SELECT row_to_json(t) FROM (
  SELECT e.page_id::text AS page_id, e.slot_name, e.created_at::text AS del_at,
         e.op, p.name AS page_name, s.domain AS domain,
         e.rendered_html AS existing,
         COALESCE(e.next_html, l.rendered_html) AS incoming
  FROM ev e
  JOIN uniqname u ON u.page_id=e.page_id AND u.slot_name=e.slot_name
  JOIN pages p ON p.id=e.page_id
  JOIN sites s ON s.id=p.site_id
  LEFT JOIN livesingle l ON l.page_id=e.page_id AND l.slot_name=e.slot_name
  WHERE e.op='overwrite' AND e.rendered_html IS NOT NULL
    AND COALESCE(e.next_html, l.rendered_html) IS NOT NULL
  ORDER BY e.page_id, e.slot_name, e.created_at
  LIMIT $CHUNK OFFSET $OFFSET
) t;" 2>/dev/null)
  [ -z "$rows" ] && break
  printf '%s\n' "$rows" >> "$OUT"
  OFFSET=$((OFFSET+CHUNK))
  echo -n "." >&2
done
echo >&2
wc -l "$OUT"
