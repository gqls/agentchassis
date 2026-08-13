# 269 — `siblingSignatures` renders methods BARE, so one bundle shows two spellings for one symbol and the bare one silently resolves to the wrong body

**Filed:** 2026-08-12, `silent_hero_logo_readers` lane, **at the council's direction**
(round `ac23f2f7-9230-403c-8f20-4e18623c1849`, `bug_historian` seat, reviewing `bugs_open/267`).
**Status:** **FIXED IN THE TREE 2026-08-13, NOT YET LIVE.** All three halves done (rendered handle,
canonical de-duplication, first-wins exactness). Go changes are inert until the next chassis roll, so
this stays in `/bugs_open/` — a bundle assembled right now still offers ambiguous bare handles.
Council: **APPROVED first round 2026-08-13**, `Council-Reviewed: e5809ca9-d718-44f6-8d27-6d8cd656dd28` — 13 seats approve, 2 advisory objections, none high-severity (both answered in §8). **§6's `[UNMEASURED]` is discharged — 48 of 1,175 methods, see §6b.**

> **Numbering:** 268 was taken by another session the same day. Checked 269 free at filing.
> Resolve by slug, not by number.

---

## 1. Why this is a new file and not a line in another one

**It already existed as prose, in a CLOSED file, and that is the defect being corrected here.**

It was recorded on 2026-08-12 as `bugs_closed/261` §8 follow-up 1 — "all real, all separate, none
fixed" — which was the right call for 261's author, who was declining to widen a fix, not declining
to track a defect. But `bugs_closed/` is not swept. A defect that lives only inside a closed file's
follow-up list is **undiscoverable by any future grep of `bugs_open/`**, which is where this estate
looks for what is biting.

The council put it plainly (`bug_historian`, medium): *"Deferring this to a risk-note rather than a
filed, tracked follow-up bug is how this class of defect has historically survived past merge — the
risk prose disappears; a filed bug does not."* Its `missing` field asked exactly the question this
file answers: does the divergence have an open bug number distinct from 261, or does it only live
inside a closed file? It only did. Now it does not.

`bugs_open/267`'s author (this lane) explicitly declined to **fix** it while editing the same
function, on the grounds that 261 had already decided not to fold it in. That was right, and it was
also incomplete: *not folding in the fix* and *not filing the bug* are two different decisions, and
they were made as one.

## 2. The defect

`siblingSignatures` (`platform/orchestration/actions/diagnose_assemble_bundle_action.go`) renders
every function in a scoped file as `f.Path + ":" + fn.Name` — the **bare** name, receiver discarded:

```go
line := fmt.Sprintf("- `%s:%s` — `%s`\n", f.Path, fn.Name, fn.Signature)
```

The section's own heading tells the model these are handles it may use:

> `## Same-file signatures (siblings of the in-scope symbols — name these in next_scope to read their bodies)`

For a method, that handle is **ambiguous**. `analysis.spanOf` matches a bare name against
`fi.Functions` and takes the **first** hit, so in a file where two types share a method name, the
bundle offers one handle for two different bodies and the model gets whichever the analyser listed
first. It does not error. It returns the wrong function's source, labelled as the right one — into
a section headed "In-scope code" that the verdict model treats as ground truth.

This is the same class `bugs_closed/261` was filed for and fixed *on the other side*: the canonical
spelling `(*Recv).Name` is what `code_symbols_actions.go:598` writes, and 261 taught
`analysis.splitReceiver` to accept it precisely so a receiver-qualified handle would disambiguate.
The reader learned the grammar; this writer never did.

### 2a. A second, quieter consequence

The in-scope de-duplication is keyed on `named[fn.Name]`. A method that IS in scope under its
receiver-qualified handle is therefore **not** suppressed from the sibling list — so it can be
listed as a sibling of itself, spending budget in a section whose whole purpose is to show the model
what it has *not* already seen. (`bugs_closed/261` §8.1 records this half too.)

## 3. `bugs_open/267` made this MORE reachable, and that is on the record

267 (fixed in the tree 2026-08-12, commit `139dcc3ca`, not yet live) changed `siblingSignatures` so
that a file whose whole-file body was omitted for size now has **every** signature listed, where
previously it was skipped entirely as "whole file already included". That is correct for 267 — it is
the map the model needs to sub-divide its scope — and it means **more bundles will now emit more
bare method names**.

