# PLAN 2026-07-31 — `bugs_open/145`: give `ReadSymbolBody` a boundary of its own

**Workstream opened 2026-07-31 (session "bugfix 12").** Target: `bugs_open/145` —
`ReadSymbolBody`'s whole-file fallback reads ANY path off disk, with no kind check.
Taken because it was filed **OPEN, UNOWNED**, no doc anywhere references
`bugs_open/145`, and `internal/analysis/` + `diagnose_assemble_bundle_action.go`
are both clean in the tree and last touched 2026-07-29 by the layer-1b commit that
*surfaced* the bug rather than one working it.

## Is the bug still valid? YES — verified first-hand, 2026-07-31

`internal/analysis/symbolbody.go:44-59` is byte-for-byte what the filing quotes.
`os.ReadFile` at :51 runs **before** any membership check; the `namePart == ""`
branch at :57-59 returns the whole file; the only `Output` consultation
(`findFile`, :61) sits **after** both.

## THE FILING'S REACHABILITY CLAIM IS WRONG, AND THE BUG IS WORSE THAN FILED

145 says the bare-path branch is *"currently unreachable by accident of the writer"*,
reasoning only about `scopeFromCodeResults` (`code_symbols.symbol` is `NOT NULL`).

That analysis stops one producer short. **The scope fallback chain's FIRST and
highest-priority source is `route.scope`** (`diagnose_assemble_bundle_action.go:139-140`),
which is `diagnose.EncodeScope(decision.NextScope)`
(`diagnose_route_action.go:226`), which is built from **`Verdict.NextScope` — the
LLM verdicter's own `next_scope` JSON array** (`pkg/diagnose/verdict_wire.go:29`).
`namedScope` (`pkg/diagnose/loop.go:398-418`) only trims and dedupes those strings;
`nextScope` (:345-392) adds call-graph neighbours to them. **Nothing anywhere
validates that an entry is a path the analyser knows, or a path at all.**

And it is not merely reachable — **the bundle text ASKS for it.** `siblingSignatures`
emits, verbatim (`diagnose_assemble_bundle_action.go:597`):

> `- _(+%d more in this file — put the bare file path in next_scope to see it whole)_`

So the whole-file branch is a **live, advertised capability driven by the least
trustworthy producer in the loop**, not dead code reachable only by accident.

Two consequences that change the fix:

1. **Fix candidate 1 in the filing (drop the branch / move it to an explicit
   `ReadWholeFile`) would break a documented feature the bundle instructs the model
   to use.** It must not be taken as written.
2. The §7D resolver does not save us. A path-shaped entry absent from
   `knownScopeIdentities` is classified FUZZY, sent to embedding search, and
   **fails open to its original string** when nothing clears
   `resolver_min_similarity` (`diagnose_route_action.go:512-600`). Enrichment, not
   a boundary — its own contract is "no worse than not resolving".

## The decisive piece of prior art: the ORIGINAL caller already had this guard

`ReadSymbolBody` was extracted from contextkit's `cmd/assembler`. That caller —
`docs019_…/go_files/contextkit/cmd/assembler/main.go:178-200` — resolves the path
against the analysis **first** and refuses before reading:

```go
fi, ok := byPath[path]
if !ok {
    w("> scope %q: path not found in analysis — skipped\n\n", sc)
    continue
}
...
if sym == "" {                              // whole file
    ... enumerate fi.Functions / fi.Types ...
    src, err := analysis.ReadSymbolBody(*root, *an, path)
```

**So the invariant existed; the chassis port dropped it.** `diagnose_assemble_bundle`
calls `ReadSymbolBody` straight off the scope list with no `byPath` check. This is
not a missing feature — it is a **precondition lost in a port**, which is exactly
the failure the architecture seat generalised to "reachable by any future scope-entry
producer".

The register already states the intent, too — CTXK-002: *"Intentionally Go-only."*
The code does not enforce what its own register entry claims.

## Decision: fix candidate 2, in its strongest form — Output membership, not extension

**Make the analyser `Output` the authority for what is readable, for BOTH branches,
and check it BEFORE touching disk.** Move the existing `findFile` call above
`os.ReadFile` and make it unconditional.

Why this and not the filing's three candidates as written:

| | |
|---|---|
| **vs. candidate 1** (drop the branch) | Breaks the advertised `next_scope` bare-path feature. Rejected on evidence. |
| **vs. candidate 2 as filed** (bound by `.go` extension) | An extension allow-list is a second copy of the analyser's own inclusion rule, and it drifts. `Output` membership *is* that rule — Go-only, and additionally minus `vendor/`, `testdata/`, `_test.go`, download-duplicates and `excluded()` paths (`analyse.go:80-99`). A `.go` check would still admit a vendored or excluded file. |
| **vs. candidate 3** (size bound) | The filing calls it weakest and this repo calls "the operator must remember `maxBodyChars`" a defect. No new knob. |

What it makes unrepresentable: **the function can no longer name a file the analyser
did not parse.** No allow-list to maintain; if the analyser ever learns another
language the boundary widens with it, automatically and correctly. Directory
traversal (`../../etc/passwd`) is closed as a side effect — it cannot be in `Output`.

It is also a **single choke point**: every path in the function goes through
`findFile`, so there is no second door to guard.

## Blast radius, named (owner ruling 2026-07-29 #3 — tell the consumers)

- **One live consumer**: `diagnose_assemble_bundle_action.go:201`. Verified — `go list
  ./...` yields **0** contextkit packages; the archived copy has its own `go.mod`.
- **The archived `cmd/assembler`** already enforces the same precondition in its
  caller, so if it is ever revived the change is a **no-op** for it. This answers the
  filing's requested negative control.
- **`code_symbols_actions.go:401`** (`flattenSymbols`) calls `SliceLines` directly
  while iterating `out.Files` — already bounded by `Output` by construction. No
  sibling defect.

## What the guarantee change actually is (RFC trigger test, ruling 2026-07-29 #1)

`ReadSymbolBody`'s guarantee **narrows**: from "any path resolvable under `root`" to
"a file the analyser parsed". That is a restriction that removes a capability nothing
needs (arbitrary repo file content has its own action,
`diagnose_read_repo_files_action.go`, with its own `max_file_bytes`). It does not make
a shared mechanism able to do something it previously could not, which is the stated
RFC trigger. **Council gate, not RFC** — submitted alongside the commit.

## Phases

1. ~~Confirm still valid; establish real reachability; find prior art.~~ **DONE.**
2. The edit: reorder + unconditional `findFile` in `ReadSymbolBody`.
3. Tests that assert the *mechanism fired*, not merely that things are quiet.
4. Correct the two stale comments the filing asked whoever took this to fix.
5. Council gate submission; register CTXK-002 in the same commit as the seam;
   LANDMINES entry + sync.
6. Build, roll, pod-verify, then close 145 → `bugs_closed/`.
