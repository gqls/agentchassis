-- FILE: SQL_2026-08-15_d2_selftarget_heroes.sql
--
-- OWNER DECISION 2 (2026-08-15, in chat): for the 10 self-target hero
-- buttons left by the resolution re-run — "Add an in-page anchor for each
-- one, delete the useless ones."
--
-- DISPOSITION, decided by evidence not by list (queries in NOTES 2026-08-15
-- afternoon):
--   ANCHOR (1): gamesdesign game-jelly-invaders — the ONE page whose hero is
--     a plain hero ABOVE its game (game lives in a separate `section` slot,
--     id gameCanvas). cta_url := '#gameCanvas'; label "Play Jelly Invaders"
--     already present, so the template's {{if and .cta_text .cta_url}} gate
--     opens and a real button renders.
--   DELETE LABELS (7): auto-battler + economy-simulator (gamesdesign),
--     bridging-loan/fee-analyser/portfolio/rate-forecaster/stamp-duty
--     (mortgagecalculator) — on all seven the TOOL IS INSIDE THE HERO, the
--     label text appears NOWHERE in the rendered html, and the template has
--     no matching literal — i.e. the label keys are consumed only by the
--     never-opening anchor gate. Deleting them is render-neutral and stops
--     the census counting vestigial labels for ever.
--   KEEP, UNTOUCHED (2): mortgagecalculator tool-affordability and
--     vetcomparison index — their label text APPEARS in the render while the
--     template holds no such literal, so the data key is load-bearing
--     (tool-UI text: the calculator submit button / the search control).
--     Deleting would blank a live control. NOT useless; nothing to do.
--
-- History note: the archive trigger stores OLD on overwrite, so the deleted
-- keys remain recoverable from page_component_history.

BEGIN;

-- 1. The anchor.
UPDATE page_components pc SET
  content_data = pc.content_data || '{"cta_url":"#gameCanvas"}'::jsonb,
  updated_at = now()
FROM pages p
WHERE pc.page_id=p.id AND p.status='active' AND pc.slot_name='hero'
  AND p.site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND p.name='game-jelly-invaders'
  AND pc.locked_at IS NULL;

-- 2. The seven vestigial label deletions.
UPDATE page_components pc SET
  content_data = pc.content_data - 'cta_text' - 'secondary_cta' - 'primary_cta',
  updated_at = now()
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE pc.page_id=p.id AND p.status='active' AND pc.slot_name='hero'
  AND pc.locked_at IS NULL
  AND (s.domain, p.name) IN (
    ('gamesdesign.co.uk','game-auto-battler'),
    ('gamesdesign.co.uk','game-economy-simulator'),
    ('mortgagecalculator.co.uk','tool-bridging-loan'),
    ('mortgagecalculator.co.uk','tool-fee-analyser'),
    ('mortgagecalculator.co.uk','tool-portfolio'),
    ('mortgagecalculator.co.uk','tool-rate-forecaster'),
    ('mortgagecalculator.co.uk','tool-stamp-duty'));

DO $$
DECLARE n int;
BEGIN
  -- the anchor row now carries label AND url
  SELECT count(*) INTO n FROM page_components pc
   JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE s.domain='gamesdesign.co.uk' AND p.name='game-jelly-invaders'
     AND pc.slot_name='hero' AND p.status='active'
     AND pc.content_data->>'cta_url'='#gameCanvas'
     AND COALESCE(pc.content_data->>'cta_text','') <> '';
  IF n <> 1 THEN RAISE EXCEPTION 'jelly-invaders anchor not set (n=%)', n; END IF;

  -- the seven deletions took: no label key remains on those rows
  SELECT count(*) INTO n FROM page_components pc
   JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE p.status='active' AND pc.slot_name='hero'
     AND (s.domain,p.name) IN (
       ('gamesdesign.co.uk','game-auto-battler'),
       ('gamesdesign.co.uk','game-economy-simulator'),
       ('mortgagecalculator.co.uk','tool-bridging-loan'),
       ('mortgagecalculator.co.uk','tool-fee-analyser'),
       ('mortgagecalculator.co.uk','tool-portfolio'),
       ('mortgagecalculator.co.uk','tool-rate-forecaster'),
       ('mortgagecalculator.co.uk','tool-stamp-duty'))
     AND (pc.content_data ? 'cta_text' OR pc.content_data ? 'primary_cta');
  IF n <> 0 THEN RAISE EXCEPTION '% rows still carry a label key', n; END IF;

  -- the two KEEP rows are untouched: labels still present
  SELECT count(*) INTO n FROM page_components pc
   JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE p.status='active' AND pc.slot_name='hero'
     AND (s.domain,p.name) IN (('mortgagecalculator.co.uk','tool-affordability'),
                               ('vetcomparison.uk','index'))
     AND pc.content_data ? 'cta_text';
  IF n <> 2 THEN RAISE EXCEPTION 'KEEP rows disturbed (n=%)', n; END IF;
END $$;

COMMIT;
