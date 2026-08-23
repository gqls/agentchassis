-- 573 — the FRONT DOOR stops reporting success over an empty queue: domain-submitter's
--       `create_research_item` sets `on_dedup: "error"`, so a submission that queues no
--       work ends FAILED, naming the item that already holds the key.
--
--       bugs_open/326, the second half. 572 made a re-submission WORK; this makes the
--       remaining refusal LEGIBLE.
--
-- ============================================================================
-- ⚠ _HOLD — DO NOT APPLY UNTIL THE CHASSIS ROLL CARRYING `on_dedup` IS LIVE.
--            THIS IS A REAL ORDERING CONSTRAINT, NOT A PREFERENCE.
-- ============================================================================
-- `create_work_item` is StrictConfig (create_work_item_action.go, opted in by
-- bugs_open/234). Under StrictConfig an unrecognised config key is a HARD DEFINITION
-- ERROR, not a warning — and `ValidateWorkflow` runs on EVERY MESSAGE
-- (platform/messaging/processor.go), not once at seed time.
--
-- So applying this against a binary that does not yet declare `on_dedup` does not degrade
-- gracefully and does not fail once: it fails EVERY domain-submitter run, for as long as
-- it is applied. The front door stops entirely.
--
-- Apply only after:
--   1. the Go change is committed and built (`make build-agent-chassis` builds committed
--      HEAD, so commit first);
--   2. the fleet has rolled;
--   3. the running chassis is confirmed to carry it — ask the service, do not infer:
--        kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--          | grep -m1 'build provenance'
--      then `git merge-base --is-ancestor <the 326 commit> <the stamped sha>`.
--      An empty grep means "not in range" (it is a STARTUP line and it scrolls), not
--      "unstamped" — fall back to the binary probe, with a control sha that must be
--      ABSENT and one that must be PRESENT.
--
-- The `_HOLD` suffix keeps the migration runner's `--apply` from taking this file with the
-- rest of the directory while still listing it, which is the documented idiom for exactly
-- this case: it IS the change, held back for ordering.
--
-- ============================================================================
-- WHAT IT DOES, in plain terms
-- ============================================================================
-- `on_dedup` says what the step should do when its write queued NOTHING. The default,
-- and every other step's behaviour, is to shrug and report success — which is right for a
-- mid-pipeline handoff, where several producers converging on one open item is the whole
-- point of the dedup.
--
-- The front door is the one place where it is wrong. An operator running
-- `082_submit_domain_unified.sh` has asked for a build. If no build was queued, "COMPLETED,
-- no error" is not a summary of what happened — it is the opposite of it. With
-- `on_dedup: "error"` the orchestration ends FAILED and `orchestration_states.error`
-- carries the answer rather than a silence:
--
--     on_dedup=error: no work was queued for item_key "research_example.uk" — an open item
--     already covers it: 3f2a… (status claimed, 0.4h old). The work this step exists to
--     queue is already in hand; if it is stalled, resolve or cancel that item and try
--     again (bugs_open/326)
--
-- AFTER 572, THIS SHOULD BE RARE AND IT SHOULD ALWAYS BE TRUE. 572 exempts this step from
-- the anti-churn brake, so the only remaining way to queue nothing is a genuine open
-- holder — i.e. a build for this domain really is already running. That is a fact worth
-- telling an operator, and it is the one case where refusing is not a regression in
-- disguise.
--
-- SCOPE: this step only. Mid-chain handoffs keep the default and must keep it — a stage
-- item converging on one open row is normal there, and failing those runs would turn a
-- working dedup into an outage.
--
-- ============================================================================
-- APPLY (by hand, after the roll)
-- ============================================================================
BEGIN;

-- RE-RUN SAFETY, with the snapshot INSIDE the guard (council round 1, correlation
-- f610741f, debug_historian, medium — and it flagged THIS file as the higher risk
-- of the two, correctly: a _HOLD migration is applied BY HAND, out of band, by an
-- operator following a runbook, which is exactly the setting where an accidental
-- second apply happens).
--
-- An unconditional `SELECT snapshot_agent(...)` before a fenced UPDATE takes a
-- second snapshot on re-run whose reason still says "pre-update" while describing
-- an already-updated row — corrupting the rollback lineage at the moment someone
-- is reaching for it. Gate it on the same pre-state marker that drives the UPDATE
-- and a re-run becomes a true no-op instead.
DO $$
DECLARE current_val TEXT; dupes INT;
BEGIN
    SELECT count(*) INTO dupes FROM agent_definitions
     WHERE type = 'domain-submitter' AND is_active
       AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF dupes <> 1 THEN
        RAISE EXCEPTION '573 REFUSED: domain-submitter has % active definition rows, expected 1 — '
            'only the higher version is loaded at runtime, so a version-blind UPDATE could '
            'patch a row that governs nothing while this migration reported success.', dupes;
    END IF;

    SELECT default_config->'workflow'->'steps'->'create_research_item'->'config'->>'on_dedup'
      INTO current_val
      FROM agent_definitions
     WHERE type = 'domain-submitter' AND is_active
       AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF current_val = 'error' THEN
        RAISE NOTICE '573 NO-OP: on_dedup is already "error"; skipping the snapshot so the '
            'rollback lineage is not polluted with a post-update row labelled pre-update.';
    ELSIF current_val IS NOT NULL THEN
        RAISE EXCEPTION '573 REFUSED: on_dedup is already %, not NULL and not "error" — '
            'somebody set it deliberately. Resolve by hand.', current_val;
    ELSE
        PERFORM snapshot_agent('domain-submitter',
            'bugs_open/326: on_dedup=error so a submission that queues nothing cannot report COMPLETED');
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{workflow,steps,create_research_item,config,on_dedup}', '"error"'::jsonb),
       updated_at = NOW()
 WHERE type = 'domain-submitter'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

DO $$
DECLARE v TEXT; act TEXT;
BEGIN
    SELECT default_config->'workflow'->'steps'->'create_research_item'->'config'->>'on_dedup',
           default_config->'workflow'->'steps'->'create_research_item'->>'action'
      INTO v, act
      FROM agent_definitions
     WHERE type = 'domain-submitter'
       AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF act IS DISTINCT FROM 'create_work_item' THEN
        RAISE EXCEPTION '573 FAILED: create_research_item is not a create_work_item step (action=%) — jsonb_set would have written the key onto nothing', act;
    END IF;
    IF v IS DISTINCT FROM 'error' THEN
        RAISE EXCEPTION '573 FAILED: on_dedup is %, expected "error"', COALESCE(v,'NULL');
    END IF;
    RAISE NOTICE '573 OK: domain-submitter.create_research_item on_dedup=error';
END $$;

COMMIT;