267 also added `analysis.SymbolSizes`, which prints handles in the **canonical** spelling. So since
`139dcc3ca` a single bundle can carry both grammars for one method:

```
_(body omitted — … Name symbols instead; the largest that would fit are
  `big_file.go:(*Big).Delta` (2435 chars) …)_          ← canonical, disambiguates

## Same-file signatures …
- `big_file.go:Delta` — `func (r *Big) Delta()`        ← bare, does not
```

The council named this as the platform's recurring
*"one call site of a shared judgement gets the rigorous fix; the sibling stays heuristic"* shape
(`016b` §9; `bugs_open/093` is the indexed case). It is right that this is another instance.

**Two spellings side by side is not itself the bug** — it is the tell that makes the bug visible,
and it is why this got filed today rather than surviving another month inside a closed file.

## 4. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Render the canonical handle, and de-duplicate on it.** Call
   `analysis.CanonicalSymbolName(fn)` — which exists as of `17734b699`, see candidate 2 — and key
   `named[]` on the same string. Do NOT re-inline `"(" + fn.Receiver.Type + ")." + fn.Name` here;
   that is the copy the council objected to, and a fourth one would be worse than the defect.
   The ambiguity stops being representable rather than being warned about.
2. **Better: give the grammar ONE owner — and HALF OF THIS IS ALREADY DONE.**
   > **UPDATED 2026-08-12, later the same evening (council round 2 of `ac23f2f7`).** Four seats
   > independently — `reuse_agent` (medium), `constitution`, `tooling_provenance`, `architecture` —
   > objected that this was being *filed* when it should have been *extracted*, and they were right.
   > `reuse_agent`'s phrasing is the one to keep: *"the plan names the exact prior collapse history
   > (`bugs_closed/189`) as evidence duplication is a known recurring failure on this codebase, then
   > proceeds to add a third copy anyway … a smaller, cheaper reuse fix than most of what gets
   > deferred to a bug number."* The deferral reason did not survive contact either: "the third copy
   > is in the live row-writing path" argues against changing `code_symbols_actions.go`, **not**
   > against giving the grammar an owner.
   >
   > **`analysis.CanonicalSymbolName(fn FuncDef) string` now exists** (commit `17734b699`) — the
   > inverse of `splitReceiver`, which had only one half. `SymbolSizes` calls it, so there is no
   > third copy and never was one in a shipped tree.

   **So what remains for THIS bug is two call sites, not three:**
   - `code_symbols_actions.go:598` — the live row-writing path. Behaviour-identical to convert, and
     it still wants its own review rather than riding a bug patch. **It should call
     `analysis.CanonicalSymbolName`.**
   - `siblingSignatures`' `fmt.Sprintf("- \`%s:%s\` — …", f.Path, fn.Name)` — §2 above, the actual
     defect this file is about, which is a *missing* call rather than a divergent copy.

   ⚠ **Do not read "the helper exists" as "the bug is half fixed".** The helper closes the DRIFT
   risk; the defect in §2 is that this writer does not call it, and a bundle still offers ambiguous
   bare handles today.
3. Weaker: leave the writer alone and make `spanOf` refuse an ambiguous bare name instead of taking
   the first hit. Turns a silent wrong answer into a loud one — genuinely better than today — but it
   breaks every legitimate bare-name lookup in a file that happens to have a collision, and it
   leaves the bundle still printing a handle it knows to be ambiguous.

## 5. Verification

**The disconfirming observation needs a COLLISION fixture, and that is the whole test.** A file with
two types sharing a method name, both listed as siblings:

- the bundle must offer `path:(*A).Greet` and `path:(*B).Greet`, never a bare `path:Greet`;
- `ReadSymbolBody` on each offered handle must return **different** bodies (the assertion that a
  bare handle cannot satisfy — `internal/analysis/symbolbody_test.go`'s `symbolBodyFixture` already
  carries exactly this pair, `Greeter.Greet` and `(*Helper).Greet`, and `TestSymbolSizes` already
  asserts it for the other producer);
- a method that IS in scope under its canonical handle must NOT appear in the sibling list (§2a).

**A test with no collision in the fixture asserts nothing here** — every spelling resolves when only
one candidate exists, which is why this survived: it is invisible on ordinary input.

