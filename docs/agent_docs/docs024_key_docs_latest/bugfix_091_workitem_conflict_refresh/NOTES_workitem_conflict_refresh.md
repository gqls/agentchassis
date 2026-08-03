# NOTES — bugs_open/091 candidate 1

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-02 23:45 — picking the bug, and how much of `bugs_open/` is actually taken

Swept all 55 files in `bugs_open/`. `scripts/who-owns.py` returns "OWNED or
recently active" for almost everything (it is deliberately conservative), so the
discriminating check was the second one: **grep the 19 live `.jsonl` session
transcripts for `bugs_open/NNN`** — `who-owns` reads COMMITS and is blind to a
session that is mid-fix with nothing committed yet.

Only three files had no owning workstream AND no mention in any live transcript:
**080** (residual data repair, site-specific), **091**, **158** (item 1 is
explicitly an RFC, item 3 needs an owner decision — not closeable this session).
091 it is: its remaining candidate is a change to a helper every detector in the
fleet calls, which is the "robust and applicable to the framework" shape.

Rejected along the way, with reasons worth keeping:

* **093** — looked ideal (council escalated it twice). It is **blocked on
  `bugs_open/083`**, not on code: the fix shipped in v1.0.1172 and has never
  executed, because the only thing that runs it is `improvement-sweep`, disabled
  since 2026-05-02. A bug can be "fixed, live, and still open" for want of a
  cadence, and the file says so — read to the end before starting.
* **122** (WCAG) and **087** (page-rebuild section plan) — both actively owned
  (dartsonline_traffic, gemini_content_provider). 087 in particular reads as
  unowned at the top and has three dated sections of another lane's work below.
* **134** — real but latent (the agent has never run) and one instance fleet-wide.

## 2026-08-02 23:55 — the measurement that resized the bug

The file rates this Medium and says "a delay, not a loss". Measured against the
live DB it is worse: `evidence-freshness` is **enabled** and ran at 18:36:07Z
today, and **four of the five open `stale_evidence` items name the wrong facts**
(table in the PLAN; query in the RUNBOOK). leopardess's item names
`C4-orchestration-state-records` while the fact that is actually drifting is
`C4-agent-definitions-catalogue` — a different fact entirely. vonc's item
describes drift that no longer exists.

Three query missteps, all recorded in the RUNBOOK rather than here because they
will recur: `orchestration_states` has no `agent_type` column (it is
`owner_agent_type`); `jsonb_array_elements` on a scalar aborts the whole
statement, so a missing type guard reads as "no rows" rather than as an error;
and the table is retention-clocked, so exactly ONE `evidence-freshness` run
exists at any moment and the comparison is only possible on the same day.

## 2026-08-03 00:30 — the design turn: `DO UPDATE` would have re-created the bug

First sketch followed the bug file literally: `ON CONFLICT … DO UPDATE SET spec,
summary, updated_at`. Writing the test for it is what caught the problem.
**`DO UPDATE` affects a row, so `RowsAffected()` returns 1** — and
`insertWorkItem` returns `rows > 0`, which `work_item_created` is set from. The
literal fix would have made `work_item_created` start reporting `true` for a
write that created nothing: **091's own defect, arriving inside 091's fix.**

So: a separate `UPDATE` in the conflict branch (the shared INSERT stays
byte-identical for ~20 callers), and a three-state outcome instead of a bool.

Second turn, same shape: `refreshOnConflict` was going to be a **field on
`workItem`**, exactly as the bug file proposes. But then a caller can set the
field and still call `insertWorkItem`, whose single bool cannot express a refresh
— a silent wrong answer, at the one call site most likely to be copied. Made it a
**parameter of a new `writeWorkItem`** instead, so `insertWorkItem` cannot receive
it and the mistake does not compile. This is a deliberate deviation from the
filed candidate and the reason is written into the code.

## 2026-08-03 00:50 — MISSTEP: I widened a shared statement and broke twenty tests I had never opened

Added `parent_item_id` to the shared INSERT unconditionally. `go build` was
clean. **The package test suite then failed 20 tests across 8 files** —
`save_sections_prune_floor_test.go`, `tool_render_path_test.go`,
`page_role_upsert_test.go`, `nav_rebuild_request_test.go` and others — because
sqlmock matches the **argument count**, and every one of those expectations lists
16 args positionally. None of those tests has anything to do with this bug, and
several are in lanes other sessions are working right now (175, 178).

The fix was not to edit twenty files. It was to notice what the failure was
telling me: **a shared statement widened for one caller charges every caller.**
`parent_item_id` is now appended as `$17` only when a parent is actually set, so
a caller that never asked for it sends the identical statement it always sent,
and the only test files touched are the three whose behaviour genuinely moved.

Two lessons, the second more useful than the first:

1. `go build` says nothing about a shared SQL statement's blast radius. The test
   suite is the only thing that measures it, and it must be the WHOLE package.
2. **The size of a test breakage is a measurement of the seam, not an obstacle to
   it.** My first instinct was "update the twenty expectations"; that would have
   shipped a real widening and made three other lanes' files carry my change.

## 2026-08-03 01:10 — the guards are MUTATION-PROVEN, and one of them found a blind spot

Four mutations, each run and each confirmed to fail the suite (a guard that is
merely present is not a guard — a test that passes with the rule removed was
testing nothing):

| mutation | test that caught it |
|---|---|
| policy check removed → default policy refreshes | `DropOnConflict_IssuesNoUpdate`, `InsertWorkItem_CannotRefresh` |
| held-status clause dropped from the refresh predicate | `RefreshStatement_GuardsTerminalAndHeldRows` |
| a refresh reports `Inserted: true` | `RefreshOnConflict_UpdatesTheOpenItem` |
| `recurrenceExpected` cleared on gap-plan items | `GapPlanWorkItem_IsRecurrenceExpected` |

