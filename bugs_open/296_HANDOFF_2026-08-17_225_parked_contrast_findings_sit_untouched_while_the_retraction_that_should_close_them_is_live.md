# 296 — 225 parked `contrast_failure` findings sit untouched while the retraction that should close them is live, wired, and running daily

**Filed** 2026-08-17 by the brochure contrast front (`bugs_open/113` lane), after
`bugs_open/122` was **closed** on the owner's ruling of 2026-08-15 and its parked backlog
turned out to have no owner.
**Severity** medium. Nothing is *breaking*; the risk is the opposite — **225 real, measured
contrast failures are invisible in every "is the fleet healthy" reading**, because
`deferred` is not a status anybody's dashboard or sweep looks at (`bugs_open/083`).
**Class** ownership + possible silent no-op. **Not yet a diagnosed root cause — see §4.**

**START HERE IF YOU ARE PICKING THIS UP.** §4 is the whole job: three candidate
explanations, ranked, each with a cheap check. Do those before writing any code. It is
entirely possible the answer is "(a), and they will drain on their own" — in which case
this file closes with a measurement and no fix at all.

> **ANSWERED 2026-08-17 (later the same day) — §8 is the answer; read it before §4.**
> **The answer is (a) AND (b), and it is NOT (c): there is no defect in the retraction.**
> (a) explains the six days of stillness — *no site was due for re-audit in that window at
> all*. (b) is the substantive truth and it is the bad news: **all four parked pairings
> re-measured in a live browser still fail at exactly their parked ratio**, so when the
> rotation does reach these sites the retraction will correctly close **approximately none
> of them**. The park is honest and **will not drain**. What this file asks for is
> therefore not a mechanism — it is an **owner for 225 live contrast defects**.
> The retraction mechanism itself is confirmed live and correct; it has simply never yet
> had a single opportunity to run.

> **CORRECTED 2026-08-18 — the prediction above was WRONG, and §9 is now the current state.**
> The rotation swept all 14 remaining sites overnight and the retraction closed **40** of the
> parked rows, not "approximately none". The park **is** draining — just not all of it. What
> survives (185) has now been re-measured by the machine and correctly declined, so the
> "needs an owner" half stands and is if anything better evidenced. **Two claims in this file
> are now false and are struck through where they appear: §8.3's "approximately none", and
> §5's "css-patch-agent has never processed a single work item" (it has now processed 58,
> completed all 58, and I verified two at the live page).** Owner decision doc:
> `brochure_component_library/DECISION_INPUT_2026-08-18_the_186_durable_contrast_failures.md`.

---

## 1. What is actually true, measured 2026-08-17

```sql
SELECT status, count(*), max(updated_at) FROM site_work_items
WHERE item_type='contrast_failure' GROUP BY 1;
--  deferred  | 225 | 2026-08-11 12:31:22
--  cancelled |   1 | 2026-08-14 16:36:54
```

**225 rows, untouched for six days. Zero have ever completed.** `attempt_count = 0` — they
were parked before any handler saw them, never attempted.

| domain | parked |
|---|---|
| vonc.com | 38 |
| robot-hands.com | 33 |
| idea.uk | 27 |
| mortgagecalculator.co.uk | 22 |
| lendzy.co.uk | 18 |
| ai-agent-orchestration.com | 17 |
| dartsonline.com | 17 |
| finetuning.uk | 11 |
| *(+ others)* | |

## 2. Why they are parked — a deliberate decision, not a bug

Migration `389` (owner decision 2026-08-11) parked them so that `improvement-sweep` could be
re-enabled for the page re-renders without dragging contrast items through
`css-patch-agent`, where `bugs_open/213`'s false-complete defect was then unfixed.
**Promoting 220 of them would have converted an honest backlog into 220 false closures.**
The park was correct and its reasoning still reads well.

**But the park was always conditional on someone doing the follow-up work, and
`bugs_open/122` closed without doing it.** That is the gap this file exists to name.

## 3. Everything the park was waiting for now EXISTS — which is what makes this odd

Each verified in code today, not inferred:

