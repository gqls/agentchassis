# RUNBOOK — Gemini content provider

*Every command that had to be got right, with its gotcha attached. When one
changes, change it HERE.*

---

## 0. Where the two providers are configured

They are different mechanisms and conflating them cost the first attempt time.

```bash
# content-creator-agent — a k8s service, provider in a configmap (git-tracked)
grep -A4 'ai_service:' deployments/kustomize/services/content-creator-agent/overlays/production/uk_001/configs/configmap-content-creator.yaml
```

```sql
-- page-content-writer — a chassis agent definition, provider in the DB.
-- The step is NESTED INSIDE the loop's sub_workflow. Verified 2026-07-27.
SELECT v->'config'->'ai_service'            AS step_ai_service,
       (v->'config'->>'max_tokens')         AS step_max_tokens,
       length(v->'config'->>'prompt_template') AS tmpl_chars,
       a.updated_at
FROM agent_definitions a,
     jsonb_each(a.default_config->'workflow'->'steps'->'process_sections_loop'
                ->'config'->'sub_workflow'->'steps') AS e(k,v)
WHERE a.type = 'page-content-writer'
  AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND e.k = 'generate_content';
```

> **CORRECTED 2026-07-27.** This section first gave the path as
> `workflow → steps → generate_content → config → ai_service`. **That returns all
> NULLs** — `generate_content` is not a top-level step. The real path is
> `workflow → steps → process_sections_loop → config → sub_workflow → steps →
> generate_content`. Caught by running it: four NULL columns and no error, because
> a wrong `->` path in Postgres yields NULL rather than failing. **A jsonb path
> that returns NULL has not told you the value is absent — it may have told you
> the path is wrong.** Found the real one by walking down from
> `jsonb_object_keys`, then `jsonb_each` over `steps` (it is an OBJECT keyed by
> step name, not an array — `jsonb_array_elements` errors with "cannot extract
> elements from an object").

**Gotchas.**
- `default_config->'workflow'->'steps'` is an **object**, not an array. Use
  `jsonb_each`, not `jsonb_array_elements`.
- The writer's own budget is **`max_tokens: 8000`** on that step — not one of
  content-creator's 100/1200/3000/6000 tiers. That matters: 8000 is roomy enough
  that the 07-24 starvation may never have bitten the writer at all (see NOTES).
- Root `ai_service` is shadowed per-key by the step block
  (`bugs_closed/009`), so patching the root alone changes nothing.
- The prompt template on that step is **12,570 chars** and growing — it was
  described as 7.8K in another workstream's notes. Re-measure rather than
  quoting; the reserve is sized off prompt complexity.

## 1. Probe the live key — do this BEFORE choosing a model

```bash
# list what this key can reach (no model arg)
scripts/gemini-probe.sh --from-pod

# full probe: tier table + which thinking knob is accepted
scripts/gemini-probe.sh --from-pod gemini-pro-latest
scripts/gemini-probe.sh --from-pod gemini-flash-lite-latest
```

Or with the key in hand: `GEMINI_API_KEY=... scripts/gemini-probe.sh <model>`.

**Gotchas.**
- The key exists **only in the cluster** (`personae-default-secrets` →
  `GEMINI_API_KEY`, wired in `patch-deployment.yaml` as `optional: true`). There
  is no local copy, so the probe needs `kubectl` auth. It never prints the key.
- A model appearing in the `models` listing is **not** proof this key can call
  it. Google closes pinned generations to newly-issued keys with a 404
  ("no longer available to new users"), and the listing does not reflect that.
  The tier probe is the only proof.
- Read the tier table for `CHARS=0` (or absurdly small) with `THINK_TOK>0` —
  that is thinking eating the ceiling, the 2026-07-24 failure. **The reserve to
  configure is the largest `THINK_TOK` observed, with headroom.**
- Use a prompt with something to think about. An empty-ish prompt under-reports
  thinking and will size the reserve too small.

## 2. Config keys the Gemini client accepts

In the `ai_service` block, either the configmap or the agent-definition step:

```yaml
ai_service:
  provider: "gemini"
  model: "gemini-pro-latest"        # REQUIRED, no default (see PLAN D4)
  api_key_env_var: "GEMINI_API_KEY"
  thinking_reserve_tokens: 8192     # optional, default 8192
  thinking_level: "low"             # optional, Gemini 3.x  ) at most ONE
  thinking_budget_tokens: 512       # optional, Gemini 2.5  ) of these
  embedding_model: "text-embedding-004"  # optional
```

**Gotchas.**
- `thinking_level` and `thinking_budget_tokens` are the same control on two
  generations and are **mutually exclusive** — setting both is refused at
  construction (it would be a 400 on every call).
- Set **neither** unless the probe says the model accepts it. The reserve carries
  the default case with no knob.
- An unknown key in an `ai_service` block is **silently ignored** — a dead key
  looks exactly like a live one. After any config change, prove the value reached
  the model rather than assuming (§5).
- `model` has no default on purpose. An absent model is now a construction error.

## 3. Ship the Go change (it is inert until you do)

```bash
# commit first — build-* builds from committed HEAD, not the working tree
git add platform/aiservice/gemini.go platform/aiservice/gemini_thinking_test.go scripts/gemini-probe.sh
git commit platform/aiservice/gemini.go platform/aiservice/gemini_thinking_test.go scripts/gemini-probe.sh -m "..."

# bump IMAGE_TAG (makefile ~line 16) — a same-tag rebuild ships the node's stale binary
make build-agent-chassis        && make push-agent-chassis        && make deploy-agent-chassis
make build-content-creator-agent && make push-content-creator-agent && make deploy-content-creator-agent
```

**Both** images are needed: `page-content-writer` runs inside the chassis,
`content-creator-agent` is its own service. Flipping the writer's DB config
against a chassis that predates the fix reproduces 07-24 exactly.

No orchestration dispatch within ~300s of a chassis pod restart — the spawn is
silently dropped.

## 4. Verify the deploy against the running pod, never git and never the tag

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system "$POD" -- sh -c \
  'strings /app/agent-chassis | grep -c "thinking consumed the entire output ceiling"'
```

**Gotcha — make the marker discriminating.** That string is *created* by this
change and appears nowhere else, so a non-zero count proves the new binary.
Pair it with a **negative control** that must be 0 — the string the old client
could not produce:

```bash
kubectl exec -n ai-persona-system "$POD" -- sh -c \
  'strings /app/agent-chassis | grep -c "thinking_reserve_tokens"'   # expect >0

# For a build AFTER 2026-07-27 14:xx, also grep the 110 candidate-1 rename —
# it is the only way to tell a pre-110 binary from a post-110 one:
kubectl exec -n ai-persona-system "$POD" -- sh -c \
  'strings /app/agent-chassis | grep -c "__sent_wire_max_output_tokens"'      # expect >0 post-110
kubectl exec -n ai-persona-system "$POD" -- sh -c \
  'strings /app/agent-chassis | grep -c "__sent_visible_budget_tokens"'       # expect 0 post-110
```

A Go *comment* is not in the binary, and a typed constant may not be either —
grep a string literal the code actually formats.

## 5. Prove the reserve reached the wire (not just that the binary shipped)

The point of the fix is a bigger `maxOutputTokens`. **`llm_call_log` does not
record that**, and the column names below are not what I first wrote here.

```sql
-- The real column names are max_tokens / output_tokens (\d llm_call_log).
SELECT created_at, model, provider, max_tokens, output_tokens, success, error_message
FROM llm_call_log
WHERE provider = 'gemini'
ORDER BY created_at DESC LIMIT 20;
```

> **CORRECTED 2026-07-27 — twice, and the second correction is `bugs_open/110`.**
> This section originally named columns `sent_max_tokens` and
> `usage_output_tokens`. **Neither exists** — the table has `max_tokens` and
> `output_tokens`. It also said the caller's ask "is written back separately as
> `__sent_visible_budget_tokens`", implying that was queryable. **It was not
> persisted at all**: that key, and `__usage_thinking_tokens`,
> `__usage_total_tokens` and `__sent_thinking_reserve_tokens`, have **no reader
> outside `platform/aiservice/`** and no column to land in. Filed as `110`, which
> also carries the fix for the worse half: `max_tokens` was being fed the
> reserve-inflated total, giving one column two provider-dependent meanings.

**Current contract, after `110` candidate 1** (in code, **inert until the next
chassis roll** — v1.0.1173 still logs the inflated total):

- `llm_call_log.max_tokens` = the caller's **visible-text budget**, the same
  meaning as for `anthropic`/`ollama`. For `page-content-writer` that is **8000**,
  not 16192. If you see 16192 on a Gemini row, that row predates the roll.
- The wire ceiling is in `__sent_wire_max_output_tokens` and thinking is in
  `__usage_thinking_tokens` — **both in-process only, persisted nowhere** until
  `110` candidate 2's migration.
- **Check the errors, not the arithmetic.** For a thinking model
  `output_tokens == max_tokens` cannot express truncation, because visible output
  may legitimately exceed the caller's budget (the API ceiling is the inflated
  total). A real Gemini cut arrives as `success = false` with an `error_message`
  naming thinking. That is the authoritative signal.

## 6. Flip content-creator (P5)

Config-only and live immediately — but only meaningful once §3/§4 are done.

```bash
# git-tracked change, then let kustomize apply it; or patch live and re-sync git
# after (the 07-24 pattern) — either way the tracked file must end up matching.
grep -A6 'ai_service:' deployments/kustomize/services/content-creator-agent/overlays/production/uk_001/configs/configmap-content-creator.yaml
```

Then generate one real post and read it. Do not judge on a status:
`complete` is not proof the work happened.

## 7. Flip page-content-writer (P6)

```sql
-- 1. back the row up FIRST (the 07-24 pattern: bak_agent_definitions_pcw_20260724)
CREATE TABLE bak_agent_definitions_pcw_20260727 AS
SELECT * FROM agent_definitions WHERE type = 'page-content-writer';

-- 2. patch the STEP block, guarded on updated_at so a concurrent session's
--    write cannot be silently overwritten. Several threads touch this row (it
--    was written at 13:44:56 on 2026-07-27 by the architecture-review re-seed).
--    Substitute the updated_at you read in §0. NOTE the nested path — the
--    shorter one you might guess at silently no-ops (see the correction in §0).
--
--    MERGE with `||` onto the existing ai_service object. Do NOT replace it.
UPDATE agent_definitions a
SET default_config = jsonb_set(
      a.default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,ai_service}',
      (a.default_config->'workflow'->'steps'->'process_sections_loop'->'config'
        ->'sub_workflow'->'steps'->'generate_content'->'config'->'ai_service')
      || '{"provider":"gemini","model":"gemini-pro-latest","api_key_env_var":"GEMINI_API_KEY"}'::jsonb),
    updated_at = now()
