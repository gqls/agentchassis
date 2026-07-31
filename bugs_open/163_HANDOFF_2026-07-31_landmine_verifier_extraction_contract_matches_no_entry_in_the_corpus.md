# 163 — the landmine-verifier's symbol lookup cannot answer a path-bearing query, and it invents a stale-index cause for its own blindness

**Filed 2026-07-31. OPEN, UNOWNED.** Found by the `bugfix_145` lane while doing the
routine thing: appending a `LANDMINES.md` entry and letting RFC_005 §3.2's verifier check
it. **The verifier is the mechanism that decides whether a landmine entry is trustworthy,
so a false negative here degrades the corpus every session reads at start-up.**

Owning lane, for routing: `architecture_review` (RFC_005 §3.2 is theirs;
`scripts/landmines-verify-dispatch.sh`, `scripts/trigger-landmine-verifier.sh` and the
`landmine-verifier` agent were added with it). I have not touched their code — this file
is the account, not a competing fix.

## The symptom, verbatim

`doc_notes`, `categories ? 'landmine-verification'`, 2026-07-31 16:01:44Z:

> **last verified (landmine-verifier): NEEDS_HUMAN_REVIEW.** Core orchestration files
> (`diagnose_assemble_bundle_action.go`, `loop.go`, `verdict_wire.go`) and key references
> (`scopeFromCodeResults`, `route.scope`) confirmed present, but `ReadSymbolBody`,
> `findFile`, `namedScope`, `nextScope`, and `NextScope` returned 0 rows — **most likely
> because the code index predates the bugfix_145 commit** (index at d98010e8 /
> 2026-07-28, entry added 2026-07-31); a human should re-verify once the index catches up.

## The stated cause is FALSE, and that is the worst part

Every symbol it says returned 0 rows **is in the index, at the very commit it names**:

```sql
SELECT symbol, path, kind, commit_sha FROM code_symbols
 WHERE symbol IN ('ReadSymbolBody','findFile','namedScope','nextScope');
```
```
 findFile       | internal/analysis/symbolbody.go | func | d98010e8bc9e0dd098da8b7c614cb5f81be6e281
 ReadSymbolBody | internal/analysis/symbolbody.go | func | d98010e8…
 namedScope     | pkg/diagnose/loop.go            | func | d98010e8…
 nextScope      | pkg/diagnose/loop.go            | func | d98010e8…
```

`ReadSymbolBody` and `findFile` have existed since long before 2026-07-28, so "the index
predates the commit" could not have explained them even in principle. **The verifier
narrated a plausible, checkable, wrong cause in the same confident voice as its true
findings** — and the remedy it prescribes ("re-verify once the index catches up") would
never have worked, because nothing was stale.