- **The retraction mechanism is built and SHARED.** `platform/orchestration/actions/work_item_retraction.go`
  — audit-path retraction, extracted as a shared helper after `write_render_audit_findings`
  became the second producer to hand-roll it (council `a43b63d6…`, `architecture` seat).
- **It is wired for this exact item type.** `retractResolvedContrastFindings`, referenced
  throughout `write_render_audit_findings_action.go`.
- **It deliberately reaches PARKED rows.** Its own header, point 3: *"`deferred` is NOT in
  `workItemClosedStatuses`, so a retraction closes PARKED items too — a stated decision
  (WII-016), not a side effect."* Confirmed at source: `workItemClosedStatuses`
  (`work_items_common.go:83-89`) is `complete, verified, rejected, wont_fix, cancelled`, and
  the candidate loader is `WHERE … status NOT IN (…)`. **`deferred` rows are candidates.**
- **The precondition the 122 lane named is MET.** `b2fca2f8f` said the blocker was that the
  audit *"reports how many pages it covered, not WHICH"* and *"needs `pages_audited`
  identities in the adapter summary"*. `bugs_open/242` shipped exactly that — live on
  v1.0.1288, behaviourally proven — and `pages_audited` is now read at
  `write_render_audit_findings_action.go:169,423`.
- **The driver is running.** `site-render-audit-rotation` is `enabled`, and last triggered
  **2026-08-17 11:09** (today).

**So: a live, wired retraction that is designed to drain the park, a met precondition, and a
running driver — and 225 rows that have not moved in six days.** That is the anomaly.

## 4. THE JOB — three candidates, ranked, each with its cheap check

**I have NOT diagnosed this.** What follows is honest ranking, not a finding. Do (a) first;
it is the cheapest and, on the evidence, the most likely.

**(a) The rotation simply has not re-observed those pages yet.** It is a *rotation* — it
covers a slice per pass, and 122's ink fix only landed 2026-08-15. If so there is **no bug**:
the rows drain as the rotation reaches them, and this file closes with a measurement.
> **Check:** find the rotation's recent runs and the `pages_audited` identities they carried,
> and intersect with the pages the 225 rows name. If the intersection is empty, it is (a).
> `SELECT … FROM site_work_items WHERE item_type='contrast_failure' AND status='deferred'`
> gives the page paths; the audit's stored summary gives `pages_audited`.
> **Positive control required:** find at least ONE contrast row that DID get retracted, or
> you cannot tell "not yet observed" from "observation happens and retracts nothing".
> If no such row exists anywhere, that is itself the finding.

**(b) The defects are still genuinely present, so retraction correctly declines.** 122's fix
was **article-body ink** — one line, 97 placements. The 225 rows span many selectors and 8+
sites. Most may still be failing for reasons 122 never addressed (the primary-as-ink family,
`features_open/026`). If so, **there is still no bug here** — but there IS a real,
unaddressed contrast population, and it needs an owner. That is the outcome I would bet on
second, and it is the one that matters most for the fleet.
> **Check:** take three rows from three different sites and re-measure the named selector on
> the live page — `scripts/render_audit.py <url>` and read the selector, or
> `scripts/probe_reveal_open_state.py` if the state needs interaction.
> **Do not** conclude from the stored row; it is six days old.

**(c) The retraction runs but excludes them for a reason the header does not describe.**
Candidates: `observed` arriving false (the header's point 1 — an unavailable observation is
inert by design); the `batch_id IS DISTINCT FROM $batch` guard at
`write_render_audit_findings_action.go:204`; or the caller's `decide` function not treating
a parked row as retractable even though the loader returns it.
> **Check:** read `retractResolvedContrastFindings`' `decide` and the `observed` argument's
> source, then look for the retraction's own log line on a recent rotation run. **A silent
> no-op and a correct decline look identical from the outside** — that is this whole file's
> shape, and the reason (c) is worth ruling out explicitly rather than assuming.

## 5. What NOT to do

