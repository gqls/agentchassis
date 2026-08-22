# NOTES — live-object declaration drift

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-22 — session start: the lane exists because bug 317 did NOT

Asked to fix `bugs_open/317`. **It is not there.** Closed and moved to `bugs_closed/`
on 2026-08-21 (`c12133e2e`), one day before this session.

I did not take the close banner's word for it. Three legs, re-checked today:

1. **Live column** — the artefact, not the file:
   ```sql
   SELECT name, enabled, (regexp_match(pre_query, 'NOT IN \(([^)]*)\)'))[1]
   FROM scheduled_tasks WHERE name='claimed-item-timeout';
   ```
   `enabled=t`, exclusion list = **14 item types, `dark_section_audit` present**. [MEASURED 2026-08-22]

2. **The union of both gate rosters is exactly those 14.** I nearly filed a false finding here —
   see the misstep below.

3. **The guard is green at committed HEAD.** `git archive HEAD` into a clean tree (not the
   shared dirty one), `go test ./platform/orchestration/actions/
   -run TestClaimTimeoutExclusionCoversBothCompletionGates -count=1` → **ok, exit 0**, at
   `45b728b01`. [MEASURED 2026-08-22]

317's fix holds. The lane below is a DIFFERENT defect, found while verifying it.

### MISSTEP 1 — I counted 12 where the live list said 14, and briefly believed the guard was broken

`grep -rn "RegisterVerifier(" .../discovery_checks/` returned **11** verifier types.
`noChangeGates` holds **1** (`dark_section_audit`). 11 + 1 = 12, against a live list of 14.
The two unaccounted types were `hardcoded_section_colors` and `needs_brand_head_assets`, and
for a few minutes I had a reverse-direction guard failure that the test manifestly was not
reporting.

**The error:** I grepped one spelling of the registration call. There are **two** —
`RegisterVerifier(` and `RegisterVerifierWithPolicy(`, the second added when RFC_017 made
verifier error-handling fail-closed with an opt-in field. Both funnel into the same registry
(`verifiers.go:143-149`), and `RegisteredVerifierItemTypes()` returns both. 13 verifiers + 1
gate-1b entry = **14**, exactly the live list.

**The cheap check that would have caught it:** grep the REGISTRY, not the call site —
`RegisteredVerifierItemTypes()` is the function the guard itself uses, and reading its
implementation names every writer of the map in one step. A grep for one call spelling
measures my guess about the API, not the API. Logged to `WRONG_CALLS.md`.

### The real finding: a guard that reads history and calls it a specification

While chasing the above I read the guard, and the thing it reads is not the live object.

`claim_timeout_exclusion_lockstep_test.go:48` globs
`*_claimed_item_timeout_generic_evidence.sql` and parses `item_type NOT IN (...)` out of the
newest match. That match is `482`, and in `482` the clause exists **only in a comment**
(line 27) — deliberately, and the file says so in its own header: the migration applies a
`replace()` of a tail fragment, so the applied SQL never contains the full clause, and the
comment is nominated as the declaration for the test to parse.

Meanwhile the live object has since been edited again by
`524_claimed_item_timeout_honours_the_cooldown.sql` — **a filename the glob does not match.**

So: the guard reads a comment in a historical file; the sweep reads
`scheduled_tasks.pre_query`; nothing ties the two.

### Confirmed absences (the disconfirming checks, both came back negative)

- `schema_migrations` records a **checksum of the FILE** (`run-migrations.sh:265,453`), never
  of the live object. It answers "was this file applied", never "does the live object still
  match what we believe".
- **Nothing in the repo reads `pg_get_functiondef`, `pg_trigger` or
  `information_schema.triggers`** — `grep -rn` over `*.go|*.py|*.sh` excluding
  `sql_for_agents`: **zero hits**. There is no live-object drift check of any kind.

Both of these could have come out the other way, and would have shrunk or killed the finding.

### Census of the class [MEASURED 2026-08-22]

