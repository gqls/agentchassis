-- 169_tool_improver_refusal_path.sql — when a component write is refused, fail honestly.
--
-- Companion to the Go component write guard (bugs_open/012 fix candidate (a)).
--
-- The guard in platform/orchestration/actions/component_write_guard.go now
-- REFUSES to overwrite content_components.html_template with a structurally
-- collapsed replacement, and returns an error so the component is left intact.
-- That alone prevents the destruction. It does not, by itself, produce an
-- honest outcome: tool-improver's `update_component` step had NO error_step, so
-- a refusal would run to routeToErrorStepOrFail -> failWorkflow. The
-- orchestration dies, the work item is left for the reaper, no note is written,
-- and the next thread reads a generic FAILED with no idea a fix was rejected as
-- mangled.
--
-- bugs_open/012 candidate (a) and 016b §9 both spell out the required outcome:
-- "leave the row untouched, fail the item honestly (needs_human_review), write
-- a NOTE recording the refusal; never a silent success." This migration wires
-- the second and third of those.
--
-- Route: update_component --(error)--> refuse_mangled_write --> note_refusal --> complete
--
--   refuse_mangled_write  fail_work_item with status_override=needs_human_review.
--                         FailWorkItemAction prefers the real error stored by
--                         routeToErrorStep at __step_error.message over the
--                         config literal, so the item records exactly which
--                         structural check refused the write.
--   note_refusal          append_doc_note against the tool's travelling NOTES,
--                         body taken from the same __step_error.message. No LLM
--                         in the refusal path — a deterministic record of why.
--
-- Both new steps fall forward to the next stage on their own error, so a
-- failure to log can never swallow the refusal itself.
--
-- Ordering note: this is DB config and is LIVE IMMEDIATELY, whereas the Go
-- guard is inert until a chassis image ships. Applying this first is safe and
-- inert in the other direction too — the new steps are only ever reached by an
-- error_step transition that nothing currently triggers.

BEGIN;

SELECT snapshot_agent('tool-improver', '169_tool_improver_refusal_path: pre-update');

DO $$
DECLARE
  n int := 0;
BEGIN
  -- Idempotency gate. Re-running this migration must be a 0-row no-op, not a
  -- silent re-apply of the same jsonb_set (needle-gate discipline; raised by
  -- the council gate's debug-historian seat).
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type='tool-improver' AND is_active
    AND default_config #>> '{workflow,steps,update_component,error_step}' = 'refuse_mangled_write';
  IF n > 0 THEN
    RAISE NOTICE '169: already applied (% row(s) already route update_component to refuse_mangled_write) — no-op', n;
    RETURN;
  END IF;

  -- Pre-condition: the step we are attaching an error route to must exist and
  -- must not already have a DIFFERENT one (another thread may have wired this
  -- its own way — in that case stop and let a human reconcile).
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type='tool-improver' AND is_active
    AND default_config #> '{workflow,steps,update_component}' IS NOT NULL
    AND default_config #>> '{workflow,steps,update_component,error_step}' IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '169: expected exactly 1 active tool-improver with update_component and no error_step, found % — reconcile by hand', n;
  END IF;

  UPDATE agent_definitions
  SET default_config = jsonb_set(
        jsonb_set(
          jsonb_set(
            default_config,
            '{workflow,steps,update_component,error_step}',
            '"refuse_mangled_write"'::jsonb,
            true),
          '{workflow,steps,refuse_mangled_write}',
          $json${
            "action": "fail_work_item",
            "description": "The component write guard refused a structurally-collapsed rewrite. Fail the item for a human instead of reporting success.",
            "config": {
              "work_item_id": "input_data.work_item_id",
              "status_override": "needs_human_review",
              "error_message": "tool-improver produced a structurally-collapsed component; the write was refused and the component left untouched"
            },
            "output_field": "refusal",
            "next_step": "note_refusal",
            "error_step": "note_refusal"
          }$json$::jsonb,
          true),
        '{workflow,steps,note_refusal}',
        $json${
          "action": "append_doc_note",
          "description": "Record the refusal on the tool's travelling NOTES so the next run sees that a fix was rejected as mangled, not merely that something failed.",
          "config": {
            "subject_type": "tool",
            "subject_key_field": "tool_data.function",
            "note_body_field": "__step_error.message",
            "note_categories": ["fix", "refused", "llm-truncation"],
            "note_source": "tool-improver",
            "note_site_id_field": "input_data.site_id",
            "created_by": "component-write-guard"
          },
          "output_field": "refusal_note",
          "next_step": "complete",
          "error_step": "complete"
        }$json$::jsonb,
        true)
  WHERE type='tool-improver' AND is_active
    -- Gated on the pre-state, so the UPDATE itself is a no-op on re-run
    -- rather than relying on the guard above alone.
    AND default_config #>> '{workflow,steps,update_component,error_step}' IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '169: expected to update 1 tool-improver row, updated %', n;
  END IF;

  -- Verify all three writes landed on the paths they were aimed at. A jsonb_set
  -- that misses its path is a silent no-op, which is precisely the class of bug
  -- this migration exists to make impossible downstream.
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type='tool-improver' AND is_active
    AND default_config #>> '{workflow,steps,update_component,error_step}' = 'refuse_mangled_write'
    AND default_config #>> '{workflow,steps,refuse_mangled_write,action}' = 'fail_work_item'
    AND default_config #>> '{workflow,steps,refuse_mangled_write,config,status_override}' = 'needs_human_review'
    AND default_config #>> '{workflow,steps,note_refusal,action}' = 'append_doc_note'
    AND default_config #>> '{workflow,steps,note_refusal,config,note_body_field}' = '__step_error.message';
  IF n <> 1 THEN
    RAISE EXCEPTION '169: post-update verification failed, matched % rows', n;
  END IF;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## tool-improver now fails honestly when a component write is refused
Context: the Go component write guard (bugs_open/012 candidate (a)) refuses to overwrite content_components.html_template with a structurally-collapsed replacement and returns an error, leaving the component intact.
Gap this closes: update_component had no error_step, so a refusal reached failWorkflow — orchestration FAILED, work item left to the reaper, no note. A thread reading that would see a generic failure, not "a fix was rejected as mangled".
Change: update_component.error_step -> refuse_mangled_write (fail_work_item, status_override=needs_human_review) -> note_refusal (append_doc_note on the tool''s NOTES) -> complete. Both new steps fall forward on their own error, so a logging failure cannot swallow the refusal.
Body/error text for both comes from __step_error.message, which routeToErrorStep populates with the guard''s actual reason — no LLM in the refusal path.
Snapshot taken (snapshot_agent tool-improver). Migration 169.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '169_tool_improver_refusal_path'
);

COMMIT;
