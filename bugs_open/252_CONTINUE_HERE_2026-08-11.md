# CONTINUE HERE — the disk-pressure lane (`bugs_open/252`)

**Last updated 2026-08-13 14:15Z.** Cold-start handoff for a new chat. The bug file
(`252_HANDOFF_2026-08-11_disk_is_invisible_to_the_scheduler…`) holds the evidence and
the fix candidates; **this file holds only what is still owed and what will mislead
you.** Read the bug file's newest block — **`🔄 2026-08-13 14:05Z`** — first; it
confirms the root cause by direct observation and **corrects the 20:45Z block below
it**, which is kept struck-through rather than deleted. Everything here assumes it.

**The one-paragraph version.** Disk was invisible to the scheduler; candidate 1 fixed
that and is live. What remains is that **image GC's trigger and the eviction trigger
are the same line**, so the reclaim cycle's trough sits against the eviction threshold
and a roll landing in that trough gets its pod rejected. Every remaining lever is node
config — the owner's. **One thread-side change is built, committed and waiting on an
apply: candidate 1b (`85e8818dd`).**

> **✅ THE MARGIN IS A SAWTOOTH, NOT A DECLINE — AND WE HAVE NOW WATCHED THE TEETH.**
> `[MEASURED 2026-08-13 14:01Z]` Fleet-worst headroom is **2.42 GB** and …1148 sits at
> **9.28 GB**, recovered from **0.73 GB** 17 hours earlier. Image GC fired on exactly
> the three pressured nodes (…1148's cached chassis tags went 9 → 3, its image list
> dropped below the 50 cap, imageFs 16.0 → 10.1 GB) while the two unpressured nodes
> grew instead. `DiskPressure` False fleet-wide, no rejections in the window.
>
> **⛔ This banner previously read "THE MARGIN IS GOING THE WRONG WAY" and that was
> WRONG — I had sampled the trough of a cycle twice and drawn a trend through it.**
> The disconfirming check was to hold still and re-measure. See the bug file's
> `🔄 2026-08-13 14:05Z` block and `WRONG_CALLS.md`.
>
> **The bug is CONFIRMED and is better stated than before:** GC works, but its
> trigger (85% used) *is* the eviction line (15% available), so the cycle's trough
> sits against the eviction threshold — and a roll landing in that trough is exactly
> the admission rejection that opened this lane on 08-11. Nothing is hours from
> failure. **Every remaining lever is still an owner action**, and candidate 3a is now
> evidenced rather than theoretical: its whole effect is to raise the trough.

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

**Candidate 3a — lower `imageGCHighThresholdPercent` to ~70 (low ~60).** Highest
actionable leverage: images are **11.2–17.2 GB per node** (re-measured 20:45Z, up
from 10.5–16.5 at 17:27Z) and are reclaimable in principle, but at 85 the kubelet
cannot start reclaiming until it is already at the
eviction line. Re-verified **20:45Z on all five nodes** as still `85/80` with
`evictionHard.imagefs.available: 15%` and `imageMinimumGCAge: 2m0s` — so GC is held
back **only** by the threshold, not by image age. **This is a kubelet/node-pool
change, i.e. an owner action, not a thread's.**

**Candidate 3b — set `imageMaximumGCAge` (NEW, and it is the better half of 3).**
`[MEASURED 20:45Z, all five nodes]` **`imageMaximumGCAge: 0s` — age-based image GC is
DISABLED fleet-wide.** The 17:30Z block checked `imageMinimumGCAge` and rightly said
age is not holding GC back; that is true and it is not the whole setting. With
`maxAge` at 0 there is exactly **one** reclaim trigger — crossing 85% — and 85% *is*
the eviction line. Setting it to e.g. `168h` retires week-old unused images
**regardless of pressure**, which is the actual shape here: stale tags piling up
between rolls. Needs the `ImageMaximumGCAge` gate — **beta and on by default since
1.30, cluster is 1.31**, so nothing to enable. Composes with 3a; 3b is the one that
stops the margin decaying between rolls. Owner action (kubelet config).

