# RFC_027 — the `path:Symbol` handle grammar has no single owner, and this is its fourth bug

**Raised** 2026-08-13 by the `silent_hero_logo_readers` lane, out of the council round on
`bugs_open/267_…_bundle_advises_a_whole_file_reread_that_can_never_fit`
(submission `ac23f2f7-9230-403c-8f20-4e18623c1849`, **APPROVED** on round 3 with one advisory
objection). **The `architecture` seat raised it in rounds 2 AND 3, both times asking for exactly
this file** rather than a `bugs_open/` cross-reference — *"worth an explicit RFC-scheduling flag"*.
`reuse_agent`, `constitution` and `tooling_provenance` objected on the same mechanism in round 2.

**Status: OPEN — needs an owner ruling on whether this is worth a consolidation, and if so who owns
it.** Nothing here blocks anything. The 267 fix is committed (`139dcc3ca`, `17734b699`) and approved
on its own terms; the seat was explicit that *"the debt is not created by it, only illuminated by it
a fourth time."*

---

## 1. What the thing IS, before any argument about it

The diagnosis loop talks about code using a short text handle: a file path, a colon, and a symbol
name. `platform/orchestration/coordinator.go:storeActionResult`. That is a **handle** — it is how a
scope entry is written, how the code index stores a row, how the bundle labels a body it renders,
and how an LLM asks for the next thing it wants to read.

For a **method** the handle needs the receiver too, because two types in one file can share a method
name. The estate's spelling for that is `(*SagaCoordinator).applyResponseToState` — parentheses
outside, pointer star inside.

So the grammar has two halves. Something must **write** a handle, and something must **read** one
back. They must agree exactly, or a handle resolves to nothing, or — worse — to the wrong body.

## 2. The rule this estate already applies to grammars like it

`internal/analysis`'s own file header states it, having been made to learn it twice:

> *"one function keeps owning the convention"*

`SplitSymbol` was exported after a third hand-rolled copy of the split had drifted (`i > 0` against
`i >= 0`). `SliceLines` was exported so an indexer writing a body and a reader slicing one could not
disagree byte-for-byte. Both were collapsed onto a single owner **after** a defect, not before.

## 3. How the current case measures against it

**The reader half now has an owner. The writer half still does not.**

| | who | state |
|---|---|---|
| **parse** a handle's path/name | `analysis.SplitSymbol` | one owner (collapsed, `bugs_closed/189`) |
| **parse** a receiver-qualified name | `analysis.splitReceiver` | one owner (taught both spellings, `bugs_closed/261`) |
| **slice** a body from a span | `analysis.SliceLines` | one owner (exported for exactly this) |
| **find** a file's FileInfo | `analysis.FindFile` | one owner (exported 2026-08-12, this round) |
| **write** a canonical method handle | `analysis.CanonicalSymbolName` **and** `code_symbols_actions.go:598` | **TWO producers** |

`CanonicalSymbolName` was extracted during this round — four seats objected that filing it was the
wrong call when the extraction was two lines. It is the inverse of `splitReceiver`, which had only
ever had one half. **`code_symbols_actions.go:598` was deliberately not converted**: it sits in the
live `code_symbols` row-writing path and deserves review on its own terms rather than riding a bug
patch. `bugs_open/269` names it as the call site that fix should use.

**Verified, not asserted** (2026-08-13, `grep -rn 'Receiver\.Type|Receiver != nil' --include=*.go
internal/ platform/ pkg/ cmd/`, excluding tests): exactly **three** sites touch `Receiver.Type`.
Two are the handle producers above. The third, `analysis/analyse.go:377-382`, builds the rendered
**signature** (`func (r *T) Name(...)`) — a different artefact with a different grammar, and worth
naming here so a future consolidation does not wrongly sweep it in.

## 4. The case for a ruling: four bugs, one mechanism

