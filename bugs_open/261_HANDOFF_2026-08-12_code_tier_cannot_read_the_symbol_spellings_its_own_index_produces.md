# 261 — the diagnosis code tier could not read the symbol spellings its own index produces

**Filed:** 2026-08-12, `silent_hero_logo_readers` lane (follow-on from commission item 2).
**Status:** **FIXED + council-APPROVED + committed (`6911c2da4`) + LIVE on `agent-chassis:v1.0.1293`
(2026-08-12 19:14Z, both replicas). OPEN pending the BEHAVIOURAL proof** — a post-roll bundle that
actually renders a method body. Verification run in flight, see §6.

**Live, verified at the artefact and NOT at the tag** (2026-08-12 ~20:15Z):

```
kubectl -n ai-persona-system get pods -l app=agent-chassis   # v1.0.1293, 2 replicas, both Ready
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=3000 | grep -m3 'build provenance'
#   -> git_commit 7a1887e3163af75ce5eb5c6cb67ba2c9be37d88e, on BOTH pods
git merge-base --is-ancestor 6911c2da4 7a1887e3   # -> YES, the fix is in the running build
```

⚠ **With a control that discriminates, because my first one did not.** I initially "controlled" with
a commit from earlier the same day — but the build is from 19:57 BST and *everything* I committed
today precedes it, so that control could not have come out false. The real control is a commit made
**after** the build (`81c508bca`): `git merge-base --is-ancestor 81c508bca 7a1887e3` correctly
returns NOT an ancestor. **A control drawn from the wrong side of the boundary is not a control.**

**Council:** `6b0cc25b-1368-4fe2-87f0-bb3aa87019c0` — **APPROVED round 1**, submitted 14:26Z, decided
14:32Z (six minutes), 12 seats, 4 abstained, *"1 advisory objection — none high-severity"*. The
`architecture` seat signalled `point_fix`: *"existing architecture being made to work as designed,
not new architecture being added under cover of a fix"* — which answers the RFC question in the
negative. **The one medium objection was right about something and is what produced §7b below.**
The commit carries `Council-Submitted:` because it predated the verdict; `098` credits it at report
time and forward-only forbids an amend.

> ⚠ **NUMBER COLLISION, and it happened during this filing.** The commit `6911c2da4`, the submission
> JSON and three comments inside `internal/analysis/symbolbody.go` all say **`bugs_open/260`**. They
> mean THIS case. Between my checking that 260 was free and my writing the file, another session took
> 260 for `one_mistyped_llm_field_silently_degrades_a_whole_component_render`. Forward-only forbids an
> amend, so the stale references stand and a follow-up commit corrects the code comments to 261.
> **Resolve by slug, and `git log` the file path — not the number.** This is exactly the hazard
> CLAUDE.md documents; the interval that beat me was under an hour.

---

## 1. Symptom

A `090` diagnosis run reaches `UNVERIFIABLE` while the evidence it needed is sitting in the code
index the whole time. The bundle shows, where a function body should be:

```
_(body unavailable — could not be read from the analysed checkout: ReadSymbolBody: symbol
"(*SagaCoordinator).applyResponseToState" not found in platform/orchestration/coordinator.go.
This is a TOOLING failure, NOT evidence about the code: draw no conclusion from this symbol's absence.)_
```

The marker is honest — it was added by `bugs_closed/164` precisely so a defect could not read as
coverage — but the verdict prompt's cite-or-abstain rule still has nothing to cite, so the run
abstains. **The loop is not wrong; it is starved.**

Motivating run: `dbcc4259-ab84-494b-a48b-1df647209a40` (a 090 on `bugs_open/236`), COMPLETED
2026-08-11 18:42Z, 4 iterations, verdict `UNVERIFIABLE`. Its `needed_evidence`, verbatim:

