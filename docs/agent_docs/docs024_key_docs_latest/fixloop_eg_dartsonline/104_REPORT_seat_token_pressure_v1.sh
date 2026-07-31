#!/usr/bin/env bash
# 104_REPORT_seat_token_pressure_v1.sh — ADVISORY. Which council seats are close to
# their token cap, and which rounds gated only because a seat was cut off.
#
# WHY THIS EXISTS (bugs_open/138 fix candidate 2, 2026-07-30).
# 138's mechanism: a review TRUNCATED at max_tokens is recovered, marked `degraded`,
# and a degraded `object` gates the round to REVISE regardless of the severities
# that survived. So an over-long ADVISORY review silently becomes a BLOCKING one —
# and because a high object-rate is also the documented kill-switch for retiring a
# noisy seat, **a seat can be pulled for being noisy when it was being cut off.**
# Candidate 1 (live 2026-07-29, FIX-055) made the CAUSE legible in `decided_by` and
# in `metadata.gated_by_truncation`. Candidate 2 is this: nothing measures the RATE.
#
# THE LEADING INDICATOR IS THE POINT, NOT THE LAGGING ONE.
# Counting truncations would report ~0 today and never fire, because candidate 3
# raised the cap on every seat that had actually truncated. That does not mean the
# door is shut — a cap raise MOVES the cliff, it does not remove it (proved on this
# very seat: `review_architecture` was raised to 16000 and a longer prompt
# reintroduced truncation against the new cap within hours). So the headline here is
# HEADROOM — output_tokens as a fraction of the seat's CURRENT cap, reported as both
# the peak (the near-miss) and the p95 (the routine pressure). A seat whose longest
# review landed at 99% of its cap has not truncated yet and is one sentence away.
#
# WHAT IT MEASURES AND CANNOT MEASURE — read this before quoting a number.
#
#  1. ONLY CALLS AT THE SEAT'S CURRENT LIVE CAP COUNT. `llm_call_log.max_tokens` is
#     recorded per call, and caps have been raised mid-window (editquality 8000 ->
#     16000 on 07-28, guidelines/prior_art on 07-29). Mixing them produces a p95 of
#     a ratio whose denominator changed, which reads as alarming nonsense: a naive
#     query showed council-gate/review_editquality at "95% of a 16000 cap" when the
#     16000 rows peak at 63% and the 95% belonged to the retired 8000 rows. Rows at
#     a superseded cap say nothing about present risk and are excluded.
#
#  2. THE COUNCIL CANNOT BE ATTRIBUTED for the two fix-lane councils. Every review
#     call before 2026-07-26 14:54 logged `agent_type='generic'`; from 15:03 the same
#     calls log `council-gate`; `fix-proposer` has NEVER appeared. So a call cannot
#     be traced to fix-proposer vs council-gate, and any figure split by agent_type
#     silently truncates its own history at that relabelling. This report therefore
#     keys on (SEAT, CAP) — the unit the risk actually belongs to — and names the
#     councils that currently hold that pair. It does NOT claim the calls came from
#     them. Where a seat name is shared by councils with DIFFERENT caps, the
#     populations are genuinely separate and appear as separate rows.
#
#  3. `checks` and `code_checks` ARE NOT PERSISTED (0 of 3,106 stored reviews carry
#     either key). They sit near the tail of most templates, so truncation destroys
#     them too — invisibly, and this report cannot see it. A seat that asked no
#     questions and a seat whose questions were cut off look identical.
#
#  4. A NULL `gated_by_truncation` MEANS "WRITTEN BY A PRE-FIX BINARY", NOT "CLEAN".
#     Section 3 reports the two populations separately and never sums them. The
#     field is emitted unconditionally by the live code precisely so absence and
#     false stay distinguishable.
#
# TWO THRESHOLDS, BOTH MEASURED, BECAUSE THEY MEAN DIFFERENT THINGS.
#   * NEAR-MISS — peak >= 95% of cap. Truncation IS a tail event, so the maximum is
#     the primary indicator: one review that came within 5% of the cap means the
#     next slightly longer one is cut. Anchored on the live distribution — both
#     populations that have truncated peak at 100% by construction, and the next
#     peaks down are 99.2, 98.8, 96.6, 94.1. The 95 cut sits in that gap.
#   * PRESSURE — p95 >= 85% of cap. This is the BODY of the distribution near the
#     cap: not one long review but a seat that routinely writes to its ceiling.
#     The two truncating populations sit at p95 96.1 and 85.7; none below 85 has
#     ever truncated.
# Reporting them separately matters: the first pair this report flagged on p95
# alone had 4 attributable calls, while `review_guardian` — 118 attributable calls
# and a 99.2% peak — read "ok" under a p95-only rule. A single blended threshold
# hid the row with the real evidence behind two rows built on inference.
# MIN_N = 20: below that a p95 IS the top one or two observations, so those rows are
# listed as "watch" — a 100%-of-cap peak on n=4 is one verbose review, not a
# distribution.
#
# NOT A PARITY LINT, and deliberately not: 102_LINT_council_seat_parity.py compares
# each seat against its OWN council's family and explicitly declines to flag
# cross-council divergence as drift, because councils legitimately differ (different
# remits, different seat sets, an owner ruling that experience-planner omits
# tolerate_truncation). Section 4 does not call divergence a violation. It prints
# divergence NEXT TO the truncation evidence, so a reader can see which divergence
# is the one that matters — that is a different claim from "these should match".
#
# Usage:  ./104_REPORT_seat_token_pressure_v1.sh [days]        (default 14)
#         P95_FLAG=90 PEAK_FLAG=98 MIN_N=30 ./104_REPORT_seat_token_pressure_v1.sh 7
set -u