> ## CORRECTED TWICE on 2026-07-31, within the hour. Read this banner and not the sections below it for the cause.
>
> **My first two root causes were both wrong, and both were about the footprint convention.
> The convention is irrelevant.** Kept rather than deleted, because the two refutations are
> what located the real defect.
>
> - **Theory 1 (filed): the comma split mangles parenthesised symbols.** Real parsing defect
>   (`split_footprints`, `scripts/landmines_lib.py:44`), measured 17 of 100 entries affected —
>   **and not the cause.** Refuted by corr `113fd03f`: rewriting my footprint into clean,
>   comma-safe `path:Symbol` form produced the *same* 0 rows and the *same* invented
>   staleness cause.
> - **Theory 2: then separate the path and symbol into their own items.** Refuted by corr
>   `f7056e8a`: same result again. I had asserted that form "works" from reading the SQL
>   instead of waiting for the verdict — the identical error, twice in one file.
>
> ### CONFIRMED CAUSE — two mechanisms that cannot both be satisfied
>
> **1. `derive_checks` emits `path:Symbol` no matter how the footprint is written.** Its
> prompt *defines* kind `"symbol"` as "a file path followed by a colon and a Go identifier",
> so the model reconstructs that form from the neighbouring path item. Read straight out of
> the `f7056e8a` run's `collected_data->'derived'->'result'->'code_checks'`:
>
> ```json
> {"kind":"ls",     "query":"internal/analysis/symbolbody.go"}
> {"kind":"symbol", "query":"internal/analysis/symbolbody.go:ReadSymbolBody"}
> {"kind":"symbol", "query":"pkg/diagnose/loop.go:namedScope"}
> ```
>
> **2. `symbolTokenClause` cannot answer a path-bearing query.**
> `platform/orchestration/actions/diagnose_code_lookup_action.go:781-797` tokenises the query
> on non-identifier characters and requires **every** token as a substring of the **`symbol`
> column** — path tokens included. So query 2 above becomes:
>
> ```sql
> symbol ILIKE '%internal%' AND symbol ILIKE '%analysis%' AND symbol ILIKE '%symbolbody%'
>   AND symbol ILIKE '%go%'  AND symbol ILIKE '%ReadSymbolBody%'
> ```
>
> | query as executed | rows |
> |---|---|
> | the path-bearing form above | **0** |
> | bare `ReadSymbolBody` | **1** |
> | bare `findFile` | **1** |
>
> The token-AND is itself a *correct* fix for receiver forms (`Type.Method` vs
> `(*Type).Method`, per the comment at `:534-542`, a measured false negative on run
> `90e989d5`). It simply never anticipated a path in the query, and `identifierTokens`
> cheerfully tokenises `internal`, `analysis`, `go`.
>
> ### Fleet measurement — every symbol check ever run has returned 0 rows
>
> ```sql
> WITH runs AS (SELECT orchestration_id,
>        jsonb_array_elements(collected_data->'derived'->'result'->'code_checks') AS chk
>   FROM orchestration_states
>  WHERE collected_data->'derived'->'result' ? 'code_checks')
> SELECT count(*) FILTER (WHERE chk->>'kind'='symbol')                       AS symbol_checks,
>        count(*) FILTER (WHERE chk->>'kind'='symbol' AND chk->>'query' LIKE '%/%') AS path_bearing
>   FROM runs;
> ```
> → **23 symbol checks across 9 runs; 23 of 23 path-bearing.** All history. So **no landmine
> entry has ever had a symbol mechanically confirmed**, the `ls` (bare path) checks are the
> only thing that has ever resolved, and **there is no footprint form an author can write to
> avoid it.** That last point is why this is worse than a convention mismatch: the two
> contracts are unsatisfiable together, and the corpus cannot route around them.
>
> Everything below this banner about the paren/comma convention describes a **real but
> INERT** defect: fixing it changes nothing until mechanism 1 or 2 changes. Read it as fix
> candidate 5, not as the cause.

## An adjacent real-but-inert defect: the extraction contract and the footprint convention also disagree

`derive_checks` (the `landmine-verifier` agent's `execute_llm_prompt` step) instructs:

> kind `"symbol"` — for a `path:Symbol` reference (**a file path followed by a colon and a
> Go identifier**)

But `split_footprints` (`scripts/landmines_lib.py:44`) **splits the footprint field on
commas**, and the house convention writes symbols parenthesised after the path. So mine
arrived as:

```
internal/analysis/symbolbody.go (ReadSymbolBody     ← unbalanced paren
findFile)                                            ← bare identifier, trailing paren
pkg/diagnose/loop.go (namedScope
nextScope)
```

A lookup for `findFile)` or `internal/analysis/symbolbody.go (ReadSymbolBody` matches
nothing. The files with no parenthesised symbols (`verdict_wire.go`) and the bare
identifiers that happened to survive intact (`scopeFromCodeResults`, `route.scope`)
resolved — which is exactly the pattern in the verdict, and why it looked like a
selective staleness problem rather than a parsing one.

## Measured scope — this is not about my entry

```
entries: 100 | footprints naming a .go file: 45
  written as `file.go (Symbol)` paren form: 17
  written as `file.go:Symbol`  colon form :  0
```

**Zero of 100 entries use the form the verifier's prompt asks for.** 17 use the paren form
that the comma split actively mangles. So the extraction contract matches nothing in the
corpus it was built to check.

Consistent with the outcome distribution: **5 of the 6 landmine-verification verdicts to
date are `NEEDS_HUMAN_REVIEW`** (`SELECT count(*) FILTER (WHERE body LIKE
'%NEEDS_HUMAN_REVIEW%') FROM doc_notes WHERE categories ? 'landmine-verification'` → 5/6).
Small n, and the mechanism is days old — quoted as corroboration, **not** as the finding.

## The experiment trail, kept because the refuted step is the useful one

Three verdicts on the **same entry**, one hour apart, differing only in footprint form. All
three are in `doc_notes` (`categories ? 'landmine-verification'`, `subject_key` ending
`…readsymbolbody-s-whole-file-branch…`) — read them side by side, because the value is in
what they eliminate:

| corr | footprint form | verdict |
|---|---|---|
| `16:01:44Z` | house paren form — `` `symbolbody.go` (`ReadSymbolBody`, `findFile`) `` | 0 rows; "most likely because the code index predates the bugfix_145 commit" |
| `113fd03f` | comma-safe `path:Symbol` — the form `derive_checks` asks for | **0 rows, same invented cause** → kills the comma-split theory |
| `f7056e8a` | path and symbols as **separate** items | **0 rows, same invented cause** → kills the "just separate them" workaround too |

Two independent negative results on the author-side variable are what forced the search down
into `derive_checks` and `symbolTokenClause`, where the cause actually is. **The footprint
form is not a variable that can affect this outcome** — mechanism 1 in the banner normalises
every form back to `path:Symbol` before the lookup sees it.

**The colon form does preserve SessionStart path matching** — `matches()`
(`scripts/landmines-session-start.py:68-71`) is substring both ways, so `path.go` matches
`path.go:Symbol`. That was never the problem either.

**This entry's footprint is left in the separated form** (`…symbolbody.go`,
`ReadSymbolBody`, `findFile`). Not because it verifies — it does not — but because it is the
form that survives `split_footprints` intact, so it will start working the moment mechanism
1 or 2 is fixed, and it costs the SessionStart hook nothing.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Stop matching path tokens against the `symbol` column.** In `symbolTokenClause`, split
   the query on the LAST colon (or on `/`-bearing tokens) and match the path part against
   `path` and the name part against `symbol`. `ReadSymbolBody`'s own `splitSymbol`
   (`internal/analysis/symbolbody.go:76-82`) is the convention already in the tree — reuse
   it rather than writing a third parser. **Makes the bad state unrepresentable:** a
   path-bearing symbol query can no longer be silently unsatisfiable.
