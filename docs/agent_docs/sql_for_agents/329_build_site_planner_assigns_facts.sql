-- 329 — build-site-planner: show the verified-fact roster and ask for per-section
-- fact assignments.
--
-- bugs_open/151 candidate 1, config half 2 of 3 (328 wiring, 329 planner prompt,
-- 330 writer prompt). Three anchored edits to plan_site's prompt_template:
--   1. a "Verified Facts (evidence base)" section before "## Canonical Page
--      Types" — the roster (id: claim, writer-visible facts only). The planner
--      already RECEIVES site_specs; it was never SHOWN the facts (measured
--      2026-08-06: neither planner's config references evidence_base).
--   2. the JSON example's sections array shows the object form.
--   3. RULES gains rule 17: assign each fact to exactly ONE section, spread
--      them, "facts": [] = deliberately factless, plain strings when no roster.
--
-- The edits operate on the ->> extracted template (raw string, no jsonb ::text
-- escaping games) and are written back with jsonb_set/to_jsonb.
--
-- SAFE AGAINST EITHER ROLL ORDER, stated not assumed: ValidateSitePlanAction on
-- an OLD binary passes object-form entries through untouched and
-- extractSectionEntries has always taken the name and dropped other keys — so a
-- plan emitted against this prompt before the Go half rolls loses the
-- assignments (facts dropped at write) and nothing else. The NEW binary against
-- the OLD prompt sees plain strings = NULL assignments = pre-existing behaviour.
-- Still: image first, then this file, so no plan silently sheds its assignments.
--   Pod check (one exec, control from a different file, invariant under this
--   change):
--   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
--     'grep -ac "normalised object-form sections" /app/agent-chassis; \
--      grep -ac "site_plan_sections lookup failed" /app/agent-chassis'

SELECT snapshot_agent('build-site-planner', '329_build_site_planner_assigns_facts.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_329 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Pre-check: each anchor occurs EXACTLY ONCE in the live template, and the
-- edit has not already been applied.
DO $$
DECLARE
    t text;
    c1 int; c2 int; c3 int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '329: build-site-planner plan_site prompt_template not found';
    END IF;
    IF position('Verified Facts (evidence base)' in t) > 0 THEN
        RAISE EXCEPTION '329: already applied — Verified Facts roster present';
    END IF;

    c1 := (length(t) - length(replace(t, '## Canonical Page Types', ''))) / length('## Canonical Page Types');
    c2 := (length(t) - length(replace(t, '"sections": ["hero", "features", "testimonials", "call-to-action"]', ''))) / length('"sections": ["hero", "features", "testimonials", "call-to-action"]');
    c3 := (length(t) - length(replace(t, 'Over-decomposing is cheap (a few unused icons); under-decomposing is expensive (manual cleanup of botched output).', ''))) / length('Over-decomposing is cheap (a few unused icons); under-decomposing is expensive (manual cleanup of botched output).');

    IF c1 <> 1 OR c2 <> 1 OR c3 <> 1 THEN
        RAISE EXCEPTION '329: anchor counts must all be exactly 1, got %/%/% — the live prompt has drifted; re-derive the anchors from the live row', c1, c2, c3;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(replace(replace(
                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                $anc1$## Canonical Page Types$anc1$,
                $rep1$## Verified Facts (evidence base)

{{if .site_specs.specs.evidence_base}}{{if .site_specs.specs.evidence_base.facts}}These are the verified facts registered for this site. Assign each fact you want the site to state to exactly ONE section, using the object form of section entries (RULES, rule 17). The section a fact is assigned to becomes the ONLY place the site states it; a fact assigned nowhere is stated nowhere. Never give two sections three or more of the same facts — restating the same facts in sibling sections is the duplication defect this assignment exists to prevent. A section that should state no verified facts gets "facts": [].

{{range .site_specs.specs.evidence_base.facts}}{{if .writer_line}}- {{.id}}: {{.claim}}
{{end}}{{end}}{{else}}No verified facts are registered for this site — use plain string section entries and no facts keys.{{end}}{{else}}No verified facts are registered for this site — use plain string section entries and no facts keys.{{end}}

## Canonical Page Types$rep1$),
                $anc2$"sections": ["hero", "features", "testimonials", "call-to-action"]$anc2$,
                $rep2$"sections": ["hero", {"name": "features", "facts": ["F1-example-id"]}, "testimonials", "call-to-action"]$rep2$),
                $anc3$Over-decomposing is cheap (a few unused icons); under-decomposing is expensive (manual cleanup of botched output).$anc3$,
                $rep3$Over-decomposing is cheap (a few unused icons); under-decomposing is expensive (manual cleanup of botched output).
17. Section entries may be plain strings ("hero") or objects ({"name": "features", "facts": ["F1-…"]}). Use the object form when the Verified Facts section above lists facts: put every fact you want stated into exactly ONE section's "facts" list, using the IDs exactly as listed there (an ID not in that list is ignored). Spread facts across sections — never give two sections three or more of the same facts. "facts": [] marks a section that deliberately states no verified facts. When no Verified Facts are listed, use plain strings only.$rep3$)
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify: all three edits present in the written-back template.
DO $$
DECLARE
    t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('Verified Facts (evidence base)' in t) = 0
       OR position('"facts": ["F1-example-id"]' in t) = 0
       OR position('17. Section entries may be plain strings' in t) = 0 THEN
        RAISE EXCEPTION '329: verify failed — one or more edits missing from the written-back template';
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad
--   SET default_config = b.default_config
--   FROM agent_definitions_bak_329 b
--   WHERE ad.id = b.id;
