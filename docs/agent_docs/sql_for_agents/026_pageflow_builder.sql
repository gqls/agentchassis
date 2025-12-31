changed from multipage-website-builder version 3


UPDATE agent_definitions
SET
    type = 'pageflow-builder',
    display_name = 'PageFlow Builder',
    description = 'Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time.',
    version = 1,  -- Reset to v1 since it's a new type
    updated_at = NOW()
WHERE type = 'multipage-website-builder'
  AND version = 3;

-- Step 2: Verify the rename worked
SELECT
    id,
    type,
    version,
    display_name,
    description,
    is_active,
    status
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Step 3: Check that intake-orchestrator will find it (matches %-builder pattern)
SELECT
    type,
    display_name,
    description
FROM agent_definitions
WHERE type LIKE '%-builder'
  AND is_active = true
ORDER BY type;

-- Step 4: Verify there's no conflict with old multipage-website-builder entries
SELECT
    type,
    version,
    display_name,
    is_active
FROM agent_definitions
WHERE type LIKE '%multipage%' OR type LIKE '%pageflow%'
ORDER BY type, version;