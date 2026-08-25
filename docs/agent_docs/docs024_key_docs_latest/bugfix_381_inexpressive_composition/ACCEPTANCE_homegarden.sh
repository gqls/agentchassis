#!/usr/bin/env bash
# ACCEPTANCE_homegarden.sh — the four reads that decide bugs_open/381, for the validation build.
#
# WHY IT EXISTS AS A FILE. Written and DRY-RUN 2026-08-25 10:47Z, while the build was still at
# hop two, precisely so it is not being debugged at the moment it matters. Q4 reads a table with a
# ~25h retention (MEASURED 2026-08-25: oldest row 1d00:55, 0 rows beyond 48h) — discovering a SQL
# error then would cost the only chance to see what the planner was told.
#
# EVERY QUERY REPORTS "not yet" RATHER THAN AN EMPTY RESULT. That is deliberate and is the whole
# design: on this lane an ambiguous zero has been misread three times in two days. "NONE PLACED
# YET" and "planner has not run yet" are different claims from "0", and the difference is the
# result.
#
# Usage:  bash ACCEPTANCE_homegarden.sh            (defaults to the homegarden.uk build)
#         SITE=<uuid> bash ACCEPTANCE_homegarden.sh
set -o pipefail
SITE="${SITE:-5904bd0f-33fd-4212-9c1b-50b28fe72fdb}"   # homegarden.uk, built 2026-08-25 10:21:49Z
SINCE="${SINCE:-2026-08-25 10:00:00+00}"
echo "### bugs_open/381 acceptance — site=$SITE   run at $(date -u '+%Y-%m-%d %H:%M:%SZ')"
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<SQL
\pset format unaligned
\pset tuples_only on
SET statement_timeout='90s';
\echo ''
\echo '=== Q1 HEADLINE — did the planner choose one of the three new components? ==='
SELECT COALESCE((SELECT string_agg(f||' x'||n::text,', ') FROM (
  SELECT cc.function AS f, count(*) AS n FROM page_components pc
    JOIN content_components cc ON cc.id=pc.component_id JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='$SITE' AND cc.function IN ('checklist','period-calendar','comparison-table')
   GROUP BY cc.function) z), 'NONE PLACED YET');
\echo ''
\echo '=== Q2 the full composition the planner chose, page by page ==='
SELECT COALESCE((SELECT string_agg(line,E'\n') FROM (
  SELECT p.name||': '||string_agg(cc.function,' > ' ORDER BY pc.position) AS line
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
    JOIN content_components cc ON cc.id=pc.component_id
   WHERE p.site_id='$SITE' GROUP BY p.name) y), '(no pages yet)');
\echo ''
\echo '=== Q3 structure — compare against garden-tools baseline: 0 tables / 0 content lists / 0 strong ==='
SELECT 'pages='||count(DISTINCT p.id)::text||' sections='||count(pc.id)::text
  ||' with_list='||count(*) FILTER (WHERE pc.rendered_html ~* '<(ul|ol)[\s>]')::text
  ||' with_table='||count(*) FILTER (WHERE pc.rendered_html ~* '<table[\s>]')::text
  ||' with_h3='||count(*) FILTER (WHERE pc.rendered_html ~* '<h3[\s>]')::text
  ||' with_strong='||count(*) FILTER (WHERE pc.rendered_html ~* '<strong[\s>]')::text
FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id WHERE p.site_id='$SITE';
\echo ''
\echo '=== Q4 ⚠ EXPIRES ~25h AFTER THE PLANNER RUNS — was it actually TOLD the capability? ==='
SELECT COALESCE((SELECT 'expresses_token='||(prompt_rendered LIKE '%[expresses:%')::text
   ||' prose_only_token='||(prompt_rendered LIKE '%[prose only]%')::text
   ||' rule19='||(prompt_rendered LIKE '%MATCH STRUCTURE TO PROMISE%')::text||' | at '||created_at::text
  FROM llm_call_log WHERE agent_type='build-site-planner' AND created_at > '$SINCE'
  ORDER BY created_at DESC LIMIT 1), 'planner has not run yet');
