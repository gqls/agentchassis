-- FILE: SQL_2026-08-16_fundamentallyai_tools_page_type.sql
--
-- The tools INDEX page is mis-typed as `tool`, which bars it from the primary nav
-- no matter what its in_header flag says. Retype it to `content`, matching four
-- other sites, then rebuild the nav.
--
-- THE MECHANISM, read at source (populate_nav_tables_action.go:~330):
--   neverPrimaryTypes := map[string]bool{"blog-post": true, "tool": true, "entity-page": true}
--   neverPrimary := neverPrimaryTypes[page.PageType] || (isChildPageURL(page.URL) && !isSectionIndexType(page.PageType))
--   if neverPrimary { if InHeader || InFooter { utility = append(...) } ... }
-- The comment above it is explicit: these types are barred "regardless of
-- in_header flag". So setting in_header=true was necessary and NOT sufficient —
-- the page went to the `utility` group, i.e. the FOOTER, which is exactly where
-- it appeared: two /tools.html links, both in the footer, none in <nav>.
--
-- WHY THIS IS A MIS-TYPE AND NOT A RULE TO FIGHT. The rule is right: individual
-- tool pages belong under a parent listing, not in the top menu. /tools.html IS
-- that parent listing. It is typed as one of its own children. Note the URL rule
-- already gets this right — isChildPageURL matches "/tools/" and /tools.html does
-- not match — so only the page_type clause misfires.
--
-- FLEET PRECEDENT, measured not assumed: five sites have a `tools` page.
--   ai-agent-orchestration.com  content
--   finetuning.uk               content
--   idea.uk                     content
--   robot-hands.com             content
--   fundamentallyai.com         tool     <- the only one
-- POSITIVE CONTROL at the served artefact — all three checked show Tools in the
-- top nav:
--   idea.uk          "Tools About Contact Guides News"
--   robot-hands.com  "Tools About News MatchMatrix Selection Guide Learning Center"
--   finetuning.uk    "Home Services About Tools Contact Use Cases ..."
--
-- WHY `content` AND NOT `section-index` (which isSectionIndexType would also
-- allow): section-index would work, but four sites already use `content` for this
-- exact page and none uses section-index. Matching the fleet beats inventing a
-- second convention for one site.
--
-- SIDE EFFECTS CHECKED, not assumed:
--   - rebuild_policy is 'generic' here AND on the precedent sites, so the build
--     path is governed by that column, not by page_type — unchanged.
--   - site_specs aspect='tools' has 0 rows mentioning /tools.html.
--   - evidence_base F14's source SQL counts page_type='tool' AND name <> 'tools',
--     so it already excluded this page BY NAME: the count stays 5 either way.
--     Re-asserted below, because a fact that silently changes value is the worst
--     outcome of a retype.

BEGIN;

DO $chk$
DECLARE pt text; h bool; rp text; n int;
BEGIN
  SELECT page_type, in_header, rebuild_policy INTO pt, h, rp FROM pages
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools' AND status='active';
  IF NOT FOUND THEN RAISE EXCEPTION 'no active tools page'; END IF;
  IF pt <> 'tool' THEN RAISE EXCEPTION 'page_type is already %, not tool — re-read before changing', pt; END IF;
  IF NOT h THEN RAISE EXCEPTION 'in_header is false — set the declaration first, retyping alone will not place it'; END IF;
  IF rp <> 'generic' THEN RAISE EXCEPTION 'rebuild_policy is %, not generic — retyping may change the build path', rp; END IF;

  -- the fleet convention must still be what this change is matching
  SELECT count(*) INTO n FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE p.name='tools' AND p.status='active' AND p.page_type='content';
  IF n < 3 THEN RAISE EXCEPTION 'only % sites use content for a tools page — the precedent this relies on has moved', n; END IF;
END $chk$;

UPDATE pages SET page_type='content'
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools' AND status='active';

-- rebuild the nav with the correct handler (nav-updater runs populate_nav_tables,
-- the ONLY writer of site_nav_items, then chains to render_site_components)
INSERT INTO site_work_items
  (site_id, page_id, item_type, item_key, summary, spec, handler_agent, pipeline,
   priority, severity, approval_mode, max_attempts, source, created_by, status)
SELECT
  '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd', p.id, 'nav_drift',
  'nav_rebuild:199733a8-ac9c-4c30-b2ce-65ecdac6f3bd:tools-20260816-retyped',
  'Nav membership for tools — retyped from tool to content so it is primary-eligible; rebuild nav and chrome',
  jsonb_build_object(
    'reason','nav_membership_declared',
    'fix','A page has nav membership declared (in_header/in_footer). Rebuild site_nav_items from pages and re-render chrome so the link ships.',
    'page_id', p.id,
    'page_url', p.url,
    'page_name','tools',
    'in_header', true,
    'in_footer', false,
    'requested_by','brochure_contrast_front_thread',
    'context','Third attempt, and the first two are why this one names its cause. (1) in_header=true alone put the link in the FOOTER: page_type=tool is in neverPrimaryTypes, which the source says applies "regardless of in_header flag". (2) A rebuild filed against page-build-handler completed green having re-rendered one file and never touched site_nav_items. Now retyped to content, matching four other sites whose Tools entry is verified live in their top nav.'
  ),
  'nav-updater','build',30,'medium','auto',3,
  'operator:brochure_component_library','brochure_contrast_front_thread','triaged'
FROM pages p
WHERE p.site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND p.name='tools'
ON CONFLICT DO NOTHING;

DO $post$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools'
     AND page_type='content' AND in_header;
  IF n <> 1 THEN RAISE EXCEPTION 'post: retype did not take'; END IF;

  -- F14 must still be TRUE after the retype, or the honesty register now lies
  SELECT count(*) INTO n FROM pages
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND status='active'
     AND page_type='tool' AND name <> 'tools';
  IF n <> 5 THEN RAISE EXCEPTION 'post: F14-interactive-tools now reads % but is registered as 5 — fix the fact before shipping', n; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by='brochure_contrast_front_thread' AND item_type='nav_drift'
     AND status='triaged' AND handler_agent='nav-updater';
  IF n <> 1 THEN RAISE EXCEPTION 'post: expected exactly 1 open nav_drift on nav-updater, found %', n; END IF;
END $post$;

SELECT name, page_type, in_header, nav_order FROM pages
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools';

COMMIT;

-- VERIFY AT THE SERVED <nav>, NOT at a link count anywhere on the page. The last
-- attempt "passed" a grep for href="/tools.html" while both hits were in the FOOTER:
--   curl -s https://fundamentallyai.com/index.html \
--     | python3 -c "import re,sys;h=sys.stdin.read();m=re.search(r'(?is)<nav[^>]*>.*?</nav>',h);print(re.sub(r'\s+',' ',re.sub(r'<[^>]+>',' ',m.group(0))))"
-- Expect seven items including Tools. Check a second page too — chrome is shared,
-- so one page proves the render, not the deploy.
