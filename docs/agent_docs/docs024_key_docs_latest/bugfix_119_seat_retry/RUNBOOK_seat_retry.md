# RUNBOOK — `bugs_open/119` seat retry

Every command here was needed to get something right. Gotchas are attached to the
command, not kept separately.

## R1 — How often is a council round thrown away by a parse failure?

The headline figure for this bug. Note `body` is **text**, not jsonb, so it must be cast.

```sql
WITH r AS (SELECT created_at, body::jsonb AS b FROM diagnosis_artifacts WHERE kind='council_report')
SELECT date_trunc('week', created_at)::date AS wk,
       count(*) AS rounds,
       count(*) FILTER (WHERE jsonb_typeof(b->'unreadable')='array') AS with_unreadable,
       count(*) FILTER (WHERE b->>'decided_by' LIKE 'unreadable reviewer%') AS voided
FROM r GROUP BY 1 ORDER BY 1;
```

**Gotcha 1 — do NOT use `jsonb_array_length(b->'unreadable')`.** It errors with
`cannot get array length of a scalar` partway through: the key is an **array** on 38
rows, **null** on 264 and **absent** on 122. Test with `jsonb_typeof(...)='array'`.

**Gotcha 2 — `with_unreadable` and `voided` are different questions.** A round can carry
an unreadable seat and still be decided by a real gating objection. Only
`decided_by LIKE 'unreadable reviewer%'` means the round was spent on the parse failure.

## R2 — Which seats, and is it one bad prompt or the whole roster?

```sql
WITH r AS (SELECT created_at, body::jsonb AS b FROM diagnosis_artifacts WHERE kind='council_report')
SELECT jsonb_array_elements_text(b->'unreadable') AS seat, count(*),
       min(created_at)::date AS first, max(created_at)::date AS last
FROM r WHERE jsonb_typeof(b->'unreadable')='array' GROUP BY 1 ORDER BY 2 DESC;
```

Answer on 2026-08-01: **seven** distinct seats, so roster-wide, not one prompt.

## R3 — Was an unreadable seat TRUNCATED or MALFORMED? (the discriminating query)

This is the one that reframed the bug. The council report names the step as
`review_x.result`; the stored value lives at `collected_data->'review_x'`.

```sql
SELECT jsonb_pretty(collected_data->'review_editquality')
FROM orchestration_states WHERE orchestration_id='<oid from R4>';
--  {"type":"text","result":"","__truncated":true,"__truncated_output_tokens":8000}
```

**Gotcha — a truncation with an EMPTY partial looks like nothing at all.** `result` is
`""`, so `length(...)` is 0 and it is tempting to read it as "the seat never ran". It ran;
the model spent the whole budget before emitting visible text. `__truncated` is the tell.

**Gotcha — most of the history is already gone.** 36 of 39 unreadable instances had no
`orchestration_states` row within four days. Do not plan an investigation around
retention; capture the specimen the day you find it.

## R4 — Find the orchestrations behind the voided rounds

```sql
SELECT created_at::timestamp(0), correlation_id, orchestration_id, left(body::jsonb->>'decided_by',50)
FROM diagnosis_artifacts WHERE kind='council_report' AND body::jsonb->>'decided_by' LIKE 'unreadable%'
ORDER BY created_at DESC LIMIT 12;
```

Reading it: repeated rows with the **same correlation** are one submission burning
several rounds. `c5219a69` appears three times in ten minutes — that is the cost shape,
not three separate incidents.

## R5 — Which steps declare which output key? (the blast radius)

```sql
WITH steps AS (
  SELECT ad.type, s.key AS step_name, s.value AS step
  FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(key,value)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.value->>'action'='execute_llm_prompt'
)
SELECT step->'config'->>'output_format' AS output_format_was_dead,
       step->'config'->>'output_type'   AS output_type_was_live,
       count(*) AS steps, count(DISTINCT type) AS agents
FROM steps GROUP BY 1,2 ORDER BY 3 DESC;
```

**Gotcha — the `is_active AND NOT is_snapshot AND deleted_at IS NULL` triple is not
optional.** Without it the snapshot rows inflate every count and you will quote a blast
radius several times too large.

