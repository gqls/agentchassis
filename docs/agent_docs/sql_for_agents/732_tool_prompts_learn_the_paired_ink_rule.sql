-- 732 — teach the two tool-writing prompts the on-colour pairing rule
--
-- bugs_open/458. The renderer emits paired ink tokens (--color-primary-text for
-- text ON a primary fill; --color-primary-ink / --color-accent-ink for those
-- colours used AS text on the page). The component-creator prompt already
-- teaches this and uses the two-level var() form. The two TOOL-writing prompts
-- never learned it: their whole colour vocabulary is fills and page inks, so an
-- LLM filling a button with --color-primary has to invent its text colour and
-- reaches for the page ground.
--
-- Measured 2026-09-03 over active unforked content_components:
--   non-tool: 151 components, 0 with a primary fill inked with the page ground
--   tool:     261 components, 148 with exactly that shape
-- and 9 of 59 palettes score under 3.0:1 for it (7 under 1.25:1), so on those
-- sites the label is not readable at all.
--
-- SURGICAL: anchored on the verbatim rule line in each prompt, and it ABORTS if
-- either anchor has moved rather than writing a prompt it did not read.
-- Idempotent: re-running is a no-op once the new text is in place.

BEGIN;

-- ---------------------------------------------------------------- guard: pre
DO $guard$
DECLARE
    n_gen int;
    n_imp int;
BEGIN
    SELECT count(*) INTO n_gen FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}'
           LIKE '%4. Use CSS custom properties for colours: var(--color-primary), var(--color-secondary)%';

    SELECT count(*) INTO n_imp FROM agent_definitions
     WHERE type = 'tool-improver' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,improve_tool,config,prompt_template}'
           LIKE '%3. Use CSS custom properties (var(--color-primary) etc) for colours%';

    -- Already applied? Then both anchors are gone and both replacements present.
    IF n_gen = 0 AND n_imp = 0 THEN
        IF EXISTS (SELECT 1 FROM agent_definitions
                    WHERE type IN ('tool-generator','tool-improver') AND is_active
                      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
                      AND default_config::text LIKE '%--color-primary-ink%') THEN
            RAISE NOTICE '732: already applied, nothing to do';
            RETURN;
        END IF;
    END IF;

    IF n_gen <> 1 THEN
        RAISE EXCEPTION '732 ABORT: expected exactly 1 live tool-generator carrying the rule-4 anchor, found %. The prompt has been edited since 2026-09-03; re-read it and re-anchor rather than overwriting a prompt this migration has not seen.', n_gen;
    END IF;
    IF n_imp <> 1 THEN
        RAISE EXCEPTION '732 ABORT: expected exactly 1 live tool-improver carrying the rule-3 anchor, found %.', n_imp;
    END IF;
END
$guard$;

-- Pre-mutation backup (council round 1, debug_historian, corr 0fd2ca6b): a
-- guard/verify sandwich is not a backup, and these two rows govern every tool
-- generated fleet-wide. TWO-ARG overload deliberately -- it writes
-- agent_definitions_backup; the one-arg form writes an is_snapshot row into
-- agent_definitions itself (LANDMINES, "snapshot_agent has TWO overloads").
-- Verify the snapshot holds the PRE-change config, not merely that one exists:
--   SELECT type, snapshot_taken_at,
--          NOT (default_config::text LIKE '%--color-primary-ink%') AS has_old
--   FROM agent_definitions_backup WHERE type IN ('tool-generator','tool-improver')
--   ORDER BY snapshot_taken_at DESC LIMIT 2;   -- has_old must be true for both
SELECT snapshot_agent('tool-generator', '732_tool_prompts_learn_the_paired_ink_rule.sql: pre-update');
SELECT snapshot_agent('tool-improver',  '732_tool_prompts_learn_the_paired_ink_rule.sql: pre-update');

-- ------------------------------------------------------------ tool-generator
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_tool_html,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}',
            '4. Use CSS custom properties for colours: var(--color-primary), var(--color-secondary), var(--color-accent), var(--color-background), var(--color-surface), var(--color-text), var(--color-text-muted), var(--color-border)',
            '4. Use CSS custom properties for colours: var(--color-primary), var(--color-secondary), var(--color-accent), var(--color-background), var(--color-surface), var(--color-text), var(--color-text-muted), var(--color-border). PAIR EVERY FILL WITH ITS OWN INK, never with the page ground. Text ON a --color-primary fill is var(--color-primary-text, #fff); text ON a --color-accent fill is var(--color-accent-text, var(--color-text)). --color-primary or --color-accent used AS text on the page is var(--color-primary-ink, var(--color-primary)) or var(--color-accent-ink, var(--color-accent)). Write the two-level var(name, fallback) form exactly as shown, never the bare name. NEVER ink a fill with var(--color-background) or var(--color-surface): you cannot see this site palette, and on a site whose primary sits near its own ground that pairing renders the label invisible.'
        )),
        false)
WHERE type = 'tool-generator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ------------------------------------------------------------- tool-improver
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,improve_tool,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,improve_tool,config,prompt_template}',
            '3. Use CSS custom properties (var(--color-primary) etc) for colours — never hardcode hex values',
            '3. Use CSS custom properties (var(--color-primary) etc) for colours — never hardcode hex values. Pair every fill with its own ink, never with the page ground: text ON a --color-primary fill is var(--color-primary-text, #fff), text ON a --color-accent fill is var(--color-accent-text, var(--color-text)), and --color-primary or --color-accent used AS text on the page is var(--color-primary-ink, var(--color-primary)) or var(--color-accent-ink, var(--color-accent)). Always the two-level var(name, fallback) form. If the tool you are improving inks a fill with var(--color-background) or var(--color-surface), that is a defect to fix even when it was not the reported issue: on a site whose primary sits near its ground it is invisible text.'
        )),
        false)
WHERE type = 'tool-improver' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- --------------------------------------------------------------- guard: post
DO $verify$
DECLARE
    ok_gen boolean;
    ok_imp boolean;
BEGIN
    -- #>> not ::text: ::text is the JSON SERIALISATION, so an embedded quote is
    -- stored escaped and a LIKE against it returns a clean FALSE rather than an
    -- error (raised by the bugs_open/450 lane, 2026-09-03). These two needles
    -- carry no quotes, so both forms agree today -- the extraction is used
    -- because the shape is wrong, not because this needle is.
    SELECT default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}' LIKE '%--color-primary-ink%'
           AND default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}' LIKE '%--color-primary-text, #fff%'
      INTO ok_gen
      FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    SELECT default_config #>> '{workflow,steps,improve_tool,config,prompt_template}' LIKE '%--color-primary-ink%'
      INTO ok_imp
      FROM agent_definitions
     WHERE type = 'tool-improver' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF NOT COALESCE(ok_gen, false) THEN
        RAISE EXCEPTION '732 ABORT: tool-generator prompt does not carry the paired-ink rule after the update; nothing committed.';
    END IF;
    IF NOT COALESCE(ok_imp, false) THEN
        RAISE EXCEPTION '732 ABORT: tool-improver prompt does not carry the paired-ink rule after the update; nothing committed.';
    END IF;
    RAISE NOTICE '732 OK: both tool-writing prompts now carry the paired-ink rule';
END
$verify$;

COMMIT;
