# 252 — disk is invisible to the scheduler, so nodes sit ~1–3 GB from the eviction line and reject pods mid-roll

**Filed:** 2026-08-11 by the `bugfix_236_site_availability` lane, after a chassis
pod was rejected during the 09:23Z roll with
`Pod was rejected: The node had condition: [DiskPressure]`, and the owner asked
whether this was a host-count or a distribution problem.
**Status:** **OPEN, UNOWNED — candidate 1 APPLIED, LIVE AND PROVEN 2026-08-12 13:00Z
(census 0 → 5).** Still open because the root blindness remains: `ephemeral-storage`
is scheduled against 35.1 GB allocatable while real free space is 2.8–9.3 GB, so the
scheduler can still place onto a nearly-full node. Outstanding: **1b** (make the
runner spread a rule, not an outcome — see the 13:00Z block), and **2–4** (owner
decisions). **Severity:** medium — self-healing today, but it
silently removes a replica during a roll and the margin is thin enough that it
recurs several times a day.
**Answer to the question that prompted it: neither.** It is not too few hosts, and
the distribution is not "wrong" so much as **unguided** — nothing in the fleet
tells the scheduler that disk exists.

## The measurement, all `[MEASURED 2026-08-11 12:25Z]`

```
CPU   actual   1–5%    requested 29–63%      <- enormous headroom
MEM   actual   7–42%   requested  6–42%      <- enormous headroom
DISK  used    62–82%   requested   0%        <- and 0% is not a typo
```

**0 of 143 running containers declare an `ephemeral-storage` request or limit.**

```
kubectl get pods -A -o json | python3 -c "…"    # full one-liner in the RUNBOOK below
  running containers: 143
  declaring an ephemeral-storage REQUEST (what the scheduler uses): 0
  declaring an ephemeral-storage LIMIT   (what caps a runaway pod):  0
```

Per-node headroom above the **hard** eviction line (`imagefs.available < 15%`,
and imagefs shares the single 41.4 GB filesystem with nodefs, so 15% = 6.2 GB):

| node | available | % | headroom above eviction |
|---|---|---|---|
| …1148 | 15.9 GB | 38.4% | 9.68 GB |
| …1149 | 11.4 GB | 27.6% | 5.23 GB |
| **…6832** | **7.6 GB** | **18.2%** | **1.34 GB** |
| …6833 | 9.6 GB | 23.1% | 3.36 GB |
| …1336 | 8.5 GB | 20.4% | 2.24 GB |

It fires, repeatedly, and it is not historical — live event, 24 minutes before
this file was written:

```
Warning EvictionThresholdMet  node/prod-instance-17722135234001149  Attempting to reclaim ephemeral-storage
Normal  NodeHasDiskPressure   node/prod-instance-17722135234001149
Normal  NodeHasNoDiskPressure node/prod-instance-17722135234001149   (5 minutes later)
```

`DiskPressure` transitions today: …1148 at 09:28:41Z, …6832 at 09:29:20Z,
…1149 at 12:08:30Z.

## Root cause — three things compounding, in the order that matters

**1. `ephemeral-storage` is not a scheduled resource here, so placement cannot be
correct by construction.** The Kubernetes scheduler only accounts for disk if pods
*request* it. With 0/143 requesting, it balances CPU and memory and is blind to
disk. That is why both GitHub Actions runner pods — the two largest ephemeral
consumers in the fleet at 2.3 GB and 2.3 GB — are on the **same** node (…1149),
and a third (2.8 GB) sits with `ollama-adapter` on …1336. Nothing did that wrong;
nothing could have done it right.

**2. Image GC fires at exactly the eviction threshold, so it never gets to work
first.** From the live kubelet config:

```
evictionHard: {"imagefs.available":"15%", "nodefs.available":"10%", …}
imageGCHighThresholdPercent: 85     <-- 85% used == 15% available == the eviction line
imageGCLowThresholdPercent:  80
```

Reclaim is supposed to be the cheap defence that runs *before* pods are harmed. At
85/15 the kubelet starts garbage-collecting images and starts refusing/evicting
pods **at the same instant**. There is no reclaim headroom by design.