`scheduled_tasks`: **97 rows, 47 enabled, 35 with a `pre_query`, 24 enabled with one.**
Of the 35, **8 enabled rows embed a literal quoted vocabulary** in their predicate
(`pre_query ~* 'IN \s*\(\s*'''`): `claimed-item-timeout`, `database-cleanup`,
`detected-item-promoter`, `held-pair-canary-escalation`, the three
`site-discovery-rotation-*`, and `site-render-audit-rotation`.

Go guards that read a repo migration file as a specification — **9 test files**, of which
**7 assert a property of a LIVE DB-resident object**:

| # | guard | live object |
|---|---|---|
| 1 | `TestClaimTimeoutExclusionCoversBothCompletionGates` | `scheduled_tasks.pre_query` |
| 2 | `TestGoAndSQLAgreeOnTheCooldownBoundary` (506) | dispatch read SQL |
| 3 | `TestDivergenceVocabularyMatchesMigration344` | trigger `trg_site_component_archive` |
| 4 | `page_component_divergence_test.go:219` (357) | trigger pair `trg_page_component_artefact_archive_upd/_del` |
| 5 | `TestValidDocSubjectTypes_LockstepWithMigrationCheck` | CHECK constraint `doc_plans_subject_type_check` |
| 6 | `TestRenderSlotNameConfigKeyMatchesTheSeededWorkflow` | `agent_definitions` workflow (reads a SEED) |
| 7 | `TestMigration302CarriesTheCanonicalPredicateVerbatim` | dispatch predicate SQL |

`write_experience_pattern_test.go` and `contact_info_no_fabrication_test.go` also read the
migrations dir but may be asserting repo-corpus conventions, where the repo genuinely IS the
subject. **[UNCLASSIFIED]** — pending the plan's judgement, not asserted either way here.

### A weaker claim I am NOT making

I checked whether `524` introduced a cooldown-boundary drift against the `506` guard. It does
**not**: `524` *stamps* `retry_after` on the reset arm, it does not *read* it with a `<`/`<=`
boundary. Recording this because the analogy was tempting and wrong, and 016b's own history
says the untreated sibling is where the next incident comes from — but only when the mechanism
actually transfers. It did not here.

### Prior art checked before filing

- `bugs_open/136` (live config says `*_domain` where code reads `*_pipeline`) — **neighbour, not
  duplicate.** Its surface is `agent_definitions` step-config KEYS that nothing reads; mine is a
  live SQL/DDL object whose only guard reads an untied repo copy. 136 brings reusable machinery:
  `cmd/config-key-audit`, `scripts/audit-config-keys.sh`.
- `needs_diagnosis` queue: **0 rows** at `awaiting_diagnosis` before I filed. No duplicate.
- `bugs_open/341` is the only open bug on the `claimed-item-timeout` sweep and is **OWNED and
  ACTIVE** (`scripts/who-owns.py 341` → `bugfix_307_terminal_write_contract`, 17 commits/14d).
  Not touched.
- LANDMINES: the `doc_subjects subject_type` entry (line ~657) independently documents instance
  5 and its two-enforcement-point shape. Relevant, and its "grep the VALUE you are adding, not
  the table you are changing" check is the same family.

### Diagnosis loop

Filed per the OWNER RULING of 2026-07-31 (a `bugs_open/` file asserting a cross-cutting or
structural root cause is not filed until it has been through the loop).
`CORRELATION_ID=c236fbb4-ca7f-4540-8170-8b806f40fc54`,
**`RUN_CORRELATION_ID=c8ec6478-5a54-4a16-aaf1-1e3373684ba0`** ← the key the artifacts carry.

---

## 2026-08-22 (later) — three live objects checked against their files. All three AGREE. The class is LATENT.

I went looking for damage and did not find it. Recording that plainly, because the finding below
is stronger *without* a damage claim and would be worth less if I stretched for one.

| live object | checked how | result |
|---|---|---|
| `scheduled_tasks.claimed-item-timeout.pre_query` | `regexp_match` on the live column | 14 types = union of both rosters ✅ agrees |
| function `site_component_history_archive` (mig 344) | `pg_get_functiondef` | all three verdicts + the md5 clause present ✅ agrees |
| function `page_component_artefact_archive` (mig 357) | `pg_get_functiondef`, all four of the guard's needles | 1 occurrence each, all four ✅ agrees |