DAYS="${1:-14}"
P95_FLAG="${P95_FLAG:-85}"
PEAK_FLAG="${PEAK_FLAG:-95}"
MIN_N="${MIN_N:-20}"

psql () { kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
            psql -U clients_user -d clients_db "$@"; }

echo "══ council seat token pressure ══ window: last ${DAYS}d ── flag: peak >= ${PEAK_FLAG}% or p95 >= ${P95_FLAG}% of cap, over >= ${MIN_N} calls"
echo "   (bugs_open/138 candidate 2. Only calls AT the seat's current live cap are counted — see header note 1.)"
echo

echo "1. HEADROOM at the live cap — the leading indicator. Flagged rows first."
psql -c "
WITH live AS (
  SELECT a.type AS council, s.key AS seat,
         (s.value->'config'->'ai_service'->>'max_tokens')::int AS cap
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.key LIKE 'review_%'
    AND (s.value->'config'->'ai_service'->>'max_tokens') IS NOT NULL
), pairs AS (
  SELECT seat, cap, string_agg(DISTINCT council, ',') AS councils_holding
  FROM live GROUP BY 1,2
), calls AS (
  SELECT step_name AS seat, max_tokens AS cap, agent_type, created_at,
         output_tokens::numeric / max_tokens AS frac
  FROM llm_call_log
  WHERE created_at > now() - interval '${DAYS} days'
    AND step_name LIKE 'review_%' AND max_tokens > 0 AND output_tokens IS NOT NULL
), agg AS (
  SELECT p.seat, p.cap, p.councils_holding, count(c.frac) AS n,
         round(100*(percentile_cont(0.95) WITHIN GROUP (ORDER BY c.frac))::numeric,1) AS p95_pct,
         round(100*max(c.frac),1) AS peak_pct,
         count(*) FILTER (WHERE c.frac >= 1) AS truncated,
         to_char(min(c.created_at),'MM-DD') || '..' || to_char(max(c.created_at),'MM-DD') AS span,
         count(c.frac) FILTER (WHERE c.agent_type = ANY(string_to_array(p.councils_holding,','))) AS n_holder,
         string_agg(DISTINCT c.agent_type, ',') AS logged_as
  FROM pairs p LEFT JOIN calls c ON c.seat = p.seat AND c.cap = p.cap
  GROUP BY 1,2,3
)
SELECT CASE WHEN n = 0 THEN 'no calls'
            WHEN truncated > 0 THEN 'FLAG truncated'
            WHEN n < ${MIN_N} THEN 'watch n<${MIN_N}'
            WHEN peak_pct >= ${PEAK_FLAG} THEN 'FLAG near-miss'
            WHEN p95_pct >= ${P95_FLAG} THEN 'FLAG pressure'
            ELSE 'ok' END AS status,
       seat, cap, n, n_holder, p95_pct, peak_pct, truncated, span, logged_as
FROM agg
ORDER BY (truncated > 0) DESC,
         (n >= ${MIN_N} AND (peak_pct >= ${PEAK_FLAG} OR p95_pct >= ${P95_FLAG})) DESC,
         peak_pct DESC NULLS LAST, seat;"

echo "   n is every call at that (seat, cap) from ANY council — header note 2 says why"
echo "   the council cannot be recovered. n_holder counts only the calls LABELLED with"
echo "   a council that still holds the pair: exact for feature-designer and the"
echo "   experience councils, a LOWER BOUND for the fix lane, because 'generic' rows"
echo "   (everything before 07-26 14:54) cannot be assigned and are excluded from it."
echo "   Read n_holder against n before acting: review_editquality@8000 flags on 227"
echo "   calls of which only 4 belong to feature-designer, the one council still at"
echo "   that cap — the other 223 are the fix lane's own history BEFORE its raise to"
echo "   16000. Same seat, same cap, largely shared prompt, so the transfer is"
echo "   plausible; it is still an INFERENCE about feature-designer, not a measurement"
echo "   of it. A flag with n_holder small is a reason to look, not a finding."

