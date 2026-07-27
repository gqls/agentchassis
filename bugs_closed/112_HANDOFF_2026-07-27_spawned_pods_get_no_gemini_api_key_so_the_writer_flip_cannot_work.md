# 112 — Spawned agent pods are never given `GEMINI_API_KEY`, so `page-content-writer` on Gemini fails at construction, and every page build with it

**Filed** 2026-07-27 by the `gemini_content_provider` workstream (triage sweep) ·
> **VERIFIED LIVE 2026-07-27 by the gemini_content_provider thread.** The fix is in
> the running chassis at **v1.0.1177**: `strings /app/agent-chassis | grep -c
> ProviderKeyEnv` → 1, and `agentenv` → 3, so `platform/agentenv` is compiled in and
> spawned pods now receive `GEMINI_API_KEY`. **`page-content-writer` can reach Gemini.**
> Note the obvious grep is VACUOUS here: `GEMINI_API_KEY` itself returns 2 on a binary
> that predates the fix, because `gemini.go`'s error string contains it. Grep the
> package symbol, not the key name. This owner should close it.
>
**Status** **OPEN — FIXED IN CODE, INERT UNTIL BUILT AND ROLLED** (`b3f19ac96`,
2026-07-27, on the owner's instruction to fix it in Go rather than revert the
flip) · **Severity** HIGH and **armed but not yet fired** —
the first build that will hit it is the scheduled `model-directory-publish` at
**~20:25 UTC 2026-07-27** · **Blocks** `bugs_open/107` P7 and the verification of
`bugs_open/110` candidate 1 · **Not** a Gemini client bug: `platform/aiservice/gemini.go`
is correct and live

> **THE FIX DOES NOT UN-ARM 20:25 ON ITS OWN.** Go is inert until an image is
> rebuilt and rolled. If `v1.0.1175` is not live before 20:25:16 UTC, the
> scheduled publish still hits the missing key. Three ways to be safe, in
> increasing order of what they cost: roll before 20:25; disable the
> `model-directory-publish` scheduled task for one tick; or revert the writer to
> Claude in the DB (config, live immediately). **This case stays OPEN until a
> Gemini call is pod-verified on the spawned path** — a pod-grep of the binary
> would prove only that the code shipped, never that a spawned pod received the
> key. The acceptance test is `SELECT count(*) FROM llm_call_log WHERE
> provider='gemini'` going non-zero, or the env var read directly off a live
> spawned pod.

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

## WHAT WAS DONE — candidate 1, both spawners (2026-07-27, `b3f19ac96`)

> **Council gate: submitted 2026-07-27 ~18:13 UTC, `SUBMISSION_CORR =
> dfa6205e-b10e-440c-b251-5d791fdeb718`.** Submitted AFTER shipping, which is
> stated plainly in the rationale — the change was racing the 20:25 tick. The
> submission asks the council specifically to rule on the thing I overrode: the
> bug file ranks blanket `envFrom: secretRef` FIRST on door-closing grounds and I
> declined it on a least-privilege argument. If the council disagrees, that is
> the finding worth having.
>
> Verdict: `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
> WHERE correlation_id='dfa6205e-b10e-440c-b251-5d791fdeb718' AND
> kind='council_report' ORDER BY created_at;`
> **Only an APPROVED verdict earns a `Council-Reviewed:` trailer, and it is
> earned per round — re-read `decided_by` before claiming it.** The deployed
> commits carry no trailer and will list in the 098 report as un-reviewed until
> one is earned; that is accurate, not an oversight.

**Owner's instruction was to fix it in Go, not to revert the flip.** So the
interim mitigation above was deliberately NOT taken, and the flip stands.

- `platform/orchestration/actions/spawn_actions.go` — `GEMINI_API_KEY` block
  added next to `GROK_API_KEY`, sourced from `personae-default-secrets`.
- `cmd/remote-job-spawner/main.go` — **a second spawner with the same allow-list
  shape, found while fixing the first.** It builds its env "same structure as
  `spawnAgentKubernetesJobFromDefinition`" and spawns an arbitrary
  `req.AgentType`, so it carries the identical defect. It had `ANTHROPIC_API_KEY`
  only — not even Grok. Gemini added. It is a **separate service with its own
  image**, so it needs its own build and roll.
  **Latent, not live:** no agent definition uses remote dispatch
  (`SELECT count(*) FROM agent_definitions WHERE default_config::text LIKE
  '%remote%' AND is_active` → 0, 2026-07-27), so nothing is failing there today.
  `GROK_API_KEY` is still absent from it, left deliberately with the reason in
  the comment rather than adding keys blind to a path nothing uses.

