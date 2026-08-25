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

⚠⚠⚠ RETRACTED — AN EARLIER VERSION OF THIS GUIDE WARNED THAT 17 section-index PAGES WOULD
   NO-OP AND THAT YOU SHOULD "CHECK handler_agent FIRST". BOTH LANES THAT GAVE ME THAT
   WARNING RETRACTED IT WITH MEASUREMENTS, AND LEAVING IT IN WOULD HAVE BEEN WORSE THAN
   HAVING NO WARNING: it points away from this bug and hands you a ready-made excuse for a
   finding that is actually ours. handler_agent=page-build-handler on these pages is CORRECT.

   [MEASURED 2026-08-25 11:40Z, verified independently here] 206's no-op needs a page with NO
   layout from any source (it dies at ready_count == 0). It builds a page that HAS sections
   perfectly well. The real predictor is an EMPTY pages.sections, not page_type:
     homegarden.uk  18 pages at 3 sections, 1 at 4, 1 at 5  -> fine;  ONE at 0 (blog-post)
     garden-tools   sections=0 -> 5 pages, 0 ever built;  every page WITH sections built
                    (its one non-deployed page carrying sections is `contact`, status
                     needs_rebuild — a rebuild state, not a build failure)
   So exposure here is 1 page in 21, not 17. april-index — the very pairing I was told would
   fail — is build_status=deployed at /april/index.html.

   THE ONE-QUERY RISK SET, use this rather than a role list:
     SELECT name, page_type, jsonb_array_length(sections) AS sections_len
     FROM pages WHERE site_id='<site>' ORDER BY sections_len, name;   -- anything at 0

   KEEP AS A DISCRIMINATOR, unchanged and still right: if you ever DO see
   "no sections ready to build", that is mis-ROUTING and not thin content. Just do not
   expect it here.

⚠⚠ RETRACTED — I CLAIMED "SEVENTEEN THIN INDEX PAGES" AND MY OWN DATA REFUTES IT.
   I wrote that the planner had "dissolved the promise across seventeen pages so no single
   page has to express it". That is FALSE, and the disconfirming evidence was in my hands
   before I wrote it (april-index already showed [LIST][H3]).

   [MEASURED 2026-08-25 11:45Z] the month pages are NOT thin:
     april-index       generic-text-block  2,822 chars  4 <li>  4 <h3>
     august-index      generic-text-block  2,994 chars  3 <li>  4 <h3>
     comparisons-index generic-text-block  2,149 chars  4 <li>  3 <h3>
     4 of 4 bodies DISTINCT — not boilerplate. August is genuinely about August:
     "August is usually the driest month of the growing season, so the jobs that matter
      most this month are about protecting what you already have rather than starting
      anything new."

   And n_sections=3 is a PLANNER CHOICE, not a default — the discriminator the
   loanzy lane asked for: other sites' section-index pages run 1-2 sections
   (["hero","blog-listing"], ["hero","guide-list"], ["hero","category-listing"]).
   homegarden is the ONLY site with ["hero","generic-text-block","content-listing"], and
   the extra member is the PROSE BLOCK. So the planner chose to give every month page a
   prose block, and the writer filled each with a month-specific structured piece.

   THE HONEST READING: the planner answered "month by month" with a per-month PAGE
   architecture rather than one page carrying period-calendar, and each page keeps its own
   promise with real structure. That is the fix WORKING — arguably better than one calendar
   page — not a promise being dodged. What it does raise is a genuine question about
   period-calendar's NECESSITY for this shape, which is a finding about the component, not
   about the planner.

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