[MEASURED 2026-08-22] Each of these could have come out otherwise — that is the point of running
them. None did.

### But the mechanism sharpened, and the sharper form does not need damage

Reading `TestDivergenceVocabularyMatchesMigration357`
(`page_component_divergence_test.go:217-240`) makes the real defect obvious:

```go
src, err := os.ReadFile(".../357_page_component_artefact_archive.sql")
if err != nil { t.Skipf(...) }                       // ← unreadable ⇒ PASSES
for _, needle := range []string{ ... } {
    if !strings.Contains(s, needle) {
        t.Errorf("migration 357 no longer contains %q — the DB-side verdict has drifted")
    }
}
```

**Migration 357 is an applied, frozen file. It CANNOT stop containing those needles** — editing an
applied migration is forbidden by this estate's own rule, precisely because `schema_migrations`
holds its checksum. So the assertion "migration 357 no longer contains X" describes an event that
policy has already made impossible.

The guard is therefore **not capable of failing in the direction it exists to detect.** It is not a
weak guard against live drift; it is a guard against a different thing entirely (someone editing a
frozen file), wearing the error message of the first. That is the same shape as the memory index's
*a quiet-test passes when the RULE is gone* and *a mutation that PASSES may have hit a guard in
series* — a check whose green is uninformative.

Two aggravations in the same function:
- `t.Skipf` on an unreadable file. A rename makes the guard pass silently. Instance 2
  (`TestGoAndSQLAgreeOnTheCooldownBoundary`, migration 506) has the identical construct.
- The Go half of the same test **is** real — it reads `page_component_divergence.go` and asserts the
  live source compares `rendered_html_digest <> md5(pc.rendered_html)`. So one test contains a
  genuine assertion and a vacuous one, and the genuine half makes the vacuous half look load-bearing.

### The live trigger SET has already outgrown the file the guard reads

`pg_trigger` on `page_components` returns **three** triggers on `page_component_artefact_archive`:

```
trg_page_component_artefact_archive_upd   ← declared in 357
trg_page_component_artefact_archive_del   ← declared in 357
trg_page_component_content_archive_upd    ← NOT in 357; added by 552
```

`552_content_only_update_archives_too.sql` added the third. The guard reads 357 and the Go comment
at `page_component_divergence.go:9` describes a *pair*. So the description the guard checks is
**already incomplete** with respect to the live object — the function body still matches, but the
set of firing conditions does not.

⚠ **552 belongs to an ACTIVE lane** — it is `bugs_open/355`'s build (RFC_042 option (c), ruled
2026-08-22). **Not touched, and must not be.** It is cited here only as evidence that live objects
accumulate past the file a guard reads, which is the whole claim. Its author did nothing wrong.

### What this does to the fix

It rules out the tempting cheap fix. "Point the guard at the newest migration instead of a pinned
one" (what instance 1's glob already tries) does **not** work: it still reads history, the glob
must guess a filename convention, and 552's name would not match a `*_page_component_artefact_archive.sql`
glob any more than 524's matched the claim-timeout one. The declaration has to stop being a
migration file.

---

## 2026-08-22 — instance 5 checked, and it is the best guard in the class. That matters for the fix.

`TestValidDocSubjectTypes_LockstepWithMigrationCheck` (`doc_subjects_common_test.go:141`) does
**not** pin a filename. Its helper `newestConstraintValues` (`:73-120`) walks the whole
`sql_for_agents` directory, parses the numeric prefix, and keeps the **newest** migration that
recreates the named CHECK constraint — then asserts Go's `validDocSubjectTypes` equals the
**union** of the newest `doc_plans_subject_type_check` and `doc_notes_subject_type_check` values.

Live check [MEASURED 2026-08-22, `pg_get_constraintdef`]:

