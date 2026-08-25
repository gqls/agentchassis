-- 632_site_lock_exception_list_ROLLBACK.sql — reverses 632_site_lock_exception_list.sql
--
-- ⚠ ROLL BACK 633 FIRST. If the held config half has been applied, dropping this
-- column leaves `find_dispatchable_site`'s pre_query referencing a column that no
-- longer exists, and the build pipeline's site selection fails fleet-wide. The
-- guard below refuses in that case rather than discovering it in production.

BEGIN;

DO $pre$
DECLARE
    v_refs bigint;
BEGIN
    SELECT count(*) INTO v_refs
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND s.value::text ILIKE '%lock_except_item_ids%';
    IF v_refs > 0 THEN
        RAISE EXCEPTION '632 ROLLBACK REFUSED: % live step config(s) still reference lock_except_item_ids. Roll back 633 first, or find_dispatchable_site will fail fleet-wide on a missing column.', v_refs;
    END IF;
END;
$pre$;

ALTER TABLE sites DROP COLUMN IF EXISTS lock_except_item_ids;

DO $guard$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
                WHERE table_name='sites' AND column_name='lock_except_item_ids') THEN
        RAISE EXCEPTION '632 ROLLBACK: column still present';
    END IF;
    RAISE NOTICE '632 ROLLBACK OK: sites.lock_except_item_ids dropped; locked_at/locked_by untouched.';
END;
$guard$;

COMMIT;
