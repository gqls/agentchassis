-- ROLLBACK for 732. Removes the appended paired-ink sentences, restoring each
-- rule line to its 2026-09-03 wording. Anchored on the appended text, so it is
-- a no-op if 732 was never applied.
BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_tool_html,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}',
            '. PAIR EVERY FILL WITH ITS OWN INK, never with the page ground. Text ON a --color-primary fill is var(--color-primary-text, #fff); text ON a --color-accent fill is var(--color-accent-text, var(--color-text)). --color-primary or --color-accent used AS text on the page is var(--color-primary-ink, var(--color-primary)) or var(--color-accent-ink, var(--color-accent)). Write the two-level var(name, fallback) form exactly as shown, never the bare name. NEVER ink a fill with var(--color-background) or var(--color-surface): you cannot see this site palette, and on a site whose primary sits near its own ground that pairing renders the label invisible.',
            ''
        )),
        false)
WHERE type = 'tool-generator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,improve_tool,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,improve_tool,config,prompt_template}',
            '. Pair every fill with its own ink, never with the page ground: text ON a --color-primary fill is var(--color-primary-text, #fff), text ON a --color-accent fill is var(--color-accent-text, var(--color-text)), and --color-primary or --color-accent used AS text on the page is var(--color-primary-ink, var(--color-primary)) or var(--color-accent-ink, var(--color-accent)). Always the two-level var(name, fallback) form. If the tool you are improving inks a fill with var(--color-background) or var(--color-surface), that is a defect to fix even when it was not the reported issue: on a site whose primary sits near its ground it is invisible text.',
            ''
        )),
        false)
WHERE type = 'tool-improver' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;
