# PLAN — collapse the third `path:Symbol` parser onto `analysis.SplitSymbol`

**Bug:** `bugs_open/189`, slug `siblingsignatures_hand_rolls_the_path_symbol_split`.
**Started 2026-08-04.** Pre-fix HEAD: `3646a4f2d`.

> **Number ambiguity, stated once.** `189` names TWO unrelated open bugs — the other is
> `resolving_a_locked_positional_slot_duplicates_it_on_the_page`. Every reference in this
> lane uses the slug. `scripts/who-owns.py 189` prints this trap itself.

## Scale, stated honestly up front

The code change is five lines in one function. The surrounding work that earns its place is
the golden test (it *is* the proof of behaviour preservation), three documents that become
false the moment the collapse lands, and one council round. Work deliberately **not** done:
a `pattern-check.py` rule (calibrated and declined — §3), a `LANDMINES.md` entry, a
`WRONG_CALLS.md` row, a SUMMARY. A plan that inflated this would be worse than one that says
so.

## Why this bug, and why now

Filed at the council gate's explicit direction. Reviewing `bugs_closed/163`'s fix (corr
`da1d3a40-8ec1-4ea1-9c65-8c585ec2d013`, APPROVED), the `bug_historian` seat objected at
MEDIUM that the plan left a third duplicate of the split-symbol logic unfixed while
exporting and fixing the other two — `016b` §9's *"One call site of a shared judgement gets
the rigorous fix; the sibling stays heuristic"*. 163's lane deliberately did not widen its
diff after approval and recorded the deferral in concept-register CTXK-002 instead. **A
register entry is not a fix**; this lane is what discharges the objection.

## The finding that shaped the fix — checked, and stronger than the bug file claims

The bug file says the divergence is *unreachable* (no producer composes a leading-colon
entry). Reading the function established something stronger: it is **unobservable**.

`inScope` is written at `:644-651` and read at exactly two sites, `:658` and `:673`, both
keyed by `f.Path` drawn from `out.Files`. So `SplitSymbol`'s parse of `":Foo"` (path `""`)
and the inline copy's parse (path `":Foo"`, marked `"*"`) are both unmatchable keys.
**Even the pathological case — an analysed file literally named `:Foo` — renders
identically**, because `:658` reads `ok && !named["*"]`, and a `"*"`-marked key excludes the
file exactly as an absent key does.

Consequence, and it is the load-bearing one: the collapse is a pure de-duplication with
provably zero observable change, so **the leading-colon edge can be decided on its merits
instead of preserved by mimicry.**

## The fix

Candidate 1's mechanism (call `analysis.SplitSymbol`), with the edge decided as an explicit
**skip** rather than the bug file's `if path == "" { path, name = s, "" }`:

```go
path, name := analysis.SplitSymbol(s)
if path == "" {
    continue
}
```

**Why skip, by "what makes the bad state unrepresentable":**

- The bad state was *two grammars over one scope list*. After this there is no second
  grammar to drift from.
- The edge is decided on its own grounds: `ReadSymbolBody` — **the other consumer of the
  same `scope` slice**, 600 lines up — rejects the identical shape with
  `"ReadSymbolBody: empty path in symbol %q"` (`symbolbody.go:53`). After the fix the two
  consumers *agree* about `":Foo"`. The bug file's candidate 1 verbatim would keep them
  disagreeing (whole-file here, error there) — merely invisibly — by preserving an edge rule
  whose only purpose is to imitate the code being deleted. That is the tension the bug file
  itself flags.
- It is still behaviour-preserving, **proven not asserted** (§2), so it is not candidate 2:
  no rendered note, no observable change to bundle rendering.

**Fallback if the council objects:** candidate 1 verbatim is extensionally equal to the
deleted code for every input (case analysis: `i<0` and `i==0` both yield `(s,"")` under
both; `i>0` splits identically and `s[:i]` is non-empty so the override never fires).
Acceptable second choice. Candidate 3 (leave it) is what the council already objected to.

## §2 — how behaviour preservation is PROVEN

Sequencing is the method; the assertions alone would not be evidence.

1. **Baseline captured first.** `TestSiblingSignatures_ScopeEntryParsing` was written and run
   **green against the unchanged function** at `3646a4f2d`. The golden literals therefore
   describe the OLD code.
2. **Collapse applied**, and the *identical, untouched* test file re-run green. The new code
   reproducing the old code's full output byte-for-byte is the evidence.
3. **Mutations, because a passing test proves nothing until it has been seen to fail.**
   Three, each predicted to fail a *named* case before it was run. Results in NOTES.

Full-output golden literals, not `Contains()` probes: this feeds bundle rendering, so a
formatting drift is a real regression.

**Stated limitation, also carried in a test comment:** no output-level test can distinguish
the old `i > 0` treatment from `SplitSymbol`+skip — that unobservability is the §"finding"
above. The test pins the *decision* (case 4) and the *shared grammar* (cases 1, 5, 6), which
is everything observable.

## §3 — the framework-level check: CALIBRATED, then DECLINED

The standing instruction is to prefer a fix applicable to the framework over the individual
case, so a `pattern-check.py` rule against a fourth copy was assessed properly rather than
waved off — and it does not clear that file's own bar.

