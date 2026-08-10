# HANDOFF — loancalculator.co.uk · the `bugs_open/227` PLATFORM thread · **CLOSED** (written 2026-08-10 evening)

**Supersedes `HANDOFF_2026-08-10b_continue_here.md`**, whose §3 ("the one thing still owed") is
now done. Nothing on this job is outstanding.

> ⚠ **This directory has more than one live lane.** This file is **only** the `bugs_open/227`
> platform job. The site's COPY/VOICE thread is `HANDOFF_2026-08-09b_continue_here.md` —
> different job, not superseded by this one.

```
site         loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis      v1.0.1283 (rolled 2026-08-10 ~21:44Z; releases are whole-fleet, owner runs make release)
the site     DONE — 26/26 pages voice H, serving, calculators golden. NOT this file's business.
the job      bugs_open/227 — BOTH defects fixed, live, and PROVEN ON BOTH ARMS. Nothing owed.
```

Plain-prose read-out for talking about it:
`SUMMARY_2026-08-10_a_plan_the_council_rejects_now_leaves_nothing_behind.md`.

## What shipped, and where its proof lives

| | what | proof |
|---|---|---|
| **345** | the site brief becomes data (`load_brief` reads `doc_notes` per experience) — kills the hardcoded single-site diagnosis, 48 hits → 0 | two runs, opposite directions, keyed only on `subject_key` (`c3976aab` / `72f540d3`, 08-09) |
| **363** | the plan persists ONLY from the council's approved branch | **arm 1** corr `d81aa5f4` (a veto writes nothing) · **arm 2** corr `c4127fe7` (an escalated run writes nothing) |
| **370** | three descriptions 363 left describing the graph it replaced | in-transaction whole-row verify; 0 retired claims remain |

All config-only: live on apply, no image, no roll, and the council gate refuses DB config
client-side. **Re-verified intact at v1.0.1283** — third confirmation that a config-only fix
survives a rebuild.

## The two arms, because the difference between them is the whole lesson

**Arm 1 — a VETOED composition is never written** (corr `d81aa5f4`, 14:40–14:46Z). Round 1 drew
a real `veto from feasibility`; the plan-row count stayed at 0 through `compose`, through the
veto and through the entire reframe round, reaching 1 only when round 2 was approved. Under the
old graph that run writes TWO rows and the first is the vetoed one, `is_current`. The
mid-flight sample at `review_journeys` is the load-bearing reading — the old edge was
`compose → persist_plan → review_journeys`, so that is the moment the two graphs disagree.

**Arm 2 — a run that ENDS non-approved writes nothing** (corr `c4127fe7`, 22:04–22:10Z).
`compose` succeeded with a real **10,498 b** plan; all five seats ran; council returned
`rejected` / `veto from feasibility` / `should_reframe=false`; the run ended
`complete_escalated`; `doc_plans` kept exactly one row — the earlier approved one, `updated_at`
unchanged, not superseded.

**⚠ The figure that makes arm 2 a proof is the 10,498 b, not the row count.** An earlier attempt
(14:51Z, killed by the fleet-wide API cap) returned `complete_refused`, no new row, plan of
record unchanged — the exact sentence we wanted, and worthless, because `compose` never
returned so the old graph writes nothing either. **The pass only counts when a write was
POSSIBLE:** `compose` succeeded and `collected_data->'proposal'->>'result'` is non-empty at the
end. Any experience-planner run finishing in well under ~7 minutes did not run — read
`collected_data->'__step_error'` (a failed step shows `COMPLETED` with `error` NULL).

## If you ever need to re-run this verification

Recipe with every gotcha: `RUNBOOK_loancalculator_couk.md` §"Observing a COUNCIL-VETOED
experience-planner run". Fixture: `probe_363_veto_arm_brief.sql` in this directory (kept out of
`sql_for_agents/` so the migration runner cannot sweep it). In short: seed an unbuildable
experience under a probe `subject_key` (`load_brief` keys on `subject_key`, **not** `site_id`,
so nothing real is touched), verify the brief actually loads before spending a run, then either
sample mid-flight (arm 1) or cap `council_decide.config.max_rounds` to 1 and let it escalate
(arm 2) — **restoring `max_rounds` to 5 afterwards and reading it back from the live row.**

**A veto is not terminal by design** — `reframe` is told to demote the vetoed feature to
coming-soon and try again, and `applyCouncilCaps`
(`platform/orchestration/actions/diagnose_council_decide_action.go:663`) escalates only on a
**second** rejection. So "wait for a natural veto" waits for something the design avoids. This
is filed as a landmine.

## Cleanup — done, do not redo

- Probe brief `doc_notes` row **deleted** (the fixture SQL is committed, so it is reproducible).
- Probe plan `6ebe06f5` **demoted, not deleted** — `is_current=false`, reason in `notes`. It is
  the evidence for arm 1 (its 7,661 b = the `reframe` response exactly, which is what proved
  persist reads the *latest* composition). `live-lender-approval-race` has **0** current plans.
- Intake work item cancelled. Note the trigger inserts a fresh one per fire, so a previous
  cancel does not cover a later probe.
- `max_rounds` back at **5**, read back from the live row.
- `LANDMINES.md` entry synced to `doc_notes` **and verified** — `landmine-verifier` run
  `8d0d8f20`, verdict **STILL_VALID**, having independently re-read `applyCouncilCaps`
  L663-671 and confirmed the `shouldReframe` guard the entry quotes. (Fired per-entry with
  `scripts/trigger-landmine-verifier.sh '<source slug>'` rather than
  `landmines-verify-dispatch.sh`, because the wrapper re-syncs the whole file and would have
  spent a run on another lane's in-progress entries. Its own caveat, worth knowing: the code
  index is **.go-only**, so agent-definition rows, shell scripts and config keys came back
  "not answerable" — 4 of its 10 checks. It cannot verify the half of a landmine that lives
  in the database.)
- `WRONG_CALLS.md` carries the near misses; `NOTES` has the full record including the missteps.

## What was deliberately NOT done, and who owns it

**Running the planner mutates the plan of record — there is no dry-run mode.** That is how a
test displaced vonc's plan on 08-09 (restored by hand, held across two rolls since). Closing it
means route (b): an opt-in `set_current_when` on the shared `write_doc_plan` action so a plan
can be written not-current. That is a **platform seam** — it owes an architecture round and a
concept-register entry, and per the owner's 2026-08-02 ruling the new authority should ship as
an opt-in field with the unsafe default OFF, not as a documented contract. Not this lane's to
take unilaterally.

Also unclosed by design: `complete_escalated` no longer surfaces a persisted plan (363's stated,
deliberate loss). The escalated draft lives in the run's own `collected_data` and in
`llm_call_log`, by correlation id. Route (b) is what would make it durable in `doc_plans`.

## If you are starting cold

`bugs_open/227` (top correction block, then §"How to verify a fix" — the two arms are separated
there), then 363's header, then this file. `NOTES_loancalculator_couk.md` `## 2026-08-10
afternoon` and `## 2026-08-10 evening` have the evidence and the missteps;
`README_where_we_are.md` has the owner's plain-prose version; the SUMMARY above is the read-out.
