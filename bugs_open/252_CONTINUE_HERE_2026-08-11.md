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

| deployment | manifest | LIVE in cluster |
|---|---|---|
| `github-actions-runner` (2 replicas) | 4Gi on `runner` | **not applied** — pods from 07-21 / 08-11 09:24 |
| `github-actions-runner-vmsites` | 4Gi on `runner` | **not applied** — pod from 07-16 |
| `ollama-adapter` | 1Gi on `model-pull` **and** `ollama` | **partly** — init has it live since the 17:13Z roll; the app container's 1Gi landed at 18:00Z and is not rolled |
| `ollama-eval` | 1Gi on `ollama` | **not applied** — pod from 06-28 |

Commits: `301161274` (the requests + a correction to what they buy),
`eab8e7fe8` (the not-applied note), `838ffa163` (wrong-container fix + WRONG_CALLS).

**Fleet census to watch — it must stop being 0, and should reach 5** (2 runner
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

- **Apply candidate 1.** Not done because at 15:05Z the cluster was busy with other
  sessions' work (38 in-flight orchestrations, 27 LLM calls in 15 min, runners
  finishing CI). It needs no special action — the four overlays pin exactly the
  images and replica counts already running, so any ordinary roll picks it up, as
  `ollama-adapter` just demonstrated. Apply commands are in the bug file.
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

The diagnosis is solid and measured; **the fix is one-third done and none of it is
proven in the cluster yet.** Until that census returns 5, the correct sentence is
"the manifests declare it" — not "the fleet requests it".