| | values | n |
|---|---|---|
| live `doc_plans_subject_type_check` | tool, pipeline, experience, action, experience-pattern, component | 6 |
| live `doc_notes_subject_type_check` | …+ landmine, decision | 8 |
| union | 8 | |
| Go `validDocSubjectTypes` | tool, pipeline, experience, action, experience-pattern, component, landmine, decision | 8 |

**Agrees.** The 6-vs-8 asymmetry between the two tables is deliberate and documented (a landmine
has no shared-contract shape for `doc_plans`), and the test models it correctly as a union.

### Why this changes the recommendation

This is the shape the rest of the class should already have had, and it is **already in the repo**.
It removes the two weakest links the other guards have — no pinned filename, no glob that a later
migration can slip past by being named differently (`524`, `552` both did exactly that). It derives
a *current declaration* from append-only history by construction, which is the thing I was about to
propose inventing a `declarations/` directory for.

So the honest framing of the fix is **not** "build a new declaration mechanism". It is:

1. **Generalise `newestConstraintValues`** from "newest migration declaring this CHECK constraint"
   to "newest migration declaring this *named live object*" — the same scan, parameterised by what
   is being declared. One helper, shared, instead of five bespoke readers.
2. **Add the half nobody has: compare that derived declaration to the LIVE object.** This cannot
   live in `go test` (no DB), which is why it has to be a runtime auditor — and why a pre-commit
   hook cannot do it either (at commit time the migration is unapplied; RFC_006 ruled on exactly
   this and the answer was a daily CronJob).

### Score so far: 4 live objects checked, 4 agree

`claimed-item-timeout.pre_query` · `site_component_history_archive` · `page_component_artefact_archive`
· both `subject_type` CHECKs. **The class is LATENT — filed for the door, not the damage**, the same
basis 317 itself was filed on. What is NOT latent is that three of the guards cannot fail in the
direction they exist to detect, and the live trigger set has already outgrown migration 357.

---

## 2026-08-22 — census COMPLETE: 6 of 6 live objects checked against their declaring file. All 6 AGREE.

| # | guard | live object | how checked | result |
|---|---|---|---|---|
| 1 | `TestClaimTimeoutExclusionCoversBothCompletionGates` | `scheduled_tasks.claimed-item-timeout.pre_query` | `regexp_match` on live column | ✅ 14 = union of both rosters |
| 2 | `TestGoAndSQLAgreeOnTheCooldownBoundary` (506) | dispatch read predicate | — **not separately measured**; `524` stamps rather than reads, so no boundary to compare | ⚠ [UNMEASURED] |
| 3 | `TestDivergenceVocabularyMatchesMigration344` | fn `site_component_history_archive` | `pg_get_functiondef` | ✅ 3 verdicts + md5 clause |
| 4 | `TestDivergenceVocabularyMatchesMigration357` | fn `page_component_artefact_archive` | `pg_get_functiondef`, all 4 needles | ✅ 1 occurrence each |
| 5 | `TestValidDocSubjectTypes_LockstepWithMigrationCheck` | 2 CHECK constraints | `pg_get_constraintdef` | ✅ Go 8 = union(6, 8) |
| 6 | `TestRenderSlotNameConfigKeyMatchesTheSeededWorkflow` | `agent_definitions` `page-content-writer` | count of the setting in live `default_config` | ✅ 2 occurrences, seed demands ≥2 |
| 7 | `TestMigration302CarriesTheCanonicalPredicateVerbatim` | `agent_definitions` `build-site-planner` | grep live config for the built predicate | ✅ exactly 1, verbatim |

Instance 2 is honestly marked UNMEASURED rather than counted as agreeing — I have no live artefact
for it that I checked, and padding the score would be the exact dishonesty this file exists to
prevent.

### Instance 7's own comment is the best statement of the bug in the repo

`links_shipped_predicate_test.go:100-104`:

> *"If migration 302's file is ever archived/moved, move this assertion onto its successor rather
> than deleting it — **the LIVE row was written from that text**, and this test is the only thing
> tying the row's predicate to the builder."*

