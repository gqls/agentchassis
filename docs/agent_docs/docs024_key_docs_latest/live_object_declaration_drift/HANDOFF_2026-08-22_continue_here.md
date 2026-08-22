# HANDOFF — live-object declaration drift · start here

**Written 2026-08-22** at the end of the founding session. **This is the lane's COLD-START doc.**
Read this, then `NOTES_live_object_declaration_drift.md` (evidence + every misstep) and
`RUNBOOK_live_object_declaration_drift.md` (the commands, with their gotchas).

Lane directory: `docs/agent_docs/docs024_key_docs_latest/live_object_declaration_drift/`
Bug: `bugs_open/363_HANDOFF_2026-08-22_go_guards_assert_the_text_of_append_only_migrations_so_they_cannot_fail_when_the_live_object_moves.md`

---

## 1. What this lane is, in one paragraph

Some platform behaviour lives in the database, not in Go: a scheduled task's `pre_query`, a trigger
function, a CHECK constraint, an `agent_definitions` workflow. A Go test cannot see any of it — there
is no database when `go test` runs. Several guards bridged that by reading the **migration file** that
once created the object and asserting Go agrees with it. That cannot work: a migration is append-only
history (`schema_migrations` records a checksum of the FILE, so an applied migration must never be
edited), while the live object accumulates every later migration. Three guards were therefore
asserting something the checksum rule had already made impossible; two more `t.Skipf`'d to green when
their file was unreadable.

## 2. Status — READ THIS BEFORE ANYTHING ELSE

**◑ BOTH PHASES BUILT. PHASE 2 IS NOT DEPLOYED. `bugs_open/363` is OPEN and should stay open.**

| | state |
|---|---|
| Phase 1 (Go guards → declaration) | **BUILT, council-APPROVED, mutation-proven** — `873575ecf` |
| Phase 2 (declaration → live object) | **BUILT and proven against the live DB — `18661b3c7` — but NOT DEPLOYED** |
| Deploy | **image unbuilt/unpushed, CronJob unapplied. Nothing runs on a schedule.** |
| Council | phase 1 corr `b3676918…` APPROVED r2. **Phase 2 REFUSED as out of scope** — not forced. |
| Register | **SQLC-002** · LANDMINES entry live |

**THE NEXT ACTION IS NOT CODE — IT IS A DEPLOY, AND IT IS THE OWNER'S.** Releases here are
whole-fleet (`make release`). `make build-live-declaration-drift-check` is wired and builds from
committed HEAD. Saying this plainly because "INERT until the roll" is itself a documented trap: it
makes the correct next action look premature. It is not premature; it is simply not a session's to
take.

**After the release, verify AT THE ARTEFACT — a green make target is not a deployed check:**

```bash
kubectl -n ai-persona-system get cronjob live-declaration-drift-check \
  -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}'
kubectl -n ai-persona-system create job --from=cronjob/live-declaration-drift-check ldd-manual-1
# then read the POD's exitCode — a Job is not a pod, and a log line is not an exit code
kubectl -n ai-persona-system get pod -l job-name=ldd-manual-1 \
  -o jsonpath='{.items[0].status.containerStatuses[0].state.terminated.exitCode}'
```
Expect **0** and one `doc_notes` row with `categories ? 'live-declaration-drift'`. Exit **2** means
it could not look (that is correct behaviour, not a pass). A **missing** row means the job did not
run — never "nothing is wrong".

### What phase 2 proved before it was committed (both outcomes, one session)

| | induced | got |
|---|---|---|
| clean | nothing | `probed 5 live object(s) (2 scheduled_task, 1 trigger_bindings, 2 trigger_fn); 0 finding(s)` exit 0 |
| D1 | drop a declared item type | exit 1, names object **and** fragment |
| D2 | declare 2 trigger bindings vs a live 3 | exit 1 `live count is 3, declared 2` |
| D3 | probe a nonexistent task | **exit 2, NOT clean** |
| D4 | no database configured | **exit 2, NOT clean** |

**D2 is the important one** — that declaration was INERT in phase 1 (the council's `bug_historian`
objection) and is now demonstrably read.

### ⚠ The council gate refuses this whole CLASS, and that is worth the owner's attention

Phase 2 was REFUSED: *"no edit touches the review scope"*. Every file is `cmd/`, `build/`,
`deployments/` or `makefile`. **Every check service in this family has that shape, so 17+ daily fleet
checks as of 2026-08-22 are structurally unreviewable.** `FORCE=1` would buy a review the owner's
scope ruling says should not be bought, so this lane did not. The lever is
`scripts/council-scope.sh` (single-sourced), and it is the owner's call — see `bugs_open/363`
§"the class of checks nobody reviews".

## 3. The single most important thing to know

