# HANDOFF — live-object declaration drift · **START HERE** (2026-08-26)

**This supersedes `HANDOFF_2026-08-22_continue_here.md`**, which was wrong about the single most
important fact for two days (it said phase 2 was undeployed; it had shipped on 08-23). Read this
first, then `NOTES_…md` (evidence + every misstep) and `RUNBOOK_…md` (commands with their gotchas).

- Lane: `docs/agent_docs/docs024_key_docs_latest/live_object_declaration_drift/`
- Bug: `bugs_open/363_HANDOFF_2026-08-22_go_guards_assert_the_text_of_append_only_migrations_so_they_cannot_fail_when_the_live_object_moves.md`
- Register: **SQLC-002** (`docs026_concept_register/register/sql-change-management.md`)

---

## 1. What this lane is, in one paragraph

Some platform behaviour lives in the database, not in Go: a scheduled task's `pre_query`, a trigger
function, a CHECK constraint, an `agent_definitions` workflow. A Go test cannot see any of it — there
is no database when `go test` runs. Several guards bridged that by reading the **migration file** that
once created the object. That cannot work: a migration is append-only history (`schema_migrations`
checksums the FILE, so an applied migration must never be edited) while the live object accumulates
every later migration. `platform/livespec` is the replacement — the declaration of each guarded live
object, in a file that is **allowed to change** — plus a daily auditor that compares it to the real
database.

## 2. Status — BOTH PHASES LIVE. Verify it yourself; do not trust this line

| | state |
|---|---|
| Phase 1 (Go guards → declaration) | **LIVE** since 2026-08-22, council `b3676918` APPROVED r2 |
| Phase 2 (declaration → live object) | **LIVE** since 2026-08-23 — CronJob `live-declaration-drift-check`, `0 7 * * *` |
| Declarations | **12** as of 2026-08-26 (`LiveAuditOnlyDeclarations = 8`, `MaxDeclarations = 24`) |
| Latest deployed image | `v1.0.1341`; its 08-26 run probed **10** objects, 0 findings |
| Council | `59c08f16` **APPROVED** (08-25 landing) · `80f84c54` **PENDING** (08-26 pairing) |
| `bugs_open/363` | **OPEN**, for the residuals in §5 — *not* for the deploy |

```bash
# the standing health check — one row per run INCLUDING CLEAN ONES,
# so a MISSING row means the job did not run, never "nothing is wrong"
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT created_at, left(body,120) FROM doc_notes WHERE categories ? 'live-declaration-drift' ORDER BY created_at DESC LIMIT 3;"
```

**The probed COUNT is the deploy evidence, not the image tag.** It went `5 → 10` on 08-26 when the
fleet build carried the 08-25 declarations out. It will read **12** after the next release (the 08-26
pairing). If you expect 12 and see 10, the release has not happened yet — that is the check.

## 3. The three things most likely to mislead you

**(a) `FragmentMatch` is presence-only, and `Max` does NOT fix it.** A declaration using
`Mode: FragmentMatch` asserts each declared `Fragment` is PRESENT. So the auditor sees the live object
**lose** a declared value and is **blind to it gaining** an undeclared one — printing the same
`0 finding(s)` either way. A `Max` bounds one value's occurrences, never the **size of the set**, and
a newly added value is in nobody's fragment list. Only a **paired `CountEqual`** sees addition.
Enforced now by `TestEveryFragmentMatchDeclarationIsGainVisibleOrWaived`. *Self-bounding exception:* a
fragment holding a **whole rendered clause including its terminator** (`claimed-item-timeout` declares
`item_type NOT IN ('a', …, 'n')` with the closing paren) already sees addition.

**(b) The live object can misdescribe ITSELF, and the auditor structurally cannot catch that.** The
live `claimed-item-timeout` `pre_query` still carries, above the clause it governs, a comment saying
the contract is *"the LOCKSTEP TWIN of the `RegisterVerifier()` calls"* and naming
`TestRegisteredVerifiersMatchClaimTimeoutExclusion`. Both false since migration `482` (2026-08-19):
the contract is the **union of both completion-gate rosters**, and that test no longer exists.
**That sentence is the original cause of `bugs_closed/317`.** A values-comparison auditor **passes**
it: the clause matches; the prose lies. So this estate's correct standing advice — *read the live row,
the repo file is history* — hands you the wrong contract here.

**(c) `PhaseGoSide` / `PhaseLiveAudit` say WHO CHECKS, not WHETHER anything does.** Both are checked.
`PhaseLiveAudit` means *only the daily auditor reads it, because a unit test has no database* — it is
**not** inert. The doc comment said "INERT" until 08-25; the `bugs_open/375` lane read it backwards and
a council seat lost an objection to it.

## 4. Do not touch — active lanes owning objects cited here

