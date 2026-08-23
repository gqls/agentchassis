-- 576 — required-fields-missing-handler: the `tomb` CTE that 574 added may not register a
--       retirement on an EMPTY-slot match. CONFIG ONLY — live on apply.
--
--       Acting on a MEDIUM advisory objection from the council round that APPROVED 574
--       (correlation d48c0a89-9ff8-4286-bfe9-2690dc13d5bc, round 1, 2026-08-23). The seat
--       was right, and it named the consequence precisely: a false positive here produces
--       "an incorrect `stale` close — exactly the outcome this migration exists to stop."
--
-- ============================================================================
-- THE EDGE, in plain terms
-- ============================================================================
-- 574 gave the router a `tomb` CTE: "is there a build_status='removed' row at this
-- (page, slot)?" A yes is the POSITIVE evidence that lets the router close a finding as
-- `stale`, on the reasoning that the component was deliberately retired.
--
-- It matches the slot the same way the `comp` CTE does, which is 410's original shape:
--
--     COALESCE(pc2.slot_name, '') = COALESCE(item.spec->>'slot_name', '')
--
-- If BOTH sides are absent, that compares `'' = ''` and matches. So an item carrying no
-- `slot_name` could be "retired" by an unrelated removed row that also carries no
-- `slot_name`, anywhere on the same page — and be CLOSED on it.
--
-- Inherited from `comp`, where the same comparison is merely a failed lookup. In `tomb`
-- it is the difference between parking a finding and closing it, which is the whole
-- subject of bugs_open/367.
--
-- ============================================================================
-- MEASURED FIRST — it is NOT reachable today, and that is why this is a guard, not a fix
-- ============================================================================
-- [MEASURED 2026-08-23, live] both halves of the trap are empty, and either one alone
-- would be enough to prevent it:
--
--   * page_components rows with an empty/NULL slot_name, ANY build_status:  0
--   * of the 38 build_status='removed' rows, those with an empty slot_name: 0
--   * required_fields_missing items with no spec.slot_name:                 2
--     — and BOTH also lack spec.page_name, so they classify `malformed` first and never
--       reach the stale arm at all (0 items have a page_name but no slot_name).
--
-- So nothing is broken in production and no row's disposition changes. What this buys is
-- that the bad state stops depending on a population that happens to be empty. The
-- estate's own rule is to make it unrepresentable rather than merely unpopulated, and the
-- guard is one clause.
--
-- ============================================================================
-- THE CHANGE
-- ============================================================================
-- `tomb` additionally requires the ITEM to name a slot. An item with no slot cannot now
-- find a tombstone, so it PARKS rather than closing — the safe direction, and the same
-- direction 574 chose everywhere else.
--
-- Deliberately NOT changed: the `comp` CTE keeps 410's comparison. Narrowing it would
-- change how findings RESOLVE for every producer, which is a behaviour change well beyond
-- this objection, and `comp` failing to match is already the safe outcome (it parks).
--
-- ROLLBACK: 576_tomb_cte_cannot_match_on_an_empty_slot_ROLLBACK.sql
-- ============================================================================

BEGIN;

DO $$
DECLARE
    v_row_count int;
    v_q         text;
    v_new_q     text;
    c_old CONSTANT text :=
        'AND COALESCE(pc2.slot_name, '''') = COALESCE(item.spec->>''slot_name'', '''')) AS retired)';
    c_new CONSTANT text :=
        'AND COALESCE(pc2.slot_name, '''') = COALESCE(item.spec->>''slot_name'', '''') '
        'AND COALESCE(item.spec->>''slot_name'', '''') <> '''') AS retired)';
BEGIN
    SELECT count(*) INTO v_row_count FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_row_count <> 1 THEN
        RAISE EXCEPTION '576: expected exactly ONE active non-snapshot row, found % — resolve before patching', v_row_count;
    END IF;

    SELECT default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query'
      INTO v_q FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('tomb AS (SELECT EXISTS' in v_q) = 0 THEN
        RAISE EXCEPTION '576: 574 is not applied (no tomb CTE) — apply 574 first';
    END IF;
    IF position(c_new in v_q) > 0 THEN
        RAISE EXCEPTION '576: already applied';
    END IF;
    IF (length(v_q) - length(replace(v_q, c_old, ''))) / length(c_old) <> 1 THEN
        RAISE EXCEPTION '576: the tomb slot-match anchor does not occur exactly once — another lane has edited this agent. Re-read the live row and re-anchor.';
    END IF;

    PERFORM snapshot_agent('required-fields-missing-handler'::text,
                           '576: guard the tomb CTE against an empty-slot match (council objection on d48c0a89)'::text);

    v_new_q := replace(v_q, c_old, c_new);
    IF v_new_q = v_q THEN
        RAISE EXCEPTION '576: the rewrite was a no-op';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
               '{workflow,steps,classify,config,query}', to_jsonb(v_new_q), false)
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

DO $$
DECLARE v_q text; v_steps jsonb; v_n int;
BEGIN
    SELECT default_config -> 'workflow' -> 'steps',
           default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query'
      INTO v_steps, v_q FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('AND COALESCE(item.spec->>''slot_name'', '''') <> ''''' in v_q) = 0 THEN
        RAISE EXCEPTION '576 VERIFY: the empty-slot guard is absent';
    END IF;

    -- NEGATIVE CONTROLS: 574's work must all still be there. A replace() that ate the CTE
    -- would satisfy the check above by deleting the thing it guards.
    IF position('tomb AS (SELECT EXISTS' in v_q) = 0
    OR position('target_not_dispatchable' in v_q) = 0
    OR position('AS target_state' in v_q) = 0
    OR position('COALESCE(pc.build_status, ''pending'') <> ''removed''' in v_q) = 0
    OR position('plan_src AS (' in v_q) = 0
    OR position('has_open_tool_recreation' in v_q) = 0 THEN
        RAISE EXCEPTION '576 VERIFY: 574''s changes or a pre-existing CTE was lost';
    END IF;
    IF position('pc.build_status = ''deployed''' in v_q) > 0 THEN
        RAISE EXCEPTION '576 VERIFY: the deployed-only predicate has reappeared — 410 was re-applied over 574';
    END IF;

    SELECT count(*) INTO v_n FROM jsonb_object_keys(v_steps);
    IF v_n <> 22 THEN
        RAISE EXCEPTION '576 VERIFY: expected 22 steps, found % — the route wiring has changed under this migration', v_n;
    END IF;
    IF v_steps #>> '{route_resolved,config,else_step}'         IS DISTINCT FROM 'route_not_dispatchable'
    OR v_steps #>> '{route_not_dispatchable,config,then_step}' IS DISTINCT FROM 'park_not_dispatchable'
    OR v_steps #>> '{route_not_dispatchable,config,else_step}' IS DISTINCT FROM 'route_owned'
    OR v_steps #>> '{park_not_dispatchable,config,status}'     IS DISTINCT FROM 'needs_human_review' THEN
        RAISE EXCEPTION '576 VERIFY: 574''s route wiring is not intact';
    END IF;

    RAISE NOTICE '576 OK: the tomb CTE now requires the item to name a slot; 574 intact, 22 steps.';
END $$;

COMMIT;
