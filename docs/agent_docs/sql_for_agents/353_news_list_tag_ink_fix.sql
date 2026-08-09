-- 353 — .news-list-tag: use `text` as the ink instead of `text_muted`
--
-- WHY. The shared `news-listing` component styles its topic chips as
--   `color: var(--color-text-muted); background: var(--color-border);`
-- i.e. the MUTED slot as an ink on top of the BORDER slot used as a fill.
-- Neither slot was authored for that job, and the pair fails WCAG AA on
-- 7 of the 8 sites that render the component. Measured 2026-08-09 from the
-- SERVED stylesheets (not from site_specs):
--
--   site                        today   after this change
--   idea.uk                     2.25    9.67
--   robot-hands.com             3.47    9.38
--   relojistas.com              3.84   12.55
--   dartsonline.com             3.94   10.94
--   webdesign.co.uk             4.13   11.99
--   fundamentallyai.com         4.30   11.47
--   gaswholesalers.com          4.32   11.01
--   ai-agent-orchestration.com  4.95   12.88   (the only one passing today)
--
-- AA needs 4.5 for body text. After: 8 of 8 pass, minimum 9.38.
-- The no-theme fallback pair improves too: #64748b on #e2e8f0 = 3.86 (fails)
-- becomes #475569 on #e2e8f0 = 6.15.
--
-- WHY ONLY THE INK, AND NOT THE FILL AS WELL. bugs_open/122's 2026-07-29 note
-- proposed "surface as its fill and text as its ink". The ink half is right and
-- is what this does. The FILL half was measured and rejected: `surface` against
-- the section's `background` is 1.04–1.22 on all eight sites, so a surface-filled
-- chip stops reading as a chip at all. Keeping `border` as the fill preserves the
-- pill's existing distinctness (1.11–1.62, unchanged) and needs one declaration
-- rather than two. Both candidates clear AA on 8/8; this is the smaller change to
-- a shared fleet component.
--
-- SCOPE. The template uses --color-text-muted five times. This changes ONE of
-- them (the chip). The other four are muted text on the section BACKGROUND —
-- the slot's designed pairing — and none of them appears in the render audit's
-- failure list. The anchor below (the colour line PLUS the background line that
-- follows it) occurs exactly once; verified before writing this file.
--
-- BLAST RADIUS. 9 placements across 8 sites, all deployed. This is config, so it
-- is live on write — but it only reaches a visitor when a page RE-RENDERS: the
-- rule ships twice, once inline in each page's <style> (from this template) and
-- once in styles.css (frozen, bugs_closed/072). The inline copy is later in the
-- document than the <link> (byte 38425 vs 8412 on robot-hands /news/), so at
-- equal specificity the inline copy WINS. A page re-render therefore repairs the
-- page WITHOUT needing a stylesheet rebuild. Until a page re-renders it keeps
-- the old chip; nothing gets worse in the meantime.
--
-- BACKUP taken before this runs: bak_cc_newslisting_20260809 (1 row, 4462 bytes).
-- ROLLBACK: 353_news_list_tag_ink_fix_ROLLBACK.sql
--
-- Filed from bugs_open/113's three-site audit; the class belongs to bugs_open/122,
-- where this component is 181 of that bug's 442 measured failures (41%).

\set ON_ERROR_STOP on
BEGIN;

-- Guard: refuse if the anchor is not present exactly once (someone else edited it).
DO $guard$
DECLARE n int;
BEGIN
  SELECT (length(html_template) - length(replace(html_template,
            E'color: var(--color-text-muted, #64748b);\n  background: var(--color-border, #e2e8f0);','')))
         / length(E'color: var(--color-text-muted, #64748b);\n  background: var(--color-border, #e2e8f0);')
    INTO n
    FROM content_components WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';
  IF n IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION 'news-list-tag anchor found % times, expected 1 — template changed since 2026-08-09, re-measure before applying', n;
  END IF;
END
$guard$;

UPDATE content_components
   SET html_template = replace(html_template,
         E'color: var(--color-text-muted, #64748b);\n  background: var(--color-border, #e2e8f0);',
         E'color: var(--color-text, #475569);\n  background: var(--color-border, #e2e8f0);'),
       updated_at = now()
 WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';

-- Verify: the chip line is gone, the other four muted uses survive, size is -6 bytes.
DO $verify$
DECLARE muted int; bytes int; chip int;
BEGIN
  SELECT (length(html_template)-length(replace(html_template,'--color-text-muted','')))/length('--color-text-muted'),
         length(html_template),
         (length(html_template)-length(replace(html_template,E'color: var(--color-text, #475569);\n  background: var(--color-border, #e2e8f0);','')))
         / length(E'color: var(--color-text, #475569);\n  background: var(--color-border, #e2e8f0);')
    INTO muted, bytes, chip
    FROM content_components WHERE id = '11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f';
  IF chip <> 1        THEN RAISE EXCEPTION 'new chip rule not written (found % times)', chip; END IF;
  IF muted <> 4       THEN RAISE EXCEPTION 'expected 4 remaining --color-text-muted uses, found %', muted; END IF;
  IF bytes <> 4456    THEN RAISE EXCEPTION 'expected 4456 bytes after edit, got %', bytes; END IF;
  RAISE NOTICE 'OK: chip ink now --color-text, % muted uses left, % bytes', muted, bytes;
END
$verify$;

COMMIT;
