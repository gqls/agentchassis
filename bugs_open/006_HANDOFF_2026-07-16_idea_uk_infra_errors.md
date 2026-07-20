# HANDOFF — three errors surfaced during idea.uk VM-site work (route each to its own chat)

**Created 2026-07-16 from the `idea_uk_vm_site` workstream.** Each section is an INDEPENDENT
error, self-contained, handable to a separate chat in any order. All three were found while doing
other work and deliberately left unfixed. **None is an outage.** Parent workstream docs:
`docs024_key_docs_latest/idea_uk_vm_site/` (PLAN / RUNBOOK / RUNNING_NOTES). idea.uk site_id =
`1244516d-014d-421c-88c6-090bb1e9552a`; DB access
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

---

## A. Self-hosted GitHub Actions runner: one replica crash-looping (lost redundancy)

**Severity: latent single-point-of-failure. NOT an outage — the sibling replica is healthy, so
site→B2 deploys still work.**

### Symptom
`kubectl -n ai-persona-system get pods` shows two `github-actions-runner-5c44ddb44d-*` replicas:
- `…-5pqdv` — `1/1 Running`, healthy, 17d.
- `…-lhg9l` — `0/1 CrashLoopBackOff`, **4906 restarts over 18d**.

### Root cause (node-level container runtime, not the app)
The crashing pod's container never starts. `lastState.terminated` (exitCode 128, StartError):
```
failed to create containerd task: … OCI runtime create failed: runc create failed:
expected cgroupsPath to be of format "slice:prefix:name" for systemd cgroups,
got "/kubepods/burstable/pod…/…" instead
```
i.e. a **cgroup-driver mismatch on the node `lhg9l` is scheduled onto** (kubelet/containerd disagree:
systemd vs cgroupfs). It is node-specific — the healthy replica landed on a good node.

### Why it matters
The runner IS the site deploy path: "commit is deploy" — the runner runs `b2 sync` per commit to
`b2://portfolio-sites/<domain>`. Redundancy is gone: **if `5pqdv` dies, `lhg9l` cannot take over on
its bad node, and fleet-wide B2 deploys stop.** Also 4906 restarts = constant churn/log noise.

### Fix (infra)
Reschedule `lhg9l` off the bad node (cordon/drain or a nodeAffinity/anti-affinity so both replicas
land on nodes with a consistent cgroup driver), or fix that node's containerd `SystemdCgroup`
setting to match the kubelet. Then confirm both replicas reach `1/1 Running`.

### Note
This reinforces the idea.uk migration rationale: moving idea.uk to VM-pull (RUNBOOK §3a) takes it off
this B2-via-runner path entirely.

---

## B. Generated contact forms POST to a dead endpoint (`/contact`) — fleet-wide

**Severity: silent data loss on every generated site with a contact form. Latent; pre-existing;
found while planning the idea.uk cutover.**

### The defect
`apply_gap_plan_action.go:465` emits a `contact-form` section whose stored component HTML is
`<form class="contact-form" action="/contact" method="POST">` (`k8s/bk_page_components.sql:140`).
Generated sites are **static** (B2 / VM nginx serving files). **There is no `/contact` backend
anywhere in the chassis** — no handler, no submissions table. So every submission on every generated
site's contact form is silently lost (405/404 or nothing), fleet-wide. idea.uk has a deployed
`/contact.html` with exactly this form.

### Why it matters for idea.uk specifically
After the VM cutover (RUNBOOK §3), nginx proxies only the tool's reserved paths to `:8080`; `/contact`
is not one of them, so the static `/contact.html` form stays dead. The tool's real intake is
`/request` (now hardened — see the parent RUNBOOK §4 / task).

### Fix options (decide per scope)
1. **Chassis-wide:** give the generated contact-form a real backend, OR change the generated form to a
   `mailto:` / a link to a working intake, so it is never born dead. Whatever backend is built should
   ship with the honeypot + rate limit the idea.uk `/request` handler now has (a ready prototype).
2. **idea.uk-only (quick):** repoint `/contact.html`'s form at the tool's `/request` (or make it a
   `mailto:idea.uk@contactforsales.com`) before/at cutover so the live contact page works.

### Where
`apply_gap_plan_action.go:465`, `k8s/bk_page_components.sql:140`; idea.uk `contact` page
(`pages` where name='contact').

---

## C. `page-build-handler` claim-timeout churn: work done, item not marked complete

**Severity: ergonomics / wasted compute. NOT an outage — the work succeeds. Pre-existing hygiene item
also listed in the main handoff backlog.**

