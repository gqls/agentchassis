-- 672_d4_governor_stage_a_hardening.sql — the 671 council round-1 revisions (corr 80df0963).
-- Three seats' asks, each folded in rather than argued with:
--   · editquality/debug_historian (low): scheduled_tasks rows are NOT single-flight — two
--     overlapping runs of spend-governor-state could race the old/new level comparison and
--     double-write or miss a level-change note. Remedy: a transaction-scoped advisory lock,
--     woven into the FIRST CTE the chain references (an unreferenced pure-SELECT CTE may
--     never be evaluated — the lock must sit on the spine).
--   · debug_historian (low): the 671 verify proved the stored text RUNS but never exercised
--     the level-change INSERT arm (budget NULL → always level 0). Remedy: this file's verify
--     drives a synthetic level change end-to-end (budget 0.01 → level 3 + note fires →
--     budget back to NULL → level 0 + note fires) and then deletes its own two synthetic
--     doc_notes rows by captured id — the probe cleans up after itself.
--   · tooling_provenance (low): the new pipeline had no travelling PLAN. Remedy: an initial
--     doc_plans row (subject 'pipeline'/'spend-governor') pointing at the lane PLAN §D4 and
--     the migrations, so the next session has a plan to load, not just DDL to reverse-read.
-- The reuse_agent/prior_art HIGH+MEDIUM objections are answered in the resubmission with
-- evidence, not with edits: platform/governance/fuel.go is a per-TASK fuel header budget
-- (74 lines, abstract units, two call sites: coordinator + content-creator) — orthogonal to
-- an account-month dollar governor; fleet-step-token-pressure and council-seat-token-pressure
-- audit per-step/per-seat max_tokens CAPS against truncation — neither prices tokens nor
-- aggregates spend. Full quotes in the round-2 submission.
--
-- Guard: refuses unless the live pre_query md5 is exactly 671's shipped text
-- (1c371a335ea3ca97c661e60164047396) — a drifted row means another hand edited it; stop.

BEGIN;

DO $$
DECLARE m text;
BEGIN
  SELECT md5(pre_query) INTO m FROM scheduled_tasks WHERE name='spend-governor-state';
  IF m IS NULL THEN
    RAISE EXCEPTION '672 REFUSED: spend-governor-state task not found — 671 not applied.';
  END IF;
  IF m = '838f8cd1cad9705f9e6651cf04dafab6' THEN
    RAISE EXCEPTION '672 REFUSED: already applied (replay) — the stored text is already the hardened one.';
  END IF;
  IF m <> '1c371a335ea3ca97c661e60164047396' THEN
    RAISE EXCEPTION '672 REFUSED: pre_query md5 % is not 671''s shipped text — drifted, investigate before overwriting.', m;
  END IF;
  IF EXISTS (SELECT 1 FROM doc_plans WHERE subject_type='pipeline' AND subject_key='spend-governor' AND is_current) THEN
    RAISE EXCEPTION '672 REFUSED: spend-governor doc_plans row already exists (replay).';
  END IF;
END $$;

-- 1. The hardened pre_query: identical chain, with the advisory lock woven into cfg (the
--    first CTE every later CTE references, so it always evaluates).
UPDATE scheduled_tasks SET pre_query = $PRE$
WITH cfg AS (
  SELECT c.* FROM governor_config c,
       (SELECT pg_advisory_xact_lock(hashtext('spend-governor-state'))) lock_taken
  WHERE c.id = 1),
spend AS (SELECT * FROM governor_spend_mtd),
new AS (
  SELECT CASE
           WHEN cfg.monthly_budget_usd IS NULL THEN 0
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l3_pct/100 THEN 3
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l2_pct/100 THEN 2
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l1_pct/100 THEN 1
           ELSE 0
         END AS lvl,
         spend.mtd_usd, spend.unpriced_io_tokens, spend.month
  FROM cfg, spend),
old AS (SELECT shed_level FROM governor_state WHERE id = 1),
noted AS (
  INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
  SELECT 'pipeline', 'spend-governor',
         format('spend-governor: shed level %s -> %s (month-to-date $%s of budget $%s; unpriced io tokens %s)',
                old.shed_level, new.lvl, new.mtd_usd,
                COALESCE((SELECT monthly_budget_usd::text FROM cfg), 'UNSET'), new.unpriced_io_tokens),
         '["spend-governor"]'::jsonb, 'scheduled_tasks:spend-governor-state'
  FROM old, new WHERE old.shed_level <> new.lvl
  RETURNING 1),
upd AS (
  UPDATE governor_state s
     SET shed_level = new.lvl, month = new.month, mtd_usd = new.mtd_usd,
         unpriced_io_tokens = new.unpriced_io_tokens, computed_at = now()
    FROM new WHERE s.id = 1
  RETURNING s.shed_level)
