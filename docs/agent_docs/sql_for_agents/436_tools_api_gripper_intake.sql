-- ============================================================
-- Migration : 436_tools_api_gripper_intake.sql
-- Purpose   : the gripper-dossier intake's three tables for the
--             tools-api service — the second tool on the island,
--             beside gauntlet_rounds (198/276).
--
-- TARGET    : the ISLAND Postgres, NOT clients_db.
--   ssh root@toolsapisuk.vs.mythic-beasts.com
--   cd /opt/island && docker compose exec -T postgres \
--       psql -U tools_api -d tools_api -v ON_ERROR_STOP=1 < 436_tools_api_gripper_intake.sql
--   then ledger it in island_migrations (198's precedent).
--
-- READERS / WRITERS (internal/tools-api/store/gripper.go):
--   gripper_chat_sessions   — /session inserts; /chat claims a turn
--                             (turns+1 WHERE status='active' AND turns<30
--                             AND tokens<60000) and appends transcript/spec
--                             with jsonb ||; /submit closes it
--                             (active→submitted); poller expires idle ones.
--   gripper_report_requests — /submit inserts (pending, expires_at +24h,
--                             next_check_at +2min); GET /requests serves
--                             pending|pulled and marks pulled; the poller
--                             drives pending/pulled → fulfilled → emailed
--                             or → failed|expired → apology, every move a
--                             guarded UPDATE (WHERE status = expected).
--   gripper_daily_turns     — one row per UTC day; the conditional upsert
--                             IS the global daily turn cap (2,000 default).
--
-- WHY THE spec COLUMN HOLDS THE CLUSTER'S FIELD NAMES: the cluster's
-- report-builder workflow reads the work-item spec by name (mass_kg,
-- travel_mm, surface_material, ip_min, cycle_rate, mounting,
-- part_geometry, application) and the work-item spec IS this column
-- verbatim, via GET /requests → pull_report_requests. See
-- internal/tools-api/gripper/spec.go.
--
-- PII: email, client_ip_hash and user_agent are nullable BY DESIGN — the
-- poller nulls them 90 days after a terminal state (owner-accepted Q7);
-- transcripts are dropped 24h after last activity. The /requests feed
-- never carries email (DESIGN §5.1).
--
-- Site rows: gripper_report_requests.site_id and gripper_chat_sessions.site_id
-- reference the island's MINIMAL sites table (id/domain/status). CORS resolves
-- the Origin against it, so robot-hands.com must be present with
-- status='deployed' and its REAL cluster site id (00ff3af5-dad8-4770-9f70-
-- 3edc267a3c92, from clients_db.sites) — the INSERT below does that,
-- idempotently, the way island_db_prep.sql seeded vonc.com.
-- ============================================================

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                 WHERE table_schema='public' AND table_name='gauntlet_rounds') THEN
    RAISE EXCEPTION 'Migration 436: gauntlet_rounds does not exist — is this the ISLAND tools_api DB? (apply 198 first)';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                 WHERE table_schema='public' AND table_name='sites') THEN
    RAISE EXCEPTION 'Migration 436: sites table missing — apply island_db_prep.sql first';
  END IF;
END $$;