**The fourth one exposed a blind spot in the harness that is worth more than the
mutation was.** Clearing `recurrenceExpected` failed the direct unit test but
**not** the three behavioural `applyNewPage`/`applyRetypeExisting` tests — and it
should have, because with the flag off `insertWorkItem` issues an anti-churn
`SELECT COUNT(*)` those mocks do not expect.

Cause, read rather than guessed: `load_work_item_actions.go:1250` is
`if err == nil && terminalCount > 0` — **the probe's error is swallowed by
design**. So an unexpected query returns an error, the error is discarded, and
the test passes. Consequence, stated plainly because it will mislead somebody
else: **no behavioural sqlmock test in this package can detect a change to
`recurrenceExpected`.** The gap-plan adoption is therefore covered by a direct
assertion on the built `workItem`, not by the behavioural tests, which are blind
to precisely the thing most likely to go wrong in that adoption. Filed as a
landmine and named in the submission's own `risks` block.

This is the [[a-mutation-that-passes-may-have-hit-a-guard-in-series]] shape with
a twist: the mutation passed not because a second guard caught it, but because
the *observer* was deaf. "No test failed" was not evidence.

## 2026-08-03 01:25 — submitted to the council

`SUBMISSION_CORR = 8e7357ae-9f8d-49bf-81c0-669d9a97a205`, 7 edits. The `risks`
block carries the measurements rather than asking the reviewers to take them —
including the harness blind spot above, and the one judgement I actively want
challenged: `needs_human_review` is deliberately NOT in the held list, so an item
a human is mid-way through reading can change under them. It is a queue, not a
claim, and the alternative is leaving the record false — but that is a judgement,
not a measurement, and it is stated as one.

## 2026-08-03 01:55 — council APPROVED (13 reviewers, 0 unreadable, 6 advisory objections)

`decided_by: approved with 6 advisory objection(s) — none high-severity`. Four were
checkable, so they were checked rather than banked. **Read the report by
CORRELATION, not by `doc_notes … ORDER BY created_at DESC LIMIT 1`** — that
returned another lane's REVISE verdict for a completely different submission
(`4a7f0877`, the unpublish primitive), which for a few seconds read as *my*
verdict. The runbook query is `diagnosis_artifacts WHERE correlation_id=…
AND kind='council_report'`.

### Acted on

**1. "Is the negative test vacuous?" — three seats independently (editquality,
guardian, debug_historian).** All three cited the landmine I filed myself: a test
asserting a query is NOT issued passes vacuously against `insertWorkItem`, because
the anti-churn probe swallows the mock's error. **They were right to ask and the
answer is no — proven by mutation, not by reading.** Forcing the default policy
down the refresh path fails the test with:

```
writeWorkItem: refresh failed for stale_evidence:…: all expectations were
already fulfilled, call to Query '…UPDATE site_work_items…'
```

i.e. `refreshOpenWorkItem` **propagates** its error (only `sql.ErrNoRows` means
"nothing matched"), so an unexpected query surfaces as a returned error and the
`t.Fatalf` fires. The distinction from the probe is exactly the `err == nil` test
one swallows and the other does not. Written into the test as a comment with the
mutation recorded, so the next reader does not have to re-derive it.

**2. "The silent no-op depends on one caller remembering to check `Recorded()`"
— bug_historian.** Sustained, and it is 091's own shape one level down: a shared
mechanism whose correctness depends on each adopter remembering something. The
`Warn` moved INTO `refreshOpenWorkItem`, so the fail-loud surface belongs to the
writer. A caller may add its own; it can no longer be the only one.

**3. "Parameterise the status lists, do not interpolate" — constitution.**
Sustained. `sqlInList` exists only because `insertWorkItem`'s `ON CONFLICT … WHERE`
predicate MUST be literal for partial-index inference to resolve `idx_swi_dedup`.
The refresh is a plain `WHERE` and has no such constraint, so the exemption never
extended to it — I had simply copied the sibling. Now `status <> ALL($5::text[])`,
with the literal built the way `depends_on` already is. PREPAREd **and EXECUTEd**
against the live schema (0 rows, as expected) — a PREPARE alone would not have
proven the array literal binds.

**4. "Log when a refresh lands on a `needs_human_review` row" — guardian, echoed
by architecture.** Done. This is the one judgement I flagged as wanting challenge,
and the seats' answer was "not blocking, but make it observable rather than
documented". It now says so every time it happens.

**5. "'368, still growing' contradicts your own drop from 380" —
prior_art_librarian.** Caught a real inconsistency in my submission. Re-measured:
368 open, **50 raised in 7 days, 2 in 24h**. The queue is accumulating at ~7/day
and the net fall is drain outpacing intake. "Growing" was wrong about the count and
right about the intake; the correction is in the bug file, marked.

### Not acted on, deliberately

**tooling_provenance (low): no `doc_notes` row capturing the decision.** It is
covered — the two `LANDMINES.md` entries are synced into `doc_notes` by
`landmines-sync.py --apply` (run), which is the ruled system of record (D10). Writing
a second hand-authored row is explicitly forbidden by that ruling.

**guidelines seat flagged a GUIDELINE GAP rather than objecting:** the documented
work-item dedup rule says "use DELETE+INSERT, not ON CONFLICT", which the entire
existing helper contradicts. Not this change's to resolve — but it is the rule the
`a5b70424` guidelines seat cited against `apply_gap_plan`, so the rule and the
platform have been out of step for a while. Worth someone's RFC; noted, not taken.
