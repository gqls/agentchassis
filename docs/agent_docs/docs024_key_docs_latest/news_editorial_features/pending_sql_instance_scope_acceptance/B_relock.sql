-- news_editorial_features lane — 2026-08-25
-- RE-LOCK the two evidence-timeseries instances after the instance-scope
-- conversion has been delivered and VERIFIED at the served artefact.
--
-- > **CORRECTED 2026-08-25 (v2): v1 COULD NOT RE-LOCK, and that is the
-- > DANGEROUS half of the pair.** v1 set `lock_type='permanent'` and
-- > `locked_by` but never restored `locked_at`. Automation's writability test is
-- >     (locked_at IS NULL OR (lock_type='timed' AND … expired))
-- > so against a CORRECTED A (which nulls `locked_at`), v1 would have left both
-- > rows fully AGENT-WRITABLE while displaying as `permanent` in the admin
-- > dashboard — a silently unlocked flagship row that every sweep on the estate
-- > may overwrite, and that reads as locked to anyone who checks.
-- > v1 happened to restore the correct state on 2026-08-25 ONLY because the v1
-- > A had left `locked_at` intact, i.e. two defects cancelling. Caught when the
-- > A defect was diagnosed; neither script had ever been run before that day.
--
-- Run this as soon as verification passes. An unlocked flagship row is exposed
-- to every sweep on the estate, and this lane has already been hit once by an
-- improvement-loop misfire (2026-08-22).
--
-- Restores EXACTLY what was there before, per row — including the ORIGINAL
-- `locked_at`, captured 2026-08-25 before the unlock. Deliberately NOT NOW():
-- the lock's provenance is when the human locked it, and rewriting it would
-- destroy the only record of that, as well as resetting the age that any
-- lock-review sweep reads.
-- Idempotent: unconditional on the two ids, so it lands the target state from
-- ANY intermediate state (fully unlocked, half-cleared, or already restored).

BEGIN;

UPDATE page_components pc
   SET locked_at       = v.locked_at::timestamptz,
       locked_by       = 'news_editorial_features-lane',
       lock_type       = 'permanent',
       lock_expires_at = NULL
  FROM (VALUES
    ('d344585f-6f79-4a18-a1fe-3116b68a4c52', '2026-08-19 15:17:43.126181+00'),
    ('ea6b4ca7-7717-4e29-ae1c-88844040b0d2', '2026-08-20 16:59:30.895515+00')
  ) AS v(id, locked_at)
 WHERE pc.id = v.id::uuid
RETURNING pc.id, pc.slot_name, pc.locked_at, pc.lock_type, pc.locked_by;

-- Assert the end state INSIDE the transaction. A block of SELECTs cannot stop a
-- COMMIT (ON_ERROR_STOP ignores a non-empty result), so this is a DO/RAISE:
-- if either row is still agent-writable, the re-lock did not take and the
-- transaction must not commit a half-restored lock.
DO $$
DECLARE writable_count int;
BEGIN
  SELECT count(*) INTO writable_count
    FROM page_components
   WHERE id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
                'ea6b4ca7-7717-4e29-ae1c-88844040b0d2')
     AND (locked_at IS NULL OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL
                                AND lock_expires_at < NOW()));
  IF writable_count > 0 THEN
    RAISE EXCEPTION 'RE-LOCK FAILED: % of 2 rows still agent-writable — refusing to commit', writable_count;
  END IF;
END $$;

COMMIT;

-- Raw read-back: both rows must show permanent / news_editorial_features-lane,
-- the ORIGINAL locked_at, and agent_writable = f.
SELECT id, slot_name, locked_at, lock_type, locked_by,
       (locked_at IS NULL OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL
                              AND lock_expires_at < NOW())) AS agent_writable,
       component_version_id IS NOT NULL AS stamped,
       length(rendered_html) AS html_bytes, updated_at
  FROM page_components
 WHERE id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
              'ea6b4ca7-7717-4e29-ae1c-88844040b0d2');