SQL
echo ""
echo "How to read it:"
echo "  Q1 non-zero            -> the lane's open item CLOSES."
echo "  Q1 none + Q4 all true  -> the planner WAS told and declined. The interesting negative;"
echo "                            read Q2 to see what it chose instead. NOT a wiring failure."
echo "  Q1 none + Q4 false     -> the capability never reached the prompt. That IS a wiring failure."
echo "  Q4 'has not run yet'   -> nothing is decided; do not read Q1/Q3 as a result."
echo "  ⚠ pages that never build are bugs_open/206, not this fix (garden-tools lost 5 of 12)."
echo ""
cat <<'WARN'

⚠⚠⚠ THE MISATTRIBUTION THAT WOULD ACTUALLY COST US — READ BEFORE JUDGING A THIN SITE.
   [MEASURED 2026-08-25 11:34Z, mint snapshot in evidence/] 17 of this build's 21 pages are
   page_type=section-index, filed by reconcile_site_plan at handler_agent=page-build-handler.
   THAT IS THE EXACT ROLE+HANDLER PAIRING THAT NO-OPPED ON garden-tools.uk.
   page-build-handler has NO ensure_page_section_layout step — that exists only in
   directory-build-handler's workflow — and the section-index routing fix (efec862f4,
   09:58:33Z) is NOT in the running pods (v1.0.1337, started 09:27:48Z, 31 min earlier).
   ⚠ SO IF THIS SITE SERVES ~4 OF 21 PAGES WITH THE MONTH PAGES MISSING, IT WILL READ AS
     "the planner promised a month-by-month structure and did not deliver it" — which is
     THIS BUG'S EXACT SYMPTOM SENTENCE, AND IT WOULD NOT BE THIS BUG.
     It is bugs_open/206 residual class (b) at 17x scale. The tell: check handler_agent and
     the error string BEFORE concluding anything about composition. The string
     "page-build-handler no-op: no sections ready to build" is mis-ROUTING, not thin content.
   Flagged independently by BOTH the loanzy_uk_example_site and bugfix_206 lanes.

⚠⚠ THE INTERACTION NEITHER LANE PREDICTED: the planner expressed "month by month" as
   SEVENTEEN section-index pages rather than as one page carrying period-calendar. It placed
   period-calendar ONCE, on the landing page. So if those 17 no-op, the site ends with the
   component placed but the calendar STRUCTURE gone — a structural promise satisfied at site
   level, routing straight into the one page role with no builder. Whether that is a 381
   concern, a 206 concern or a third thing is an open question, recorded not answered.

WARN
echo "⚠⚠ COMPARISON-TABLE IS UNEXERCISED ON THIS BUILD, NOT FAILED, IF IT DOES NOT APPEAR."
echo "   [MEASURED 2026-08-25 10:53-10:54Z by the loanzy_uk_example_site lane, orchestration"
echo "   5937f08b-63ad-4de2-a5ea-97b17cacbb04] the vertical landscape was synthesised from TWO"
echo "   sources, not three:"
echo "     rhs.org.uk          success  6 sources  quality good"
echo "     gardenersworld.com  success  6 sources  quality good"
echo "     which.co.uk         success  0 SOURCES  quality NONE  (\"Crawl completed but no"
echo "                                                usable page content was found.\")"
echo "   Both contributing publishers are CALENDAR/TASK-structured how-to sites. The one"
echo "   comparison publisher in the draw contributed nothing. So the landscape is biased"
echo "   TOWARD period-calendar and checklist and AWAY from comparison, for reasons that have"
echo "   nothing to do with whether comparison-table works."
echo "   PRIMARY SIGNAL on this build: period-calendar and checklist placement."
echo "   comparison-table absent => report as UNEXERCISED. Do NOT record it as a negative;"
echo "   it needs a build whose vertical actually is comparison-shaped."
