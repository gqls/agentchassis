-- 106_thunder_unreachable_counter.sql
-- DB: clients_db   (public.thunder_instances lives in clients_db, NOT templates_db)
--
-- Adds a per-instance consecutive-unreachable probe counter for the
-- thunder-training-monitor. The monitor's probe (ssh_get_status) returns
-- reachable:false as a VALID answer when a box is down. A single down-probe is
-- not grounds to decommission (could be a transient SSH/network blip), so the
-- monitor counts CONSECUTIVE unreachable probes and only treats the instance as
-- 'lost' (-> mark the run failed + decommission) once the count crosses a
-- threshold. Each scheduler tick is a fresh sub-agent that can't hold a count in
-- memory, so the state must live on the row.
--
--   consecutive_unreachable_probes : incremented on each unreachable probe,
--                                    reset to 0 on any reachable probe.
--   last_probe_at                  : observability — when the monitor last
--                                    probed this instance.
--
-- Written/reset by the record_probe_streak action (mode=bump | mode=reset).
-- Idempotent (ADD COLUMN IF NOT EXISTS). The existing
-- trg_thunder_instances_updated_at trigger maintains updated_at on UPDATE, so
-- the streak UPDATEs do not set it.
--
-- NOTE on numbering: a project DOC named 106_claude_anthropic_skill.md exists —
-- it is not a SQL migration, but confirm 106 is the next free number in your
-- migration runner and renumber this file if the SQL sequence differs.

BEGIN;

ALTER TABLE public.thunder_instances
    ADD COLUMN IF NOT EXISTS consecutive_unreachable_probes integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_probe_at                  timestamp with time zone;

-- Verify both columns exist before committing.
DO $$
BEGIN
    IF (SELECT count(*)
          FROM information_schema.columns
         WHERE table_schema = 'public'
           AND table_name   = 'thunder_instances'
           AND column_name IN ('consecutive_unreachable_probes', 'last_probe_at')) <> 2 THEN
        RAISE EXCEPTION 'migration 106 failed: expected both columns on public.thunder_instances';
    END IF;
END $$;

COMMIT;

-- Rollback (manual):
--   ALTER TABLE public.thunder_instances
--     DROP COLUMN IF EXISTS consecutive_unreachable_probes,
--     DROP COLUMN IF EXISTS last_probe_at;
