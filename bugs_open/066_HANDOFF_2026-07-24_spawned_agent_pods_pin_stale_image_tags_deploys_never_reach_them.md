# BUG 066 — spawned agent pods pin stale image tags; chassis deploys never reach them

**Filed:** 2026-07-24 · gauntlet_dead_cta / feature-builder B4 shakeout
**Status:** **FIXED IN CODE 2026-07-27 (`c0d7c3a71`) — STILL OPEN: INERT until a chassis
image rolls past v1.0.1174.** The defect is reproducible in production until then, so this
stays in `/bugs_open/`. Post-roll verification is § "How to verify" below; do that, then move
the file to `/bugs_closed/`.
**Severity:** high — a rolled chassis image does NOT reach any agent that runs as a
spawned dedicated pod, and the standard deploy verification cannot see the gap.

## Symptom
After rolling agent-chassis v1.0.1155 (formatGeneratedGo fix, bugs_closed/065) and
pod-grepping the deployment pod GREEN (discriminating symbol present), the very next
feature-implementer run (round 6, orch `6d187f71`) failed with the exact error the fix
removes. The spawned implementer pod was running **v1.0.1151**.

## Mechanism
Agents that run as dedicated pods (the `isRepoCloningAgent` spawn class — fix-proposer,
feature-implementer/-orchestrator, and any `spawn_agent` that materialises its own pod)
take their image from **`agent_definitions.image_repository` + `image_tag`** — NOT from
the agent-chassis Deployment.

> **CORRECTED 2026-07-27 — the original next sentence was false, and the correction is
> the whole reason the fix looks the way it does.** It read: *"`make deploy-agents` / a
> kustomize apply updates the Deployment (and its newTag), but **nothing updates the
> per-agent rows**, so every roll widens the gap."*
>
> `make deploy-agents` **does** update the per-agent rows, and always has:
> `makefile:1028` calls `update-agent-images-v2`, which runs
> `UPDATE agent_definitions SET image_repository = …, image_tag = …`.
> `deploy-100-bootstrap-agents` (makefile:519) syncs too. There were **four** copies of
> that UPDATE in the makefile and a fifth in `scripts/deploy/update-agent-images.sh`
> (written Aug 2025, never wired to anything).
>
> **The rows went four tags stale anyway, and that is the finding.** A deploy-time sync
> is a property of one deploy *path*, not of the system. Paths that move the cluster
> without it: `kubectl apply -k …/agent-chassis/overlays/production/uk_001/` — written as
> a comment at the foot of `deploy-agents` itself, i.e. documented as the shortcut;
> `kubectl set image` (`scripts/deploy/deploy-agents.sh`); and `kubectl rollout undo`.
> Undo is the worst case, because it is exactly when spawned pods most need to follow the
> chassis **down**, and a forward-only sync cannot take them there.
>
> *What caught it:* reading the tail of the `deploy-agents` recipe when I went to edit it.
> *The cheap check that would have caught it first:* `grep -n "agent_definitions" makefile`
> — one command, run only afterwards. Logged in `WRONG_CALLS.md` 2026-07-27.

**The durable statement of the bug, after the correction:** the running tag was recorded in
two places that no mechanism kept in step — the Deployment (enforced by Kubernetes) and a
database column (written by whichever deploy path happened to be used). Only the second one
decided what a spawned pod ran.

## Census (2026-07-24, live — as filed)
```sql
SELECT COALESCE(image_tag,'(null)') tag, count(*)
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND image_repository LIKE '%agent-chassis%'
GROUP BY 1 ORDER BY 2 DESC;
-- v1.0.1151 | 168     ← the fleet
-- v1.0.1155 |   2     ← feature-implementer(+orchestrator), hand-updated this session
-- latest    |   1
```

**2026-07-27, the same two records moving independently in the OTHER direction** — caught
by accident mid-session, and the reason the fix covers both: the rows were updated at
`13:44:56` recording v1.0.1173, while the chassis pod's `startTime` was `13:45:31`. For
those 35 seconds the row **led** the Deployment, and spawned pods created 13:49–14:04 were
already on v1.0.1173 while the chassis served v1.0.1172. A sync that only ever writes
forward cannot make these agree; only resolving at spawn time can.

## Why the standard verification misses it
The CLAUDE.md deploy check — `kubectl exec <deployment pod> … strings | grep` — verifies
the DEPLOYMENT's binary. For a spawned agent that check is a **false green**: it proves
the image exists, not that the agent will run it. (Logged in WRONG_CALLS 2026-07-24;
016b §9 pattern added.)

## The fix (`c0d7c3a71` + `e96d42226`, 2026-07-27) — INERT until a roll past v1.0.1174

**Council gate: APPROVED round 1** — corr `3e146ef2-a072-40a8-86be-f6cd940a95f9`, five
advisory objections, none high-severity, `unreadable: 0`. Four were acted on in `e96d42226`
(bound SQL parameters; a log on the silent repository-mismatch fallback; `POD_NAME` honoured
as the house convention ahead of the hostname fallback; and lookups attached to three absence
claims that had been asserted without one). The scope objection — that the row-sync hygiene is
bundled into a fix that disclaims it as the cause — is accurate and was the owner's explicit
call; recorded, not unpicked.

**Half 1 — the fix. `platform/orchestration/actions/agent_image.go` (new).**
A spawned chassis pod now runs the image **its spawner is running**. `resolveAgentImage()`
asks Kubernetes what image *this* pod is on and hands it to the child:

