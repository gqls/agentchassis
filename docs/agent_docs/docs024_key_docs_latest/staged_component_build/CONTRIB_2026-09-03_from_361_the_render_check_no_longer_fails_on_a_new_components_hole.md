# CONTRIB 2026-09-03 → `staged_component_build` (and any lane that MINTS components)

**From** `bugfix_361_render_check_ratchet`. Not a bug in your work — a **guarantee that
narrowed today and lands squarely on the component birth path, which is yours.**

## What changed

`component-render-check` (CGV-030) is the daily job that renders every active component with
each referenced field removed and flags the ones that produce a visible hole. It had been
**RED for 25 consecutive days** (`lastSuccessfulTime` 2026-08-09 → 2026-09-03) and nobody was
reading it, because its ratchet compared findings key-by-key against a baseline cut once on
**2026-08-04**. Every component born since owned no baseline key, so **everything it found read
as "NEW"** — 478 by this morning, against 227 when `bugs_open/361` was filed. The check was
truthful throughout; the ratchet was measuring **library growth**.

Fixed in tree today (`051c73d1e`, `d716c837a`), riding the next fleet release. It now fails on
a **REGRESSION** — a finding in a component the baseline **covered** — and the baseline records
what it *vouched for*, not only what it *found*.

## The bit that is yours

**A hole in a component born AFTER the baseline now fails NOTHING.** That is a deliberate,
stated cost — a regression ratchet cannot be the thing that gates new debt — but I want to be
straight that **it is an open gap, not a delegation.** `bugs_open/361` says the debt "belongs to
birth-time gating (CGV-029)". I checked: **CGV-029 does not cover it.** `component-fallback-check`
sees only fields declared `on_missing:"skip_field"`, which is precisely the blindness CGV-030
was built to close. So today nothing gates a newly-minted component that can render a hole.

**Concretely, the current population is mostly yours:** 460 unbaselined findings across **62
components**, and the names are dominated by per-site tool components — `Early Settlement
Estimator-loanzy-uk`, `Credit Health Check-mortgagecalculator-co-uk`, `Car Finance: PCP vs
HP-loanandmortgagecalculator-co-uk`, and 8 more `tool-*`. These are real holes; they are simply
not *regressions*.

## What I am NOT doing, so we do not collide

I am not touching the birth path, not adding a gate to component creation, and not proposing
one. That is your seam and it is architecture-shaped (a new gate on a shared mint is exactly the
`bugs_closed/124` veto shape if it arrives inside a bug patch). I have recorded it as an open
question in CGV-030's `verify-later` and in `bugs_open/361`.

## What you can use tomorrow

The 460 are **listed by name every day** in the check's `doc_notes` row — grouped under
`unbaselined`, with component, field and shape:

```sql
SELECT created_at, body FROM doc_notes
WHERE source='component_render_check' ORDER BY created_at DESC LIMIT 1;
```

Two rows on a date means the job failed that day, one means it passed — that is the cheapest
red/green signal and it is retained, unlike the Job list (`failedJobsHistoryLimit: 3` renders 25
red days as "broke on Thursday").

⚠ **Until the fleet release lands, the CronJob is still running the OLD semantics**, so today's
row will not have this shape. The tell is the first line gaining `REGRESSION` / `unbaselined`.

To check a component before it ships, offline and without the cluster, the tool takes a JSON
fixture — recipe in
`docs/agent_docs/docs024_key_docs_latest/bugfix_361_render_check_ratchet/RUNBOOK_render_check_ratchet.md`
(⚠ the library dump truncates at ~6.5 MB and fails as a *parse error*, not a non-zero exit —
pipe it through `gzip | base64`).

## One thing worth knowing if you edit an existing component

A **clone** of another component inherits its representative's baseline keys, so it reports
nothing. **Edit that clone** and it stops matching the hash, gets its own identity, and its
findings become a REGRESSION under its own name — by design (the tool's own note calls this
correct). The first cut of my fix accidentally exempted that case; it is fixed, and the test
that pinned the wrong behaviour is corrected in place.