SELECT (SELECT shed_level FROM upd)  AS shed_level,
       (SELECT count(*) FROM noted)  AS level_changed
$PRE$
WHERE name = 'spend-governor-state';

-- 2. The travelling PLAN for the new pipeline.
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('pipeline', 'spend-governor',
'D4 LLM SPEND GOVERNOR — design of record.

WHAT: deliberate LLM spend shedding before the account hard cap, in the owner-ruled order
(2026-08-31): L1 sheds routine maintenance, L2 adds new site builds, L3 adds research.
LLM-free item types are never shed. Client work stays protected via D2 (separate lane).

PARTS: governor_model_prices (verified USD rates; cache writes at the 1h rate — err high) ·
governor_work_class_map (item_type -> class + llm_bearing; unmapped types default to
maintenance+llm_bearing at enforcement time) · governor_config (single row; enabled=false and
budget NULL keep everything inert) · governor_state (computed_at IS the heartbeat — stale
means the task is not running, never "all fine") · governor_spend_mtd (the single-sourced
meter; unpriced models are surfaced, never dropped) · scheduled task spend-governor-state
(120s, advisory-locked, doc_notes row on level CHANGE only).

STAGE B (not built): a claim-step check in Go refusing shed-class llm-bearing claims with
reason spend_governor_shed — opt-in, default off, its own council round. Option C (trigger
interval <=25s) stays gated on the governor being live and exercised once.

PRIOR ART, ruled out with evidence (council corr 80df0963 r2): platform/governance/fuel.go
is a per-TASK fuel header budget (abstract units, two call sites) — different scope, different
unit, different question; fleet-step-token-pressure and council-seat-token-pressure audit
max_tokens CAPS vs truncation — not spend.

FULLER DESIGN: dispatch_throughput lane PLAN §"D4 — LLM SPEND GOVERNOR"; migrations 671
(stage A) + 672 (hardening); rollback 671_..._ROLLBACK.sql (guarded, refuses while enabled).',
'migration:672', 'dispatch_throughput lane');

-- 3. Verify: md5 changed as intended; the lock is in the stored text; then the synthetic
--    level-change drive, self-cleaning.
DO $$
DECLARE q text; lvl int; changed int; before_notes int; after_notes int; synth_ids uuid[];
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='spend-governor-state';
  IF position('pg_advisory_xact_lock' in q) = 0 THEN
    RAISE EXCEPTION '672 VERIFY: advisory lock absent from stored pre_query';
  END IF;

  SELECT count(*) INTO before_notes FROM doc_notes WHERE subject_key='spend-governor';

  -- Drive level 0 -> 3: any positive budget is below August's ~$2k spend.
  UPDATE governor_config SET monthly_budget_usd = 0.01 WHERE id = 1;
  EXECUTE q;
  SELECT shed_level INTO lvl FROM governor_state WHERE id=1;
  IF lvl <> 3 THEN RAISE EXCEPTION '672 VERIFY: synthetic budget 0.01 gave level %, expected 3', lvl; END IF;

  -- Drive 3 -> 0 by restoring the shipped-inert budget.
  UPDATE governor_config SET monthly_budget_usd = NULL WHERE id = 1;
  EXECUTE q;
  SELECT shed_level INTO lvl FROM governor_state WHERE id=1;
  IF lvl <> 0 THEN RAISE EXCEPTION '672 VERIFY: budget back to NULL gave level %, expected 0', lvl; END IF;

  SELECT count(*) INTO after_notes FROM doc_notes WHERE subject_key='spend-governor';
  IF after_notes - before_notes <> 2 THEN
    RAISE EXCEPTION '672 VERIFY: expected exactly 2 synthetic level-change notes, got %', after_notes - before_notes;
  END IF;

  -- The probe cleans up after itself: delete exactly the two synthetic rows just written.
  SELECT array_agg(id) INTO synth_ids FROM (
    SELECT id FROM doc_notes WHERE subject_key='spend-governor'
    ORDER BY created_at DESC LIMIT 2) t;
  DELETE FROM doc_notes WHERE id = ANY(synth_ids);

  PERFORM 1 FROM doc_plans WHERE subject_type='pipeline' AND subject_key='spend-governor' AND is_current;
  IF NOT FOUND THEN RAISE EXCEPTION '672 VERIFY: doc_plans row missing'; END IF;

  RAISE NOTICE '672 OK: lock woven in, level-change INSERT arm proven both directions (0->3->0, 2 notes fired and cleaned), travelling PLAN in place, config back to shipped-inert.';
END $$;

COMMIT;
