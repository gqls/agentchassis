# HANDOFF — continue here: `bugs_open/375`, the unguarded completion writer

**Rewritten 2026-08-25 (end of session)** by the lane's first working session. Supersedes
`HANDOFF_2026-08-25_continue_here.md` (same day, morning) and `HANDOFF_2026-08-24_start_here.md`
(the original cold start) — both kept for the record, neither current.

> **In one line:** three council-approved changes shipped, nothing is pending, and **the bug is still
> OPEN and should be**. What is left is §3 — and only one item there is blocked on anything.

---

## 1. What this bug is, in plain terms

A **work item** is one recorded defect on one site. A **verifier** re-runs that defect's own
predicate immediately before the item is stamped `complete`, and refuses the stamp if the defect is
still there. It exists to stop a handler reporting success without having fixed anything.

There are **three** writers of `complete`, and they have never agreed:

| writer | consults the verifier? |
|---|---|
| `CompleteWorkItemAction` | yes, always |
| `UpdateWorkItemStatusAction` | **never did.** Since 2026-08-24: only when that STEP sets `verify_before_complete: true`. **No step does.** |
| the `claimed-item-timeout` sweep | **no** — writes the row directly; held off a type only by `livespec.ClaimedItemTimeoutExclusions` (`bugs_closed/317`) |

`verifier_coverage_test.go` goes green regardless of any of that. **That is the bug**, and it is
still true today for every completion through writer 2.

## 2. Verified state `[2026-08-25, each with how it was checked]`

| thing | state | how |
|---|---|---|
| `WII-030`'s gate in the binary | **LIVE**, chassis `v1.0.1337`, both pods, started 09:27Z | literal probe of `/proc/1/exe` with a must-be-PRESENT and a must-be-ABSENT control. ⚠ the Go identifier `updateStatusVerifyConfigKey` is correctly **absent** — never probe for that |
| steps arming `verify_before_complete` | **0** of 200 live agents | recursive `jsonb_path_query` |
| declared unarmed `complete` arms | **6**, and the live set MATCHES | `go run ./cmd/config-key-audit --unarmed-verified-completers` → `[]`, exit 0 |
| that clean result | **DEMAND-CONTROLLED both ways** | drop a real entry → `undeclared`; add a ghost → `stale`; both exit 1, same input |
| item types reachable by those arms | **7**, none with a verifier | live **UNION `site_work_items_archive`** — the live table alone says 5 |
| `verifier_not_consulted` rows | **0 — and UNINFORMATIVE** | the demand control is empty: no verifier exists for any of the 7, so the record *cannot* fire. Not a pass, not a fail |
| the verifier | written, **unregistered**, mutation-proven ×4 | §3a |
| council | **3 submissions, all APPROVED** (one after a REVISE) | §4 |
| the bug | **OPEN, correctly** | §5 |

⚠ **Snapshots on a tree many sessions share. Re-run before acting on any of it.**

### 2a. ⚠ NOTHING FROM 2026-08-25 NEEDS A CHASSIS ROLL — do not wait for one

The chassis is `v1.0.1337`, built at **09:27Z**, i.e. **before** today's commits. That is fine and
it is worth being explicit, because the reflex on this estate is "Go change ⇒ wait for the roll":

- **Candidate 4's Go half is a TEST.** It runs at build time. It is effective now, for anybody who
  runs `go test`. Nothing at runtime reads `livespec.UnarmedVerifiedCompleters`.
- **The audit mode is a CLI** run from the repo (`cmd/config-key-audit`), not a deployed service.
  Effective now.
- **The verifier is UNREGISTERED**, so although it compiles into the chassis it is unreachable at
  runtime. A roll ships it as dead weight and changes nothing.

The only thing that ever needed a roll was `WII-030`'s gate, and that shipped in `v1.0.1337`.

## 3. WHAT IS LEFT — three items

### 3a. Register the verifier — five steps, and only step 1 is blocked

The verifier for `required_fields_missing` is **written, mutation-proven and council-approved, and
deliberately not switched on**. The full sequence lives in
`discovery_checks/verify_required_fields_missing.go`'s own header — **that header is the runbook**,
and the LANDMINE *"Registering a verifier is NOT a one-line change"* carries it too.

| # | step | state |
|---|---|---|
| **1** | Add `required_fields_missing` to `livespec.ClaimedItemTimeoutExclusions` **AND ship a migration amending the live `scheduled_tasks.pre_query`** for `claimed-item-timeout` | **BLOCKED — see below** |
| 2 | Scope-test licence in `write_audit_findings_verifier_join_test.go` `optedIn` | ✅ **DONE** |
| 3 | Remove the type from `itemTypesWithoutVerifiers` | not done (correctly — it is still unregistered) |
| 4 | Lifecycle posture for the file | ✅ **DONE** (`PostureObserves`) |
| 5 | `WII-031`'s lockstep will fail on **three** arms of `CQ-023`'s router — arm or acknowledge each | not done |

