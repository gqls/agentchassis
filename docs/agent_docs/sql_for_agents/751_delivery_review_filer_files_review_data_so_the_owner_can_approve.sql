-- 751: delivery-review-filer must file `review_data`, or the owner cannot approve.
--
-- bugs_open/474. Found on the FIRST run of delivery-review-filer ever (the
-- owner-authorised rehearsal on idea.uk, 2026-09-03) — before the owner clicked,
-- not after.
--
-- THE DEFECT. The admin screen derives everything it submits from the item's
-- `spec.review_data`:
--   App.tsx:519   editedReviewData is set only if spec.checkpoint && spec.review_data
--   App.tsx:762   handleApprove returns early: "No review data to approve"
--   App.tsx:1295  but the button SHOWS on `isCheckpoint` alone
-- so the owner gets a live-looking "Approve & Continue" that does nothing, with no
-- error naming what is missing or whose fault it is. HandleApproveWorkItem would
-- refuse it a layer further back too (ReviewData is binding:"required").
--
-- WHY ONLY THIS PRODUCER. Every other checkpoint item is filed by
-- `checkpoint_for_review`, which always writes review_data. delivery-review-filer
-- is the one that files a checkpoint through `create_work_item`, which does not.
-- [MEASURED 2026-09-03, all history] across every row with spec->>'checkpoint'
-- = 'true', exactly ONE lacks review_data: the single needs_delivery_review row.
-- A second producer with the same gap would have appeared in that census.
--
-- WHY `review_data` POINTS AT `input_data`, having read the resolver rather than
-- guessed at it. create_work_item builds the spec in three layers
-- (create_work_item_action.go:266-305): `spec_data` copies a map's entries in
-- FLAT; `spec_paths` resolves each value and assigns `specMap[key] = val`, also
-- FLAT; `spec_literal` is verbatim constants with no resolution. So none of the
-- three can ASSEMBLE a nested object out of separate input fields — but
-- `spec_paths` can point a key at a map that already exists, because
-- ExtractNestedField returns whatever is at the path, map included.
-- `input_data` IS that map, and it holds exactly site_id, domain, site_url and
-- brief — which is precisely what the owner is being asked to look at. So this is
-- one key, no dispatcher has to remember anything, and there is no placeholder.
--
-- The rejected alternative was `spec_paths: {"review_data": "input_data.review_data"}`
-- with every dispatcher composing the object. It has one merit — an unresolved
-- spec_path is a HARD ERROR, so a forgetful dispatcher would fail loudly at file
-- time instead of silently producing an unapprovable item — but it moves work onto
-- every future caller to fix a defect none of them caused.
--
-- ⚠ NO ROLL NEEDED. Agent config: live the moment it applies. 474's candidate 2
-- (make the button's visibility condition match its action, in App.tsx) is a
-- frontend change, is NOT in this file, and is not required for the owner to
-- approve.
--
-- ROLLBACK: 751_..._ROLLBACK.sql.

BEGIN;

-- ── snapshots ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS bak_751_delivery_review_filer_20260903 AS
SELECT * FROM agent_definitions
WHERE type = 'delivery-review-filer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS bak_751_review_items_20260903 AS
SELECT * FROM site_work_items WHERE item_type = 'needs_delivery_review';

-- ── anchor guard: refuse if the step is not the shape this was written against ─
DO $$
DECLARE n int; cfg jsonb;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '751: expected exactly 1 live delivery-review-filer row, found %', n;
  END IF;

  SELECT default_config->'workflow'->'steps'->'file_review'->'config' INTO cfg
    FROM agent_definitions WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cfg IS NULL THEN
    RAISE EXCEPTION '751: file_review.config is absent — the step moved';
  END IF;
  IF cfg->>'item_type' <> 'needs_delivery_review' THEN
    RAISE EXCEPTION '751: file_review files %, expected needs_delivery_review', cfg->>'item_type';
  END IF;
  IF (cfg->'spec_paths') ? 'review_data' THEN
    RAISE EXCEPTION '751: review_data already configured — already applied, or another session got here first';
  END IF;
  IF COALESCE(cfg->'spec_literal'->>'checkpoint','') <> 'true' THEN
    RAISE EXCEPTION '751: spec_literal no longer carries checkpoint:true — the approve handler 400s without it';
  END IF;
END $$;

-- ── the change: one spec_path, pointing at the map the filer was given ──────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,file_review,config,spec_paths}',
      (default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths')
        || jsonb_build_object('review_data', 'input_data'),
      false),
    updated_at = NOW()
WHERE type='delivery-review-filer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── unblock the item already filed, from its OWN spec ───────────────────────
-- Nothing invented: built from the three fields the row already carries, which is
-- what the config above will now produce by itself.
UPDATE site_work_items
SET spec = spec || jsonb_build_object('review_data', jsonb_build_object(
             'domain',   spec->>'domain',
             'site_url', spec->>'site_url',
             'brief',    spec->>'brief')),
    updated_at = NOW()
WHERE item_type = 'needs_delivery_review'
  AND status = 'needs_human_review'
  AND NOT (spec ? 'review_data');

-- ── verify: DO/RAISE, because ON_ERROR_STOP will NOT abort a COMMIT on a
--    SELECT that merely returns the wrong rows ─────────────────────────────────
DO $$
DECLARE bad int; patched int; cfgpath text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths'->>'review_data'
    INTO cfgpath
    FROM agent_definitions WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfgpath IS DISTINCT FROM 'input_data' THEN
    RAISE EXCEPTION '751 VERIFY: spec_paths.review_data = %, expected input_data', cfgpath;
  END IF;

  SELECT count(*) INTO bad FROM site_work_items
   WHERE item_type='needs_delivery_review' AND status='needs_human_review'
     AND NOT (spec ? 'review_data');
  IF bad <> 0 THEN
    RAISE EXCEPTION '751 VERIFY: % open delivery review item(s) still lack review_data', bad;
  END IF;

  SELECT count(*) INTO patched FROM site_work_items
   WHERE item_type='needs_delivery_review'
     AND spec->'review_data'->>'site_url' = spec->>'site_url'
     AND spec->'review_data'->>'site_url' IS NOT NULL;
  IF patched < 1 THEN
    RAISE EXCEPTION '751 VERIFY: no item carries a review_data mirroring its own site_url';
  END IF;

  -- the thing that must be UNDISTURBED: without checkpoint:true the approve
  -- handler 400s, and its error steers the owner to RESOLVE, which never opens
  -- the gate (platform/delivery/prepare.go, ReviewItemRequiredSpec).
  IF EXISTS (SELECT 1 FROM site_work_items
              WHERE item_type='needs_delivery_review'
                AND COALESCE(spec->>'checkpoint','') <> 'true') THEN
    RAISE EXCEPTION '751 VERIFY: a delivery review item lost checkpoint:true';
  END IF;
  IF EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='delivery-review-filer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                AND COALESCE(default_config->'workflow'->'steps'->'file_review'
                             ->'config'->'spec_literal'->>'checkpoint','') <> 'true') THEN
    RAISE EXCEPTION '751 VERIFY: the filer stopped stamping checkpoint:true';
  END IF;

  RAISE NOTICE '751: filer now files review_data from input_data; % open item(s) patched from their own spec', patched;
END $$;

COMMIT;
