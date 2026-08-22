# 363 — seven Go guards assert the TEXT of an append-only migration file, so they cannot fail in the direction they exist to detect

> ## ◑ PHASE 1 FIXED, LIVE-IN-CODE and council-APPROVED 2026-08-22 — the guards no longer read history. PHASE 2 NOT BUILT, so this stays OPEN.
>
> **What shipped** (`873575ecf`, council `b3676918` APPROVED at round 2 — 13 reviews, 10 approve, 3
> advisory, none high; registration `e03fbde6d` + LANDMINES): `platform/livespec` holds each guarded
> live object's declaration in a file that is ALLOWED to change, and **4** guards now assert against
> it instead of against a frozen migration. Both `t.Skipf` silent-greens are gone. Register **SQLC-002**.
>
> **Proven, not asserted.** The tripwire was written BEFORE the conversions and fired on **exactly the
> four files** they cover, then passed once they were converted — so it has been observed red and
> green on the same day. Mutation battery **6 of 6** behaved, each run singly with a revert between.
>
> **⚠ WHY THIS IS STILL OPEN.** Phase 1 closes the half where a guard asserted something that *could
> not fail* (three asserted the text of a file the checksum rule forbids editing). It does **NOT**
> close the live-drift half: nothing compares a declaration to the LIVE object. A migration editing a
> guarded object now leaves livespec stale **with no tell at all** — the guards compare Go to livespec
> and the migration changed a third thing neither reads. That is why the LANDMINES entry shipped with
> phase 1 rather than with phase 2.
>
> **Phase 2** (a read-only `--live-declaration-drift` mode on `cmd/config-key-audit` + a daily check
> image on the `shared-output-fields-check` pattern + `RELEASE_IMAGES`) is a separate council round and
> is NOT built. The `trigger_bindings` declaration is inert until it lands, and is COUNTED
> (`DeferredDeclarations`) so it cannot read as guarded.
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

## Relations

`bugs_closed/317` (where this was found; its fix is sound and re-verified live 2026-08-22) ·
`bugs_open/136` (the neighbouring live-config-vs-code class, and its reusable auditor) ·
`bugs_open/341`, `bugs_open/355` (active lanes owning objects cited here as evidence — do not touch) ·
`RFC_006` (the owner ruling that a pre-commit hook cannot gate live config; `SingleOwner` + daily
CronJob) · `RFC_022` / WFA-013 (`audit-optional-key-budget.sh`, the worked auditor+cron+parity-test
shape) · LANDMINES `Dedup index ↔ Go list lockstep` and the `doc_subjects subject_type` entry (the
same "two hand-maintained copies of one vocabulary" family).
