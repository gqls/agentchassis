-- 516 — the two tool BUILD steps declare `related_pages` with the `?`
--       OPTIONAL-EXPLICIT marker: the suggestion's own path or NOTHING, never
--       the whole-tree search. This is the fix for `bugs_open/330` (090
--       CONFIRMED). CONFIG ONLY — live on apply.
--
--       ⚠ _HOLD, AND THE ORDERING CONDITION IS NOT A ROLL. Read "ORDERING"
--       below before applying: this migration CONSUMES the demand control that
--       migration 512's verification depends on, and 512 is applied but NOT yet
--       verified.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- `create_tool_component` and `deploy_tool_to_site` both declare `related_pages`
-- as an Optional input, and migration 211 wired both to
-- `input_data.spec.related_pages` — the add_tool work item's own suggestion.
--
-- But most add_tool specs do not carry the key (4 of the 5 most recent
-- tool-generator runs when 330 was filed). A declared path that resolves to
-- nothing does NOT stop the resolver: the field falls through to the whole-tree
-- search, which finds every `related_pages` in the tree — all of them inside
-- `load_brand_context.specs.tools.suggestions[N].related_pages` — and takes the
-- shallowest-first winner, `suggestions[0]`'s.
--
-- On webdesign.co.uk that produced NINE different tools cross-linked to the SAME
-- TWO pages (`learn-algorithms-p-values-explained`, `learn-algorithms-bayesian-
-- theory`) — which are correct for exactly one tool, the A/B Test Significance
-- Calculator, because it is `suggestions[0]`. One tool's right answer was handed
-- to nine others. 32 items; 0 reached live content (22 failed, 6 wont_fix, 2
-- triaged, 2 unresolved), so the cost so far is wasted writer work and a
-- polluted queue, not wrong links on served pages.
--
-- ============================================================================
-- WHY `?` IS THE RIGHT TOOL, AND WHY "A DECISION, NOT A WIRE" IS NOT
-- ============================================================================
-- The lane's own live-class ledger currently reads, for this pair: "Refusal is
-- the DESIRED outcome, so it needs a recorded decision, not a wire." That was
-- written before the `?` marker was live and it is superseded — the two halves
-- of this field's correct behaviour cannot both be had without a wire:
--
--   * the spec DOES carry related_pages (a real, live case): the value must be
--     used. Un-declaring the field, or 512's `input_fields` shape, would lose it.
--   * the spec does NOT carry it: the answer must be ABSENT, and absence is
--     already a handled, documented state downstream (see below).
--
-- `"related_pages?": "input_data.spec.related_pages"` is precisely those two
-- halves: resolve from the named path, or leave the field absent — never the
-- whole-tree search, never the nested-object fallback, never the deprecated
-- bridge. A PLAIN wire cannot do it: it is what is wired today, and its miss is
-- exactly the fall-through this bug is about.
--
-- ============================================================================
-- WHY ABSENCE IS SAFE HERE — read at the consumer, first-hand, 2026-08-21
-- ============================================================================
-- Both steps' actions reach the same helper, `relatedPagesFromInputs`
-- (`create_tool_cross_link_items.go:443`), from three call sites
-- (`create_tool_component_action.go:527`, `deploy_tool_action.go:271,548`).
-- With the field absent, `inputs.GetRaw("related_pages")` is nil and the helper
-- falls back to reading THE SAME declared path directly
-- (`ExtractNestedField(collected, "input_data.spec.related_pages")`) — also nil.
-- So absence propagates, and nothing reaches for a search.
--
-- Absence is then not merely tolerated but INSTRUMENTED
-- (`create_tool_cross_link_items.go:142-151`):
--     if len(req.relatedPages) == 0 {
--         // Not a defect: a suggestion may legitimately name no related pages.
--         logger.Info("emitToolCrossLinkItems: no related_pages on the suggestion, ...")
--         recordCrossLinkSkip(ctx, params, logger, req, "no_related_pages", "info", ...)
--         return 0
--     }
-- — a counted, reasoned skip at info. That is the acknowledgement RFC_029
-- §10.15's adoption gate requires, and it is recorded in
-- `architecture_review/optional_explicit_wire_acks.json` in the same commit.
--
-- ============================================================================
-- SCOPE — two steps, and why the second one is in
-- ============================================================================
-- The conflict instrument has logged this class for `tool-generator` only
-- (27 rows, last 2026-08-20 09:50:10Z). `tool-deployer`/`deploy_tool` has ZERO
-- logged rows, so its inclusion is PROPHYLACTIC and is stated as such: it
-- carries the identical unmarked wire to the identical helper, so the identical
-- substitution is available to it the moment its spec omits the key. Leaving one
-- of two identical armed wires unmarked is how a class gets "fixed" and then
-- rediscovered. Both are one line each and both are behaviour-identical whenever
-- the path resolves.
--
-- ============================================================================
-- ORDERING — why this is a _HOLD, and the condition is NOT "after a roll"
-- ============================================================================
-- The binary precondition is already MET: the `?` parser is live on `v1.0.1321`
-- (ecc419bd1 an ancestor of build revision 0483e7f4e; both pods one digest).
-- So this is NOT held for a roll.
--
-- It is held because **applying it destroys another migration's demand control.**
-- Migration 512 (tool-generator/`reason`, applied 2026-08-20 17:38:34Z) is
-- APPLIED BUT UNVERIFIED, and its stated test is:
--     pass = `reason` rows 0 **while `related_pages` KEEPS FIRING**
--            (if both go quiet, the instrument died and the zero means nothing)
-- `related_pages` is that control. Apply 516 first and both classes go quiet
-- together, permanently, and 512 can never be verified — only assumed.
--
-- [MEASURED 2026-08-21 ~11:4xZ] tool-generator runs since 512's boundary: **0**
-- (18 h). So 512's zero today is NO DEMAND, not no defect, and this file cannot
-- jump the queue on the grounds that "the class looks quiet".
--
-- APPLY ONLY WHEN BOTH HOLD:
--   1. tool-generator has RUN since 2026-08-20 17:38:34Z:
--      SELECT count(*), max(created_at) FROM orchestration_states
--       WHERE owner_agent_type='tool-generator' AND created_at > '2026-08-20 17:38:34Z';
--      -- must be > 0, and state n: n runs cannot detect a residual rarer than ~1 in n
--   2. On those runs, 512 reads PASS — `reason` 0 with `related_pages` still firing:
--      SELECT context->>'field', count(*), max(occurred_at) FROM agent_error_log
--       WHERE error_code='RESOLVER_CONFLICTING_CANDIDATES' AND agent_type='tool-generator'
--         AND occurred_at > '2026-08-20 17:38:34Z' GROUP BY 1;
-- Record 512's verdict in the lane NOTES BEFORE applying this. After applying,
-- 330's own test (bug file §7) needs its own tool-generator demand — and the
-- instrument-alive control then has to come from ANOTHER agent's live class
-- (bdl/`commit_sha` is the obvious one), because this file removes the local one.
--
-- The runner refuses `--record-only` on a _HOLD sidecar; record the apply in the
-- lane NOTES instead.
--
-- ROLLBACK: 516_tool_related_pages_optional_explicit_ROLLBACK.sql
-- ============================================================================

