# CONTRIB 2026-08-09 — the A2 blocker is refuted: spawned pods DO get S3, through a gate your agent type simply failed

From the session the owner has been driving on bugs_open/198 and follow-ups. This
corrects the ⚠⚠ 2026-08-08 warning in `CONTRIB_2026-08-03_acceptance_renders…` (commit
`86581e265`), which told this lane not to seed the A2 critic until an architecture
question about `execute_vision_prompt`'s contract was settled. **The premise is wrong;
the action's contract does not have to give. A2 needs one line, not an architecture
round.** A dated correction now sits under that warning; this file carries the evidence.

## The mechanism (verified in code AND against every live pod)

`spawn_actions.go:2556` — the chassis spawner (`spawnAgentKubernetesJobFromDefinition`,
the only in-cluster creator of `agent-*` Jobs) injects the full S3 environment
conditionally:

```go
if isStorageEnabledAgent(agentDef.Type) || agentDef.Category == "orchestrator" || agentDef.Category == "code-driven" {
```

- `isStorageEnabledAgent` (line 3039): a hardcoded 12-type allow-list (site-publisher,
  html-developer, visual-designer, image-generator, content-creator, content-researcher,
  domain-analyst, site-architect, website-builder, asset-deployer,
  training-data-preparer, artefact-collector).
- `Category` is `agent_definitions.category`, scanned straight from the DB
  (`getAgentDefinition`, line ~2168).
- What passes the gate gets: `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` /
  `B2_APPLICATION_KEY_ID` / `B2_APPLICATION_KEY` copied **as literal values from the
  spawner's own environment** (which is why the 08-05 chassis-overlay env fix was
  load-bearing, not a no-op), plus `S3_ENDPOINT` / `S3_REGION` / `IMAGE_BUCKET` /
  `ASSETS_BUCKET` / `S3_USE_PATH_STYLE` via configMapKeyRef → `storage-config`.

## The census (2026-08-09, all running spawned pods, zero counterexamples)

| type | category | on the list? | storage env |
|---|---|---|---|
| build-dispatch-loop | orchestrator | no | present |
| image-build-handler | orchestrator | no | present |
| feed-ingester | code-driven | no | present |
| content-feed-orchestrator | code-driven | no | present |
| deployer-agent | data-driven | no | absent |
| page-rerender / section-editor / tool-auditor / tool-improver | specialist | no | absent |
| tool-acceptance-agent (the 08-08 test pod) | tools | no | absent |

10/10 types predicted by the gate. The four `present` rows are on it by CATEGORY, not
by the list — the category clause is live and doing most of the work.

## Why the 08-08 test failed — and why the generalisation didn't hold

`tool-acceptance-agent` is category `tools` (the only definition in that category) and
not on the list → its pod genuinely had no S3 env. Every observation in `86581e265` is
true OF THAT POD; "agents run in spawned pods with NO S3 credentials" is false as a
class statement. Three words called "orchestrator" invite exactly this confusion, and
the gate reads only the third: the pod label `spawned-by: orchestrator` (on EVERY
spawned pod — verified on a storage-less section-editor pod), the workflow
`processing_mode: "orchestrator"` (near-universal), and `agent_definitions.category`
(24 distinct values across ~185 active definitions; `orchestrator` covers 21).

## What unblocks A2

Either of:
1. **One Go line**: add the critic's type to `isStorageEnabledAgent` — ships with the
   next chassis roll (pod-grep it before trusting).
2. **Config-only, live immediately**: seed the new agent definition with category
   `code-driven` IF that is an honest description — do not bend the category to smuggle
   credentials; if it isn't honest, take the Go line.

## Side findings, so nobody trips on them next

- The `valueFrom`/`secretKeyRef` entries in html-developer's and visual-designer's
  `env_vars` columns are **decorative**: both spawners' custom-env structs carry only
  `Name`/`Value`, so `valueFrom` is silently dropped. Those types work because they are
  on the allow-list. Copying that env_vars pattern for a new agent yields nothing.
- `cmd/remote-job-spawner` has **no storage block at all** — the same two-spawner drift
  class bugs_open/112 fixed for provider keys. A storage-enabled agent dispatched
  remotely would fail exactly the way the 08-08 pod did.
- Storage credentials land as **literal values in pod specs** (readable by anyone with
  pod read). `43c1801d6` stopped logging them; it did not change how they travel.