> *"The bundle never renders the bodies of `persistAwaitingStateWithRetry`, `processAwaitResponse`,
> or `applyResponseToState` — only `storeActionResult`'s body and a bare signature line for
> `applyResponseToState` are present."*

## 2. Root cause — two gaps in one function, `analysis.spanOf`

`analysis.ReadSymbolBody` is the single door a `path:Symbol` scope entry goes through to become
source text. It could resolve neither of the two spellings that actually reach it.

**GAP 1 — METHODS (312 of 335 failures).** The indexer writes every Go method receiver-qualified:

```go
// platform/orchestration/actions/code_symbols_actions.go:598
name = "(" + fn.Receiver.Type + ")." + fn.Name      // -> (*SagaCoordinator).applyResponseToState
```

`diagnose_assemble_bundle`'s `scopeFromCodeResults` (`:725-746`) concatenates that column verbatim
into a scope entry (`path + ":" + symbol`). `spanOf` then split on the last dot and compared the
**raw** prefix against `receiverType()`:

```go
wantRecv, wantName = name[:i], name[i+1:]              // wantRecv = "(*SagaCoordinator)"
if wantRecv != "" && receiverType(fn) != wantRecv {    // receiverType = "SagaCoordinator"
    continue                                           // -> never matches. Ever.
}
```

All **1,170** indexed methods were unreadable by their own canonical name.

**GAP 2 — PACKAGE-LEVEL VALUES (22 of 335).** `FileInfo.Values` has been populated since
`bugs_open/223` phase 2 and the indexer writes **1,238** `var`/`const` rows. `spanOf` searched only
`Functions` then `Types` — it never looked at `Values`, so every one answered "symbol not found".
The same drift as 223 itself, in the opposite direction: there the writer was behind the reader,
here a reader was left behind the writer.

## 3. Why this is not an edge case — our own documentation taught the failing input

The parenthesised form is the **canonical** one in this estate, and the bundle propagates it:

- the bundle's own `## Index search results` section renders rows as
  `check_image_source_unsatisfiable.go : (*ImageSourceUnsatisfiableCheck).Run` and
  `plan_sections_action.go : (*sourceResolver).ensureAssets`, and invites the model to name symbols
  in `next_scope`;
- `LANDMINES.md`, entry added **2026-08-11** by the `diagnosis_schema_visibility` lane, instructs a
  `code_request` author in terms: *"asking for a method by its bare name is not a narrower query, it
  is an **unanswerable** one. Name it `(*Receiver).Method` …"*.

**Following our own documented remedy is what triggered the defect.** That entry is correct for the
index-query path and wrong for the body-read path, and nothing distinguished the two. It has now
been corrected in place.

The reason it survived: `symbolbody_test.go` asserted the **dotted** `Type.Method` form — a spelling
**no producer emits**. Two checks blind in the same way agreed with each other.

## 4. Measurement

All figures live, **as at 2026-08-12 ~15:30 BST**, over every bundle ever assembled.

| | |
|---|---|
| bundles all-history | 460 |
| symbol-read failures | **335**, across **47** distinct diagnosis runs |
| — receiver form `(*T).M` | **312** |
| — package-level `var`/`const` | **22** |
| — genuinely absent symbol | **1** |

```sql
WITH f AS (SELECT correlation_id,
       (regexp_matches(body,'ReadSymbolBody: symbol "([^"]+)" not found','g'))[1] AS sym
       FROM diagnosis_artifacts WHERE kind='bundle')
SELECT count(*), count(DISTINCT correlation_id),
       count(*) FILTER (WHERE sym ~ '^\(\*?[A-Za-z0-9_]+\)\.') AS receiver_form
FROM f;
```

