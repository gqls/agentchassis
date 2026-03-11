-- Fall back to town or just name + UK when postcode missing
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,search_practice,config,query_template}',
        '"{{.business_record.business.name}} {{if .business_record.business.postcode}}{{.business_record.business.postcode}}{{else if .business_record.business.town}}{{.business_record.business.town}}{{end}} veterinary practice UK"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';