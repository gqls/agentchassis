-- Update the html-developer agent with the complete workflow
UPDATE agent_definitions
SET default_config = '{
    "agent_type": "html-developer",
    "processing_mode": "task",
    "workflow": {
        "start_step": "generate_html",
        "steps": {
            "generate_html": {
                "action": "generate_html",
                "description": "Generate HTML from context",
                "config": {
                    "model": "claude-3-opus-20240229",
                    "temperature": 0.7
                },
                "next_step": "process_html"
            },
            "process_html": {
                "action": "process_html",
                "description": "Process and enhance HTML",
                "config": {
                    "minify": true,
                    "add_meta": true,
                    "responsive": true
                },
                "next_step": "validate_html"
            },
            "validate_html": {
                "action": "validate_html",
                "description": "Validate HTML structure",
                "next_step": "store_html"
            },
            "store_html": {
                "action": "store_result",
                "description": "Store HTML to S3",
                "config": {
                    "storage_type": "s3",
                    "content_field": "final_html",
                    "content_type": "text/html",
                    "path_template": "websites/{{.ClientID}}/{{.CorrelationID}}/index.html"
                },
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow",
                "description": "Complete HTML generation"
            }
        }
    },
    "storage": {
        "enabled": true,
        "type": "s3",
        "bucket_env": "ASSETS_BUCKET",
        "public_access": true,
        "auto_store": true
    }
}'::jsonb
WHERE type = 'html-developer';