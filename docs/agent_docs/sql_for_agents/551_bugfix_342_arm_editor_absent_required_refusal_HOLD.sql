-- 551_bugfix_342_arm_editor_absent_required_refusal_HOLD.sql
--
-- bugs_open/342 — arm the REFUSAL half on section-editor's apply_edit step:
-- an edit whose render left a schema-required source:"llm" field EMPTY is
-- refused at the ONE persist switch, and the live section keeps what it had.
-- The required_fields_missing item is filed BEFORE the refusal fires (the
-- emit is inside the render branches), so a refused edit still leaves a queue
-- entry saying why.
--
-- ⚠ _HOLD: ORDERING-CRITICAL — THE RUNNER MUST NOT TAKE THIS (the 502 shape).
-- The Go half (refuseAbsentRequiredFields / refusePersistForAbsentRequired in
-- mistyped_llm_fields_gate.go, the gate in ApplySectionEditAction) is committed
-- with this file but INERT until a chassis image built from that commit has
-- rolled. A config key naming behaviour the running binary does not have is a
-- no-op that LOOKS applied — the worst of both states.
--
-- APPLY ONLY AFTER the running chassis carries the Go half:
--   1. Find the running stamp (the startup log line scrolls on agent-chassis —
--      go to the binary):
--        kubectl get pods -n ai-persona-system -l app=agent-chassis \
--          -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image --no-headers
--        kubectl -n ai-persona-system exec <pod> -- sh -c 'grep -aq "<candidate-sha>" /proc/1/exe && echo PRESENT'
--        (run a nonsense-string control in the same breath)
--   2. git merge-base --is-ancestor <the commit that ships this file> <stamp>
--   3. Probe BOTH replicas for the new refusal literal:
--        grep -aq "refusing to persist" /proc/1/exe        # positive
--        (plus the nonsense control)
--   Then drop the _HOLD suffix in the same commit that records the apply.
--
-- ⚠ INTERACTION, STATED (bugs_open/344): section-editor's apply_edit step
-- declares no error_step, and the dispatch loop's completion can trample a
-- failure flag — so the DRIVING work item of a refused edit may still read
-- `complete` until 344 lands. The refusal's protection does not depend on
-- that: the live page is untouched either way, and the
-- required_fields_missing item is already on the queue. Do NOT "fix" this by
-- adding an error_step here without reading 344's candidates first — its
-- candidate 1 (completion refuses a future retry_after) is the contained one.
--
-- WHY OPT-IN (owner ruling 2026-08-02 §2): an edit that leaves a required
-- field empty SUCCEEDS today — the blank section ships and page assembly
-- drops it. Refusing is new authority over content that currently persists,
-- so the unsafe side is the default and THIS migration is the visible act of
-- switching it on for the one live consumer.
--
-- CANARY AFTER APPLY (the bug file's verification): one edit against a
-- component with a required source:"llm" field deliberately absent must
-- refuse — item filed, live section byte-identical; one clean edit must
-- persist (positive control — an arm that only stops edits is not a fix).

SELECT snapshot_agent('section-editor', 'migration 551: pre-update (bugs_open/342 editor refusal arm)');

BEGIN;

-- ── Pre-conditions ──────────────────────────────────────────────────────────
DO $$
DECLARE
    live_rows integer;
    step_action text;
    existing  text;
BEGIN
    SELECT count(*) INTO live_rows
    FROM agent_definitions
    WHERE type='section-editor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF live_rows IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION 'MIGRATION 551: expected exactly 1 live section-editor row, found % — resolve first.', live_rows;
    END IF;

    -- The step must exist and carry the action this key gates; jsonb_set on a
    -- wrong path would otherwise invent config for a step nothing runs.
    SELECT default_config#>>'{workflow,steps,apply_edit,action}' INTO step_action
    FROM agent_definitions
    WHERE type='section-editor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF step_action IS DISTINCT FROM 'apply_section_edit' THEN
        RAISE EXCEPTION 'MIGRATION 551: steps.apply_edit.action is % not apply_section_edit — the workflow has moved. Re-derive.', COALESCE(step_action,'ABSENT');
    END IF;

    -- Double-apply refusal.
    SELECT default_config#>>'{workflow,steps,apply_edit,config,refuse_absent_required_fields}' INTO existing
    FROM agent_definitions
    WHERE type='section-editor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF existing IS NOT NULL THEN
        RAISE EXCEPTION 'MIGRATION 551: refuse_absent_required_fields already present (%) — refusing to double-apply.', existing;
    END IF;
END $$;

-- ── The write ───────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,apply_edit,config,refuse_absent_required_fields}',
        'true'::jsonb),
    version    = version + 1,
    updated_at = now()
WHERE type='section-editor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── Post-conditions ─────────────────────────────────────────────────────────
DO $$
DECLARE
    armed text;
    strip text;
BEGIN
    SELECT default_config#>>'{workflow,steps,apply_edit,config,refuse_absent_required_fields}',
           default_config#>>'{workflow,steps,apply_edit,config,strip_literal_markdown}'
    INTO armed, strip
    FROM agent_definitions
    WHERE type='section-editor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF armed IS DISTINCT FROM 'true' THEN
        RAISE EXCEPTION 'MIGRATION 551: read back % for refuse_absent_required_fields, expected true.', COALESCE(armed,'ABSENT');
    END IF;
    -- Sibling-key sanity: the write must not have displaced the existing step
    -- config (jsonb_set replaces the LEAF, but a wrong path replaces a BRANCH).
    IF strip IS NULL THEN
        RAISE EXCEPTION 'MIGRATION 551: strip_literal_markdown vanished from apply_edit config — the jsonb_set path clobbered the config object. Restore from the snapshot.';
    END IF;
END $$;

COMMIT;
