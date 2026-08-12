# CONTINUE HERE — the disk-pressure lane (`bugs_open/252`)

**Last updated 2026-08-12 17:35Z.** Cold-start handoff for a new chat. The bug file
(`252_HANDOFF_2026-08-11_disk_is_invisible_to_the_scheduler…`) holds the evidence and
the fix candidates; **this file holds only what is still owed and what will mislead
you.** Read the bug file's newest block — `🔎 STATE AS OF 2026-08-12 17:30Z` — first;
it re-ranks the candidates and everything below assumes it.

## Where this came from

The owner asked why a chassis pod was evicted during the 08-11 09:23Z roll, and
whether it meant too few hosts or bad distribution. Answer: **neither.** CPU sits at
1–5% and memory at 7–42%, but **0 of 143 containers requested `ephemeral-storage`**,
so the scheduler balanced CPU and memory and was blind to disk. Compounded by
`imageGCHighThresholdPercent: 85` being *exactly* the `imagefs.available < 15%`
eviction line, so reclaim never runs before evictions. Filed as 252.

Also: the pod was **not evicted** — `Pod was rejected` is the kubelet refusing to
*admit* a newly-scheduled pod onto an already-tainted node. Different mechanism,
different fix. Do not re-derive this.

## ✅ Candidate 1 is DONE — applied, live, proven. Do not redo it.

The owner applied it 2026-08-12 12:55Z. **Census went 0 → 5** and all four
deployments carry their request in the live pod:

| deployment | request | live |
|---|---|---|
| `github-actions-runner` ×2 | 4Gi on `runner` | ✅ …1336 and …1149 |
| `github-actions-runner-vmsites` | 4Gi on `runner` | ✅ **moved off …1149** to …6833 |
| `ollama-adapter` | 1Gi on `model-pull` + `ollama` | ✅ both containers |
| `ollama-eval` | 1Gi on `ollama` | ✅ serving since 12:58Z |

It worked: the two 2.3 GB runners had been co-located on …1149 and are now on
separate nodes, and …1149 recovered 0.82 GB → 3.43 GB at the time. Sizing was right
— measured actual usage is 2.33 GB and 2.15 GB against the 4Gi requests.

```bash
# the census — must stay at 5
kubectl get pods -A -o json | python3 -c "
import json,sys;d=json.load(sys.stdin)
print(sum(1 for p in d['items'] if p['status'].get('phase') in ('Running','Pending')
      for c in p['spec']['containers']
      if ((c.get('resources') or {}).get('requests') or {}).get('ephemeral-storage')))"
```

## What is still owed, in the order the measurement now supports

**Candidate 3 — lower `imageGCHighThresholdPercent` to ~70 (low ~60).** Highest
actionable leverage: images are 10.5–16.5 GB per node and are reclaimable in
principle, but at 85 the kubelet cannot start reclaiming until it is already at the
eviction line. Re-verified 17:30Z as still `85/80` with
`evictionHard.imagefs.available: 15%` and `imageMinimumGCAge: 2m0s` — so GC is held
back **only** by the threshold, not by image age. **This is a kubelet/node-pool
change, i.e. an owner action, not a thread's.**

**Candidate 1b — give the runners a `topologySpreadConstraints` or `podAntiAffinity`.**
Cheap, manifest-only, unowned, and nobody has started it. Today's good placement is a
**scheduler score, not a rule**: measured 13:00Z, `github-actions-runner` has
`affinity: None` and `topologySpreadConstraints: None`, and with 35.1 GB allocatable
per node putting both 4Gi replicas back on one node stays entirely legal. The next
roll may quietly undo candidate 1's main win and nothing would report it.

**Open question, and it is now the biggest one — what is "other"?** Per node:
images 10.5–16.5 GB, pod writable layers 0.44–2.72 GB, and **"other" 14.3–17.7 GB,
the largest category, 35–43% of every disk.** `[UNVERIFIED]` composition — plausibly
node OS, `/var/log/pods`, containerd metadata, but **I did not open a node to look
and it must not be repeated as fact until someone does.** It matters because no
candidate on file addresses it, and if ~15 GB/node is structurally unavailable then
these 41.4 GB disks have only ~26 GB of usable budget and **candidate 4 (grow the
disks) is much stronger than the filing's "buys margin, closes no door".**

