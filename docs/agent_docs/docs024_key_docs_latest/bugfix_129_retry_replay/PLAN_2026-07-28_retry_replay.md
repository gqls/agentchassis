# PLAN — a retry must be a REPLAY of the original request, not a reconstruction

**Workstream opened 2026-07-28 (bugsearch 4 thread).** Bug: `bugs_open/129` —
*"the spawned child ADOPTS the parent's orchestration row and silently declines
the work"*. Ownership checked before starting: `scripts/who-owns.py 129_HANDOFF…`
→ *"some activity in 14d but no clear owner"*; the citing dirs
(`architecture_review`, `bugs_sweep_2026_07`, `gauntlet_dead_cta`) cite it as
cross-reference only, and `bugs_sweep`'s own handoff names **091** as its next
item, not 129.

---

## What 129 filed, and what it actually is

129 filed the **child-side symptom**, with pod logs: the child receives a request,
loads state by `orchestration_id`, finds the **parent's** row at
`AWAITING_RESPONSES`, concludes "already awaiting — nothing to do", logs
`ProcessMessage completed successfully`, and never replies. The parent times out
after three retries.

Its fix candidates were ordered child-side (1: key the child's state by its own
identity; 2: don't decline the request you are currently holding; 3: reply
something). **All three treat the child as the defect.** It is not. The child did
exactly what it was told: it was *handed the parent's orchestration id*.

**Root cause, read from the code and confirmed by the whole retried population:**

`platform/orchestration/coordinator.go` → `handleRecoverableError` (the live
retry path, reached from `retryExpiredAwaitedRequest`) does not resend the
original request. It **synthesises a new one** out of the *awaiting* orchestration's
own state:

```go
retryRequest := &types.RequestMessage{
    Headers: types.RequestHeaders{
        OrchestrationID:   state.OrchestrationID,   // ← the PARENT's id
        OrchestrationName: state.OrchestrationName,
        …
        Action:            "execute",               // ← never the original action
    },
    Body: map[string]interface{}{                   // ← the original body is GONE
        "is_retry":      true,
        "retry_version": awaited.RetryVersion,
    },
}
```

Three defects in one message, all from the same cause — *the retry is built from
what the waiter knows, not from what was sent*:

1. **Identity.** The child gets the parent's `orchestration_id`, so
   `getOrCreateState` Priority 1 (`GetState(execCtx.OrchestrationID)`) resolves the
   parent's row, `handleOrchestrationStatus` hits `StatusAwaitingResponses` and
   returns `ErrWaitingForResponse`. **This is 129's swallow, exactly.** No
   `ParentOrchestrationID` is set either, so the "Creating child orchestration"
   branch that would have rescued it cannot fire.
2. **Payload.** The original body never travels. Even with the identity fixed, the
   child would execute its workflow on `{is_retry:true, retry_version:N}`.
3. **Action.** Always `"execute"`, never the original (`initialize` for a spawn,
   the configured target action for a call).

The code already **knows** this is wrong, in a comment eight lines above:

> ```go
> // Adapter actions (git_commit, etc.) need to re-execute the step, not send
> // a retry message, because adapters need the full payload (files, etc.)
> ```

Adapters got an exemption. Agents never did.

## The measurement — the population, not an example

Every figure below is from the live `clients_db`, 2026-07-28, and is a **query, not
an argument**.

```sql
SELECT CASE WHEN COALESCE(target_agent_id,'')='' AND requests_topic LIKE 'system.adapter%'
            THEN 'adapter: re-executes step (already correct)'
            ELSE 'agent: synthesised retry (poisoned)' END AS path,
       count(*) AS retried_14d, count(*) FILTER (WHERE retry_version=3) AS exhausted
FROM awaited_requests WHERE sent_at > now() - interval '14 days' AND retry_version > 0
GROUP BY 1;
```

| path | retried, 14d | exhausted the budget |
|---|---|---|
| agent: synthesised retry (poisoned) | **430** | **294 (68%)** |
| adapter: re-executes step | **0** | 0 |

**Every retry the fleet has sent in fourteen days took the poisoned path.** The
adapter exemption is dead code for retries. And all-history the outcome
distribution is the shape of a retry that cannot work — 93 at `retry_version` 1,
45 at 2, and **294 at 3**. A retry that recovers decays; this one accumulates at
the cap.

