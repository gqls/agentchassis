-- 670: the 19 current-plan logo prompts that already carry build-site-planner's wordmark
-- licence (bugs_open/417 §2-§3: 27 current-plan logo prompts fleet-wide; 19 mention
-- wordmark, 10 verbatim "no text outside the wordmark"). OWNER RULING 2026-08-31 (loanzy
-- lane, decision 2): REWRITE the stored instructions — 669 alone stops the next sites and
-- repairs none of these. Two edits, both surgical:
--   (a) the 10 verbatim rows: the licensing phrase is replaced outright;
--   (b) ALL wordmark-mentioning rows: a dated governing clause is APPENDED, voiding any
--       remaining wordmark language in situ (varied phrasings make surgical edits per-row
--       unsafe; an appended override is idempotent-guardable and leaves the original
--       auditable). Nothing regenerates now — the rewritten prompt is only read at the
--       next logo (re)generation, so this changes no served pixels today.
-- Locked rows are excluded by predicate (0 exist today — measured 2026-08-31 — the
-- predicate is belt for the day one appears).
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE v_expected int; v_done int;
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='bak_670_plan_imagery_wordmark') THEN
    RAISE EXCEPTION '670 probe: backup table exists — already applied';
  END IF;
  SELECT count(*) INTO v_expected FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id=spi.plan_id
   WHERE spi.kind='logo' AND sp.is_current AND spi.locked_at IS NULL
     AND spi.prompt ILIKE '%wordmark%';
  IF v_expected = 0 THEN
    RAISE EXCEPTION '670 probe: zero wordmark-mentioning current logo prompts — nothing to do (already applied?)';
  END IF;
  EXECUTE 'CREATE TABLE bak_670_plan_imagery_wordmark AS
    SELECT spi.* FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id
    WHERE spi.kind=''logo'' AND sp.is_current AND spi.locked_at IS NULL
      AND spi.prompt ILIKE ''%wordmark%''';
  -- (a) replace the verbatim licence, longest variant first ("… itself", then the bare
  -- phrase — rehearsal caught 3 rows carrying it without "itself")
  UPDATE site_plan_imagery spi
     SET prompt = replace(replace(prompt,
           'no text outside the wordmark itself', 'no lettering or words of any kind'),
           'no text outside the wordmark',        'no lettering or words of any kind')
    FROM site_plans sp
   WHERE sp.id=spi.plan_id AND spi.kind='logo' AND sp.is_current AND spi.locked_at IS NULL
     AND spi.prompt LIKE '%no text outside the wordmark%';
  -- (b) append the governing clause to every wordmark-mentioning row not yet carrying it
  UPDATE site_plan_imagery spi
     SET prompt = prompt || ' — OWNER RULING 2026-08-31 (migration 670): render a text-free mark with no lettering or words of any kind; any earlier wording in this prompt that permits or presupposes a wordmark is void. The brand name is set in HTML beside the logo, never rendered in the image.'
    FROM site_plans sp
   WHERE sp.id=spi.plan_id AND spi.kind='logo' AND sp.is_current AND spi.locked_at IS NULL
     AND spi.prompt ILIKE '%wordmark%'
     AND spi.prompt NOT LIKE '%migration 670%';
  -- verify (counts the LICENCE, not the prohibition — 417 §2):
  SELECT count(*) INTO v_done FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id=spi.plan_id
   WHERE spi.kind='logo' AND sp.is_current
     AND spi.prompt LIKE '%no text outside the wordmark%';
  IF v_done <> 0 THEN
    RAISE EXCEPTION '670 verify: % rows still carry the verbatim licence', v_done;
  END IF;
  SELECT count(*) INTO v_done FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id=spi.plan_id
   WHERE spi.kind='logo' AND sp.is_current AND spi.locked_at IS NULL
     AND spi.prompt ILIKE '%wordmark%' AND spi.prompt NOT LIKE '%migration 670%';
  IF v_done <> 0 THEN
    RAISE EXCEPTION '670 verify: % wordmark-mentioning rows missed the override', v_done;
  END IF;
  RAISE NOTICE '670 applied: % rows backed up and rewritten', v_expected;
END $$;
COMMIT;
