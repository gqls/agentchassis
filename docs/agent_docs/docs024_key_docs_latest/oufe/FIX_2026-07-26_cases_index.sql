-- ============================================================================
-- oufe.com — build the cases index (/cases/index.html)
--
-- WHY
--   The page existed as an empty shell: planned, zero sections, zero
--   page_components, parked at needs_human_review ("not_built"). Meanwhile the
--   header nav item "Cases", the header CTA and a footer link all pointed at it,
--   so every one of them 404'd on a live site.
--
--   Hand-authored rather than sent to the content writer, for the reason the
--   site itself now states: an index that describes what a case study here IS
--   should say it precisely, and this one has to carry the illustration-not-
--   authority framing exactly (migration 223). It is also short.
--
-- POSTURE (owner direction 2026-07-26, PLAN §7)
--   Lead with mechanism; present a real case as clearly-marked illustration —
--   "a possibly inaccurate case study" — never as a definitive account. And say
--   what is not written yet rather than implying a fuller library exists.
--
-- Component: generic-text-block (8d81e665…), the same one migration 182 used for
-- legal pages — content_data {heading, content}, no required numeric fields, so
-- it cannot trip the bugs_open/073 empty-required-stat build failure.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, build_status)
SELECT
  p.id,
  '8d81e665-3ee0-443d-a873-690268c15fbb'::uuid,
  1,
  'main',
  jsonb_build_object(
    'heading', 'Cases',
    'content',
      '<p>A case here is a worked example, not a verdict. We take a real situation, explain the mechanism being used, and show the arithmetic that follows from it. The point is that you should be able to apply the same reasoning to the next situation you meet.</p>' ||
      '<p><strong>Treat every case on this site as a possibly inaccurate case study.</strong> We try hard to source real and current information, and we name the document behind each figure and the date we read it. That is there so you can check us — it is not a guarantee that we got it right. A source can be wrong, our reading of it can be wrong, and situations move faster than pages do.</p>' ||
      '<h3>In preparation</h3>' ||
      '<p><strong>Thames Water.</strong> A restructuring of a regulated utility, which makes it unusually well documented: the regulator publishes, and the court publishes. We are working through the primary documents now. Nothing is published here yet, because we do not have the verified figures — and a plausible estimate is not a substitute for a sourced one.</p>' ||
      '<h3>What comes first</h3>' ||
      '<p>Ahead of the cases, we are writing the mechanism explainers they depend on: how a restructuring plan can bind a class that voted against it, how the creditor waterfall decides who is paid and in what order, and what the statutory framework actually permits. Those can be explained exactly, using figures that are openly hypothetical, and they are the part that transfers to every case rather than just one.</p>' ||
      '<p>If you want to see the arithmetic behind a waterfall before any of that is written, the recovery simulator lets you set a capital structure and watch which class the value runs out in.</p>'
  ),
  'pending'
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'cases-index'
  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id = p.id);

-- Record the section on the page so the reconciler treats it as composed rather
-- than as an empty shell to rebuild.
UPDATE pages SET
  sections = '["generic-text-block"]'::jsonb,
  meta_description = 'Case studies in UK restructuring: worked examples of the mechanism, with every figure sourced and dated so you can check it.',
  updated_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND name = 'cases-index';

COMMIT;

-- Then render and deploy:
--   ./docs/agent_docs/docs024_key_docs_latest/oufe/TRIGGER_rerender_page.sh cases-index oufe.com
-- and afterwards confirm the live page and that the nav no longer 404s:
--   curl -s -o /dev/null -w '%{http_code}\n' https://oufe.com/cases/index.html
