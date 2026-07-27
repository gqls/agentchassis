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
```

A Go *comment* is not in the binary, and a typed constant may not be either —
grep a string literal the code actually formats.

## 5. Prove the reserve reached the wire (not just that the binary shipped)

The point of the fix is a bigger `maxOutputTokens`. `llm_call_log` records it:

```sql
SELECT created_at, model, provider,
       sent_max_tokens, usage_output_tokens, error
FROM llm_call_log
WHERE provider = 'gemini'
ORDER BY created_at DESC LIMIT 20;
```

`sent_max_tokens` is now what actually went on the wire (caller's budget +
reserve), and the caller's own ask is written back separately as
`__sent_visible_budget_tokens`. **Consequence to know:** the
"`output_tokens == max_tokens` means the completion was CUT" rule goes quiet for
a thinking model, because visible output is being compared against a total that
includes the reserve. That rule was always a proxy for the finish reason; the
client now returns a typed `*TruncatedError` on `finishReason=MAX_TOKENS`
directly, which is authoritative. **Check the errors, not the arithmetic.**

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
--    write cannot be silently overwritten. Several threads touch this row.
--    Substitute the updated_at you read in §0. NOTE the nested path — the
--    shorter one you might guess at silently no-ops (see the correction in §0).
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,ai_service}',
      '{"provider":"gemini","model":"gemini-pro-latest","api_key_env_var":"GEMINI_API_KEY"}'::jsonb),
    updated_at = now()
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND updated_at = '<the updated_at you read>';
-- expect UPDATE 1. UPDATE 0 means someone else wrote the row — re-read, do not retry blind.
-- Then RE-READ with §0's query: jsonb_set on a path whose parent does not exist
-- is a SILENT NO-OP that still reports UPDATE 1. The row count proves the guard
-- held, not that the value landed.

-- 3. confirm the Voice & Style block survived (it is separate from the provider
--    and must NOT be lost in a provider change)
SELECT default_config->'workflow'->'steps'->'generate_content'->'config'->>'prompt_template' LIKE '%Voice & Style%'
FROM agent_definitions WHERE type = 'page-content-writer' AND is_active;
```

Then rebuild **one** page and read the copy. Rollback is the backup table, or
just re-point provider/model to `anthropic` / `claude-sonnet-4-6`.

## 8. Run the tests

```bash
go test ./platform/aiservice/ -run Gemini -v
```

`TestEveryProviderDecodesItsStopSignal` (in `stop_signal_test.go`) is the
structural CI guard: any type in `platform/aiservice` with a `GenerateText`
method must reference `TruncatedError` in its own file.
