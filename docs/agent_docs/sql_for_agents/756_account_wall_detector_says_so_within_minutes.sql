-- 756_account_wall_detector_says_so_within_minutes.sql — the thing the spend governor
-- structurally cannot see, announced within two minutes instead of found by a reviewer forty
-- minutes later.
--
-- WHY. 2026-09-04 11:21:12Z–11:56:48Z: the Anthropic account's prepaid credit ran out; every
-- LLM call fleet-wide failed with HTTP 400 "Your credit balance is too low to access the
-- Anthropic API" (91 of 107 council reviewer calls, plus landmine-verifier, tool-improver,
-- feed-triage, build-briefing, diagnose-agent). The spend governor stood at level 0, $647 of a
-- $2,000 budget — it meters SPEND AGAINST THE BUDGET and has no view of the account's
-- REMAINING BALANCE, so the fifth account-wall blackout arrived with the governor green. It
-- was found 40 minutes later by a peer lane reading a failed council run, and misattributed
-- to the governor. The owner's ruling is "loudly"; this makes the account wall loud too.
--
-- WHAT. A 120 s scheduled task in the governor's own shape (fire_message=false, one statement,
-- advisory-locked, writes doc_notes ONLY on a state change):
--   * WALL:    failed calls carrying the credit-balance string in the last 5 minutes, and no
--              'account-wall' note in the last 30 minutes → one note (subject_key
--              spend-governor, categories [spend-governor, account-wall]).
--   * CLEARED: no such failure in the last 5 minutes, at least one success in the last 5
--              minutes, and the newest account-wall note has no 'account-wall-cleared' note
--              after it → one cleared note.
-- The session-start banner (scripts/governor-session-start.py) reads the newest account-wall
-- note and shouts while it is uncleared. The state-change alarm (753) is untouched.
--
-- READS llm_call_log; WRITES doc_notes only. Never touches llm_call_log (it is the training
-- corpus). VERIFY: drives both branches against SYNTHETIC rows inside this transaction
-- (a fake failed call, then a fake success), asserts the notes, deletes exactly the synthetic
-- rows and notes by id, and asserts llm_call_log's count is back where it started.
-- Rollback: 756_..._ROLLBACK.sql (deletes the task; leaves any notes it wrote — they are history).

BEGIN;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='account-wall-detector') THEN
    RAISE EXCEPTION '756 REFUSED: account-wall-detector already exists (replay).';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='llm_call_log' AND column_name='error_message') THEN
    RAISE EXCEPTION '756 REFUSED: llm_call_log.error_message missing — the signature column this reads.';
  END IF;
END $$;

INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, target_topic, pre_query, enabled, timeout_seconds, fire_message)
VALUES (
  'account-wall-detector',
  'Announces (doc_notes spend-governor / account-wall) when the Anthropic API refuses calls for credit balance, and when it clears. The spend governor cannot see the account balance; this can see its symptom. 756.',
  120, 'generic', 'system.agent.scheduled.requests',
  $PRE$
WITH lock AS (SELECT pg_advisory_xact_lock(hashtext('account-wall-detector'))),
recent AS (
  SELECT count(*) FILTER (WHERE NOT success AND error_message ILIKE '%credit balance is too low%') AS failing,
         count(*) FILTER (WHERE success) AS ok,
         min(created_at) FILTER (WHERE NOT success AND error_message ILIKE '%credit balance is too low%') AS first_fail,
         max(created_at) FILTER (WHERE NOT success AND error_message ILIKE '%credit balance is too low%') AS last_fail
  FROM llm_call_log, lock WHERE created_at > now() - interval '5 minutes'),
last_wall AS (SELECT id, created_at FROM doc_notes WHERE subject_key='spend-governor' AND categories ? 'account-wall' ORDER BY created_at DESC LIMIT 1),
last_clear AS (SELECT created_at FROM doc_notes WHERE subject_key='spend-governor' AND categories ? 'account-wall-cleared' ORDER BY created_at DESC LIMIT 1),
wall AS (
  INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
  SELECT 'pipeline', 'spend-governor',
         format('ACCOUNT WALL: the Anthropic API is refusing calls — "credit balance is too low" — %s failed call(s) in the last 5 minutes (first %s, last %s). The spend governor CANNOT see this: it meters spend against the budget, not the account balance. OWNER: top up the prepaid balance (Plans & Billing). Sessions: every LLM-bearing step is failing until it clears; council submissions will end at complete_invalid — do not re-trigger until an account-wall-cleared note follows this one.',
                r.failing, r.first_fail, r.last_fail),
         '["spend-governor","account-wall"]'::jsonb, 'scheduled_tasks:account-wall-detector'
  FROM recent r
  WHERE r.failing > 0
    AND NOT EXISTS (SELECT 1 FROM doc_notes WHERE subject_key='spend-governor' AND categories ? 'account-wall' AND created_at > now() - interval '30 minutes')
  RETURNING 1),
cleared AS (
  INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
  SELECT 'pipeline', 'spend-governor',
         format('ACCOUNT WALL CLEARED: no credit-balance refusals in the last 5 minutes and %s successful call(s). Council submissions that ended at complete_invalid during the wall can be re-triggered now.', r.ok),
         '["spend-governor","account-wall-cleared"]'::jsonb, 'scheduled_tasks:account-wall-detector'
  FROM recent r, last_wall w
  WHERE r.failing = 0 AND r.ok > 0
    AND COALESCE((SELECT created_at FROM last_clear) < w.created_at, true)   -- no cleared note yet, or the last one predates the last wall
  RETURNING 1)
SELECT (SELECT failing FROM recent) AS failing_5m, (SELECT count(*) FROM wall) AS wall_noted, (SELECT count(*) FROM cleared) AS cleared_noted
$PRE$,
  true, 60, false);

