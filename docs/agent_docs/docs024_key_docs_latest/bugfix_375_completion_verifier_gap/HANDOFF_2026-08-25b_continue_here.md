# HANDOFF — continue here: `bugs_open/375`, after the owner's four rulings

**Written 2026-08-25 (evening)** by the lane's first working session. Supersedes
`HANDOFF_2026-08-25_continue_here.md` (same day, morning — keep it for its §5 traps).

> **In one line:** the gate is LIVE and inert by design; **candidate 4 is BUILT and council-APPROVED**;
> a verifier is **WRITTEN and deliberately UNREGISTERED** behind a five-step prerequisite whose first
> step is a live migration. The bug is still OPEN and should be. What is left is §4.

---

## 1. The owner's rulings, 2026-08-25, and what each one produced

| ruling | what was done |
|---|---|
| **1. Do not close 375 until it is fixed** | Recorded in bug §9c/§9d with the reasoning and what *would* let it close. Nothing was closed. |
| **2. Build candidate 4** | **DONE.** `livespec.UnarmedVerifiedCompleters` + a build-time lockstep + `config-key-audit --unarmed-verified-completers`. Council `3083d182` **APPROVED r1**. Register `WII-031`. |
| **3. Explain candidate 2 further** | Explained in chat; not started. It remains architecture-scope. §4c below. |
| **4. Write a verifier** | **WRITTEN, NOT REGISTERED.** `required_fields_missing`, mutation-proven. Register `WII-032`. Registration is a five-step sequence written into the verifier's own file; step 1 is a live migration. §4a below. **Council round 1 → REVISE, and it found a REAL defect** (an empty field declaration computed to "nothing missing" and would have certified an emptied schema as repaired). Fixed `43277271a`; **round 2 verdict PENDING — READ IT FIRST**, §2a below. |
| **plus: file the `image_url_404` bug** | **NOT FILED — the premise was refuted.** See §5. Contributed into `bugs_open/033` instead (commit `243684746`). |

## 2. Verified state `[2026-08-25, each with how it was checked]`

| thing | state | how |
|---|---|---|
| `WII-030`'s gate in the binary | **LIVE, `v1.0.1337`, both pods** | literal probe on `/proc/1/exe` with a must-be-present and a must-be-absent control; the Go identifier `updateStatusVerifyConfigKey` is correctly **absent** |
| steps arming `verify_before_complete` | **0** of 200 live agents | recursive `jsonb_path_query` |
| declared unarmed `complete` arms | **6**, and the live set MATCHES | `go run ./cmd/config-key-audit --unarmed-verified-completers` → `[]`, exit 0 |
| that clean result | **DEMAND-CONTROLLED both ways** | dropping a real entry → `undeclared`; adding a ghost → `stale`; both exit 1, same input |
| item types reachable by those arms | **7**, none with a verifier | live **UNION `site_work_items_archive`** — the live table alone says 5 |
| the verifier | **written, unregistered, mutation-proven ×3** | swapping the lifecycle axis fails 6 tests; neutering `Grades` fails the 213 assertion; making an unreadable schema resolve fails the fail-closed test |
| council | `3083d182` **APPROVED r1** (candidate 4) · `c8ed18c1` (verifier) — **READ IT** | `orchestration_states` by payload |
| the bug | **OPEN, correctly** | bug §9c |

⚠ **Snapshots on a shared tree. Re-run before acting.**

## 2a. ⚠ THE FIRST THING TO DO: read the pending verdict

`c8ed18c1-a694-4c80-afdc-12274634fbd2` — round 1 **REVISE** (13 seats), round 2 resubmitted on the
same correlation (orch `27f7bc39`) and **still running when this was written.**

```sql
SELECT orchestration_id, current_step, status, created_at FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'c8ed18c1-a694-4c80-afdc-12274634fbd2'
 ORDER BY created_at;              -- TWO rows: round 1 COMPLETED, round 2 is the later one
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id = 'c8ed18c1-a694-4c80-afdc-12274634fbd2' AND kind='council_report'
 ORDER BY created_at;              -- one report per round
```
⚠ **Two rounds share one correlation** (that is what `RESUBMIT_CORR` is for), so `LIMIT 1` on the
report table gives you the LATEST — order by `created_at` and check which round you are reading.

