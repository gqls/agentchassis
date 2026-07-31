-- 278 — build-dispatch-loop reads the work item's component_id COLUMN, not just the spec copy
--
-- bugs_open/154. The second half of the fix; the first half is Go
-- (LoadWorkItemsAction / setRoutingField, committed 2026-07-31).
-- Council-Reviewed: 10be5ed9-3bd0-45ed-b6bb-4385a887967d (APPROVED, 6 advisory
-- objections, none high). THIS FILE IS THE REVISED VERSION: four of those six
-- objections were about this migration and are answered inline below, marked
-- [SEAT]. Answering them cost nothing because it had not been applied yet.
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
-- The Go half exposes those columns top-level on the item map, resolved COLUMN
-- FIRST with a spec.<key> fallback. So after the image ships,
-- current_item.component_id is a STRICT SUPERSET of current_item.spec.component_id
-- — it resolves everywhere the old path did, plus the 16 rows it did not. That
-- superset property is the whole safety argument for this migration.
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
--
-- [SEAT editquality, medium] "the jsonb_set path assumes a specific depth …
-- same shape as a known landmine". Right to ask, and it was assumed rather than
-- proven when submitted. VERIFIED read-only against the live row 2026-07-31:
--   SELECT default_config #>> '{workflow,steps,process_item,config,sub_workflow,
--          steps,call_handler,config,input_mapping,component_id?}' …
--   → 'current_item.spec.component_id'
-- The path resolves at that exact depth. Step 1 below re-asserts it at APPLY
-- time rather than trusting this comment, because the row is shared and mutable.
--
-- [SEAT tooling_provenance, medium] "mutates default_config in place with a raw
-- UPDATE; the platform has snapshot_agent() for exactly this". Correct, and the
-- reuse rule applies — CONFIRMED to exist:
--   snapshot_agent(p_agent_type text) / snapshot_agent(p_agent_type text, p_reason text)
-- Now used below, replacing the hand-rolled "SELECT the old value and keep it in
-- your scrollback" step the first draft had. A snapshot in a table beats a
-- pre-image in a terminal nobody still has open.
--
-- [SEAT debug_historian, medium+low] "needle-gate SQL implements two of five
-- disciplines: no dump/backup before the UPDATE, and no read-only pre-flight
-- count asserting the occurrence count immediately before it runs (not a census
-- taken earlier in the investigation)". Both added as steps 1 and 2.
--
-- [SEAT guardian, medium] "labelled operation:add for a new SQL file rather than
-- config_change naming the owning pipeline". Noted for the submission metadata;
-- recorded here so the file itself states its surface: this is a CONFIG_CHANGE
-- to the owning pipeline **build-dispatch-loop**, delivered as a migration.
--
-- ── WHAT THIS DOES *NOT* CLOSE (raised by [SEAT bug_historian, medium], and the
--    objection was sharper than the answer I would have given) ──
-- The Go half exposes THREE columns; this migration rewires ONE mapping. So the
-- class is closed for component_id and NOT for its siblings. Measured 2026-07-31:
--   entity_id   column set on 0 rows; 1 agent reads it — asset-deployer, via
--               input_data.spec.entity_id, i.e. through the `spec` passthrough,
--               NOT through a dispatcher mapping. build-dispatch-loop maps no
--               entity_id at all.
--   affected_url column set on 0 rows; 0 agents read it anywhere.
-- So nothing is broken through them today — there is no failing population — but
-- a future creator that writes entity_id on the COLUMN hits the identical bug,
-- because asset-deployer reads the spec path directly and the Go coalesce cannot
-- reach a handler that never goes through a mapping. Deliberately NOT fixed here:
-- it would be a two-part speculative change (add an entity_id? mapping AND repoint
-- asset-deployer) against zero failing rows. When the first such row appears, that
-- is the fix. Recorded in LANDMINES.md and register WDS-014 so it is findable then.

BEGIN;

-- ── STEP 1 — PRE-FLIGHT ASSERTION [SEAT debug_historian] ──────────────────
-- A count derived mechanically HERE, at apply time, not carried in from the
-- investigation. Expect exactly 1. Anything else means the row is not the shape
-- this migration was written against — stop and re-read it.
SELECT count(*) AS rows_to_change_expect_1
FROM agent_definitions
WHERE type = 'build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,component_id?}'
      = 'current_item.spec.component_id';

-- ── STEP 2 — SNAPSHOT [SEAT tooling_provenance] ───────────────────────────
-- The platform's own versioning mechanism, not a hand-rolled pre-image.
SELECT snapshot_agent('build-dispatch-loop',
                      'bugs_open/154 — component_id? mapping to the first-class column (278)');

-- ── STEP 3 — THE CHANGE ───────────────────────────────────────────────────
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

-- ── STEP 4 — VERIFY BEFORE COMMIT ─────────────────────────────────────────
-- Expect exactly one row reading current_item.component_id. If it still reads
-- the spec path, the WHERE guard refused the row: ROLLBACK and re-read step 1.
SELECT type,
       default_config #>> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,component_id?}' AS input_mapping_after
FROM agent_definitions
WHERE type = 'build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ── ROLLBACK ──
-- Preferred: restore the snapshot taken in step 2 (that is what it is for).
-- Direct inverse, if the snapshot path is unavailable:
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
