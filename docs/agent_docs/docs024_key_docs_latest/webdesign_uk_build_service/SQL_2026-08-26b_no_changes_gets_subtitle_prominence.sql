-- SQL_2026-08-26b — OWNER RULING 2026-08-26: "No changes are included" gets SUBTITLE
-- prominence. Verbatim: "I want to put the 'No changes are included' on the home page
-- and the 'What you get page' as a full subtitle like the others and more prominent too
-- on the 'How It Works' page."
--
-- One writer_block paragraph; facts and bans untouched (the term is already attested,
-- no_changes_included). The existing hard-term-first/fairness-pairing rule stands and
-- is referenced, not restated.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{writer_block}', to_jsonb((c.data->>'writer_block') || $wb$

NO CHANGES ARE INCLUDED GETS SUBTITLE PROMINENCE (owner ruling, 2026-08-26). On the home page and on what-you-get, the sentence "No changes are included." stands as a FULL SUBTITLE in the same visual register as the other section subtitles: its own line, subtitle styling, not buried inside a paragraph. On how-it-works it is stated more prominently than today: its own early line, not a mid-paragraph clause. In every placement, keep the standing order rule: the hard term first, then the fairness argument (the files are yours to edit) immediately alongside, never as an apology. Use the sanctioned phrasing exactly; never a bare "no" in front of a banned token.$wb$)) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-26b: No-changes gets subtitle prominence on index + what-you-get, more prominent on how-it-works (owner 2026-08-26). One paragraph; facts/bans untouched.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $chk$
DECLARE wb text; pwb text;
BEGIN
  SELECT ss.data->>'writer_block' INTO wb FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data->>'writer_block' INTO pwb FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF strpos(pwb,'SUBTITLE PROMINENCE')>0 THEN RAISE EXCEPTION 'control: paragraph already present'; END IF;
  IF (length(wb)-length(replace(wb,'NO CHANGES ARE INCLUDED GETS SUBTITLE PROMINENCE','')))/length('NO CHANGES ARE INCLUDED GETS SUBTITLE PROMINENCE') <> 1
    THEN RAISE EXCEPTION 'paragraph not exactly once'; END IF;
  IF (length(wb)-length(replace(wb,'—',''))) <> (length(pwb)-length(replace(pwb,'—',''))) THEN RAISE EXCEPTION 'em-dash count moved'; END IF;
  IF (SELECT ss.data->'facts' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current)
     <> (SELECT ss.data->'facts' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1)
    THEN RAISE EXCEPTION 'facts moved'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED';
END $chk$;

COMMIT;
