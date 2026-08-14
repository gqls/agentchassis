# SUMMARY — the disk-pressure lane (bugs_open/252), 2026-08-14

Written to be read aloud. The evidence lives in the two files under `bugs_open/252_*`;
this is the story.

## What we're trying to do

Stop the cluster's nodes from running so close to full on disk that they refuse new
pods. On 11 August, during a routine deploy, one node did exactly that: it had
crossed its "too full" line moments earlier, and the pod that the deploy tried to
place on it was turned away. The deploy healed itself — a replacement pod went
elsewhere — but the same event on two nodes at once would leave a service
short-handed, and nothing would alert us. We want the margin between "normal
operation" and "node refuses work" to be wide, and to stay wide on its own.

## Where we've come from

The original question was "was that pod eviction a sign we need more hosts, or
better distribution?" The answer turned out to be neither. The machines are nearly
idle on processor and memory; the problem was that **nobody had ever told the
scheduler that disk exists**. Not one container in the fleet declared how much disk
it needs, so the scheduler placed pods by processor and memory alone — and it put
the two hungriest disk consumers, our two CI build runners, on the same node.

Fixing that (disk requests, then a placement rule) was the first half of the story.
The second half came from actually watching the system for three days. Node disk
does not drain steadily downward — it breathes. Usage climbs as deploys pull fresh
images, then the kubelet's image garbage collector wakes up, deletes old unused
images, and the margin springs back. We measured a node at 0.7 GB from the refusal
line one evening and 9.3 GB clear the next morning, with nobody touching anything.

That rhythm revealed the real defect, and it is embarrassingly simple: **the line
where cleanup starts and the line where the node refuses pods are the same line.**
Cleanup begins at 85% full; refusal begins at 15% free. Those are the same number.
So every cycle, the system rides all the way down to the refusal threshold before
it starts making space — and a deploy that lands in that window gets turned away.
That is what happened on 11 August. Along the way we also mis-called it once
ourselves: we sampled the bottom of the breathing cycle twice, drew a straight
line through it, and briefly reported a collapse that wasn't happening. The
correction is written up in the lane, because the lesson — for anything with a
cleanup cycle, expect oscillation, and re-measure before calling a trend — is
worth more than the embarrassment.

## What we've done

Four things are finished and proven on the live cluster:

1. **Disk is now a declared resource.** The big consumers (the CI runners, the
   model-serving pods) state their disk needs, so the scheduler counts disk when
   placing them. The two heavyweight runners were immediately spread onto
   separate nodes.

2. **That separation is now a rule, not luck.** A placement constraint forbids two
   runner pods on one node while any node has none. We checked it survives every
   deploy: three runner pods, three different nodes, enforced.

3. **Deploys ship all of this automatically.** The release process now applies the
   runner manifests and the node-configuration piece below, so none of it depends
   on someone remembering a manual step.

4. **The cleanup-timing fix is built and ready to deploy.** We tried the polite
   route first — the cluster-wide kubelet configuration object — and discovered
   our hosting provider silently ignores tenant writes to it: the API says
   "patched", and the values are unchanged on the very next read. So the fix ships
   instead as a small self-healing component that runs on every node, adjusts the
   kubelet's cleanup thresholds directly, and re-applies itself whenever a node is
   replaced — which matters, because these are spot machines that come and go.
   It changes two things: cleanup now starts at 70% full instead of 85%, and
   images unused for a week are removed continuously rather than only under
   pressure.

## Where we are now

The immediate danger is over and has stayed over: no pod has been refused since the
11th, and the fleet currently sits 1.2 to 4.3 GB clear of the refusal line,
breathing normally. The cleanup-timing fix is committed and will go live with the
next release. One decision is still open, and it is genuinely the owner's: each
node's system journal quietly holds 3.9 GB — about 70 days of operating-system
logs — because nobody ever capped it. Capping it at half a gigabyte returns
roughly 3.4 GB per node permanently, at the cost of keeping only about ten days
of logs. The machinery to apply that is already in place; it is one small addition
the moment the retention trade is accepted.

## Where we're going

Once the next release rolls, the breathing cycle's low point moves from "touching
the refusal line" to roughly 5 GB clear of it, and stale images stop accumulating
between deploys at all. After a week of normal operation we'll read the margins
again and confirm the cycle's floor sits where the arithmetic says it should. If
the journal cap is approved, the floor rises a further 3.4 GB on every node. The
two remaining ideas on file — per-pod disk ceilings, and simply buying bigger
disks — stay parked: the first would kill legitimate large builds mid-run, and the
second costs money to solve a problem that configuration is already solving.
