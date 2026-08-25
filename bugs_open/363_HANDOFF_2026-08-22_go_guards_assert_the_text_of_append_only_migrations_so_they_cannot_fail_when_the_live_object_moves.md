# 363 — seven Go guards assert the TEXT of an append-only migration file, so they cannot fail in the direction they exist to detect

> ## ◑ BOTH PHASES BUILT 2026-08-22. Phase 2 is NOT DEPLOYED, so this stays OPEN.
>
> **Phase 1** (`873575ecf`, council `b3676918` **APPROVED r2**): `platform/livespec` holds each guarded
> live object's declaration in a file ALLOWED to change; **4** guards assert against it instead of a
> frozen migration; both `t.Skipf` silent-greens gone. Register **SQLC-002**.
>
> **Phase 2** (`e0ff0a5b`-era commit, see `git log -- cmd/config-key-audit/livedeclarations.go`):
> `--live-declaration-drift [--report]` on the existing `cmd/config-key-audit` binary + a check image +
> a CronJob at **07:00 UTC**. **Proven end-to-end against the live database, both outcomes in one
> session:** clean = *probed 5 live objects, 0 findings, exit 0*; D1 drop a declared type → **exit 1**
> naming object and fragment; D2 declare 2 trigger bindings against a live 3 → **exit 1** *"live count
> is 3, declared 2"*; D3 nonexistent probe → **exit 2, not clean**; D4 no database → **exit 2, not
> clean**. D2 is the one that matters: that declaration was INERT in phase 1 and is now demonstrably read.
>
> **⚠ WHY THIS IS STILL OPEN — the image is NOT built or pushed and the CronJob is NOT applied.**
> Nothing runs on a schedule yet. Releases here are whole-fleet and the owner runs `make release`;
> `make build-live-declaration-drift-check` is wired and builds from committed HEAD. **After the
> release, verify at the artefact** (`kubectl get cronjob live-declaration-drift-check`, then read a
> run's exit code and its `doc_notes` row) — a green make target is not a deployed check.
>
> **⚠ The council gate REFUSED phase 2 as OUT OF SCOPE** (every file is `cmd/`/`build/`/`deployments/`/
> `makefile`; scope is `platform/`, `internal/`, `pkg/` + migrations). Not forced — the scope is an
> owner ruling. See §"the class of checks nobody reviews" below.
>
> **⚠ Even fully deployed, this does NOT catch everything.** It compares live TEXT to declared
> fragments, so it **passes** the live `claimed-item-timeout` `pre_query` whose own COMMENT states a
> contract superseded on 2026-08-19 and names a deleted test — the sentence that caused
> `bugs_closed/317`. The clause matches; the prose lies.
>
> **Unchanged and not oversold:** all **7** live objects measured 2026-08-22 AGREE. No drift exists.
> Filed for the door, not the damage.

**Filed 2026-08-22** by the `live_object_declaration_drift` lane, found while re-verifying
`bugs_closed/317`. **OPEN, UNOWNED. LATENT — 7 of 7 live objects measured and all seven agree;
filed for the door, not the damage** (the same basis `317` itself was filed on).

Workstream docs: `docs/agent_docs/docs024_key_docs_latest/live_object_declaration_drift/`
(`NOTES_…md` carries every measurement and command; `RUNBOOK_…md` the queries).

## What the thing IS, then the rule, then this case

**The thing.** Some of this platform's behaviour lives in the database, not in Go: a scheduled
task's `pre_query`, a trigger function, a CHECK constraint, an `agent_definitions` workflow. A Go
unit test cannot see any of it — there is no database when `go test` runs.

**The workaround we adopted.** Seven guards bridge that gap by reading the **migration file** that
once created the object, and asserting the Go side agrees with that file's text.

**The rule that breaks it.** A migration must never be edited once applied — `schema_migrations`
stores its checksum, so editing one makes that record a lie. Migration `482` states this in its own
header. **So the file is frozen, and the live object is not.**

**This case.** The guard therefore asserts something that policy has already made impossible.
`TestDivergenceVocabularyMatchesMigration357` fails with *"migration 357 no longer contains %q"* —
but 357 cannot stop containing it. The test is green today and would stay green if the live
trigger were redefined tomorrow. It is not a weak guard against live drift; it is a guard against a
different event, wearing the error message of the first.

## The census — **7** guards as of 2026-08-22, 4 kinds of live object, all measured

[MEASURED 2026-08-22, each against the live DB. Every one could have come out otherwise.]

| # | guard | live object | live check | result |
|---|---|---|---|---|
| 1 | `TestClaimTimeoutExclusionCoversBothCompletionGates` | `scheduled_tasks.claimed-item-timeout.pre_query` | `regexp_match` on the column | ✅ 14 = union of both gate rosters |
| 2 | `TestGoAndSQLAgreeOnTheCooldownBoundary` (506) | `scheduled_tasks.build-pipeline-trigger.pre_query` | grep for the rendered predicate | ✅ 1×, verbatim, non-strict `<=` |
| 3 | `TestDivergenceVocabularyMatchesMigration344` | fn `site_component_history_archive` | `pg_get_functiondef` | ✅ 3 verdicts + md5 clause |
| 4 | `TestDivergenceVocabularyMatchesMigration357` | fn `page_component_artefact_archive` | `pg_get_functiondef`, 4 needles | ✅ 1 occurrence each |
| 5 | `TestValidDocSubjectTypes_LockstepWithMigrationCheck` | 2 `subject_type` CHECK constraints | `pg_get_constraintdef` | ✅ Go 8 = union(6, 8) |
| 6 | `TestRenderSlotNameConfigKeyMatchesTheSeededWorkflow` | `agent_definitions` `page-content-writer` | count in live `default_config` | ✅ 2, seed demands ≥2 |
| 7 | `TestMigration302CarriesTheCanonicalPredicateVerbatim` | `agent_definitions` `build-site-planner` | grep live config | ✅ 1×, verbatim |

**Nothing has drifted.** A future session that finds drift has a NEW and much louder finding, not a
confirmation of this one.

## Why it is still a defect, in three parts that do not need damage

**(a) Three of the seven cannot fail.** 3, 4 and 7 assert `strings.Contains(<frozen file>, needle)`.
Policy forbids the only edit that could make that false.

**(b) Two pass when the file is merely renamed.** `TestGoAndSQLAgreeOnTheCooldownBoundary` and
`TestDivergenceVocabularyMatchesMigration357` call **`t.Skipf`** on a read error. A rename is a
silent green — and renames happen: see (c).

**(c) The live object already outgrows the file the guard reads.** `pg_trigger` returns **three**
triggers on `page_components` bound to `page_component_artefact_archive`:
`trg_page_component_artefact_archive_upd`, `…_del` (both declared in `357`) and
`trg_page_component_content_archive_upd` — **added by `552`, which `357` knows nothing about.**
Migration `524` did the same thing to instance 1's object under a filename its glob does not match.
⚠ **`552` belongs to the ACTIVE `bugs_open/355` lane and must not be touched** — it is cited here
only as evidence that live objects accumulate past the file a guard reads. Its author did nothing
wrong; the guard's premise is what is wrong.

## The inference the whole class rests on, in its author's own words

`links_shipped_predicate_test.go:100-104`:

> *"If migration 302's file is ever archived/moved, move this assertion onto its successor rather
> than deleting it — **the LIVE row was written from that text**, and this test is the only thing
> tying the row's predicate to the builder."*

*Live equals file, because the live object was once written from the file.* True at the instant of
application; it decays silently from then on, and nothing re-establishes it.

## Confirmed absences (both disconfirming, both came back negative)

- **`schema_migrations` records a checksum of the FILE**, never of the live object
  (`scripts/migration/run-migrations.sh:265,453`). It answers *was this file applied*, never *does
  the live object still match what we believe*.
- **Nothing in the repo reads a live object's definition at all.**
  `grep -rn "pg_get_functiondef\|pg_trigger\|information_schema.triggers"` over `*.go|*.py|*.sh`,
  excluding `sql_for_agents`: **zero hits.** There is no live-object drift check of any kind.

## Blast radius, as queries rather than arguments

- `scheduled_tasks`, all **as of 2026-08-22**: **97 rows, 47 enabled, 35 carry a `pre_query`, 24 enabled ones do.** Of the 35,
  **8** enabled rows as of 2026-08-22 embed a literal quoted vocabulary a Go list could drift from
  (`claimed-item-timeout`, `database-cleanup`, `detected-item-promoter`,
  `held-pair-canary-escalation`, three `site-discovery-rotation-*`, `site-render-audit-rotation`).
- The 7 guards above are the ones that *try*. The 8-row figure is the population that *could* need
  one. The gap between 7 and 8+ is not this bug's subject but is worth someone's attention.

## ⚠ Two fixes that look right and are not

1. **"Point the guard at the newest matching migration."** Instance 1 already does this and it did
   not help: `524` and `552` both edited the live object under filenames the relevant globs do not
   match. A glob over an append-only corpus is a guess about a naming convention.
2. **"Make the tests query the live database."** `go test` runs inside `git archive HEAD` in a clean
   context with no cluster. And a pre-commit hook cannot do it either — **at commit time the
   migration is unapplied**, which this estate already ruled on (`RFC_006`, owner 2026-08-02; the
   answer there was a marker in code plus a **daily CronJob** against the real system).

## Prior art, checked before filing

- `bugs_open/136` — neighbour, not duplicate. Its surface is `agent_definitions` step-config **keys
  nothing reads**; this is a live SQL/DDL object whose only guard reads an untied repo copy. It
  brings reusable machinery: `cmd/config-key-audit`, `scripts/audit-config-keys.sh`.
- `bugs_open/341` — the only open bug on the `claimed-item-timeout` sweep, **OWNED and ACTIVE**
  (`who-owns.py` → `bugfix_307_terminal_write_contract`). Untouched.
- `needs_diagnosis` queue: **0 rows** at `awaiting_diagnosis` when this was filed.
- LANDMINES: the `doc_plans`/`doc_notes` `subject_type` entry documents instance 5's two-enforcement-
  point shape independently.

## The best existing implementation is already in the tree, and is the thing to generalise

Instance 5 does **not** pin a filename. `newestConstraintValues`
(`doc_subjects_common_test.go:73-120`) walks the whole `sql_for_agents` directory, parses each
numeric prefix, and keeps the **newest** migration that recreates the named constraint — then
asserts Go equals the **union** of the two tables' newest declared value sets. That removes the
glob-guessing and filename-pinning that instances 1–4, 6 and 7 all suffer from.

What even it lacks is the last step: **asking the live database what it actually contains.**

## The drift channel is PROVEN, not hypothesised

[MEASURED 2026-08-22, from `git log --follow` and the migration texts]

`220_claimed_item_timeout_generic_evidence.sql` was **edited nine times after it first applied**
(created `d61b3ace1`; then `ac9f75a0c`, `ec8ad7959`, `96dd3015c`, `af2667453`, `dc4f4e6b2`,
`a60a13cbb`, `ad51ca863`, `d644723b8`, `c121d5a73`). For most of its life the file **was** the
mutable declaration and was maintained as one — which is why a guard reading it made sense. The
checksum-freeze convention later took that away, and **nothing replaced it.**

Since the freeze, the live exclusion clause has been edited by migrations the guard's glob
(`*_claimed_item_timeout_generic_evidence.sql`) **cannot see** — each doing
`SET pre_query = replace(...)` naming the previous clause verbatim:

| migration | glob match? |
|---|---|
| `322_dead_fragment_link_claim_timeout_exclusion.sql` | **NO** |
| `331_literal_markdown_claim_timeout_exclusion.sql` | **NO** |
| `374_decision_regression_claim_timeout_exclusion.sql` | **NO** |
| `524_claimed_item_timeout_honours_the_cooldown.sql` (same column) | **NO** |
| `482_claimed_item_timeout_generic_evidence.sql` | yes — all the guard ever reads |

**The guard is correct today only because `482`'s author hand-reconciled every earlier edit into a
comment.** That is a hand-maintained copy of a live vocabulary, which is the class this estate keeps
filing bugs about — and it is one careful author away from being wrong.

## The LIVE object documents a contract that was superseded three days ago — and names a deleted test

This is the sharpest instance in the file, and it was found by the diagnosis loop's Tier-1 citation
of the live column rather than by reading any repo file.

This estate's standing rule is **"the repo file is history, the live row is fact — read the live
row."** Do that here, and `scheduled_tasks.pre_query` for `claimed-item-timeout` tells you, in the
comment block immediately above the exclusion clause it governs [MEASURED 2026-08-22, live column]:

```
-- The item_type exclusion is the LOCKSTEP TWIN of the RegisterVerifier()
-- calls in platform/orchestration/actions/discovery_checks/: those item
-- types have a Go verifier that can BLOCK completion (bugs_open/017, /021)
-- and SQL cannot run it, so they keep falling through to reset.
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion pins the two together.
```

**Both sentences are now false.**

1. The contract has been the **UNION of both completion-gate rosters** since migration `482`
   (2026-08-19) — `excluded ⇔ (has a verifier) OR (has a noChangeGates entry)`. "LOCKSTEP TWIN of
   the `RegisterVerifier()` calls" is precisely the gate-2-only contract that `bugs_closed/317`
   exists to have corrected.
2. **`TestRegisteredVerifiersMatchClaimTimeoutExclusion` no longer exists.**
   `grep -rn "func TestRegisteredVerifiersMatchClaimTimeoutExclusion"` returns nothing; it was
   replaced by `TestClaimTimeoutExclusionCoversBothCompletionGates`, which had to move package to
   see both rosters. The live object names a deleted test as its own guarantee.

**Why this is the worst version of the defect, not a footnote.** That exact sentence is *the cause
of 317* — the guard's author read "LOCKSTEP TWIN of the `RegisterVerifier()` calls", believed it,
and built a gate-2-only lockstep. `482` fixed the **data** (the 14-type list) and left the
**sentence** in place. So the next author who does the correct thing — reads the live object rather
than the repo file — is handed the same wrong contract that produced the original bug, now with a
citation to a test they cannot find.

It also closes off the one escape hatch a reader might hope for: it is not merely that the repo file
is stale while the live object is true. **The live object carries its own stale declaration**, and
nothing checks that either. A drift auditor that compares live text to a declaration would not catch
this on its own — the *clause* matches; it is the *prose around it* that lies. Any fix should treat
an object's embedded contract statement as part of what is declared, or say plainly that it does not.

## Fix candidates, ordered by what closes the door

**1. RECOMMENDED — give the current declaration a home that is allowed to change, and tie both ends
to it.** A leaf package `platform/livespec` holds each guarded live object's *current* declaration
(value lists, literal fragments, a renderer for the exclusion clause) plus its identity and a probe
SQL. Unit tests import it — compile-time constants, no DB, no file I/O, no glob — closing the
Go↔declaration leg and letting the tests stop reading `sql_for_agents` for these objects entirely.
A new read-only mode on the existing `cmd/config-key-audit` binary probes the live objects and
compares them to livespec, closing the declaration↔live leg, and runs daily as a check image on the
**`shared-output-fields-check` pattern** — a Go image, not RFC_006's Python mirror, because that
service's own header records that an image "dissolves … both of RFC_006's named drift risks (a
`DECLARED_*` literal kept in step by hand, and a parity test to notice when it is not)". *Makes
unrepresentable:* a guard whose spec is a frozen file, and live drift that stays silent beyond ~24 h.
*Costs:* one package, one mode, one check service, and a standing duty — a migration touching a
guarded object updates livespec in the same commit. *Leaves open:* see the residual below.

**2. Patch each guard** (widen globs, `Skipf`→`Fatalf`, "newest file" logic for 344/357). Cheap, and
makes nothing unrepresentable: the file is still the spec, a `replace()` migration still carries no
full declaration, live drift stays invisible, and the next guard author copies the pattern. **Rejected
as primary — but the two `Skipf`→`Fatalf` one-liners are worth taking regardless.**

**3. Live-object checksums recorded at apply time.** Rejected: `scheduled_tasks` and
`agent_definitions` are *legitimately* edited straight in the database here, so a hash-vs-last-migration
pages on every lawful edit; and a hash says "something changed", not "the fragment the guard depends
on changed". Same cost as (1) without its precision.

**4. Let the unit tests read the live DB.** Violates the constraint outright — `go test` runs inside
a `git archive` context with no cluster. Listed so the council sees it was considered and why not.

**5. Do nothing for the repo-only cases.** Correct, and adopted: `write_experience_pattern_test.go`
(schema DDL, where the accumulated corpus genuinely is the canonical channel) and
`contact_info_no_fabrication_test.go` (the migration is a render *fixture*, not a spec-of-live claim,
and the file names its live watcher at lines 23-25) are **excluded**.

### What the recommendation does NOT cover, stated plainly

- **A ≤24 h detection window** between a drifting edit and the next run. Accepted: RFC_006's ruling
  is that after-the-fact-on-a-clock is the only mechanism that can see database-side change at all.
- **Objects nobody enrols.** Mitigated by a tripwire test scanning `platform/**/*_test.go` for new
  `sql_for_agents` readers outside a **reasoned** allow-list — scanning call arguments, not comments
  (the "source-scan test makes comments load-bearing" landmine), and every allow-list entry carrying
  a written reason (an allow-list without one "converts a live debt into a false all-clear").
- **The job not running.** One `doc_notes` row per run *including clean ones*, so a MISSING row reads
  as "the job did not run", never as "nothing is wrong" — and the RUNBOOK carries the standing query,
  because detection works while schedule and dispatch do not.
- **A declaration wrong in the same way the live object is wrong.** Requires simultaneous wrong edits
  to the Go roster, livespec and the live object — strictly smaller exposure than today's.

**This closes drift DETECTION, not drift PREVENTION.** Nothing can prevent it: these objects are
legitimately mutable in production by design. The tests stop lying immediately; the live tie becomes
eventually-consistent with a stated, monitored bound.

### Constraints any candidate must satisfy

- It cannot require a database at `go test` time, and it cannot be a pre-commit hook — **at commit
  time the migration is unapplied** (`RFC_006`, owner 2026-08-02).
- It must key on the **live object**, with the migration as an attribute of it. `506` writes **two**
  live objects (a `scheduled_tasks` row and an `agent_definitions` row) and its guard checks only the
  first — so a registry keyed on the file reproduces the defect inside the fix.
- New logic belongs in `platform/` (reviewable under `scripts/council-scope.sh:57-61`, and importable
  by both a unit test and the auditor) with only a thin `main` in `cmd/`; `cmd/` and `scripts/` are
  outside council scope.
- **A new check service must join `RELEASE_IMAGES` in the commit that creates it** (`makefile:78-92`
  — two services were already born outside it, and the coverage gate structurally cannot see that).

## How to verify a fix

Read the **artefact**, and make the check disconfirmable:

- **Induce the drift.** Change a live object out from under its declaration in a transaction
  (`BEGIN; ALTER …; <run auditor>; ROLLBACK;`) and require the auditor to report it. An auditor that
  has never been shown failing is not evidence it can fail — a clean run over an estate where
  nothing has drifted is exactly the uninformative green this bug is about.
- **A post-fix zero needs a demand control.** All 7 objects agree today, so "the auditor found
  nothing" is the expected result whether it works or not. The induced case above IS the demand
  control; run both and report the pair.
- If the fix ships a CronJob, remember the standing trap: **re-apply the kustomize overlay or the
  cluster keeps the old literal**, and one `doc_notes` row must be written **per run including clean
  ones**, so that a MISSING row reads as "the job did not run" rather than "nothing is wrong".


## The class of checks nobody reviews (found 2026-08-22, recorded for the owner, not decided)

Submitting phase 2 to the council gate returns:

```
REFUSED: no edit touches the review scope.
  In scope: platform/, internal/, pkg/ ... plus migrations
```

Correct, per the scope ruling — and it generalises past this bug. **Every check service in this
family is a `cmd/` binary plus a CronJob, so the entire class of daily fleet checks — 17+ services as
of 2026-08-22 — is structurally unreviewable by the gate.** The 2026-08-19 widening admitted
migrations on the argument that *"a migration IS the running system, live the moment it applies, with
no image tag to roll back."* A check service is weaker on that test: it has a tag and can be rolled
back. But it is still deployed code making daily assertions about production, and a check that is
subtly wrong reports clean for ever.

`FORCE=1` would buy a review the ruling says should not be bought, so this lane did not use it. If
the owner wants this class covered, the lever is the scope definition in `scripts/council-scope.sh`
(single-sourced — 097, the commit-msg nudge and the 098 report all read it), not a per-submission
override.

## Relations

`bugs_closed/317` (where this was found; its fix is sound and re-verified live 2026-08-22) ·
`bugs_open/136` (the neighbouring live-config-vs-code class, and its reusable auditor) ·
`bugs_open/341`, `bugs_open/355` (active lanes owning objects cited here as evidence — do not touch) ·
`RFC_006` (the owner ruling that a pre-commit hook cannot gate live config; `SingleOwner` + daily
CronJob) · `RFC_022` / WFA-013 (`audit-optional-key-budget.sh`, the worked auditor+cron+parity-test
shape) · LANDMINES `Dedup index ↔ Go list lockstep` and the `doc_subjects subject_type` entry (the
same "two hand-maintained copies of one vocabulary" family).

---

## CONTRIB 2026-08-25 — a QUEUED consumer for `ClaimedItemTimeoutExclusions`, waiting on this lane's uncommitted work

From the `bugfix_375_completion_verifier_gap` lane. **Not a finding against this bug, and not a
request to change anything here** — a heads-up that a second lane is queued behind one line of
`livespec.go`, plus one thing of yours we independently confirmed.

### What we need, and why it waits

`bugs_open/375` has a completion verifier for `required_fields_missing` — written,
mutation-proven, council-approved (`c8ed18c1`, r1 REVISE → r2 APPROVED) and **deliberately not
registered**. Registering it fails five build guards, and the load-bearing one is yours:

> `TestClaimTimeoutExclusionCoversBothCompletionGates`: *"has a registered verifier (gate 2) but is
> NOT declared in `livespec.ClaimedItemTimeoutExclusions` … Add it to the declaration AND ship a
> migration amending the live pre_query — the declaration alone changes nothing in production."*

So `375` owes exactly one entry in that slice plus a migration on the live
`claimed-item-timeout` `pre_query`. **Verified still outstanding 2026-08-25 15:34Z:**
`grep -c required_fields_missing platform/livespec/livespec.go` → **0**, and
`SELECT pre_query LIKE '%required_fields_missing%' FROM scheduled_tasks WHERE
name='claimed-item-timeout'` → **f**.

**We have not made that edit, on purpose.** `livespec.go` and `livespec_test.go` have been dirty in
the shared tree all day (first seen 10:50Z, still dirty 15:34Z; `livespec_test.go` since 08-23), and
a pathspec commit of a shared file takes the other session's work as a passenger. **Two further
reasons specific to your change**, which is why we are asking rather than editing:

1. Your uncommitted work introduces **`LiveAuditOnlyDeclarations = 5`**, asserted by `livespec_test`.
   A second session appending to the same file while a counted invariant is mid-flight is how that
   number goes quietly wrong.
2. At 10:50Z the package did not compile (`DeferredDeclarations` undefined); by 15:13Z it did. So the
   change looks finished but unlanded, and landing it is entirely your call.

**All we ask: when you commit, no action is needed from you.** `375`'s next session checks
`git status --porcelain -- platform/livespec/` and adds its own line. This note exists so that if
someone else touches that slice first, they know a consumer is queued.

⚠ **One thing worth knowing if you DO add it for us:** `375`'s own guard (`WII-031`,
`unarmed_completer_lockstep_test.go`) will then fail on three arms of `CQ-023`'s router, and arming
`close_converted` fail-closes a live route. That sequencing belongs to `375`, not here.

