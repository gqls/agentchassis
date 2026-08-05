-- Induction probes for bugs_open/196 (see RUNBOOK + HANDOFF in this directory).
-- Seed, dispatch the parent, read the PARENT's orchestration row, then DELETE both
-- rows (cleanup block at the bottom). Category 'test' after 195's convention.
--
-- Child: workflow fails ValidateWorkflow (action "complete" is unregistered; the
-- registered action is complete_workflow) => ErrWorkflowInvalid => the PERMANENT
-- branch of handleError => sendErrorResponse is the parent's ONLY answer.
-- Parent: fabricates the spawn blob via query_database (output_format=object
-- flattens row fields to the top level of the output_field, which is exactly the
-- flat requests_topic/responses_topic fallback findAgentByType reads), then
-- call_agent at the generic topic with input_data.agent_type naming the invalid
-- child (selectWorkflow resolves agent_type from input_data as its last check).

INSERT INTO agent_definitions (type, display_name, description, category, default_config, is_active)
VALUES
('test-196-invalid-child', 'bug 196 probe: invalid child',
 'Induction probe for bugs_open/196 - workflow fails validation. DELETE after use.',
 'test',
 '{"workflow": {"start_step": "only", "steps": {"only": {"action": "complete", "description": "deliberately unregistered action - fails ValidateWorkflow with WORKFLOW_INVALID"}}}}'::jsonb,
 true),
('test-196-parent', 'bug 196 probe: awaiting parent',
 'Induction probe for bugs_open/196 - call_agent at a child that fails validation. DELETE after use.',
 'test',
 '{"workflow": {"start_step": "prepare_child", "steps": {
    "prepare_child": {
      "action": "query_database",
      "config": {
        "query": "SELECT ''probe-fake-agent'' AS agent_id, ''generic'' AS agent_type, ''child'' AS role, ''system.agent.generic.requests'' AS requests_topic, ''system.agent.generic.responses'' AS responses_topic",
        "output_format": "object"
      },
      "output_field": "spawn_generic",
      "next_step": "call_child",
      "description": "Fabricate the spawn blob so call_agent resolves the generic topic without a real spawn (the spawn handshake is flaky per bugfix 003; this probe must not depend on it)"
    },
    "call_child": {
      "action": "call_agent",
      "config": {
        "agent_type": "generic",
        "target_action": "orchestrate",
        "input_mapping": {"agent_type": "input_data.child_agent_type"},
        "timeout_seconds": 120
      },
      "output_field": "child_result",
      "next_step": "finish",
      "description": "The awaited call whose failure answer is what this probe measures"
    },
    "finish": {
      "action": "complete_workflow",
      "config": {"output_fields": ["child_result"]},
      "description": "Reached only if the child failure was stamped complete (the bug) - reaching this step post-fix is the FALSIFIER"
    }
 }}}'::jsonb,
 true);

-- CLEANUP (run after both inductions, whatever their outcome):
-- DELETE FROM agent_definitions WHERE type IN ('test-196-invalid-child','test-196-parent');