BEGIN;

SELECT snapshot_agent('tool-generator',
                      '516_tool_related_pages_optional_explicit: pre-update');
SELECT snapshot_agent('tool-deployer',
                      '516_tool_related_pages_optional_explicit: pre-update');

DO $$
DECLARE
    tgt record;
    cfg jsonb;
BEGIN
    FOR tgt IN
        SELECT * FROM (VALUES
            ('tool-generator', 'save_tool',   'create_tool_component'),
            ('tool-deployer',  'deploy_tool', 'deploy_tool_to_site')
        ) AS v(agent_type, step_name, expected_action)
    LOOP
        SELECT default_config #> ARRAY['workflow','steps',tgt.step_name,'config'] INTO cfg
          FROM agent_definitions
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        IF cfg IS NULL THEN
            RAISE EXCEPTION '516: no live % step config at %.% — refusing to guess',
                tgt.agent_type, 'workflow.steps', tgt.step_name;
        END IF;

        -- The step must still run the action whose spec was read. If it has been
        -- repointed, the absence-safety argument above is about other code.
        IF (SELECT default_config #>> ARRAY['workflow','steps',tgt.step_name,'action']
              FROM agent_definitions
             WHERE type = tgt.agent_type AND is_active
               AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL)
           IS DISTINCT FROM tgt.expected_action THEN
            RAISE EXCEPTION '516: %.% no longer runs % — re-measure before applying',
                tgt.agent_type, tgt.step_name, tgt.expected_action;
        END IF;

        IF cfg ? 'related_pages?' THEN
            RAISE EXCEPTION '516: %.% already declares related_pages? — already applied',
                tgt.agent_type, tgt.step_name;
        END IF;

        -- The unmarked wire is what we are CONVERTING, so its absence means the
        -- ground has moved (211 rolled back, or someone re-pointed the field).
        IF NOT (cfg ? 'related_pages') THEN
            RAISE EXCEPTION '516: %.% has no unmarked related_pages wire to convert — '
                'migration 211''s wiring is not where this file expects it',
                tgt.agent_type, tgt.step_name;
        END IF;
        IF cfg ->> 'related_pages' IS DISTINCT FROM 'input_data.spec.related_pages' THEN
            RAISE EXCEPTION '516: %.% wires related_pages to % , not input_data.spec.related_pages '
                '— read why before converting', tgt.agent_type, tgt.step_name, cfg ->> 'related_pages';
        END IF;

        -- Rename the key in place: same path, marked. Set the marked key first,
        -- then drop the unmarked one, so no intermediate state loses the wire.
        UPDATE agent_definitions
           SET default_config = jsonb_set(
                   default_config,
                   ARRAY['workflow','steps',tgt.step_name,'config','related_pages?'],
                   to_jsonb('input_data.spec.related_pages'::text),
                   true),
               updated_at = NOW()
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        UPDATE agent_definitions
           SET default_config = default_config
                   #- ARRAY['workflow','steps',tgt.step_name,'config','related_pages'],
               updated_at = NOW()
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    END LOOP;
END $$;

