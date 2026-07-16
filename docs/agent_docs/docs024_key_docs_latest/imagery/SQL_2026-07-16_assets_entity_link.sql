-- SQL_2026-07-16_assets_entity_link.sql
--
-- Phase I3.1 (content-linked card imagery / Lane B, PLAN Phase I3, decision D2
-- user-confirmed 2026-07-08 and migration approved 2026-07-16): generic
-- content-entity linking on assets. An asset may now belong to a content
-- entity — entity_type names the table-concept ('page', 'content_feed_item',
-- 'affiliate_product', ...), entity_id is the row's uuid. Both nullable;
-- Lane A (plan-driven) assets simply leave them NULL.
--
-- No FK on entity_id (it is polymorphic across tables by design — same
-- pattern the codebase uses for site_work_items.spec references). text +
-- partial index per storage conventions. Additive-only; no rows modified.
-- Idempotent: guarded on column existence.

\set ON_ERROR_STOP on

-- Backup of the table's current shape (columns, not data — no data changes).
CREATE TABLE IF NOT EXISTS assets_schema_backup_20260716 AS
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns WHERE table_name = 'assets';

BEGIN;

DO $mig$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'assets' AND column_name = 'entity_type'
    ) THEN
        RAISE NOTICE 'assets.entity_type already exists — no-op';
        RETURN;
    END IF;

    ALTER TABLE assets ADD COLUMN entity_type text;
    ALTER TABLE assets ADD COLUMN entity_id uuid;

    RAISE NOTICE 'assets gained entity_type + entity_id';
END
$mig$;

-- Fast entity → assets lookup; only rows that are entity-linked are indexed.
CREATE INDEX IF NOT EXISTS idx_assets_entity
    ON assets (entity_type, entity_id)
    WHERE entity_type IS NOT NULL;

-- One ACTIVE asset per (site, entity, purpose) — the same uniqueness idea as
-- idx_assets_site_asset_key_unique, applied to the entity link, so a re-derive
-- supersedes rather than accumulates.
CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_entity_purpose_unique
    ON assets (site_id, entity_type, entity_id, purpose)
    WHERE entity_type IS NOT NULL AND status = 'active';

-- Verify
DO $verify$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'assets' AND column_name = 'entity_type'
    ) OR NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'assets' AND column_name = 'entity_id'
    ) THEN
        RAISE EXCEPTION 'assets entity columns missing after migration';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'assets' AND indexname = 'idx_assets_entity'
    ) THEN
        RAISE EXCEPTION 'idx_assets_entity missing after migration';
    END IF;
END
$verify$;

COMMIT;