Concentration (14d, retried, by step): `spawn_dispatch` 154 (126 exhausted),
`call_dispatch` 120 (91), then `process_item_iter_*_call_handler`, `spawn_ingester`,
`deploy_page`, `call_diagnoser` … — i.e. **`spawn_agent` and `call_agent` are
essentially the entire population**, which is what makes this fixable in two files.

`[UNMEASURED]` How many of the 108 requests that ended `processed` at
`retry_version` ≥ 1 were rescued *by the retry* rather than by a late original
response arriving. The two are indistinguishable in `awaited_requests`. So the
claim here is **"68% of retried requests exhausted the budget"**, which is
measured — *not* "retries never work", which is not.

## The fix — one invariant

> **A retry is a replay of the original request. Only `retry_version`,
> `message_id` and `timestamp` may differ.**

That makes all three defects unrepresentable at once, because none of the three
fields is reconstructed any more.

| # | change | file | why here |
|---|---|---|---|
| 1 | the sending action records the exact produced message under `result["retry_payload"]` (`{topic, key, headers, body}`) | `actions/spawn_actions.go`, `actions/call_agent.go` | the action is the only place that knows what was sent; `result[…]` is the existing convention (`requests_topic`, `responses_topic`, `target_agent_type` already travel this way) |
| 2 | `createAwaitedRequest` lifts it onto `AwaitedRequest.RequestPayload`, tagged `json:"-"` | `coordinator.go`, `state.go` | `json:"-"` keeps it **out** of the `orchestration_states.awaited_requests` JSONB, which is rewritten on every state update. The payload belongs on the per-request row, not the hot one |
| 3 | `InsertAwaitedRequest` persists it to a new nullable `awaited_requests.request_payload jsonb`; `ON CONFLICT … DO UPDATE` fills it onto a row `preRegisterAwaitedRequest` created first | `state.go` + migration | **one** write site. Spawn pre-registers before it sends, so the insert must be able to back-fill |
| 4 | `handleRecoverableError` re-produces the stored message verbatim | `coordinator.go` | `GetAwaitedRequest` is *already* called here to read `retry_version`, so the payload arrives with no extra query |
| 5 | **hard guard**: the coordinator may never emit a request whose `orchestration_id` equals the awaiting orchestration's | `coordinator.go`, `helpers.go` | this is the invariant itself. Without a stored payload it fails the request with a named error rather than sending one that will be swallowed |
| 6 | child-side: an inbound **request** whose `request_id` the loaded row is itself awaiting is misrouted by construction — never report success | `coordinator.go` | defence in depth: 129's own candidate 2, and it holds against any future sender that gets identity wrong |

### Why not the alternatives

- **"Just re-execute the step", generalising the adapter branch.** Correct by
  construction (same code rebuilds the request) and needs no storage — but for
  `spawn_agent`, which is the single largest retried step (154/430), it spawns a
  **second pod** and orphans the first, which is the `bugs_open/124` double-dispatch
  class. Rejected for the dominant case; retained as the adapter path's behaviour,
  untouched.
- **Fix the child only (129's candidates 1–3).** Leaves every retry fleet-wide
  still carrying an empty body. Converts a silent stall into a child that runs on
  `{is_retry:true}` — a wrong answer instead of no answer. Strictly worse alone.
- **Store the payload in the state JSONB.** No migration needed, but
  `orchestration_states.awaited_requests` is rewritten on every state update, so a
  body would be re-serialised many times per orchestration. Hence `json:"-"` + a
  column on the per-request table.

## Ordering, and the seam declaration

This adds a column and changes a **shared** mechanism (every awaited request in
the fleet), so per the 2026-07-28 owner ruling it is **architecture-scope** and is
going to the council gate. The ordering constraint is real and is stated here
rather than assumed: **the nullable column must exist before the binary that
writes it**, so migration first, image second. The old binary ignores the column;
the new binary tolerates NULL (that is what guard 5 is for). There is no window in
which either half breaks the other.

## How it will be verified

Induced, not inferred (`bugs_open/129` "How to verify a fix"):

1. `TRIGGER_code_indexer_v2.sh` — dispatch `index-orchestrator`.
2. The child's log must show `fetched repo source`; `spawn_indexer → call_indexer`
   must advance.
3. The discriminating pod-grep is a string this change **deletes** — see
   `RUNBOOK_retry_replay.md`; the obvious positive markers are vacuous because the
   surrounding words already exist in the binary.
