#!/usr/bin/env bash
# Export the DELETE+INSERT rebuild pairs for bugs_open/293's calibration.
#
# THE JOIN, and why it is exact rather than plausible:
#   migration 357's DELETE trigger banks OLD.rendered_html, so an archived
#   delete row IS the content that was live immediately before a rebuild.
#   A whole-page save (save_page_sections_action.go:788) DELETEs every
#   agent-writable row for the page and re-INSERTs — so the row that replaced
#   it is the LIVE page_components row, and page_components.created_at proves
#   the re-insert belongs to THAT rebuild (filtered to <60s here; 1,109 of
#   1,123 are within 5s). Disconfirming control, run 2026-08-17: ZERO live rows
#   are OLDER than the last delete of their own (page, slot) — which the
#   DELETE+INSERT model predicts and a wrong join would not.
#
# Chunked 20 rows at a time: a single json_agg row of this size dies mid-stream
# with `unexpected EOF` (NOTES_shared_template_write.md, 2026-08-17).
set -euo pipefail
OUT="${1:?usage: export_pairs.sh <out.jsonl>}"
: > "$OUT"
CHUNK=20
OFFSET=0
while :; do
  rows=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
WITH lastdel AS (
  SELECT DISTINCT ON (page_id, slot_name, position)
         page_id, slot_name, position, created_at AS del_at, rendered_html AS existing_html, application_name
  FROM page_component_history
  WHERE op='delete' AND slot_name IS NOT NULL
  ORDER BY page_id, slot_name, position, created_at DESC
),
-- SLOT NAMES REPEAT ON A PAGE (14 pages, 32 rows, measured 2026-08-17), so a
-- name-only join emits a CARTESIAN PRODUCT and manufactures pairs that never
-- existed. Restricting to page+slot groups of exactly one on BOTH sides is what
-- makes the remaining join exact; the duplicate-slot pages are a finding of
-- their own (the shipped guard keys its comparison on slot_name alone) and are
-- reported separately rather than measured here.
uniq_del AS (
  SELECT page_id, slot_name FROM lastdel GROUP BY 1,2 HAVING count(*)=1
),
uniq_live AS (
  SELECT page_id, slot_name FROM page_components WHERE slot_name IS NOT NULL GROUP BY 1,2 HAVING count(*)=1
)
SELECT row_to_json(t) FROM (
  SELECT d.page_id::text AS page_id, d.slot_name, d.del_at::text AS del_at,
         d.application_name, p.name AS page_name, s.domain AS domain,
         d.existing_html AS existing, pc.rendered_html AS incoming,
         EXTRACT(EPOCH FROM (pc.created_at - d.del_at))::numeric(10,3) AS reinsert_lag_s
  FROM lastdel d
  JOIN uniq_del ud ON ud.page_id=d.page_id AND ud.slot_name=d.slot_name
  JOIN uniq_live ul ON ul.page_id=d.page_id AND ul.slot_name=d.slot_name
  JOIN page_components pc ON pc.page_id=d.page_id AND pc.slot_name=d.slot_name
  JOIN pages p ON p.id=d.page_id
  JOIN sites s ON s.id=p.site_id
  WHERE pc.created_at >= d.del_at AND pc.created_at < d.del_at + interval '60 seconds'
    AND pc.rendered_html IS NOT NULL AND d.existing_html IS NOT NULL
  ORDER BY d.page_id, d.slot_name
  LIMIT $CHUNK OFFSET $OFFSET
) t;" 2>/dev/null)
  [ -z "$rows" ] && break
  printf '%s\n' "$rows" >> "$OUT"
  OFFSET=$((OFFSET+CHUNK))
  echo -n "." >&2
done
echo >&2
wc -l "$OUT"
