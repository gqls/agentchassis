-- 581_refuse_selector_invisible_section_birth.sql
--
-- WHAT: refuse, at BIRTH, a section-level library component that carries no
-- `section_type` — the state that makes a component invisible to the component
-- selector while looking perfectly healthy in every listing.
--
-- WHY IT IS A TRIGGER AND NOT GO: every such row ever born is
-- `created_from='manual'` (28 of 28, all-history, as of 2026-08-23); the
-- `generated` route has produced ZERO. The producer is hand-run SQL, which no
-- Go guard can reach. Only the table can.
--
-- WHY NOW (bugs_open/351): the selector queries `WHERE section_type = $1`, so a
-- NULL row is unreachable through it — silently, with no error and no work
-- item. The measured cost of that silence is TEN diverted twins as of
-- 2026-08-23 (site-suffixed near-duplicates of components the library already
-- owned, born 08-18..08-22), each one a paid LLM generation, plus the exposure
-- to bugs_open/345's retry loop. 351's predicate half is fixed, live and
-- demand-proven; this closes the door the data half came through.
--
-- THIS IS A PRECAUTION, NOT A REPAIR, AND THE DISTINCTION IS DELIBERATE. The
-- standing NULL rows are NOT touched and are NOT backfilled — that decision is
-- recorded in bugs_open/351 ("incumbents stay Path-1-only", 2026-08-23) and
-- rests on two measured harms of a backfill: load_existing_component's PRIMARY
-- query is keyed on section_type and its NULL miss deliberately routes to the
-- store's identity resolver (load_existing_component_action.go:163,182), and
-- the selector's `ORDER BY score DESC` has no secondary key, so a backfilled
-- incumbent can tie exactly with its twin and resolve nondeterministically.
-- The live harm of the standing rows is ~0: all 25 are bound and in service via
-- Paths 0/1, and two self-heals already exist in store_generated_component.
--
-- ── SCOPING: THREE TRAPS, EACH VERIFIED IN CODE BEFORE THIS WAS WRITTEN ──────
--
-- 1. `forked_from IS NULL` IS LOAD-BEARING, NOT DECORATION.
--    deploy_tool_action.go:326's fork INSERT copies `component_level` from its
--    source but does NOT list `section_type` among its 16 columns. A
--    section-level fork is therefore legitimately born NULL with `forked_from`
--    set. An unscoped trigger would break tool deployment at runtime.
--
-- 2. INSERT-ONLY, NEVER UPDATE — which is also why this is not a CHECK.
--    A CHECK constraint (even NOT VALID) is enforced on UPDATE of pre-existing
--    rows, so every template-repair write to the standing 25
--    (fix_hardcoded_colours, fix_forced_text_colours, the rerender writers…)
--    would begin failing. A BEFORE INSERT trigger leaves them freely
--    repairable, and leaves a deliberate post-birth
--    `UPDATE … SET section_type = NULL` representable as an explicit opt-out.
--
-- 3. NOT A GENERATED COLUMN AND NOT A `COALESCE(section_type, function)`
--    DEFAULT. 35 active rows deliberately carry a section_type that DIFFERS
--    from function (the split is 89 equal / 35 differ / 25 NULL of 149, as of
--    2026-08-23), so a generated column would destroy real vocabulary. And
--    `function` DEFAULTs to 'generic-text-block', so a silent COALESCE would
--    quietly pour unlabelled rows into the commonest selector pool — the one
--    place a wrong match is least likely to be noticed.
--
-- WHAT IT DOES NOT DO: it does not touch existing rows, does not run on UPDATE,
-- does not fire for tools or forks, and invents no vocabulary. It refuses, and
-- says what to supply.
--
-- COMPLETENESS NOTE: `section_type = ''` is not NULL and so is not caught here
-- — it is already refused by the existing `chk_section_type_kebab_case`, whose
-- regex requires at least one character. NULL (this trigger) and '' (that
-- CHECK) together leave only a valid kebab value admissible.
--
-- ── REVISED 2026-08-24 after council `f0cd2420` (APPROVED, 3 advisory objections) ──
--
-- Two medium advisories were acted on rather than banked, and both changed the SQL:
--
-- A. `bug_historian`: BEFORE INSERT alone leaves a MUTATION PATH open — nothing
--    stopped a later UPDATE setting section_type back to NULL on an
--    already-valid row, reproducing the identical invisibility through a
--    different door. It asked whether that was live or theoretical and noted it
--    could not be answered from the schema. ANSWERED: it is theoretical today.
--    No Go writer and no migration sets section_type to NULL — the only UPDATE
--    of the column anywhere is store_generated_component_action.go:1365,
--    `SET section_type = COALESCE(section_type, $1)`, which is a HEAL and can
--    never null it (grep over platform/ internal/ pkg/ cmd/ and
--    sql_for_agents/, 2026-08-24).
--    Closed anyway, because it costs one branch and the door is the point:
--    the trigger now also fires BEFORE UPDATE, and refuses ONLY the
--    NOT NULL -> NULL transition. It deliberately still allows an UPDATE of a
--    row whose section_type was ALREADY NULL, which is what preserves the
--    original reason for excluding UPDATE — the 25 standing rows stay freely
--    repairable by every template-repair writer.
--
-- B. `debug_historian`: CREATE TRIGGER is not idempotent (Postgres has no
--    CREATE TRIGGER IF NOT EXISTS), so a replay or partial-apply retry died on
--    "trigger already exists". The pre-check made it worse, not better: it
--    asserted ZERO triggers, so a re-run failed at MY OWN guard. Now the guard
--    ignores this migration's own trigger and a DROP ... IF EXISTS precedes the
--    CREATE, so re-applying is a no-op.
--
-- A third advisory (`editquality`, low) noted that adding a row to
-- 000_concept_index.md trips a documented landmine. Checked with that
-- landmine's own PAIR check rather than the row count it warns about: entries
-- with no index row = empty, rows with no entry = empty, CLC-029 present in
-- both.
--
-- ROLLBACK: 581_refuse_selector_invisible_section_birth_ROLLBACK.sql