- **Do not promote the 225 to `triaged`.** That is precisely what migration `389` prevented.
  ~~`css-patch-agent` **has still never processed a single work item** (measured 2026-08-12:
  0 complete, 0 failed, `attempt_count = 0` across its entire history), so promotion sends
  225 items to a handler with no track record at all.~~
  > **CORRECTED 2026-08-18 — this is the load-bearing sentence of the whole park, and it is
  > now FALSE.** `css-patch-agent` has processed **58** `contrast_failure` items since
  > 2026-08-17 17:13, **completed all 58, failed none**, writing real CSS
  > (`p.p { color: #4a4a40; }`). I verified two at the artefact, not at the status:
  > `noted.co.uk/index.html` lost both patched pairings and `noted.co.uk/contact.html`
  > measures **0 failures**. Combined with `bugs_open/213` being closed, **both halves of the
  > park's premise have expired.** Promotion is now a graded loop — the fixer proposes, the
  > render audit independently re-measures and only withdraws what actually got fixed. That
  > does NOT make bulk promotion automatically right (see the class-A caveat in the decision
  > doc), but it is no longer "a handler with no track record".
  If they should be *fixed* rather than *retracted*, that is a separate decision needing its
  own evidence — **and it is now live as an owner decision**:
  `brochure_component_library/DECISION_INPUT_2026-08-18_the_186_durable_contrast_failures.md`.
- **Do not write a second retraction.** One exists and is shared, and the council already
  told a lane not to copy-paste a third hand-rolled version.
- **Do not read "122 closed" as "contrast is fixed".** 122 fixed article-body ink across 97
  placements — genuinely — and that is a subset of what these 225 rows describe.

## 6. Related files, and how they differ

- **`bugs_open/083`** — the CLASS: discovery findings written as `detected` never reach a
  handler; the promoter runs only inside a task disabled since May. It already measured
  *"467 rows across 6 statuses"* unreachable by the sweep or its coverage report. **This file
  is one instance of that class, with a specific mechanism that should already cover it.**
  If (c) turns out to be the answer, contribute the finding to 083 as well.
- **`bugs_closed/122`** — the history: where these rows came from and why they were parked.
  Its `b2fca2f8f` carries the costing of the retraction fork and is worth reading before
  designing anything.
- **`bugs_open/242`** — the precondition, DONE: `pages_audited` identities.
- **`bugs_open/213`** — CLOSED: the false-complete defect that justified the park.
- **`features_open/026`** — the primary-as-ink family, the most likely content of (b).

## 7. Method note / substitution declared (owner ruling 2026-07-31)

**No `090` run, and this file deliberately asserts no root cause** — §4 poses a question and
ranks candidates rather than naming a mechanism, so there is no structural claim to
substantiate. Everything stated as fact in §1 and §3 is first-hand and re-runnable: live DB
counts and timestamps, the retraction header quoted verbatim, `workItemClosedStatuses` read
at source, `pages_audited` grepped at its call sites, and the scheduler row read directly.
**If the answer to §4 turns out to be (c) — a live mechanism silently declining — that IS a
structural claim and it should go through `090` before it is written up as one.**


## 8. THE ANSWER — measured 2026-08-17 16:20–16:35 UTC. It is (a) AND (b), not (c)

Every figure below is first-hand and re-runnable; the query or command is given with it.
Hypotheses I formed and then **refuted** are kept in §8.5, because they are the part a
re-reader cannot rederive.

### 8.1 (a) is CONFIRMED, and it fully explains the six days

Not one of the 16 sites holding parked rows has been re-audited since the sweep that filed
them. `reaudited_since_park` is **false for all 16**:

```sql
SELECT s.domain, count(*) FILTER (WHERE w.status='deferred') AS parked,
       r.last_selected_at,
       (r.last_selected_at > '2026-08-11 12:31:22+00') AS reaudited_since_park
FROM site_work_items w JOIN sites s ON s.id=w.site_id
LEFT JOIN site_discovery_rotation r ON r.site_id=s.id AND r.agent_type='render-audit-agent'
WHERE w.item_type='contrast_failure' AND w.status='deferred'
GROUP BY s.domain, r.last_selected_at ORDER BY parked DESC;
--  all 16 rows: last_selected_at ∈ 2026-08-10 14:54 … 2026-08-11 12:04, reaudited = f
```

