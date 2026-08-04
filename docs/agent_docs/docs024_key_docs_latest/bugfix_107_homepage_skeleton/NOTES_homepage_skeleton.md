# NOTES — bugfix 107 homepage skeleton (append-only, newest at the bottom)

## 2026-08-04 — lane opened

- Picked 107 after a fleet-wide ownership sweep: who-owns + live-transcript
  greps + `site_work_items` queue. The only recent touch was a triage `head -45`
  read at 09:17 today by an idle session.
- On the way in, found `bugs_open/121` (house voice) was fixed AND live since
  2026-07-27 — the file simply never moved. Re-verified all four layers today
  (canonical row, template placeholder, pod-grep v1.0.1251, llm_call_log
  artefact check: 8/8 today's page-content-writer prompts carry the resolved
  voice block) and closed it (`4c449273e`).
- Re-validated 107 against the live fleet (RUNBOOK §1). Strongest single fact:
  **lendzy.co.uk, built 2026-08-02 — six days after filing — has the exact
  skeleton the bug describes** (`hero > brief-explanation > info-card-grid >
  mechanism-flow > call-to-action`). The bug's original table remains accurate
  for the older sites.
- Sites that differ (vonc, gamesdesign, dartsonline) are hand-directed;
  ported sites (loancalculator family) bypass the planner entirely
  (`ported-prose`/`ported-page`). Neither refutes the claim — both are paths
  around the planner, not the planner varying.
- [INFERRED] the planner default is the cause — that is the bug file's reading
  and matches the composition census, but I have not yet read
  `plan_sections_action.go` end to end. Two read-only research agents
  dispatched (code map + docs prior art). 090 to be filed once the mechanism
  is grounded in symbols, per the 2026-07-31 owner ruling.
