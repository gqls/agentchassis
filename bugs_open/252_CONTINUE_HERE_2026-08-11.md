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

**Candidate 5 — set `SystemMaxUse=512M` in `/etc/systemd/journald.conf`. NEW, and it
is the cheapest and biggest item on the list.** `[MEASURED 08-12 18:30Z on …1148 and
…1149, root]` journald holds **3.87 and 3.85 GiB** — the largest non-container
consumer on both — because `journald.conf` contains nothing but `[Journal]` and the
default `SystemMaxUse` is *min(10% of filesystem, 4 GiB)*; 10% of 38.6 GiB = 3.86 GiB
and both nodes sit on it to within 10 MiB. Capping at 512M returns **~3.4 GiB per
node** (`[EXTRAPOLATED]` ~17 GiB fleet — three nodes were not opened), against
current headroom of 1.7–2.0 GiB. It roughly **triples** the margin on the tight
nodes, for one line. Owner's call (node config). Check `journalctl --disk-usage` and
the oldest retained entry first; reclaim with `journalctl --vacuum-size=512M`.

> **The "what is other?" question is ANSWERED — do not re-open it.** On …1148, of
> 30.7 GiB used: containerd snapshots 14.07 (which **includes** container writable
> layers), **containerd content store 5.70**, **journal 3.87**, `/usr` 4.17, `/boot`
> 0.45, calico logs 0.32, `/opt` 0.20, `/var/log/pods` **0.18**. Total ~30.5 against
> `df`'s 30.7 — it closes. Two consequences: the kubelet's `imageFs.usedBytes` is
> **exactly** the snapshotter and omits the content store, so **an image costs ~1.4×
> what the kubelet reports**; and **container log rotation is not worth touching** at
> 0.18 GiB. Full table and the arithmetic: the bug file's `🔦 18:30Z` block.

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
7. **A `du` total is only a total if you can read every subtree — and the shortfall is
   silent.** Investigating "what is other?" the first pass used the `node-exporter`
   pod (mounts host `/` read-only at `/host/root`, no pod to create). It runs as
   **nobody**, `/var/lib/containerd` is root-only, and the command ended
   `2>/dev/null`. `du -xd1` returned a confident **10.8 GiB** for the whole
   filesystem while `df` said **30.7 GiB used** — two thirds skipped, no error, a
   perfectly plausible number. Use `kubectl debug node/<node>
   --profile=sysadmin`, send stderr to **stdout, never `/dev/null`**, and
   **reconcile against `df` before believing any total.** The discrepancy is the only
   thing that catches it. Delete the debug pod afterwards.
8. **One clean roll is not proof.** The 08-12 ~14:55Z chassis roll produced no
   rejection and no `DiskPressure`, which is encouraging and nothing more — the
   original failure was intermittent and no failure rate has been established. The
   `FailedScheduling` on `alertmanager-…-0` is an **unbound PVC**, a different
   mechanism — not ours.

## One loose end with a correlation to chase

A LANDMINES entry was filed 2026-08-12 17:40Z — *"An `orchestration_states` status
count is not a measure of cluster load"*, covering trap 5 above (commit `122cb945c`,
synced to `doc_notes`). A `landmine-verifier` run was dispatched for it and **had
not returned when this handoff was written**:

- correlation `e3be27d4-4963-4809-89c6-24984b3f6909`, orchestration
  `19851554-1dca-4bc9-bc32-84e1a0f6150a`, at `spawn_verifier` as of 17:34Z.
- read the verdict with:
  ```sql
  SELECT subject_key, left(body,400), created_at FROM doc_notes
  WHERE categories ? 'landmine-verification' ORDER BY created_at DESC LIMIT 3;
  ```
- **If it refutes any part of the entry, correct the entry visibly** (strike-through
  + date), re-run `./scripts/landmines-sync.py --apply`, and log it in
  `WRONG_CALLS.md`. A refutation here is a cheap success, not a problem.

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
7.9–14.4 GB, and fleet-worst headroom has drifted back to **1.73 GB with three of
five nodes now under 2.0 GB**.

**The best remaining move is the cheapest one and nobody has made it yet:** journald
is sitting at its default cap holding ~3.87 GiB on every node measured, and capping
it returns more disk than candidates 1 and 2 could ever address combined. That is
candidate 5, it is one line, and it needs the owner.

> **Claim-shape lesson kept deliberately.** This line has read "one-third done",
> then "one-fifth done", then "done" — the denominator changed from 3 deployments to
> 5 containers under a fraction that never stated it. **A fraction over an unstated
> denominator is the shape to distrust in your own writing.**