**Round 1's finding is already fixed** (`43277271a`): `SchemaContentFields` returns `ok=true` with
ZERO fields for a v2 schema whose `fields` object is empty, so the predicate computed "nothing
missing" and would have certified an emptied schema as REPAIRED. It now errors, fail-closed. If
round 2 raises something new, the code is already on the shared branch — act on it, do not wait.

**What round 2 argues, so you can judge whether the seats accepted it:** the gating objection was
that an unregistered verifier leaves the defect unchanged. True as a description. The counter is
that registering it without step 1's migration is strictly *worse* than inert — the
claimed-item-timeout sweep would complete items straight past it, a false green rather than a
missing one — and that round 1 is itself the argument for this order, since a defect was caught
before the predicate ever graded a live claim. **If the owner would rather take that risk and arm it
now, that is a legitimate call and it is theirs, not a session's.**

## 3. What is committed

`c735bfd9c` `c94212ad3` `721465601` `35257bee2` `b6aa4853b` `e88cd0e4f` `b0c066ac5` (WII-030 and its
docs) · `243684746` (033 contrib) · `3083d182`'s code (candidate 4) · `64645d05e` (the verifier) ·
`034d43bbc` (pattern-check allow-list) · `08a44365f` (a stale comment corrected) · `7ae01c648`
(WII-031/032) · `8b700ec88` (the landmine).

## 4. WHAT IS LEFT — in the order a next session should take it

### 4a. Register the verifier — FIVE steps, and step 1 is load-bearing

The sequence is written into `discovery_checks/verify_required_fields_missing.go`'s header, and the
landmine *"Registering a verifier is NOT a one-line change"* carries it too. Summary:

1. **THE MIGRATION, FIRST.** Add `required_fields_missing` to `livespec.ClaimedItemTimeoutExclusions`
   **and ship a migration amending the live `scheduled_tasks.pre_query`** for `claimed-item-timeout`.
   Until the LIVE clause excludes the type, that sweep writes `site_work_items` directly and will
   complete items **straight past the new verifier** — `bugs_closed/317` reintroduced by the act of
   adding a guard. ⚠ **Step 1 edits `platform/livespec/livespec.go`, which is why this stopped here:**
   another session had four hunks of in-flight, non-compiling work in it on 2026-08-25. **Check that
   file is clean before starting.** Verify at the live object, not the Go declaration:
   `SELECT pre_query LIKE '%required_fields_missing%' FROM scheduled_tasks WHERE name='claimed-item-timeout';`
   then `go run ./cmd/config-key-audit --live-declaration-drift`.
2. Scope-test licence — **already done** (`write_audit_findings_verifier_join_test.go` `optedIn`).
3. Remove the type from `itemTypesWithoutVerifiers`.
4. Lifecycle posture — **already done** (`PostureObserves`).
5. `WII-031`'s lockstep will then fail on **three** arms of `CQ-023`'s router. Arm or acknowledge each.
   ⚠ **Arm `close_stale` and `close_resolved` before `close_converted`** — CQ-023 records that the
   `converted` arm fail-closes once a verifier exists for this type.

Then, and only then, the live proof becomes available:
`SELECT count(*) FROM site_work_items WHERE result->'_verification'->>'status'='verifier_not_consulted'`
can be non-zero for the first time — and a zero once registered **and** armed is a real fault.

### 4b. Candidate 2 — unify the two writers (architecture-scope, the real fix)

Not started. `bugs_closed/284` is the precedent. This lane did its first half: both paths share one
gate implementation and one row read (`loadWorkItemVerifyRow`). **Nobody has read
`CompleteWorkItemAction`'s call sites**, so the honest first step is a scoping read, not a proposal —
and the owner ruling of 2026-07-29 plus `bugs_closed/124`'s REJECTED verdict are both about exactly
the shape of shipping this inside a bug patch.

### 4c. Optional: put the new audit mode on a cadence

`--unarmed-verified-completers` is a CLI and fires only when somebody runs it. If the `undeclared`
direction is worth catching automatically it belongs beside the other nightly `config-key-audit`
CronJobs. **An owner call, not a session's** — noted in `WII-031`'s verify-later.

## 5. The `image_url_404` bug was NOT filed, and why that is the right answer

The morning handoff carried an `[OBSERVED, NOT DIAGNOSED]` note: 42 rows with an empty
`handler_agent` while the handler built for them had handled none — read as a dispatch defect.
**Both halves were wrong.**

