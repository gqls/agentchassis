# 189 — `siblingSignatures` hand-rolls the `path:Symbol` split, and the two copies now disagree on the leading-colon edge

**Filed 2026-08-04 by the `bugfix_163_symbol_lookup` lane, at the council gate's request.**
**Status: OPEN, UNOWNED. LATENT — no live failure is claimed; see § Measured.**

## Why this exists

Filed at the council's explicit direction, the same way `bugs_closed/164` was. Reviewing
163's fix (corr `da1d3a40-8ec1-4ea1-9c65-8c585ec2d013`, **APPROVED**), the `bug_historian`
seat objected at MEDIUM on exactly one ground:

> "Plan explicitly leaves a THIRD duplicate of the split-symbol logic unfixed
> (`diagnose_assemble_bundle_action.go:639-644`, differing `i>0` vs `i>=0` edge) while
> exporting/fixing the other two call sites. This is the documented shape *'One call site of
> a shared judgement gets the rigorous fix; the sibling stays heuristic'* (016b §9).
> Disclosed and tracked in CTXK-002, which mitigates but does not remove the exposure — that
> third site can still misparse a `path:Symbol` query the same way 163 did."

The seat is right that a register entry is not a fix. 163's lane deliberately did not widen
its diff after approval; this is the filing that discharges the objection instead.

## The defect

`platform/orchestration/actions/diagnose_assemble_bundle_action.go:640-643`, inside
`siblingSignatures`:

```go
for _, s := range scope {
    path, name := s, ""
    if i := strings.LastIndex(s, ":"); i > 0 {
        path, name = s[:i], s[i+1:]
    }
```

This is the **third** hand-rolled copy of the `path:Symbol` convention. The other two are now
one: `internal/analysis/SplitSymbol` (`symbolbody.go:105-114`), exported 2026-08-04 by 163's
fix precisely so that one function owns the convention — the `SliceLines` precedent recorded
in that file's own header, where `prior_art_librarian` once stopped a council extracting a
helper that already existed.

**The two copies disagree on one input.** `SplitSymbol` splits on `i >= 0`; this copy on
`i > 0`. For a scope entry `":Foo"`:

| parser | path | name |
|---|---|---|
| `SplitSymbol` | `""` | `"Foo"` |
| the inline copy | `":Foo"` | `""` — treated as a WHOLE-FILE scope entry |

## Measured — and this is why it is filed LATENT, not "biting"

```sql
-- scope entries fleet-wide that would hit the divergent branch
SELECT count(*) FROM diagnosis_artifacts WHERE kind='bundle' AND body LIKE '%:%';
```
`[UNMEASURED]` — the honest position. A leading-colon scope entry is not a form any producer
composes today: `scopeFromCodeResults` and `resolveScopeEntries` both build `path + ":" +
symbol` from non-empty paths. **So the divergence is currently unreachable, and this bug is
about the DUPLICATE, not about a live wrong answer.**

What makes it worth a ticket anyway is the producer set: `route.scope` is built from the LLM
verdict's `next_scope` with **no validation beyond trim-and-dedupe** (`pkg/diagnose/loop.go`),
and the §7D fuzzy resolver **fails open to the original string** by explicit contract
(`LANDMINES.md`, the `ReadSymbolBody` entry). So the least trustworthy producer in the loop
can emit an arbitrary string into this parser, and the two parsers reading the same scope list
will not agree about it.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Call `analysis.SplitSymbol` and decide the leading-colon edge explicitly**, e.g.
   `path, name := analysis.SplitSymbol(s); if path == "" { path, name = s, "" }` to preserve
   today's treatment verbatim while removing the second grammar. **One parser, one edge
   decision, written down.** Cheapest and closes the class.
2. Same, but make a leading-colon entry an explicit skip with a rendered note — arguably more
   correct (`":Foo"` names no file), but it is a behaviour change to bundle rendering and
   should be its own decision, not a side effect of de-duplication.
3. Leave it and rely on CTXK-002's note. **Weakest** — that is exactly what the council
   objected to, and a register entry cannot stop the next port re-copying the grammar.

## How to verify a fix

