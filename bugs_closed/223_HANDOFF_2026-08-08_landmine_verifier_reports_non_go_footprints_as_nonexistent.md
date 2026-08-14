# 223 — the landmine-verifier reports every NON-Go footprint as non-existent, because the code index holds only Go — and 284 of 288 entries have one

> ## ⚠ STATUS FIRST, BECAUSE READERS TRUNCATE — updated 2026-08-14
>
> **CLOSED — moved to `bugs_closed/` 2026-08-14 at the owner's direction**, under his
> 2026-08-12 restoration of the fixed-AND-live bar (superseding the 2026-08-06
> keep-in-place direction the previous banner cited). Both phases live and proven since
> v1.0.1284/1286; its residual finding became `bugs_closed/254` (also closed and live).
>
> **BOTH PHASES ARE LIVE AND BEHAVIOURALLY PROVEN. Nothing technical remains owed.**
>
> **PHASE 2 LIVE on `v1.0.1284`, proven 2026-08-11 09:55Z** (`027bf28a0` + `c7c9dd87f`,
> council APPROVED all reviewers, `3af67677-601e-4181-ad09-17c7a789f995`). The index now
> holds `var` and `const`. The same landmine entry that on 08-08 drew *"`metaCommentaryPatterns`
> and `placeholderPatterns` **no longer resolve as standalone symbols (possibly inlined or
> renamed)**"* today returns **`STILL_VALID` — "confirmed present at expected line ranges"**,
> and the evidence line's missing-kind warning has **retired itself** (`kinds with NO rows:
> type` — `var`/`const` gone). Nothing was ever renamed. Verbatim before/after and the
> full five-criterion reconciliation: the **PHASE 2 ACCEPTANCE** section at the foot.
>
> **PHASE 1 LIVE AND PROVEN on `v1.0.1279`, 2026-08-10 14:41Z** (`1058b5366` + `362c7c091`,
> council APPROVED `495df717-4010-491f-aec0-92c13aaf3809`; seed 365 applied and recorded).
> The entry that drew the flat false `STALE` on 08-08 was re-fired and returned
> `NEEDS_HUMAN_REVIEW` with three checks rendered `NOT ANSWERABLE BY THIS INDEX`.
>
> **⚠ TWO CORRECTIONS TO THIS FILE'S OWN §4a, both of which mislead in the direction of a
> false alarm:** (1) *"var+const should appear near **1,371**"* — **WRONG, the figure is
> 1,204.** 1,371 was measured without the `exclude_patterns: ["docs/"]` that
> `analyse_repo_local` actually passes; comparing against it reads a healthy index as 12%
> of rows silently dropped, which is the exact symptom of the identity collision the census
> exists to detect (`WRONG_CALLS.md`, 2026-08-11). (2) *"phase 2 adds NO NEW STRING LITERAL
> so it cannot be pod-grepped"* — true, but **no longer the constraint it was**:
> `bugs_open/153`'s build stamp landed in the same window, so
> `kubectl logs <pod> | grep 'build provenance'` + `git merge-base --is-ancestor` dates any
> commit exactly, greppable spelling or not.
>
> **What remains is NOT this lane's and NOT technical:** `RFC_022` awaits an owner ruling,
> and the `derive_checks` follow-up (§5.1) is RFC_005's mechanism to route, not to patch
> here. Full account: **the PHASE 1 and PHASE 2 sections at the foot of this file**, and
> `docs/agent_docs/docs024_key_docs_latest/bugfix_223_index_answerability/`.
>
> **If you are here to read a verdict rather than to fix this:** every landmine verdict
> stored up to 2026-08-09 must be read under this file's own caveat — for a non-Go
> footprint, `STILL_VALID` is **not** evidence for the entry any more than `STALE` was
> evidence against it. Only the verdict's prose reasoning carries signal.

**Filed 2026-08-08. OPEN, UNOWNED.** Found by the `provocation_pipeline` lane doing the
routine thing: appending two `LANDMINES.md` entries and letting RFC_005 §3.2's verifier
check them. Both verdicts were wrong in the same direction, and **one of them is
self-refuting** — it denies the existence of the very `doc_notes` category it was written
into, by the very script chain that wrote it.

**This is the residual class after `bugs_closed/163`, not a regression of it.** 163 fixed
path-bearing *symbol* queries for Go symbols (`parseSymbolQuery`/`symbolClauseFor`, live
v1.0.1245) and its fix is working. This is a different mechanism with the same surface:
the index it queries **contains no non-Go rows at all**, so there is no query form and no
fix to `symbolTokenClause` that could ever resolve a `.py`, `.sh`, `.sql`, table-name or
config-value footprint.

**Owning lane, for routing:** `architecture_review` (RFC_005 §3.2 is theirs — the
verifier agent, `landmines-verify-dispatch.sh`, `trigger-landmine-verifier.sh`), with
`bugfix_163_symbol_lookup` as the lane that already knows this code. I have not touched
their code; this file is the account, not a competing fix.

> ## CORRECTED 2026-08-08, within the hour, by the verifier itself. **The title overstates and the cause does not.**
>
> **What stands, and is now independently corroborated.** The index really is Go-only and
> the blind spot is real. Two further verdicts arrived after filing, and one of them
> *verified this bug's own landmine entry* and reached the same conclusion unprompted:
>
> > **STILL_VALID.** "The lookup results themselves demonstrate the entry's thesis: non-Go
> > footprints (scripts, doc_notes categories, bug references) returned **0 rows** while
> > Go-side components resolved abundantly, **confirming the index is Go-only and the
> > described blind spot persists.**"
>
> **What is WRONG: "reports EVERY non-Go footprint as non-existent."** It does not. Given
> the identical 0-row input, the conclusion the verifier draws **varies run to run** —
> three verdicts within one hour, all on non-Go footprints, all on v1.0.1267:
>
> | verdict | what it concluded from 0 rows |
> |---|---|
> | on the sync-ordering entry | **STALE** — "do not exist anywhere in the indexed codebase; the entire described workflow has no footprint" (flatly false) |
> | on the prose-columns entry, first run | hedged — "either lives outside indexed scope **or has been removed**" |
> | on the prose-columns entry, re-run | **correct abstention** — "cannot be mechanically verified" |
> | on this bug's own landmine entry | **correct, and reasoned about the blindness itself** |
>
> **So the defect is one layer up from where I filed it.** The blind spot is
> *deterministic*; the **inference drawn from it is stochastic**, ranging from a correct
> abstention to a flat assertion of non-existence, with nothing binding the conclusion to
> the blindness. That is the same shape as `163`'s invented staleness cause, and — noted
> because it is the second instance found on one day — the same shape as the provocation
> gate's safety boolean (`provocation_pipeline/HANDOFF_2026-08-08b` §4): **a model-authored
> conclusion sitting over a mechanical blind spot, with no structural guard between them.**
>
> **This makes fix candidate 1 stronger, not weaker.** "Ask the model to abstain when it
> cannot check" is already what it does *most* of the time — 3 of 4 verdicts here — and the
> failure mode is precisely that you cannot tell which run you got. An entry is degraded by
> the one flat STALE, not rescued by the three careful ones. **Only a structural bar on the
> inference removes it.**
>
> **And a consequence I had not seen, which matters for reading the corpus:** a
> `STILL_VALID` on a non-Go-footprint entry is **not evidence FOR that entry either** — the
> footprints were equally unchecked. My own entry passed because the model reasoned about
> its thesis, not because anything was verified. For 284 of 288 entries, *both* verdict
> directions are uninformative about the footprints, and only the prose reasoning carries
> signal.
>
> Original title kept rather than rewritten, so the overstatement stays visible.

**Why it matters more than a wrong verdict on one entry.** `LANDMINES.md` is the
system of record for traps that fire when you touch a thing (owner ruling D10), it is
synced into `doc_notes` for council seats and agents to read, and a `SessionStart` hook
puts matching entries in front of every new session. **A verdict of `STALE` /
"does not exist" is precisely the signal a future session would use to discount or delete
a valid entry** — so a false STALE does not merely fail to verify, it actively argues for
removing protection that is correct.

## The two verdicts, verbatim

Both from `doc_notes`, `categories ? 'landmine-verification'`, on chassis **v1.0.1267**:

**1. On my entry about the provocation pool's prose columns** (2026-08-08 17:44:02Z) —
`NEEDS_HUMAN_REVIEW`. It correctly confirmed the Go half (it read `loadGateCandidates` and
quoted the column precedence — **a genuinely useful catch, see §"credit where due"**), then:

> …but three footprint items (`calibration.vonc.com`,
> `319_provocation_gate_calibration_harness.sql`, `provocation-gate-calibration`) returned
> zero results in the symbol index, so the calibration fixture infrastructure **either
> lives outside indexed scope or has been removed.**

Those three are, respectively: a value in `provocations.domain`, a file under
`docs/agent_docs/sql_for_agents/`, and a value in `agent_definitions.type`. None is a Go
symbol; none could ever be in the index. The disjunction it offers is at least honest.

**2. On my entry about the sync/verify ordering trap** — flatly `STALE`:

> **last verified (landmine-verifier): STALE.** None of the three scripts
> (`landmines-sync.py`, `landmines-verify-dispatch.sh`, `trigger-landmine-verifier.sh`),
> the `--apply` flag, the `NEEDS_VERIFICATION:` output, or the `landmine-verification`
> category exist anywhere in the indexed codebase; **the entire described workflow has no
> footprint.**

## Every clause of that is false, and the verdict disproves itself

```bash
$ git ls-files scripts/landmines-sync.py scripts/landmines-verify-dispatch.sh scripts/trigger-landmine-verifier.sh
scripts/landmines-sync.py
scripts/landmines-verify-dispatch.sh
scripts/trigger-landmine-verifier.sh          # all three tracked at HEAD