**Why step 1 is first and load-bearing.** The `claimed-item-timeout` sweep writes `site_work_items`
directly. Until the **LIVE** clause excludes this type, that sweep completes items **straight past
the new verifier** — `bugs_closed/317` reintroduced *by the act of adding a guard*. A false green is
worse than a missing one. **The council accepted this argument explicitly** (round 2, after a HIGH
objection that an unregistered verifier changes nothing), so the sequence is the reviewed plan, not
a session's preference.

**What blocks it, WHO holds it, and how to tell when it clears.** Step 1 edits
`platform/livespec/livespec.go`, held by the **`live_object_declaration_drift` lane**
(`bugs_open/363`, still OPEN) — identified from the diff's own content, since uncommitted work is
invisible to `who-owns.py`. **Re-checked 2026-08-25 15:34Z: still dirty** (`livespec.go` since
~10:50Z, `livespec_test.go` since 08-23). A pathspec commit of a shared file takes their
half-written work as a passenger.

⚠ **Two reasons beyond the passenger rule not to force it.** Their change introduces
`LiveAuditOnlyDeclarations = 5`, **asserted by `livespec_test`** — appending to a file whose counted
invariant is mid-flight is how that number goes quietly wrong. And their lane is the *authority* on
the phase-2 fact this lane used to overturn a council objection, so letting them land first means
the two accounts agree rather than race.

**A note is already filed in their account** (`bugs_open/363`, CONTRIB 2026-08-25, commit
`ac7c75c9b`): it states the one-line ask, says we are deliberately not making the edit, and asks
nothing of them — `375`'s next session adds its own line once the file is free.

~~⚠ **`go test ./platform/livespec/` FAILS, and it is NOT theirs and NOT yours.**
`TestNoNewMigrationFileReadersOutsideTheAllowList` fails identically at committed HEAD…
A third lane's breakage. Do not debug it.~~

> **RESOLVED 2026-08-25 16:31Z, and this line was stale within about an hour of being written —
> which is the point.** The `bugs_open/333` lane fixed it at `559e60bd0`: their door test no longer
> reads migration 488's file, the frozen path is a Go literal (a checksummed migration cannot
> change, so the copy cannot go stale), every Go-side assertion kept, and **neither `livespec` file
> was touched** — they left the allow-list alone precisely because it is `363`'s dirty WIP.
> **Verified here at committed HEAD rather than taken on their word:**
> `scripts/verify-head-builds.sh --test ./platform/livespec/ ./platform/orchestration/actions/` →
> both `ok` at `559e60bd0`. The other breakage (`advisedIdentityPin` in `actions`) is resolved too.
> **So run those tests normally; a red result now is yours.**
>
> ⚠ **The lesson is about this file, not about them.** I spent this session correcting three other
> people's statements that outlived their truth — a "phase 2 has not shipped" comment, `CQ-023`'s
> fail-close warning, a landmine footprint naming a deleted symbol — and then wrote one myself that
> rotted in **under ninety minutes**. A "known broken, do not debug" note is the most perishable
> thing a handoff can carry and the most confidently obeyed. **Date every one of them, and re-run
> the check before believing your own file.**

**Check before starting:**
```bash
git status --porcelain -- platform/livespec/     # empty = clear to proceed
```

**Then verify at the LIVE object, never at the Go declaration:**
```sql
SELECT pre_query LIKE '%required_fields_missing%' AS live_clause_has_it
FROM scheduled_tasks WHERE name = 'claimed-item-timeout';
```
```bash
go run ./cmd/config-key-audit --live-declaration-drift
```

⚠ **Step 5 has its own ordering.** Arming `close_converted` is the one `CQ-023` warns fail-closes a
live route. **Arm `close_stale` and `close_resolved` first**, read that router's close paths, and
treat `close_converted` as a separate decision — or take the acknowledgement arm, which is why
`WII-031` demands a reason rather than a switch.

**Once registered AND an arm is armed**, this becomes the live proof for the first time:
```sql
SELECT count(*) FROM site_work_items WHERE result->'_verification'->>'status' = 'verifier_not_consulted';
```
A zero there *once armed* is a real fault. A zero **now** means nothing — the record cannot fire.

### 3b. Candidate 2 — unify the two writers (the real fix, architecture-scope)

**Not started, and the honest next step is a scoping read rather than a proposal.**

Everything this lane shipped *manages* the asymmetry — an opt-in switch, a runtime marker, a build
guard, a declaration, a drift audit. Five mechanisms describing one duplication. Unify the writers
and all five stop being necessary, because "which door did this go through" stops being a question.

