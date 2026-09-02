# PLAN — calendar_component

Scope, decisions, and the open architectural question this lane inherits.

## 0. What "calendar" means on this platform, and what it does NOT mean

Established by research (2026-09-02) before this lane existed. Three distinct "calendar"
things live in the codebase and get confused with each other by name alone:

1. **`period-calendar`** — a generic PAGE COMPONENT (VIZ-017, migration
   `605_period_calendar_component.sql`). An ordered list of NAMED PERIODS
   (months/quarters/seasons/weeks), no dates, no numeric fields, a real `<ol>`. **This is
   what "calendar development" means on this platform, and this lane owns it.** Built as
   one of three sibling components (`checklist`, `comparison-table` — NOT this lane's
   scope) in `bugs_closed/381`.
2. **The `editorial_design_uplift` lane's Phase E "timeline"** — explicitly NOT this.
   Fact-fed, dated, cited real-world events. Boundary agreed 2026-08-25 (a year in a
   period label means you want theirs, not `period-calendar`). If a request needs dates
   or citations, route it to that lane, not here.
3. **"Calendar" as editorial subject matter** (`darts-calendar-density`, PDC tournament
   calendars on dartsonline.com) — an article topic, rendered through the unrelated
   `evidence-timeseries` chart component. Nothing to do with the `period-calendar` UI
   component. A grep for "calendar" will keep surfacing this — don't conflate it with (1).

## 1. Current state (inherited, verified fresh 2026-09-02)

- `period-calendar` is LIVE, council-approved (`c134b0e9`), registered VIZ-017.
- Adoption, queried live 2026-09-02: **2 placements** — `homegarden.uk` `/index.html`
  (landing, 2026-08-26) and `farmerinsurance.uk`
  `/blog/seasonal-farm-insurance-risks.html` (blog-post, 2026-08-31). The second
  placement is new and unrecorded anywhere before this session — see NOTES.
- Its two siblings remain at **zero** placements each (`checklist`, `comparison-table`),
  8+ days and multiple site builds after launch. Real, current, unexplained — not this
  lane's components, but a comparison point worth carrying forward.

## 2. The open item this lane inherits, unowned until now

`homegarden.uk`'s build (2026-08-25) surfaced something neither the `bugfix_381` lane
nor the `bugfix_206` lane claimed: told (informally, via the mission brief) to cover a
gardening year "month by month" with no calendar named explicitly, the planner satisfied
the promise at the SITE level — 17 pages, one per month plus a "this month" index, wired
into nav — rather than at the PAGE level with the `period-calendar` component built to
satisfy exactly that promise. Recorded by the `loanzy_uk_example_site` lane (its NOTES,
2026-08-25, "the interaction neither lane predicted") but explicitly not filed —
"belongs to neither bug".

This is a real design question for a calendar-development thread to own:

- Is a 17-page month-by-month structure a legitimate alternative to one page carrying
  `period-calendar`, or is it the planner reaching for a heavier, harder-to-maintain
  shape when a lighter component would satisfy the same promise more cheaply?
- Those 17 pages depend on `directory-build-handler`/`section-index` builder coverage
  (`bugs_open/206`, still open) — if that population no-ops, the site's whole calendar
  disappears. `period-calendar` has no such dependency: it renders inline off
  writer-authored content, nothing else has to build first.
- Nobody has decided whether the planner SHOULD be nudged toward the cheaper component
  when both would satisfy the promise, or whether both are legitimate and the choice is
  content-dependent (a full 17-page month guide vs. a compact `period-calendar` list).

**Next action for this lane:** read the current `build-site-planner` prompt's
`period-calendar` description (see RUNBOOK) to see whether it already nudges toward the
lighter component, before proposing anything.

## 3. Decisions and their reasons

None yet — this is a fresh thread. Log corrections/decisions here as they're made, not
silently.
