-- SQL_2026-08-25f — the FAQ's dissatisfaction QUESTION may not carry a banned token.
--
-- WHY. Wave-2 faq failed with 1 blocker: the writer asked "What if I don't like the
-- finished site? Can I get my money back?" and the promise-shape ban fired on
-- "money back" - correctly: the negation guard scans the clause for do-not/never
-- and a QUESTION is not a denial, so the bare token reads as a promise. The ban is
-- right (narrowed 2026-08-18 exactly so denials pass); the steering gap is that
-- nothing told the writer the question FORM is the trap. One sentence in
-- writer_block; facts and bans untouched.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{writer_block}', to_jsonb(replace(c.data->>'writer_block',
      $n$call them the full terms, never the refund position (a rewrite was blocked on exactly that phrase, 2026-08-18).$n$,
      $n$call them the full terms, never the refund position (a rewrite was blocked on exactly that phrase, 2026-08-18). And in the FAQ, the dissatisfaction question itself must not carry a banned token: ask "What if I don't like the finished site?", never "Can I get my money back?". A question is not a denial, so the gate reads the bare token as a promise and blocks the page (it did, 2026-08-25); the answer then states the sanctioned denial and points at what the customer CAN do: the files are theirs to edit, and anyone they hire can take them on.$n$))) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-25f: FAQ dissatisfaction question must not carry a banned token (wave-2 faq blocker). One writer_block sentence; facts and bans untouched.',
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
  IF strpos(pwb,'never the refund position (a rewrite was blocked on exactly that phrase, 2026-08-18).')=0
    THEN RAISE EXCEPTION 'control: anchor sentence absent before'; END IF;
  IF strpos(wb,'the dissatisfaction question itself must not carry a banned token')=0
    THEN RAISE EXCEPTION 'addition missing'; END IF;
  IF (length(wb)-length(replace(wb,'—',''))) <> (length(pwb)-length(replace(pwb,'—',''))) THEN RAISE EXCEPTION 'em-dash count moved'; END IF;
  IF (SELECT ss.data->'facts' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current)
     <> (SELECT ss.data->'facts' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1)
    THEN RAISE EXCEPTION 'facts moved'; END IF;
  IF (SELECT ss.data->'banned_claims' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current)
     <> (SELECT ss.data->'banned_claims' FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1)
    THEN RAISE EXCEPTION 'banned_claims moved'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED: one writer_block sentence added';
END $chk$;

COMMIT;
