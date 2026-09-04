-- ROLLBACK for 752 (bugs_open/474). Restores 751's shape:
-- spec_paths.review_data = 'input_data' (the whole map).
--
-- ⚠ That shape is the one the council objected to — unconstrained, and giving the
-- backfilled item and future items two different shapes for one field. It is not
-- broken (the owner can still approve), so this rollback is SAFE to run; it just
-- reinstates a design the round found worse. Use it to reach a known state while
-- something else is diagnosed.

BEGIN;

DO $$
DECLARE sp jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths'
    INTO sp FROM agent_definitions WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF sp IS NULL OR jsonb_typeof(sp) <> 'object' THEN
    RAISE EXCEPTION '752 ROLLBACK: spec_paths absent or not an object (%)', jsonb_typeof(sp);
  END IF;
  IF sp->>'review_data' IS DISTINCT FROM 'input_data.review_data' THEN
    RAISE EXCEPTION '752 ROLLBACK: review_data is %, not 752''s value — nothing to roll back', sp->>'review_data';
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,file_review,config,spec_paths}',
      COALESCE(default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths', '{}'::jsonb)
        || jsonb_build_object('review_data', 'input_data'),
      false),
    updated_at = NOW()
WHERE type='delivery-review-filer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE sp jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths'
    INTO sp FROM agent_definitions WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF sp->>'review_data' IS DISTINCT FROM 'input_data' THEN
    RAISE EXCEPTION '752 ROLLBACK VERIFY: review_data = %, expected input_data', sp->>'review_data';
  END IF;
  IF NOT (sp ? 'brief' AND sp ? 'domain' AND sp ? 'site_url') THEN
    RAISE EXCEPTION '752 ROLLBACK VERIFY: spec_paths lost a sibling key: %', sp;
  END IF;
  RAISE NOTICE '752 ROLLBACK: review_data restored to input_data (751''s shape)';
END $$;

COMMIT;
