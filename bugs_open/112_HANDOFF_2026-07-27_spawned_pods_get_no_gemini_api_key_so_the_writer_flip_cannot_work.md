# 112 — Spawned agent pods are never given `GEMINI_API_KEY`, so `page-content-writer` on Gemini fails at construction, and every page build with it

**Filed** 2026-07-27 by the `gemini_content_provider` workstream (triage sweep) ·
**Status** OPEN, unowned · **Severity** HIGH and **armed but not yet fired** —
the first build that will hit it is the scheduled `model-directory-publish` at
**~20:25 UTC 2026-07-27** · **Blocks** `bugs_open/107` P7 and the verification of
`bugs_open/110` candidate 1 · **Not** a Gemini client bug: `platform/aiservice/gemini.go`
is correct and live

---

## Symptom (predicted, precisely, and not yet observed)

`bugs_open/107` P6 flipped `page-content-writer`'s `generate_content` step to
Gemini in the live DB at 2026-07-27 14:34 UTC. That step will now fail at client
construction with:

```
API key environment variable 'GEMINI_API_KEY' is not set or empty
```

`platform/aiservice/gemini.go:157-160`. The step has **no `error_step`** (verified
against the live `agent_definitions` row), so the failure propagates:
`generate_content` → `process_sections_loop` → the writer orchestration FAILS →
the page build fails. **Every section of every page.**

It has not fired yet only because no page build has reached the writer since the
flip. The one build that ran (15:46, `dartsonline/grip-styles`) died three steps
earlier on a missing section spec, and `SELECT count(*) FROM llm_call_log WHERE
provider='gemini'` is **0**.

## Root cause

**A spawned agent pod's environment is an explicit allow-list built in Go, not an
inheritance of the spawner's environment.** The main `agent-chassis` Deployment
gets its keys through `envFrom: secretRef: personae-default-secrets`, which
carries `GEMINI_API_KEY`. Spawned pods do not use that `secretRef` at all.

`platform/orchestration/actions/spawn_actions.go:2440-2518` appends each secret
env var by name. It injects `ANTHROPIC_API_KEY` and `GROK_API_KEY` from
`personae-default-secrets`:

```go
{
    Name: "ANTHROPIC_API_KEY",
    ValueFrom: &corev1.EnvVarSource{
        SecretKeyRef: &corev1.SecretKeySelector{
            LocalObjectReference: corev1.LocalObjectReference{
                Name: "personae-default-secrets",
            },
            Key: "ANTHROPIC_API_KEY",
        },
    },
},
```

There is **no `GEMINI_API_KEY` block**. `grep -rn "GEMINI_API_KEY" --include=*.go`
returns **two hits in the entire tree**, both inside `platform/aiservice/gemini.go`
(a doc comment and an error string). Nothing anywhere provisions it into a pod.

The provider switch was only ever wired for `content-creator-agent`, which is a
**standalone Deployment** with its own patch
(`deployments/kustomize/services/content-creator-agent/overlays/production/uk_001/patch-deployment.yaml:25-29`
adds `GEMINI_API_KEY` explicitly). That is why P5 succeeded and why P6 cannot:
**the two agents reach their API keys by completely different routes**, and only
one of them was given Gemini.

## Evidence (all 2026-07-27, live)

| check | result |
|---|---|
| `page-content-writer` runs in a spawned pod | `pod=agent-page-content-writer-47953cd4-k8bmn` and 4 more, from `orchestration_states.collected_data->'__execution_context__'->'sender'->>'pod_name'` |
| `page-build-handler.spawn_content_writer` | `{"action": "spawn_agent", "config": {"agent_type": "page-content-writer"}}` |
| main chassis pod has the key | `agent-chassis-5994dc6d6c-pt8v9` → `PRESENT len=53` |
| a live **spawned** pod has it | `agent-build-dispatch-loop-b62d9c1a-pnnt9` (image `v1.0.1174`) → **ABSENT** |
| that spawned pod's env list | `ANTHROPIC_API_KEY`, `GROK_API_KEY`, `FIRECRAWL_API_KEY`, ... and **no `GEMINI_API_KEY`** |
| the secret does hold it | `personae-default-secrets` keys include `GEMINI_API_KEY` |
| writer's live provider | `{"model": "gemini-pro-latest", "provider": "gemini", "max_tokens": 8000, "api_key_env_var": "GEMINI_API_KEY"}` |
| Gemini rows logged so far | **0** |