-- The site the intake serves. Idempotent on id; the cluster id is the one
-- the CORS layer stamps on every row, so it must be the REAL one.
INSERT INTO sites (id, domain, status)
VALUES ('00ff3af5-dad8-4770-9f70-3edc267a3c92', 'robot-hands.com', 'deployed')
ON CONFLICT (id) DO NOTHING;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema='public' AND table_name='gripper_chat_sessions') THEN
    RAISE NOTICE 'Migration 436: gripper_chat_sessions already exists — skipping';
  ELSE
    CREATE TABLE gripper_chat_sessions (
        id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
        site_id          uuid        NOT NULL REFERENCES sites(id),
        created_at       timestamptz NOT NULL DEFAULT now(),
        last_activity_at timestamptz NOT NULL DEFAULT now(),
        client_ip_hash   text,
        user_agent       text,
        turns            integer     NOT NULL DEFAULT 0,
        input_tokens     integer     NOT NULL DEFAULT 0,
        output_tokens    integer     NOT NULL DEFAULT 0,
        transcript       jsonb       NOT NULL DEFAULT '[]'::jsonb,
        spec             jsonb       NOT NULL DEFAULT '{}'::jsonb,
        status           text        NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active','submitted','expired','blocked'))
    );
    CREATE INDEX gripper_chat_sessions_status_activity_idx
        ON gripper_chat_sessions (status, last_activity_at);
    RAISE NOTICE 'Migration 436: gripper_chat_sessions created';
  END IF;

  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema='public' AND table_name='gripper_report_requests') THEN
    RAISE NOTICE 'Migration 436: gripper_report_requests already exists — skipping';
  ELSE
    CREATE TABLE gripper_report_requests (
        id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(), -- IS the report id/slug
        site_id             uuid        NOT NULL REFERENCES sites(id),
        session_id          uuid        REFERENCES gripper_chat_sessions(id),  -- NULL in plain-form mode
        email               text,                                              -- nulled after retention
        spec                jsonb       NOT NULL,
        report_url          text,
        status              text        NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending','pulled','fulfilled','emailed',
                                              'email_failed','failed','expired')),
        created_at          timestamptz NOT NULL DEFAULT now(),
        expires_at          timestamptz NOT NULL,
        first_pulled_at     timestamptz,
        last_pulled_at      timestamptz,
        fulfilled_at        timestamptz,
        emailed_at          timestamptz,
        failure_notified_at timestamptz,
        email_attempts      integer     NOT NULL DEFAULT 0,
        next_check_at       timestamptz NOT NULL,
        client_ip_hash      text,
        user_agent          text
    );
    -- The poller's three lanes all filter on status + next_check_at; rows in
    -- a resting state (emailed / email_failed) never need it.
    CREATE INDEX gripper_report_requests_due_idx
        ON gripper_report_requests (next_check_at)
        WHERE status IN ('pending','pulled','fulfilled','failed','expired');
    -- The cluster's pull: pending|pulled by created_at.
    CREATE INDEX gripper_report_requests_feed_idx
        ON gripper_report_requests (created_at)
        WHERE status IN ('pending','pulled');
    RAISE NOTICE 'Migration 436: gripper_report_requests created';
  END IF;

  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema='public' AND table_name='gripper_daily_turns') THEN
    RAISE NOTICE 'Migration 436: gripper_daily_turns already exists — skipping';
  ELSE
    CREATE TABLE gripper_daily_turns (
        day    date    PRIMARY KEY,
        turns  integer NOT NULL DEFAULT 0
    );
    RAISE NOTICE 'Migration 436: gripper_daily_turns created';
  END IF;
END $$;

-- Assert the end state; a block that skipped for the wrong reason looks
-- identical to one that worked (276's precedent).
DO $$
DECLARE n_tables int; n_idx int; n_site int;
BEGIN
  SELECT count(*) INTO n_tables FROM information_schema.tables
   WHERE table_schema='public'
     AND table_name IN ('gripper_chat_sessions','gripper_report_requests','gripper_daily_turns');
  IF n_tables <> 3 THEN
    RAISE EXCEPTION 'Migration 436: expected 3 tables, found %', n_tables;
  END IF;
  SELECT count(*) INTO n_idx FROM pg_indexes
   WHERE indexname IN ('gripper_report_requests_due_idx','gripper_report_requests_feed_idx',
                       'gripper_chat_sessions_status_activity_idx');
  IF n_idx <> 3 THEN
    RAISE EXCEPTION 'Migration 436: expected 3 indexes, found %', n_idx;
  END IF;
  SELECT count(*) INTO n_site FROM sites
   WHERE id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND domain='robot-hands.com' AND status='deployed';
  IF n_site <> 1 THEN
    RAISE EXCEPTION 'Migration 436: robot-hands.com site row missing or not deployed';
  END IF;
  RAISE NOTICE 'Migration 436 verified: 3 tables, 3 indexes, robot-hands.com deployed';
END $$;

-- Ledger entry
-- tables : gripper_chat_sessions, gripper_report_requests, gripper_daily_turns
-- fk     : *.site_id -> sites(id); gripper_report_requests.session_id -> gripper_chat_sessions(id)
-- index  : gripper_chat_sessions_status_activity_idx (status, last_activity_at)
--          gripper_report_requests_due_idx (next_check_at) partial
--          gripper_report_requests_feed_idx (created_at) partial
-- seed   : sites += robot-hands.com (cluster id 00ff3af5-…) status deployed
-- reads  : store.Gripper (internal/tools-api/store/gripper.go)
-- writes : same; every transition guarded on status
