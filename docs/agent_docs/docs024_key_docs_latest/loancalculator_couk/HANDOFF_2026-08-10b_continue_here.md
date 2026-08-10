# HANDOFF — loancalculator.co.uk · **the `bugs_open/227` PLATFORM thread** · continue here (written 2026-08-10 mid-afternoon)

**Supersedes `HANDOFF_2026-08-10_continue_here.md`** for the 227 job. That file's §"What is
DONE" and its two estate facts stay accurate; what it listed as owed has now been **half
closed by observation and half narrowed**, and its recommended route to the other half —
"wait for a natural veto" — turns out to be waiting for something the system is designed to
avoid. Read this file's §3 before spending a run on it.

> ⚠ **This directory has more than one live lane.** This file is **only** the
> `bugs_open/227` platform job. The site's COPY/VOICE thread is
> `HANDOFF_2026-08-09b_continue_here.md` — different job, not superseded by this one.

```
site         loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis      v1.0.1279 (pods rolled 2026-08-10 ~13:42Z by another lane; config survived, re-verified)
the site     DONE — 26/26 pages voice H, serving, calculators golden. NOT this file's business.
the job      bugs_open/227 — both defects FIXED AND LIVE. Veto arm PROVEN. One narrow arm owed.
```

## ⛔ READ FIRST — THE FLEET IS BLOCKED, AND IT IS NOT THIS LANE'S DOING

**At 14:51Z on 2026-08-10 the Anthropic account hit its configured usage cap.** Every
LLM-driven agent on the estate now fails at its first model call:

```
API request failed with status 400: {"type":"invalid_request_error",
 "message":"You have reached your specified API usage limits.
            You will regain access on 2026-09-01 at 00:00 UTC."}
```

Observed on **two agent types in the same minute** — this lane's `experience-planner`
(`compose`) and another session's `council-gate` (`review_architecture`, request_id
`req_011CduFpnewdmmf9ak88Ww2t`). **Last successful call estate-wide: 14:51:45.067Z**
(`SELECT max(created_at) FILTER (WHERE success) FROM llm_call_log` — use the `success`
column, not a test on `response_text`). Nothing is broken and nothing is lost; queued work
will simply fail until the owner raises the cap on the billing side. **Do not diagnose an
agent failure today without checking this first** —
`SELECT collected_data->'__step_error'->>'message' FROM orchestration_states WHERE
collected_data ? '__step_error' ORDER BY updated_at DESC LIMIT 3;`

**Independently corroborated, and already written up by another lane** — see `LANDMINES.md`,
the July "specified API usage limits" entry, which `bugfix_236_site_availability` extended
this afternoon with the recurrence (diagnosed from a standalone service *outside* the
cluster, `6a4fbab21`, which is what establishes it as ACCOUNT-level rather than a chassis
credential fault). Two consequences from their entry worth carrying: **size the outage by the
absence of any success, never by the error count** (a quiet fleet produces only ~5 error rows,
which reads like noise); and **every lane's "submit to the council before or alongside
committing" obligation is currently unsatisfiable** — a `Council-Submitted:` trailer written
now names a run that will never reach a verdict, so record in your lane docs that a fresh
submission is owed. This lane's commits today are config + docs only, which the gate refuses
client-side anyway, so nothing here is affected.

Everything below that needs a model run is blocked by this. Everything config-only is done.

---

## 1. State in one paragraph