Note the spawned pod already runs `v1.0.1174` — `bugs_open/066`'s fix is working,
so this is **not** a stale-image problem. The binary is right; the environment is not.

## Fix candidates, ordered by what closes the door

**1 — Add the `GEMINI_API_KEY` block to `spawn_actions.go` next to the other two.**
Ten lines, mirrors `ANTHROPIC_API_KEY` exactly, secret key already exists. Go, so
it needs a build and roll. This is the obvious fix and it is correct, but note it
leaves the class open: the next provider added will fail the same way.

**2 — Make the spawned pod's secret env derive from the same source as the
Deployment's** (`envFrom: secretRef: personae-default-secrets`) instead of a
hand-maintained list. This makes the bad state unrepresentable: a key added to the
secret reaches spawned pods without a code change. It widens what a spawned pod
can see, which is a deliberate security posture question — the existing code
comments show at least one key (`GITHUB_READ_TOKEN`) is *deliberately* scoped to
one agent type, so a blanket `envFrom` would need that carve-out kept.

**3 — Refuse the flip at config time.** Have the seeding/validation path reject an
`ai_service.api_key_env_var` that the target agent's pod will not receive. Largest
change; turns a runtime outage into a refused config. Worth considering because
this failure is invisible until a build runs, and builds are scheduled hours apart.

**Interim mitigation available with no code at all:** revert
`page-content-writer.generate_content` to `anthropic` /
`claude-sonnet-4-6` (DB config, live immediately). That un-arms the fleet-wide
build failure. `bugs_open/107`'s client fix stays live and correct; only the *flip*
is reverted. The workstream's own cost data (NOTES, 5x2 run comparison) independently
argues for that anyway: Gemini spends ~1,815 thinking tokens per section against
Claude's zero, roughly **10x the billable output tokens** per section, and the writer
runs per section across the whole estate.

## How to verify a fix

1. **The key reaches a spawned pod.** Not a pod-grep of the binary — an env check
   on a *spawned* pod, because the binary was never the problem:
   `kubectl -n ai-persona-system exec <a running agent-*-* pod> -- sh -c 'echo ${#GEMINI_API_KEY}'`
   must be non-zero. Use a spawned pod, never `agent-chassis-*`, which has always
   had the key and therefore proves nothing.
2. **The writer actually calls Gemini.** `SELECT provider, model, max_tokens,
   output_tokens, success FROM llm_call_log WHERE agent_type='page-content-writer'
   ORDER BY created_at DESC LIMIT 5` — `provider='gemini'` rows must appear. Until
   one does, nothing about the Gemini writer path is verified.
3. **Induce the failing branch.** Point a scratch step at a bogus
   `api_key_env_var` and confirm the construction error still names the variable.
   A green build proves the key arrived, not that the guard works.

## Pointers

`bugs_open/107` (the client fix this blocks the verification of — that fix is live
and correct) · `bugs_open/110` (candidate 1 is live but unverifiable until a Gemini
row exists) · `bugs_open/066` (spawned pods pin stale image tags — same pod class,
different mechanism; its fix is live and confirmed here) ·
`docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/`

Transferable pattern for `016b` §9: **"the same credential reaches two agents by
two different routes, and only one route was updated"** — a Deployment's `envFrom`
and a spawner's hand-built env list are two separate provisioning paths for the
same secret, and a change to the secret satisfies neither automatically. Testing
the standalone service tells you nothing about the spawned one.