| bug | what broke | which half |
|---|---|---|
| `bugs_closed/189` | the path/name split hand-rolled three times, copies disagreed on a leading colon | parse |
| `bugs_closed/261` | the reader could not parse the spelling its own index writes — **301 unreadable bodies across 44 diagnosis runs**, not one a genuinely absent symbol | write ↔ parse disagreement |
| `bugs_open/267` | over-budget advice; its fix was about to add a **third** copy of the writer | write |
| `bugs_open/269` | `siblingSignatures` writes methods bare, so a bundle offers an ambiguous handle that silently resolves to the **wrong** body | write |

The `architecture` seat's reading, verbatim, because it is the argument:

> *"each of these bugs is being closed as an isolated point fix rather than as evidence that the
> handle grammar needs ONE authoritative implementation the whole estate calls. The cost of not
> changing: every new consumer of a symbol handle (`SymbolSizes` today, whatever reads
> `code_symbols.body` next) is one more chance to reintroduce the exact collision defect 261 fixed."*

**What makes this worth an owner's minute rather than a fifth bug file:** 261's cost was measurable
and large (301 lost bodies) and its failure mode was *silent* — the loop reported "symbol not found"
and every reader took that at face value, including a landmine-verifier that answered "no longer
resolves as a standalone symbol". 269's is worse in kind: it returns a *plausible wrong body* rather
than an error.

## 5. What is actually being asked

Not a redesign. The question is narrow and has a cheap answer either way:

1. **Is the writer half worth collapsing onto `CanonicalSymbolName`?** That is one call site
   (`code_symbols_actions.go:598`) plus one missing call (`siblingSignatures`, which is
   `bugs_open/269`'s actual defect). Behaviour-identical for the first; a fix for the second.
2. **If so, does it need this track at all, or is it two ordinary council rounds?** The seat's own
   signal was `ARCHITECTURE_SIGNAL: insufficient` for the point fix — it wanted the *pattern*
   visible, not the fix escalated. A defensible ruling is "no RFC needed, just do it, and stop
   filing it."
3. **Is there a shape rule worth writing down?** The candidate: *when a grammar has a parser in a
   shared package, the builder belongs beside it.* Three of the four bugs above are one half of a
   pair living somewhere the other half could not see.

## 6. What this RFC does NOT claim

- **It does not claim the four bugs are the same defect.** They are four defects in one mechanism.
  A ruling that says "four bugs in a year on a grammar this small is acceptable" is a legitimate
  outcome and would close this file.
- **It does not claim a consumer is broken today by having two producers.** The two producers
  currently agree, and `TestCanonicalSymbolName` pins the helper against
  `code_symbols_actions.go:598`'s format with literal handles. The risk is drift, which is
  prospective. **`bugs_open/269` is a live defect; this file is not.**
- **[UNMEASURED] how many methods sit in a file with a name collision** — the population where 269
  can actually return the wrong body. `bugs_open/269` §6 carries the query and flags that its column
  shapes are written from memory of `code_symbols_actions.go:598` and want a `\d code_symbols` first.
  Whoever takes this should measure it; the ruling may well turn on the number.
- **Council precedent checked** (2026-08-13, `diagnosis_artifacts` where `kind='council_report'` and
  the body names `diagnose_assemble_bundle`): 18 prior rounds, one **REJECTED** (`fca1071b`,
  2026-08-10). That veto is about dispatch resolution across four packages and merely *names* this
  file as a sibling missing a `deleted_at`/`is_snapshot` guard — which this file already carries. **No
  prior ruling conflicts with the position above.** Recorded because two seats asked for this lookup
  twice and it had not been run.

## 7. Relations

- `bugs_open/269` — the live defect, and the concrete work item. **Read this first if you want
  something to do.**
- `bugs_closed/261` — the measured cost of the two halves disagreeing (301 bodies, 44 runs).
- `bugs_closed/189` — the first collapse, and the precedent for the rule.
- `bugs_open/267` — the round that surfaced this; register **DIAG-043**.
- Council round `ac23f2f7-9230-403c-8f20-4e18623c1849` — rounds 2 and 3 of the `architecture`
  objection, plus `reuse_agent`/`constitution`/`tooling_provenance` on the same mechanism.
