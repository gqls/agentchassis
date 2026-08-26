-- SQL_2026-08-26c — sharpen: on the HOME page the sentence OPENS a subtitle.
-- The prominence wave delivered what-you-get (subtitle opens with the sentence + own h3)
-- and how-it-works (own h3), but index carries it only mid-flow inside the cta-subtitle.
BEGIN;
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned, jsonb_set(c.data, '{writer_block}', to_jsonb(replace(c.data->>'writer_block',
    'On the home page and on what-you-get, the sentence "No changes are included." stands as a FULL SUBTITLE in the same visual register as the other section subtitles: its own line, subtitle styling, not buried inside a paragraph.',
    'On the home page and on what-you-get, the sentence "No changes are included." stands as a FULL SUBTITLE in the same visual register as the other section subtitles, and it OPENS the subtitle element: it is the first sentence of the hero subheadline or of the offer-in-short lead, exactly as what-you-get now does, never mid-flow inside another sentence run.'))) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-26c: home-page no-changes subtitle must OPEN the element (index under-delivered the prominence wave).',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;
DO $chk$
DECLARE wb text; pwb text;
BEGIN
  SELECT ss.data->>'writer_block' INTO wb FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data->>'writer_block' INTO pwb FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF strpos(pwb,'not buried inside a paragraph.')=0 THEN RAISE EXCEPTION 'control: needle absent before'; END IF;
  IF strpos(wb,'OPENS the subtitle element')=0 THEN RAISE EXCEPTION 'sharpened clause missing'; END IF;
  IF (length(wb)-length(replace(wb,'—',''))) <> (length(pwb)-length(replace(pwb,'—',''))) THEN RAISE EXCEPTION 'em-dash moved'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED';
END $chk$;
COMMIT;
