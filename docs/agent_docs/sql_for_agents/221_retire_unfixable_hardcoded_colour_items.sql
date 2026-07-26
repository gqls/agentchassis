-- 221_retire_unfixable_hardcoded_colour_items.sql
--
-- bugs_open/077 — the data half. The code half (the detector now partitions its
-- population by the handler's remit and files the residue as a capability_gap)
-- ships in the same commit; this file cleans up what the OLD detector left
-- behind, and cannot do so before that code is live. See SEQUENCING below.
--
-- WHY
-- ---
-- The `hardcoded_section_colors` detector matched ANY hex background — 3/4/6/8
-- digit, light or dark, inline style="" attributes included — in any component
-- carrying a <style> tag. Its handler, ReplaceHardcodedColors, rewrites only DARK
-- SIX-DIGIT hexes and two-colour Ndeg gradients, and only INSIDE <style> blocks.
-- On a site where the detector's whole population sits outside that remit, the
-- fixer runs, changes nothing, and the two-strike rule
-- (load_work_item_actions.go:1041) parks the item as:
--
--     [unresolved after 2 attempts] Found 8 components with hardcoded hex colors …
--
-- That label asserts the handler FAILED. It never had anything to do. 'unresolved'
-- is the fleet-wide "needs investigation" signal for every other item type, so
-- rows that are correctly unfixable devalue it everywhere.
--
-- WHICH ROWS, AND WHY THE PREDICATE IS SHAPED LIKE THIS
-- ----------------------------------------------------
-- No hardcoded domain list. A site qualifies only if a SQL predicate that is
-- STRICTLY WIDER than ReplaceHardcodedColors on every axis finds nothing:
--
--   * no <style> boundary       (the Go transform only enters <style> blocks)
--   * no trailing-terminator    (the Go regexes require \s*[;}\n] after the hex)
--   * no restriction to the detector's own population (any unlocked component)
--
-- Everything the transform could change is therefore inside what this matches, so
-- a count of ZERO here proves a remit of zero there. It can only ever be
-- conservative: a site whose remit is genuinely empty might be skipped, but a site
-- with fixable work can never be retired. That direction is the one that matters —
-- retiring a fixable item would hide real work.
--
-- MEASURED LIVE 2026-07-26 (this exact predicate, per site carrying an item):
--
--   ai-agent-orchestration.com  unresolved   0   <- retired
--   finetuning.uk               unresolved   0   <- retired
--   gaswholesalers.com          unresolved   0   <- retired
--   leopardessconsulting.co.uk  unresolved   1       left alone
--   vonc.com                    unresolved   1       left alone
--   robot-hands.com             unresolved   3       left alone (x3 rows)
--   robot-hands.com             detected     3       left alone
--   gamesdesign.co.uk           complete     2       already terminal
--
-- So: 3 rows. CORRECTION to the figures in bugs_open/077's own table, which listed
-- webdesign.co.uk and dartsonline.com among the zero-remit sites: both have
-- detector matches but carry NO work item, so neither appears here; and
-- webdesign.co.uk is not provably zero under this wider predicate anyway.
--
-- WHY 'wont_fix' AND NOT DELETE
-- -----------------------------
-- 'wont_fix' is the existing terminal spelling of "no fix is possible" (written
-- today by apply_gap_plan_action.go:753 for not_actionable). It is in
-- workItemTerminalStatuses (work_items_common.go:29-44), so the row stops
-- occupying its idx_swi_dedup slot and the corrected detector is free to file the
-- honest capability_gap on its next pass. The finding itself is preserved — the
-- summary is rewritten to say what is actually true, and the spec records why.
--
-- SEQUENCING — APPLY THIS *AFTER* THE IMAGE ROLL, NOT BEFORE
-- ---------------------------------------------------------
-- Retiring the row frees the dedup slot. If the OLD detector is still deployed
-- when that happens, its next discovery pass over one of these sites re-files the
-- same dishonest item and the cleanup is undone. Roll the chassis past the commit
-- that ships the partitioning check, verify against the pod, then apply.
--
-- ROLLBACK
-- --------
--   UPDATE site_work_items
--      SET status  = 'unresolved',
--          summary = regexp_replace(summary, '^No fixable colours[^—]*— ', ''),
--          spec    = spec - 'retired_by' - 'retired_reason' - 'retired_at'
--    WHERE spec->>'retired_by' = '221_retire_unfixable_hardcoded_colour_items.sql';
--   -- Restores the label, not the truth of it.

BEGIN;

-- ---------------------------------------------------------------------------
-- Guard: refuse a second application rather than silently rewriting summaries
-- a second time (the leading-tag strip is not idempotent against its own output).
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM site_work_items
         WHERE spec->>'retired_by' = '221_retire_unfixable_hardcoded_colour_items.sql'
    ) THEN
        RAISE EXCEPTION '221: already applied — hardcoded_section_colors rows already carry the retirement marker';
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Retire the rows whose handler could never have cleared them.
-- ---------------------------------------------------------------------------
WITH zero_remit_sites AS (
    SELECT s.id
      FROM sites s
     WHERE NOT EXISTS (
        SELECT 1
          FROM page_components pc
          JOIN pages p ON pc.page_id = p.id
         WHERE p.site_id = s.id
           AND pc.locked_at IS NULL
           AND (
                -- strictly wider than the Go bgSingleRe / bgColorRe: no <style>
                -- boundary, no trailing terminator
                pc.rendered_html ~ 'background(-color)?\s*:\s*#[0-4][0-9a-fA-F]{5}'
                -- strictly wider than the Go gradientRe
             OR pc.rendered_html ~ 'linear-gradient\s*\(\s*[0-9]+deg\s*,\s*#[0-9a-fA-F]{3,8}\s*,\s*#[0-9a-fA-F]{3,8}\s*\)'
           )
     )
)
UPDATE site_work_items wi
   SET status  = 'wont_fix',
       summary = 'No fixable colours: the colour fixer''s remit is empty on this site — '
                 || regexp_replace(wi.summary, '^\[[^\]]*\]\s*', ''),
       spec    = COALESCE(wi.spec, '{}'::jsonb) || jsonb_build_object(
                    'retired_by',     '221_retire_unfixable_hardcoded_colour_items.sql',
                    'retired_reason', 'detector population entirely outside ReplaceHardcodedColors remit (bugs_open/077)',
                    'retired_at',     now()
                 ),
       updated_at = now()
 WHERE wi.item_type = 'hardcoded_section_colors'
   AND wi.item_key  = 'hardcoded_section_colors'          -- discovery-check rows only;
                                                          -- the design-audit producer uses a
                                                          -- different key shape and a
                                                          -- per-page, LLM-authored summary
   AND wi.status NOT IN ('complete', 'verified', 'wont_fix', 'rejected', 'cancelled')
   AND wi.site_id IN (SELECT id FROM zero_remit_sites);

-- ---------------------------------------------------------------------------
-- Report what was actually touched. Not a guard — a zero here is a legitimate
-- outcome on a cluster where the corrected detector has already run.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    n integer;
BEGIN
    SELECT count(*) INTO n
      FROM site_work_items
     WHERE spec->>'retired_by' = '221_retire_unfixable_hardcoded_colour_items.sql';
    RAISE NOTICE '221: retired % hardcoded_section_colors item(s) whose handler had an empty remit', n;
END $$;

COMMIT;