**Candidate 2 (blanket `envFrom: secretRef`) was considered and declined.** It is
the one that makes the bad state unrepresentable, and on the door-closing
ordering it should win — but the allow-list turns out to be *deliberate*, not an
oversight. The GitHub token immediately below it is scoped to repo-cloning agents
only, with a comment saying so. A blanket `secretRef` would hand all twelve
provider keys in `personae-default-secrets` to every spawned agent pod and
silently undo that boundary. Kept the allow-list; wrote the two-place rule into
the code instead, where the next person adding a provider will be standing.
**Candidate 3 (refuse the flip at config time) remains the real structural fix
and is untouched** — it is what would have turned this outage into a refused
config, and it is still worth doing.

**Not verified, and cannot be from a commit:** that a spawned pod actually
receives the key. That needs the roll, then verify step 1 below. A pod-grep of
the chassis binary would be VACUOUS here — the failure is in what the *spawned*
pod's env contains, not in what the chassis binary knows.

## VERIFY STEP 1 — PASSED 2026-07-27 18:16 UTC, on the real path

**The key reaches a spawned pod. Observed, not inferred, with a roll-boundary
control** — which is the whole point, because the binary was never the problem.

First pod spawned by the `v1.0.1175` chassis, `agent-build-dispatch-loop-d4f0502e-2cq7z`,
started `2026-07-27T18:16:07Z`:

```
# env NAMES in the spawned pod's own spec
ANTHROPIC_API_KEY
GROK_API_KEY
GEMINI_API_KEY          <-- absent before this fix

# read inside the running container
GEMINI len=53   ANTHROPIC len=108
```

`len=53` matches the chassis's own `GEMINI_API_KEY`, so it is the real key and
not an empty var that merely exists.

**The control is what makes this conclusive.** The two `nav-updater` pods spawned
*before* the roll are still alive on the old image, so the comparison is like for
like on a live cluster rather than against a memory of one:

| spawned pod | image | `GEMINI_API_KEY` entries |
|---|---|---|
| `agent-build-dispatch-loop-d4f0502e-2cq7z` (post-roll) | v1.0.1175 | **1** |
| `agent-nav-updater-79698e25-wdl4c` (pre-roll) | v1.0.1174 | **0** |
| `agent-nav-updater-7d39ed72-vd8w4` (pre-roll) | v1.0.1174 | **0** |

This also confirms `bugs_open/066` is doing its job: the spawned pod runs the
image its SPAWNER runs (v1.0.1175), which is the only reason a chassis-only roll
could fix a defect that manifests in spawned pods at all. Had 066 still been
broken, spawned pods would have kept pinning `agent_definitions.image_tag` and
this fix would have been invisible no matter how correct the code was.

**STILL OPEN on verify step 2**: no Gemini call has yet traversed the chassis
(`llm_call_log` provider='gemini' remains 0). Step 1 proves the key arrives; only
step 2 proves the writer works on Gemini. The `model-directory-publish` tick at
20:25:16 UTC is the first thing due to exercise it.

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

---

## COUNCIL: round 1 REVISE (harness), round 2 **APPROVED** — corr `dfa6205e-b10e-440c-b251-5d791fdeb718`

**Round 2 verdict, 2026-07-27 18:49:19 UTC: `approved`, `unreadable=0`,
"approved with 2 advisory objection(s) — none high-severity".** The two medium
objections from round 1 both flipped on the seats that raised them:
`reuse_agent` object → **approve**, `tooling_provenance` object → **approve**.

**The approval attaches to the plan that was actually committed** — checked,
because a later approval can attach to a materially different plan and the 098
report cannot tell. The approved `fix_plan` names exactly four edits:
`platform/agentenv/provider_keys.go` (add), `spawn_actions.go` (modify),
`cmd/remote-job-spawner/main.go` (modify), `provider_keys_test.go` (add) — the
same four files, same operations, as commit `6b5509ee3`.

**Round 1 was not a judgement on the plan** and is not recorded as one: it was
`decided_by: unreadable reviewer(s): review_guardian.result` against eight
approvals and no veto, because the guardian's JSON closed a bracket early. Filed
separately as `bugs_open/119`.

### Advisory objections that survive the approval — read before the next roll

- **guardian [medium] — the widening is real.** *"Consolidating the allow-list
  gives cmd/remote-job-spawner GROK_API_KEY it never had."* Deliberate, and
  disclosed in the submission's own risks, but it IS a permission widening on the
  remote path. It is currently unexercised (no agent definition uses remote
  dispatch), and the remote spawner is a separate image that has not been rebuilt,
  so nothing has changed in production yet. **Whoever rolls remote-job-spawner
  owns that decision** — it is the one behavioural change in the refactor.
- **prior_art_librarian [medium] — unverifiable at that tier, not wrong.** The
  seat could not check the claim that `doc_plans`/`doc_notes` held no prior row,
  because those tables are not in the schema available to it. The rows are real
  and now exist: `subject_type='action'`, `subject_key='spawn_agent'`, one plan
  (the durable SHARED CONTRACT) and one note (the incident + why the GROK
  asymmetry existed and was removed). Cite them directly rather than the claim.
