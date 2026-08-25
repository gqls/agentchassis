#!/bin/bash
# capture_reconcile_mint.sh <domain>
#
# The bugs_open/206 closure assertion, captured at the artefact. Owner's retraction of 2026-08-25
# moved that proof onto "the next greenfield build", so this exists to be pointed at one.
#
# THE ASSERTION: a page whose role is entity-directory must be minted with
# handler_agent='directory-build-handler', by created_by='reconcile_site_plan', with nobody having
# set the handler by hand. A hand re-triage fixes the page and proves nothing.
#
# ⚠ THREE TRAPS BUILT IN, each one measured by the 206 lane or this one:
#  1. UNION site_work_items WITH site_work_items_archive. Closing a row ARCHIVES it out of the live
#     table, so a census run after the build has moved on reads a confident ZERO for rows that
#     existed. Never query the live table alone for a population you did not watch appear.
#  2. Discriminate on spec->>'page_role'. NOT page_type / page_id / filename / domain — measured
#     absent on 134 of 134 reconcile-minted rows fleet-wide, so those filters return a confident
#     zero for the very population they exist to count.
#  3. pages is the AUTHORITY for what a page IS; join it on (site_id, page_name). The item's own
#     spec says what reconcile THOUGHT it was filing.
#  4. ⚠ FOUND BY VALIDATING THIS SCRIPT, 2026-08-25 — needs_page HAS MANY PRODUCERS AND ONLY ONE
#     CARRIES page_role. Fleet-wide `[MEASURED 2026-08-25 11:06Z, site_work_items UNION archive]`:
#     1,438 needs_page rows, **46 distinct created_by values**, and only **451 (31.4%)** carry a
#     page_role key at all — 387 of those are reconcile_site_plan (which carries it on 387 of 387).
#     The other big automated producers carry NONE: page-rerender 414 rows / 0, image-build-handler
#     262 / 0, render_directory 103 / 0, render_news_section 80 / 0. `page_type` is on 0 of 1,438.
#     CONSEQUENCE: filtering on page_role alone silently narrows a needs_page census to ~a third of
#     it; NOT filtering on created_by mixes five automated producers with different spec shapes. This
#     script prints every row and states WHY each is n/a rather than dropping it — which is how the
#     second producer on garden-tools (six image-build-handler rows) was noticed at all.
set -o pipefail
DOM="${1:?usage: capture_reconcile_mint.sh <domain>}"
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
echo "### reconcile-mint capture for $DOM at $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
$PSQL -c "
WITH items AS (
  SELECT site_id, item_type, item_key, status, handler_agent, created_by, created_at, spec, error
    FROM site_work_items
  UNION ALL
  SELECT site_id, item_type, item_key, status, handler_agent, created_by, created_at, spec, error
    FROM site_work_items_archive
)
SELECT i.created_at::time(0) AS minted,
       i.spec->>'page_name'         AS page_name,
       i.spec->>'page_role'         AS role_in_spec,
       p.page_type                  AS role_in_pages,
       COALESCE(NULLIF(i.handler_agent,''),'(none)') AS handler_agent,
       COALESCE(jsonb_array_length(p.sections),-1) AS sections_len,
       i.created_by,
       i.status,
       CASE
         WHEN i.spec->>'page_role' IS NULL                     THEN 'n/a  (no page_role on this row)'
         WHEN i.spec->>'page_role' <> 'entity-directory'        THEN 'n/a  (not the asserted role)'
         WHEN i.handler_agent = 'directory-build-handler'       THEN '*** PASS — routed to directory-build-handler ***'
         ELSE '*** FAIL — entity-directory at ' || COALESCE(NULLIF(i.handler_agent,''),'(none)') || ' ***'
       END AS verdict_206,
       left(COALESCE(i.error,''),60) AS err
FROM items i
JOIN sites s ON s.id = i.site_id
LEFT JOIN pages p ON p.site_id = i.site_id AND p.name = i.spec->>'page_name'
WHERE s.domain = '$DOM' AND i.item_type = 'needs_page'
ORDER BY i.created_at, i.spec->>'page_name';"
n=$($PSQL -tA -c "
WITH items AS (SELECT site_id,item_type FROM site_work_items UNION ALL SELECT site_id,item_type FROM site_work_items_archive)
SELECT count(*) FROM items i JOIN sites s ON s.id=i.site_id WHERE s.domain='$DOM' AND i.item_type='needs_page';" 2>&1)
echo ">> needs_page rows found (live UNION archive): $n"
[ "$n" = "0" ] && echo "!! ZERO ROWS. That is UNKNOWN, not a pass — reconcile may not have run yet. Do NOT read it as 'nothing to route'."
echo
echo "### CLASS (b) RISK SET — pages the planner left with NO sections. THIS is the predictor."
$PSQL -c "
SELECT p.name, p.page_type, jsonb_array_length(p.sections) AS sections_len, p.build_status,
       COALESCE(p.deployed_at::text,'(never stamped)') AS deployed_at
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='$DOM' AND COALESCE(jsonb_array_length(p.sections),0)=0
ORDER BY p.name;"
echo ">> Any row above is a candidate for bugs_open/206 residual class (b): page-build-handler cannot"
echo "   FILL a missing layout, so it no-ops with 'no sections ready to build'. Empty result = no"
echo "   class (b) exposure on this site."
echo ">> ⚠ DEMONSTRATED, not proposed: on homegarden.uk 2026-08-25 this predicate named ONE page of 21"
echo "   (blog-post, sections_len 0) and that page — and only that page — no-opped at 11:42:10Z, while"
echo "   17 section-index pages sharing the 'failing' role built fine. The ROLE-based version of this"
echo "   prediction was wrong by 17x; see WRONG_CALLS.md 2026-08-25."
echo ">> ⚠ DO NOT pair this with 'deployed_at IS NULL' as the outcome measure. garden-tools' contact"
echo "   page HAS sections, serves 57KB, and carries deployed_at = NULL at build_status='needs_rebuild'."
echo "   A page can serve while its own deploy is unstamped. Verify the outcome by fetching the URL."
echo
echo ">> Reminder: 'no sections ready to build' on a later failure is AMBIGUOUS — (a) wrong/missing"
echo "   builder, or (b) a builder that cannot fill a missing layout (ensure_page_section_layout"
echo "   exists only in directory-build-handler's workflow). The string cannot tell you which."
