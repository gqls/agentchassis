-- 0NN_council_gate_orchestrator.sql — run the council as a SPAWNED POD, not
-- inline on the shared chassis slot. 2026-07-27. Applies to clients_db.
-- Owner direction 2026-07-27: "coerce the council into our standard agent
-- workflow framework — spawning agents to run the workflows with the thin
-- orchestrator wrapper on top. We don't want to redesign the chassis to
-- accommodate it but treat it as a workflow."
--
-- ██ STATUS 2026-07-28 (evening): PROVEN END TO END, AND STILL **NOT** THE ██
-- ██ DEFAULT — deliberately. Council verdict REVISE; objections accepted.   ██
--
-- PROVEN. A real submission ran through this wrapper end to end on 2026-07-28:
-- published 15:20:07Z, wrapper at call_council/AWAITING_RESPONSES by ~15:20:44
-- (37s), the 16 seats executed in a dedicated pod, council_report 15:26:59Z,
-- wrapper COMPLETED 15:30:52Z (~10.75 min). Generic-lane LAG stayed 0
-- throughout. The mechanism works and the lane benefit is real.
--
-- NOT ADOPTED, and this is the load-bearing part. The proposal to flip 097's
-- default went to the gate (corr f5da8f65-a3ec-4d16-8254-3dbfcb76953c) and came
-- back REVISE, gated by the guardian: 11 reviewers, abstained 5, unreadable 0
-- (a real verdict, not harness noise). Its HIGH objection was aimed squarely at
-- an EARLIER VERSION OF THIS HEADER, which asserted "the blocker was never in
-- this wrapper":
--   bugs_open/003 and 029 (spawn-machinery defects) are still OPEN, and the
--   council path would inherit them for the first time. The evidence I had
--   resolved an ADJACENT bug (the response-lane treadmill), not those. Asserting
--   otherwise while overwriting a DO-NOT-FLIP warning is exactly the
--   drift-vs-live-behaviour mismatch this gate exists to catch. It was right.
-- Accepted, not argued. The urgency had gone anyway: the dedicated council lane
-- already stopped councils blocking other work, and replicas=2 already gives
-- two concurrent councils.
--
-- BEFORE ANY FUTURE THREAD FLIPS THE DEFAULT, meet these first:
--   1. bugs_open/003 + 029 closed, or explicitly accepted by the owner;
--   2. Anthropic concurrency headroom MEASURED (currently UNMEASURED);
--   3. bugs_open/124 fixed — duplicate chains double response-lane load.
-- Then change the 097 default and this header in the SAME edit.
-- Decision + full objections: doc_notes subject_key='council-gate-orchestrator'.
--
-- ── superseded, kept because the wrong turns are the useful part ────────────
-- ██ STATUS 2026-07-27: APPLIED AND LIVE, BUT **NOT** THE DEFAULT PATH. ██
-- Tested end to end the day it was written. The lane fix WORKS — measured
-- QUEUE DEPTH (LAG) 0 on system.agent.generic.requests while the council ran in
-- its own pod, where the inline path necessarily holds LAG >= 1 for 4-9 minutes.
-- The spawn handshake UNDERNEATH it does not: the run stuck at spawn_council /
-- AWAITING_RESPONSES and never reached call_council. The archetype this copies
-- fails the same way in 2 of its 4 runs over the whole retained history — "in
-- daily use" was not a reliability measurement, and that error is logged in
-- WRONG_CALLS.md.
--   CAUSE: still open. My first explanation — that the child answered 1.92s
--   before the parent began listening (19:50:24.189 vs 19:50:26.109) and the
--   reply was discarded — was REFUTED by the diagnosis loop (corr eb8df254) and
--   is wrong: persistAwaitingStateWithRetry (coordinator.go:1863-1879) returns
--   early on "Response already arrived during state persist - continuing", and
--   processResponseClaimWithRetry retries the claim for exactly that case. It is
--   a genuine NON-RESPONSE, fleet-wide (same shape on
--   build-pipeline-trigger/spawn_dispatch), i.e. the bugs_open/003 family.
--   Full correction + next scope: bugs_open/096 "Candidate 4".
-- >>> DO NOT flip 097_TRIGGER_council_review_v1.sh's default to this agent until
-- >>> the spawn race is fixed. See bugs_open/096 "Candidate 4" for the evidence
-- >>> and bugs_open/003 + 029 for the probable family.
--
-- ██ WHY ██
-- bugs_open/096: a council run holds the request lane for its WHOLE duration,
-- so every other dispatch on that lane waits behind it. The mechanism is not
-- council-specific but councils are the observed cause every time:
--   * every one of the 16 seats is `execute_llm_prompt`, registered
--     IsLocal:true (registry.go:307-312), so the seats run INLINE;
--   * consumeRequestLane calls processMessage as a BLOCKING call and commits
--     the offset after it returns (agentbase/agent.go:611-676);
--   * continueExecution is a strictly one-step-at-a-time state machine
--     (coordinator.go:837-975).
-- Measured 2026-07-27: 19 council runs, 4-9 min each; SEVEN of the last 25
-- started <=1.3s after the previous one ENDED (i.e. were already queued), and
-- eight ran back-to-back 18:09-18:54 with hand-off gaps of 0.1-20s.
--
-- The council gate's OWN guidelines seat objects to exactly this shape in other
-- people's plans (its prompt, this directory's 0NN_council_gate.sql):
--   "anything doing substantive work (LLM calls, crawls, heavy DB, minutes of
--    runtime) must run in a spawned pod via a parent (processing_mode:
--    'orchestrator' + spawn_agent), never inline on a shared chassis slot"
-- This seed makes the gate obey its own rule.
--
-- ██ WHAT THIS IS NOT ██
-- NO Go, NO image, NO chassis roll. DB config only, so it is LIVE the moment it
-- is applied and reverts by restoring the snapshot. That is a materially better
-- risk profile than the alternative considered first (extra EXTRA_REQUEST_TOPICS
-- lanes), which needed a chassis roll — and a roll kills in-flight
-- orchestrations (bugs_open/003; it happened during today's 19:22 UTC roll,
-- which left council-gate orchestration f849afaf wedged at review_guardian).
--
-- ██ WHY THIS SUPERSEDES THE MULTI-LANE IDEA ██
-- With the wrapper the generic lane is held only for spawn+call. Measured on
-- the archetype over the last 7 days: diagnose-orchestrator 527s avg vs its
-- child diagnose-agent 519s => ~8s of wrapper overhead (small sample, n=2/3).
-- So N councils run concurrently through ONE lane, and the "give councils 3
-- lanes" config change is not needed. Do not build it.
--
-- ██ THE ARCHETYPE ██ (read live from agent_definitions, not from a file)
-- diagnose-orchestrator = spawn_agent -> call_agent -> complete_workflow.
-- The closer precedent is fix-implementer-orchestrator -> fix-implementer,
-- because fix-implementer is ITSELF processing_mode:'orchestrator' with a
-- multi-step workflow — exactly council-gate's shape — and it is proven e2e
-- (feature-builder B4 -> PR #3 merged).
--
-- ██ FACTS CHECKED BEFORE WRITING THIS, each of which would have broken it ██
--  1. CONCURRENT SPAWNS DO NOT COLLIDE. spawn_actions.go:2364 does a
--     Get-then-Delete on an existing k8s Job BY NAME. If the name were derived
--     from type+role, the second concurrent council would delete the first
--     one's pod. It is not: agentID := uuid.New().String() (:53) and jobName :=
--     "agent-<type>-<agentID[:8]>" (:2360). Fresh UUID per spawn.
--  2. THE SPAWNED POD GETS THE ANTHROPIC KEY. All 16 seats are
--     provider=anthropic, model=claude-sonnet-5 (verified live). Spawned pods
--     receive ANTHROPIC_API_KEY from the shared allow-list
--     (platform/agentenv/provider_keys.go:52) plus personae-prod-config and the
--     DB secrets. Ollama is NOT on the council's path at all — the other 26
--     steps are deterministic (conditional/query_database/diagnose_*/
--     append_doc_note/complete_workflow), there is no RAG or embedding step,
--     and code_lookup is deliberately not mirrored from fix-proposer.
--  3. THE CHILD RUNS ITS FULL WORKFLOW. call_agent's default target_action is
--     "process" (call_agent.go:455-470), which routes to workflowMode "task"
--     (processor.go:1155-1163, because from_agent_id is set). "task" only
--     overrides the workflow if default_config.task_workflow exists
--     (processor.go:1124-1133); council-gate has none, so it runs its full
--     default_config.workflow. This is why the archetype works.
--  4. THE LANE IS ACTUALLY RELEASED. call_agent returns await_response:true
--     (call_agent.go:637-649) -> processAwaitResponse -> AWAITING_RESPONSES ->
--     continueExecution RETURNS. The goroutine is free.
--  5. NO READINESS RACE. The child's job.<id>.requests topic holds the message
--     until the pod boots, so call_agent cannot outrun the spawn.
--  6. NO INPUT CONTRACT TO FAIL. council-gate.input_contract IS NULL, so
--     ValidateInputContract is skipped. ResolveInputMapping copies whatever is
--     at each path, so the nested `plan` object forwards intact.
--  7. THE IMAGE IS THE SPAWNER'S, not the (never-updated) agent_definitions
--     row — bugs_open/066, resolved and live in v1.0.1177.
--
-- ██ WHAT THIS DOES NOT FIX ██
-- A single council is still ~6 minutes of 16 SEQUENTIAL seats. This buys
-- concurrency, not latency. Making the seats concurrent needs the `fan_out`
-- action, which today has a validation contract (validation/workflow.go:98-105)
-- and a fuel cost (governance/fuel.go:17) but NO HANDLER IN THE REGISTRY, so a
-- fan_out step fails at getActionHandler. It is half-built, not absent. That is
-- a separate piece of work, and it is the same "treat it as a workflow" move.
--
-- ██ ROLLBACK ██
--   DELETE FROM agent_definitions WHERE type='council-gate-orchestrator';
--   -- and point 097_TRIGGER_council_review_v1.sh back at council-gate
--   -- (TARGET_AGENT_TYPE env override, default in the script).
-- The council-gate row itself is NOT restructured by this seed, so reverting
-- the wrapper restores the previous behaviour exactly.

BEGIN;

-- Idempotent re-apply path (no-op on first apply — the type does not exist yet).
SELECT snapshot_agent('council-gate-orchestrator', 'pre-update: council gate wrapper (bugs_open/096)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='council-gate-orchestrator' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'council-gate-orchestrator',
    'Council Gate Orchestrator (spawns the council into its own pod)',
    'Thin wrapper: spawns council-gate as a DEDICATED k8s Job and forwards the submission verbatim, so the 16 sequential LLM seats run in their own pod instead of holding the shared chassis request lane for 4-9 minutes (bugs_open/096). Mirrors diagnose-orchestrator -> diagnose-agent and fix-implementer-orchestrator -> fix-implementer. Releases the lane in ~8s, which lets councils run concurrently with each other and with ordinary dispatches. Does not change what the council does or how it judges.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "review"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'spawn_council',
      'processing_mode', 'orchestrator',
      -- Must exceed the child's own timeout (council-gate: 600s) with room for
      -- the spawn. Mirrors fix-implementer-orchestrator's 2000/1800 pairing.
      -- continueExecution's stale-orchestration guard uses timeout_seconds * 3.
      'timeout_seconds', 2000,
      'steps', jsonb_build_object(

        'spawn_council', jsonb_build_object(
          'action', 'spawn_agent',
          'description', 'Spawn a dedicated council-gate pod. Job name carries a fresh UUID, so concurrent councils cannot collide.',
          'next_step', 'call_council',
          'config', jsonb_build_object(
            'role', 'reviewer',
            'agent_type', 'council-gate'
          )
        ),

        'call_council', jsonb_build_object(
          'action', 'call_agent',
          'description', 'Hand the submission to the spawned council and await its verdict. await_response defaults true, which is what releases the request lane.',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'agent_type', 'council-gate',
            'target_role', 'reviewer',
            'timeout_seconds', 1800,
            -- The four fields 097_TRIGGER_council_review_v1.sh publishes. Names
            -- must match what the council's own steps read: persist_submission
            -- resolves plan_field over input_data.plan, and every reviewer
            -- prompt templates {{.input_data.rationale}}. `submitter` is
            -- optional (the trigger defaults it to "unnamed") — marked ? so a
            -- caller that omits it does not fail the whole mapping.
            'input_mapping', jsonb_build_object(
              'fix_correlation_id', 'input_data.fix_correlation_id',
              'rationale',          'input_data.rationale',
              'plan',               'input_data.plan',
              'submitter?',         'input_data.submitter'
            )
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Forward the council verdict unchanged. The artifacts (council_report, doc_notes) are written by the CHILD under the submission correlation, so the trail is unaffected by the wrapper.',
          'config', jsonb_build_object('result_from', 'call_council')
        )
      )
    ))
