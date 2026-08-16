-- 417 — image-build-handler: asset_id? becomes asset_id! (STRICT) on
-- call_asset_deployer — RFC_029 §9 D3's named first adopter
--
-- ⚠⚠ HOLD: ORDERING-CRITICAL — DO NOT APPLY UNTIL THE CHASSIS IMAGE CARRYING
-- THE `!` MARKER PARSER (RFC_029 Phase 1, this migration's own commit) HAS
-- ROLLED. The runner's SIDECAR_RE excludes this file from --apply on purpose;
-- apply it BY HAND after verifying the binary:
--
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the RFC_029 Phase 1 commit> <the stamped sha>
--
-- WHY THE ORDER IS LOAD-BEARING (both failure shapes are silent or fleet-loud):
-- an older binary reads "asset_id!" as an ORDINARY destination field named
-- "asset_id!". If the path resolves, the child receives a key named "asset_id!"
-- and NO "asset_id" — so its own resolver falls back to the whole-tree search,
-- which is the EXACT regression this marker exists to prevent, reintroduced by
-- the migration that was meant to forbid it. If the path does not resolve, the
-- un-suffixed field is REQUIRED and hard-fails the step on the store_asset
-- refusal branches that migration 401 deliberately kept non-fatal.
--
-- WHAT THIS CHANGES (RFC_029 §9 D3, owner-delegated ruling 2026-08-15): the
-- ruling names asset_id on the two 401/402 callers as the `!` marker's first
-- adopter, "converting the two shipped fixes from current-state accidents into
-- enforced invariants — a regression that re-exposes asset_id to the search
-- becomes a loud failure instead of a silent guess."
--
-- MEASURED before writing this file (2026-08-15, live DB): 13 of 13
-- asset-deployer children of image-build-handler parents in the retained
-- orchestration_states window (2026-08-14, post-401) carry asset_id in their
-- resolved input_data; ZERO refusal-branch spawns (locked site / no asset URL —
-- the branches 401's own text kept non-fatal via `?`) were observed. After this
-- flip a refusal-branch spawn hard-fails call_asset_deployer LOUDLY instead of
-- spawning a deployer with no asset. That is a real behaviour change on a
-- measured-zero branch, and it is the intended one: per the owner's stated
-- ranking (RFC_029 §9), a loud absence beats a silent guess.
--
-- ⚠ THE SECOND NAMED ADOPTER IS *NOT* FLIPPED, AND MUST NOT BE. The ruling also
-- names build-dispatch-loop's repair-path mapping (migration 402), but that
-- mapping is SHARED BY EVERY DISPATCHED ITEM TYPE — 402's own text measures
-- exactly one (item_type, handler_agent) pair fleet-wide carrying spec.asset_id
-- against 636+ item types flowing through the same call_handler step. Its `?`
-- is doing per-item-type optionality; a strict marker there would hard-fail
-- every non-asset dispatch in the fleet. Recorded as a dated correction on
-- RFC_029 §9 D3 (same commit as this file); the repair path keeps `?` and is
-- covered by Phase 1/2 of the search change instead.
--
-- Idempotent: fenced on the `asset_id?` key still being present; a replay
-- matches no row. DB-only; snapshot-prefixed.
--
-- TWO PRE-APPLY CHECKS, MEASURED (council run ae2a88a7 on RFC_029 Phase 1,
-- debug_historian seat, 2026-08-15; both re-run 2026-08-16 ~10:30Z):
--   (1) The two-active-rows-per-agent-type trap (LANDMINES: "any UPDATE
--       agent_definitions … WHERE type = <x>" — four types carry TWO active
--       rows and only the higher version loads) is NOT APPLICABLE here:
--       image-build-handler has exactly ONE active, non-snapshot, non-deleted
--       row (version 1). Re-check before applying — the WHERE below hits every
--       active row, so a second one would be patched too:
--         SELECT count(*), array_agg(version) FROM agent_definitions
--          WHERE type='image-build-handler' AND is_active
--            AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;   -- expect 1
--   (2) The snapshot_agent call below is the TWO-ARG overload, and
--       pg_get_functiondef confirms it writes to agent_definitions_backup
--       (the one-arg form writes an is_snapshot row into agent_definitions —
--       LANDMINES "snapshot_agent has TWO overloads"). So the pre-image for the
--       ROLLBACK lives in agent_definitions_backup, and the check that matters
--       is that it holds the PRE-change key, not merely that a row exists:
--         SELECT snapshot_taken_at, snapshot_reason,
--                default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}' ? 'asset_id?' AS has_old
--           FROM agent_definitions_backup WHERE type='image-build-handler'
--          ORDER BY snapshot_taken_at DESC LIMIT 1;                           -- expect has_old = t
--       Ledger (schema_migrations, keyed by FILENAME) checked 2026-08-16: this
--       file is unclaimed; the other 417 (brief_fidelity_auditor) is applied and
--       does not collide because the ledger keys on the full filename.

BEGIN;

SELECT snapshot_agent('image-build-handler',
    '417_image_build_handler_asset_id_goes_strict_HOLD.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config #- '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id?}',
         '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id!}',
         COALESCE(
           default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id?}',
           '"asset_stored.asset_id"'::jsonb
         ),
         true
       ),
       updated_at = now()
 WHERE type = 'image-build-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}' ? 'asset_id?';

-- Verify: exactly one live row, carrying asset_id! and not asset_id?
DO $$
DECLARE mapping jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}' INTO mapping
    FROM agent_definitions
    WHERE type = 'image-build-handler'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF mapping IS NULL THEN
        RAISE EXCEPTION '417: no live image-build-handler config at {...call_asset_deployer,config,input_mapping}';
    END IF;
    IF NOT (mapping ? 'asset_id!') THEN
        RAISE EXCEPTION '417: asset_id! missing after update — mapping is %', mapping;
    END IF;
    IF mapping ? 'asset_id?' THEN
        RAISE EXCEPTION '417: asset_id? still present after update — both spellings would leave the strict one authoritative in the binary but the intent ambiguous to every reader. mapping is %', mapping;
    END IF;
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 417: image-build-handler's asset_id mapping goes STRICT (`!`) — RFC_029 §9 D3's first adopter
The `!` marker (RFC_029 Phase 1) means: explicit resolution only; if the source path does not resolve, the call_asset_deployer step fails loudly; the field is never left absent for the child asset-deployer's whole-tree search to guess at. Happy path unchanged (13/13 recent spawns resolve asset_stored.asset_id). The store_asset refusal branches (locked / no asset URL), which migration 401 kept non-fatal with `?`, now fail the step loudly — measured zero occurrences in the retained window, and a loud absence beats a silent guess per the owner's ranking.
NOT adopted on build-dispatch-loop's repair mapping (the ruling's other named adopter): that mapping is shared by 636+ item types and its `?` is per-item-type optionality — a strict marker there would hard-fail every non-asset dispatch. See RFC_029 §9's dated correction.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'hand-applied-after-roll');

COMMIT;