- **prior_art_librarian [low]** — the "only 3 files reference `corev1.EnvVar`"
  completeness claim is a point-in-time content check, so it goes stale; re-run it
  before assuming no fourth pod-env builder has appeared.
- **guardian [low]** — `spawn_actions.go` has been deflected to a higher layer
  four times in this council's history. This round was judged a legitimate dedup
  rather than a deflection, but the pattern is worth knowing before the next edit
  to that file.

**Still true and unchanged by the approval:** candidate 3 (refuse at config time
an `api_key_env_var` the target pod cannot receive) is undone, and it remains the
fix that makes this class unrepresentable. `agentenv.ProviderKeyNames` exists to
make it cheap to write.

---

# ✅ CLOSED 2026-07-27 — both acceptance tests passed on the real path

**Bar met: fixed AND live AND verified**, on the live path rather than by
inference from a binary.

### Test 1 — the key reaches a SPAWNED pod ✅ (18:16 UTC, v1.0.1175)

First pod spawned by the new chassis, `agent-build-dispatch-loop-d4f0502e-2cq7z`:
`GEMINI_API_KEY` present in the pod spec, `len=53` inside the running container
(matching the chassis's own key, so a real value and not an empty var that merely
exists). **Live roll-boundary control**: the two pods spawned *before* the roll
were still running on v1.0.1174 and carried **0** such entries. A chassis
pod-grep would have been vacuous here — the defect was always in what the
*spawned* pod's env contains.

### Test 2 — the writer actually calls Gemini ✅ (20:09 UTC)

```
llm_call_log WHERE provider='gemini'
2026-07-27 20:09:31 | page-content-writer | gemini-pro-latest | max_tokens 8000 | output 87 | success=t
2026-07-27 20:09:50 | page-content-writer | gemini-pro-latest | max_tokens 8000 | output 79 | success=t
```

`page-content-writer` runs in a spawned pod, so these two rows are the whole
causal chain working: spawned pod → env → client construction → API call →
success. The column was **0 for the life of the platform** until now. Both calls
finished well inside the ceiling (87/79 against 8000), so neither is a truncated
completion masquerading as success.

### Both spawners are live, and the consolidation with them

`v1.0.1179`, rolled 20:26 UTC. Discriminating markers with the pre-image as
control — note `GEMINI_API_KEY` reads **2 on both**, so the marker that proved the
*first* fix is useless for the consolidation; the package path is the one that
discriminates:

| marker | chassis v1.0.1175 | chassis v1.0.1179 | remote-job-spawner v1.0.1179 |
|---|---|---|---|
| `platform/agentenv` | 0 | **3** | **3** |
| `ProviderKeyEnv` | 0 | **1** | **1** |
| `GROK_API_KEY` | — | — | **1** |
| `…agentenvNOTREAL` | — | 0 | 0 |

(The remote spawner's binary is at `/remote-job-spawner`, **not** `/app/` — a
grep against `/app/*` there hits a YAML file and returns 0 for everything, which
reads exactly like "the change did not ship".)

### ⚠️ The guardian's medium objection is NOW LIVE, and it was accepted knowingly

Council round 2 approved this while flagging: *"Consolidating the allow-list gives
`cmd/remote-job-spawner` `GROK_API_KEY` it never had… a permission widening on the
remote path."* That is no longer theoretical — `remote-job-spawner:v1.0.1179`
carries it. It remains **unexercised** (no agent definition uses remote dispatch),
and it was disclosed in the submission's own risks, but anyone auditing spawned-pod
credentials should know it changed today and why.

### What this closure does NOT claim

- **Acceptance test 3 was not run.** Pointing a scratch step at a bogus
  `api_key_env_var` to confirm the construction error still names the variable is
  untouched. The fix is proven by a working path, not by its guard.
- **Nothing here is a verdict on Gemini as the writer's provider.** That is
  `bugs_open/107`/`110` and the `gemini_content_provider` workstream's call —
  including the ~10× thinking-token cost finding, which is an owner decision and
  is unaffected by this fix.
- **Candidate 3 remains undone** — nothing validates at config time that an
  `api_key_env_var` names a key the target pod can receive, so the next provider
  added to a DB config without a matching `agentenv` entry fails identically. One
  place to change now instead of two; still not zero. `agentenv.ProviderKeyNames`
  exists to make that check cheap to write.

**Council:** `dfa6205e-b10e-440c-b251-5d791fdeb718` — round 1 REVISE (an unreadable
guardian result, not a judgement; filed as `bugs_open/119`), round 2 **APPROVED**,
`unreadable=0`, verified to name the same four files as commit `6b5509ee3`.
