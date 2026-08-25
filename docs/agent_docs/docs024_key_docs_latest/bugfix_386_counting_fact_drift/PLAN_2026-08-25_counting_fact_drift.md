# PLAN 2026-08-25 — bugs_open/386: counting-fact drift

Lane opened 2026-08-25. Bug file:
`bugs_open/386_HANDOFF_2026-08-24_refreshing_a_counting_fact_turns_every_page_that_rendered_the_old_value_into_an_unregistered_claim.md`
(filed by the `bugs_open/364` lane, handed over as a residual by the closed `bugs_open/380` lane).

## The defect, in one paragraph

The evidence register holds counting facts — `SELECT count(*)` metrics re-read by a daily
`evidence-freshness` job. When the job re-reads one it **overwrites the value and discards the
previous reading** (`refresh_evidence_base_action.go:534-544`). Every already-deployed page that
rendered yesterday's number is then convicted as an `unregistered_number` at `error` severity
(`validate_page_content.go:1324-1326`), and error is failing (`:420`, `:449`) — so a page whose
only fault is being a day old cannot be rebuilt to fix itself.

## D1 — the owner's ruling governs the default shape (2026-08-25, bug file §4b, commit 2a9091c7d)

A counting fact is expressed as **"at least" N**, or the claim is **cancelled or minimised** —
prefer not publishing a live counter in prose at all. Mechanically this promotes fix candidate 3
(`tolerance: "gte"`) to the default and puts *don't mint the claim* above it.

**The reason is worth keeping, because it reframes the bug:** a monotonic counter's honest form
IS a lower bound. "We have collected 11,646 items" is false a minute later; "at least 11,000"
stays true for months and needs no re-render. The drift is a symptom of publishing an exact value
that was only ever true at one instant.

**The ruling does NOT supersede candidate 1** (fact-value history / a range anchored on
verification time). §4b says so explicitly: it remains the durable fix, because it makes the bad
state unrepresentable rather than merely tolerated. So this lane runs the ruling first (config,
live immediately) and the durable fix second (Go, inert until a roll) — not one instead of the
other.

## D2 — a bare `gte` is not the fix, and arming one without the control is the mistake

`numberSupported` gates a fact on its `context_terms` appearing in the ±70-byte window, and that
test is `strings.Contains` — a **substring**, not a word match (`claims.go:1086-1096`, read
2026-08-25). So a `gte` fact vouches for **every smaller number** anywhere near a term it names.

`[MEASURED 2026-08-25]` The live proof the ruling cites is real and now larger than the ruling
says: `ai-agent-orchestration.com` carries `aao-orchestrations` at `gte` with the single broad
term `["orchestration"]` — **value 7281 today, not the 4068 §4b quotes**. The ceiling of what that
one fact silently supports has risen by 3,213 since that figure was taken, by the very mechanism
this bug is about. Correction filed back to the 364 lane.

Therefore, for every fact this lane converts: round the published figure **down** to a stable
threshold and register `gte` at *that* (not at today's count, which re-arms the drift on the next
tick); give the fact **narrow** `context_terms` ("feed items collected", not "items"); and run the
false-negative control below **before** arming, never after.

## D3 — candidate 1 is re-designed as exact former-value history, not a monotonic range

The bug file's candidate 1 says "a fact that is `count(*)` and monotonic is supported by any value
between its previous and current reading". Two problems, both measured:

- **Counters on this estate go down.** `sql_for_agents/218_evidence_facts_for_043_sites.sql:306`
  records `work-items-completed` falling 1,267 → 1,051 because the ledger reaps. A `monotonic`
  flag is false the first time that happens.
- **A range vouches for values nobody published.** `[anchor, current]` supports every number in
  between, which is the accidental-support gradient `bugs_open/364` §2 warns about.

Replace the range with **exact matching against retained former values**: strictly tighter, no
monotonicity assumption, and it vouches only for numbers the register actually held.

`[MEASURED 2026-08-25]` The history is reconstructible today — the refresh supersedes rather than
overwrites the spec row (`refresh_evidence_base_action.go:1289-1318`), leaving **315** superseded
`evidence_base` rows across **15** sites back to **2026-07-16**. fundamentallyai's
`F9-feed-items-collected` reconstructs daily, and **11513 — the exact value its convicted page
renders — is in the archive dated 2026-08-23**. So the backfill payload exists; it is not inferred.

