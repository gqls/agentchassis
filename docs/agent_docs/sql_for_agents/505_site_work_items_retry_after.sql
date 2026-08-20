-- 505 — site_work_items gains a not-claimable-before stamp, and adopts
-- reaper_policies for its retry numbers.
--
-- bugs_open/307, owner ruling 2026-08-18: "a transient blip should return the
-- item to queued." The retry machinery already existed and already ran; what it
-- lacked was a WAIT. A re-triaged item was claimable on the very next dispatch
-- tick, so during the 2h43m git-adapter outage of 2026-08-17, 88 of 100 items
-- burned all three attempts inside the burst. Retry-without-delay is equivalent
-- to no retry for any outage longer than a few dispatch cycles.
--
-- WHY A COLUMN AND NOT A STATUS. Every parking state in this estate was measured
-- and rejected for this job:
--   * 'blocked' — the live `feasibility-recheck` task (enabled, every 600s)
--     releases EVERY blocked row whose handler_agent exists, with no timestamp
--     condition, and sets error = NULL. A cooldown parked there survives at most
--     ten minutes and loses its reason. (That is also why the table holds zero
--     blocked rows and always has: not unused, continuously drained.)
--   * 'deferred' — absent from idx_swi_dedup's exclusion list, so the row keeps
--     its (site_id, item_key) slot and the detector that filed it cannot re-file
--     it: undispatchable and un-refilable at once.
--   * a NEW status — invisible to idx_swi_dedup, to the detected-item-promoter's
--     success floor, and to the stale-orchestration reaper's status enumeration.
-- A nullable timestamp on a row that stays 'triaged' perturbs none of them.
--
-- SAFE AGAINST THE RUNNING BINARY. Nothing reads or writes this column until the
-- chassis that carries the failure-write contract rolls; NULL means "claimable
-- now", which is what every existing row gets. The four read sites gain
-- `(retry_after IS NULL OR retry_after <= NOW())`, which is a tautology while the
-- column is entirely NULL — so this file and 506 may be applied in any order,
-- before or after the roll, without a window in which dispatch stops.
--
-- NO INDEX. The predicate is a residual filter behind the existing partial
-- indexes (idx_swi_handler, idx_swi_site_pending, both WHERE status IN
-- ('triaged','approved')), which already reduce the candidate set to tens of
-- rows; NOW() is STABLE-not-IMMUTABLE so it cannot be indexed usefully anyway.
--
-- Idempotent throughout (IF NOT EXISTS / ON CONFLICT), so a migration-runner
-- replay after a direct psql apply is a no-op.

BEGIN;

-- ── 1. The stamp ────────────────────────────────────────────────────────────
ALTER TABLE site_work_items
    ADD COLUMN IF NOT EXISTS retry_after timestamptz;

COMMENT ON COLUMN site_work_items.retry_after IS
    'bugs_open/307: not-claimable-before stamp written by the work-item failure '
    'ladder. NULL = claimable now. Honoured by claim_work_item, '
    'LoadWorkItemsAction, build-pipeline-trigger, find_dispatchable_site and '
    'build-dispatch-watchdog. A triaged row with a future retry_after is WAITING, '
    'not stuck.';

-- ── 2. The numbers, declared rather than hard-coded ─────────────────────────
-- RFC_018 (architecture_review/RFC_018_reaper_accounting_as_a_shared_mechanism.md)
-- built reaper_policies for exactly this and named this queue as the second
-- consumer it was waiting for: "stale-work-item-reaper / claimed-item-timeout
-- (site_work_items): closest fit; already has attempt_count — adopt
-- reaper_policies for its numbers first, executor second." Owner decision
-- 2026-08-08 #3: "a task type declares its own ceiling by inserting a row."
--
-- So the backoff lives here, not as a literal in Go. An operator retunes a queue
-- with an INSERT and no build. The ladder reads the item_type row if one exists,
-- else this '__default__' row, else its code fallback of 30 minutes.
--
-- park_after is set to 3 to match site_work_items.max_attempts' own default; it
-- is recorded for consistency with the sibling consumer and is NOT read by the
-- ladder, which takes its ceiling from the row's own max_attempts (69 live rows
-- deliberately run max_attempts=1 as a one-shot lane, and a queue-wide ceiling
-- must not override that).
INSERT INTO reaper_policies (queue, item_type, park_after, backoff_minutes, stale_after_minutes, notes)
VALUES ('site_work_items', '__default__', 3, 30, 20,
        'bugs_open/307: retry backoff for the work-item failure ladder. 30 min x attempt '
        '(30m then 60m on a max_attempts=3 item), chosen to outlive the median adapter '
        'outage while staying far inside stale-work-item-reaper''s 48h ceiling on a '
        'triaged row. Per-item_type rows override this one.')
ON CONFLICT (queue, item_type) DO NOTHING;

-- ── 3. Verify (DO/RAISE — a SELECT cannot stop a COMMIT) ────────────────────
DO $do$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'site_work_items' AND column_name = 'retry_after') THEN
    RAISE EXCEPTION '505: site_work_items.retry_after was not added';
  END IF;

  -- The whole "safe against the running binary" claim rests on this: if any row
  -- already carried a non-NULL stamp, the read-side predicates in 506 would
  -- start excluding work the moment they applied.
  IF (SELECT count(*) FROM site_work_items WHERE retry_after IS NOT NULL) <> 0 THEN
    RAISE EXCEPTION '505: retry_after is unexpectedly populated on % rows — 506''s predicates would not be inert',
      (SELECT count(*) FROM site_work_items WHERE retry_after IS NOT NULL);
  END IF;

  IF to_regclass('reaper_policies') IS NULL THEN
    RAISE EXCEPTION '505: reaper_policies is missing — migration 335 must be applied first';
  END IF;

  IF (SELECT count(*) FROM reaper_policies
       WHERE queue = 'site_work_items' AND item_type = '__default__') <> 1 THEN
    RAISE EXCEPTION '505: the default backoff policy row was not seeded';
  END IF;
END
$do$;

COMMIT;
