-- 581_refuse_selector_invisible_section_birth_ROLLBACK.sql
--
-- Reverses 581. Safe and complete: the migration created exactly one trigger
-- and one function and changed NO data, so dropping both restores the prior
-- state exactly. Nothing needs un-backfilling because nothing was backfilled.
--
-- After this runs, a section-level non-forked component can once again be born
-- with a NULL section_type and be silently invisible to the component selector.
-- That is the state bugs_open/351 documents; if you are rolling this back,
-- record why there, because the next reader will otherwise re-derive it.

BEGIN;

DROP TRIGGER IF EXISTS trg_cc_refuse_null_section_type ON content_components;
DROP FUNCTION IF EXISTS refuse_selector_invisible_section();
-- Pre-revision names, in case an older build of 581 was the one applied.
DROP TRIGGER IF EXISTS trg_cc_refuse_null_section_type_birth ON content_components;
DROP FUNCTION IF EXISTS refuse_selector_invisible_section_birth();

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
   WHERE c.relname = 'content_components'
     AND NOT t.tgisinternal
     AND t.tgname IN ('trg_cc_refuse_null_section_type', 'trg_cc_refuse_null_section_type_birth');
  IF n <> 0 THEN
    RAISE EXCEPTION '581 ROLLBACK: the trigger is still present after DROP.';
  END IF;
  RAISE NOTICE '581 ROLLBACK: trigger and function removed; no data was changed by 581 or by this file.';
END $$;

COMMIT;