**3. The disks are small for the workload.** 41.4 GB per node, of which cached
images are already 10.0–15.6 GB (24–37%). The heavy tenants are not ours-per-se:
`ollama:latest` 3.25 GB, `cephcsi` 1.02 GB, `browser-runner-adapter` 0.87 GB per
tag, `github-actions-runner` 0.60 GB per tag.

> **REFUTED, and worth recording because it was my first theory:** chassis image
> churn is NOT the cause. `IMAGE_TAG` is bumped for every build, so I expected a
> pile of stale chassis tags. Measured: the chassis image is **0.09 GB**, and the
> worst node caches **9** tags = 0.8 GB. The tag churn this repo worries about
> costs under a gigabyte fleet-wide. Do not "fix" this by pruning chassis tags.

## What actually happened to the chassis pod — and it is NOT an eviction

The message was `Pod was rejected: The node had condition: [DiskPressure]`. That
is the kubelet **refusing to admit a newly-scheduled pod**, not the kubelet killing
a running one to reclaim space. The distinction matters because it changes the
fix:

1. a node crosses `imagefs.available < 15%` (a CI build, or image pulls during a roll);
2. the kubelet reports `DiskPressure`, which taints the node;
3. a rolling deploy — the 09:23Z chassis roll — places a new pod there anyway,
   because **the scheduler does not know about disk** and the taint had not yet
   propagated to it;
4. the kubelet on arrival sees the condition and **rejects** the pod;
5. the ReplicaSet makes a replacement elsewhere and the roll completes.

So the observable impact today is a **transient loss of one replica during a
roll**, self-healed. The reason it is worth filing anyway: with …6832 at 1.34 GB
of headroom, the same event on two nodes at once during a roll would leave a
deployment short for as long as the pressure lasts, and nothing alerts on it.

## 🔎 STATE AS OF 2026-08-12 17:30Z — **THE DISK BUDGET, MEASURED BY CATEGORY. IT RE-RANKS THE CANDIDATES.**

Taken after a fresh chassis build rolled (pods `agent-chassis-6588556967-{msvvv,wp74f}`,
created ~14:55Z on …1149 and …6833).

**The roll went clean.** No `Pod was rejected`, no `DiskPressure`, no
`EvictionThresholdMet` in the event window — the failure that opened this lane did
not recur on the first roll after candidate 1. **This is weak evidence and must not
be quoted as proof.** The original was intermittent, one roll is a sample of one,
and I did not establish the failure rate this sample could detect. (The one
`FailedScheduling` present is `alertmanager-…-0`, an **unbound PVC** — a different
mechanism, not ours, do not chase it.)

### The measurement nobody in this lane had taken `[MEASURED 17:27Z]`

Every node is 41.4 GB. `node.fs` and `node.runtime.imageFs` report **identical
capacity and available**, so they are one filesystem and the subtraction below is
valid (checked, not assumed):

| node | images | pod writable layers | **other** | free | above evict |
|---|---|---|---|---|---|
| …1148 | 15.1 | 0.44 | **17.7** | 8.2 | 2.00 |
| …1149 | 16.5 | 2.63 | **14.3** | 7.9 | **1.73** |
| …6832 | 10.5 | 0.44 | **16.0** | 14.4 | 8.22 |
| …6833 | 15.9 | 2.72 | **14.8** | 8.0 | **1.80** |
| …1336 | 14.6 | 0.48 | **16.8** | 9.5 | 3.27 |

**Total pod writable-layer usage across the WHOLE FLEET is 6.77 GB, over 138 pods.**
Per node that is 0.44–2.72 GB. Meanwhile images are 10.5–16.5 GB per node, and
**"other" — neither images nor pod writable layers — is 14.3–17.7 GB per node, the
largest single category, 35–43% of every disk.**

**What this does to the fix candidates:**

- **Candidates 1 and 2 (requests, limits) govern the SMALLEST category.** Requests
  and limits can only ever act on the 0.44–2.72 GB of pod writable layers. Candidate
  1 was still right and still worth doing — it is what separates the two 2.3 GB
  runners — but it can never be the answer to "the node is 1.7 GB from eviction".
