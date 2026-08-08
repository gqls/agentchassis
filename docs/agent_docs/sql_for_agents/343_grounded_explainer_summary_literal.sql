-- ============================================================================
-- 343_grounded_explainer_summary_literal.sql
--
-- bugs_open/136 §4/§5.2 — the one key that was BITING, resolved by OWNER
-- DECISION 2026-08-08: option A + D (static literal + repair the rows), option
-- C (implement template rendering in create_work_item) explicitly declined.
--
-- grounded-explainer's create_review_item step set `summary_template`, a key
-- create_work_item has never read. The action falls back summary → config
-- summary → item_type (create_work_item_action.go:133-139), so both live
-- grounded_draft_review items reached the human-review queue captioned with
-- their own item_type. The fallback is what hid it: a loud empty caption would
-- have been noticed on day one.
--
-- Why a LITERAL and not template rendering (the owner's call, with the reasons
-- recorded): all ten other live create_work_item steps use a static summary;
-- rendering would be a shared-action capability with ONE consumer, which has
-- executed ZERO orchestrations; and a template naming a missing key fails on
-- the path to a human. If interpolated captions ever have real demand, build
-- it then, on evidence.
--
-- Why the topic is NOT in the repaired captions: it is unrecoverable. The same
-- step's `spec_fields` is also a dead key (bugs_open/136 §3), so both rows'
-- spec is EMPTY — nothing was ever captured to interpolate. The generic
-- caption is the truth of what the rows can say.
--
-- Two parts:
--   1. definition: drop summary_template, set summary literal  (A)
--   2. data: recaption the two existing rows                    (D)
--
-- The repo seed (224_grounded_explainer_agent.sql:183) is corrected in the
-- same commit — a replayed seed would silently reintroduce the dead key
-- (the bugs_open/134 lesson: fix the seed as well as the live row).
--
-- Live immediately (DB config; no image roll involved).
-- Verify (both must hold):
--   SELECT default_config->'workflow'->'steps'->'create_review_item'->'config'
--     FROM agent_definitions WHERE type='grounded-explainer' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- has "summary", no "summary_template"
--   SELECT summary FROM site_work_items WHERE item_type='grounded_draft_review';
--   -- no row equals the item_type
-- ============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- Guard: the step still carries the dead key. If another session has already
-- fixed or restructured it, stop rather than blindly re-shaping their config.
-- ---------------------------------------------------------------------------
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'create_review_item'->'config'
    INTO cfg
    FROM agent_definitions
   WHERE type='grounded-explainer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '343: grounded-explainer create_review_item config not found — definition moved, do not apply blind';
  END IF;
  IF NOT cfg ? 'summary_template' THEN
    RAISE EXCEPTION '343: summary_template already absent — another session got here first; verify and skip';
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 1 (A). Definition: summary literal in, summary_template out.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,create_review_item,config}',
         (default_config->'workflow'->'steps'->'create_review_item'->'config')
           - 'summary_template'
           || jsonb_build_object('summary', 'Grounded explainer draft ready for review'),
         false),
       updated_at = now()
 WHERE type='grounded-explainer' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 2 (D). Data: recaption the two rows born under the fallback.
-- Keyed on summary = item_type so a row a human has since retitled is left
-- alone; scoped by item_type so nothing else can match.
-- ---------------------------------------------------------------------------
UPDATE site_work_items
   SET summary = 'Grounded explainer draft ready for review',
       updated_at = now()
 WHERE item_type = 'grounded_draft_review'
   AND summary = 'grounded_draft_review';

-- ---------------------------------------------------------------------------
-- Verify inside the transaction, loudly — a SELECT cannot stop a COMMIT
-- (the migration-runner landmine), so both checks RAISE.
-- ---------------------------------------------------------------------------
DO $$
DECLARE cfg jsonb; bad int;
BEGIN
  SELECT default_config->'workflow'->'steps'->'create_review_item'->'config'
    INTO cfg
    FROM agent_definitions
   WHERE type='grounded-explainer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg ? 'summary_template' OR NOT cfg ? 'summary' THEN
    RAISE EXCEPTION '343: definition edit did not land as intended: %', cfg;
  END IF;
  SELECT count(*) INTO bad FROM site_work_items
   WHERE item_type='grounded_draft_review' AND summary=item_type;
  IF bad > 0 THEN
    RAISE EXCEPTION '343: % row(s) still captioned with their item_type', bad;
  END IF;
END $$;

COMMIT;
