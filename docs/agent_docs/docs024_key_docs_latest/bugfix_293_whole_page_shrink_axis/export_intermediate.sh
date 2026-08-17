#!/usr/bin/env bash
# INTERMEDIATE pairs for bugs_open/293: consecutive archive events for the same
# (page, slot). Each archive row banks the content that was live UNTIL that
# moment, so event[i].rendered_html → event[i+1].rendered_html is the write that
# happened in between.
#
# WEAKER THAN THE TERMINAL JOIN, deliberately included. The terminal join
# (export_pairs.sh) has page_components.created_at to prove the re-insert
# belongs to that rebuild; here there is nothing to prove the slot was
# re-inserted by the SAME rebuild rather than re-created later, so a dropped-then-
# restored slot yields a pair that never existed as a write. That is why this set
# is used only to LOOK FOR hollowings — every refusal is hand-inspected — and
# never to count false refusals.
#
# Restricted to (page, slot) groups with a single position, because slot names
# repeat on 14 pages and a name-only join across duplicates is a cartesian
# product (it manufactured a refusal on the first run of this calibration).
set -euo pipefail
OUT="${1:?usage: export_intermediate.sh <out.jsonl>}"
: > "$OUT"
CHUNK=10
OFFSET=0
while :; do
  rows=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
WITH ev AS (
  SELECT page_id, slot_name, created_at, op, rendered_html,
         lead(created_at) OVER w AS next_at, lead(rendered_html) OVER w AS next_html
  FROM page_component_history WHERE slot_name IS NOT NULL AND op IS NOT NULL
  WINDOW w AS (PARTITION BY page_id, slot_name ORDER BY created_at)
), uniqname AS (
  SELECT page_id, slot_name FROM page_component_history
  WHERE op='delete' AND slot_name IS NOT NULL GROUP BY 1,2 HAVING count(DISTINCT position)=1
)
SELECT row_to_json(t) FROM (
  SELECT e.page_id::text AS page_id, e.slot_name, e.created_at::text AS del_at,
         e.op, p.name AS page_name, s.domain AS domain,
         e.rendered_html AS existing, e.next_html AS incoming,
         EXTRACT(EPOCH FROM (e.next_at - e.created_at))::numeric(12,1) AS gap_s
  FROM ev e
  JOIN uniqname u ON u.page_id=e.page_id AND u.slot_name=e.slot_name
  JOIN pages p ON p.id=e.page_id
  JOIN sites s ON s.id=p.site_id
  WHERE e.next_html IS NOT NULL AND e.rendered_html IS NOT NULL
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