### And one thing of yours we confirmed independently, before your rename lands

Your uncommitted comment says phase 2 is deployed and that leaving a constant called `Deferred`
*"would have been this lane's own defect, a written statement outliving its truth"*. **We reached the
same conclusion from the other side, and it had already cost something.**

`claim_timeout_exclusion_lockstep_test.go`'s header still said *"nothing compares livespec to the
LIVE `scheduled_tasks.pre_query`. That is the phase-2 auditor, and until it ships…"*. A council
`prior_art_librarian` seat quoted that sentence back as a **MEDIUM objection** against a correct
claim in `375`'s candidate-4 round (corr `3083d182`, 2026-08-25). We checked rather than argued:
`livespec.Declarations` carries `scheduled_task.claimed-item-timeout.exclusions` with its `ProbeSQL`,
and `compareAllDeclarations` iterates **every** declaration with no `Phase` filter
(`livedeclarations.go:129`). The claim was right; the comment was stale. **Corrected in place**
(`08a44365f`), with what the staleness cost recorded beside it.

⚠ **A naming trap we nearly inverted, and which your rename only half-fixes:** `PhaseGoSide` is the
**CHECKED** state and `PhaseLiveAudit` is the **INERT** one — the opposite of what the names suggest.
We briefly read it backwards, which would have made the seat right. Renaming `DeferredDeclarations`
→ `LiveAuditOnlyDeclarations` is a real improvement; the two `Phase` constants still read backwards
to a stranger. **Your call, your lane** — recorded because we lost time to it and a reviewer lost an
objection to it.