## 6. Blast radius — ~~[UNMEASURED]~~ MEASURED, see §6b. Original note kept below for the caveat it carried

Not measured at filing. What matters is not how many methods are rendered bare (all of them) but how
many are rendered bare **in a file that contains a name collision**, since only those can resolve
wrongly. That is answerable from the code index without running anything:

```sql
-- files where one symbol name is held by more than one receiver: the only files
-- where a bare handle can silently return the wrong body
SELECT path, symbol, count(*) AS receivers
FROM code_symbols
WHERE kind = 'method'
GROUP BY path, regexp_replace(symbol, '^\(.*\)\.', ''), symbol
HAVING count(*) > 1
ORDER BY receivers DESC LIMIT 50;
```

⚠ **Check the column shapes before trusting that** (`\d code_symbols`): it is written from memory of
`code_symbols_actions.go:598`'s `"(" + Receiver.Type + ")." + Name` format, not from the live schema,
and the `kind` vocabulary is asserted, not verified. Whoever picks this up should run `\d` first —
this estate has a standing landmine about exactly that.

## 7. Relations

- `bugs_closed/261` — where this was first recorded (§8 follow-up 1) and where the *reader* half was
  fixed. **The one to read first.**
- `bugs_open/267` — increased its reach and put the two spellings in one bundle; §7c records the
  decision not to fold this in and why. Commit `139dcc3ca`.
- `bugs_open/093` — the indexed case of the "one call site guarded, the sibling unpatched" shape.
- `016b` §9 — the transferable pattern.
- Council round `ac23f2f7-9230-403c-8f20-4e18623c1849` — the `bug_historian` objection that produced
  this file (round 1), and the four-seat reuse objection that extracted `CanonicalSymbolName` out of
  it (round 2).
- **The `architecture` seat's wider reading, recorded here because it asked to be visible to whoever
  schedules RFC work:** this is *at minimum the fourth* bug against one underlying mechanism — an
  ad-hoc grammar for `path:Symbol` handles and receiver-qualified method spelling, spread across
  `internal/analysis` and `code_symbols_actions.go`. The four: `bugs_closed/189` (the split
  hand-rolled three times), `bugs_closed/261` (the reader could not parse what the writer produced),
  `bugs_open/267` (whose own fix illuminated it a fourth time), and this file. Its verdict was
  `ARCHITECTURE_SIGNAL: insufficient` for the point fix — correctly — with the debt noted anyway:
  *"each of these bugs is being closed as an isolated point fix rather than as evidence that the
  handle grammar needs ONE authoritative implementation the whole estate calls … the debt is not
  created by it, only illuminated by it a fourth time."* **`CanonicalSymbolName` is the first
  instalment of that authoritative implementation, not a substitute for the RFC.**

## 6b. MEASURED 2026-08-13 — 17 collision groups, 48 methods, and one of them is the diagnosis loop's own file

**The control ran first, because without it the whole measurement is unreadable.** The query strips a
`(Recv).` prefix to get a bare name; if `code_symbols` did not actually store the parenthesised
spelling, the regex would strip nothing and every answer would be wrong in a way that still looks
like an answer:

| | count |
|---|---|
| `kind='method'` rows total | **1,175** |
| …stored parenthesised (`(%).%`) | **1,175** |
| …stored unparenthesised | **0** |

So the spelling assumption holds exactly, and the strip is doing real work.

**The population where a bare handle can return the wrong body:**

| | count |
|---|---|
| collision groups (same file, same bare name, ≥2 receivers) | **17** |
| methods inside them | **48** of 1,175 = **4.1%** |

```sql
SELECT count(*) AS colliding_name_groups, sum(n) AS methods_affected
FROM (SELECT count(*) AS n FROM code_symbols WHERE kind='method'
      GROUP BY repo, path, regexp_replace(symbol,'^\(.*\)\.','')
      HAVING count(*) > 1) x;
```

⚠ **4.1% is the FLOOR of the harm, not the rate of it.** For a group of `n` receivers, a bare handle
resolves to the first and is wrong for the other `n−1`. The worst groups here are **six-way**, so a
bare handle in them is wrong **5 times in 6**:

| file | bare name | receivers |
|---|---|---|
| `discovery_checks/check_integrity.go` | `Name` | **6** |
| `discovery_checks/check_integrity.go` | `Run` | **6** |
| `discovery_checks/check_news_feed.go` | `Run` | 5 |
| `discovery_checks/check_news_feed.go` | `Name` | 5 |
| **`pkg/diagnose/loop.go`** | **`String`** | **2** — `(Outcome).String` / `(Tier).String` |
| `platform/errors/errors.go` | `Error` | 2 — `(*AgentError).Error` / `(*DomainError).Error` |
| `platform/orchestration/actions/query_agent_definitions_actions.go` | `Next`/`Close`/`Err` | 2 each |

**Read that fifth row twice.** `pkg/diagnose/loop.go` is the diagnosis loop's own source. A diagnosis
*of the diagnosis loop* — which is what `bugs_open/267` and `bugs_closed/261` both were — could be
handed `(Tier).String`'s body while asking about `(Outcome).String`, and nothing in the bundle would
say so. The interface-implementation pattern that produces these groups (`Name`/`Run` across sibling
check types, `Error`/`Unwrap` across error types) is *exactly* the shape a diagnosis reaches for when
it is following a dispatch path.

**What this does NOT say.** It does not say 48 wrong bodies have been served. It says 48 methods are
in the population where the defect can fire, and nothing measures how often the bundle actually
offered one of them — the sibling section's per-file cap means many were never listed at all. The
honest claim is the population and the per-group odds; the incidence is unmeasured and would need a
scan of `diagnosis_artifacts.body` for bare-handle lines, which is a different query and not one this
fix needs.

## 7b. FIXED 2026-08-13 — three halves, each mutation-verified alone

Candidate 1 as filed, calling `analysis.CanonicalSymbolName` rather than re-inlining the grammar
(candidate 2's helper, extracted in `17734b699` at four council seats' request).

| # | what | mutation that proves it load-bearing |
|---|---|---|
| 1 | the rendered handle is canonical — `` `path:(*Beta).Handle` `` not `` `path:Handle` `` | revert to `fn.Name` → all three new tests fail |
| 2 | de-duplication keys on the canonical name too (§2a: a method in scope canonically was listed as its own sibling) | disable the `named[canon]` arm → the §2a test fails, alone |
| 3 | **first-wins exactness** — a BARE scope entry resolves the way `spanOf` does, so only the FIRST same-named method is suppressed and the other is still listed | suppress every same-named method → the exactness test fails, alone |

**Half 3 is the one that is easy to get wrong in the safe-looking direction.** Suppressing every
method sharing the bare name is conservative and it *hides a sibling the model has not seen* — which
inverts the purpose of a section that exists to show what retrieval missed. The fix tracks first-wins
in `fi.Functions` order, which is the same order `spanOf` scans, so the de-duplication is exact
rather than merely cautious.

**The fixture carries a real collision and asserts it before anything else** (§5's requirement): two
types with a `Handle` method, built by running the **real analyser** over a temp checkout rather than
hand-describing spans. The load-bearing assertion is not a string match — it **resolves both offered
handles through `ReadSymbolBody` and requires different bodies**, which is precisely what the bare
spelling could never satisfy.

## 8. Council — APPROVED first round, and the two objections were both worth answering

`e5809ca9-d718-44f6-8d27-6d8cd656dd28`, 2026-08-13. **13 seats approve, 2 advisory objections
(`bug_historian` and `prior_art_librarian`, both medium), none high-severity.** Neither blocks; both are
answered here rather than waved at.

### 8a. `bug_historian`: *"the plan asserts `code_symbols_actions.go:598` is 'independent' but does not establish it is SAFE"*

**A fair hit, and the answer was already sitting in my own control measurement — I ran that census to
validate the blast-radius query and did not notice it doubled as the safety proof for the other
producer.**

`code_symbols_actions.go:594-598` reads:

```go
kind := "func"
name := fn.Name
if fn.Receiver != nil {
    kind = "method"
    name = "(" + fn.Receiver.Type + ")." + fn.Name
}
```

It emits the **canonical** form unconditionally for every method. And that is not just what the source
says — it is what the data says: **1,175 of 1,175 `kind='method'` rows are stored parenthesised, 0 are
not** (§6b's control). So the indexer **cannot** produce the collision-driven wrong answer this fix
removes: it has never written a bare method handle.

**So the honest statement is stronger than "independent": it is independent AND provably correct on this
defect, measured, not inferred.** The residual risk is *drift* — two producers of one grammar can
diverge later — which is a different failure from the *ambiguity* fixed here, and is exactly
`architecture_review/RFC_027`'s question. The seat is right that "independent" alone did not establish
this; it does now.

### 8b. `prior_art_librarian`: *"the landmine on the two grammars was not read"*

**Correct, and it is this lane's own landmine** (added 2026-08-12, footprint
`code_symbols_actions.go ~:593-640` + `spanOf`/`splitReceiver` + `siblingSignatures`). Read now, and it
turns out to prescribe exactly what mutation testing made me do independently:

> *"assert it with a fixture that uses **the spelling the producer really emits** — not the one the
> function already implements. That is how this survived: `symbolbody_test.go` asserted the dotted
> `Type.Method` form, a spelling **no producer has ever written**, so the test and the code were blind in
> the same direction and agreed with each other."*

That is the same failure I hit in the 267 round: `TestCanonicalSymbolName`'s first draft asserted only a
round trip through `splitReceiver`, which accepts *both* spellings, so it passed on the wrong one. The
mutation caught it and the fix was to anchor on **literal handles quoting what the producer emits**. Two
independent routes to one rule — worth noting because the landmine would have got me there for free had
I grepped it by symbol rather than by the file I was editing.

Also from the same entry, checked against this fix: *"the trap inside the fix — widening the accepted
SPELLING must not widen the MATCH."* This change moves in the opposite direction (it **narrows what is
EMITTED**, and touches no matching rule), so that trap is not in play.