**The live object can misdescribe ITSELF, and a values-comparison auditor will pass it.**

The live `scheduled_tasks.pre_query` for `claimed-item-timeout` still carries, above the clause it
governs:

> *"The item_type exclusion is the LOCKSTEP TWIN of the `RegisterVerifier()` calls … 
> `TestRegisteredVerifiersMatchClaimTimeoutExclusion` pins the two together."*

Both are false since migration `482` (2026-08-19): the contract is the **union of both completion-gate
rosters**, and `grep -rn "func TestRegisteredVerifiersMatchClaimTimeoutExclusion"` returns **nothing**.
**That sentence is the original cause of `bugs_closed/317`** — its author read it and built a
gate-2-only lockstep — and it is still there. So "read the live row, the repo file is history", this
estate's correct standing advice, hands you the wrong contract. A phase-2 auditor comparing live text
to a declared clause **passes** this: the clause matches; the prose lies.

**Owed, small, unclaimed:** a migration correcting that comment in the live `pre_query`. Not done
because that column belongs to active lanes (see §6). Whoever next edits it should carry the fix.

## 4. What phase 1 actually shipped

`platform/livespec/livespec.go` — the declaration of each guarded live object in a file that is
**allowed to change**, with its `ProbeSQL` (how phase 2 will read it) and required/forbidden
fragments. **5** declarations as of 2026-08-22. Keyed on the **live object**, never the migration —
`506` writes two live objects, so a file-keyed registry would reproduce the defect inside the fix.

Four guards converted to read it, both `Skipf`s gone:

| file | was |
|---|---|
| `platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go` | globbed `*_claimed_item_timeout_generic_evidence.sql`, parsed a **comment** |
| `platform/orchestration/actions/complete_work_item_retry_guard_test.go` | read `506`, `t.Skipf` on missing |
| `platform/orchestration/actions/page_component_divergence_test.go` | read `357`, `t.Skipf` on missing |
| `platform/orchestration/actions/site_component_divergence_test.go` | read `344` |

`platform/livespec/livespec_test.go` — self-guards plus
`TestNoNewMigrationFileReadersOutsideTheAllowList`, which scans **AST string literals, never comments**
and requires both a `sql_for_agents` literal and a read call. Allow-list = **5** files as of
2026-08-22, each with a written reason.

`DeferredDeclarations` counts entries nothing can check yet (**1** today: the trigger **binding
count**) so an inert declaration cannot read as guarded.

## 5. Phase 2 — BUILT (this section is now the record of how, not a to-do)

A read-only `--live-declaration-drift [--report]` mode on the **existing** `cmd/config-key-audit`
binary, plus a daily Go check image on the **`shared-output-fields-check` pattern** (an image, NOT
RFC_006's Python mirror — that service's own header explains why), plus `RELEASE_IMAGES`.

**Verified reuse facts, so you need not re-derive them:**
- `dbConn()` at `cmd/config-key-audit/fleetdb.go:52`, `writeDocNote()` at `:110`. Both exist.
- **No** existing mode reads `scheduled_tasks` — `grep 'scheduled_tasks' cmd/config-key-audit/*.go` → **0 hits** as of 2026-08-22.
- `UnknownConfigKeys`/`ListDeclaredConfigKeys`/`ListRemovedConfigKeys` (`datahelpers/action_inputs.go:247/435/335`) do key-**set** membership over `agent_definitions`, not text matching. No engine to extend.
- Of **17** `*-check` services as of 2026-08-22, none probes a trigger, CHECK constraint or `pre_query`.
- Council scope (`scripts/council-scope.sh:57-61`) is `^(platform|internal|pkg)/` + appliable migrations. **`cmd/`, `scripts/`, `deployments/`, `makefile` are OUT** — logic belongs in `platform/`, thin `main` in `cmd/`.
- **A new check service MUST join `RELEASE_IMAGES` in the commit that creates it** (`makefile:78-92`; two already fell into that hole and the coverage gate structurally cannot see it).

**Constraints that killed the alternatives** (do not re-propose them): `go test` has no database; a
pre-commit hook cannot gate live config because **at commit time the migration is unapplied**
(RFC_006, owner 2026-08-02); live-object checksums page on every lawful direct edit.

**Phase 2 needs its OWN council round.** The r2 approval covers phase 1 only.

## 6. Do not touch — active lanes owning objects cited here

- **`bugs_open/341`** and **`bugs_closed/307`** — the `bugfix_307_terminal_write_contract` lane owns
  the `claimed-item-timeout` `pre_query` (migration `524`).
- **`bugs_open/355`** — owns migration `552`, which added the third `page_components` trigger binding.

Both are cited as **evidence** in `bugs_open/363` and were not modified. Re-run
`./scripts/who-owns.py <n>` **plus `git status`** at every phase boundary — the script reads commits,
so a session mid-fix is invisible.

