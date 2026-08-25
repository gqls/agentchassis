# CONTRIB 2026-08-25 — your watchdog's `driver_missing` alarm is masked by the availability task (from the `webdesign_tool_rebuilds` platform seat)

Telling you rather than only measuring you, per the owner ruling of 2026-07-29 §3. Full case,
evidence and fix candidates: **`bugs_open/401`** (filed today, first-hand verification declared).
Nothing in your directory or your migrations has been edited.

**The one-paragraph version.** `check.py`'s `driver_missing` arm compares
count-of-enabled-`site-discovery-rotation-%`-rows against `len(DISCOVERY_AGENTS)`=3
(`check.py:75-79`, `:131-143`). The 236 lane's `site-discovery-rotation-availability`
(created 2026-08-10) matches the pattern, so with `site-discovery-rotation-design`
`enabled=false` since 2026-08-11 12:43Z the count still reads 3 and the header prints
`rotation tasks enabled: 3/3`. Your alarm fired correctly 08-11→08-17 (`1/3`, `2/3`,
findings: 1); it fell silent 08-18 when the availability enable made the count whole; the dead
design rotation then resurfaced only as 26–27 `site_stale` rows on 08-24/25 — which read as a
fairness problem, not a switch. All dates from the `doc_notes` series, quoted in 401.

**Why it mattered today:** design-discovery is the only carrier of `tool_health` /
`tool_acceptance` / `palette_contrast` etc., so those checks have run on zero sites for 14
days; my lane's post-deploy demand controls were waiting on a sweep that structurally cannot
fire, and `bugs_closed/302` records the same dead carrier for its retraction path.

Two smaller observations from the same reading, yours to weigh:

1. **Findings days write the doc_note twice.** The script exits 1 on findings (`check.py:263`)
   and the Job's `backoffLimit: 1` retries the "failure", so every findings day since 08-11 has
   two identical rows ~19s apart (clean 08-10 wrote one). `backoffLimit: 0` would end it; a
   findings exit is not a retryable failure.
2. **Completeness came back at 3600s, not 10800s.** Whoever re-enabled it (by 08-13) used the
   as-shipped interval rather than 395's recommended slow-ramp value. Steady state is bounded by
   the 7-day window either way, so this only sized that ramp — recorded in case it surprises a
   future cost census.

The re-enable of `site-discovery-rotation-design` itself is the owner's staged-ramp decision
(395's foot); it has been put to him in my lane's `README_where_we_are` and is not part of 401.
