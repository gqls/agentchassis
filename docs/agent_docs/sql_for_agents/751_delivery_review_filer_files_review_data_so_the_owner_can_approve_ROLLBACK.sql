-- ROLLBACK for 751 (bugs_open/474).
--
-- ⚠ THIS RESTORES A STATE IN WHICH THE OWNER CANNOT APPROVE A DELIVERY REVIEW.
-- The item renders, the "Approve & Continue" button shows, and pressing it does
-- nothing but print "No review data to approve". Use this only to reach a known
-- state while something else is diagnosed — never as a repair.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '751 ROLLBACK: expected exactly 1 live delivery-review-filer row, found %', n;
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,file_review,config,spec_paths}',
      (default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths') - 'review_data',
      false),
    updated_at = NOW()
WHERE type='delivery-review-filer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Only the rows 751 patched: an item whose review_data mirrors its own spec.
-- An item filed by the FIXED config carries the same shape, so this is scoped by
-- status as well — a completed/approved review is left alone.
UPDATE site_work_items
SET spec = spec - 'review_data', updated_at = NOW()
WHERE item_type='needs_delivery_review'
  AND status='needs_human_review'
  AND spec->'review_data'->>'site_url' = spec->>'site_url';

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'file_review'->'config' INTO cfg
    FROM agent_definitions WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF (cfg->'spec_paths') ? 'review_data' THEN
    RAISE EXCEPTION '751 ROLLBACK VERIFY: spec_paths.review_data still present: %', cfg->'spec_paths';
  END IF;
  IF COALESCE(cfg->'spec_literal'->>'checkpoint','') <> 'true' THEN
    RAISE EXCEPTION '751 ROLLBACK VERIFY: checkpoint:true was disturbed — the item would 400 on approve';
  END IF;
  RAISE NOTICE '751 ROLLBACK: review_data removed from the filer and from open items';
END $$;

COMMIT;
