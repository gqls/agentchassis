# CONTINUE HERE — the disk-pressure lane (`bugs_open/252`), 2026-08-11 18:05Z

Cold-start handoff for a new chat. The bug file itself
(`252_HANDOFF_2026-08-11_disk_is_invisible_to_the_scheduler…`) holds the evidence
and the fix candidates; **this file holds only what is still owed and what will
mislead you.** Read the bug file's `⚠ STATE AS OF 2026-08-11 18:00Z` block first.

## Where this came from

The owner asked why a chassis pod was evicted during the 09:23Z roll, and whether
it meant too few hosts or bad distribution. Answer: **neither.** CPU sits at 1–5%
and memory at 7–42%, but **0 of 143 containers requested `ephemeral-storage`**, so
the scheduler balanced CPU and memory and was blind to disk. That is why both
2.3 GB CI runners were on one node. Compounded by `imageGCHighThresholdPercent: 85`
being *exactly* the `imagefs.available < 15%` eviction line, so reclaim never runs
before evictions. Filed as 252.

Also note the pod was **not evicted** — `Pod was rejected` is the kubelet refusing
to *admit* a newly-scheduled pod onto an already-tainted node. Different mechanism,
different fix. Do not re-derive this.

## State of candidate 1 (the only one acted on)

> **✅ DONE 2026-08-12 13:00Z — the owner applied it; census is 5.** The table below
> is history. Read the bug file's `✅ STATE AS OF 2026-08-12 13:00Z` block for the
> outcome, the before/after headroom, and **candidate 1b**, which is what this lane
> now owes.

| deployment | manifest | LIVE in cluster |
|---|---|---|
| `github-actions-runner` (2 replicas) | 4Gi on `runner` | ✅ live 12:55Z — now on …1336 and …1149 |
| `github-actions-runner-vmsites` | 4Gi on `runner` | ✅ live 12:55Z — **moved off …1149** to …6833 |
| `ollama-adapter` | 1Gi on `model-pull` **and** `ollama` | ✅ live overnight, both containers |
| `ollama-eval` | 1Gi on `ollama` | ✅ live 12:58Z, serving |

Commits: `301161274` (the requests + a correction to what they buy),
`eab8e7fe8` (the not-applied note), `838ffa163` (wrong-container fix + WRONG_CALLS).

**Fleet census to watch — it read 0, now reads 1, and must reach 5** (2 runner
replicas + vmsites + ollama-adapter + ollama-eval):

```bash
kubectl get pods -A -o json | python3 -c "
import json,sys;d=json.load(sys.stdin)
print(sum(1 for p in d['items'] if p['status'].get('phase') in ('Running','Pending')
      for c in p['spec']['containers']
      if ((c.get('resources') or {}).get('requests') or {}).get('ephemeral-storage')))"
```

## The three traps in this lane

1. **`grep -c ephemeral-storage` on a rendered overlay is not validation.** It
   returned `1` for all four services and I read four 1s as four successes; on
   `ollama-adapter` the field was on the wrong container. **Parse the YAML and
   print per container** — the one-liner is in the bug file's 18:00Z block.
   A pod's effective request is `max(sum of containers, max of initContainers)`, so
   the wrong placement still produces the right pod-level number and hides itself.
2. **Requests do NOT stop pods landing on a nearly-full node**, and the bug file
   originally said they did (corrected in place). `ephemeral-storage` is scheduled
   against **allocatable** — 35.1 GB — while real free space is 7.6–14.1 GB. The
   21–27 GB gap is images plus system, charged to no request. Requests only stop
   the big writable-layer consumers being co-located.
3. **`kubectl describe node` prints `ephemeral-storage 0 (0%)` whether nothing
   requests it or nothing exists.** Those read identically. Use the census above.

## What is still owed

- **Apply candidate 1** — three deployments left (`github-actions-runner`,
  `github-actions-runner-vmsites`, `ollama-eval`); `ollama-adapter` is done.
  > **The reason it is still owed CHANGED on 08-12.** It was a load judgement at
  > 15:05Z on 08-11 (38 in-flight orchestrations, 27 LLM calls/15 min, CI running).
  > That is gone: measured 08-12 12:37Z the runners are idle between ~20 s jobs,
  > LLM traffic is **1** call/15 min, and both ollama pods have served **0**
  > inference requests in 2 hours. The blocker now is **local permission** — the
  > apply was refused by the Claude Code classifier at 12:38Z, cluster untouched.
  > A session with the permission, or the owner running it directly, unblocks it.

  Verify first, then apply (`diff` is the honest pre-check — the whole diff should
  be `generation` plus one `+ ephemeral-storage` line, nothing else):
  ```bash
  cd /home/ant/projects/agentchassis
  for d in github-actions-runner github-actions-runner-vmsites ollama-eval; do
    kubectl diff  -k deployments/kustomize/services/$d/overlays/production/uk_001
    kubectl apply -k deployments/kustomize/services/$d/overlays/production/uk_001
  done
  ```
  Then re-run the census below — it must go **1 → 5**.
- **Candidates 2–4 are owner decisions, not thread work.** (2) limits — deliberately
  not added: a limit evicts the pod that exceeds it, so a large build would die
  mid-run, and there is no measured worst-case build size yet. (3) lower
  `imageGCHighThresholdPercent` to ~70 so reclaim precedes eviction — a node-pool
  change. (4) grow the 41.4 GB disks — costs money, closes no door, do it *with*
  1–3 and never instead.

## Everything else from this session is CLOSED — do not reopen

- **`bugs_open/236` (522 half)** — fixed, live, drill-proven both halves (filed on a
  real 90-second cookly.uk outage, self-cleared on restore), council **APPROVED**
  `7177fb02`. Its handoff is closed and says so.
- **`bugs_open/236` (hero/logo half)** — **owned by another lane, actively.** I only
  contributed the retention finding: `AWAITING_RESPONSES` rows live **4 hours**,
  which is why its evidence keeps evaporating. Do not take it.
- **Cloudflare lane** — notified in their NOTES (apex detector now exists; route
  changes lag both ways; the 404-token cannot read DNS).
- **090 explainer** — written at the owner's request:
  `docs024_key_docs_latest/fixloop_eg_dartsonline/SUMMARY_090_diagnosis_loop_what_it_is.md`.

## The honest summary line

**Candidate 1 is done, live and proven: the fleet requests disk (census 5), the two
co-located runners are on separate nodes, and the 0.82 GB node is back to 3.43 GB.**
That sentence is now earned at the pod, not at the manifest.

**What it does NOT mean, and this is the live trap in this lane:** the co-location
cannot recur is **false**. There is no anti-affinity and no topology spread on the
runners — today's good placement is a scheduler score, not a rule, and 2×4Gi on one
35.1 GB-allocatable node remains entirely legal. **Candidate 1b** closes that; it is
the cheapest remaining item and nobody owns it.

> **The claim's history, kept because the shape is the lesson.** 08-11: "one-third
> done and none of it proven in the cluster". 08-12 12:40Z: "one-fifth done, that
> fifth proven" — the denominator changed from 3 deployments to 5 containers under a
> fraction that never stated it. 08-12 13:00Z: done and proven. A fraction over an
> unstated denominator is the shape to distrust in your own writing.
