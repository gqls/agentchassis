-- Bump batch step config batch_size as fallback
-- (in case Go input_data override isn't deployed)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_batch,config,batch_size}',
        '100'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';