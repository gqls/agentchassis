# 410 — three independent seams fail toward the quiet default, complete green, and ship nothing

Filed 2026-08-26 by the `dartsonline_traffic` lane, at the `news_editorial` lane's suggestion —
they found the third instance and declined to fix it inside a feature commit, correctly (035 §6.1's
own scope veto). Two of the three are mine. Nobody owns the seam, which is why it is filed rather
than carried.

**This is a pattern file, not a new investigation.** Each instance is documented where it was found;
what is new is the direction they share and what that implies about which defaults are safe.

## The three, one week, three lanes

| # | seam | what happens | where |
|---|---|---|---|
| 1 | a listing re-rendered in assemble mode after its card image lands | the stored array is re-rendered **verbatim**, empty image fields included. Three re-renders, no change, item `complete` | `bugs_open/384` |
| 2 | a `page_rerender` carrying a reason the Go reader does not know | `keyReason=""`, `scoped=false` → unscoped, **assemble-only**, completes green | `bugs_open/404` |
| 3 | `loadStoredSections`' row scan fails | `logger.Warn(...); continue` — the function returns **fewer sections, or none**, with **no error**, and the page renders empty | verified this filing: `rerender_page_sections_action.go:1206`, scan branch at +32 |

## The property they share, and why it is worse than three bugs

**Every one fails toward assemble/skip — the quiet default — and the artefact is left looking
freshly built.** Not blank, not erroring, not obviously stale: *rebuilt*. A completed work item, a
new deploy stamp, a page that renders.

**That default is correct, and that is the problem.** Assemble-only is the right ordinary behaviour
— it cannot escalate a page to the content writer, and it cannot destroy hand-placed in-body imagery
(a loss this estate has already paid for). `rerender-pages` has produced **6,428** `page_rerender`
items of which **3** carry a reason at all `[MEASURED 2026-08-26]`: assemble is the overwhelming
norm and should be. So **the estate's safest mode is also its silent-failure mode**, and every drift,
every unknown value and every skipped row lands in it. Failing toward *re-resolve* would be
self-announcing — too many re-renders, someone notices. Failing toward assemble announces nothing by
construction.

**The one confirmed instance of items actually taking the silent branch**: 7 of 19 `literal_markdown`
work items predate migration 473, which is what taught the gate that reason. They took
`else_step: render_page` and completed green (`bugs_open/404`, 2026-08-26). Whether those pages were
later repaired by another route is **not** established.

## What this is NOT

Not "add more checks". All three seams sit *downstream* of correct detection — in 384 the asset was
right, in 404 the item was filed correctly, in 3 the query returned rows. **The defect is in the
handling, and the handling reports success.** A checker that watches the artefact catches these only
after the fact and only if it is enabled, which is `bugs_open/399`'s finding one seam along.

## Fix candidates, ordered by what closes the door

1. **Make "I did not understand this" a refusal, not a fallback.** An unknown reason, an unparseable
   row, a mode that cannot satisfy the request — each should fail closed and file, not degrade to
   assemble. The cost is real (a refusal on the fleet's busiest pipeline is loud) and that is the
   point: instance 3 sits on `rerender_page_sections`, so this needs its own review, not a feature
   commit. **This is the door-closing fix and the expensive one.**
2. **Parity between a vocabulary and its readers, asserted at commit time** — `bugs_open/404`'s
   candidate 0, reading the *live* condition rather than a pasted copy, and proven by adding a value
   to the fixture and requiring the test to go red. Narrower than 1, needs no design decision.
3. **Make the silent skips countable.** Instance 3 returns fewer rows than it selected and says so
   only in a log. Returning a count, or erroring when `scanned < selected`, converts an invisible
   loss into a number a caller can assert on. Cheapest of the three; catches the class rather than
   the instances.

## Verification, and the trap in it

Whatever ships, **prove it can fail**: add an unknown reason to a fixture and require a refusal;
make a scan fail and require an error. A test asserting over today's known values passes on the day
it is written and can never do anything else — the shape all three lanes logged separately this week.

## 090 substitution, stated

Not run through the loop. Instance 3 read first-hand at the file:line above; instances 1 and 2 are
documented in their own files with their own evidence, and this file deliberately points rather than
restates. What is NOT established: whether instance 3 has ever fired in production — I verified the
branch exists and reports nothing, not that a scan has failed.

## Relations

`bugs_open/384` (instance 1, fixed at the framework) · `bugs_open/404` (instance 2, latent, candidate
0 waiting) · `bugs_open/399` (the same "detectable rather than unrepresentable" argument one seam
along) · `features_open/035` §6.1 (the scope veto that correctly kept instance 3 out of a feature
commit) · `LANDMINES.md` "a stale PAGE holds every improvement since it rendered"

---

## CORRECTIONS 2026-08-26, hours after filing — an attribution, a citation, and a real reproduction

**1. The 6,428 / 3 figure is NOT mine and my `[MEASURED 2026-08-26]` marker implied it was.**
It was measured and supplied by the **`bugs_open/384` lane**; I took it from a message and stamped it
with the estate's own "I measured this" marker. That marker is supposed to distinguish a figure I
checked from one I relayed, and using it on a relayed number is the precise failure the marker rule
exists to prevent — a reader auditing this file would have come to me for the query and I do not have
it. **Corrected: the control is the 384 lane's, dated 2026-08-26, and I have not independently
re-run it.** The argument it supports — that assemble is the correct and overwhelming norm, which is
exactly why every drift lands there — is unaffected; its provenance is.

**2. `rerender_page_sections_action.go:1206` has already expired. Cite the symbol.**
When I filed this, `:1206` was the `rows.Scan` branch. It is now the `loadStoredSections` **function
signature**; the Warn-and-continue is at `:1238`, moved ~32 lines by `bd811fa93` (*"035 P1:
loadStoredSections reads the composition pair"*) **the same afternoon**. Verified just now.

> **Cite it as `loadStoredSections`' `rows.Scan` error branch** (`logger.Warn("rerender_page_sections:
> row scan failed"); continue`). This file is under active edit from at least two lanes, so any line
> number written here expires — including the one I just corrected it to.

Same family as this repo's standing rule against citing `HEAD~1`: **a reference that moves is not a
reference.** Cheap check before quoting a line: `grep -n "<the distinctive string>" <file>`.

**3. The third seam now has a real reproduction, not a hypothetical one.**
`bd811fa93` added two columns to that SELECT. **Six tests went red reporting *"expected exactly one
section, got 0"* — and not one said *"scan mismatch"***. So a genuine change to a genuine query
presented as **an empty page rather than an error**, and the tests encoded the symptom while losing
the cause. That is the seam firing under a routine, correct edit, which is stronger than the
argument the file otherwise makes from construction alone. Supplied by the `news_editorial` lane.

**What this does to candidate 3** (make the silent skips countable): it moves from *cheapest* to
*best evidenced*. Had `loadStoredSections` returned `scanned < selected` as an error, or even a
count, those six tests would have named the cause on the first run instead of six times reporting
its symptom.
