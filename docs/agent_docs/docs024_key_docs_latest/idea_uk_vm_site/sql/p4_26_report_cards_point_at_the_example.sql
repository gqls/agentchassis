-- p4_26 — point /report.html's six cards at the specimen instead of at themselves
--
-- APPLY ONLY AFTER /report/example/index.html RETURNS 200. These links currently
-- promise something that does not exist; pointing them at a page that is not yet
-- live would replace a self-link with a 404, which is worse.
--
-- The defect: all six cards carry link_url '/report', which serves the SAME page
-- they sit on — byte-identical, 44,394 bytes. So "See an example", "View specimen
-- report", "Read more" and "See the method" all reload the page the reader is
-- already on. Every one returns 200, so no link checker has ever complained:
-- the same class as the hero CTA fixed in p4_20, where a wrong destination and a
-- dead destination are indistinguishable to anything that only checks status.
--
-- Two of those labels were the sharper problem — "See an example" and "View
-- specimen report" promised an artefact the site did not have at all. It has one
-- now (p4_24), so they can be honoured rather than merely reworded.
--
-- Five cards go to the example, which genuinely demonstrates each of their
-- claims. "A specific next step" keeps a request-a-report label but is pointed
-- at the FORM ANCHOR rather than at '/report' — it is a CTA, not a demonstration.

\set page '41333d74-0c5a-4e12-b942-50ba4df793e6'

-- BEFORE
SELECT c->>'title' AS card, c->>'link_url' AS url, c->>'link_label' AS label
FROM page_components pc, LATERAL jsonb_array_elements(pc.content_data->'cards') c
WHERE pc.page_id = :'page' AND pc.position = 3;

BEGIN;

UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{cards}', (
      SELECT jsonb_agg(
               CASE WHEN c->>'title' = 'A specific next step'
                    THEN c || '{"link_url":"/report.html#request-a-report","link_label":"Request a report"}'::jsonb
                    ELSE c || '{"link_url":"/report/example/index.html","link_label":"See it in the example report"}'::jsonb
               END ORDER BY ord)
      FROM jsonb_array_elements(pc.content_data->'cards') WITH ORDINALITY AS t(c, ord)
    )),
    updated_at = now()
WHERE pc.page_id = :'page' AND pc.position = 3 AND pc.locked_at IS NULL;

COMMIT;

-- AFTER — no card may still point at '/report'
SELECT c->>'title' AS card, c->>'link_url' AS url, c->>'link_label' AS label
FROM page_components pc, LATERAL jsonb_array_elements(pc.content_data->'cards') c
WHERE pc.page_id = :'page' AND pc.position = 3;

SELECT count(*) AS cards_still_self_linking
FROM page_components pc, LATERAL jsonb_array_elements(pc.content_data->'cards') c
WHERE pc.page_id = :'page' AND pc.position = 3 AND c->>'link_url' = '/report';
