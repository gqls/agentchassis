-- 571 — improvement-loop's two site-wide `create_work_item` steps DECLARE that they
--       supply no page and no component: `page_id?` / `component_id?` — the caller's own
--       path or NOTHING, never the whole-tree search. CONFIG ONLY — live on apply.
--
--       Found by the RFC_029 step-5 gate on its first day live (2026-08-22 18:34–18:40Z),
--       which is the first time this defect was ever visible from outside.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- `improvement-loop` runs a site sweep: it calls the discovery agents, then files ONE
-- site-wide work item — "Re-assemble and deploy pages after improvement fixes"
-- (`insert_rerender_item`), or a "not converging" note (`record_not_converging`). Both are
-- items ABOUT A WHOLE SITE. Neither is about a page.
--
-- The action they call, `create_work_item`, declares `page_id` and `component_id` as
-- OPTIONAL inputs (create_work_item_action.go: Optional = spec_data, parent_item_id,
-- page_id, component_id, summary). Other callers legitimately file page-scoped items.
-- These two steps wire NEITHER — and an Optional field that no config wires does not
-- become absent. It falls through to the whole-tree search, which hunts the run's entire
-- collected_data for anything called `page_id`.
--
-- By that point the tree holds the completeness-discovery report: **65 findings carrying 34
-- DISTINCT page ids** (measured 2026-08-23 on orchestration 64dae8dd, robot-hands.com).
-- Pre-flip the search ranked them and took `findings[0].page_id` — **one arbitrary page's
-- id, attached to a site-wide rerender item**. Silently: no error, no warning, a `complete`
-- run.
--
-- ============================================================================
-- THE HYPOTHESIS THIS REFUTES — it is NOT an identifier problem (owner's question, 08-23)
-- ============================================================================
-- Asked whether the same page ids appear on different sites, and whether the estate needs a
-- joint `[site, page]` key or a better page-id scheme. **Measured, and no:**
--
--   839 pages / 839 distinct ids           -> `pages.id` is a globally unique UUID (PK)
--   ids appearing on >1 site: 0            -> cross-site id collision does not happen
--   (site_id, name) is ALREADY unique      -> the joint key the question imagines exists
--   the 34 ambiguous ids: all ONE site     -> the ambiguity is entirely WITHIN one site
--
-- Page NAMES do repeat across sites (`index` on 28 sites, `about` on 20), so a name without
-- a site is ambiguous — but an ID never is. A better identifier scheme would fix NOTHING
-- here: every one of the 34 candidates was a valid, distinct, correctly-scoped page id. The
-- request was underspecified, not the identity. "Give me page_id" has no answer when the
-- tree legitimately holds 34 of them, and no identifier design can supply one.
--
-- **So the fix belongs where the ambiguity is: in what the step DECLARES it takes.**
--
-- ============================================================================
-- WHAT THIS CHANGES
-- ============================================================================
-- Both steps gain `"page_id?"` and `"component_id?"` pointing at `input_data.<field>` —
-- the sweep's own request. Today improvement-loop's input_data carries neither (0 of 2
-- retained runs, as of 2026-08-23), so both resolve to ABSENCE, which is the correct value
-- for a site-wide item. If a future caller ever scopes a sweep to one page, that page — and
-- only that page — is what the item will carry.
--
-- Unlike migration 516, this does NOT convert an existing wire: there is no unmarked twin to
-- remove. It adds a declaration where the absence of one was the defect.
--
-- ⚠ WHY NOT FIX `create_work_item`'S SPEC INSTEAD: other callers legitimately supply
-- page_id/component_id. Dropping them from Optional would break those; this is a per-CALLER
-- statement and belongs in the caller's config.
--
-- ⚠ WHAT THIS DOES NOT DO: it fixes two steps of one agent, not the class. The measured
-- population is **137 of 298 live steps (46%) that run an action WITH a spec and leave >=1
-- declared field unwired** (as of 2026-08-21, the flip's council round). The flip already
-- makes every one of them refuse rather than guess; per-step declarations like this one are
-- the follow-up, and `bugs_open/330` candidate 2 is where that class is tracked.
--
-- ORDERING: none. The `?` parser has been live since v1.0.1321 (2026-08-20), so this
-- migration's keys are parsed the moment they land. No _HOLD needed.
--
-- ROLLBACK: 571_improvement_loop_declares_it_supplies_no_page_ROLLBACK.sql
-- ============================================================================

BEGIN;

DO $do$
DECLARE
    tgt        record;
    cfg        jsonb;
    changed    int := 0;
    verified   int := 0;
BEGIN
    FOR tgt IN
        SELECT s.key AS step_name
          FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'create_work_item'
         ORDER BY s.key
    LOOP
        SELECT s.value->'config' INTO cfg
          FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.key = tgt.step_name;

        IF cfg IS NULL THEN
            RAISE EXCEPTION '571: improvement-loop.%.config is missing — refusing to guess', tgt.step_name;
        END IF;

        -- Idempotency: already-marked is not a failure, but it must not be silent.
        IF cfg ? 'page_id?' OR cfg ? 'component_id?' THEN
            RAISE NOTICE '571: improvement-loop.% already declares a marked key — skipping', tgt.step_name;
            CONTINUE;
        END IF;

        -- The premise of this migration is that these steps wire NEITHER field. If a future
        -- author has wired one, their intent outranks this file's assumption: stop and be read.
        IF cfg ? 'page_id' OR cfg ? 'component_id' THEN
            RAISE EXCEPTION '571: improvement-loop.% already wires page_id/component_id unmarked '
                '(page_id=%, component_id=%) — read why before converting',
                tgt.step_name, cfg->>'page_id', cfg->>'component_id';
        END IF;

        -- The site-wide premise itself, asserted rather than assumed: these steps must be
        -- filing a SITE item. A step that had gained a page scope would make this fix wrong.
        IF cfg->>'site_id' IS DISTINCT FROM 'site_record.site_id' THEN
            RAISE EXCEPTION '571: improvement-loop.% does not wire site_id to site_record.site_id (got %) '
                '— its scope has changed; re-read before applying',
                tgt.step_name, cfg->>'site_id';
        END IF;

        UPDATE agent_definitions
           SET default_config = jsonb_set(
                   jsonb_set(default_config,
                       ARRAY['workflow','steps',tgt.step_name,'config','page_id?'],
                       to_jsonb('input_data.page_id'::text), true),
                   ARRAY['workflow','steps',tgt.step_name,'config','component_id?'],
                   to_jsonb('input_data.component_id'::text), true),
               updated_at = NOW()
         WHERE type = 'improvement-loop' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        changed := changed + 1;
        RAISE NOTICE '571: improvement-loop.% now declares page_id? / component_id?', tgt.step_name;
    END LOOP;

    -- A discovery-driven migration that finds NOTHING must shout, not succeed quietly.
    IF changed = 0 THEN
        SELECT count(*) INTO verified
          FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'create_work_item'
           AND s.value->'config' ? 'page_id?';
        IF verified > 0 THEN
            RAISE NOTICE '571: ALREADY APPLIED — % step(s) already marked, nothing to do', verified;
        ELSE
            RAISE EXCEPTION '571: found ZERO create_work_item steps on improvement-loop — the walk is '
                'wrong or the agent has changed shape. Refusing to report success on an empty apply';
        END IF;
    END IF;

    -- ── VERIFY, in the same transaction, with a NEGATIVE CONTROL ────────────
    -- 1. every create_work_item step now carries BOTH marked keys with the right values
    SELECT count(*) INTO verified
      FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
     WHERE d.type = 'improvement-loop' AND d.is_active
       AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
       AND s.value->>'action' = 'create_work_item'
       AND s.value->'config'->>'page_id?' = 'input_data.page_id'
       AND s.value->'config'->>'component_id?' = 'input_data.component_id';

    IF verified < 2 THEN
        RAISE EXCEPTION '571: only % of the 2 expected create_work_item steps carry both marked keys', verified;
    END IF;

    -- 2. NEGATIVE CONTROL: the rest of each config must be intact. A jsonb_set that
    --    clobbered a sibling key would still pass check 1.
    IF EXISTS (
        SELECT 1 FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'create_work_item'
           AND NOT (s.value->'config' ? 'site_id'
                AND s.value->'config' ? 'item_type'
                AND s.value->'config' ? 'source'
                AND s.value->'config' ? 'item_key_prefix'
                AND s.value->'config' ? 'spec_literal')
    ) THEN
        RAISE EXCEPTION '571: a pre-existing config key was lost from a create_work_item step';
    END IF;

    -- 3. NEGATIVE CONTROL: no UNMARKED twin may exist — it would mean the marker is not the
    --    only authority for the field and the search could still be reached.
    IF EXISTS (
        SELECT 1 FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'create_work_item'
           AND (s.value->'config' ? 'page_id' OR s.value->'config' ? 'component_id')
    ) THEN
        RAISE EXCEPTION '571: an unmarked page_id/component_id twin survives — the marker is not sole authority';
    END IF;

    RAISE NOTICE '571: applied. improvement-loop files site-wide items that take a page only from '
        'their own request, never from a search of the tree.';
END
$do$;

COMMIT;