> **CORRECTION, and it is mine.** I first recorded this as *"321 failures, and NOT ONE was a
> genuinely absent symbol"* — into the council submission and the commit message, **before** I had
> checked. The 20 non-receiver names *looked* like values, and I classified them by reading them.
> Querying `code_symbols.kind` for all 20 returned **19**, not 20: `controllerAddress`
> (`platform/kafka/topic_manager.go:318`) is a plain `func`, absent from the analysed snapshot
> because the index sat at `46b507ed1` (2026-08-11 18:49) while the function arrived in `e1f960ac2`
> (2026-08-12 14:20). That one is an **index-staleness** case — `bugs_closed/108`'s class, not this
> one — and **the fix does not cover it.** So the honest claim is **334 of 335 (99.7%)**, not 100%.
> The check that caught it is the one I should have run before writing the sentence: ask the index
> for `kind`, never infer a kind from a name.

> **The census is a moving target, which is itself the finding.** Measured 321 failures / 44 runs at
> ~14:50 BST and 335 / 47 at ~15:30 — **14 further lost bodies in forty minutes.** Any figure here is
> a snapshot; re-run the query rather than quoting it.

## 5. The fix (committed `6911c2da4`, NOT live)

`internal/analysis/symbolbody.go` — one function, one new package-local helper:

1. **`splitReceiver`** normalises `(*T).M`, `(T).M`, `*T.M` and `T.M` all to receiver `T`, and
   `spanOf` uses it instead of the inline last-dot split.
2. **`spanOf` searches `fi.Values`** after Functions and Types, on the bare-name branch only.

**Fixed at the reader, not at the producers**, deliberately: there are two independent producers of
a scope entry — the code-search fallback (our code) and an LLM's `next_scope` copied from whatever
the bundle showed it (not our code) — so normalising at `scopeFromCodeResults` would fix one and
leave the other broken. `symbolbody.go` already owns this grammar by precedent: `SplitSymbol` and
`SliceLines` were **exported** rather than re-implemented for exactly this reason
(`bugs_closed/163`, `bugs_closed/189` — and 189 was filed at the council gate's own direction when a
third hand-rolled copy of the split had drifted).

**Blast radius is bounded: exactly one live caller**, `diagnose_assemble_bundle_action.go:242`
(`grep -rn ReadSymbolBody --include=*.go platform/ internal/ pkg/ cmd/ | grep -v _test`). This
matches the correction already recorded in `bugs_closed/145` — `cmd/assembler` is archived reference
code under its own `go.mod`, outside this module's build.

**The 145 read boundary is untouched.** Output membership still decides which *paths* are readable;
this normalises a *symbol name* within an already-admitted file and cannot name a path.
`TestReadSymbolBodyRefusesUnanalysedPaths`, including its traversal cases, was re-run unchanged.

### How it was proven

- The regression test was written against **unmodified HEAD** and **measured failing there first**,
  on exactly the five broken spellings, while the two dotted forms passed.
- **The negative controls were proven load-bearing by mutation**: disabling the receiver comparison
  (`if false && wantRecv != "" …`) makes `(*Nope).Greet` resolve to `Greeter`'s body and the test
  fails. Without that, a fix that ignored the receiver entirely would pass every positive case and
  silently break disambiguation.
- A **premise guard** fails loudly if the analyser ever stops recording `Values`, so the var/const
  cases cannot pass vacuously.
- The fail-then-pass **and** the mutation both ran in a throwaway `git archive HEAD` checkout, so the
  shared tree never held a mutation — this lane's own §7 lesson from `038211dd8`.
- `go build ./...` clean; `go test ./internal/analysis/ ./platform/orchestration/actions/` green.

## 6. Verification — step 1 DONE, step 2 IN FLIGHT

Per `bugs_open/236` §3's rule about unfalsifiable zeros, do not call the loop fixed on a quiet count.