echo
echo "2. What a flagged row costs, and the two ways out."
echo "   Raising the cap moves the cliff; shortening the output removes the pressure."
echo "   Measured on review_architecture (the only seat that has had both applied):"
echo "   8000->16000 AND a length budget together took peak output to 4,443 tokens —"
echo "   28% of the new cap — because the outputs got SHORTER, not just the ceiling"
echo "   higher. Neither half alone was tested in isolation, so this cannot say which"
echo "   did more; it CAN say the pair works and that a cap raise alone is a bet."
echo

echo "3. ROUND-level truncation gates. Two populations, never summed (header note 4)."
psql -c "
WITH rep AS (
  SELECT id, created_at, body::jsonb AS b, metadata->>'gated_by_truncation' AS flag
  FROM diagnosis_artifacts
  WHERE kind='council_report' AND created_at > now() - interval '${DAYS} days'
), rv AS (
  SELECT rep.id, rep.created_at, rep.flag, rep.b->>'decision' AS decision,
         rep.b->>'decided_by' AS decided_by,
         COALESCE((r->>'degraded')::boolean,false) AS degraded,
         r->>'verdict' AS verdict,
         EXISTS (SELECT 1 FROM jsonb_array_elements(
                   CASE WHEN jsonb_typeof(r->'objections')='array'
                        THEN r->'objections' ELSE '[]'::jsonb END) o
                 WHERE lower(btrim(COALESCE(o->>'severity',''))) NOT IN ('low','medium')
                ) AS has_gating_obj,
         COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(r->'objections')='array'
                  THEN r->'objections' ELSE '[]'::jsonb END),0) AS n_obj
  FROM rep, LATERAL jsonb_array_elements(rep.b->'reviews') r
), g AS (
  SELECT *, (verdict='object' AND (degraded OR n_obj=0 OR has_gating_obj)) AS gates,
            (verdict='object' AND degraded AND NOT has_gating_obj) AS trunc_only
  FROM rv
), per AS (
  SELECT id, min(created_at) AS at, min(flag) AS flag, min(decision) AS decision,
         count(*) FILTER (WHERE gates AND NOT trunc_only) AS merits,
         count(*) FILTER (WHERE gates AND trunc_only) AS trunc
  FROM g GROUP BY id
)
SELECT CASE WHEN flag IS NULL THEN 'replayed (pre-fix binary wrote it)'
            ELSE 'reported by the live code' END AS source,
       count(*) AS rounds,
       count(*) FILTER (WHERE trunc > 0 AND merits = 0) AS gated_ONLY_by_truncation,
       count(*) FILTER (WHERE trunc > 0 AND merits > 0) AS mixed,
       count(*) FILTER (WHERE flag = 'true') AS flag_true
FROM per GROUP BY 1 ORDER BY 1;"

echo "   'replayed' re-derives the rule from the stored reviews[] — it is what the"
echo "   live code WOULD have said. 'reported' is what it DID say. When the second"
echo "   population is empty the fix has not rolled yet; that is a fact about the"
echo "   deploy, not about truncation."
echo

echo "4. Seats whose cap DIVERGES across councils (information, not a violation)."
psql -c "
WITH live AS (
  SELECT a.type AS council, s.key AS seat,
         (s.value->'config'->'ai_service'->>'max_tokens')::int AS cap
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.key LIKE 'review_%'
    AND (s.value->'config'->'ai_service'->>'max_tokens') IS NOT NULL
)
SELECT seat, string_agg(council || '=' || cap, ', ' ORDER BY cap DESC, council) AS caps
FROM live GROUP BY seat HAVING count(DISTINCT cap) > 1 ORDER BY seat;"

echo "   A divergence matters when the LOW side is a seat this report flagged, or"
echo "   when the high side was raised BECAUSE it truncated. 099_SYNC_gate_roster.py"
echo "   mirrors fix-proposer -> council-gate ONLY; the other four councils are"
echo "   synced by nothing, so a fix ruled on for one can sit unapplied on another."
echo

echo "── automatic half: scheduled task 'council-seat-token-pressure' (CTE-only,"
echo "   no LLM, no message fired) writes a doc_notes row when section 1 flags a"
echo "   pair, at most one per 24h. Read the notes it has written:"
echo "     SELECT created_at, body FROM doc_notes WHERE categories ? 'seat-token-pressure'"
echo "      ORDER BY created_at DESC LIMIT 5;"
echo "   The threshold lives in ONE place — that task's pre_query. This script"
echo "   reports numbers and does not re-encode it; read it there:"
echo "     SELECT pre_query FROM scheduled_tasks WHERE name='council-seat-token-pressure';"
