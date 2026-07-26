# RUNBOOK — bugs_closed/040 candidate 2, induced-fault probe

Every command here had to be got right once. The gotcha is attached to the command;
when one changes, change it **here**, not in your scrollback.

Case file: `bugs_closed/040_HANDOFF_2026-07-20_failed_page_build_leaves_page_deployed_and_partially_composed.md`

---

## 1. What the probe proves

`update_work_item_status` falls back to the routed step error when the workflow supplied
no `error_message` literal — but never for `complete`. Three arms, one run:

| item | ends | `error` must read | tests |
|---|---|---|---|
| PROBE A `d418240c-…` | `failed` | `step boom failed: …` | the fallback fires |
| PROBE B `b0b0b0b0-1111-4222-8333-444444444444` | `needs_human_review` | `cand2 probe literal: …` | a configured literal still **wins** |
| PROBE C `1b001fec-…` | `complete` | **BLANK** | the `complete` exclusion holds |

B and C both run *after* `boom` has set `__step_error`, which is never cleared — that is
the point. Agent type `scratch-cand2-probe`; `item_type='scratch_cand2_probe'` has no
handler, so the dispatch loop cannot pick these up by accident.

## 2. Reset before every run

`attempt_count` is **not** reset by the workflow, and `max_attempts` will eventually
refuse the update — a second run against un-reset items silently proves nothing.

```sql
UPDATE site_work_items
   SET status='detected', error='', attempt_count=0, result=NULL, max_attempts=9
 WHERE item_type='scratch_cand2_probe';
```

## 3. Publish

**Two silent traps, one discipline.** Both known-bad forms exit without publishing:

| form | what happens |
|---|---|
| `printf … \| kubectl run -i --rm … -- kcat -P …` | exits **0**, publishes **nothing** |
| `kubectl run … --image=edenhill/kcat … -- sh -c '…'` | the image's **entrypoint is kcat**, so `sh -c …` arrives as kcat *arguments* → usage text, `-b <broker,..> missing` |

The working form puts the payload in the **container command**, adds `--command` to
replace the entrypoint, and makes the container print its own success:

```bash
CORR=$(uuidgen); ORCH=$(uuidgen)
PAYLOAD='{"action":"orchestrate","config":{"agent_type":"scratch-cand2-probe"},"input_data":{"item_a":"d418240c-f88f-480a-85c8-b328c901b7f5","item_b":"b0b0b0b0-1111-4222-8333-444444444444","item_c":"1b001fec-8e4e-4e4d-b6d3-2eb17d9e4c4c"}}'
echo "CORR=$CORR ORCH=$ORCH"
kubectl -n kafka run "kcat-cand2-$(date +%s)" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach --quiet --command -- \
  sh -c "printf '%s' '$PAYLOAD' | kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORR -H request_id=$(uuidgen) -H message_id=$(uuidgen) \
    -H orchestration_id=$ORCH -H orchestration_name=cand2-probe \
    -H step_name=start -H client_id=demo_client -H message_type=request \
    -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"
```

**`PUBLISH_OK` must appear.** A publish with no positive confirmation is not evidence of
a publish. Save `CORR` and `ORCH`.

## 4. Confirm it is actually on the topic, and at which offset

```bash
kubectl -n kafka run "kcat-read-$(date +%s)" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach --quiet --command -- \
  sh -c "kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests -p 0 -o <last-committed-offset> -c 8 -e -f '%o %T %h\n'"
```

Find your `correlation_id` in the output; the leading number is your offset. Compare it
with the consumer's `CURRENT-OFFSET` (§5) to see how many messages are ahead of you.

## 5. Is the lane stuck, or just busy? — the discriminator §5 of the handoff lacked

```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- bash -c \
 '/opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group'
```

`CURRENT-OFFSET` is the **next** offset to consume, so the message being processed right
now is `CURRENT-OFFSET` itself. Sample twice, 30 s apart — one reading cannot tell
"stalled" from "busy".

**A frozen `CURRENT-OFFSET` is NOT sufficient evidence of a stall.** Read the head
message's `orchestration_id` header (§4) and ask whether its orchestration is alive:

```sql
SELECT orchestration_id, status, current_step, created_at, updated_at, now()-updated_at AS idle
FROM orchestration_states WHERE orchestration_id = '<head message orchestration_id>';
```

- `EXECUTING_STEP` with a small `idle` → **head-of-line blocking, not a stall.** The lane
  is one in-order partition; a 16-seat `council-gate` run at the head holds every message
  behind it for as long as it takes. Wait; do not re-fire.
- no row, or `idle` growing without bound → then you have a genuine stall.

Measured 2026-07-26 21:52Z: offset frozen at 105272 for 5+ min, pod alive with 0 restarts
— and 105272's orchestration `f5c5d809-…` was `EXECUTING_STEP` at
`review_improvement_guardian`, updated 33 s earlier. Busy, not stuck.

## 6. Assert

```sql
SELECT summary, status, attempt_count, COALESCE(NULLIF(error,''),'<<BLANK>>') AS error
FROM site_work_items WHERE item_type='scratch_cand2_probe' ORDER BY summary;
```

To distinguish the **prefix branch** from a plain copy — the work item alone **cannot**,
because a prefix-if-absent branch and no branch produce identical output:

```sql
SELECT collected_data->'__step_error'->>'failed_step',
       collected_data->'__step_error'->>'message'
FROM orchestration_states WHERE orchestration_id = '<the ORCH you published>';
```

If `message` already starts with `step `, the prefix branch was **skipped** (correctly).

## 7. Pod-grep the deployment — with both controls in one command

A grep for a literal your own change introduced is unfalsifiable on its own.

```bash
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c '
  strings /app/agent-chassis | grep -c "no error_message literal"        # created by cand2 -> 1
  strings /app/agent-chassis | grep -c "build is short of its plan"      # positive control -> 1
  strings /app/agent-chassis | grep -c "candidate two placeholder xyzzy" # negative control -> 0'
```

The command exits **1** — that is the negative control's `grep -c` finding nothing, which
is the intended result, not a failure.

## 8. Cleanup — leave nothing behind

```sql
DELETE FROM site_work_items WHERE item_type='scratch_cand2_probe';
DELETE FROM agent_definitions WHERE type='scratch-cand2-probe';
```
