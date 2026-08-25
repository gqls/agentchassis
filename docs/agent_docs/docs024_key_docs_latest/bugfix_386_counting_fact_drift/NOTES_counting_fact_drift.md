# NOTES — bugs_open/386 counting-fact drift

Append-only, newest at the bottom. Technical record: what was tried, what the system actually
said, and every misstep.

---

## 2026-08-25 — lane opened; what was read first-hand

Read end to end rather than grepped: `numberSupported` (`claims.go:1069-1121`), `EvidenceFact`
(`:74-96`), the sql refresh arm (`refresh_evidence_base_action.go:490-547`), the supersede
(`:1289-1318`), the gate (`validate_page_content.go:420`, `:449`, `:1320-1333`), and the series
machinery (`claims_series.go`).

Mechanism confirmed at the code: on any non-`fresh` outcome the refresh sets
`fact["value"] = live; fact["verified_at"] = today` and the previous reading is not kept anywhere
in the fact. The bug file's account is accurate.

### Misstep avoided by reading rather than assuming — the series machinery is NOT a drop-in

First instinct was that `claims_series.go` already solved this: it holds many dated observations
per fact, each with its own source, matched exactly. It is the right *shape*, but
`numberSupported` consults it only when `f.Value == nil` (`claims.go:1077`), and `IsSeries()` keys
on `len(Observations) > 0` (`claims_series.go:82-84`). So pushing superseded readings into
`observations` would (a) never be consulted while the fact keeps a current value, and (b) flip
every armed fact into a series, changing behaviour in `ValidateSeries` and the chart path. History
has to be a **distinct field**. Recording this because "reuse the existing mechanism" was the
right instinct and the wrong conclusion, and only the second read showed why.

### Correction to the bug file — candidate 1's premise is false on this estate

The bug file says counting facts "increase every day" and proposes a `monotonic` flag. But
`sql_for_agents/218_evidence_facts_for_043_sites.sql:306` records `work-items-completed` falling
1,267 → 1,051 because the ledger reaps. A `monotonic` flag would be false the first time that
happens, and a range `[previous, current]` vouches for every intermediate value nobody published.
Exact matching against *retained former values* is strictly tighter and needs no monotonicity
assumption. Caught by reading the repo's own migration, not by the census.

### Correction to CLM-027 and the bugfix_380 handoff — the discriminator does not exist

Both say a rotation finding whose `nearest_fact_id` has a `verified_at` newer than the page's
render is a stale render rather than an invention. `grep -rn nearest_fact` repo-wide returns
**zero Go hits**: it is a field in the auditor LLM's output JSON
(`claims_verification/SEED_claims_auditor.sql:70`), so it exists only on the LLM audit path — not
on the Go `unregistered_number` path that actually blocks the rebuild. And at the build gate the
comparison is undecidable in principle: the gate runs now against the current register, so the
fact's `verified_at` and "now" are both today. To be corrected in the register entry visibly, not
just here.

### The census `[MEASURED 2026-08-25]`

295 facts fleet-wide in current `evidence_base` specs; 29 sql-sourced number facts on 6 sites; 13
of those `exact` — fundamentallyai ×6, leopardess ×2, robot-hands ×2, vonc ×3. Only 2 facts
fleet-wide carry no `context_terms`, and both are already `exact`, so the degrade-to-exact rule is
a no-op there.

Disconfirmable, and it came out the interesting way: the archive **does** hold the history. 315
superseded `evidence_base` rows across 15 sites back to 2026-07-16, and fundamentallyai's
`F9-feed-items-collected` reconstructs daily — **11513, the exact value the convicted page
renders, is there dated 2026-08-23**. Had the archive been pruned, the whole backfill half of the
plan would have been impossible and today's stale pages would have needed a different remedy.

Second disconfirmable check: the counters moved again overnight. The bug file recorded the
register at 11646 / 10416 / 437 / 503 on 08-24; today it reads 11828 / 10600 / 454 / 542. The
mechanism re-arms daily, as claimed — had they been unchanged, the "self-inflicted and periodic"
argument would have needed re-examining.

## 2026-08-25 later — the owner ruled, and the ruling's own figure is stale

Owner ruling arrived as bug file §4b (commit `2a9091c7d`), relayed by the `bugs_open/364` session:
a counting fact is expressed as "at least" N, or the claim is cancelled or minimised. That
promotes candidate 3 (`gte`) to the default and puts *don't mint the claim* above it. §4b
explicitly does **not** supersede candidate 1, so the lane runs the ruling first and the durable
fix second.

§4b's caveat is the load-bearing part and it checks out: `numberSupported` gates on
`context_terms` via `strings.Contains` against a ±70-byte window (`claims.go:1086-1096`), a
substring test, so a `gte` fact vouches for every smaller number near a term it names.

