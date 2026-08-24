#!/usr/bin/env bash
# Watch v3 — THREE states, not two.
#
# v2 fired its STOP condition on a transient query failure: the test was
# [ "$BAD" != "0" ], and an empty string is != "0", so "I could not measure"
# raised the same alarm as "the bad thing happened". That is worse than no alarm,
# because the one real firing would then arrive indistinguishable from noise.
#
# So: a value is only compared when it IS a number. Absent is UNKNOWN, retried,
# and — if it persists — reported as a MEASUREMENT FAILURE with its own exit code.
#   exit 0 = adoption observed        exit 2 = stop condition (a real reading)
#   exit 3 = could not measure        exit 1 = window elapsed
Q() { kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -t -c "$1" 2>/dev/null | tr -d '\r' | head -1; }
num() { [[ "$1" =~ ^[0-9]+$ ]]; }
POPQ="SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0;"
BADQ="SELECT count(pc.component_version_id) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0;"
ADQ="SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.function='adopted-fragment';"
unmeasured=0
for i in $(seq 1 60); do
  A=$(Q "$ADQ"); P=$(Q "$POPQ"); B=$(Q "$BADQ")
  if ! num "$A" || ! num "$P" || ! num "$B"; then
    sleep 5; A=$(Q "$ADQ"); P=$(Q "$POPQ"); B=$(Q "$BADQ")   # one retry
  fi
  if ! num "$A" || ! num "$P" || ! num "$B"; then
    unmeasured=$((unmeasured+1))
    echo "[$(date -u +%H:%M:%S)] UNKNOWN — query returned no value (consecutive=$unmeasured). NOT an alarm."
    if [ "$unmeasured" -ge 5 ]; then echo "### MEASUREMENT FAILURE: 5 consecutive ticks could not read the database. Nothing is known about the guard."; exit 3; fi
    sleep 60; continue
  fi
  unmeasured=0
  RAN=$(kubectl -n ai-persona-system logs -l app=agent-chassis --since=5m --tail=20000 2>/dev/null | grep -c "enrichSectionsWithComponentIDs: invoked")
  num "$RAN" || RAN="?"
  echo "[$(date -u +%H:%M:%S)] adopted=$A population=$P population_stamped=$B seam_invocations_5m=$RAN"
  if [ "$B" -gt 0 ]; then echo "!!! STOP CONDITION (a real reading of $B): a population row acquired a stamp — splice hygiene failed"; exit 2; fi
  if [ "$A" -gt 0 ]; then echo ">>> FIRST ADOPTION OBSERVED ($A rows)"; exit 0; fi
  sleep 60
done
echo "=== window elapsed. Read seam_invocations_5m first: all-zero means the seam never ran, so zero adoptions say NOTHING about correctness. ==="
exit 1