- The bug's grep shape (`LastIndex(.*":")`) over `platform/ internal/ pkg/ cmd/` returns
  **4 sites**: the defect, the owner (`symbolbody.go`), and two legitimate *different*
  conventions — `agent_image.go:221` (docker `repo:tag`, with registry-port logic) and
  `companies_house_fetch_accounts_action.go:536` (iXBRL `ns5:` prefix stripping).
- Widening to the other spellings (`Index`, `IndexByte`, `SplitN`, `Split`, `Cut`) gives
  **13 colon-split sites, the large majority unrelated**: CSS declarations, font stacks,
  `page:variant` names, aspect ratios `"16:9"`, `base:arg` in `queryresolve.go:158`.

A check keyed on the lexical shape fires overwhelmingly on ordinary good work — exactly the
DECLINED bar recorded in `pattern-check.py`'s own "unsupported figure" block ("anything that
fired on ordinary work was cut"). The precedent that *does* work,
`check_handrolled_shipped_predicate`, keys on a **domain literal** (`build_status`) that
uniquely names the shared predicate. The colon split has no such token: its meaning lives in
the **provenance of the string variable** (a scope entry vs a docker ref), which no regex can
see. An allow-list version would grow with every future legitimate colon split — a check
that taxes ordinary work to guard a convention with, post-fix, zero duplicates.

**What guards against a fourth copy instead, honestly:** (a) there is nothing left to copy
*from* — the likeliest fourth copy was a cut-paste of the third; (b) `SplitSymbol`'s header
and CTXK-002 name the single owner (discoverability, not enforcement — *a doc comment is not
a control*); (c) the mechanism that actually caught copies two and three — the council's
`bug_historian` seat matching `016b` §9 — stays armed, and the instance line this lane
appends sharpens that match; (d) the close-out records both census greps so a future audit
re-runs them in seconds, with the caveat that **a grep proves absence only for the spelling
it searches.**

## §4 — docs this fix carries, and the ones it must not

**In the fix commit** (the register must not assert a falsehood for even a window —
`bugs_open/161`'s lesson, and the 2026-07-29 ruling's condition (2) by parity):

- **CTXK-002** — append to the "SEAM WIDENED 2026-08-03" paragraph, do not erase its history.
- **`symbolbody.go` header** — its "left alone deliberately … whoever next edits
  siblingSignatures should collapse it" becomes false at this commit.
- **`016b` §9** — one *instance* line on the existing entry (`bugs_closed/164`'s precedent).
  A **new** §9 entry is NOT warranted: no new transferable pattern, only an instance of an
  indexed one.

**NOT warranted, with reasons.** `LANDMINES.md`: the entry test is *"would a symptom-free
session touching this get it wrong without the entry?"* — after the collapse there is no
trap (one owner, explicit edge, accurate register), and an entry naming the new edit would
recreate 163's self-contamination finding for zero protection. `WRONG_CALLS.md`: the bar is a
claim written down that turned out false; add a row only if one occurs (see NOTES — one near
miss was caught *before* it was published, which is NOTES material, not a wrong call).
**SUMMARY: expected never** — the five-headings test would reproduce this PLAN, and rarity is
part of the design.

## §5 — council gate

**Not architecture-scope.** The 2026-07-29 ruling narrows that to changes altering what a
shared mechanism **GUARANTEES**. This deletes a duplicate of an existing exported helper with
provably preserved observable behaviour; `SplitSymbol`'s contract is untouched; no new
namespace, key or contract; no config half to order against a binary, so the ordering
exemption is irrelevant. Condition (2) — register in the same commit — is honoured anyway.
Consumers whose guarantee changes: none. The 163 lane, which filed this at the council's
direction, is told by the close-out referencing their round.

**Submitted 2026-08-04, `SUBMISSION_CORR=89bc06d7-2414-4c03-b79f-d85e5f5d9454`**, before the
commit. Commit carries `Council-Submitted:`, never `Council-Reviewed:` on an unread verdict.

## §6 — verification, and what "closed" means for a latent defect

- `gofmt -l` on both touched files, `go build ./...`, `go test` on both packages.
- **The bug's own census grep, with its reading corrected.** It says *"should return the ONE
  remaining site after the fix, not two"*. Raw hits are **3**, not 1: `symbolbody.go` (the
  owner) plus two unrelated conventions. The bug counts *convention* sites (2 → 1). The
  close-out must say this, or a future verifier reads 3 as a failure.
- **Closure.** The defect is a **source-level** property — a divergent duplicate — and it
  dies at commit, grep-verifiable at HEAD. The behavioural divergence was never reachable and
  is proven unobservable, so there is no live failure for a roll to extinguish. Pod-grep
  cannot prove this one live (it adds and removes no string literal; `SplitSymbol` is already
  in the binary via `ReadSymbolBody`), so "live" is established by **tag ancestry** — whether
  the commit that set the deployed `IMAGE_TAG` is a descendant of the fix — recorded as
  `[INFERRED]` with `bugs_open/153`'s provenance limit named, not silently skipped.
