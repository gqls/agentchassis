# TODO: Orchestration `collected_data` growth causes OOM-kills and lost work

**Severity:** High — currently causing pod OOM-kills and dropped responses
**Status:** Diagnosed, not yet fixed
**Found:** 2026-05-19 during gaswholesalers news debugging session

## Observation

Component-quality-auditor orchestrations show `collected_data` of **18 MB each**
at step `create_regen_items_iter_8_check_quality`. Six identical-size
orchestrations all stuck on the same iteration step:

```sql
SELECT orchestration_id, current_step, pg_size_pretty(LENGTH(collected_data::text)::bigint)
FROM orchestration_states
WHERE owner_agent_type = 'component-quality-auditor'
  AND status IN ('AWAITING_RESPONSES', 'EXECUTING_STEP');
-- All 6 rows: 18 MB collected_size
```

Pod limits: `memory: 512Mi`. Loading a single 18MB `collected_data` plus the
Go runtime baseline (~80-150MiB) plus working buffers for LLM calls and
JSON encoding/decoding gets very close to the 512Mi cap. Auditor pods
have been observed being OOM-killed mid-processing:

```
Last State:     Terminated
  Reason:       OOMKilled
  Exit Code:    137
```

## Consequence chain

When an agent pod is OOM-killed:

1. Mid-processing work is lost
2. If the kafka request offset was already committed before the OOM, the
   request is NOT redelivered — work is gone permanently
3. If the agent's response wasn't produced before the OOM, the parent
   orchestration waits forever and eventually times out
4. Build-dispatch-loop FAILS its iteration
5. Generic parent FAILS its call_dispatch with "Request X timed out
   after 3 retries"

Today's session showed exactly this pattern: parent orchestrations
timed out at call_dispatch even though the children completed (per DB
state) — but the kafka response message was missing from the topic.

The OOM correlates with the empty topic because both stem from the
agent being killed during the publish phase.

## Suspected contributors to collected_data bloat

Without code-level profiling yet, the most likely contributors:

| Field | Why it grows | Mitigation |
|---|---|---|
| `__raw_message__` | Stores the entire inbound kafka message including the body. Duplicates information that's already structured in `input_data`. | Drop `__raw_message__` once parsing is complete |
| `processing_history` | JSONB array, one entry per status change per step. The auditor's iterative pattern (`iter_0`, `iter_1`...) compounds this. | Cap to last N entries, or move to a separate audit table |
| LLM responses | A single Claude response with tool use can be 50KB+. Across iterations, these accumulate in `collected_data` under each step's output_field. | Strip large fields from collected_data after the step has consumed them; only retain what downstream steps need |
| Component output across iterations | The auditor processes multiple components per run; each iteration adds its result to collected_data without releasing prior iterations' details. | Streaming-style writes (DB updates per iteration with old data discarded) |

## Investigation steps when picking this up

1. Dump one of the 18MB rows to a file and inspect what's actually in
   there:

   ```bash
   psql -h <host> -U <user> -d clients_db -tA -c "
     SELECT jsonb_pretty(collected_data)
     FROM orchestration_states
     WHERE orchestration_id = '914cc2ce-1278-42eb-9adc-174af4a52d54'
   " > /tmp/auditor_collected.json
   wc -c /tmp/auditor_collected.json
   # Then look at top-level keys and their sizes
   jq -r 'to_entries | map({key: .key, size: (.value | tostring | length)}) | sort_by(.size) | reverse' \
     /tmp/auditor_collected.json | head -30
   ```

2. Find which field is dominant. Almost certainly one or two large fields
   are responsible for the bulk.

3. Identify the lifecycle of those fields:
   - When is the data first added to collected_data?
   - Which subsequent steps actually read it?
   - When can it be safely dropped?

4. Patch the relevant action(s) to release the field once no longer needed.

## Short-term mitigations to apply before fixing root cause

While the proper fix is sized:

- **Raise the per-pod memory limit** from 512Mi to 1Gi or 2Gi. Nodes have
  headroom (11%/5%/21%/8%/37% used per `kubectl top nodes`). This
  prevents OOM-kills while the proper fix is developed.
- **Add `GOMEMLIMIT`** to the deployment env vars at ~88% of the k8s limit
  so Go GCs aggressively before the kernel OOMs:
  ```yaml
  env:
  - name: GOMEMLIMIT
    value: "880MiB"     # for a 1Gi limit
  - name: GOGC
    value: "75"
  ```
- **Strip debug symbols** from the binary: `go build -ldflags="-s -w" -trimpath`

These don't address the underlying bloat but stop the cascading failures.

## Why this matters more than it looks

OOM-kills are not just "the pod dies and gets restarted, no big deal".
They cause:

- **Phantom-completed orchestrations** (DB says done, no kafka response)
- **Cascading parent timeouts** that look like response-routing bugs
- **Hours-long debugging sessions** chasing routing issues that are
  actually memory-pressure consequences

This was today's failure mode. Until collected_data is bounded, periodic
OOM-cascades will mask other issues.

## Related issues already documented

- `031_locks.md` — leases on awaited_requests, related to claim race
- `015_batch_processing_architecture_v2.md` — iterator patterns are the
  context where collected_data accumulates fastest
- `016_debugging_guide_v2.md` Section 9 — should include an entry
  pointing to this todo

## Cross-reference: consumer group bug from same session

The consumer group bug (line 3152 of `platform/agentbase/agent.go` using
`a.AgentID` for the response consumer group) is structurally separate
from this memory issue but worth fixing in the same chassis update since
both require a chassis rebuild + deploy.
