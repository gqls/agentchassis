-- D14 pilot (2026-07-17): supersede the 3 LIVE articles' D13-era photographic
-- content heroes so the next content_image_missing pass regenerates them as
-- kind=content_hero (Banana, flat duotone override — style-guide row 361f2ed7).
-- Regeneration recipe per the I3 gate-fix handoff: the check only GENERATES
-- when no active content hero exists, so supersede → generate → land →
-- flag_rebuild re-renders the article → next pass re-derives the card
-- (origin_asset_id mismatch). Pilot = the 3 deployed articles only (the other
-- 6 rows are excluded by the F2.1 eligibility filter and are R6's
-- build-or-retire decision).
--
-- Site: robot-hands.com 00ff3af5-dad8-4770-9f70-3edc267a3c92

UPDATE assets
   SET status = 'superseded',
       updated_at = now()
 WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
   AND asset_key IN (
       'content_hero_tool_grip_force_friction_calculator_guide',
       'content_hero_tool_gripper_payload_calculator_guide',
       'content_hero_tool_gripper_cycle_time_estimator_guide')
   AND status = 'active';
