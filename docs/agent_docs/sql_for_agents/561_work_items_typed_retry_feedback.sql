-- 561_work_items_typed_retry_feedback.sql
--
-- bugs_open/345 — give retry feedback its OWN channel, because the one it is
-- using is a free-text column that twenty writers and several humans write to.
--
-- WHAT IS WRONG TODAY. 345's candidate 1 shipped (Go loader + migration 533's
-- prompt block + migration 555's dispatcher mapping) and it WORKS — item
-- ceea0c07 was refused at 12:18:43Z on 2026-08-22, re-dispatched at 12:51
-- carrying `last_error`, and completed at 12:53:07. But the value it carries is
-- `site_work_items.error`, and the prompt asserts a provenance that column
-- cannot support: "Your previous output for this component was refused by
-- validation and was NOT stored."
--
-- [MEASURED 2026-08-22, live clients_db] That claim is false much of the time:
--   * TWENTY write sites across TEN files write `site_work_items.error`,
--     including three in internal/core-manager/admin/site_admin_handlers.go —
--     the human operator HTTP path — plus hand-run SQL by lanes.
--   * Of 799 items fleet-wide passing the loader's gate (non-blank error,
--     completed_at IS NULL): 11 (1.4%) are validation rejections, 383 are other
--     steps' failures, and 405 are NEITHER — they are human/lane administrative
--     notes. Real stored values: 'HELD 2026-08-18 by the loanzy_uk_example_site
--     lane: …' (33 rows), '[cancelled 2026-08-20 by the 215 same-name canary
--     (corr 313368d2): test run, not a build request…' (32),
--     'Claim timed out (attempts exhausted)' (16).
--   * Narrowed to the ONLY population that reaches a reader today
--     (item_type='needs_new_component'): 17 items, of which 6 (35%) are NOT
--     validation rejections — 3 token-cap truncations (bugs_open/337's
--     population, where the correct remedy is "be shorter", not "change what it
--     says was wrong") and 3 human administrative notes.
--
-- WHY A COLUMN AND NOT A CLASSIFIER. Classifying at the loader would mean
-- matching error TEXT — a lexical comparison over a column written freely by
-- twenty code paths and by people. That is the defect class this estate's
-- council gate has caught in three consecutive rounds, and it would have to be
-- re-got-right for every future producer. A dedicated column with exactly ONE
-- writer makes the question "is this actionable feedback about my own previous
-- output?" unanswerable-wrongly: there is nothing to classify. `error` keeps
-- its twenty writers and its meaning, untouched.
--
-- ADDITIVE AND INERT. Nothing reads this column until the Go loader does, and
-- nothing writes it until store_generated_component_action.go does. Either half
-- may land first: with the column absent the loader reads nothing; with the Go
-- half absent the column is simply never populated. No ordering constraint is
-- claimed here (owner ruling 2026-07-29 retired the condition that would have
-- let one be claimed).
--
-- SHAPE: {"code": "...", "message": "...", "at": "...",
--         "orchestration_id": "...", "step": "..."}
--   code   — the producer's OWN classification, not a guess from the text
--   message— the actionable report, verbatim, capped by the READER not here
--   at     — so a stale record is visible as stale
--
-- site_work_items is 9,919 rows [MEASURED 2026-08-22], and ADD COLUMN with no
-- DEFAULT is metadata-only on this Postgres, so the lock is momentary.
--
-- ROLLBACK: 561_..._ROLLBACK.sql drops the column.

BEGIN;

DO $guard$
DECLARE
  n int;
BEGIN
  -- Pre-state: the column must not already exist. A second session adding a
  -- column of the same name with a DIFFERENT shape is the failure this catches
  -- — and it must ABORT, not silently adopt whatever is there.
  SELECT count(*) INTO n
    FROM information_schema.columns
   WHERE table_schema='public' AND table_name='site_work_items'
     AND column_name='retry_feedback';

  IF n <> 0 THEN
    RAISE EXCEPTION '561: site_work_items.retry_feedback already exists — another session has added it. Read its shape before assuming this migration is a no-op.';
  END IF;

  -- The table itself must be the one we think it is.
  SELECT count(*) INTO n
    FROM information_schema.columns
   WHERE table_schema='public' AND table_name='site_work_items'
     AND column_name IN ('error','completed_at','attempt_count','max_attempts','item_type');

  IF n <> 5 THEN
    RAISE EXCEPTION '561: site_work_items is not the expected shape (found % of the 5 anchor columns) — re-read the schema before editing', n;
  END IF;
END
$guard$;

ALTER TABLE site_work_items ADD COLUMN retry_feedback jsonb;

COMMENT ON COLUMN site_work_items.retry_feedback IS
  'bugs_open/345: TYPED retry feedback for the handler that produced the refused artefact. '
  'Exactly one writer (store_generated_component_action.go recordValidationRejection). '
  'NOT a general error field — that is site_work_items.error, which has ~20 writers including '
  'human operators, and which must never be piped into a prompt as if it described the '
  'writer''s own output. Shape: {code, message, at, orchestration_id, step}.';

-- Verify. A DO/RAISE block, never a bare SELECT: ON_ERROR_STOP ignores a
-- non-empty result set, so a verify block of SELECTs cannot stop the COMMIT.
DO $verify$
DECLARE
  t text;
BEGIN
  SELECT data_type INTO t
    FROM information_schema.columns
   WHERE table_schema='public' AND table_name='site_work_items'
     AND column_name='retry_feedback';

  IF t IS NULL THEN
    RAISE EXCEPTION '561 VERIFY: retry_feedback was not created';
  END IF;
  IF t <> 'jsonb' THEN
    RAISE EXCEPTION '561 VERIFY: retry_feedback is %, expected jsonb', t;
  END IF;

  -- It must be NULL everywhere: this migration adds a channel, it does not
  -- backfill one. Backfilling from `error` would import exactly the 405
  -- human notes this whole change exists to keep out.
  PERFORM 1 FROM site_work_items WHERE retry_feedback IS NOT NULL LIMIT 1;
  IF FOUND THEN
    RAISE EXCEPTION '561 VERIFY: retry_feedback is already populated — this migration does not backfill';
  END IF;

  RAISE NOTICE '561 OK: site_work_items.retry_feedback created (jsonb, NULL on all % rows)',
    (SELECT count(*) FROM site_work_items);
END
$verify$;

COMMIT;
