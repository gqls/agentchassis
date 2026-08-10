# BEFORE artefact — one verifier run on v1.0.1277, captured verbatim, 2026-08-10 10:58Z

This is the paired *before* half of this lane's acceptance test. It is not a
reconstruction: it is `orchestration_states.collected_data` from a run dispatched by this
lane on purpose, on the entry it appended minutes earlier, whose footprint is almost
entirely the class the index cannot represent.

- entry: `LANDMINES.md#a-landmine-s-identity-is-its-title-slugged-and-cut-at-80-characters-333-of-356-e`
- correlation: `cb68e33e-d454-46ef-94a9-df85cc93bea4`
- binary: chassis `v1.0.1277` (both replicas), index at commit `b2371b4b`, 20h old, refreshed 2h ago
- how to re-read it:
  ```sql
  SELECT jsonb_pretty(collected_data->'lookup'), collected_data->'verdict'->'result'
    FROM orchestration_states
   WHERE collected_data::text LIKE '%title-slugged%' AND collected_data ? 'lookup'
   ORDER BY created_at DESC LIMIT 1;
  ```

## What the `derive_checks` step asked for — 8 checks, and their answerability

| # | kind | query | rows | could it EVER have matched? |
|---|---|---|---|---|
| 1 | ls | `scripts/landmines_lib.py` | 0 | **no** — no `.py` row exists |
| 2 | content | `slugify` | **6** | not the Python one — matched Go `slugifyPathSegments`, `slugifyForCompositionName` |
| 3 | ls | `scripts/landmines-sync.py` | 0 | **no** — no `.py` row exists |
| 4 | content | `doc_notes` | 24 | yes — Go SQL strings |
| 5 | content | `subject_key` | 24 | yes — Go SQL strings |
| 6 | content | `landmine-verification` | 0 | **no** — the string lives only in `.sh`/`.py` |
| 7 | content | `landmine-verifier` | 0 | **no** — same |
| 8 | content | `load_entry` | 0 | **no** — it is a workflow step name in `agent_definitions`, not Go |

**Five of eight checks were unanswerable by construction. Nothing in the rendered answer
says so, and one of the three answerable ones answered a different question.**

## The scope line the action DID render

```
(index freshness: commit b2371b4b (ref 087_towards_multiple_domains), committed 20h ago;
 refreshed 2h ago. The index mirrors the last pushed tip … local unpushed work is never visible.)
(index scope: 5837 symbols, source bodies indexed for all of them)
```

Freshness: stated. Body coverage: stated. **Language coverage: not stated. Kind coverage:
not stated.** Those are the two facts this run needed and the only two it was not given.

## The wording that does the damage, verbatim

```
[code_check 1] kind=ls query="scripts/landmines_lib.py" — confirms the landmines_lib module still exists at this path
  answered: 0 rows — no indexed path has that prefix, out of 5837 indexed symbols. The query was RUN; this is not an unanswered question.

[code_check 6] kind=content query="landmine-verification" — confirms the landmine-verification component or reference exists
  answered: 0 rows — searched the bodies and declarations of 5837 indexed symbols (5837 with bodies). The query was RUN and found nothing; this is not an unanswered question.
```

Every clause is true. Together they assert the strongest available reading — *this is a
real answer, not an unanswered question* — about a query that could not have returned a
row whatever the state of the repository. That sentence was written for
`bugs_closed/108` defect B, where empty answers were being read as silence, and it fixed
that. It is now the mechanism of the opposite error.

## The flattering answer, verbatim — check 2

```
[code_check 2] kind=content query="slugify" — confirms the slugify function exists in landmines_lib.py
  - platform/orchestration/actions/adopt_verbatim.go : slugifyPathSegments   [body] func slugifyPathSegments(segs []string) string {
  - platform/orchestration/actions/resolve_composition_helpers.go : slugifyForCompositionName   [body] func slugifyForCompositionName(s string) string {
  … 4 more
```

The check's stated purpose was *"confirms the slugify function exists in
landmines_lib.py"*. It came back with six confident Go hits from two unrelated files. The
verdict caught it — *"only Go-side `slugifyPathSegments` and `slugifyForCompositionName`
… which are unrelated"* — but catching it was luck of the draw, not a guard. **A `content`
check aimed at a non-Go file can be answered by a same-named Go symbol, which is a false
positive with citations.**

## The verdict — the CAREFUL branch, and it is still a total loss

`NEEDS_HUMAN_REVIEW`:

> The core footprint files — `scripts/landmines_lib.py` and `scripts/landmines-sync.py` —
> returned 0 rows in the code_symbols index (checks 1 and 3), meaning they are **either not
> present at the current ref or not indexed** (the index covers Go symbols heavily but
> **may not** cover Python scripts). … Since the Python scripts are the mechanical heart of
> the landmine … **there is no way to confirm or deny** that the described behavior still
> holds. The scripts **may simply be outside the index scope (which appears Go-centric)**.

Read what this run cost and what it bought. It spent two LLM calls and eight index
queries to arrive at a disjunction — *not present, or not indexed* — that **one census
would have collapsed to a fact**. The model then *guessed* the census correctly ("appears
Go-centric") from the shape of its own results, and hedged everything on that guess.

This is 223's good case: no false `STALE`, no invented rename. And it is still a wasted
round, because the entry ends with no verification and a reader ends with
`NEEDS_HUMAN_REVIEW`, which here means *the wrong question was asked* and not *a human
should look*. **The stochastic half of 223 is which of those two outcomes you get; the
deterministic half is that the action never told the model what it can see.** Only the
second is fixable in one place, and fixing it removes the input the first varies on.

## Acceptance, restated against this artefact

Re-fire this same entry after the fix and require:

1. checks 1 and 3 render as **not answerable by this index**, naming the extension census
   — not "the query was RUN";
2. checks 4 and 5 still return their 24 rows each (**the fix must not buy abstention by
   checking less**);
3. check 2 carries a caveat that a `content` match is a Go symbol and cannot confirm a
   Python function;
4. the verdict is an explicit unverifiable-by-index abstention, not a hedge built on a
   guess about scope.
