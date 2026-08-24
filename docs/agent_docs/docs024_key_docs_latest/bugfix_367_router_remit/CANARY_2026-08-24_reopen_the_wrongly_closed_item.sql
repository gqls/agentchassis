-- CANARY (bugs_open/367) — re-open the ONE item this bug wrongly closed, so the new
-- disposition is observed on a real item in production rather than only at the route.
--
-- WHY THIS IS A REPAIR, not just a test. Item 562788c3 was closed `complete` on
-- 2026-08-23 17:09Z by the defect: the router could not resolve a non-deployed component
-- and reported it as gone. The finding is STILL TRUE — re-checked 2026-08-24: component
-- 0a1498b3 (page ai-agent-roi-estimator, slot tool-cta) has build_status='pending',
-- locked_at NULL, 9,220 chars of rendered_html, and BOTH named fields absent from
-- content_data, untouched since 2026-07-17. So the row is a false negative sitting in the
-- "actioned" bucket, and re-opening it is the correct disposal of the damage.
--
-- SAFE TO RUN because, re-checked in the same session:
--   * 0 non-terminal rows hold that item_key (a `complete` status is excluded from
--     idx_swi_dedup, so the key released when the wrong close happened) — no collision;
--   * no other session has open required_fields_missing work on that page (the 7 open
--     items there are other types, none routed at this handler);
--   * migrations 574 + 576 are applied and the VERIFY sidecar passes on the live row,
--     so the router will now PARK this rather than close it.
--
-- EXPECTED OUTCOME: dispatch within ~120s → classify → route `target_not_dispatchable`,
-- target_state `pending` → park_not_dispatchable → status `needs_human_review`, with the
-- repair paths in `error`. Assert at the ROUTE in orchestration_states, not at the status.
--
-- attempt_count = 0 is LOAD-BEARING: the claim gate requires attempt_count < max_attempts.

BEGIN;

DO $$
DECLARE n int; v_status text; v_html int;
BEGIN
    -- The finding must still be true, or re-opening it files noise.
    SELECT length(COALESCE(rendered_html,'')) INTO v_html FROM page_components
     WHERE id = '0a1498b3-a066-4d50-8a9f-f97b281830a1'
       AND COALESCE(build_status,'pending') <> 'removed'
       AND locked_at IS NULL
       AND (content_data->>'headline') IS NULL
       AND (content_data->>'trust_note') IS NULL;
    IF v_html IS NULL THEN
        RAISE EXCEPTION 'CANARY: the finding is no longer true (component repaired, locked or retired) — do NOT re-open the item; close this canary instead';
    END IF;

    -- The dedup key must be free, or the UPDATE violates idx_swi_dedup.
    SELECT count(*) INTO n FROM site_work_items
     WHERE item_key = 'required_fields_missing:f438eca6-e37d-4f55-b5ae-0bd909e2dc92:tool-cta'
       AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
    IF n <> 0 THEN
        RAISE EXCEPTION 'CANARY: % non-terminal row(s) already hold that dedup key — another producer has re-raised it; watch that row instead', n;
    END IF;

    SELECT status INTO v_status FROM site_work_items WHERE id = '562788c3-c9e9-4e8b-9967-c16dc9b8ed36';
    IF v_status IS DISTINCT FROM 'complete' THEN
        RAISE EXCEPTION 'CANARY: item is %, not complete — it has already been re-opened or reaped', COALESCE(v_status,'MISSING');
    END IF;

    UPDATE site_work_items
       SET status = 'triaged', attempt_count = 0, updated_at = NOW(),
           completed_at = NULL,
           result = COALESCE(result,'{}'::jsonb) || jsonb_build_object(
               'reopened_by', 'bugs_open/367 canary 2026-08-24',
               'reopened_because', 'closed complete on 2026-08-23 17:09Z by the 367 defect; the finding is still true and migrations 574+576 now park it instead')
     WHERE id = '562788c3-c9e9-4e8b-9967-c16dc9b8ed36';

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN RAISE EXCEPTION 'CANARY: expected to update exactly 1 row, updated %', n; END IF;
    RAISE NOTICE 'CANARY OK: 562788c3 re-opened at triaged/attempt_count 0. Expect target_not_dispatchable within ~120s.';
END $$;

COMMIT;