- **Matches on repository.** Only a row naming the same image this pod already runs is
  corrected; an agent deliberately running something else is untouched.
- **`default_config.pin_image_tag = true` is the deliberate override** — the supported
  form of this file's own interim rule, and the SQL sync honours the identical flag (a JSON
  boolean, matched the same way in both places), so a pin means one thing everywhere.
- **Fail-safe.** No cluster, RBAC denied, slow API → falls back to the row, i.e. exactly
  the old behaviour. Success cached for the process; **failure** retried at most every 60 s,
  so one transient error at startup cannot demote a pod for its whole life.
- **No RBAC change needed:** `ai-persona-app` already has `pods: get` (`rbac-job-spawner.yaml`).
  It does **not** have `deployments: get` — verified with `kubectl auth can-i --as=…` — which
  is why this reads its own pod rather than the Deployment.
- Applied at **both** spawn sites: the in-cluster Job (`spawn_actions.go`) and the
  `DispatchRequest` that `remote-job-spawner` builds its pod from (`dispatch_actions.go`);
  the far side has no row to consult, so the correction must happen before the message is sent.
- 24 unit cases, both drift directions, the pin, a non-boolean pin value, a sidecar pod, a
  different repository, a digest reference, and no-self-images.

**Half 2 — hygiene, explicitly NOT the fix.** The row is what every RUNBOOK census reads, so
it must not lie. The five unscoped copies of the sync collapse onto one scoped implementation,
`scripts/deploy/update-agent-images.sh`, which reads the tag from the **live Deployment**
(a make variable is a request; the Deployment is an outcome) and no longer rewrites
`is_snapshot` rollback copies, soft-deleted rows, or the `image_repository` of an agent that
deliberately runs another image. Measured live: the old `WHERE 1=1` hit **183** rows where
**180** is correct.

**New: `scripts/check-agent-image-drift.sh`** (also `make check-agent-image-drift`) — read-only,
separates the three questions the old census conflated: what the Deployment runs, what the rows
say, what spawned pods actually run.

### Rejected, with reasons (so they are not re-proposed)
- **An env var carrying the tag.** The chassis Deployment already has two:
  `AGENT_IMAGE_TAG=v1.0.82` and, via `personae-prod-config`, `agent_image_tag=v1.0.44` —
  both measured inside a pod running v1.0.1173, both read by **no Go code**
  (`grep -rn "AGENT_IMAGE" --include=*.go .` → nothing), both ~1,100 versions stale. A
  duplicated tag rots here; that is evidence, not a prediction.
- **A chassis-startup write-back** of the running tag over the rows. During a rolling deploy
  the old and new pods would fight over the column, and every spawned pod runs this same
  binary, so the write would fan out fleet-wide on every spawn.
- **Deploy-time sync alone.** It already existed and 066 happened anyway — see the correction
  above. Kept as hygiene, not as the fix.

## How to verify (OWED — do this at the next roll, then close the case)
1. **Pod-grep a string the fix created**, not a comment, with a negative control:
   ```
   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
     'strings /app/agent-chassis | grep -c "bugs_open/066: agent_definitions.image_tag trails"'
   ```
   ≥ 1. (A comment is not in the binary — this string is a live log message.)
2. **Induce the failing branch.** A green happy path proves deployment, not correctness:
   snapshot then set one non-critical agent's `image_tag` to a deliberately stale value, fire
   it, and confirm **both**: the spawned pod's `spec.containers[0].image` carries the
   **chassis's** tag, and the chassis log carries
   `bugs_open/066: agent_definitions.image_tag trails …` naming the stale row value.
   ```
   kubectl get pod <spawned> -n ai-persona-system -o jsonpath='{.spec.containers[0].image}'
   ```
3. **Confirm the escape hatch:** set `default_config.pin_image_tag = true` on that same row →
   the spawned pod now runs the pinned tag. Then restore the row.
4. `make deploy-agents` → `scripts/check-agent-image-drift.sh` reports zero drift, and the
   snapshot row was **not** rewritten.

Only when 1–4 pass does this file move to `/bugs_closed/`.

## Interim rule (until the roll)
Unchanged and still required: any fix that must run inside a spawn-class agent — after the
image roll, UPDATE that agent's `image_tag` row (snapshot first) and verify the SPAWNED pod's
image. The deployment pod-grep is necessary but not sufficient. **After the roll this rule
retires**, and its deliberate form becomes `default_config.pin_image_tag`.

## Residual traps found while fixing this (not fixed here)
- **`scripts/deploy/deploy-agents.sh`** rolls the chassis with `kubectl set image`, bypassing
  kustomize entirely, and patches the two dead configmap keys. Its default argument is
  `IMAGE_TAG=${1:-v1.0.44}` — which is exactly the value still sitting in the live configmap,
  so it has been run at least once and nothing has corrected it since. Harmless after this
  fix; still a trap for a reader.
- **`spawn_group.go:248`** writes a step-config `image_tag` that **nothing reads** — a dead
  override. Deliberately not wired: making dead code live is a behaviour change this bug did
  not ask for. Deleting it is the other reasonable call.
- The dead `AGENT_IMAGE_TAG` / `AGENT_IMAGE_REPOSITORY` env vars in
  `agent-chassis/overlays/production/uk_001/patch-deployment.yaml` (and the vet-intel /
  business-intel patches, and `047-base-configs`) are still there, still unread.