**Relations:** `bugs_open/375` §10a/§10e · register `WII-030`/`WII-031`/`WII-032` ·
`bugs_closed/317` (why the exclusion list is load-bearing at all) ·
`docs024_key_docs_latest/bugfix_375_completion_verifier_gap/HANDOFF_2026-08-25b_continue_here.md` §3a

## CONTRIB 2026-08-25 (from the `bugfix_333_owned_page_door` lane) — your allow-list guard caught my test at HEAD; fixed on my side without touching your dirty files, and one livespec Declaration is now OWED once your rename lands

`TestNoNewMigrationFileReadersOutsideTheAllowList` was failing at committed HEAD on
`platform/orchestration/actions/work_item_owned_page_door_test.go` (mine, 08-24): it read migration
488's SQL to pin the door's jsonpath against the `{workflow,steps,load_page_record,config,refuse_owned_page}`
path 488 wrote. Your guard is right — the file-read half of that assertion can never fail. Fixed by
quoting the frozen path as a Go literal (a copy of a checksummed artefact cannot go stale) and keeping
every Go-side assertion; **I did not add an allow-list entry, because `livespec_test.go` and
`livespec.go` are both yours and dirty**, and the 375 lane's CONTRIB above already records why nobody
should touch them mid-rename.

**Owed to `platform/livespec`, after your rename lands — a Declaration for the door's live object:**
`agent_definitions` type `page-build-handler` must carry `workflow.steps.load_page_record.config.refuse_owned_page = true`
(the key the WII-028 door reads via `jsonb_path_exists`; migration 488 wrote it). Its live tie is your
daily auditor; its Go tie is the door test above. I will file it as a separate small file in the
package (the 375 lane's `unarmed_completers.go` precedent) so it does not collide with your file — but
only after `LiveAuditOnlyDeclarations` lands, since it changes the counted set. If you would rather add
it yourself in the same commit as the rename, the Declaration is one entry; say so in this file and I
will stand down. Nothing else needed from you.
