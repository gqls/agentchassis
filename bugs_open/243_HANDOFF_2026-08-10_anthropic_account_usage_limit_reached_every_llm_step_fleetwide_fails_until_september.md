# 243 — the Anthropic account hit its usage limit at 14:51:47Z today; EVERY LLM step fleet-wide now fails, and the API says access returns 2026-09-01

**Filed:** 2026-08-10 ~17:00 BST by the `bugfix_153_build_provenance` lane, which found it
by having its own council submission die on it. **Status: OPEN, unowned, LIVE RIGHT NOW.**
**Class:** account/billing state, not a code defect — the same family as `bugs_open/202`
(Gemini quota 429 blocking page builds), one provider along.
**Severity:** the fleet's entire LLM capability is down. This is an **owner action**, not a
debugging task — see §5.

> **THIS IS NOT A CODE BUG AND MUST NOT BE "FIXED" IN CODE.** The provider is asserting a
> billing state about our account in its own error body. Per CLAUDE.md's diagnosis norm, a
> `090` run was **not** performed and the reason is stated: the root cause is named by the
> upstream service, there is no mechanism to diagnose. What IS open is a decision (§5), which
> is the owner's.

## 1. The error, verbatim

```
step review_editquality failed: failed to execute action execute_llm_prompt:
AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 endpoint=<nil>:
API request failed with status 400: {"type":"error","error":{"type":"invalid_request_error",
"message":"You have reached your specified API usage limits. You will regain access on
2026-09-01 at 00:00 UTC."},"request_id":"req_011CduR3gLr5nj1nr21HVhun"}
```

Note **400**, not 429. Nothing in our retry/classifier machinery treats a 400 as transient,
which is correct here — retrying cannot help for three weeks — but it means the failure is
**terminal on first contact** and burns the step's attempts immediately.

## 2. The cutover, measured to the second `[MEASURED]`

`llm_call_log` gives an unusually clean boundary — two calls 2.7 seconds apart:

```sql
SELECT created_at, agent_type, success, left(error_message,70)
FROM llm_call_log WHERE created_at > now() - interval '5 hours' ORDER BY created_at DESC;
```

| time (UTC) | agent | success |
|---|---|---|
| 2026-08-10 14:51:45.067 | council-gate | **t** — the last successful LLM call in the fleet |
| 2026-08-10 14:51:47.814 | council-gate | **f** — the first refusal |

Since that boundary, **zero successes**:

```sql
SELECT success, count(*), count(DISTINCT agent_type)
FROM llm_call_log WHERE created_at > '2026-08-10 14:51:46+00' GROUP BY 1;
-- f | 5 | 3      (no 't' row at all)
```

Affected agent types so far: `council-gate` (3), `experience-planner` (1),
`tool-recreation-handler` (1). That list is small only because **fleet LLM volume has
collapsed** — 40 calls in the 14:00 hour, 3 in the 16:00 hour — not because the failure is
selective.

## 3. Blast radius — every LLM step we own `[MEASURED]`

```sql
SELECT s.v->'config'->'ai_service'->>'provider' AS provider,
       count(*) AS steps, count(DISTINCT type) AS agents
FROM agent_definitions, LATERAL jsonb_each(default_config->'workflow'->'steps') s(k,v)
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND s.v->'config' ? 'ai_service' GROUP BY 1;
-- anthropic | 127 | 55
-- (null)    |  11 | 10
```

**127 configured LLM steps across 55 live agents, and every single one is `anthropic`.** All
127 name the same credential, `ANTHROPIC_API_KEY` — there is no second key and no non-Anthropic
step anywhere in live config. The 11 null rows are steps whose `ai_service` block omits a
provider, not steps using a different one.

> **TRAP for anyone re-measuring this.** `provider` is nested inside `ai_service`, NOT a
> direct child of `config`. The obvious query —
> `WHERE v->'config'->>'provider' = 'anthropic'` — returns **0 rows and no error**, which
> reads exactly like "nothing uses Anthropic". I wrote that query first and briefly believed
> it. The correct path is `v->'config'->'ai_service'->>'provider'`. Landmine filed.

The council gate specifically: all **17** of its `review_*`/`gate_*` steps use
`{"model":"claude-sonnet-5","provider":"anthropic","max_tokens":16000|8000,
"api_key_env_var":"ANTHROPIC_API_KEY"}`. There is no seat that could still run.

## 4. What this breaks, and the part that is genuinely dangerous