$ grep -c '\-\-apply' scripts/landmines-sync.py                    # 6
$ grep -c 'NEEDS_VERIFICATION' scripts/landmines-verify-dispatch.sh # 2
```
```sql
SELECT count(*) FROM doc_notes WHERE categories ? 'landmine-verification';  -- 32
```

**The verdict asserting that `landmine-verification` does not exist is itself row 32 of
that category**, written by the script chain it says has no footprint. A claim cannot be
refuted more cheaply than by the artefact that carries it.

## CONFIRMED CAUSE — the index is Go-only, measured

```sql
SELECT CASE WHEN path LIKE '%.go' THEN '.go' WHEN path LIKE '%.py' THEN '.py'
            WHEN path LIKE '%.sh' THEN '.sh' WHEN path LIKE '%.sql' THEN '.sql'
            WHEN path LIKE '%.md' THEN '.md' ELSE 'other' END AS ext,
       count(*) AS symbols, count(DISTINCT path) AS files
  FROM code_symbols GROUP BY 1 ORDER BY 2 DESC;
```
```
 ext | symbols | files
-----+---------+-------
 .go |    5755 |   668
(1 row)          <-- ONE row. No .py, no .sh, no .sql, no .md, no 'other'.
```

`code_symbols` is **100% Go**. So for a non-Go footprint, both halves of 163's repaired
lookup return 0: the `path` column has no such path, and the `symbol` column has no such
identifier. 163's fix made a 0-row answer *honest* about the predicate it ran
(`-- searched: …`) and made a path-qualified miss re-run name-only — both good, and
neither can conjure a row class the index never ingested. **The verifier then narrates
that structural absence as evidence of non-existence**, which is the same
plausible-confident-wrong voice 163 documented, arriving through a different door.

## Measured scope — this is the common case, not the edge

```sql
SELECT CASE WHEN subject_key LIKE '%.go%' THEN 'names a .go path/symbol (verifiable)'
            ELSE 'NON-Go footprint (unverifiable by the index)' END AS kind,
       count(*) AS footprint_rows, count(DISTINCT source) AS entries
  FROM doc_notes WHERE subject_type='landmine' AND source LIKE 'LANDMINES.md%' GROUP BY 1;
