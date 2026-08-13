# 269 — `siblingSignatures` renders methods BARE, so one bundle shows two spellings for one symbol and the bare one silently resolves to the wrong body

**Filed:** 2026-08-12, `silent_hero_logo_readers` lane, **at the council's direction**
(round `ac23f2f7-9230-403c-8f20-4e18623c1849`, `bug_historian` seat, reviewing `bugs_open/267`).
**Status:** OPEN, not started. No code written.

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

## 6. Blast radius — [UNMEASURED], and here is the query that would settle it

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
