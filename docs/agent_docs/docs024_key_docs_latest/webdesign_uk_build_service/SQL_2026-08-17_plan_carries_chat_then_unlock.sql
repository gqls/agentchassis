-- 2026-08-17 — owner: "The chat box lock can come off."
--
-- DONE IN THE SAFE ORDER, because unlocking ALONE would have deleted it.
-- Measured before writing this: the CURRENT plan (site_plans 6a3e6d1b,
-- is_current) holds contact = [hero, contact-info] ONLY. The chat box survives
-- today purely because it is a LOCKED row that the bugs_open/285 fix merges
-- into the assembled list (that run recorded locked_merge_count=1). Remove the
-- lock with the plan unchanged and the next rebuild assembles [hero,
-- contact-info], finds no locked row to merge, and save_page_sections removes
-- the section with no lock left to block it — the exact 2026-08-11 deletion,
-- repeated.
--
-- So: put the chat box in the PLAN first (structural protection, and the shape
-- the owner asked for — "through the framework, spec and planner"), THEN clear
-- the lock. Net effect: the section is carried by design rather than by a lock,
-- which is strictly better than the state before.
--
-- Reversible: DELETE the plan row and re-set locked_at to restore today's state.

BEGIN;

-- 1. the plan carries the chat box (idempotent)
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
SELECT sp.id, 'contact', 2, 'chat-input-box'
  FROM site_plans sp JOIN sites s ON s.id = sp.site_id
 WHERE s.domain = 'webdesign.uk' AND sp.is_current
   AND NOT EXISTS (
     SELECT 1 FROM site_plan_sections x
      WHERE x.plan_id = sp.id AND x.page_name = 'contact' AND x.component_name = 'chat-input-box');

-- 2. only now, release the lock (the guard keys on locked_at)
UPDATE page_components pc
   SET locked_at = NULL, lock_type = NULL, locked_by = NULL, updated_at = NOW()
  FROM pages p, sites s
 WHERE pc.page_id = p.id AND p.site_id = s.id
   AND s.domain = 'webdesign.uk' AND p.name = 'contact' AND pc.slot_name = 'chat-input-box';

-- 3. verify, DO/RAISE so a failure aborts the COMMIT
DO $$
DECLARE n_plan int; n_locked int;
BEGIN
  SELECT count(*) INTO n_plan
    FROM site_plan_sections sps JOIN site_plans sp ON sp.id = sps.plan_id JOIN sites s ON s.id = sp.site_id
   WHERE s.domain='webdesign.uk' AND sp.is_current AND sps.page_name='contact' AND sps.component_name='chat-input-box';
  IF n_plan <> 1 THEN
    RAISE EXCEPTION 'plan does not carry chat-input-box exactly once (got %) — refusing to leave it unlocked AND unplanned', n_plan;
  END IF;

  SELECT count(*) INTO n_locked
    FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE s.domain='webdesign.uk' AND p.name='contact' AND pc.slot_name='chat-input-box' AND pc.locked_at IS NOT NULL;
  IF n_locked <> 0 THEN RAISE EXCEPTION 'lock not released'; END IF;
END $$;

COMMIT;
