# BUG 066 — spawned agent pods pin stale image tags; chassis deploys never reach them

**Filed:** 2026-07-24 · gauntlet_dead_cta / feature-builder B4 shakeout · **OPEN**
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
the agent-chassis Deployment. `make deploy-agents` / a kustomize apply updates the
Deployment (and its newTag), but **nothing updates the per-agent rows**, so every roll
widens the gap.

## Census (2026-07-24, live)
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
**Exposure nuance (honest):** rows whose agents only ever execute inside the shared
chassis deployment are inert carriers of the stale tag; the ACTIVE harm is on the
spawned-pod class. The verified harm instance is the feature-implementer (round 6).
Which of the 168 are spawn-class needs the census refined against the spawn gate's
predicate before any bulk fix.

## Why the standard verification misses it
The CLAUDE.md deploy check — `kubectl exec <deployment pod> … strings | grep` — verifies
the DEPLOYMENT's binary. For a spawned agent that check is a **false green**: it proves
the image exists, not that the agent will run it. (Logged in WRONG_CALLS 2026-07-24;
016b §9 pattern added.)

## Fix candidates
1. **Deploy-time sync:** `deploy-agents` (or a companion migration/script) updates
   `agent_definitions.image_tag` for chassis-image rows in the same step that bumps the
   Deployment — one authority for "current tag". Needs the spawn-class nuance decided:
   update all chassis-image rows, or only spawn-class ones.
2. **Indirection:** spawn path resolves the tag from one place (a config row/configmap)
   instead of per-agent columns; per-agent pinning becomes an explicit override.
3. **Advisory check:** a discovery/pattern check flagging active agents whose image_tag
   trails the deployed chassis tag (cheap, catches drift without changing the deploy).

## How to verify a fix
Roll a chassis tag; then `SELECT DISTINCT image_tag` for active chassis-image agents
must equal the rolled tag (or the documented override); fire a spawn-class agent and
`kubectl get pod <spawned> -o jsonpath='{.spec.containers[0].image}'` must show it.

## Interim rule (until fixed)
Any fix that must run inside a spawn-class agent: after the image roll, UPDATE that
agent's `image_tag` row (snapshot first) and verify the SPAWNED pod's image — the
deployment pod-grep is necessary but not sufficient.
