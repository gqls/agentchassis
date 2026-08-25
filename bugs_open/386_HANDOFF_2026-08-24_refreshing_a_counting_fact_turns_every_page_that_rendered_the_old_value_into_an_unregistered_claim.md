# 386 — refreshing a counting fact turns every page that already rendered the OLD value into an "unregistered claim"

**Filed** 2026-08-24 by the `bugs_open/364` lane, found by a fleet claims census run for that
bug's residual. **Not** 364's mechanism — filed separately so the two are not conflated.
**Severity** medium. Nothing false is published; honest pages are convicted of dishonesty,
and the conviction is at `error` severity, which refuses a rebuild.
**Class** evidence-register drift / detector false positive with a moving reference.
**Status** OPEN, unowned, not fixed. **Owner ruling recorded 2026-08-25 — see §4b** ("at least", or don't make the claim), with the caveat that decides how it must be implemented. Diagnosis below is first-hand and cited; no `090` run.

## 1. The symptom

`cmd/claimscan` over live `rendered_html`, fundamentallyai.com against its own current
register, 2026-08-24 — five findings, all in one `evidence-chart` component on `capabilities`:

```
NUMBER  capabilities  evidence-chart  "11513"  …Feed items collected 11513 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "10194"  …Feed items with a credibility assessment 10194 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "428"    …Council review rounds sent back for revision 428 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "483"    …Council review rounds approved 483 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "23"     …Council review rounds rejected outright by a guardian veto 23 verified 2026-08-23…
```

Every one carries its own `verified <date>` stamp. These are not invented figures — they are
this platform's own metrics, rendered from the register by the component that exists to render
the register.

## 2. Root cause

The register has moved and the page has not. Live `site_specs` (`aspect='evidence_base'`,
`is_current`) for `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`, read 2026-08-24:

| page renders | register now holds |
|---|---|
| 11513 | **11646** |
| 10194 | **10416** |
| 428 | **437** |
| 483 | **503** |
| 23 | 25 |

These are **counting facts** — `SELECT count(*) …` metrics that increase every day. Each refresh
of the register silently invalidates every deployed page that rendered the previous value.
`numberSupported` (`claims.go`) compares the page's number against the fact's **current** value,
with `tolerance` deciding the slack; an `exact` counting fact is wrong the moment the counter ticks.

**The blast radius is every site with a counting fact, not this one page.** The tighter the fact
hygiene (`exact` rather than `gte`), the sooner an honest page is convicted — the same perverse
gradient `bugs_open/364` §2 records for its own mechanism.

## 3. Why it matters more than five findings

- The finding is filed at `error` severity by `validate_page_content`, which **refuses the
  rebuild** — so a page whose only fault is being a day old cannot be rebuilt to fix itself
  without the writer happening to regenerate a different number.
- It is **self-inflicted and periodic**: the register refresh is a scheduled job, so this
  re-arms on its own cadence. Nobody has to do anything wrong for it to recur.
- It is **indistinguishable, in the queue, from a real fabrication.** A reviewer reading
  `unregistered_number "11513"` has no way to tell "the counter moved" from "the writer invented
  a figure" without going to the register and diffing. That is the honest cost: it spends the
  credibility of the finding class.

## 4. Fix candidates, ordered by what closes the door

1. **Make a counting fact's support a range anchored on its verification time**, not a point.
   A fact that is `count(*)` and monotonic is supported by any value between its previous and
   current reading. Needs the register to record that a fact is monotonic — it currently cannot
   say so. This makes the bad state unrepresentable rather than merely rarer.
2. **Re-render on fact refresh.** Whatever writes a new fact value queues a `page_rerender` for
   every page whose stored copy cites the old one. Closes the door on the *page* rather than on
   the *detector*, and is the direction that keeps published copy true.
   ⚠ This is the expensive one and it races the sweep — see the `LANDMINES.md` entry
   "A data repair RACES the sweep that publishes it".
3. **Widen tolerance on counting facts** (`gte`, or `approx_pct`). Cheapest, and it is what the
   fleet has drifted into doing by accident — but a broad `gte` silently vouches for unrelated
   numbers in the same window, which is the accidental-support mechanism `bugs_open/364` §2 warns
   about. **This one trades a false positive for a false negative; do not take it without measuring
   what it starts vouching for.**
4. Suppress the check on components that render the register. Rejected as written: it hides the
   drift instead of resolving it, and an `evidence-chart` is exactly where a wrong figure matters.

## 4b. OWNER RULING 2026-08-25 — "at least", or don't make the claim

**The ruling, in the owner's terms:** a counting fact should be expressed as **"at least" N**, or the
claim should be **cancelled or minimised** — i.e. prefer not publishing a live counter in prose at all.

**What that means mechanically:** it selects fix candidate 3 (`tolerance: "gte"`) as the default
shape for a counting fact, and adds a preference above it — *don't mint the claim* — which candidate
list did not contain because I had framed the choice as "how do we make the check tolerate this"
rather than "should the page say it".

**It is the right call for a reason worth stating: a monotonic counter's honest form IS a lower
bound.** "We have collected 11,646 items" is false one minute later; "at least 11,000 items" stays
true for months and needs no re-render. The drift this bug describes is a symptom of publishing an
exact value that was only ever true at one instant.

### ⚠ The one caveat that decides the implementation — a broad `gte` buys silence, not accuracy

`numberSupported` matches a fact against a number only when the fact's `context_terms` appear in the
±70-byte window around it — **and that test is `strings.Contains`, i.e. a substring, not a word
match.** A `gte` fact therefore vouches for **every number smaller than its value** anywhere near a
term it names. This is measured, not theoretical: ai-agent-orchestration.com carries the fact
`aao-orchestrations`, `gte`, with the single broad `context_terms ["orchestration"]` — and it
silently supports **every** figure below its value sitting next to that word. That is why the first
draft of `bugs_open/364`'s test passed with the fix reverted, and why that bug's §2 calls it
accidental support.

> **CORRECTED 2026-08-25 by the lane that took this bug, and the correction is the bug eating its own
> warning.** I originally wrote the value inline as **`4068`** with no date. Read live the next day it
> was **`7281`** — the silently-vouched-for ceiling had risen **3,213 in about a day**, on the very
> fact I was using to illustrate the danger. A bare count in a landmine-shaped warning is the same
> trap one level up (CLAUDE.md: *a count of things must carry the date it was counted*). **The number
> is now deliberately not quoted**: the load-bearing fact is the SHAPE — one `gte` fact, one broad
> term, a ceiling that climbs on its own — and any figure written here is wrong by tomorrow.

**So the ruling must be implemented with tight `context_terms`, not with a bare `gte`.** A counting
fact converted to `gte` with a broad term is not a fix; it is the check being switched off for a
whole vocabulary, and it will read as "no findings" for ever after.

**Concretely, for whoever takes this:**
1. Round the published figure **down** to a stable threshold and register it `gte` at that
   threshold — not at today's exact count, which reintroduces the drift on the next tick.
2. Give the fact **narrow, specific** `context_terms` (`"feed items collected"`, not `"items"`), and
   check what else on the site sits near those words before committing to them.
3. Where a page does not need the number at all, take the owner's stronger option and remove it —
   that is the only version with no ongoing cost.
4. Candidate 1 (a monotonic range anchored on verification time) remains the durable fix and is
   **not** superseded by this ruling; it is what makes the bad state unrepresentable rather than
   merely tolerated. Candidate 2 (re-render on refresh) stays the right answer for keeping published
   copy true, and is the expensive one.

**Still open after this ruling:** the register cannot currently say a fact is monotonic, so nothing
stops the next author registering an exact counting fact. Until it can, this ruling is a convention
enforced by review, not by the schema — and `bugs_open/364` §5b's lesson applies: a comment is not a
control on a tree this many sessions share.

> ### ⚠ CORRECTED 2026-08-25 — MY MONOTONICITY PREMISE IS FALSE, and the refutation was already in the repo
>
> Points 1 and 4 above, and the "a monotonic counter's honest form IS a lower bound" reasoning in the
> ruling, all assume a counting fact only ever **rises**. **It does not.** `sql_for_agents/218_evidence_facts_for_043_sites.sql:48`
> — written **2026-07-24**, a month before I wrote this, about **this same site** — states it outright:
> *"work items completed is NOT MONOTONIC: 1,267 on 07-24, 1,051 today"*, because its ledger reaps.
>
> **This breaks the ruling's cheap half, not just my candidate.** "At least N" is a false statement
> the moment a reaping counter falls below N, so a floor is not automatically the honest form — it is
> the honest form for a genuinely accumulating count and a *new* false claim for a reaping one. **Ask
> which kind the fact is before applying the ruling to it.** A `monotonic` flag would have encoded my
> wrong premise into the schema.
>
> The lane that took this bug has replaced the range with **exact matching against retained former
> values**, which is strictly tighter, assumes no monotonicity, and vouches only for numbers the
> register actually held. That is feasible because the refresh **supersedes rather than overwrites**
> (`refresh_evidence_base_action.go:1289-1318`), leaving **315** superseded `evidence_base` rows
> across 15 sites back to 2026-07-16 `[MEASURED 2026-08-25]` — so the history the durable fix needs is
> already on disk and does not have to be reconstructed by guesswork.
>
> **How I got it wrong: I did not grep before asserting.** The counter-example was in a migration
> named for the site in my own evidence. Second time in this lane — the URL-form trap that refuted my
> `bugs_open/387` filing was likewise already written down. `WRONG_CALLS.md` 2026-08-25.

> ### Two further findings from the owning lane, 2026-08-25 — both narrow what this ruling can reach
>
> **1. The ruling is already implemented, by hand, on five facts that predate it — so nobody needs to
> invent the template.** `F1-live-sites` is `gte 26` with the writer_line *"more than 10 live
> production sites run on this platform (live count {value}; state a FLOOR, never the exact
> number)"*; same shape on `F2-council-seats`, `C1-records-verified`, `C4-agent-definitions-catalogue`.
> All carry **narrow multi-word** `context_terms`, so they are also the existence proof that the
> ruling is implementable **without** the accidental-support hole warned about above.
>
> **2. ⚠ The ruling cannot reach the case that FILED this bug.** The five facts in §1's evidence
> (F9–F13) have **no `writer_line` at all** — they never reach `composeWriterBlock`, yet still convict
> via `numberSupported`. The stale `11513` is frozen into the `evidence-chart` component's
> `content_data`, written 2026-08-23: **a stored snapshot from the component whose job is rendering
> the register, not LLM prose.** "Express it as at least N" has nothing to attach to — there is no
> sentence, only a chart. Candidate 1 is what covers that case, and for a chart that stamps its own
> *"verified 2026-08-23"* it is the semantically correct answer rather than a tolerance widening,
> because that sentence is true for ever and needs no re-render. Note an assemble-mode rerender would
> republish the same bytes; only `rerender_sections` recomputes `content_data`.
>
> **And the exposure is much smaller than the 13 exact `sql` facts.** Fast movers are F9/F10
> (+~180/day) and F11/F12, plus leopardess `C1-ch-vet-mirror` and `C1-records-enriched`. The rest
> count **enumerable** things whose writer_line names every item (`vonc-archetypes` 8,
> `rh-manufacturers` 6, `vonc-guides` 4, `vonc-tools` 6, `rh-grippers` 10) — and `F14-interactive-tools`
> whose writer_line says *"an EXACT count — do not round it or state a floor"*. **Converting those to
> `gte` would be the accidental-support mistake above, for zero benefit.**

## 5. How to verify

```bash
# the drift, per site — any row where a page's rendered figure is not the register's current one
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT f->>'value', f->>'tolerance', f->>'kind'
FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
WHERE ss.site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
  AND ss.aspect='evidence_base' AND ss.is_current;"
```
Then `cmd/claimscan` over that site's live `rendered_html` — **assert the exported row count
against the DB before trusting the scan**, it truncates silently (`WRONG_CALLS.md`, 2026-08-24).

## 6. Relations

- `bugs_open/364` (the census that found this; different mechanism — that one is about *whose*
  number it is, this one is about *when* the number was true).
- CLM-016 / CLM-014 (the surface gate and how this layer is measured), `numberSupported` and
  `EvidenceFact.Tolerance` in `platform/orchestration/datahelpers/claims.go`.
- `refresh_evidence_base_action.go` / `evidence_citations.go` — the two live writers of the register.