**But the figure §4b cites is stale, by the very mechanism this bug describes.** It quotes
`ai-agent-orchestration.com` carrying `4068 gte / context_terms ["orchestration"]`. Read live
2026-08-25: the fact is `aao-orchestrations`, the terms are indeed the single broad
`["orchestration"]` — and the value is **7281**. The ceiling of what that one fact silently
supports has risen by 3,213 since the figure was taken. So the caveat is *understated*, not wrong.
Reported back to the 364 lane. This is the third document in two days to quote a counting fact
that had moved by the time it was read, which is itself the argument for the ruling.

## 2026-08-25 later still — the 13 facts do NOT take one remedy, and the motivating case takes the ruling's *other* half

Prompted by an FYI from the `bugs_open/387` session (its lane: the writer shipped the literal
`NNN` placeholder 14 times because the unscoped prompt carries only `writer_block`, never the fact
values). That made me check how each exposed fact's number actually reaches the page before
designing anything, and the answer splits the population.

`[MEASURED 2026-08-25]` `writer_line` + `writer_block_managed` for all 29 sql-sourced number facts:

**(a) The ruling is already implemented on 5 facts, by hand, and it works.** `F1-live-sites` is
`gte` 26 with `writer_line` = *"more than 10 live production sites … (live count {value}; state a
FLOOR, never the exact number)"*. Same shape on `F2-council-seats` ("more than a dozen"),
`C1-records-verified` ("more than 2,000"), `C4-agent-definitions-catalogue` ("more than 150
… ({value} at the last live count)"). A rounded-down floor, with the live value available to the
substituter and an explicit instruction not to state it. **That is the owner's ruling, in
production, predating the ruling.** Phase A copies this template rather than inventing one — and it
is also evidence the ruling is implementable without the accidental-support hole, because all four
carry narrow multi-word terms.

**(b) The five facts that actually convict the page have NO `writer_line` at all.** F9, F10, F11,
F12, F13 — precisely the bug's §1 evidence. `composeWriterBlock` composes from `writer_line`, so
these five contribute **nothing** to the writer's instructions while still being used by
`numberSupported` to convict. The numbers on the convicted page were therefore never written under
instruction from these facts.

Where they came from instead:
```
capabilities | evidence-chart | evidence-chart | comp_updated 2026-08-23 | deployed 2026-08-24
old_in_content_data = t | old_in_rendered = t | new_in_content_data = f
```
The stale value is **frozen into `content_data`**, written on 08-23 when the register said 11513;
today's 11828 is absent from it. So this is a stored snapshot produced by the component that exists
to render the register.

**Three consequences, and they reorder the lane:**

1. **The ruling's prose remedy cannot reach the motivating case.** There is no `writer_line` to
   rewrite, and "express it as at least N" is a *prose* instruction; the convicted content is a
   chart. §4b anticipated this by explicitly preserving candidate 1 — for this component class
   candidate 1 is not a tolerance widening, it is the *semantically correct* answer, because the
   chart already renders its own `verified 2026-08-23` stamp. "11513 verified 2026-08-23" is a true
   statement for ever, and needs no re-render at all. The register simply cannot currently agree.
2. **An assemble-mode rerender would republish the stale bytes.** Only a regenerating rerender
   (`rerender_sections`, i.e. reason ∈ `image_landed` / `section_data_resolved` / `cta_links_stale`)
   recomputes `content_data`. Any Phase C design must pick the reason deliberately; the default
   route is the one that cannot fix this.
3. **The real exposure is far smaller than 13.** Of the 13 `exact` sql facts, the fast movers are
   fundamentallyai F9/F10 (+~180 a day) and F11/F12 (+17 / +39 a day), plus leopardess
   `C1-ch-vet-mirror` and `C1-records-enriched`. The rest are small counts of *enumerable* things —
   `vonc-archetypes` 8 (and its writer_line names all eight), `vonc-guides` 4, `vonc-tools` 6,
   `rh-manufacturers` 6 (names all six), `rh-grippers` 10, and `F14-interactive-tools` 5 whose
   writer_line says *"an EXACT count — do not round it or state a floor"*. For those, exact is the
   honest form and the ruling's stronger option does not apply; converting them to `gte` would be
   the accidental-support mistake for no benefit.

So: **the ruling applies cleanly to about two prose facts on leopardess; the bug's own motivating
damage takes Phase B.** I had committed the order as "ruling first, durable fix second" an hour
ago. That is right as a default and wrong for the case that filed the bug — recorded here rather
than quietly re-ordered.

`writer_block_managed` is `true` on fundamentallyai and leopardess, unset on robot-hands and vonc —
though both unmanaged sites already use `{value}` in every writer_line, which is worth passing back
to 387: whatever blocks unmanaged sites from machine substitution, it is not the absence of
`{value}` in their lines.
