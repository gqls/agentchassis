-- Extract all field paths referenced in workflows
WITH workflow_json AS (
    SELECT type, default_config->'workflow'->'steps' as steps
    FROM agent_definitions
    WHERE default_config->'workflow' IS NOT NULL
)
SELECT DISTINCT
    type,
    path_value
FROM workflow_json,
     LATERAL (
              SELECT jsonb_path_query(steps, '$.**.agent_type_field')::text as path_value
              UNION ALL
              SELECT jsonb_path_query(steps, '$.**.default_from')::text
              UNION ALL
              SELECT jsonb_path_query(steps, '$.**.content_field')::text
              UNION ALL
              SELECT jsonb_path_query(steps, '$.**.iterate_over')::text
              UNION ALL
              SELECT jsonb_path_query(steps, '$.**.*_from')::text
              UNION ALL
              SELECT jsonb_path_query(steps, '$.**.*_field')::text
         ) paths
WHERE path_value IS NOT NULL
  AND path_value != 'null'
ORDER BY type, path_value;