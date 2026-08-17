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
  `css-patch-agent` **has still never processed a single work item** (measured 2026-08-12:
  0 complete, 0 failed, `attempt_count = 0` across its entire history), so promotion sends
  225 items to a handler with no track record at all. If they should be *fixed* rather than
  *retracted*, that is a separate decision needing its own evidence.
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