To NAME the agents (the 2026-07-29 owner ruling requires telling consumers, not just
counting them), swap the select for
`string_agg(DISTINCT type, ', ' ORDER BY type)` with `output_format='json'`.

## R6 — Did a JSON-declaring step actually get JSON? (the denominator)

```sql
WITH json_fields AS (
  SELECT DISTINCT s.value->>'output_field' AS f
  FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(key,value)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.value->>'action'='execute_llm_prompt' AND s.value->'config'->>'output_format'='json'
    AND s.value->>'output_field' IS NOT NULL
), outs AS (
  SELECT os.orchestration_id, e.key AS field, e.value AS val
  FROM orchestration_states os, LATERAL jsonb_each(os.collected_data) e(key,value)
  WHERE jsonb_typeof(e.value)='object'
)
SELECT COALESCE(val->>'type','(no type key)') AS result_type, count(*)
FROM outs JOIN json_fields ON json_fields.f = outs.field GROUP BY 1 ORDER BY 2 DESC;
```

**Run this BEFORE quoting any failure count.** On 2026-08-01 it returns 782 json / 2 text,
which is what turns "2 failures" from alarming into 0.25%.

## R7 — Prove the fix by mutation, not by a green run

```bash
cp platform/orchestration/actions/ai_actions.go "$SCRATCH/ai_actions.go.bak"
# ...remove the output_format fallback, or the vocabulary gate...
go test ./platform/orchestration/actions/ -run 'TestGetOutputType|TestSeatConfigNow'   # expect FAIL
cp "$SCRATCH/ai_actions.go.bak" platform/orchestration/actions/ai_actions.go
go test ./platform/orchestration/actions/ -run 'TestGetOutputType|TestSeatConfigNow'   # expect ok
```

**Gotcha — restore from the backup immediately.** This is a shared working tree; another
session reading a mutated `ai_actions.go` would see a defect that does not exist.

## R8 — A green build in this tree is not a green HEAD

```bash
rm -rf "$SCRATCH/headtest" && mkdir -p "$SCRATCH/headtest"
git archive HEAD | tar -x -C "$SCRATCH/headtest"
cp platform/orchestration/actions/ai_actions.go "$SCRATCH/headtest/platform/orchestration/actions/"
cp platform/orchestration/actions/ai_output_contract_test.go "$SCRATCH/headtest/platform/orchestration/actions/"
cd "$SCRATCH/headtest" && go build ./platform/... ./internal/...
```

**Gotcha — build `./platform/... ./internal/...`, not `./...`.** `./...` links every
`cmd/` binary and fills `/tmp`; the resulting `no space left on device` reads exactly like
a build failure and is not one. Delete `headtest` afterwards — it is ~GB.

## R9 — Watch the council run (find it by payload, never by the printed id)

```sql
SELECT orchestration_id, status, current_step, updated_at::timestamp(0)
FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

Read the verdict **keyed on your own correlation** — reading the latest note is how a
thread commits someone else's approval as its own:

```sql
SELECT body FROM doc_notes WHERE categories ? 'council-gate'
  AND body LIKE '%<SUBMISSION_CORR>%' ORDER BY created_at DESC LIMIT 1;
```

## R10 — Post-roll verification (RUN 2026-08-02; passed on v1.0.1228)

Loop over **every** replica, never `-o jsonpath='{.items[0]...}'` — one pod of N is not the
deployment (`logs-deploy-reads-one-pod-of-n`).

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name} {end}'); do
  echo "== $POD"
  kubectl -n ai-persona-system exec "$POD" -- sh -c '
    B=/app/agent-chassis
    echo "pipeline ctl (pre-existing, >=1): $(strings $B | grep -c "Ensure valid JSON syntax")"
    echo "re-ask RECOVERED        (>=1): $(strings $B | grep -c "bugs_open/119 re-ask RECOVERED")"
    echo "RETRY marker            (>=1): $(strings $B | grep -c "RETRY (bugs_open/119")"
    echo "Keep every citation     (>=1): $(strings $B | grep -c "Keep every citation")"
    echo "spelling ctl (near-miss,  =0): $(strings $B | grep -c "bugs_open/119 re-ask RECOVEREDX")"
  '
done
```