- **Candidate 3 (image GC) governs the middle category** and is now clearly the
  highest-leverage *actionable* item: 10.5–16.5 GB per node is reclaimable in
  principle, and at `imageGCHighThresholdPercent: 85` reclaim still cannot start
  until the eviction line is already reached. **Re-verified 17:30Z, unchanged:**
  `85/80`, `evictionHard.imagefs.available: 15%`, and `imageMinimumGCAge: 2m0s` —
  so GC is *not* being held back by image age. It is held back only by the threshold.
- **The largest category is addressed by NO candidate on file.** `[UNVERIFIED]` what
  "other" is composed of — plausibly the node OS, container logs under
  `/var/log/pods`, and containerd metadata, but **I did not open a node to look** and
  nobody should repeat this as fact until someone does. **This is now the top open
  question in this lane**, because if ~15 GB per node is structurally unavailable
  then a 41.4 GB disk has only ~26 GB of usable budget, and candidate 4 (grow the
  disks) is much stronger than "buys margin, closes no door" allowed.

### Headroom is degrading again, and the distribution has tightened

Fleet-worst headroom: **1.34 GB** at filing (08-11 12:25Z) → **0.82 GB** (08-12
12:37Z) → **2.81 GB** after candidate 1 (13:00Z) → **1.73 GB** now (17:27Z).

The more telling change is the spread. At filing, one node was under 2.24 GB. Now
**three of five are under 2.0 GB** (…1149 1.73, …6833 1.80, …1148 2.00). Candidate 1
redistributed the pod layers, but consumption fleet-wide is outpacing reclaim, which
is exactly the shape candidate 3 exists to fix.

### Candidate 1's sizing was right, and it is doing its job

Actual usage of the tenants given requests: `github-actions-runner-…-nrssq` **2.33 GB**
against a 4Gi request, `…-vmsites-…-tp44p` **2.15 GB** against 4Gi. They are the top
two ephemeral consumers in the fleet by an order of magnitude (third place is
0.216 GB). And they are now on **different nodes** — before, both sat on …1149,
i.e. ~4.5 GB on the node that is today at 1.73 GB of headroom.

> **Chassis tag churn is still NOT the problem — the filing's refutation holds at a
> second measurement.** 8 distinct chassis tags cached fleet-wide (`v1.0.1283`
> through `v1.0.1291`), 0.09 GB each, **2.1 GB total across all five nodes**. Do not
> "fix" this by pruning chassis tags.

## ✅ STATE AS OF 2026-08-12 13:00Z — **CANDIDATE 1 IS APPLIED, LIVE AND PROVEN. CENSUS = 5.**

The owner ran the apply at 12:55Z. All three deployments rolled successfully;
`ollama-eval` came back at 12:58:36Z and is serving. **Supersedes every state block
below.**

```
CENSUS (containers requesting ephemeral-storage) = 5     [MEASURED 2026-08-12 13:00Z]
  github-actions-runner-cbb64d544-l4tqc          runner      4Gi   node …1336
  github-actions-runner-cbb64d544-nrssq          runner      4Gi   node …1149
  github-actions-runner-vmsites-6f897c7d8f-tp44p runner      4Gi   node …6833
  ollama-adapter-745d456cff-cbqrn                ollama      1Gi   node …1336
  ollama-eval-5bff8c445f-tqrq5                   ollama      1Gi   node …6832
  (+ ollama-adapter model-pull initContainer, 1Gi)
```

**The co-location broke up, which is the specific thing candidate 1 was for.**
Before: `…-9n5p7` **and** `…-vmsites` were both on …1149, the node with 0.82 GB of
headroom. After: all three runner pods are on three different nodes — …1336, …1149,
…6833.

**And the tightest node recovered** `[MEASURED before 12:37Z → after 13:00Z]`:

| node | before | after | Δ | note |
|---|---|---|---|---|
| **…1149** | **0.82 GB** | **3.43 GB** | **+2.61** | vmsites' writable layer left — attributable |
| …6833 | 7.82 GB | 2.81 GB | −5.01 | took vmsites, pulled the runner image — attributable |
| …6832 | 4.18 GB | 9.26 GB | +5.08 | [UNATTRIBUTED] plausibly the old eval pod's layer freed, not measured |
| …1336 | 3.67 GB | 4.68 GB | +1.01 | [UNATTRIBUTED] |
| …1148 | 3.63 GB | 3.08 GB | −0.55 | [UNATTRIBUTED] |

