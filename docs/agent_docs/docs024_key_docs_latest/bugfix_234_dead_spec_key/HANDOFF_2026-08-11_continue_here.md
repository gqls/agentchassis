# HANDOFF — bugfix 234 lane, 2026-08-11. Cold start for a fresh chat.

## STEP 0 — has it already landed? (this lane is DONE; do not redo it)

`bugs_open/234` is **fixed, live, behaviourally proven, and council-APPROVED**. The file
stays in `bugs_open/` per the owner ruling of 2026-08-06. **Nothing technical is owed.**
Before acting on anything below, re-read `bugs_open/234`'s status header — it is written to
be true at a glance, and this tree moves under you.

Verify in 30 seconds (do NOT trust this document's figures — it is a snapshot):
```sql
-- the fix, at a filed row: any improvement_rerender_* row created after 2026-08-10 10:49Z
SELECT item_key, spec, created_at FROM site_work_items
 WHERE created_by='improvement-loop' AND item_key LIKE 'improvement_rerender%'
 ORDER BY created_at DESC LIMIT 3;   -- expect spec to carry refresh_site_components
-- the class guard is clean fleet-wide
SELECT body FROM doc_notes WHERE subject_key='removed-config-keys' ORDER BY created_at DESC LIMIT 1;
```

## What was wrong, and what shipped

Three live `create_work_item` steps set a config key `spec` that **no version of the action
has ever read** (it composes the item spec from `spec_data`/`spec_paths`/`spec_literal`), so
every item they filed carried `spec = '{}'`. The one with teeth: improvement-loop's
`refresh_site_components: true` never reached the rerender gate — 17 rows, none noticed, over
eight days.

| what | where | state |
|---|---|---|
| data fix | migration **364** (+ seeds 054/291/269) | applied, recorded, guards mutation-proven |
| the class fix | `ActionInputSpec.RemovedConfigKeys` (register **SCR-007**) + `StrictConfig` on `create_work_item` | LIVE |
| the standing guard | `removed-config-keys-check` CronJob, 06:25 UTC, runs the **Go** binary | deployed, green |
| adoption #2/#3 | migration **370** + `update_page_status` retires `notes_field`, `validation_issues_field`, `commit_from` | applied + LIVE |
| council | corr `3eb0d1f1-6929-4131-bbef-c636256aa667` | **APPROVED r4**, 0 MISMATCH in `098` |

## What is OPEN (all owner decisions — none is blocked work)

1. **`StrictConfig` on `update_page_status`.** Its census is clean (9 steps, 0 unrecognised).
   Deliberately NOT flipped: under the RFC_021 Q1 protocol every strict flip is its own
   adoption with its own census recorded in its own commit. Offered to the owner 2026-08-11
   and **not chosen** — do not take it as implied.
2. **A 4th `RemovedConfigKeys` adoption** anywhere else. Same protocol. The full-inventory
   query (the guardian's round-4 ask) is in this lane's RUNBOOK — run it, not just the
   zero-count, before any flip.
3. **Nothing else.** The lane's own work is finished.

## The five things most likely to burn the next session

1. **A strict / removed-key rejection happens in the CHASSIS, BEFORE any agent spawns.** So
   there is **no witness pod and no orchestration row** — the only trace is a chassis log
   line. Three canary firings were lost to pollers watching for a pod, and a real fleet
   outage supplied a believable wrong explanation. Grep the chassis log.
2. **The chassis rotates its `build provenance` startup line away within minutes**, so
   CLAUDE.md's "ask the service what it is running" recipe returns nothing here. And
   grepping a binary for **your** commit fails by design: it carries **one** sha, its build
   point. Correct method in this lane's RUNBOOK — extract the binary, find the stamp among
   recent commits (with a must-be-absent control), then `git merge-base --is-ancestor`.
   **Prefer behaviour to provenance where you can get it.**
3. **The council verdict query printed by the trigger does not filter by correlation** —
   it returns whichever lane finished last. Corrected at source in `RUNBOOK_council_gate.md`;
   confirm YOUR run is `COMPLETED` first, then filter, with `LIMIT 3` because a resubmission
   reuses the correlation.
4. **A submission's `file` field is schema-validated** — any whitespace and the round dies at
   `persist_submission` with `error=NULL` and an empty `execution_path`, before a single seat
   sees it. Overflow goes in `rationale`/`sketch`.
5. **A witness agent carrying a deliberately bogus key poisons the very census the adoption
   protocol depends on.** Delete it afterwards, always.

## Missteps worth reading before you trust anything here

Four in `WRONG_CALLS.md` (2026-08-11), each with the cheap check that would have caught it:
a fleet census written **into a Go comment** (false within the hour, because ~30 sessions
share this tree); a filtered count leaking into its own denominator; prose in the schema-
validated path field; and citing the **wrong provenance mechanism** for `commit_from` — a
seat caught that one, and its stated reason was imprecise while its instinct was right.
That last one is the general lesson: **check an objection even when its reason is wrong.**

## Where the rest lives

- case file: `bugs_open/234_HANDOFF_2026-08-09_…md` (status header first)
- lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_234_dead_spec_key/`
  — `SUMMARY_2026-08-11` is the read-out, `NOTES` holds the missteps, `RUNBOOK` the commands
- the architecture ruling: `architecture_review/RFC_021` (§1 is the four-state contract)
- register: **SCR-007** in `docs026_concept_register/register/adopting-and-scraping.md`