**✅ Candidate 1b is DONE as code — committed `85e8818dd`, awaiting an apply.** Both
runner deployments now carry a shared `workload: gha-runner` pod label and an
identical `maxSkew: 1` / hostname / `DoNotSchedule` / `nodeTaintsPolicy: Honor`
spread constraint. **The shared label is the load-bearing part and this brief got it
wrong** — the two deployments have *different* `app:` values, so the obvious
`app: github-actions-runner`-scoped constraint would have spread that deployment's
replicas and **still allowed the pairing that opened 252** (a runner replica beside
the *vmsites* pod). Verified by parsing the rendered YAML per container, `kubectl
diff -k` clean (`generation` + label + constraint, nothing else), selector untouched,
and a deadlock check — a 4Gi surge pod fits on all five nodes by request accounting.
**Applying rolls three CI runner pods, so it is an owner action:**
```bash
kubectl apply -k deployments/kustomize/services/github-actions-runner/overlays/production/uk_001
kubectl apply -k deployments/kustomize/services/github-actions-runner-vmsites/overlays/production/uk_001
kubectl -n ai-persona-system get pods -o wide -l workload=gha-runner   # must be 3 pods, 3 nodes
```

**Candidate 5 — set `SystemMaxUse=512M` in `/etc/systemd/journald.conf`. NEW, and it
is the cheapest and biggest item on the list.** `[MEASURED 08-12 18:30Z on …1148 and
…1149, root]` journald holds **3.87 and 3.85 GiB** — the largest non-container
consumer on both — because `journald.conf` contains nothing but `[Journal]` and the
default `SystemMaxUse` is *min(10% of filesystem, 4 GiB)*; 10% of 38.6 GiB = 3.86 GiB
and both nodes sit on it to within 10 MiB. Capping at 512M returns **~3.4 GiB per
node** (`[EXTRAPOLATED]` ~17 GiB fleet — three nodes were not opened), against
**current headroom of 0.60–0.79 GB on the three tight nodes** (re-measured 20:45Z;
this line previously read 1.7–2.0 GiB and that is now badly out of date). At tonight's
margin it does not triple the headroom — **it multiplies it by roughly five**, and it
is still one line. Owner's call (node config). Check `journalctl --disk-usage` and
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
8. **`kubectl get nodes -o json` CANNOT census images — `.status.images` is capped at
   `nodeStatusMaxImages` (50 here) and truncates SILENTLY.** The only tell is a node
   reporting *exactly* 50; three of ours do. A repo showing **0 copies on a node may
   simply have fallen off the end**. This caught me mid-session: I read
   "`browser-runner-adapter`: 8 tags on …1148, 0 on …1149" off that field and nearly
   filed per-node tag concentration as a finding. Any tag count from this field is a
   **lower bound**, never a census — including the 17:30Z chassis-tag figure, whose
   conclusion survives but whose number is a bound.
   Also: **summing `sizeBytes` is not disk usage in either direction.** The sum is
   *smaller* than the kubelet's `imageFs` figure on every node (0.32–0.70×) because
   `sizeBytes` is compressed and the snapshotter is unpacked — while separately,
   summing one repo's tags overstates their marginal cost because they share layers.
   **A real per-image census needs the node opened as root** (the 18:30Z method).
   Deliberately not done tonight: it means a `kubectl debug node` pod, and the node
   worth opening (…1148) is 0.73 GB from eviction — pulling a debug image onto it to
   measure it is the wrong trade at this margin. Cheap again once headroom recovers.
9. **One clean roll is not proof.** The 08-12 ~14:55Z chassis roll produced no
   rejection and no `DiskPressure`, which is encouraging and nothing more — the
   original failure was intermittent and no failure rate has been established. The
   `FailedScheduling` on `alertmanager-…-0` is an **unbound PVC**, a different
   mechanism — not ours.

## ✅ The loose end is CLOSED — both verifier verdicts are in, neither refutes anything

`[CHECKED 20:45Z]` Entry 2 (the `du` landmine, corr `253cf06c`) returned
**`NEEDS_HUMAN_REVIEW` at 18:50Z — exactly as predicted below**, and for exactly the
predicted reason: "the entire footprint … lives outside the .go-only code index and
could not be mechanically confirmed or contradicted; the entry is internally
consistent but unverifiable from available evidence" (6 checks, 1 matched, 5 matched
nothing in scope). Entry 1 stands at `STILL_VALID` with 2 of 4 checks NOT ANSWERABLE.

