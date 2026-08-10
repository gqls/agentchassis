# 243 — tool-acceptance's `look` step has no storage client on EITHER execution path, so the vision half of acceptance has never run

**Filed 2026-08-10** by `staged_component_build`, at the owner's direction, after the
batch-8 `tool-setup-builder` S6 run surfaced it. Symptom first recorded in that lane's
NOTES (`## 2026-08-10 (second session, later)`) and the 08-10 handoff's defect item 9.

**Status: OPEN. Diagnosed first-hand, fix candidates below, nothing committed as a fix.**

## Declared substitute for the 090 loop (owner ruling 2026-07-31)

This file asserts a root cause without a 090 run. The substitute: the owner directly named
the mechanism to check ("storage is injected at spawn time if the type is listed in the
spawn action code") and every link in the chain below was then read first-hand from the
live code, the live DB and the running pods, with the citation inline at each step. The
chain has no inferred link: the failing line, the nil's origin, the env gate, the spawn
list, the live category row, the per-run `processing_node`, the ConfigMap key→env mapping,
and the deliberate absence of the env on the standing chassis are each individually
quoted. The one inference I did make mid-investigation — "the runs execute inline on the
standing chassis" generalised from one run's shared `owner_agent_id` — was WRONG for 20 of
26 runs and was corrected by reading `processing_node`, which is the column the overlay
comment says to read before concluding anything about which env matters.

## Symptom (measured, could have come out otherwise)

Every `tool-acceptance-agent` orchestration in the retained window ends
`complete_no_look` with the identical `__step_error`:

```
step look failed: failed to execute action execute_vision_prompt:
execute_vision_prompt: no storage client — cannot download screenshots
```

Grouped over all rows (`owner_agent_type='tool-acceptance-agent'`): **26 of 26**
`complete_no_look` carry that message; the only other row is 1 `complete_error` from an
unrelated `request_run` timeout. The grouping by `__step_error` is what distinguishes
"uniformly broken" from "a designed skip branch" — the bare `complete_no_look` count
cannot. Retention bounds the claim: earliest retained row 2026-08-09, so "every run in the
retained window", not "every run ever". (`site_work_items` shows 51 complete
`acceptance_run` items all-history, so the true failure count is likely ~2× the retained.)

**What is NOT lost:** the check results. They are produced by `request_run` and land in
`browser_run.response` before `look` runs, and the judge acts on them. What is lost is the
screenshot/vision pass — the half that catches what a selector cannot see (a tool that is
present, correct and invisible).

## Root cause — one gate, two execution paths that both fail it

`execute_vision_prompt` hard-errors when `params.StorageClient == nil`
(`platform/orchestration/actions/execute_vision_prompt_action.go:87`). That client is the
agent's own: `agentbase/agent.go:335` hands `a.storageClient` to the coordinator
(`coordinator.go:101`, then into every action at `coordinator.go:1703`), and
`agentbase/agent.go:316` builds it **only when `IMAGE_BUCKET` is set** — otherwise it logs
"Storage client not configured (IMAGE_BUCKET not set)" and leaves it nil.

`orchestration_states.processing_node` splits the 26 failing runs across two environments,
and **neither has `IMAGE_BUCKET`**:

| path | runs | why the env is absent |
|---|---|---|
| spawned `agent-tool-acceptance-agent-<hash>` pods (the overnight `acceptance_run` due-sweep) | **20** | spawn injects storage env only when `isStorageEnabledAgent(type) \|\| category ∈ {orchestrator, code-driven}` (`spawn_actions.go:2556`). `tool-acceptance-agent` is **not** in the 12-name `storageAgents` list (`spawn_actions.go:3040-3053`), and its live `agent_definitions.category` is **`tools`** — so no branch fires and nothing storage-shaped is injected |
| standing `agent-chassis-*` pods (manual kcat dispatches on the generic topic, e.g. `tool_acceptance_run.sh`) | **6** | the deployment **deliberately** carries no `IMAGE_BUCKET`/`S3_ENDPOINT` — owner ruling 2026-08-08, recorded in `deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml` (~line 100): an earlier revision added them (`820a033c0`, 2026-08-05) and was reverted. The B2 credentials ARE present there (older, `019cf8d94`, explicitly not part of that revert); it is the bucket/endpoint config that is absent, and `agentbase` gates on the bucket |

