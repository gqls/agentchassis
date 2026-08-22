-- 551_bugfix_342_arm_editor_absent_required_refusal.sql
--
-- ✅ APPLIED BY HAND 2026-08-22 ~18:5xZ, AFTER the v1.0.1326 roll, and the _HOLD
-- suffix was dropped in the same commit to say so. The precondition it was held
-- for was met and checked AT THE ARTEFACT, not at the tag: both replicas carry
-- the config key `refuse_absent_required_fields` and the chrome literal
-- "REQUIRED content field(s) absent — refusing to store", each of which had
-- ZERO occurrences in the tree before the commits that introduced them (checked
-- with `git grep` at the parent commit, which is what makes them valid positive
-- controls); two negative controls behaved (a nonsense literal absent, and the
-- deleted regex-fallback literal absent).
-- ⚠ A THIRD PROBE ARM WAS INVALID AND IS RECORDED RATHER THAN QUIETLY DROPPED:
-- "refusing to persist" was ALREADY present in three other files before this
-- change, so it would have read PRESENT whatever shipped. See WRONG_CALLS.md,
-- 2026-08-22 — a probe arm is only a control if the literal is new.
--
-- Verified independently of this file's own post-check: the key reads true, BOTH
-- sibling keys (strip_literal_markdown, allow_rendered_html_transform) survived
-- the jsonb_set, version 1→2, no other agent gained the key, and the CHROME
-- refusal is still armed nowhere (its deliberate state — see 550).
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
-- CANARY AFTER APPLY (the bug file's verification, sharpened by council
-- 3626629a round 1, bug_historian): one edit against a component with a
-- required source:"llm" field deliberately absent must refuse — and check
-- THREE things, not one:
--   (a) the live section is byte-identical (md5 the rendered_html before and
--       after; this is the protection itself);
--   (b) a required_fields_missing item exists naming those fields (the record
--       must survive the refusal — refusing must never be why a defect goes
--       unrecorded);
--   (c) the DRIVING work item's terminal status is READ, not assumed benign.
--       Expect it to read `complete` until bugs_open/344 lands — the dispatch
--       loop's mark_complete tramples a failure flag — and record what it
--       actually said. The queryable fingerprint 344 gives for the trample is
--       `retry_after > completed_at` on a `complete` row.
-- Positive control in the same run: a clean edit must still persist. An arm
-- that only stops edits is not a fix.
--
-- OPTIONAL-KEY BUDGET (RFC_022, N=10; architecture seat's advisory on the same
-- round): apply_section_edit declares 7 OPTIONAL inputs and this key is a
-- CONFIG key, which `cmd/config-key-audit --optional-key-budget` does not
-- count — so the action stays at 7 of 10 and no accumulated-surface review is
-- owed. Measured 2026-08-22.

SELECT snapshot_agent('section-editor', 'migration 551: pre-update (bugs_open/342 editor refusal arm)');

BEGIN;

-- ── Pre-conditions ──────────────────────────────────────────────────────────
DO $$
DECLARE
    live_rows integer;
    step_action text;
    existing  text;
BEGIN
    -- ⚠ THE DUPLICATE-ACTIVE-ROW LANDMINE, CHECKED RATHER THAN ASSERTED
    -- (council 3626629a round 1, prior_art_librarian, GATING/high): four agent
    -- types carry TWO active definition rows and only the higher version is
    -- ever loaded, so an UPDATE-by-type can silently write the row nobody
    -- reads. Measured 2026-08-22: the four are chief-strategist,
    -- content-creator, content-creator-contact and site-component-architect —
    -- section-editor is NOT among them and has exactly ONE live row. This
    -- guard is what keeps that true at apply time rather than at review time.
    -- Its correct response to a duplicate is to ABORT, not to write both:
    -- which row the loader picks is a question this migration must not guess.
    SELECT count(*) INTO live_rows
    FROM agent_definitions
    WHERE type='section-editor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF live_rows IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION 'MIGRATION 551: expected exactly 1 live section-editor row, found % — section-editor has joined the duplicate-active-row set (chief-strategist, content-creator, content-creator-contact, site-component-architect as at 2026-08-22). Only the higher version loads, so writing both would arm a row nobody reads. Resolve the duplication first.', live_rows;
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

-- ── The write, with the row-count assertion BAKED IN ────────────────────────
-- Council 3626629a round 2, debug_historian: needle-gate discipline wants the
-- "exactly one live row" expectation asserted by the UPDATE itself, not only
-- by a precondition that ran beforehand — between the two, another session can
-- insert a second active row, and a WHERE-clause precondition cannot see that.
-- GET DIAGNOSTICS makes the write its own witness.
DO $$
DECLARE
    n_updated integer;
BEGIN
    UPDATE agent_definitions
    SET default_config = jsonb_set(default_config,
            '{workflow,steps,apply_edit,config,refuse_absent_required_fields}',
            'true'::jsonb),
        version    = version + 1,
        updated_at = now()
    WHERE type='section-editor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    GET DIAGNOSTICS n_updated = ROW_COUNT;
    IF n_updated IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION 'MIGRATION 551: the UPDATE touched % rows, expected exactly 1 — a second active section-editor row appeared between the precondition and the write, and only the higher version loads. Aborting rather than arming a row nobody reads.', n_updated;
    END IF;
END $$;

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
