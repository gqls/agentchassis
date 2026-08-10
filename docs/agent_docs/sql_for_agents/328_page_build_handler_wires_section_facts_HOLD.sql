-- *** _HOLD (2026-08-07): SLICE B of RFC_016 — do NOT apply in a blanket --apply run. ***
-- Council 902a8563 REJECTED shipping this alongside Slice A (guardian veto on
-- breadth). It applies only after: Slice A (327+329, applied 2026-08-07) has been
-- observed on real plan rows, this slice has its OWN council round, and a human
-- has read the v4 prompt plaintext (compliance seat's ask; 330 only). To apply:
-- rename away the _HOLD suffix, psql -f, then --record-only under the new name.
-- 328 — page-build-handler: wire plan-time fact assignments into plan_sections.
--
-- bugs_open/151 candidate 1, config half 1 of 3 (328 wiring, 329 planner prompt,
-- 330 writer prompt). Adds ONE key to the plan_sections step config:
--
--   "section_facts": "spec_sections.section_facts"
--
-- load_page_sections_from_spec emits section_facts (aligned with sections) ONLY
-- when its authoritative tier (site_plan_sections) served the list; plan_sections
-- consumes it ONLY when this config key names it — the feature is opt-in at the
-- step config, per the owner ruling of 2026-08-02 (new behaviour on a shared seam
-- ships as an opt-in field).
--
-- > CORRECTED 2026-08-10 (council round a06ff850, debug_historian): this file
-- > used to claim "exactly ONE live agent wires spec_sections into
-- > plan_sections (measured 2026-08-06)". Re-measured by ACTION, not step
-- > name: TWO agents run plan_sections (page-build-handler, page-content-
-- > writer) and NEITHER wires spec_sections. page-build-handler is still the
-- > right and only target, on different grounds, verified 2026-08-10:
-- > call_content_writer's input_mapping passes its section_plan to the
-- > writer, whose check_section_plan keeps a caller-supplied plan VERBATIM
-- > ("Build the section plan the caller did not supply", bugs_open/087) —
-- > and 30/30 writer runs in the retained orchestration window carried a
-- > caller-supplied plan, so the writer's own plan_sections is a fallback
-- > that does not fire on the build path. Transit preserves the stamped
-- > fields: resolve_internal_links mutates section maps IN PLACE and returns
-- > the same slice (resolve_internal_links_action.go), and select_sections'
-- > extract_fields copies whole entries.
--
-- *** DO NOT APPLY until a chassis image containing the Go half is LIVE ***
-- Pod check (both strings in ONE exec; the control is a pre-existing literal
-- from a DIFFERENT file, invariant under this change):
--   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
--     'grep -ac "assigned fact id matches no current evidence_base fact" /app/agent-chassis; \
--      grep -ac "site_plan_sections lookup failed" /app/agent-chassis'
--   -> first count >= 1 (new), second >= 1 (control). `strings` is absent from
--   these images — use grep -ac (LANDMINES:503).
-- Applying early is NOT fatal (PlanSectionsInputSpec is CheckConfig without
-- StrictConfig, so an old binary warns on the unknown key and continues), but
-- the key would be dead config until the roll — image first, config second.

SELECT snapshot_agent('page-build-handler', '328_page_build_handler_wires_section_facts.sql: pre-update');

BEGIN;

-- Ordering guard (owner ruling 2026-08-10, decision 4): the slice applies
-- 362 -> 328 -> 330, enforced here rather than by a filename or a handoff. A
-- bare SELECT cannot stop a COMMIT, hence DO/RAISE. Without 362 the planner
-- still re-composes built pages freely, and wiring consumption first would
-- measure a mechanism whose input is still being destroyed upstream.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'build-site-planner' AND is_active
          AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config::text LIKE '%Only when the briefing explicitly asks for a page to be redesigned%'
    ) THEN
        RAISE EXCEPTION '328: ordering precondition failed -- seed 362 (build-site-planner realised-sections prompt) is not applied; apply 362 -> 328 -> 330';
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_328 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'page-build-handler' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_sections,config,section_facts}',
        '"spec_sections.section_facts"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Guard: exactly one row updated and the key reads back.
DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'page-build-handler' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND default_config->'workflow'->'steps'->'plan_sections'->'config'->>'section_facts'
          = 'spec_sections.section_facts';
    IF n <> 1 THEN
        RAISE EXCEPTION '328: expected exactly 1 page-build-handler row carrying the section_facts wiring, found %', n;
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad
--   SET default_config = b.default_config
--   FROM agent_definitions_bak_328 b
--   WHERE ad.id = b.id;
-- or surgically:
--   UPDATE agent_definitions
--   SET default_config = default_config #- '{workflow,steps,plan_sections,config,section_facts}'
--   WHERE type='page-build-handler' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
