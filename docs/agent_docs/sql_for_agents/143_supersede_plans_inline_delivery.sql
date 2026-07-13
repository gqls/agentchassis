-- 143_supersede_plans_inline_delivery.sql — Option B: PLANs surrender to
-- delivered reality (decision 2026-07-10). DB-only. doc_plans supersede
-- (never edit history in place); no snapshot_agent (data tables, not
-- agent_definitions) — the retired rows ARE the rollback.
--
-- DECISION ON RECORD: every generated PLAN declared delivery "Path 1 —
-- extracted to /tools/assets/<fn>.js on rerender" with a matching asset_loads
-- criterion, but the render path ships all JS INLINE and the extraction was
-- never built. Tier-2's first live sweep (cd0d9731) therefore failed BOTH
-- tools on `asset`. Chosen remedy: criteria must describe what the system
-- DELIVERS, not what it aspires to — supersede the PLANs to inline delivery
-- and drop the asset check (144 fixes the composer so future PLANs are born
-- honest). If extraction ships later, PLANs supersede forward again.
--
-- Also fixed while superseding (drop-rate-tuner only): the interaction
-- check's selectors were composer-invented kebab-case (#drop-chance,
-- #stat-median); the generated tool uses camelCase. Corrected to the REAL ids
-- verified on the live page 2026-07-10: #dropChance (range input, min 0.1
-- max 25) and #statMedian. Second sighting of the invented-selector class —
-- exactly what the anchor rule was built to catch.
--
-- Closes the two Tier-2 work items: resolution was PLAN-side, no tool change
-- needed, so they are cancelled with the resolution recorded (leaving them
-- would send tool-improver chasing a stale contract).

BEGIN;

DO $$
DECLARE
    fn         text;
    old_id     uuid;
    old_body   text;
    new_body   text;
    asset_line text;
    s          int;
    p          int;
    rest       text;
    canon      text := '## Delivery mechanism' || E'\n' ||
        'Inline — all JS and CSS ship inside the page HTML. Superseded 2026-07-10 (decision on record): the Path 1 asset extraction (/tools/assets/<function>.js) was designed but never built; acceptance criteria must describe delivered reality, so the asset_loads check is removed. If extraction ships later, supersede forward again.' || E'\n\n';
