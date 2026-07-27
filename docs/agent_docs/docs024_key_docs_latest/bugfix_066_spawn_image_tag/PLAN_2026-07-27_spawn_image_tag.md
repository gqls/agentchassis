# PLAN — bugs_open/066: spawned agent pods pin stale image tags

**Opened:** 2026-07-27 · **Case:** `bugs_open/066` (filed 2026-07-24 by the
feature-builder B4 shakeout, then handed to "whoever owns deploys")
**State:** fixed in code `c0d7c3a71`, **INERT** until a chassis roll past v1.0.1174.

## The problem

An agent that runs as a dedicated spawned pod takes its container image from
`agent_definitions.image_repository` + `image_tag`. The agent-chassis Deployment takes its
image from kustomize. Those are two records of one fact, and only the second is enforced by
Kubernetes. On 2026-07-24 the fleet sat at v1.0.1151 while the Deployment ran v1.0.1155 and
a feature-implementer run failed on a bug that had already shipped — minutes after a green
deployment pod-grep, which for this class of agent proves only that the image *exists*.

## The correction that shaped the fix

> **CORRECTED 2026-07-27, before any code was written.** My first reading of the makefile
> concluded that `make deploy-agents` never syncs the rows and that the deploy-time sync
> existed three times and was wired into the roll *zero* times. **That was false.**
> `deploy-agents` calls `update-agent-images-v2` at `makefile:1028` and always has.
>
> The rows went four tags stale anyway, and that is a **better** finding: a deploy-time sync
> is a property of one deploy *path*, not of the system. `kubectl apply -k …` (the shortcut
> written as a comment at the foot of `deploy-agents` itself), `kubectl set image`
> (`scripts/deploy/deploy-agents.sh`) and `kubectl rollout undo` all move the cluster with no
> make involved. **Undo is the decisive case**: it is exactly when spawned pods most need to
> follow the chassis *down*, and a sync that only ever writes forward cannot take them there.
>
> Caught by reading the tail of the `deploy-agents` recipe when I went to edit it. The cheap
> check: `grep -n "agent_definitions" makefile`. Logged in `WRONG_CALLS.md`; the owner had
> already taken a decision on the false premise, and was told.

## Decisions, and why

| # | Decision | Reason |
|---|---|---|
| 1 | **Resolve the spawn image at spawn time from the RUNNING chassis pod.** | The tag a spawned chassis pod should run is not a fact about a database row; it is a fact about the process doing the spawning. This is the only form that survives every deploy path, including rollback. |
| 2 | **Match on repository, not on "is it the chassis".** | An agent whose row names a different image is left completely alone, with no list of "chassis-class agents" to maintain — the kind of list that drifts. |
| 3 | **Read our own pod, not the Deployment.** | `ai-persona-app` already has `pods: get`; it does **not** have `deployments: get` (verified `kubectl auth can-i --as=…`). Reading the Deployment would need an RBAC change *and* would return the new tag mid-rollout while the pod still ran the old one. |
| 4 | **Fail-safe to the row, always.** | The worst case of a broken lookup is the bug we already had, never a failed spawn. Success cached for the process; **failure** retried at most every 60 s, so one transient error at startup cannot demote a pod for its whole life. |
| 5 | **`default_config.pin_image_tag` as the deliberate override.** | 066's own interim rule depends on being able to pin a row; the fix would otherwise silently take that away. In `default_config` rather than a new column because **zero rows pin today** — a column plus migration is speculative until something does. Reversal trigger: the first real pin, or any need to constrain it in the schema. |
| 6 | **Keep the deploy-time row sync as hygiene, not as the fix.** | The row is what every RUNBOOK census reads; a record that lies is its own defect. But it is explicitly not what closes the door — see the correction above. |
| 7 | **Do not add an env var carrying the tag.** | The chassis Deployment already has two, `AGENT_IMAGE_TAG=v1.0.82` and `agent_image_tag=v1.0.44`, measured inside a pod running v1.0.1173, read by no Go code. A duplicated tag rots here; that is evidence, not a prediction. |
| 8 | **Do not have the chassis write the running tag back over the rows at startup.** | During a rolling deploy the old and new pods would fight over the column, and every spawned pod runs this same binary, so the write would fan out fleet-wide on every spawn. |
| 9 | **Do not wire `spawn_group.go:248`.** | It sets a step-config `image_tag` that nothing reads — a dead override. Making dead code live is a behaviour change this bug did not ask for. Flagged in the case file as a residual; deleting it is the other reasonable call. |
| 10 | **Do not roll it myself** (owner, this session). | Commit + council + docs; the next roll picks it up. 066 stays OPEN until then — "fixed" is not "live", and the defect is reproducible in production until the roll. |

## Phasing

- **P1 — resolver + call sites + tests.** DONE (`c0d7c3a71`). 24 unit cases pass;
  `platform/`, `internal/`, `pkg/` build.
- **P2 — row sync consolidation + drift check.** DONE, same commit. Predicate proven
  read-only against production; both branches of the drift check exercised.
- **P3 — council gate.** Submitted, corr `3e146ef2-a072-40a8-86be-f6cd940a95f9`.
- **P4 — the roll.** NOT ours. Whoever rolls next runs the four verification steps in
  `bugs_open/066` § "How to verify", then moves the case to `/bugs_closed/`.

## What would falsify the design

If a spawned pod ever needs to run a *different* chassis build from its spawner as normal
operation — a canary, a per-agent rollout — then repository-matching is wrong and the pin
becomes load-bearing rather than an escape hatch. Nothing does that today (175/175 active
rows, one image, one tag).