BEGIN;

-- Guard 1: this table had NO non-internal triggers as of 2026-08-23, so there is
-- nothing to interact with. If that has changed, stop and look before adding one.
--
-- ⚠ It excludes THIS migration's own trigger, so a replay is a no-op rather than
-- a failure at the guard. That was the shape of the first version and it is the
-- opposite of idempotent: it made re-running fail EARLIER and more confusingly.
DO $$
DECLARE n int; names text;
BEGIN
  SELECT count(*), string_agg(t.tgname, ', ') INTO n, names
    FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid
   WHERE c.relname = 'content_components' AND NOT t.tgisinternal
     AND t.tgname <> 'trg_cc_refuse_null_section_type';
  IF n <> 0 THEN
    RAISE EXCEPTION '581: content_components has % other non-internal trigger(s) (%); expected 0. Read them before adding another.', n, names;
  END IF;
END $$;

CREATE OR REPLACE FUNCTION refuse_selector_invisible_section()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  -- Not the shape we guard: a tool, a fork, or a row that declares its type.
  IF NEW.component_level <> 'section'
     OR NEW.forked_from IS NOT NULL
     OR NEW.section_type IS NOT NULL
  THEN
    RETURN NEW;
  END IF;

  -- Reaching here, NEW is a section-level, non-forked row with section_type NULL.
  --
  -- ⚠ THE ONE ALLOWED CASE, and it is the whole reason this is not a CHECK:
  -- an UPDATE of a row that was ALREADY NULL. That is a template repair to one
  -- of the standing rows (fix_hardcoded_colours, fix_forced_text_colours, the
  -- rerender writers …), none of which touch section_type. Refusing it would
  -- freeze exactly the rows most in need of maintenance.
  IF TG_OP = 'UPDATE' AND OLD.section_type IS NULL THEN
    RETURN NEW;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    RAISE EXCEPTION
      'content_components: refusing to CLEAR section_type on a section-level library component (function=%, was %). The component selector queries "WHERE section_type = $1", so nulling this makes the row silently unreachable while it stays healthy-looking in every listing — the same defect as birthing it without one. If a component is being retired, set is_active = false instead. See bugs_open/351 and migration 581.',
      COALESCE(NEW.function, '(null)'), quote_literal(OLD.section_type)
      USING ERRCODE = '23514';
  ELSE
    RAISE EXCEPTION
      'content_components: a section-level library component must declare section_type (got NULL for function=%). Without it the component selector cannot see this row: it queries "WHERE section_type = $1", so the row would be silently unreachable while looking healthy in every listing. Supply section_type (kebab-case; usually the same string a page plan would ask for). Forks and component_level=''tool'' rows are exempt. See bugs_open/351 and migration 581.',
      COALESCE(NEW.function, '(null)')
      USING ERRCODE = '23514';
  END IF;
END $$;

COMMENT ON FUNCTION refuse_selector_invisible_section() IS
  'bugs_open/351 / migration 581: refuses a section-level, non-forked component that would be INVISIBLE to the component selector — born without section_type, or updated to clear it. Deliberately ALLOWS an UPDATE of a row that was already NULL, so the 25 pre-existing such rows stay repairable; that carve-out is why this is a trigger and not a CHECK.';

-- Postgres has no CREATE TRIGGER IF NOT EXISTS, so the DROP is what makes a
-- replay or partial-apply retry a no-op instead of an error.
DROP TRIGGER IF EXISTS trg_cc_refuse_null_section_type ON content_components;
DROP TRIGGER IF EXISTS trg_cc_refuse_null_section_type_birth ON content_components;  -- the pre-revision name
DROP FUNCTION IF EXISTS refuse_selector_invisible_section_birth();                    -- ditto

CREATE TRIGGER trg_cc_refuse_null_section_type
  BEFORE INSERT OR UPDATE ON content_components
  FOR EACH ROW
  EXECUTE FUNCTION refuse_selector_invisible_section();

