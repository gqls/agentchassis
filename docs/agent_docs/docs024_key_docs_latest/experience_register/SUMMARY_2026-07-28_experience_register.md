# Experience register — where we are, 2026-07-28

*Second in the series. The first is `SUMMARY_2026-07-26_experience_register.md`, written when
nothing had been built and the design had just been corrected by harvesting. This one is a
different read-out, which is why it exists: the thing now runs, and running it taught us
something we could not have learned any other way.*

## What we're trying to do

Every time we build a site, the way its parts *behave* is invented from scratch. A card that
opens something, a list that reveals detail in place, a button that promises a page — each build
re-decides all of it, and nothing anywhere records what was decided. That has two costs. The
small one is repetition. The large one is that **a promise nobody wrote down cannot be checked**:
if no record says this card is supposed to lead somewhere real, then a card leading nowhere is
not detectably wrong. It just looks like a card.

The experience register is a library of those behaviours, held once and copied per site. Each
entry says what a thing must do, in terms specific enough to be tested — and the copy for a
particular site says which real page each promise points at. Then "this button goes nowhere" stops
being a matter of somebody noticing.

## Where we've come from

The design was written on 24 July from first principles. On 26 July we went and read nine live
implementations across two sites instead, and the design was wrong in sixteen places — it had no
way to say *"this must NOT be a control"*, which turned out to be the single most valuable clause
in the real code. That finding is the workstream's founding lesson: **build these things from what
shipped, not from what ought to have shipped.**

Since then: the database tables, a validator that checks an entry when it is written rather than
only when it is run, two independent council reviews (both approved), and the pipeline that puts
an entry in.

## What we've done

**The register exists and has things in it.** Nine entries, each with its own provenance document,
written yesterday and today. Every one went in through the validating path — none was inserted by
hand — so the library's contents are exactly the set that passed its own rules. That was deliberate
from the start: a library whose contents never satisfied its own contract is no evidence the
contract means anything.

Nothing in it claims to be approved. Approval is a verdict something else issues; the database now
physically refuses to store an entry as approved unless at least one of its tests can actually run.
That constraint exists because two reviewers pointed out that my earlier answer — "we'll add that
check later" — was a promise, not a safeguard.

**And the first real run refused six of the nine.** Every refusal was correct. One entry used a
placeholder nothing could ever fill, so its test would have typed the placeholder text into the box
and passed. Five checks carried a "minimum count" setting that **no part of our system reads** —
including one named `at_least_two_cards` that in fact asserted "at least one", on the very clause
that says a carousel's arrows should only appear when there are two or more cards. And all nine
entries had put a placeholder into the field that decides *when the pattern gets chosen*, which
would have made them unselectable for ever.

Two of the six failures were mine, in the checker rather than the entries, and both were the same
mistake I'd made twice already this week: a hand-written list of things to look at, which quietly
omitted something. Fixed the way the others were — stop keeping a list.

## Where we are now

The register holds **nine entries, twenty-nine working checks, and twenty-three checks it has
honestly marked as impossible for us to run today.** That last number is the interesting one.

We already believed our testing was missing some capabilities — that was written down on Sunday as
a rough count from reading. The register now answers it from data, and sorted by what each missing
capability would buy:

| what we can't test | clauses blocked | entries affected |
|---|---|---|
| **whether an element carries an attribute** | **7** | **7 of 9** |
| waiting for something slow (we assert after 0.3s; the real thing takes 8–23s) | 3 | 2 |
| whether a linked page actually loads | 2 | 2 |
| keyboard focus behaviour | 2 | 1 |
| asserting something is *absent* | 2 | 2 |
| "at least N of these" | 2 | 2 |
| the rest (navigation, ordering, failure states, emptiness) | 5 | 5 |

The top row is the finding. Checking whether an element carries an attribute is blocked on seven
clauses in seven of the nine entries — and it is *the anti-dead-link rule*, the one found in six
independent implementations, the reason this workstream was started. **We cannot currently check
the rule the register exists to enforce.** Nobody knew that in this form yesterday; it took writing
the entries down in a machine-readable way and asking the system what it could do with them.

That is a prioritised list of work on our testing harness, derived from evidence rather than
opinion, and it is the first thing the register has produced that we did not already know.

## Where we're going

**Next is binding**: taking an entry and pointing it at one real site's actual pages and elements,
then running its checks by themselves. That is the point where this stops being a library and
starts being a safety net — where the four dead carousel links I found by hand on Sunday become
something the system notices without anybody looking.

**Then the harness work**, in the order the table above gives. Attribute assertion first; it is
one capability and it unblocks a third of everything currently deferred.

**Then the planner**, so that a new site is built *from* the register rather than having entries
retro-fitted to it.

One honest caveat: the corrections made today are committed but not yet in the running system, so
the deferral count above slightly undercounts. It will correct itself at the next build, with no
action needed.

---

> **CORRECTION, same day, after the next build shipped.** The table above was written from a
> deferral count I had marked as incomplete, and the completed count changes its force. The
> register holds **38 deferrals, not 23** — the missing 15 were the binding-level ones, which only
> get recorded from v1.0.1189 onward.
>
> **Attribute assertion blocks 13 clauses across all nine entries**, not 7 across 7. It is a third
> of everything deferred, and there is no entry in the register it does not affect. The order of
> the table is unchanged; the gap between first and second roughly doubled.
>
> Also since writing: the **consumer** is built. The register can now run a bound entry's checks
> against a live page and record what happened — with the rule that verification requires a
> *pass*, never merely an absence of failures, because a document whose checks were all skipped
> also has no failures.