-- ---------------------------------------------------------------- verify: both branches, synthetic rows, all cleaned
DO $$
DECLARE q text; f int; w int; c int; n0 bigint; fake_id uuid; fake2_id uuid; wall_id uuid; clear_id uuid; notes0 int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='account-wall-detector';
  EXECUTE 'PREPARE p756 AS ' || q; EXECUTE 'DEALLOCATE p756';
  SELECT count(*) INTO n0 FROM llm_call_log;
  SELECT count(*) INTO notes0 FROM doc_notes WHERE subject_key='spend-governor';
  IF EXISTS (SELECT 1 FROM doc_notes WHERE subject_key='spend-governor' AND categories ? 'account-wall' AND created_at > now() - interval '30 minutes') THEN
    RAISE EXCEPTION '756 VERIFY: an account-wall note from the last 30 minutes already exists — the wall branch cannot be driven cleanly now; retry later';
  END IF;

  -- (a) no failure, no note either way (the quiet state)
  EXECUTE q INTO f, w, c;
  IF w <> 0 OR c <> 0 THEN RAISE EXCEPTION '756 VERIFY: quiet state wrote notes (wall=% cleared=%)', w, c; END IF;

  -- (b) a synthetic credit-balance failure → exactly one WALL note
  INSERT INTO llm_call_log (agent_type, model, success, error_message, created_at)
  VALUES ('__756_verify__', 'claude-sonnet-5', false, 'API request failed with status 400: {"type":"error","error":{"type":"invalid_request_error","message":"Your credit balance is too low to access the Anthropic API."}}', now())
  RETURNING id INTO fake_id;
  EXECUTE q INTO f, w, c;
  IF f < 1 OR w <> 1 THEN RAISE EXCEPTION '756 VERIFY: synthetic failure should write one wall note; failing=% wall=%', f, w; END IF;
  SELECT id INTO wall_id FROM doc_notes WHERE subject_key='spend-governor' AND categories ? 'account-wall' ORDER BY created_at DESC LIMIT 1;
  IF wall_id IS NULL OR position('ACCOUNT WALL' in (SELECT body FROM doc_notes WHERE id=wall_id)) = 0 THEN RAISE EXCEPTION '756 VERIFY: wall note missing or malformed'; END IF;
  -- (c) a second tick while still failing → NO second note (30-minute dedupe)
  EXECUTE q INTO f, w, c;
  IF w <> 0 THEN RAISE EXCEPTION '756 VERIFY: second tick during the wall wrote another note'; END IF;
  -- (d) failure gone, a success present → exactly one CLEARED note
  DELETE FROM llm_call_log WHERE id = fake_id;
  INSERT INTO llm_call_log (agent_type, model, success, created_at) VALUES ('__756_verify__', 'claude-sonnet-5', true, now()) RETURNING id INTO fake2_id;
  EXECUTE q INTO f, w, c;
  IF f <> 0 OR c <> 1 THEN RAISE EXCEPTION '756 VERIFY: recovery should write one cleared note; failing=% cleared=%', f, c; END IF;
  SELECT id INTO clear_id FROM doc_notes WHERE subject_key='spend-governor' AND categories ? 'account-wall-cleared' ORDER BY created_at DESC LIMIT 1;
  -- (e) another quiet tick → nothing
  EXECUTE q INTO f, w, c;
  IF w <> 0 OR c <> 0 THEN RAISE EXCEPTION '756 VERIFY: quiet tick after recovery wrote notes (wall=% cleared=%)', w, c; END IF;

  -- clean up EXACTLY the synthetic rows and notes; the training corpus must be untouched
  DELETE FROM llm_call_log WHERE id = fake2_id;
  DELETE FROM doc_notes WHERE id IN (wall_id, clear_id);
  IF (SELECT count(*) FROM llm_call_log) <> n0 THEN RAISE EXCEPTION '756 VERIFY: llm_call_log count changed (% -> %)', n0, (SELECT count(*) FROM llm_call_log); END IF;
  IF (SELECT count(*) FROM doc_notes WHERE subject_key='spend-governor') <> notes0 THEN RAISE EXCEPTION '756 VERIFY: spend-governor note count changed'; END IF;
  RAISE NOTICE '756 OK: quiet → nothing; synthetic wall → one note; repeat → none; recovery → one cleared note; quiet → nothing; llm_call_log and doc_notes restored exactly.';
END $$;

COMMIT;