-- ── VERIFY: INDUCE the refusal, then prove it does NOT fire where it must not ──
--
-- A verify block of bare SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a
-- non-empty result), so this is DO/RAISE and it INDUCES the violation rather
-- than asserting the trigger's existence. Every probe row is written inside a
-- subtransaction that is deliberately aborted, so nothing survives this block.
DO $$
DECLARE
  refused    boolean := false;
  cleared    boolean := false;
  fork_src   uuid;
  standing   uuid;
  tag        text := 'zz-581-probe-' || replace(gen_random_uuid()::text, '-', '');
BEGIN
  SELECT id INTO fork_src FROM content_components WHERE component_level = 'section' LIMIT 1;
  IF fork_src IS NULL THEN
    RAISE EXCEPTION '581 VERIFY: no section-level row exists to fork from — the controls cannot run, so this migration is unverified.';
  END IF;

  SELECT id INTO standing FROM content_components
   WHERE component_level = 'section' AND forked_from IS NULL AND section_type IS NULL LIMIT 1;

  BEGIN
    -- (1) INDUCE — the case this migration exists for. Must be refused.
    BEGIN
      INSERT INTO content_components (name, function, html_template, component_level, section_type, forked_from)
      VALUES (tag || '-a', 'zz-probe-null', '<section>x</section>', 'section', NULL, NULL);
    EXCEPTION WHEN check_violation THEN
      refused := true;
    END;
    IF NOT refused THEN
      RAISE EXCEPTION '581 VERIFY: a section-level row with NULL section_type was ACCEPTED — the trigger is inert.';
    END IF;

    -- (2) CONTROL — the same row WITH a section_type must be accepted.
    --     Disconfirming result: refusal, i.e. the predicate is too wide.
    INSERT INTO content_components (name, function, html_template, component_level, section_type, forked_from)
    VALUES (tag || '-b', 'zz-probe-ok', '<section>x</section>', 'section', 'zz-probe-ok', NULL);

    -- (3) CONTROL — a FORK is legitimately born NULL (deploy_tool_action.go:326).
    --     Disconfirming result: refusal, i.e. this migration just broke tool deployment.
    INSERT INTO content_components (name, function, html_template, component_level, section_type, forked_from)
    VALUES (tag || '-c', 'zz-probe-fork', '<section>x</section>', 'section', NULL, fork_src);

    -- (4) CONTROL — a TOOL never carries section_type.
    --     Disconfirming result: refusal, i.e. create_tool_component is broken.
    INSERT INTO content_components (name, function, html_template, component_level, section_type, forked_from)
    VALUES (tag || '-d', 'zz-probe-tool', '<div>x</div>', 'tool', NULL, NULL);

    -- (5) CONTROL — an UPDATE of a STANDING NULL row must still succeed, or the
    --     repair writers are broken. This is the NOT-VALID-CHECK trap, induced.
    IF standing IS NOT NULL THEN
      UPDATE content_components SET updated_at = updated_at WHERE id = standing;
    ELSE
      RAISE NOTICE '581 VERIFY: no standing NULL row to UPDATE-probe (the set may have healed) — control (5) skipped.';
    END IF;

    -- (6) INDUCE the UPDATE arm — clearing a SET section_type must be refused.
    --     This is the mutation path council f0cd2420's bug_historian seat named:
    --     BEFORE INSERT alone left it open. Row -b (control 2) has a section_type,
    --     so nulling it is exactly the transition the arm exists to refuse.
    --     Disconfirming result: the UPDATE succeeds, i.e. the arm is inert and the
    --     door is still open through a different write path.
    cleared := false;
    BEGIN
      UPDATE content_components SET section_type = NULL WHERE name = tag || '-b';
    EXCEPTION WHEN check_violation THEN
      cleared := true;
    END;
    IF NOT cleared THEN
      RAISE EXCEPTION '581 VERIFY: clearing section_type on a section-level row was ACCEPTED — the UPDATE arm is inert.';
    END IF;

    -- Undo controls (2)(3)(4). This always fires; it is the rollback, not a failure.
    RAISE EXCEPTION 'ZZ581_PROBE_ROLLBACK';
  EXCEPTION
    WHEN raise_exception THEN
      IF SQLERRM <> 'ZZ581_PROBE_ROLLBACK' THEN
        RAISE;   -- a real verification failure — let it abort the migration
      END IF;
  END;

  RAISE NOTICE '581 VERIFY: PASS — BOTH refusals induced (birth, and clearing), and 4 controls (labelled, fork, tool, standing-row UPDATE) all behaved as required.';
END $$;

-- Record what the standing population was when the door closed, so a later
-- reader can tell healing from drift. NOT an assertion: another session may
-- heal a row between now and then, and that must not fail a migration.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
   WHERE is_active AND component_level = 'section' AND forked_from IS NULL AND section_type IS NULL;
  RAISE NOTICE '581: % active section-level non-forked rows still carry a NULL section_type (was 25 as of 2026-08-23). They are deliberately NOT backfilled — see bugs_open/351.', n;
END $$;

COMMIT;