- **`bugs_open/341`** / **`bugs_closed/307`** own the `claimed-item-timeout` `pre_query` (migration `524`).
- **`bugs_open/355`** owns migration `552` (the third `page_components` trigger binding).
- **`bugs_open/375`** owes one line to `livespec.ClaimedItemTimeoutExclusions` (`required_fields_missing`)
  plus a migration. **Deliberately left to them** — registering it fails three arms of CQ-023's router
  and arming `close_converted` fail-closes a live route. Their sequencing, not ours.
- **`bugs_open/333`** — its owed Declaration was **absorbed on 08-25** (`workflow.page-build-handler.refuse_owned_page`).
  It has been told to stand down.

Re-run `./scripts/who-owns.py <n>` **and `git status`** at every phase boundary — the script reads
commits, so a session mid-fix is invisible.

## 5. What is left — why `bugs_open/363` stays OPEN

1. **The two `trigger_fn` body declarations are still gain-blind**, waived in writing. A **fourth
   verdict literal** added to `site_component_history_archive` or `page_component_artefact_archive`
   would be invisible to its entry. `classifySiteComponentArtefact` is the Go mirror that would then
   disagree — that is the tie we have. Closing it needs a different probe shape (counting literals in
   a PL/pgSQL body is noise, not a vocabulary size), and was not attempted.
2. **The stale live prose in (3b)** — needs a migration on a column owned by another lane.
3. **`bugs_open/375`'s registration** — its call.
4. **Council `80f84c54` — RESUBMITTED 2026-08-26 09:01Z, verdict PENDING. Read it, and act on a REVISE.**
   Run orchestration `8a6ad1cf-7762-4591-8f75-9e1f1c57d6e5`; the code is already on the shared branch.
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='80f84c54-1854-4fb6-a003-11af1889d20d' AND kind='council_report' ORDER BY created_at;
   ```
   **⚠ If that returns nothing, check the RUN before assuming it is queued — a DEAD round looks
   identical to a pending one.**
   `SELECT current_step, status FROM orchestration_states WHERE orchestration_id='8a6ad1cf-7762-4591-8f75-9e1f1c57d6e5';`
   `complete_invalid` means it died, and the reason is in `collected_data->>'__step_error'` — the
   `error` column is NULL (`bugs_open/354`).

   **Why the first attempt died (kept, because the diagnosis is reusable): it did not fail review; it
   never got one.**
   Its run ended `complete_invalid` with
   *"no reviewer produced a readable opinion (6 abstained, 11 unreadable)"* — the **fleet-wide
   Anthropic provider outage** of 2026-08-26, not anything about the submission.
   **The `Council-Submitted:` trailer on `083d3096e` stays TRUE** (it asserts submission, never a
   verdict), so nothing needs amending — forward-only holds.
   **Resubmitting is on the SAME correlation so the trail accumulates** (done 09:01Z):
   ```bash
   RESUBMIT_CORR=80f84c54-1854-4fb6-a003-11af1889d20d \
     ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
     docs/agent_docs/docs024_key_docs_latest/live_object_declaration_drift/SUBMISSION_2026-08-26_countequal_pairing.json
   ```
   **Check the provider is actually back FIRST** — at 08:57 UTC it was *flapping*, not down: 67
   `endpoint unavailable` errors in 15 minutes **alongside** 38 successful calls, so a green single
   call proves nothing. Require a clean window:
   ```sql
   SELECT count(*) FROM agent_error_log
   WHERE occurred_at > now() - interval '15 minutes' AND error_message ILIKE '%endpoint unavailable%';
   ```
   ⚠ **`complete_invalid` reads EXACTLY like "still queued"** — no verdict row, no error column
   (`error` is NULL; the reason is in `collected_data->>'__step_error'`). Diagnose it by comparing
   `collected_data` keys against a healthy run: a healthy council has ten `review_*` keys, a dead one
   has **none** and only the `gate_*` relevance steps. That comparison is what distinguishes "the
   seats disagreed" from "the seats never ran".
5. **The 8-vs-7 coverage gap from the original filing** — 8 enabled `scheduled_tasks` rows embed a
   literal vocabulary a Go list could drift from; only some are declared. Never this bug's subject,
   still nobody's.
6. **The unreviewable class** — every check service is `cmd/` + a CronJob, so 17+ daily fleet checks
   are structurally outside council scope. Owner's call; the lever is `scripts/council-scope.sh`.

## 6. Adding a declaration — the checklist that survives review

1. **Probe it live first, with a disconfirming control.** A probe returning the number you wanted is
   evidence only if it could have returned something else. Run the same probe against a deliberately
   wrong path/name and require a different answer. Both go in `Provenance` with the date.
2. **Choose the mode by which direction can fail.** Enumerable vocabulary that can grow → pair a
   `CountEqual` on the value count. Otherwise → write a real waiver in `gainBlindnessWaivers`.
   The guard will refuse a thin waiver, a stale one, and one that duplicates a pairing.
3. **Bump `LiveAuditOnlyDeclarations`** if `Phase: PhaseLiveAudit`. It is asserted in **two** packages
   — `platform/livespec` and `cmd/config-key-audit` — so both move or HEAD breaks.
4. **`go vet ./platform/livespec/ ./cmd/config-key-audit/`.** `go build ./...` **cannot** see a broken
   test file and will tell you a half-landed rename is fine.
5. **`./scripts/verify-head-builds.sh --with <each file> --test`** before committing.
6. **Run the demand control after committing** (§7), and report the clean run *and* the induced one.

## 7. Running the auditor by hand + the demand control

```bash
PW=$(kubectl -n ai-persona-system get secret personae-platform-secrets \
       -o jsonpath='{.data.CLIENTS_DB_PASSWORD}' | base64 -d)
