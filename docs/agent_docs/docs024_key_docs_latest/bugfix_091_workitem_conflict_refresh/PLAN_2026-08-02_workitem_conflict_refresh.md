# PLAN — `bugs_open/091` candidate 1: a repeat finding must refresh the open record, not vanish

**Started** 2026-08-02, by the bug-sweep session that picked 091 off `bugs_open/`.
**Bug:** `bugs_open/091_HANDOFF_2026-07-26_second_drift_dropped_by_dedup_and_reported_as_raised.md`
**Owning remit (stated by the bug):** `work_item_completion_integrity` for the shared
helper; `claims_verification` for the V4 evidence layer. Neither has committed to
either area since 2026-07-27, `who-owns.py 091` identifies no owning workstream, and
091 appears in none of the 19 live session transcripts (checked 2026-08-02 23:50).

---

## Why this one, and why now

Candidate 2 (report honestly) shipped on v1.0.1177. Candidates 1 and 3 are owed.
Candidate 3 is ruled out by the bug itself (it multiplies rows into the
`needs_human_review` queue that `bugs_open/033`'s owner ruling says must not fill —
368 rows today). **Candidate 1 is the only remaining shape**, and the bug states it
precisely: `ON CONFLICT … DO UPDATE` behind an explicit opt-in, never as a blanket
change to a helper every detector in the fleet calls.

## The measurement that resized this bug — it is ACTIVE, not a latency annoyance

The filed text says "this is a delay, not a loss" and rates it **Medium**. Measured
today against the live DB, that undersells it. The `evidence-freshness` scheduled task
is **enabled and ran at 2026-08-02 18:36:07Z**. Comparing the durable record each open
`stale_evidence` item carries against what that run actually found:

| site | item filed | the item says drifted | today's run found drifted | record correct? |
|---|---|---|---|---|
| leopardessconsulting.co.uk | 07-26 | `C4-orchestration-state-records` | `C4-agent-definitions-catalogue` | **NO — a different fact entirely** |
| fundamentallyai.com | 07-26 | F11, F12, F13 | F9, F10, F11, F12, F13 | **NO — two drifting facts invisible** |
| ai-agent-orchestration.com | 07-26 | agent-definitions, agent-types, orchestrations | agent-definitions, agent-types | **NO — over-reports a fact that re-synced** |
| vonc.com | 08-01 | `vonc-tools` | *(nothing)* | **NO — describes drift that no longer exists** |
| oufe.com | 07-27 | 12 CIT-* | the same 12 CIT-* | yes |

**Four of five open items are factually wrong today.** The human who opens the
leopardess item is sent to fix a fact that is not the one that moved. That is not a
delay; it is a durable record actively misdirecting the only consumer it has
(`handler_agent='human-review'`, HITL-terminal — no machine ever re-reads it).

Query is in the RUNBOOK. The whole comparison is possible only because the *run* is
retained for ~24h — so re-run it the same day or it reads as empty.

## What is being built

### A — `conflictPolicy` on the shared work-item writer (the door-closer)

`insertWorkItem` (`platform/orchestration/actions/load_work_item_actions.go`) is the
one door ~20 call sites use. It ends in

```sql
ON CONFLICT (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (<terminal>)
DO NOTHING
```

which correctly refuses a second OPEN row for a key — and, as a side effect, throws
away the *finding*. Add an opt-in second policy that keeps the row single and brings
its **description** up to date.

**Three design commitments, each because the obvious version is wrong:**

1. **A second statement, not a `DO UPDATE` clause on the existing one.** The hot path
   every other caller uses stays byte-identical. The refresh is an explicit `UPDATE`
   run only in the conflict branch, inside the caller's existing transaction, with the
   same terminal-status predicate — so a row that went terminal between the two
   statements is not resurrected, it simply does not match and the caller is told
   nothing was recorded.