1. ✅ **Live in the running build** — done, with a discriminating control. See the status block above.
2. 🔄 **The behavioural proof, in flight.** Re-ran the 090 on `236` §5 with the **symptom text
   verbatim from the run that failed**, so this is a controlled re-run: same question, same three
   methods in scope, fixed harness.
   - intake correlation `ab65485f-e00a-4dc2-90de-ba8ba9c275ef`
   - **`RUN_CORRELATION_ID=eddaf1af-b44d-4bc0-8485-5885056042cd`** ← the artefacts are under THIS
   - dispatched 2026-08-12 ~20:20Z. Prior run took ~19 minutes for 4 iterations.
3. **The pass condition is POSITIVE, not an absence**: a bundle under `eddaf1af-…` that renders
   `(*SagaCoordinator).applyResponseToState`'s **body** (~4,746 chars), not merely its name. A drop
   in the failure count is weaker evidence — it also falls if nobody asked.

```sql
-- did a method body actually render this time?
SELECT iteration, length(body) AS len,
       (body LIKE '%func (s *SagaCoordinator) applyResponseToState%') AS body_rendered,
       (body LIKE '%ReadSymbolBody: symbol%not found%')              AS still_failing
FROM diagnosis_artifacts
WHERE correlation_id='eddaf1af-b44d-4bc0-8485-5885056042cd' AND kind='bundle'
ORDER BY iteration;

-- the verdict is NOT in diagnosis_artifacts for a 090 — it is on the run's own row
SELECT collected_data->'verdict' FROM orchestration_states
WHERE correlation_id='eddaf1af-b44d-4bc0-8485-5885056042cd';
```

⚠ **Two ways to misread the result.** (a) A verdict of `UNVERIFIABLE` **for a different reason** is
not a failure of this fix — read `needed_evidence` before concluding anything. (b) A `CONFIRMED` on
236's mechanism is a bonus, not the test: this run exists to prove *the harness can read method
bodies*, and it passes on that even if the hypothesis about 236 turns out false.

```sql
-- after the roll: failures should stop accruing on NEW bundles
SELECT date_trunc('hour', created_at) AS hr, count(*)
FROM diagnosis_artifacts
WHERE kind='bundle' AND body LIKE '%ReadSymbolBody: symbol%not found%'
GROUP BY 1 ORDER BY 1 DESC LIMIT 12;
```

## 7. What this unblocks

- **`bugs_open/236` §5's root cause**, which is otherwise stuck: two `090` runs have now failed on
  the same code-path question, and a third would have failed identically.
- Every future diagnosis whose evidence is a **method body** — which, at 1,170 indexed methods
  against 3,781 plain funcs, is roughly a quarter of the callable surface.
- It also closes a question `bugs_closed/…` commission item 5 explicitly left open. That lane's PLAN
  §3: *"whether the code tier has an analogous blind spot is unexamined `[UNMEASURED]`"*. It is
  measured now, and the answer is yes — though the mechanism is **not** the one item 5 fixed in the
  schema tier (see §9).

## 7b. THREE producers, not two — and the middle one rewrites a working entry into a broken one

Added 2026-08-12 after the council's `prior_art_librarian` seat objected that the submission audited
only two producers of a `path:Symbol` handle. It was right, and the third one matters:

- `knownScopeIdentities` (`diagnose_route_action.go:541`) builds the "already exact, pass through
  untouched" set from the **analyser Output**'s `functions` and `types`, as **BARE** names
  (`syms[path+":"+name] = true`).
- So an entry written in the **index** spelling — the spelling the bundle's own index-results
  section shows the model — is **not recognised as exact**, and falls through to the fuzzy resolver.
- `resolveScopeEntries` (`:651`) searches `code_symbols` and re-emits `add(path + ":" + symbol)`
  at `:700` — **the index spelling again** — logging `"diagnose_route: resolved fuzzy scope entry"`.

> **The enrichment step converts a scope entry into the one spelling the body reader could not read,
> and reports it as resolved.** A model that wrote the bare name — which worked — could have it
> rewritten into the failing form by the step whose stated contract is *"no worse than not
> resolving"*.

**This is why the fix belongs at the reader.** Three producers emit the failing spelling and only
two are code we control; `spanOf` is the only place that sees all three.

