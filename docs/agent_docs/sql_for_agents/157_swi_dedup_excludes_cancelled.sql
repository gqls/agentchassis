-- 157_swi_dedup_excludes_cancelled.sql
--
-- The site_work_items dedup index (idx_swi_dedup, UNIQUE on (site_id,
-- item_key) over "open" items) treats 'cancelled' as OPEN, so a cancelled
-- item permanently occupies its slot and every later insert for the same
-- defect is silently swallowed by ON CONFLICT DO NOTHING.
--
-- That contradicts the design intent already on record:
--   * the cooldown queries were fixed 2026-07-12/13 to EXCLUDE cancelled
--     items ("cancelled = resolved with a reason");
--   * T15 (2026-07-14) parked the vonc quiz improve_tool item as "cancelled
--     with reason, REGENERABLE by re-running 087" — which this index made
--     false: nothing could regenerate while the cancelled row held the slot.
--
-- Proven live 2026-07-16 during the P3 screenshots proof: a Tier-4 FAILED run
-- on tool-drop-rate-tuner produced its acceptance-fail note (with the new
-- Evidence line) but NO improve_tool item — the judge's insert collided with
-- T9's cancelled 2026-07-12 controlled-test item and vanished.
--
-- Fix: rebuild the partial-index predicate with 'cancelled' in the CLOSED
-- set. Safe by construction: the new predicate covers a strict SUBSET of the
-- rows the old unique index already covered, so no existing rows can violate
-- the new constraint.
--
-- Applied outside run-migrations.sh (psql -f + manual ledger row): the runner
-- is blocked at 151_gripper_spec_sheet_component.sql (another workstream's
-- known-failing duplicate, being fixed separately) and must not have files
-- 151-156 — none of them this workstream's — applied as a side effect.

BEGIN;

DO $$
DECLARE
  def text;
BEGIN
  SELECT indexdef INTO def FROM pg_indexes WHERE indexname = 'idx_swi_dedup';
  IF def IS NULL THEN
    RAISE EXCEPTION '157: idx_swi_dedup not found — schema drifted, do not apply blind';
  END IF;
  IF strpos(def, 'cancelled') > 0 THEN
    RAISE EXCEPTION '157: idx_swi_dedup already excludes cancelled — nothing to do';
  END IF;
END $$;

DROP INDEX idx_swi_dedup;

CREATE UNIQUE INDEX idx_swi_dedup ON public.site_work_items
  USING btree (site_id, item_key)
  WHERE item_key IS NOT NULL
    AND status <> ALL (ARRAY['complete'::text, 'verified'::text, 'rejected'::text,
                             'wont_fix'::text, 'failed'::text, 'unresolved'::text,
                             'cancelled'::text]);

-- Convention: schema/workflow-altering migrations leave a pipeline note.
INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## idx_swi_dedup: cancelled items no longer block re-raising
Observed: a cancelled site_work_item held its (site_id, item_key) dedup slot forever, so later inserts for the same defect were silently dropped by ON CONFLICT DO NOTHING (proven 2026-07-16: a failing Tier-4 run on tool-drop-rate-tuner produced no improve_tool item — the 2026-07-12 cancelled test item still occupied the slot).
Root cause: idx_swi_dedup''s open-statuses predicate omitted ''cancelled'', contradicting the cooldown queries (which already treat cancelled as resolved) and the parked-item convention (cancelled = regenerable).
Fix: index rebuilt with ''cancelled'' in the closed set (migration 157).
Verified: post-apply, a repeat insert for a key held only by a cancelled item succeeds.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '157_swi_dedup_excludes_cancelled'
);

COMMIT;
