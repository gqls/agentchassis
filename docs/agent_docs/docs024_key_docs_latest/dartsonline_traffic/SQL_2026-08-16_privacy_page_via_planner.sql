-- SQL_2026-08-16_privacy_page_via_planner.sql
--
-- Ask the framework to create dartsonline.com's privacy page, via the route that
-- ACTUALLY EXISTS: content-gap-planner -> apply_gap_plan -> new_page, which creates
-- the pages row and then files needs_content_page for the copy.
--
-- Modelled directly on SQL_2026-07-29n_news_page.sql in this directory, which used the
-- same route on this same site and produced /news/index.html (serving 200 today).
--
-- ── WHY THE COPY WILL BE VERBATIM, AND WHERE THAT IS ENFORCED ───────────────
-- The owner approved the wording on 2026-08-15 and supplied the controller identity.
-- It is already registered: site_specs aspect 'evidence_base'
-- (12c022db-3338-40c4-b3a7-1542d413ad70, created 2026-08-16 by
-- apply_privacy_evidence_base.py) with the copy inline in `writer_block` — which is
-- the key the writer consumes. `supplied_copy` alone would NOT work: it has zero Go
-- readers. Verified at write time: body_in_writer_block=t.
--
-- ⚠ That is a PROMPT-LEVEL instruction, not an enforcement. The model can still
-- deviate. Verify the served page against the draft before calling this done — the
-- noted.co.uk lane checked 22 assertions by hand and so must we.
--
-- ── STATUS = 'triaged', DELIBERATELY ───────────────────────────────────────
-- `detected` items do NOT drain on this site — one of mine sat 10 hours untouched on
-- 2026-08-14. 29n used 'triaged' for the same reason.
--
-- ── THE ONE THING THAT MUST BE RIGHT ───────────────────────────────────────
-- page_type is a ROUTING KEY, not a label. bugs_open/081 is the case where a mistyped
-- page was DEPLOYED and has no repair path. The convention across the 7 fleet sites
-- that already have this page is page_type='content', name='privacy',
-- url='/privacy.html'. VERIFY THE ROW BEFORE THE BUILD ITEM RUNS (query at the foot).
--
-- ⚠ `approach` is read from the PLAN THE LLM PRODUCES (apply_gap_plan_action.go:127),
-- so asserting it here is a strong steer, not a guarantee. Check which branch ran.

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
) VALUES (
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'dartsonline-traffic-workstream',
  'content',
  'needs_content_planning',
  'high',
  'dartsonline.com has no privacy page, which blocks the affiliate applications and leaves analytics running with no policy',
  jsonb_build_object(
    'check',       'missing_privacy_page',
    'page_name',   'privacy',
    'page_type',   'content',
    'category',    'content_completeness',
    'approach',    'new_page',
    'description', 'dartsonline.com serves Google Tag Manager on every page and has no privacy policy at all (/privacy.html returns 404, measured 2026-08-16). The owner is applying to two affiliate networks and a privacy page is a precondition. Seven other sites in this estate already carry this page as name=privacy, url=/privacy.html, page_type=content, sections=["generic-text-block"], in_header=false, in_footer=true. Create it to match.',
    'suggestion',  'Create a new page named ''privacy'' with page_type ''content'' (that exact literal — it is a routing key and the wrong value silences the gates that key on it), url ''/privacy.html'', a single generic-text-block section, in_header=false, in_footer=true, nav_label ''Privacy''. THE COPY IS ALREADY WRITTEN AND APPROVED BY THE OWNER: it is in this site''s evidence_base writer_block, between the BEGIN/END OWNER-SUPPLIED PRIVACY COPY markers. Reproduce it WORD FOR WORD. Do not paraphrase it, do not summarise it, do not add reassuring sentences, and do not add sections it does not contain — it is a legal document and every sentence in it was checked against what this site actually does. In particular: do NOT invent retention periods, do NOT claim a cookie consent banner exists (there is none yet), and do NOT describe the contact form as sending data to a server (it opens the visitor''s own email client). This site is a darts PUBLICATION: it holds no stock, takes no payments and ships nothing.'
  ),
  70,
  'content-gap-planner',
  'triaged',
  'dartsonline-traffic-workstream',
  'missing_privacy_page:5fe8785b-223d-41a3-88ee-c07187622381',
  'auto'
)
ON CONFLICT DO NOTHING;

-- ── verification, in order ─────────────────────────────────────────────────
-- 1. the item moved:
--   SELECT status, attempt_count, updated_at FROM site_work_items
--   WHERE item_key='missing_privacy_page:5fe8785b-223d-41a3-88ee-c07187622381';
--
-- 2. WHICH BRANCH RAN (approach is the LLM's, not ours):
--   SELECT spec->>'approach', result FROM site_work_items
--   WHERE item_key='missing_privacy_page:5fe8785b-223d-41a3-88ee-c07187622381';
--
-- 3. the row, BEFORE it builds — page_type must be exactly 'content':
--   SELECT name, url, page_type, status, in_header, in_footer, sections
--   FROM pages p JOIN sites s ON s.id=p.site_id
--   WHERE s.domain='dartsonline.com' AND p.name='privacy';
--
-- 4. the SERVED page, against the draft — not the item status:
--   curl -s https://dartsonline.com/privacy.html | sed -e 's/<[^>]*>//g'
--   and diff the prose against DRAFT_2026-08-15_privacy_copy_for_owner_approval.md.
--   Specifically confirm the controller line reads
--   'Fine Tuning, of Fleetside, West Molesey, East Surrey'.
--
-- 5. the footer link appears — and remember the 2026-08-15 landmine: a nav rebuild
--   refreshes stored chrome on EVERY page and redeploys only SOME. Grade at the
--   served bytes across the whole sitemap, never at pages.rendered_footer.