That is the inference the whole class rests on: *live == file, because the live object was once
written from the file.* True at the instant of application; it decays silently from then on. The
author saw the tie and named it, and the tie is still not asserted anywhere.

### What the census does and does not license

**Does:** the defect is real, structural, and affects seven guards across four kinds of live object
(`scheduled_tasks.pre_query`, trigger functions, CHECK constraints, `agent_definitions` workflows).
Three of the seven cannot fail in the direction they exist to detect, because they assert the text
of a file that policy forbids editing. Two more (`506`, `357`) `t.Skipf` or pass when the file is
unreadable, so a rename is a silent green.

**Does not:** there is no damage to point at. No live object has drifted. Anyone reading this should
file/fix it on the "close the door" argument — the same basis `bugs_closed/317` itself used — and
**not** claim a live fault. If a future session finds drift, that is a new and much louder finding.

---

## 2026-08-22 — CORRECTION to the census table above: instance 2 is now MEASURED, and it agrees

> **CORRECTED 2026-08-22:** the row for instance 2 above says `⚠ [UNMEASURED]`. That was true when
> written and is now superseded — I found the live artefact and checked it. **Leaving the original
> row in place** rather than editing it, per the working-docs rule; this block is the correction.

`506_dispatch_reads_honour_retry_after.sql` writes **two** live objects (its own header, line 12):
`scheduled_tasks.build-pipeline-trigger.pre_query` (line 37) and an `agent_definitions.default_config`
via `jsonb_set` (line 57).

The Go canonical renderer is `workItemRetryNotPendingSQL` (`work_items_common.go:355`), which for
alias `wi` produces `(wi.retry_after IS NULL OR wi.retry_after <= NOW())`.

Live [MEASURED 2026-08-22]:

```
$ ... SELECT pre_query FROM scheduled_tasks WHERE name='build-pipeline-trigger';
      1 (wi.retry_after IS NULL OR wi.retry_after <= NOW())
```

**Exactly one occurrence, verbatim, non-strict `<=`** — matching Go, and specifically NOT the strict
`<` the guard warns about. ✅ agrees.

### Final score: 7 of 7 live objects agree. Nothing has drifted anywhere in the class.

That is the whole census, with no UNMEASURED rows left and no instance excluded for being awkward.
The bug is filed on the door, not on damage — and a future session finding drift should treat that
as a NEW and much louder finding, not as confirmation of this one.

### One more design input for the fix

`506` is the case that shows why the auditor cannot be keyed on "one migration → one object". A
single migration writes a `scheduled_tasks` row **and** an `agent_definitions` row, and the guard
that reads it checks only the first. Whatever registry the fix introduces has to be keyed on the
**live object**, with the migration as an attribute of it — not the other way round. Keying on the
file reproduces the defect in the fix.

### Council scope, checked before designing (`scripts/council-scope.sh:57-61`)

In scope: `^(platform|internal|pkg)/` and `^docs/agent_docs/sql_for_agents/[0-9]{3}_[A-Za-z0-9_]+\.sql$`.
**`cmd/` and `scripts/` are OUT of scope.** The existing precedent auditor lives in
`cmd/config-key-audit`, i.e. outside review. So the new logic should sit in `platform/` (reviewable,
importable by both a test and a runtime auditor) with only a thin `main` in `cmd/` — that is a
design consequence of the gate's scope, not a style preference.

---

## 2026-08-22 — number collision: I wrote the bug as 362 and another lane had already taken it

Checked the highest bug number (`361`), wrote the file as `362`, and by the time I saved it
`bugs_open/362_..._two_tool_writers_persist_rendered_html_without_link_repair...md` existed and was
**committed and marked OWNED** by the `webdesign_tool_rebuilds` lane (`30c03b6f0`). Renumbered to
**363** before committing; mine was still untracked so it cost a `mv` and one `sed`.

**The check:** a bug number is not reserved by checking it — only by committing the file. On this
tree the gap between "I picked a number" and "I saved the file" is long enough to lose the race, and
CLAUDE.md already records that several numbers name two unrelated cases for exactly this reason.
**Re-check the max immediately before `git add`, and commit the file early to claim it** — the
evidence sections are worth committing before the fix candidates are written anyway.

