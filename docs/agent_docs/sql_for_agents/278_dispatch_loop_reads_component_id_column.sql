-- 278 — build-dispatch-loop reads the work item's component_id COLUMN, not just the spec copy
--
-- bugs_open/154. The second half of the fix; the first half is Go
-- (LoadWorkItemsAction / setRoutingField, committed 2026-07-31).
--
-- WHY. site_work_items carries component_id, entity_id and affected_url as real
-- columns, but LoadWorkItemsAction built current_item from a SELECT listing only
-- page_id among them. So the only path a dispatcher could reference was
-- current_item.spec.component_id — a copy the creating agent had to remember to
-- duplicate into the spec JSONB. tool-auditor populated the column and not the
-- blob, so ALL FOUR of its improve_tool items were structurally undispatchable:
-- the optional "component_id?" mapping resolved nothing, ResolveInputMapping
-- silently skips an unresolved optional path, and tool-improver's load_tool then
-- hard-errored on `input_data.component_id resolved to nil`. Items from three
-- other creators (16 of 16, all spec-only) ran clean. The creator that used the
-- schema properly was the one whose items could not be routed.
--
-- The Go half now exposes those columns top-level on the item map, resolved
-- COLUMN FIRST with a spec.<key> fallback, so this single path is correct for
-- BOTH populations — the 4 column-only rows and the 235 spec-only ones. The
-- coalesce has to live in Go because input_mapping resolves exactly one source
-- path per destination and has no coalesce syntax.
--
-- ══════════════════════════════════════════════════════════════════════════
-- ORDER IS LOAD-BEARING. DO NOT APPLY THIS UNTIL THE IMAGE IS LIVE.
-- ══════════════════════════════════════════════════════════════════════════
-- The coalesce is in the BINARY; this path is in the DB, which is live
-- immediately. Applying this against a chassis that predates the Go change
-- strands the 235 spec-only rows — current_item.component_id would not exist
-- and every one of them would lose the value that works today. That is a
-- strictly worse outage than the bug being fixed.
--
-- Gate on the RUNNING POD, on every replica, never on the tag and never on git
-- (a roll is not evidence your fix shipped — bugs_open/153). The positive
-- control in the same exec is what makes a 0 meaningful:
--
--   kubectl exec -n ai-persona-system <pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c "routing field left unset"; \
--      strings /app/agent-chassis | grep -c "LoadWorkItemsAction: Starting"'
--
--   first  = the symbol THIS change added        → must be >= 1
--   second = a pre-existing symbol (the control) → if this is 0 the grep itself
--            is broken and the first number means nothing
--
-- ── the key really does carry a question mark ──
-- "component_id?" — the ? is the optional marker and is part of the JSON key.
-- Distinct from bugs_open/134, where a ? leaked into a `params` key, where it
-- means nothing and silently makes the key inert. In input_mapping it is real:
-- ResolveInputMapping trims the suffix and treats the field as skippable.
-- Getting this wrong writes a SECOND key and leaves the original in place.

BEGIN;

-- Snapshot the current mapping before touching it, so a rollback is exact
-- rather than reconstructed from memory.
SELECT type,
       default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
                     ->'steps'->'call_handler'->'config'->'input_mapping' AS input_mapping_before
FROM agent_definitions
WHERE type = 'build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,component_id?}',
      '"current_item.component_id"'::jsonb,
      false   -- create_missing=false: if the key is absent the mapping is not
              -- the shape this migration was written against; fail loud rather
              -- than inventing a key next to whatever is really there.
    ),
    updated_at = now()
WHERE type = 'build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,component_id?}'
      = 'current_item.spec.component_id';   -- idempotent + refuses a drifted row

-- VERIFY BEFORE COMMIT. Expect exactly one row reading current_item.component_id.
SELECT type,
       default_config #>> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,component_id?}' AS input_mapping_after
FROM agent_definitions
WHERE type = 'build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ── ROLLBACK (same shape, inverted) ──
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(default_config,
--       '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,component_id?}',
--       '"current_item.spec.component_id"'::jsonb, false)
-- WHERE type='build-dispatch-loop'
--   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── VERIFY THE FIX AT THE ARTEFACT, not at this row ──
-- Dispatch a tool-auditor-raised improve_tool item and assert it gets PAST
-- load_tool. Retention trap (bugs_open/154's own warning, and it bit the fixing
-- session): orchestration_states history is short — rows for a run fired at
-- 18:14 were gone by ~18:40. Capture this when it happens; do not plan to query
-- it back later. site_work_items and diagnosis_artifacts outlive it.
--
--   SELECT current_step, status FROM orchestration_states
--   WHERE owner_agent_type = 'tool-improver' ORDER BY created_at DESC LIMIT 3;
--
--   SELECT left(id::text,8), status, created_by, error
--   FROM site_work_items
--   WHERE item_type='improve_tool' AND created_by='tool-auditor'
--   ORDER BY created_at DESC LIMIT 5;
