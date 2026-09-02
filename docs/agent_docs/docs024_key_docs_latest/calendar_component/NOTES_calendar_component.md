# NOTES — calendar_component

Append-only. Newest at the bottom.

## 2026-09-02 — lane opened

Session renamed "calendar" by the owner; asked to research everything discussed about
calendars fleet-wide and take ownership of calendar development if no dedicated thread
already existed.

**Grepped fleet-wide** (docs, bugs_open/closed, concept register, Go code, memory).
No bug, no workstream directory, no concept-register entry is calendar-specific in the
sense of a dedicated lane. "Calendar" hits fell into three unrelated buckets (see PLAN
§0): the `period-calendar` component, the `editorial_design_uplift` lane's
explicitly-distinct fact-fed timeline, and "calendar" as editorial subject matter
(dartsonline PDC pages, using an unrelated chart component). Only the first is
calendar-shaped work in the UI/component sense.

**Traced `period-calendar`'s history.** Built 2026-08-24/25 as the second half of
`bugs_closed/381` (the "planner composes pages from components that cannot express the
page it planned" bug), alongside two unrelated siblings (`checklist`,
`comparison-table`). Closed 2026-08-25 after a false start — an earlier draft of the
closure claimed "the calendar SHIPPED" on `homegarden.uk` when that domain was still a
parked stub returning HTTP 200 for every path (`WRONG_CALLS.md`, 2026-08-25 entry). The
final closure re-verified with a proper invented-path 404 control before landing.

**Checked current adoption, live DB, 2026-09-02** — see RUNBOOK for the query.
**2 placements**, not the 1 the closing lane last recorded on 2026-08-25:
- `homegarden.uk` `/index.html` (landing), 2026-08-26 — the build that closed the bug.
- `farmerinsurance.uk` `/blog/seasonal-farm-insurance-risks.html` (blog-post),
  2026-08-31 — **new, nobody had written this down before now.** First adoption on a
  second site, and the first on a `blog-post` rather than a `landing` page.

`checklist` and `comparison-table` both confirmed at **0** placements (explicit LEFT
JOIN, not a silent GROUP BY omission — see RUNBOOK for why that distinction matters
here). Not this lane's components, but worth carrying forward as a comparison point:
whatever makes `period-calendar` get chosen and its siblings not is worth understanding
before assuming the pattern will hold as more sites build.

**Found the open architectural question.** The `loanzy_uk_example_site` lane recorded,
on 2026-08-25, that `homegarden.uk`'s planner satisfied a "month by month" brief at the
SITE level (17 separate month pages) rather than reaching for the PAGE-level
`period-calendar` component built for exactly that promise — and explicitly declined to
file it, since it belongs to neither the 381 nor the 206 bug. Nobody has picked it up
since (checked: no later mention anywhere in the fleet as of today, `grep -rn
"calendar-shaped"` returns only the three hits already accounted for). This is now this
lane's first open item — see PLAN §2.

**Not done yet:** read the current `build-site-planner` prompt text for
`period-calendar` to see whether it already nudges toward the lighter component, before
proposing any change.