**Candidates 2 and 4 remain owner decisions.** (2) limits — deliberately not added: a
limit evicts the pod that exceeds it, so a large build would die mid-run, and there
is still no measured worst-case build size. It also governs the smallest category.
(4) grow the 41.4 GB disks — costs money; see the "other" question above before
pricing it.

## The traps in this lane

1. **`grep -c ephemeral-storage` on a rendered overlay is not validation.** It
   returned `1` for all four services and the previous session read four 1s as four
   successes; on `ollama-adapter` the field was on the wrong container. A pod's
   effective request is `max(sum of containers, max of initContainers)`, so the wrong
   placement still produces the right pod-level number and hides itself. **Parse the
   YAML and print per container** — one-liner in the bug file's 18:00Z block.
2. **Requests do NOT stop pods landing on a nearly-full node.** `ephemeral-storage`
   is scheduled against **allocatable** — 35.1 GB — while real free space is
   7.9–14.4 GB. The gap is images plus system, charged to no request. Requests only
   stop the big writable-layer consumers being co-located.
3. **`kubectl describe node` used to print `ephemeral-storage 0 (0%)` whether nothing
   requested it or nothing existed** — those read identically. Since candidate 1 it
   discriminates (4Gi/1Gi/4Gi/5Gi across four nodes), but prefer the census above.
4. **`kubectl diff -k` is the honest pre-apply check.** For a request-only change the
   entire diff should be `generation: N→N+1` plus one `+ ephemeral-storage` line. Run
   it immediately before applying — another session may have moved the overlay.
5. **An in-flight orchestration count needs an `updated_at` bucket, or it overstates
   load ~20×.** On 08-12, 46 of 48 `RUNNING`/`EXECUTING_STEP`/`INITIALIZED` rows had
   not been touched in over 2 hours. Statuses there are **UPPERCASE** — a lowercase
   `NOT IN ('completed',…)` filter matches nothing and silently returns every row.
   `[UNVERIFIED]` whether those stale rows are `bugs_open/029`'s hung spawns.
6. **The `build provenance` line is a STARTUP line and scrolls.** Empty output from
   `logs -l app=agent-chassis --tail=400 | grep 'build provenance'` means **"not in
   range", not "unstamped"** — confirmed again 08-12 at 85 minutes of pod age. Fall
   back to the binary probe, with a control, per CLAUDE.md.
7. **One clean roll is not proof.** The 08-12 ~14:55Z chassis roll produced no
   rejection and no `DiskPressure`, which is encouraging and nothing more — the
   original failure was intermittent and no failure rate has been established. The
   `FailedScheduling` on `alertmanager-…-0` is an **unbound PVC**, a different
   mechanism — not ours.

## Everything else from the originating session is CLOSED — do not reopen

- **`bugs_open/236` (522 half)** — fixed, live, drill-proven both halves, council
  **APPROVED** `7177fb02`. Its handoff is closed and says so.
- **`bugs_open/236` (hero/logo half)** — **owned by another lane, actively.** Only
  contribution was the retention finding: `AWAITING_RESPONSES` rows live **4 hours**,
  which is why its evidence keeps evaporating. Do not take it.
- **Cloudflare lane** — notified in their NOTES (apex detector now exists; route
  changes lag both ways; the 404-token cannot read DNS).
- **090 explainer** — `docs024_key_docs_latest/fixloop_eg_dartsonline/SUMMARY_090_diagnosis_loop_what_it_is.md`.

## The honest summary line

**Candidate 1 is done, live and proven — the fleet requests disk, and the two big
consumers are on separate nodes.** But the lane is not fixed and 252 stays OPEN:
requests are scheduled against 35.1 GB allocatable while real free space is
7.9–14.4 GB, fleet-worst headroom has drifted back to **1.73 GB with three of five
nodes now under 2.0 GB**, and the per-category measurement shows the candidates on
file govern the two smaller categories while the largest — ~15 GB/node of "other" —
is unexplained and unaddressed.

> **Claim-shape lesson kept deliberately.** This line has read "one-third done",
> then "one-fifth done", then "done" — the denominator changed from 3 deployments to
> 5 containers under a fraction that never stated it. **A fraction over an unstated
> denominator is the shape to distrust in your own writing.**