This lane did its **first half** without setting out to: both paths now share one gate
implementation and one row read (`loadWorkItemVerifyRow`). What nobody has done is **read
`CompleteWorkItemAction`'s call sites** to judge whether the merge is feasible — its inputs come
through `datahelpers.ExtractActionInputs` with a `CompleteWorkItemInputSpec`, while
`UpdateWorkItemStatusAction` reads `params.StepConfig.Config` directly and declares no spec at all.
Until somebody reads that, "have the second delegate to the first" is a sentence, not a plan.

⚠ **Do not ship it inside a bug patch.** `bugs_closed/124` drew a REJECTED verdict for exactly that
shape, and the owner ruling of 2026-07-29 puts a change to what a shared completion path guarantees
at architecture scope. Precedent for how to do it: `bugs_closed/284` (duplicate writers unified with
a structural single-definition test to stop a fourth copy appearing). Destination:
`architecture_review/` as an RFC that either proposes the merge or argues it down.

### 3c. Optional — put the new audit mode on a cadence (an owner call)

`--unarmed-verified-completers` fires only when somebody runs it. If the `undeclared` direction is
worth catching automatically it belongs beside the other nightly `config-key-audit` CronJobs.
Recorded in `WII-031`'s verify-later. **Not a session's call.**

## 4. What shipped, and the council record

| submission | what | verdict |
|---|---|---|
| `7a6add95` | `WII-030` — the opt-in verifier gate on writer 2, plus the bypass record | **APPROVED r1** (12 seats), 2 medium objections, both acted on |
| `3083d182` | `WII-031` — candidate 4: the build-time lockstep + live-drift audit | **APPROVED r1** (10 seats) |
| `c8ed18c1` | `WII-032` — the `required_fields_missing` verifier | **r1 REVISE → r2 APPROVED** (13 seats, *"all reviewers approve"*) |

**Round 1 of `c8ed18c1` earned its keep and is the most transferable thing in this lane.**
`SchemaContentFields` returns `ok=true` with **ZERO** fields for a v2 schema whose `fields` object is
empty (`component_schema_fields.go:78`). `missingRequiredValueFields` then found nothing missing and
the verifier said **Resolved** — so a component whose field declarations had been **emptied** (the
silent-loss class of `bugs_open/012`, `/021`) would have been **certified as repaired by the guard
written to catch it**. Fixed (`43277271a`), fail-closed under RFC_017.

**The asymmetry is the lesson:** a DETECTOR's `continue` on an unreadable declaration is correct; a
VERIFIER running the identical arithmetic certifies a repair. ⚠ The fix is deliberately **local** —
measured: `SchemaContentFields` has **10** non-test callers and **exactly one resolves**; hardening
the shared helper would break the nine that legitimately want the permissive reading.

## 5. Why the bug is still OPEN

The bar is **fixed AND live**. The gate is live. The bug is **not fixed**: every completion through
`update_work_item_status` is still unverified, no arm is armed, and the two writers are still two.

What changed is that the trap now **fails the build** instead of only leaving a row marker, that a
verifier exists for somebody to register, and that three documents which used to mislead now tell the
truth. **None of that makes the bug's own claim false.** Closing on "the mechanism shipped" is the
`bugs_open/021` §INSTANCE 2 error one level along.

## 6. Traps this lane has paid for — read before touching anything

1. **`site_work_items` is a ROLLING WINDOW.** UNION `site_work_items_archive` for any
   "all-history"/"ever" claim. Cost this lane a figure published to six files, **and a positive
   control that shared the blind spot**. (`WRONG_CALLS.md`; bug §8.)
2. **Enumerate what a helper can RETURN, not what you expect it to return.** The `{"fields":{}}`
   defect sat in the gap between two conditions I had already handled. **Mutate PER SHAPE** — three
   of four shapes in the fix's test were already caught, and an aggregate mutation would have hidden
   that only one was the hole.
3. **grep LANDMINES for the SYMBOL you are about to build on.** The entry predicting that defect was
   written 2026-08-03 and its footprint named the very file whose predicate I was reusing. The
   SessionStart hook only matches files already dirty, so a shared helper is never shown to you.
4. **A mock cannot assert SQL text.** sqlmock returns whatever you queued regardless of the
   statement — put the column list or predicate in the **expectation**. Caught twice here.
5. **Mutate, don't trust green.** M1–M14 on record. ⚠ The terminal-decision guard sits in SERIES on
   this arm, so fixtures must be `detected`/`claimed`/`triaged` or a mutation reads as covered.
6. **Do not register a verifier from a test `init()`** — the registry has no removal and a
   cross-package contract test reads it. Use the `verifierLookup` seam.
