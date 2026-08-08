# 223 — the landmine-verifier reports every NON-Go footprint as non-existent, because the code index holds only Go — and 284 of 288 entries have one

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