**Why not, and why that is correct rather than broken.** Read the driver's own config
(`SELECT interval_seconds, pre_query FROM scheduled_tasks WHERE name='site-render-audit-rotation'`):
it fires **hourly**, takes **`LIMIT 1` site per fire**, and a site is eligible only when its
stamp is **older than 7 days**. There are **23 eligible sites and 0 never-audited**. All 23
were swept 08-10/08-11, so **between 2026-08-11 12:04 and 2026-08-17 14:54 not one site was
due** — the rotation was correctly idle. The first parked-row site (`robot-hands.com`,
stamped 08-10 14:54) came due **today at 14:54**. Six days of stillness is precisely what
this design predicts; it is not evidence of anything failing.

### 8.2 The positive control §4(a) demanded does not exist — and that is arithmetic, not a finding

**Zero `contrast_failure` rows have ever been retracted, anywhere, ever.** A retraction sets
`status='complete'` (`resolveWorkItems`, `work_items_common.go:401-415`); the type has only
`deferred` (225) and `cancelled` (1), and that one cancelled row carries `result = {}`, so it
was not closed by the retraction path either.

§4(a) said "if no such row exists anywhere, that is itself the finding." **It is not, and
here is why the check as written could not have come out any other way:**

| event | when |
|---|---|
| last render audit of any site holding contrast findings | **2026-08-11 12:04** |
| retraction code committed (`5639a1103`) | **2026-08-12 20:03** |

Code cannot run before it is committed. **Every audit that could ever have retracted these
rows predates the retraction's existence by at least a day.** The zero is not a silent
no-op; the mechanism has never had one opportunity. The only render audit since is today's
`remortgagecalculator.uk` run, which **skipped** (zero deployed pages — `bugs_open/299`) and
so correctly retracted nothing.

**It is live now**, so the forward-looking claim holds — binary probe on the running chassis
(`v1.0.1305`, pod `agent-chassis-5bd56bdd9b-6sb8t`), **with both controls**:

```bash
kubectl -n ai-persona-system exec $POD -- grep -ac "is no longer below its contrast threshold" /proc/1/exe  # 1  (must be present)
kubectl -n ai-persona-system exec $POD -- grep -ac "retraction_scope_pages" /proc/1/exe                     # 1  (must be present)
kubectl -n ai-persona-system exec $POD -- grep -ac "zzz_not_a_real_symbol" /proc/1/exe                      # 0  (must be absent)
```

### 8.3 (b) is CONFIRMED — and this is the finding that matters

Four parked pairings, four different sites, **re-measured today in a real headless browser**
(`python3 scripts/render_audit.py <url>` — a live render, not the stored row). **All four
still fail, at exactly the parked ratio:**

| site / page | parked selector | parked | **live today** |
|---|---|---|---|
| dartsonline.com `/about.html` | `H3.H3` | 1.06:1 | **1.06:1** ✗ |
| finetuning.uk `/about.html` | `A.cta-btn` | 1.00:1 | **1.00:1** ✗ |
| idea.uk `/about.html` | `P.info-card-grid__subtitle` | 3.35:1 | **3.35:1** ✗ |
| vonc.com `/about.html` | `A.gauntlet-btn-primary` | 1.76:1 | **1.76:1** ✗ |

