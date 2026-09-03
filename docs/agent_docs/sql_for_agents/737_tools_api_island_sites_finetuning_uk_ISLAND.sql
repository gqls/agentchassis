-- 737 — tools-api on the ISLAND: allow finetuning.uk through CORSMiddleware.
--
-- WHY THE `_ISLAND` SUFFIX (council 63be72d1 round 2, guardian, gating). scripts/migration/
-- run-migrations.sh sweeps this directory against clients_db and treats any file with an
-- UPPERCASE suffix as a sidecar it reports and never applies (SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$').
-- A plain-named island file IS picked up: 198_tools_api_gauntlet_rounds.sql is in clients_db's
-- schema_migrations (applied_by run-migrations.sh, 2026-08-08 18:14:54) and `gauntlet_rounds`
-- exists in clients_db as a result [MEASURED 2026-09-03]. This file's INSERT would have SUCCEEDED
-- there too (the core `sites` table has no NOT NULL column without a default beyond id/domain/
-- status), leaving a garbage core site row. The suffix is the only thing that keeps the runner off.
-- ROLLBACK (island): DELETE FROM sites WHERE domain = 'finetuning.uk';
--
-- TARGET    : the ISLAND Postgres (toolsapisuk.vs.mythic-beasts.com), NOT clients_db.
--             Same target and ledger as 198 / 276 / 436:
--   cd /opt/island && docker compose exec -T postgres psql -U tools_api -d tools_api -v ON_ERROR_STOP=1 < this_file.sql
--   then ledger it in island_migrations (198's precedent).
--
-- WHY. tools-api's browser route groups (gauntlet, gripper, and the playground
-- added 2026-09-03 for finetuning.uk's demo chat) sit behind CORSMiddleware,
-- which allows an Origin only when the island's OWN `sites` table holds that
-- domain at status='deployed' (`store.ActiveSiteByOrigin`). That table is
-- `island_db_prep.sql`'s minimal id/domain/status table, NOT the cluster's
-- `sites`. Measured on the island 2026-09-03 11:47Z: it holds robot-hands.com
-- and vonc.com only. Council round 1 (corr 63be72d1) checked the CLUSTER's
-- `sites` (finetuning.uk = deployed there) and read the prerequisite as met —
-- wrong table; this file is the prerequisite. Without this row every browser
-- call from https://finetuning.uk to /api/v1/tools/playground/chat is 403
-- ("origin not allowed") and the feature is dead on arrival with every line of
-- code correct (the guardian seat's exact concern).
--
-- Idempotent: re-running is a no-op. Rollback: DELETE FROM sites WHERE domain='finetuning.uk';

BEGIN;

-- `id` is uuid NOT NULL with NO default on the island (\d sites, 2026-09-03), so it is supplied.
INSERT INTO sites (id, domain, status)
SELECT gen_random_uuid(), 'finetuning.uk', 'deployed'
WHERE NOT EXISTS (SELECT 1 FROM sites WHERE domain = 'finetuning.uk');

-- Verify: exactly one row, deployed. A non-empty SELECT cannot stop a COMMIT
-- (ON_ERROR_STOP ignores results), so this is a DO/RAISE.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM sites WHERE domain = 'finetuning.uk' AND status = 'deployed';
  IF n <> 1 THEN
    RAISE EXCEPTION 'island sites: finetuning.uk deployed rows = %, want 1', n;
  END IF;
END $$;

COMMIT;
