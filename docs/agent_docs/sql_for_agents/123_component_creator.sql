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