**Fleet-worst headroom went 0.82 GB → 2.81 GB, a 3.4× improvement in the number
that actually causes the bug.** Note the fleet did not gain disk — the improvement
is redistribution, exactly as the 15:00Z correction predicted.

`kubectl describe node` now discriminates, where the filing warned it could not:
allocated `ephemeral-storage` reads 4Gi (12%) / 1Gi (3%) / 4Gi (12%) / 5Gi (15%)
across four nodes, and `0 (0%)` on …1148 only. Disk is a scheduled resource here
for the first time.

### ⚠ NEW FINDING — the separation is NOT guaranteed, and candidate 1 does not make it so

`[MEASURED 13:00Z]` `github-actions-runner` has **`affinity: None`** and
**`topologySpreadConstraints: None`**. The scheduler spread the three runners by
score, not by rule. With 35.1 GB allocatable per node and 4Gi requests, putting
both runner replicas back on one node is **entirely representable** — the next roll
may legally undo today's placement, and nothing would report it.

So the correct sentence is: *candidate 1 made disk visible and today's placement is
good*, **not** *the co-location cannot recur*. This is the same trap the 15:00Z
correction caught, one level up — requests fixed the input to the decision, they did
not constrain the decision.

**Candidate 1b (new, unowned): add `topologySpreadConstraints` (or a
`podAntiAffinity`) on the two runner deployments** so the spread is a rule rather
than an outcome. Cheap, manifest-only, closes the door that candidate 1 only leans
shut. Ranked above candidates 2–4 by the "what makes the bad state unrepresentable"
test, and unlike 3 and 4 it needs no owner-level node-pool change or spend.

## ⚠ STATE AS OF 2026-08-12 12:40Z — CENSUS IS **1**, AND THE APPLY IS PERMISSION-BLOCKED

**Supersedes the 18:00Z block below for state; that block's *reasoning* still stands.**

**(a) `ollama-adapter` is now fully live, both containers.** The census below returns
**1** rather than 0 — the 18:00Z app-container fix rolled overnight (pod
`ollama-adapter-745d456cff-cbqrn`, age 14h at time of writing) and the live pod
carries `ephemeral-storage: 1Gi` on the `ollama` container *and* on the
`model-pull` initContainer. `kubectl diff -k` on that overlay shows **no drift**,
so it needs no apply. Remaining to reach 5: the two `github-actions-runner`
replicas, `github-actions-runner-vmsites`, `ollama-eval`.

**(b) The margin got worse, on the node that matters** `[MEASURED 2026-08-12 12:37Z]`:

| node | available | % | headroom above evict | vs 08-11 12:25Z |
|---|---|---|---|---|
| …1148 | 9.8 GB | 23.8% | 3.63 GB | −6.05 GB |
| **…1149** | **7.0 GB** | **17.0%** | **0.82 GB** | −4.41 GB |
| …6832 | 10.4 GB | 25.1% | 4.18 GB | +2.84 GB |
| …6833 | 14.0 GB | 33.9% | 7.82 GB | +4.46 GB |
| …1336 | 9.9 GB | 23.9% | 3.67 GB | +1.43 GB |

**…1149 at 0.82 GB is tighter than the worst node in the original filing** (…6832
at 1.34 GB), and it is hosting **both** `github-actions-runner-…-9n5p7` *and*
`github-actions-runner-vmsites` — precisely the co-location candidate 1 exists to
break up, now sitting on the least headroom in the fleet. No `DiskPressure` or
`EvictionThresholdMet` events in the current event window.

**(c) The 15:05Z "cluster is busy" deferral no longer applies** `[MEASURED 12:37Z]`:
runners idle (jobs are ~20 s; last completed 12:35:15Z, all `Succeeded`), **1** LLM
call in the preceding 15 minutes (vs 27 on 08-11), and **0** `/api/generate|chat|embed`
requests to either ollama pod in 2 hours — only 30-second `/api/tags` health polls.

