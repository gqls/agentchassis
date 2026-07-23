-- p3_06_lock_home_cta_sections.sql — idea.uk: lock the home-page CTA sections after p3_05 verified LIVE.
--
-- p3_05 set explicit CTA urls in content_data and re-rendered; verified on the live page 2026-07-23
-- (all four body CTAs point at /report.html / /tools.html#audience-check / /tools.html). Lock the four
-- sections so a rebuild / re-plan / LLM copy pass cannot clobber the explicit funnel URLs.
--
-- LIMIT (stated honestly): a lock does NOT stop a reason='cta_links_stale' rerender — that path runs
-- applyCTARecompute (rerender_page_sections_action.go:287) which writes a hub URL into plan.ResolvedData
-- and merges it LAST, overriding stored content_data. The generic resolver would pick a content hub
-- (guides-index/news-index), NOT /report.html. So the cta-link-integrity workstream must NOT fire a
-- blind cta_links_stale recompute on idea.uk's index (noted in RUNNING_NOTES §X.9 + a doc_note here).
-- The lock DOES protect against the more likely clobbers: rebuild, re-plan, and LLM copy rewrite.
--
-- Also cancels the p3_05 work item, which is stale (the page-rerender queue was dead ~15h so the
-- rerender was direct-fired via Kafka instead; the work item was never dispatched).

\set ON_ERROR_STOP on

BEGIN;

UPDATE page_components pc
SET locked_at = now(),
    locked_by = 'idea.uk home-CTA funnel fix (p3_05/p3_06) — explicit tool targets, do not auto-recompute',
    lock_type = 'permanent'
FROM content_components cc
WHERE cc.id = pc.component_id
  AND cc.name IN ('hero','brief-explanation','tool-list','call-to-action')
  AND pc.page_id = 'b147b925-a389-491a-8cce-ef6ba926d86a';

-- (No work-item cancel: the p3_05 page_rerender item was picked up by the revived queue and
--  COMPLETED — a second, idempotent section_data_resolved rerender producing the same result as
--  the Kafka direct-fire. Harmless.)

INSERT INTO doc_notes (subject_type, subject_key, body, categories, created_at)
VALUES (
  'pipeline', 'cta_link_integrity',
  'Home-page (/index.html) CTA sections hero/brief-explanation/tool-list/call-to-action carry EXPLICIT '
  || 'business-intent CTA urls (funnel to the paid /report.html tool + /tools.html#audience-check) set by '
  || 'p3_05 and locked by p3_06. Do NOT fire a cta_links_stale recompute on this page: applyCTARecompute '
  || 'would replace them with a content hub. See idea_uk_vm_site/RUNNING_NOTES §X.9.',
  '["cta-link-integrity","idea.uk","do-not-recompute"]'::jsonb,
  now()
);

COMMIT;

SELECT cc.name, pc.lock_type, pc.locked_at IS NOT NULL AS locked
FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
WHERE pc.page_id='b147b925-a389-491a-8cce-ef6ba926d86a'
  AND cc.name IN ('hero','brief-explanation','tool-list','call-to-action') ORDER BY pc.position;