Supporting facts, each checked:

- The spawn block's ConfigMap mapping is complete and correct for the listed types:
  `storage-config` keys `S3-ENDPOINT`/`S3-REGION`/`image_bucket`/`assets_bucket`/
  `S3_USE_PATH_STYLE` map to env names `S3_ENDPOINT`/`S3_REGION`/`IMAGE_BUCKET`/
  `ASSETS_BUCKET`/`S3_USE_PATH_STYLE` (`spawn_actions.go:2588-2634`), plus AWS/B2
  credentials as direct values (`spawn_actions.go:2568-2580`). So membership of the list
  is sufficient — a listed type gets everything `agentbase` needs.
- **The screenshots exist.** The browser-runner adapter has its own client and uploads
  fine — my run's `browser_run.response.renders` names
  `s3://personae-prod-uk001-images/acceptance-evidence/<site>/tool-setup-builder/…_desktop.png`
  and `…_mobile.png`. The orchestration-side download is the only missing piece.
- **Blast radius is exactly this agent.** A `default_config` scan over live
  `agent_definitions` finds `tool-acceptance-agent` is the **only** active agent whose
  workflow uses `execute_vision_prompt`. render-audit's vision runs adapter-side and its
  recent orchestrations complete with no step error on the same chassis pods.
- The overlay comment's reason 2 ("agents do not execute in this deployment") is true for
  the sweep and **false for manual dispatches** — 6 of 26 runs, including both of this
  lane's S6 runs, executed inline on standing chassis pods. Its own instruction ("read
  `processing_node` before concluding anything about which env matters") is what caught
  this; the comment should not be weakened, but its reason-2 wording overstates.

## Fix candidates, ordered by what closes the door

1. **Add `"tool-acceptance-agent"` to `storageAgents` (`spawn_actions.go:3040`).** One
   line. Fixes the scheduled sweep — 20 of the 26 observed runs, and 100% of the
   unattended path, which is the one that matters "for ever". Consistent with BOTH halves
   of the 2026-08-08 ruling: the list is the sanctioned mechanism for granting storage to
   a specific spawned type, rather than spreading env across the whole chassis
   deployment. Additive, opt-in, changes no shared guarantee → normal council gate, no
   RFC (2026-07-29 ruling). Verify after the next roll by the spawn log line
   `Injecting storage credentials {agent_type: tool-acceptance-agent}` and a sweep run
   reaching `complete` (not `complete_no_look`), with `__step_error` empty.
2. **The manual path needs an owner decision, and candidate 1 does not touch it.** kcat
   dispatches on the generic topic run inline on the standing chassis, which stays
   bucket-less under the 08-08 ruling. Options: (a) accept that manual runs keep losing
   the `look` step (check half still lands — it is how all of this lane's S6 greens were
   read); (b) reshape the manual trigger to go through the spawn path like the sweep;
   (c) reopen the 08-08 ruling for the chassis deployment. (a) costs nothing today and
   leaves a knowingly-degraded path; do not choose it silently.
3. **Make the degradation visible either way**: `complete_no_look` reads as COMPLETED and
   nobody noticed 26 consecutive losses. Whatever else is done, the `look` failure should
   mark the run's summary (e.g. `vision: skipped(<reason>)`) so a green with no vision is
   distinguishable from a green with vision — this is the same "a silent gate either did
   not look or approved" shape the memory index already records.

## How to verify (whoever fixes)

```sql
-- before: every retained sweep run says complete_no_look + the storage message
SELECT processing_node, current_step,
       collected_data->'__step_error'->>'message'
  FROM orchestration_states
 WHERE owner_agent_type='tool-acceptance-agent'
 ORDER BY created_at DESC LIMIT 5;
```

After candidate 1 rolls: dispatch one acceptance run (the due-sweep raises them
overnight, or `tool_acceptance_run.sh` — but note a MANUAL run tests the manual path,
which candidate 1 does not fix; to test the fix you need a SPAWNED run). Expect
`current_step='complete'`, no `__step_error`, and the vision verdict populated. The spawn
env can be checked directly on the pod while it lives:
`kubectl exec <agent-tool-acceptance-agent-…> -- env | grep -E "IMAGE_BUCKET|S3_ENDPOINT"`.