Not logged to WRONG_CALLS: nothing false was asserted and nothing was published wrong. It is a
coordination hazard, which is this file's job, not that one's.

---

## 2026-08-22 — the design pass came back; grounding it caught two false migration numbers

The plan is good and I am taking its recommendation. But a subagent's report is another doc — no
seam shows where its measuring stopped — so I re-ran its load-bearing new claims before letting any
of them into `bugs_open/363`. Two of them were wrong.

### ✅ CONFIRMED, and it upgrades the bug from "could drift" to "the channel is proven"

**`220_claimed_item_timeout_generic_evidence.sql` was edited NINE times after it first applied.**
`git log --follow` on the file: created by `d61b3ace1` ("generic completion evidence for the
claim-timeout sweep"), then `ac9f75a0c`, `ec8ad7959`, `96dd3015c`, `af2667453`, `dc4f4e6b2`,
`a60a13cbb`, `ad51ca863`, `d644723b8`, `c121d5a73`. So for most of its life this file **was** the
mutable declaration and was maintained as one — which is exactly why the guard was written to read
it, and exactly what the checksum-freeze convention later took away without anything replacing it.

**Three migrations edited the live exclusion clause under filenames the guard's glob cannot see.**
[MEASURED 2026-08-22] Each does a `SET pre_query = replace(...)` naming the previous clause verbatim:

| migration | filename | matches glob `*_claimed_item_timeout_generic_evidence.sql`? |
|---|---|---|
| `322_dead_fragment_link_claim_timeout_exclusion.sql` | `…_claim_timeout_exclusion.sql` | **NO** |
| `331_literal_markdown_claim_timeout_exclusion.sql` | `…_claim_timeout_exclusion.sql` | **NO** |
| `374_decision_regression_claim_timeout_exclusion.sql` | `…_claim_timeout_exclusion.sql` | **NO** |
| `524_claimed_item_timeout_honours_the_cooldown.sql` | (same column, cooldown stamp) | **NO** |
| `482_claimed_item_timeout_generic_evidence.sql` | — | yes (this is all the guard ever reads) |

The guard is right today only because `482`'s author hand-reconciled every earlier edit into a
comment. **That is a hand-maintained copy of a live vocabulary — the exact class this estate keeps
filing bugs about.**

### ❌ TWO CLAIMS REFUTED — `269` and `341` do not touch this at all

The report listed six "live edit vehicles": `269, 322, 331, 341, 374, 524`. Checked each:

- `269` is `269_deduplicate_sections_handler.sql` — **no mention of `claimed-item-timeout`.**
- `341` is `341_domain_strategist_refresh_safe_and_premise_fields.sql` — **no mention.** (The number
  is also confusable with `bugs_open/341`, which *is* about this sweep. Plausibly that collision is
  where it came from.)

So the true count is **four**, not six. The argument is unharmed — three glob-invisible edits to the
clause is already decisive — but the figure that goes in the bug file is four, and I have written
four. **Had I pasted the report through, `bugs_open/363` would have carried two fabricated
citations**, in a file whose entire purpose is to complain about unverified declarations.

### ✅ CONFIRMED — every reuse target the plan leans on exists as described

- `deployments/kustomize/services/shared-output-fields-check/base/cronjob.yaml` exists and its header
  argues the Go-image-over-Python-mirror case verbatim, including that an image "dissolves … both of
  RFC_006's named drift risks (a `DECLARED_*` literal kept in step by hand, and a parity test to
  notice when it is not)" and that these checks **connect to Postgres directly, never `kubectl exec`**
  (no pods/exec RBAC — a kubectl-only tool fails in a way that looks CLEAN).
- `writeDocNote` exists at `cmd/config-key-audit/fleetdb.go:110`.
- The makefile warning is real and emphatic (`makefile:78-92`): two check services were *already*
  born outside `RELEASE_IMAGES`, the coverage gate structurally cannot see such an omission, and
  **"A NEW CHECK SERVICE MUST BE ADDED HERE IN THE COMMIT THAT CREATES IT."**

