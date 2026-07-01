-- FILE: migrations/NNN_component_creator_preserve_field_names.sql
--
-- F1-prompt (Part 2 of 2): add a regeneration field-name-preservation rule to
-- the component-creator generate_template prompt. This is the anti-drift anchor
-- that complements the StoreGeneratedComponentAction field-contract guard —
-- same pattern as the tool-doc-header rule on tool-improver.
--
-- PATH NOTE (072 nested-prompt trap): the prompt_template lives at the TOP
-- LEVEL of default_config (default_config->>'prompt_template'), NOT at
-- workflow.steps.generate_template.config.prompt_template — that step config
-- holds only ai_service + input_fields. A migration on the step path would be
-- a silent no-op. This migration edits the top-level key.
--
-- DORMANCY: the rule is wrapped in {{if .existing_field_names}}...{{end}}, so
-- it renders to nothing until the companion load-existing-component step (Part
-- 1) feeds `existing_field_names` into generate_template's collected data and
-- adds it to that step's input_fields. Until Part 1 lands this migration is a
-- safe render-time no-op (it only inserts dormant prompt text).
--
-- Convention (per 019_tool_library / 016 debugging guide): snapshot-first,
-- anchored, idempotent, drift-checked, live-row only.

BEGIN;

DO $mig$
DECLARE
    p      text;
    anchor text := '== COMPONENT CONTRACT';   -- unique, ASCII (avoids the em-dash); "== END CONTRACT ==" does not match
    marker text := '== REGENERATION FIELD-NAME RULE ==';
    n      int;
    block  text := $blk$
{{if .existing_field_names}}
== REGENERATION FIELD-NAME RULE ==
This component ALREADY EXISTS and pages across the site have stored content keyed to its current field names. You are REGENERATING it, not creating a new one. In input_schema.fields and as {{placeholder "..."}} tokens you MUST reuse these exact field names, spelled identically (same case, same underscores):
{{.existing_field_names}}
You MAY add new fields, and you MAY restyle or restructure the template and CSS freely. You MUST NOT rename or drop any of the existing field names listed above. Renaming or removing one strands the stored content for every page that uses this shared component and renders those sections blank. If a field seems unneeded, keep its name present and reachable rather than removing it.
== END REGENERATION FIELD-NAME RULE ==
{{end}}

$blk$;
BEGIN
    SELECT default_config->>'prompt_template' INTO p
    FROM agent_definitions
    WHERE type = 'component-creator' AND deleted_at IS NULL AND is_active = true
      AND (is_snapshot IS NULL OR is_snapshot = false);

    IF p IS NULL THEN
        RAISE EXCEPTION 'component-creator: top-level prompt_template not found (check the path / 072 trap)';
    END IF;

    -- idempotent: bail if the rule is already present
    IF position(marker IN p) > 0 THEN
        RAISE NOTICE 'component-creator: regeneration rule already present — no change';
        RETURN;
    END IF;

    -- drift guard: the anchor must appear exactly once
    n := (length(p) - length(replace(p, anchor, ''))) / length(anchor);
    IF n <> 1 THEN
        RAISE EXCEPTION 'component-creator: anchor "%" found % times (expected 1) — prompt drifted, aborting', anchor, n;
    END IF;

    -- snapshot only once we know we are going to change something
    PERFORM snapshot_agent('component-creator',
        'F1-prompt: add regeneration field-name preservation rule');

    UPDATE agent_definitions
    SET default_config = jsonb_set(
            default_config, '{prompt_template}',
            to_jsonb(replace(p, anchor, block || anchor)), true)
    WHERE type = 'component-creator' AND deleted_at IS NULL AND is_active = true
      AND (is_snapshot IS NULL OR is_snapshot = false);

    RAISE NOTICE 'component-creator: regeneration field-name rule inserted before the component contract';
END
$mig$;

-- verify (expect rule_present = t, and the placeholder token preserved)
SELECT position('== REGENERATION FIELD-NAME RULE ==' IN (default_config->>'prompt_template')) > 0 AS rule_present,
       position('{{.existing_field_names}}' IN (default_config->>'prompt_template')) > 0 AS injects_names
FROM agent_definitions
WHERE type = 'component-creator' AND deleted_at IS NULL AND is_active = true;

COMMIT;
