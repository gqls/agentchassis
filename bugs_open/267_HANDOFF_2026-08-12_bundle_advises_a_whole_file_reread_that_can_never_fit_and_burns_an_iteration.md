# 267 — the bundle advises a whole-file re-read that cannot fit, and the loop spends an iteration obeying it

**Filed:** 2026-08-12, `silent_hero_logo_readers` lane. Found by the run that verified
`bugs_closed/261` — this is the defect that was **behind** it.
**Status:** ~~OPEN, not started~~ → **FIXED IN THE TREE, NOT YET LIVE** (2026-08-12 evening).
Candidates 1 and 2 both implemented, plus two neighbouring instances of the same shape that
the filing did not name (see §7). Go changes are inert until a chassis image is rebuilt and
rolled, so this file stays in `/bugs_open/` — the defect is still reproducible in production.
Council gate: `Council-Submitted: ac23f2f7-9230-403c-8f20-4e18623c1849` (verdict not yet read).

> **Numbering:** 262–266 were all taken by other sessions while `261` was being closed **this
> afternoon**. Checked 267 free at filing. Resolve by slug.

---

## 1. Symptom

A `090` run reaches `UNVERIFIABLE` reporting *"body omitted, too large for the bundle"*, having spent
an iteration asking for a whole file that could never have fitted.

Worked case, run `eddaf1af-b44d-4bc0-8485-5885056042cd`, iteration 3:

```
### platform/orchestration/coordinator.go
_(body omitted — 169139 chars, and 0 of the 60000-char body budget is already spent.
  It was found; it did not fit. Put THIS SYMBOL ALONE in next_scope to read it whole.)_

> **This section is INCOMPLETE.** 0 of 1 in-scope symbol(s) rendered with a body.
```

**The marker is honest and the accounting is correct** — this is `bugs_closed/164` working as
designed, separating "did not fit" (coverage) from "could not be read" (defect). The defect is not
the report. **It is that the loop was ADVISED to make a request that was arithmetically impossible.**

## 2. Root cause — advice offered unconditionally for an outcome that is conditional

Two places in `diagnose_assemble_bundle_action.go` tell the model to re-request a whole file:

- the sibling-signatures section: *"put the bare file path in next_scope to see it whole"*
- the over-cap marker itself: *"Put THIS SYMBOL ALONE in next_scope to read it whole."*

Neither is conditioned on the file fitting. `coordinator.go` is **169,139 chars against a
60,000-char budget** — 2.8×. **No arrangement of scope makes that request succeed**, and the second
message is the sharper failure: it is printed *by the cap*, which at that moment knows both numbers,
and it still advises the re-request. Alone or not alone, it will not fit.

The cost is not just a wasted step. The loop had **3 iterations**; it spent one of them on this, and
its final verdict then asked for exactly the right four symbols — **as `next_scope`, with no
iteration left to spend it**:

```
"next_scope": ["…coordinator.go:persistAwaitingStateWithRetry", "…:storeActionResult",
               "…:processAwaitResponse", "…:applyResponseToState"]
```

Those four bodies total ~13,400 chars — they fit inside the 60,000 budget four times over. **The
loop worked the answer out and ran out of road.**

## 3. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Do not advise what the numbers already refute.** The over-cap marker has `chars` and `budget`
   in hand. When `chars > budget`, drop the "read it whole" sentence and instead name the file's
   largest symbols — the bundle already computes per-file symbol lists for the sibling section, so
   the data is in scope at that point. This closes the door rather than warning about it.
2. **Make the sibling section's advice conditional too** — *"put the bare file path in next_scope"*
   should not be printed for a file whose whole-file size exceeds the body budget.
3. Weaker: raise the cap. Rejected — 169k against 60k is not a tuning problem, and a bigger cap
   spends the verdict model's context on a file it was never going to read.

## 4. Verification

The disconfirming observation is cheap and specific: a bundle whose scope is a bare path to an
oversized file should **not** contain the string "read it whole". Confirm the advice still appears
for a file that *does* fit — otherwise the fix is a blanket removal rather than a conditional one,
and that is the failure mode to guard against.

## 4b. MEASURED 2026-08-12 (~20:05Z) — the `[UNMEASURED]` above is now discharged