2. **`DO UPDATE`'s `RowsAffected` would re-break candidate 2.** An UPDATE affects one
   row, so `rows > 0` would start reading `true` again for a write that created
   nothing — reintroducing the exact dishonesty 091 was filed about, inside its own
   fix. The writer therefore returns a **three-state outcome** (`Inserted` /
   `Refreshed` / neither), and `work_item_created` keeps meaning *created*.
3. **The policy is a parameter of a new `writeWorkItem`, not a field on `workItem`.**
   If it were a struct field, a caller could set it and still call `insertWorkItem`,
   whose single `bool` cannot express a refresh — a silent wrong answer. As a
   parameter the old function literally cannot receive it. The type system does the
   work rather than a comment asking people to be careful.

**Not refreshed, deliberately:** `status`, `priority`, `handler_agent`, `severity`. A
human may have moved any of them on the open row. The finding owns its own
description (`summary`, `spec`) and nothing else. And the refresh **skips a row a
handler currently holds** (`claimed`, `diagnosing`) — mutating the spec under a
running handler is a different bad state, and this seam should not create it.

### B — the two council findings 091 records, both answered by the same change

The bug's STATUS block carries two reviewer objections from council `a5b70424`,
both about `apply_gap_plan_action.go`, and both owed:

* `guidelines`: its three hand-rolled `INSERT INTO site_work_items … ON CONFLICT DO
  NOTHING` statements violate the work-item dedup rule;
* `reuse_agent`: they should route through `insertWorkItem` and inherit its dedup
  semantics and its boolean rather than hand-roll.

They are right, and the code already says so in a comment at
`apply_gap_plan_action.go:555-558`. **Why the hand-rolling exists** is the useful
finding: those items set `parent_item_id`, and the shared `workItem` struct has no
field for it — so the shared door was unusable for anyone who needed a parent, and
three call sites walked round it. Adding `parentItemID` to the helper removes the
reason, then routes all three through it.

**Behaviour is preserved deliberately**: they adopt with `recurrenceExpected: true`.
A gap plan re-requesting a page build is an ACTION REQUEST, which is exactly the case
the anti-churn heuristics were opted out of in `bugs_open/024`. Without that flag,
adoption would silently start suppressing gap-plan items within 3h of a terminal
predecessor and branding them `unresolved` after two — a behaviour change riding in
on a bug patch, which is the thing this repo keeps being bitten by.

## Scope and venue

**This is a platform seam** (a new capability on a shared mechanism), so:

- it goes through the **council gate** before/alongside the commit;
- it is **registered in the concept register in the same commit that ships it**, with
  its landmine — condition (2) of the ordering exemption, which stands;
- it is **not** an RFC. Per the owner ruling of 2026-07-29, an RFC is triggered when a
  change alters what the shared mechanism *guarantees*. `dropOnConflict` is the
  default and is byte-identical to today; nothing reaches the new path until a caller
  names it. Additive-and-inert, not additive-and-guarantee-changing.
- **No ordering constraint is claimed.** HEAD is shared; another session's build ships
  this whenever it happens. Review here is after the fact, by design.

## Verification — the bug prescribes it and it is not a build

091 §"How to verify a fix" is explicit: **not by re-reading the code.** Induce it.
Drift a second, different fact on a site that already has an open `stale_evidence`
item, and require BOTH:

- the open item's `spec->'drifted'` now names the newly drifted fact, and
- the run's `work_item_created` reads **false** (nothing was *created*), with
  `work_item_refreshed` true.

A clean sweep proves nothing — the happy path reported true before and after.

## Order of work

1. ~~Confirm the bug is live and unowned~~ — done, see the table above.
2. Build A, with unit tests that MUTATE the guard to prove it fires.
3. Build B on top of A's `parentItemID`.
4. `go build` + tests against `git archive HEAD`, not the dirty tree.
5. Council submission; commit with `Council-Submitted:`.
6. Register the seam + landmine in the same commit.
7. Roll, then induce the fault live. Close only when live — a fix committed but inert
   leaves the defect reproducible, which is the `bugs_closed/` bar.
