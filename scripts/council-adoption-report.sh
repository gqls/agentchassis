#!/usr/bin/env bash
# council-adoption-report.sh — did the 2026-07-27 council changes change BEHAVIOUR?
#
# Three times this session I wrote that the honest test is not whether the text is
# in place but whether the seats use it, and left that as a note. A note is not an
# instrument. This is the instrument.
#
# WHAT IT CAN AND CANNOT SEE. The stored council_report keeps only
# {reviewer, verdict, objections, missing, notes, degraded} — `checks` and
# `code_checks` are NOT persisted (verified 2026-07-27: 0 of 2,138 stored reviews
# carry either key). So this CANNOT measure whether a seat issued a SQL check. It
# measures what survives: the text of notes and objections. A seat that queried its
# minutes and said nothing about it is invisible here, and that is a real blind spot,
# not an oversight — record it when quoting these numbers.
#
# CUTOVER. Do not guess this. Take it from the config rows themselves:
#   SELECT type, updated_at FROM agent_definitions
#    WHERE type IN ('council-gate','fix-proposer','feature-designer')
#      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
# The first draft of this script hardcoded ~13:00 from memory and was 45 minutes
# early, which silently reclassified five pre-change council runs as post-change
# evidence. The true value was 13:44:56 — and the five runs at 13:00-13:29 are
# BASELINE, not signal. A cutover you assumed is a result you invented.
#
# Rows before CUT are the baseline; after, the test. Small n for a while — do not
# over-read one round.
set -u
CUT="${CUT:-2026-07-27 13:44:56+00}"
psql () { kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
            psql -U clients_user -d clients_db "$@"; }

echo "── council adoption report ── cutover: $CUT"
echo

echo "1. Volume either side of the cutover (is there enough to read yet?)"
psql -c "
WITH r AS (SELECT created_at, COALESCE(source_agent,'generic') AS agent,
                  jsonb_array_elements(body::jsonb->'reviews') AS rv
           FROM diagnosis_artifacts WHERE kind='council_report')
SELECT agent,
       count(*) FILTER (WHERE created_at <  '$CUT') AS before_reviews,
       count(*) FILTER (WHERE created_at >= '$CUT') AS after_reviews
FROM r GROUP BY agent ORDER BY 3 DESC, 2 DESC;"

echo "2. THE headline: does the guardian cite PRIOR DEFLECTIONS when it invokes"
echo "   the stability preference? (before this change it had no way to know)"
echo "   FIXED 2026-07-27 late: 'invoked' and 'cited' were two INDEPENDENT filters"
echo "   over all guardian reviews, so the pair was never a subset — yet it was"
echo "   quoted as '6 of 90' in the handoff, memory and the owner's decision doc."
echo "   The real intersection was 2 of 90 (2.2%), not 6 (6.7%): 4 of the 6 cited"
echo "   precedent WITHOUT invoking the preference. both_invoked_and_cited is the"
echo "   headline; the other two columns are kept only to show the gap."
psql -c "
WITH r AS (SELECT created_at, jsonb_array_elements(body::jsonb->'reviews') AS rv
           FROM diagnosis_artifacts WHERE kind='council_report'),
g AS (SELECT created_at,
             COALESCE(rv->>'notes','') || ' ' ||
             COALESCE((SELECT string_agg(o->>'problem',' ') FROM jsonb_array_elements(rv->'objections') o),'') AS txt
      FROM r WHERE rv->>'reviewer'='guardian'),
f AS (SELECT created_at,
        (txt ILIKE '%higher%layer%' OR txt ILIKE '%battle-tested%'
         OR txt ILIKE '%foundational%') AS stab,
        (txt ILIKE '%deflect%' OR txt ILIKE '%prior council%'
         OR txt ILIKE '%council_report%'
         OR txt ILIKE '%previous submission%'
         OR txt ~* 'sent upward|already been sent|has been before|repeatedly sent') AS prec
      FROM g)