**Same drift, fourth instance:** `knownScopeIdentities` iterates `{"functions", "types"}` and not
`values` — the identical omission `spanOf` had. Post-fix its only cost is a wasted embedding call
and `code_symbols` read per package-level `var`/`const` entry, since the body resolves either way.
Follow-up, not a defect. See §8.

## 8. Follow-ups deliberately NOT folded in

All real, all separate, none fixed:

1. **`siblingSignatures` renders `fn.Name` bare** (`diagnose_assemble_bundle_action.go:695`), so the
   same method appears under a *different* spelling from the one the index section shows. The model
   is shown two grammars for one symbol and told to pick. It also means the in-scope dedup
   (`named[fn.Name]`) does not suppress a method that IS in scope under its receiver-qualified name,
   so it can be listed as a sibling of itself.
2. **The per-file sibling cap of ~10 signatures** hid the three functions this run needed behind
   `_(+79 more in this file — put the bare file path in next_scope to see it whole)_`. Iteration 4's
   scope had collapsed to five symbols, three of them copies of a trivial `getMapKeys`.
3. **`knownScopeIdentities` omits `values`** (`diagnose_route_action.go:541`, see §7b) — a
   package-level `var`/`const` entry is never recognised as exact, so every one costs a needless
   embedding call and `code_symbols` read. Cosmetic once this fix rolls; it loses no evidence.
4. **A precedent check against council history for `symbolbody.go`/`spanOf` was asked for by two
   seats and not run.** The two prior council rounds on this file (`bugs_closed/163`, `189`) are
   both *reuse* rulings, not this defect, but that is from the bug files rather than from
   `diagnosis_artifacts`. Owed.

## 9. Corrections to the handoff that routed this work

`silent_hero_logo_readers/HANDOFF_2026-08-12_continue_here.md` §3-4 got the conclusion right and two
details wrong. Both are recorded because the second one cost a search.

- **"The index held four bodies; the bundle rendered one."** True of **iteration 4 only**. Iteration
  3 rendered `persistAwaitingStateWithRetry` *and* `processAwaitResponse` with full bodies and
  carried no INCOMPLETE notice at all; iteration 2 rendered `applyResponseToState`. No single
  iteration ever held all four at once, and **the bundle does not accumulate across iterations** —
  the verdict reads the last one. Reading one bundle and generalising to "the tier" is the same
  shape as the wrong call §7 of that handoff already logs.
- **"The likely target is the code-tier gather/assemble path beside `gatherSchema`."** Wrong file,
  wrong package. The defect is in `internal/analysis`, two packages away from the assemble action.
  The handoff's own warning — *"find the producer first and do not assume it is symmetrical with the
  schema tier"* — was right, and its own guess was symmetrical anyway.
- **This is NOT the same defect as item 5's.** Item 5 fixed a *rendering* blind spot where a
  filtered-out table and a non-existent table read identically. Here the rendering is honest and
  already says "TOOLING failure"; what failed is *resolution*, upstream of it. Same consequence
  (the verdict abstains on an absence), different mechanism, different tier, different fix.

## 10. Relations

- `bugs_open/236` — the case whose diagnosis this unblocks.
- `bugs_closed/145` — `ReadSymbolBody`'s read boundary; untouched here, and its regression re-run.
- `bugs_closed/163`, `bugs_closed/189` — the two prior collapses onto `SplitSymbol`; the precedent
  for fixing the grammar at its owner rather than at a call site.
- `bugs_closed/164` — the honest "body unavailable" marker, which is what made this diagnosable at
  all rather than silent.
- `bugs_open/223` phase 2 — added the `var`/`const` kinds this reader never learned.
- `bugs_closed/108` / the index-freshness landmine — the class the single uncovered failure belongs
  to.
- `LANDMINES.md` — the receiver-qualified entry, corrected in place today.
