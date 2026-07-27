-- ============================================================================
-- 238_experience_register_writer_agent.sql — the owning pipeline for the
-- experience register's write path.
--
-- WHY THIS EXISTS, AND WHY IT IS A SEPARATE, NAMED THING
--   `write_experience_pattern` shipped in v1.0.1175 with no caller. The council
--   that reviewed it (corr 2e71f640, APPROVED) asked for exactly this, twice:
--
--     guardian:  "Risk #2 states nothing currently calls write_experience_pattern
--                 — no workflow step names it. That's good for containment now,
--                 but it means the owning pipeline for this surface is still
--                 undetermined. When a workflow-JSON step is later added to
--                 invoke this action, that edit must come back as operation
--                 config_change naming the owning pipeline explicitly — flagging
--                 so it isn't waved through as 'just wiring.'"
--
--     guardian (missing): "A concrete plan/ticket for the workflow step that
--                 will actually invoke WriteExperiencePatternAction, with its
--                 owning pipeline named."
--
--   So it is named here rather than smuggled into an existing agent: the
--   experience register is a library with its own lifecycle, and an entry
--   arriving through some other agent's workflow would make its provenance a
--   matter of archaeology.
--
-- THE ONE THING THIS WORKFLOW EXISTS TO MAKE IMPOSSIBLE
--   Risk #3 of that same submission: "a workflow that writes an entry and
--   forgets the paired write_doc_plan step produces a register row with no
--   provenance, and nothing currently detects that."
--
--   The two writes are therefore ONE workflow, not two calls a caller must
--   remember to make in order. `write_pattern` hands off to `write_travelling_doc`
--   and the run does not complete without it. An entry in the register always has
--   a document saying where it came from and what it is for, because there is no
--   path that produces one without the other.
--
--   This is the same reasoning as migration 230's approval constraint: a rule
--   somebody must remember is a schema defect wearing a documentation costume.
--   Here the shape of the workflow is the enforcement.
--
-- WHAT IT DOES NOT DO
--   It does not approve anything. write_experience_pattern always writes
--   'draft', and refuses a caller that supplies a status at all. Approval is a
--   council verdict; 'proven' is the first live green run of bound criteria.
--   Nothing in this file can shortcut either.
--
-- INPUT
--   input_data: {
--     experience_pattern: { ...the entry... },   -- validated by the action
--     doc_plan_body: "markdown"                  -- its travelling doc
--   }
--
-- PREREQUISITE (checked, not assumed — v1.0.1175, pod agent-chassis-566bf56b78-jtjnj):
--   kubectl exec -n ai-persona-system <pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c write_experience_pattern'   -> 16
--   and a negative control returning 0, since a bare non-zero count proves only
--   that some string exists. Migrations 218 and 230 are applied.
-- ============================================================================

BEGIN;

INSERT INTO agent_definitions (
  id, type, display_name, description, category, status, is_active, default_config
)
SELECT
  gen_random_uuid(),
  'experience-register-writer',
  'Experience register writer',
  'The only pipeline that puts an entry into the experience register. Validates the entry through write_experience_pattern (which refuses a supplied status, refuses site-specific values in a base entry, and refuses a template whose placeholders do not close) and writes its travelling doc in the same run, so a register row cannot exist without provenance. Always writes draft: approval is a council verdict, not a field.',
  'documentation',
  'experimental',
  true,
  $cfg${
  "workflow": {
    "start_step": "write_pattern",
    "processing_mode": "orchestrator",
    "timeout_seconds": 300,
    "steps": {
      "write_pattern": {
        "action": "write_experience_pattern",
        "config": {
          "pattern_field": "input_data.experience_pattern",
          "created_by": "experience-register-writer",
          "source": "harvest"
        },
        "next_step": "write_travelling_doc",
        "output_field": "pattern_write",
        "description": "Validate and write the register entry. Refuses on any shape or criteria error; deferred checks are recorded, never dropped, never counted as a pass. Always writes status draft."
      },

      "write_travelling_doc": {
        "action": "write_doc_plan",
        "config": {
          "subject_type": "experience-pattern",
          "subject_key_field": "input_data.experience_pattern.name",
          "plan_body_field": "input_data.doc_plan_body",
          "plan_source": "harvest",
          "created_by": "experience-register-writer"
        },
        "next_step": "complete",
        "output_field": "doc_write",
        "description": "The entry PROVENANCE doc, keyed by the pattern name. In the same workflow as the write on purpose: an entry with no document is the gap this pipeline exists to close, and a caller that must remember a second call is not a control."
      },

      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": ["pattern_write", "doc_write"]
        }
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'experience-register-writer'
    AND COALESCE(is_snapshot, false) = false
    AND deleted_at IS NULL
);

-- Guard: assert the post-conditions inside the transaction.
DO $guard$
DECLARE
    cfg jsonb;
    steps jsonb;
BEGIN
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'experience-register-writer'
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '238: the experience-register-writer definition was not created';
    END IF;

    steps := cfg->'workflow'->'steps';

    IF steps->'write_pattern'->>'action' <> 'write_experience_pattern' THEN
        RAISE EXCEPTION '238: start step does not call write_experience_pattern';
    END IF;

    -- The load-bearing assertion: the pattern write must hand off to the doc
    -- write. If a later edit points write_pattern straight at complete, this
    -- file's whole reason for existing is gone and the guard says so.
    IF steps->'write_pattern'->>'next_step' <> 'write_travelling_doc'
       OR steps->'write_travelling_doc'->>'action' <> 'write_doc_plan' THEN
        RAISE EXCEPTION '238: the pattern write no longer hands off to the travelling-doc write — an entry could be created with no provenance, which is the one thing this pipeline exists to prevent';
    END IF;

    IF steps->'write_travelling_doc'->'config'->>'subject_type' <> 'experience-pattern' THEN
        RAISE EXCEPTION '238: the doc step writes the wrong subject_type';
    END IF;

    -- Nothing here may assert a status. Belt and braces: the action refuses one
    -- anyway, but a config that tried would be a statement of intent worth
    -- failing on.
    IF cfg::text LIKE '%"status": "approved"%' OR cfg::text LIKE '%"status": "proven"%' THEN
        RAISE EXCEPTION '238: this workflow must not assert an approval status';
    END IF;
END
$guard$;

COMMIT;

-- Verify
SELECT type, status, is_active,
       default_config->'workflow'->'steps'->'write_pattern'->>'action'        AS step_1,
       default_config->'workflow'->'steps'->'write_travelling_doc'->>'action' AS step_2
FROM agent_definitions
WHERE type = 'experience-register-writer'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Rollback recipe (hand-run):
--   UPDATE agent_definitions SET is_active = false, deleted_at = now()
--    WHERE type = 'experience-register-writer' AND COALESCE(is_snapshot,false) = false;
