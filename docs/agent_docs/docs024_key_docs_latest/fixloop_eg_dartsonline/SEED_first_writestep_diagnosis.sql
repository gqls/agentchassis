-- Seed a hand-authored CONFIRMED diagnosis for the first end-to-end write-step run.
-- Real defect: generate_image_actions.go:594, inside loadAgentDefinitionForImageAction,
-- a raw fmt.Printf debug line that (a) bypasses the structured zap logger and
-- (b) names the WRONG function (loadAgentDefinitionForAction — a copy-paste from
-- ai_actions.go). Genuine defect, single-file, zero behavioural risk. The
-- diagnosis is hand-written (the diagnosis LOOP is separately proven); proposer,
-- council, and implementer that consume it all run for real.
\set corr '11111111-e2e2-4a1b-9c3d-000000000001'

INSERT INTO orchestration_states (
    orchestration_id, correlation_id, client_id, status, current_step,
    workflow_plan, collected_data, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    :'corr'::uuid,
    'demo_client',
    'COMPLETED',
    'diagnose_emit',
    '{}'::jsonb,
    jsonb_build_object(
      'diagnosis', jsonb_build_object(
        'status', 'CONFIRMED',
        'conclusion',
          E'CONFIRMED - misleading debug log in the image agent-definition loader.\n\nSYMPTOM: pod logs for image-generation work emit a line "DEBUG: loadAgentDefinitionForAction called with agentType=..." from a function that is NOT loadAgentDefinitionForAction, making the logs actively misleading when tracing image-agent definition loads.\n\nMECHANISM (static, cited): platform/orchestration/actions/generate_image_actions.go, function loadAgentDefinitionForImageAction, contains at its top:\n    fmt.Printf("DEBUG: loadAgentDefinitionForAction called with agentType=%s, db type=%T\\n", agentType, db)\nTwo defects in one line:\n1. WRONG FUNCTION NAME. The enclosing function is loadAgentDefinitionForImageAction, but the string says loadAgentDefinitionForAction - a verbatim copy-paste from the sibling ai_actions.go:722, which defines the correctly-named loadAgentDefinitionForAction. So the log attributes this loader''s activity to the wrong function.\n2. RAW fmt.Printf, NOT THE STRUCTURED LOGGER. The line writes to stdout via fmt.Printf instead of the project''s zap logger. It carries no orchestration_id/correlation_id and cannot be filtered or levelled, violating the logging constitution ("put the run id in log lines"; use the structured logger, not ad-hoc prints).\n\nFIX (constrained, single file): in generate_image_actions.go, replace the raw fmt.Printf debug line in loadAgentDefinitionForImageAction with a structured logger call at an emitted level naming the CORRECT function and carrying agentType - or, since this helper takes db interface{} with no logger in scope and the line is pure debug noise, remove it. Either removes the misleading output. No behaviour other than logging changes.\n\nSCOPE: this fix covers generate_image_actions.go only. A sibling raw fmt.Printf debug line exists at ai_actions.go:722 (there the name is correct); it is a separate instance, out of scope for this constrained fix.',
        'seeded_by', 'hand-authored for first write-step end-to-end run (2026-07-12); diagnosis loop separately proven'
      )
    ),
    now(), now()
);

SELECT correlation_id,
       collected_data->'diagnosis'->>'status' AS status,
       left(collected_data->'diagnosis'->>'conclusion', 70) AS conclusion_head
FROM orchestration_states WHERE correlation_id = :'corr'::uuid;
