-- 610 — DROP content_components.usage_count
--
-- ⚠⚠ _HOLD: DO NOT LET THE RUNNER TAKE THIS. Apply BY HAND, and only after the
--          precondition below is verified at the ARTEFACT. It is named _HOLD because a
--          banner cannot hold a file back — the runner's --apply takes EVERY pending
--          file in the directory, and a migration's own guard checks DRIFT, not ORDER.
--
-- PRECONDITION, and it is the whole reason this file is held:
--   store_generated_component_action.go's birth INSERT named `usage_count` in its column
--   list until commit <this commit>. Against a binary built BEFORE that commit, dropping
--   this column makes every component creation fail with
--   `column "usage_count" of relation "content_components" does not exist`.
--
--   So: DO NOT APPLY until a chassis build containing that commit is LIVE, proven at the
--   artefact, not at git:
--
--     kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user \
--       -d clients_db -c "SELECT service, git_commit FROM service_binary_capabilities \
--       WHERE service LIKE '%chassis%' ORDER BY last_seen_at DESC LIMIT 1;"
--     git merge-base --is-ancestor <this commit> <that git_commit>   # must be YES
--     git merge-base --is-ancestor <a LATER commit> <that git_commit> # must be NO (control)
--
--   The control matters. On 2026-08-24 this lane's first ancestry control was a commit
--   that ALSO predated the build, so both arms returned YES and the check proved nothing.
--
--   ⚠ THE ANCESTRY CHECK IS NOT ENOUGH ON ITS OWN, and the council's debug_historian seat was
--   right to say so (round ac7b62e6, medium): a roll is not evidence, and edit 1 REMOVES text
--   rather than adding any, so there is no new marker to grep for. An absence-grep is this
--   estate's documented false-negative trap — a silently failing grep looks exactly like a
--   successful removal.
--
--   BUT A DISCRIMINATING PROBE DOES EXIST, because the removed text was itself inside a
--   string literal (the SQL), and it has been PROVEN to discriminate while the pre-fix
--   binary was still running — which is the only time that proof can ever be obtained:
--
--     POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --             -o jsonpath='{.items[0].metadata.name}')
--     # MUST BE ABSENT once edit 1 is live (it is the old INSERT column list):
--     kubectl -n ai-persona-system exec $POD -- grep -aq 'usage_count, avg_quality_score' /proc/1/exe
--     # CONTROL, MUST BE PRESENT — proves the grep is not silently failing:
--     kubectl -n ai-persona-system exec $POD -- grep -aq 'INSERT INTO content_components' /proc/1/exe
--
--   [VERIFIED 2026-08-25, against pod agent-chassis-67fd9c76f5-2g8kw running the PRE-fix
--   binary] the first probe returned PRESENT and the control returned PRESENT. So the
--   absence of that literal is a real signal and not merely an untested expectation.
--   ⚠ Do NOT substitute a positive probe for the new code: the obvious candidate
--   ('is_active,' followed by 'avg_quality_score') was tested the same day and matched the
--   PRE-fix binary too, so it does not discriminate. Absence-of-the-old + control is the check.
--
-- LEDGER (council editquality, low): a _HOLD file is applied by hand, so the runner will not
--   record it. Record it the same way 609 was, immediately after applying:
--     ./scripts/migration/run-migrations.sh --record-only \
--        610_content_components_drop_dead_usage_count_HOLD.sql \
--        --note "applied by hand <date>; edit-1 probe: old INSERT literal ABSENT, control PRESENT; guard passed with N rows/max M"
--   Without that, a later state audit of schema_migrations shows 610 as never applied.
--
-- ⚠ SECOND PRECONDITION, ADDED 2026-08-25 — RE-GREP THE INSERT, DO NOT TRUST THIS FILE'S
--   DATE. The `bugs_open/388` lane is actively editing store_generated_component_action.go
--   (its fix moves component IDENTITY resolution off the LLM's emitted `function` field).
--   That is the very file whose INSERT column list this migration depends on. The runtime
--   guard below reads DATA and cannot see CODE, so it cannot catch a re-introduced column
--   reference. Before applying, run:
--
--     grep -n 'usage_count' platform/orchestration/actions/store_generated_component_action.go
--
--   Anything other than the explanatory comment near the creation path — in particular any
--   occurrence inside an INSERT column list or VALUES clause — means a writer is back and
--   this migration must NOT run until it is removed and that removal is live.
--   (`usage_count` carries DEFAULT 0, so omitting it from the INSERT is behaviour-identical;
--   there is never a reason to name it.)
--
-- WHY DROP RATHER THAN LEAVE IT COMMENTED. 609 makes the column honest to a reader of
-- `\d+`. It does not make the bad state unrepresentable: the values are still there, still
-- plausible-looking integers (20, 19, 12, 7, …), and still one `SELECT` away from being
-- quoted as evidence in a bug file. This estate's own rule is to rank fixes by what makes
-- the bad state impossible rather than merely discouraged. Nothing reads the column
-- (measured: no Go reader, no DB function, view or trigger, no frontend), and it carries
-- NO INDEX, so the drop costs nothing and removes the trap permanently.
--
-- WHAT IS LOST: nothing of value. The values are a frozen snapshot of resolution attempts
-- on one of three paths — they miss ~1,900 real bindings and count non-usages (the two
-- largest belong to components with ZERO bindings). The ROLLBACK restores the column but
-- NOT the numbers, and that is deliberate: restoring wrong numbers would be worse than
-- restoring none. Full evidence: bugs_closed/378.

\set ON_ERROR_STOP on
BEGIN;

-- GUARD 1 — idempotency. Already dropped is a success, not a failure.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'content_components' AND column_name = 'usage_count'
  ) THEN
    RAISE NOTICE 'content_components.usage_count already dropped — nothing to do.';
  END IF;
END $$;

-- GUARD 2 — THE LIVE-WRITER CHECK, and it can actually fire.
-- If the column has moved since the counter was killed, some binary is still writing it,
-- which means the precondition above is NOT met and dropping would break creations.
-- The frozen snapshot as of 2026-08-24/25 is: exactly 12 section-level rows non-zero,
-- maximum value 20. Either number moving means a writer is live.
DO $$
DECLARE
  n_nonzero int;
  max_val   int;
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name='content_components' AND column_name='usage_count') THEN

    SELECT count(*), COALESCE(max(usage_count), 0) INTO n_nonzero, max_val
      FROM content_components
     WHERE is_active AND forked_from IS NULL
       AND component_level = 'section' AND COALESCE(usage_count, 0) > 0;

    IF n_nonzero > 12 OR max_val > 20 THEN
      RAISE EXCEPTION
        'ABORT: usage_count has MOVED since it was killed (% non-zero rows, max %; expected <=12 and <=20). A binary is still incrementing it, so the old birth INSERT is probably live too and this DROP would break component creation. Verify the chassis build BEFORE retrying.',
        n_nonzero, max_val;
    END IF;

    RAISE NOTICE 'GUARD: PASS — % non-zero rows, max % (frozen as expected)', n_nonzero, max_val;
  END IF;
END $$;

ALTER TABLE content_components DROP COLUMN IF EXISTS usage_count;

-- VERIFY — DO/RAISE, because a verify block of SELECTs cannot stop the COMMIT.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'content_components' AND column_name = 'usage_count'
  ) THEN
    RAISE EXCEPTION 'VERIFY FAILED: content_components.usage_count still exists after the DROP';
  END IF;
  RAISE NOTICE 'VERIFY: PASS — content_components.usage_count is gone';
END $$;

COMMIT;
