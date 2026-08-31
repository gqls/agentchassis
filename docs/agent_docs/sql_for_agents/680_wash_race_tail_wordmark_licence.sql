-- 680: the race tail of 670 (bugs_open/417). Migration 669 fixed build-site-planner's
-- worked-example logo prompt at 2026-08-31 12:36:14+00; site_plan_imagery row
-- b56182fa-cdfe-4b9a-b1c8-606ea9fa39ea (boxingonline.com, the FIRST PAID SITE) was
-- created at 12:36:55+00 — 41 seconds later, by a planner run already in flight — and
-- so carries the old licence. Its logo asset (20ce80fb-…, created 12:56:10+00) was
-- generated from it and reads "BOXING NEWS" on a site called Boxing Online: the
-- farmerinsurance "Farm Shield Info" mechanism, second occurrence, live on 19 pages.
--
-- WHY 670 COULD NOT HAVE CAUGHT IT, and why this is a class point rather than a
-- straggler: the row says "no text OTHER THAN the wordmark itself"; the exemplar said
-- "no text OUTSIDE the wordmark itself". 670's surgical arm keyed on the literal
-- `outside`, so a PARAPHRASE was invisible to it. The model rewords the exemplar.
-- Therefore any fix that matches this licence by literal string is a FLOOR, not a
-- bound — which is the argument for the Go guard this migration accompanies
-- (generate_image_actions.go applies the policy to every kind=logo generation, at the
-- one point every prompt from every producer must pass). This migration is only the
-- PRE-ROLL protection: the Go half is inert until the next chassis build rolls, and a
-- regeneration for this site may fire before then (owner ruled it fixed before the
-- delivery email).
--
-- Shape follows 670 arm (b): an APPENDED dated governing clause, idempotent-guardable,
-- leaving the original wording auditable rather than rewriting it away.
--
-- DELIBERATELY NOT INCLUDED — a widened regex "safety net" arm over other unclaused
-- rows. The concept census of all 28 current-plan logo prompts (2026-08-31) found
-- exactly ONE row carrying licence-without-name, and ~8 rows that name their exact
-- wordmark ON PURPOSE (cv1 'CareerPrep', idea.uk, oufe, relojistas, robot-hands,
-- webdesign.uk, lendzy, loanzy). A regex broad enough to catch a paraphrase is broad
-- enough to void a deliberate worded mark, and that damage would be silent. Re-running
-- the concept census by hand is the correct instrument; a regex is not.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE v_n int;
BEGIN
  UPDATE site_plan_imagery spi
     SET prompt = prompt || ' — OWNER RULING 2026-08-31 (migration 680, race tail of 670): render a text-free mark with no lettering or words of any kind; any earlier wording in this prompt that permits or presupposes a wordmark is void. The brand name is set in HTML beside the logo, never rendered in the image.'
    FROM site_plans sp
   WHERE sp.id = spi.plan_id
     AND spi.id = 'b56182fa-cdfe-4b9a-b1c8-606ea9fa39ea'
     AND spi.kind = 'logo'
     AND sp.is_current
     AND spi.locked_at IS NULL
     AND spi.prompt LIKE '%no text other than the wordmark itself%'
     AND spi.prompt NOT LIKE '%migration 6%';
  GET DIAGNOSTICS v_n = ROW_COUNT;
  IF v_n <> 1 THEN
    RAISE EXCEPTION '680: expected exactly 1 race-tail row, updated % — already washed, locked, superseded, or the prompt text changed. Re-run the CONCEPT census (not a literal grep) before forcing.', v_n;
  END IF;
  RAISE NOTICE '680 applied: race-tail row b56182fa washed with the 670 governing clause';
END $$;
COMMIT;