```
```
 NON-Go footprint (unverifiable by the index)   | 1116 rows | 284 entries
 names a .go path/symbol (verifiable)           |  255 rows | 141 entries
```

**1,116 of 1,371 footprint rows (81%) can never be resolved, spanning 284 of 288
entries.** (Entries appear in both buckets when they carry a mixed footprint list, which
is why the entry counts sum past 288 — and a mixed entry is the dangerous one, because a
partly-confirmed verdict reads as diligent.) The corpus is overwhelmingly *about* tables,
commands, config values, migrations and scripts — that is what a landmine guards — so the
unverifiable class is not a tail, it is the corpus.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Teach the verifier what it cannot check, and forbid the inference.** Classify each
   footprint before looking it up; for a non-Go footprint, emit `NOT_CHECKABLE_BY_INDEX`
   and **structurally prevent** it contributing to `STALE`. Cheapest, and it is the only
   candidate that makes "absent from a Go index" unable to *mean* "does not exist". A
   verdict that can only say STALE about things it can actually see is the invariant worth
   buying.
2. **Give it a non-index check for the classes that have one.** A footprint naming a
   repo path is answerable by `git ls-files`; a table or column by
   `information_schema`; an `agent_definitions.type` by a row lookup. Real coverage
   rather than an honest abstention — but it is new capability per class, and each class
   needs its own care.
3. **Ingest non-Go files into `code_symbols`.** Widest reach, worst blast radius: the
   index feeds `diagnose_code_lookup` fleet-wide, so changing what it contains changes
   every diagnosis run's search space. **Architecture-scope by the 2026-07-29 ruling** —
   it changes what a shared mechanism *guarantees* to every consumer, not just this one.
   Do not let it ride in as a bug patch.
4. **Suppress the verdict wording only** (stop saying "does not exist", keep the rest).
   Rejected as a stand-alone: it treats a false conclusion as a phrasing problem, and
   `WRONG_CALLS.md` has a standing entry on exactly that move. Worth doing *with* 1.

## How to verify a fix

Re-run the verifier on **this file's own two motivating entries** — the provocation
prose-columns entry and the sync-ordering entry — and require: no `STALE` attributable to
a non-Go footprint, and the three genuinely-checkable Go facts still confirmed (the fix
must not buy abstention by checking less). Then re-run over an entry with a **mixed**
footprint list, which is the case a partial confirmation flatters.

**A verdict is not the artefact — read it out of `doc_notes`:**
```sql
SELECT created_at, left(body,300) FROM doc_notes
 WHERE categories ? 'landmine-verification' AND subject_key LIKE '%<slug>%'
 ORDER BY created_at DESC LIMIT 1;