Over all **486** bundles ever assembled:

| shape | iterations | runs |
|---|---|---|
| a body omitted for size (any) | 49 | 26 |
| …and the bundle also says "read it whole" | 44 | — |
| **zero bodies rendered, cause = CAP only** | **6** | **5** |
| zero bodies rendered, cause = resolver only (`bugs_closed/261`'s class, now fixed) | 6 | — |
| zero bodies rendered, both causes | 1 | — |

**The load-bearing number is 6 iterations across 5 runs** — an iteration that returned *no code at
all* because everything in scope was too big. At 3 iterations per run that is a third of the run's
budget, spent on a request the bundle itself advised and its own arithmetic refuted.

⚠ **Do NOT quote the 44.** That is co-occurrence — "read it whole" appears somewhere in the bundle
and *a* body was omitted — and the phrase is printed by the over-cap marker itself, so the two are
nearly the same event by construction. It is not 44 wasted iterations.

⚠ **The separation matters and the naive query gets it wrong.** `**This section is INCOMPLETE.** 0 of
N in-scope` is emitted for BOTH causes. Filtering only on that gives 13 iterations and silently
folds `261`'s resolver failures into `267`'s cap failures — two different bugs, one of them already
fixed. The discriminators are the marker words: `did not fit` (cap) versus `could not be read`
(resolver). My first count made exactly this mistake and read 13.

```sql
SELECT count(*) FILTER (WHERE body ~ '\*\*This section is INCOMPLETE\.\*\* 0 of [0-9]+ in-scope'
                          AND body LIKE '%did not fit%' AND body NOT LIKE '%could not be read%') AS cap_only,
       count(*) FILTER (WHERE body ~ '\*\*This section is INCOMPLETE\.\*\* 0 of [0-9]+ in-scope'
                          AND body LIKE '%could not be read%' AND body NOT LIKE '%did not fit%') AS resolver_only
FROM diagnosis_artifacts WHERE kind='bundle';
```

**Live confirmation while this file was being written:** the seeded re-run
`36bd1b42-29b5-4094-9264-94ea80c6194a` hit the cap on its **iteration 2**, having been given a
perfectly-sized seed for iteration 1. So this is current behaviour, not history.

## 5. Why this was invisible until today

It sat **behind** `bugs_closed/261`. Until this afternoon the code tier could not resolve
receiver-qualified method names at all, so a scope naming those methods failed on *resolution* and
never reached the *cap*. Fixing the resolver is what let a run get far enough to hit this.

**That is the expected shape of removing a blocker, not a regression** — but it is exactly the shape
that gets misreported as one. The 261 fix is independently proven: same run, iteration 1, **12 of 12
bodies rendered on the identical scope where 3 of 12 rendered pre-fix**.

## 6. Relations

- `bugs_closed/261` — the resolver defect in front of this one; its §8 records the neighbouring
  sibling-signature cap issue.
- `bugs_closed/164` — the omitted-vs-unreadable distinction. **Working correctly here**; this bug is
  about the advice, not the accounting.
- `bugs_open/236` — the case still waiting on all of this. Re-run seeded with the four symbols is in
  flight as `36bd1b42-29b5-4094-9264-94ea80c6194a`.
- `bugs_open/174` — `seed_scope` confiscated in transit; the reason that seeded re-run checks for the
  "This scope was NOT chosen" marker rather than assuming its seed arrived.

---

## 7. FIXED IN THE TREE 2026-08-12 (evening) — what was changed, and the two extra instances

Candidates 1 and 2 as filed, and **the census in §2 was incomplete: there were four unconditional
invitations, not two.** Grepping for the *shape* rather than the two known strings found the other
two, and both are the same defect read by the same model in the same bundle.

| # | where | was | now |
|---|---|---|---|
| 1 | over-cap marker (`overCapAdvice`) | "Put THIS SYMBOL ALONE in next_scope to read it whole" — always | conditional on `len(body)` vs the **whole** budget |
| 2 | sibling "+N more" | "put the bare file path in next_scope to see it whole" — always | offered only when the file fits (`os.Stat`), unknown → old wording |
| 3 | **coverage SUMMARY line** — NOT in the filing | "N did not fit … — re-request them singly in next_scope" for the whole set | three arms: all / some / none of the omitted would fit singly |
| 4 | **`inScope[path]["*"]` in `siblingSignatures`** — NOT in the filing | a bare-path entry is "whole file already included; no siblings to add" | false exactly when the whole file did **not** render, so that file's signatures are now listed |

**#4 is the one worth reading twice.** It is not an advice string at all, and it is why candidate 1's
"name the file's largest symbols" was only half a fix: the one file the model must sub-divide was
*also* the one file whose symbol list the bundle suppressed. Refusing the request and then
withholding the map would have moved the dead end, not closed it.

The refusal now names what to ask for instead, with exact sizes. Rendered output, from
`TestOverCapAdvice_OversizedWholeFileIsNotOfferedWhole`:

```
### big_file.go

_(body omitted — 32362 chars, and 0 of the 5295-char body budget is already spent. It was found;
  it did not fit. That is larger than the WHOLE 5295-char budget, so NO next_scope can render this
  path — do not re-request it. Name symbols instead; the largest that would fit are
  `big_file.go:Beta` (4795 chars), `big_file.go:(*Big).Delta` (2435 chars), `big_file.go:Gamma`
  (411 chars) (4 symbols in this file; its function signatures are listed under "Same-file
  signatures" below, subject to that section's own cap).)_
```

### 7a. New machinery, and why it lives in `internal/analysis`

`analysis.SymbolSizes(fi, src)` ranks a file's symbols by body size, and
`analysis.FindFile` is the former unexported `findFile`, exported rather than re-walked at the call
site (`bugs_closed/189` is that drift's first instance). Registered as **DIAG-043**.

**Sizes are produced by CALLING `SliceLines`, not by re-deriving its span arithmetic.** That costs
one split of the file per symbol and it buys the property that matters: an advertised size is *by
construction* the number the cap will later compare, so the two cannot drift. A suggested symbol
whose size was computed a second way would be this same bug with an extra indirection — an
impossible request, arrived at more politely.

Handles are printed in the spelling `ReadSymbolBody` resolves — methods receiver-qualified as
`(*Recv).Name`, the canonical `code_symbols` form. `TestSymbolSizes` asserts the **round trip**
(every offered handle resolves, and to a body of exactly the advertised length) rather than
comparing against the fixture's own text, which would let the test and the code agree with each
other while both were wrong about the span convention.

### 7b. §4's disconfirming observation, discharged — and the guard it names

§4 asked for two things and the second is the load-bearing one: *"Confirm the advice still appears
for a file that does fit — otherwise the fix is a blanket removal rather than a conditional one."*

- `TestOverCapAdvice_AchievableRerequestKeepsTheOriginalAdviceVerbatim` — a body that fits the
  budget **alone** and was dropped only alongside what was already spent keeps the original sentence
  verbatim. Both halves of that fixture are asserted before the test looks at anything.
- `TestSiblingSignatures_BareFilePathOfferedOnlyWhenItCouldFit` — one fixture, three runs, **only
  the budget differs**: over → suppressed, under → offered, unknown → byte-identical to offered.
- `bugs_closed/164`'s pre-existing byte-identity control (`TestBundleBodyCap_FittingScopeIsByte
  IdenticalToThePreFixFormat`) still passes, so a scope that fits renders exactly as before.

**Every guard was mutation-verified individually.** Making each of the four conditionals
unconditional in turn fails exactly the test written for it and no other; making `SymbolSizes` emit
bare method names fails exactly the collision assertion. A guard in series with another would have
shown up as a mutation that passed, and none did.

### 7c. What is NOT fixed, and one thing this makes worse before it makes it better

- **`siblingSignatures` still renders methods with a BARE name** while the new marker renders them
  canonically, so a single bundle can now show **two spellings for one method**. It is a real defect
  — a bare name silently resolves to the *wrong* body on a method-name collision — and this change
  makes it MORE reachable, because a whole-file entry that used to be skipped now lists every
  signature in the file.
  > **CORRECTED 2026-08-12, after council round `ac23f2f7`.** This paragraph originally ended
  > "already recorded in `bugs_closed/261` §8.1 … left there rather than quietly widened into this
  > fix", and treated that as sufficient. The `bug_historian` seat objected that a defect living
  > only in a CLOSED file's follow-up list is undiscoverable by any sweep of `bugs_open/`, and it
  > was right: *not folding in the fix* and *not filing the bug* are two different decisions, and
  > this lane made them as one. **Now filed as `bugs_open/269`**, still unfixed, and the non-fold
  > decision stands unchanged.
- **Bundles that hit the whole-file arm now carry MORE signature text** under a fair-shared
  6,000-char cap, which can push other scoped files toward their own "+N more" sooner. Judged worth
  it: the file being sub-divided is precisely the one the model must see, and DIAG-032's per-file
  share plus 600-char floor already exists for this.
- **`bugs_closed/261` §8.2** — the per-file sibling cap of ~10 signatures — is untouched, and it is
  the next thing to bite this same path.

### 7d. Verification when it rolls

Go changes are inert until a chassis image is rebuilt and rolled. When it has:

```sql
-- has any live bundle taken the whole-file arm? (0 until the roll)
SELECT count(*) FROM diagnosis_artifacts
 WHERE kind='bundle' AND body LIKE '%NO next_scope can render this path%';
```

Then re-run §4b's `cap_only` query. **It should stop growing, not go to zero** — the 6 historical
iterations stay in the table for ever, so a zero there would mean the query is wrong, not that the
bug is fixed. Bound it on `created_at` after the roll.

> **CORRECTED 2026-08-12, and this is a WRONG_CALLS-grade error, not a tidy-up.** This paragraph
> originally ended *"take that timestamp from the service's own build-provenance line rather than
> from the tag"* — and I had also written that instruction into the council submission, using it to
> dismiss the `debug_historian` seat's request for a deploy-verification step as "superseded lore".
>
> **That recipe is INOPERATIVE on `agent-chassis`, which is the service this fix ships in.** The
> `build provenance` line is emitted once at startup and the chassis rotates its container log away
> within minutes: measured 2026-08-11, pods started 09:23Z and by 10:07Z `logs --since=24h |
> grep -c "build provenance"` returned **0 on both replicas** — rotated, not unstamped. The
> `LANDMINES.md` entry saying so is dated 2026-08-11 and names this binary explicitly, in triplicate
> across `bugs_open/153`, `cmd/agent-chassis/main.go:53` and register **BLD-019**.
>
> **What caught it:** the council's `debug_historian` and `prior_art_librarian` seats, both HIGH, in
> round 2. The second one named the failure exactly — *"the claim needs the landmine body read
> before it can be trusted, not asserted from memory"*. I had read CLAUDE.md's rewritten §"Building
> & deploying images", quoted its first half, and skipped the sentence four lines later that names
> `agent-chassis` as the measured exception. **I was correcting a reviewer while being wrong.**
>
> ⚠ **And the obvious fallback is worse, not better.** Probing `/proc/1/exe` for *your* commit's sha
> returns absent on a binary that genuinely contains your change: the binary carries **one** commit —
> the build point — not its ancestors. Three absents in a row reads as "my fix did not ship" and is
> actually "wrong test".

**So verify this fix by BEHAVIOUR, which the landmine prescribes for the chassis and which happens to
be free here — the fix's whole output is a string only the new code can emit:**

```sql
-- (a) the witness: a phrase that exists in no previous binary
SELECT count(*) AS new_code_fired
  FROM diagnosis_artifacts
 WHERE kind='bundle' AND created_at > '<roll>'
   AND body LIKE '%NO next_scope can render this path%';

-- (b) THE DEMAND CONTROL, without which (a) is unreadable
SELECT count(*) AS bundles_since_roll,
       count(*) FILTER (WHERE body LIKE '%did not fit%') AS omissions_since_roll
  FROM diagnosis_artifacts WHERE kind='bundle' AND created_at > '<roll>';
```

**Read them as a pair.** `new_code_fired = 0` has two causes with opposite meanings: the code is not
live, or nothing over-budget has been scoped since the roll. Only `omissions_since_roll > 0` with
`new_code_fired = 0` means the fix did not ship. `omissions_since_roll = 0` means the test has not
been run yet, and asserting anything from it would be the unfalsifiable check this estate has already
logged itself doing.
