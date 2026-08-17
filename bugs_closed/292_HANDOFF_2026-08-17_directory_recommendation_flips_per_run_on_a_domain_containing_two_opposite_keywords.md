# 292 — a site's directory recommendation FLIPS BETWEEN RUNS when its domain contains two keywords with opposite verdicts

**Status: CLOSED 2026-08-17 — FIXED AND LIVE.** Shipped in **v1.0.1307** (pods started
17:05:46Z). Evidence, and note it is an ANCESTRY argument rather than a binary read: the
tag-bump commit `e24dc0e6c` (16:36:35Z) is what the running image was built from, and
`git merge-base --is-ancestor e0d662243 e24dc0e6c` returns true. The `build provenance`
startup line had already scrolled out of range (~1h on a busy service), which is why the
documented fallback was used.
**Council-approved** unanimously at round 1, corr `d9ca49ae-1c5d-476c-9059-361ed95531bb`.

> **Do NOT verify this the way I first tried.** Grepping the binary for `e0d662243` returns
> ABSENT on a perfectly good build: the binary carries ONE stamp — the commit it was built
> from — not every ancestor. A discovery grep for `[0-9a-f]{40}` is worse (20 hits, none a
> real commit — Go's internal digit table). Both mistakes and the working recipe are logged
> in `WRONG_CALLS.md`, 2026-08-17.

Prior status (kept for the record): FIXED IN TREE, COMMITTED, INERT UNTIL THE NEXT FLEET ROLL.
The running binary at filing time was **v1.0.1305**, which carried the DEFECT — and which the
cluster then re-served from cache for a further day because the tag was reused (`aa9c7b74f`).

**Filed 2026-08-17 by the portfolio_positioning lane**, found as a *pre-flight* for the
Phase C pilot — before dispatching a build, not after a symptom. There is no failure row
anywhere for this; nothing recorded it, because both outcomes look like normal operation.

## Why this is filed without a `090` run (declared substitution, per the owner ruling of 2026-07-31)

The ruling allows first-hand verification in place of the diagnosis loop **if the session
says so plainly**. What was done instead, and why it is stronger here:

- **The defect is REPRODUCED, not reasoned about.** A test asserting the correct outcome
  fails on the unfixed code — `mortgage-refinance.co.uk` resolved to
  `Recommended=false, kind=""` **on iteration 1** — and passes 600×3 after the fix.
- **Reachability is COMPUTED against the real register**, not assumed (table below).
- The blast radius is one function, read in full. There is no "the cause may live somewhere
  else" question for the loop to answer, which is the thing `090` is for.

## Mechanism

`matchVerticalDirectory` (`platform/orchestration/actions/feed_directory_recommendation_action.go`)
builds a signal list, then returns on the **first exact match**:

```go
signals := []string{lower(industry), lower(siteType), lower(category)}
for keyword := range verticalDirectoryMap {                 // <-- RANDOMISED
    if strings.Contains(domainLower, strings.ReplaceAll(keyword, " ", "")) {
        signals = append(signals, keyword)                  // appends ALL matches
    }
}
for _, signal := range signals {
    if config, ok := verticalDirectoryMap[signal]; ok { return &config }   // FIRST wins
    ...
}
```

`verticalDirectoryMap` **deliberately mixes recommending and NOT-recommending entries** —
`"mortgage"`/`"savings"`/`"banking"`/`"insurance"`/`"health insurance"` recommend, and
`"finance"` explicitly does not ("too generic to choose a provider class; a wrong directory
is worse than none"). Go randomises `for k := range map`. So when a domain contains two
keywords on opposite sides, **which one lands first in `signals` is random per run, and the
verdict follows it.**

**The sharpest part: the file already knows this.** The PARTIAL-match arm twenty lines below
carries a comment explaining that map order is random and this map mixes verdicts, and picks
the longest key for exactly that reason. The same defect was sitting one level up, in the
loop that BUILDS the list. A guard was written for the inner case and the outer case was
left — which is why "we already handled that" is not evidence that a class is closed.

## Reachability — computed against the live portfolio register, 2026-08-17

| domain | keywords matched | verdict |
|---|---|---|
| `remortgagecalculator.uk` (**the Phase C pilot**) | `mortgage` | recommended — **deterministic, pilot unaffected** |
| **`mortgage-refinance.co.uk`** (M4, same register entry as the pilot) | **`mortgage` + `finance`** | **FLIPS per run** |
| `remortgagequotation.co.uk`, `remortgagequote.uk`, `fixedmortgagerates.co.uk`, `adversecreditmortgage.co.uk`, `loanandmortgagecalculator.co.uk` | `mortgage` | recommended, deterministic |
| `equityreleasecalculator.co.uk` | none | no directory (no signal) |

`"refinance"` contains `"finance"` — that is the whole trap, and it is invisible reading the
domain list. Any future `*finance*` domain in a recommending vertical inherits it.

**Why the domain signal is load-bearing at all** (measured, and it makes this worse than it
looks): all four comparable finance sites on the estate carry `industry` **NULL** and
`site_type` `"interactive-platform"` — neither matches the map. So for this family the
domain-derived signal is the **only** one that can fire; there is no explicit signal to fall
back on.

## Damage, stated honestly

**No site has been mis-recommended yet** — the flag is only reachable through the two
consumers wired by migration 432 (2026-08-15), `improvement-sweep` has been disabled
fleet-wide since 2026-08-14, and no finance site has been built. `SELECT` over
`site_specs` shows **zero** sites carrying any finance `content_features` key. This is a
defect caught before it produced damage, which is the only reason the table above is a
prediction rather than a list of casualties.

**What it would have looked like if it had fired**: a remortgage-refinance site silently
built with no lender directory, on one run; with one, on the next. No error, no work item —
`recommended:false` means **NO WRITE** by design (the safe default for an opt-in flag), so
the absence is indistinguishable from "this site was never evaluated". The discovery checks
cannot see it either: they gate on the flag being present, so a site that never got the flag
is never flagged as missing its directory.

## The fix (in tree)

Append **at most one** domain-derived keyword, chosen by the same rule the partial matcher
uses — longest key wins, lexicographic tie-break. For `mortgage-refinance.co.uk`:
`mortgage` (8) beats `finance` (7).

Longest-wins is the **right** answer here, not merely a stable one: the specific provider
class is what a remortgage site needs, and `"finance"` is not-recommended precisely because
it is too generic to pick one. Explicit classification signals still precede any
domain-derived one, so a site the classifier actually understood is never overridden by a
coincidence in its domain string (pinned by a second test).

## How to verify after the next roll

```bash
go test ./platform/orchestration/actions/ -run TestMatchVerticalDirectory_ -count=3
```
Both tests must pass. `-count=3` is not decoration: map order is randomised **per range
statement**, so the pre-fix failure is probabilistic — a single run could pass by luck.
Each case iterates 200×, making a 50/50 flip effectively impossible to hide.

At the artefact, once rolled: `kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq
"<the fix commit sha>" /proc/1/exe` (with a must-be-present and a must-be-absent control),
or read the service's own `build provenance` line and
`git merge-base --is-ancestor <fix-commit> <the stamp>`.

## Relations

- `bugs_open/209` — **same mechanism, different site**: `unified_extractor.go:494` resolves
  a deploy source by ranging a map. Two independent instances now; a third would argue for
  a lint rule rather than a third bug file.
- `DIR-001` (`docs026_concept_register/register/directory-pipeline.md`) — the subsystem.
- `RFC_031` — the enrichment-splice trigger; same action, different concern.
- `portfolio_positioning/NOTES_portfolio_positioning.md`, 2026-08-17 entry — the pre-flight
  that turned this up, and the reason the Phase C pilot was **not** blocked by it.
