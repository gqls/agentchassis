# NOTES — `bugs_open/145` symbolbody boundary

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-31 — session "bugfix 12", opening

### Picking the bug

`scripts/who-owns.py` returned **`VERDICT: OWNED or recently active.` for every open
bug** — it counts any commit mentioning the number, and on this tree that is all of
them. **[MISSTEP]** I ran it 32 times before noticing the verdict line carried no
signal at that granularity. What actually discriminated was the pair
*(last commit touching the bug FILE, does it name an owning workstream)*:

```bash
for f in bugs_open/*.md; do n=$(basename "$f" | grep -oE '^[0-9]+')
  git log -1 --format='%ad' --date=short -- "$f"
  python3 scripts/who-owns.py "$n" | sed -n '/likely OWNING/,/^$/p' | grep -c 'none identified'
done
```

Eleven bugs had no named workstream. Of those, 145 was the only one whose own file says
**"OPEN, UNOWNED"** in as many words, and `grep -rn "bugs_open/145" docs/` returned
**nothing** — no thread has written about it anywhere. Both files it touches were clean
in the tree.

Rejected on ownership evidence, worth recording so nobody re-walks it:
- **143** (`derive_card_asset` lock ordering) — filed by the `bugfix_131_og_card` lane,
  and *while I was working* another session's uncommitted `assetLock` helper appeared in
  `platform/orchestration/actions/`, which is 143's stated fix shape. Actively owned.
- **091, 097, 111** — `bugs-sweep` lane (MEMORY names 091 as its next item; 111's own
  file carries a bugs-sweep status block and blocks on 117/118).
- **080/081** — need an owner call, per their files.

### The filing's reachability claim is wrong, and I nearly inherited it

145 says the bare-path branch is *"currently unreachable by accident of the writer"*,
having followed `scopeFromCodeResults` (whose `code_symbols.symbol` is `NOT NULL`).

**[VERIFIED by reading the whole chain]** That is the wrong producer. `route.scope` is
the FIRST entry in the bundle's scope fallback chain
(`diagnose_assemble_bundle_action.go:139-140`) → `diagnose.EncodeScope(decision.NextScope)`
(`diagnose_route_action.go:226`) → `Verdict.NextScope`, the **LLM verdicter's own JSON**
(`pkg/diagnose/verdict_wire.go:29`). `namedScope` (`pkg/diagnose/loop.go:409-415`) trims
and dedupes; that is the entire validation.

And the bundle **asks** for it — `diagnose_assemble_bundle_action.go:597`:

> `- _(+%d more in this file — put the bare file path in next_scope to see it whole)_`

So the branch is a live, advertised capability driven by the least trustworthy producer
in the loop. Two things follow, and the second is the one that mattered:

1. **145's fix candidate 1 must NOT be taken as written.** "Drop the branch or move it to
   an explicit `ReadWholeFile`" would break a documented capability. I had started
   towards candidate 1 because the filing ranks it first under "makes the bad state
   unrepresentable" — reading `:597` is what stopped me. **[MISSTEP AVOIDED, barely:
   I was ordering by the filing's own ranking rather than by what the code does.]**
2. The §7D resolver is not a filter. An unresolvable path-shaped entry **fails open to
   its original string** by explicit contract (`diagnose_route_action.go:512-600`).
   It looks like the safety net; it is enrichment.

### The piece of prior art that decided the design

`ReadSymbolBody` was extracted from contextkit's `cmd/assembler`, and **that caller
already has this guard** (`docs019_…/contextkit/cmd/assembler/main.go:178-200`):

```go
fi, ok := byPath[path]
if !ok { w("> scope %q: path not found in analysis — skipped\n\n", sc); continue }
```

So the invariant existed and **the chassis port dropped it**. That reframes the fix from
"add a check" to "restore a precondition, and put it where a port cannot lose it" —
which is exactly what the architecture seat meant by *"a generic hazard reachable by any
future scope-entry producer"*. It also answers the negative control 145 asked for: if
`cmd/assembler` is ever revived, the change is a **no-op** for it.

### Why Output membership beat the filing's `.go` extension test

`Output` **is** the analyser's inclusion rule (`analyse.go:80-99`): Go, non-test, minus
`vendor/`, `testdata/`, download-duplicates, `excluded()`. An extension test would be a
second, drifting copy of that rule — and would still admit a vendored or excluded file.
**[MEASURED, not argued]** the regression test includes `f_test.go`,
`vendor/dep/dep.go` and `testdata/sample.go` precisely because all three are `.go` and
all three leaked pre-fix.

### The severity is worse than filed, and this one is measured

145 scopes the leak to files committed in the repo. **[MEASURED 2026-07-31]** With a real
file placed outside the checkout, the pre-fix code returned it:

```
ReadSymbolBody("../outside.go") LEAKED a file from outside the analysed root
```

