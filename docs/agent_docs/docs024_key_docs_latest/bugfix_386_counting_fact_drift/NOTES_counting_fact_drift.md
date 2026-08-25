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