7. **A pathspec commit takes SAME-FILE passengers, possibly half-written.** Twice the right move was
   a new file (`livespec/unarmed_completers.go`) or stopping at the file boundary (step 1). **The
   tell: a deletion count in `git diff --numstat` you cannot account for.**
8. **A stale document is read as ground truth by reviewers.** THREE instances in two days — a
   `phase-2 auditor has not shipped` sentence, `CQ-023`'s fail-close warning, and a landmine
   footprint naming a symbol that no longer exists. All three were correct when written; all three
   misled a council seat. ⚠ Naming trap the first exposed: **`PhaseGoSide` is the CHECKED state and
   `PhaseLiveAudit` is the INERT one.**
9. **The code index is pinned at `e347c5ad` (2026-08-23)**, so a landmine-verifier verdict on newer
   files reports "0 rows" as staleness, not absence.
10. **Do NOT repair a working-tree build failure caused by another session's WIP.** On 2026-08-25 the
    tree would not build because a committed test named a symbol another session's *uncommitted*
    rename had removed. Pointing the test at the new name fixes the tree and **breaks HEAD**, because
    the rename is not committed — a peer lane did exactly that (`6d3e0027e`, HEAD broken ~3h49m,
    restored `8b9128131`) under a commit message reading "builds again". **The diagnosis, not a
    discipline:** `scripts/verify-head-builds.sh --test <pkg>` builds committed HEAD alone. **HEAD
    green ⇒ the breakage is somebody's WIP and is not yours to fix.** This lane hit the identical
    failure hours earlier and used `--with <my files>` instead, which is the only reason it never
    recurred here. Recorded fleet-wide as the fourth occurrence of the HEAD-vs-tree class in
    `LANDMINES.md` (`c021e52c3`).
11. **A "known broken, do not debug" note is the most perishable thing you can write** — mine rotted
    in under 90 minutes (§3a). Re-run the check before believing this file.
12. **Other lanes broke HEAD twice on 2026-08-25 — both since FIXED** (`559e60bd0` and the `actions`
    one); kept as a dated event because the lesson is the habit, not the state: (`TestNoNewMigrationFileReadersOutsideTheAllowList` from
    `bugs_open/333`'s file; `advisedIdentityPin` undefined in `actions`). If a package fails for you,
    check whose it is before debugging. `scripts/verify-head-builds.sh --with <your files>` is how
    this lane tested throughout.

## 7. What is NOT established — do not quietly inherit it

- **Whether any of the 7 types SHOULD have a verifier.** Two are on `verifier_coverage_test.go`'s
  `catMechanical` backlog; that is an invitation, not a judgement.
- **Whether the 578 unguarded completions contain FALSE completions.** Nobody has re-run those
  predicates. `bugs_open/367` found one by accident — that is one, not a rate. Its own measurement,
  probably its own `090`.
- **Whether candidate 2 is feasible.** Cited by shape from `bugs_closed/284`, never verified.
- **Whether a shared "guard exclusion" struct should cover writers 2 and 3.** A council seat raised
  it twice, low severity; **unactioned and stated as such**, not resolved.
- **Whether `098` has credited these commits.** Trailers are present and well-formed on all of them;
  the report is slow and was never run to completion. `098_REPORT_unreviewed_commits_v1.sh 3`.

## 8. Where everything is

| what | where |
|---|---|
| the bug | `bugs_open/375_HANDOFF_2026-08-23_…md` — read **§7c, §8, §9, §10** |
| `WII-030`'s gate | `platform/orchestration/actions/update_work_item_status_verification.go` |
| candidate 4 (`WII-031`) | `platform/livespec/unarmed_completers.go` · `platform/orchestration/actions/unarmed_completer_lockstep_test.go` · `cmd/config-key-audit/unarmedcompleters.go` |
| the verifier (`WII-032`) | `platform/orchestration/actions/discovery_checks/verify_required_fields_missing.go` (+ `_test.go`) — **its header IS the registration runbook** |
| the third writer's precedent | `platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go`, `bugs_closed/317` |
| this lane's standing five | `PLAN_2026-08-24_…`, `RUNBOOK_…`, `NOTES_…`, `README_where_we_are.md`, `SUMMARY_2026-08-24_…` |
| register | `WII-030`, `WII-031`, `WII-032`; `CQ-023` (corrected) |
| landmines | two entries: *"…protects a type only on the paths that ASK"* and *"Registering a verifier is NOT a one-line change"*; plus a corrected footprint on the *"unreadable config computes to HEALTHY"* entry |
| wrong calls | `WRONG_CALLS.md` — three entries, 2026-08-24 ×2 and 2026-08-25 |
| the `image_url_404` finding | `bugs_open/033`, CONTRIB 2026-08-25 — **not** a bug of its own; the premise was refuted |
