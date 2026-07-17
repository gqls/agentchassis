-- 161_supersede_loot_plan_interaction_shape.sql — the loot-table PLAN gets a
-- REAL interaction check (the 148 precedent). DB-only. doc_plans supersede
-- (never edit history in place); the retired row IS the rollback.
--
-- WHY. tool-loot-table-balancer was the first tool born on claude-sonnet-5
-- (2026-07-17, orchestration f81bdcc9, chassis v1.0.1128). Everything landed
-- correctly EXCEPT the shape of the interaction check: compose_plan emitted
--
--   {"id":"add-item","type":"click","selector":"#ltbAddItem",
--    "expect":"#ltbRows .ltb-row:nth-child(4)"}
--
-- "click" is not a check type in the Tier-4 vocabulary (the types are
-- page_status_ok / selector_exists / selector_count / no_console_errors /
-- no_horizontal_overflow / interaction), and "expect" is an OBJECT, not a
-- string. The runner therefore SKIPS it honestly ("click not implemented") —
-- no fake pass, but the tool's actual behaviour goes untested, which is the
-- one thing Tier 4 exists to prevent.
--
-- The selectors themselves were REAL (#ltbAddItem, #ltbRows, .ltb-row all
-- occur in the generated HTML): the no-invention rule held on the new model.
-- Only the shape was improvised — because the composer prompt described
-- interactions in prose and never showed the JSON. Every well-formed
-- interaction check to date was hand-written in a migration (143/148), so the
-- gap had never been exercised. Migration 160 fixed the prompt; this migration
-- fixes the one PLAN that was born before it.
--
-- PROBED BEFORE AUTHORED (the 148 rule — never write a criterion you have not
-- watched pass). Real Chromium against the live deployed page, 2026-07-17:
--   * #ltbRows .ltb-row:nth-child(4) does NOT exist before the click (so the
--     expect is not vacuously true), on desktop AND mobile;
--   * clicking #ltbAddItem produces it: "interaction produced the expected
--     result", on desktop AND mobile.

BEGIN;

DO $$
DECLARE
    fn        text := 'tool-loot-table-balancer';
    site      uuid := 'e33263f4-74f8-494f-b191-546845dbbddf';  -- gamesdesign.co.uk
    old_id    uuid;
    old_body  text;
    new_body  text;
    rest      text;
    s         int;
    e         int;
    -- Dollar-quoted: preserves backslashes and newlines verbatim.
    new_fence text := $fence$```criteria
{ "profiles": ["desktop","mobile"],
  "container": ".tool-container",
  "checks": [
    {"id":"boots","type":"selector_exists","selector":".tool-container"},
    {"id":"console","type":"no_console_errors"},
    {"id":"status","type":"page_status_ok"},
    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
    {"id":"add-item","type":"interaction",
      "steps":[{"action":"click","selector":"#ltbAddItem"}],
      "expect":{"selector":"#ltbRows .ltb-row:nth-child(4)"}}
  ]
}
```$fence$;
BEGIN
    SELECT id, body INTO old_id, old_body FROM doc_plans
    WHERE subject_type='tool' AND subject_key=fn AND is_current;
    IF old_id IS NULL THEN
        RAISE EXCEPTION '161: no current PLAN for %', fn;
    END IF;

    -- Guard: we are superseding the version we actually diagnosed.
    IF strpos(old_body, '"type":"click"') = 0 THEN
        RAISE EXCEPTION '161: the mis-shaped click check is not in the % PLAN — re-inspect before superseding', fn;
    END IF;

    -- Replace the WHOLE criteria fence (line surgery is what produced half of
    -- these defects).
    s := strpos(old_body, '```criteria');
    IF s = 0 THEN
        RAISE EXCEPTION '161: no ```criteria fence in the % PLAN', fn;
    END IF;
    rest := substr(old_body, s);
    e := strpos(substr(rest, 12), chr(10) || '```');   -- closing fence, past the opener
    IF e = 0 THEN
        RAISE EXCEPTION '161: unterminated criteria fence in the % PLAN', fn;
    END IF;
    e := e + 11;                                       -- offset back into rest
    new_body := substr(old_body, 1, s - 1) || new_fence || substr(rest, e + 4);

    UPDATE doc_plans SET is_current = false, superseded_at = now() WHERE id = old_id;
    INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
    VALUES ('tool', fn, new_body, 'human', '161_supersede_loot_plan_interaction_shape');

    INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, created_by)
    VALUES ('tool', fn, site,
        '## PLAN superseded: interaction check reshaped to the Tier-4 vocabulary — ' || fn || E'\n' ||
        'Observed: this tool was the first born on claude-sonnet-5 (2026-07-17). Its birth PLAN carried {"id":"add-item","type":"click","selector":"#ltbAddItem","expect":"<string>"} — "click" is not a Tier-4 check type and "expect" must be an object, so the runner skipped the check and the tool''s behaviour went untested.' || E'\n' ||
        'Root cause: the compose_plan prompt described interaction checks in prose without showing the JSON shape; every well-formed interaction check to date was hand-written in a migration (143/148), so the gap had never been exercised. The selectors were REAL — the no-invention rule held; only the shape was improvised.' || E'\n' ||
        'Fix: check reshaped to {"type":"interaction","steps":[{"action":"click","selector":"#ltbAddItem"}],"expect":{"selector":"#ltbRows .ltb-row:nth-child(4)"}} (this migration). The composer itself was fixed at birth by migration 160, so future PLANs emit the right shape.' || E'\n' ||
        'Verified: probed in real Chromium against the live deployed page before being written here — the 4th row does NOT exist before the click (the expect is not vacuously true) and the click produces it, on desktop AND mobile.' || E'\n' ||
        'Categories: migration',
        '["migration"]'::jsonb, 'human', '161_supersede_loot_plan_interaction_shape');

    RAISE NOTICE '161: % superseded (old %, new body % chars)', fn, old_id, length(new_body);
END $$;

COMMIT;
