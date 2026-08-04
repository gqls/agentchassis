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