### The one design point I am keeping from my own census over the report's framing

The report treats instance 5 as "(iii) judgement". I agree, but the reason matters for the edit
list: `newestConstraintValues` is **the shape to generalise**, not merely a case to patch. Its
corpus scan is what makes 322/331/374 visible where a glob is not.

---

## 2026-08-22 — the diagnosis loop returned UNVERIFIABLE, and that is NOT a confirmation

Run `c8ec6478-5a54-4a16-aaf1-1e3373684ba0`, item reached `complete`. **But `complete` is not proof
the work happened** — and here it wasn't. Reading the diagnoser's own response:

```
status      : UNVERIFIABLE
conclusion  : NOT CONFIRMED (stopped: iteration-cap)
stopped_by  : iteration-cap
summary     : Diagnosis NOT confirmed (stopped: iteration-cap). Best-effort trail
              attached for a human; no fix proposed.
```

**So `bugs_open/363` is NOT carried by the loop.** It is carried by my own first-hand verification —
7 live objects measured, the git history of 220, the three glob-invisible migrations, and both
disconfirming checks. The 2026-07-31 owner ruling allows exactly that (run the loop, or state plainly
why first-hand verification substituted); what it does not allow is writing up an UNVERIFIABLE run as
a verdict. It is not one, and this file says so.

**Why it failed, in its own words** (iteration 1 citation): the run got **no seed scope**, fell back
to symbol search on the symptom text, and drew `MarkDecommissioned`, `idleMonitor`,
`ClaimEmailAttempt`, `notifyTopicsReady` — none of my seven guards. It then re-proposed almost the
same scope for five iterations and hit the cap. My fault: `090` takes `SEED_SCOPE` (`:117`) and warns
`nothing to key coverage on … dispatching blind` (`:308-309`). Logged to `WRONG_CALLS.md`.

**The general lesson, worth more than this run:** a symptom whose evidence lives in **test files and
migration files** has no runtime symbol named after it — the defect *is* that a guard reads the wrong
thing. So the fallback scope structurally cannot find this class, which is the class most in need of
an independent read.

### But its Tier-1 citations corroborated three facts, and handed me the best finding in the file

Independently cited by the loop, from the live system:

1. `schema_migrations` columns: `filename, applied_at, checksum, applied_by, notes` — confirms the
   ledger keys on the FILE and its checksum, never on the live object.
2. `524_claimed_item_timeout_honours_the_cooldown.sql | 2026-08-21 18:44:22 | 30af2a80…` and
   `506_dispatch_reads_honour_retry_after.sql | 2026-08-20 14:14:47 | 3501f696…` — both applied.
3. It quoted the **live `scheduled_tasks.pre_query`**, and that quote is the finding I had missed.

### The finding I had missed, and the loop surfaced: the LIVE object's own contract statement is stale

The live column carries, immediately above the exclusion clause:

> *"The item_type exclusion is the LOCKSTEP TWIN of the `RegisterVerifier()` calls … 
> `TestRegisteredVerifiersMatchClaimTimeoutExclusion` pins the two together."*

**Both halves are false as of 2026-08-19.** The contract is the union of both gate rosters (mig 482),
and `grep -rn "func TestRegisteredVerifiersMatchClaimTimeoutExclusion"` returns **nothing** — the test
was replaced by `TestClaimTimeoutExclusionCoversBothCompletionGates`.

I had been reading the live column for its *data* (the 14-type list, which is correct) and never read
the *prose* around it. **That sentence is the cause of bug 317** — its author read it, believed it and
built a gate-2-only lockstep. `482` corrected the data and left the sentence.

This is the one thing that defeats the obvious framing of my own bug. It is not "the repo file is
stale, the live object is true": **the live object carries its own stale declaration too**, and a
drift auditor comparing live text to a declared clause would pass it, because the clause matches and
it is the prose that lies. Recorded as a constraint on the fix, not a reason to abandon it.