SELECT CASE WHEN created_at < '$CUT' THEN 'before' ELSE 'after' END AS era,
       count(*)                                     AS guardian_reviews,
       count(*) FILTER (WHERE stab)                 AS invoked_stability,
       count(*) FILTER (WHERE stab AND prec)        AS both_invoked_and_cited,
       round(100.0*count(*) FILTER (WHERE stab AND prec)
             / NULLIF(count(*) FILTER (WHERE stab),0), 1) AS pct_of_invoked,
       count(*) FILTER (WHERE prec AND NOT stab)    AS cited_but_did_not_invoke
FROM f GROUP BY 1 ORDER BY 1 DESC;"
echo "   CAVEAT, both directions. OVER-counts: '%deflect%' matches the word alone,"
echo "   and the NEW prompt itself says 'deflected upward', so a seat echoing its"
echo "   own instructions scores a citation. UNDER-counts: the 14:18 guardian"
echo "   reasoned correctly about recurrence WITHOUT quoting a past report and"
echo "   scored zero. Read the review text before trusting either number at low n."

echo "3. Do the historians cite the new case index (a title, a bug slug, a wrong call)?"
psql -c "
WITH r AS (SELECT created_at, jsonb_array_elements(body::jsonb->'reviews') AS rv
           FROM diagnosis_artifacts WHERE kind='council_report'),
h AS (SELECT created_at, rv->>'reviewer' AS seat,
             COALESCE(rv->>'notes','') || ' ' ||
             COALESCE((SELECT string_agg(o->>'problem',' ') FROM jsonb_array_elements(rv->'objections') o),'') AS txt
      FROM r WHERE rv->>'reviewer' IN ('bug_historian','debug_historian'))
SELECT seat, CASE WHEN created_at < '$CUT' THEN 'before' ELSE 'after' END AS era,
       count(*) AS reviews,
       count(*) FILTER (WHERE txt ~* 'bugs_(open|closed)|016b|WRONG_CALLS|§9|HANDOFF_') AS cited_a_source
FROM h GROUP BY 1,2 ORDER BY 1, 2 DESC;"

echo "4. The new seat: has review_architecture run, and what did it say?"
echo "   (signal lives in the FIRST LINE of notes — the only field that persists)"
psql -c "
WITH r AS (SELECT created_at, jsonb_array_elements(body::jsonb->'reviews') AS rv
           FROM diagnosis_artifacts WHERE kind='council_report')
SELECT created_at::timestamp(0) AS ts, rv->>'verdict' AS verdict,
       substring(COALESCE(rv->>'notes','') from 'ARCHITECTURE_SIGNAL:[^\n]*') AS signal_line,
       jsonb_array_length(COALESCE(rv->'objections','[]'::jsonb)) AS objections
FROM r WHERE rv->>'reviewer'='architecture' ORDER BY created_at DESC LIMIT 15;"

echo "5. Is the new seat producing NOISE? (a seat that objects to everything, or"
echo "   emits no signal line, is not earning its place — pull it if so.)"
echo "   NB: match ARCHITECTURE_SIGNAL case-SENSITIVELY. The pre-fix prompt carried a"
echo "   lowercase 'architecture_signal' JSON field, so an ILIKE here returns TRUE"
echo "   for the very version being replaced — a vacuous check that cost me a wrong"
echo "   'already patched' reading on 2026-07-27."
psql -c "
WITH r AS (SELECT jsonb_array_elements(body::jsonb->'reviews') AS rv
           FROM diagnosis_artifacts WHERE kind='council_report')
SELECT count(*) AS runs,
       count(*) FILTER (WHERE rv->>'verdict'='object')                          AS objected,
       count(*) FILTER (WHERE COALESCE(rv->>'notes','') !~ 'ARCHITECTURE_SIGNAL') AS missing_signal_line,
       count(*) FILTER (WHERE COALESCE(rv->>'notes','') ~ 'DEFLECTIONS: [0-9]')   AS gave_a_deflection_count,
       count(*) FILTER (WHERE rv->>'degraded' = 'true')                          AS truncated
FROM r WHERE rv->>'reviewer'='architecture';"

echo
echo "Reading it: (2) is the one that matters — stability objections that cite"
echo "precedent. (5) is the kill switch: high object-rate with no signal line means"
echo "confident noise. Remember (2)-(3) are TEXT matches on what a seat chose to"
echo "mention; silence is not proof it did not look."
