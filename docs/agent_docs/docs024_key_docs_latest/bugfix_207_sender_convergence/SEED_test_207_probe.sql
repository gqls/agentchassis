-- Induction probe for bugs_open/207 (adapted from bugfix_196's proven v2 recipe —
-- SEED_test_196_probes.sql; their missteps are pre-applied here: void topic so the
-- parent parks, 600s call timeout, dispatch payload must carry child_agent_type).
-- Seed, create the void topic, dispatch parent, read R, dispatch the FLAT failing
-- child (inline workflow, no seed needed — deadline-exceeded via
-- local_action_timeout_seconds, RSH-004's proven machinery), capture the child's
-- response envelope from legacy system.generic.responses (the reply headers are
-- dropped at intake on this probe path — 196's documented wrinkle), then re-publish
-- it to system.agent.generic.responses to drive the parent's retry_version.
-- DELETE the row and the void topic after (cleanup at bottom).

INSERT INTO agent_definitions (type, display_name, description, category, default_config, is_active)
VALUES
('test-207-parent', 'bug 207 probe: awaiting parent',
 'Induction probe for bugs_open/207 - parks awaiting a request nothing real answers. DELETE after use.',
 'test',
 '{"workflow": {"start_step": "prepare_child", "steps": {
    "prepare_child": {
      "action": "query_database",
      "config": {
        "query": "SELECT ''probe-fake-agent'' AS agent_id, ''generic'' AS agent_type, ''child'' AS role, ''system.agent.test-207-void.requests'' AS requests_topic, ''system.agent.generic.responses'' AS responses_topic",
        "output_format": "object"
      },
      "output_field": "spawn_generic",
      "next_step": "call_child",
      "description": "Fabricate the spawn blob so call_agent resolves a topic without the flaky spawn handshake; the topic is VOID so the parent parks"
    },
    "call_child": {
      "action": "call_agent",
      "config": {
        "agent_type": "generic",
        "target_action": "orchestrate",
        "input_mapping": {"agent_type": "input_data.child_agent_type"},
        "timeout_seconds": 600
      },
      "output_field": "child_result",
      "next_step": "finish",
      "description": "The awaited call whose recoverable-vs-terminal answer is what this probe measures"
    },
    "finish": {
      "action": "complete_workflow",
      "config": {"output_fields": ["child_result"]},
      "description": "Reached only if the child failure is treated as step output"
    }
 }}}'::jsonb,
 true);

-- CLEANUP (run after the induction, whatever its outcome):
-- DELETE FROM agent_definitions WHERE type = 'test-207-parent';
-- and delete kafka topic system.agent.test-207-void.requests