Added to `bugs_open/363` as its own section.

---

## 2026-08-22 — council round 1: REVISE, gated by editquality, and the gating objection was RIGHT

Corr `b3676918-9eee-4b9f-85f3-749e16b3d033`. **14 reviews, 3 abstained, not truncation-gated.**
10 approve / 4 object. `decided_by: gating objection from editquality`.

### The gating defect, which I had not seen and should have

> **editquality [HIGH], edit 2:** *"The tripwire's allow-list … omits the two guards the diagnosis
> names explicitly as broken — `complete_work_item_retry_guard_test.go:231` and
> `page_component_divergence_test.go:221`. If those files still read `sql_for_agents` paths, the new
> `TestNoNewMigrationFileReadersOutsideTheAllowList` fails on introduction; if it doesn't catch them,
> it gives false confidence exactly where the diagnosis says confidence is unwarranted."*

**Correct, and it is my own cited landmine turned back on me.** My round-1 plan put the tripwire in
commit A and deferred converting the two silently-skipping guards to a commit B that existed only in
prose. So at the moment the tripwire lands, those two files still read the corpus: the test either
goes red on introduction, or I add them to the allow-list — allow-listing precisely the two guards
the bug calls broken. My own plan quoted the rule against that ("an allow-list without reasons
converts a live debt into a false all-clear") and then walked into it, because the A/B split hid the
overlap.

**The fix is structural, not cosmetic:** round 2 converts all four defective guards in the SAME plan
as the tripwire, so every remaining allow-list entry is a guard that legitimately keeps a repo-side
check. And phase 2 (the auditor, image, CronJob, makefile) is split into its own round — which also
answers editquality's medium, that the deployment files were hidden inside a dockerfile comment so
"the image builds but never deploys".

### The other objection that found something real

> **debug_historian [MEDIUM], edit 5:** the plan's own evidence found THREE triggers bound to
> `page_component_artefact_archive` where 357 declares two — *"the sketch's Declaration … only checks
> the function BODY text, not the trigger-binding count, so this specific drift-in-progress (which the
> plan's own evidence surfaced) is not itself covered by the new mechanism."*

**Right, and sharp.** I found that drift, put it in the bug as evidence, and then designed a fix that
would not have detected it. livespec now carries a binding-COUNT declaration.

### Objections answered by going and measuring, having skipped the search first time

- **reuse_agent [MEDIUM ×2]:** no reuse search shown. Fair — I asserted reuse of the binary without
  checking its comparison engine. Ran it: `grep 'scheduled_tasks' cmd/config-key-audit/*.go` → **ZERO
  hits**; `UnknownConfigKeys`/`ListDeclaredConfigKeys`/`ListRemovedConfigKeys` are all in
  `datahelpers/action_inputs.go:247/435/335` and do key-**set** membership over `agent_definitions`
  step config, not text matching over live object definitions. Of **17** existing `*-check` services,
  none probes a trigger, CHECK constraint or `pre_query`. Different comparison, no engine to extend.
- **prior_art_librarian [MEDIUM]:** verify `dbConn`/`writeDocNote` exist rather than asserting.
  They do — `cmd/config-key-audit/fleetdb.go:52` and `:110`.
- **architecture [LOW]:** livespec has no growth boundary unlike RFC_022's budget. Added
  `MaxDeclarations` with a test.
- **debug_historian [LOW]:** exact-string equality on rendered SQL is a false-alarm generator.
  Changed to fragment containment.
- **guardian [LOW]:** confirm the deleted glob/regex has no other consumer — grep, not assumption.

### The lesson worth keeping

**A phased plan can create a defect that neither phase has.** Both of my commits were individually
sound; the gap was the interval between them, and it existed only because I described phase B in
prose instead of listing it as edits. The council could see the plan I *wrote*, not the plan I
*meant* — which is the correct thing for it to review, and the reason the objection landed.

Round 2 resubmitted on the same trail (`RESUBMIT_CORR`), run correlation
`98e657be-e266-44bb-a548-e6fcb968071d`.
