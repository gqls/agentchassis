# 252 — disk is invisible to the scheduler, so nodes sit ~1–3 GB from the eviction line and reject pods mid-roll

**Filed:** 2026-08-11 by the `bugfix_236_site_availability` lane, after a chassis
pod was rejected during the 09:23Z roll with
`Pod was rejected: The node had condition: [DiskPressure]`, and the owner asked
whether this was a host-count or a distribution problem.
**Status:** OPEN, UNOWNED. **Severity:** medium — self-healing today, but it
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

## Fix candidates, ordered by what closes the door

1. **Declare `ephemeral-storage` requests on the heavy tenants** (the three
   `github-actions-runner` deployments, `ollama-*`, `browser-runner-adapter`).
   This is the one that makes the bad state *unrepresentable*: once disk is a
   scheduled resource, the scheduler stops co-locating two 2.3 GB CI runners, and
   it stops placing new pods on a node that has no disk left. Cheap, per-workload,
   no cluster-wide change.
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
