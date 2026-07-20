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

---

## Verification pass, 2026-07-20 (bugfix-006 thread)

All three re-checked against the live cluster and the current tree. **A confirmed and worse, B real
but its stated mechanism is STALE, C's open question answered.** Evidence inline below.

### A — CONFIRMED, degraded further

```
$ kubectl -n ai-persona-system get pods | grep runner
github-actions-runner-5c44ddb44d-5pqdv           1/1  Running            0     22d
github-actions-runner-5c44ddb44d-lhg9l           0/1  CrashLoopBackOff   6365  23d
github-actions-runner-vmsites-5bf4b47c57-zdtjb   1/1  Running            0     4d1h
```
Restarts have gone **4906 → 6365** since filing (18d → 23d). Diagnosis unchanged; still a
single-point-of-failure on the fleet deploy path. Needs node-level access (cordon/drain or fix that
node's containerd `SystemdCgroup`), so it is an owner/infra action, not a code change.

Note: a **third** runner deployment now exists (`github-actions-runner-vmsites`, healthy, 4d) that
did not exist when this was filed.

### B — > **CORRECTED 2026-07-20: real, but not for the reason stated. Nothing posts to `/contact`.**

The defect (contact forms on live sites silently discard submissions) is **confirmed and fleet-wide**.
The stated *cause* no longer holds, so the fix options as written aim at a string that is not there.

Measured across every live form component:
```sql
SELECT substring(rendered_html from 'action="([^"]*)"'), count(*)
FROM page_components WHERE rendered_html ILIKE '%<form%' GROUP BY 1;
-- (empty) 12 | #contact 9 | /audience-check 1 | /request 1  →  action="/contact": ZERO
```
The component template is **not** hardcoded. `k8s/bk_content_components.sql:134` renders
`action="{{.form_action}}"`, and `form_action` comes from each component's own `content_data`.
**No Go code sets `form_action` at all** (`grep -rn 'form_action' --include='*.go' platform/ internal/`
returns only an unrelated file-header comment) — so there is no default and no validation. Whatever
the content LLM happens to emit is what ships:

| form_action | sites |
|---|---|
| `#contact` | ai-agent-orchestration, dartsonline, gaswholesalers, leopardess, relojistas, robot-hands, vetcomparison, vonc |
| `""` (empty) | finetuning.uk, leopardess, robot-hands |
| `mailto:idea.uk@contactforsales.com` | idea.uk (hand-fixed) |

`#contact` and `""` both POST to the current URL. On a static host that is a 405/404 — the message is
lost and the visitor gets no error they can act on. Only 6 of 24 form components have any JS submit
handler, and those are the tool calculators, not the contact forms — so these submissions are real
navigations, not intercepted.

**So: 10 of 11 live sites have a contact form that cannot deliver a message.** Severity as filed is
right; `apply_gap_plan_action.go:465` and `k8s/bk_page_components.sql:140` are the wrong fix sites.
The real fix is a **default + a validation check on `form_action`** (an unset/`#`/empty action on a
`contact-form` should fail a discovery check), not building a `/contact` backend.

**On idea.uk's mailto — staged, and deliberately so.** Its `content_data.form_action` is the mailto
while its `rendered_html` still carries `action="#contact"`.

> **CORRECTED 2026-07-20, same thread, within the hour:** I first wrote this up as an unnoticed
> instance of the "content_data edits do not hold until re-render" landmine. It is not. The
> idea.uk thread applied it source-only **on purpose** and documented why
> (`idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §Q, 2026-07-17): it publishes on the next
> contact-page build, and forcing one was rejected because (a) a single-page rebuild of an
> already-`deployed` page bounces to `needs_human_review` at attempt 0, and (b) the site was not
> live, so there was no user-visible gap. **Caught by:** searching the docs before building the fix,
> at the owner's prompting — the RUNNING_NOTES entry was four days old and directly on point.
> The lesson is the one this file keeps relearning: *check whether a thread already decided this*
> before reporting their deliberate decision as your discovery.

**Two things this search settled, which change the fix rather than just the wording:**

1. **The approach is already an owner decision.** Owner chose **"Convert form → mailto"**
   on 2026-07-17 (§Q). So the fleet default should be a `mailto:` built from the site's own contact
   address — the pattern already proven on idea.uk — not an invented convention and not a new backend.
2. **This bug is the sanctioned owner of the fleet fix.** §Q ends: *"the fleet-wide dead-form fix
   stays its own thread, `aaa_fails_to_mend/006 §B`"* — i.e. here. The idea.uk thread scoped itself
   to one site on purpose. No concurrent thread holds this.

**Operational constraint the same notes record, which the fleet fix must respect:** a single-page
rebuild of an already-`deployed` page bounces to `needs_human_review` at attempt 0. So a Go default
alone will NOT repair the 11 existing sites — it fixes newly-rendered components only. Remediating
what is already deployed is a separate, deliberate step (a data migration over `content_data` plus a
render, through the review gate), and should be costed as such rather than assumed to follow.

### C — the open question is ANSWERED: the sweep is a scheduled task, not Go

The addendum grepped `platform/` and concluded "not in the Go tree" — correct, but it is not absent,
it is **in the database**. `scheduled_tasks` row **`claimed-item-timeout`**, `interval_seconds=120`,
enabled, firing normally (last trigger was seconds before this check). Its `pre_query` has **two
branches with different windows**:

1. **auto-complete on evidence — 15 minutes.** Marks the item `complete` where the targeted artifact
   is provably written (`error = 'Auto-completed: work verified done despite lost response'`).
2. **reset — 40 minutes.** `claimed → triaged` (or `failed` at max attempts) with
   `'Claim timed out — handler pod likely died'`.

**This fully explains the 16-minute stall.** The reset window is **40 minutes**, not 10 — so at 16
minutes nothing was due to happen yet. The addendum's inference ("its window is at minimum longer
than 16 minutes") was right; the figure is 40. The operator-recovery SQL above uses a 10-minute
predicate, which is *more aggressive than the live reaper by design* — fine for the
nothing-was-done case it is scoped to, but it is not what the platform would have done on its own.

**NEW, and the part still worth fixing: the evidence branch covers 3 item types out of 18.** It tests
only `needs_content_page`, `page_rerender` and `needs_design`. Every other item type falls through to
the 40-minute reset **even when its work succeeded** — which is exactly the churn this section
describes. Measured over 14 days:

| item_type | timed out | auto-completed |
|---|---|---|
| `page_rerender` | 64 | 30 |
| `needs_content_page` | 0 | 6 |
| **`needs_page`** | **42** | **0** |
| `phantom_internal_link` | 23 | 0 |
| `content_rewrite` | 7 | 0 |
| *(11 further types)* | 15 | 0 |

`needs_page` — **the exact item type in this section's original symptom** — has 42 timeouts and has
never once been auto-completed, because no branch knows what artifact proves a `needs_page` done.
So C is not fixed; it is fixed for three item types and live for the other fifteen. The structural
fix this section asks for (complete atomically, or retry the completion write idempotently) is still
the right one, and would make the per-item-type evidence branches unnecessary rather than needing
fifteen more of them.

### Found while verifying C — filed separately as `bugs_open/048`

Four `scheduled_tasks` have not fired in **79 days** while `enabled=true`, because a fifth task in
their concurrency group consumes the group's only slot on every tick and then does nothing. Not a
duplicate of `029` (that is hung orchestrations holding real slots). See `048`.