kubectl -n ai-persona-system port-forward svc/postgres-clients 15432:5432 >/dev/null 2>&1 &
go build -o /tmp/cka ./cmd/config-key-audit/
PG_CLIENTS_HOST=127.0.0.1 PG_CLIENTS_PORT=15432 CLIENTS_DB_PASSWORD="$PW" \
  /tmp/cka --live-declaration-drift ; echo "exit=$?"
```

⚠ **Omit `--report` by hand** — with it the run writes a `doc_notes` row identical to the CronJob's,
and that row is the fleet's evidence the *scheduled* job ran.
⚠ **`PG_CLIENTS_PORT` matters** — `dbConn()` defaults to 5432.

**THE DEMAND CONTROL — a clean run proves nothing on its own**, because everything agrees today.
Induce on the **declaration** side, never production, and induce in the direction the instrument
watches:

| induce | required |
|---|---|
| `CountEqual`: declare 7 against the live 8 | exit 1, *"live count is 8, declared 7"* |
| `FragmentMatch`: declare a value that is **not** live | exit 1, naming object + fragment |
| ~~`FragmentMatch`: **remove** a declared fragment~~ | **exit 0 — and that is CORRECT.** Asserting *less* is still satisfied. This is not a broken auditor; it is the wrong induction. |

> ⚠ **RESTORE WITH `cp`, NOT `git checkout`.** The RUNBOOK's recipe ends
> `git checkout platform/livespec/livespec.go`. On this shared tree that **silently discards any
> uncommitted work in that file** — mine, on 2026-08-26, ~20 minutes after I wrote the landmine
> warning about it. `cp <file> /tmp/x.bak` first, restore from that, and
> `git status --porcelain -- <path>` before any checkout. Note `go test -run <name>` prints
> **`ok … [no tests to run]`** when the function is gone, so a deleted test reads as a passing one.

## 8. Facts a new session should not re-derive

[All measured 2026-08-26 unless dated otherwise. **A census goes stale by ADDITION** — re-check before quoting.]

- **12** declarations; **8** auditor-only; `MaxDeclarations = 24`.
- Live counts: `doc_notes_subject_type_check` **8** values, `doc_plans_subject_type_check` **6**
  (deliberately narrower — a landmine has no shared-contract shape to put in a plan; migration `273`
  refuses to rebuild `doc_notes` without `'landmine'`).
- `page-content-writer` `slot_name_from` = **2** occurrences; `build-site-planner` predicate = **1**;
  `page-build-handler` `refuse_owned_page` = **1** row (bogus-path control: **0**).
- `pg_trigger`: **3** triggers bound to `page_component_artefact_archive`; migration `357` declares 2.
- `scheduled_tasks` (2026-08-22): **97** rows, **47** enabled, **35** with a `pre_query`, **24**
  enabled with one, **8** embedding a literal vocabulary.
- The `platform/orchestration/actions` suite has been **RED at bare HEAD** (undeclared
  `WORK_ITEM_STATUS_OVERRIDE_REFUSED`, `bugs_open/358`) — reproduced with **no overlay**, so if you see
  it, check whether it is yours before debugging. On this tree a red package is as likely to be
  another session's half-finished commit as a real regression.

## 9. Missteps this lane has already paid for — all in `WRONG_CALLS.md`

1. Grepped one spelling of `RegisterVerifier` and invented a 12-vs-14 discrepancy. *Build your set the
   way the guard builds its set.*
2. Dispatched `090` with no `SEED_SCOPE` on a symptom whose evidence lives in test files; burned all
   five iterations, returned `UNVERIFIABLE`.
3. A line-oriented completeness check found 7 readers where there were 9. *Count it the way the
   consumer counts it.*
4. **`go build ./...` said fine while a renamed symbol had HEAD half-broken** — it does not build test
   files. `go vet` names it in one line.
5. **Read a correct instrument as a broken one** by inducing drift in the direction `FragmentMatch`
   does not watch. *A control that does not fire is a claim about the control, not yet about the
   system* — and chasing it produced this lane's best finding.
6. **Ran `git checkout` on my own uncommitted work**, 20 minutes after writing the landmine against it.
7. **Submitted a council plan that asserted ~80% of the diff it claimed.** APPROVED, but the objection
   was right. *A submission is evidence, not a summary of evidence.*
