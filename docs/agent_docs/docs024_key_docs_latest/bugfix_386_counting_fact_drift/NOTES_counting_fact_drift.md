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

## 2026-08-25 — M3 run properly, and THREE wrong turns of my own first

The premise that gates the whole durable fix — *every stale rendered value was once the register's
current value* — is **CONFIRMED, 5 of 5**. But I reached it after measuring the wrong surface twice
and drawing a false conclusion once, and the missteps are more instructive than the result.

### Misstep 1 — I scanned `content_data`, which is not a published surface

First premise test regexed `page_components.content_data`. It returned three numbers matching no
register value, one of which (`125`) I started to chase as a possible real finding. It is part of
`"…(122/125 on the first full day; most traffic is search-engine crawlers — always say so)"` — a
register fact's **claim text including its writer guidance**, staged in `content_data`.

Checked whether that guidance reaches the page: `guidance_in_content_data = t` on all three
components, `guidance_in_rendered_html = f` on all three. So it is not published and there is no
second defect — but `content_data` is a staging field holding the whole register snapshot, and
scanning it over-reports by construction.

### Misstep 2 — I then regexed raw `rendered_html`, which is chart markup

Second test extracted every 3+ digit number from `rendered_html` and reported 13 values per page as
"NEVER a register value → real finding": 127, 320, 555, 600, 620, 625, 666, 700, 766, 875, 1200,
9375, and 8125. **Every one is SVG geometry** — `evidence-chart` draws a chart, so viewBox bounds
and coordinates are in the markup. The identical list on all three pages was the tell and I should
have read it as one.

Worse, `8125` and `9375` *did* match former F10 values, so a coordinate coincided with a real
former reading. Had I trusted that, I would have "confirmed" the premise partly on chart geometry.
That is the accidental-support failure mode running in reverse, inside my own measurement.

**The rule: the scanners do not read markup, they read extracted text blocks. Do not hand-roll a
third text-extraction formulation** — which is exactly what the bugfix_380 handoff §3 warns about
when it notes a shared `page_visible_text(uuid)` would stop 601's query being a third formulation.
Use `cmd/claimscan`, which runs the same engine as the gate.

### Misstep 3 — "three pages are convicted" was false

From `content_data LIKE '%11646%'` I concluded three pages were convicted, and said so. The real
engine finds **one**. `digital-asset-recovery` and `index` store 11646 in `content_data` and never
render it — the chart draws only some of the snapshot it is given. **Conviction requires the value
to be RENDERED**, and `content_data` presence does not imply it. Both pages are `page_type=landing`,
which is *not* on `editorialPageTypes` (guide, blog-post, news-index, tool, game), so the skip list
is not the explanation and I should not have reached for it.

### The real measurement

Export honesty asserted first: 91 TSV rows against 91 in the DB, stderr 0 bytes — the documented
4-column recipe with `page_type`, without which every page reads UNKNOWN and the tool disagrees with
the gate it exists to predict.

`go run ./cmd/claimscan -evidence <live> -components <tsv>` → **5 findings across 91 components**,
all on `capabilities` / `evidence-chart`, all five carrying their own `verified 2026-08-23` stamp —
precisely the bug file's §1 list, no more and no less. Zero practice claims, zero suppressed.

Premise test against the archive, all five values:

| fact | value | first held | last held | current? |
|---|---|---|---|---|
| F9-feed-items-collected | 11513 | 2026-08-23 | 2026-08-23 | no |
| F10-feed-items-scored | 10194 | 2026-08-23 | 2026-08-23 | no |
| F11-council-rounds-revise | 428 | 2026-08-23 | 2026-08-23 | no |
| F12-council-rounds-approved | 483 | 2026-08-23 | 2026-08-23 | no |
| F13-council-rounds-rejected | 23 | 2026-08-21 | 2026-08-23 | no |

5 of 5 were the register's own value on the day the component was written (08-23 17:26), and none
is current. **Disconfirmable and it could have failed**: a rounded or paraphrased figure would have
returned no row. So history-exact covers 100% of the motivating case, and Phase B is provably
sufficient for it rather than merely plausible.

### Engine parity FAILED, and it matters for how these numbers are read

`git diff 4c996e1b5..HEAD -- platform/orchestration/datahelpers/` is **not** empty: `claims.go` is
**122 lines ahead** of the rolled chassis, from `52958897f` — the 364 lane's Phase 2 / RFC_053,
moving the number scan to COMPONENT grain, committed 11:50 today and not yet rolled.

So the claimscan run above predicts the **post-roll** gate, not the live one. That is the right
engine to design Phase B against (it ships after a roll too), but it must not be quoted as "what
the fleet does today".

Checked for collision: `52958897f` adds `normaliseSurfaceKey` and surface/grain handling and touches
**none** of `numberSupported`, `ContextTerms`, `Tolerance`, `seriesSupports`. Phase B lives entirely
in the fact-matching path, so this is same-file proximity, not a design conflict — but two sessions
editing `claims.go` means whoever commits takes both edits, and no hook prevents a same-file
passenger. Flagged to the 364 session.

## 2026-08-25 — my claims.go half was SWEPT into another lane's commit, and the cost is a reporting one

`63d95be1f` is my Go-slice commit. It carries five files. It does **not** carry
`platform/orchestration/datahelpers/claims.go`, which is the half the whole design lives in — and I
had named that path explicitly on `git commit`.

Not a loss, and not a mystery: the file was already committed, by `001211abf` (the 364 lane's
"I described a peer's fix … corrected in all five places"). That session swept my uncommitted
working-tree changes into its own commit. This is the exact failure CLAUDE.md describes — committing
per task stops *you* sweeping up *others'* WIP; it cannot stop a session still running `git add -A`
from sweeping up *yours*. By the time my pathspec commit ran there was nothing left to commit for
that path, so it silently carried five files instead of six.

Verified rather than assumed, because a sweep can also mangle: HEAD's `claims.go` carries exactly
one each of `RetainHistory bool`, `History []FactHistoryEntry`, `func (f *EvidenceFact)
historySupports` and the `if f.historySupports(val)` arm, and both full packages pass fresh with
`-count=1` at the HEAD combination. The code is intact.

**The cost is not the code, it is the audit trail, and it is worth naming because it is invisible
otherwise:**

1. **My commit message describes code that is not in my commit.** Anyone bisecting or reviewing
   `63d95be1f` will find the tests and the register entry but not the mechanism they test.
2. **The council trailer is now on the wrong commit.** `098` joins commits to verdicts by trailer.
   My `Council-Submitted: 18dba069` sits on the commit that carries the *tests*; the in-scope
   platform file landed in `001211abf`, which carries no trailer at all. So the coverage report will
   show an in-scope commit as unreviewed while the review that actually covers it is credited
   elsewhere. Nothing dishonest has been written — the trailer asserts a submission, and that
   submission is real and does cover this code — but the join is wrong and no amend can fix it
   (forward-only).
3. **A same-file passenger is the one thing a pathspec commit cannot exclude**, and this is the
   mirror image: not a passenger riding *in*, but my own work riding *out* under someone else's
   message.

Nothing to undo. Recorded here, reported to the 364 lane so their commit's contents are not a
surprise to them either, and the correlation is written down in both this file and CLM-028 so the
verdict can still be traced by hand when `098` shows the gap.

**The check, for anyone whose task spans a file another lane is active in:** after a pathspec
commit, read the yellow scope block and count the files against what you named. It is advisory and
it never blocks, which is exactly why it is easy to scroll past — but a path you named that is
*absent* from it means the file was committed by someone else between your edit and your commit, and
that is worth knowing before you write "shipped" anywhere.