And on every one of those four pages the **live failure count now EXCEEDS the parked count**
— dartsonline 6 vs 2, finetuning 4 vs 3, idea.uk 11 vs 2, vonc 23 vs 8. (Expected: the filed
set was always a subset of the measured set — the `max_items` cap and the locked-component
skip both drop measured-and-failing findings, which is exactly what
`retractResolvedContrastFindings`' point (2) is built to respect.)

**So: when the rotation reaches these sites over the coming week, the retraction will
~~correctly decline on approximately all 225 rows.~~ **CORRECTED 2026-08-18: it declined on
185 and RETRACTED 40** — see §9.** The park is honest. **It will not drain FULLY,
and nothing is wrong with the machinery that would drain it.** `bugs_closed/122` fixed
article-body ink across 97 placements and that repair is real — it is simply a different
population from these, most of which look like the primary-as-ink family
(`features_open/026`): note vonc's `rgb(124,60,255)` on `rgb(252,92,125)` and idea.uk's
`rgb(131,124,114)` on `rgb(232,223,204)`.

### 8.4 What this file actually needs — and it is not a mechanism

The three "do not do this" items in §5 all still stand, and §8.3 removes the last reason to
consider a fourth. **What is left is a decision, not an engineering task:** 225 measured,
still-reproducing, WCAG-AA contrast failures across 16 live sites, sitting in a status no
dashboard reads (`bugs_open/083`), which no mechanism will ever clear because **they are
true**. They need an owner or an explicit "we accept these" ruling. Contribute the
`deferred`-is-invisible half to `083`; the defect population itself belongs with
`features_open/026`.

### 8.5 Refuted along the way — recorded so nobody re-walks them

- **[REFUTED] "The retraction is silently declining" (candidate (c)).** Not needed to
  explain anything, and disproved by §8.2's dates. I read `decide`, the `observed`
  argument's source and the `batch_id` guard at source before testing it — all correct.
- **[REFUTED] "Sites flip to `building` during a build, so the rotation's
  `status IN ('active','deployed')` predicate excludes them."** Plausible — `building` *is*
  a valid status (`v3_site_actions.go:360-363`) — but **no live workflow ever sets it**:
  the only three `update_site_status` steps in active `agent_definitions` all set
  `deployed`. Query in §8.6.
- **[REFUTED, and the reasoning was the interesting part] "A claimed build item blocked the
  rotation at the two ticks that found nothing."** My retrospective test
  (`claimed_at <= t AND updated_at >= t`) returned 0 for both ticks — **but that test is
  structurally blind to the case it was looking for.** A row that is *currently* claimed has
  `updated_at == claimed_at`, so a claim held across `t` fails `updated_at >= t` and is
  counted as absent. A reconstruction of "was it claimed then" from `updated_at` cannot see a
  held claim. I only caught this by watching live rows. **Do not reconstruct claim state from
  `updated_at`.**

### 8.6 A real secondary observation, measured, not asserted

The rotation's `NOT EXISTS (… status='claimed' AND pipeline='build')` guard is **sampled once
per hour** against a value that flips on a **~20-second** timescale. Polled read-only every
20s while two due sites were mid-build: **9 of 14 samples had EVERY due site blocked**, i.e.
the pre-query would have returned nothing. `undeployed_asset` items hold a claim for ~25s of
every ~26s during a burst.

**This is a delay, not starvation** — a skipped site keeps its old stamp, stays due, and is
picked on the first lucky tick (expected wait a few hours). **[SAMPLE: 14 × 20s during an
active build burst — this is NOT a steady-state rate, and I have not measured one.]** Worth
knowing before anyone reads "the rotation didn't pick it up this hour" as a defect. Queries:

```sql
-- which statuses live config actually sets (refutes the `building` hypothesis)
SELECT DISTINCT ad.type, s.value->'config'->>'status'
FROM agent_definitions ad,
     LATERAL jsonb_path_query(ad.default_config,'$.**.steps') AS steps,
     LATERAL jsonb_each(steps) AS s(key,value)
WHERE s.value->>'action'='update_site_status'
  AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
--  pageflow-builder | deployed ; rerender-pages | deployed ; site-work-orchestrator | deployed
```

To run the selector without stamping anything (the pre_query contains a data-modifying CTE,
so `BEGIN … ROLLBACK` is mandatory — a bare run **advances the rotation and costs a site its
turn**):

```bash
{ echo "BEGIN;"; psql -t -A -c "SELECT pre_query FROM scheduled_tasks WHERE name='site-render-audit-rotation';"; echo ";"; echo "ROLLBACK;"; } | psql
```

## 9. CURRENT STATE — measured 2026-08-18 18:12 UTC. The park is draining; 185 are durable

The rotation swept **all 14** remaining affected sites between 2026-08-17 17:12 and
2026-08-18 15:22. This section supersedes §8's forward-looking claims.

```sql
SELECT status,
       count(*) FILTER (WHERE created_at <  '2026-08-12') AS original_park,
       count(*) FILTER (WHERE created_at >= '2026-08-12') AS filed_since,
       count(*) FILTER (WHERE result ? 'resolved_by')     AS by_retraction
FROM site_work_items WHERE item_type='contrast_failure' GROUP BY 1;
--  deferred  | 185 |  0 |  0
--  complete  |  40 | 58 | 40
--  cancelled |   1 |  0 |  0
```

**Of the original 226: 40 retracted, 185 still failing, 1 cancelled.**

### 9.1 The retraction behaved exactly as specified

Every audit run shows `retracted == retracted_parked`, i.e. **every closure drained a parked
row** — read from `collected_data->'findings_written'` on the runs still in
`orchestration_states`. Per-site: dartsonline 14/17, lendzy 13/18, cookly 7/7, idea.uk 4/27,
vonc 1/38. Worked verification: `dartsonline.com/about.html` measured **6 contrast failures
on 08-17 and 1 on 08-18**, with the 1.06:1 heading family gone — so the closure was correct
and the page was genuinely repaired.

### 9.2 Why my §8.3 prediction was wrong, and what would have caught it

I predicted "approximately none". **The mechanism I missed was sitting in my own evidence.**
I measured the `page_rerender` churn on these very sites — and used it *only* to explain why
the rotation's `NOT EXISTS (claimed build item)` guard kept skipping ticks. It never occurred
to me that the same traffic was **repairing the pages the audit was about to re-measure**.
The cheap check: **before predicting a queue will not drain, ask what is currently writing to
the artefact its closure depends on.** Logged in `WRONG_CALLS.md`.

### 9.3 What survives is durable, and this is now well evidenced

**All 185 remaining rows sit on sites the audit re-measured in this pass and deliberately did
NOT retract.** They are not unobserved; they were re-observed and still fail. Nor can a
re-render save them — every affected site has been re-rendered heavily since 08-15 (vonc 44,
robot-hands 93, ai-agent-orchestration 150, webdesign 956), and a re-render reproduces the
same hard-coded pairing by construction.

| severity | rows | sites | colour families |
|---|---|---|---|
| A. **invisible** (≤1.2:1) | 60 | 7 | 18 |
| B. severely unreadable (1.2–3.0) | 40 | 6 | 19 |
| C. fails AA, readable (3.0–4.0) | 59 | 8 | 18 |
| D. marginal (4.0–4.5) | 26 | 5 | 9 |

65 colour families across 14 sites. Dominant shapes: `robot-hands` white-on-white `A.cta-btn`
(1.00:1, `bugs_open/113`'s family); `vonc` `rgba(255,255,255,0.7)` on brand purple (19 rows,
`features_open/026`); `idea.uk` `rgb(131,124,114)` on cream (21 rows).

### 9.4 The decision this file exists to force is now WITH THE OWNER

`brochure_component_library/DECISION_INPUT_2026-08-18_the_186_durable_contrast_failures.md`
— four costed options, recommending release of classes C+D (85 rows) to the now-working
fixer while holding class A, on the ground that patching *text* colour where the *background*
failed to render yields a passing ticket and a still-wrong button.

**This file stays OPEN** until that decision lands: the mechanism question is answered, the
ownership question — which was always the point — is not.

### 9.5 Two audit blind spots, recorded and NOT fixed

- **`pool` / `system` sites are never audited.** The selector takes `active`/`deployed` only:
  today 23 deployed + 2 active vs **17 `pool` + 1 `system`**, which hold 2 deployed pages
  between them. Live blind spot is 2 pages; the gap is unguarded and silent.
  **`building` is a non-issue** — a valid enum value (`v3_site_actions.go:360-363`) that no
  live workflow writes, so no site is ever in it (all three `update_site_status` steps set
  `deployed`).
- **`[UNMEASURED]` page coverage on large sites.** Observed `retraction_scope_pages` was 5–24
  while `webdesign.co.uk` has 108 deployed pages. The run I needed was already pruned from
  `orchestration_states`, so I could **not** read its `truncated`/`pages_total` and assert
  nothing either way. Verify on the next `webdesign.co.uk` audit.

### 9.6 Latency reduced

The 7-day re-audit window was the confirmation bottleneck (a repair waits up to a week to be
graded). Cut to **3 days** by migration `469` on the owner's instruction. Three not two: the
selector competes with build work and was measured turned away on **9 of 14** samples during
a burst, so a 2-day window would sit at real throughput and start slipping turns silently.