**(d) The apply was attempted at 12:38Z and was REFUSED by the local Claude Code
permission classifier**, not by the cluster. Nothing was changed. This is the only
reason candidate 1 is still one-fifth done — it is no longer a judgement call about
load. It needs the owner to run the three commands, or to grant the permission.

> **Pre-apply check worth keeping — `kubectl diff -k` answers the actual worry.**
> The 15:05Z block reasons at length that the overlays pin the running images and
> replica counts, verified by hand. `kubectl diff -k <overlay>` proves it in one
> command and is disconfirmable: for all three pending deployments the *entire*
> diff is `generation: N→N+1` plus one `+ ephemeral-storage:` line. No image
> change, no replica change, nothing else. Run it immediately before applying —
> another session may have moved the overlay since.

> **CHECK REFINEMENT — an in-flight orchestration count needs an AGE filter, or it
> overstates load by ~20×.** The 15:05Z deferral rested partly on "38 in-flight
> orchestrations". Today the same shape of query returns 48 rows in
> `RUNNING`/`EXECUTING_STEP`/`INITIALIZED` — of which **46 have not been touched in
> over 2 hours** and only **2** are active in the last 15 minutes. Bucket by
> `updated_at` before reading a status count as current load:
> ```sql
> SELECT CASE WHEN updated_at > now() - interval '15 minutes' THEN 'active'
>             WHEN updated_at > now() - interval '2 hours' THEN 'recent'
>             ELSE 'stale' END, count(*)
> FROM orchestration_states
> WHERE status IN ('RUNNING','EXECUTING_STEP','INITIALIZED') GROUP BY 1;
> ```
> Statuses are **UPPERCASE** in this table — a lowercase `NOT IN ('completed',…)`
> filter matches nothing and silently returns every row. [UNVERIFIED] whether those
> 46 stale rows are the hung-spawn defect of `bugs_open/029`; not investigated here,
> and not this lane's to take.

## ⚠ STATE AS OF 2026-08-11 18:00Z — PARTLY APPLIED, AND ONE CONTAINER WAS WRONG

**Supersedes the 15:05Z block below.** Two things changed.

**(a) A release at 17:13Z rolled `ollama-adapter` and picked the change up** — the
live `model-pull` initContainer now carries `ephemeral-storage: 1Gi`. So this does
land through the ordinary release path, as predicted. The runners and `ollama-eval`
were NOT rolled and still read `** NOT APPLIED **`; their pods date from 07-16,
07-21 and 06-28, so they are waiting for a roll that touches them.

**(b) The 1Gi on `ollama-adapter` went on the WRONG CONTAINER, and I did not catch
it for three hours.** `ollama-adapter/base/deployment.yaml` has two near-identical
`resources:` blocks — the `model-pull` initContainer at ~line 51 and the `ollama`
app container at ~line 85 — and I patched the first believing it was the second.
Fixed 18:00Z; the app container now has its own request.

> **Why it stayed invisible, which is the transferable part.** The pod's
> *effective* request is `max(sum of app containers, max of initContainers)`, so
> with 1Gi on the init and 0 on the app the pod still requested 1Gi. **The number
> was right and the placement was wrong**, so every pod-level check agreed with me.
> It still mattered: the kubelet ranks disk evictions by *usage above request*, and
> the long-running container — the one that would actually be evicted — had no
> request of its own.
>
> **My validation could not have caught it.** I ran
> `kubectl kustomize <overlay> | grep -c ephemeral-storage` and got `1` for every
> service, and read four 1s as four successes. A count cannot say WHICH container
> holds the field. The check that found it, and the one to use:
>
> ```bash
> kubectl kustomize <overlay> | python3 -c "
> import sys,yaml
> for doc in yaml.safe_load_all(sys.stdin):
>     if not doc or doc.get('kind')!='Deployment': continue
>     sp=doc['spec']['template']['spec']
>     for key in ('initContainers','containers'):
>         for c in sp.get(key) or []:
>             r=(c.get('resources') or {}).get('requests') or {}
>             print(key, c['name'], r.get('ephemeral-storage','** MISSING **'))"
> ```
>
> `ollama-eval`'s initContainer is still without one, deliberately: it pulls into
> the PVC, and the app container's 1Gi already sets the pod's effective request.

