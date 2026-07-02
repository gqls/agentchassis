--- prompt template fix to primary colour css names

-- Step 3: apply the fix
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb(REPLACE(
                default_config->>'prompt_template',
                '7. CSS VARIABLES AVAILABLE:
           --color-primary, --color-primary-hover, --color-primary-text
           --color-secondary, --color-accent
           --color-text, --color-text-muted, --color-heading
           --color-background, --color-surface, --color-card-bg, --color-border
           --color-header-bg, --color-header-text
           --color-footer-bg, --color-footer-text, --color-white
           --container-max-width (1200px), --spacing-section (5rem 2rem)
           --border-radius, --shadow',
                '7. CSS VARIABLES AVAILABLE — USE ONLY THESE NAMES:
           Palette (all begin with --color-):
             --color-primary, --color-primary-hover, --color-primary-text
             --color-secondary, --color-accent
             --color-text, --color-text-muted, --color-heading
             --color-background, --color-surface, --color-card-bg, --color-border
             --color-header-bg, --color-header-text
             --color-footer-bg, --color-footer-text, --color-white
           Layout:
             --container-max-width (1200px), --spacing-section (5rem 2rem)
             --border-radius, --shadow

           STRICT RULE: Do NOT invent CSS variable names. The only permitted colour
           variable pattern is --color-{role} as listed above. Variables like
           --primary-color, --secondary-color, --accent-color, --background-color,
           --text-color, --border-color are WRONG and will produce broken output
           because they are undefined in every deployed stylesheet. Always use
           --color-primary, --color-secondary, --color-accent etc.'
                 ))
                     ),
    updated_at = now()
WHERE type = 'component-creator';

-- Step 4: verify the patch landed
SELECT SUBSTRING(
               default_config->>'prompt_template'
    FROM POSITION('7. CSS VARIABLES' IN (default_config->>'prompt_template'))
    FOR 250
) AS new_section7
FROM agent_definitions
WHERE type = 'component-creator';

---

BEGIN;
DO $mig$
DECLARE p text; n int;
BEGIN
SELECT default_config->>'prompt_template' INTO p FROM agent_definitions
WHERE type='component-creator' AND deleted_at IS NULL AND is_active=true AND (is_snapshot IS NULL OR is_snapshot=false);
IF p IS NULL THEN RAISE EXCEPTION 'prompt not found'; END IF;
  n := (length(p) - length(replace(p, '{{if .existing_field_names}}', ''))) / length('{{if .existing_field_names}}');
  IF n = 0 THEN RAISE NOTICE 'no dead block — nothing to clean'; RETURN; END IF;
  PERFORM snapshot_agent('component-creator', 'F1-prompt cleanup: remove dead existing_field_names block');
UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{prompt_template}',
                               to_jsonb(regexp_replace(p, '\{\{if \.existing_field_names\}\}.*?\{\{end\}\}\s*', '', 'g')), true)
WHERE type='component-creator' AND deleted_at IS NULL AND is_active=true AND (is_snapshot IS NULL OR is_snapshot=false);
RAISE NOTICE 'removed % dead existing_field_names block(s)', n;
END $mig$;
-- verify: dead gone (0), live remains (t)
SELECT position('{{if .existing_field_names}}' IN (default_config->>'prompt_template')) AS dead_pos,
       position('{{if .existing_component.field_names}}' IN (default_config->>'prompt_template')) > 0 AS live_present
FROM agent_definitions WHERE type='component-creator' AND deleted_at IS NULL AND is_active=true;
COMMIT;