- **Every page build, content write, classifier, planner and reviewer** — 55 agents.
- **The council gate is DOWN**, which means CLAUDE.md's standing review norm is currently
  unsatisfiable. Every thread that submits from now gets an orchestration that ends
  `current_step='complete_invalid'`.
- **THE FAILURE MODE LOOKS LIKE LATENCY, AND THAT IS THE TRAP.** CLAUDE.md tells submitters to
  budget ~30 minutes and explicitly says *"a missing orchestration row is almost always
  latency, not a dropped dispatch — do not retry on that evidence"*. That advice is correct in
  general and **actively misleading today**: the run is not slow, it is dead, and the thread
  that waits patiently will wait for ever. Three submissions have already died this way:

```sql
SELECT current_step, status, count(*), min(created_at), max(created_at)
FROM orchestration_states WHERE created_at > now() - interval '6 hours'
  AND collected_data->'input_data'->>'fix_correlation_id' IS NOT NULL GROUP BY 1,2;
-- complete_rejected | COMPLETED | 1 | 11:04 | 11:04
-- complete_approved | COMPLETED | 1 | 11:20 | 11:20
-- complete_invalid  | COMPLETED | 3 | 14:42 | 16:51
```

Correlations that died: `b67eb26a-14ef-45d7-b755-3e489fd57ef0` (14:42, `review_architecture`),
`7177fb02-51c5-4c2a-bb02-10aa27ae85ca` (16:50), `44fa6a98-acaa-46b5-9ada-f0c34ca5475d`
(16:51, this lane's). **Note the 14:42 one predates the measured cutover** — its
`review_architecture` step ran later in the workflow than the 14:51 boundary, so it is the
same cause, not a second one.

**How to tell, in one query** — do this before assuming your council run is merely queued:

```sql
SELECT current_step, status, collected_data->'__step_error'->>'message'
FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<YOUR_CORR>';
```
`complete_invalid` + a `usage limits` message = this bug, not your submission. Your `fix_plan`
artifact will have persisted normally, which is why the submission itself looks fine.

## 5. Fix candidates — ordered, but the first is the only real one

1. **Owner raises or resets the Anthropic account spend limit.** The message is
   *"your specified API usage limits"* — a self-imposed cap on our own account, not an
   Anthropic-side suspension, so it is very likely a console setting the owner can change in
   minutes. **This is the whole fix and everything below is contingency.** It needs the
   owner: no thread has console access.
2. **Add a second provider as a fallback and let the fleet fail over.** The platform already
   supports multiple providers (`bugs_closed/`'s gemini work, and `platform/aiservice` has
   per-provider clients), and `bugs_open/202` records the mirror-image situation with Gemini.
   **But note what §3 measured: 127/127 steps name one provider and one key.** There is no
   fallback configured anywhere, so this is a real build, not a config flip — and doing it
   under an outage, to live config, on a shared tree, is how mistakes ship. Propose it as
   follow-up work, do not improvise it today.
3. **Make the failure legible rather than silent** — the genuinely useful code-adjacent
   follow-up, and the only one this lane would recommend building. A 400 whose body contains
   `usage limits` is a distinct, non-retryable, fleet-scope condition and should be surfaced
   as such: a named error class, and ideally a work item or alert, so the next thread does not
   spend 30 minutes reading a dead run as a slow one (§4). Today the only way to learn this is
   to query `__step_error` by hand, knowing to.
4. **Do NOT** bump retries, widen the transient classifier, or "just resubmit". A 400 with a
   three-week horizon is correctly terminal; making it retryable would burn quota we do not
   have and turn one clean failure into a storm.

## 6. How to verify it is over

```sql
-- the fleet's own liveness signal: any successful LLM call at all
SELECT max(created_at) FROM llm_call_log WHERE success;
```
If that is newer than the moment the owner changed the setting, the fleet is back. A positive
control costs one council submission — but do not spend one until the query above has moved,
because a submission during the outage is a guaranteed `complete_invalid`.

**Do not close this on the calendar.** `2026-09-01` is the API's stated auto-restore date, so
this bug will appear to fix itself in three weeks whether or not anyone acts. If it closes
that way, candidate 1 was never done and the same cap will be hit again — record which of the
two actually happened.

## 7. Related

- `bugs_open/202` — the Gemini 429 equivalent, filed 2026-08-05, same shape one provider
  along: a quota state blocking page builds fleet-wide, with the decision belonging to the
  owner. Read together, the two make the case for candidate 2 far better than either alone:
  **we have now had a single-provider outage twice in five days.**
- CLAUDE.md § "Council review of platform changes" — the ~30-minute latency guidance whose
  interaction with this outage is described in §4.
