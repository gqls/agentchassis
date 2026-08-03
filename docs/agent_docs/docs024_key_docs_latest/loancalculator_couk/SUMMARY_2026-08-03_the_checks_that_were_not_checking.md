# SUMMARY — 2026-08-03 · the checks that were not checking

*Fifth in the series. Follows `SUMMARY_2026-08-02b_ours_to_rebuild.md`.*

---

## What we're trying to do

Take a hand-built website of twenty-seven pages and twelve financial calculators,
and make it something our platform owns and can improve — without changing a
single number any of those calculators produces, and without anyone having to take
that on trust.

The calculators are the hard part. They are public money tools: someone works out
whether to consolidate their debts on one of them. So the standard we set was that
every change has to be *proved*, not reviewed and waved through, and the proof has
to be something a stranger can re-run.

## Where we've come from

The site was adopted whole, page by page, as exact copies. Then each calculator was
rewritten as a proper component and proved to compute identically to the original
across three sets of inputs. Then the pages were broken into their parts, so the
platform can edit the words. Then the twelve calculators were locked, so the
platform can edit the words and *only* the words.

That left a short list of four genuine bugs, found during the rewrite and
deliberately left alone at the time — because fixing a bug in the middle of proving
something is unchanged means neither claim can be checked. This session was that
list.

## What we've done

All four are fixed and live: money that printed to three decimal places, a car
finance tool that did nothing at all at zero per cent, a debt consolidation checker
that counted a debt towards what you owe but not towards what it costs you, and a
"better option" verdict that was told by colour alone.

But the substance of the session is not the four fixes. It is that **our own
checks said green, and for half the fixes that green meant nothing.**

The harness that drives every calculator and compares the answers works out what to
type by scaling the number already in each box — the same, double, half. That is a
sensible way to drive any calculator without hand-writing a script for each one,
and it means it never types anything far from what the page already shows. The car
finance bug only appears at zero per cent, and no doubling or halving of 8.9 gets
you to zero. The consolidation bug only appears when a box is left *empty*, and the
harness fills every box it can find. Both bugs were unreachable. The harness had
been reporting a confident pass over both of them all along, and would have gone on
doing so after the fix.

So we built a second, smaller check that types the awkward values on purpose — and,
because a check nobody has watched fail is worth very little, one that rebuilds each
calculator *as it was before the fix* and requires the same test to come out
differently. All four fixes now clear that bar. Three deliberate "nothing should
change here" cases confirm the working paths were left alone.

Three near-misses were caught on the way, and they are the same shape as each other:

- A new piece of text added to a calculator would have shipped as an **empty
  space**. We had assumed the system falls back to the default written in the
  component's own definition. It does not — it only ever reads what is stored
  against that specific page, and a value it was never given comes out blank with
  no error. The accessibility badge — the entire fix — would have been invisible on
  a page passing every automatic check we have.
- The before-and-after check **stopped being a check the moment we committed the
  fix**, because its "before" was "the latest commit". Caught only because we had
  already changed it to compare the actual readings rather than pass/fail; the
  earlier version would have reported all four as proven.
- The tool that renders "before" and "after" was **caching by which calculator, not
  which version**, so it handed back the old one twice and called them identical —
  wrong for exactly the two fixes that changed only code, which is to say wrong
  precisely where a wrong answer reads *"nothing to do, your fix is already there."*

And then the deploy itself. We followed our own written procedure to rebuild the
pages, and the job came back **complete** with the pages unchanged — the fixes
sitting in the database doing nothing. When the rebuilder wants to know which
calculator a slot holds, it looks the slot up *by name*; when we took this site
apart we named its slots by position, deliberately, so a missing paragraph would
name itself in the error. Those names match nothing, so the rebuilder finds nothing,
quietly keeps what was already there, and reports success.

That is filed as bug 182, and the diagnosis loop **confirmed it on the first pass**,
reaching the mechanism independently and citing the same lines — plus one step we
had missed, which sharpens the fix. It reaches six sites; this one is the extreme
case at every slot, and the other five are the more dangerous shape, rebuilding most
of a page and silently freezing one or two pieces.

## Where we are now

The four bugs are fixed, live, and proved on the real public pages rather than a
local copy — eight test cases, all passing, driven against production. Ten of the
twelve calculators are byte-for-byte what they were; the two that changed did so
only in the intended ways, and **no arithmetic anywhere on the site is different**
except the rounding we set out to correct. A fresh baseline is recorded and
self-verifies. All twelve calculators are locked again.

We also know something we did not know this morning: the platform cannot rebuild
this site's pages from their components at all, and it says "complete" when it
tries. Everything here goes through a different route — the same one all
twenty-seven pages were originally published through — which is written down, has a
safety check that refuses to write unless it can first reproduce what is already
there, and is now proven twice.

Two traps found here are written into the fleet-wide landmines file, because
neither is about this site: the empty-field one, and the fact that a
scaled-from-defaults test harness can never reach a boundary bug.

## Where we're going

The next thing on the list is the header's link list, which is still maintained by
hand — twenty-five links lifted from the original site. Add a page and it does not
appear; remove one and a dead link stays. Generating it from the pages themselves is
the obvious next mechanism, and it was deliberately kept separate from the
decomposition so that a fault would not have two candidate causes.

After that there is a small tidy owed on one calculator's template, which now costs
almost nothing, and one content question for the owner: a tools page that nothing
links to and that duplicates the calculator on the front page.

One decision is genuinely open. The consolidation fix currently **refuses** to give
a verdict on a half-filled debt row, showing the total you owe — which is a fact —
and dashes where the interest comparison would be. The alternative was to quietly
leave that debt out of the sums. We chose refusal because leaving it out answers a
different question from the one the reader asked without saying so, and this is a
page about being mis-sold. It is a small change if the owner prefers the other.

And the open platform question: whether bug 182 gets fixed at source. Until it
does, this site and five others cannot receive a rebuild from their components, and
the failure looks exactly like a success.