```

## Why no `090` run (owner ruling 2026-07-31, stated substitute)

The ruling requires the diagnosis loop for a filed structural root cause, or a plain
statement of the equivalent first-hand verification substituted. Substituted deliberately
here, because the evidence is decisive and disconfirmable in one step: the failing
verdicts are **stored artefacts** on the current binary, not a reconstruction; the census
above would have shown a non-`.go` row had one existed; and the self-refutation (the
verdict living in the category it denies) needs no interpretation. The cause is also not
displaced from the symptom — it is the scope of the table the verifier reads. Filing 090
would re-derive a one-query census. **Anyone who disputes the cause should re-run the
`code_symbols` census, which is the claim's single point of failure.**

## Credit where due, because it changes how this should be read

The same verifier **caught a real error in one of these two entries within the hour** — my
prose-columns entry had the gate's column precedence backwards, and the verdict found it
by reading `loadGateCandidates` and quoting `provocation_gate_action.go:663`. My own md5
check had passed 9 of 9 and was structurally incapable of failing (`WRONG_CALLS.md`,
2026-08-08). **So this is not a case for switching the verifier off** — it is doing the
job on the half it can see, and doing it well. It is a case for stopping it from
converting its own blind spot into a verdict against the corpus.

## Related

- `bugs_closed/163` — the Go-symbol half, fixed and live v1.0.1245. Read it first: this
  file assumes its fix is working, and it is.
- `bugs_open/181` — `code_lookup` row caps are silent while a sibling in the same function
  reports. Same family: the lookup layer's absences are indistinguishable from findings.
- `LANDMINES.md` — the entry added alongside this file, so a session reading a `STALE`
  verdict tomorrow does not act on it.
- `architecture_review/RFC_005` §3.2 — the mechanism's own design doc.

---

# THIRD FAILURE MODE, 2026-08-08 late — a **Go** footprint fails too, when it names a package-level `var`

Contributed by the `bugfix_209_deploy_purpose_keyed_source` lane (I do not own this
bug — this is evidence into the shared account, not a competing fix). It matters
because this file currently says of `bugs_closed/163`: *"this file assumes its fix
is working, and it is."* **For `var` declarations it is not.**

## What happened

Re-fired the verifier for the 221 lane's entry (its handoff's outstanding loose
end, unblocked once fleet credit returned — last credit/quota error 20:13Z, run
dispatched 22:32Z). Verdict:

> **NEEDS_HUMAN_REVIEW.** Core footprint file and all five checker functions still
> exist at expected paths, but `metaCommentaryPatterns` and `placeholderPatterns`
> **no longer resolve as standalone symbols (possibly inlined or renamed)**…

**Both symbols exist at HEAD, unrenamed and not inlined** —
`validate_page_content.go:105` (`placeholderPatterns`) and `:1229`
(`metaCommentaryPatterns`), each `var X = []struct{…}`.

## The cause is the indexer's KIND COVERAGE, not staleness

Staleness was the obvious suspect and it is **not** the cause. The verifier read
commit `93c576963` (2026-08-07 09:31), ~38h behind HEAD — but both vars long
predate that commit, so a stale index would still have seen them.

One query settles it. `code_symbols` holds **no `var` kind at all**:

```sql
SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 2 DESC;
--  func 3592 | method 1114 | struct 973 | alias 40 | interface 36   (total 5755)
```

So a package-level `var` is **unrepresentable** in the index. Any landmine
footprint naming one resolves to nothing, 100% of the time, on a current index, in
a language the verifier is supposed to handle.

**Disconfirming control, in the same run:** the three `func` footprints on the
entry I filed the same evening (`ExtractActionInputs`, `findFieldRecursive`,
`findStorageURI`) all resolved, and that entry returned **STILL_VALID**. So this is
not "the verifier is broken" — it is the same shape this bug already describes:
a gap in the table it reads, converted into a verdict against the corpus.

## Why this one is worse than the non-Go case

The non-Go case at least returns a verdict whose wrongness a reader may suspect.
Here the verdict **names a specific, plausible, false mechanism** — *"possibly
inlined or renamed"* — which is precisely what a session would act on: go looking
for the rename, fail to find it, and conclude the entry is stale. The phrasing
manufactures a hypothesis the evidence never supported.

## What this adds to the fix

Whatever scoping fix this bug takes, the verifier must distinguish **"the symbol is
absent from the code"** from **"the index cannot represent symbols of this kind"**,
and must not emit a rename/inline hypothesis it has no evidence for. Cheapest
correct behaviour today: treat a footprint that matches no indexable kind as
**UNVERIFIABLE**, never as a change in the code. (An `UNVERIFIABLE` verdict means
"wrong question asked", not "premise false" — the distinction this estate has been
bitten by before.)

**Do not downgrade the 221 entry on this verdict.** It was re-checked by hand and
is accurate.

**2026-08-09 addendum — a SECOND consumer of the same var-blindness: the diagnosis
loop.** A 090 run the same morning (`e952039b`, filed by the 209 lane on the
Defaults-shadow mechanism, `bugs_open/231`) stopped at UNVERIFIABLE with this as
its stated gap 1: *"the actual var declaration of DeployImageAssetInputSpec —
specifically its Defaults map — is NOT in the bundle. The content search for the
identifier only surfaced its two USE sites … never the literal spec definition."*
`var DeployImageAssetInputSpec = …` is a package-level `var` — unrepresentable in
`code_symbols` (no `var` kind), so the loop's `diagnose_code_lookup` can fetch
every *use* of a spec but never its *declaration*. Consequence worth stating: the
diagnosis loop is structurally unable to CONFIRM any hypothesis whose deciding
evidence is the CONTENT of a package-level `var` (a spec's Defaults, a pattern
table, a policy map) — it will stop at UNVERIFIABLE with a "still needed" naming
exactly the declaration it cannot see. Whoever fixes this bug's lookup layer fixes
both consumers at once.

---

## Recurrence 2026-08-09 — a third instance, self-refuting in the same way, from the `bugfix_201` lane

Not a new mechanism; recorded here as **frequency evidence**, because this bug is
still OPEN and UNOWNED and the rate at which it fires is the argument for fixing it.

A `LANDMINES.md` entry filed 2026-08-08 by the `201` lane — footprint
`scripts/landmines-sync.py`, `scripts/landmines-verify-dispatch.sh`,
`scripts/trigger-landmine-verifier.sh`, `LANDMINES.md`, `doc_notes`,
`landmine-verification` — came back:

> **last verified (landmine-verifier): STALE.** All four footprint scripts/files
> (landmines-sync.py, landmines-verify-dispatch.sh, trigger-landmine-verifier.sh,
> LANDMINES.md) are absent from the code_symbols index, and the
> `landmine-verification` category string has zero hits across 5755 indexed
> symbols — **the entire toolchain this entry warns about appears removed or
> relocated.**

**All three scripts had been executed successfully in the minutes before that
verdict was written** — and the verdict itself was delivered *by that toolchain*,
into the `landmine-verification` category it declares has zero hits. Same
self-refutation as this bug's opening case, independently reproduced.

**The index composition, measured rather than inferred** `[MEASURED]` — worth
pinning here because the bug's headline says "holds only Go" and this is the query
that proves it, with every alternative in the same row so a future divergence
cannot hide:

```sql
SELECT count(*) FILTER (WHERE path LIKE '%.go')  AS go,    -- 5755
       count(*) FILTER (WHERE path LIKE '%.sh')  AS sh,    --    0
       count(*) FILTER (WHERE path LIKE '%.py')  AS py,    --    0
       count(*) FILTER (WHERE path LIKE '%.md')  AS md,    --    0
       count(*) FILTER (WHERE path LIKE '%.sql') AS sql,   --    0
       count(*)                                  AS total  -- 5755
