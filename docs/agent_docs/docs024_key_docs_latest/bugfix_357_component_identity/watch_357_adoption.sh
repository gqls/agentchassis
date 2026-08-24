#!/usr/bin/env bash
# bugs_open/357 — watch for the first adoption after arming, with real controls.
#
# THREE STATES, NOT TWO (v3). An earlier version tested [ "$BAD" != "0" ], and an
# empty string is != "0" — so a transient query failure raised the same alarm as a
# real defect. A guard that cries wolf on its own plumbing teaches its reader to
# discount the one firing it exists for. A value is compared only when it is a
# number; a blank is retried, then reported UNKNOWN, and five in a row are a
# MEASUREMENT FAILURE saying plainly that nothing is known.
#   exit 0 = adoption observed   exit 2 = stop condition (on a REAL reading)
#   exit 3 = could not measure   exit 1 = window elapsed
#
# THE DEMAND CONTROL IS DB-BASED (v4), and this matters twice over. v3 asked the
# LOGS whether the seam had run and got 0 while the database showed 209 saves
# through it since arming — a kubectl log grep is not a census (wrong pod,
# retention, level). And it asked the wrong question anyway: the seam runs
# constantly on ordinary pages, so "did the seam run" says nothing about whether
# ADOPTION had an opportunity. What does is the fallback signature — a page saved
# with exactly one component row, no <section in its html, and no data-component.
ARMED_AT="2026-08-24 16:15:00Z"
Q() { kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -t -c "$1" 2>/dev/null | tr -d '\r' | head -1; }
num() { [[ "$1" =~ ^[0-9]+$ ]]; }

ADQ="SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.function='adopted-fragment';"
POPQ="SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0;"
BADQ="SELECT count(pc.component_version_id) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0;"
CANDQ="SELECT count(*) FROM page_components pc WHERE pc.created_at > '$ARMED_AT' AND (SELECT count(*) FROM page_components x WHERE x.page_id = pc.page_id) = 1 AND pc.rendered_html NOT ILIKE '%<section%' AND pc.rendered_html !~ 'data-component=\"[^\"]+\"';"
SAVEQ="SELECT count(*) FROM page_components WHERE created_at > '$ARMED_AT' AND content_brief IS NOT NULL;"

unmeasured=0
for i in $(seq 1 60); do
  A=$(Q "$ADQ"); P=$(Q "$POPQ"); B=$(Q "$BADQ")
  if ! num "$A" || ! num "$P" || ! num "$B"; then sleep 5; A=$(Q "$ADQ"); P=$(Q "$POPQ"); B=$(Q "$BADQ"); fi
  if ! num "$A" || ! num "$P" || ! num "$B"; then
    unmeasured=$((unmeasured+1))
    echo "[$(date -u +%H:%M:%S)] UNKNOWN — query returned no value (consecutive=$unmeasured). NOT an alarm."
    if [ "$unmeasured" -ge 5 ]; then echo "### MEASUREMENT FAILURE: five consecutive ticks could not read the database. Nothing is known about the guard."; exit 3; fi
    sleep 60; continue
  fi
  unmeasured=0
  C=$(Q "$CANDQ"); S=$(Q "$SAVEQ"); num "$C" || C="?"; num "$S" || S="?"
  echo "[$(date -u +%H:%M:%S)] adopted=$A population=$P population_stamped=$B adoption_candidates=$C saves_since_arming=$S"
  if [ "$B" -gt 0 ]; then echo "!!! STOP CONDITION (a real reading of $B): a population row acquired a stamp — splice hygiene failed"; exit 2; fi
  if [ "$A" -gt 0 ]; then echo ">>> FIRST ADOPTION OBSERVED ($A rows)"; exit 0; fi
  sleep 60
done
echo "=== window elapsed. Read adoption_candidates FIRST: 0 means no qualifying page was ever saved, so zero adoptions say nothing about correctness. saves_since_arming shows whether the seam is alive at all. ==="
exit 1
