-- p4_16_report_copy_pass.sql — /report.html: light copy pass now the engine matches the promise.
--
-- Owner ruled option B (extend the engine) and the extended binary is DEPLOYED (verified running
-- since 15:11). The page's claims about assessment, sources and honesty are now TRUE — but the
-- copy UNDERSELLS: it never mentions the second half of what a customer receives (the further
-- ideas generated around their business, web-checked and ranked, with what didn't make the cut
-- and why). This pass adds that half and nothing else structural. Sections are NOT locked
-- (verified before writing), so a plain content_data edit + section_data_resolved rerender lands.
--
-- Deliberately light: the quoted paragraphs the owner cares about are kept intact; one paragraph
-- is appended to the body, one sentence extended in the closing CTA.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/report.html'
    AND pc.lock_type IS NOT NULL;
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % report.html section(s) are locked — a rerender would discard the edit (save_page_sections preserves locked rows). Unlock first.', n;
  END IF;
END
$guard$;

-- Body: append the further-ideas half after the "A report typically covers" paragraph.
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{content}', to_jsonb(
      (pc.content_data->>'content') ||
      '<p>The report does not stop at the idea you sent. Around it, we generate further ideas for your business — combinations of what you already have with what is now possible — and put each through the same discipline: a second, independent AI critiques them, the promising ones are checked against the live web (do competitors exist, would people actually pay), and the survivors are ranked, each with its own cheap first test and its own sources. You also see what did not make the cut, and why — sometimes the most useful page in the report.</p>'
    ))
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/report.html'
  AND pc.slot_name = 'generic-text-block'
  AND position('The report does not stop at the idea you sent' in pc.content_data->>'content') = 0;

-- Closing CTA: mention both halves.
UPDATE page_components pc
SET content_data = pc.content_data || jsonb_build_object(
      'subheadline',
      'The Verified Idea Report costs £29 and gives you a structured, research-based assessment of your idea — the market, the competition, what makes it defensible, and a sensible next step, with sources you can check — plus further ideas for your business, tested the same way. It''s an information service, not professional advice. If we don''t think there''s enough to work with, we say so.'
    )
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/report.html'
  AND pc.slot_name = 'call-to-action';

DO $guard2$
DECLARE c text;
BEGIN
  SELECT pc.content_data->>'content' INTO c
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/report.html'
    AND pc.slot_name = 'generic-text-block';
  IF position('The report does not stop at the idea you sent' in c) = 0 THEN
    RAISE EXCEPTION 'ABORT: body edit did not land.';
  END IF;
END
$guard2$;

COMMIT;

SELECT pc.slot_name, length(pc.content_data->>'content') AS body_len,
       left(pc.content_data->>'subheadline', 60) AS sub
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/report.html'
  AND pc.slot_name IN ('generic-text-block','call-to-action')
ORDER BY pc.position;