- Unit: `siblingSignatures` over a scope list containing `a.go:Foo`, `a.go`, and `":Foo"` —
  assert the first two are unchanged **byte-for-byte** against today's output (this function
  feeds bundle rendering, so a formatting drift is a real regression), and that the third
  behaves as candidate 1 or 2 says, deliberately.
- Negative control: a scope entry with **no** colon must still map to a whole-file entry.
- `grep -rn "LastIndex(.*\":\")" --include=*.go platform/ internal/ pkg/` should return the
  ONE remaining site after the fix, not two.

## Related

- **`bugs_closed/163`** — the fix that exported `SplitSymbol` and the council round that
  ordered this filing; its close-out names this duplicate as a stated deferral.
- **`bugs_closed/164`** — the precedent: an adjacent defect found while editing a function
  must be filed, not left "for a reviewer to say so".
- **CTXK-002** (concept register) — records the export and names this site as the remaining
  in-build duplicate.
- **016b §9** — *"One call site of a shared judgement gets the rigorous fix; the sibling stays
  heuristic"*, the pattern the seat matched this to.
- **Filing method, declared** (CLAUDE.md 2026-07-31 ruling): this asserts a **code-local**
  claim — two named parsers, one divergent edge — verified by reading both functions and
  grepping the tree for other copies, not by the `090` loop. The part that is uncertain
  (whether the divergent branch is reachable in practice) is marked `[UNMEASURED]` above
  rather than argued.

---

## CLOSED 2026-08-04 — collapsed onto `analysis.SplitSymbol`; council APPROVED round 1

