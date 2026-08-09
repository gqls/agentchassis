-- ROLLBACK for 353 — restore .news-list-tag's original muted ink.
--
-- Restores from the row backed up before 353 ran. Prefer this over a reverse
-- replace(): the backup is the byte-exact pre-change template, so it also undoes
-- anything else 353's replace() might have touched.
--
-- NOTE this restores the template ONLY. Pages re-rendered while 353 was live keep
-- the new chip in their stored rendered_html until they re-render again — the
-- component template is not the artefact a visitor reads. See 353's header.

\set ON_ERROR_STOP on
BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM bak_cc_newslisting_20260809
   WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';
  IF n <> 1 THEN
    RAISE EXCEPTION 'backup bak_cc_newslisting_20260809 has % matching rows, expected 1 — cannot roll back safely', n;
  END IF;
END
$guard$;

UPDATE content_components c
   SET html_template = b.html_template,
       updated_at    = now()
  FROM bak_cc_newslisting_20260809 b
 WHERE c.id = b.id
   AND c.id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';

DO $verify$
DECLARE bytes int; muted int;
BEGIN
  SELECT length(html_template),
         (length(html_template)-length(replace(html_template,'--color-text-muted','')))/length('--color-text-muted')
    INTO bytes, muted
    FROM content_components WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';
  IF bytes <> 4462 THEN RAISE EXCEPTION 'expected 4462 bytes after rollback, got %', bytes; END IF;
  IF muted <> 5    THEN RAISE EXCEPTION 'expected 5 --color-text-muted uses after rollback, found %', muted; END IF;
  RAISE NOTICE 'OK: rolled back to % bytes, % muted uses', bytes, muted;
END
$verify$;

COMMIT;
