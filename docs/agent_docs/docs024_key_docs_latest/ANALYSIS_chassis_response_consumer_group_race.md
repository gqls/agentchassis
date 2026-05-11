# Analysis — Chassis response-topic consumer group race condition

**Date observed:** 2026-05-10
**Status:** Discovery, not yet remediated. For deep review against previous architecture decisions and future plans before any change is made.
**Triggering observation:** Two chassis pods (`jpzjz` and `bpwpp`) each processed the same `spawn_image_gen` response for orchestration `c32d257e-…`, causing `call_logo_gen` to fail. See section 1 below.

---

## 1. What was observed (facts, not interpretation)

Orchestration `c32d257e-fd72-4682-9860-a6cc8376fe40` failed at `call_logo_gen`. The audit trail shows two state transitions on `call_logo_gen` ~500ms apart, the second flipping to FAILED:

```
13 → 14   EXECUTING_STEP   call_logo_gen   16:23:14.771
14 → 15   EXECUTING_STEP → FAILED   call_logo_gen   16:23:15.270
```

Chassis logs show two distinct chassis pods each running `ProcessResponse` on the same spawn-response message, ~215ms apart:

```
16:23:10.250Z   pod jpzjz (agent_id 62300d3b…)   ProcessResponse
16:23:10.465Z   pod bpwpp (agent_id ea264562…)   ProcessResponse
```

`kafka-consumer-groups.sh --describe --all-groups` filtered on `system.agent.generic.responses` returns **~85 consumer groups** subscribed to the topic. Only **three** have a live `CONSUMER-ID` (three currently-running chassis pods); the rest are stale groups with old offsets and no members. Each group ID is a UUID matching the pattern of `agent_id`.

Chassis pod env (`KAFKA_CONSUMER_GROUP=generic-requests-group`) shows the **requests** topic uses a shared, stable group name. The responses topic does not — each pod is its own group.

`collected_data` for the failed orchestration has both `image_gen_agent` and `spawn_image_gen` keys, both holding the same spawn response object. The agent definition says `spawn_image_gen.output_field = "image_gen_agent"`, so only `image_gen_agent` should be present. The duplicate `spawn_image_gen` key is consistent with two pods each writing the spawn result.

## 2. What the code confirms

The code in `platform/agentbase/client.go`, `server.go`, `consumer.go`, and `helpers.go` shows the actual mechanism:

- **`AgentServer`** (requests path) is constructed with `consumerGroup` passed in by whatever wires it up. The env-var `KAFKA_CONSUMER_GROUP=generic-requests-group` shows this gets a stable, shared name for the generic chassis pool.
- **`AgentClient`** (responses path) is constructed with the *same* `consumerGroup` string passed in. The client appends `"-responses"` to it when logging but the actual `groupID` passed to `kafka.NewConsumer` is whatever the constructor caller hands it.
- **`kafka.NewConsumer`** in `consumer.go` uses the `groupID` parameter verbatim as the kafka-go `ReaderConfig.GroupID`. Whatever the caller gives it becomes the kafka group.

