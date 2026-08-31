-- 666 — the data-handling commitments into the privacy policy, where a reader looks for them
--
-- Companion to 665. The owner's 2026-08-26 decisions are contractual in the terms; three of them
-- are ALSO data-protection statements and a privacy notice is where a reader (and a regulator)
-- looks for retention, deletion and location. `[MEASURED 2026-08-31]` the policy said nothing about
-- training documents or the trained model in either place.
--
-- Extends the two sections that already exist rather than bolting on a new one:
--   "How we store your data"     <- where the documents physically sit during training and handover
--   "How long we keep your data" <- the 30-day default and the one-week deletion-on-request
--
-- Text is the verbatim `writer_line` of ft-data-location, ft-retention-default and
-- ft-deletion-window. Nothing composed.
--
-- ⚠ NO UNLOCK NEEDED and none performed: unlike `terms`, this page carries no component lock and
-- `rebuild_policy` is 'generic'. Asserted below rather than assumed.
--
-- ⚠ AND THAT IS A REAL ASYMMETRY, LEFT FOR THE OWNER. Because this page IS rebuildable, a future
-- page-build can regenerate it and silently drop or reword these commitments — the same exposure
-- 665 protects `terms` from by keeping `rebuild_policy='owned'`. Locking it would match terms, but
-- that is a NEW protection rather than a restoration, so it is his call and not taken here.
--
-- Rollback: 666_..._ROLLBACK.sql

BEGIN;

DO $$
DECLARE lk timestamptz; pol text;
BEGIN
    SELECT pc.locked_at, COALESCE(p.rebuild_policy,'generic') INTO lk, pol
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name='privacy-policy';
    IF lk IS NOT NULL THEN
        RAISE EXCEPTION '666: privacy-policy IS locked (locked_at=%) — this file assumes it is not; unlock/relock like 665 instead', lk;
    END IF;
    IF pol = 'owned' THEN
        RAISE EXCEPTION '666: privacy-policy is rebuild_policy=owned — re-read 665 before editing';
    END IF;
END $$;

UPDATE page_components pc
   SET content_data = jsonb_set(pc.content_data, '{content}',
         to_jsonb(replace(replace(pc.content_data->>'content', 'We do not use your data to train AI models without a separate, explicit agreement with you.</p>', 'We do not use your data to train AI models without a separate, explicit agreement with you.</p><p><strong>Documents you send us for fine-tuning.</strong> During training your documents sit on a rented GPU machine; for the handover they sit in our storage. That is the whole of it.</p>'), 'If you would like us to delete your data, contact us and we will do so unless we have a legal obligation to retain it.</p>', 'If you would like us to delete your data, contact us and we will do so unless we have a legal obligation to retain it.</p><p><strong>Fine-tuning documents and models.</strong> We keep your documents and your model for 30 days after we hand it over, then delete them. Ask us to delete your documents and your model sooner and we will, within a week.</p>'))),
       updated_at = now()
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain='finetuning.uk' AND p.name='privacy-policy'
   AND pc.content_data->>'content' NOT LIKE '%Documents you send us for fine-tuning%';

DO $$
DECLARE has_all bool; n int;
BEGIN
    SELECT count(*), bool_and(
             pc.content_data->>'content' LIKE '%rented GPU machine%' AND
             pc.content_data->>'content' LIKE '%30 days after we hand it over%' AND
             pc.content_data->>'content' LIKE '%within a week%')
      INTO n, has_all
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name='privacy-policy';
    IF n <> 1 THEN RAISE EXCEPTION '666: expected 1 privacy-policy component, found %', n; END IF;
    IF NOT has_all THEN RAISE EXCEPTION '666: the three data commitments are not all present'; END IF;
    RAISE NOTICE '666 OK: location, retention and deletion stated in the privacy policy';
END $$;

COMMIT;
