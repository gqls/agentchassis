-- SQL_2026-07-12_add_sprite_sheet_kind.sql
--
-- Phase I2.0 (sprite-sheet bullets, SCOPE_I2_sprite_sheets.md): extend
-- site_plan_imagery's chk_kind with 'sprite_sheet'. The Go mirror
-- (validImageryKinds in write_site_plan_action.go) gains the same value in
-- the paired code commit — constraint and mirror always change together.
--
-- text + CHECK per storage conventions (no native enums). Idempotent: the
-- guard skips if sprite_sheet is already allowed.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS site_plan_imagery_chk_backup_20260712 AS
SELECT conname, pg_get_constraintdef(oid) AS def
FROM pg_constraint WHERE conrelid = 'site_plan_imagery'::regclass;

BEGIN;

DO $mig$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'site_plan_imagery'::regclass
          AND conname = 'chk_kind'
          AND pg_get_constraintdef(oid) LIKE '%sprite_sheet%'
    ) THEN
        RAISE NOTICE 'chk_kind already includes sprite_sheet — no-op';
        RETURN;
    END IF;

    ALTER TABLE site_plan_imagery DROP CONSTRAINT chk_kind;
    ALTER TABLE site_plan_imagery ADD CONSTRAINT chk_kind CHECK (
        kind IN ('logo', 'hero', 'illustration', 'icon', 'infographic', 'sprite_sheet')
    );
    RAISE NOTICE 'chk_kind extended with sprite_sheet';
END
$mig$;

-- Verify
DO $verify$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'site_plan_imagery'::regclass
          AND conname = 'chk_kind'
          AND pg_get_constraintdef(oid) LIKE '%sprite_sheet%'
    ) THEN
        RAISE EXCEPTION 'chk_kind does not include sprite_sheet after migration';
    END IF;
END
$verify$;

COMMIT;