## D4 — the discriminator CLM-027 points at does not exist in code

`nearest_fact_id` has **zero Go hits repo-wide** (grepped 2026-08-25); it is a field in the
auditor LLM's JSON output (`claims_verification/SEED_claims_auditor.sql:70`). And
"`verified_at` newer than the page's render" is undecidable at the build gate anyway: the gate runs
now, against the current register, so both timestamps are today. Correct CLM-027 and the 380
handoff visibly rather than only noting it here.

## D5 — the 090 substitution, stated

CLAUDE.md's default is to run the diagnosis loop before asserting a cross-cutting root cause.
This lane substitutes first-hand verification and says so: the mechanism is a ~50-line path in one
package plus one action, re-read end to end at the file:lines above, and it was *executed* by the
2026-08-24 fleet claimscan census that filed the bug. **One premise is not locally verifiable** —
"every stale rendered value was once the register's current value". M3 below tests it; if it fails,
run `090` before committing any Go slice.

## Scope, measured — the lane is small

`[MEASURED 2026-08-25]` 295 facts fleet-wide in current `evidence_base` specs. **29** are
sql-sourced number facts, on **6** sites. Of those **13 are `exact`** — the blast radius:
fundamentallyai ×6, leopardess ×2, robot-hands ×2, vonc ×3. The other 16 are `gte` and all carry
`context_terms`, so the degrade-to-exact rule at `claims.go:1099` does not catch them.

Consumers of the seam, enumerated rather than asserted: `numberSupported` has exactly **2**
non-test call sites (`claims.go:916` → `ScanUnregisteredNumbers`; `claims_stats.go:327` →
`ScanStatClaims`), reaching the build gate, both discovery checks, the save floor, the revalidator
and `cmd/claimscan`.

## Phases

**Phase 0 (now, docs only).** This lane's standing five. No register touched.

**Phase A — implement the ruling on the 13 exact facts (config/migration, live immediately).**
Gated on the false-negative control, which runs FIRST:
- **M8 (the control, mandatory before arming).** Offline, with the fleet's exact engine (assert
  `git diff <live build stamp> HEAD -- platform/orchestration/datahelpers/` is empty first), diff
  `cmd/claimscan` findings between the live register and a candidate rounded-down `gte` register.
  **Every finding that disappears is something newly vouched for — read each in ≥300 chars of
  context.** If a genuine invention disappears, the terms are too broad: narrow and re-measure.
- **M3 — classify today's convictions.** Export each of the 6 sites' components (assert the
  exported row count against `count(*)` first; the export truncates at exit 0), scan, and join
  each NUMBER finding's value against the reconstructed history. Disconfirming result: a
  stale-looking value matching **no** former value → pages carry rounded or paraphrased figures,
  and D3's design changes.
- Then a migration per site setting the rounded-down `gte` and narrow terms. Council: yes
  (appliable migrations in scope since 2026-08-19). Where a page does not need the number at all,
  take the owner's stronger option and remove it — the only version with no ongoing cost.

**Phase B — the durable fix (Go, inert until the roll, own council round).** Opt-in
`retain_history` + capped `history` on `EvidenceFact`; the refresh appends the outgoing
`{value, verified_at}` before overwriting; `numberSupported` treats an exact match against a
history entry, under the same context-term gate, as supported. This also answers what §4b leaves
open — the register cannot currently say a fact is monotonic, so the convention is enforced by
review rather than by the schema, and "a comment is not a control on a tree this many sessions
share".

**Phase C — re-render on fact refresh (Go, gated).** Only with a bounded structural target set
(`site_plan_sections.assigned_fact_ids`), and only through the existing canonical helper
`insertPageRerenderItem` (`create_rerender_items_action.go:123-166`) — never a second copy of that
row shape.

**Phase D — queue hygiene.** The already-open false items will **not** self-close: the
revalidator's `armGateClaimsStillPresent` rule ("the standard moved, not the copy") is a
deliberate integrity control and must not be "fixed". A human rules and cancels — record the item
ids first, because closing archives the row out of the table you would query.
