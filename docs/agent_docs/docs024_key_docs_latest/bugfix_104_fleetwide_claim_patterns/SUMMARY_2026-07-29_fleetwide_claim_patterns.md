# SUMMARY — 2026-07-29 — the honesty check now covers the whole estate

Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and the technical log in `NOTES_fleetwide_claim_patterns.md`.

## What we're trying to do

We run fifteen websites that publish claims — prices, counts, capabilities, promises
about our own reliability. We have a checker that refuses to build a page containing a
claim we have already established is false. The aim of this piece of work was to make
that checker cover the whole estate, instead of only the sites where somebody had
remembered to arm it.

## Where we've come from

The checker worked and had worked for weeks. The patterns it looks for were stored per
site, so a lesson learned on one site could not reach any other, and a brand-new site
started with no protection at all. Seven of fifteen sites had some patterns. Eight had
none — including the two we would least want exposed: vetcomparison, which once
published invented prices for three thousand named real veterinary practices, and
idea.uk, the only site taking money from customers.

This was not an unnoticed gap. It had been *filed*, months earlier, as a deliberate
decision with a numeric trigger: keep patterns per-site until at least two sites have
evidence registers. That was the right call when only one site had one. Nine sites have
one now. The trigger fired months ago, nobody was watching it, and a deferral with a
condition nobody re-reads is indistinguishable from a permanent decision.

## What we've done

We measured the proposed fix before shipping it, against every page of every site —
about nine hundred pieces of live copy — using a tool that runs the real checker rather
than an approximation of it. **That measurement is the reason this went well.** It found
that the patterns as proposed would have broken three sites, and that most of what they
blocked would have been us telling the truth: four of seven findings fired on honest,
hedged sentences like *"where manufacturer data has not been independently verified,
that is stated explicitly"*. One pattern out of ten caused all four. It is the strongest
of the ten, and it is now deliberately left out until the checker can tell "this is
verified" from "this is *not* verified".

With that pattern removed and one other narrowed on the owner's ruling, nine patterns
now apply to all fifteen sites, including sites with no register of their own, and
including the sixteenth site on the day someone creates it. It is live, verified in the
running system rather than inferred from a version number, and covered by tests that
prove both directions: the overclaim fails the build, and five kinds of legitimate
sentence still pass.

We also put it through the reviewer council, three rounds, and it is approved. The
approval is the least interesting part. The review found four things we had not: a flaw
in how we measured the blast radius, a missing way to switch the checker off without a
software release, a silent failure mode in the pattern loader, and then a silent failure
mode in the off-switch we had just added. All four are fixed. Three of them are the same
shape as the original bug — something that looks identical whether it is working or
broken.

## Where we are now

Fifteen sites of fifteen are covered. Nothing on the estate fails a build today; the set
was measured to fire on zero of the nine hundred live components. If a pattern turns out
to be wrong, it can be switched off across the whole estate in seconds through
configuration, without a rebuild — and switching it off now announces itself in the logs,
naming the site, so a quietly disarmed site cannot go unnoticed.

The bug is closed. Two things are open and neither is a defect. The strongest pattern
stays out until somebody writes a proper negation guard, which is a separate piece of
work nobody owns. And two reviewers noted that a change of this reach should have a
filed RFC — the substance exists, the artefact does not — which is a decision about
process, for the owner, not something a thread should wave through on its own authority.

## Where we're going

Three things, in the order they matter.

The negation guard is the real prize. It would let the strongest pattern back in, and it
would make the three remaining negatively-phrased patterns safe rather than merely
lucky — today they fire on nothing, which is a fact about our current copy, not a
property of the patterns.

Second, the post-publication half of this layer still does not run at all. Arming a site
arms its *build*; the sweep that would catch drift on already-published pages has been
switched off since early May. That is a different bug, already filed, and it is the
reason "the site is protected" should be read as "new copy is checked".

Third, the thing worth generalising: the deferral that caused all of this had a written
numeric trigger and no watcher. We now have a sensor for a neighbouring version of that
problem. We do not have one for this version — a decision whose stated condition has
quietly come true.