2. **Make the verifier unable to invent a cause it has not checked.** The `verify` step
   should receive the *queries it ran* and be forbidden from attributing a 0-row result to
   index staleness unless it has read `code_symbols.commit_sha` and compared it. "0 rows
   for X" is a fact; "most likely because the index is stale" is a hypothesis rendered in
   the same voice, and the corpus cannot tell them apart. **This one outranks (1) in
   damage even though (1) is the mechanical cause** — the lookup bug produced a wrong
   answer, but the *narration* is what sends a human to wait for a re-index that was never
   needed, and it will keep doing so for the next unrelated 0-row cause.
3. **Emit the derived `code_checks` into the verdict body.** A verdict that showed its
   query as `internal/analysis/symbolbody.go:ReadSymbolBody` would have been
   self-diagnosing, and I would not have needed three rounds to find this.
4. **Reconcile the extraction contract with the corpus, in one direction, and say which.**
   `derive_checks` asks for `path:Symbol`; **0 of 100 entries use it** and 17 use the paren
   form the comma split mangles. Pick one and put it in CLAUDE.md's landmines bullet.
5. **Make `split_footprints` paren-aware** — do not split on a comma inside `(...)`. Fixes
   the 17 paren entries without a corpus rewrite. Weakest alone: with (1) unfixed it just
   produces well-formed queries that still return 0.

## How to verify a fix

- Re-run the verifier over an entry using the **paren** form (16 others exist) and confirm
  its symbols resolve.
- Negative control: an entry naming a genuinely absent symbol must still come back
  unconfirmed — and must say *"not found in the index"* without a fabricated cause.
- Assert the mechanism fired: a `NEEDS_HUMAN_REVIEW` for a real absence and a `CONFIRMED`
  for a present one, from the same run shape.

## Related

- `RFC_005` §3.2 (the dispatch this rides on) — `architecture_review/`.
- **`WRONG_CALLS.md`, 2026-07-31 (same lane):** running `landmines-sync.py --apply` before
  `landmines-verify-dispatch.sh` **consumes the `NEEDS_VERIFICATION` signal**, so the
  wrapper fires nothing and says "Nothing needs verification". CLAUDE.md names only the
  inner script. That is a second, independent way this verifier silently does not run, and
  it should probably be fixed in the same pass.
- Family: MEMORY's *"a claim about behaviour is NOT the behaviour"* and *"a PASS from a
  BLIND check outlives the blindness"* — this is the same shape with the sign flipped: a
  FAIL from a blind check, carrying an invented explanation that will outlive it.
