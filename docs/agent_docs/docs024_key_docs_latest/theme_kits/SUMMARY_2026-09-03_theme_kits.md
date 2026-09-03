# SUMMARY — theme kits, 2026-09-03

The first summary for this lane. Written because Phase 1 went live and the measurements
that followed changed what the work is worth, which is exactly the kind of turn a
summary exists to mark. **New file each time — never an edit of this one.**

---

## What we're trying to do

Give a new site a named "look" it can start from — colours, fonts, page shapes, header
and footer bundled together and pickable by name — and let the site change any of it
freely afterwards. Eventually, be able to make one of those looks out of a site we have
already built, or one we admire elsewhere.

The rule that governs it is the owner's: a look is a starting point and never a
constraint. The machine that designs a site may ignore our set of looks entirely if it
judges better.

## Where we've come from

This was not new ground, which we discovered early and which reshaped the job. The
platform already had a composable design system built in April: palettes, layouts and
font sets combined into a per-site design record, resolved by one agent and rendered by
another. It is live, and it is what our own pitch material calls "composable themes".

So the work became a thin, named bundle sitting *over* that system rather than a
replacement for it. We deliberately avoided calling the new table "themes", because the
word already means "one site's design record" in a dozen places in the code and a second
meaning would have made every reference ambiguous.

We also inherited a precedent worth naming: when the design-resolving agent was first
built, an early draft gave it ownership of navigation and page structure, and that was
walked back within a day. This work follows that line. A kit is something other
mechanisms consult. It takes no write authority away from anyone.

## What we've done

Phase 1 is built, reviewed hard, and live since 2 September. It is a registry of four
named kits, an action that stamps one onto a site, a table of reusable page shapes that
replaces a hardcoded list in the program, and one change to how a site's layout is chosen
so a kit's layout can win.

Two reviews found eight defects before any of it touched the live database, which was
the cheapest possible place to find them. The best of the eight was a uniqueness rule
that could never have fired, making the part of the setup that was supposed to prevent
duplicate rows dead code. We proved the fix by deliberately breaking it.

The owner's ruling arrived mid-build and changed a default. The action originally
deferred to whatever a site already had, which sounded conservative and was in fact a
no-op on most of the fleet, because the classifier had already filled those values in. It
now writes the kit's values and marks them as defaults, so a later reader can tell a
default from a decision. The single thing it will never overwrite is an explicit human
pin.

Since then we have measured what the thing is actually worth, and corrected the record
where it was wrong.

## Where we are now

Live, adopted by nothing, and worth less than we thought.

A kit bundles four dimensions. **Three of them cannot change how a site looks.** Colours
never reach the stylesheet, because a later rendering step writes the eight main colours
itself; we proved that at the artefact on a real site that served none of its
hand-chosen colours. Page shapes apply to fewer than one live page in eighteen, because
the page planner overwhelmingly chooses for itself. And the header and footer are a
no-op: all four kits point at exactly the components an unpinned site already gets.

That leaves the page layout as the only dimension a kit moves. And two of our four kits
pick a layout the system would have chosen anyway — one of them being the layout it falls
back to when it cannot decide, so that kit dresses the default up as a choice. Exactly
one of the four reaches a look the system cannot otherwise reach, and it does so as a
workaround for a defect in how we tag layouts, not as a design decision. We have written
it down that way.

The internal review council asked for changes, correctly. The claim I made about a fix
was true of the code and invisible in my summary of it, and a reviewer can only judge
what is in front of them. Resubmitted today with the guard shown, every disputable figure
carrying a query the reader can run, and one of my own earlier claims retracted as false
rather than quietly dropped.

Three of my written claims turned out to be right in conclusion and false in reason —
including two that named a component that does not exist. None was caught externally.
They surfaced because a reviewer queried an unrelated sentence and the re-measurement
swept them up. That is the calibration lesson from this phase, and it is worth more than
the code: **a conclusion you already believe is where you stop checking the evidence for
it.**

> **CORRECTED the same afternoon, and left visible because the error is the point.** The
> sentence above is itself wrong. The two components DO exist. I had filtered a table on
> its `function` column and drawn a conclusion about its `name` column, in a table where
> those two hold near-identical vocabularies — so I retracted a claim that was true. What
> caught it was grepping the repository before telling another team: seventy files name
> those components and a migration carries guards against overwriting them, and a thing
> that does not exist does not need guards.
>
> So the count is four errors, not three, and the fourth is the worst kind — a retraction
> reads as "someone checked", which makes it more authoritative than the claim it
> replaces. The real lesson is one rung further along than the sentence above: when the
> conclusion keeps surviving while your reasons for it keep failing, the conclusion is
> coming from somewhere other than the evidence you are citing. Go and find where.
>
> The substance of this summary is unaffected. The finding that all four kits pin a
> header and footer an unpinned site already gets still holds, and so does everything in
> the section above about what a kit can and cannot move.

Nothing here is dangerous. No site uses a kit, an unused kit is inert, and every value
stays overridable. The three sites we did touch were verified colour-by-colour before and
after and are unchanged.

## Where we're going

Two open questions for the owner, and one piece of work we are waiting on.

The waiting is on a fleet scoring tool being built in another lane, which will say for
each layout whether no site of that shape exists or whether sites of that shape exist and
keep losing. Only the second kind is worth a kit. Until it lands we are deliberately not
reseeding, because picking by taste is how we got the current four.

The first question is small: a dead feature that lets a palette be specified when a site
is submitted should be either built properly or removed. Building it alone would still
put no colour on a site, so our recommendation is to remove it.

The second question is the real one. Given that three of four dimensions cannot move, is
a kit the right vehicle at all, or is the whole value in making more layouts reachable —
a cheaper and more direct piece of work? Phase 1 is built and harmless, so there is no
cost to leaving it standing while that gets decided. But we would rather ask than build
Phase 2 on an assumption we have just measured as thin.