FROM agent_definitions d
WHERE d.type = 'diagnose-orchestrator'
  AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
ON CONFLICT (type, version) DO UPDATE
   SET default_config = EXCLUDED.default_config,
       description    = EXCLUDED.description,
       updated_at     = now();

-- Reap finished council pods promptly. council-gate.idle_timeout_seconds is 0,
-- which falls through to the 3600s default at spawn_actions.go:2691-2693 — so
-- every finished council pod would loiter for an HOUR. At 19 runs/day that is a
-- lot of idle pods once councils can overlap. 900s leaves generous headroom over
-- the longest observed run (520s) and over the child's own 600s timeout.
-- Targeted single-column UPDATE, NOT a re-seed: a whole-object rewrite here is
-- the config-clobber class (multi_session_coordination/FINDING_2026-07-17).
UPDATE agent_definitions
   SET idle_timeout_seconds = 900, updated_at = now()
 WHERE type = 'council-gate' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND idle_timeout_seconds <> 900;

COMMIT;

-- Verify:
--   SELECT type, default_config->'workflow'->>'start_step',
--          default_config->'workflow'->>'timeout_seconds'
--     FROM agent_definitions WHERE type='council-gate-orchestrator' AND is_active;
--   SELECT type, idle_timeout_seconds FROM agent_definitions
--    WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Then ONE test submission before retargeting the trigger for everyone:
--   TARGET_AGENT_TYPE=council-gate-orchestrator \
--     ./097_TRIGGER_council_review_v1.sh <submission.json>
-- What proves it worked (all four, not just the verdict):
--   1. the wrapper reaches AWAITING_RESPONSES within seconds of publish;
--   2. `kubectl get jobs -n ai-persona-system | grep agent-council-gate` shows a
--      dedicated pod;
--   3. an UNRELATED dispatch published while the council runs starts immediately;
--   4. the council_report artifact lands under the submission correlation.