`filepath.Join(root, "../outside.go")` resolves and `os.ReadFile` succeeds. So the leak
was never bounded by the repo at all. **[NOTE ON HOW I GOT THIS]** my first version of
the test asserted traversal against paths I had never created, so it passed against the
*pre-fix* code too — a quiet pass that proved nothing. Restructuring the fixture to two
levels (`base/repo` as the analysed root, `base/outside.go` as a real file) is what
turned an `[INFERRED]` claim into a measured one. **A traversal assertion whose target
does not exist is indistinguishable from a working guard.**

### Both-directions verification

- Against the fix: `ok github.com/gqls/agentchassis/internal/analysis`, and the whole
  `platform/orchestration/actions` package passes (0.240s) on clean HEAD + my three files.
- Against HEAD's pre-fix `symbolbody.go` in an isolated module: **FAILS on all six** —
  five leaked bodies plus the out-of-root file.

### Shared-tree friction, both instances someone else's

- `go vet ./platform/orchestration/actions/` failed on
  `derive_brand_head_assets_test.go:108` (`assetLock`). **Not mine, and not at HEAD** —
  another session's uncommitted WIP. Checked by extracting `git archive HEAD` to scratch
  and vetting there: clean but for a long-standing `unreachable code` notice.
- `go build ./...` is amber at HEAD anyway: `cmd/reasoningset/main.go:504` unused
  variables, and a mixed-package directory under `docs024…/traffic_probe/`. Neither mine;
  neither in the chassis binary's path. Recorded so the next thread does not chase them.
- Checked all three of my files for a **same-file passenger** before committing
  (`git diff` on each): only my changes present.

### Misstep: I disarmed the landmine verifier by following CLAUDE.md literally

Ran `landmines-sync.py --apply` as CLAUDE.md instructs; it printed
`NEEDS_VERIFICATION:LANDMINES.md#readsymbolbody-…`. Then ran the wrapper that acts on
that signal, which answered **"Nothing needs verification"** — because the wrapper greps
the output of the `--apply` *it* runs, and mine had already written the row. **The signal
is consumed by the write.** Exit 0, reassuring totals, no check performed. Recovered with
`trigger-landmine-verifier.sh` directly (corr `329710d1`). Full entry in `WRONG_CALLS.md`;
the fix for the next session is in this workstream's RUNBOOK. CLAUDE.md's landmines
bullet names only the inner script — flagged, not silently edited, since that file is
co-edited.

### On not running 090 (declared, per the owner ruling of 2026-07-31)

That ruling requires a `bugs_open/` file asserting a cross-cutting root cause to go
through the diagnosis loop **or for the filing session to state plainly why it
substituted equivalent first-hand verification.** I am not filing a new cross-cutting
claim — 145 was filed by the architecture seat and I am *narrowing* it (correcting its
reachability claim and raising its severity). My substitute, stated so it can be judged:
I read every function in the chain end to end rather than grepping it
(`ReadSymbolBody` → `:201` consumer → `:597` footer → `:139-140` fallback chain →
`EncodeScope` → `namedScope`/`nextScope` → `VerdictWire.NextScope`), and I have a test
that **reproduces the leak against the pre-fix binary and fails when the fix is present**
— which is stronger evidence than a verdict, because it is re-runnable by anyone. The
durable claims are also going to the council gate (`bce4caab`) and to the landmine
verifier (`329710d1`), i.e. two independent readers, not none.

### Council gate

Submitted `bce4caab-17b6-4bbb-ba6b-d1e18f196156` **before** committing; committed with
`Council-Submitted:` since forward-only forbids the amend that `Council-Reviewed:` would
need. Guarantee-change test applied (owner ruling 2026-07-29 #1): this **narrows**
`ReadSymbolBody` from "any path under root" to "a file the analyser parsed" — a
restriction, not a new shared capability — so council gate, not RFC. Consumers named in
the submission's `risks` per ruling #3. Register CTXK-002 updated **in the same commit
as the seam**, per the surviving condition (2) of the ordering exemption.

Three things I deliberately did **not** do, all disclosed to the council rather than
quietly dropped:
- **No per-read size bound** (145's candidate 3). An analysed Go file is still unbounded
  at the read, capped only by `maxBodyChars` at the bundle.
- **Did not fix an adjacent pre-existing wart I found**: the bundle loop `break`s rather
  than `continue`s when a body would exceed `maxBodyChars`, so **one oversized body
  silently drops every remaining scope entry**. Separate concern, separate blast radius.
  Flagged to the council in case a reviewer disagrees.
- **Did not edit the archived contextkit copy.** CTXK-002 carries a "byte-for-byte"
  sync obligation, so I measured it first: **already 29 diff lines out of step** before
  my change (archived has unexported `sliceLines`; the chassis exported `SliceLines` on
  2026-07-27). It is in no build and its caller has the guard, so I recorded the drift
  in the register rather than editing a historical snapshot.
