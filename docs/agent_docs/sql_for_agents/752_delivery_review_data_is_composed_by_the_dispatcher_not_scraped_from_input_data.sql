-- 752: narrow delivery-review-filer's `review_data` from "the whole input_data
-- map" to an object the DISPATCHER composes.
--
-- bugs_open/474. This supersedes the shape 751 shipped, acting on the council's
-- REVISE of that migration (correlation 95eaf57c-74d4-46f3-bad3-f8ef46bc1df8).
-- Forward-only: 751 stands in history and its backfill was correct.
--
-- ── WHAT 751 GOT WRONG, and both objections were right ─────────────────────
--
-- 751 set `spec_paths.review_data = 'input_data'`, pointing the key at the whole
-- map the filer happens to be given.
--
--  (a) UNCONSTRAINED SHAPE (guardian, medium). That copies whatever else
--      input_data carries — orchestration bookkeeping, envelope fields, anything
--      a future step adds — verbatim into `spec.review_data`, which the admin
--      screen renders DIRECTLY to the owner. A future unrelated change to this
--      agent's input_data would silently change what he is shown, with nothing
--      reviewing that coupling. Today it also drags in `site_id`, a uuid he has
--      no use for.
--
--  (b) TWO SHAPES FOR ONE FIELD (guardian, low — and the sharper form of (a)).
--      751's backfill of the already-filed row built {domain, site_url, brief}
--      from that row's own spec, while its config change would give every FUTURE
--      row {site_id, domain, site_url, brief}. One field, two shapes, diverging
--      silently at the boundary between the patched item and the next one.
--
-- ── THE FIX, and why the loud-failure property is worth the cost ────────────
--
-- `spec_paths.review_data = 'input_data.review_data'`: the dispatcher composes
-- the object explicitly, and it is exactly {domain, site_url, brief} — the same
-- three fields 751 backfilled, so the two shapes become one.
--
-- The cost is that every dispatcher must now supply it. That is the whole point.
-- create_work_item treats a configured spec_path that does not resolve as a HARD
-- ERROR (create_work_item_action.go:287-296, "refusing to create a work item with
-- an incomplete spec"), so a dispatcher that forgets gets a loud refusal AT
-- DISPATCH, in front of the operator who just ran it — instead of silently filing
-- an item the owner cannot approve and will not understand. Trading a silent
-- dead-end for a loud failure is the trade this bug exists to make.
--
-- I rejected this shape when writing 751, on the grounds that it "moves work onto
-- every future caller to fix a defect none of them caused". The council's (a) and
-- (b) change that balance: an unconstrained payload rendered to the owner is a
-- worse defect than three fields a caller must name, and the RUNBOOK recipe is
-- the only caller today.
--
-- ── TWO ROBUSTNESS GAPS IN 751, ALSO FROM THE ROUND, CLOSED HERE ────────────
--
--   * `jsonb_set(..., create_missing => false)` returns the document UNCHANGED if
--     the path is absent — a silent no-op reporting success (editquality, high).
--     751 survived only because its DO/RAISE verify would have caught it. Here
--     the anchor guard asserts spec_paths EXISTS and is an object, before the
--     write rather than after.
--   * `jsonb || NULL` is NULL, so concatenating onto an absent spec_paths would
--     WIPE it rather than add a key (debug_historian, medium). COALESCE closes it
--     by construction.
--
-- ── AND THE PROCESS POINT, recorded rather than argued away ─────────────────
--
-- debug_historian objected (high) that 751 was APPLIED BY HAND BEFORE the council
-- saw it, so the seats audited a change already live. That is accurate. The
-- reason was that the owner was blocked on an unapprovable review item and the
-- fix was one config key; the mitigation was a snapshot, an anchor guard, a
-- DO/RAISE verify and a committed ROLLBACK file — which the submission failed to
-- list as an edit, so the seats could not see it. **The reviewer could only judge
-- what was in front of them, and what was in front of them was a bare jsonb_set.**
-- This file is submitted BEFORE it is applied.
--
-- ⚠ NO ROLL NEEDED — agent config, live on apply. Requires the RUNBOOK dispatch
-- recipe to carry `review_data` in input_data; that edit ships in the same commit.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_752_delivery_review_filer_20260903 AS
SELECT * FROM agent_definitions
WHERE type = 'delivery-review-filer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── anchor guard ────────────────────────────────────────────────────────────
DO $$
DECLARE n int; sp jsonb;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '752: expected exactly 1 live delivery-review-filer row, found %', n;
  END IF;

  SELECT default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths'
    INTO sp
    FROM agent_definitions WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  -- The gap editquality named: jsonb_set with create_missing=false is a silent
  -- no-op on an absent path, so assert the path exists BEFORE writing to it.
  IF sp IS NULL OR jsonb_typeof(sp) <> 'object' THEN
    RAISE EXCEPTION '752: spec_paths is absent or not an object (%) — jsonb_set would no-op silently', jsonb_typeof(sp);
  END IF;
  IF sp->>'review_data' IS DISTINCT FROM 'input_data' THEN
    RAISE EXCEPTION '752: spec_paths.review_data is %, expected 751''s value ''input_data'' — 751 not applied, or already superseded', sp->>'review_data';
  END IF;
END $$;

-- ── the change ──────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,file_review,config,spec_paths}',
      COALESCE(default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths', '{}'::jsonb)
        || jsonb_build_object('review_data', 'input_data.review_data'),
      false),
    updated_at = NOW()
WHERE type='delivery-review-filer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── verify ──────────────────────────────────────────────────────────────────
DO $$
DECLARE sp jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'file_review'->'config'->'spec_paths'
    INTO sp FROM agent_definitions WHERE type='delivery-review-filer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF sp->>'review_data' IS DISTINCT FROM 'input_data.review_data' THEN
    RAISE EXCEPTION '752 VERIFY: spec_paths.review_data = %, expected input_data.review_data', sp->>'review_data';
  END IF;
  -- the three that must be UNDISTURBED — the || could have replaced the object
  IF NOT (sp ? 'brief' AND sp ? 'domain' AND sp ? 'site_url') THEN
    RAISE EXCEPTION '752 VERIFY: spec_paths lost a sibling key: %', sp;
  END IF;
  IF EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='delivery-review-filer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                AND COALESCE(default_config->'workflow'->'steps'->'file_review'
                             ->'config'->'spec_literal'->>'checkpoint','') <> 'true') THEN
    RAISE EXCEPTION '752 VERIFY: the filer stopped stamping checkpoint:true — the approve handler 400s without it';
  END IF;

  -- The already-filed item keeps the shape 751 gave it, and that shape is now
  -- the SAME one future items get. Assert the convergence rather than assume it.
  IF EXISTS (SELECT 1 FROM site_work_items
              WHERE item_type='needs_delivery_review'
                AND (SELECT count(*) FROM jsonb_object_keys(spec->'review_data')) <> 3) THEN
    RAISE EXCEPTION '752 VERIFY: an existing review item does not carry exactly the 3 fields the dispatcher now composes';
  END IF;

  RAISE NOTICE '752: review_data is now composed by the dispatcher (domain, site_url, brief) — one shape, and a forgetful dispatcher fails LOUDLY at file time';
END $$;

COMMIT;