FROM code_symbols;
```

Not a single non-Go row exists, so **no `.sh`/`.py`/`.md`/`.sql` footprint can ever
verify** — the STALE is structurally guaranteed and carries no information about the
entry. It is the `UNVERIFIABLE` case this bug's closing section already argues for.

**A second, separate trap found in the same session, worth knowing for whoever fixes
this:** `landmines-sync.py --apply` computes its `NEEDS_VERIFICATION` list as
new-or-changed *relative to rows already in `doc_notes`*. Running it directly —
which **CLAUDE.md instructs** — writes the rows and so consumes the signal;
`landmines-verify-dispatch.sh` afterwards prints *"Nothing needs verification"* and
dispatches nothing, permanently, for that entry. So the population of entries with
**no verdict at all** is larger than the population that got a wrong one, and both
are invisible. Find them with:

```sql
SELECT DISTINCT n.source FROM doc_notes n WHERE n.categories ? 'landmine'
  AND NOT EXISTS (SELECT 1 FROM doc_notes v
                  WHERE v.categories ? 'landmine-verification' AND v.subject_key = n.source);
```

Full entry (with the manual recovery path) is in `LANDMINES.md`, 2026-08-08,
*"`landmines-sync.py --apply` CONSUMES the NEEDS_VERIFICATION signal"*.

**Do not downgrade the `201` lane's other entry on this bug's account** — the
slot-rename entry, whose footprint is Go symbols and DB tables, verified
**STILL_VALID** in the same batch. The verifier is right when the index can see the
footprint, which is what makes the wrong verdicts hard to spot.

---

# PHASE 1 FIXED AND COMMITTED 2026-08-10 — `1058b5366`, council `495df717-4010-491f-aec0-92c13aaf3809`

**STATUS: the Go half is committed and INERT until the next chassis roll; seed 365 is
APPLIED and live. This bug stays OPEN until the roll, because the defect is still
reproducible on `v1.0.1277`.** Owned and worked by the
`bugfix_223_index_answerability` lane —
`docs/agent_docs/docs024_key_docs_latest/bugfix_223_index_answerability/` (PLAN, NOTES,
RUNBOOK, README_where_we_are, and a verbatim BEFORE artefact).

## The cause stands, re-measured, and one clause of it was too narrow

Re-ran this file's own census today. Unchanged in composition, grown in size: **5,837 rows,
100% `.go`, 682 paths, no `var`/`const`/`type` row.** A fact this file does not record and
which changes how the fix should be read: **`code_symbols`' own CHECK constraint already
permits `var`, `const` and `type`**, and the reader's `codeKindList` already treats them as
code. The missing kinds are an **unfinished write path**, not a design gap — which is why
indexing them is not architecture-scope, and is phase 2 of this lane rather than a deferral.

## Two failure modes this file does not contain, both found while fixing it

1. **THE BLIND SPOT ALSO PRODUCES FALSE POSITIVES.** `ls` is a path-*prefix* listing over an
   index of Go symbols and it presents as a directory listing. Measured: `scripts/` returns
   **110 indexed paths** (Go programs under `scripts/documentation_project/`,
   `scripts/goscripts/`, …) while every `.py` and `.sh` **directly** under `scripts/` — the
   three files the motivating entry actually named — is invisible. So a check written at
   directory altitude comes back with a generous listing that reads as **confirmation that
   the footprint resolves**. That is worse than a false `STALE`: a wrong accusation invites
   checking, while a flattering partial confirmation reads as diligence. This is the shape
   §"Measured scope" warns about for mixed entries, arriving through the `ls` kind instead.
2. **A `content` check aimed at a non-Go file can be ANSWERED BY A SAME-NAMED GO SYMBOL.**
   In a run captured today, `content: slugify` — stated purpose *"confirms the slugify
   function exists in landmines_lib.py"* — returned **six** confident Go hits
   (`slugifyPathSegments` in `adopt_verbatim.go`, `slugifyForCompositionName` in
   `resolve_composition_helpers.go`). A false positive **with citations**. The verdict caught
   it, by reading them; nothing made it.

## And the CAREFUL branch is also a total loss, which strengthens this file's argument

This file's correction says the failure mode is that you cannot tell which run you got. A
run dispatched on purpose today (v1.0.1277, banked verbatim as
`EVIDENCE_2026-08-10_prefix_run_verbatim.md`) drew the good branch and it is still a wasted
round. Five of its eight derived checks were unanswerable by construction; the verdict
reached:

> …returned 0 rows …, meaning they are **either not present at the current ref or not
> indexed** (the index covers Go symbols heavily but **may not** cover Python scripts) …
> there is no way to confirm or deny …

Two LLM calls and eight index queries to produce a disjunction that **one census collapses
to a fact** — and it then *guessed* the census correctly ("appears Go-centric") and hedged
everything on the guess. So the cost is not only the wrong verdicts: every careful verdict
is paying for the same missing sentence.

## What phase 1 changed

- **The shared lookup action states what the corpus CAN represent**, from a live per-extension
  and per-kind census (one `GROUP BY` per run). A 0-row answer whose target class the corpus
  cannot hold renders **`NOT ANSWERABLE BY THIS INDEX … It is NOT evidence that the target is
  absent, removed, renamed or inlined, and it must not contribute to a verdict of STALE`**,
  replacing *"The query was RUN; this is not an unanswered question."* A symbol miss names the
  kinds with **zero** rows — which is what removes the invented *"possibly inlined or
  renamed"*. A non-empty `ls` answer says what it is a listing **of**. **Every sentence is
  computed**, so all of them retire themselves when the corpus widens; a test asserts that in
  advance.
- **Four additive return keys** — `checks_with_rows`, `checks_unanswerable`,
  `no_code_evidence`, `evidence_line`. Measured before adding: 0 live definitions referenced
  any of them, or `checks_run`. **LANDMINE: `checks_run > 0` does not mean anything was
  verified** — it counts checks that executed, unanswerable ones included, and always did.
- **Seed 365 is the structural bar this file asked for.** `run_checks → gate_evidence`
  (`conditional_branch` on `no_code_evidence`) `→ {verify_unverifiable | verify}`. A round
  that confirmed nothing against indexed code **cannot reach** the STALE-bearing prompt;
  the new branch's vocabulary is `UNVERIFIABLE | NEEDS_HUMAN_REVIEW` — and per this file's own
  correction, `STILL_VALID` is absent from it too, because with no checkable evidence both
  directions are uninformative. The evidence branch keeps its full vocabulary plus rules
  bounding what a `STALE` may rest on (the mixed case).
- **Every verdict row now carries the action's own census**, appended by an opt-in
  `append_doc_note` key (`note_body_suffix_field`) — mechanical, so the model whose verdict
  it qualifies cannot soften or omit it. This is the half that protects the *reader*, and it
  is why candidate 1 was ranked above the branch rather than below it.

## What phase 1 does NOT fix — stated so nobody reads coverage into it

**The `var`/`const` blind spot itself is untouched.** The 2026-08-09 addendum's class — a
hypothesis whose deciding evidence is the CONTENT of a package-level `var` — still stops at
`UNVERIFIABLE`. It now stops **honestly**, naming the index as the limit instead of inventing
a rename. Phase 2 indexes those kinds (measured **1,173** package-level specs, ~+20% corpus)
and is its own commit and its own council round, because its blast radius is the corpus —
embeddings, prune cohorts, and every diagnosis run's search space — not the rendering.

Non-Go ingestion stays **excluded** as architecture-scope, per this file's candidate 3 and
the 2026-07-29 ruling. Evidence it is a separate intent rather than an oversight: the
reader-side D12 guard (`kindDoc`, `docBlockHeader`) is built and inert, and the CHECK
constraint does **not** admit `'doc'` — so ingestion needs a schema change and is its own
lane by construction.

## To the other consumers of this seam — you are being told, not merely measured

The change is in `answerCodeCheck`/`emptyAnswer`, so it reaches four consumers:
**`fix-proposer`** (and through the 099 roster mirror, every council seat's `code_lookup`
step), **`feature-designer`**, **`landmine-verifier`**, and **the diagnosis loop's runtime
lane**, which calls `answerCodeCheck` directly at `diagnose_load_runtime_action.go:479` and
is invisible to any `agent_definitions` query.

**What changed about your guarantee:** a 0-row answer whose class this index cannot hold no
longer renders as an answered question, and each round now ends with a one-line census of
what it established. For an in-corpus query nothing changes. Your `cite-or-abstain` gets the
fact it was missing. Nothing parses the answer text mechanically (measured), so no consumer
needs a config change. **If you branch on evidence, use `no_code_evidence`, never
`checks_run`.**

`architecture_review` owns RFC_005 §3.2: the verifier's verdict **vocabulary** changed on the
no-evidence branch (`UNVERIFIABLE` appears; `STALE` and `STILL_VALID` cannot). Nothing parses
status, but it is your mechanism.

## How to close this

1. **Roll.** Then pod-grep with the negative control this lane banked: `NOT ANSWERABLE BY
   THIS INDEX` is **0** on `v1.0.1277` and must be ≥1 in every replica, beside a positive
   control (`this answer is CAPPED` → 1) that proves the grep works.
2. **Re-fire the two motivating entries** and require: no `STALE` attributable to an
   unverifiable footprint; the genuinely-checkable Go facts **still confirmed** (the fix must
   not buy abstention by checking less — the paired before-artefact pins `doc_notes` and
   `subject_key` at ~24 rows each); the verdict an explicit unverifiable-by-index abstention;
   and the persisted row ending `[code-lookup evidence: …]`.
3. **Then a MIXED-footprint entry**, which is the case a partial confirmation flatters: it
   must route to `verify`, confirm what it confirmed, and name what it could not check in the
   same verdict.

**Already proven, on the unchanged binary:** seed 365 is safe ahead of the roll —
`evidence_gate` recorded `{"condition_met": false, "next_step_override": "verify"}`, the run
COMPLETED, the verdict was unchanged and no suffix was written. Three ways to fail, none
taken.


---

# ACCEPTANCE — PHASE 1 PROVEN LIVE, 2026-08-10 14:41Z, chassis `v1.0.1279`

**Artefact first, not the tag.** Both replicas grepped, with the negative control this lane
banked *before* writing any code:

```
NOT ANSWERABLE BY THIS INDEX   → 1   (was 0 on v1.0.1277 — this is the change)
every match above comes from   → 1     not a directory listing      → 1
UNREPRESENTABLE here           → 1     code-lookup evidence         → 2
resolved EMPTY                 → 1   (append_doc_note's suffix)
this answer is CAPPED          → 1   POSITIVE control (181) — proves the grep works
NOT ANSWERABLE BY THAT INDEX   → 0   NEGATIVE control — a string never added
```

## The motivating entry, before and after

**BEFORE (2026-08-08, v1.0.1267)** — flatly `STALE`:

> None of the three scripts … exist anywhere in the indexed codebase; **the entire described
> workflow has no footprint.**

**AFTER (2026-08-10 14:41Z, v1.0.1279)** — same entry, same index, `NEEDS_HUMAN_REVIEW`:

```
[code_check 1] kind=ls query="scripts/landmines-sync.py"
  NOT ANSWERABLE BY THIS INDEX: the corpus holds NO .py file at all — the indexed corpus
  holds only: .go (5837 rows). The query was executed and returned 0 rows, and it COULD NOT
  have returned a row whatever the state of the repository. This is UNKNOWN. It is NOT
  evidence that the target is absent, removed, renamed or inlined, and it must not
  contribute to a verdict of STALE …
[code-lookup evidence: 7 check(s) ran; 1 matched indexed code; 3 NOT ANSWERABLE by this
 index; 3 ran and matched nothing in scope. Scope: 5837 symbols … kinds with NO rows:
 const, type, var.]
```

and the persisted verdict now states the fact instead of guessing it: *"The entire footprint
(Python and shell scripts) falls outside the Go-only code_symbols index."*

**Four entries re-fired; all four persisted rows carry the `[code-lookup evidence: …]`
suffix.** The fix did NOT buy abstention by checking less — the provocation entry's Go half
was still confirmed by name (*"Core Go footprint items (`loadGateCandidates`,
`gate_provocation`, `provocations` table references) confirmed present in
`provocation_gate_action.go` at commit b2371b4b"*) while its non-Go half was named as
unverifiable in the same verdict.

## The third false-positive mode, caught live — and it is COMMON, not rare

The `toolgolden.py` entry (footprint: three `.py` files) produced check
`content: VECTORS`, meant to confirm a **Python constant**. It matched **8 Go rows** —
`vectorSearchCodeSymbols`, `pgvectorString`, `RAGIndexAction`, … — because an ILIKE
substring match on `VECTORS` hits `vectorSearch` and `pgvector`. Pre-fix, that is a
confident confirmation of something that was never checked. Post-fix the new caveat fired,
and the verdict used it:

> The single code_check that returned indexed rows (check 2, searching for 'VECTORS')
> **matched only .go symbols related to vector search in the platform orchestration layer,
> which are unrelated to the VECTORS constant described in the entry.**

## ⚠ A MEASURED LIMITATION OF THIS FIX — the gate's TRUE branch has never fired

`no_code_evidence` is `checks_with_rows == 0`, and across **4 of 4** acceptance runs today
`checks_with_rows` was 1, 3, 1, 1 — **never 0**, so every round routed to `verify` and
`verify_unverifiable` has not executed in production. Reason, measured rather than guessed:
**one accidental substring match is enough**, and `derive_checks` reliably emits at least one
bare-identifier `content` query that hits some Go symbol (`VECTORS`, `doc_notes`, …).

So, honestly: **the branch is a backstop that has not yet been reached, and the evidence
layer is doing all the protective work** — the per-check `NOT ANSWERABLE` lines, the
"shares a name" caveat, and the mechanical suffix on every persisted row. That ordering was
this lane's ranking (candidate 1 above candidate 2) and the acceptance data supports it.

The gate is *correct* — its resolution is proven at build time by a lockstep test, and its
FALSE branch is proven live — but a reader should not credit it with work it has not done.
**Follow-up worth its own round, in `architecture_review`'s mechanism rather than here:**
`derive_checks` should not emit a bare `content` query for a footprint item it can already
see is a non-Go file, which would both cut wasted checks and make `no_code_evidence`
reachable when the footprint really is entirely unverifiable.
