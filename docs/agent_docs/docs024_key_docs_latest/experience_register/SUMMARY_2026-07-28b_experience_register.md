# Experience register — where we are, 2026-07-28b

*Third in the series, written at the owner's request. The 07-28 summary was written this morning
before the thing had ever run. It has run now, and running it changed what I can honestly claim —
so this one leads with the problems rather than the progress.*

## What we're trying to do

Every time we build a site, the way its parts *behave* gets invented again. A card that opens
something, a list that reveals detail in place, a button that promises a page. Nothing records what
was decided, and that has one cost that matters far more than repetition: **a promise nobody wrote
down cannot be checked.** If no record says this card is supposed to lead somewhere real, a card
leading nowhere isn't detectably wrong. It just looks like a card.

The register is a library of those behaviours, held once and copied per site, in terms specific
enough to be tested.

## Where we've come from

Designed on 24 July from first principles. On 26 July we read nine live implementations instead,
and the design was wrong in sixteen places. Since then: the tables, a validator, two council
approvals, the pipeline that writes entries, the path that binds them to a real site, and the part
that runs the checks.

## What we've done

The register holds **nine entries**, each with a provenance document, every one written through the
validating path rather than inserted by hand — so its contents are exactly the set that passed its
own rules. Nothing in it claims to be approved; the database physically refuses to store an entry as
approved unless at least one of its tests can actually run.

And on 28 July it ran against a real page for the first time.

## Where we are now — the problems, which are the useful part

**Six of the nine entries were refused on first contact with real data**, and every refusal was
correct. One test referred to a value nothing defined — it would have typed the placeholder's own
name into the box and passed. Five tests carried a "must be at least N" setting that **nothing in
our system reads**, including one called *at_least_two_cards* that asserted "at least one" on the
very rule that a carousel's arrows appear only with two or more cards. And all nine had put a
placeholder in the field that decides *when* a pattern gets chosen, which would have made them
permanently unselectable — a failure with no symptom at all.

**Then the first run against a live page called a good page broken.** Three of the four things it
found were mine:

1. **A rule we had already written down as impossible was executed anyway.** One check cannot
   succeed on that component — it looks for the data file's address in the page's own HTML, and
   that page loads its data from a separate script file. We had recorded that. The runner ignored
   the record, ran it, and reported a failure. **The register's first verdict in its life was
   "broken", about a page that is fine.** This workstream's standing rule is *fix the tooling,
   never the page*, and here the tooling was the one lying.

2. **A check that needed a real browser was judged without one** — and that judgement is
   *automatically favourable*. It would have counted as a pass while checking nothing. A
   verification could have rested on it.

3. **A setting I invented and never implemented.** The runner advertised an option that no code
   read. That is precisely the failure our own notes warn about — an unknown setting is silently
   ignored, so a dead one looks exactly like a live one — committed by me, in a file I had written
   the day before.

4. **The entry itself claimed more than it checked.** A test named for counting rows asserted only
   that the container existed — which another test already asserted. Two tests, one fact, and the
   one with the informative name was the empty one.

**And a correction to this morning's summary.** I wrote that the write-time validator catches
"checks the platform cannot execute". That is too strong, and this run is the counter-example. The
validator can tell whether a *kind* of check exists. It cannot tell whether a check can ever
*succeed*, because that depends on how the page is built — **and validation sees the template, never
the page.** So there are two different kinds of impossible, and only one is catchable before you
run anything. That is why running it in report-only mode has to be a first-class thing rather than
a debugging convenience.

**The honest current reading.** Against the page it was harvested from, that entry has five clauses
and **exactly one is checkable today.** It passes. The other four are held back, each naming what
blocks it. That is a thin result and the system now says so rather than rounding it up to a tick.

**What we still cannot check.** Thirty-eight clauses across the nine entries are marked impossible
for us today, and the biggest single blocker — **whether an element carries an attribute** — accounts
for thirteen of them, across **all nine entries**. It is the dead-link rule, the one found
independently reinvented in six places, the reason this workstream exists. We cannot currently check
the rule the register was built to enforce, and now we know that precisely rather than roughly.

**Two things are not finished, and I want them stated plainly rather than implied.** The bind and
run code is committed but **not deployed and not called by any workflow** — what has been proven is
the machinery and the entries, not the wiring. And the lifecycle is **blocked at its first gate**:
approval-per-experience was designed in July and never built, so every entry is a draft, every copy
is a proposal, and nothing can reach "verified" except in report-only mode. I am building that
approval step now, at your instruction.

## Where we're going

**The approval seat**, in progress — a small review panel that judges one entry: does every clause
name something a machine could observe, is anything dishonest, do the tests actually assert what the
clauses claim, and is this already in the register under another name.

**Then wiring**, so binding and checking happen through the platform rather than on my machine.

**Then the tooling gap, in the order the evidence gives.** Attribute assertion first: one capability,
a third of everything currently blocked, and no entry it doesn't affect.

**Then the planner**, so a new site is built *from* the register instead of having entries
retro-fitted to it.

One pattern worth naming, because it recurred five times this week and I only caught the last two
myself: **every defect I introduced was a hand-written list that quietly went stale** — of fields, of
documents to search, of allowed types. Each is now derived from the real thing, or the build fails
when someone adds something and doesn't classify it. The fix that generalised wasn't any of the
corrections; it was removing the second copy of the truth.