## 7. Verify the state before trusting this document

```bash
# phase 1 still green (run from the repo, NOT a scratchpad extract — see below)
go test ./platform/livespec/ ./platform/orchestration/actions/ -count=1

# the tripwire must be capable of failing: add a throwaway sql_for_agents reader
# to any platform test, re-run, expect it to fire and NAME the file, then revert.
```

⚠ **`git archive HEAD` into a scratchpad**: the tmpfs at `/tmp` is shared (~30 sessions) and a full
`/tmp` makes commands **look failed when they succeeded**. This box's scratchpad is on
`/dev/nvme0n1p2`, not tmpfs (`df` it before relying on either) — that is why the phase-1 runs in NOTES
are sound. A council seat raised this; it was checked, not argued.

⚠ **A RED `platform/orchestration/actions` SUITE IS PROBABLY NOT YOURS.** Checked 2026-08-22 at the
end of this session: the shared working tree carried another lane's in-flight change and the package
panicked with `Expected number of values to match number of columns: expected 5, actual 6` in
`TestRerenderPageSections_StructuralCarryMakesANotReadySectionRerender` — a sqlmock column-count
mismatch from the **357 lane** adding `page_components.component_version_id` (RFC_046). That lane also
has an uncommitted anchor widening in `page_component_divergence_test.go`, a test this lane never
touched.

**Committed HEAD was GREEN**, verified in a clean `git archive HEAD` extract, and all five of this
lane's tests passed in the dirty tree too. So before debugging a red suite: run **your** tests by
name, then extract HEAD and run the failing test there. On this tree a red package is as likely to be
someone else's half-finished commit as a real regression.

## 8. Open items, honestly listed

1. **Phase 2** — designed, not built, needs its own council round. The main job.
2. **The stale live prose** (§3) — a migration correcting the `claimed-item-timeout` comment.
3. **The landmine verifier returned `UNVERIFIABLE`** for the entry (corr `71f6262c`) — *"all footprint
   items … fall outside the Go-symbol-only code index"*. That is a **limitation of the verifier**, not
   a defect in the entry: the footprint is SQL objects, doc directories and test filenames. Do not
   re-run it expecting a different answer, and do not read the non-verdict as a problem.
4. **Two guards remain repo-side by design** — `v3_render_slot_name_test.go` (a seed lint) and
   `links_shipped_predicate_test.go` (migration 302 is the operator's paste source). Both keep genuine
   repo value; their **live** tie is phase 2. They are allow-listed with reasons.
5. **`bugs_open/363` stays OPEN** until phase 2 lands.

## 9. Three missteps recorded in `WRONG_CALLS.md` — read them, they were all cheap to avoid

1. **Grepped one spelling of `RegisterVerifier`** and invented a 12-vs-14 discrepancy. The registry has
   two writers (`RegisterVerifier`, `RegisterVerifierWithPolicy`) because RFC_017 shipped the policy
   form as an opt-in field — the estate's own good practice created the second spelling. *Build your
   set the way the guard builds its set.*
2. **Dispatched `090` past its own `dispatching blind` warning**, with no `SEED_SCOPE`. A symptom whose
   evidence is in **test files** has no runtime symbol, so the fallback symbol search found nothing
   relevant and burned all five iterations. The run returned `UNVERIFIABLE` — recorded as such, never
   as support.
3. **My completeness check was line-oriented** and found 7 readers where there are 9, missing a file
   this very bug names as broken (`filepath.Join(...)` and `os.ReadFile(path)` are on different lines).
   *Count it the way the consumer counts it* — the tripwire walks the AST, so the honest check was to
   run the tripwire.

## 10. Facts a new session should not re-derive

[All MEASURED 2026-08-22 — re-check any figure before quoting it; a census goes stale by ADDITION.]

- **7** guards in the class; **all 7** live objects checked and **all 7 AGREE**. There is **no damage**. Filed for the door.
- **9** readers of a `sql_for_agents` path in `platform/**`: 4 converted, 5 allow-listed. Four more mention it only in comments/fixtures.
- Migration `220` was edited **nine times after it applied**; `322`, `331`, `374` each amended the live exclusion clause under filenames the old glob could not match; `524` edited the same column.
- `pg_trigger`: **3** triggers bound to `page_component_artefact_archive`; migration `357` declares **2**.
- `scheduled_tasks`: **97** rows, **47** enabled, **35** with a `pre_query`, **24** enabled with one; **8** enabled rows embed a literal vocabulary.
- `schema_migrations` checksums the **FILE**. Nothing in the repo reads `pg_get_functiondef`/`pg_trigger`/`information_schema.triggers` — **0** hits.