**Fix commit `a2f54802c`.** Council `89bc06d7-2414-4c03-b79f-d85e5f5d9454` — **APPROVED, round
1**, "1 advisory objection, none high-severity"; 12 reviewers, 5 abstained, 0 unreadable,
`gated_by_truncation: false`. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_189_split_symbol_duplicate/`.

### What shipped — candidate 1's mechanism, with the edge decided as a SKIP

```go
path, name := analysis.SplitSymbol(s)
if path == "" {
    continue
}
```

Not candidate 1 verbatim (`if path == "" { path, name = s, "" }`), and the reason is a finding
this file did not have.

### This file's premise was right, and one notch too weak — the divergence is UNOBSERVABLE

§ Measured argues the divergence is **unreachable** (no producer composes a leading-colon
entry). True, and it is also **unobservable**, which is the stronger fact that changed the fix:

`inScope` is written at `:644-651` and read at exactly **two** sites, `:658` and `:673`, both
keyed by `f.Path` drawn from `out.Files`. So `SplitSymbol`'s parse of `":Foo"` (path `""`) and
this file's inline parse (path `":Foo"`, marked `"*"`) are **both unmatchable keys**. Even the
pathological case — an analysed file literally named `:Foo` — renders identically, because
`:658` reads `ok && !named["*"]` and a `"*"`-marked key excludes the file exactly as an absent
key does.

**That is a licence, not an anticlimax.** With no behaviour to preserve, the edge could be
decided on its merits instead of mimicked: an entry with no path names no file, so it is
skipped — the same judgement `ReadSymbolBody` makes on **the same scope slice**
(`"ReadSymbolBody: empty path in symbol %q"`, `symbolbody.go:53`). The two consumers of one
scope list now agree where they previously disagreed in silence. Candidate 1 verbatim would
have kept that second inconsistency for ever, invisibly, to imitate code being deleted.

### Behaviour preservation — proven, not asserted

`TestSiblingSignatures_ScopeEntryParsing` (six cases, **full-output golden literals**, not
`Contains()` probes — this feeds bundle rendering, so formatting drift is a real regression):

1. written and run **green against the UNCHANGED function** at `3646a4f2d`;
2. collapse applied, the **identical untouched test file** re-run — byte-identical, green;
3. three mutations, each predicted to fail a *named* case before it ran, each did:
   `LastIndex`→`Index` → the last-colon case only; halves swapped → cases 1/5/6; the edge made
   matchable → case 4 (and 5, unpredicted and correct — case 5 carries the edge entry too).

The golden literals therefore describe the OLD code, and the new code reproducing them is the
evidence. Case 5 specifically pins that a no-path entry cannot inflate the fair-share
denominator.

### The advisory objection, CHECKED not banked

`editquality` (MEDIUM) and `bug_historian` (`missing`) both said the same thing: the claim that
`route.scope` carries unvalidated model-authored strings was **cited, not verified**, and the
`ReadSymbolBody` landmine warns about exactly that reasoning shape. Fair. Checked afterwards,
and **CONFIRMED**:

- `pkg/diagnose/loop.go:408-416` (`namedScope`) — the only filtering is `TrimSpace`, a
  non-empty test, and a dedupe map. Nothing inspects shape; `":Foo"` survives verbatim.
- `diagnose_route_action.go` §7D — its own contract: *"the remaining entries **fail-open to
  their prose labels**"*, and §7D is *"no worse than not resolving"*.

So the untrusted producer is real. That **strengthens** the fix: it is precisely why the two
readers of that list must not disagree.

### § How to verify a fix — one correction for whoever re-runs it

This file says the grep *"should return the ONE remaining site after the fix, not two"*. **The
raw grep returns THREE**, and all three are correct:

```
$ grep -rn 'LastIndex(.*":")' --include=*.go platform/ internal/ pkg/
companies_house_fetch_accounts_action.go:536   # iXBRL "ns5:" prefix — different convention
agent_image.go:221                             # docker repo:tag — different convention
internal/analysis/symbolbody.go:130            # THE OWNER
```

"One site" counts *`path:Symbol` convention* sites (2 → 1). Read three raw hits as **pass**, not
failure. Wider spelling census (`Index`/`IndexByte`/`SplitN`/`Split`/`Cut`): **13 sites**, the
large majority unrelated conventions. And the standing caveat: **a grep proves absence only for
the spelling it searches.**

### A `pattern-check.py` rule against a FOURTH copy — calibrated and DECLINED

Measured, not waved off. The 13 colon-split sites are overwhelmingly legitimate different
conventions (docker tags, CSS declarations, aspect ratios `"16:9"`, `base:arg`), so a lexical
rule fires on ordinary good work — the bar that file's own DECLINED "unsupported figure" block
sets. The working precedent, `check_handrolled_shipped_predicate`, only works because it keys on
a **domain literal** (`build_status`); the colon split has no discriminating token, because its
meaning lives in the *provenance of the variable*, which no regex sees. What guards instead:
there is nothing left to copy from, `SplitSymbol`'s header and CTXK-002 name the single owner,
and the `bug_historian` seat that caught copies two and three stays armed — now with an instance
line in `016b` §9 sharpening the match.

### Docs corrected in the fix commit, not later

`CTXK-002` (its "left deliberately" deferral struck through, not deleted, with the false premise
named) and the `SplitSymbol` header (which instructed the next editor to do exactly this).
`016b` §9 gained an **instance** line on the existing entry, not a new entry. No `LANDMINES.md`
entry: after the collapse there is no trap left to warn about. No SUMMARY: the five headings
would reproduce the PLAN.

### What is NOT claimed

The running image at close was **`v1.0.1251`, built minutes BEFORE `a2f54802c`** — so this fix
is **not in a running binary yet**; it ships on the next build, which takes committed HEAD.

Closed anyway, deliberately. The house bar is *"fixed AND live … because the defect is still
reproducible until it ships"*, and that rationale is the test: this defect is a **source-level**
property (one convention, two parsers) whose divergence was **proven unobservable**, so there is
no runtime behaviour to reproduce before the roll or after it. The duplicate dies at commit and
is grep-verifiable at HEAD.

Pod-grep cannot arbitrate it either way — the change adds and removes **no string literal**, and
`SplitSymbol` is already in every binary via `ReadSymbolBody`. Anyone needing to know whether a
given image carries it must use **tag ancestry** (is the commit that set that `IMAGE_TAG` a
descendant of `a2f54802c`), which is `[INFERRED]`, not measured — `bugs_open/153` is why.

### For the `bugfix_163_symbol_lookup` lane

Your deferral is discharged; the `bug_historian` MEDIUM from corr
`da1d3a40-8ec1-4ea1-9c65-8c585ec2d013` is closed. Worth knowing: the premise you deferred on —
"folding it in changes behaviour" — was false, and one read of two lines would have shown it.
That is recorded in `016b` §9 as the transferable half, not as a criticism: *"folding it in would
change behaviour" is a measurement, not a reason.*