The wiring layer that constructs `AgentClient` (not in the files I've seen — would be in `cmd/chassis-startup` or wherever `NewAgentClient` is called) is passing a per-pod UUID as the group, not the stable `generic-requests-group` name. That's the proximal cause.

I have **not** seen the wiring code that does this, so the exact line where the misconfiguration happens is still TBD. The README_consumer_groups.md fragment lists historical groups including ones like `chief-strategist-group-e556f9ef` (per-agent-id) and `generic-requests-group` (shared), suggesting the inconsistency has been present for a while across multiple agent types — not a recent regression.

## 3. Why this produces the failure mode observed

Kafka consumer-group semantics: within one group, partitions are distributed among consumers and each message goes to exactly one consumer in the group. Across groups, each group gets a full copy of every message.

If every chassis pod is in its own per-pod group on the responses topic:

- Every response message is delivered to every chassis pod.
- Each pod independently runs `ProcessResponse`, which advances the orchestration state in postgres.
- Whichever pod gets there first advances version N → N+1.
- The second pod tries to advance from a version that's now stale, hits whatever optimistic-concurrency check the coordinator has, and produces a failed transition (audit row 14 → 15).

This explains:
- Why two pods log `ProcessResponse` for the same message.
- Why two audit transitions appear close together.
- Why the second transition is FAILED.
- Why `collected_data` has duplicate keys (both writes happened, but one was the canonical output_field and the other was the step name).
- Why no error appears in `agent_error_log` — the second pod's failure is in the orchestration state machine layer, not in an action that calls the error-log writer.

## 4. Why this hasn't been catastrophic

Most failure modes from this race are silent:

- **Idempotent updates.** Many response handlers do an UPSERT or `jsonb_set` style write that produces the same result regardless of which pod runs it. Both pods "succeed", offsets get committed twice, no visible damage.
- **One pod processes faster.** When the second pod arrives at a step that's already advanced, its DB write fails on the version check, the pod logs a warning at most, and the message is offset-committed. The first pod's work stands.
- **Per-spawn topics shield child→parent calls.** The `job.<stable-id>.requests`/`.responses` topics used between spawned parent and child are per-orchestration and consumed only by that one pair. Those *might* be configured with a per-spawn group that effectively has one member, which would be safe by accident — needs verification.

The case that bites is when the timing makes the second pod's stale-state write produce a *FAILED* transition on a step the workflow engine considers terminal. `call_logo_gen` happened to be in that category today.

## 5. What this means for the migration

The Phase 2F migration's new `spawn_asset_deployer → call_asset_deployer` pair is structurally identical to `spawn_image_gen → call_logo_gen`. It will hit the same race when exercised. Runtime testing of the migration is blocked until either:

(a) the consumer-group misconfiguration is fixed, so each response goes to exactly one chassis pod, or
(b) we deliberately scale the chassis to 1 replica during testing to dodge the race.

(b) is a quick hack that confirms the migration works in isolation but doesn't validate concurrent behaviour.

## 6. Discussion points to revisit deeply

These are not decisions. Each one deserves to be examined against existing decisions and the future direction.

### 6a. The intended consumer-group model

The system has at least three distinct kafka topic patterns:

- **Fixed shared topics** (`system.agent.generic.requests`, `system.agent.<adapter-type>.process`) — many pods need to share work. Shared consumer group is correct.
- **Per-spawn dedicated topics** (`job.<stable-id>.requests`/`.responses`) — one parent and one child. Either a single-consumer setup or a per-spawn group works.
- **The generic responses topic** (`system.agent.generic.responses`) — this is where the misconfiguration shows up. Today every pod gets every message.

The question to discuss: is the generic responses topic supposed to behave like a shared work pool (one pod handles each response, all pods share the load) or like a fan-out broadcast (every pod sees every response, for some reason)? Today it's accidentally the second. The behaviour the system relies on is the first.

If the answer is "shared work pool", the fix is: every chassis pod joins one stable consumer group on this topic, just like they do for the requests topic.

### 6b. Whether the per-spawn topics have the same defect

`job.<stable-id>.requests` and `job.<stable-id>.responses` are created per spawn. The parent's chassis pod (or any chassis pod) consumes the responses topic; the child consumes the requests topic.

Worth verifying:
- Does the parent's consumer on the per-spawn responses topic use a stable group, a per-pod group, or a per-spawn group?
- If per-spawn, that's effectively single-consumer because no other parent is interested in those responses. Safe by construction.
- If per-pod, the parent would also race here when multiple chassis pods are running.

The fact that spawn-response *did* race suggests this might be per-pod. But the responses topic on which the duplicate processing happened in the logs was `system.agent.generic.responses`, not a `job.*.responses` topic. So this is open.

### 6c. The role of optimistic-concurrency in the state machine

The audit table has `old_version` and `new_version` columns. The version increments on every transition. This suggests the saga coordinator already does some form of CAS-style update (`UPDATE … WHERE version = old_version`). If it does, the second writer should fail cleanly and not flip the state to FAILED.

That a FAILED transition was recorded means either:

- The CAS check exists but the failure path on stale-version writes is mapping to FAILED instead of "no-op, drop the duplicate response".
- The CAS check doesn't exist and both pods successfully wrote, and the FAILED came from some other path I haven't traced (e.g., one pod's `CallAgentAction` errored on a topic-resolution issue because the other pod had already mutated the state).

This is a separate issue from the consumer-group misconfiguration but interacts with it. Fixing the group makes the second writer disappear; hardening the CAS path makes the system robust to spurious duplicate deliveries from any cause (kafka retries on rebalance, etc.).

### 6d. The 82 abandoned consumer groups

Stale groups don't consume, but they sit in kafka's group metadata and clutter `--describe --all-groups`. Each one is a previous chassis pod's per-pod group from before that pod terminated. Worth a cleanup once the model is right, and possibly a scheduled sweep — but not urgent.

### 6e. The architectural question raised in the previous discussion

> if a pod goes down it is expected that another will take over the processing — I don't know how far we've tested this, but it isn't or shouldn't eventually be one pod one message I don't think. Please critically assess where we are with this or if we should perhaps be a different mechanism like perhaps if a message isn't replied to then a new message is sent to a different container

Two separable concerns:

1. **Failover when a pod dies mid-processing.** Kafka's standard consumer-group rebalancing handles this: when a pod drops out of the group, kafka reassigns its partitions to surviving members. Messages whose offsets weren't committed get reread by whichever pod inherits the partition. This is what you want and it requires the shared-group model. The current per-pod groups *don't* give failover — they give duplicate delivery.

2. **Timeout-and-retry for genuinely lost replies.** Different mechanism, different scope. `awaited_requests.timeout_at` exists for this. When a child agent never responds at all (the request was dropped, the child crashed, etc.), the timeout fires and the parent's orchestration can decide what to do — retry, fail, escalate. Mixing this with the consumer-group fix would conflate two different problems; they should be designed independently.

The current discovery is purely about (1). (2) is worth examining separately at some point but isn't the cause of what we saw today.

## 7. What I'm not concluding

- **Not** that the migration is involved. The migration is committed and SQL-verified. The runtime path it adds will suffer the same race once we get past the logo failure, but the race is upstream of any step the migration changed.
- **Not** that the consumer-group fix is trivial. I haven't seen the wiring code that constructs `AgentClient`. The actual change might be one line; it might also reveal other places that depend on the per-pod-group behaviour for reasons I haven't traced yet.
- **Not** that the chassis is the only place this pattern exists. Other agent types listed in the README_consumer_groups.md fragment also appear to have per-agent-id groups on their responses topics (e.g., `chief-strategist-group-e556f9ef`). If any of them run with multiple replicas, they'd race too. If they all run as single replicas (which most spawned-Job pods do, since the Job creates exactly one), it doesn't matter for them — but the generic chassis Deployment has 3 replicas, which is why we hit it here.

## 8. Open questions to answer before designing the fix

1. **What's the intended model for the generic responses topic?** Shared work pool, or fan-out? The system's behaviour up to today *relies* on shared (only one pod should advance state per response), so this is mostly a confirmation question.

2. **What's the model for `job.<stable-id>.responses`?** If a Job pod is always single-replica, this might not have mattered yet — but if the chassis Deployment is the one consuming those topics, we have the same multi-consumer problem there.

3. **Does the saga coordinator's `ProcessResponse` use optimistic concurrency?** If yes, why did the duplicate produce a FAILED rather than a no-op? If no, that's a separate hardening item.

4. **Are there callers that genuinely *want* fan-out delivery?** I can't think of a good reason — observability/audit systems would have their own pipeline — but worth confirming nobody's leaning on it.

5. **What does the wiring code look like?** The constructor call to `NewAgentClient` somewhere in `cmd/chassis-startup` or the bootstrap path is where the group ID gets chosen. Reading that code would make the fix concrete.

6. **What about the requests-topic group for non-generic agents?** The README fragment shows historical groups like `chief-strategist-group-e556f9ef` — agent-type-suffixed-with-id. Are those per-agent-id (one consumer per group, racy if scaled) or per-agent-type (shared across replicas)? Separate from today's problem but related.

## 9. Suggested next steps (for discussion, not action)

In rough order:

1. **Trace the wiring.** Find where `NewAgentClient` is called for the generic chassis and read what `consumerGroup` argument it receives. That tells us where the per-pod group name is coming from.
2. **Confirm the per-spawn topic group model.** Same kind of trace for the per-spawn topic subscribers.
3. **Verify the saga coordinator's CAS behaviour.** Read `ProcessResponse` in `coordinator.go` — does it check `old_version` before update? What does it do when the check fails?
4. **Design the fix.** Most likely a one-line change to pass `"generic-chassis-responses"` (or similar stable name) instead of an agent_id-derived string. But the design needs to satisfy the questions above, not just patch the symptom.
5. **Plan the rollout.** The fix changes consumer-group membership. Existing in-flight messages on per-pod groups will be left behind (those groups stop having consumers); new messages flow to the new shared group from the new offset. Worth thinking about what's in flight at the moment of switchover.
6. **Decide what to do about Phase 2F runtime testing in the meantime.** Either wait for the fix, or scale chassis to 1 replica for an isolated test of the new spawn+call pair.

---

*End of analysis. Nothing in this document changes anything. Material for deep review and discussion before any code/config change.*
