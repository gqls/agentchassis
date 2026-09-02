-- 687 — build-site-planner: fix candidates #1 and #3 from bugs_open/428
-- (site-planner LLM knowingly defers strategy-named entity roles, citing its
-- own final say).
--
-- WHAT THIS CHANGES (surgical, anchored on the LIVE text, aborts on drift):
--  1. "## Domain Strategy" block: {{.site_specs.specs.strategy}} becomes
--     {{toJSON .site_specs.specs.strategy}}. toJSON is an EXISTING template
--     function (data_helpers.go RenderPromptTemplate funcMap, already live —
--     this migration adds no new capability, only uses one already deployed
--     to every prompt on the estate). Bug 428 §1 traced the current bare
--     interpolation to a real defect: Go's %v formatting of the nested
--     map/slice strategy data produces an unquoted, unstructured dump
--     ("recommended_page_types:[map[page_type:index reasoning:…] …]"). The
--     model reads it correctly regardless (428 §2's own citation proves
--     this), so this is NOT the omission's cause — it is an independent
--     legibility fix, worth shipping so the next thread auditing "did the
--     model see X" can read the prompt text directly.
--  2. The FINAL SAY rule gains a concrete requirement: any named
--     recommended_page_types entry NOT included in `pages` must be named,
--     by its exact page_type, in strategy_notes with a real per-type reason
--     — not the generic "keeping it lean" framing bug 428 §2 found
--     (boxingonline's own call named both roles then gave one shared
--     rationale for dropping both). This does not remove the model's
--     licensed final say (RFC-level enforcement was explicitly NOT chosen —
--     see 428 §6 candidate 1's softer option), it makes the "note why"
--     obligation the prompt already imposes concrete enough that an omission
--     is either a named decision or a visible gap, never a silent one.
--
-- NOT a _HOLD: toJSON is already live in every deployed agent-chassis binary
-- (confirmed by reading data_helpers.go's current, committed funcMap — no Go
-- change accompanies this migration), so there is no image-roll ordering
-- constraint. Applies immediately on submission/approval.
--
-- Council-Submitted: <fill in after 097_TRIGGER>

SELECT snapshot_agent('build-site-planner', '687_build_site_planner_strategy_json_and_omission_reason.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_687 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    t text;
    c1 int; c2 int; nrows int;
BEGIN
    -- Pre-flight (640's own council-raised lesson): SELECT INTO takes the
    -- FIRST of N rows silently; a duplicate active row would half-apply.
    SELECT count(*) INTO nrows FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF nrows <> 1 THEN
        RAISE EXCEPTION '687: expected exactly 1 active build-site-planner row BEFORE writing, found %', nrows;
    END IF;

    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '687: build-site-planner plan_site prompt_template not found';
    END IF;
    IF position('{{toJSON .site_specs.specs.strategy}}' in t) > 0 THEN
        RAISE EXCEPTION '687: already applied — toJSON strategy interpolation present';
    END IF;

    c1 := (length(t) - length(replace(t,
        $a1$## Domain Strategy
{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}$a1$,
        ''))) / length($a1$## Domain Strategy
{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}$a1$);

    c2 := (length(t) - length(replace(t,
        $a2$You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.$a2$,
        ''))) / length($a2$You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.$a2$);

    IF c1 <> 1 OR c2 <> 1 THEN
        RAISE EXCEPTION '687: live prompt has drifted (anchor counts %/%; both must be 1) — re-derive this seed from the live row', c1, c2;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(replace(
                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                $a1$## Domain Strategy
{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}$a1$,
                $r1$## Domain Strategy
{{if .site_specs.specs.strategy}}{{toJSON .site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}$r1$),
                $a2$You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.$a2$,
                $r2$You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment — but for EVERY page_type listed in Domain Strategy's recommended_page_types that you do NOT include in `pages`, name that exact page_type in strategy_notes together with a real per-type reason (e.g. "entity-page: deferred, no fighter content yet to aggregate into profiles" — not a generic "keeping the structure lean" that could justify dropping any number of types at once). An omitted named type with no per-type reason in strategy_notes is a gap, not a decision.$r2$)
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('{{toJSON .site_specs.specs.strategy}}' in t) = 0
       OR position('name that exact page_type in strategy_notes' in t) = 0
       OR position('{{.site_specs.specs.strategy}}{{else}}' in t) > 0 THEN
        RAISE EXCEPTION '687: verify failed — the toJSON interpolation or the per-type reason requirement is missing, or the old bare interpolation survived';
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run): restore default_config from
-- agent_definitions_bak_687, or from the snapshot_agent row this file took
-- first — see docs/agent_docs/sql_for_agents/687_..._ROLLBACK.sql.
