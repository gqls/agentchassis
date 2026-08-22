# 363 — seven Go guards assert the TEXT of an append-only migration file, so they cannot fail in the direction they exist to detect

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

## The census — 7 guards, 4 kinds of live object, all measured

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

- `scheduled_tasks`: **97 rows, 47 enabled, 35 carry a `pre_query`, 24 enabled ones do.** Of the 35,
  **8 enabled rows embed a literal quoted vocabulary** a Go list could drift from
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

## Fix candidates

*(Pending: the diagnosis loop's independent read — run correlation
`c8ec6478-5a54-4a16-aaf1-1e3373684ba0` — and the design pass. To be filled in by this lane before
council submission; do not treat this section's absence as "no candidates exist".)*

Two constraints any candidate must satisfy, both already established above:
- it cannot require a database at `go test` time, and it cannot be a pre-commit hook;
- it must key on the **live object**, with the migration as an attribute of it. Migration `506`
  writes **two** live objects (a `scheduled_tasks` row and an `agent_definitions` row) and its guard
  checks only the first — so a registry keyed on the file reproduces the defect inside the fix.

Design consequence of `scripts/council-scope.sh:57-61`: scope is `^(platform|internal|pkg)/` plus
appliable migrations; **`cmd/` and `scripts/` are out of review.** New logic belongs in `platform/`
(reviewable, and importable by both a unit test and a runtime auditor) with only a thin `main` in
`cmd/`.

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
