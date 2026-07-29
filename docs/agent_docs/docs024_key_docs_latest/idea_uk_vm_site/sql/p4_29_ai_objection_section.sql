-- p4_29 — answer "couldn't I just ask an AI myself?" on /report.html
--
-- The objection every visitor already has. Copy approved by the owner
-- (DRAFT_2026-07-28_ai_objection_and_example_place.md, section A) after two
-- rounds: v1 was too long and leaned on stacked negatives, which reads as
-- machine-written. This is the short, positive rewrite.
--
-- Placed at position 4: after the cards that say what the report covers, before
-- the call to action. The reader learns what it is, what is in it, why it beats
-- the obvious alternative, and only then is asked to buy.
--
-- generic-text-block requires BOTH `heading` and `content` — supplying only
-- `content` looks complete and silently escalates the whole page to the LLM
-- writer (RUNBOOK TRAP 1b, hit today building the specimen page). The guard
-- below refuses rather than repeating it.
--
-- No unique index on (page_id, position), checked — so the set-based shift of
-- the CTA and form down one is safe in a single statement.

\set page '41333d74-0c5a-4e12-b942-50ba4df793e6'

BEGIN;

UPDATE page_components SET position = position + 1
WHERE page_id = :'page' AND position >= 4;

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, build_status)
SELECT :'page', cc.id, 4, cc.function,
  jsonb_build_object(
    'heading', 'Couldn''t I just ask an AI myself?',
    'content', $h$<p>Up to a point, yes — prompt one of the big tools for an evening and you will get something worth reading.</p>
<p>The work is in the checking. We research every report with live web searches and put the source under each finding, so you can go and read it yourself. A second AI, from a different company, pulls the first one's work apart before you ever see it. Then a person reads the lot.</p>
<p>We ask the same questions every time: what the problem is, who has it, who else is solving it, where you are strong, where you are exposed, and one thing to do next.</p>
<p>And it will tell you when the answer is no. <a href="/report/example/index.html">Watch that happen in the example report →</a></p>$h$),
  'pending'
FROM content_components cc WHERE cc.id = '8d81e665-3ee0-443d-a873-690268c15fbb';

COMMIT;

-- Refuse to go further if any section is missing a required field.
DO $g$
DECLARE missing text;
BEGIN
  SELECT string_agg(cc.function || '.' || f.key, ', ') INTO missing
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  CROSS JOIN LATERAL jsonb_each(cc.input_schema->'fields') AS f(key,val)
  WHERE pc.page_id = '41333d74-0c5a-4e12-b942-50ba4df793e6'
    AND (f.val->>'required')::boolean IS TRUE
    AND NOT (pc.content_data ? f.key);
  IF missing IS NOT NULL THEN
    RAISE EXCEPTION 'ABORT: required field(s) missing: % — the render would escalate to the LLM writer', missing;
  END IF;
END $g$;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id, 'manual-p4_29', 'page_rerender', 'medium',
  'Re-render /report.html with the new "could I just ask an AI?" section',
  jsonb_build_object('domain','idea.uk','page_id',p.id::text,'filename','report.html',
                     'page_name',p.name,'reason','section_data_resolved'),
  85, 'page-rerender', 'triaged', 'idea.uk vm 7 session, p4_29',
  'page_rerender_report_p4_29_' || p.site_id::text
FROM pages p WHERE p.id = :'page';

SELECT position, slot_name, coalesce(content_data->>'heading', left(content_data::text,40)) AS heading
FROM page_components WHERE page_id = :'page' ORDER BY position;
