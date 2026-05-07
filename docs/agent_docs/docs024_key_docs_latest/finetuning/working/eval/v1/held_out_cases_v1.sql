SELECT row_to_json(t)
FROM (
         SELECT DISTINCT ON (orchestration_id)
             jsonb_build_array(
             jsonb_build_object('role', 'user',      'content', prompt_rendered),
             jsonb_build_object('role', 'assistant', 'content', cleaned_response)
             ) AS messages,
             jsonb_build_object(
             'source_log_id',    id::text,
             'agent_type',       agent_type,
             'step_name',        step_name,
             'orchestration_id', orchestration_id,
             'model',            model,
             'created_at',       to_char(created_at AT TIME ZONE 'UTC',
             'YYYY-MM-DD"T"HH24:MI:SS"Z"')
             ) AS metadata
         FROM (
             SELECT
             id, agent_type, step_name, orchestration_id, model, created_at,
             prompt_rendered,
             regexp_replace(
             regexp_replace(trim(response_text),
             E'^```(?:json|JSON)?\\s*\\n?', '', 'n'),
             E'\\s*\\n?```\\s*$', '', 'n'
             ) AS cleaned_response
             FROM llm_call_log
             WHERE agent_type = 'page-content-writer'
             AND step_name  = 'process_sections_loop_iter_0_generate_content'
             AND success    = true
             AND (response_text LIKE '{%' OR response_text LIKE E'```%')
             AND prompt_rendered IS NOT NULL
             AND length(prompt_rendered) > 100
             AND length(response_text) > 10
             -- Held-out: created after the iter_0 training export
             AND created_at > '2026-04-23 14:54:32 UTC'::timestamptz
             -- Defensive: exclude any source_log_id we actually trained on
             AND id::text NOT IN (
             SELECT metadata->>'source_log_id'
             FROM training_exports.rows
             WHERE export_id = '146a9a12-c953-48eb-bf1f-c1856e5f13b7'::uuid
             )
             ORDER BY orchestration_id, created_at
             ) cleaned
             LIMIT 50
     ) t;