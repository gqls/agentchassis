# 164 — the diagnosis bundle's body cap `break`s instead of skipping, so one oversized symbol silently drops the whole rest of scope

**Filed 2026-07-31 by the `bugfix_145` lane, AT THE COUNCIL'S EXPLICIT REQUEST.** OPEN,
UNOWNED.

I found this while fixing `bugs_open/145` in the same function, disclosed it in my
submission's `risks` as "an adjacent pre-existing wart I am not fixing", and left it for a
reviewer to rule on. The council gate's **`bug_historian` seat objected to exactly that
choice** (corr `bce4caab`, round 1, severity **medium**), and it is right:

> "The plan's own risk section discloses, but declines to fix, an adjacent bug in the exact
> file/loop it is editing… This is not a hypothetical near-miss: it is a byte-for-byte match
> to the indexed transferable pattern **'A hard cap that silently discards its input's tail
> rewrites meaning — and the tail is whatever was composed LAST (2026-07-20)'** (016b §9),
> which is precisely the silent-drop-during-render shape this council exists to catch. The
> author found it while auditing the very function this edit modifies, named it accurately,
> and then left it for 'a reviewer to say so' rather than filing it."
>
> `missing`: "No work item or bug filing accompanies the disclosed maxBodyChars/break
> silent-truncation defect… it should be filed now (referencing the 016b §9 hard-cap
> pattern) rather than left as a reviewer's-discretion footnote."

**So this file exists because a review seat caught a thread (me) doing the thing the seat
was built to catch. Recording that plainly, because the seat's value is the reason to keep
paying for it.**

## The defect

`platform/orchestration/actions/diagnose_assemble_bundle_action.go:197-213` — the loop that
renders in-scope code into the diagnosis bundle:

```go
b.WriteString("## In-scope code\n\n")
total, truncated := 0, false
included := 0
for _, sym := range scope {
    body, err := analysis.ReadSymbolBody(repoRoot, anaOut, sym)
    if err != nil {
        logger.Warn("diagnose_assemble_bundle: could not read body", zap.String("symbol", sym), zap.Error(err))
        continue                                    // ← an ERROR skips and carries on
    }
    if total+len(body) > maxBodyChars {
        truncated = true
        break                                       // ← a CAP HIT abandons the whole rest of scope
    }
    fmt.Fprintf(&b, "### %s\n```go\n%s\n```\n\n", sym, body)
    total += len(body)
    included++
}
```

The two failure paths are inconsistent, and the more common one is the harsher: a symbol
that cannot be READ is skipped (`continue`) and its siblings still render, but a symbol
that merely does not FIT ends the loop (`break`) and **every remaining scope entry is
dropped**, however small.

## Why that is worse than it looks

- **Scope is SORTED, so what gets dropped is not random.** `nextScope`/`namedScope` end with
  `sort.Strings(next.Symbols)` (`pkg/diagnose/loop.go:390`, `:416`). So the casualties are
  whatever sorts last — an alphabetical tail, not a "least relevant" tail. One 60,000-char
  file under `internal/` can silently evict everything under `pkg/` and `platform/`.
- **The verdicter is never told.** `truncated` is set, but the bundle text gets no marker at
  this site — contrast `siblingSignatures`, which does write `_(further files omitted — cap
  reached)_` when *its* cap trips (`:604`). So the model reads a bundle that looks like the
  complete answer to its own `next_scope` request.
- **It corrupts the loop's control signal, not just its content.** The diagnosis loop's
  convergence guard measures whether scope NARROWS. A verdict formed on a silently truncated
  bundle names its `next_scope` from what it could see, so the loop can converge on the
  wrong region and record it as progress. This is the mechanism that makes it a *loop* bug
  and not only a rendering bug.
- **One oversized body is enough**, and `maxBodyChars` defaults to 60,000 while
  `ReadSymbolBody`'s whole-file branch is unbounded at the read (see `bugs_closed/145`) —
  and the bundle **advertises** the whole-file form to the model at `:597` ("put the bare
  file path in next_scope to see it whole"). So the platform invites the model to request
  the exact input that triggers this.

## Measured

`[UNMEASURED — deliberately, and here is why]` I have **not** established how often this
fires in production. The honest reason: attributing a past truncation needs the bundle
artefacts joined against the scope that produced them, and `truncated` is not surfaced per
run in a way I could query in the time I had. **Do not quote this file as evidence that it
has fired.** What IS established is the code path, the sort order, and the absent marker —
all above, all first-hand. The first thing the fixing thread should do is measure it:

```sql
-- bundles at/near the cap, as a rate, with their scope size
SELECT date_trunc('day', created_at) AS d, count(*) AS bundles,
       count(*) FILTER (WHERE length(body) >= 59000) AS near_cap
  FROM diagnosis_artifacts WHERE kind='bundle' GROUP BY 1 ORDER BY 1 DESC LIMIT 14;
```

(`length(body) >= 59000` is a proxy, not the flag — check `metadata` for a `truncated`/
`symbols_in_scope` vs `included` pair first; `:304` and `:328` log both, so the fields may
already exist on the artefact.)

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **`continue`, not `break`, plus a per-symbol refusal line in the bundle.** Skip the body
   that does not fit, keep going, and write `### <sym>\n_(body omitted — %d chars, cap %d)_`
   where it would have gone. The model then sees exactly which symbols it asked for and did
   not get, and can re-ask for them singly. **Makes the bad state unrepresentable:** there is
   no longer a silent difference between "not in scope" and "did not fit".
2. **Report `included` vs `len(scope)` in the bundle text**, not only in the log line at
   `:304`. A count the verdicter can read is the difference between a truncation it can
   route around and one it cannot see.
3. **Budget per symbol rather than first-come-first-served** — `maxBodyChars/len(scope)`
   with slack, the allocator `siblingSignatures` already uses for its per-file share
   (`:590-600`). Reuses an in-file convention instead of inventing one, and stops the
   alphabetically-first symbol spending the whole budget.
4. Bound the read itself so a single body cannot approach the cap. Weakest, and it is the
   candidate `145` already declined for being a knob rather than a boundary — but worth
   noting the two bugs meet here.

Candidates 1 and 2 together are ~10 lines and testable without the cluster.

## How to verify a fix

- A scope of `[big.go, small.go]` where `big.go` alone exceeds the cap must still render
  `small.go`, and must name `big.go` as omitted **in the bundle text**, not only in a log.
- Negative control: an unchanged scope that fits entirely must produce a **byte-identical**
  bundle to before the fix — otherwise the change moves every existing diagnosis's baseline.
- Assert the mechanism fired: a test that passes when the cap is never hit is asserting
  nothing. Force the cap, then assert on the omission marker's presence AND on the
  subsequent symbol's body being present.

## Related

- **`bugs_closed/145`** — same function, same loop; the boundary fix that exposed this.
  `145`'s `risks` section is where I first named it, and its `NOTES` records the decision to
  split rather than fold it in.
- **016b §9**, "A hard cap that silently discards its input's tail rewrites meaning — and
  the tail is whatever was composed LAST (2026-07-20)" — the indexed pattern this matches.
  Add this instance to that entry when fixing.
- Family: `bugs_open/012` (an `output_tokens == max_tokens` truncation persisted as success)
  and MEMORY's *"a `complete` work item is not a repaired artefact"*. Every member of this
  family is a cap that reported success.