Its second point — whether `CanonicalSymbolName`, `spanOf` and `splitReceiver` exist as the rationale
assumes — verified: `symbolbody.go:164`, `:262`, `:323`, with `CanonicalSymbolName` added in
`17734b699`.

### 8c. `debug_historian` (low): no stated pod-verification step

Now stated, and with the recipe corrected by this lane's own 2026-08-13 experience — see §9. The seat
cited the pre-2026-08-11 `strings`-the-binary lore; the current method is the build stamp plus an
ancestry test, **with the log-range precheck first**, because on `agent-chassis` the stamp is a startup
line and the recipe is TIME-LIMITED rather than inoperative (`bugs_closed/267` §9).

## 9. Verification when it rolls — and it needs a COLLISION file or it proves nothing

**① Is the code live?** Immediately after the roll (the window is minutes, not hours):

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system logs $POD --tail=100000 | head -1              # PRECHECK: a startup line?
kubectl -n ai-persona-system logs $POD --tail=100000 | grep -m1 'build provenance'
git merge-base --is-ancestor a3fee59b8 <the git_commit from that line> && echo "269 IS IN THE BINARY"
```
Plus a control: any descendant of the stamp must read ABSENT.

**② Has the behaviour changed?** A bare method handle should stop appearing in NEW bundles:

```sql
SELECT count(*) FILTER (WHERE body ~ '- `[^`]+\.go:[A-Z][A-Za-z0-9_]*` — `func \(') AS bare_method_handles,
       count(*)                                                                       AS bundles_in_window
FROM diagnosis_artifacts WHERE kind='bundle' AND created_at > '<roll>';
```

⚠ **`bare_method_handles = 0` is only evidence if `bundles_in_window > 0`.** The demand control is not
optional — `bugs_closed/267` §9 records this exact zero being unreadable without it on 2026-08-13.

⚠ **And the bundle must have scoped a COLLISION file** for the fix's point to be exercised at all;
§6b names them (`discovery_checks/check_integrity.go` is the richest, six-way). A clean result from a
bundle over a collision-free file demonstrates nothing — which is precisely how this defect survived
`bugs_closed/261`'s fix.

## 10. Relations

- `bugs_closed/261` — where this was first recorded (§8 follow-up 1) and where the READER half was fixed.
- **`LANDMINES.md`, "The code index and the body reader are TWO grammars for one `path:Symbol` handle"** —
  this lane's own entry, footprinted on both producers. Cited at `prior_art_librarian`'s objection; its
  fixture rule is the one mutation testing rediscovered here.
- `bugs_closed/267` — increased this defect's reach, and §9 there is where the deploy-proof recipe was
  learned.
- `architecture_review/RFC_027` — the open owner question. This fix narrows it from three producers of
  the spelling to two; §8a establishes the remaining one is currently correct, so the RFC is about drift,
  not a live defect.