-- ============================================================================
-- VERIFY — a DO block, not a SELECT: ON_ERROR_STOP does not abort a COMMIT on a
-- non-empty result set, so a verify made of SELECTs cannot stop a bad apply
-- (LANDMINES / RFC_006).
-- ============================================================================
DO $$
DECLARE
    tgt record;
    got text;
    stale boolean;
BEGIN
    FOR tgt IN
        SELECT * FROM (VALUES
            ('tool-generator', 'save_tool'),
            ('tool-deployer',  'deploy_tool')
        ) AS v(agent_type, step_name)
    LOOP
        SELECT default_config #>> ARRAY['workflow','steps',tgt.step_name,'config','related_pages?'],
               default_config #> ARRAY['workflow','steps',tgt.step_name,'config'] ? 'related_pages'
          INTO got, stale
          FROM agent_definitions
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        IF got IS DISTINCT FROM 'input_data.spec.related_pages' THEN
            RAISE EXCEPTION '516 VERIFY FAILED: %.% related_pages? reads % , expected input_data.spec.related_pages',
                tgt.agent_type, tgt.step_name, COALESCE(got, '<null>');
        END IF;
        -- Both spellings present would leave the marker inert: ExtractActionInputs
        -- drops the unmarked twin of a marked key, but a reader of the config
        -- would see two wires and not know which governs.
        IF stale THEN
            RAISE EXCEPTION '516 VERIFY FAILED: %.% still carries the UNMARKED related_pages key',
                tgt.agent_type, tgt.step_name;
        END IF;
    END LOOP;

    IF (SELECT count(*) FROM agent_definitions_backup
         WHERE type IN ('tool-generator','tool-deployer')
           AND snapshot_reason LIKE '516_tool_related_pages%') < 2 THEN
        RAISE EXCEPTION '516 VERIFY FAILED: expected a snapshot row per agent in agent_definitions_backup';
    END IF;

    RAISE NOTICE '516 OK: both tool build steps declare related_pages? = input_data.spec.related_pages';
END $$;

COMMIT;