> **CORRECTED 2026-08-02 — the negative control this section used to prescribe could NEVER
> have worked.** It grepped for `adds format-specific instructions based on output_type`, a
> line the change deleted — but that line is a **Go comment**, and comments are not
> compiled. It returns 0 against every binary ever built, before or after the roll, and
> reads as a pass. Worse, this change removed **no string literal at all** (check with
> `git diff <first>^..<last> -- <file> | grep '^-'` — all comments and code structure), so a
> removal-based negative control was not available here in the first place.

**So pick controls that can actually fail, and know which question each answers:**

| control | answers |
|---|---|
| a pre-existing string (`Ensure valid JSON syntax`) | does `strings`+grep work on this binary at all |
| a string added by the **LAST** commit of the series (`Keep every citation`) | is the image at least as new as the whole series — the image-age question a removed string would have answered |
| a deliberate near-miss of your own marker (`…RECOVEREDX`) | does grep return 0 when absent, i.e. is a `1` real |

A positive control proves the pipeline and **never** the spelling; a mis-cased grep reads as
"not shipped" (`grep -ic`).

Then the behavioural check, which outranks the grep — **with its denominator**, or a `0` is
unreadable:

```sql
SELECT count(*), max(created_at) FROM llm_call_log WHERE error_message LIKE 'RETRY (bugs_open/119%';
-- ALWAYS pair it with: how many in-scope calls even happened?
SELECT count(*) FROM llm_call_log WHERE created_at > '<roll ts>';
```

**Gotcha — 0 retries out of 4 total calls is not evidence of anything.** On 2026-08-02 the
fleet ran 4 LLM calls in the 32 min after the roll, none of them an `execute_llm_prompt`
step with a JSON declaration. Report that as `[UNMEASURED — 0 denominator]`, not as a pass.

## R11 — Did a JSON-declaring step receive the JSON instructions? (edit A's observable)

`prompt_rendered` is assigned **after** `appendOutputInstructions` (ai_actions.go:341 → 473),
so the logged prompt does show which block was appended.

```sql
WITH json_steps AS (
  SELECT DISTINCT ad.type AS agent_type, s.key AS step_name
  FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(key,value)
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.value->>'action'='execute_llm_prompt' AND s.value->'config'->>'output_format'='json'
)
SELECT CASE WHEN l.created_at > '<roll ts>' THEN 'AFTER' ELSE 'BEFORE' END AS era, count(*) AS calls,
       count(*) FILTER (WHERE l.prompt_rendered LIKE '%CRITICAL OUTPUT FORMAT - JSON:%') AS got_json_block,
       count(*) FILTER (WHERE l.prompt_rendered LIKE '%CRITICAL OUTPUT FORMAT:%')        AS got_default_block
FROM llm_call_log l JOIN json_steps j USING (agent_type, step_name) GROUP BY 1;
```

Baseline before the roll: **9,061 of 9,063 got the DEFAULT block** — 25 agents, four months.

**Gotcha — do NOT detect on a sentence like `Ensure valid JSON syntax`; it matches your own
submission.** A council submission's rationale is rendered *into* every seat's prompt, so a
phrase you quote while explaining the bug becomes a false positive in `prompt_rendered`.
That first cut returned 31 of 9,063, and all 31 were my own two council rounds (exactly 2
per seat, all on the submission date, `prompt_template` matching 0 times). Detect on the
block **header** (`CRITICAL OUTPUT FORMAT - JSON:`) — text no rationale would quote in
passing — and confirm with `prompt_template`, which the appended block never touches.

**Gotcha — `CRITICAL OUTPUT FORMAT:` is not a prefix-match trap, but check it.** The JSON
block reads `CRITICAL OUTPUT FORMAT - JSON:`, so the two headers are distinguishable; a
`LIKE '%CRITICAL OUTPUT FORMAT:%'` does not match the JSON one.

**Gotcha — the `(agent_type, step_name)` join silently drops pre-07-26 history.** Council
seats logged under `agent_type='generic'` until 2026-07-26 and `'council-gate'` after, so
1,798 older seat calls never join to any `agent_definitions.type`. The denominator is
therefore conservative and skewed recent — fine for "was the block appended", wrong for any
absolute historical volume.
