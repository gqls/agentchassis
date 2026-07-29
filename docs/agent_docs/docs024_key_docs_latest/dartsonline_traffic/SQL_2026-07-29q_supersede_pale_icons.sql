-- SQL_2026-07-29q_supersede_pale_icons.sql
--
-- Make the corrected prompts actually produce anything.
--
-- SQL ...p fixed the seven icon prompts so they ask for light linework on a dark
-- navy ground instead of grey on #EEEEEE. On its own that changes nothing,
-- because the mechanism that raises generation work only fires for a plan row
-- with NO active asset:
--
--   check_unfulfilled_imagery_plan.go:67 -> hasActiveAssetForAssetKey()
--   imagery_helpers.go:96-104:
--       SELECT COUNT(*) FROM assets
--        WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
--
-- Every one of those seven keys already has an active asset — a correct
-- rendering of the old, wrong prompt. So the corrected instruction would sit
-- there unread indefinitely. **A fixed prompt is not a regenerated image.**
-- Superseding the stale assets is what turns one into the other.
--
-- WHAT IS SUPERSEDED, AND WHY IT IS SAFE
-- Only assets whose PROMPT specified the pale ground, and only icons. Measured
-- before writing: the served homepage carries exactly one <img> (the logo) and
-- one CSS background-image (hero-home.jpg), and all 17 generated icons are
-- undeployed — they are the 17 `undeployed_asset` items sitting 'detected'
-- since 2026-07-28. So nothing on the live site references any of them and
-- superseding breaks no page.
--
-- Deliberately NOT touched:
--   hero_home        in use on the live homepage, and it is the reference the
--                    new style guide anchors to. Superseding the anchor would
--                    be self-defeating.
--   the other heroes  their prompts were rewritten in ...p, but a hero is
--                    expensive and three of them are fine as they are. Their
--                    corrected prompts apply to the next regeneration, whenever
--                    that is chosen deliberately.
--   the 8 orphan icons (icon_accessories, icon_brands, icon_dartboard,
--                    icon_fast_shipping, icon_shipping, icon_soft_tip,
--                    icon_spec_complete, icon_specialist) — they belong to no
--                    current plan row, so no amount of superseding regenerates
--                    them. They are dead weight, not a defect to fix here.
--
-- WHAT HAPPENS NEXT, and what to check. The next discovery pass raises
-- `needs_imagery` items for the seven superseded keys plus `icon_independent`
-- (new key, never had an asset). Those land at status='detected'. That is the
-- bugs_open/083 queue where things go to die, so they will need promoting to
-- 'triaged' the same way the page items were — per site, as data.
--
-- **READ EVERY GENERATED PNG BEFORE WIRING IT.** A green status is not a
-- picture. The whole reason these seven need doing again is that seventeen
-- images were generated, marked active, and are unusable — and nobody looked
-- until today.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

CREATE TABLE IF NOT EXISTS bak_20260729q_assets AS
SELECT * FROM assets WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381';

UPDATE assets SET status = 'superseded', updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND status = 'active'
  AND purpose = 'icon'
  AND asset_key IN (
    'icon_all_brands',
    'icon_darts_only',
    'icon_expert_guides',
    'icon_free_shipping',      -- key deleted from the plan in ...p; the asset
                               -- would otherwise linger as an active orphan
    'icon_player_knowledge',
    'icon_setup_builder',
    'icon_specialist_range'
  );

COMMIT;

-- ── verification ───────────────────────────────────────────────────────────
--   SELECT asset_key, status FROM assets
--   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND purpose='icon'
--   ORDER BY status, asset_key;
--
-- Then, after the next discovery pass, the items it should have raised —
-- 7 superseded keys are 6 live plan rows (icon_free_shipping was deleted from
-- the plan) plus icon_independent, so expect 7 items, not 8:
--   SELECT item_key, status FROM site_work_items
--   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--     AND item_type='needs_imagery' ORDER BY item_key;
--
-- If that returns nothing after a pass has run, do NOT assume the check is
-- broken: confirm a discovery pass actually ran for this site first
-- (site_work_items.batch_id / the agent's own log), because "no items" and
-- "no run" are indistinguishable from the queue alone.