### Symptom
A `needs_page` item for idea.uk `index` **rendered and deployed successfully** (verified: the
tool-list card flipped to `/audience-check`, page redeployed), but the work item stayed `claimed`,
then reverted with `error = "Claim timed out — handler pod likely died"` and `attempt_count`
incremented — so the dispatch loop **re-claimed and redundantly re-rendered the already-correct
page**. Left alone it burns handler pods on repeat work and can eventually exhaust `max_attempts`
(→ `unresolved`) despite the artefact being correct. Observed repeatedly 2026-07-14/15 across idea.uk
page builds; the `page-build-handler` pods are short-lived and some die mid-run.

### The pattern (claim-timeout-AFTER-success)
The handler does the work — renders, commits, marks the page `deployed` — but the **pod dies before
marking the work item `complete`**. The claim-timeout sweep then reverts `claimed → triaged` and it
re-runs. So "died" is misleading: the page build succeeded; only the item-status write was lost.

### What to determine (I did not)
Why `page-build-handler` pods die near the end of a build (OOM? liveness probe killing a
still-working pod? node eviction? — connects to error A's node issues and to
`003_HANDOFF_spawn_lost_child_response.md`'s Kafka-broker-2 node network path). And structurally:
the item should be marked `complete` **atomically with / before** the last deploy step, or the
completion write should be idempotently retried, so a late pod death doesn't orphan a finished build.

### Workaround used
Manually set the verifiably-complete item to `complete` (the render was confirmed correct in
`page_components` + deployed HTML) to stop the redundant re-renders. Not a fix — the next build can hit
the same churn.

### Where
`page-build-handler` workflow / the claim-timeout reaper (find via
`grep -rn "Claim timed out" --include='*.go'`); dispatch/claim logic in
`claim_work_item_action.go`. Related: main backlog "claim-timeout retries on long builds";
`002_HANDOFF…errors_to_fix.md` §A / `003_HANDOFF_spawn_lost_child_response.md` (node-level pod deaths).

---

## Priorities (suggested)
1. **A** — infra, quick, removes a real single-point-of-failure on the whole fleet's deploy path.
2. **C** — connects to the broader pod-death investigation (003); the atomic-completion fix is
   self-contained and stops wasted rebuilds.
3. **B** — decide chassis-wide vs idea.uk-only; the idea.uk-only repoint is a 5-minute change to make
   the live contact page work at cutover.

### C addendum, 2026-07-20 — the same class, but STALLED rather than churning

A `page_rerender` item for idea.uk `tools` sat at `status='claimed'`,
`claimed_by='build-dispatch-loop'`, for **16 minutes** with `attempt_count=0`, `handled_by` empty
and `result='{}'`. Its sibling item (`index`, created in the same INSERT, same second) was claimed
and completed in ~90 seconds. Nothing appeared in the chassis logs for it.

**What differs from the churn described above:** there, the work succeeded and only the status write
was lost, and the sweep reverted `claimed → triaged` so it re-ran. Here the work never started and
the item did **not** revert within 16 minutes. `ClaimWorkItemAction` (`claim_work_item_action.go:96-105`)
sets `claimed_at = NOW()`, but a grep of `platform/` for a requeue predicate on `claimed_at`
(`claimed_at <`, `stale`, `claim_timeout`) returns **nothing** — so whatever performs the sweep this
file describes is not in the Go tree, and its window is at minimum longer than 16 minutes. Left
alone, a claim that dies before doing any work appears to stall indefinitely.

Fleet check at the time: exactly **one** item was `claimed` and older than 10 minutes — this one. So
it is not a systemic pile-up; it is a single lost claim with no visible recovery path.

**Operator recovery that worked** (unblocked it; the item then completed in under a minute):
```sql
UPDATE site_work_items
SET status='triaged', claimed_by=NULL, claimed_at=NULL, updated_at=now()
WHERE id='<item>' AND status='claimed' AND claimed_at < now() - interval '10 minutes';
```
Safe when `attempt_count=0`, `result='{}'` and no handler log line exists — i.e. nothing was done, so
there is nothing to duplicate. Do **not** apply it blind to the churn case above, where the work
*has* succeeded: there the correct action is to mark the item `complete`, not to requeue it.

**Open question for whoever takes C:** where does the claim-timeout sweep actually live, and what is
its window? The two observed behaviours (revert-after-success vs stall-with-no-work) may have
different causes and should not be assumed to be one bug.