**No refutation, so no entry needs correcting and nothing is owed to `WRONG_CALLS.md`
from this.** Both entries' substantive claims rest on the first-hand cluster
measurement, as the note below already says. The prediction landing is mild evidence
the model of the verifier is right: **for a landmine about DB behaviour, shell tools
or kubectl, this verifier confirms vocabulary and nothing more.** Do not re-dispatch.

## The original note on that loose end (kept — its reasoning is the reusable part)

Two LANDMINES entries were filed from this lane on 2026-08-12, both synced to
`doc_notes`:

1. *"An `orchestration_states` status count is not a measure of cluster load"*
   (`122cb945c`) — trap 5 above. Verifier **returned `STILL_VALID`** (corr
   `e3be27d4`, 17:34Z).
2. *"A `du` total is only a total if you could READ every subtree"* (`6aa6a57d7`) —
   trap 7 above. Verifier dispatched, corr
   `253cf06c-f9b6-4de2-9e3b-71bf183823a7`, **not returned when this was written**.

```sql
SELECT subject_key, left(body,600), created_at FROM doc_notes
WHERE categories ? 'landmine-verification' ORDER BY created_at DESC LIMIT 3;
```

> **⚠ Do not read `STILL_VALID` here as "the claim was checked" — it was not, and the
> verdict body says so if you read it.** `landmine-verifier` runs against a **Go-only
> code index** (7,302 `.go` symbols). For entry 1 it confirmed the *tables* exist and
> are queried, and reported the actual claims — the UPPERCASE status values, the
> `updated_at` staleness — as **"NOT ANSWERABLE by this index" (2 of 4 checks)**. It
> also describes indexed commit `46b507ed` from **2026-08-11 17:49Z**, "the last
> pushed tip, not the present tree". So for a landmine about **DB behaviour, shell
> tools or kubectl**, this verifier can confirm that the nouns exist and nothing more.
> Entry 2 is about `du`, `kubectl debug` and file permissions and should be expected
> to come back `NEEDS_HUMAN_REVIEW` for the same reason. **Both entries' substantive
> claims were measured first-hand against the live cluster in this session — that is
> the real evidence; the verifier is corroboration of vocabulary, not of mechanism.**
> **If either verdict does refute something, correct the entry visibly**
> (strike-through + date), re-run `./scripts/landmines-sync.py --apply`, and log it in
> `WRONG_CALLS.md`. A refutation is a cheap success, not a problem.

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

**Candidate 1 is live and proven. Candidate 1b is committed (`85e8818dd`) and awaiting
an apply** — it is a kustomize change, so the 08-13 chassis roll did **not** deliver
it even though the deployed commit contains it. Between them the fleet requests disk
and the two big consumers would be separated by a rule rather than by luck. **That is
the whole of what a thread can do here, and the lane is not fixed.** 252 stays OPEN.

**The bug is confirmed, and the 08-13 measurement states it better than the filing
did:** image GC works, but `imageGCHighThresholdPercent: 85` *is* the
`imagefs.available < 15%` eviction line, so the reclaim cycle's **trough sits against
the eviction threshold**. A roll landing in that trough is the admission rejection of
08-11. Not a slow drain to failure — a recurring window of exposure.

**Every remaining lever is node configuration, i.e. the owner's**, and the two
cheapest are each one line:

- **Candidate 3a — `imageGCHighThresholdPercent` 85 → ~70.** Now evidenced rather
  than theoretical: its entire effect is to **raise the trough** so the cycle never
  approaches the eviction line. This is the fix for the mechanism as understood.
- **Candidate 5 — `SystemMaxUse=512M`.** Journald's ~3.87 GiB per node is **never
  touched by image GC** — a permanent tax that lowers the whole sawtooth. Freeing it
  raises every point in the cycle. Unaffected by the correction above.

Neither has been made. **The margin is healthy at this instant (2.42–9.28 GB) and
that is not the point** — it was 0.60 GB seventeen hours ago and will be again.

> **Claim-shape lesson kept deliberately.** This line has read "one-third done",
> then "one-fifth done", then "done" — the denominator changed from 3 deployments to
> 5 containers under a fraction that never stated it. **A fraction over an unstated
> denominator is the shape to distrust in your own writing.**