WHERE a.type = 'page-content-writer'
  AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND a.updated_at = '<the updated_at you read>';
-- expect UPDATE 1. UPDATE 0 means someone else wrote the row — re-read, do not retry blind.


-- 3. RE-READ. Two independent reasons the row count is not proof:
--    (a) jsonb_set on a path whose parent does not exist is a SILENT NO-OP that
--        still reports UPDATE 1;
--    (b) a wholesale replace would have dropped sibling keys (see below).
--    Also confirms the Voice & Style block survived — it is separate from the
--    provider and must NOT be lost in a provider change.
SELECT v->'config'->'ai_service'                       AS ai_service_now,
       (v->'config'->>'prompt_template') LIKE '%Voice & Style%' AS style_block_intact,
       length(v->'config'->>'prompt_template')         AS tmpl_chars
FROM agent_definitions a,
     jsonb_each(a.default_config->'workflow'->'steps'->'process_sections_loop'
                ->'config'->'sub_workflow'->'steps') AS e(k,v)
WHERE a.type='page-content-writer' AND a.is_active
  AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND e.k='generate_content';
-- ai_service_now MUST still contain max_tokens: 8000.
```

> **CORRECTED 2026-07-27 — this section would have quietly halved the writer's
> budget.** The first version replaced the whole `ai_service` object with
> `{"provider","model","api_key_env_var"}`. **`max_tokens: 8000` lives inside that
> same block**, so the replace would have dropped it, and `NewGeminiClient` would
> have fallen back to the client's 2048 default — a 4x cut to the writer's output
> budget, invisible in the diff, presenting later as truncated page sections.
> Caught by reading the row before writing it: `step_max_tokens` came back NULL
> from a query looking at the step config, because the key is one level in, under
> `ai_service`. **`jsonb_set` with a literal object is a REPLACE, not a merge** —
> use `||` on the existing object whenever the block you are patching has siblings
> you did not enumerate.

Then rebuild **one** page and read the copy. Rollback is the backup table, or
just re-point provider/model to `anthropic` / `claude-sonnet-4-6`.

## 8. Run the tests

```bash
go test ./platform/aiservice/ -run Gemini -v
```

`TestEveryProviderDecodesItsStopSignal` (in `stop_signal_test.go`) is the
structural CI guard: any type in `platform/aiservice` with a `GenerateText`
method must reference `TruncatedError` in its own file.
