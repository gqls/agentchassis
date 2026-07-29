-- 265_asset_ingest_staging.sql
--
-- Staging table for OPERATOR-SUPPLIED asset bytes — the ingress half of the
-- asset-amend path (bugs_open/131 og-card slug: three sites' stored "logo" is
-- not a logo, and the platform had NO path for a human to supply a corrected
-- image; every image ever stored arrived from a generator).
--
-- WHY A TABLE AND NOT THE WORK-ITEM SPEC. Bytes must not ride Kafka: the
-- kafka-go writer caps messages at 1 MiB (producer.go never sets BatchBytes),
-- and the recorded doctrine is "heavy artifacts live in the DB, retrievable
-- by id" (concept register, system-architecture). The work item carries only
-- this table's row id. BYTEA precedent: chassis_intake_events.payload (249).
--
-- FLOW: scripts/amend-asset.sh base64s a file into `content` via psql stdin
-- and queues a site_work_items row ({mode:'ingest_upload', staging_id}) in
-- the same transaction → build-dispatch → asset-deployer →
-- ingest_staged_asset action (chassis ≥ the roll that carries it) validates
-- (sha256, image-decode), uploads to S3 at a NEW key, amends the assets row
-- in place, and marks this row 'ingested'.
--
-- STATUS VOCABULARY:
--   pending    — staged, not yet claimed
--   processing — claimed by a running ingest (atomic pending→processing)
--   ingested   — done; consumed_at set
--   refused    — the platform declined (locked asset, not an image,
--                cross-site mismatch); `error` says why
--   failed     — infrastructure/integrity failure (sha mismatch, S3, DB)
--
-- RETENTION: operator-driven, low volume. Rows with status != 'pending'
-- older than 7 days may be purged (same policy family as
-- chassis_intake_events); no automated purge is installed by this migration.
--
-- ORDERING: this table is inert until the chassis image carrying
-- ingest_staged_asset is rolled — safe to apply any time. 266 (the
-- asset-deployer mode) is NOT: apply it only after the pod-grep passes.
--
-- ROLLBACK: DROP TABLE asset_ingest_staging;  (no other object depends on it)

\set ON_ERROR_STOP on

DO $guard$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables
                WHERE table_schema = 'public' AND table_name = 'asset_ingest_staging') THEN
        RAISE EXCEPTION 'asset_ingest_staging already exists — this migration is not re-runnable; drop it first if you mean to recreate';
    END IF;
END
$guard$;

BEGIN;

CREATE TABLE asset_ingest_staging (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id     uuid NOT NULL REFERENCES sites(id),
    asset_key   text NOT NULL,
    purpose     text,
    content     bytea NOT NULL,
    sha256      text NOT NULL,
    note        text,
    created_by  text NOT NULL,
    status      text NOT NULL DEFAULT 'pending',
    error       text,
    consumed_at timestamptz,
    created_at  timestamptz DEFAULT now(),
    CONSTRAINT asset_ingest_staging_status_check
        CHECK (status IN ('pending', 'processing', 'ingested', 'refused', 'failed'))
);

COMMENT ON TABLE asset_ingest_staging IS
    'Operator-supplied asset bytes awaiting ingest by the ingest_staged_asset action (asset-amend path, bugs_open/131 og-card). Loader: scripts/amend-asset.sh.';
COMMENT ON COLUMN asset_ingest_staging.sha256 IS
    'Hex digest computed by the loader over the original file; the action re-computes over content and refuses on mismatch.';

CREATE INDEX idx_asset_ingest_staging_pending
    ON asset_ingest_staging (created_at) WHERE status = 'pending';

DO $verify$
DECLARE
    v_count int;
BEGIN
    SELECT count(*) INTO v_count
      FROM information_schema.columns
     WHERE table_name = 'asset_ingest_staging'
       AND column_name IN ('id','site_id','asset_key','purpose','content','sha256',
                           'note','created_by','status','error','consumed_at','created_at');
    IF v_count <> 12 THEN
        RAISE EXCEPTION 'asset_ingest_staging has % of 12 expected columns', v_count;
    END IF;
    RAISE NOTICE 'asset_ingest_staging created (12 columns, pending index, status check)';
END
$verify$;

COMMIT;
