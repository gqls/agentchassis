-- Multipage Wrapper
UPDATE agent_definitions
SET
    input_contract = '{
        "required": ["final_html", "input_data"],
        "expects": {
            "final_html": "object with html property"
        }
    }'::jsonb,
    output_contract = '{
        "produces": "site_files",
        "format": {
            "type": "object",
            "description": "Map of filename to HTML content for all pages"
        }
    }'::jsonb
WHERE type = 'multipage-wrapper';

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,wrap_multipage,config,index_html_field}',
            '"final_html"'::jsonb
                     )
WHERE type = 'multipage-wrapper';