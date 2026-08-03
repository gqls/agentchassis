# PLAN — `bugs_open/163`: the code lookup's symbol arm cannot answer a path-bearing query

**Opened 2026-08-03.** Lane owner: this thread. Bug filed 2026-07-31 by the
`bugfix_145` lane and left OPEN, UNOWNED for three days.

## What is broken, in one paragraph

`symbolTokenClause` (`platform/orchestration/actions/diagnose_code_lookup_action.go:773-797`)
tokenises a `code_check` query into maximal `[A-Za-z0-9_]` runs and requires **every** token
as an `ILIKE` substring of the **`symbol` column**. The `landmine-verifier`'s `derive_checks`
prompt *defines* kind `symbol` as `path:Symbol`. `code_symbols.symbol` never contains a path.
So the producer's contract and the executor's predicate are **unsatisfiable together**, and
every path-bearing symbol check returns 0 rows by construction — then renders as a confident
"the query was RUN and matched none".

## Why the fix goes in Go and not in the prompt or the corpus

The `path:Symbol` contract is precise, already documented, and already what the corpus's
upstream producers emit (`scopeFromCodeResults`, `resolveScopeEntries`, the §7D fuzzy
resolver all compose `path + ":" + symbol`). The prompt is right. The executor is wrong.
Fixing Go changes no prompt, rewrites no corpus, and repairs all three consumers at once:

| consumer | call site |
|---|---|
| council verify tier | `diagnose_code_lookup_action.go:395` |
| diagnosis loop runtime lane | `diagnose_load_runtime_action.go:483` |
| `landmine-verifier` `run_checks` | the `diagnose_code_lookup` action directly |

## Decisions, and their reasons

1. **Split on the LAST colon; bind path→`path`, name→`symbol`.** This is not my idea — it is
   the fix already recorded in 016b §9 `:398-401` and in 163 `:185-190`. Cited, not re-argued.
2. **Reuse by EXPORT, not extraction.** `splitSymbol` (`internal/analysis/symbolbody.go:105`)
   already owns the last-colon convention. The precedent is one function below it: `SliceLines`
   was exported rather than re-implemented, after `prior_art_librarian` caught a council
   proposing to extract a helper that already existed (`WRONG_CALLS.md:7940-7982`).
3. **A left part only counts as a path if it looks like one** (contains `/`, or ends in a
   source extension). Everything without a colon keeps today's behaviour exactly, so the
   `51e0776fb` receiver-form fix (`Type.Method` ↔ `(*Type).Method`, measured false negative on
   run `90e989d5`) cannot regress.
4. **A line reference is not a symbol.** Re-measuring the corpus found a third convention
   nobody modelled: 12 `.go:` footprints are line numbers or ranges (`spawn_actions.go:3066`,
   `run_checks_action.go:773-774`). A naive last-colon split sends `2730` to the `symbol`
   column and reproduces this very bug in a new costume. Detect and degrade to a path check.
5. **A miss at a path discloses; it does not deny.** If the path-qualified query is empty,
   re-run name-only and report both facts. A real symbol can then never read as absent merely
   because its file moved — which is precisely the shape that produced the false verdict.
6. **Narrate the predicate, not just the query.** The query already reaches the model
   (`:394`); the *predicate* does not. That is what lets a per-check fact override the
   run-level staleness banner.

## Explicitly out of scope, and why

- **`scripts/landmines_lib.py`'s comma/paren split.** Not on the verifier's path at all:
  `derive_checks` reads the raw `entry.body` (`landmines_lib.py:166`), never
  `split_footprints`' output. 163's "adjacent defect" section is inert for this reason —
  which is a *stronger* statement than the empirical refutations already recorded there.
- **The live `verify` prompt.** A shared mechanism owned by `architecture_review` via
  RFC_005 §3.2. Tell them; do not edit it (owner ruling 2026-07-29 §3).
- **`bugs_open/181`'s row caps.** Same file, same function, same arm — but another lane's,
  and in flight. A disposition is filed into their bug file rather than silence
  (`bugs_closed/164` is the council ruling that made that mandatory in this family).

## Phasing

1. Export `analysis.SplitSymbol`; make `symbolTokenClause` path-aware.
2. Name-only fallback + predicate narration in the `symbol` arm; correct the wire-format doc.
3. Tests: the 163 regression, the line-ref form, a negative control; prove each by mutation.
4. Council gate (submit before/alongside the commit; `Council-Submitted:` trailer).
5. Commit narrowly → build → roll → prove at both pods with three probes.
6. End-to-end: re-run the verifier over an entry the blind version could not confirm.
7. Register, landmines, 016b, WRONG_CALLS; close to `bugs_closed/`.

## Corrections to the originating brief

> **CORRECTION 2026-08-03 (to `bugs_open/163` itself, not to my own plan).** Two citations in
> the bug file have drifted: `splitSymbol` is at `symbolbody.go:105-114`, not `:76-82` (the
> 145 fix moved it); `split_footprints` is at `landmines_lib.py:51`, not `:44`. Neither
> changes the diagnosis.

> **CORRECTION 2026-08-03 — the bug's fleet figure cannot be reproduced, and that is expected,
> not a refutation.** 163 records "23 of 23 symbol checks path-bearing, all history" from
> `orchestration_states`. That table is retention-clocked; today it holds 32 checks across 5
> runs and **0** of kind `symbol`. I do not quote 23/23 as current. The live-index test
> (path-bearing → 0 rows, correct split → 1 row, and `symbol LIKE '%/%'` → 0 of 4,992) carries
> the claim instead, and does not depend on retention.