`bugs_open/227` had two defects, both fixed, live, and config-only. **Defect 1** (the planner's
prompts hardcoded one site's diagnosis) — migration **345**, applied 08-09, proven live in both
directions. **Defect 2** (a plan persisted `is_current` before the council voted, with nothing
demoting a vetoed one) — migration **363**, applied 08-10 morning. As of this afternoon 363 is
**proven for the approved arm AND for the vetoed round**; what remains unobserved is only a run
that *ends* non-approved. Migration **370** (this session) fixes three strings 363 left behind
that describe the graph it replaced.

## 2. What this session PROVED — corr `d81aa5f4-a732-4fb3-b438-4ff496ef7ba2`

A veto cannot be ordered up, so one was **seeded**: a realistic owner brief for an experience a
static host genuinely cannot deliver (live partner-lender decisions API polled from the page
with a key in the query string, a presence counter by postcode, per-visitor state written
server-side and read back cross-device, and an owner line refusing a coming-soon label).
Filed through 345's own channel — `doc_notes`, which `load_brief` selects by **`subject_key`,
not `site_id`**, so no real experience could be touched. Fixture:
`probe_363_veto_arm_brief.sql` in this directory (kept out of `sql_for_agents/` so the
migration runner can never sweep it up). Recipe and gotchas: RUNBOOK §"Observing a
COUNCIL-VETOED experience-planner run".

| time (UTC) | step | `doc_plans` rows for the key |
|---|---|---|
| 14:40:33 | `compose` returns 12,189 b | 0 |
| **14:40:48** | **`EXECUTING_STEP @ review_journeys`** | **0** |
| 14:42:42 | council round 1 → **`veto from feasibility`** | 0 |
| 14:44:05 | `reframe` returns 7,661 b | 0 |
| 14:44:12 | `EXECUTING_STEP @ review_journeys` (round 2) | 0 |
| 14:45:51 | council round 2 → `approved with 2 advisory objection(s)` | 0 |
| 14:46:26 | `COMPLETED @ complete` | **1** |

**Under the old graph this run writes TWO rows and the FIRST is the vetoed one, `is_current`** —
bug 227's second defect verbatim, reproduced deliberately and now absent. The veto was a real
one: no server for the write endpoint, no cross-device store, an API key exposed in client JS.

**The mid-flight sample is the load-bearing reading**, not the final count. The old edge was
`compose → persist_plan → review_journeys`, so `review_journeys` with the count still at
baseline is the moment the two graphs disagree — and it discriminates on a run of any length,
which the row count does not (see the 08-10 morning correction).

**A bonus check nobody had run, and it was the dangerous one.** The persisted body is
**7,661 b = the `reframe` response exactly**, not compose's 12,189 b. 363's header verified
"compose, recompose AND reframe all write `proposal`" against compose+recompose runs only. Had
`reframe` written its own field, moving the write later would have persisted the **vetoed**
draft on approval — silent wrong content, not a missing row.

## 3. ⚠ THE ONE THING STILL OWED — and why the last handoff's route cannot deliver it

**Owed: a run that ENDS non-approved (`complete_escalated`/`complete_refused`) leaving no row.**

The previous handoff said "either wait for a natural veto, or seed an unbuildable experience".
Both were tried today and **neither reaches this arm, because a veto is not terminal by
design**:

- `reframe`'s prompt: *"If the vetoed feature admits no honest minimal-real version, demote it
  to a labelled coming-soon panel and move the real version to the LATER list — that is an
  acceptable honest MVP."* It is instructed to converge on something approvable, and it did.
- `applyCouncilCaps` (`platform/orchestration/actions/diagnose_council_decide_action.go:663`):
  `shouldReframe := rejected && rejectedCount <= 1 && round < maxRounds` — escalation needs a
  **second** rejection in the same run.

**The route that does work** (set up today, blocked by the cap before the council ran): cap the
council to one round, so any non-approved round-1 verdict routes straight to
`complete_escalated`.

```sql
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,'{workflow,steps,council_decide,config,max_rounds}', to_jsonb(1))
 WHERE type='experience-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,council_decide,config,max_rounds}' = to_jsonb(5)
RETURNING default_config #>> '{workflow,steps,council_decide,config,max_rounds}';
```
Then fire `092_TRIGGER_experience_plan.sh loancalculator.co.uk live-lender-approval-race "the
live lender approval race board"` (the brief is already seeded and drew a veto today), assert
**no new row and the plan of record unchanged** (`6ebe06f5…`, 7,661 b), and **restore
`max_rounds` to 5, reading it back from the live row.** It is a shared row — keep the window to
one run and arm the restore before you fire.

⚠ **THE RESULT THAT LOOKS LIKE A PASS AND IS NOT ONE.** Today's attempt returned
`COMPLETED @ complete_refused`, no new row, plan of record unchanged — and proves nothing,
because `compose` died on the API cap and the old graph writes nothing on that run either.
**Any experience-planner run finishing in well under ~7 minutes did not run.** Read
`collected_data->'__step_error'`; the status says `COMPLETED` with `error` NULL regardless.
This was the third check-that-cannot-fail on this bug in three days (`WRONG_CALLS.md`
2026-08-10) — the habit that catches all three is to state what the FAILING version of *this*
run would look like before reading the result.

**[INFERRED, not observed]** in the meantime: no non-approved terminal can reach a write.
`check_approved.config.then_step` is the **only** step-target reference to `persist_plan` in
the live row. Scan the target fields, not the raw text — 370 put the words into two
descriptions, so `default_config::text` now contains the literal 3 times and that count means
nothing:

```sql
SELECT s.key AS step, y.k AS field
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s,
     LATERAL (SELECT k, v FROM (
        SELECT 'next_step' k, s.value->>'next_step' v
        UNION ALL SELECT 'config.'||c.key, c.value#>>'{}' FROM jsonb_each(COALESCE(s.value->'config','{}'::jsonb)) c
                   WHERE jsonb_typeof(c.value)='string'
     ) x WHERE v='persist_plan') y
WHERE d.type='experience-planner' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL;   -- expect exactly 1 row
```

## 4. Migration 370 — three strings 363 left describing the graph it replaced

`370_experience_planner_escalation_descriptions_catch_up_with_363.sql` (+ ROLLBACK). Config
only, dry-run proven then applied, verified live (0 retired claims remain). No behaviour
change; three prose strings:

- `complete_escalated.description` said *"The current (rejected) plan **stays is_current** but
  MUST NOT be built"* — the exact state 363 made unrepresentable. A reader goes looking for a
  rejected plan of record and, not finding one, concludes someone already demoted it.
- `complete_escalated.config.success_message` — same claim, and this one travels to whoever
  reads the escalation.
- `recompose.description` — named the dead `persist` edge.

Its verify block sweeps the **whole row** for each retired claim rather than re-reading the
three paths it set: a `jsonb_set` on a wrong path silently adds a key, so "my new string is
there" is not "the old one is gone".

## 5. Housekeeping — done, do not redo

- Probe work item `needs_experience_plan:live-lender-approval-race` → `cancelled` (it was a
  test artefact; leaving it open would hold the key on `idx_swi_dedup`).
- The probe brief `doc_notes` row **is deliberately still there** (`5730bcf3-4e62-49d3-991e-a2e16e417975`,
  categories `["experience-brief","test-artefact","bugs-227-veto-probe"]`) — the arm in §3
  re-fires the same key and needs it. **Delete it when §3 is closed.**
- The approved probe plan `6ebe06f5-…` (7,661 b) is `is_current` for `live-lender-approval-race`,
  a key nothing builds from. Harmless; goes with the brief when §3 closes.
- `max_rounds` is back at **5** — verified by reading the live row, not by assuming the UPDATE
  landed.
- A new `LANDMINES.md` entry ("A council VETO on an experience plan is NOT terminal…") is
  synced to `doc_notes`. **`landmines-sync.py --apply` flagged it `NEEDS_VERIFICATION`** — the
  `landmine-verifier` pass needs a model call, so it is blocked by the cap. Run
  `landmines-verify-dispatch.sh` once the cap lifts.
- `WRONG_CALLS.md` carries the near miss; `NOTES` has the full session record; 363's header and
  `bugs_open/227` both carry the result and the narrowed remainder.

## 6. If you are starting cold

Read this file, then `bugs_open/227` (its top correction block first, then §"How to verify a
fix", which now has the two arms separated), then 363's header. `NOTES_loancalculator_couk.md`
`## 2026-08-10 afternoon` has the evidence and the missteps; `README_where_we_are.md` has the
owner's plain-prose version, including the API cap.