### The 15:05Z block, kept for the reasoning

## ⚠ SUPERSEDED 15:05Z — COMMITTED BUT **NOT APPLIED**

`301161274` adds `ephemeral-storage` requests to both runner deployments (4Gi)
and both ollama pods (1Gi). **Those manifests are inert.** A resource change to a
Deployment's pod template only takes effect when the pods are recreated, and
nothing has recreated them.

**Why I did not apply it, and this is a judgement to re-make, not a blocker:**
applying restarts the pods, and at 15:05Z the cluster was busy with other
sessions' work — **38 in-flight orchestrations**, 27 LLM calls in the preceding 15
minutes, and all three runner pods had completed CI `deploy` jobs within the last
two minutes. Restarting `ollama-adapter` takes the eval path down for a few
minutes (`OLLAMA_LOAD_TIMEOUT` is 10m and the models reload from the PVC), and
restarting the runners kills whatever CI job is in flight. Many sessions share
this cluster and none of them asked for that.

**It does not need a special action.** The four overlays pin exactly the images
and replica counts that are already running (verified 2026-08-11: runners
`v1.0.948` / `v1.0.1126`, ollama `latest`, replicas 2/1/1/1), so the next ordinary
roll of these deployments picks the requests up with no drift risk — which also
means this is safe to include in a whole-fleet release rather than needing a
one-service apply.

**To apply deliberately at a quiet moment:**

```bash
for d in github-actions-runner github-actions-runner-vmsites ollama-adapter ollama-eval; do
  kubectl apply -k deployments/kustomize/services/$d/overlays/production/uk_001
done
# then PROVE it, at the pod and not at the manifest:
kubectl get pods -A -o json | python3 -c "
import json,sys;d=json.load(sys.stdin)
print([(p['metadata']['name'][:40], ((c.get('resources') or {}).get('requests') or {}).get('ephemeral-storage'))
       for p in d['items'] if p['status'].get('phase')=='Running'
       for c in p['spec']['containers']
       if any(k in p['metadata']['name'] for k in ('github-actions-runner','ollama'))])"
```

The fleet-wide census is the real check — it must stop being 0:

```bash
kubectl get pods -A -o json | python3 -c "
import json,sys;d=json.load(sys.stdin)
print('containers requesting ephemeral-storage:', sum(1 for p in d['items'] if p['status'].get('phase') in ('Running','Pending') for c in p['spec']['containers'] if ((c.get('resources') or {}).get('requests') or {}).get('ephemeral-storage')))"
```

Expect **5** once applied (2 runner replicas + 1 vmsites + ollama-adapter +
ollama-eval), not 4 — `github-actions-runner` runs two replicas.

## Fix candidates, ordered by what closes the door

> **⚠ RE-RANKED 2026-08-12 17:30Z by the per-category measurement above — read that
> block before acting on this list.** The order below was written when "images plus
> system" was one undifferentiated 21–27 GB lump. Split, it becomes: pod writable
> layers 0.44–2.72 GB/node (what 1 and 2 govern), images 10.5–16.5 GB/node (what 3
> governs), and **"other" 14.3–17.7 GB/node, which no candidate here addresses.**
> Working order now: **3** (highest actionable leverage) → **1b** (cheap, closes the
> co-location door properly) → **4** (now stronger than stated below) → **2**
> (smallest category, and it can evict a live build). **1 is DONE.**

