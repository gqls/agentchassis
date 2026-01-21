-- ============================================================================
-- Replace all dated Claude model names with aliases
--
-- This makes the configuration more resilient to Anthropic model updates.
-- When a new model version is released, only the alias resolver in code
-- needs updating, not every agent definition.
-- ============================================================================

-- First, let's see what models are currently in use
SELECT DISTINCT
    regexp_matches(default_config::text, '"(claude-[a-z0-9-]+)"', 'g') as model_names
FROM agent_definitions
WHERE default_config::text ~* 'claude-';

-- ============================================================================
-- Claude 4.5 family - replace dated versions with aliases
-- ============================================================================

-- claude-sonnet-4-5-YYYYMMDD → claude-sonnet-4-5
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-sonnet-4-5-[0-9]{8}"',
        '"claude-sonnet-4-5"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-sonnet-4-5-[0-9]{8}';

-- claude-haiku-4-5-YYYYMMDD → claude-haiku-4-5
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-haiku-4-5-[0-9]{8}"',
        '"claude-haiku-4-5"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-haiku-4-5-[0-9]{8}';

-- claude-opus-4-5-YYYYMMDD → claude-opus-4-5
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-opus-4-5-[0-9]{8}"',
        '"claude-opus-4-5"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-opus-4-5-[0-9]{8}';

-- ============================================================================
-- Claude 4 family (non-4.5) - replace dated versions with aliases
-- ============================================================================

-- claude-sonnet-4-YYYYMMDD → claude-sonnet-4
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-sonnet-4-[0-9]{8}"',
        '"claude-sonnet-4"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ '"claude-sonnet-4-[0-9]{8}"'
  AND default_config::text !~ 'claude-sonnet-4-5';  -- Don't match 4-5

-- claude-opus-4-YYYYMMDD → claude-opus-4
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-opus-4-[0-9]{8}"',
        '"claude-opus-4"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ '"claude-opus-4-[0-9]{8}"'
  AND default_config::text !~ 'claude-opus-4-5';

-- ============================================================================
-- Claude 3.5 family - replace dated versions with aliases
-- ============================================================================

-- claude-3-5-sonnet-YYYYMMDD → claude-3-5-sonnet
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-3-5-sonnet-[0-9]{8}"',
        '"claude-3-5-sonnet"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-3-5-sonnet-[0-9]{8}';

-- claude-3-5-haiku-YYYYMMDD → claude-3-5-haiku
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-3-5-haiku-[0-9]{8}"',
        '"claude-3-5-haiku"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-3-5-haiku-[0-9]{8}';

-- ============================================================================
-- Claude 3 family (legacy) - replace dated versions with aliases
-- ============================================================================

-- claude-3-opus-YYYYMMDD → claude-3-opus
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-3-opus-[0-9]{8}"',
        '"claude-3-opus"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-3-opus-[0-9]{8}';

-- claude-3-sonnet-YYYYMMDD → claude-3-sonnet
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-3-sonnet-[0-9]{8}"',
        '"claude-3-sonnet"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-3-sonnet-[0-9]{8}';

-- claude-3-haiku-YYYYMMDD → claude-3-haiku
UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        '"claude-3-haiku-[0-9]{8}"',
        '"claude-3-haiku"',
        'g'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text ~ 'claude-3-haiku-[0-9]{8}';

-- ============================================================================
-- Verify the changes
-- ============================================================================

-- Show all model references after update
SELECT type,
       regexp_matches(default_config::text, '"(claude-[a-z0-9-]+)"', 'g') as model
FROM agent_definitions
WHERE default_config::text ~* 'claude-'
ORDER BY type;

-- Confirm no dated models remain
SELECT type, default_config::text
FROM agent_definitions
WHERE default_config::text ~ 'claude-[a-z0-9-]+-[0-9]{8}';