- The empty handler is **deliberate and documented three times** in `check_image_url_404.go`
  (`:274`: *"HandlerAgent intentionally empty — flag-only. Repairing a stale reference means removing
  or repointing it, which no image generator can decide."*).
- The handler has **not** handled 0 rows — it handled **3**. They were archived out of the rolling
  window.

And the useful half: on the one occasion three rows were hand-assigned to that handler, **it
escalated all three straight back to `needs_human_review`.** So the obvious remedy — "give every
`HandlerAgent: \"\"` type a dispatch route" — is refuted on this type by direct evidence.
`bugs_open/033`'s header already asks exactly this contract question and says two council seats
disagreed, so the finding went **into that file** (`243684746`), not into a new bug.

## 6. Traps this lane has paid for

1. **`site_work_items` is a ROLLING WINDOW.** UNION `site_work_items_archive` for any
   "all-history"/"ever" claim. It cost this lane a figure published to six files, and a positive
   control that shared the blind spot. (`WRONG_CALLS.md`, bug §8.)
2. **Enumerate what a helper can RETURN, not what you expect it to return.** `SchemaContentFields` returns `ok=true` with a ZERO-LENGTH map for `{"fields":{}}`; I had handled unparseable JSON and `ok==false` and the gap between those two conditions was the defect. Mutate PER SHAPE — three of four shapes in the fix's test were already caught, and an aggregate mutation would have hidden that only one was the hole.
3. **grep LANDMINES for the SYMBOL you are about to build on.** The entry that predicted this defect was written 2026-08-03 and its footprint names this type's detector, whose predicate I was reusing. The SessionStart hook only matches files already dirty, so a shared helper is never shown to you.
4. **A mock cannot assert SQL text.** sqlmock returns whatever you queued regardless of the
   statement — put the column list or predicate in the **expectation**. Measured twice: values-only
   assertions passed mutations that dropped a column and that swapped the lifecycle axis.
5. **Mutate, don't trust green.** M1–M13 on record. ⚠ The terminal-decision guard sits in SERIES on
   this arm, so fixtures must be in `detected`/`claimed`/`triaged` or a mutation reads as covered.
6. **Do not register a verifier from a test `init()`** — the registry has no removal and a
   cross-package contract test reads it. Use the `verifierLookup` seam.
7. **A pathspec commit takes SAME-FILE passengers, possibly half-written.** Twice today the right
   move was a new file (`livespec/unarmed_completers.go`) or stopping at the file boundary (step 1).
8. **A stale comment is read as ground truth by reviewers.** A sentence in
   `claim_timeout_exclusion_lockstep_test.go` saying phase 2 was unbuilt cost a council seat a MEDIUM
   objection against a correct claim. Corrected in place with the cost recorded. ⚠ And the naming
   trap it exposed: **`PhaseGoSide` is the CHECKED state; `PhaseLiveAudit` is the INERT one.**
9. **The code index is pinned at `e347c5ad` (2026-08-23)**, so a landmine-verifier verdict on newer
   files reports "0 rows" as staleness, not absence.
10. **HEAD was broken by another lane today** (`TestNoNewMigrationFileReadersOutsideTheAllowList`,
   from `bugs_open/333`'s test file). Not this lane's; the session editing `livespec.go` is likely
   fixing it. If `./platform/livespec/...` fails for you, check whose it is before debugging.

## 7. Where everything is

| what | where |
|---|---|
| the bug | `bugs_open/375_HANDOFF_2026-08-23_…md` — read **§7c, §8, §9, §10** |
| WII-030's gate | `platform/orchestration/actions/update_work_item_status_verification.go` |
| candidate 4 | `platform/livespec/unarmed_completers.go` · `platform/orchestration/actions/unarmed_completer_lockstep_test.go` · `cmd/config-key-audit/unarmedcompleters.go` |
| the verifier | `platform/orchestration/actions/discovery_checks/verify_required_fields_missing.go` (+ `_test.go`) — **its header IS the registration runbook** |
| the third writer's precedent | `platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go`, `bugs_closed/317` |
| this lane's five | `PLAN_2026-08-24_…`, `RUNBOOK_…`, `NOTES_…`, `README_where_we_are.md`, `SUMMARY_2026-08-24_…` |
| register | `WII-030`, **`WII-031`**, **`WII-032`**; `CQ-023` (corrected) |
| landmines | two entries: *"Registering a verifier protects a type only on the paths that ASK"* and *"Registering a verifier is NOT a one-line change"* |
| wrong calls | `WRONG_CALLS.md`, two entries dated 2026-08-24 |