1. **Declare `ephemeral-storage` requests on the heavy tenants** (the two
   `github-actions-runner` deployments, `ollama-*`, `browser-runner-adapter`).
   **DONE 2026-08-11** for the runners (4Gi each) and both ollama pods (1Gi).

   > **CORRECTED 2026-08-11 15:00Z, by the session that wrote this line.** It
   > first said this "makes the bad state *unrepresentable* … it stops placing new
   > pods on a node that has no disk left." **The first half is right and the
   > second half is wrong**, and the difference is worth more than the fix.
   >
   > `ephemeral-storage` requests are scheduled against the node's **allocatable**
   > figure, not against real free space. Measured 2026-08-11 15:00Z:
   >
   > | node | REAL free | what the scheduler counts |
   > |---|---|---|
   > | …1148 | 14.1 GB | 35.1 GB allocatable |
   > | …1149 | 10.3 GB | 35.1 GB |
   > | …6832 | 13.2 GB | 35.1 GB |
   > | …6833 | **7.6 GB** | 35.1 GB |
   > | …1336 | **7.8 GB** | 35.1 GB |
   >
   > The ~21–27 GB gap is cached images plus system usage, and **it is charged to
   > no pod's request**, so the scheduler cannot see it. Requests therefore
   > bin-pack against a figure that is 3–4× the truth.
   >
   > **What candidate 1 actually buys:** the two big writable-layer consumers stop
   > being co-located, which is the specific thing that happened. **What it does
   > NOT buy:** the scheduler will still cheerfully place a pod on …6833 at 7.6 GB
   > free, because it believes that node has ~31 GB spare. The only defences
   > against *that* are reactive (the `DiskPressure` taint, which is what rejected
   > the chassis pod in the first place) or candidates 3 and 4. **Do not let this
   > file be read as "fixed by requests".**
2. **Declare `ephemeral-storage` limits on the same tenants.** Requests fix
   placement; limits fix blame. Today a runaway build is capped by nothing, and
   the kubelet's disk-eviction ranking is *usage above request* — with every
   request at 0 that ranking is arbitrary between equals.
3. **Move `imageGCHighThresholdPercent` down to ~70** (low ~60) so reclaim has
   real headroom before the eviction line at 15% available. This is a kubelet
   config change and therefore a node-pool-level action, i.e. the owner's, not a
   thread's.
4. **Grow the disks** from 41.4 GB. Buys margin, closes no door — and is the only
   candidate here that costs money. Worth doing *with* 1–3, never instead of.
5. **Do NOT reach for "add more hosts."** CPU sits at 1–5% and memory at 7–42%.
   A sixth node adds 41 GB of disk and five more nodes' worth of image cache to
   maintain; it does not make disk visible to the scheduler.

## How to verify a fix

```bash
# 1. disk becomes a scheduled resource (the number that must stop being 0)
kubectl get pods -A -o json | python3 -c "
import json,sys;d=json.load(sys.stdin);r=sum(1 for p in d['items'] if p['status'].get('phase') in ('Running','Pending') for c in p['spec']['containers'] if ((c.get('resources') or {}).get('requests') or {}).get('ephemeral-storage')); print('containers requesting ephemeral-storage:', r)"

# 2. headroom above the hard eviction line, per node
for n in $(kubectl get nodes -o name | cut -d/ -f2); do
  kubectl get --raw "/api/v1/nodes/$n/proxy/stats/summary" | python3 -c "
import json,sys;d=json.load(sys.stdin);fs=d['node']['fs']
print(d['node']['nodeName'][-4:], round((fs['availableBytes']-0.15*fs['capacityBytes'])/1e9,2),'GB above evict')"
done

# 3. the condition stops recurring
kubectl get events -A --sort-by=.lastTimestamp | grep -iE "diskpressure|EvictionThresholdMet"
```

**Do not verify with `kubectl describe node` alone** — it prints `ephemeral-storage
0 (0%)` in Allocated resources whether nothing requests it or nothing exists, and
those read identically.

## Provenance of this filing

**This asserts a structural, cross-cutting cause and did NOT go through the `090`
diagnosis loop** — stated plainly per the owner ruling of 2026-07-31 rather than
silently omitted. The reason: `090` reads the repo's Go code and the clients
database, and this defect lives in neither. It is Kubernetes pod specs and live
kubelet configuration, which that loop cannot see. Substituted first-hand
verification instead, all of it disconfirmable and reproduced above: the kubelet
`configz` for the thresholds, the kubelet `stats/summary` for real filesystem
usage on every node, `kubectl get events` for a live firing, and a fleet-wide
census for the 0/143. My first hypothesis (chassis tag churn) was **refuted** by
that same census, which is the evidence that these numbers could have come out
otherwise.