BEGIN
    FOREACH fn IN ARRAY ARRAY['tool-xp-curve-designer','tool-drop-rate-tuner'] LOOP
        SELECT id, body INTO old_id, old_body FROM doc_plans
        WHERE subject_type='tool' AND subject_key=fn AND is_current;
        IF old_id IS NULL THEN
            RAISE EXCEPTION '143: no current PLAN for %', fn;
        END IF;

        -- 1) drop the asset check (exact byte-run guard, 136-style)
        asset_line := '{"id":"asset","type":"asset_loads","path":"/tools/assets/' || fn || '.js"},';
        IF strpos(old_body, asset_line) = 0 THEN
            RAISE EXCEPTION '143: asset check line not found verbatim in % PLAN — inspect before superseding', fn;
        END IF;
        new_body := replace(old_body, asset_line, '');

        -- 2) replace the whole Delivery mechanism section (robust to wrapping)
        s := strpos(new_body, '## Delivery mechanism');
        IF s = 0 THEN
            RAISE EXCEPTION '143: Delivery mechanism heading not found in % PLAN', fn;
        END IF;
        rest := substr(new_body, s);
        p := strpos(substr(rest, 2), E'\n## ');
        IF p = 0 THEN
            RAISE EXCEPTION '143: no section after Delivery mechanism in % PLAN', fn;
        END IF;
        new_body := substr(new_body, 1, s - 1) || canon || substr(rest, p + 2);

        -- 3) drop-rate-tuner only: correct the invented interaction selectors
        IF fn = 'tool-drop-rate-tuner' THEN
            IF strpos(new_body, '"selector":"#drop-chance"') = 0
               OR strpos(new_body, '"expect":{"selector":"#stat-median"}') = 0 THEN
                RAISE EXCEPTION '143: expected invented selectors not found in drop-rate PLAN';
            END IF;
            new_body := replace(new_body, '"selector":"#drop-chance"', '"selector":"#dropChance"');
            new_body := replace(new_body, '"expect":{"selector":"#stat-median"}', '"expect":{"selector":"#statMedian"}');
        END IF;

        -- 4) supersede: retire old, insert new current
        UPDATE doc_plans SET is_current = false, superseded_at = now()
        WHERE id = old_id;
        INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
        VALUES ('tool', fn, new_body, 'human', '143_supersede_plans_inline_delivery');

        -- 5) correction note in the tool's own stream
        INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, created_by)
        VALUES ('tool', fn, 'e33263f4-74f8-494f-b191-546845dbbddf',
            '## PLAN superseded: inline delivery (Option B) — ' || fn || E'\n' ||
            'Observed: Tier-2 sweep cd0d9731 failed the asset_loads check — /tools/assets/' || fn || '.js is never referenced; the JS ships inline.' || E'\n' ||
            'Root cause: the PLAN asserted the designed-but-never-built Path 1 extraction' ||
            CASE WHEN fn = 'tool-drop-rate-tuner'
                 THEN '; its interaction selectors were also composer-invented kebab-case (#drop-chance, #stat-median) vs the real camelCase ids (#dropChance, #statMedian).'
                 ELSE '.' END || E'\n' ||
            'Fix: Delivery mechanism superseded to inline; asset check removed' ||
            CASE WHEN fn = 'tool-drop-rate-tuner'
                 THEN '; interaction selectors corrected to ids verified on the live page.'
                 ELSE '.' END || E'\n' ||
            'Verified: anchors probed against the deployed page 2026-07-10; the Tier-2 work item is cancelled (PLAN-side resolution, no tool change).' || E'\n' ||
            'Categories: migration',
            '["migration"]'::jsonb, 'human', '143_supersede_plans_inline_delivery');

        RAISE NOTICE '143: % superseded (old %, new body % chars)', fn, old_id, length(new_body);
    END LOOP;
END $$;

-- 6) cancel the two Tier-2 items (tolerant: they may have been claimed)
UPDATE site_work_items
SET status = 'cancelled',
    result = jsonb_build_object('resolution',
      'Resolved by PLAN supersede (143, Option B): asset criterion removed (inline delivery decision); drop-rate interaction selectors corrected. No tool change needed.'),
    updated_at = now()
WHERE item_key IN (
    'tool_acceptance:tool-xp-curve-designer:e33263f4-74f8-494f-b191-546845dbbddf',
    'tool_acceptance:tool-drop-rate-tuner:e33263f4-74f8-494f-b191-546845dbbddf')
  AND status IN ('detected','triaged');

-- Guards: two current PLANs, fences intact, no asset check, no invented ids.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM doc_plans
    WHERE subject_type='tool'
      AND subject_key IN ('tool-xp-curve-designer','tool-drop-rate-tuner')
      AND is_current
      AND body LIKE '%```criteria%'
      AND strpos(body, '"type":"asset_loads"') = 0
      AND strpos(body, 'Inline — all JS and CSS ship inside the page HTML') > 0;
    IF n <> 2 THEN RAISE EXCEPTION '143: expected 2 corrected current PLANs, found %', n; END IF;

    SELECT count(*) INTO n FROM doc_plans
    WHERE subject_key='tool-drop-rate-tuner' AND is_current
      AND strpos(body, '#dropChance') > 0 AND strpos(body, '#statMedian') > 0
      AND strpos(body, '#drop-chance') = 0;
    IF n <> 1 THEN RAISE EXCEPTION '143: drop-rate selector correction incomplete'; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT subject_key, is_current, length(body), body LIKE '%asset_loads%' AS still_has_asset
--   FROM doc_plans WHERE subject_key IN ('tool-xp-curve-designer','tool-drop-rate-tuner')
--   ORDER BY subject_key, created_at;
-- Rollback: flip is_current back to the retired rows and delete the new ones
--   (ids in the NOTICEs above); un-cancel the items.
